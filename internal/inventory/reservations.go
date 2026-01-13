package inventory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
)

const DefaultReservationDuration = 15 * time.Minute

// CreateReservation reserves tickets for a specific duration during checkout
func (h *InventoryHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var req CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.TicketClassID == 0 {
		writeError(w, http.StatusBadRequest, "Ticket class ID is required")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	// Get session ID from header
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID := fmt.Sprintf("user_%d", userID)

	// Start transaction
	tx := h.db.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if !committed {
			tx.Rollback()
		}
	}()

	// Lock the ticket class row to prevent race conditions
	var ticketClass models.TicketClass
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&ticketClass, req.TicketClassID).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusNotFound, "Ticket class not found")
		return
	}

	// Check if ticket class is saleable
	if !h.isTicketClassSaleable(&ticketClass) {
		tx.Rollback()
		writeError(w, http.StatusBadRequest, "Ticket class is not available for sale")
		return
	}

	// Check min/max per order
	if ticketClass.MinPerOrder != nil && req.Quantity < *ticketClass.MinPerOrder {
		tx.Rollback()
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Minimum %d tickets required per order", *ticketClass.MinPerOrder))
		return
	}
	if ticketClass.MaxPerOrder != nil && req.Quantity > *ticketClass.MaxPerOrder {
		tx.Rollback()
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Maximum %d tickets allowed per order", *ticketClass.MaxPerOrder))
		return
	}

	// Calculate available quantity
	availableQty := h.calculateAvailableQuantity(&ticketClass)
	if availableQty < req.Quantity {
		tx.Rollback()
		writeError(w, http.StatusConflict, fmt.Sprintf("Only %d tickets available", availableQty))
		return
	}

	// Check if session already has a reservation for this ticket class
	var existingReservation models.ReservedTicket
	if err := tx.Where("ticket_id = ? AND session_id = ? AND expires > ?",
		req.TicketClassID, sessionID, time.Now()).
		First(&existingReservation).Error; err == nil {
		// Update existing reservation
		existingReservation.QuantityReserved = req.Quantity
		existingReservation.Expires = time.Now().Add(DefaultReservationDuration)
		if err := tx.Save(&existingReservation).Error; err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "Failed to update reservation")
			return
		}

		// Load relations for response
		var ticketClassForResponse models.TicketClass
		var event models.Event
		tx.First(&ticketClassForResponse, req.TicketClassID)
		tx.First(&event, ticketClass.EventID)

		tx.Commit()
		response := h.convertToReservationResponse(&existingReservation, ticketClassForResponse.Name, event.Title)
		writeJSON(w, http.StatusOK, response)
		return
	}

	// Create new reservation
	reservation := models.ReservedTicket{
		TicketID:         req.TicketClassID,
		EventID:          ticketClass.EventID,
		QuantityReserved: req.Quantity,
		SessionID:        sessionID,
		Expires:          time.Now().Add(DefaultReservationDuration),
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to create reservation")
		return
	}

	// Load relations for response
	var event models.Event
	tx.First(&event, ticketClass.EventID)

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to complete reservation")
		return
	}
	committed = true

	response := h.convertToReservationResponse(&reservation, ticketClass.Name, event.Title)
	writeJSON(w, http.StatusCreated, response)
}

// GetReservation retrieves details of a specific reservation
func (h *InventoryHandler) GetReservation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reservationID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	var reservation models.ReservedTicket
	if err := h.db.First(&reservation, reservationID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Reservation not found")
		return
	}

	// Load ticket class and event
	var ticketClass models.TicketClass
	var event models.Event
	h.db.First(&ticketClass, reservation.TicketID)
	h.db.First(&event, reservation.EventID)

	response := h.convertToReservationResponse(&reservation, ticketClass.Name, event.Title)
	writeJSON(w, http.StatusOK, response)
}

// ListUserReservations lists all active reservations for a session
func (h *InventoryHandler) ListUserReservations(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID := fmt.Sprintf("user_%d", userID)

	var reservations []models.ReservedTicket
	query := h.db.Where("session_id = ? AND expires > ?", sessionID, time.Now())

	// Optional filter by event
	if eventID := r.URL.Query().Get("event_id"); eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	if err := query.Order("created_at DESC").Find(&reservations).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch reservations")
		return
	}

	var responses []ReservationResponse
	for _, res := range reservations {
		var ticketClass models.TicketClass
		var event models.Event
		h.db.First(&ticketClass, res.TicketID)
		h.db.First(&event, res.EventID)
		responses = append(responses, h.convertToReservationResponse(&res, ticketClass.Name, event.Title))
	}

	writeJSON(w, http.StatusOK, ReservationListResponse{
		Reservations: responses,
		Total:        int64(len(responses)),
	})
}

// ValidateReservation checks if a reservation is still valid
func (h *InventoryHandler) ValidateReservation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reservationID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	var reservation models.ReservedTicket
	if err := h.db.First(&reservation, reservationID).Error; err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":   false,
			"reason":  "Reservation not found",
			"expired": true,
		})
		return
	}

	isExpired := time.Now().After(reservation.Expires)
	if isExpired {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":   false,
			"reason":  "Reservation has expired",
			"expired": true,
		})
		return
	}

	// Check if ticket class still has availability
	var ticketClass models.TicketClass
	if err := h.db.First(&ticketClass, reservation.TicketID).Error; err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"reason": "Ticket class not found",
		})
		return
	}

	if !h.isTicketClassSaleable(&ticketClass) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"reason": "Ticket class is no longer available for sale",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":          true,
		"reservation_id": reservation.ID,
		"quantity":       reservation.QuantityReserved,
		"expires_at":     reservation.Expires,
		"time_remaining": time.Until(reservation.Expires).String(),
	})
}

// ExtendReservation extends the expiration time of a reservation
func (h *InventoryHandler) ExtendReservation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reservationID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	var req ExtendReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Minutes <= 0 || req.Minutes > 30 {
		writeError(w, http.StatusBadRequest, "Extension must be between 1 and 30 minutes")
		return
	}

	var reservation models.ReservedTicket
	if err := h.db.First(&reservation, reservationID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Reservation not found")
		return
	}

	// Check if already expired
	if time.Now().After(reservation.Expires) {
		writeError(w, http.StatusBadRequest, "Cannot extend expired reservation")
		return
	}

	// Extend expiration
	reservation.Expires = reservation.Expires.Add(time.Duration(req.Minutes) * time.Minute)
	if err := h.db.Save(&reservation).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to extend reservation")
		return
	}

	// Load relations for response
	var ticketClass models.TicketClass
	var event models.Event
	h.db.First(&ticketClass, reservation.TicketID)
	h.db.First(&event, reservation.EventID)

	response := h.convertToReservationResponse(&reservation, ticketClass.Name, event.Title)
	writeJSON(w, http.StatusOK, response)
}
