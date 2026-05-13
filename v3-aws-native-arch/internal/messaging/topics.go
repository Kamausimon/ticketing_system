package messaging

// Topic name constants — identical to the former kafka/topics.go so call sites
// only need to change the import alias, not the constant references.
const (
	OrderCreatedTopic          = "order.created"
	OrderPaymentConfirmedTopic = "order.payment_confirmed"
	OrderConfirmedTopic        = "order.confirmed"
	OrderCancelledTopic        = "order.cancelled"
	OrderCompletedTopic        = "order.completed"

	ReservationExpiredTopic = "reservation.expired"
	InventoryReleasedTopic  = "inventory.released"

	EventCancelledTopic = "event.cancelled"

	PaymentCompletedTopic = "payment.completed"
	TicketsGeneratedTopic = "tickets.generated"

	NotificationEmailTopic = "notification.email.requested"

	RefundApprovedTopic  = "refund.approved"
	RefundCompletedTopic = "refund.completed"
)
