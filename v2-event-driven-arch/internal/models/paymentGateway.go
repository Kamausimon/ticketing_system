package models

import "gorm.io/gorm"

type PaymentGateway struct {
	gorm.Model
	ProviderName string
	ProviderURL  string
	IsOnSite     bool
	CanRefund    bool
	Name         string
}
