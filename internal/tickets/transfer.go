package tickets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// TransferTicket handles transferring a ticket to another person
func (h *TicketHandler) TransferTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Parse request
	var req TransferTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewHolderName == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_holder_name is required")
		return
	}

	if req.NewHolderEmail == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_holder_email is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check ticket status
	if ticket.Status != models.TicketActive {
		middleware.WriteJSONError(w, http.StatusBadRequest, "only active tickets can be transferred")
		return
	}

	// Check if event allows transfers (in production, add this to Event model)
	// For now, we'll allow all transfers

	// Update ticket holder
	ticket.HolderName = req.NewHolderName
	ticket.HolderEmail = req.NewHolderEmail

	if err := h.db.Save(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to transfer ticket")
		return
	}

	// In production, you would:
	// 1. Send email to new holder with ticket details
	// 2. Send confirmation to original holder
	// 3. Log the transfer for audit purposes

	response := map[string]interface{}{
		"message":          "Ticket transferred successfully",
		"ticket_number":    ticket.TicketNumber,
		"new_holder_name":  ticket.HolderName,
		"new_holder_email": ticket.HolderEmail,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTransferHistory handles getting the transfer history of a ticket
func (h *TicketHandler) GetTransferHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership or organizer access
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		// Check if user is the event organizer
		var event models.Event
		h.db.Preload("OrderItem.TicketClass").First(&ticket)
		if err := h.db.Where("id = ? AND account_id = ?",
			ticket.OrderItem.TicketClass.EventID, user.AccountID).First(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// In production, you would have a ticket_transfer_history table
	// For now, return mock data or current holder info
	response := map[string]interface{}{
		"message":        "Transfer history not yet implemented",
		"ticket_number":  ticket.TicketNumber,
		"current_holder": ticket.HolderName,
		"current_email":  ticket.HolderEmail,
		"transfers":      []interface{}{}, // Would contain transfer history
	}

	json.NewEncoder(w).Encode(response)
}
