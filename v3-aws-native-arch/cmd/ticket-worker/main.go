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
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.TicketWorkerQueue)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.Messaging.TopicARNs)

	log.Println("ticket-worker started, listening on", cfg.Messaging.TicketWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("ticket-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := handlePaymentConfirmed(ctx, db, publisher, msg); err != nil {
			log.Printf("ticket-worker: failed to handle event %s: %v", msg.MessageID, err)
			continue
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("ticket-worker: failed to delete message: %v", err)
		}
	}
}

func handlePaymentConfirmed(ctx context.Context, db *gorm.DB, publisher *messaging.SNSPublisher, msg *messaging.Message) error {
	var evt apievents.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var order models.Order
	if err := db.Preload("OrderItems.TicketClass").First(&order, evt.OrderID).Error; err != nil {
		return fmt.Errorf("order %d not found: %w", evt.OrderID, err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range order.OrderItems {
		var count int64
		tx.Model(&models.Ticket{}).Where("order_item_id = ?", item.ID).Count(&count)
		if count > 0 {
			continue
		}

		for i := 0; i < item.Quantity; i++ {
			ticket := models.Ticket{
				OrderItemID:  item.ID,
				TicketNumber: fmt.Sprintf("TKT-%d-%d-%d-%d", item.TicketClass.EventID, evt.OrderID, item.ID, i),
				HolderName:   fmt.Sprintf("%s %s", order.FirstName, order.LastName),
				HolderEmail:  order.Email,
				QRCode:       fmt.Sprintf("QR-%d-%d-%d", item.TicketClass.EventID, evt.OrderID, i),
				Status:       models.TicketActive,
			}
			if err := tx.Create(&ticket).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("create ticket: %w", err)
			}
		}
	}

	now := time.Now()
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":       models.OrderFulfilled,
		"completed_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update order status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	orderIDStr := fmt.Sprintf("%d", evt.OrderID)

	confirmed := apievents.OrderConfirmedEvent{
		OrderID:   orderIDStr,
		AccountID: fmt.Sprintf("%d", evt.AccountID),
		EventID:   fmt.Sprintf("%d", evt.EventID),
		Timestamp: now,
	}
	confirmedPayload, err := json.Marshal(confirmed)
	if err != nil {
		return fmt.Errorf("marshal order.confirmed: %w", err)
	}
	if err := publisher.Publish(ctx, messaging.OrderConfirmedTopic, orderIDStr, confirmedPayload); err != nil {
		return fmt.Errorf("publish order.confirmed: %w", err)
	}

	generated := apievents.TicketsGeneratedEvent{
		OrderID:   evt.OrderID,
		AccountID: evt.AccountID,
		EventID:   evt.EventID,
		Timestamp: now,
	}
	generatedPayload, err := json.Marshal(generated)
	if err != nil {
		return fmt.Errorf("marshal tickets.generated: %w", err)
	}
	if err := publisher.Publish(ctx, messaging.TicketsGeneratedTopic, orderIDStr, generatedPayload); err != nil {
		return fmt.Errorf("publish tickets.generated: %w", err)
	}

	log.Printf("ticket-worker: generated tickets for order %d", evt.OrderID)
	return nil
}
