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

// ListUserTickets handles listing all tickets for a user
func (h *TicketHandler) ListUserTickets(w http.ResponseWriter, r *http.Request) {
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
	filter := parseTicketFilter(r)

	// Build query
	query := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ?", user.AccountID)

	// Apply filters
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

	response := TicketListResponse{
		Tickets:    ticketResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// ListEventTickets handles listing all tickets for an event (organizer view)
func (h *TicketHandler) ListEventTickets(w http.ResponseWriter, r *http.Request) {
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

	// Parse filters
	filter := parseTicketFilter(r)
	eventIDUint := uint(eventID)
	filter.EventID = &eventIDUint

	// Build query
	query := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("ticket_classes.event_id = ?", eventID)

	// Apply filters
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

	response := TicketListResponse{
		Tickets:    ticketResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTicketStats handles getting ticket statistics
func (h *TicketHandler) GetTicketStats(w http.ResponseWriter, r *http.Request) {
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

	// Base query for user's tickets
	baseQuery := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ?", user.AccountID)

	var stats TicketStats

	// Total tickets
	baseQuery.Count(&stats.TotalTickets)

	// Active tickets
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ? AND tickets.status = ?", user.AccountID, models.TicketActive).
		Count(&stats.ActiveTickets)

	// Used tickets
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ? AND tickets.status = ?", user.AccountID, models.TicketUsed).
		Count(&stats.UsedTickets)

	// Cancelled tickets
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ? AND tickets.status = ?", user.AccountID, models.TicketCancelled).
		Count(&stats.CancelledTickets)

	// Refunded tickets
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.account_id = ? AND tickets.status = ?", user.AccountID, models.TicketRefunded).
		Count(&stats.RefundedTickets)

	// Calculate check-in rate
	if stats.TotalTickets > 0 {
		stats.CheckInRate = float64(stats.UsedTickets) / float64(stats.TotalTickets) * 100
	}

	json.NewEncoder(w).Encode(stats)
}

// Helper function to parse ticket filter from request
func parseTicketFilter(r *http.Request) TicketFilter {
	filter := TicketFilter{
		Page:  1,
		Limit: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := models.TicketStatus(status)
		filter.Status = &s
	}

	if eventID := r.URL.Query().Get("event_id"); eventID != "" {
		if id, err := strconv.ParseUint(eventID, 10, 64); err == nil {
			eid := uint(id)
			filter.EventID = &eid
		}
	}

	if orderID := r.URL.Query().Get("order_id"); orderID != "" {
		if id, err := strconv.ParseUint(orderID, 10, 64); err == nil {
			oid := uint(id)
			filter.OrderID = &oid
		}
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = &date
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.SearchTerm = strings.TrimSpace(search)
	}

	return filter
}

// Helper function to apply ticket filters to query
func applyTicketFilters(query *gorm.DB, filter TicketFilter) *gorm.DB {
	if filter.Status != nil {
		query = query.Where("tickets.status = ?", *filter.Status)
	}

	if filter.EventID != nil {
		query = query.Where("ticket_classes.event_id = ?", *filter.EventID)
	}

	if filter.OrderID != nil {
		query = query.Where("orders.id = ?", *filter.OrderID)
	}

	if filter.StartDate != nil {
		query = query.Where("tickets.created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("tickets.created_at <= ?", *filter.EndDate)
	}

	if filter.SearchTerm != "" {
		searchPattern := "%" + filter.SearchTerm + "%"
		query = query.Where(
			"tickets.ticket_number ILIKE ? OR tickets.holder_name ILIKE ? OR tickets.holder_email ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	return query
}
