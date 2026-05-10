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
	"ticketing_system/internal/outbox"

	gokafka "github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func main() {
	db := database.Init()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	consumer := kafka.NewConsumer(brokers, kafkatopics.InventoryReleasedTopic, "waitlist-worker")
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("waitlist-worker started, listening on", kafkatopics.InventoryReleasedTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("waitlist-worker: read error: %v", err)
			continue
		}

		if err := handleInventoryReleased(db, msg); err != nil {
			log.Printf("waitlist-worker: failed for %s: %v", msg.Key, err)
			// Do not commit — Kafka redelivers so we don't miss a notification.
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("waitlist-worker: failed to commit offset: %v", err)
		}
	}
}

func handleInventoryReleased(db *gorm.DB, msg gokafka.Message) error {
	var evt apievents.InventoryReleasedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Load event and ticket class details needed for the notification email.
	var event models.Event
	if err := db.Preload("Venue").First(&event, evt.EventID).Error; err != nil {
		return fmt.Errorf("load event %d: %w", evt.EventID, err)
	}

	var ticketClass models.TicketClass
	var ticketClassName string
	var ticketPrice float64
	if err := db.First(&ticketClass, evt.TicketClassID).Error; err == nil {
		ticketClassName = ticketClass.Name
		ticketPrice = float64(ticketClass.Price) / 100.0
	}

	venueName := ""
	if len(event.Venue) > 0 {
		venueName = event.Venue[0].VenueName
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://ticketing.local"
	}
	purchaseURL := fmt.Sprintf("%s/events/%d", baseURL, event.ID)

	// Find waiting entries whose requested quantity fits within what was released.
	// Priority DESC → FIFO within same priority, which is the fair queue order.
	var entries []models.WaitlistEntry
	query := db.Where(
		"event_id = ? AND status = 'waiting' AND quantity <= ?",
		evt.EventID, evt.QuantityReleased,
	).Order("priority DESC, created_at ASC")

	// If the release is for a specific ticket class, match that class OR
	// entries that didn't specify a class (they'll take whatever is available).
	query = query.Where("ticket_class_id = ? OR ticket_class_id IS NULL", evt.TicketClassID)

	if err := query.Find(&entries).Error; err != nil {
		return fmt.Errorf("query waitlist for event %d: %w", evt.EventID, err)
	}

	if len(entries) == 0 {
		return nil
	}

	remaining := evt.QuantityReleased
	notified := 0
	now := time.Now()
	expires := now.Add(24 * time.Hour)

	for _, entry := range entries {
		if entry.Quantity > remaining {
			// Not enough inventory left for this entry — skip (they stay in queue).
			continue
		}

		// Mark as notified inside a small transaction so the status update and
		// the outbox email write are atomic.
		tx := db.Begin()
		entry.Status = "notified"
		entry.NotifiedAt = &now
		entry.ExpiresAt = &expires
		if err := tx.Save(&entry).Error; err != nil {
			tx.Rollback()
			log.Printf("waitlist-worker: save entry %d: %v", entry.ID, err)
			continue
		}

		body := buildNotificationBody(entry, event, ticketClassName, ticketPrice, expires, purchaseURL, venueName)
		if err := outbox.QueueEmail(tx, []string{entry.Email},
			fmt.Sprintf("Tickets available: %s", event.Title),
			body,
			"waitlist_notification",
		); err != nil {
			tx.Rollback()
			log.Printf("waitlist-worker: queue email for entry %d: %v", entry.ID, err)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("waitlist-worker: commit entry %d: %v", entry.ID, err)
			continue
		}

		remaining -= entry.Quantity
		notified++

		if remaining <= 0 {
			break
		}
	}

	log.Printf("waitlist-worker: notified %d waitlist entries for event %d (ticket class %d, reason: %s)",
		notified, evt.EventID, evt.TicketClassID, evt.Reason)
	return nil
}

func buildNotificationBody(
	entry models.WaitlistEntry,
	event models.Event,
	ticketClassName string,
	price float64,
	expires time.Time,
	purchaseURL string,
	venueName string,
) string {
	ticketLine := ""
	if ticketClassName != "" {
		ticketLine = fmt.Sprintf("Ticket type:  %s (%.2f KES each)\n", ticketClassName, price)
	}
	venueLine := ""
	if venueName != "" {
		venueLine = fmt.Sprintf("Venue:        %s\n", venueName)
	}

	return fmt.Sprintf(`Hi %s,

Good news! %d ticket(s) for %s are now available for you.

Event details:
%sDate:         %s
%s
You have until %s to complete your purchase — after that your spot will be released to the next person in the queue.

Purchase here: %s

If you no longer need the tickets, you can ignore this email.

Ticketing System`,
		entry.Name,
		entry.Quantity,
		event.Title,
		ticketLine,
		event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"),
		venueLine,
		expires.Format("Monday, January 2, 2006 at 3:04 PM"),
		purchaseURL,
	)
}
