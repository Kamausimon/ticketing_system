package inventory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

type InventoryHandler struct {
	db       *gorm.DB
	_metrics *analytics.PrometheusMetrics // Reserved for future instrumentation
}

func NewInventoryHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *InventoryHandler {
	return &InventoryHandler{
		db:       db,
		_metrics: metrics,
	}
}

// Request types
type CreateReservationRequest struct {
	TicketClassID uint   `json:"ticket_class_id"`
	Quantity      int    `json:"quantity"`
	SessionID     string `json:"session_id"`
}

type BulkAvailabilityRequest struct {
	TicketClassIDs []uint `json:"ticket_class_ids"`
}

type ConvertReservationRequest struct {
	ReservationID uint `json:"reservation_id"`
	OrderID       uint `json:"order_id"`
}

type ExtendReservationRequest struct {
	Minutes int `json:"minutes"`
}

// Response types
type AvailabilityResponse struct {
	TicketClassID     uint         `json:"ticket_class_id"`
	TicketClassName   string       `json:"ticket_class_name"`
	TotalQuantity     *int         `json:"total_quantity"`
	QuantitySold      int          `json:"quantity_sold"`
	QuantityReserved  int          `json:"quantity_reserved"`
	QuantityAvailable int          `json:"quantity_available"`
	IsAvailable       bool         `json:"is_available"`
	IsPaused          bool         `json:"is_paused"`
	IsHidden          bool         `json:"is_hidden"`
	SaleStartDate     *time.Time   `json:"sale_start_date,omitempty"`
	SaleEndDate       *time.Time   `json:"sale_end_date,omitempty"`
	Price             models.Money `json:"price"`
	Currency          string       `json:"currency"`
}

type EventInventoryResponse struct {
	EventID        uint                   `json:"event_id"`
	EventName      string                 `json:"event_name"`
	TicketClasses  []AvailabilityResponse `json:"ticket_classes"`
	TotalSold      int                    `json:"total_sold"`
	TotalReserved  int                    `json:"total_reserved"`
	TotalAvailable int                    `json:"total_available"`
}

type ReservationResponse struct {
	ID              uint      `json:"id"`
	TicketClassID   uint      `json:"ticket_class_id"`
	TicketClassName string    `json:"ticket_class_name"`
	EventID         uint      `json:"event_id"`
	EventName       string    `json:"event_name"`
	Quantity        int       `json:"quantity"`
	SessionID       string    `json:"session_id"`
	ExpiresAt       time.Time `json:"expires_at"`
	IsExpired       bool      `json:"is_expired"`
	TimeRemaining   string    `json:"time_remaining"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReservationListResponse struct {
	Reservations []ReservationResponse `json:"reservations"`
	Total        int64                 `json:"total"`
}

type BulkAvailabilityResponse struct {
	Availability []AvailabilityResponse `json:"availability"`
}

type ReleaseResponse struct {
	ReservationID uint   `json:"reservation_id"`
	Released      bool   `json:"released"`
	Message       string `json:"message"`
}

type CleanupResponse struct {
	ReleasedCount  int       `json:"released_count"`
	ReservationIDs []uint    `json:"reservation_ids"`
	CleanedAt      time.Time `json:"cleaned_at"`
}

// Helper functions
func (h *InventoryHandler) calculateAvailableQuantity(ticketClass *models.TicketClass) int {
	if ticketClass.QuantityAvailable == nil {
		return 999999 // Unlimited
	}

	// Calculate reserved quantity
	var reservedQty int
	h.db.Model(&models.ReservedTicket{}).
		Where("ticket_id = ? AND expires > ?", ticketClass.ID, time.Now()).
		Select("COALESCE(SUM(quantity_reserved), 0)").
		Scan(&reservedQty)

	available := *ticketClass.QuantityAvailable - ticketClass.QuantitySold - reservedQty
	if available < 0 {
		return 0
	}
	return available
}

func (h *InventoryHandler) isTicketClassSaleable(ticketClass *models.TicketClass) bool {
	now := time.Now()

	// Check if paused or hidden
	if ticketClass.IsPaused || ticketClass.IsHidden {
		return false
	}

	// Check sale dates
	if ticketClass.StartSaleDate != nil && now.Before(*ticketClass.StartSaleDate) {
		return false
	}
	if ticketClass.EndSaleDate != nil && now.After(*ticketClass.EndSaleDate) {
		return false
	}

	return true
}

func (h *InventoryHandler) getReservedQuantity(ticketClassID uint) int {
	var reservedQty int
	h.db.Model(&models.ReservedTicket{}).
		Where("ticket_id = ? AND expires > ?", ticketClassID, time.Now()).
		Select("COALESCE(SUM(quantity_reserved), 0)").
		Scan(&reservedQty)
	return reservedQty
}

func (h *InventoryHandler) convertToAvailabilityResponse(ticketClass *models.TicketClass) AvailabilityResponse {
	reservedQty := h.getReservedQuantity(ticketClass.ID)
	availableQty := h.calculateAvailableQuantity(ticketClass)
	isAvailable := h.isTicketClassSaleable(ticketClass) && availableQty > 0

	return AvailabilityResponse{
		TicketClassID:     ticketClass.ID,
		TicketClassName:   ticketClass.Name,
		TotalQuantity:     ticketClass.QuantityAvailable,
		QuantitySold:      ticketClass.QuantitySold,
		QuantityReserved:  reservedQty,
		QuantityAvailable: availableQty,
		IsAvailable:       isAvailable,
		IsPaused:          ticketClass.IsPaused,
		IsHidden:          ticketClass.IsHidden,
		SaleStartDate:     ticketClass.StartSaleDate,
		SaleEndDate:       ticketClass.EndSaleDate,
		Price:             ticketClass.Price,
		Currency:          ticketClass.Currency,
	}
}

func (h *InventoryHandler) convertToReservationResponse(reservation *models.ReservedTicket, ticketClassName, eventName string) ReservationResponse {
	isExpired := time.Now().After(reservation.Expires)
	timeRemaining := ""
	if !isExpired {
		duration := time.Until(reservation.Expires)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		timeRemaining = fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return ReservationResponse{
		ID:              reservation.ID,
		TicketClassID:   reservation.TicketID,
		TicketClassName: ticketClassName,
		EventID:         reservation.EventID,
		EventName:       eventName,
		Quantity:        reservation.QuantityReserved,
		SessionID:       reservation.SessionID,
		ExpiresAt:       reservation.Expires,
		IsExpired:       isExpired,
		TimeRemaining:   timeRemaining,
		CreatedAt:       reservation.CreatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
