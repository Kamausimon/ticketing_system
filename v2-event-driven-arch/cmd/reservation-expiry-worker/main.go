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

	"gorm.io/gorm"
)

const pollInterval = 1 * time.Minute

func main() {
	db := database.Init()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("reservation-expiry-worker started, polling every %s", pollInterval)

	// Run once immediately on startup so we don't wait a full minute after a restart.
	if err := expireReservations(ctx, db, producer); err != nil {
		log.Printf("reservation-expiry-worker: initial run error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := expireReservations(ctx, db, producer); err != nil {
				log.Printf("reservation-expiry-worker: error: %v", err)
			}
		}
	}
}

// expireReservations finds all reservations whose Expires timestamp is in the
// past, deletes them in a single transaction, and publishes one event per
// expired row so downstream workers (waitlist) can react.
func expireReservations(ctx context.Context, db *gorm.DB, producer *kafka.Producer) error {
	var expired []models.ReservedTicket
	if err := db.Where("expires <= ?", time.Now()).Find(&expired).Error; err != nil {
		return fmt.Errorf("query expired reservations: %w", err)
	}

	if len(expired) == 0 {
		return nil
	}

	ids := make([]uint, len(expired))
	for i, r := range expired {
		ids[i] = r.ID
	}

	// Delete all expired rows in one statement.
	if err := db.Where("id IN ?", ids).Delete(&models.ReservedTicket{}).Error; err != nil {
		return fmt.Errorf("delete expired reservations: %w", err)
	}

	log.Printf("reservation-expiry-worker: released %d expired reservations", len(expired))

	// Publish one event per expired reservation.
	// Downstream: InventoryReleasedTopic is consumed by the waitlist worker (issue #5).
	now := time.Now()
	for _, r := range expired {
		if err := publishExpired(ctx, producer, r, now); err != nil {
			// Log and continue — the rows are already deleted, so skipping
			// the event only delays the waitlist notification, it doesn't
			// corrupt inventory state.
			log.Printf("reservation-expiry-worker: publish failed for reservation %d: %v", r.ID, err)
		}
	}

	return nil
}

func publishExpired(ctx context.Context, producer *kafka.Producer, r models.ReservedTicket, now time.Time) error {
	expiredEvt := apievents.ReservationExpiredEvent{
		ReservationID:    r.ID,
		TicketClassID:    r.TicketID,
		EventID:          r.EventID,
		SessionID:        r.SessionID,
		QuantityReleased: r.QuantityReserved,
		Timestamp:        now,
	}
	expiredPayload, err := json.Marshal(expiredEvt)
	if err != nil {
		return fmt.Errorf("marshal reservation_expired: %w", err)
	}
	key := fmt.Sprintf("%d", r.ID)
	if err := producer.Publish(ctx, kafkatopics.ReservationExpiredTopic, key, expiredPayload); err != nil {
		return fmt.Errorf("publish reservation_expired: %w", err)
	}

	releasedEvt := apievents.InventoryReleasedEvent{
		TicketClassID:    r.TicketID,
		EventID:          r.EventID,
		QuantityReleased: r.QuantityReserved,
		Reason:           "reservation_expired",
		Timestamp:        now,
	}
	releasedPayload, err := json.Marshal(releasedEvt)
	if err != nil {
		return fmt.Errorf("marshal inventory_released: %w", err)
	}
	// Key by ticket class so Kafka partitions all events for the same ticket
	// class to the same partition — waitlist workers process them in order.
	tcKey := fmt.Sprintf("%d", r.TicketID)
	if err := producer.Publish(ctx, kafkatopics.InventoryReleasedTopic, tcKey, releasedPayload); err != nil {
		return fmt.Errorf("publish inventory_released: %w", err)
	}

	return nil
}
