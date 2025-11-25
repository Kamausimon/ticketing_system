package venues

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/models"
)

type CreateVenueRequest struct {
	VenueName        string           `json:"venue_name"`
	VenueType        models.VenueType `json:"venue_type"`
	VenueCapacity    int              `json:"venue_capacity"`
	VenueSection     string           `json:"venue_section,omitempty"`
	VenueLocation    string           `json:"venue_location"`
	Address          string           `json:"address,omitempty"`
	City             string           `json:"city"`
	State            string           `json:"state"`
	Country          string           `json:"country"`
	ZipCode          string           `json:"zip_code,omitempty"`
	ParkingAvailable bool             `json:"parking_available"`
	ParkingCapacity  *int             `json:"parking_capacity,omitempty"`
	IsAccessible     bool             `json:"is_accessible"`
	HasWifi          bool             `json:"has_wifi"`
	HasCatering      bool             `json:"has_catering"`
	ContactEmail     *string          `json:"contact_email,omitempty"`
	ContactPhone     *string          `json:"contact_phone,omitempty"`
	Website          *string          `json:"website,omitempty"`
}

func (h *VenueHandler) CreateVenue(w http.ResponseWriter, r *http.Request) {
	var req CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if err := validateCreateVenueRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create venue model
	venue := models.Venue{
		VenueName:        req.VenueName,
		VenueType:        req.VenueType,
		VenueCapacity:    req.VenueCapacity,
		VenueSection:     req.VenueSection,
		VenueLocation:    req.VenueLocation,
		Address:          req.Address,
		City:             req.City,
		State:            req.State,
		Country:          req.Country,
		ZipCode:          req.ZipCode,
		ParkingAvailable: req.ParkingAvailable,
		ParkingCapacity:  req.ParkingCapacity,
		IsAccessible:     req.IsAccessible,
		HasWifi:          req.HasWifi,
		HasCatering:      req.HasCatering,
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,
		Website:          req.Website,
	}

	if err := h.db.Create(&venue).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create venue")
		return
	}

	response := convertToVenueResponse(&venue)
	respondWithJSON(w, http.StatusCreated, response)
}

func validateCreateVenueRequest(req *CreateVenueRequest) error {
	if strings.TrimSpace(req.VenueName) == "" {
		return &ValidationError{Message: "venue name is required"}
	}

	if req.VenueCapacity <= 0 {
		return &ValidationError{Message: "venue capacity must be greater than 0"}
	}

	// Validate venue type
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
		if req.VenueType == vt {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return &ValidationError{Message: "invalid venue type"}
	}

	if strings.TrimSpace(req.VenueLocation) == "" {
		return &ValidationError{Message: "venue location is required"}
	}

	if strings.TrimSpace(req.City) == "" {
		return &ValidationError{Message: "city is required"}
	}

	if strings.TrimSpace(req.Country) == "" {
		return &ValidationError{Message: "country is required"}
	}

	// Validate parking capacity if parking is available
	if req.ParkingAvailable && req.ParkingCapacity != nil && *req.ParkingCapacity <= 0 {
		return &ValidationError{Message: "parking capacity must be greater than 0 if parking is available"}
	}

	return nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
