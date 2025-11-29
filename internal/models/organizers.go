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
	// Payment and bank details
	PaymentGatewayID    *uint           `gorm:"index"`
	PaymentGateway      *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	BankAccountName     string
	BankAccountNumber   string
	BankCode            string
	BankCountry         string
	IsPaymentConfigured bool `gorm:"default:false"`
	// Verification and approval
	IsVerified         bool   `gorm:"default:false"`
	VerificationStatus string // "pending", "approved", "rejected"
	RejectionReason    string
}
