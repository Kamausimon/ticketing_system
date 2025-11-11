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
	Price               float32 `gorm:"not null"`
	Currency            string  `gorm:"not null;default:'KSH'"`
	MaxPerOrder         *int
	MinPerOrder         *int `gorm:"default:1"`
	QuantityAvailable   *int
	QuantitySold        int `gorm:"default:0"`
	StartSaleDate       *time.Time
	EndSaleDate         *time.Time
	SalesVolume         float32
	OrganizerFeesVolume float32
	IsPaused            bool `gorm:"default:false"`
	IsHidden            bool `gorm:"default:false"`
	SortOrder           int  `gorm:"default:0"`
	RequiresApproval    bool `gorm:"default:false"`
}
