package promotions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetPromotionDetails handles getting detailed information about a specific promotion
func (h *PromotionHandler) GetPromotionDetails(w http.ResponseWriter, r *http.Request) {
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

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get promotion with relationships
	var promotion models.Promotion
	if err := h.db.Preload("Event").Preload("Organizer").
		Where("id = ?", promotionID).First(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Get user for authorization check
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if user can view this promotion
	// If promotion is not public and user is not the owner, deny access
	if !promotion.IsPublic {
		if promotion.OrganizerID != nil {
			// Check if organizer's account_id matches user's account_id
			var organizer models.Organizer
			if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
				middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
				return
			}
		}
	}

	response := convertToPromotionResponse(promotion)
	json.NewEncoder(w).Encode(response)
}

// GetPromotionByCode handles getting a promotion by its code
func (h *PromotionHandler) GetPromotionByCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get code from URL
	vars := mux.Vars(r)
	code := vars["code"]
	if code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.Preload("Event").Preload("Organizer").
		Where("UPPER(code) = UPPER(?)", code).First(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Only return public promotions or require auth for private ones
	if !promotion.IsPublic {
		userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
		if userID == 0 {
			middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var user models.User
		if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
			return
		}

		if promotion.OrganizerID != nil {
			// Check if organizer's account_id matches user's account_id
			var organizer models.Organizer
			if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
				middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
				return
			}
		}
	}

	response := convertToPromotionResponse(promotion)
	json.NewEncoder(w).Encode(response)
}

// GetPromotionUsageDetails handles getting detailed usage information
func (h *PromotionHandler) GetPromotionUsageDetails(w http.ResponseWriter, r *http.Request) {
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
	if promotion.OrganizerID != nil {
		// Check if organizer's account_id matches user's account_id
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
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

// DeletePromotion handles deleting a promotion (soft delete)
func (h *PromotionHandler) DeletePromotion(w http.ResponseWriter, r *http.Request) {
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

	// Verify ownership
	if promotion.OrganizerID != nil {
		// Check if organizer's account_id matches user's account_id
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Can't delete if it has been used
	var usageCount int64
	h.db.Model(&models.PromotionUsage{}).Where("promotion_id = ?", promotionID).Count(&usageCount)
	if usageCount > 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "cannot delete promotion that has been used. Consider deactivating instead")
		return
	}

	// Soft delete
	if err := h.db.Delete(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to delete promotion")
		return
	}

	response := map[string]interface{}{
		"message": "Promotion deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}
