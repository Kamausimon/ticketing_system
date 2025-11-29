package organizers

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// BankDetailsRequest represents bank account details submission
type BankDetailsRequest struct {
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankCode          string `json:"bank_code"`
	BankCountry       string `json:"bank_country"`
}

// BankDetailsResponse represents the response after saving bank details
type BankDetailsResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// UpdateBankDetails handles bank account details submission
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

	// Update bank details
	updates := map[string]interface{}{
		"bank_account_name":   req.BankAccountName,
		"bank_account_number": req.BankAccountNumber,
		"bank_code":           req.BankCode,
		"bank_country":        req.BankCountry,
	}

	if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update bank details")
		return
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

	bankDetails := map[string]interface{}{
		"bank_account_name":   organizer.BankAccountName,
		"bank_account_number": organizer.BankAccountNumber,
		"bank_code":           organizer.BankCode,
		"bank_country":        organizer.BankCountry,
	}

	json.NewEncoder(w).Encode(bankDetails)
}
