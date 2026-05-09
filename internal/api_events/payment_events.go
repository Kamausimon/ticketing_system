package apievents

import "time"

// PaymentCompletedEvent is published (via outbox) when Intasend confirms a
// payment as COMPLETE. The ticket-worker consumes this to create tickets and
// publish TicketsGeneratedEvent.
type PaymentCompletedEvent struct {
	OrderID   uint      `json:"order_id"`
	AccountID uint      `json:"account_id"`
	EventID   uint      `json:"event_id"`
	Amount    int64     `json:"amount"`   // cents
	Currency  string    `json:"currency"`
	Provider  string    `json:"provider"` // "intasend", "stripe"
	Timestamp time.Time `json:"timestamp"`
}

// TicketsGeneratedEvent is published by the ticket-worker after all tickets for
// an order have been created. The ticket-email-worker consumes this to generate
// PDFs and send the ticket confirmation emails.
type TicketsGeneratedEvent struct {
	OrderID   uint      `json:"order_id"`
	AccountID uint      `json:"account_id"`
	EventID   uint      `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}
