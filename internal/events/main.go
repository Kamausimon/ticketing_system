package events

import (
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// EventHandler handles all event-related operations
type EventHandler struct {
	db *gorm.DB
}

// NewEventHandler creates a new event handler
func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{db: db}
}

// Common event request/response structures
type EventResponse struct {
	ID                      uint                 `json:"id"`
	Title                   string               `json:"title"`
	Location                string               `json:"location"`
	Description             string               `json:"description"`
	StartDate               time.Time            `json:"start_date"`
	EndDate                 time.Time            `json:"end_date"`
	OnSaleDate              *time.Time           `json:"on_sale_date"`
	Status                  models.EventStatus   `json:"status"`
	Category                models.EventCategory `json:"category"`
	Currency                string               `json:"currency"`
	MaxCapacity             *int                 `json:"max_capacity"`
	IsLive                  bool                 `json:"is_live"`
	IsPrivate               bool                 `json:"is_private"`
	MinAge                  *int                 `json:"min_age"`
	LocationAddress         *string              `json:"location_address"`
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
	SalesVolume             float32              `json:"sales_volume"`
	OrganizerFeesVolume     float32              `json:"organizer_fees_volume"`
	OrganizerFeeFixed       float32              `json:"organizer_fee_fixed"`
	OrganizerFeePercentage  float32              `json:"organizer_fee_percentage"`
	Organizer               OrganizerSummary     `json:"organizer"`
	Venues                  []VenueSummary       `json:"venues"`
	Images                  []EventImageResponse `json:"images"`
	CreatedAt               time.Time            `json:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
}

type OrganizerSummary struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	About    string  `json:"about"`
	LogoPath *string `json:"logo_path"`
}

type VenueSummary struct {
	ID            uint             `json:"id"`
	VenueName     string           `json:"venue_name"`
	VenueType     models.VenueType `json:"venue_type"`
	VenueLocation string           `json:"venue_location"`
	Capacity      int              `json:"capacity"`
}

type EventImageResponse struct {
	ID        uint   `json:"id"`
	ImagePath string `json:"image_path"`
}

// Pagination and filtering structures
type EventListParams struct {
	Page      int                   `json:"page"`
	Limit     int                   `json:"limit"`
	Category  *models.EventCategory `json:"category"`
	Location  *string               `json:"location"`
	StartDate *time.Time            `json:"start_date"`
	EndDate   *time.Time            `json:"end_date"`
	Status    *models.EventStatus   `json:"status"`
	Search    *string               `json:"search"`
	SortBy    string                `json:"sort_by"`    // "date", "popularity", "price"
	SortOrder string                `json:"sort_order"` // "asc", "desc"
}

type EventListResponse struct {
	Events     []EventResponse `json:"events"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
