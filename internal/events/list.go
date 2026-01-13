package events

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// ListEvents handles listing events with filtering and pagination
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	params := parseEventListParams(r)

	// Build the query
	query := h.db.Model(&models.Event{}).
		Where("status != ?", models.EventCancelled)

	// Apply filters
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

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.Search != nil {
		searchTerm := "%" + strings.ToLower(*params.Search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Only show published events for public listing (unless specific status filter is applied)
	if params.Status == nil {
		query = query.Where("status = ?", models.EventLive)
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
	case "popularity":
		if params.SortOrder == "desc" {
			query = query.Order("sales_volume DESC")
		} else {
			query = query.Order("sales_volume ASC")
		}
	case "created":
		if params.SortOrder == "desc" {
			query = query.Order("created_at DESC")
		} else {
			query = query.Order("created_at ASC")
		}
	default:
		query = query.Order("start_date ASC") // Default sorting
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

	response := EventListResponse{
		Events:     eventResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// ListOrganizerEvents handles listing events for a specific organizer
func (h *EventHandler) ListOrganizerEvents(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	params := parseEventListParams(r)

	// Build the query for organizer's events
	query := h.db.Model(&models.Event{}).
		Where("organizer_id = ?", organizer.ID)

	// Apply filters (same as public listing but include all statuses)
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

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.Search != nil {
		searchTerm := "%" + strings.ToLower(*params.Search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
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
		if params.SortOrder == "desc" {
			query = query.Order("status DESC")
		} else {
			query = query.Order("status ASC")
		}
	case "created":
		if params.SortOrder == "desc" {
			query = query.Order("created_at DESC")
		} else {
			query = query.Order("created_at ASC")
		}
	default:
		query = query.Order("created_at DESC") // Default: newest first
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

	response := EventListResponse{
		Events:     eventResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// parseEventListParams parses query parameters for event listing
func parseEventListParams(r *http.Request) EventListParams {
	params := EventListParams{
		Page:      1,
		Limit:     20,
		SortBy:    "date",
		SortOrder: "asc",
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			params.Limit = limit
		}
	}

	// Parse category
	if categoryStr := r.URL.Query().Get("category"); categoryStr != "" {
		category := models.EventCategory(categoryStr)
		params.Category = &category
	}

	// Parse location
	if location := r.URL.Query().Get("location"); location != "" {
		params.Location = &location
	}

	// Parse start date
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			params.StartDate = &startDate
		}
	}

	// Parse end date
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			params.EndDate = &endDate
		}
	}

	// Parse status
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.EventStatus(statusStr)
		params.Status = &status
	}

	// Parse search
	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = &search
	}

	// Parse sort by
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	// Parse sort order
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		params.SortOrder = sortOrder
	}

	return params
}

// convertToEventResponse converts a models.Event to EventResponse
func convertToEventResponse(event models.Event) EventResponse {
	// Convert organizer
	organizer := OrganizerSummary{
		ID:       event.Organizer.ID,
		Name:     event.Organizer.Name,
		About:    event.Organizer.About,
		LogoPath: event.Organizer.LogoPath,
	}

	// Convert venues
	venues := make([]VenueSummary, len(event.Venue))
	for i, venue := range event.Venue {
		venues[i] = VenueSummary{
			ID:            venue.ID,
			VenueName:     venue.VenueName,
			VenueType:     venue.VenueType,
			VenueLocation: venue.VenueLocation,
			Capacity:      venue.VenueCapacity,
		}
	}

	// Convert images
	images := make([]EventImageResponse, len(event.EventImages))
	for i, img := range event.EventImages {
		images[i] = EventImageResponse{
			ID:        img.ID,
			ImagePath: img.ImagePath,
		}
	}

	return EventResponse{
		ID:                      event.ID,
		Title:                   event.Title,
		Location:                event.Location,
		Description:             event.Description,
		StartDate:               event.StartDate,
		EndDate:                 event.EndDate,
		OnSaleDate:              event.OnSaleDate,
		Status:                  event.Status,
		Category:                event.Category,
		Currency:                event.Currency,
		MaxCapacity:             event.MaxCapacity,
		IsLive:                  event.IsLive,
		IsPrivate:               event.IsPrivate,
		MinAge:                  event.MinAge,
		LocationAddress:         event.LocationAddress,
		LocationCountry:         event.LocationCountry,
		BgType:                  event.BgType,
		BgColor:                 event.BgColor,
		TicketBorderColor:       event.TicketBorderColor,
		TicketBgColor:           event.TicketBgColor,
		TicketTextColor:         event.TicketTextColor,
		TicketSubTextColor:      event.TicketSubTextColor,
		BarcodeType:             event.BarcodeType,
		IsBarcodeEnabled:        event.IsBarcodeEnabled,
		EnableOfflinePayment:    event.EnableOfflinePayment,
		PreOrderMessageDisplay:  event.PreOrderMessageDisplay,
		PostOrderMessageDisplay: event.PostOrderMessageDisplay,
		Tags:                    event.Tags,
		SalesVolume:             event.SalesVolume,
		OrganizerFeesVolume:     event.OrganizerFeesVolume,
		OrganizerFeeFixed:       event.OrganizerFeeFixed,
		OrganizerFeePercentage:  event.OrganizerFeePercentage,
		Organizer:               organizer,
		Venues:                  venues,
		Images:                  images,
		CreatedAt:               event.CreatedAt,
		UpdatedAt:               event.UpdatedAt,
	}
}
