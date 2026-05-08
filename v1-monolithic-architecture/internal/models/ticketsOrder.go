package models

import "gorm.io/gorm"

type TicketOrder struct {
	gorm.Model
	OrderID  uint   `gorm:"not null;index"`
	Order    Order  `gorm:"foreignKey:OrderID"`
	TicketID uint   `gorm:"not null;index"`
	Ticket   Ticket `gorm:"foreignKey:TicketID"`
}
