package organizers

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// PaymentGatewayConfigRequest represents payment gateway configuration
type PaymentGatewayConfigRequest struct {
	PaymentGatewayID uint   `json:"payment_gateway_id"`
	Config           string `json:"config"` // JSON string containing gateway-specific config
}

// PaymentGatewayConfigResponse represents the response after saving payment config
type PaymentGatewayConfigResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ConfigurePaymentGateway handles payment gateway setup
func (h *OrganizerHandler) ConfigurePaymentGateway(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Parse request
	var req PaymentGatewayConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.PaymentGatewayID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "payment_gateway_id is required")
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

	// Verify payment gateway exists
	var paymentGateway models.PaymentGateway
	if err := h.db.Where("id = ?", req.PaymentGatewayID).First(&paymentGateway).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "payment gateway not found")
		return
	}

	// Check if account already has this payment gateway configured
	var existingConfig models.AccountPaymentGateway
	existingConfigErr := h.db.Where("account_id = ? AND payment_gateway_id = ?", user.AccountID, req.PaymentGatewayID).First(&existingConfig).Error

	if existingConfigErr == nil {
		// Update existing configuration
		if err := h.db.Model(&existingConfig).Update("config", req.Config).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update payment gateway configuration")
			return
		}
	} else {
		// Create new configuration
		newConfig := models.AccountPaymentGateway{
			AccountID:        user.AccountID,
			PaymentGatewayID: req.PaymentGatewayID,
			Config:           req.Config,
		}
		if err := h.db.Create(&newConfig).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to save payment gateway configuration")
			return
		}
	}

	// Update organizer to mark payment as configured
	updates := map[string]interface{}{
		"payment_gateway_id":    req.PaymentGatewayID,
		"is_payment_configured": true,
	}

	if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update organizer payment status")
		return
	}

	response := PaymentGatewayConfigResponse{
		Message: "Payment gateway configured successfully",
		Status:  "success",
	}

	json.NewEncoder(w).Encode(response)
}

// GetPaymentGatewayConfig retrieves payment gateway configuration
func (h *OrganizerHandler) GetPaymentGatewayConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get payment gateway configurations
	var configs []models.AccountPaymentGateway
	if err := h.db.Where("account_id = ?", user.AccountID).Preload("PaymentGateway").Find(&configs).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve payment configurations")
		return
	}

	response := map[string]interface{}{
		"is_configured":      organizer.IsPaymentConfigured,
		"current_gateway_id": organizer.PaymentGatewayID,
		"configurations":     configs,
	}

	json.NewEncoder(w).Encode(response)
}
