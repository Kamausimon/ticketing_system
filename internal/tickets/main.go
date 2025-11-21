package tickets

import (
	"fmt"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// TicketHandler handles all ticket-related operations
type TicketHandler struct {
	db *gorm.DB
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{db: db}
}

// TicketResponse represents the ticket response structure
type TicketResponse struct {
	ID            uint                `json:"id"`
	TicketNumber  string              `json:"ticket_number"`
	OrderID       uint                `json:"order_id"`
	OrderNumber   string              `json:"order_number"`
	EventID       uint                `json:"event_id"`
	EventTitle    string              `json:"event_title"`
	EventDate     time.Time           `json:"event_date"`
	EventLocation string              `json:"event_location"`
	TicketClass   string              `json:"ticket_class"`
	HolderName    string              `json:"holder_name"`
	HolderEmail   string              `json:"holder_email"`
	QRCode        string              `json:"qr_code"`
	Status        models.TicketStatus `json:"status"`
	CheckedInAt   *time.Time          `json:"checked_in_at,omitempty"`
	PdfPath       string              `json:"pdf_path,omitempty"`
	Price         float64             `json:"price"`
	Currency      string              `json:"currency"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// TicketListResponse represents a paginated list of tickets
type TicketListResponse struct {
	Tickets    []TicketResponse `json:"tickets"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

// TicketFilter represents filtering options for tickets
type TicketFilter struct {
	Page       int
	Limit      int
	Status     *models.TicketStatus
	EventID    *uint
	OrderID    *uint
	StartDate  *time.Time
	EndDate    *time.Time
	SearchTerm string // Search by ticket number, holder name, email
}

// TicketStats represents ticket statistics
type TicketStats struct {
	TotalTickets     int64   `json:"total_tickets"`
	ActiveTickets    int64   `json:"active_tickets"`
	UsedTickets      int64   `json:"used_tickets"`
	CancelledTickets int64   `json:"cancelled_tickets"`
	RefundedTickets  int64   `json:"refunded_tickets"`
	CheckInRate      float64 `json:"check_in_rate"`
}

// GenerateTicketsRequest represents the request to generate tickets
type GenerateTicketsRequest struct {
	OrderID uint `json:"order_id"`
}

// CheckInRequest represents the check-in request
type CheckInRequest struct {
	TicketNumber string `json:"ticket_number"`
	EventID      uint   `json:"event_id"`
}

// TransferTicketRequest represents the ticket transfer request
type TransferTicketRequest struct {
	NewHolderName  string `json:"new_holder_name"`
	NewHolderEmail string `json:"new_holder_email"`
}

// ValidateTicketRequest represents the validation request
type ValidateTicketRequest struct {
	TicketNumber string `json:"ticket_number"`
	EventID      uint   `json:"event_id"`
}

// CheckInStats represents check-in statistics
type CheckInStats struct {
	TotalTickets int64      `json:"total_tickets"`
	CheckedIn    int64      `json:"checked_in"`
	NotCheckedIn int64      `json:"not_checked_in"`
	CheckInRate  float64    `json:"check_in_rate"`
	LastCheckIn  *time.Time `json:"last_check_in,omitempty"`
}

// Helper function to convert models.Ticket to TicketResponse
func convertToTicketResponse(ticket models.Ticket) TicketResponse {
	response := TicketResponse{
		ID:           ticket.ID,
		TicketNumber: ticket.TicketNumber,
		HolderName:   ticket.HolderName,
		HolderEmail:  ticket.HolderEmail,
		QRCode:       ticket.QRCode,
		Status:       ticket.Status,
		CheckedInAt:  ticket.CheckedInAt,
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    ticket.UpdatedAt,
	}

	if ticket.PdfPath != nil {
		response.PdfPath = *ticket.PdfPath
	}

	// Add order item details if loaded
	if ticket.OrderItem.ID > 0 {
		response.OrderID = ticket.OrderItem.OrderID
		response.OrderNumber = fmt.Sprintf("ORD-%d", ticket.OrderItem.OrderID)
		response.Price = float64(ticket.OrderItem.UnitPrice)

		// Add ticket class details
		if ticket.OrderItem.TicketClass.ID > 0 {
			response.TicketClass = ticket.OrderItem.TicketClass.Name
			response.EventID = ticket.OrderItem.TicketClass.EventID
			response.Currency = ticket.OrderItem.TicketClass.Currency

			// Add event details if loaded
			if ticket.OrderItem.TicketClass.Event.ID > 0 {
				response.EventTitle = ticket.OrderItem.TicketClass.Event.Title
				response.EventDate = ticket.OrderItem.TicketClass.Event.StartDate
				response.EventLocation = ticket.OrderItem.TicketClass.Event.Location
			}
		}
	}

	return response
}

// Helper function to generate unique ticket number
func generateTicketNumber(eventID, orderID, ticketID uint) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TKT-%d-%d-%d-%d", eventID, orderID, ticketID, timestamp)
}

// Helper function to generate QR code data
func generateQRCodeData(ticketNumber string) string {
	// In production, this would generate an actual QR code image
	// For now, return the data that would be encoded
	return fmt.Sprintf("TICKET:%s:VERIFIED:%d", ticketNumber, time.Now().Unix())
}
