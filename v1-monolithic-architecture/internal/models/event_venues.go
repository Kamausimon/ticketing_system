package models

import (
	"time"

	"gorm.io/gorm"
)

type EventVenues struct {
	gorm.Model
	VenueID     uint   `gorm:"not null;uniqueIndex:idx_event_venue"`
	Venue       Venue  `gorm:"foreignKey:VenueID"`
	EventID     uint   `gorm:"not null;uniqueIndex:idx_event_venue"`
	Event       Event  `gorm:"foreignKey:EventID"`
	VenueRole   string `gorm:"default:'primary'"`
	SetupTime   *time.Time
	EventTime   *time.Time
	CleanupTime *time.Time
}
