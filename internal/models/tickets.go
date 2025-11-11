package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketActive    TicketStatus = "active"
	TicketUsed      TicketStatus = "used"
	TicketCancelled TicketStatus = "cancelled"
	TicketRefunded  TicketStatus = "refunded"
)

type Ticket struct {
	gorm.Model
	OrderItemID  uint      `gorm:"not null;index"`
	OrderItem    OrderItem `gorm:"foreignKey:OrderItemID"`
	TicketNumber string    `gorm:"unique;not null"`
	QRCode       string    `gorm:"unique"`
	BarcodeData  string
	HolderName   string
	HolderEmail  string
	Status       TicketStatus `gorm:"default:'active'"`
	CheckedInAt  *time.Time
	CheckedInBy  *uint
	UsedAt       *time.Time
	RefundedAt   *time.Time
	PdfPath      *string
}
