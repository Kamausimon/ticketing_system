package inventory

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// ReleaseReservation manually releases a reservation (e.g., when checkout is cancelled)
func (h *InventoryHandler) ReleaseReservation(w http.ResponseWriter, r *http.Request) {
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

	// Delete the reservation
	if err := h.db.Delete(&reservation).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to release reservation")
		return
	}

	writeJSON(w, http.StatusOK, ReleaseResponse{
		ReservationID: uint(reservationID),
		Released:      true,
		Message:       fmt.Sprintf("Released %d tickets", reservation.QuantityReserved),
	})
}

// ReleaseExpiredReservations is a background job that releases all expired reservations
func (h *InventoryHandler) ReleaseExpiredReservations(w http.ResponseWriter, r *http.Request) {
	var expiredReservations []models.ReservedTicket
	if err := h.db.Where("expires <= ?", time.Now()).Find(&expiredReservations).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch expired reservations")
		return
	}

	if len(expiredReservations) == 0 {
		writeJSON(w, http.StatusOK, CleanupResponse{
			ReleasedCount:  0,
			ReservationIDs: []uint{},
			CleanedAt:      time.Now(),
		})
		return
	}

	var reservationIDs []uint
	for _, res := range expiredReservations {
		reservationIDs = append(reservationIDs, res.ID)
	}

	// Delete expired reservations
	if err := h.db.Where("id IN ?", reservationIDs).Delete(&models.ReservedTicket{}).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to release expired reservations")
		return
	}

	writeJSON(w, http.StatusOK, CleanupResponse{
		ReleasedCount:  len(reservationIDs),
		ReservationIDs: reservationIDs,
		CleanedAt:      time.Now(),
	})
}

// StartReservationCleanup starts a background goroutine that automatically cleans up expired reservations
func (h *InventoryHandler) StartReservationCleanup() {
	go func() {
		ticker := time.NewTicker(60 * time.Minute) // Check every minute
		defer ticker.Stop()

		log.Println("🧹 Reservation cleanup background job started")

		for range ticker.C {
			var expiredReservations []models.ReservedTicket
			if err := h.db.Where("expires <= ?", time.Now()).Find(&expiredReservations).Error; err != nil {
				log.Printf("⚠️  Error fetching expired reservations: %v", err)
				continue
			}

			if len(expiredReservations) > 0 {
				// Group by event and ticket class to notify waitlist
				eventTickets := make(map[uint]map[uint]int) // eventID -> ticketClassID -> quantity

				var reservationIDs []uint
				for _, res := range expiredReservations {
					reservationIDs = append(reservationIDs, res.ID)

					// Track released quantity by event and ticket class
					if _, exists := eventTickets[res.EventID]; !exists {
						eventTickets[res.EventID] = make(map[uint]int)
					}
					eventTickets[res.EventID][res.TicketID] += res.QuantityReserved
				}

				if err := h.db.Where("id IN ?", reservationIDs).Delete(&models.ReservedTicket{}).Error; err != nil {
					log.Printf("⚠️  Error releasing expired reservations: %v", err)
				} else {
					log.Printf("🧹 Released %d expired reservations", len(reservationIDs))

					// Notify waitlist for each event/ticket class
					for eventID, tickets := range eventTickets {
						for ticketClassID, quantity := range tickets {
							tcID := ticketClassID
							h.autoNotifyWaitlist(eventID, &tcID, quantity)
						}
					}
				}
			}
		}
	}()
}

// ConvertReservationToOrder converts a reservation to an actual order
// This should be called by the orders module when checkout completes
func (h *InventoryHandler) ConvertReservationToOrder(w http.ResponseWriter, r *http.Request) {
	var req ConvertReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ReservationID == 0 || req.OrderID == 0 {
		writeError(w, http.StatusBadRequest, "Reservation ID and Order ID are required")
		return
	}

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

	// Find the reservation
	var reservation models.ReservedTicket
	if err := tx.First(&reservation, req.ReservationID).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusNotFound, "Reservation not found")
		return
	}

	// Verify reservation hasn't expired
	if time.Now().After(reservation.Expires) {
		tx.Rollback()
		writeError(w, http.StatusBadRequest, "Reservation has expired")
		return
	}

	// Update ticket class sold count
	if err := tx.Model(&models.TicketClass{}).
		Where("id = ?", reservation.TicketID).
		UpdateColumn("quantity_sold", gorm.Expr("quantity_sold + ?", reservation.QuantityReserved)).
		Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to update ticket class")
		return
	}

	// Delete the reservation
	if err := tx.Delete(&reservation).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to remove reservation")
		return
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to complete reservation conversion")
		return
	}
	committed = true

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"reservation_id": req.ReservationID,
		"order_id":       req.OrderID,
		"quantity":       reservation.QuantityReserved,
		"message":        "Reservation successfully converted to order",
	})
}

// ReleaseSessionReservations releases all reservations for a specific session
func (h *InventoryHandler) ReleaseSessionReservations(w http.ResponseWriter, r *http.Request) {
	// Get session ID from authenticated user
	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID := fmt.Sprintf("user_%d", userID)

	var reservations []models.ReservedTicket
	if err := h.db.Where("session_id = ?", sessionID).Find(&reservations).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch reservations")
		return
	}

	if len(reservations) == 0 {
		writeJSON(w, http.StatusOK, CleanupResponse{
			ReleasedCount:  0,
			ReservationIDs: []uint{},
			CleanedAt:      time.Now(),
		})
		return
	}

	var reservationIDs []uint
	totalQuantity := 0
	for _, res := range reservations {
		reservationIDs = append(reservationIDs, res.ID)
		totalQuantity += res.QuantityReserved
	}

	// Delete all reservations for this session
	if err := h.db.Where("session_id = ?", sessionID).Delete(&models.ReservedTicket{}).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to release reservations")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"released_count":  len(reservationIDs),
		"reservation_ids": reservationIDs,
		"total_quantity":  totalQuantity,
		"session_id":      sessionID,
		"cleaned_at":      time.Now(),
	})
}

// GetReservationsByEvent gets all active reservations for an event
func (h *InventoryHandler) GetReservationsByEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var reservations []models.ReservedTicket
	if err := h.db.Where("event_id = ? AND expires > ?", eventID, time.Now()).
		Order("created_at DESC").
		Find(&reservations).Error; err != nil {
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
