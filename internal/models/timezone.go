package models

import "gorm.io/gorm"

type Timezone struct {
	gorm.Model
	Name        string `gorm:"not null;uniqueIndex"`
	DisplayName string `gorm:"not null"`
	Offset      string `gorm:"not null"` // e.g., "+03:00", "-05:00"
	IanaName    string `gorm:"not null"` // e.g., "Africa/Nairobi", "America/New_York"
	IsActive    bool   `gorm:"default:true"`
}
