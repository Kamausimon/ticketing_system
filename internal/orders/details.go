package orders

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetOrderDetails handles getting a single order's details
func (h *OrderHandler) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order with relations
	var order models.Order
	query := h.db.Preload("Event").Preload("OrderItems.TicketClass").
		Where("id = ?", orderID)

	// Only allow users to see their own orders unless they're an organizer/admin
	if user.Role != models.RoleAdmin {
		query = query.Where("account_id = ?", user.AccountID)
	}

	if err := query.First(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// If user is an organizer, verify they own the event
	if user.Role == models.RoleOrganizer && order.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	response := convertToOrderResponse(order)
	json.NewEncoder(w).Encode(response)
}

// GetOrderSummary handles getting a summary of an order (lightweight version)
func (h *OrderHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order (basic info only, no relations)
	var order models.Order
	query := h.db.Where("id = ?", orderID)

	if user.Role != models.RoleAdmin {
		query = query.Where("account_id = ?", user.AccountID)
	}

	if err := query.First(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Get item count
	var itemCount int64
	h.db.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Count(&itemCount)

	summary := map[string]interface{}{
		"id":             order.ID,
		"order_number":   generateOrderNumber(order.ID),
		"event_id":       order.EventID,
		"status":         order.Status,
		"payment_status": order.PaymentStatus,
		"total_amount":   float64(order.Amount),
		"currency":       order.Currency,
		"item_count":     itemCount,
		"order_date":     order.CreatedAt,
	}

	json.NewEncoder(w).Encode(summary)
}
