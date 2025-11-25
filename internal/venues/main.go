package venues

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

type VenueHandler struct {
	db *gorm.DB
}

func NewVenueHandler(db *gorm.DB) *VenueHandler {
	return &VenueHandler{db: db}
}

// Response types
type VenueResponse struct {
	ID               uint             `json:"id"`
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
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

type VenueListResponse struct {
	Venues     []VenueResponse `json:"venues"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type AvailabilityResponse struct {
	VenueID   uint               `json:"venue_id"`
	VenueName string             `json:"venue_name"`
	Available bool               `json:"available"`
	Conflicts []ConflictingEvent `json:"conflicts,omitempty"`
}

type ConflictingEvent struct {
	EventID   uint   `json:"event_id"`
	Title     string `json:"title"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Helper functions
func convertToVenueResponse(venue *models.Venue) VenueResponse {
	response := VenueResponse{
		ID:               venue.ID,
		VenueName:        venue.VenueName,
		VenueType:        venue.VenueType,
		VenueCapacity:    venue.VenueCapacity,
		VenueSection:     venue.VenueSection,
		VenueLocation:    venue.VenueLocation,
		Address:          venue.Address,
		City:             venue.City,
		State:            venue.State,
		Country:          venue.Country,
		ZipCode:          venue.ZipCode,
		ParkingAvailable: venue.ParkingAvailable,
		ParkingCapacity:  venue.ParkingCapacity,
		IsAccessible:     venue.IsAccessible,
		HasWifi:          venue.HasWifi,
		HasCatering:      venue.HasCatering,
		ContactEmail:     venue.ContactEmail,
		ContactPhone:     venue.ContactPhone,
		Website:          venue.Website,
		CreatedAt:        venue.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        venue.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return response
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
