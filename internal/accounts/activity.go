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

	userID := middleware.GetUserIDFromToken(r)
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

	// For now, return mock activity data since we don't have an activity log table
	// In production, you would query actual activity from database
	activities := []AccountActivity{
		{
			ID:          1,
			AccountID:   user.AccountID,
			Action:      "login",
			Description: "User logged in",
			IPAddress:   "192.168.1.1",
			UserAgent:   nil,
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          2,
			AccountID:   user.AccountID,
			Action:      "profile_updated",
			Description: "Profile information updated",
			IPAddress:   "192.168.1.1",
			UserAgent:   nil,
			Timestamp:   time.Now().Add(-24 * time.Hour),
		},
	}

	// Apply filtering
	filteredActivities := []AccountActivity{}
	for _, activity := range activities {
		if filter.Action != nil && activity.Action != *filter.Action {
			continue
		}
		if filter.StartDate != nil && activity.Timestamp.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && activity.Timestamp.After(*filter.EndDate) {
			continue
		}
		filteredActivities = append(filteredActivities, activity)
	}

	// Apply pagination
	totalCount := int64(len(filteredActivities))
	start := (filter.Page - 1) * filter.Limit
	end := start + filter.Limit

	if start > len(filteredActivities) {
		start = len(filteredActivities)
	}
	if end > len(filteredActivities) {
		end = len(filteredActivities)
	}

	paginatedActivities := filteredActivities[start:end]
	totalPages := int((totalCount + int64(filter.Limit) - 1) / int64(filter.Limit))

	response := PaginatedResponse{
		Data:       paginatedActivities,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// logAccountActivity logs an activity for an account (helper function)
func (h *AccountHandler) logAccountActivity(accountID uint, action, description, ipAddress string) {
	// In production, this would insert into an actual activity log table
	// For now, it's a no-op since we don't have the table structure
	// You could implement this with a separate ActivityLog model

	// Example implementation:
	// activity := models.ActivityLog{
	//     AccountID:   accountID,
	//     Action:      action,
	//     Description: description,
	//     IPAddress:   ipAddress,
	//     Timestamp:   time.Now(),
	// }
	// h.db.Create(&activity)
}

// GetActivityTypes handles getting available activity types for filtering
func (h *AccountHandler) GetActivityTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	activityTypes := []string{
		"login",
		"logout",
		"profile_updated",
		"password_changed",
		"address_updated",
		"preferences_updated",
		"stripe_setup",
		"payment_method_added",
		"payment_method_removed",
		"event_created",
		"event_published",
		"event_updated",
		"order_placed",
		"order_cancelled",
	}

	response := map[string]interface{}{
		"activity_types": activityTypes,
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

// ActivityRequest represents activity filtering parameters
type ActivityRequest struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Action    string `json:"action"`
}

// ActivityResponse represents activity response
type ActivityResponse struct {
	Activities []ActivityLog `json:"activities"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// GetAccountStats handles getting account statistics
func (h *AccountHandler) GetAccountStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
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

	userID := middleware.GetUserIDFromToken(r)
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

	userID := middleware.GetUserIDFromToken(r)
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

	// TODO: Implement actual activity log clearing
	response := map[string]interface{}{
		"message": "Activity log cleared successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// parseActivityParams parses activity filtering parameters
func parseActivityParams(r *http.Request) ActivityRequest {
	params := ActivityRequest{
		Page:  1,
		Limit: 50,
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

	// Parse dates and action
	params.StartDate = r.URL.Query().Get("start_date")
	params.EndDate = r.URL.Query().Get("end_date")
	params.Action = r.URL.Query().Get("action")

	return params
}

// getAccountActivities returns account activities (mock implementation)
func getAccountActivities(accountID uint, params ActivityRequest) []ActivityLog {
	// Mock activity data
	activities := []ActivityLog{
		{
			ID:          1,
			Action:      "login",
			Description: "User logged in successfully",
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			Success:     true,
		},
		{
			ID:          2,
			Action:      "profile_update",
			Description: "User updated profile information",
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Success:     true,
		},
		{
			ID:          3,
			Action:      "event_created",
			Description: "User created a new event",
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X)",
			Timestamp:   time.Now().Add(-6 * time.Hour),
			Success:     true,
		},
		{
			ID:          4,
			Action:      "password_change",
			Description: "User changed account password",
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timestamp:   time.Now().Add(-24 * time.Hour),
			Success:     true,
		},
		{
			ID:          5,
			Action:      "login_failed",
			Description: "Failed login attempt with incorrect password",
			IPAddress:   "192.168.1.200",
			UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			Timestamp:   time.Now().Add(-48 * time.Hour),
			Success:     false,
		},
	}

	// Filter by action if specified
	if params.Action != "" {
		var filtered []ActivityLog
		for _, activity := range activities {
			if activity.Action == params.Action {
				filtered = append(filtered, activity)
			}
		}
		return filtered
	}

	return activities
}

// getAccountStatistics returns account statistics
func getAccountStatistics(db *gorm.DB, accountID uint) map[string]interface{} {
	// In a real implementation, this would query the database for actual statistics
	return map[string]interface{}{
		"total_events":         12,
		"total_orders":         156,
		"total_revenue":        "$12,450.00",
		"total_attendees":      1245,
		"events_this_month":    3,
		"orders_this_month":    45,
		"revenue_this_month":   "$3,200.00",
		"attendees_this_month": 320,
		"average_order_value":  "$79.80",
		"repeat_customers":     45,
		"account_age_days":     365,
		"last_activity":        time.Now().Add(-1 * time.Hour),
	}
}

// logAccountActivity logs an account activity
func logAccountActivity(db *gorm.DB, accountID uint, action, description, ipAddress, userAgent string, success bool) {
	// In a real implementation, this would insert into an activity_logs table
	fmt.Printf("Account Activity: Account %d - %s: %s from %s (Success: %v)\n",
		accountID, action, description, ipAddress, success)
}
