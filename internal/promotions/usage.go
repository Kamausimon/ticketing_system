package promotions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// RecordPromotionUsage handles recording a promotion usage after order completion
func (h *PromotionHandler) RecordPromotionUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		PromotionID    uint  `json:"promotion_id"`
		OrderID        uint  `json:"order_id"`
		DiscountAmount int64 `json:"discount_amount"` // in cents
		OriginalAmount int64 `json:"original_amount"` // in cents
		FinalAmount    int64 `json:"final_amount"`    // in cents
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify order belongs to user
	var order models.Order
	if err := h.db.Where("id = ? AND account_id = ?", req.OrderID, user.AccountID).First(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, req.PromotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check if already recorded
	var existingUsage models.PromotionUsage
	if err := h.db.Where("promotion_id = ? AND order_id = ?", req.PromotionID, req.OrderID).
		First(&existingUsage).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "promotion usage already recorded")
		return
	}

	// Record usage
	usage := models.PromotionUsage{
		PromotionID:    req.PromotionID,
		OrderID:        req.OrderID,
		AccountID:      user.AccountID,
		DiscountAmount: models.Money(req.DiscountAmount),
		OriginalAmount: models.Money(req.OriginalAmount),
		FinalAmount:    models.Money(req.FinalAmount),
		UsedAt:         time.Now(),
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save usage record
	if err := tx.Create(&usage).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to record usage")
		return
	}

	// Update promotion counters
	promotion.UsageCount++
	promotion.TotalRevenue += models.Money(req.FinalAmount)
	promotion.TotalDiscount += models.Money(req.DiscountAmount)
	promotion.LastUsageCheck = &usage.UsedAt

	// Track metrics
	if h.metrics != nil {
		h.metrics.TrackPromotionUsage(
			fmt.Sprintf("%d", req.PromotionID),
			promotion.Code,
			float64(req.DiscountAmount)/100.0,
			order.Currency,
		)
	}

	// Check if exhausted
	if !promotion.IsUnlimited && promotion.UsageLimit != nil {
		if promotion.UsageCount >= *promotion.UsageLimit {
			promotion.Status = models.PromotionExhausted
			promotion.PrecomputedActive = false
		}
	}

	if err := tx.Save(&promotion).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update promotion")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to record promotion usage")
		return
	}

	response := map[string]interface{}{
		"message":   "Promotion usage recorded successfully",
		"usage_id":  usage.ID,
		"promotion": convertToPromotionResponse(promotion),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPromotionUsage handles getting usage history for a promotion
func (h *PromotionHandler) GetPromotionUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion and verify ownership
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil && *promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get usage records
	var usages []models.PromotionUsage
	query := h.db.Preload("Promotion").Preload("Order").Preload("Account").
		Where("promotion_id = ?", promotionID).
		Order("used_at DESC").
		Offset((page - 1) * limit).
		Limit(limit)

	if err := query.Find(&usages).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch usage records")
		return
	}

	// Count total
	var totalCount int64
	h.db.Model(&models.PromotionUsage{}).Where("promotion_id = ?", promotionID).Count(&totalCount)

	// Convert to response
	usageResponses := make([]PromotionUsageResponse, len(usages))
	for i, usage := range usages {
		usageResponses[i] = PromotionUsageResponse{
			ID:             usage.ID,
			PromotionCode:  usage.Promotion.Code,
			OrderID:        usage.OrderID,
			AccountEmail:   usage.Account.Email,
			DiscountAmount: usage.DiscountAmount,
			OriginalAmount: usage.OriginalAmount,
			FinalAmount:    usage.FinalAmount,
			UsedAt:         usage.UsedAt,
		}
	}

	// Calculate total pages
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"usages":      usageResponses,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// GetPromotionStats handles getting promotion statistics
func (h *PromotionHandler) GetPromotionStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
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
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil && *promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get unique users count
	var uniqueUsers int64
	h.db.Model(&models.PromotionUsage{}).
		Where("promotion_id = ?", promotionID).
		Distinct("account_id").
		Count(&uniqueUsers)

	// Calculate average discount
	var avgDiscount float64
	if promotion.UsageCount > 0 {
		avgDiscount = float64(promotion.TotalDiscount) / float64(promotion.UsageCount)
	}

	// Calculate average order size
	var avgOrderSize models.Money
	if promotion.UsageCount > 0 {
		avgOrderSize = models.Money(int64(promotion.TotalRevenue) / int64(promotion.UsageCount))
	}

	// Calculate conversion rate (if we track views)
	conversionRate := 0.0
	if promotion.ConversionRate != nil {
		conversionRate = *promotion.ConversionRate
	}

	response := map[string]interface{}{
		"promotion_id":       promotion.ID,
		"code":               promotion.Code,
		"usage_count":        promotion.UsageCount,
		"unique_users":       uniqueUsers,
		"total_revenue":      promotion.TotalRevenue,
		"total_discount":     promotion.TotalDiscount,
		"average_discount":   avgDiscount,
		"average_order_size": avgOrderSize,
		"conversion_rate":    conversionRate,
		"status":             promotion.Status,
	}

	json.NewEncoder(w).Encode(response)
}

// RevokePromotionUsage handles revoking a promotion usage (e.g., after refund)
func (h *PromotionHandler) RevokePromotionUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		UsageID uint `json:"usage_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get usage record
	var usage models.PromotionUsage
	if err := h.db.Preload("Promotion").First(&usage, req.UsageID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "usage record not found")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check authorization
	if usage.Promotion.OrganizerID != nil && *usage.Promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update promotion counters
	var promotion models.Promotion
	if err := tx.First(&promotion, usage.PromotionID).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	if promotion.UsageCount > 0 {
		promotion.UsageCount--
	}
	promotion.TotalRevenue -= usage.FinalAmount
	promotion.TotalDiscount -= usage.DiscountAmount

	// If was exhausted, reactivate
	if promotion.Status == models.PromotionExhausted {
		promotion.Status = models.PromotionActive
		promotion.PrecomputedActive = true
	}

	if err := tx.Save(&promotion).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update promotion")
		return
	}

	// Delete usage record
	if err := tx.Delete(&usage).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to revoke usage")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to revoke promotion usage")
		return
	}

	response := map[string]interface{}{
		"message": "Promotion usage revoked successfully",
	}

	json.NewEncoder(w).Encode(response)
}
