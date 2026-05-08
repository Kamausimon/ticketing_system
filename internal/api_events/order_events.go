// event models
package apievents

import "time"

type OrderCreatedEvent struct {
	OrderID       string           `json:"order_id"`     // Generated after creation
	OrderNumber   string           `json:"order_number"` // Generated order reference
	AccountID     string           `json:"account_id"`   // From the user
	EventID       string           `json:"event_id"`
	FirstName     string           `json:"first_name"`
	LastName      string           `json:"last_name"`
	Email         string           `json:"email"`
	Amount        float64          `json:"amount"`       // Calculated subtotal
	TaxAmount     float64          `json:"tax_amount"`   // Calculated tax
	TotalAmount   float64          `json:"total_amount"` // Final total
	Currency      string           `json:"currency"`
	Status        string           `json:"status"` // OrderPending
	PaymentStatus string           `json:"payment_status"`
	PaymentMethod string           `json:"payment_method"`
	IsBusiness    bool             `json:"is_business"`
	BusinessName  *string          `json:"business_name,omitempty"`
	Items         []OrderEventItem `json:"items"` // What was ordered
	Timestamp     time.Time        `json:"timestamp"`
}

type OrderEventItem struct {
	TicketClassID   string  `json:"ticket_class_id"`
	TicketClassName string  `json:"ticket_class_name"`
	Quantity        int     `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TotalPrice      float64 `json:"total_price"`
}

type OrderPaymentConfirmedEvent struct {
	OrderID       string    `json:"order_id"`
	AccountID     string    `json:"account_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	Timestamp     time.Time `json:"timestamp"`
}

type OrderConfirmedEvent struct {
	OrderID   string    `json:"order_id"`
	AccountID string    `json:"account_id"`
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderCancelledEvent struct {
	OrderID   string    `json:"order_id"`
	AccountID string    `json:"account_id"`
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderCompletedEvent struct {
	OrderID   string    `json:"order_id"`
	AccountID string    `json:"account_id"`
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}
