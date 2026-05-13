package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/messaging"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	snsClient, sqsClient := messaging.NewAWSClients(ctx)
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.PaymentWorkerQueue)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.Messaging.TopicARNs)

	log.Println("payment-worker started, listening on", cfg.Messaging.PaymentWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("payment-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue // long-poll timeout, no message
		}

		if err := handleOrderCreated(ctx, db, publisher, msg); err != nil {
			log.Printf("payment-worker: failed to handle event %s: %v", msg.MessageID, err)
			// Do not delete — SQS will redeliver after the visibility timeout.
			continue
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("payment-worker: failed to delete message: %v", err)
		}
	}
}

func handleOrderCreated(ctx context.Context, db *gorm.DB, publisher *messaging.SNSPublisher, msg *messaging.Message) error {
	var evt apievents.OrderCreatedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var orderID uint
	fmt.Sscanf(evt.OrderID, "%d", &orderID)

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order %d not found: %w", orderID, err)
	}

	if evt.PaymentMethod != "offline" {
		log.Printf("payment-worker: order %d awaiting %s payment", orderID, evt.PaymentMethod)
		return nil
	}

	now := time.Now()
	if err := db.Model(&order).Updates(map[string]interface{}{
		"status":              models.OrderPaid,
		"payment_status":      models.PaymentCompleted,
		"is_payment_received": true,
		"completed_at":        &now,
	}).Error; err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	confirmed := apievents.OrderPaymentConfirmedEvent{
		OrderID:       evt.OrderID,
		AccountID:     evt.AccountID,
		Amount:        evt.TotalAmount,
		PaymentMethod: evt.PaymentMethod,
		TransactionID: fmt.Sprintf("offline-%d", orderID),
		Timestamp:     time.Now(),
	}
	payload, err := json.Marshal(confirmed)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := publisher.Publish(ctx, messaging.OrderPaymentConfirmedTopic, evt.OrderID, payload); err != nil {
		return fmt.Errorf("publish payment_confirmed: %w", err)
	}

	log.Printf("payment-worker: order %d marked paid (offline)", orderID)
	return nil
}
