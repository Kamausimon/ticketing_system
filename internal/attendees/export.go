package attendees

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ticketing_system/internal/models"

	"github.com/jung-kurt/gofpdf"
)

// ExportAttendeeList exports attendee list in specified format
func (h *AttendeeHandler) ExportAttendeeList(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv" // Default to CSV
	}

	// Get attendees
	var attendees []models.Attendee
	if err := h.db.Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ?", uint(eventID)).
		Order("last_name, first_name").
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
		return
	}

	switch ExportFormat(format) {
	case ExportCSV:
		h.exportCSV(w, attendees)
	case ExportPDF:
		h.exportPDF(w, attendees)
	default:
		h.exportJSON(w, attendees)
	}
}

// exportCSV exports attendees as CSV
func (h *AttendeeHandler) exportCSV(w http.ResponseWriter, attendees []models.Attendee) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=attendees_%s.csv", time.Now().Format("20060102")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "First Name", "Last Name", "Email", "Ticket Number", "Ticket Type", "Checked In", "Arrival Time", "Is Refunded"}
	writer.Write(header)

	// Write data
	for _, attendee := range attendees {
		checkedIn := "No"
		arrivalTime := ""
		if attendee.HasArrived {
			checkedIn = "Yes"
			if attendee.ArrivalTime != nil {
				arrivalTime = attendee.ArrivalTime.Format("2006-01-02 15:04:05")
			}
		}

		refunded := "No"
		if attendee.IsRefunded {
			refunded = "Yes"
		}

		ticketNumber := ""
		ticketType := ""
		if attendee.Ticket.ID > 0 {
			ticketNumber = attendee.Ticket.TicketNumber
			if attendee.Ticket.OrderItem.ID > 0 && attendee.Ticket.OrderItem.TicketClass.ID > 0 {
				ticketType = attendee.Ticket.OrderItem.TicketClass.Name
			}
		}

		row := []string{
			fmt.Sprintf("%d", attendee.ID),
			attendee.FirstName,
			attendee.LastName,
			attendee.Email,
			ticketNumber,
			ticketType,
			checkedIn,
			arrivalTime,
			refunded,
		}
		writer.Write(row)
	}
}

// exportJSON exports attendees as JSON
func (h *AttendeeHandler) exportJSON(w http.ResponseWriter, attendees []models.Attendee) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=attendees_%s.json", time.Now().Format("20060102")))

	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	json.NewEncoder(w).Encode(responses)
}

// exportPDF exports attendees as PDF
func (h *AttendeeHandler) exportPDF(w http.ResponseWriter, attendees []models.Attendee) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Attendee List")
	pdf.Ln(12)

	// Event info (if available)
	if len(attendees) > 0 && attendees[0].Event.ID > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, fmt.Sprintf("Event: %s", attendees[0].Event.Title))
		pdf.Ln(10)
	}

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 220, 255)
	pdf.CellFormat(10, 8, "ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "First Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Last Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, "Email", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Ticket Number", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Ticket Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Checked In", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Refunded", "1", 1, "C", true, 0, "")

	// Table data
	pdf.SetFont("Arial", "", 8)
	for _, attendee := range attendees {
		checkedIn := "No"
		if attendee.HasArrived {
			checkedIn = "Yes"
		}

		refunded := "No"
		if attendee.IsRefunded {
			refunded = "Yes"
		}

		ticketNumber := ""
		ticketType := ""
		if attendee.Ticket.ID > 0 {
			ticketNumber = attendee.Ticket.TicketNumber
			if attendee.Ticket.OrderItem.ID > 0 && attendee.Ticket.OrderItem.TicketClass.ID > 0 {
				ticketType = attendee.Ticket.OrderItem.TicketClass.Name
			}
		}

		pdf.CellFormat(10, 6, fmt.Sprintf("%d", attendee.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, attendee.FirstName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, attendee.LastName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, attendee.Email, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, ticketNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, ticketType, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, checkedIn, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, refunded, "1", 1, "C", false, 0, "")
	}

	// Summary
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Total Attendees: %d", len(attendees)))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=attendees_%s.pdf", time.Now().Format("20060102")))

	if err := pdf.Output(w); err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// ExportBadgeData exports data for badge printing
func (h *AttendeeHandler) ExportBadgeData(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var attendees []models.Attendee
	if err := h.db.Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ? AND is_refunded = ?", uint(eventID), false).
		Order("last_name, first_name").
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
		return
	}

	type BadgeData struct {
		FullName     string `json:"full_name"`
		Email        string `json:"email"`
		TicketNumber string `json:"ticket_number"`
		TicketType   string `json:"ticket_type"`
		EventTitle   string `json:"event_title"`
		QRCode       string `json:"qr_code"`
	}

	badges := make([]BadgeData, len(attendees))
	for i, attendee := range attendees {
		fullName := fmt.Sprintf("%s %s", attendee.FirstName, attendee.LastName)
		ticketType := ""
		ticketNumber := ""
		qrCode := ""
		eventTitle := ""

		if attendee.Event.ID > 0 {
			eventTitle = attendee.Event.Title
		}

		if attendee.Ticket.ID > 0 {
			ticketNumber = attendee.Ticket.TicketNumber
			qrCode = attendee.Ticket.QRCode
			if attendee.Ticket.OrderItem.ID > 0 && attendee.Ticket.OrderItem.TicketClass.ID > 0 {
				ticketType = attendee.Ticket.OrderItem.TicketClass.Name
			}
		}

		badges[i] = BadgeData{
			FullName:     fullName,
			Email:        attendee.Email,
			TicketNumber: ticketNumber,
			TicketType:   ticketType,
			EventTitle:   eventTitle,
			QRCode:       qrCode,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(badges)
}
