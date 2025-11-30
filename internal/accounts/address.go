package accounts

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// GetAccountAddress handles getting user's address information
func (h *AccountHandler) GetAccountAddress(w http.ResponseWriter, r *http.Request) {
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

	// Return address information
	address := map[string]interface{}{
		"address1":    account.Address1,
		"address2":    account.Address2,
		"city":        account.City,
		"county":      account.County,
		"postal_code": account.PostalCode,
	}

	json.NewEncoder(w).Encode(address)
}

// UpdateAccountAddress handles updating user's address information
func (h *AccountHandler) UpdateAccountAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
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

	// Update address fields
	if req.Address1 != nil {
		trimmed := strings.TrimSpace(*req.Address1)
		if trimmed == "" {
			account.Address1 = nil
		} else {
			account.Address1 = &trimmed
		}
	}

	if req.Address2 != nil {
		trimmed := strings.TrimSpace(*req.Address2)
		if trimmed == "" {
			account.Address2 = nil
		} else {
			account.Address2 = &trimmed
		}
	}

	if req.City != nil {
		trimmed := strings.TrimSpace(*req.City)
		if trimmed == "" {
			account.City = nil
		} else {
			account.City = &trimmed
		}
	}

	if req.County != nil {
		trimmed := strings.TrimSpace(*req.County)
		if trimmed == "" {
			account.County = nil
		} else {
			account.County = &trimmed
		}
	}

	if req.PostalCode != nil {
		trimmed := strings.TrimSpace(*req.PostalCode)
		if trimmed == "" {
			account.PostalCode = nil
		} else {
			account.PostalCode = &trimmed
		}
	}

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update address")
		return
	}

	// Log activity
	h.logAccountActivity(account.ID, "address_updated", "Address information updated", getClientIP(r))

	response := map[string]interface{}{
		"message": "Address updated successfully",
		"address": map[string]interface{}{
			"address1":    account.Address1,
			"address2":    account.Address2,
			"city":        account.City,
			"county":      account.County,
			"postal_code": account.PostalCode,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// AddressUpdateRequest represents address update request
type AddressUpdateRequest struct {
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2"`
	City       string `json:"city"`
	County     string `json:"county"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// ClearAccountAddress handles clearing account address information
func (h *AccountHandler) ClearAccountAddress(w http.ResponseWriter, r *http.Request) {
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

	// Clear address fields
	account.Address1 = nil
	account.Address2 = nil
	account.City = nil
	account.County = nil
	account.PostalCode = nil

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to clear address")
		return
	}

	response := map[string]interface{}{
		"message": "Address cleared successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetSupportedCountries returns list of supported countries
func (h *AccountHandler) GetSupportedCountries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	countries := []map[string]interface{}{
		{"code": "KE", "name": "Kenya"},
		{"code": "US", "name": "United States"},
		{"code": "GB", "name": "United Kingdom"},
		{"code": "CA", "name": "Canada"},
		{"code": "AU", "name": "Australia"},
		{"code": "ZA", "name": "South Africa"},
		{"code": "NG", "name": "Nigeria"},
		{"code": "GH", "name": "Ghana"},
		{"code": "UG", "name": "Uganda"},
		{"code": "TZ", "name": "Tanzania"},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"countries": countries,
	})
}
