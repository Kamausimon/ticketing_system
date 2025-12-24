package promotions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// CreatePromotion handles creating a new promotion
func (h *PromotionHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req CreatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateCreatePromotionRequest(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// If event-specific, verify ownership
	if req.EventID != nil {
		var event models.Event
		if err := h.db.Where("id = ? AND account_id = ?", *req.EventID, user.AccountID).First(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied to event")
			return
		}
	}

	// Generate code if requested
	code := req.Code
	if req.GenerateCode || code == "" {
		code = generatePromotionCode()
	} else {
		code = strings.ToUpper(strings.TrimSpace(code))
	}

	// Check if code already exists
	var existingPromo models.Promotion
	if err := h.db.Where("code = ?", code).First(&existingPromo).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "promotion code already exists")
		return
	}

	// Get or determine organizer_id
	organizerID := req.OrganizerID
	if organizerID == nil {
		// Look up organizer record for this user's account
		var organizer models.Organizer
		if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "organizer profile not found for this account")
			return
		}
		organizerID = &organizer.ID
	}

	// Verify organizer exists and matches user's account
	var organizer models.Organizer
	if err := h.db.Where("id = ? AND account_id = ?", *organizerID, user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "invalid organizer")
		return
	}

	// Create promotion
	promotion := models.Promotion{
		Code:               code,
		Name:               req.Name,
		Description:        req.Description,
		Type:               req.Type,
		Status:             models.PromotionDraft,
		Target:             req.Target,
		DiscountPercentage: req.DiscountPercentage,
		EventID:            req.EventID,
		OrganizerID:        organizerID,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		EarlyBirdCutoff:    req.EarlyBirdCutoff,
		UsageLimit:         req.UsageLimit,
		PerUserLimit:       req.PerUserLimit,
		PerOrderLimit:      req.PerOrderLimit,
		IsPublic:           req.IsPublic,
		FirstTimeCustomers: req.FirstTimeCustomers,
		CreatedBy:          userID,
	}

	// Convert amounts from cents to Money type
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

	if req.FreeQuantity != nil {
		promotion.FreeQuantity = req.FreeQuantity
	}

	// Handle ticket class IDs (store as JSON)
	if len(req.TicketClassIDs) > 0 {
		ticketClassJSON, err := json.Marshal(req.TicketClassIDs)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to process ticket class IDs")
			return
		}
		promotion.TicketClassIDs = string(ticketClassJSON)
	}

	// Handle event categories (store as JSON)
	if len(req.EventCategories) > 0 {
		categoriesJSON, err := json.Marshal(req.EventCategories)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to process event categories")
			return
		}
		promotion.EventCategories = string(categoriesJSON)
	}

	// Set unlimited flag
	if req.UsageLimit == nil {
		promotion.IsUnlimited = true
	}

	// Save to database
	if err := h.db.Create(&promotion).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create promotion")
		return
	}

	// Load relationships for response
	h.db.Preload("Event").Preload("Organizer").First(&promotion, promotion.ID)

	response := convertToPromotionResponse(promotion)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ClonePromotion handles cloning an existing promotion
func (h *PromotionHandler) ClonePromotion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		PromotionID uint   `json:"promotion_id"`
		NewCode     string `json:"new_code,omitempty"`
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

	// Get original promotion
	var original models.Promotion
	if err := h.db.First(&original, req.PromotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Verify ownership if organizer-specific
	if original.OrganizerID != nil && *original.OrganizerID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Generate new code
	newCode := req.NewCode
	if newCode == "" {
		newCode = generatePromotionCode()
	} else {
		newCode = strings.ToUpper(strings.TrimSpace(newCode))
	}

	// Check if new code already exists
	var existingPromo models.Promotion
	if err := h.db.Where("code = ?", newCode).First(&existingPromo).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "new code already exists")
		return
	}

	// Clone promotion
	cloned := models.Promotion{
		Code:               newCode,
		Name:               original.Name + " (Copy)",
		Description:        original.Description,
		Type:               original.Type,
		Status:             models.PromotionDraft,
		Target:             original.Target,
		DiscountPercentage: original.DiscountPercentage,
		DiscountAmount:     original.DiscountAmount,
		FreeQuantity:       original.FreeQuantity,
		MinimumPurchase:    original.MinimumPurchase,
		MaximumDiscount:    original.MaximumDiscount,
		EventID:            original.EventID,
		OrganizerID:        original.OrganizerID,
		StartDate:          time.Now(),
		EndDate:            time.Now().AddDate(0, 1, 0),
		UsageLimit:         original.UsageLimit,
		PerUserLimit:       original.PerUserLimit,
		PerOrderLimit:      original.PerOrderLimit,
		IsUnlimited:        original.IsUnlimited,
		IsPublic:           original.IsPublic,
		FirstTimeCustomers: original.FirstTimeCustomers,
		CreatedBy:          userID,
	}

	// Save cloned promotion
	if err := h.db.Create(&cloned).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to clone promotion")
		return
	}

	// Load relationships
	h.db.Preload("Event").Preload("Organizer").First(&cloned, cloned.ID)

	response := convertToPromotionResponse(cloned)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// validateCreatePromotionRequest validates the create promotion request
func validateCreatePromotionRequest(req *CreatePromotionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Description == "" {
		return fmt.Errorf("description is required")
	}

	if req.Type == "" {
		return fmt.Errorf("type is required")
	}

	if req.Target == "" {
		return fmt.Errorf("target is required")
	}

	// Event ID is always required
	if req.EventID == nil {
		return fmt.Errorf("event_id is required")
	}

	// Validate target-specific requirements
	switch req.Target {
	case models.TargetSpecificTicket:
		if len(req.TicketClassIDs) == 0 {
			return fmt.Errorf("ticket_class_ids is required when target is 'specific_ticket'")
		}
	case models.TargetCategory:
		if len(req.EventCategories) == 0 {
			return fmt.Errorf("event_categories is required when target is 'category'")
		}
	}

	// Validate discount configuration based on type
	switch req.Type {
	case models.PromotionPercentage, models.PromotionEarlyBird, models.PromotionBulk:
		if req.DiscountPercentage == nil {
			return fmt.Errorf("discount_percentage is required for this type")
		}
		if *req.DiscountPercentage < 1 || *req.DiscountPercentage > 100 {
			return fmt.Errorf("discount_percentage must be between 1 and 100")
		}
	case models.PromotionFixedAmount:
		if req.DiscountAmount == nil {
			return fmt.Errorf("discount_amount is required for fixed amount type")
		}
		if *req.DiscountAmount <= 0 {
			return fmt.Errorf("discount_amount must be greater than 0")
		}
	case models.PromotionFreeTickets:
		if req.FreeQuantity == nil {
			return fmt.Errorf("free_quantity is required for free tickets type")
		}
		if *req.FreeQuantity <= 0 {
			return fmt.Errorf("free_quantity must be greater than 0")
		}
	}

	// Validate dates
	if req.StartDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}
	if req.EndDate.IsZero() {
		return fmt.Errorf("end_date is required")
	}
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}

	return nil
}
