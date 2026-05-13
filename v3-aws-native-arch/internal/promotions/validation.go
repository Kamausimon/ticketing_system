package promotions

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// ValidatePromotionCode handles validating a promotion code during checkout
func (h *PromotionHandler) ValidatePromotionCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req ValidatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "code is required")
		return
	}

	if req.OrderAmount <= 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "order_amount is required")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.Preload("Event").Where("UPPER(code) = UPPER(?)", req.Code).First(&promotion).Error; err != nil {
		response := ValidatePromotionResponse{
			Valid:       false,
			Code:        req.Code,
			Message:     "Invalid promotion code",
			ErrorReason: "not_found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate promotion
	valid, errorReason, message := h.validatePromotion(&promotion, &req)
	if !valid {
		response := ValidatePromotionResponse{
			Valid:       false,
			Code:        req.Code,
			Message:     message,
			ErrorReason: errorReason,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate discount
	orderAmount := models.Money(req.OrderAmount)
	discountAmount := calculatePromotionDiscount(&promotion, orderAmount)
	finalAmount := orderAmount - discountAmount

	// Ensure final amount is not negative
	if finalAmount < 0 {
		finalAmount = 0
	}

	response := ValidatePromotionResponse{
		Valid:          true,
		PromotionID:    promotion.ID,
		Code:           promotion.Code,
		DiscountAmount: int64(discountAmount),
		FinalAmount:    int64(finalAmount),
		Message:        "Promotion code is valid",
	}

	json.NewEncoder(w).Encode(response)
}

// validatePromotion performs comprehensive validation checks
func (h *PromotionHandler) validatePromotion(promotion *models.Promotion, req *ValidatePromotionRequest) (bool, string, string) {
	now := time.Now()

	// Check status
	if promotion.Status != models.PromotionActive {
		return false, "inactive", "Promotion code is not active"
	}

	// Check dates
	if now.Before(promotion.StartDate) {
		return false, "not_started", "Promotion has not started yet"
	}
	if now.After(promotion.EndDate) {
		return false, "expired", "Promotion code has expired"
	}

	// Check early bird cutoff
	if promotion.Type == models.PromotionEarlyBird && promotion.EarlyBirdCutoff != nil {
		if now.After(*promotion.EarlyBirdCutoff) {
			return false, "early_bird_expired", "Early bird promotion period has ended"
		}
	}

	// Check usage limit
	if !promotion.IsUnlimited && promotion.UsageLimit != nil {
		if promotion.UsageCount >= *promotion.UsageLimit {
			return false, "exhausted", "Promotion code has reached its usage limit"
		}
	}

	// Check minimum purchase
	if promotion.MinimumPurchase != nil {
		if models.Money(req.OrderAmount) < *promotion.MinimumPurchase {
			return false, "minimum_not_met", "Order amount does not meet minimum purchase requirement"
		}
	}

	// Check event targeting
	if promotion.Target == models.TargetEvent && promotion.EventID != nil {
		if req.EventID == nil || *req.EventID != *promotion.EventID {
			return false, "wrong_event", "Promotion code is not valid for this event"
		}
	}

	// Check per-user limit (if account ID provided)
	if req.AccountID > 0 && promotion.PerUserLimit != nil {
		var userUsageCount int64
		h.db.Model(&models.PromotionUsage{}).
			Where("promotion_id = ? AND account_id = ?", promotion.ID, req.AccountID).
			Count(&userUsageCount)

		if userUsageCount >= int64(*promotion.PerUserLimit) {
			return false, "user_limit_exceeded", "You have reached the maximum usage limit for this promotion"
		}
	}

	// Check first-time customer restriction
	if promotion.FirstTimeCustomers && req.AccountID > 0 {
		var previousOrdersCount int64
		h.db.Model(&models.Order{}).
			Where("account_id = ? AND status IN ?", req.AccountID,
				[]models.OrderStatus{models.OrderPaid, models.OrderFulfilled}).
			Count(&previousOrdersCount)

		if previousOrdersCount > 0 {
			return false, "not_first_time", "This promotion is only valid for first-time customers"
		}
	}

	return true, "", ""
}

// CheckPromotionEligibility checks if a user is eligible for a promotion
func (h *PromotionHandler) CheckPromotionEligibility(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		Code    string `json:"code"`
		EventID *uint  `json:"event_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.Where("UPPER(code) = UPPER(?)", req.Code).First(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Build validation request
	validateReq := ValidatePromotionRequest{
		Code:        req.Code,
		EventID:     req.EventID,
		OrderAmount: 1000, // Dummy amount for eligibility check
		AccountID:   user.AccountID,
	}

	// Validate
	valid, errorReason, message := h.validatePromotion(&promotion, &validateReq)

	response := map[string]interface{}{
		"eligible":     valid,
		"code":         req.Code,
		"message":      message,
		"error_reason": errorReason,
		"usage_count":  promotion.UsageCount,
		"usage_limit":  promotion.UsageLimit,
		"is_unlimited": promotion.IsUnlimited,
	}

	// If eligible, include usage info
	if valid {
		var userUsageCount int64
		h.db.Model(&models.PromotionUsage{}).
			Where("promotion_id = ? AND account_id = ?", promotion.ID, user.AccountID).
			Count(&userUsageCount)

		response["user_usage_count"] = userUsageCount
		response["user_usage_limit"] = promotion.PerUserLimit
	}

	json.NewEncoder(w).Encode(response)
}
