package attendees

import (
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// AttendeeHandler handles all attendee-related operations
type AttendeeHandler struct {
	db       *gorm.DB
	_metrics *analytics.PrometheusMetrics // Reserved for future instrumentation
}

// NewAttendeeHandler creates a new attendee handler
func NewAttendeeHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *AttendeeHandler {
	return &AttendeeHandler{
		db:       db,
		_metrics: metrics,
	}
}

// AttendeeResponse represents the attendee response structure
type AttendeeResponse struct {
	ID                     uint       `json:"id"`
	OrderID                uint       `json:"order_id"`
	EventID                uint       `json:"event_id"`
	EventTitle             string     `json:"event_title"`
	TicketID               uint       `json:"ticket_id"`
	TicketNumber           string     `json:"ticket_number"`
	TicketClassName        string     `json:"ticket_class_name"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	Email                  string     `json:"email"`
	HasArrived             bool       `json:"has_arrived"`
	ArrivalTime            *time.Time `json:"arrival_time,omitempty"`
	AccountID              uint       `json:"account_id"`
	IsRefunded             bool       `json:"is_refunded"`
	PrivateReferenceNumber int        `json:"private_reference_number"`
	CreatedAt              time.Time  `json:"created_at"`
}

// AttendeeListResponse represents paginated attendees
type AttendeeListResponse struct {
	Attendees  []AttendeeResponse `json:"attendees"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

// AttendeeFilter represents filtering options
type AttendeeFilter struct {
	Page       int
	Limit      int
	EventID    *uint
	HasArrived *bool
	IsRefunded *bool
	SearchTerm string // Search by name or email
}

// CheckInRequest represents check-in request
type CheckInRequest struct {
	TicketNumber string `json:"ticket_number"`
	CheckedInBy  uint   `json:"checked_in_by"`
}

// BulkCheckInRequest represents bulk check-in request
type BulkCheckInRequest struct {
	TicketNumbers []string `json:"ticket_numbers"`
	CheckedInBy   uint     `json:"checked_in_by"`
}

// UpdateAttendeeRequest represents attendee update request
type UpdateAttendeeRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// AttendanceStats represents attendance statistics
type AttendanceStats struct {
	TotalAttendees  int64           `json:"total_attendees"`
	CheckedIn       int64           `json:"checked_in"`
	NotCheckedIn    int64           `json:"not_checked_in"`
	RefundedTickets int64           `json:"refunded_tickets"`
	CheckInRate     float64         `json:"check_in_rate"`
	ArrivalTrend    []HourlyArrival `json:"arrival_trend,omitempty"`
}

// HourlyArrival represents arrivals per hour
type HourlyArrival struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// ExportFormat represents export file format
type ExportFormat string

const (
	ExportCSV   ExportFormat = "csv"
	ExportExcel ExportFormat = "excel"
	ExportPDF   ExportFormat = "pdf"
)

// Helper function to convert models.Attendee to AttendeeResponse
func convertToAttendeeResponse(attendee models.Attendee) AttendeeResponse {
	response := AttendeeResponse{
		ID:                     attendee.ID,
		OrderID:                attendee.OrderID,
		EventID:                attendee.EventID,
		TicketID:               attendee.TicketID,
		FirstName:              attendee.FirstName,
		LastName:               attendee.LastName,
		Email:                  attendee.Email,
		HasArrived:             attendee.HasArrived,
		ArrivalTime:            attendee.ArrivalTime,
		AccountID:              attendee.AccountID,
		IsRefunded:             attendee.IsRefunded,
		PrivateReferenceNumber: attendee.PrivateReferenceNumber,
		CreatedAt:              attendee.CreatedAt,
	}

	// Add event title if loaded
	if attendee.Event.ID > 0 {
		response.EventTitle = attendee.Event.Title
	}

	// Add ticket information if loaded
	if attendee.Ticket.ID > 0 {
		response.TicketNumber = attendee.Ticket.TicketNumber
		if attendee.Ticket.OrderItem.ID > 0 && attendee.Ticket.OrderItem.TicketClass.ID > 0 {
			response.TicketClassName = attendee.Ticket.OrderItem.TicketClass.Name
		}
	}

	return response
}
