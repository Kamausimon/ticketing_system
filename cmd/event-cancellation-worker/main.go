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
	"ticketing_system/internal/notifications"

	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	_, sqsClient := messaging.NewAWSClients(ctx)
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.EventCancellationWorkerQueue)
	notifService := notifications.NewNotificationService(cfg)

	log.Println("event-cancellation-worker started, listening on", cfg.Messaging.EventCancellationWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("event-cancellation-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := handleEventCancelled(ctx, db, notifService, msg); err != nil {
			log.Printf("event-cancellation-worker: failed to handle event %s: %v", msg.MessageID, err)
			// Do not delete — SQS redelivers so we can retry.
			continue
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("event-cancellation-worker: failed to delete message: %v", err)
		}
	}
}

func handleEventCancelled(_ context.Context, db *gorm.DB, notifService *notifications.NotificationService, msg *messaging.Message) error {
	var evt apievents.EventCancelledEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

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
			log.Printf("event-cancellation-worker: failed to cascade order %d: %v", order.ID, err)
			continue
		}
		notifyAttendee(notifService, order, evt)
	}

	log.Printf("event-cancellation-worker: cascaded cancellation for event %d (%d orders)", evt.EventID, len(orders))
	return nil
}

func cascadeOrder(db *gorm.DB, order models.Order, evt apievents.EventCancelledEvent, now time.Time) error {
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

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":       models.OrderCancelled,
		"cancelled_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("cancel order: %w", err)
	}

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
