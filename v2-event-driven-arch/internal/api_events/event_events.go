package apievents

import "time"

// EventCancelledEvent is published when an organiser cancels a live/published event.
// The event-cancellation-worker consumes this to cascade refunds, ticket
// cancellations, and attendee notifications.
type EventCancelledEvent struct {
	EventID     uint      `json:"event_id"`
	OrganizerID uint      `json:"organizer_id"`
	Title       string    `json:"title"`
	Timestamp   time.Time `json:"timestamp"`
}
