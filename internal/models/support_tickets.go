package models

import (
	"time"

	"gorm.io/gorm"
)

// SupportTicketStatus represents the status of a support ticket
type SupportTicketStatus string

const (
	TicketStatusOpen       SupportTicketStatus = "open"
	TicketStatusInProgress SupportTicketStatus = "in_progress"
	TicketStatusResolved   SupportTicketStatus = "resolved"
	TicketStatusClosed     SupportTicketStatus = "closed"
)

// SupportTicketPriority represents the priority level of a ticket
type SupportTicketPriority string

const (
	TicketPriorityCritical SupportTicketPriority = "critical"
	TicketPriorityHigh     SupportTicketPriority = "high"
	TicketPriorityMedium   SupportTicketPriority = "medium"
	TicketPriorityLow      SupportTicketPriority = "low"
)

// SupportTicketCategory represents the category of a support ticket
type SupportTicketCategory string

const (
	TicketCategoryPayment        SupportTicketCategory = "payment"
	TicketCategoryBooking        SupportTicketCategory = "booking"
	TicketCategoryAccount        SupportTicketCategory = "account"
	TicketCategoryEvent          SupportTicketCategory = "event"
	TicketCategoryTechnical      SupportTicketCategory = "technical"
	TicketCategoryRefund         SupportTicketCategory = "refund"
	TicketCategoryGeneral        SupportTicketCategory = "general"
	TicketCategoryFeatureRequest SupportTicketCategory = "feature_request"
)

// SupportTicket represents a support ticket from a customer or organizer
type SupportTicket struct {
	gorm.Model
	TicketNumber string                `gorm:"type:varchar(20);unique;not null;index" json:"ticket_number"`
	Subject      string                `gorm:"type:varchar(255);not null" json:"subject"`
	Description  string                `gorm:"type:text;not null" json:"description"`
	Category     SupportTicketCategory `gorm:"type:varchar(50);not null;index" json:"category"`
	Priority     SupportTicketPriority `gorm:"type:varchar(20);default:'medium';index" json:"priority"`
	Status       SupportTicketStatus   `gorm:"type:varchar(20);default:'open';index" json:"status"`

	// Submitter info
	UserID      *uint  `gorm:"index" json:"user_id,omitempty"`
	User        *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Email       string `gorm:"type:varchar(255);not null;index" json:"email"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber string `gorm:"type:varchar(50)" json:"phone_number,omitempty"`

	// Related entities (optional)
	OrderID     *uint      `gorm:"index" json:"order_id,omitempty"`
	Order       *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	EventID     *uint      `gorm:"index" json:"event_id,omitempty"`
	Event       *Event     `gorm:"foreignKey:EventID" json:"event,omitempty"`
	OrganizerID *uint      `gorm:"index" json:"organizer_id,omitempty"`
	Organizer   *Organizer `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`

	// Assignment and resolution
	AssignedToID    *uint      `gorm:"index" json:"assigned_to_id,omitempty"`
	AssignedTo      *User      `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolvedByID    *uint      `json:"resolved_by_id,omitempty"`
	ResolvedBy      *User      `gorm:"foreignKey:ResolvedByID" json:"resolved_by,omitempty"`
	ResolutionNotes string     `gorm:"type:text" json:"resolution_notes,omitempty"`

	// AI classification metadata
	AIClassified      bool    `gorm:"default:false" json:"ai_classified"`
	AIPriority        string  `gorm:"type:varchar(20)" json:"ai_priority,omitempty"`
	AIConfidenceScore float64 `gorm:"type:decimal(5,4)" json:"ai_confidence_score,omitempty"`
	AIReasoning       string  `gorm:"type:text" json:"ai_reasoning,omitempty"`

	// Comments/Updates
	Comments []SupportTicketComment `gorm:"foreignKey:TicketID" json:"comments,omitempty"`
}

// SupportTicketComment represents a comment/update on a support ticket
type SupportTicketComment struct {
	gorm.Model
	TicketID    uint          `gorm:"not null;index" json:"ticket_id"`
	Ticket      SupportTicket `gorm:"foreignKey:TicketID" json:"-"`
	UserID      *uint         `gorm:"index" json:"user_id,omitempty"`
	User        *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comment     string        `gorm:"type:text;not null" json:"comment"`
	IsInternal  bool          `gorm:"default:false" json:"is_internal"` // Internal notes not visible to customer
	AuthorName  string        `gorm:"type:varchar(255)" json:"author_name"`
	AuthorEmail string        `gorm:"type:varchar(255)" json:"author_email"`
}
