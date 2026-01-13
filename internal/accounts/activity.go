package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// GetAccountActivity handles getting user's account activity
func (h *AccountHandler) GetAccountActivity(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	filter := parseActivityFilter(r)

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Build query
	query := h.db.Model(&models.AccountActivity{}).Where("account_id = ?", user.AccountID)

	// Apply filters
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}
	if filter.Success != nil {
		query = query.Where("success = ?", *filter.Success)
	}
	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", *filter.EndDate)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to count activities")
		return
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	var activities []models.AccountActivity
	if err := query.Order("timestamp DESC").Offset(offset).Limit(filter.Limit).Find(&activities).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch activities")
		return
	}

	totalPages := int((totalCount + int64(filter.Limit) - 1) / int64(filter.Limit))

	// Convert to response format
	responseActivities := make([]map[string]interface{}, len(activities))
	for i, activity := range activities {
		responseActivities[i] = map[string]interface{}{
			"id":          activity.ID,
			"account_id":  activity.AccountID,
			"user_id":     activity.UserID,
			"action":      activity.Action,
			"category":    activity.Category,
			"description": activity.Description,
			"ip_address":  activity.IPAddress,
			"user_agent":  activity.UserAgent,
			"success":     activity.Success,
			"severity":    activity.Severity,
			"resource":    activity.Resource,
			"resource_id": activity.ResourceID,
			"timestamp":   activity.Timestamp,
			"created_at":  activity.CreatedAt,
		}
	}

	response := PaginatedResponse{
		Data:       responseActivities,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// logAccountActivity logs an activity for an account (helper function)
func (h *AccountHandler) logAccountActivity(accountID uint, action, description, ipAddress string) {
	activity := models.AccountActivity{
		AccountID:   accountID,
		Action:      action,
		Category:    getCategoryFromAction(action),
		Description: description,
		IPAddress:   ipAddress,
		Success:     true,
		Severity:    models.SeverityInfo,
		Timestamp:   time.Now(),
	}
	h.db.Create(&activity)
}

// LogActivityWithDetails logs an activity with full details
func (h *AccountHandler) LogActivityWithDetails(accountID uint, userID *uint, action, category, description, ipAddress, userAgent string, success bool, severity string) {
	activity := models.AccountActivity{
		AccountID:   accountID,
		UserID:      userID,
		Action:      action,
		Category:    category,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Success:     success,
		Severity:    severity,
		Timestamp:   time.Now(),
	}
	h.db.Create(&activity)
}

// getCategoryFromAction determines the category based on action
func getCategoryFromAction(action string) string {
	switch action {
	case "login", "logout", "login_failed", "password_reset_request", "password_reset":
		return models.ActivityCategoryAuth
	case "2fa_enabled", "2fa_disabled", "2fa_verified", "2fa_failed", "password_changed":
		return models.ActivityCategorySecurity
	case "profile_updated", "address_updated", "email_verified":
		return models.ActivityCategoryProfile
	case "preferences_updated", "settings_updated":
		return models.ActivityCategorySettings
	case "event_created", "event_published", "event_updated", "event_deleted":
		return models.ActivityCategoryEvent
	case "order_placed", "order_cancelled", "order_refunded":
		return models.ActivityCategoryOrder
	case "payment_processed", "payment_failed", "payment_method_added", "payment_method_removed":
		return models.ActivityCategoryPayment
	case "ticket_generated", "ticket_transferred", "ticket_checked_in":
		return models.ActivityCategoryTicket
	case "refund_requested", "refund_approved", "refund_processed":
		return models.ActivityCategoryRefund
	default:
		return "general"
	}
}

// GetActivityTypes handles getting available activity types for filtering
func (h *AccountHandler) GetActivityTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	activityTypes := []string{
		models.ActionLogin,
		models.ActionLoginFailed,
		models.ActionLogout,
		models.ActionPasswordChanged,
		models.ActionPasswordResetRequest,
		models.ActionPasswordReset,
		models.Action2FAEnabled,
		models.Action2FADisabled,
		models.Action2FAVerified,
		models.ActionProfileUpdated,
		models.ActionAddressUpdated,
		models.ActionPreferencesUpdated,
		models.ActionEmailVerified,
		models.ActionEventCreated,
		models.ActionEventPublished,
		models.ActionEventUpdated,
		models.ActionEventDeleted,
		models.ActionOrderPlaced,
		models.ActionOrderCancelled,
		models.ActionOrderRefunded,
		models.ActionPaymentProcessed,
		models.ActionPaymentFailed,
		models.ActionPaymentMethodAdded,
		models.ActionPaymentMethodRemoved,
		models.ActionTicketGenerated,
		models.ActionTicketTransferred,
		models.ActionRefundRequested,
		models.ActionRefundApproved,
	}

	categories := []string{
		models.ActivityCategoryAuth,
		models.ActivityCategoryProfile,
		models.ActivityCategoryEvent,
		models.ActivityCategoryOrder,
		models.ActivityCategoryPayment,
		models.ActivityCategorySecurity,
		models.ActivityCategoryTicket,
		models.ActivityCategoryRefund,
		models.ActivityCategorySettings,
	}

	response := map[string]interface{}{
		"activity_types": activityTypes,
		"categories":     categories,
	}

	json.NewEncoder(w).Encode(response)
}

// parseActivityFilter parses query parameters for activity filtering
func parseActivityFilter(r *http.Request) ActivityFilter {
	filter := ActivityFilter{
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

	// Parse action
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}

	// Parse category
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = &category
	}

	// Parse success filter
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		switch successStr {
		case "true":
			success := true
			filter.Success = &success
		case "false":
			success := false
			filter.Success = &success
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

	return filter
}

// GetAccountStats handles getting account statistics
func (h *AccountHandler) GetAccountStats(w http.ResponseWriter, r *http.Request) {
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

	// Get user to find account ID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get account statistics
	stats := getAccountStatistics(h.db, user.AccountID)

	json.NewEncoder(w).Encode(stats)
}

// LogActivity handles manual activity logging
func (h *AccountHandler) LogActivity(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	var req struct {
		Action      string `json:"action"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user to find account ID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Log the activity
	logAccountActivity(h.db, user.AccountID, req.Action, req.Description, r.RemoteAddr, r.UserAgent(), true)

	response := map[string]interface{}{
		"message": "Activity logged successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// ClearActivityLog handles clearing account activity logs (admin only)
func (h *AccountHandler) ClearActivityLog(w http.ResponseWriter, r *http.Request) {
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

	// Get user and verify admin role
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleAdmin {
		middleware.WriteJSONError(w, http.StatusForbidden, "admin access required")
		return
	}

	// Parse request to get account ID to clear
	var req struct {
		AccountID uint   `json:"account_id"`
		OlderThan string `json:"older_than"` // Optional: clear only old records
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccountID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	// Build delete query
	query := h.db.Where("account_id = ?", req.AccountID)

	// If older_than specified, only delete old records
	if req.OlderThan != "" {
		if olderThan, err := time.Parse("2006-01-02", req.OlderThan); err == nil {
			query = query.Where("timestamp < ?", olderThan)
		}
	}

	// Count records to be deleted
	var count int64
	query.Model(&models.AccountActivity{}).Count(&count)

	// Delete activities
	if err := query.Delete(&models.AccountActivity{}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to clear activity log")
		return
	}

	// Log this action
	h.LogActivityWithDetails(user.AccountID, &userID, "activity_log_cleared",
		models.ActivityCategoryAdmin,
		fmt.Sprintf("Cleared %d activity records for account %d", count, req.AccountID),
		getClientIP(r), r.UserAgent(), true, models.SeverityInfo)

	response := map[string]interface{}{
		"message":         "Activity log cleared successfully",
		"records_deleted": count,
	}

	json.NewEncoder(w).Encode(response)
}

// getAccountStatistics returns account statistics
func getAccountStatistics(db *gorm.DB, accountID uint) map[string]interface{} {
	stats := make(map[string]interface{})

	// Get account age
	var account models.Account
	if err := db.Where("id = ?", accountID).First(&account).Error; err == nil {
		accountAge := int(time.Since(account.CreatedAt).Hours() / 24)
		stats["account_age_days"] = accountAge
		stats["account_created_at"] = account.CreatedAt
	}

	// Total events (for organizers)
	var totalEvents int64
	db.Model(&models.Event{}).Joins("JOIN organizers ON organizers.id = events.organizer_id").
		Where("organizers.account_id = ?", accountID).Count(&totalEvents)
	stats["total_events"] = totalEvents

	// Published events
	var publishedEvents int64
	db.Model(&models.Event{}).Joins("JOIN organizers ON organizers.id = events.organizer_id").
		Where("organizers.account_id = ? AND events.status = ?", accountID, "published").Count(&publishedEvents)
	stats["published_events"] = publishedEvents

	// Total orders
	var totalOrders int64
	db.Model(&models.Order{}).Where("account_id = ?", accountID).Count(&totalOrders)
	stats["total_orders"] = totalOrders

	// Total tickets
	var totalTickets int64
	db.Model(&models.Ticket{}).Joins("JOIN orders ON orders.id = tickets.order_id").
		Where("orders.account_id = ?", accountID).Count(&totalTickets)
	stats["total_tickets"] = totalTickets

	// Events this month
	monthStart := time.Now().AddDate(0, 0, -30)
	var eventsThisMonth int64
	db.Model(&models.Event{}).Joins("JOIN organizers ON organizers.id = events.organizer_id").
		Where("organizers.account_id = ? AND events.created_at >= ?", accountID, monthStart).Count(&eventsThisMonth)
	stats["events_this_month"] = eventsThisMonth

	// Orders this month
	var ordersThisMonth int64
	db.Model(&models.Order{}).Where("account_id = ? AND created_at >= ?", accountID, monthStart).Count(&ordersThisMonth)
	stats["orders_this_month"] = ordersThisMonth

	// Last activity
	var lastActivity models.AccountActivity
	if err := db.Where("account_id = ?", accountID).Order("timestamp DESC").First(&lastActivity).Error; err == nil {
		stats["last_activity"] = lastActivity.Timestamp
		stats["last_activity_action"] = lastActivity.Action
	}

	// Total activities logged
	var totalActivities int64
	db.Model(&models.AccountActivity{}).Where("account_id = ?", accountID).Count(&totalActivities)
	stats["total_activities"] = totalActivities

	// Failed activities (potential security issues)
	var failedActivities int64
	db.Model(&models.AccountActivity{}).Where("account_id = ? AND success = ?", accountID, false).Count(&failedActivities)
	stats["failed_activities"] = failedActivities

	// Login statistics
	var totalLogins int64
	var failedLogins int64
	db.Model(&models.LoginHistory{}).Where("account_id = ? AND success = ?", accountID, true).Count(&totalLogins)
	db.Model(&models.LoginHistory{}).Where("account_id = ? AND success = ?", accountID, false).Count(&failedLogins)
	stats["total_logins"] = totalLogins
	stats["failed_logins"] = failedLogins

	// Last login
	var lastLogin models.LoginHistory
	if err := db.Where("account_id = ? AND success = ?", accountID, true).Order("login_at DESC").First(&lastLogin).Error; err == nil {
		stats["last_login"] = lastLogin.LoginAt
		stats["last_login_ip"] = lastLogin.IPAddress
		stats["last_login_location"] = lastLogin.Location
	}

	return stats
}

// logAccountActivity logs an account activity
func logAccountActivity(db *gorm.DB, accountID uint, action, description, ipAddress, userAgent string, success bool) {
	severity := models.SeverityInfo
	if !success {
		severity = models.SeverityWarning
	}

	activity := models.AccountActivity{
		AccountID:   accountID,
		Action:      action,
		Category:    getCategoryFromAction(action),
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Success:     success,
		Severity:    severity,
		Timestamp:   time.Now(),
	}

	db.Create(&activity)
}
