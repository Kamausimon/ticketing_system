package models

import "gorm.io/gorm"

// RefundLineItem represents individual items/tickets being refunded
// This allows tracking partial refunds at the ticket level
type RefundLineItem struct {
	gorm.Model

	// Link to parent refund
	RefundRecordID uint         `gorm:"not null;index"`
	RefundRecord   RefundRecord `gorm:"foreignKey:RefundRecordID"`

	// What's being refunded
	OrderItemID uint      `gorm:"not null;index"`
	OrderItem   OrderItem `gorm:"foreignKey:OrderItemID"`

	TicketID *uint   `gorm:"index"` // Specific ticket (for ticket-level refunds)
	Ticket   *Ticket `gorm:"foreignKey:TicketID"`

	// Refund details
	Quantity     int   `gorm:"not null"` // How many tickets from this order item
	RefundAmount Money `gorm:"not null"` // Amount for this line item (in cents)

	// Optional line-item specific reason
	Reason      *string // Additional context for this specific item
	Description string  `gorm:"not null"` // Human readable description

	// Soft delete
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name
func (RefundLineItem) TableName() string {
	return "refund_line_items"
}
