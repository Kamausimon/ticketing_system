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

	consumer := kafka.NewConsumer(brokers, kafkatopics.OrderCreatedTopic, "payment-worker")
	defer consumer.Close()

	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("payment-worker started, listening on", kafkatopics.OrderCreatedTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // clean shutdown
			}
			log.Printf("payment-worker: read error: %v", err)
			continue
		}

		if err := handleOrderCreated(ctx, db, producer, msg); err != nil {
			log.Printf("payment-worker: failed to handle event %s: %v", msg.Key, err)
			// Do not commit — Kafka will re-deliver on restart.
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("payment-worker: failed to commit offset: %v", err)
		}
	}
}

func handleOrderCreated(ctx context.Context, db *gorm.DB, producer *kafka.Producer, msg gokafka.Message) error {
	var evt apievents.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var orderID uint
	fmt.Sscanf(evt.OrderID, "%d", &orderID)

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order %d not found: %w", orderID, err)
	}

	// Offline payments are confirmed immediately — no payment gateway involved.
	// Stripe / MPesa payments are confirmed later via their webhooks.
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

	if err := producer.Publish(ctx, kafkatopics.OrderPaymentConfirmedTopic, evt.OrderID, payload); err != nil {
		return fmt.Errorf("publish payment_confirmed: %w", err)
	}

	log.Printf("payment-worker: order %d marked paid (offline)", orderID)
	return nil
}
