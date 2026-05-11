package ticketclasses

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// CreateTicketClass handles creating a new ticket class for an event
func (h *TicketClassHandler) CreateTicketClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["eventId"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
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

	// Parse request
	var req CreateTicketClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateCreateRequest(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = event.Currency
	}
	if req.MinPerOrder == nil {
		minOrder := 1
		req.MinPerOrder = &minOrder
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Determine hidden status (support both is_hidden and is_visible)
	isHidden := false
	if req.IsHidden != nil {
		isHidden = *req.IsHidden
	} else if req.IsVisible != nil {
		isHidden = !(*req.IsVisible) // is_visible is inverse of is_hidden
	}

	// Create ticket class
	priceInCents := int64(req.Price * 100)
	ticketClass := models.TicketClass{
		EventID:             uint(eventID),
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		Price:               models.Money(priceInCents),
		Currency:            req.Currency,
		QuantityAvailable:   req.QuantityAvailable,
		MaxPerOrder:         req.MaxPerOrder,
		MinPerOrder:         req.MinPerOrder,
		StartSaleDate:       req.StartSaleDate,
		EndSaleDate:         req.EndSaleDate,
		SortOrder:           sortOrder,
		IsHidden:            isHidden,
		IsPaused:            false,
		QuantitySold:        0,
		SalesVolume:         0,
		OrganizerFeesVolume: 0,
	}

	if err := h.db.Create(&ticketClass).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create ticket class")
		return
	}

	response := h.convertToResponse(&ticketClass)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Ticket class created successfully",
		"ticket_class": response,
	})
}

func validateCreateRequest(req CreateTicketClassRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if req.QuantityAvailable != nil && *req.QuantityAvailable < 0 {
		return fmt.Errorf("quantity available cannot be negative")
	}
	if req.MaxPerOrder != nil && *req.MaxPerOrder <= 0 {
		return fmt.Errorf("max per order must be greater than 0")
	}
	if req.MinPerOrder != nil && *req.MinPerOrder <= 0 {
		return fmt.Errorf("min per order must be greater than 0")
	}
	if req.MinPerOrder != nil && req.MaxPerOrder != nil && *req.MinPerOrder > *req.MaxPerOrder {
		return fmt.Errorf("min per order cannot be greater than max per order")
	}
	if req.StartSaleDate != nil && req.EndSaleDate != nil && req.EndSaleDate.Before(*req.StartSaleDate) {
		return fmt.Errorf("end sale date must be after start sale date")
	}
	return nil
}
