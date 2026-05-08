package models

import "gorm.io/gorm"

type Organizer struct {
	gorm.Model
	AccountID           uint    `gorm:"not null;index"`
	Account             Account `gorm:"foreignKey:AccountID"`
	Name                string
	About               string
	Email               string
	Phone               string
	ConfirmationKey     string
	Facebook            string
	Twitter             string
	LogoPath            *string
	IsEmailConfirmed    bool `gorm:"default:false"`
	ShowTwitterWidget   bool
	ShowFacebookWidget  bool
	TaxName             string
	TaxValue            float32
	TaxPin              string
	ChargeTax           int
	PageHeaderBgColor   string
	PageBgColor         string
	PageTextColor       string
	EnableOrganizerPage bool
	// Bank details for PAYOUTS only (NOT for collecting payments from customers)
	// Platform collects all customer payments, then uses these details to pay organizers
	PaymentGatewayID    *uint           `gorm:"index"` // DEPRECATED: Not used in centralized payment model
	PaymentGateway      *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	BankAccountName     string          // Account holder name for payouts
	BankAccountNumber   string          // Encrypted account number for payouts
	BankCode            string          // Encrypted bank/SWIFT code for payouts
	BankCountry         string          // Country code for payout destination
	IsPaymentConfigured bool            `gorm:"default:false"` // DEPRECATED: Use bank details presence instead
	// Verification and approval
	IsVerified         bool   `gorm:"default:false"`
	VerificationStatus string `gorm:"type:varchar(50);default:'pending'"` // "pending", "kyc_scheduled", "kyc_completed", "approved", "rejected"
	RejectionReason    string `gorm:"type:text"`
	// KYC tracking
	KYCStatus      string  `gorm:"type:varchar(50);default:'pending'"` // "pending", "scheduled", "completed", "failed"
	KYCNotes       string  `gorm:"type:text"`                          // Admin notes from KYC process
	KYCCompletedAt *string // When KYC was completed
}
