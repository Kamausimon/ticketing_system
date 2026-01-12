package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// UserHandler handles admin user management operations
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new admin user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// UserListResponse represents a user in the list
type UserListResponse struct {
	ID            uint      `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	Phone         *string   `json:"phone"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	LastLoginAt   *int64    `json:"last_login_at"`
}

// UserDetailsResponse represents detailed user information
type UserDetailsResponse struct {
	UserListResponse
	AccountID        uint       `json:"account_id"`
	Isconfirmed      bool       `json:"is_confirmed"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at"`
	TokenVersion     int        `json:"token_version"`
	UpdatedAt        time.Time  `json:"updated_at"`
	HasTwoFactorAuth bool       `json:"has_two_factor_auth"`
}

// UpdateRoleRequest represents a role update request
type UpdateRoleRequest struct {
	Role   string `json:"role"`   // "customer", "organizer", "admin"
	Reason string `json:"reason"` // Optional reason for the change
}

// UpdateStatusRequest represents a status update request
type UpdateStatusRequest struct {
	IsActive bool   `json:"is_active"`
	Reason   string `json:"reason"` // Optional reason for the change
}

// ListUsers returns all users with pagination (Admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	role := r.URL.Query().Get("role")
	isActive := r.URL.Query().Get("is_active")

	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.User{})

	if role != "" {
		query = query.Where("role = ?", role)
	}
	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get users
	var users []models.User
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	// Convert to response format
	var userList []UserListResponse
	for _, user := range users {
		userList = append(userList, UserListResponse{
			ID:            user.ID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Email:         user.Email,
			Username:      user.Username,
			Phone:         user.Phone,
			Role:          string(user.Role),
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
			LastLoginAt:   user.LastLoginAt,
		})
	}

	response := map[string]interface{}{
		"users": userList,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetUserDetails returns detailed information about a specific user (Admin only)
func (h *UserHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if user has 2FA enabled
	var twoFactorAuth models.TwoFactorAuth
	hasTwoFA := h.db.Where("user_id = ? AND enabled = ?", user.ID, true).First(&twoFactorAuth).Error == nil

	response := UserDetailsResponse{
		UserListResponse: UserListResponse{
			ID:            user.ID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Email:         user.Email,
			Username:      user.Username,
			Phone:         user.Phone,
			Role:          string(user.Role),
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
			LastLoginAt:   user.LastLoginAt,
		},
		AccountID:        user.AccountID,
		Isconfirmed:      user.Isconfirmed,
		EmailVerifiedAt:  user.EmailVerifiedAt,
		TokenVersion:     user.TokenVersion,
		UpdatedAt:        user.UpdatedAt,
		HasTwoFactorAuth: hasTwoFA,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateUserRole updates a user's role (Admin only)
func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Parse request
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"customer":  true,
		"organizer": true,
		"admin":     true,
	}
	if !validRoles[req.Role] {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid role. Must be: customer, organizer, or admin")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Prevent changing own role
	adminUserID := middleware.GetUserIDFromToken(r)
	if adminUserID == user.ID {
		middleware.WriteJSONError(w, http.StatusForbidden, "cannot change your own role")
		return
	}

	// Update role
	oldRole := user.Role
	if err := h.db.Model(&user).Update("role", req.Role).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update user role")
		return
	}

	// Log the activity (optional)
	// You can add activity logging here

	response := map[string]interface{}{
		"message":  "User role updated successfully",
		"user_id":  user.ID,
		"old_role": oldRole,
		"new_role": req.Role,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateUserStatus activates or deactivates a user account (Admin only)
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Parse request
	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Prevent changing own account status
	adminUserID := middleware.GetUserIDFromToken(r)
	if adminUserID == user.ID {
		middleware.WriteJSONError(w, http.StatusForbidden, "cannot change your own account status")
		return
	}

	// Update status
	if err := h.db.Model(&user).Update("is_active", req.IsActive).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update user status")
		return
	}

	// If deactivating, invalidate all tokens
	if !req.IsActive {
		h.db.Model(&user).Update("token_version", gorm.Expr("token_version + 1"))
	}

	status := "activated"
	if !req.IsActive {
		status = "deactivated"
	}

	response := map[string]interface{}{
		"message":   "User account " + status + " successfully",
		"user_id":   user.ID,
		"is_active": req.IsActive,
	}

	json.NewEncoder(w).Encode(response)
}

// SearchUsers searches for users by name, email, or username (Admin only)
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "search query is required")
		return
	}

	// Search users
	var users []models.User
	searchPattern := "%" + query + "%"
	if err := h.db.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(50).Find(&users).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	// Convert to response format
	var userList []UserListResponse
	for _, user := range users {
		userList = append(userList, UserListResponse{
			ID:            user.ID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Email:         user.Email,
			Username:      user.Username,
			Phone:         user.Phone,
			Role:          string(user.Role),
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
			LastLoginAt:   user.LastLoginAt,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userList,
		"count": len(userList),
	})
}

// GetUserStats returns statistics about users (Admin only)
func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if user is admin
	if !h.isAdmin(w, r) {
		return
	}

	// Get total users
	var totalUsers int64
	h.db.Model(&models.User{}).Count(&totalUsers)

	// Get users by role
	var roleStats []struct {
		Role  string
		Count int64
	}
	h.db.Model(&models.User{}).Select("role, COUNT(*) as count").Group("role").Scan(&roleStats)

	// Get active users
	var activeUsers int64
	h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers)

	// Get verified emails
	var verifiedEmails int64
	h.db.Model(&models.User{}).Where("email_verified = ?", true).Count(&verifiedEmails)

	// Get users with 2FA enabled
	var twoFAUsers int64
	h.db.Model(&models.TwoFactorAuth{}).Where("enabled = ?", true).Count(&twoFAUsers)

	// Get recent registrations (last 30 days)
	var recentRegistrations int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	h.db.Model(&models.User{}).Where("created_at >= ?", thirtyDaysAgo).Count(&recentRegistrations)

	response := map[string]interface{}{
		"total_users":          totalUsers,
		"active_users":         activeUsers,
		"inactive_users":       totalUsers - activeUsers,
		"verified_emails":      verifiedEmails,
		"users_with_2fa":       twoFAUsers,
		"recent_registrations": recentRegistrations,
		"users_by_role":        roleStats,
	}

	json.NewEncoder(w).Encode(response)
}

// isAdmin checks if the current user is an admin
func (h *UserHandler) isAdmin(w http.ResponseWriter, r *http.Request) bool {
	userID := middleware.GetUserIDFromToken(r)

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return false
	}

	if user.Role != models.RoleAdmin {
		middleware.WriteJSONError(w, http.StatusForbidden, "admin access required")
		return false
	}

	return true
}
