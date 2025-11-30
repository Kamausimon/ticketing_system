package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketClass struct {
	gorm.Model
	EventID             uint   `gorm:"not null;index"`
	Event               Event  `gorm:"foreignKey:EventID"`
	Name                string `gorm:"not null"`
	Description         string
	Price               Money  `gorm:"not null"` // Use Money type instead of float32
	Currency            string `gorm:"not null;default:'KSH'"`
	MaxPerOrder         *int
	MinPerOrder         *int `gorm:"default:1"`
	QuantityAvailable   *int
	QuantitySold        int `gorm:"default:0"`
	Version             int `gorm:"default:0"` // Optimistic locking version field
	StartSaleDate       *time.Time
	EndSaleDate         *time.Time
	SalesVolume         Money `gorm:"default:0"` // Use Money type instead of float32
	OrganizerFeesVolume Money `gorm:"default:0"` // Use Money type instead of float32
	IsPaused            bool  `gorm:"default:false"`
	IsHidden            bool  `gorm:"default:false"`
	SortOrder           int   `gorm:"default:0"`
	RequiresApproval    bool  `gorm:"default:false"`
}
