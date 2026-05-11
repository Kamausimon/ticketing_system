package models

import (
	"time"

	"gorm.io/gorm"
)

// TicketTransferHistory tracks all transfers of a ticket
type TicketTransferHistory struct {
	gorm.Model
	TicketID        uint      `gorm:"not null;index"`
	Ticket          Ticket    `gorm:"foreignKey:TicketID"`
	FromHolderName  string    `gorm:"not null"`
	FromHolderEmail string    `gorm:"not null"`
	ToHolderName    string    `gorm:"not null"`
	ToHolderEmail   string    `gorm:"not null"`
	TransferredBy   uint      `gorm:"not null"` // User ID who initiated the transfer
	TransferredAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	TransferReason  string    // Optional reason for transfer
	IPAddress       string    // IP address from which transfer was made
	UserAgent       string    // Browser/client info
}
