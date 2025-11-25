package attendees

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// CheckInAttendee checks in an attendee
func (h *AttendeeHandler) CheckInAttendee(w http.ResponseWriter, r *http.Request) {
	var req CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TicketNumber == "" {
		http.Error(w, "Ticket number is required", http.StatusBadRequest)
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Find ticket
		var ticket models.Ticket
		if err := tx.Where("ticket_number = ?", req.TicketNumber).First(&ticket).Error; err != nil {
			return fmt.Errorf("ticket not found")
		}

		// Check ticket status
		if ticket.Status != models.TicketActive {
			return fmt.Errorf("ticket is not active (status: %s)", ticket.Status)
		}

		// Find attendee
		var attendee models.Attendee
		if err := tx.Where("ticket_id = ?", ticket.ID).First(&attendee).Error; err != nil {
			return fmt.Errorf("attendee not found")
		}

		// Check if already checked in
		if attendee.HasArrived {
			return fmt.Errorf("attendee already checked in at %v", attendee.ArrivalTime)
		}

		// Check if refunded
		if attendee.IsRefunded {
			return fmt.Errorf("ticket has been refunded")
		}

		// Update attendee
		now := time.Now()
		attendee.HasArrived = true
		attendee.ArrivalTime = &now

		if err := tx.Save(&attendee).Error; err != nil {
			return err
		}

		// Update ticket
		ticket.CheckedInAt = &now
		ticket.CheckedInBy = &req.CheckedInBy
		ticket.Status = models.TicketUsed

		return tx.Save(&ticket).Error
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Check-in successful",
		"ticket":  req.TicketNumber,
	})
}

// BulkCheckIn checks in multiple attendees
func (h *AttendeeHandler) BulkCheckIn(w http.ResponseWriter, r *http.Request) {
	var req BulkCheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.TicketNumbers) == 0 {
		http.Error(w, "Ticket numbers are required", http.StatusBadRequest)
		return
	}

	successes := []string{}
	failures := map[string]string{}

	for _, ticketNumber := range req.TicketNumbers {
		err := h.db.Transaction(func(tx *gorm.DB) error {
			var ticket models.Ticket
			if err := tx.Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
				return fmt.Errorf("ticket not found")
			}

			if ticket.Status != models.TicketActive {
				return fmt.Errorf("ticket is not active")
			}

			var attendee models.Attendee
			if err := tx.Where("ticket_id = ?", ticket.ID).First(&attendee).Error; err != nil {
				return fmt.Errorf("attendee not found")
			}

			if attendee.HasArrived {
				return fmt.Errorf("already checked in")
			}

			if attendee.IsRefunded {
				return fmt.Errorf("ticket refunded")
			}

			now := time.Now()
			attendee.HasArrived = true
			attendee.ArrivalTime = &now

			if err := tx.Save(&attendee).Error; err != nil {
				return err
			}

			ticket.CheckedInAt = &now
			ticket.CheckedInBy = &req.CheckedInBy
			ticket.Status = models.TicketUsed

			return tx.Save(&ticket).Error
		})

		if err != nil {
			failures[ticketNumber] = err.Error()
		} else {
			successes = append(successes, ticketNumber)
		}
	}

	response := map[string]interface{}{
		"total":           len(req.TicketNumbers),
		"successes":       len(successes),
		"failures":        len(failures),
		"success_tickets": successes,
		"failed_tickets":  failures,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UndoCheckIn reverses a check-in
func (h *AttendeeHandler) UndoCheckIn(w http.ResponseWriter, r *http.Request) {
	ticketNumber := r.URL.Query().Get("ticket_number")
	if ticketNumber == "" {
		http.Error(w, "Ticket number is required", http.StatusBadRequest)
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var ticket models.Ticket
		if err := tx.Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
			return fmt.Errorf("ticket not found")
		}

		var attendee models.Attendee
		if err := tx.Where("ticket_id = ?", ticket.ID).First(&attendee).Error; err != nil {
			return fmt.Errorf("attendee not found")
		}

		if !attendee.HasArrived {
			return fmt.Errorf("attendee has not checked in")
		}

		attendee.HasArrived = false
		attendee.ArrivalTime = nil

		if err := tx.Save(&attendee).Error; err != nil {
			return err
		}

		ticket.CheckedInAt = nil
		ticket.CheckedInBy = nil
		ticket.Status = models.TicketActive

		return tx.Save(&ticket).Error
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Check-in reversed successfully"})
}
