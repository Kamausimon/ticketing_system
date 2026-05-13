package models

import (
	"time"

	"gorm.io/gorm"
)

type Attendee struct {
	gorm.Model
	OrderID                uint   `gorm:"not null;index"`
	Order                  Order  `gorm:"foreignKey:OrderID"`
	EventID                uint   `gorm:"not null;index"`
	Event                  Event  `gorm:"foreignKey:EventID"`
	TicketID               uint   `gorm:"not null;index"`
	Ticket                 Ticket `gorm:"foreignKey:TicketID"`
	FirstName              string
	LastName               string
	Email                  string
	HasArrived             bool
	ArrivalTime            *time.Time
	AccountID              uint    `gorm:"not null;index"`
	Account                Account `gorm:"foreignKey:AccountID"`
	IsRefunded             bool    `gorm:"default:false"`
	PrivateReferenceNumber int
}
