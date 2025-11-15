package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionPayment TransactionType = "payment"      // Customer pays
	TransactionRefund  TransactionType = "refund"       // Money back to customer
	TransactionPayout  TransactionType = "payout"       // Platform pays organizer
	TransactionFee     TransactionType = "platform_fee" // Platform commission
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionFailed    TransactionStatus = "failed"
	TransactionHeld      TransactionStatus = "held"    // Waiting for settlement
	TransactionSettled   TransactionStatus = "settled" // Paid to organizer
	TransactionCancelled TransactionStatus = "cancelled"
)

type Money int64

func (m Money) String() string {
	return fmt.Sprintf("%.2f", float64(m)/100)
}

type Percentage int64

func (p Percentage) ToPercent() float64 {
	return float64(p) / 100
}

type PaymentTransaction struct {
	gorm.Model

	Amount   Money             `gorm:"not null"`
	Currency string            `gorm:"not null;default:'KSH'"`
	Type     TransactionType   `gorm:"not null;index"`
	Status   TransactionStatus `gorm:"not null;index"`

	OrderID          *uint           `gorm:"index"`
	Order            *Order          `gorm:"foreignKey:OrderID"`
	PaymentGatewayID *uint           `gorm:"index"`
	PaymentGateway   *PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	OrganizerID      *uint           `gorm:"index"`
	Organizer        *Organizer      `gorm:"foreignKey:OrganizerID"`

	ExternalTransactionID *string `gorm:"index"`
	ExternalReference     *string

	ProcessedAt *time.Time
	SettledAt   *time.Time

	Description string
	Notes       *string

	ParentTransactionID *uint
	ParentTransaction   *PaymentTransaction `gorm:"foreignKey:ParentTransactionID"`
}
