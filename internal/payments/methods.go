package payments

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// SavePaymentMethod saves a customer payment method
func (h *PaymentHandler) SavePaymentMethod(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID   uint    `json:"account_id"`
		Type        string  `json:"type"` // "card", "mpesa"
		DisplayName string  `json:"display_name"`
		PhoneNumber *string `json:"phone_number,omitempty"`
		Last4       *string `json:"last4,omitempty"`
		ExpiryMonth *int    `json:"expiry_month,omitempty"`
		ExpiryYear  *int    `json:"expiry_year,omitempty"`
		ExternalID  *string `json:"external_id,omitempty"` // Intasend/Stripe ID
		IsDefault   bool    `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If setting as default, unset other defaults
	if req.IsDefault {
		h.DB.Model(&models.PaymentMethod{}).
			Where("account_id = ? AND is_default = true", req.AccountID).
			Update("is_default", false)
	}

	paymentMethod := models.PaymentMethod{
		AccountID:               req.AccountID,
		Type:                    models.PaymentMethodType(req.Type),
		Status:                  models.PaymentMethodActive,
		DisplayName:             req.DisplayName,
		IsDefault:               req.IsDefault,
		CardLast4:               req.Last4,
		CardExpiryMonth:         req.ExpiryMonth,
		CardExpiryYear:          req.ExpiryYear,
		MpesaPhoneNumber:        req.PhoneNumber,
		ExternalPaymentMethodID: req.ExternalID,
		IsVerified:              true,
	}

	if err := h.DB.Create(&paymentMethod).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save payment method")
		return
	}

	writeJSON(w, http.StatusCreated, convertToPaymentMethodResponse(&paymentMethod))
}

// GetPaymentMethods returns customer's saved payment methods
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	accountIDStr := r.URL.Query().Get("account_id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var methods []models.PaymentMethod
	if err := h.DB.Where("account_id = ? AND deleted_at IS NULL", accountID).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch payment methods")
		return
	}

	var responses []PaymentMethodResponse
	for _, m := range methods {
		responses = append(responses, convertToPaymentMethodResponse(&m))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"payment_methods": responses,
		"total":           len(responses),
	})
}

// DeletePaymentMethod soft deletes a payment method
func (h *PaymentHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	methodID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	if err := h.DB.Delete(&models.PaymentMethod{}, methodID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete payment method")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// SetDefaultPaymentMethod sets a payment method as default
func (h *PaymentHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	methodID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	var method models.PaymentMethod
	if err := h.DB.First(&method, methodID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Payment method not found")
		return
	}

	// Unset other defaults for this account
	h.DB.Model(&models.PaymentMethod{}).
		Where("account_id = ? AND id != ?", method.AccountID, methodID).
		Update("is_default", false)

	// Set this as default
	method.IsDefault = true
	if err := h.DB.Save(&method).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update default")
		return
	}

	writeJSON(w, http.StatusOK, convertToPaymentMethodResponse(&method))
}

// UpdatePaymentMethodExpiry updates card expiry date
func (h *PaymentHandler) UpdatePaymentMethodExpiry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	methodID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	var req struct {
		ExpiryMonth int `json:"expiry_month"`
		ExpiryYear  int `json:"expiry_year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var method models.PaymentMethod
	if err := h.DB.First(&method, methodID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Payment method not found")
		return
	}

	method.CardExpiryMonth = &req.ExpiryMonth
	method.CardExpiryYear = &req.ExpiryYear
	method.Status = models.PaymentMethodActive

	now := time.Now()
	method.VerifiedAt = &now

	if err := h.DB.Save(&method).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update expiry")
		return
	}

	writeJSON(w, http.StatusOK, convertToPaymentMethodResponse(&method))
}
