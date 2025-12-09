package organizers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/security"
)

// BankDetailsRequest represents bank account details submission
// These details are used for PAYOUTS ONLY - not for collecting customer payments.
// The platform collects all customer payments through its own payment gateway,
// then transfers funds to organizers using these bank details.
type BankDetailsRequest struct {
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankCode          string `json:"bank_code"`    // Bank SWIFT/Sort code
	BankCountry       string `json:"bank_country"` // ISO country code
}

// BankDetailsResponse represents the response after saving bank details
type BankDetailsResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// UpdateBankDetails handles bank account details submission for payouts.
// This is where organizers specify where they want to receive their earnings.
// All customer payments are collected by the platform, not by individual organizers.
func (h *OrganizerHandler) UpdateBankDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Parse request
	var req BankDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.BankAccountName == "" || req.BankAccountNumber == "" || req.BankCode == "" || req.BankCountry == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "all bank details are required")
		return
	}

	// Get user and organizer
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Encrypt sensitive bank details
	if h.encryption == nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "encryption service not available")
		return
	}

	encryptedAccountNumber, encryptedBankCode, err := h.encryption.EncryptBankDetails(
		req.BankAccountNumber,
		req.BankCode,
	)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to encrypt bank details: %v", err))
		return
	}

	// Check if bank details are being changed (not first time setup)
	isUpdate := organizer.BankAccountNumber != ""

	// Update bank details with encrypted values
	updates := map[string]interface{}{
		"bank_account_name":   req.BankAccountName,
		"bank_account_number": encryptedAccountNumber,
		"bank_code":           encryptedBankCode,
		"bank_country":        req.BankCountry,
	}

	if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update bank details")
		return
	}

	// Send notification email if this is an update (not first time setup)
	if isUpdate && h.notifications != nil {
		go h.sendBankDetailsChangeNotification(organizer, user, r)
	}

	response := BankDetailsResponse{
		Message: "Bank details updated successfully",
		Status:  "success",
	}

	json.NewEncoder(w).Encode(response)
}

// GetBankDetails retrieves stored bank details for organizer
func (h *OrganizerHandler) GetBankDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get user and organizer
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Decrypt sensitive bank details
	var accountNumber, bankCode string
	if h.encryption != nil && organizer.BankAccountNumber != "" {
		var err error
		accountNumber, bankCode, err = h.encryption.DecryptBankDetails(
			organizer.BankAccountNumber,
			organizer.BankCode,
		)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to decrypt bank details: %v", err))
			return
		}
	} else {
		// Fallback for unencrypted data (backward compatibility)
		accountNumber = organizer.BankAccountNumber
		bankCode = organizer.BankCode
	}

	// Return decrypted bank details with masked account number for display
	bankDetails := map[string]interface{}{
		"bank_account_name":        organizer.BankAccountName,
		"bank_account_number":      accountNumber, // Full number for editing
		"bank_account_number_mask": security.MaskBankAccountNumber(accountNumber),
		"bank_code":                bankCode,
		"bank_country":             organizer.BankCountry,
	}

	json.NewEncoder(w).Encode(bankDetails)
}

// sendBankDetailsChangeNotification sends email notification when bank details are changed
func (h *OrganizerHandler) sendBankDetailsChangeNotification(organizer models.Organizer, user models.User, r *http.Request) {
	// Get all users in the same account for notification
	var accountUsers []models.User
	if err := h.db.Where("account_id = ?", user.AccountID).Find(&accountUsers).Error; err != nil {
		fmt.Printf("Failed to get account users: %v\n", err)
		return
	}

	// Send notification to all users in the account
	for _, accountUser := range accountUsers {
		emailData := map[string]interface{}{
			"Name":           accountUser.FirstName + " " + accountUser.LastName,
			"OrganizerName":  organizer.Name,
			"ChangedBy":      user.FirstName + " " + user.LastName,
			"ChangedByEmail": user.Email,
			"IPAddress":      getClientIP(r),
			"Timestamp":      fmt.Sprintf("%s", r.Context().Value("timestamp")),
			"SupportEmail":   h.notifications.GetSupportEmail(),
		}

		if err := h.notifications.SendBankDetailsChangeNotification(accountUser.Email, emailData); err != nil {
			fmt.Printf("Failed to send bank details change notification to %s: %v\n", accountUser.Email, err)
		}
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fallback to RemoteAddr
	return r.RemoteAddr
}
