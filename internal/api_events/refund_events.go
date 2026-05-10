package apievents

import "time"

// RefundApprovedEvent is published when a refund is approved and ready for
// gateway processing. The refund-processor-worker calls Intasend on receipt.
// Only the refund ID is carried — the worker loads full details from the DB.
type RefundApprovedEvent struct {
	RefundID     uint      `json:"refund_id"`
	RefundNumber string    `json:"refund_number"`
	Timestamp    time.Time `json:"timestamp"`
}

// RefundCompletedEvent is published by the refund-processor-worker after
// Intasend confirms the refund. Downstream consumers (analytics, notifications)
// can subscribe to this without coupling to the payment gateway.
type RefundCompletedEvent struct {
	RefundID         uint      `json:"refund_id"`
	RefundNumber     string    `json:"refund_number"`
	ExternalRefundID string    `json:"external_refund_id"`
	OrderID          uint      `json:"order_id"`
	Timestamp        time.Time `json:"timestamp"`
}
