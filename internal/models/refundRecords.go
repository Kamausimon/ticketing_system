package models

import (
	"time"

	"gorm.io/gorm"
)

// RefundReason - Why was the refund issued?
type RefundReason string

const (
	RefundEventCancelled  RefundReason = "event_cancelled"  // Event was cancelled
	RefundCustomerRequest RefundReason = "customer_request" // Customer requested refund
	RefundChargeback      RefundReason = "chargeback"       // Credit card chargeback
	RefundFraud           RefundReason = "fraud_prevention" // Fraud detection
	RefundSystemError     RefundReason = "system_error"     // Technical error
	RefundGoodwill        RefundReason = "goodwill"         // Goodwill gesture
)

// RefundStatus - Current state of the refund
type RefundStatus string

const (
	RefundRequested  RefundStatus = "requested"  // Refund requested
	RefundApproved   RefundStatus = "approved"   // Approved for processing
	RefundProcessing RefundStatus = "processing" // Being processed
	RefundCompleted  RefundStatus = "completed"  // Successfully refunded
	RefundFailed     RefundStatus = "failed"     // Failed to process
	RefundRejected   RefundStatus = "rejected"   // Request denied
)

// RefundType - What kind of refund is this?
type RefundType string

const (
	RefundFull    RefundType = "full"    // Full order refund
	RefundPartial RefundType = "partial" // Partial refund
	RefundTicket  RefundType = "ticket"  // Individual ticket refund
)

// RefundRecord represents a refund transaction
// This connects to PaymentRecords and impacts SettlementRecords
type RefundRecord struct {
	gorm.Model

	// Basic refund info
	RefundNumber string       `gorm:"unique;not null;index"` // Unique refund ID (REF-2024-001)
	RefundType   RefundType   `gorm:"not null;index"`        // Type of refund
	RefundReason RefundReason `gorm:"not null;index"`        // Why refund was issued
	Status       RefundStatus `gorm:"not null;index"`        // Current status

	// What's being refunded
	OrderID uint  `gorm:"not null;index"` // Original order
	Order   Order `gorm:"foreignKey:OrderID"`
	EventID uint  `gorm:"not null;index"` // Event this affects
	Event   Event `gorm:"foreignKey:EventID"`

	// Who's involved
	AccountID   uint      `gorm:"not null;index"` // Customer account
	Account     Account   `gorm:"foreignKey:AccountID"`
	OrganizerID uint      `gorm:"not null;index"` // Organizer impacted
	Organizer   Organizer `gorm:"foreignKey:OrganizerID"`

	// Money amounts (in cents)
	OriginalAmount  Money  `gorm:"not null"`               // Original payment amount
	RefundAmount    Money  `gorm:"not null"`               // Amount being refunded
	OrganizerImpact Money  `gorm:"not null"`               // Amount deducted from organizer
	Currency        string `gorm:"not null;default:'KSH'"` // Currency

	// Processing info
	PaymentGatewayID *uint           `gorm:"index"` // Gateway used for refund
	PaymentGateway   *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	ExternalRefundID *string         `gorm:"index"` // Gateway refund transaction ID

	// Approval workflow
	RequestedBy     *uint `gorm:"index"` // Who requested refund
	RequestedByUser *User `gorm:"foreignKey:RequestedBy"`
	ApprovedBy      *uint `gorm:"index"` // Who approved refund
	ApprovedByUser  *User `gorm:"foreignKey:ApprovedBy"`

	// Timestamps
	RequestedAt time.Time  `gorm:"not null"` // When requested
	ApprovedAt  *time.Time // When approved
	ProcessedAt *time.Time // When processed
	CompletedAt *time.Time // When completed
	FailedAt    *time.Time // When failed

	// Settlement impact
	AffectsSettlement  bool `gorm:"default:true"`  // Impacts organizer settlement?
	SettlementAdjusted bool `gorm:"default:false"` // Settlement adjusted?

	// Descriptions and notes
	Description     string  `gorm:"not null"` // Human readable description
	InternalNotes   *string // Admin notes
	RejectionReason *string // Why rejected

	// Line items for partial refunds
	RefundLineItems []RefundLineItem `gorm:"foreignKey:RefundRecordID"`

	// Soft delete
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
