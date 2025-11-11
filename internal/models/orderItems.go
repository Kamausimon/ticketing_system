package models

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model
	OrderID          uint        `gorm:"not null;index"`
	Order            Order       `gorm:"foreignKey:OrderID"`
	TicketClassID    uint        `gorm:"not null;index"`
	TicketClass      TicketClass `gorm:"foreignKey:TicketClassID"`
	Quantity         int         `gorm:"not null"`
	UnitPrice        float32     `gorm:"not null"`
	TotalPrice       float32     `gorm:"not null"`
	Discount         *float32
	PromoCodeUsed    *string
	GeneratedTickets []Ticket `gorm:"foreignKey:OrderItemID"`
}
