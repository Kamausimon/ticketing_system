package orders

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// SearchOrders handles searching orders with advanced filtering
func (h *OrderHandler) SearchOrders(w http.ResponseWriter, r *http.Request) {
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

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get search query (optional - can filter without searching)
	searchQuery := r.URL.Query().Get("q")

	// Parse filter parameters
	filter := parseOrderFilter(r)

	// Build query
	query := h.db.Model(&models.Order{}).Preload("Event").Preload("OrderItems.TicketClass")

	// Filter by account (users see their own orders)
	if user.Role != models.RoleAdmin {
		query = query.Where("account_id = ?", user.AccountID)
	}

	// Apply search if query provided and not wildcard
	if searchQuery != "" && searchQuery != "*" {
		search := "%" + strings.ToLower(searchQuery) + "%"
		query = query.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR CAST(id AS TEXT) LIKE ?",
			search, search, search, search,
		)
	}

	// Apply email filter (exact match)
	if filter.Email != "" {
		query = query.Where("LOWER(email) = ?", strings.ToLower(filter.Email))
	}

	// Apply other filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.PaymentStatus != nil {
		query = query.Where("payment_status = ?", *filter.PaymentStatus)
	}
	if filter.EventID != nil {
		query = query.Where("event_id = ?", *filter.EventID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit).Order("created_at DESC")

	// Execute query
	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	// Convert to response
	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = convertToOrderResponse(order)
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(filter.Limit) - 1) / int64(filter.Limit))

	response := OrderSearchResponse{
		Query:      searchQuery,
		Orders:     orderResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// SearchOrganizerOrders handles searching orders for organizer's events
func (h *OrderHandler) SearchOrganizerOrders(w http.ResponseWriter, r *http.Request) {
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

	// Get user to verify organizer role
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer && user.Role != models.RoleAdmin {
		middleware.WriteJSONError(w, http.StatusForbidden, "organizer access required")
		return
	}

	// Get search query (optional - can filter without searching)
	searchQuery := r.URL.Query().Get("q")

	// Parse filter parameters
	filter := parseOrderFilter(r)

	// Build query - join with events to filter by organizer
	query := h.db.Model(&models.Order{}).
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.account_id = ?", user.AccountID).
		Preload("Event").
		Preload("OrderItems.TicketClass")

	// Apply search if query provided and not wildcard
	if searchQuery != "" && searchQuery != "*" {
		search := "%" + strings.ToLower(searchQuery) + "%"
		query = query.Where(
			"LOWER(orders.email) LIKE ? OR LOWER(orders.first_name) LIKE ? OR LOWER(orders.last_name) LIKE ? OR CAST(orders.id AS TEXT) LIKE ? OR LOWER(events.title) LIKE ?",
			search, search, search, search, search,
		)
	}

	// Apply email filter (exact match)
	if filter.Email != "" {
		query = query.Where("LOWER(orders.email) = ?", strings.ToLower(filter.Email))
	}

	// Apply other filters
	if filter.Status != nil {
		query = query.Where("orders.status = ?", *filter.Status)
	}
	if filter.PaymentStatus != nil {
		query = query.Where("orders.payment_status = ?", *filter.PaymentStatus)
	}
	if filter.EventID != nil {
		query = query.Where("orders.event_id = ?", *filter.EventID)
	}
	if filter.StartDate != nil {
		query = query.Where("orders.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("orders.created_at <= ?", *filter.EndDate)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit).Order("orders.created_at DESC")

	// Execute query
	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	// Convert to response
	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = convertToOrderResponse(order)
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(filter.Limit) - 1) / int64(filter.Limit))

	response := OrderSearchResponse{
		Query:      searchQuery,
		Orders:     orderResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// OrderSearchResponse represents search results for orders
type OrderSearchResponse struct {
	Query      string          `json:"query"`
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
