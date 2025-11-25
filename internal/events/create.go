package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// Event creation request structure
type CreateEventRequest struct {
	Title                   string               `json:"title"`
	Location                string               `json:"location"`
	Description             string               `json:"description"`
	StartDate               time.Time            `json:"start_date"`
	EndDate                 time.Time            `json:"end_date"`
	OnSaleDate              *time.Time           `json:"on_sale_date"`
	Category                models.EventCategory `json:"category"`
	Currency                string               `json:"currency"`
	MaxCapacity             *int                 `json:"max_capacity"`
	IsPrivate               bool                 `json:"is_private"`
	MinAge                  *int                 `json:"min_age"`
	LocationAddress         *string              `json:"location_address"`
	LocationAddressLine     *string              `json:"location_address_line"`
	LocationCountry         *string              `json:"location_country"`
	BgType                  string               `json:"bg_type"`
	BgColor                 string               `json:"bg_color"`
	TicketBorderColor       string               `json:"ticket_border_color"`
	TicketBgColor           string               `json:"ticket_bg_color"`
	TicketTextColor         string               `json:"ticket_text_color"`
	TicketSubTextColor      string               `json:"ticket_sub_text_color"`
	BarcodeType             string               `json:"barcode_type"`
	IsBarcodeEnabled        bool                 `json:"is_barcode_enabled"`
	EnableOfflinePayment    bool                 `json:"enable_offline_payment"`
	PreOrderMessageDisplay  *string              `json:"pre_order_message_display"`
	PostOrderMessageDisplay *string              `json:"post_order_message_display"`
	Tags                    string               `json:"tags"`
	OrganizerFeeFixed       float32              `json:"organizer_fee_fixed"`
	OrganizerFeePercentage  float32              `json:"organizer_fee_percentage"`
	VenueIDs                []uint               `json:"venue_ids"`
}

type CreateEventResponse struct {
	Message string `json:"message"`
	EventID uint   `json:"event_id"`
	Status  string `json:"status"`
}

// CreateEvent handles event creation
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if user is an organizer
	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can create events")
		return
	}

	// Get organizer profile
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Check if organizer is verified
	if !organizer.IsEmailConfirmed {
		middleware.WriteJSONError(w, http.StatusForbidden, "organizer email must be verified to create events")
		return
	}

	// Parse request
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := validateEventRequest(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create event
	event := models.Event{
		Title:                   strings.TrimSpace(req.Title),
		Location:                strings.TrimSpace(req.Location),
		Description:             strings.TrimSpace(req.Description),
		StartDate:               req.StartDate,
		EndDate:                 req.EndDate,
		OnSaleDate:              req.OnSaleDate,
		OrganizerID:             organizer.ID,
		AccountID:               user.AccountID,
		Category:                req.Category,
		Currency:                req.Currency,
		MaxCapacity:             req.MaxCapacity,
		Status:                  models.EventDraft, // Start as draft
		IsLive:                  false,
		IsPrivate:               req.IsPrivate,
		MinAge:                  req.MinAge,
		LocationAddress:         req.LocationAddress,
		LocationAddressLine:     req.LocationAddressLine,
		LocationCountry:         req.LocationCountry,
		BgType:                  req.BgType,
		BgColor:                 req.BgColor,
		TicketBorderColor:       req.TicketBorderColor,
		TicketBgColor:           req.TicketBgColor,
		TicketTextColor:         req.TicketTextColor,
		TicketSubTextColor:      req.TicketSubTextColor,
		BarcodeType:             req.BarcodeType,
		IsBarcodeEnabled:        req.IsBarcodeEnabled,
		EnableOfflinePayment:    req.EnableOfflinePayment,
		PreOrderMessageDisplay:  req.PreOrderMessageDisplay,
		PostOrderMessageDisplay: req.PostOrderMessageDisplay,
		Tags:                    req.Tags,
		OrganizerFeeFixed:       req.OrganizerFeeFixed,
		OrganizerFeePercentage:  req.OrganizerFeePercentage,
		SalesVolume:             0,
		OrganizerFeesVolume:     0,
	}

	// Save event
	if err := h.db.Create(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	// Track metrics for event creation
	if h.metrics != nil {
		h.metrics.TrackEventCreated(string(req.Category), fmt.Sprintf("%d", organizer.ID))
	}

	// Associate venues if provided
	if len(req.VenueIDs) > 0 {
		for _, venueID := range req.VenueIDs {
			// Verify venue exists
			var venue models.Venue
			if err := h.db.Where("id = ?", venueID).First(&venue).Error; err != nil {
				continue // Skip invalid venue IDs
			}

			// Create venue association
			eventVenue := models.EventVenues{
				EventID:   event.ID,
				VenueID:   venueID,
				VenueRole: "primary",
			}
			h.db.Create(&eventVenue)
		}
	}

	response := CreateEventResponse{
		Message: "Event created successfully",
		EventID: event.ID,
		Status:  string(event.Status),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// validateEventRequest validates the event creation request
func validateEventRequest(req CreateEventRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.Location == "" {
		return fmt.Errorf("location is required")
	}
	if req.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if req.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}
	if req.StartDate.Before(time.Now()) {
		return fmt.Errorf("start date must be in the future")
	}
	if req.Currency == "" {
		req.Currency = "KSH" // Default currency
	}

	// Validate category
	validCategories := []models.EventCategory{
		models.CategoryMusic, models.CategoryConference, models.CategorySeminar,
		models.CategoryTradeShow, models.CategoryProductLaunch, models.CategoryTeamBuilding,
		models.CategoryCorporateMeeting, models.CategoryCorporateRetreat, models.CategorySports,
		models.CategoryEducational, models.CategoryFestival, models.CategoryArt,
	}

	validCategory := false
	for _, cat := range validCategories {
		if req.Category == cat {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return fmt.Errorf("invalid event category")
	}

	return nil
}
