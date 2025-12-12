package ticketclasses

import (
	"ticketing_system/internal/models"
	"time"
)

// convertToResponse converts a TicketClass model to response format
func (h *TicketClassHandler) convertToResponse(tc *models.TicketClass) TicketClassResponse {
	// Calculate available quantity
	availableQty := 0
	if tc.QuantityAvailable != nil {
		availableQty = *tc.QuantityAvailable - tc.QuantitySold
		if availableQty < 0 {
			availableQty = 0
		}
	} else {
		// Unlimited availability
		availableQty = -1
	}

	// Determine status
	status := h.calculateStatus(tc, availableQty)

	// Convert price from cents to dollars
	priceFloat := float64(tc.Price) / 100.0
	salesVolumeFloat := float64(tc.SalesVolume) / 100.0

	return TicketClassResponse{
		ID:                tc.ID,
		EventID:           tc.EventID,
		Name:              tc.Name,
		Description:       tc.Description,
		Price:             priceFloat,
		Currency:          tc.Currency,
		QuantityAvailable: tc.QuantityAvailable,
		QuantitySold:      tc.QuantitySold,
		MaxPerOrder:       tc.MaxPerOrder,
		MinPerOrder:       tc.MinPerOrder,
		StartSaleDate:     tc.StartSaleDate,
		EndSaleDate:       tc.EndSaleDate,
		SalesVolume:       salesVolumeFloat,
		IsPaused:          tc.IsPaused,
		SortOrder:         tc.SortOrder,
		IsHidden:          tc.IsHidden,
		AvailableQuantity: availableQty,
		Status:            status,
		CreatedAt:         tc.CreatedAt,
		UpdatedAt:         tc.UpdatedAt,
	}
}

// calculateStatus determines the current status of a ticket class
func (h *TicketClassHandler) calculateStatus(tc *models.TicketClass, availableQty int) TicketClassStatus {
	status := TicketClassStatus{
		IsAvailable: true,
		OnSale:      true,
		SoldOut:     false,
	}

	// Check if paused
	if tc.IsPaused {
		status.IsAvailable = false
		status.OnSale = false
		status.Reason = "Sales paused by organizer"
		return status
	}

	// Check if sold out
	if tc.QuantityAvailable != nil && availableQty <= 0 {
		status.IsAvailable = false
		status.OnSale = false
		status.SoldOut = true
		status.Reason = "Sold out"
		return status
	}

	// Check sale dates
	now := time.Now()
	if tc.StartSaleDate != nil && now.Before(*tc.StartSaleDate) {
		status.IsAvailable = false
		status.OnSale = false
		status.Reason = "Sales have not started yet"
		return status
	}

	if tc.EndSaleDate != nil && now.After(*tc.EndSaleDate) {
		status.IsAvailable = false
		status.OnSale = false
		status.Reason = "Sales have ended"
		return status
	}

	// Check visibility
	if tc.IsHidden {
		status.IsAvailable = false
		status.OnSale = false
		status.Reason = "Hidden by organizer"
		return status
	}

	return status
}
