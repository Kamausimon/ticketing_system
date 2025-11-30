package models

import (
	"time"

	"gorm.io/gorm"
)

// WaitlistEntry represents a user waiting for sold-out tickets
type WaitlistEntry struct {
	gorm.Model
	EventID       uint         `gorm:"not null;index"`
	Event         Event        `gorm:"foreignKey:EventID"`
	TicketClassID *uint        `gorm:"index"` // NULL means any ticket class
	TicketClass   *TicketClass `gorm:"foreignKey:TicketClassID"`
	Email         string       `gorm:"not null;index"`
	Name          string       `gorm:"not null"`
	Phone         *string
	Quantity      int    `gorm:"not null;default:1"`
	Status        string `gorm:"default:'waiting';index"` // waiting, notified, converted, expired
	NotifiedAt    *time.Time
	ConvertedAt   *time.Time
	ExpiresAt     *time.Time
	Priority      int    `gorm:"default:0"` // Higher priority = notified first
	SessionID     string `gorm:"index"`
	UserID        *uint  `gorm:"index"`
	User          *User  `gorm:"foreignKey:UserID"`
}
