package inventory

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetTicketAvailability returns the availability status for a specific ticket class
func (h *InventoryHandler) GetTicketAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ticket class ID")
		return
	}

	var ticketClass models.TicketClass
	if err := h.db.First(&ticketClass, ticketClassID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Ticket class not found")
		return
	}

	response := h.convertToAvailabilityResponse(&ticketClass)
	writeJSON(w, http.StatusOK, response)
}

// GetEventInventory returns inventory status for all ticket classes in an event
func (h *InventoryHandler) GetEventInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var event models.Event
	if err := h.db.First(&event, eventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	var ticketClasses []models.TicketClass
	if err := h.db.Where("event_id = ?", eventID).Order("sort_order ASC").Find(&ticketClasses).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch ticket classes")
		return
	}

	var ticketClassResponses []AvailabilityResponse
	totalSold := 0
	totalReserved := 0
	totalAvailable := 0

	for _, tc := range ticketClasses {
		response := h.convertToAvailabilityResponse(&tc)
		ticketClassResponses = append(ticketClassResponses, response)
		totalSold += response.QuantitySold
		totalReserved += response.QuantityReserved
		totalAvailable += response.QuantityAvailable
	}

	eventInventory := EventInventoryResponse{
		EventID:        event.ID,
		EventName:      event.Title,
		TicketClasses:  ticketClassResponses,
		TotalSold:      totalSold,
		TotalReserved:  totalReserved,
		TotalAvailable: totalAvailable,
	}

	writeJSON(w, http.StatusOK, eventInventory)
}

// GetInventoryStatus returns detailed inventory status for a ticket class
func (h *InventoryHandler) GetInventoryStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketClassID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ticket class ID")
		return
	}

	var ticketClass models.TicketClass
	if err := h.db.Preload("Event").First(&ticketClass, ticketClassID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Ticket class not found")
		return
	}

	response := h.convertToAvailabilityResponse(&ticketClass)

	// Add additional status information
	status := map[string]interface{}{
		"ticket_class":  response,
		"event_id":      ticketClass.EventID,
		"event_name":    ticketClass.Event.Title,
		"is_saleable":   h.isTicketClassSaleable(&ticketClass),
		"max_per_order": ticketClass.MaxPerOrder,
		"min_per_order": ticketClass.MinPerOrder,
	}

	writeJSON(w, http.StatusOK, status)
}

// BulkCheckAvailability checks availability for multiple ticket classes at once
func (h *InventoryHandler) BulkCheckAvailability(w http.ResponseWriter, r *http.Request) {
	var req BulkAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.TicketClassIDs) == 0 {
		writeError(w, http.StatusBadRequest, "No ticket class IDs provided")
		return
	}

	if len(req.TicketClassIDs) > 50 {
		writeError(w, http.StatusBadRequest, "Maximum 50 ticket classes can be checked at once")
		return
	}

	var ticketClasses []models.TicketClass
	if err := h.db.Where("id IN ?", req.TicketClassIDs).Find(&ticketClasses).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch ticket classes")
		return
	}

	var responses []AvailabilityResponse
	for _, tc := range ticketClasses {
		responses = append(responses, h.convertToAvailabilityResponse(&tc))
	}

	writeJSON(w, http.StatusOK, BulkAvailabilityResponse{
		Availability: responses,
	})
}
