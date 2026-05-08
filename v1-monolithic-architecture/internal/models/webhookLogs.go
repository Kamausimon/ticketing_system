package models

import (
	"time"

	"gorm.io/gorm"
)

// WebhookProvider defines the source of the webhook
type WebhookProvider string

const (
	WebhookIntasend WebhookProvider = "intasend"
	WebhookStripe   WebhookProvider = "stripe"
	WebhookMpesa    WebhookProvider = "mpesa"
	WebhookPaypal   WebhookProvider = "paypal"
	WebhookPesapal  WebhookProvider = "pesapal"
	WebhookFlutter  WebhookProvider = "flutterwave"
	WebhookOther    WebhookProvider = "other"
)

// WebhookStatus defines the processing status
type WebhookStatus string

const (
	WebhookReceived  WebhookStatus = "received"  // Webhook received
	WebhookProcessed WebhookStatus = "processed" // Successfully processed
	WebhookFailed    WebhookStatus = "failed"    // Processing failed
	WebhookRetrying  WebhookStatus = "retrying"  // Retry in progress
	WebhookIgnored   WebhookStatus = "ignored"   // Intentionally ignored
	WebhookDuplicate WebhookStatus = "duplicate" // Duplicate event
)

// WebhookLog stores all webhook events from payment gateways
// Critical for debugging payment issues and ensuring idempotency
type WebhookLog struct {
	gorm.Model

	// Webhook identification
	Provider  WebhookProvider `gorm:"not null;index"`
	EventID   string          `gorm:"not null;index"` // External event ID (for deduplication)
	EventType string          `gorm:"not null;index"` // e.g., "payment_intent.succeeded"
	Status    WebhookStatus   `gorm:"not null;index;default:'received'"`

	// Raw webhook data
	Payload       string `gorm:"type:text;not null"` // Full JSON payload
	Headers       string `gorm:"type:text"`          // HTTP headers (for debugging)
	RequestMethod string `gorm:"default:'POST'"`     // Usually POST
	RequestPath   string // Webhook endpoint path

	// Processing information
	ProcessedAt    *time.Time // When webhook was processed
	ProcessingTime *int       // Processing time in milliseconds
	RetryCount     int        `gorm:"default:0"` // Number of retry attempts
	LastRetryAt    *time.Time

	// Success/failure tracking
	Success      bool    `gorm:"index;default:false"`
	ErrorMessage *string `gorm:"type:text"` // Error if processing failed
	StackTrace   *string `gorm:"type:text"` // Stack trace for debugging

	// Related entities (if identified from payload)
	OrderID              *uint `gorm:"index"` // Related order
	PaymentTransactionID *uint `gorm:"index"` // Related transaction
	PaymentRecordID      *uint `gorm:"index"` // Related payment record
	AccountID            *uint `gorm:"index"` // Related account
	OrganizerID          *uint `gorm:"index"` // Related organizer

	// External references
	ExternalTransactionID *string `gorm:"index"` // Gateway transaction ID
	ExternalReference     *string `gorm:"index"` // Additional reference

	// Security
	SignatureValid  bool    `gorm:"default:false"` // Signature verification result
	SignatureHeader *string // Signature from header
	IPAddress       string  `gorm:"index"` // Source IP (for security)
	UserAgent       *string

	// Idempotency
	IdempotencyKey *string `gorm:"index"` // For preventing duplicate processing
	IsDuplicate    bool    `gorm:"default:false;index"`

	// Metadata
	Environment string  `gorm:"default:'production'"` // production/sandbox
	APIVersion  *string // Gateway API version
	Notes       *string `gorm:"type:text"` // Additional notes

	// Response sent back to webhook
	ResponseStatus int     `gorm:"default:200"` // HTTP status sent back
	ResponseBody   *string // Response body sent

	// Soft delete support (for audit trail)
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// IsRetryable checks if webhook can be retried
func (wh *WebhookLog) IsRetryable() bool {
	return wh.Status == WebhookFailed && wh.RetryCount < 5
}

// ShouldRetry checks if retry is needed based on time elapsed
func (wh *WebhookLog) ShouldRetry() bool {
	if !wh.IsRetryable() {
		return false
	}

	// Exponential backoff: 1min, 5min, 15min, 1hr, 4hr
	backoffMinutes := []int{1, 5, 15, 60, 240}
	if wh.RetryCount >= len(backoffMinutes) {
		return false
	}

	if wh.LastRetryAt == nil {
		wh.LastRetryAt = &wh.CreatedAt
	}

	elapsed := time.Since(*wh.LastRetryAt)
	backoff := time.Duration(backoffMinutes[wh.RetryCount]) * time.Minute
	return elapsed >= backoff
}
