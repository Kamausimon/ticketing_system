package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderPending       OrderStatus = "pending"
	OrderPaid          OrderStatus = "paid"
	OrderFulfilled     OrderStatus = "fulfilled"
	OrderCancelled     OrderStatus = "cancelled"
	OrderRefunded      OrderStatus = "refunded"
	OrderPartialRefund OrderStatus = "partial_refund"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type Order struct {
	gorm.Model
	AccountID           uint    `gorm:"not null;index"`
	Account             Account `gorm:"foreignKey:AccountID"`
	FirstName           string
	LastName            string
	Email               string
	BusinessName        *string
	BusinessTaxNumber   *string
	BusinessAddressLine *string
	TicketPdfPath       *string
	OrderPreference     string
	TransactionID       *uint
	Discount            *float32
	BookingFee          *float32
	OrganizerBookingFee *float32
	OrderDate           *time.Time
	Notes               *string
	IsDeleted           bool `gorm:"default:false"`
	IsCancelled         bool `gorm:"default:false"`
	IsPartiallyRefunded bool `gorm:"default:false"`
	Amount              float32
	AmountRefunded      *float32
	EventID             uint  `gorm:"not null;index"`
	Event               Event `gorm:"foreignKey:EventID"`
	PaymentGatewayID    *uint
	PaymentGateway      PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`
	IsPaymentReceived   bool           `gorm:"default:false"`
	IsBusiness          bool
	TaxAmount           float32
	Status              OrderStatus   `gorm:"not null;default:'pending';index"`
	PaymentStatus       PaymentStatus `gorm:"not null;default:'pending'"`
	OrderItems          []OrderItem   `gorm:"foreignKey:OrderID"`
	Currency            string        `gorm:"not null;default:'KSH'"`
	CompletedAt         *time.Time
	CancelledAt         *time.Time
	RefundedAt          *time.Time
}
