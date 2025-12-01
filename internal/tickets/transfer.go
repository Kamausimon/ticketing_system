package tickets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

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

	// Store original holder info for history
	originalHolderName := ticket.HolderName
	originalHolderEmail := ticket.HolderEmail

	// Update ticket holder
	ticket.HolderName = req.NewHolderName
	ticket.HolderEmail = req.NewHolderEmail

	// Begin transaction to save ticket and log transfer history
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to transfer ticket")
		return
	}

	// Log transfer in history
	transferHistory := models.TicketTransferHistory{
		TicketID:        uint(ticketID),
		FromHolderName:  originalHolderName,
		FromHolderEmail: originalHolderEmail,
		ToHolderName:    req.NewHolderName,
		ToHolderEmail:   req.NewHolderEmail,
		TransferredBy:   userID,
		TransferReason:  req.TransferReason,
		IPAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
	}

	if err := tx.Create(&transferHistory).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to log transfer history")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to complete transfer")
		return
	}

	// Log activity for audit
	activity := models.AccountActivity{
		AccountID:   user.AccountID,
		UserID:      &userID,
		Action:      models.ActionTicketTransferred,
		Category:    models.ActivityCategoryTicket,
		Description: fmt.Sprintf("Ticket %s transferred from %s to %s", ticket.TicketNumber, originalHolderEmail, req.NewHolderEmail),
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Success:     true,
		Severity:    models.SeverityInfo,
		Resource:    "ticket",
		Timestamp:   time.Now(),
	}
	h.db.Create(&activity)

	// In production, you would:
	// 1. Send email to new holder with ticket details
	// 2. Send confirmation to original holder

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

	// Fetch transfer history from database
	var transferHistory []models.TicketTransferHistory
	if err := h.db.Where("ticket_id = ?", ticketID).Order("transferred_at DESC").Find(&transferHistory).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch transfer history")
		return
	}

	// Build transfer history response
	type TransferRecord struct {
		ID              uint      `json:"id"`
		FromHolderName  string    `json:"from_holder_name"`
		FromHolderEmail string    `json:"from_holder_email"`
		ToHolderName    string    `json:"to_holder_name"`
		ToHolderEmail   string    `json:"to_holder_email"`
		TransferredAt   time.Time `json:"transferred_at"`
		TransferReason  string    `json:"transfer_reason,omitempty"`
	}

	transfers := make([]TransferRecord, len(transferHistory))
	for i, th := range transferHistory {
		transfers[i] = TransferRecord{
			ID:              th.ID,
			FromHolderName:  th.FromHolderName,
			FromHolderEmail: th.FromHolderEmail,
			ToHolderName:    th.ToHolderName,
			ToHolderEmail:   th.ToHolderEmail,
			TransferredAt:   th.TransferredAt,
			TransferReason:  th.TransferReason,
		}
	}

	response := map[string]interface{}{
		"ticket_number":    ticket.TicketNumber,
		"current_holder":   ticket.HolderName,
		"current_email":    ticket.HolderEmail,
		"transfer_count":   len(transfers),
		"transfer_history": transfers,
	}

	json.NewEncoder(w).Encode(response)
}
