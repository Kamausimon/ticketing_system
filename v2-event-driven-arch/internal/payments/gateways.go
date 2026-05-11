package payments

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetAvailableGateways returns list of configured payment gateways
func (h *PaymentHandler) GetAvailableGateways(w http.ResponseWriter, r *http.Request) {
	var gateways []models.PaymentGateway
	if err := h.db.Find(&gateways).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch payment gateways")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"gateways": gateways,
		"total":    len(gateways),
	})
}

// GetGatewayConfig returns configuration for a specific gateway (organizer only)
func (h *PaymentHandler) GetGatewayConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gatewayID := vars["gateway_id"]

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and account
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get organizer to verify permissions
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		writeError(w, http.StatusForbidden, "only organizers can configure payment gateways")
		return
	}

	// Get gateway configuration for this account
	var accountGateway models.AccountPaymentGateway
	if err := h.db.Preload("PaymentGateway").
		Where("account_id = ? AND payment_gateway_id = ?", user.AccountID, gatewayID).
		First(&accountGateway).Error; err != nil {
		writeError(w, http.StatusNotFound, "gateway configuration not found")
		return
	}

	// Return sanitized configuration (don't expose sensitive keys)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"gateway_id":   accountGateway.PaymentGatewayID,
		"gateway_name": accountGateway.PaymentGateway.Name,
		"provider":     accountGateway.PaymentGateway.ProviderName,
		"configured":   accountGateway.Config != "",
		"message":      "Gateway configuration retrieved successfully",
	})
}

// UpdateGatewayConfig updates or creates gateway configuration for an organizer
func (h *PaymentHandler) UpdateGatewayConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gatewayID := vars["gateway_id"]

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and account
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify organizer permissions
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		writeError(w, http.StatusForbidden, "only organizers can configure payment gateways")
		return
	}

	// Parse request body
	var req struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if gateway exists
	var gateway models.PaymentGateway
	if err := h.db.First(&gateway, gatewayID).Error; err != nil {
		writeError(w, http.StatusNotFound, "payment gateway not found")
		return
	}

	// Update or create account gateway configuration
	var accountGateway models.AccountPaymentGateway
	result := h.db.Where("account_id = ? AND payment_gateway_id = ?", user.AccountID, gatewayID).
		First(&accountGateway)

	if result.Error != nil {
		// Create new configuration
		accountGateway = models.AccountPaymentGateway{
			AccountID:        user.AccountID,
			PaymentGatewayID: gateway.ID,
			Config:           req.Config,
		}
		if err := h.db.Create(&accountGateway).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create gateway configuration")
			return
		}
	} else {
		// Update existing configuration
		if err := h.db.Model(&accountGateway).Update("config", req.Config).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update gateway configuration")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Gateway configuration updated successfully",
	})
}
