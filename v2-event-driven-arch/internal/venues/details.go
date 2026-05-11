package venues

import (
	"net/http"
	"strconv"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func (h *VenueHandler) GetVenueDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var venue models.Venue
	if err := h.db.First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venue")
		return
	}

	response := convertToVenueResponse(&venue)
	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) GetVenueStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
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

	// Get event statistics for this venue
	var stats struct {
		TotalEvents      int64 `json:"total_events"`
		UpcomingEvents   int64 `json:"upcoming_events"`
		PastEvents       int64 `json:"past_events"`
		TotalTicketsSold int64 `json:"total_tickets_sold"`
	}

	// Count total events at this venue
	h.db.Table("events").
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ?", venueID).
		Count(&stats.TotalEvents)

	// Count upcoming events
	h.db.Table("events").
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ? AND start_date > NOW()", venueID).
		Count(&stats.UpcomingEvents)

	// Count past events
	h.db.Table("events").
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ? AND end_date < NOW()", venueID).
		Count(&stats.PastEvents)

	// Count total tickets sold for events at this venue
	h.db.Table("tickets").
		Joins("JOIN events ON tickets.event_id = events.id").
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ?", venueID).
		Count(&stats.TotalTicketsSold)

	response := map[string]interface{}{
		"venue_id":   venue.ID,
		"venue_name": venue.VenueName,
		"statistics": stats,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) GetVenueEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
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

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := r.URL.Query().Get("status") // upcoming, past, ongoing

	// Build query
	query := h.db.Model(&models.Event{}).
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ?", venueID)

	switch status {
	case "upcoming":
		query = query.Where("start_date > NOW()")
	case "past":
		query = query.Where("end_date < NOW()")
	case "ongoing":
		query = query.Where("start_date <= NOW() AND end_date >= NOW()")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count events")
		return
	}

	// Fetch events
	var events []models.Event
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("start_date ASC").Find(&events).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"venue_id":    venue.ID,
		"venue_name":  venue.VenueName,
		"events":      events,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	respondWithJSON(w, http.StatusOK, response)
}
