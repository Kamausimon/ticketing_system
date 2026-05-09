package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Analytics watches multiple topics — one goroutine per topic,
	// each with its own consumer so offsets are tracked independently.
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		runConsumer(ctx, db, brokers, kafkatopics.OrderCreatedTopic, "analytics-worker-created", handleOrderCreatedAnalytics)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runConsumer(ctx, db, brokers, kafkatopics.OrderConfirmedTopic, "analytics-worker-confirmed", handleOrderConfirmedAnalytics)
	}()

	log.Println("analytics-worker started")
	wg.Wait()
}

type handler func(ctx context.Context, db *gorm.DB, msg gokafka.Message) error

func runConsumer(ctx context.Context, db *gorm.DB, brokers []string, topic, groupID string, h handler) {
	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	log.Printf("analytics-worker listening on %s", topic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("analytics-worker [%s]: read error: %v", topic, err)
			continue
		}

		if err := h(ctx, db, msg); err != nil {
			log.Printf("analytics-worker [%s]: handler error: %v", topic, err)
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("analytics-worker [%s]: commit error: %v", topic, err)
		}
	}
}

func handleOrderCreatedAnalytics(ctx context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	// Increment order count on the event's stats row.
	db.Model(&models.EventStats{}).
		Where("event_id = ?", eventID).
		UpdateColumn("total_orders", gorm.Expr("total_orders + 1"))

	log.Printf("analytics-worker: order.created recorded for event %d", eventID)
	return nil
}

func handleOrderConfirmedAnalytics(ctx context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.OrderConfirmedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	// Increment confirmed (paid) orders on the event's stats row.
	db.Model(&models.EventStats{}).
		Where("event_id = ?", eventID).
		UpdateColumn("tickets_sold", gorm.Expr("tickets_sold + 1"))

	log.Printf("analytics-worker: order.confirmed recorded for event %d", eventID)
	return nil
}
