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

	consumer := kafka.NewConsumer(brokers, kafkatopics.EventCancelledTopic, "event-cancellation-worker")
	defer consumer.Close()

	notifService := notifications.NewNotificationService(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("event-cancellation-worker started, listening on", kafkatopics.EventCancelledTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("event-cancellation-worker: read error: %v", err)
			continue
		}

		if err := handleEventCancelled(ctx, db, notifService, msg); err != nil {
			log.Printf("event-cancellation-worker: failed to handle event %s: %v", msg.Key, err)
			// Do not commit — Kafka will redeliver so we can retry.
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("event-cancellation-worker: failed to commit offset: %v", err)
		}
	}
}

func handleEventCancelled(_ context.Context, db *gorm.DB, notifService *notifications.NotificationService, msg gokafka.Message) error {
	var evt apievents.EventCancelledEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Load all paid or fulfilled orders for this event.
	// Pending or already-cancelled orders are excluded.
	var orders []models.Order
	if err := db.
		Preload("OrderItems.GeneratedTickets").
		Where("event_id = ? AND status IN ?", evt.EventID, []string{
			string(models.OrderPaid),
			string(models.OrderFulfilled),
		}).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("query orders for event %d: %w", evt.EventID, err)
	}

	if len(orders) == 0 {
		log.Printf("event-cancellation-worker: no paid orders for event %d, nothing to cascade", evt.EventID)
		return nil
	}

	now := time.Now()
	for _, order := range orders {
		if err := cascadeOrder(db, order, evt, now); err != nil {
			// Log per-order failures but continue so one bad order doesn't
			// block all other attendees from being processed.
			log.Printf("event-cancellation-worker: failed to cascade order %d: %v", order.ID, err)
			continue
		}

		// Email the attendee outside the transaction — a send failure should
		// not undo the DB state (refund record already created).
		notifyAttendee(notifService, order, evt)
	}

	log.Printf("event-cancellation-worker: cascaded cancellation for event %d (%d orders)", evt.EventID, len(orders))
	return nil
}

// cascadeOrder cancels one order: creates a refund record, marks the order
// cancelled, and cancels all its tickets — all in a single transaction.
func cascadeOrder(db *gorm.DB, order models.Order, evt apievents.EventCancelledEvent, now time.Time) error {
	// Idempotency: skip if a refund for this order already exists with reason event_cancelled.
	var existing int64
	db.Model(&models.RefundRecord{}).
		Where("order_id = ? AND refund_reason = ?", order.ID, models.RefundEventCancelled).
		Count(&existing)
	if existing > 0 {
		return nil
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create refund record directly in approved state — event cancellations
	// don't need manual approval, they go straight to processing.
	refundNumber := fmt.Sprintf("REF-EV%d-ORD%d-%d", evt.EventID, order.ID, now.Unix())
	refund := models.RefundRecord{
		RefundNumber:    refundNumber,
		RefundType:      models.RefundFull,
		RefundReason:    models.RefundEventCancelled,
		Status:          models.RefundApproved,
		OrderID:         order.ID,
		EventID:         evt.EventID,
		AccountID:       order.AccountID,
		OrganizerID:     evt.OrganizerID,
		OriginalAmount:  order.TotalAmount,
		RefundAmount:    order.TotalAmount,
		OrganizerImpact: order.TotalAmount,
		Currency:        order.Currency,
		RequestedAt:     now,
		ApprovedAt:      &now,
		Description:     fmt.Sprintf("Event '%s' was cancelled by organiser", evt.Title),
	}
	if err := tx.Create(&refund).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create refund record: %w", err)
	}

	// Cancel the order.
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":       models.OrderCancelled,
		"cancelled_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("cancel order: %w", err)
	}

	// Cancel all tickets belonging to this order's items.
	for _, item := range order.OrderItems {
		if len(item.GeneratedTickets) == 0 {
			continue
		}
		ticketIDs := make([]uint, len(item.GeneratedTickets))
		for i, t := range item.GeneratedTickets {
			ticketIDs[i] = t.ID
		}
		if err := tx.Model(&models.Ticket{}).
			Where("id IN ?", ticketIDs).
			Update("status", models.TicketCancelled).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("cancel tickets for order item %d: %w", item.ID, err)
		}
	}

	return tx.Commit().Error
}

func notifyAttendee(notifService *notifications.NotificationService, order models.Order, evt apievents.EventCancelledEvent) {
	subject := fmt.Sprintf("Important: %s has been cancelled", evt.Title)
	body := fmt.Sprintf(
		"Hi %s,\n\nWe're sorry to inform you that %s has been cancelled.\n\n"+
			"A full refund of %s %.2f will be processed to your original payment method within 5-10 business days.\n\n"+
			"Your order reference is: #%d\n\nWe apologise for any inconvenience caused.",
		order.FirstName, evt.Title, order.Currency, float64(order.TotalAmount)/100, order.ID,
	)
	if err := notifService.SendPlainEmail([]string{order.Email}, subject, body); err != nil {
		log.Printf("event-cancellation-worker: failed to email attendee %s for order %d: %v", order.Email, order.ID, err)
	}
}
