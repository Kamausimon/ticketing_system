package models

import (
	"time"

	"gorm.io/gorm"
)

// SettlementStatus defines the status of a settlement batch
type SettlementStatus string

const (
	SettlementPending        SettlementStatus = "pending"          // Created but not processed
	SettlementAwaitingEvent  SettlementStatus = "awaiting_event"   // Waiting for event completion
	SettlementHoldingPeriod  SettlementStatus = "holding_period"   // In post-event holding period
	SettlementReadyToProcess SettlementStatus = "ready_to_process" // Ready for payout
	SettlementProcessing     SettlementStatus = "processing"       // Being processed by payment system
	SettlementCompleted      SettlementStatus = "completed"        // Successfully sent to organizers
	SettlementFailed         SettlementStatus = "failed"           // Failed to process
	SettlementCancelled      SettlementStatus = "cancelled"        // Manually cancelled
	SettlementPartial        SettlementStatus = "partial"          // Some payments succeeded, some failed
	SettlementDisputed       SettlementStatus = "disputed"         // Under dispute investigation
	SettlementWithheld       SettlementStatus = "withheld"         // Withheld due to issues
)

// SettlementFrequency defines how often settlements occur
type SettlementFrequency string

const (
	SettlementPostEvent SettlementFrequency = "post_event" // After event completion + holding period
	SettlementWeekly    SettlementFrequency = "weekly"     // Weekly (only for completed events)
	SettlementBiWeekly  SettlementFrequency = "bi_weekly"  // Every 2 weeks
	SettlementMonthly   SettlementFrequency = "monthly"    // Monthly (recommended)
	SettlementQuarterly SettlementFrequency = "quarterly"  // Every 3 months
	SettlementManual    SettlementFrequency = "manual"     // Manual approval required
)

// SettlementTrigger defines what triggers a settlement
type SettlementTrigger string

const (
	TriggerEventCompletion SettlementTrigger = "event_completion" // Event finished successfully
	TriggerScheduledDate   SettlementTrigger = "scheduled_date"   // Regular schedule (weekly/monthly)
	TriggerManualRequest   SettlementTrigger = "manual_request"   // Organizer/admin request
	TriggerDisputeResolved SettlementTrigger = "dispute_resolved" // After dispute resolution
)

// SettlementRecord represents a batch payment to organizers
// This is like a "payroll run" - paying multiple organizers at once
// CRITICAL: Settlements only occur AFTER events complete + holding period
type SettlementRecord struct {
	gorm.Model

	// Settlement identification
	SettlementBatchID string              `gorm:"unique;not null;index"` // Unique batch identifier
	Description       string              `gorm:"not null"`              // Human readable description
	Status            SettlementStatus    `gorm:"not null;index"`        // Current status
	Frequency         SettlementFrequency `gorm:"not null"`              // Settlement frequency
	Trigger           SettlementTrigger   `gorm:"not null"`              // What triggered this settlement

	// Event completion requirements
	EventID                 *uint      `gorm:"index"` // Specific event (for post-event settlements)
	Event                   *Event     `gorm:"foreignKey:EventID"`
	EventCompletedAt        *time.Time // When the event actually finished
	EventCompletionVerified bool       `gorm:"default:false"` // Admin verified event completed successfully

	// Risk management & holding period
	HoldingPeriodDays      int        `gorm:"not null;default:7"` // Days to hold funds post-event
	HoldingPeriodStartDate *time.Time // When holding period started
	HoldingPeriodEndDate   *time.Time // When holding period ends
	EarliestPayoutDate     *time.Time // Earliest date settlement can be processed

	// Dispute & risk tracking
	HasActiveDisputes bool    `gorm:"default:false"` // Are there active disputes?
	DisputeCount      int     `gorm:"default:0"`     // Number of disputes
	ChargebackCount   int     `gorm:"default:0"`     // Number of chargebacks
	RefundAmount      Money   `gorm:"default:0"`     // Total refunds issued
	WithholdingReason *string // Why settlement is withheld

	// Time period this settlement covers
	PeriodStartDate time.Time `gorm:"not null;index"` // Start of settlement period
	PeriodEndDate   time.Time `gorm:"not null;index"` // End of settlement period

	// Settlement totals (for quick reference)
	TotalOrganizers     int    `gorm:"not null;default:0"`     // Number of organizers paid
	TotalAmount         Money  `gorm:"not null;default:0"`     // Total amount being settled
	TotalPaymentRecords int    `gorm:"not null;default:0"`     // Number of payment records included
	Currency            string `gorm:"not null;default:'KSH'"` // Settlement currency

	// Processing information
	InitiatedBy     *uint      `gorm:"index"` // User who initiated settlement
	InitiatedByUser *User      `gorm:"foreignKey:InitiatedBy"`
	ApprovedBy      *uint      `gorm:"index"` // User who approved settlement (for manual approvals)
	ApprovedByUser  *User      `gorm:"foreignKey:ApprovedBy"`
	ApprovedAt      *time.Time // When settlement was approved
	ProcessedAt     *time.Time // When settlement was processed
	CompletedAt     *time.Time // When all payments completed
	FailedAt        *time.Time // When settlement failed

	// External system references
	ExternalBatchID  *string         `gorm:"index"` // External payment system batch ID
	PaymentGatewayID *uint           `gorm:"index"` // Gateway used for settlement
	PaymentGateway   *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`

	// Settlement items (individual organizer payments)
	SettlementItems []SettlementItem `gorm:"foreignKey:SettlementRecordID"`

	// Metadata
	Notes             *string // Additional notes
	InternalReference *string // Internal tracking reference
	RiskScore         *int    // Risk assessment score (0-100)

	// Soft delete support
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// SettlementItem represents a payment to a single organizer within a settlement batch
// Each item is tied to specific events and includes post-event verification
type SettlementItem struct {
	gorm.Model

	// Parent settlement
	SettlementRecordID uint             `gorm:"not null;index"`
	SettlementRecord   SettlementRecord `gorm:"foreignKey:SettlementRecordID"`

	// Organizer being paid
	OrganizerID uint      `gorm:"not null;index"`
	Organizer   Organizer `gorm:"foreignKey:OrganizerID"`

	// Event tracking (critical for post-event settlements)
	EventID         uint        `gorm:"not null;index"` // Event this settlement item relates to
	Event           Event       `gorm:"foreignKey:EventID"`
	EventStatus     EventStatus // Event status at settlement time
	EventEndDate    time.Time   `gorm:"not null"` // When event ended
	EventVerifiedAt *time.Time  // When event completion was verified

	// Risk assessment for this specific item
	HasDisputes        bool    `gorm:"default:false"` // Disputes for this organizer's event
	RefundAmountIssued Money   `gorm:"default:0"`     // Refunds issued for this event
	ChargebackAmount   Money   `gorm:"default:0"`     // Chargebacks for this event
	RiskHoldApplied    bool    `gorm:"default:false"` // Additional risk hold applied
	RiskHoldReason     *string // Reason for risk hold

	// Payment amount calculation
	GrossAmount       Money  `gorm:"not null"`  // Total earnings before deductions
	PlatformFeeAmount Money  `gorm:"default:0"` // Platform fees deducted
	RefundDeduction   Money  `gorm:"default:0"` // Deducted for refunds/chargebacks
	AdjustmentAmount  Money  `gorm:"default:0"` // Manual adjustments (+/-)
	NetAmount         Money  `gorm:"not null"`  // Final amount being paid
	Currency          string `gorm:"not null;default:'KSH'"`

	// Item status (can be different from parent settlement)
	Status SettlementStatus `gorm:"not null;index"`

	// External references
	ExternalTransactionID *string `gorm:"index"` // Gateway transaction ID
	ExternalReference     *string // Additional reference

	// Organizer bank details (snapshot at time of settlement)
	BankAccountNumber string  `gorm:"not null"`
	BankName          string  `gorm:"not null"`
	BankCode          *string // Bank/routing code
	AccountHolderName string  `gorm:"not null"`

	// Processing timestamps
	ProcessedAt   *time.Time // When this item was processed
	CompletedAt   *time.Time // When payment completed
	FailedAt      *time.Time // When payment failed
	FailureReason *string    // Why payment failed

	// Link to related payment records
	PaymentRecords []PaymentRecord `gorm:"many2many:settlement_payment_records"` // PaymentRecords included in this settlement

	// Metadata
	Description string  `gorm:"not null"` // Description of this settlement item
	Notes       *string // Additional notes
}
