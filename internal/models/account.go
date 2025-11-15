package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	FirstName            string
	LastName             string
	Email                string
	TimezoneID           *int
	DateFormatID         *int
	DateTimeFormatID     *int
	CurrencyID           *int
	LastIP               *string
	LastLoginDate        *time.Time
	Address1             *string
	Address2             *string
	City                 *string
	County               *string
	PostalCode           *string
	IsActive             bool `gorm:"default:true"`
	IsBanned             bool `gorm:"default:false"`
	StripeAccessToken    *string
	StripeRefreshToken   *string
	StripeSecretKey      *string
	StripePublishableKey *string
	StripeDataRaw        *string
	PaymentGatewayID     uint
	PaymentGateway       PaymentGateway
}
