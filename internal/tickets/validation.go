package tickets

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// ValidateTicket handles validating a ticket at entry
func (h *TicketHandler) ValidateTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req ValidateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TicketNumber == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket_number is required")
		return
	}

	if req.EventID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user is organizer of the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", req.EventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied: you are not the organizer of this event")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Where("ticket_number = ?", req.TicketNumber).First(&ticket).Error; err != nil {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Ticket not found",
			"reason":  "invalid_ticket",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Verify ticket is for the correct event
	if ticket.OrderItem.TicketClass.EventID != req.EventID {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Ticket is not for this event",
			"reason":  "wrong_event",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check ticket status
	if ticket.Status == models.TicketUsed {
		response := map[string]interface{}{
			"valid":         false,
			"message":       "Ticket has already been used",
			"reason":        "already_used",
			"checked_in_at": ticket.CheckedInAt,
		}
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	if ticket.Status == models.TicketCancelled {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Ticket has been cancelled",
			"reason":  "cancelled",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if ticket.Status == models.TicketRefunded {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Ticket has been refunded",
			"reason":  "refunded",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Ticket is valid
	response := map[string]interface{}{
		"valid":         true,
		"message":       "Ticket is valid",
		"ticket_number": ticket.TicketNumber,
		"holder_name":   ticket.HolderName,
		"holder_email":  ticket.HolderEmail,
		"event_title":   ticket.OrderItem.TicketClass.Event.Title,
		"event_date":    ticket.OrderItem.TicketClass.Event.StartDate,
		"ticket_class":  ticket.OrderItem.TicketClass.Name,
		"status":        ticket.Status,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidateTicketByQR handles validating a ticket by scanning QR code
func (h *TicketHandler) ValidateTicketByQR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		QRCode  string `json:"qr_code"`
		EventID uint   `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.QRCode == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "qr_code is required")
		return
	}

	if req.EventID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user is organizer of the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", req.EventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get ticket by QR code
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Where("qr_code = ?", req.QRCode).First(&ticket).Error; err != nil {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Invalid QR code",
			"reason":  "invalid_qr",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Verify ticket is for the correct event
	if ticket.OrderItem.TicketClass.EventID != req.EventID {
		response := map[string]interface{}{
			"valid":   false,
			"message": "QR code is not for this event",
			"reason":  "wrong_event",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Use the same validation logic as ValidateTicket
	if ticket.Status == models.TicketUsed {
		response := map[string]interface{}{
			"valid":         false,
			"message":       "Ticket has already been used",
			"reason":        "already_used",
			"checked_in_at": ticket.CheckedInAt,
		}
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	if ticket.Status != models.TicketActive {
		response := map[string]interface{}{
			"valid":   false,
			"message": "Ticket is not active",
			"reason":  "invalid_status",
			"status":  ticket.Status,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Ticket is valid
	response := map[string]interface{}{
		"valid":         true,
		"message":       "Ticket is valid",
		"ticket_number": ticket.TicketNumber,
		"holder_name":   ticket.HolderName,
		"holder_email":  ticket.HolderEmail,
		"event_title":   ticket.OrderItem.TicketClass.Event.Title,
		"ticket_class":  ticket.OrderItem.TicketClass.Name,
		"status":        ticket.Status,
	}

	json.NewEncoder(w).Encode(response)
}
