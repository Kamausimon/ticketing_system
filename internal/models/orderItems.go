package models

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model
	OrderID          uint        `gorm:"not null;index"`
	Order            Order       `gorm:"foreignKey:OrderID"`
	TicketClassID    uint        `gorm:"not null;index"`
	TicketClass      TicketClass `gorm:"foreignKey:TicketClassID"`
	Quantity         int         `gorm:"not null"`
	UnitPrice        Money       `gorm:"not null"` // Use Money type instead of float32
	TotalPrice       Money       `gorm:"not null"` // Use Money type instead of float32
	Discount         *Money      // Use Money type instead of float32
	PromoCodeUsed    *string
	GeneratedTickets []Ticket `gorm:"foreignKey:OrderItemID"`
}
