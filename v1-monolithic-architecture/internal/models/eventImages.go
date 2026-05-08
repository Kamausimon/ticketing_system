package models

import "gorm.io/gorm"

type EventImages struct {
	gorm.Model
	ImagePath string
	EventID   uint    `gorm:"not null;index"`
	Event     Event   `gorm:"foreignKey:EventID"`
	AccountID uint    `gorm:"not null;index"`
	Account   Account `gorm:"foreignKey:AccountID"`
	UserID    uint    `gorm:"not null;index"`
	User      User    `gorm:"foreignKey:UserID"`
}
