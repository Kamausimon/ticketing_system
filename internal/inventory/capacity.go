package inventory

import (
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// CapacityLevel represents the current capacity status
type CapacityLevel string

const (
	CapacityAvailable CapacityLevel = "available" // > 20% remaining
	CapacityLow       CapacityLevel = "low"       // 10-20% remaining
	CapacityCritical  CapacityLevel = "critical"  // 1-10% remaining
	CapacitySoldOut   CapacityLevel = "sold_out"  // 0% remaining
	CapacityUnlimited CapacityLevel = "unlimited" // No quantity limit
)

// CapacityWarning represents a capacity warning response
type CapacityWarning struct {
	Level             CapacityLevel `json:"level"`
	Message           string        `json:"message"`
	PercentRemaining  float64       `json:"percent_remaining"`
	QuantityRemaining int           `json:"quantity_remaining"`
	IsUrgent          bool          `json:"is_urgent"`
	RecommendedAction string        `json:"recommended_action,omitempty"`
}

// CapacityStatus represents the complete capacity status for a ticket class or event
type CapacityStatus struct {
	TicketClassID     *uint            `json:"ticket_class_id,omitempty"`
	TicketClassName   string           `json:"ticket_class_name,omitempty"`
	EventID           uint             `json:"event_id"`
	EventName         string           `json:"event_name"`
	TotalCapacity     *int             `json:"total_capacity"`
	Available         int              `json:"available"`
	Sold              int              `json:"sold"`
	Reserved          int              `json:"reserved"`
	PercentSold       float64          `json:"percent_sold"`
	PercentAvailable  float64          `json:"percent_available"`
	Warning           *CapacityWarning `json:"warning,omitempty"`
	CanPurchase       bool             `json:"can_purchase"`
	WaitlistAvailable bool             `json:"waitlist_available"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// EventCapacityResponse represents aggregated capacity for an entire event
type EventCapacityResponse struct {
	EventID         uint             `json:"event_id"`
	EventName       string           `json:"event_name"`
	OverallCapacity CapacityStatus   `json:"overall_capacity"`
	TicketClasses   []CapacityStatus `json:"ticket_classes"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// GetTicketClassCapacity returns real-time capacity status with warnings
func (h *InventoryHandler) GetTicketClassCapacity(w http.ResponseWriter, r *http.Request) {
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

	status := h.calculateCapacityStatus(&ticketClass)
	writeJSON(w, http.StatusOK, status)
}

// GetEventCapacity returns aggregated capacity status for all ticket classes in an event
func (h *InventoryHandler) GetEventCapacity(w http.ResponseWriter, r *http.Request) {
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

	var ticketClassStatuses []CapacityStatus
	totalAvailable := 0
	totalSold := 0
	totalReserved := 0
	var totalCapacity *int

	for _, tc := range ticketClasses {
		status := h.calculateCapacityStatus(&tc)
		ticketClassStatuses = append(ticketClassStatuses, status)

		totalAvailable += status.Available
		totalSold += status.Sold
		totalReserved += status.Reserved

		// Aggregate total capacity (if any ticket class has a limit)
		if tc.QuantityAvailable != nil {
			if totalCapacity == nil {
				cap := 0
				totalCapacity = &cap
			}
			*totalCapacity += *tc.QuantityAvailable
		}
	}

	// Calculate overall event capacity
	overallStatus := CapacityStatus{
		EventID:       event.ID,
		EventName:     event.Title,
		TotalCapacity: totalCapacity,
		Available:     totalAvailable,
		Sold:          totalSold,
		Reserved:      totalReserved,
		UpdatedAt:     time.Now(),
	}

	if totalCapacity != nil && *totalCapacity > 0 {
		overallStatus.PercentSold = float64(totalSold) / float64(*totalCapacity) * 100
		overallStatus.PercentAvailable = float64(totalAvailable) / float64(*totalCapacity) * 100
		overallStatus.Warning = h.generateCapacityWarning(totalAvailable, *totalCapacity)
		overallStatus.CanPurchase = totalAvailable > 0
	} else {
		overallStatus.PercentSold = 0
		overallStatus.PercentAvailable = 100
		overallStatus.CanPurchase = true
	}

	// Check if waitlist is available (when sold out)
	overallStatus.WaitlistAvailable = overallStatus.Available == 0

	response := EventCapacityResponse{
		EventID:         event.ID,
		EventName:       event.Title,
		OverallCapacity: overallStatus,
		TicketClasses:   ticketClassStatuses,
		UpdatedAt:       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// MonitorCapacity returns real-time capacity monitoring data (for admin dashboards)
func (h *InventoryHandler) MonitorCapacity(w http.ResponseWriter, r *http.Request) {
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

	// Calculate monitoring metrics
	type MonitoringData struct {
		CapacityStatus
		SalesVelocity      float64    `json:"sales_velocity"` // Tickets per hour
		EstimatedSoldOutAt *time.Time `json:"estimated_sold_out_at"`
		ActiveReservations int        `json:"active_reservations"`
		ReservationRate    float64    `json:"reservation_rate"` // Reservations per hour
	}

	var monitoringResults []MonitoringData

	for _, tc := range ticketClasses {
		status := h.calculateCapacityStatus(&tc)

		// Calculate sales velocity (last 24 hours)
		var recentSales int64
		h.db.Model(&models.Ticket{}).
			Where("ticket_class_id = ? AND created_at > ?", tc.ID, time.Now().Add(-24*time.Hour)).
			Count(&recentSales)

		velocity := float64(recentSales) / 24.0 // tickets per hour

		// Calculate active reservations
		var activeReservations int64
		h.db.Model(&models.ReservedTicket{}).
			Where("ticket_id = ? AND expires > ?", tc.ID, time.Now()).
			Count(&activeReservations)

		// Estimate sold-out time
		var estimatedSoldOut *time.Time
		if velocity > 0 && status.Available > 0 {
			hoursUntilSoldOut := float64(status.Available) / velocity
			soldOutTime := time.Now().Add(time.Duration(hoursUntilSoldOut) * time.Hour)
			estimatedSoldOut = &soldOutTime
		}

		monitoringResults = append(monitoringResults, MonitoringData{
			CapacityStatus:     status,
			SalesVelocity:      velocity,
			EstimatedSoldOutAt: estimatedSoldOut,
			ActiveReservations: int(activeReservations),
			ReservationRate:    velocity * 1.2, // Typically 20% more reservations than sales
		})
	}

	response := map[string]interface{}{
		"event_id":         event.ID,
		"event_name":       event.Title,
		"ticket_classes":   monitoringResults,
		"monitored_at":     time.Now(),
		"refresh_interval": "30s", // Recommended refresh interval
	}

	writeJSON(w, http.StatusOK, response)
}

// Helper: Calculate capacity status for a ticket class
func (h *InventoryHandler) calculateCapacityStatus(ticketClass *models.TicketClass) CapacityStatus {
	reserved := h.getReservedQuantity(ticketClass.ID)
	available := h.calculateAvailableQuantity(ticketClass)

	status := CapacityStatus{
		TicketClassID:   &ticketClass.ID,
		TicketClassName: ticketClass.Name,
		EventID:         ticketClass.EventID,
		TotalCapacity:   ticketClass.QuantityAvailable,
		Available:       available,
		Sold:            ticketClass.QuantitySold,
		Reserved:        reserved,
		CanPurchase:     h.isTicketClassSaleable(ticketClass) && available > 0,
		UpdatedAt:       time.Now(),
	}

	// Load event name
	var event models.Event
	if err := h.db.First(&event, ticketClass.EventID).Error; err == nil {
		status.EventName = event.Title
	}

	// Calculate percentages
	if ticketClass.QuantityAvailable != nil && *ticketClass.QuantityAvailable > 0 {
		totalCapacity := *ticketClass.QuantityAvailable
		status.PercentSold = float64(ticketClass.QuantitySold) / float64(totalCapacity) * 100
		status.PercentAvailable = float64(available) / float64(totalCapacity) * 100
		status.Warning = h.generateCapacityWarning(available, totalCapacity)
	} else {
		// Unlimited capacity
		status.PercentSold = 0
		status.PercentAvailable = 100
		status.Warning = &CapacityWarning{
			Level:   CapacityUnlimited,
			Message: "Unlimited capacity available",
		}
	}

	// Check if waitlist should be available
	status.WaitlistAvailable = available == 0 && ticketClass.QuantityAvailable != nil

	return status
}

// Helper: Generate capacity warning based on availability
func (h *InventoryHandler) generateCapacityWarning(available, total int) *CapacityWarning {
	if total == 0 {
		return nil
	}

	percentRemaining := float64(available) / float64(total) * 100

	if available == 0 {
		return &CapacityWarning{
			Level:             CapacitySoldOut,
			Message:           "This event is completely sold out",
			PercentRemaining:  0,
			QuantityRemaining: 0,
			IsUrgent:          true,
			RecommendedAction: "Join the waitlist to be notified if tickets become available",
		}
	} else if percentRemaining <= 10 {
		return &CapacityWarning{
			Level:             CapacityCritical,
			Message:           fmt.Sprintf("Only %d tickets remaining! Almost sold out", available),
			PercentRemaining:  percentRemaining,
			QuantityRemaining: available,
			IsUrgent:          true,
			RecommendedAction: "Purchase now - tickets are selling fast",
		}
	} else if percentRemaining <= 20 {
		return &CapacityWarning{
			Level:             CapacityLow,
			Message:           fmt.Sprintf("%d tickets remaining - selling quickly", available),
			PercentRemaining:  percentRemaining,
			QuantityRemaining: available,
			IsUrgent:          true,
			RecommendedAction: "Don't wait - secure your tickets now",
		}
	}

	return &CapacityWarning{
		Level:             CapacityAvailable,
		Message:           fmt.Sprintf("%d tickets available", available),
		PercentRemaining:  percentRemaining,
		QuantityRemaining: available,
		IsUrgent:          false,
	}
}
