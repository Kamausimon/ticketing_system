package payments

import (
	"encoding/json"
	"net/http"
	"os"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

type PaymentHandler struct {
	DB                     *gorm.DB
	IntasendPublishableKey string
	IntasendSecretKey      string
	IntasendWebhookSecret  string
	IntasendTestMode       bool
	// StripeSecretKey      string // Commented - for future international expansion
	// StripePublishableKey string
	// StripeWebhookSecret  string
}

func NewPaymentHandler(db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{
		DB:                     db,
		IntasendPublishableKey: os.Getenv("INTASEND_PUBLISHABLE_KEY"),
		IntasendSecretKey:      os.Getenv("INTASEND_SECRET_KEY"),
		IntasendWebhookSecret:  os.Getenv("INTASEND_WEBHOOK_SECRET"),
		IntasendTestMode:       os.Getenv("INTASEND_TEST_MODE") == "true",
		// StripeSecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		// StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		// StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}

// Request types
type InitiatePaymentRequest struct {
	OrderID       uint    `json:"order_id"`
	Amount        int64   `json:"amount"` // Amount in cents
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`         // "mpesa", "card", "bank"
	PhoneNumber   *string `json:"phone_number,omitempty"` // For M-Pesa
	Email         string  `json:"email"`
	CallbackURL   *string `json:"callback_url,omitempty"`
}

type RefundPaymentRequest struct {
	PaymentRecordID uint   `json:"payment_record_id"`
	Amount          int64  `json:"amount"` // Amount to refund in cents
	Reason          string `json:"reason"`
}

type VerifyPaymentRequest struct {
	TransactionID string `json:"transaction_id"`
}

// Response types
type PaymentResponse struct {
	Success           bool    `json:"success"`
	TransactionID     string  `json:"transaction_id,omitempty"`
	Status            string  `json:"status"`
	PaymentURL        *string `json:"payment_url,omitempty"`         // For card payments
	CheckoutRequestID *string `json:"checkout_request_id,omitempty"` // For M-Pesa STK Push
	Message           string  `json:"message"`
}

type RefundResponse struct {
	Success  bool   `json:"success"`
	RefundID string `json:"refund_id,omitempty"`
	Amount   int64  `json:"amount"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

type PaymentStatusResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        int64   `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	CompletedAt   *string `json:"completed_at,omitempty"`
}

type PaymentMethodResponse struct {
	ID          uint    `json:"id"`
	Type        string  `json:"type"`
	DisplayName string  `json:"display_name"`
	IsDefault   bool    `json:"is_default"`
	Last4       *string `json:"last4,omitempty"`
	ExpiryMonth *int    `json:"expiry_month,omitempty"`
	ExpiryYear  *int    `json:"expiry_year,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	IsVerified  bool    `json:"is_verified"`
}

type WebhookEventResponse struct {
	Received  bool   `json:"received"`
	Processed bool   `json:"processed"`
	Message   string `json:"message"`
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

func convertToPaymentMethodResponse(pm *models.PaymentMethod) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:          pm.ID,
		Type:        string(pm.Type),
		DisplayName: pm.DisplayName,
		IsDefault:   pm.IsDefault,
		Last4:       pm.CardLast4,
		ExpiryMonth: pm.CardExpiryMonth,
		ExpiryYear:  pm.CardExpiryYear,
		PhoneNumber: pm.MpesaPhoneNumber,
		IsVerified:  pm.IsVerified,
	}
}
