package ticketclasses

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// ListTicketClasses handles listing all ticket classes for an event
func (h *TicketClassHandler) ListTicketClasses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Get all ticket classes for the event
	var ticketClasses []models.TicketClass
	query := h.db.Where("event_id = ?", eventID)

	// Apply filters
	if includeHidden := r.URL.Query().Get("include_hidden"); includeHidden != "true" {
		query = query.Where("is_hidden = ?", false)
	}

	if err := query.Order("sort_order ASC, created_at ASC").Find(&ticketClasses).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch ticket classes")
		return
	}

	// Convert to response
	var responses []TicketClassResponse
	for _, tc := range ticketClasses {
		responses = append(responses, h.convertToResponse(&tc))
	}

	response := TicketClassListResponse{
		TicketClasses: responses,
		Total:         int64(len(responses)),
		EventID:       uint(eventID),
		EventTitle:    event.Title,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTicketClass handles getting details of a specific ticket class
func (h *TicketClassHandler) GetTicketClass(w http.ResponseWriter, r *http.Request) {
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

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	response := h.convertToResponse(&ticketClass)
	json.NewEncoder(w).Encode(response)
}
