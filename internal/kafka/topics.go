package kafka

const (
	OrderCreatedTopic          = "order.created"
	OrderPaymentConfirmedTopic = "order.payment_confirmed"
	OrderConfirmedTopic        = "order.confirmed"
	OrderCancelledTopic        = "order.cancelled"
	OrderCompletedTopic        = "order.completed"

	ReservationExpiredTopic = "reservation.expired"
	InventoryReleasedTopic  = "inventory.released"
)
