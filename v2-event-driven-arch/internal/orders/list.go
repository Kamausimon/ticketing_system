package orders

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

// ListOrders handles listing orders with filtering and pagination
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
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

	// Parse filter parameters
	filter := parseOrderFilter(r)

	// Build query
	query := h.db.Model(&models.Order{}).Preload("Event").Preload("OrderItems.TicketClass")

	// Filter by account (users see their own orders)
	if user.Role != models.RoleAdmin {
		query = query.Where("account_id = ?", user.AccountID)
	}

	// Apply filters
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
	if filter.SearchTerm != "" {
		search := "%" + strings.ToLower(filter.SearchTerm) + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			search, search, search)
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

	response := OrderListResponse{
		Orders:     orderResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// ListOrganizerOrders handles listing orders for an organizer's events
func (h *OrderHandler) ListOrganizerOrders(w http.ResponseWriter, r *http.Request) {
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

	// Parse filter parameters
	filter := parseOrderFilter(r)

	// Build query - join with events to filter by organizer
	query := h.db.Model(&models.Order{}).
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.account_id = ?", user.AccountID).
		Preload("Event").
		Preload("OrderItems.TicketClass")

	// Apply filters
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
	if filter.SearchTerm != "" {
		search := "%" + strings.ToLower(filter.SearchTerm) + "%"
		query = query.Where("LOWER(orders.email) LIKE ? OR LOWER(orders.first_name) LIKE ? OR LOWER(orders.last_name) LIKE ?",
			search, search, search)
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

	response := OrderListResponse{
		Orders:     orderResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// GetOrderStats handles getting order statistics
func (h *OrderHandler) GetOrderStats(w http.ResponseWriter, r *http.Request) {
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

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	stats := OrderStats{}

	// Build base query with account filter for non-admins
	buildQuery := func() *gorm.DB {
		query := h.db.Model(&models.Order{})
		if user.Role != models.RoleAdmin {
			query = query.Where("account_id = ?", user.AccountID)
		}
		return query
	}

	// Get total orders
	buildQuery().Count(&stats.TotalOrders)

	// Get total revenue
	var revenueResult struct {
		Total float64
	}
	buildQuery().
		Select("COALESCE(SUM(amount), 0) as total").
		Where("status = ? OR status = ?", models.OrderPaid, models.OrderFulfilled).
		Scan(&revenueResult)
	stats.TotalRevenue = revenueResult.Total

	// Count by status - use fresh queries each time
	buildQuery().Where("status = ?", models.OrderPending).Count(&stats.PendingOrders)
	buildQuery().Where("status = ? OR status = ?", models.OrderPaid, models.OrderFulfilled).Count(&stats.CompletedOrders)
	buildQuery().Where("status = ?", models.OrderCancelled).Count(&stats.CancelledOrders)
	buildQuery().Where("status = ?", models.OrderRefunded).Count(&stats.RefundedOrders)

	// Calculate average order value
	if stats.CompletedOrders > 0 {
		stats.AverageOrder = stats.TotalRevenue / float64(stats.CompletedOrders)
	}

	json.NewEncoder(w).Encode(stats)
}

// parseOrderFilter parses query parameters into OrderFilter
func parseOrderFilter(r *http.Request) OrderFilter {
	filter := OrderFilter{
		Page:  1,
		Limit: 20,
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	// Parse status
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.OrderStatus(statusStr)
		filter.Status = &status
	}

	// Parse payment status
	if paymentStatusStr := r.URL.Query().Get("payment_status"); paymentStatusStr != "" {
		paymentStatus := models.PaymentStatus(paymentStatusStr)
		filter.PaymentStatus = &paymentStatus
	}

	// Parse event ID
	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		if eventID, err := strconv.ParseUint(eventIDStr, 10, 32); err == nil {
			id := uint(eventID)
			filter.EventID = &id
		}
	}

	// Parse start date
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	// Parse end date
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse search term
	filter.SearchTerm = r.URL.Query().Get("search")

	// Parse email filter
	filter.Email = r.URL.Query().Get("email")

	return filter
}
