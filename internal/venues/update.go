package venues

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type UpdateVenueRequest struct {
	VenueName        *string           `json:"venue_name,omitempty"`
	VenueType        *models.VenueType `json:"venue_type,omitempty"`
	VenueCapacity    *int              `json:"venue_capacity,omitempty"`
	VenueSection     *string           `json:"venue_section,omitempty"`
	VenueLocation    *string           `json:"venue_location,omitempty"`
	Address          *string           `json:"address,omitempty"`
	City             *string           `json:"city,omitempty"`
	State            *string           `json:"state,omitempty"`
	Country          *string           `json:"country,omitempty"`
	ZipCode          *string           `json:"zip_code,omitempty"`
	ParkingAvailable *bool             `json:"parking_available,omitempty"`
	ParkingCapacity  *int              `json:"parking_capacity,omitempty"`
	IsAccessible     *bool             `json:"is_accessible,omitempty"`
	HasWifi          *bool             `json:"has_wifi,omitempty"`
	HasCatering      *bool             `json:"has_catering,omitempty"`
	ContactEmail     *string           `json:"contact_email,omitempty"`
	ContactPhone     *string           `json:"contact_phone,omitempty"`
	Website          *string           `json:"website,omitempty"`
}

func (h *VenueHandler) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var req UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate update request
	if err := validateUpdateVenueRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find existing venue
	var venue models.Venue
	if err := h.db.First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venue")
		return
	}

	// Update fields if provided
	if req.VenueName != nil {
		venue.VenueName = *req.VenueName
	}
	if req.VenueType != nil {
		venue.VenueType = *req.VenueType
	}
	if req.VenueCapacity != nil {
		venue.VenueCapacity = *req.VenueCapacity
	}
	if req.VenueSection != nil {
		venue.VenueSection = *req.VenueSection
	}
	if req.VenueLocation != nil {
		venue.VenueLocation = *req.VenueLocation
	}
	if req.Address != nil {
		venue.Address = *req.Address
	}
	if req.City != nil {
		venue.City = *req.City
	}
	if req.State != nil {
		venue.State = *req.State
	}
	if req.Country != nil {
		venue.Country = *req.Country
	}
	if req.ZipCode != nil {
		venue.ZipCode = *req.ZipCode
	}
	if req.ParkingAvailable != nil {
		venue.ParkingAvailable = *req.ParkingAvailable
	}
	if req.ParkingCapacity != nil {
		venue.ParkingCapacity = req.ParkingCapacity
	}
	if req.IsAccessible != nil {
		venue.IsAccessible = *req.IsAccessible
	}
	if req.HasWifi != nil {
		venue.HasWifi = *req.HasWifi
	}
	if req.HasCatering != nil {
		venue.HasCatering = *req.HasCatering
	}
	if req.ContactEmail != nil {
		venue.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != nil {
		venue.ContactPhone = req.ContactPhone
	}
	if req.Website != nil {
		venue.Website = req.Website
	}

	venue.UpdatedAt = time.Now()

	if err := h.db.Save(&venue).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update venue")
		return
	}

	response := convertToVenueResponse(&venue)
	respondWithJSON(w, http.StatusOK, response)
}

func validateUpdateVenueRequest(req *UpdateVenueRequest) error {
	if req.VenueName != nil && strings.TrimSpace(*req.VenueName) == "" {
		return &ValidationError{Message: "venue name cannot be empty"}
	}

	if req.VenueCapacity != nil && *req.VenueCapacity <= 0 {
		return &ValidationError{Message: "venue capacity must be greater than 0"}
	}

	if req.VenueType != nil {
		validTypes := []models.VenueType{
			models.VenueIndoor,
			models.VenueOutdoor,
			models.VenueSportsArena,
			models.VenueConference,
			models.VenueConvectionCenter,
			models.VenueHotel,
			models.VenueResort,
			models.VenueBreweryWinery,
			models.VenueRestaurant,
			models.VenueTheatre,
		}

		isValidType := false
		for _, vt := range validTypes {
			if *req.VenueType == vt {
				isValidType = true
				break
			}
		}

		if !isValidType {
			return &ValidationError{Message: "invalid venue type"}
		}
	}

	if req.VenueLocation != nil && strings.TrimSpace(*req.VenueLocation) == "" {
		return &ValidationError{Message: "venue location cannot be empty"}
	}

	if req.City != nil && strings.TrimSpace(*req.City) == "" {
		return &ValidationError{Message: "city cannot be empty"}
	}

	if req.Country != nil && strings.TrimSpace(*req.Country) == "" {
		return &ValidationError{Message: "country cannot be empty"}
	}

	if req.ParkingCapacity != nil && *req.ParkingCapacity <= 0 {
		return &ValidationError{Message: "parking capacity must be greater than 0"}
	}

	return nil
}
