package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"sync"
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

	_, sqsClient := messaging.NewAWSClients(ctx)

	// Each topic gets its own SQS queue so messages are tracked independently.
	consumers := []struct {
		queueURL string
		handler  handler
	}{
		{cfg.Messaging.AnalyticsOrderCreatedQueue, handleOrderCreated},
		{cfg.Messaging.AnalyticsOrderConfirmedQueue, handleOrderConfirmed},
		{cfg.Messaging.AnalyticsPaymentCompletedQueue, handlePaymentCompleted},
		{cfg.Messaging.AnalyticsOrderCancelledQueue, handleOrderCancelled},
		{cfg.Messaging.AnalyticsEventCancelledQueue, handleEventCancelled},
		{cfg.Messaging.AnalyticsTicketsGeneratedQueue, handleTicketsGenerated},
	}

	var wg sync.WaitGroup
	for _, c := range consumers {
		wg.Add(1)
		c := c
		go func() {
			defer wg.Done()
			runConsumer(ctx, db, messaging.NewSQSConsumer(sqsClient, c.queueURL), c.handler)
		}()
	}

	log.Println("analytics-worker started")
	wg.Wait()
}

type handler func(ctx context.Context, db *gorm.DB, msg *messaging.Message) error

func runConsumer(ctx context.Context, db *gorm.DB, consumer *messaging.SQSConsumer, h handler) {
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("analytics-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := h(ctx, db, msg); err != nil {
			log.Printf("analytics-worker: handler error: %v", err)
			// Delete anyway — analytics errors shouldn't stall the consumer.
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("analytics-worker: failed to delete message: %v", err)
		}
	}
}

func todayFilter(db *gorm.DB, eventID uint) *gorm.DB {
	today := time.Now().Truncate(24 * time.Hour)
	return db.Model(&models.EventStats{}).
		Where("event_id = ? AND granularity = ? AND day = ?", eventID, models.StatsDaily, today)
}

func handleOrderCreated(_ context.Context, db *gorm.DB, msg *messaging.Message) error {
	var evt apievents.OrderCreatedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	todayFilter(db, eventID).
		UpdateColumn("check_out_start", gorm.Expr("check_out_start + 1"))

	log.Printf("analytics-worker: order.created for event %d", eventID)
	return nil
}

func handleOrderConfirmed(_ context.Context, db *gorm.DB, msg *messaging.Message) error {
	var evt apievents.OrderConfirmedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	todayFilter(db, eventID).
		UpdateColumn("tickets_sold", gorm.Expr("tickets_sold + 1"))

	log.Printf("analytics-worker: order.confirmed for event %d", eventID)
	return nil
}

func handlePaymentCompleted(_ context.Context, db *gorm.DB, msg *messaging.Message) error {
	var evt apievents.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
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

func handleOrderCancelled(_ context.Context, db *gorm.DB, msg *messaging.Message) error {
	var evt apievents.OrderCancelledEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var eventID uint
	fmt.Sscanf(evt.EventID, "%d", &eventID)

	var order models.Order
	if err := db.Preload("OrderItems").First(&order, evt.OrderID).Error; err != nil {
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

func handleEventCancelled(_ context.Context, _ *gorm.DB, msg *messaging.Message) error {
	var evt apievents.EventCancelledEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	log.Printf("analytics-worker: event.cancelled — event %d (%q) by organiser %d", evt.EventID, evt.Title, evt.OrganizerID)
	return nil
}

func handleTicketsGenerated(_ context.Context, _ *gorm.DB, msg *messaging.Message) error {
	var evt apievents.TicketsGeneratedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	log.Printf("analytics-worker: tickets.generated for order %d (event %d)", evt.OrderID, evt.EventID)
	return nil
}
