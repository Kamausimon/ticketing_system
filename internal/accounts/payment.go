package accounts

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// SetupStripeIntegration handles setting up Stripe payment integration
func (h *AccountHandler) SetupStripeIntegration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req StripeIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.StripeSecretKey == "" || req.StripePublishableKey == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "stripe secret key and publishable key are required")
		return
	}

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if user is an organizer
	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can set up payment integration")
		return
	}

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Update Stripe credentials
	account.StripeAccessToken = &req.StripeAccessToken
	account.StripeRefreshToken = &req.StripeRefreshToken
	account.StripeSecretKey = &req.StripeSecretKey
	account.StripePublishableKey = &req.StripePublishableKey

	// Set up default payment gateway if not exists
	if account.PaymentGatewayID == nil || *account.PaymentGatewayID == 0 {
		// Find or create Stripe payment gateway
		var gateway models.PaymentGateway
		if err := h.db.Where("name = ?", "Stripe").First(&gateway).Error; err != nil {
			// Create Stripe gateway if not exists
			gateway = models.PaymentGateway{
				ProviderName: "Stripe",
				ProviderURL:  "https://api.stripe.com",
				IsOnSite:     false,
				CanRefund:    true,
				Name:         "Stripe",
			}
			h.db.Create(&gateway)
		}

		account.PaymentGatewayID = &gateway.ID
	}

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to setup stripe integration")
		return
	}

	// Log activity
	h.logAccountActivity(account.ID, "stripe_setup", "Stripe payment integration configured", getClientIP(r))

	response := map[string]interface{}{
		"message": "Stripe integration setup successfully",
		"gateway": map[string]interface{}{
			"provider":   "Stripe",
			"is_active":  true,
			"can_refund": true,
			"setup_date": account.UpdatedAt,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetPaymentMethods handles getting user's payment methods
func (h *AccountHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
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

	// For now, return mock payment methods since we don't have a payment_methods table
	// In production, you would query actual payment methods from Stripe/database
	paymentMethods := []PaymentMethod{}

	// Get account to check if Stripe is configured
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err == nil {
		if account.StripeSecretKey != nil {
			// Mock Stripe payment method
			paymentMethods = append(paymentMethods, PaymentMethod{
				ID:           1,
				Type:         "stripe",
				Last4:        "4242",
				Brand:        "visa",
				ExpiryMonth:  12,
				ExpiryYear:   2025,
				IsDefault:    true,
				StripeCardID: "card_mock_stripe_id",
			})
		}
	}

	response := map[string]interface{}{
		"payment_methods": paymentMethods,
		"count":           len(paymentMethods),
	}

	json.NewEncoder(w).Encode(response)
}

// GetPaymentGatewaySettings handles getting payment gateway settings
func (h *AccountHandler) GetPaymentGatewaySettings(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is an organizer
	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can view payment gateway settings")
		return
	}

	// Get account with payment gateway
	var account models.Account
	if err := h.db.Preload("PaymentGateway").Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Prepare settings response
	settings := map[string]interface{}{
		"stripe_configured": account.StripeSecretKey != nil && account.StripePublishableKey != nil,
		"gateway_name":      "",
		"gateway_active":    false,
		"can_refund":        false,
		"setup_date":        nil,
	}

	if account.PaymentGateway != nil && account.PaymentGateway.ID > 0 {
		settings["gateway_name"] = account.PaymentGateway.Name
		settings["gateway_active"] = true
		settings["can_refund"] = account.PaymentGateway.CanRefund
		settings["setup_date"] = account.UpdatedAt
	}

	// Mask sensitive information
	if account.StripePublishableKey != nil {
		publishableKey := *account.StripePublishableKey
		if len(publishableKey) > 8 {
			masked := publishableKey[:8] + strings.Repeat("*", len(publishableKey)-8)
			settings["stripe_publishable_key_masked"] = masked
		}
	}

	json.NewEncoder(w).Encode(settings)
}

// StripeConnectRequest represents Stripe Connect setup request
type StripeConnectRequest struct {
	BusinessType  string `json:"business_type"`
	BusinessName  string `json:"business_name"`
	BusinessTaxID string `json:"business_tax_id"`
	ReturnURL     string `json:"return_url"`
	RefreshURL    string `json:"refresh_url"`
}

// PaymentGatewayResponse represents payment gateway information
type PaymentGatewayResponse struct {
	ID                   uint   `json:"id"`
	Name                 string `json:"name"`
	IsActive             bool   `json:"is_active"`
	HasStripeConnect     bool   `json:"has_stripe_connect"`
	StripeAccountStatus  string `json:"stripe_account_status"`
	CanReceivePayments   bool   `json:"can_receive_payments"`
	RequiresVerification bool   `json:"requires_verification"`
}

// GetPaymentGatewayInfo handles getting payment gateway information
func (h *AccountHandler) GetPaymentGatewayInfo(w http.ResponseWriter, r *http.Request) {
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

	// Get account with payment gateway
	var account models.Account
	if err := h.db.Preload("PaymentGateway").Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Build payment gateway response
	paymentGateway := PaymentGatewayResponse{
		ID:                   getGatewayID(account.PaymentGateway),
		Name:                 getGatewayName(account.PaymentGateway),
		IsActive:             true, // Default since field doesn't exist
		HasStripeConnect:     account.StripeAccessToken != nil,
		StripeAccountStatus:  getStripeAccountStatus(&account),
		CanReceivePayments:   canReceivePayments(&account),
		RequiresVerification: requiresVerification(&account),
	}

	json.NewEncoder(w).Encode(paymentGateway)
}

// SetupStripeConnect handles Stripe Connect onboarding
func (h *AccountHandler) SetupStripeConnect(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is an organizer
	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can setup payment processing")
		return
	}

	// Parse request
	var req StripeConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: Implement actual Stripe Connect setup
	// This would involve:
	// 1. Creating a Stripe Connect account
	// 2. Generating onboarding link
	// 3. Storing account credentials securely

	response := map[string]interface{}{
		"message":         "Stripe Connect setup initiated",
		"onboarding_url":  "https://connect.stripe.com/setup/...", // Mock URL
		"account_id":      "acct_placeholder",                     // Mock account ID
		"return_url":      req.ReturnURL,
		"refresh_url":     req.RefreshURL,
		"setup_completed": false,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CompleteStripeSetup handles Stripe Connect setup completion
func (h *AccountHandler) CompleteStripeSetup(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Implement actual Stripe Connect completion
	// This would involve:
	// 1. Verifying the Stripe account setup
	// 2. Storing access tokens securely
	// 3. Updating account status

	response := map[string]interface{}{
		"message":        "Stripe Connect setup completed",
		"account_status": "active",
		"can_receive":    true,
	}

	json.NewEncoder(w).Encode(response)
}

// DisconnectStripe handles disconnecting Stripe account
func (h *AccountHandler) DisconnectStripe(w http.ResponseWriter, r *http.Request) {
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

	// Clear Stripe credentials
	account.StripeAccessToken = nil
	account.StripeRefreshToken = nil
	account.StripeSecretKey = nil
	account.StripePublishableKey = nil
	account.StripeDataRaw = nil

	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to disconnect Stripe")
		return
	}

	response := map[string]interface{}{
		"message": "Stripe account disconnected successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getStripeAccountStatus(account *models.Account) string {
	if account.StripeAccessToken == nil {
		return "not_connected"
	}
	// In a real implementation, this would check actual Stripe account status
	return "active"
}

func canReceivePayments(account *models.Account) bool {
	return account.StripeAccessToken != nil && account.StripeSecretKey != nil
}

func requiresVerification(account *models.Account) bool {
	// In a real implementation, this would check Stripe account verification requirements
	return account.StripeAccessToken != nil && account.StripeSecretKey == nil
}

func getGatewayID(gateway *models.PaymentGateway) uint {
	if gateway == nil {
		return 0
	}
	return gateway.ID
}

func getGatewayName(gateway *models.PaymentGateway) string {
	if gateway == nil {
		return ""
	}
	return gateway.Name
}
