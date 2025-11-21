package refunds

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

type RefundHandler struct {
	DB *gorm.DB
	// Payment handler for processing actual refunds through gateway
	IntasendSecretKey     string
	IntasendWebhookSecret string
	IntasendTestMode      bool
}

func NewRefundHandler(db *gorm.DB, intasendSecret, webhookSecret string, testMode bool) *RefundHandler {
	return &RefundHandler{
		DB:                    db,
		IntasendSecretKey:     intasendSecret,
		IntasendWebhookSecret: webhookSecret,
		IntasendTestMode:      testMode,
	}
}

// Request types
type RefundRequestRequest struct {
	OrderID      uint                  `json:"order_id"`
	RefundType   models.RefundType     `json:"refund_type"` // "full", "partial", "ticket"
	RefundReason models.RefundReason   `json:"reason"`
	Description  string                `json:"description"`
	LineItems    []RefundLineItemInput `json:"line_items,omitempty"` // For partial refunds
}

type RefundLineItemInput struct {
	OrderItemID uint   `json:"order_item_id"`
	TicketID    *uint  `json:"ticket_id,omitempty"` // For ticket-level refunds
	Quantity    int    `json:"quantity"`
	Amount      int64  `json:"amount"` // Amount in cents
	Reason      string `json:"reason,omitempty"`
}

type RefundApprovalRequest struct {
	Approved        bool    `json:"approved"`
	InternalNotes   *string `json:"internal_notes,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"` // Required if approved=false
}

type RefundProcessRequest struct {
	PaymentGatewayID uint `json:"payment_gateway_id"`
}

// Response types
type RefundResponse struct {
	Success      bool   `json:"success"`
	RefundID     uint   `json:"refund_id,omitempty"`
	RefundNumber string `json:"refund_number,omitempty"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	Message      string `json:"message"`
}

type RefundStatusResponse struct {
	RefundNumber     string                 `json:"refund_number"`
	Status           string                 `json:"status"`
	RefundType       string                 `json:"refund_type"`
	RefundReason     string                 `json:"refund_reason"`
	OriginalAmount   int64                  `json:"original_amount"`
	RefundAmount     int64                  `json:"refund_amount"`
	Currency         string                 `json:"currency"`
	RequestedAt      string                 `json:"requested_at"`
	ApprovedAt       *string                `json:"approved_at,omitempty"`
	ProcessedAt      *string                `json:"processed_at,omitempty"`
	CompletedAt      *string                `json:"completed_at,omitempty"`
	ExternalRefundID *string                `json:"external_refund_id,omitempty"`
	LineItems        []RefundLineItemDetail `json:"line_items,omitempty"`
}

type RefundLineItemDetail struct {
	OrderItemID uint   `json:"order_item_id"`
	TicketID    *uint  `json:"ticket_id,omitempty"`
	Quantity    int    `json:"quantity"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type RefundListResponse struct {
	Refunds []RefundSummary `json:"refunds"`
	Total   int             `json:"total"`
}

type RefundSummary struct {
	ID           uint   `json:"id"`
	RefundNumber string `json:"refund_number"`
	OrderID      uint   `json:"order_id"`
	Status       string `json:"status"`
	RefundType   string `json:"refund_type"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	RequestedAt  string `json:"requested_at"`
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
