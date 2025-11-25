package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// UpdateOrderStatus handles updating the status of an order
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Parse request
	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("Event").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Check permissions
	if user.Role != models.RoleAdmin && user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// If organizer, verify they own the event
	if user.Role == models.RoleOrganizer && order.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, req.Status) {
		middleware.WriteJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("cannot transition from %s to %s", order.Status, req.Status))
		return
	}

	// Update order status
	oldStatus := order.Status
	order.Status = req.Status
	if req.Status == models.OrderFulfilled && order.CompletedAt == nil {
		now := time.Now()
		order.CompletedAt = &now
	}

	if err := h.db.Save(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update order status")
		return
	}

	// Track metrics for order completion
	if h.metrics != nil && req.Status == models.OrderFulfilled && oldStatus != models.OrderFulfilled {
		// Track revenue
		h.metrics.TrackRevenue(float64(order.Amount), order.Currency, fmt.Sprintf("%d", order.EventID), "")

		// Track completed order
		paymentMethod := "unknown"
		h.metrics.TrackOrderCompleted(paymentMethod, float64(order.Amount), order.Currency, time.Since(order.CreatedAt))
	}

	response := map[string]interface{}{
		"message": "Order status updated successfully",
		"order":   convertToOrderResponse(order),
	}

	json.NewEncoder(w).Encode(response)
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("OrderItems").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Check ownership or admin
	if user.Role != models.RoleAdmin && order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order can be cancelled
	if order.Status == models.OrderCancelled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "order is already cancelled")
		return
	}
	if order.Status == models.OrderFulfilled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "cannot cancel fulfilled orders")
		return
	}

	// Start transaction
	tx := h.db.Begin()

	// Update order status
	order.Status = models.OrderCancelled
	order.IsCancelled = true
	now := time.Now()
	order.CancelledAt = &now

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to cancel order")
		return
	}

	// Return tickets to inventory
	for _, item := range order.OrderItems {
		var ticketClass models.TicketClass
		if err := tx.First(&ticketClass, item.TicketClassID).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update inventory")
			return
		}

		ticketClass.QuantitySold -= item.Quantity
		if err := tx.Save(&ticketClass).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update inventory")
			return
		}
	}

	tx.Commit()

	// Track metrics
	if h.metrics != nil {
		h.metrics.OrdersFailed.WithLabelValues("cancelled").Inc()
	}

	response := map[string]interface{}{
		"message": "Order cancelled successfully",
		"order":   convertToOrderResponse(order),
	}

	json.NewEncoder(w).Encode(response)
}

// RefundOrder handles refunding an order
func (h *OrderHandler) RefundOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Parse request
	var req RefundOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Only admins and organizers can process refunds
	if user.Role != models.RoleAdmin && user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("Event").Preload("OrderItems").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// If organizer, verify they own the event
	if user.Role == models.RoleOrganizer && order.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order can be refunded
	if order.Status != models.OrderPaid && order.Status != models.OrderFulfilled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "only paid or fulfilled orders can be refunded")
		return
	}

	// Calculate refund amount
	refundAmount := req.Amount
	if refundAmount == 0 || refundAmount > float64(order.Amount) {
		refundAmount = float64(order.Amount) // Full refund
	}

	// Update order
	order.Status = models.OrderRefunded
	amountRefunded := float32(refundAmount)
	order.AmountRefunded = &amountRefunded
	now := time.Now()
	order.RefundedAt = &now

	if refundAmount < float64(order.Amount) {
		order.IsPartiallyRefunded = true
	}

	if err := h.db.Save(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to process refund")
		return
	}

	// Track metrics
	if h.metrics != nil {
		h.metrics.RefundsTotal.WithLabelValues(order.Currency, req.Reason).Add(refundAmount)
	}

	// TODO: Process actual refund through payment gateway

	response := map[string]interface{}{
		"message":       "Order refunded successfully",
		"order":         convertToOrderResponse(order),
		"refund_amount": refundAmount,
	}

	json.NewEncoder(w).Encode(response)
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(currentStatus, newStatus models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderPending: {
			models.OrderPaid,
			models.OrderCancelled,
		},
		models.OrderPaid: {
			models.OrderFulfilled,
			models.OrderRefunded,
			models.OrderCancelled,
		},
		models.OrderFulfilled: {
			models.OrderRefunded,
		},
		models.OrderCancelled:     {}, // Terminal state
		models.OrderRefunded:      {}, // Terminal state
		models.OrderPartialRefund: {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}
