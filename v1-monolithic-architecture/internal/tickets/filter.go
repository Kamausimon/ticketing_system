package tickets

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

// AdvancedTicketFilter represents advanced filtering options for organizers
type AdvancedTicketFilter struct {
	TicketFilter
	TicketClassID    *uint
	MinPrice         *float64
	MaxPrice         *float64
	IsCheckedIn      *bool
	CheckedInBefore  *time.Time
	CheckedInAfter   *time.Time
	TransferStatus   *string // "original", "transferred", "received"
	RefundStatus     *string // "refunded", "not_refunded"
	OrderStatus      *models.OrderStatus
	PaymentStatus    *models.PaymentStatus
	TicketClassNames []string // Filter by multiple ticket class names
}

// FilterEventTicketsAdvanced handles advanced ticket filtering for organizers
func (h *TicketHandler) FilterEventTicketsAdvanced(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Parse advanced filters
	filter := parseAdvancedTicketFilter(r)
	eventIDUint := uint(eventID)
	filter.EventID = &eventIDUint

	// Build query
	query := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("ticket_classes.event_id = ?", eventID)

	// Apply advanced filters
	query = applyAdvancedTicketFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get tickets with all necessary preloads
	var tickets []models.Ticket
	query = query.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("tickets.created_at DESC")

	if err := query.Find(&tickets).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	// Convert to response
	ticketResponses := make([]TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = convertToTicketResponse(ticket)
	}

	// Calculate total pages
	totalPages := int(totalCount) / filter.Limit
	if int(totalCount)%filter.Limit > 0 {
		totalPages++
	}

	// Calculate statistics
	stats := calculateTicketFilterStats(h.db, eventIDUint, filter)

	response := AdvancedTicketFilterResponse{
		Tickets:    ticketResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		Stats:      stats,
	}

	json.NewEncoder(w).Encode(response)
}

// SearchEventTickets handles searching tickets within an event
func (h *TicketHandler) SearchEventTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
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

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse filters
	filter := parseTicketFilter(r)
	eventIDUint := uint(eventID)
	filter.EventID = &eventIDUint
	filter.SearchTerm = searchQuery

	// Build query
	query := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("ticket_classes.event_id = ?", eventID)

	// Apply search
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query = query.Where(
		"LOWER(tickets.ticket_number) LIKE ? OR LOWER(tickets.holder_name) LIKE ? OR LOWER(tickets.holder_email) LIKE ? OR LOWER(orders.first_name) LIKE ? OR LOWER(orders.last_name) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// Apply other filters
	query = applyTicketFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get tickets
	var tickets []models.Ticket
	query = query.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("tickets.created_at DESC")

	if err := query.Find(&tickets).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	// Convert to response
	ticketResponses := make([]TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = convertToTicketResponse(ticket)
	}

	// Calculate total pages
	totalPages := int(totalCount) / filter.Limit
	if int(totalCount)%filter.Limit > 0 {
		totalPages++
	}

	response := TicketSearchResponse{
		Query:      searchQuery,
		Tickets:    ticketResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to parse advanced ticket filter from request
func parseAdvancedTicketFilter(r *http.Request) AdvancedTicketFilter {
	filter := AdvancedTicketFilter{
		TicketFilter: parseTicketFilter(r),
	}

	// Parse ticket class ID
	if ticketClassIDStr := r.URL.Query().Get("ticket_class_id"); ticketClassIDStr != "" {
		if id, err := strconv.ParseUint(ticketClassIDStr, 10, 64); err == nil {
			tcid := uint(id)
			filter.TicketClassID = &tcid
		}
	}

	// Parse min price
	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		if price, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = &price
		}
	}

	// Parse max price
	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		if price, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = &price
		}
	}

	// Parse is_checked_in
	if checkedInStr := r.URL.Query().Get("is_checked_in"); checkedInStr != "" {
		if checked, err := strconv.ParseBool(checkedInStr); err == nil {
			filter.IsCheckedIn = &checked
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

	// Parse transfer status
	if transferStatus := r.URL.Query().Get("transfer_status"); transferStatus != "" {
		filter.TransferStatus = &transferStatus
	}

	// Parse refund status
	if refundStatus := r.URL.Query().Get("refund_status"); refundStatus != "" {
		filter.RefundStatus = &refundStatus
	}

	// Parse order status
	if orderStatus := r.URL.Query().Get("order_status"); orderStatus != "" {
		os := models.OrderStatus(orderStatus)
		filter.OrderStatus = &os
	}

	// Parse payment status
	if paymentStatus := r.URL.Query().Get("payment_status"); paymentStatus != "" {
		ps := models.PaymentStatus(paymentStatus)
		filter.PaymentStatus = &ps
	}

	// Parse ticket class names
	if classNames := r.URL.Query().Get("ticket_class_names"); classNames != "" {
		filter.TicketClassNames = strings.Split(classNames, ",")
	}

	return filter
}

// Helper function to apply advanced ticket filters to query
func applyAdvancedTicketFilters(query *gorm.DB, filter AdvancedTicketFilter) *gorm.DB {
	// Apply base filters
	query = applyTicketFilters(query, filter.TicketFilter)

	// Apply ticket class ID filter
	if filter.TicketClassID != nil {
		query = query.Where("ticket_classes.id = ?", *filter.TicketClassID)
	}

	// Apply price filters
	if filter.MinPrice != nil {
		query = query.Where("order_items.unit_price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("order_items.unit_price <= ?", *filter.MaxPrice)
	}

	// Apply check-in filters
	if filter.IsCheckedIn != nil {
		if *filter.IsCheckedIn {
			query = query.Where("tickets.checked_in_at IS NOT NULL")
		} else {
			query = query.Where("tickets.checked_in_at IS NULL")
		}
	}

	if filter.CheckedInBefore != nil {
		query = query.Where("tickets.checked_in_at < ?", *filter.CheckedInBefore)
	}
	if filter.CheckedInAfter != nil {
		query = query.Where("tickets.checked_in_at > ?", *filter.CheckedInAfter)
	}

	// Apply transfer status filter
	if filter.TransferStatus != nil {
		switch *filter.TransferStatus {
		case "original":
			query = query.Where("tickets.original_holder_email = tickets.holder_email")
		case "transferred":
			query = query.Where("tickets.original_holder_email != tickets.holder_email AND tickets.transfer_count > 0")
		case "received":
			query = query.Where("tickets.original_holder_email != tickets.holder_email")
		}
	}

	// Apply refund status filter
	if filter.RefundStatus != nil {
		if *filter.RefundStatus == "refunded" {
			query = query.Where("tickets.status = ?", models.TicketRefunded)
		} else {
			query = query.Where("tickets.status != ?", models.TicketRefunded)
		}
	}

	// Apply order status filter
	if filter.OrderStatus != nil {
		query = query.Where("orders.status = ?", *filter.OrderStatus)
	}

	// Apply payment status filter
	if filter.PaymentStatus != nil {
		query = query.Where("orders.payment_status = ?", *filter.PaymentStatus)
	}

	// Apply ticket class names filter
	if len(filter.TicketClassNames) > 0 {
		query = query.Where("ticket_classes.name IN ?", filter.TicketClassNames)
	}

	return query
}

// Helper function to calculate statistics for filtered tickets
func calculateTicketFilterStats(db *gorm.DB, eventID uint, filter AdvancedTicketFilter) TicketFilterStats {
	var stats TicketFilterStats

	baseQuery := db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("ticket_classes.event_id = ?", eventID)

	// Apply the same filters to get stats for the filtered set
	baseQuery = applyAdvancedTicketFilters(baseQuery, filter)

	// Get counts by status
	var statusCounts []struct {
		Status models.TicketStatus
		Count  int64
	}
	baseQuery.Select("tickets.status, COUNT(*) as count").
		Group("tickets.status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		switch sc.Status {
		case models.TicketActive:
			stats.ActiveCount = sc.Count
		case models.TicketUsed:
			stats.UsedCount = sc.Count
		case models.TicketCancelled:
			stats.CancelledCount = sc.Count
		case models.TicketRefunded:
			stats.RefundedCount = sc.Count
		}
		stats.TotalCount += sc.Count
	}

	// Calculate check-in rate
	if stats.TotalCount > 0 {
		stats.CheckInRate = float64(stats.UsedCount) / float64(stats.TotalCount) * 100
	}

	// Get revenue stats
	db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("ticket_classes.event_id = ? AND tickets.status NOT IN ?", eventID, []models.TicketStatus{models.TicketCancelled, models.TicketRefunded}).
		Select("COALESCE(SUM(order_items.unit_price), 0)").
		Scan(&stats.TotalRevenue)

	return stats
}

// TicketFilterStats represents statistics for filtered tickets
type TicketFilterStats struct {
	TotalCount     int64   `json:"total_count"`
	ActiveCount    int64   `json:"active_count"`
	UsedCount      int64   `json:"used_count"`
	CancelledCount int64   `json:"cancelled_count"`
	RefundedCount  int64   `json:"refunded_count"`
	CheckInRate    float64 `json:"check_in_rate"`
	TotalRevenue   float64 `json:"total_revenue"`
}

// AdvancedTicketFilterResponse represents the response for advanced ticket filtering
type AdvancedTicketFilterResponse struct {
	Tickets    []TicketResponse  `json:"tickets"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
	Stats      TicketFilterStats `json:"stats"`
}

// TicketSearchResponse represents search results for tickets
type TicketSearchResponse struct {
	Query      string           `json:"query"`
	Tickets    []TicketResponse `json:"tickets"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}
