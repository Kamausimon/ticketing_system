package models

import "gorm.io/gorm"

type Currency struct {
	gorm.Model
	Code     string `gorm:"not null;uniqueIndex;size:3"` // e.g., "USD", "KSH", "EUR"
	Name     string `gorm:"not null"`                    // e.g., "US Dollar", "Kenyan Shilling"
	Symbol   string `gorm:"not null"`                    // e.g., "$", "KSh", "€"
	IsActive bool   `gorm:"default:true"`
}
