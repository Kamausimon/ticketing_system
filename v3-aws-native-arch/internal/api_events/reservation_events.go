package apievents

import "time"

// ReservationExpiredEvent is published when a reserved seat is not converted
// to an order within the expiry window. The worker uses this to publish
// InventoryReleasedEvent so downstream consumers (waitlist) get notified.
type ReservationExpiredEvent struct {
	ReservationID   uint      `json:"reservation_id"`
	TicketClassID   uint      `json:"ticket_class_id"`
	EventID         uint      `json:"event_id"`
	SessionID       string    `json:"session_id"`
	QuantityReleased int      `json:"quantity_released"`
	Timestamp       time.Time `json:"timestamp"`
}

// InventoryReleasedEvent is published whenever seats become available again —
// either from a reservation expiry or an explicit cancellation.
// The waitlist worker (issue #5) consumes this to notify the next person in queue.
type InventoryReleasedEvent struct {
	TicketClassID    uint      `json:"ticket_class_id"`
	EventID          uint      `json:"event_id"`
	QuantityReleased int       `json:"quantity_released"`
	Reason           string    `json:"reason"` // "reservation_expired" | "order_cancelled"
	Timestamp        time.Time `json:"timestamp"`
}
