package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentRecordType string

const (
	RecordCustomerPayment PaymentRecordType = "customer_payment" // Customer pays for tickets
	RecordPlatformFee     PaymentRecordType = "platform_fee"     // Platform commission
	RecordGatewayFee      PaymentRecordType = "gateway_fee"      // Payment processor fee
	RecordOrganizerPayout PaymentRecordType = "organizer_payout" // Payment to organizer
	RecordRefund          PaymentRecordType = "refund"           // Refund to customer
	RecordChargeback      PaymentRecordType = "chargeback"       // Dispute/chargeback
	RecordAdjustment      PaymentRecordType = "adjustment"       // Manual adjustment
)

type PaymentRecordStatus string

const (
	RecordPending   PaymentRecordStatus = "pending"
	RecordCompleted PaymentRecordStatus = "completed"
	RecordFailed    PaymentRecordStatus = "failed"
	RecordCancelled PaymentRecordStatus = "cancelled"
	RecordDisputed  PaymentRecordStatus = "disputed"
)

type PaymentRecord struct {
	gorm.Model

	Amount   Money               `gorm:"not null"`               // Amount in cents
	Currency string              `gorm:"not null;default:'KSH'"` // Currency code
	Type     PaymentRecordType   `gorm:"not null;index"`         // Type of payment
	Status   PaymentRecordStatus `gorm:"not null;index"`         // Current status

	// Relationships - what this payment is for
	OrderID          *uint           `gorm:"index"` // Related order (if applicable)
	Order            *Order          `gorm:"foreignKey:OrderID"`
	EventID          *uint           `gorm:"index"` // Related event
	Event            *Event          `gorm:"foreignKey:EventID"`
	AccountID        *uint           `gorm:"index"` // Customer account
	Account          *Account        `gorm:"foreignKey:AccountID"`
	OrganizerID      *uint           `gorm:"index"` // Organizer (for payouts)
	Organizer        *Organizer      `gorm:"foreignKey:OrganizerID"`
	PaymentGatewayID *uint           `gorm:"index"` // Payment gateway used
	PaymentGateway   *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`

	// External system references
	ExternalTransactionID *string `gorm:"index"` // Gateway transaction ID
	ExternalReference     *string // Additional reference
	GatewayResponseCode   *string // Gateway response code

	// Timestamps for payment lifecycle
	InitiatedAt time.Time  `gorm:"not null"` // When payment was initiated
	ProcessedAt *time.Time // When payment was processed
	CompletedAt *time.Time // When payment was completed
	FailedAt    *time.Time // When payment failed

	// Metadata and tracking
	Description string  `gorm:"not null"` // Human readable description
	Notes       *string // Additional notes
	IPAddress   *string // Customer IP (for fraud detection)
	UserAgent   *string // Customer browser

	// Fee tracking
	PlatformFeeAmount Money `gorm:"default:0"` // Platform commission
	GatewayFeeAmount  Money `gorm:"default:0"` // Payment processor fee
	NetAmount         Money `gorm:"not null"`  // Amount after fees

	// Related records for audit trail
	ParentRecordID *uint           // Parent record (for refunds)
	ParentRecord   *PaymentRecord  `gorm:"foreignKey:ParentRecordID"`
	ChildRecords   []PaymentRecord `gorm:"foreignKey:ParentRecordID"` // Child records (refunds, adjustments)

	// Reconciliation fields
	ReconciledAt      *time.Time // When reconciled with bank/gateway
	ReconciliationRef *string    // Reconciliation reference

	// Soft delete support
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
