package venues

import (
	"net/http"
	"strconv"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type CheckAvailabilityRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (h *VenueHandler) CheckVenueAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	// Get date range from query params or request body
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	if endDate.Before(startDate) {
		respondWithError(w, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	// Check if venue exists
	var venue models.Venue
	if err := h.db.First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venue")
		return
	}

	// Find conflicting events through event_venues junction table
	var conflictingEvents []models.Event
	err = h.db.Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ?", venueID).
		Where("(start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?) OR (start_date >= ? AND end_date <= ?)",
			endDate, startDate, // Overlaps start
			endDate, endDate, // Overlaps end
			startDate, endDate, // Completely within range
		).Find(&conflictingEvents).Error

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check availability")
		return
	}

	available := len(conflictingEvents) == 0

	// Build conflicts list
	conflicts := make([]ConflictingEvent, len(conflictingEvents))
	for i, event := range conflictingEvents {
		conflicts[i] = ConflictingEvent{
			EventID:   event.ID,
			Title:     event.Title,
			StartDate: event.StartDate.Format("2006-01-02"),
			EndDate:   event.EndDate.Format("2006-01-02"),
		}
	}

	response := AvailabilityResponse{
		VenueID:   venue.ID,
		VenueName: venue.VenueName,
		Available: available,
		Conflicts: conflicts,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) GetVenueCalendar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	// Get month and year from query params
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	// Default to current month/year
	now := time.Now()
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 1900 || year > 2100 {
		year = now.Year()
	}

	// Check if venue exists
	var venue models.Venue
	if err := h.db.First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venue")
		return
	}

	// Get first and last day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Get all events for this venue in the month
	var events []models.Event
	err = h.db.Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ?", venueID).
		Where("(start_date <= ? AND end_date >= ?) OR (start_date >= ? AND start_date <= ?)",
			lastDay, firstDay, firstDay, lastDay,
		).Order("start_date ASC").Find(&events).Error

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch calendar")
		return
	}

	// Build calendar response
	type CalendarEvent struct {
		EventID   uint   `json:"event_id"`
		Title     string `json:"title"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Status    string `json:"status"`
	}

	calendarEvents := make([]CalendarEvent, len(events))
	for i, event := range events {
		status := "upcoming"
		if event.EndDate.Before(now) {
			status = "past"
		} else if event.StartDate.Before(now) && event.EndDate.After(now) {
			status = "ongoing"
		}

		calendarEvents[i] = CalendarEvent{
			EventID:   event.ID,
			Title:     event.Title,
			StartDate: event.StartDate.Format("2006-01-02"),
			EndDate:   event.EndDate.Format("2006-01-02"),
			Status:    status,
		}
	}

	response := map[string]interface{}{
		"venue_id":   venue.ID,
		"venue_name": venue.VenueName,
		"month":      month,
		"year":       year,
		"events":     calendarEvents,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) FindAvailableVenues(w http.ResponseWriter, r *http.Request) {
	// Get search criteria
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	minCapacity := r.URL.Query().Get("min_capacity")
	venueType := r.URL.Query().Get("venue_type")
	city := r.URL.Query().Get("city")

	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	if endDate.Before(startDate) {
		respondWithError(w, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	// Build base query
	query := h.db.Model(&models.Venue{})

	// Apply filters
	if venueType != "" {
		query = query.Where("venue_type = ?", venueType)
	}

	if city != "" {
		query = query.Where("LOWER(city) = LOWER(?)", city)
	}

	if minCapacity != "" {
		if cap, err := strconv.Atoi(minCapacity); err == nil {
			query = query.Where("venue_capacity >= ?", cap)
		}
	}

	// Apply amenity filters
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

	// Get all venues matching criteria
	var allVenues []models.Venue
	if err := query.Find(&allVenues).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venues")
		return
	}

	// Filter out venues with conflicting events
	var availableVenues []models.Venue
	for _, venue := range allVenues {
		var conflictCount int64
		h.db.Table("events").
			Joins("JOIN event_venues ON event_venues.event_id = events.id").
			Where("event_venues.venue_id = ?", venue.ID).
			Where("(start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?) OR (start_date >= ? AND end_date <= ?)",
				endDate, startDate,
				endDate, endDate,
				startDate, endDate,
			).Count(&conflictCount)

		if conflictCount == 0 {
			availableVenues = append(availableVenues, venue)
		}
	}

	// Convert to response format
	venueResponses := make([]VenueResponse, len(availableVenues))
	for i, venue := range availableVenues {
		venueResponses[i] = convertToVenueResponse(&venue)
	}

	response := map[string]interface{}{
		"available_venues": venueResponses,
		"total":            len(availableVenues),
		"search_criteria": map[string]interface{}{
			"start_date":   startDateStr,
			"end_date":     endDateStr,
			"min_capacity": minCapacity,
			"venue_type":   venueType,
			"city":         city,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}
