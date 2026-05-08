package ticketclasses

import (
	"time"

	"gorm.io/gorm"
)

// TicketClassHandler handles ticket class operations
type TicketClassHandler struct {
	db *gorm.DB
}

// NewTicketClassHandler creates a new ticket class handler
func NewTicketClassHandler(db *gorm.DB) *TicketClassHandler {
	return &TicketClassHandler{db: db}
}

// Request/Response structures

type CreateTicketClassRequest struct {
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Price             float64    `json:"price"`
	Currency          string     `json:"currency"`
	QuantityAvailable *int       `json:"quantity_available"`
	MaxPerOrder       *int       `json:"max_per_order"`
	MinPerOrder       *int       `json:"min_per_order"`
	StartSaleDate     *time.Time `json:"start_sale_date"`
	EndSaleDate       *time.Time `json:"end_sale_date"`
	SortOrder         *int       `json:"sort_order"`
	IsHidden          *bool      `json:"is_hidden"`
	IsVisible         *bool      `json:"is_visible"` // For convenience - inverse of is_hidden
}

type UpdateTicketClassRequest struct {
	Name              *string    `json:"name"`
	Description       *string    `json:"description"`
	Price             *float64   `json:"price"`
	QuantityAvailable *int       `json:"quantity_available"`
	MaxPerOrder       *int       `json:"max_per_order"`
	MinPerOrder       *int       `json:"min_per_order"`
	StartSaleDate     *time.Time `json:"start_sale_date"`
	EndSaleDate       *time.Time `json:"end_sale_date"`
	SortOrder         *int       `json:"sort_order"`
	IsHidden          *bool      `json:"is_hidden"`
}

type TicketClassResponse struct {
	ID                uint              `json:"id"`
	EventID           uint              `json:"event_id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Price             float64           `json:"price"`
	Currency          string            `json:"currency"`
	QuantityAvailable *int              `json:"quantity_available"`
	QuantitySold      int               `json:"quantity_sold"`
	MaxPerOrder       *int              `json:"max_per_order"`
	MinPerOrder       *int              `json:"min_per_order"`
	StartSaleDate     *time.Time        `json:"start_sale_date"`
	EndSaleDate       *time.Time        `json:"end_sale_date"`
	SalesVolume       float64           `json:"sales_volume"`
	IsPaused          bool              `json:"is_paused"`
	SortOrder         int               `json:"sort_order"`
	IsHidden          bool              `json:"is_hidden"`
	AvailableQuantity int               `json:"available_quantity"`
	Status            TicketClassStatus `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type TicketClassStatus struct {
	IsAvailable bool   `json:"is_available"`
	Reason      string `json:"reason,omitempty"`
	OnSale      bool   `json:"on_sale"`
	SoldOut     bool   `json:"sold_out"`
}

type TicketClassListResponse struct {
	TicketClasses []TicketClassResponse `json:"ticket_classes"`
	Total         int64                 `json:"total"`
	EventID       uint                  `json:"event_id"`
	EventTitle    string                `json:"event_title"`
}
