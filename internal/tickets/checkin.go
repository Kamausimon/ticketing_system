package tickets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// CheckInTicket handles checking in a ticket at the event
func (h *TicketHandler) CheckInTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req CheckInRequest
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
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ticket is for the correct event
	if ticket.OrderItem.TicketClass.EventID != req.EventID {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket is not for this event")
		return
	}

	// Check ticket status
	if ticket.Status == models.TicketUsed {
		response := map[string]interface{}{
			"status":        "already_checked_in",
			"message":       "This ticket has already been checked in",
			"checked_in_at": ticket.CheckedInAt,
		}
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	if ticket.Status == models.TicketCancelled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket has been cancelled")
		return
	}

	if ticket.Status == models.TicketRefunded {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket has been refunded")
		return
	}

	// Check if ticket is active
	if ticket.Status != models.TicketActive {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket is not active")
		return
	}

	// Check-in the ticket
	now := time.Now()
	ticket.Status = models.TicketUsed
	ticket.CheckedInAt = &now
	ticket.UsedAt = &now
	checkedInByUserID := userID
	ticket.CheckedInBy = &checkedInByUserID

	if err := h.db.Save(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to check in ticket")
		return
	}

	// Track metrics for check-in
	if h.metrics != nil {
		h.metrics.TicketsCheckedIn.WithLabelValues(
			fmt.Sprintf("%d", req.EventID),
		).Inc()
	}

	response := map[string]interface{}{
		"status":        "success",
		"message":       "Ticket checked in successfully",
		"ticket_number": ticket.TicketNumber,
		"holder_name":   ticket.HolderName,
		"holder_email":  ticket.HolderEmail,
		"checked_in_at": ticket.CheckedInAt,
		"event_title":   ticket.OrderItem.TicketClass.Event.Title,
	}

	json.NewEncoder(w).Encode(response)
}

// BulkCheckIn handles checking in multiple tickets at once
func (h *TicketHandler) BulkCheckIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		TicketNumbers []string `json:"ticket_numbers"`
		EventID       uint     `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.TicketNumbers) == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket_numbers is required")
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

	// Process each ticket
	successCount := 0
	var errors []map[string]interface{}

	for _, ticketNumber := range req.TicketNumbers {
		var ticket models.Ticket
		if err := h.db.Preload("OrderItem.TicketClass").
			Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
			errors = append(errors, map[string]interface{}{
				"ticket_number": ticketNumber,
				"error":         "ticket not found",
			})
			continue
		}

		// Verify ticket is for the correct event
		if ticket.OrderItem.TicketClass.EventID != req.EventID {
			errors = append(errors, map[string]interface{}{
				"ticket_number": ticketNumber,
				"error":         "ticket is not for this event",
			})
			continue
		}

		// Skip if already checked in
		if ticket.Status == models.TicketUsed {
			errors = append(errors, map[string]interface{}{
				"ticket_number": ticketNumber,
				"error":         "already checked in",
			})
			continue
		}

		// Skip invalid statuses
		if ticket.Status != models.TicketActive {
			errors = append(errors, map[string]interface{}{
				"ticket_number": ticketNumber,
				"error":         "ticket is not active",
			})
			continue
		}

		// Check-in the ticket
		now := time.Now()
		ticket.Status = models.TicketUsed
		ticket.CheckedInAt = &now
		ticket.UsedAt = &now
		checkedInByUserID := userID
		ticket.CheckedInBy = &checkedInByUserID

		if err := h.db.Save(&ticket).Error; err != nil {
			errors = append(errors, map[string]interface{}{
				"ticket_number": ticketNumber,
				"error":         "failed to check in",
			})
			continue
		}

		successCount++
	}

	response := map[string]interface{}{
		"message":       "Bulk check-in completed",
		"total":         len(req.TicketNumbers),
		"success_count": successCount,
		"error_count":   len(errors),
		"errors":        errors,
	}

	json.NewEncoder(w).Encode(response)
}

// GetCheckInStats handles getting check-in statistics for an event
func (h *TicketHandler) GetCheckInStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get event ID from URL
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	var stats CheckInStats

	// Total tickets for event
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("ticket_classes.event_id = ?", eventID).
		Count(&stats.TotalTickets)

	// Checked in tickets
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("ticket_classes.event_id = ? AND tickets.checked_in_at IS NOT NULL", eventID).
		Count(&stats.CheckedIn)

	// Not checked in
	stats.NotCheckedIn = stats.TotalTickets - stats.CheckedIn

	// Check-in rate
	if stats.TotalTickets > 0 {
		stats.CheckInRate = float64(stats.CheckedIn) / float64(stats.TotalTickets) * 100
	}

	// Get last check-in time
	var lastTicket models.Ticket
	if err := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("ticket_classes.event_id = ? AND tickets.checked_in_at IS NOT NULL", eventID).
		Order("tickets.checked_in_at DESC").
		First(&lastTicket).Error; err == nil {
		stats.LastCheckIn = lastTicket.CheckedInAt
	}

	json.NewEncoder(w).Encode(stats)
}

// UndoCheckIn handles undoing a check-in (for mistakes)
func (h *TicketHandler) UndoCheckIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket number from request
	ticketNumber := r.URL.Query().Get("ticket_number")
	if ticketNumber == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket_number is required")
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
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify user is organizer of the event
	if ticket.OrderItem.TicketClass.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if ticket is checked in
	if ticket.Status != models.TicketUsed {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket is not checked in")
		return
	}

	// Undo check-in
	ticket.Status = models.TicketActive
	ticket.CheckedInAt = nil
	ticket.UsedAt = nil
	ticket.CheckedInBy = nil

	if err := h.db.Save(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to undo check-in")
		return
	}

	response := map[string]interface{}{
		"message":       "Check-in undone successfully",
		"ticket_number": ticket.TicketNumber,
		"status":        ticket.Status,
	}

	json.NewEncoder(w).Encode(response)
}
