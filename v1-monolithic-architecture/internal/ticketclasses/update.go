package ticketclasses

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// UpdateTicketClass handles updating a ticket class
func (h *TicketClassHandler) UpdateTicketClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket class ID")
		return
	}

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and verify ownership
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event belongs to user
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get ticket class
	var ticketClass models.TicketClass
	if err := h.db.Where("id = ? AND event_id = ?", ticketClassID, eventID).First(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket class not found")
		return
	}

	// Parse request
	var req UpdateTicketClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build updates map for partial updates
	updates := make(map[string]interface{})

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			middleware.WriteJSONError(w, http.StatusBadRequest, "name cannot be empty")
			return
		}
		updates["name"] = trimmed
	}

	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}

	if req.Price != nil {
		if *req.Price < 0 {
			middleware.WriteJSONError(w, http.StatusBadRequest, "price cannot be negative")
			return
		}
		priceInCents := int64(*req.Price * 100)
		updates["price"] = models.Money(priceInCents)
	}

	if req.QuantityAvailable != nil {
		if *req.QuantityAvailable < ticketClass.QuantitySold {
			middleware.WriteJSONError(w, http.StatusBadRequest, "cannot set quantity below already sold quantity")
			return
		}
		updates["quantity_available"] = *req.QuantityAvailable
	}

	if req.MaxPerOrder != nil {
		if *req.MaxPerOrder <= 0 {
			middleware.WriteJSONError(w, http.StatusBadRequest, "max per order must be greater than 0")
			return
		}
		updates["max_per_order"] = *req.MaxPerOrder
	}

	if req.MinPerOrder != nil {
		if *req.MinPerOrder <= 0 {
			middleware.WriteJSONError(w, http.StatusBadRequest, "min per order must be greater than 0")
			return
		}
		updates["min_per_order"] = *req.MinPerOrder
	}

	if req.StartSaleDate != nil {
		updates["start_sale_date"] = *req.StartSaleDate
	}

	if req.EndSaleDate != nil {
		updates["end_sale_date"] = *req.EndSaleDate
	}

	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if req.IsHidden != nil {
		updates["is_hidden"] = *req.IsHidden
	}

	// Apply updates
	if len(updates) > 0 {
		if err := h.db.Model(&ticketClass).Updates(updates).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update ticket class")
			return
		}
	}

	// Reload ticket class to get updated values
	if err := h.db.Where("id = ?", ticketClassID).First(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve updated ticket class")
		return
	}

	response := h.convertToResponse(&ticketClass)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Ticket class updated successfully",
		"ticket_class": response,
	})
}

// DeleteTicketClass handles deleting a ticket class
func (h *TicketClassHandler) DeleteTicketClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket class ID")
		return
	}

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user and verify ownership
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event belongs to user
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get ticket class
	var ticketClass models.TicketClass
	if err := h.db.Where("id = ? AND event_id = ?", ticketClassID, eventID).First(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket class not found")
		return
	}

	// Check if tickets have been sold
	if ticketClass.QuantitySold > 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "cannot delete ticket class with sold tickets")
		return
	}

	// Soft delete
	if err := h.db.Delete(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to delete ticket class")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Ticket class deleted successfully",
	})
}
