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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Each topic gets its own consumer group ID so offsets are tracked independently.
	consumers := []struct {
		topic   string
		groupID string
		handler handler
	}{
		{kafkatopics.OrderCreatedTopic, "analytics-order-created", handleOrderCreated},
		{kafkatopics.OrderConfirmedTopic, "analytics-order-confirmed", handleOrderConfirmed},
		{kafkatopics.PaymentCompletedTopic, "analytics-payment-completed", handlePaymentCompleted},
		{kafkatopics.OrderCancelledTopic, "analytics-order-cancelled", handleOrderCancelled},
		{kafkatopics.EventCancelledTopic, "analytics-event-cancelled", handleEventCancelled},
		{kafkatopics.TicketsGeneratedTopic, "analytics-tickets-generated", handleTicketsGenerated},
	}

	var wg sync.WaitGroup
	for _, c := range consumers {
		wg.Add(1)
		c := c
		go func() {
			defer wg.Done()
			runConsumer(ctx, db, brokers, c.topic, c.groupID, c.handler)
		}()
	}

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
			// Commit anyway — analytics errors shouldn't stall the consumer.
			// A bad message will never succeed on retry.
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("analytics-worker [%s]: commit error: %v", topic, err)
		}
	}
}

// todayFilter scopes an EventStats update to today's daily bucket.
// Without this, every time-granularity row for the event gets updated.
func todayFilter(db *gorm.DB, eventID uint) *gorm.DB {
	today := time.Now().Truncate(24 * time.Hour)
	return db.Model(&models.EventStats{}).
		Where("event_id = ? AND granularity = ? AND day = ?", eventID, models.StatsDaily, today)
}

// handleOrderCreated fires when a customer starts checkout (order created, not yet paid).
// Maps to EventStats.CheckOutStart — the closest field for "checkout initiated".
// Bug fix: the previous version updated a non-existent "total_orders" column.
func handleOrderCreated(_ context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	todayFilter(db, eventID).
		UpdateColumn("check_out_start", gorm.Expr("check_out_start + 1"))

	log.Printf("analytics-worker: order.created for event %d", eventID)
	return nil
}

// handleOrderConfirmed fires after the ticket-worker creates tickets.
// Updates TicketsSold on today's daily stats row.
func handleOrderConfirmed(_ context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.OrderConfirmedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	todayFilter(db, eventID).
		UpdateColumn("tickets_sold", gorm.Expr("tickets_sold + 1"))

	log.Printf("analytics-worker: order.confirmed for event %d", eventID)
	return nil
}

// handlePaymentCompleted fires when Intasend confirms a payment.
// Updates GrossRevenue and SalesVolume on today's daily stats row.
func handlePaymentCompleted(_ context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	amountFloat := float64(evt.Amount) / 100.0

	todayFilter(db, evt.EventID).Updates(map[string]interface{}{
		"gross_revenue": gorm.Expr("gross_revenue + ?", amountFloat),
		"sales_volume":  gorm.Expr("sales_volume + ?", amountFloat),
	})

	log.Printf("analytics-worker: payment.completed for event %d (%.2f %s)", evt.EventID, amountFloat, evt.Currency)
	return nil
}

// handleOrderCancelled fires when a customer cancels an order.
// Decrements TicketsSold so the stats stay accurate.
func handleOrderCancelled(_ context.Context, db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.OrderCancelledEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	// Load the order to get the actual ticket quantity cancelled.
	var order models.Order
	if err := db.Preload("OrderItems").First(&order, evt.OrderID).Error; err != nil {
		// Order not found — log and continue, don't stall the consumer.
		log.Printf("analytics-worker: order.cancelled: order %s not found: %v", evt.OrderID, err)
		return nil
	}

	totalQty := 0
	for _, item := range order.OrderItems {
		totalQty += item.Quantity
	}

	if totalQty > 0 {
		todayFilter(db, eventID).
			UpdateColumn("tickets_sold", gorm.Expr("GREATEST(tickets_sold - ?, 0)", totalQty))
	}

	log.Printf("analytics-worker: order.cancelled for event %d (-%d tickets)", eventID, totalQty)
	return nil
}

// handleEventCancelled fires when an organiser cancels a live event.
// No specific EventStats column for this — we log it for audit purposes.
// A future migration can add a cancelled_at column if needed.
func handleEventCancelled(_ context.Context, _ *gorm.DB, msg gokafka.Message) error {
	var evt apievents.EventCancelledEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	log.Printf("analytics-worker: event.cancelled — event %d (%q) by organiser %d", evt.EventID, evt.Title, evt.OrganizerID)
	return nil
}

// handleTicketsGenerated fires after the ticket-worker creates all tickets for an order.
// No additional EventStats update needed — tickets_sold was already incremented
// by handleOrderConfirmed on the same order flow.
func handleTicketsGenerated(_ context.Context, _ *gorm.DB, msg gokafka.Message) error {
	var evt apievents.TicketsGeneratedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	log.Printf("analytics-worker: tickets.generated for order %d (event %d)", evt.OrderID, evt.EventID)
	return nil
}
