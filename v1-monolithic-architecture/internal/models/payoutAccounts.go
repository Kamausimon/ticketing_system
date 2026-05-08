package models

import (
	"time"

	"gorm.io/gorm"
)

// PayoutAccountType defines the type of payout account
type PayoutAccountType string

const (
	PayoutBank        PayoutAccountType = "bank"         // Bank account
	PayoutMobileMoney PayoutAccountType = "mobile_money" // M-Pesa, etc.
	PayoutPaypal      PayoutAccountType = "paypal"       // PayPal
	PayoutStripe      PayoutAccountType = "stripe"       // Stripe Connect
	PayoutOther       PayoutAccountType = "other"        // Other methods
)

// PayoutAccountStatus defines the verification status
type PayoutAccountStatus string

const (
	PayoutPending     PayoutAccountStatus = "pending"     // Awaiting verification
	PayoutVerifying   PayoutAccountStatus = "verifying"   // Verification in progress
	PayoutVerified    PayoutAccountStatus = "verified"    // Verified and active
	PayoutFailed      PayoutAccountStatus = "failed"      // Verification failed
	PayoutSuspended   PayoutAccountStatus = "suspended"   // Temporarily suspended
	PayoutDeactivated PayoutAccountStatus = "deactivated" // Deactivated by user
)

// PayoutAccount stores organizer bank account/payout details
// Separating from Organizer model allows multiple accounts and better security
type PayoutAccount struct {
	gorm.Model

	// Owner
	OrganizerID uint      `gorm:"not null;index"`
	Organizer   Organizer `gorm:"foreignKey:OrganizerID"`

	// Account identification
	AccountType PayoutAccountType   `gorm:"not null;index"`
	Status      PayoutAccountStatus `gorm:"not null;index;default:'pending'"`
	DisplayName string              `gorm:"not null"` // e.g., "KCB Bank - Main Account"
	IsDefault   bool                `gorm:"default:false;index"`

	// Bank account details (for bank transfers)
	BankName          *string
	BankCode          *string // SWIFT/IFSC/Sort code
	BankBranch        *string
	BankCountry       *string `gorm:"size:2"` // ISO country code
	AccountNumber     *string `gorm:"index"`  // Encrypted in production
	AccountHolderName *string

	// Mobile money details (M-Pesa, etc.)
	MobileProvider    *string // e.g., "Safaricom", "Airtel"
	MobilePhoneNumber *string `gorm:"index"`
	MobileAccountName *string

	// PayPal details
	PaypalEmail *string `gorm:"index"`

	// Stripe Connect details
	StripeAccountID *string `gorm:"index"` // Stripe Connected Account ID
	StripeCountry   *string // Country for Stripe account

	// Currency
	Currency string `gorm:"not null;default:'KSH'"` // Default payout currency

	// Verification
	IsVerified        bool `gorm:"default:false;index"`
	VerifiedAt        *time.Time
	VerifiedBy        *uint   // Admin who verified
	VerificationNotes *string `gorm:"type:text"`

	// Verification documents (references to uploaded files)
	DocumentPaths     *string `gorm:"type:text"` // JSON array of document paths
	VerificationToken *string `gorm:"index"`     // Token for micro-deposit verification

	// Usage tracking
	TotalPayoutsCount  int   `gorm:"default:0"` // Number of payouts sent
	TotalPayoutsAmount Money `gorm:"default:0"` // Total amount paid out
	LastPayoutAt       *time.Time
	LastPayoutAmount   *Money

	// Failure tracking
	FailedPayoutsCount int `gorm:"default:0"` // Number of failed payouts
	LastFailureAt      *time.Time
	LastFailureReason  *string `gorm:"type:text"`

	// Security and compliance
	RequiresKYC          bool    `gorm:"default:false"` // KYC verification required
	KYCStatus            *string // KYC verification status
	KYCCompletedAt       *time.Time
	IsSuspiciousActivity bool    `gorm:"default:false;index"` // Flagged for review
	SuspicionReason      *string `gorm:"type:text"`
	ReviewedBy           *uint   // Admin who reviewed
	ReviewedAt           *time.Time

	// Risk management
	DailyPayoutLimit   *Money // Maximum daily payout
	MonthlyPayoutLimit *Money // Maximum monthly payout
	RequiresApproval   bool   `gorm:"default:false"` // Manual approval required

	// External gateway references
	ExternalAccountID *string `gorm:"index"`     // External payment system ID
	ExternalMetadata  *string `gorm:"type:text"` // Additional metadata (JSON)

	// Address information (for compliance)
	AddressLine1 *string
	AddressLine2 *string
	City         *string
	State        *string
	PostalCode   *string
	Country      *string `gorm:"size:2"` // ISO country code

	// Metadata
	Notes *string `gorm:"type:text"` // Internal notes

	// Soft delete support
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// CanReceivePayout checks if account can receive payouts
func (pa *PayoutAccount) CanReceivePayout() bool {
	return pa.Status == PayoutVerified &&
		!pa.IsSuspiciousActivity &&
		pa.IsVerified
}

// RequiresReview checks if account needs manual review
func (pa *PayoutAccount) RequiresReview() bool {
	return pa.IsSuspiciousActivity ||
		pa.RequiresApproval ||
		pa.FailedPayoutsCount > 3
}

// IsWithinLimits checks if a payout amount is within limits
func (pa *PayoutAccount) IsWithinLimits(amount Money) bool {
	if pa.DailyPayoutLimit != nil && amount > *pa.DailyPayoutLimit {
		return false
	}
	if pa.MonthlyPayoutLimit != nil && amount > *pa.MonthlyPayoutLimit {
		return false
	}
	return true
}
