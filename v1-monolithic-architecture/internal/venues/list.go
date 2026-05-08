package venues

import (
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/models"
)

func (h *VenueHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&models.Venue{})

	// Apply filters
	if venueType := r.URL.Query().Get("venue_type"); venueType != "" {
		query = query.Where("venue_type = ?", venueType)
	}

	if city := r.URL.Query().Get("city"); city != "" {
		query = query.Where("LOWER(city) = LOWER(?)", city)
	}

	if country := r.URL.Query().Get("country"); country != "" {
		query = query.Where("LOWER(country) = LOWER(?)", country)
	}

	if minCapacity := r.URL.Query().Get("min_capacity"); minCapacity != "" {
		if cap, err := strconv.Atoi(minCapacity); err == nil {
			query = query.Where("venue_capacity >= ?", cap)
		}
	}

	if maxCapacity := r.URL.Query().Get("max_capacity"); maxCapacity != "" {
		if cap, err := strconv.Atoi(maxCapacity); err == nil {
			query = query.Where("venue_capacity <= ?", cap)
		}
	}

	// Amenity filters
	if r.URL.Query().Get("parking_available") == "true" {
		query = query.Where("parking_available = ?", true)
	}

	if r.URL.Query().Get("is_accessible") == "true" {
		query = query.Where("is_accessible = ?", true)
	}

	if r.URL.Query().Get("has_wifi") == "true" {
		query = query.Where("has_wifi = ?", true)
	}

	if r.URL.Query().Get("has_catering") == "true" {
		query = query.Where("has_catering = ?", true)
	}

	// Search by name or location
	if search := r.URL.Query().Get("search"); search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(venue_name) LIKE ? OR LOWER(venue_location) LIKE ? OR LOWER(city) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count venues")
		return
	}

	// Fetch venues with pagination
	var venues []models.Venue
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&venues).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venues")
		return
	}

	// Convert to response format
	venueResponses := make([]VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = convertToVenueResponse(&venue)
	}

	// Calculate total pages
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := VenueListResponse{
		Venues:     venueResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) SearchVenuesByLocation(w http.ResponseWriter, r *http.Request) {
	latitude := r.URL.Query().Get("latitude")
	longitude := r.URL.Query().Get("longitude")
	radiusKm := r.URL.Query().Get("radius_km")

	if latitude == "" || longitude == "" {
		respondWithError(w, http.StatusBadRequest, "latitude and longitude are required")
		return
	}

	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil || lat < -90 || lat > 90 {
		respondWithError(w, http.StatusBadRequest, "invalid latitude")
		return
	}

	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil || lon < -180 || lon > 180 {
		respondWithError(w, http.StatusBadRequest, "invalid longitude")
		return
	}

	radius := 50.0 // default 50km
	if radiusKm != "" {
		if r, err := strconv.ParseFloat(radiusKm, 64); err == nil && r > 0 {
			radius = r
		}
	}

	// Note: This is a simple implementation. For production, you'd use PostGIS or similar
	// For now, we'll just return venues in the same city/country
	var venues []models.Venue
	query := h.db.Model(&models.Venue{})

	// Apply additional filters if provided
	if venueType := r.URL.Query().Get("venue_type"); venueType != "" {
		query = query.Where("venue_type = ?", venueType)
	}

	if minCapacity := r.URL.Query().Get("min_capacity"); minCapacity != "" {
		if cap, err := strconv.Atoi(minCapacity); err == nil {
			query = query.Where("venue_capacity >= ?", cap)
		}
	}

	if err := query.Order("venue_capacity DESC").Limit(50).Find(&venues).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to search venues")
		return
	}

	// Convert to response format
	venueResponses := make([]VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = convertToVenueResponse(&venue)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"venues": venueResponses,
		"query": map[string]interface{}{
			"latitude":  lat,
			"longitude": lon,
			"radius_km": radius,
		},
	})
}

func (h *VenueHandler) GetVenuesByType(w http.ResponseWriter, r *http.Request) {
	venueType := r.URL.Query().Get("type")
	if venueType == "" {
		respondWithError(w, http.StatusBadRequest, "venue type is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.Venue{}).Where("venue_type = ?", venueType)

	// Apply capacity filter if provided
	if minCapacity := r.URL.Query().Get("min_capacity"); minCapacity != "" {
		if cap, err := strconv.Atoi(minCapacity); err == nil {
			query = query.Where("venue_capacity >= ?", cap)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count venues")
		return
	}

	// Fetch venues
	var venues []models.Venue
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("venue_capacity DESC").Find(&venues).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venues")
		return
	}

	// Convert to response format
	venueResponses := make([]VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = convertToVenueResponse(&venue)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := VenueListResponse{
		Venues:     venueResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	respondWithJSON(w, http.StatusOK, response)
}
