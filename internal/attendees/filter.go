package attendees

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// AdvancedAttendeeFilter represents advanced filtering options for attendees
type AdvancedAttendeeFilter struct {
	Page               int
	Limit              int
	EventID            *uint
	TicketClassID      *uint
	HasArrived         *bool
	IsRefunded         *bool
	CheckedInBefore    *time.Time
	CheckedInAfter     *time.Time
	RegistrationAfter  *time.Time
	RegistrationBefore *time.Time
	OrderStatus        *models.OrderStatus
	TicketStatus       *models.TicketStatus
	SearchTerm         string
	SortBy             string // "name", "email", "arrival_time", "registration_time"
	SortOrder          string // "asc", "desc"
}

// FilterAttendees handles advanced attendee filtering with statistics
func (h *AttendeeHandler) FilterAttendees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// Parse filters
	filter := parseAdvancedAttendeeFilter(r)

	// Build query
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass")

	// For organizers, only show attendees for their events
	if user.Role == models.RoleOrganizer {
		query = query.Joins("JOIN events ON events.id = attendees.event_id").
			Where("events.account_id = ?", user.AccountID)
	}

	// Apply filters
	query = applyAdvancedAttendeeFilters(query, filter)

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to count attendees")
		return
	}

	// Apply sorting
	switch filter.SortBy {
	case "name":
		if filter.SortOrder == "desc" {
			query = query.Order("last_name DESC, first_name DESC")
		} else {
			query = query.Order("last_name ASC, first_name ASC")
		}
	case "email":
		if filter.SortOrder == "desc" {
			query = query.Order("email DESC")
		} else {
			query = query.Order("email ASC")
		}
	case "arrival_time":
		if filter.SortOrder == "desc" {
			query = query.Order("arrival_time DESC NULLS LAST")
		} else {
			query = query.Order("arrival_time ASC NULLS LAST")
		}
	case "registration_time":
		if filter.SortOrder == "desc" {
			query = query.Order("created_at DESC")
		} else {
			query = query.Order("created_at ASC")
		}
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	query = query.Limit(filter.Limit).Offset(offset)

	// Get attendees
	var attendees []models.Attendee
	if err := query.Find(&attendees).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to fetch attendees")
		return
	}

	// Convert to response
	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	// Calculate total pages
	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	// Calculate statistics
	stats := calculateAttendeeStats(h.db, filter, user)

	response := AdvancedAttendeeFilterResponse{
		Attendees:  responses,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		Stats:      stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchAttendeesByEvent searches attendees within a specific event
func (h *AttendeeHandler) SearchAttendeesByEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get event ID
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	// Get search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "search query (q) is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify access to event
	if user.Role == models.RoleOrganizer {
		var event models.Event
		if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Parse filters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Build query
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ?", uint(eventID))

	// Apply search
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query = query.Where(
		"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone_number) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to count attendees")
		return
	}

	// Apply pagination
	offset := (page - 1) * limit
	query = query.Order("created_at DESC").Limit(limit).Offset(offset)

	// Get attendees
	var attendees []models.Attendee
	if err := query.Find(&attendees).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to fetch attendees")
		return
	}

	// Convert to response
	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := AttendeeSearchResponse{
		Query:      searchQuery,
		Attendees:  responses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to parse advanced attendee filter from request
func parseAdvancedAttendeeFilter(r *http.Request) AdvancedAttendeeFilter {
	filter := AdvancedAttendeeFilter{
		Page:      1,
		Limit:     50,
		SortBy:    "registration_time",
		SortOrder: "desc",
	}

	// Parse page
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	// Parse limit
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	// Parse event ID
	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		if eid, err := strconv.ParseUint(eventIDStr, 10, 32); err == nil {
			eid32 := uint(eid)
			filter.EventID = &eid32
		}
	}

	// Parse ticket class ID
	if ticketClassIDStr := r.URL.Query().Get("ticket_class_id"); ticketClassIDStr != "" {
		if tcid, err := strconv.ParseUint(ticketClassIDStr, 10, 32); err == nil {
			tcid32 := uint(tcid)
			filter.TicketClassID = &tcid32
		}
	}

	// Parse has_arrived
	if arrivedStr := r.URL.Query().Get("has_arrived"); arrivedStr != "" {
		if arrived, err := strconv.ParseBool(arrivedStr); err == nil {
			filter.HasArrived = &arrived
		}
	}

	// Parse is_refunded
	if refundedStr := r.URL.Query().Get("is_refunded"); refundedStr != "" {
		if refunded, err := strconv.ParseBool(refundedStr); err == nil {
			filter.IsRefunded = &refunded
		}
	}

	// Parse checked_in_before
	if beforeStr := r.URL.Query().Get("checked_in_before"); beforeStr != "" {
		if date, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			filter.CheckedInBefore = &date
		}
	}

	// Parse checked_in_after
	if afterStr := r.URL.Query().Get("checked_in_after"); afterStr != "" {
		if date, err := time.Parse(time.RFC3339, afterStr); err == nil {
			filter.CheckedInAfter = &date
		}
	}

	// Parse registration_after
	if afterStr := r.URL.Query().Get("registration_after"); afterStr != "" {
		if date, err := time.Parse(time.RFC3339, afterStr); err == nil {
			filter.RegistrationAfter = &date
		}
	}

	// Parse registration_before
	if beforeStr := r.URL.Query().Get("registration_before"); beforeStr != "" {
		if date, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			filter.RegistrationBefore = &date
		}
	}

	// Parse order status
	if orderStatus := r.URL.Query().Get("order_status"); orderStatus != "" {
		os := models.OrderStatus(orderStatus)
		filter.OrderStatus = &os
	}

	// Parse ticket status
	if ticketStatus := r.URL.Query().Get("ticket_status"); ticketStatus != "" {
		ts := models.TicketStatus(ticketStatus)
		filter.TicketStatus = &ts
	}

	// Parse search term
	if search := r.URL.Query().Get("search"); search != "" {
		filter.SearchTerm = strings.TrimSpace(search)
	}

	// Parse sort by
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	// Parse sort order
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	return filter
}

// Helper function to apply advanced attendee filters to query
func applyAdvancedAttendeeFilters(query *gorm.DB, filter AdvancedAttendeeFilter) *gorm.DB {
	// Apply event ID filter
	if filter.EventID != nil {
		query = query.Where("attendees.event_id = ?", *filter.EventID)
	}

	// Apply ticket class filter (requires join with tickets)
	if filter.TicketClassID != nil {
		query = query.Joins("JOIN tickets ON tickets.id = attendees.ticket_id").
			Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
			Where("order_items.ticket_class_id = ?", *filter.TicketClassID)
	}

	// Apply has_arrived filter
	if filter.HasArrived != nil {
		query = query.Where("attendees.has_arrived = ?", *filter.HasArrived)
	}

	// Apply is_refunded filter
	if filter.IsRefunded != nil {
		query = query.Where("attendees.is_refunded = ?", *filter.IsRefunded)
	}

	// Apply checked-in time filters
	if filter.CheckedInBefore != nil {
		query = query.Where("attendees.arrival_time < ?", *filter.CheckedInBefore)
	}
	if filter.CheckedInAfter != nil {
		query = query.Where("attendees.arrival_time > ?", *filter.CheckedInAfter)
	}

	// Apply registration time filters
	if filter.RegistrationBefore != nil {
		query = query.Where("attendees.created_at < ?", *filter.RegistrationBefore)
	}
	if filter.RegistrationAfter != nil {
		query = query.Where("attendees.created_at > ?", *filter.RegistrationAfter)
	}

	// Apply order status filter (requires join with tickets and orders)
	if filter.OrderStatus != nil {
		query = query.Joins("JOIN tickets ON tickets.id = attendees.ticket_id").
			Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.status = ?", *filter.OrderStatus)
	}

	// Apply ticket status filter (requires join with tickets)
	if filter.TicketStatus != nil {
		query = query.Joins("JOIN tickets ON tickets.id = attendees.ticket_id").
			Where("tickets.status = ?", *filter.TicketStatus)
	}

	// Apply search term
	if filter.SearchTerm != "" {
		searchPattern := "%" + strings.ToLower(filter.SearchTerm) + "%"
		query = query.Where(
			"LOWER(attendees.first_name) LIKE ? OR LOWER(attendees.last_name) LIKE ? OR LOWER(attendees.email) LIKE ? OR LOWER(attendees.phone_number) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	return query
}

// Helper function to calculate attendee statistics
func calculateAttendeeStats(db *gorm.DB, filter AdvancedAttendeeFilter, user models.User) AttendeeFilterStats {
	var stats AttendeeFilterStats

	baseQuery := db.Model(&models.Attendee{})

	// Apply organizer filter if needed
	if user.Role == models.RoleOrganizer {
		baseQuery = baseQuery.Joins("JOIN events ON events.id = attendees.event_id").
			Where("events.account_id = ?", user.AccountID)
	}

	// Apply filters
	baseQuery = applyAdvancedAttendeeFilters(baseQuery, filter)

	// Get total count
	baseQuery.Count(&stats.TotalCount)

	// Count arrived attendees
	db.Model(&models.Attendee{}).
		Where("has_arrived = ?", true).
		Count(&stats.ArrivedCount)

	// Count refunded attendees
	db.Model(&models.Attendee{}).
		Where("is_refunded = ?", true).
		Count(&stats.RefundedCount)

	// Calculate arrival rate
	if stats.TotalCount > 0 {
		stats.ArrivalRate = float64(stats.ArrivedCount) / float64(stats.TotalCount) * 100
	}

	return stats
}

// AttendeeFilterStats represents statistics for filtered attendees
type AttendeeFilterStats struct {
	TotalCount    int64   `json:"total_count"`
	ArrivedCount  int64   `json:"arrived_count"`
	RefundedCount int64   `json:"refunded_count"`
	ArrivalRate   float64 `json:"arrival_rate"`
}

// AdvancedAttendeeFilterResponse represents the response for advanced attendee filtering
type AdvancedAttendeeFilterResponse struct {
	Attendees  []AttendeeResponse  `json:"attendees"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
	Stats      AttendeeFilterStats `json:"stats"`
}

// AttendeeSearchResponse represents search results for attendees
type AttendeeSearchResponse struct {
	Query      string             `json:"query"`
	Attendees  []AttendeeResponse `json:"attendees"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}
