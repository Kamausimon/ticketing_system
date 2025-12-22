package orders

import (
	"fmt"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"ticketing_system/internal/payments"
	"time"

	"gorm.io/gorm"
)

// OrderHandler handles all order-related operations
type OrderHandler struct {
	db                  *gorm.DB
	metrics             *analytics.PrometheusMetrics
	paymentHandler      *payments.PaymentHandler
	notificationService *notifications.NotificationService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics, paymentHandler *payments.PaymentHandler, notifService *notifications.NotificationService) *OrderHandler {
	return &OrderHandler{
		db:                  db,
		metrics:             metrics,
		paymentHandler:      paymentHandler,
		notificationService: notifService,
	}
}

// OrderResponse represents the order response structure
type OrderResponse struct {
	ID                  uint                 `json:"id"`
	OrderNumber         string               `json:"order_number"`
	AccountID           uint                 `json:"account_id"`
	EventID             uint                 `json:"event_id"`
	EventTitle          string               `json:"event_title"`
	FirstName           string               `json:"first_name"`
	LastName            string               `json:"last_name"`
	Email               string               `json:"email"`
	Amount              float64              `json:"amount"`
	Currency            string               `json:"currency"`
	Status              models.OrderStatus   `json:"status"`
	PaymentStatus       models.PaymentStatus `json:"payment_status"`
	BookingFee          float64              `json:"booking_fee,omitempty"`
	OrganizerBookingFee float64              `json:"organizer_booking_fee,omitempty"`
	Discount            float64              `json:"discount,omitempty"`
	TaxAmount           float64              `json:"tax_amount"`
	TotalAmount         float64              `json:"total_amount"`
	OrderDate           time.Time            `json:"order_date"`
	CompletedAt         *time.Time           `json:"completed_at,omitempty"`
	Items               []OrderItemResponse  `json:"items"`
	IsBusiness          bool                 `json:"is_business"`
	BusinessName        string               `json:"business_name,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

// OrderItemResponse represents order item details
type OrderItemResponse struct {
	ID              uint    `json:"id"`
	TicketClassID   uint    `json:"ticket_class_id"`
	TicketClassName string  `json:"ticket_class_name"`
	Quantity        int     `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TotalPrice      float64 `json:"total_price"`
	Discount        float64 `json:"discount,omitempty"`
	PromoCodeUsed   string  `json:"promo_code_used,omitempty"`
}

// OrderListResponse represents a paginated list of orders
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// OrderFilter represents filtering options for orders
type OrderFilter struct {
	Page          int
	Limit         int
	Status        *models.OrderStatus
	PaymentStatus *models.PaymentStatus
	EventID       *uint
	StartDate     *time.Time
	EndDate       *time.Time
	SearchTerm    string // Search by email, name, order number
	Email         string // Filter by specific email address
}

// OrderStats represents order statistics
type OrderStats struct {
	TotalOrders     int64   `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	PendingOrders   int64   `json:"pending_orders"`
	CompletedOrders int64   `json:"completed_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	RefundedOrders  int64   `json:"refunded_orders"`
	AverageOrder    float64 `json:"average_order_value"`
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	EventID         uint                     `json:"event_id"`
	FirstName       string                   `json:"first_name"`
	LastName        string                   `json:"last_name"`
	Email           string                   `json:"email"`
	IsBusiness      bool                     `json:"is_business"`
	BusinessName    string                   `json:"business_name,omitempty"`
	BusinessTaxID   string                   `json:"business_tax_id,omitempty"`
	BusinessAddress string                   `json:"business_address,omitempty"`
	PromoCode       string                   `json:"promo_code,omitempty"`
	Items           []CreateOrderItemRequest `json:"items"`
	PaymentMethod   string                   `json:"payment_method"` // "stripe", "mpesa", "offline"
}

// CreateOrderItemRequest represents an item in the order
type CreateOrderItemRequest struct {
	TicketClassID uint `json:"ticket_class_id"`
	Quantity      int  `json:"quantity"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status"`
	Reason string             `json:"reason,omitempty"`
}

// ProcessPaymentRequest represents the payment processing request
type ProcessPaymentRequest struct {
	PaymentMethod string  `json:"payment_method"`
	PaymentToken  string  `json:"payment_token,omitempty"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	ReturnURL     string  `json:"return_url,omitempty"`
	CancelURL     string  `json:"cancel_url,omitempty"`
}

// RefundOrderRequest represents the refund request
type RefundOrderRequest struct {
	Reason         string  `json:"reason"`
	Amount         float64 `json:"amount,omitempty"`       // Partial refund amount, omit for full refund
	RefundItems    []uint  `json:"refund_items,omitempty"` // Specific order item IDs
	NotifyCustomer bool    `json:"notify_customer"`
}

// OrderCalculation represents order cost breakdown
type OrderCalculation struct {
	Subtotal            float64 `json:"subtotal"`
	BookingFee          float64 `json:"booking_fee"`
	OrganizerBookingFee float64 `json:"organizer_booking_fee"`
	Discount            float64 `json:"discount"`
	TaxAmount           float64 `json:"tax_amount"`
	TotalAmount         float64 `json:"total_amount"`
	PromoCodeApplied    string  `json:"promo_code_applied,omitempty"`
	Currency            string  `json:"currency"`
}

// Helper function to convert models.Order to OrderResponse
func convertToOrderResponse(order models.Order) OrderResponse {
	response := OrderResponse{
		ID:            order.ID,
		OrderNumber:   generateOrderNumber(order.ID),
		AccountID:     order.AccountID,
		EventID:       order.EventID,
		FirstName:     order.FirstName,
		LastName:      order.LastName,
		Email:         order.Email,
		Amount:        float64(order.Amount),
		Currency:      order.Currency,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		TaxAmount:     float64(order.TaxAmount),
		IsBusiness:    order.IsBusiness,
		OrderDate:     order.CreatedAt,
		CompletedAt:   order.CompletedAt,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	// Add optional fields
	if order.BookingFee != nil {
		response.BookingFee = float64(*order.BookingFee)
	}
	if order.OrganizerBookingFee != nil {
		response.OrganizerBookingFee = float64(*order.OrganizerBookingFee)
	}
	if order.Discount != nil {
		response.Discount = float64(*order.Discount)
	}
	if order.BusinessName != nil {
		response.BusinessName = *order.BusinessName
	}

	// Calculate total amount
	response.TotalAmount = response.Amount + response.BookingFee + response.TaxAmount - response.Discount

	// Add event title if loaded
	if order.Event.ID > 0 {
		response.EventTitle = order.Event.Title
	}

	// Convert order items
	response.Items = make([]OrderItemResponse, len(order.OrderItems))
	for i, item := range order.OrderItems {
		response.Items[i] = convertToOrderItemResponse(item)
	}

	return response
}

// Helper function to convert models.OrderItem to OrderItemResponse
func convertToOrderItemResponse(item models.OrderItem) OrderItemResponse {
	response := OrderItemResponse{
		ID:            item.ID,
		TicketClassID: item.TicketClassID,
		Quantity:      item.Quantity,
		UnitPrice:     float64(item.UnitPrice),
		TotalPrice:    float64(item.TotalPrice),
	}

	if item.Discount != nil {
		response.Discount = float64(*item.Discount)
	}
	if item.PromoCodeUsed != nil {
		response.PromoCodeUsed = *item.PromoCodeUsed
	}
	if item.TicketClass.ID > 0 {
		response.TicketClassName = item.TicketClass.Name
	}

	return response
}

// Helper function to generate order number
func generateOrderNumber(orderID uint) string {
	return fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), orderID)
}
