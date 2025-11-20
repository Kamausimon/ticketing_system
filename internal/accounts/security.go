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

	// For now, return mock login history since we don't have a login_history table
	// In production, you would query actual login history from database
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Create mock login history based on account info
	loginHistory := []LoginHistory{}

	if account.LastLoginDate != nil {
		loginHistory = append(loginHistory, LoginHistory{
			ID:        1,
			AccountID: account.ID,
			IPAddress: *account.LastIP,
			UserAgent: nil, // Not stored in current model
			Location:  nil, // Would need geolocation service
			Success:   true,
			Timestamp: *account.LastLoginDate,
		})
	}

	response := map[string]interface{}{
		"login_history": loginHistory,
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

	// Return security settings
	settings := map[string]interface{}{
		"two_factor_enabled":   false, // Not implemented yet
		"email_notifications":  true,  // Default
		"login_notifications":  true,  // Default
		"account_status":       "active",
		"last_password_change": user.UpdatedAt, // Approximation
		"last_login":           account.LastLoginDate,
		"last_login_ip":        account.LastIP,
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

	// TODO: Implement account locking logic
	response := map[string]interface{}{
		"message": "Account lock functionality not implemented yet",
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

	// TODO: Implement account unlocking logic
	response := map[string]interface{}{
		"message": "Account unlock functionality not implemented yet",
	}

	json.NewEncoder(w).Encode(response)
}

// getRecentLogins returns recent login records (mock data for now)
func getRecentLogins(accountID uint) []LoginRecord {
	// In a real implementation, this would fetch from a login_logs table
	return []LoginRecord{
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Location:  "Nairobi, Kenya",
			Success:   true,
		},
		{
			Timestamp: time.Now().Add(-6 * time.Hour),
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X)",
			Location:  "Nairobi, Kenya",
			Success:   true,
		},
	}
}

// logSecurityEvent logs a security-related event (mock implementation)
func logSecurityEvent(db *gorm.DB, accountID uint, action, ipAddress, userAgent string) {
	// In a real implementation, this would insert into a security_logs table
	fmt.Printf("Security Event: Account %d - %s from %s\n", accountID, action, ipAddress)
}
