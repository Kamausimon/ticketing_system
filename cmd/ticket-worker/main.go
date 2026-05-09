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
	"time"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/database"
	"ticketing_system/internal/kafka"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/models"

	gokafka "github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func main() {
	db := database.Init()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	consumer := kafka.NewConsumer(brokers, kafkatopics.OrderPaymentConfirmedTopic, "ticket-worker")
	defer consumer.Close()

	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("ticket-worker started, listening on", kafkatopics.OrderPaymentConfirmedTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("ticket-worker: read error: %v", err)
			continue
		}

		if err := handlePaymentConfirmed(ctx, db, producer, msg); err != nil {
			log.Printf("ticket-worker: failed to handle event %s: %v", msg.Key, err)
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("ticket-worker: failed to commit offset: %v", err)
		}
	}
}

func handlePaymentConfirmed(ctx context.Context, db *gorm.DB, producer *kafka.Producer, msg gokafka.Message) error {
	var evt apievents.OrderPaymentConfirmedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var orderID uint
	fmt.Sscanf(evt.OrderID, "%d", &orderID)

	// Load order with all relations needed for ticket generation.
	var order models.Order
	if err := db.Preload("OrderItems.TicketClass").First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order %d not found: %w", orderID, err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range order.OrderItems {
		// Idempotency check — skip if tickets already exist for this item.
		var count int64
		tx.Model(&models.Ticket{}).Where("order_item_id = ?", item.ID).Count(&count)
		if count > 0 {
			continue
		}

		for i := 0; i < item.Quantity; i++ {
			ticket := models.Ticket{
				OrderItemID:  item.ID,
				TicketNumber: fmt.Sprintf("TKT-%d-%d-%d-%d", item.TicketClass.EventID, orderID, item.ID, i),
				HolderName:   fmt.Sprintf("%s %s", order.FirstName, order.LastName),
				HolderEmail:  order.Email,
				QRCode:       fmt.Sprintf("QR-%d-%d-%d", item.TicketClass.EventID, orderID, i),
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

	confirmed := apievents.OrderConfirmedEvent{
		OrderID:   evt.OrderID,
		AccountID: evt.AccountID,
		EventID:   fmt.Sprintf("%d", order.EventID),
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(confirmed)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := producer.Publish(ctx, kafkatopics.OrderConfirmedTopic, evt.OrderID, payload); err != nil {
		return fmt.Errorf("publish order_confirmed: %w", err)
	}

	log.Printf("ticket-worker: generated tickets for order %d", orderID)
	return nil
}
