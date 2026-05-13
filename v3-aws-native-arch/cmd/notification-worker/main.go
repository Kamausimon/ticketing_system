package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/messaging"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"

	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	_, sqsClient := messaging.NewAWSClients(ctx)
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.NotificationWorkerQueue)
	notifService := notifications.NewNotificationService(cfg)

	log.Println("notification-worker started, listening on", cfg.Messaging.NotificationWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("notification-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := handleOrderConfirmed(ctx, db, notifService, msg); err != nil {
			log.Printf("notification-worker: failed to handle event %s: %v", msg.MessageID, err)
			continue
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("notification-worker: failed to delete message: %v", err)
		}
	}
}

func handleOrderConfirmed(_ context.Context, db *gorm.DB, notifService *notifications.NotificationService, msg *messaging.Message) error {
	var evt apievents.OrderConfirmedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var orderID uint
	fmt.Sscanf(evt.OrderID, "%d", &orderID)

	var order models.Order
	if err := db.Preload("Event").Preload("OrderItems.TicketClass").First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order %d not found: %w", orderID, err)
	}

	items := make([]notifications.OrderItem, len(order.OrderItems))
	for i, item := range order.OrderItems {
		items[i] = notifications.OrderItem{
			Name:     item.TicketClass.Name,
			Quantity: item.Quantity,
			Price:    float64(item.UnitPrice),
			Currency: order.Currency,
		}
	}

	data := notifications.OrderConfirmationData{
		CustomerName: fmt.Sprintf("%s %s", order.FirstName, order.LastName),
		OrderNumber:  evt.OrderID,
		EventName:    order.Event.Title,
		Items:        items,
		Currency:     order.Currency,
		Total:        float64(order.TotalAmount) / 100,
	}

	if err := notifService.SendOrderConfirmationEmail(order.Email, data); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	log.Printf("notification-worker: confirmation email sent for order %d", orderID)
	return nil
}
