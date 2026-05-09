package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/kafka"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"

	gokafka "github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func main() {
	db := database.Init()
	cfg := config.LoadOrPanic()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	consumer := kafka.NewConsumer(brokers, kafkatopics.OrderConfirmedTopic, "notification-worker")
	defer consumer.Close()

	notifService := notifications.NewNotificationService(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("notification-worker started, listening on", kafkatopics.OrderConfirmedTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("notification-worker: read error: %v", err)
			continue
		}

		if err := handleOrderConfirmed(ctx, db, notifService, msg); err != nil {
			log.Printf("notification-worker: failed to handle event %s: %v", msg.Key, err)
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("notification-worker: failed to commit offset: %v", err)
		}
	}
}

func handleOrderConfirmed(_ context.Context, db *gorm.DB, notifService *notifications.NotificationService, msg gokafka.Message) error {
	var evt apievents.OrderConfirmedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
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
