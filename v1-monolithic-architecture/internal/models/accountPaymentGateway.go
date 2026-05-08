package models

import "gorm.io/gorm"

type AccountPaymentGateway struct {
	gorm.Model
	AccountID        uint           `gorm:"not null;index"`
	Account          Account        `gorm:"foreignKey:AccountID"`
	PaymentGatewayID uint           `gorm:"not null;index"`
	PaymentGateway   PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	Config           string
}
