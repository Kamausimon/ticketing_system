package tickets

import (
	"fmt"
	"os"
	"path/filepath"

	"ticketing_system/internal/models"
	"ticketing_system/pkg/pdf"
	"ticketing_system/pkg/qrcode"
)

// generateTicketPDF generates a PDF for a ticket
func (h *TicketHandler) generateTicketPDF(ticket *models.Ticket) (string, error) {
	// Load full ticket data with relations
	var fullTicket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event.Venue").
		Preload("OrderItem.Order").
		Where("id = ?", ticket.ID).
		First(&fullTicket).Error; err != nil {
		return "", fmt.Errorf("failed to load ticket data: %w", err)
	}

	// Get event and venue info
	event := fullTicket.OrderItem.TicketClass.Event
	order := fullTicket.OrderItem.Order
	ticketClass := fullTicket.OrderItem.TicketClass

	// Generate QR code
	qrContent := fmt.Sprintf("TICKET:%s|EVENT:%d|ATTENDEE:%s",
		fullTicket.TicketNumber,
		event.ID,
		fullTicket.HolderName,
	)

	qrGenerator := qrcode.NewGenerator().WithSize(512)
	qrBytes, err := qrGenerator.GenerateBytes(qrContent)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Prepare ticket data for PDF
	ticketData := pdf.TicketData{
		TicketNumber:  fullTicket.TicketNumber,
		EventName:     event.Title,
		EventDate:     event.StartDate,
		EventTime:     event.StartDate.Format("3:04 PM"),
		VenueName:     event.Location, // Use event location as fallback
		VenueAddress:  event.Location,
		AttendeeName:  fullTicket.HolderName,
		AttendeeEmail: fullTicket.HolderEmail,
		TicketType:    ticketClass.Name,
		SeatNumber:    "",
		Price:         float64(fullTicket.OrderItem.UnitPrice),
		Currency:      "USD",
		QRCode:        qrBytes,
		OrderNumber:   fmt.Sprintf("ORD-%d", order.ID),
		PurchaseDate:  order.CreatedAt,
		SpecialNotes:  "",
	}

	// If venue info is available, use it
	if len(event.Venue) > 0 {
		venue := event.Venue[0]
		ticketData.VenueName = venue.VenueName
		ticketData.VenueAddress = fmt.Sprintf("%s, %s, %s %s",
			venue.Address,
			venue.City,
			venue.State,
			venue.ZipCode,
		)
	}

	// Create storage directory
	storageDir := filepath.Join("storage", "tickets", fmt.Sprintf("%d", order.ID))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate PDF
	pdfGenerator := pdf.NewTicketGenerator()
	pdfFileName := fmt.Sprintf("ticket_%s.pdf", fullTicket.TicketNumber)
	pdfPath := filepath.Join(storageDir, pdfFileName)

	if err := pdfGenerator.GenerateToFile(ticketData, pdfPath); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	fmt.Printf("✅ Generated PDF for ticket %s at %s\n", fullTicket.TicketNumber, pdfPath)

	// Return relative path for storage
	return pdfPath, nil
}

// RegeneratePDF regenerates the PDF for a ticket
func (h *TicketHandler) RegeneratePDF(ticket *models.Ticket) (string, error) {
	return h.generateTicketPDF(ticket)
}

// GenerateBatchPDFs generates PDFs for multiple tickets
func (h *TicketHandler) GenerateBatchPDFs(tickets []models.Ticket) (map[uint]string, error) {
	results := make(map[uint]string)
	errors := make([]error, 0)

	for _, ticket := range tickets {
		pdfPath, err := h.generateTicketPDF(&ticket)
		if err != nil {
			errors = append(errors, fmt.Errorf("ticket %s: %w", ticket.TicketNumber, err))
			continue
		}
		results[ticket.ID] = pdfPath
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("failed to generate %d PDFs: %v", len(errors), errors[0])
	}

	return results, nil
}

// Example PDF generation with actual library (commented out):
/*
import (
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

func (h *TicketHandler) generateTicketPDF(ticket *models.Ticket) (string, error) {
	// Generate QR code image
	qrFilePath := fmt.Sprintf("/tmp/qr_%s.png", ticket.TicketNumber)
	if err := qrcode.WriteFile(ticket.QRCode, qrcode.Medium, 256, qrFilePath); err != nil {
		return "", err
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add title
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(0, 20, ticket.OrderItem.TicketClass.Event.Title)
	pdf.Ln(25)

	// Add event details
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Date: %s", ticket.OrderItem.TicketClass.Event.StartDate.Format("January 2, 2006")))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Location: %s", ticket.OrderItem.TicketClass.Event.Location))
	pdf.Ln(15)

	// Add ticket details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, fmt.Sprintf("Ticket Number: %s", ticket.TicketNumber))
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Class: %s", ticket.OrderItem.TicketClass.Name))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Holder: %s", ticket.HolderName))
	pdf.Ln(20)

	// Add QR code
	pdf.Image(qrFilePath, 75, pdf.GetY(), 60, 60, false, "", 0, "")

	// Save PDF
	pdfPath := fmt.Sprintf("/storage/tickets/%d/ticket_%s.pdf", ticket.OrderItem.OrderID, ticket.TicketNumber)
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return "", err
	}

	return pdfPath, nil
}
*/
