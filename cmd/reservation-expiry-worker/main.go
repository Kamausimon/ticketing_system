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

	"gorm.io/gorm"
)

const pollInterval = 1 * time.Minute

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	snsClient, _ := messaging.NewAWSClients(ctx)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.Messaging.TopicARNs)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("reservation-expiry-worker started, polling every %s", pollInterval)

	if err := expireReservations(ctx, db, publisher); err != nil {
		log.Printf("reservation-expiry-worker: initial run error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := expireReservations(ctx, db, publisher); err != nil {
				log.Printf("reservation-expiry-worker: error: %v", err)
			}
		}
	}
}

func expireReservations(ctx context.Context, db *gorm.DB, publisher *messaging.SNSPublisher) error {
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

	if err := db.Where("id IN ?", ids).Delete(&models.ReservedTicket{}).Error; err != nil {
		return fmt.Errorf("delete expired reservations: %w", err)
	}

	log.Printf("reservation-expiry-worker: released %d expired reservations", len(expired))

	now := time.Now()
	for _, r := range expired {
		if err := publishExpired(ctx, publisher, r, now); err != nil {
			log.Printf("reservation-expiry-worker: publish failed for reservation %d: %v", r.ID, err)
		}
	}

	return nil
}

func publishExpired(ctx context.Context, publisher *messaging.SNSPublisher, r models.ReservedTicket, now time.Time) error {
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
	if err := publisher.Publish(ctx, messaging.ReservationExpiredTopic, key, expiredPayload); err != nil {
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
	tcKey := fmt.Sprintf("%d", r.TicketID)
	if err := publisher.Publish(ctx, messaging.InventoryReleasedTopic, tcKey, releasedPayload); err != nil {
		return fmt.Errorf("publish inventory_released: %w", err)
	}

	return nil
}
