package attendees

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ticketing_system/internal/models"
)

// ListAttendees lists all attendees with filtering
func (h *AttendeeHandler) ListAttendees(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass")

	// Parse event ID filter
	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		if eid, err := strconv.ParseUint(eventIDStr, 10, 32); err == nil {
			query = query.Where("event_id = ?", uint(eid))
		}
	}

	// Parse has_arrived filter
	if arrivedStr := r.URL.Query().Get("has_arrived"); arrivedStr != "" {
		if arrived, err := strconv.ParseBool(arrivedStr); err == nil {
			query = query.Where("has_arrived = ?", arrived)
		}
	}

	// Parse is_refunded filter
	if refundedStr := r.URL.Query().Get("is_refunded"); refundedStr != "" {
		if refunded, err := strconv.ParseBool(refundedStr); err == nil {
			query = query.Where("is_refunded = ?", refunded)
		}
	}

	// Parse search term
	if searchTerm := r.URL.Query().Get("search"); searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		http.Error(w, "Failed to count attendees", http.StatusInternalServerError)
		return
	}

	// Get attendees
	var attendees []models.Attendee
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
		return
	}

	// Convert to response
	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := AttendeeListResponse{
		Attendees:  responses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListEventAttendees lists all attendees for a specific event
func (h *AttendeeHandler) ListEventAttendees(w http.ResponseWriter, r *http.Request) {
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

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ?", uint(eventID))

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		http.Error(w, "Failed to count attendees", http.StatusInternalServerError)
		return
	}

	// Get attendees
	var attendees []models.Attendee
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
		return
	}

	// Convert to response
	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := AttendeeListResponse{
		Attendees:  responses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchAttendees searches for attendees by name or email
func (h *AttendeeHandler) SearchAttendees(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		http.Error(w, "Search term is required", http.StatusBadRequest)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")

	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket")

	if eventIDStr != "" {
		if eid, err := strconv.ParseUint(eventIDStr, 10, 32); err == nil {
			query = query.Where("event_id = ?", uint(eid))
		}
	}

	searchPattern := "%" + searchTerm + "%"
	query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
		searchPattern, searchPattern, searchPattern)

	var attendees []models.Attendee
	if err := query.Limit(20).Find(&attendees).Error; err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetAttendeeCount returns the count of attendees for an event
func (h *AttendeeHandler) GetAttendeeCount(w http.ResponseWriter, r *http.Request) {
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

	var total int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ?", uint(eventID)).
		Count(&total).Error; err != nil {
		http.Error(w, "Failed to count attendees", http.StatusInternalServerError)
		return
	}

	var checkedIn int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ? AND has_arrived = ?", uint(eventID), true).
		Count(&checkedIn).Error; err != nil {
		http.Error(w, "Failed to count checked-in attendees", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_attendees": total,
		"checked_in":      checkedIn,
		"not_checked_in":  total - checkedIn,
		"check_in_rate":   float64(checkedIn) / float64(total) * 100,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
