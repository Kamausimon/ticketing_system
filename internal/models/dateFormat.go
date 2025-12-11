package models

import "gorm.io/gorm"

type DateFormat struct {
	gorm.Model
	Format   string `gorm:"not null;uniqueIndex"` // e.g., "YYYY-MM-DD", "DD/MM/YYYY"
	Example  string `gorm:"not null"`             // e.g., "2024-12-25", "25/12/2024"
	IsActive bool   `gorm:"default:true"`
}

type DateTimeFormat struct {
	gorm.Model
	Format   string `gorm:"not null;uniqueIndex"` // e.g., "YYYY-MM-DD HH:mm", "DD/MM/YYYY HH:mm"
	Example  string `gorm:"not null"`             // e.g., "2024-12-25 14:30", "25/12/2024 02:30 PM"
	IsActive bool   `gorm:"default:true"`
}
