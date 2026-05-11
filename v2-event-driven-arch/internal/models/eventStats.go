package models

import (
	"time"

	"gorm.io/gorm"
)

type StatsGranularity string

const (
	StatsHourly  StatsGranularity = "hourly"
	StatsDaily   StatsGranularity = "daily"
	StatsWeekly  StatsGranularity = "weekly"
	StatsMonthly StatsGranularity = "monthly"
)

type EventStats struct {
	gorm.Model
	Date               time.Time `gorm:"index:idx_events_stats_lookup"`
	Day                time.Time `gorm:"type:date"`
	Hour               int       `gorm:"check:hour >=0 AND hour <=23"`
	Views              int
	UniqueViews        int
	TicketsSold        int
	SalesVolume        float32
	OrganizerFeeVolume float32
	EventID            uint  `gorm:"not null;index:idx_events_stats_lookup"`
	Event              Event `gorm:"foreignKey:EventID"`
	AddToCartCount     int
	CheckOutStart      int
	ConversionRate     float32 //tickets_sold/views
	GrossRevenue       float32
	NetRevenue         float32
	PlatformFees       float32
	PaymentFees        float32
	AverageTimeOnPage  int
	BounceRate         float32
	Granularity        StatsGranularity `gorm:"not null"`
}
