package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/messaging"
	"ticketing_system/internal/models"
	"ticketing_system/internal/outbox"

	"gorm.io/gorm"
)

const (
	intasendAPIURL     = "https://api.intasend.com/api/v1"
	intasendSandboxURL = "https://sandbox.intasend.com/api/v1"
)

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	secretKey := os.Getenv("INTASEND_SECRET_KEY")
	testMode := os.Getenv("INTASEND_TEST_MODE") == "true"

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	snsClient, sqsClient := messaging.NewAWSClients(ctx)
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.RefundProcessorWorkerQueue)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.Messaging.TopicARNs)

	log.Println("refund-processor-worker started, listening on", cfg.Messaging.RefundProcessorWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("refund-processor-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := handleRefundApproved(ctx, db, publisher, secretKey, testMode, msg); err != nil {
			log.Printf("refund-processor-worker: failed for %s: %v", msg.MessageID, err)
			// Delete anyway — Intasend errors are logged and the admin uses
			// the retry endpoint. Leaving messages undeleted would cause
			// infinite retries and potentially double-charge customers.
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("refund-processor-worker: failed to delete message: %v", err)
		}
	}
}

func handleRefundApproved(ctx context.Context, db *gorm.DB, publisher *messaging.SNSPublisher, secretKey string, testMode bool, msg *messaging.Message) error {
	var evt apievents.RefundApprovedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var refund models.RefundRecord
	if err := db.Preload("Order.PaymentRecords").First(&refund, evt.RefundID).Error; err != nil {
		return fmt.Errorf("load refund %d: %w", evt.RefundID, err)
	}

	if refund.Status == models.RefundProcessing || refund.Status == models.RefundCompleted {
		log.Printf("refund-processor-worker: refund %d already %s, skipping", refund.ID, refund.Status)
		return nil
	}

	var payment *models.PaymentRecord
	for i := range refund.Order.PaymentRecords {
		if refund.Order.PaymentRecords[i].Status == models.RecordCompleted {
			payment = &refund.Order.PaymentRecords[i]
			break
		}
	}
	if payment == nil {
		return fmt.Errorf("refund %d: no completed payment record found", refund.ID)
	}

	now := time.Now()
	refund.Status = models.RefundProcessing
	refund.ProcessedAt = &now
	if err := db.Save(&refund).Error; err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	externalRefundID, err := callIntasend(secretKey, testMode, payment, &refund)
	if err != nil {
		failedAt := time.Now()
		refund.Status = models.RefundFailed
		refund.FailedAt = &failedAt
		db.Save(&refund)
		return fmt.Errorf("intasend: %w", err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	completedAt := time.Now()
	refund.ExternalRefundID = &externalRefundID
	refund.Status = models.RefundCompleted
	refund.CompletedAt = &completedAt
	if err := tx.Save(&refund).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("save completed refund: %w", err)
	}

	if err := tx.Model(&models.Order{}).
		Where("id = ?", refund.OrderID).
		Update("status", models.OrderRefunded).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update order status: %w", err)
	}

	emailBody := fmt.Sprintf(`Dear Customer,

Your refund of %s %.2f (ref: %s) has been processed successfully.

It will appear in your account within 3–5 business days.

Ticketing System`,
		refund.Currency, float64(refund.RefundAmount)/100, refund.RefundNumber,
	)
	if err := outbox.QueueEmail(tx, []string{refund.Order.Email},
		fmt.Sprintf("Refund Completed — %s", refund.RefundNumber),
		emailBody,
		"refund_completed",
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("queue email: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	completedEvt := apievents.RefundCompletedEvent{
		RefundID:         refund.ID,
		RefundNumber:     refund.RefundNumber,
		ExternalRefundID: externalRefundID,
		OrderID:          refund.OrderID,
		Timestamp:        completedAt,
	}
	payload, _ := json.Marshal(completedEvt)
	if err := publisher.Publish(ctx, messaging.RefundCompletedTopic, fmt.Sprintf("%d", refund.ID), payload); err != nil {
		log.Printf("refund-processor-worker: publish refund.completed failed for refund %d: %v", refund.ID, err)
	}

	log.Printf("refund-processor-worker: refund %d (%s) completed, external ID: %s", refund.ID, refund.RefundNumber, externalRefundID)
	return nil
}

func callIntasend(secretKey string, testMode bool, payment *models.PaymentRecord, refund *models.RefundRecord) (string, error) {
	baseURL := intasendAPIURL
	if testMode {
		baseURL = intasendSandboxURL
	}

	body, err := json.Marshal(map[string]interface{}{
		"invoice_id": payment.ExternalTransactionID,
		"amount":     float64(refund.RefundAmount) / 100.0,
		"reason":     string(refund.RefundReason),
	})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/refunds/", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secretKey)

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, respBody)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	if trackingID, ok := result["tracking_id"].(string); ok {
		return trackingID, nil
	}
	return fmt.Sprintf("REF-%d", time.Now().Unix()), nil
}
