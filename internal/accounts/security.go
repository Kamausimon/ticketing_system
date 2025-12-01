package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ChangePassword handles changing user's password
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req SecuritySettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.CurrentPassword == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "current password is required")
		return
	}
	if req.NewPassword == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new password is required")
		return
	}
	if len(req.NewPassword) < 8 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new password must be at least 8 characters long")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "current password is incorrect")
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Update password
	user.Password = string(newPasswordHash)
	user.UpdatedAt = time.Now()

	if err := h.db.Save(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	// Log activity
	h.logAccountActivity(user.AccountID, "password_changed", "Password changed successfully", getClientIP(r))

	// Log security event
	logSecurityEvent(h.db, user.AccountID, "password_changed", getClientIP(r), r.UserAgent())

	response := map[string]interface{}{
		"message": "Password changed successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetLoginHistory handles getting user's login history
func (h *AccountHandler) GetLoginHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
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

	// Fetch actual login history from database
	var loginHistory []models.LoginHistory
	if err := h.db.Where("account_id = ?", user.AccountID).
		Order("login_at DESC").
		Limit(50).
		Find(&loginHistory).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch login history")
		return
	}

	// Convert to response format
	responseHistory := make([]map[string]interface{}, len(loginHistory))
	for i, login := range loginHistory {
		responseHistory[i] = map[string]interface{}{
			"id":               login.ID,
			"ip_address":       login.IPAddress,
			"user_agent":       login.UserAgent,
			"location":         login.Location,
			"device":           login.Device,
			"browser":          login.Browser,
			"success":          login.Success,
			"fail_reason":      login.FailReason,
			"login_at":         login.LoginAt,
			"logout_at":        login.LogoutAt,
			"session_duration": login.SessionDuration,
		}
	}

	response := map[string]interface{}{
		"login_history": responseHistory,
		"count":         len(loginHistory),
	}

	json.NewEncoder(w).Encode(response)
}

// GetSecuritySettings handles getting user's security settings
func (h *AccountHandler) GetSecuritySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
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

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Check if 2FA is enabled
	var twoFactorAuth models.TwoFactorAuth
	twoFactorEnabled := false
	if err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error; err == nil {
		twoFactorEnabled = true
	}

	// Get recent logins
	recentLogins := getRecentLogins(h.db, user.AccountID)

	// Return security settings
	settings := map[string]interface{}{
		"two_factor_enabled":   twoFactorEnabled,
		"email_notifications":  true, // Default
		"login_notifications":  true, // Default
		"account_status":       "active",
		"last_password_change": user.UpdatedAt, // Approximation
		"last_login":           account.LastLoginDate,
		"last_login_ip":        account.LastIP,
		"recent_logins":        recentLogins,
	}

	json.NewEncoder(w).Encode(settings)
}

// PasswordChangeRequest represents password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// SecuritySettingsResponse represents security settings
type SecuritySettingsResponse struct {
	TwoFactorEnabled   bool          `json:"two_factor_enabled"`
	LastPasswordChange time.Time     `json:"last_password_change"`
	LoginAttempts      int           `json:"login_attempts"`
	AccountLocked      bool          `json:"account_locked"`
	SecurityQuestions  int           `json:"security_questions_count"`
	RecentLogins       []LoginRecord `json:"recent_logins"`
}

type LoginRecord struct {
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Location  string    `json:"location"`
	Success   bool      `json:"success"`
}

// LockAccount handles account locking (admin function)
func (h *AccountHandler) LockAccount(w http.ResponseWriter, r *http.Request) {
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

	// Parse request to get target account ID
	var req struct {
		AccountID uint `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccountID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	// Get the account to lock
	var account models.Account
	if err := h.db.First(&account, req.AccountID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Check if account is already locked
	if !account.IsActive {
		middleware.WriteJSONError(w, http.StatusBadRequest, "account is already locked")
		return
	}

	// Lock the account by setting IsActive to false
	if err := h.db.Model(&account).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to lock account")
		return
	}

	// Log security event
	logSecurityEvent(h.db, req.AccountID, "account_locked", getClientIP(r), r.UserAgent())

	// Log account activity
	h.logAccountActivity(req.AccountID, "account_locked", fmt.Sprintf("Account locked by admin (User ID: %d)", userID), getClientIP(r))

	response := map[string]interface{}{
		"success":    true,
		"message":    "Account locked successfully",
		"account_id": req.AccountID,
	}

	json.NewEncoder(w).Encode(response)
}

// UnlockAccount handles account unlocking (admin function)
func (h *AccountHandler) UnlockAccount(w http.ResponseWriter, r *http.Request) {
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

	// Parse request to get target account ID
	var req struct {
		AccountID uint `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccountID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	// Get the account to unlock
	var account models.Account
	if err := h.db.First(&account, req.AccountID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Check if account is already unlocked
	if account.IsActive {
		middleware.WriteJSONError(w, http.StatusBadRequest, "account is already active")
		return
	}

	// Unlock the account by setting IsActive to true
	if err := h.db.Model(&account).Updates(map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to unlock account")
		return
	}

	// Log security event
	logSecurityEvent(h.db, req.AccountID, "account_unlocked", getClientIP(r), r.UserAgent())

	// Log account activity
	h.logAccountActivity(req.AccountID, "account_unlocked", fmt.Sprintf("Account unlocked by admin (User ID: %d)", userID), getClientIP(r))

	response := map[string]interface{}{
		"success":    true,
		"message":    "Account unlocked successfully",
		"account_id": req.AccountID,
	}

	json.NewEncoder(w).Encode(response)
}

// getRecentLogins returns recent login records from database
func getRecentLogins(db *gorm.DB, accountID uint) []LoginRecord {
	var loginHistory []models.LoginHistory
	if err := db.Where("account_id = ?", accountID).
		Order("login_at DESC").
		Limit(10).
		Find(&loginHistory).Error; err != nil {
		return []LoginRecord{}
	}

	// Convert to LoginRecord format
	records := make([]LoginRecord, len(loginHistory))
	for i, login := range loginHistory {
		location := ""
		if login.Location != nil {
			location = *login.Location
		}
		records[i] = LoginRecord{
			Timestamp: login.LoginAt,
			IPAddress: login.IPAddress,
			UserAgent: login.UserAgent,
			Location:  location,
			Success:   login.Success,
		}
	}

	return records
}

// logSecurityEvent logs a security-related event
func logSecurityEvent(db *gorm.DB, accountID uint, action, ipAddress, userAgent string) {
	activity := models.AccountActivity{
		AccountID:   accountID,
		Action:      action,
		Category:    models.ActivityCategorySecurity,
		Description: fmt.Sprintf("Security event: %s", action),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Success:     true,
		Severity:    models.SeverityWarning,
		Timestamp:   time.Now(),
	}
	db.Create(&activity)
}
