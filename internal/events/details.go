package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetEventDetails handles getting detailed event information
func (h *EventHandler) GetEventDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Get event with all related data
	var event models.Event
	if err := h.db.Preload("Organizer").
		Preload("Venue").
		Preload("EventImages").
		Where("id = ?", eventID).
		First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found")
		return
	}

	// Track event view metrics
	if h.metrics != nil {
		h.metrics.TrackEventView(fmt.Sprintf("%d", eventID))
	}

	// Check if event is accessible (live events are public, others require organizer access)
	if event.Status != models.EventLive {
		// Check if user is the organizer
		userID := middleware.GetUserIDFromToken(r)
		if userID == 0 {
			middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Get user
		var user models.User
		if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
			return
		}

		// Get organizer
		var organizer models.Organizer
		if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}

		// Check if user owns this event
		if event.OrganizerID != organizer.ID {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Convert to response format
	response := convertToEventResponse(event)

	json.NewEncoder(w).Encode(response)
}

// UpdateEvent handles event updates by organizers
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can update events")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get event
	var event models.Event
	if err := h.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or access denied")
		return
	}

	// Parse update request
	var req CreateEventRequest // Reuse the same structure
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateEventRequest(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update event fields
	event.Title = strings.TrimSpace(req.Title)
	event.Location = strings.TrimSpace(req.Location)
	event.Description = strings.TrimSpace(req.Description)
	event.StartDate = req.StartDate
	event.EndDate = req.EndDate
	event.OnSaleDate = req.OnSaleDate
	event.Category = req.Category
	event.Currency = req.Currency
	event.MaxCapacity = req.MaxCapacity
	event.IsPrivate = req.IsPrivate
	event.MinAge = req.MinAge
	event.LocationAddress = req.LocationAddress
	event.LocationAddressLine = req.LocationAddressLine
	event.LocationCountry = req.LocationCountry
	event.BgType = req.BgType
	event.BgColor = req.BgColor
	event.TicketBorderColor = req.TicketBorderColor
	event.TicketBgColor = req.TicketBgColor
	event.TicketTextColor = req.TicketTextColor
	event.TicketSubTextColor = req.TicketSubTextColor
	event.BarcodeType = req.BarcodeType
	event.IsBarcodeEnabled = req.IsBarcodeEnabled
	event.EnableOfflinePayment = req.EnableOfflinePayment
	event.PreOrderMessageDisplay = req.PreOrderMessageDisplay
	event.PostOrderMessageDisplay = req.PostOrderMessageDisplay
	event.Tags = req.Tags
	event.OrganizerFeeFixed = req.OrganizerFeeFixed
	event.OrganizerFeePercentage = req.OrganizerFeePercentage

	// Save event
	if err := h.db.Save(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update event")
		return
	}

	// Update venue associations if provided
	if len(req.VenueIDs) > 0 {
		// Clear existing venue associations
		h.db.Where("event_id = ?", event.ID).Delete(&models.EventVenues{})

		// Add new venue associations
		for _, venueID := range req.VenueIDs {
			// Verify venue exists
			var venue models.Venue
			if err := h.db.Where("id = ?", venueID).First(&venue).Error; err != nil {
				continue // Skip invalid venue IDs
			}

			// Create venue association
			eventVenue := models.EventVenues{
				EventID:   event.ID,
				VenueID:   venueID,
				VenueRole: "primary",
			}
			h.db.Create(&eventVenue)
		}
	}

	response := map[string]interface{}{
		"message":  "Event updated successfully",
		"event_id": event.ID,
		"status":   string(event.Status),
	}

	json.NewEncoder(w).Encode(response)
}

// PublishEvent handles publishing an event (changing status from draft to live)
func (h *EventHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can publish events")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get event
	var event models.Event
	if err := h.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or access denied")
		return
	}

	// Check if event can be published
	if event.Status != models.EventDraft {
		middleware.WriteJSONError(w, http.StatusBadRequest, "only draft events can be published")
		return
	}

	// Update status to live
	event.Status = models.EventLive
	event.IsLive = true

	if err := h.db.Save(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to publish event")
		return
	}

	// Track metrics for event publishing
	if h.metrics != nil {
		h.metrics.EventsPublished.Inc()
	}

	response := map[string]interface{}{
		"message":  "Event published successfully",
		"event_id": event.ID,
		"status":   string(event.Status),
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteEvent handles event deletion by organizers
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can delete events")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get event
	var event models.Event
	if err := h.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or access denied")
		return
	}

	// Check if event can be deleted (only draft events should be completely deleted)
	if event.Status == models.EventDraft {
		// Hard delete for draft events
		if err := h.db.Unscoped().Delete(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to delete event")
			return
		}
	} else {
		// Soft delete for published events (change status to cancelled)
		event.Status = models.EventCancelled
		event.IsLive = false
		if err := h.db.Save(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to cancel event")
			return
		}
	}

	response := map[string]interface{}{
		"message":  "Event deleted successfully",
		"event_id": event.ID,
	}

	json.NewEncoder(w).Encode(response)
}
