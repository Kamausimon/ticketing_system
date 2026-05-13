package venues

import (
	"net/http"
	"strconv"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func (h *VenueHandler) DeleteVenue(w http.ResponseWriter, r *http.Request) {
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

	// Check if there are any upcoming events at this venue
	var upcomingEventsCount int64
	if err := h.db.Table("events").
		Joins("JOIN event_venues ON event_venues.event_id = events.id").
		Where("event_venues.venue_id = ? AND start_date > NOW()", venueID).
		Count(&upcomingEventsCount).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check venue events")
		return
	}

	if upcomingEventsCount > 0 {
		respondWithError(w, http.StatusConflict, "Cannot delete venue with upcoming events")
		return
	}

	// Soft delete the venue (GORM's DeletedAt will be set)
	if err := h.db.Delete(&venue).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete venue")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Venue deleted successfully",
	})
}

func (h *VenueHandler) RestoreVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	// Find soft-deleted venue
	var venue models.Venue
	if err := h.db.Unscoped().Where("id = ?", venueID).First(&venue).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venue")
		return
	}

	// Check if venue is actually deleted
	if venue.DeletedAt.Time.IsZero() {
		respondWithError(w, http.StatusBadRequest, "Venue is not deleted")
		return
	}

	// Restore the venue
	if err := h.db.Model(&venue).Unscoped().Update("deleted_at", nil).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to restore venue")
		return
	}

	// Fetch the restored venue
	if err := h.db.First(&venue, venueID).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch restored venue")
		return
	}

	response := convertToVenueResponse(&venue)
	respondWithJSON(w, http.StatusOK, response)
}

func (h *VenueHandler) PermanentlyDeleteVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	// Check if there are ANY events at this venue (past or future)
	var eventsCount int64
	if err := h.db.Table("event_venues").
		Where("venue_id = ?", venueID).
		Count(&eventsCount).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check venue events")
		return
	}

	if eventsCount > 0 {
		respondWithError(w, http.StatusConflict, "Cannot permanently delete venue with associated events")
		return
	}

	// Permanently delete (hard delete)
	if err := h.db.Unscoped().Delete(&models.Venue{}, venueID).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to permanently delete venue")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Venue permanently deleted",
	})
}
