package promotions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// UpdatePromotion handles updating an existing promotion
func (h *PromotionHandler) UpdatePromotion(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	var req UpdatePromotionRequest
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

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Verify ownership
	if promotion.OrganizerID != nil && *promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Can't update if promotion is active and has usage
	if promotion.Status == models.PromotionActive && promotion.UsageCount > 0 {
		// Only allow limited updates
		if req.DiscountPercentage != nil || req.DiscountAmount != nil {
			middleware.WriteJSONError(w, http.StatusBadRequest, "cannot change discount values for active promotion with usage")
			return
		}
	}

	// Apply updates
	if req.Name != nil {
		promotion.Name = *req.Name
	}
	if req.Description != nil {
		promotion.Description = *req.Description
	}
	if req.DiscountPercentage != nil {
		promotion.DiscountPercentage = req.DiscountPercentage
	}
	if req.DiscountAmount != nil {
		amount := models.Money(*req.DiscountAmount)
		promotion.DiscountAmount = &amount
	}
	if req.MinimumPurchase != nil {
		amount := models.Money(*req.MinimumPurchase)
		promotion.MinimumPurchase = &amount
	}
	if req.MaximumDiscount != nil {
		amount := models.Money(*req.MaximumDiscount)
		promotion.MaximumDiscount = &amount
	}
	if req.EndDate != nil {
		// Can't set end date before start date
		if req.EndDate.Before(promotion.StartDate) {
			middleware.WriteJSONError(w, http.StatusBadRequest, "end date cannot be before start date")
			return
		}
		promotion.EndDate = *req.EndDate
	}
	if req.UsageLimit != nil {
		// Can't set usage limit below current usage
		if *req.UsageLimit < promotion.UsageCount {
			middleware.WriteJSONError(w, http.StatusBadRequest, "usage limit cannot be less than current usage count")
			return
		}
		promotion.UsageLimit = req.UsageLimit
		promotion.IsUnlimited = false
	}
	if req.PerUserLimit != nil {
		promotion.PerUserLimit = req.PerUserLimit
	}
	if req.IsPublic != nil {
		promotion.IsPublic = *req.IsPublic
	}

	// Save updates
	if err := h.db.Save(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update promotion")
		return
	}

	// Load relationships
	h.db.Preload("Event").Preload("Organizer").First(&promotion, promotion.ID)

	response := convertToPromotionResponse(promotion)
	json.NewEncoder(w).Encode(response)
}

// ActivatePromotion handles activating a promotion
func (h *PromotionHandler) ActivatePromotion(w http.ResponseWriter, r *http.Request) {
	h.updatePromotionStatus(w, r, models.PromotionActive)
}

// PausePromotion handles pausing a promotion
func (h *PromotionHandler) PausePromotion(w http.ResponseWriter, r *http.Request) {
	h.updatePromotionStatus(w, r, models.PromotionPaused)
}

// DeactivatePromotion handles deactivating a promotion
func (h *PromotionHandler) DeactivatePromotion(w http.ResponseWriter, r *http.Request) {
	h.updatePromotionStatus(w, r, models.PromotionCancelled)
}

// updatePromotionStatus is a helper to update promotion status
func (h *PromotionHandler) updatePromotionStatus(w http.ResponseWriter, r *http.Request, newStatus models.PromotionStatus) {
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

	// Verify ownership
	if promotion.OrganizerID != nil && *promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Validate status transition
	if !isValidStatusTransition(promotion.Status, newStatus) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid status transition")
		return
	}

	// Update status
	promotion.Status = newStatus

	// Update precomputed active flag
	if newStatus == models.PromotionActive {
		promotion.PrecomputedActive = true
	} else {
		promotion.PrecomputedActive = false
	}

	if err := h.db.Save(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update promotion status")
		return
	}

	// Load relationships
	h.db.Preload("Event").Preload("Organizer").First(&promotion, promotion.ID)

	response := convertToPromotionResponse(promotion)
	json.NewEncoder(w).Encode(response)
}

// ExtendPromotionDate handles extending the end date of a promotion
func (h *PromotionHandler) ExtendPromotionDate(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	var req struct {
		NewEndDate time.Time `json:"new_end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewEndDate.IsZero() {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_end_date is required")
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
	if promotion.OrganizerID != nil && *promotion.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Validate new end date
	if req.NewEndDate.Before(promotion.StartDate) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new end date cannot be before start date")
		return
	}

	if req.NewEndDate.Before(time.Now()) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new end date cannot be in the past")
		return
	}

	// Update end date
	promotion.EndDate = req.NewEndDate

	// If promotion was expired, reactivate it
	if promotion.Status == models.PromotionExpired {
		promotion.Status = models.PromotionActive
		promotion.PrecomputedActive = true
	}

	if err := h.db.Save(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to extend promotion")
		return
	}

	// Load relationships
	h.db.Preload("Event").Preload("Organizer").First(&promotion, promotion.ID)

	response := convertToPromotionResponse(promotion)
	json.NewEncoder(w).Encode(response)
}

// Helper function to validate status transitions
func isValidStatusTransition(from, to models.PromotionStatus) bool {
	validTransitions := map[models.PromotionStatus][]models.PromotionStatus{
		models.PromotionDraft: {
			models.PromotionActive,
			models.PromotionCancelled,
		},
		models.PromotionActive: {
			models.PromotionPaused,
			models.PromotionExpired,
			models.PromotionExhausted,
			models.PromotionCancelled,
		},
		models.PromotionPaused: {
			models.PromotionActive,
			models.PromotionCancelled,
		},
		models.PromotionExpired: {
			models.PromotionActive, // Can reactivate by extending date
			models.PromotionCancelled,
		},
	}

	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}
