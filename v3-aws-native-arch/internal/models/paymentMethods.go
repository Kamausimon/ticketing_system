package models

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethodType defines the type of payment method
type PaymentMethodType string

const (
	PaymentMethodCard   PaymentMethodType = "card"   // Credit/debit card
	PaymentMethodMpesa  PaymentMethodType = "mpesa"  // M-Pesa mobile money
	PaymentMethodBank   PaymentMethodType = "bank"   // Direct bank transfer
	PaymentMethodWallet PaymentMethodType = "wallet" // Digital wallet
	PaymentMethodCash   PaymentMethodType = "cash"   // Cash payment
	PaymentMethodOther  PaymentMethodType = "other"  // Other methods
)

// PaymentMethodStatus defines the status of a payment method
type PaymentMethodStatus string

const (
	PaymentMethodActive   PaymentMethodStatus = "active"   // Active and ready to use
	PaymentMethodExpired  PaymentMethodStatus = "expired"  // Card expired
	PaymentMethodInvalid  PaymentMethodStatus = "invalid"  // Validation failed
	PaymentMethodDisabled PaymentMethodStatus = "disabled" // User disabled
	PaymentMethodDeleted  PaymentMethodStatus = "deleted"  // Soft deleted
)

// CardBrand defines the card brand
type CardBrand string

const (
	CardVisa       CardBrand = "visa"
	CardMastercard CardBrand = "mastercard"
	CardAmex       CardBrand = "amex"
	CardDiscover   CardBrand = "discover"
	CardDinersClub CardBrand = "diners_club"
	CardJCB        CardBrand = "jcb"
	CardUnknown    CardBrand = "unknown"
)

// PaymentMethod stores customer payment methods for future use
// Critical: Never store full card numbers - only last 4 digits
type PaymentMethod struct {
	gorm.Model

	// Owner
	AccountID uint    `gorm:"not null;index"`
	Account   Account `gorm:"foreignKey:AccountID"`

	// Payment method type
	Type   PaymentMethodType   `gorm:"not null;index"`
	Status PaymentMethodStatus `gorm:"not null;index;default:'active'"`

	// Display information
	DisplayName string  `gorm:"not null"` // e.g., "Visa ending in 4242"
	Nickname    *string // User-defined nickname

	// Default payment method
	IsDefault bool `gorm:"default:false;index"`

	// Card-specific fields (only for card type)
	CardBrand       *CardBrand // Card brand (Visa, Mastercard, etc.)
	CardLast4       *string    `gorm:"size:4"` // Last 4 digits
	CardExpiryMonth *int       // Expiry month (1-12)
	CardExpiryYear  *int       // Expiry year (e.g., 2025)
	CardCountry     *string    `gorm:"size:2"` // ISO country code
	CardFingerprint *string    `gorm:"index"`  // Unique card identifier

	// M-Pesa specific fields
	MpesaPhoneNumber *string `gorm:"index"` // Phone number for M-Pesa
	MpesaAccountName *string // Account holder name

	// Bank specific fields
	BankAccountLast4  *string // Last 4 digits of account
	BankName          *string
	BankCode          *string // Bank/routing code
	BankAccountHolder *string

	// External gateway references
	StripePaymentMethodID   *string `gorm:"index"` // Stripe payment method ID
	StripeCustomerID        *string `gorm:"index"` // Stripe customer ID
	ExternalPaymentMethodID *string `gorm:"index"` // Other gateway IDs

	// Verification status
	IsVerified    bool       `gorm:"default:false"` // Verified by gateway
	VerifiedAt    *time.Time // When verified
	LastUsedAt    *time.Time // Last time used for payment
	FailureCount  int        `gorm:"default:0"` // Number of failed attempts
	LastFailureAt *time.Time // Last failure timestamp

	// Metadata
	BillingAddress *string // Billing address JSON or string
	Metadata       *string // Additional metadata (JSON)

	// Soft delete support
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// IsExpired checks if a card payment method has expired
func (pm *PaymentMethod) IsExpired() bool {
	if pm.Type != PaymentMethodCard || pm.CardExpiryMonth == nil || pm.CardExpiryYear == nil {
		return false
	}

	now := time.Now()
	// Card expires at the end of the expiry month
	expiryDate := time.Date(*pm.CardExpiryYear, time.Month(*pm.CardExpiryMonth+1), 0, 23, 59, 59, 0, time.UTC)
	return now.After(expiryDate)
}

// ShouldUpdateExpiry checks if card expiry needs updating
func (pm *PaymentMethod) ShouldUpdateExpiry() bool {
	if pm.Type != PaymentMethodCard {
		return false
	}
	return pm.IsExpired() && pm.Status == PaymentMethodActive
}
