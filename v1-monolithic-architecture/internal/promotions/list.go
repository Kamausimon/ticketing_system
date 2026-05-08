package promotions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// ListPromotions handles listing all promotions with filtering
func (h *PromotionHandler) ListPromotions(w http.ResponseWriter, r *http.Request) {
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

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Parse filters
	filter := parsePromotionFilter(r)

	// Build query - only show user's promotions or public ones
	query := h.db.Model(&models.Promotion{}).
		Preload("Event").Preload("Organizer")

	// Filter by organizer (user's own promotions)
	query = query.Where("organizer_id = ? OR is_public = ?", user.AccountID, true)

	// Apply filters
	query = applyPromotionFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get promotions
	var promotions []models.Promotion
	query = query.Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("created_at DESC")

	if err := query.Find(&promotions).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch promotions")
		return
	}

	// Convert to response
	promotionResponses := make([]PromotionResponse, len(promotions))
	for i, promo := range promotions {
		promotionResponses[i] = convertToPromotionResponse(promo)
	}

	// Calculate total pages
	totalPages := int(totalCount) / filter.Limit
	if int(totalCount)%filter.Limit > 0 {
		totalPages++
	}

	response := PromotionListResponse{
		Promotions: promotionResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// ListActivePromotions handles listing only active promotions
func (h *PromotionHandler) ListActivePromotions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse filters
	filter := parsePromotionFilter(r)
	status := models.PromotionActive
	filter.Status = &status

	// Build query - only public active promotions
	query := h.db.Model(&models.Promotion{}).
		Preload("Event").Preload("Organizer").
		Where("is_public = ? AND status = ?", true, models.PromotionActive)

	// Apply date filters (must be current)
	now := time.Now()
	query = query.Where("start_date <= ? AND end_date >= ?", now, now)

	// Apply other filters
	query = applyPromotionFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get promotions
	var promotions []models.Promotion
	query = query.Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("created_at DESC")

	if err := query.Find(&promotions).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch promotions")
		return
	}

	// Convert to response
	promotionResponses := make([]PromotionResponse, len(promotions))
	for i, promo := range promotions {
		promotionResponses[i] = convertToPromotionResponse(promo)
	}

	// Calculate total pages
	totalPages := 0
	if filter.Limit > 0 {
		totalPages = int(totalCount) / filter.Limit
		if int(totalCount)%filter.Limit > 0 {
			totalPages++
		}
	}

	response := PromotionListResponse{
		Promotions: promotionResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// ListOrganizerPromotions handles listing promotions for organizer
func (h *PromotionHandler) ListOrganizerPromotions(w http.ResponseWriter, r *http.Request) {
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

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Parse filters
	filter := parsePromotionFilter(r)

	// Build query - join with organizers table to filter by account_id
	query := h.db.Model(&models.Promotion{}).
		Preload("Event").Preload("Organizer").
		Joins("JOIN organizers ON organizers.id = promotions.organizer_id").
		Where("organizers.account_id = ?", user.AccountID)

	// Apply filters
	query = applyPromotionFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get promotions
	var promotions []models.Promotion
	query = query.Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("created_at DESC")

	if err := query.Find(&promotions).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch promotions")
		return
	}

	// Convert to response
	promotionResponses := make([]PromotionResponse, len(promotions))
	for i, promo := range promotions {
		promotionResponses[i] = convertToPromotionResponse(promo)
	}

	// Calculate total pages
	totalPages := 0
	if filter.Limit > 0 {
		totalPages = int(totalCount) / filter.Limit
		if int(totalCount)%filter.Limit > 0 {
			totalPages++
		}
	}

	response := PromotionListResponse{
		Promotions: promotionResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// SearchPromotions handles searching promotions by code or name
func (h *PromotionHandler) SearchPromotions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get search term
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "search term is required")
		return
	}

	// Parse filters
	filter := parsePromotionFilter(r)
	filter.SearchTerm = searchTerm

	// Build query - only public promotions for search
	query := h.db.Model(&models.Promotion{}).
		Preload("Event").Preload("Organizer").
		Where("is_public = ?", true)

	// Apply search
	searchPattern := "%" + searchTerm + "%"
	query = query.Where("UPPER(code) LIKE UPPER(?) OR UPPER(name) LIKE UPPER(?)",
		searchPattern, searchPattern)

	// Apply other filters
	query = applyPromotionFilters(query, filter)

	// Count total
	var totalCount int64
	query.Count(&totalCount)

	// Get promotions
	var promotions []models.Promotion
	query = query.Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Order("created_at DESC")

	if err := query.Find(&promotions).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to search promotions")
		return
	}

	// Convert to response
	promotionResponses := make([]PromotionResponse, len(promotions))
	for i, promo := range promotions {
		promotionResponses[i] = convertToPromotionResponse(promo)
	}

	// Calculate total pages
	totalPages := 0
	if filter.Limit > 0 {
		totalPages = int(totalCount) / filter.Limit
		if int(totalCount)%filter.Limit > 0 {
			totalPages++
		}
	}

	response := PromotionListResponse{
		Promotions: promotionResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to parse promotion filter from request
func parsePromotionFilter(r *http.Request) PromotionFilter {
	filter := PromotionFilter{
		Page:  1,
		Limit: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := models.PromotionStatus(status)
		filter.Status = &s
	}

	if promoType := r.URL.Query().Get("type"); promoType != "" {
		t := models.PromotionType(promoType)
		filter.Type = &t
	}

	if eventID := r.URL.Query().Get("event_id"); eventID != "" {
		if id, err := strconv.ParseUint(eventID, 10, 64); err == nil {
			eid := uint(id)
			filter.EventID = &eid
		}
	}

	if organizerID := r.URL.Query().Get("organizer_id"); organizerID != "" {
		if id, err := strconv.ParseUint(organizerID, 10, 64); err == nil {
			oid := uint(id)
			filter.OrganizerID = &oid
		}
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = &date
		}
	}

	if isPublic := r.URL.Query().Get("is_public"); isPublic != "" {
		switch isPublic {
		case "true":
			val := true
			filter.IsPublic = &val
		case "false":
			val := false
			filter.IsPublic = &val
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.SearchTerm = strings.TrimSpace(search)
	}

	return filter
}

// Helper function to apply promotion filters to query
func applyPromotionFilters(query *gorm.DB, filter PromotionFilter) *gorm.DB {
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}

	if filter.EventID != nil {
		query = query.Where("event_id = ?", *filter.EventID)
	}

	if filter.OrganizerID != nil {
		query = query.Where("organizer_id = ?", *filter.OrganizerID)
	}

	if filter.StartDate != nil {
		query = query.Where("start_date >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("end_date <= ?", *filter.EndDate)
	}

	if filter.IsPublic != nil {
		query = query.Where("is_public = ?", *filter.IsPublic)
	}

	if filter.SearchTerm != "" {
		searchPattern := "%" + filter.SearchTerm + "%"
		query = query.Where(
			"UPPER(code) LIKE UPPER(?) OR UPPER(name) LIKE UPPER(?) OR UPPER(description) LIKE UPPER(?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	return query
}
