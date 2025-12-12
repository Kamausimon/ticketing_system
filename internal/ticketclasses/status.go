package ticketclasses

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// PauseTicketClass handles pausing ticket sales for a ticket class
func (h *TicketClassHandler) PauseTicketClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket class ID")
		return
	}

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and verify ownership
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event belongs to user
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get ticket class
	var ticketClass models.TicketClass
	if err := h.db.Where("id = ? AND event_id = ?", ticketClassID, eventID).First(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket class not found")
		return
	}

	if ticketClass.IsPaused {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket class is already paused")
		return
	}

	// Pause ticket class
	if err := h.db.Model(&ticketClass).Update("is_paused", true).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to pause ticket class")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Ticket class sales paused successfully",
	})
}

// ResumeTicketClass handles resuming ticket sales for a ticket class
func (h *TicketClassHandler) ResumeTicketClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket class ID")
		return
	}

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and verify ownership
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event belongs to user
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get ticket class
	var ticketClass models.TicketClass
	if err := h.db.Where("id = ? AND event_id = ?", ticketClassID, eventID).First(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket class not found")
		return
	}

	if !ticketClass.IsPaused {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket class is not paused")
		return
	}

	// Resume ticket class
	if err := h.db.Model(&ticketClass).Update("is_paused", false).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to resume ticket class")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Ticket class sales resumed successfully",
	})
}
