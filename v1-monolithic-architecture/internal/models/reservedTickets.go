package models

import (
	"time"

	"gorm.io/gorm"
)

type ReservedTicket struct {
	gorm.Model
	TicketID         uint   `gorm:"not null;index"`
	Ticket           Ticket `gorm:"foreignKey:TicketID"`
	EventID          uint   `gorm:"not null;index"`
	Event            Event  `gorm:"foreignKey:EventID"`
	QuantityReserved int
	Expires          time.Time
	SessionID        string
}
