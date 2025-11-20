package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// GetAccountProfile handles getting user's account profile
func (h *AccountHandler) GetAccountProfile(w http.ResponseWriter, r *http.Request) {
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

	// Get account with payment gateway info
	var account models.Account
	if err := h.db.Preload("PaymentGateway").Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Convert to response format
	response := convertToAccountResponse(account)

	json.NewEncoder(w).Encode(response)
}

// UpdateAccountProfile handles updating user's basic profile information
func (h *AccountHandler) UpdateAccountProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateProfileUpdate(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
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

	// Check if email is being changed and if it's already taken
	if !strings.EqualFold(req.Email, account.Email) {
		var existingAccount models.Account
		if err := h.db.Where("email = ? AND id != ?", strings.ToLower(req.Email), account.ID).First(&existingAccount).Error; err == nil {
			middleware.WriteJSONError(w, http.StatusConflict, "email already in use")
			return
		}
	}

	// Update account fields
	account.FirstName = strings.TrimSpace(req.FirstName)
	account.LastName = strings.TrimSpace(req.LastName)
	account.Email = strings.ToLower(strings.TrimSpace(req.Email))
	account.TimezoneID = req.TimezoneID
	account.DateFormatID = req.DateFormatID
	account.CurrencyID = req.CurrencyID

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	// Log activity
	h.logAccountActivity(account.ID, "profile_updated", "Profile information updated", getClientIP(r))

	response := map[string]interface{}{
		"message": "Profile updated successfully",
		"account": convertToAccountResponse(account),
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteAccount handles account deletion (soft delete)
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
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

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Soft delete by deactivating account
	account.IsActive = false

	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to deactivate account")
		return
	}

	// Also deactivate user
	user.IsActive = false
	h.db.Save(&user)

	// Log activity
	h.logAccountActivity(account.ID, "account_deleted", "Account deactivated", getClientIP(r))

	response := map[string]interface{}{
		"message": "Account deactivated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Helper functions
func validateProfileUpdate(req UpdateProfileRequest) error {
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper regex or library
	return len(email) > 0 &&
		len(email) <= 255 &&
		strings.Contains(email, "@") &&
		strings.Contains(email, ".") &&
		email[0] != '@' &&
		email[len(email)-1] != '@'
}

func getClientIP(r *http.Request) string {
	// Try to get IP from X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
