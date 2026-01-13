package events

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// SearchEvents handles dedicated event search with advanced filtering
func (h *EventHandler) SearchEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "search query (q) is required")
		return
	}

	// Parse other parameters (reuse existing parser)
	params := parseEventListParams(r)

	// Build the query
	query := h.db.Model(&models.Event{}).
		Where("status = ?", models.EventLive) // Only show live events for public search

	// Apply search across multiple fields
	searchTerm := "%" + strings.ToLower(searchQuery) + "%"
	query = query.Where(
		"LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ? OR LOWER(tags) LIKE ?",
		searchTerm, searchTerm, searchTerm, searchTerm,
	)

	// Apply additional filters
	if params.Category != nil {
		query = query.Where("category = ?", *params.Category)
	}

	if params.Location != nil {
		query = query.Where("LOWER(location) LIKE ?", "%"+strings.ToLower(*params.Location)+"%")
	}

	if params.StartDate != nil {
		query = query.Where("start_date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("end_date <= ?", *params.EndDate)
	}

	// Count total results
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to count events")
		return
	}

	// Apply sorting (default to relevance/popularity for search)
	switch params.SortBy {
	case "date":
		if params.SortOrder == "desc" {
			query = query.Order("start_date DESC")
		} else {
			query = query.Order("start_date ASC")
		}
	case "popularity":
		query = query.Order("sales_volume DESC")
	case "created":
		query = query.Order("created_at DESC")
	default:
		// Default for search: relevance (by popularity and recent)
		query = query.Order("sales_volume DESC, created_at DESC")
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload related data
	query = query.Preload("Organizer").
		Preload("Venue").
		Preload("EventImages")

	// Execute query
	var events []models.Event
	if err := query.Find(&events).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	// Convert to response format
	eventResponses := make([]EventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = convertToEventResponse(event)
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(params.Limit) - 1) / int64(params.Limit))

	response := EventSearchResponse{
		Query:      searchQuery,
		Events:     eventResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// SearchOrganizerEvents handles event search for organizers (includes all statuses)
func (h *EventHandler) SearchOrganizerEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can access this endpoint")
		return
	}

	// Get organizer profile
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "search query (q) is required")
		return
	}

	// Parse other parameters
	params := parseEventListParams(r)

	// Build the query for organizer's events
	query := h.db.Model(&models.Event{}).
		Where("organizer_id = ?", organizer.ID)

	// Apply search
	searchTerm := "%" + strings.ToLower(searchQuery) + "%"
	query = query.Where(
		"LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ? OR LOWER(tags) LIKE ?",
		searchTerm, searchTerm, searchTerm, searchTerm,
	)

	// Apply additional filters
	if params.Category != nil {
		query = query.Where("category = ?", *params.Category)
	}

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.Location != nil {
		query = query.Where("LOWER(location) LIKE ?", "%"+strings.ToLower(*params.Location)+"%")
	}

	if params.StartDate != nil {
		query = query.Where("start_date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("end_date <= ?", *params.EndDate)
	}

	// Count total results
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to count events")
		return
	}

	// Apply sorting
	switch params.SortBy {
	case "date":
		if params.SortOrder == "desc" {
			query = query.Order("start_date DESC")
		} else {
			query = query.Order("start_date ASC")
		}
	case "status":
		query = query.Order("status DESC, created_at DESC")
	case "created":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Preload related data
	query = query.Preload("Organizer").
		Preload("Venue").
		Preload("EventImages")

	// Execute query
	var events []models.Event
	if err := query.Find(&events).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	// Convert to response format
	eventResponses := make([]EventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = convertToEventResponse(event)
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(params.Limit) - 1) / int64(params.Limit))

	response := EventSearchResponse{
		Query:      searchQuery,
		Events:     eventResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// EventSearchResponse represents search results for events
type EventSearchResponse struct {
	Query      string          `json:"query"`
	Events     []EventResponse `json:"events"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
