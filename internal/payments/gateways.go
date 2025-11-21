package payments

import (
	"net/http"
	"ticketing_system/internal/models"
)

// GetAvailableGateways returns list of configured payment gateways
func (h *PaymentHandler) GetAvailableGateways(w http.ResponseWriter, r *http.Request) {
	var gateways []models.PaymentGateway
	if err := h.DB.Find(&gateways).Error; err != nil {
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
	// TODO: Implement when organizer needs to configure their own payment gateway
	// This would allow organizers to use their own Intasend/Stripe accounts
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Gateway configuration not yet implemented",
	})
}
