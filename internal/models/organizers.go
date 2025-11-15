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
}
