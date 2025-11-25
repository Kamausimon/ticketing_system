package attendees

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetAttendeeDetails retrieves detailed information about a specific attendee
func (h *AttendeeHandler) GetAttendeeDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attendeeID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid attendee ID", http.StatusBadRequest)
		return
	}

	var attendee models.Attendee
	if err := h.db.Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Preload("Order").
		Preload("Account").
		First(&attendee, uint(attendeeID)).Error; err != nil {
		http.Error(w, "Attendee not found", http.StatusNotFound)
		return
	}

	response := convertToAttendeeResponse(attendee)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAttendeeByTicket retrieves attendee by ticket number
func (h *AttendeeHandler) GetAttendeeByTicket(w http.ResponseWriter, r *http.Request) {
	ticketNumber := r.URL.Query().Get("ticket_number")
	if ticketNumber == "" {
		http.Error(w, "Ticket number is required", http.StatusBadRequest)
		return
	}

	var ticket models.Ticket
	if err := h.db.Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	var attendee models.Attendee
	if err := h.db.Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("ticket_id = ?", ticket.ID).
		First(&attendee).Error; err != nil {
		http.Error(w, "Attendee not found", http.StatusNotFound)
		return
	}

	response := convertToAttendeeResponse(attendee)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAttendeesByOrder retrieves all attendees for a specific order
func (h *AttendeeHandler) GetAttendeesByOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var attendees []models.Attendee
	if err := h.db.Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("order_id = ?", uint(orderID)).
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
		return
	}

	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
