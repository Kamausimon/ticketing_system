package attendees

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ticketing_system/internal/models"
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
		http.Error(w, "PDF export not yet implemented", http.StatusNotImplemented)
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
