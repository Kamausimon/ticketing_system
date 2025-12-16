package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// ProcessPayment handles processing payment for an order
// DEPRECATED: Use the /api/payments/initiate endpoint instead for Intasend integration
// This endpoint is maintained for backward compatibility with offline payments only
func (h *OrderHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
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
	var req ProcessPaymentRequest
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
	if err := h.db.First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Check ownership
	if order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order is pending payment
	if order.Status != models.OrderPending || order.PaymentStatus != models.PaymentPending {
		middleware.WriteJSONError(w, http.StatusBadRequest, "order is not pending payment")
		return
	}

	// Validate payment amount
	expectedAmount := float64(order.Amount)
	if order.BookingFee != nil {
		expectedAmount += float64(*order.BookingFee)
	}
	expectedAmount += float64(order.TaxAmount)
	if order.Discount != nil {
		expectedAmount -= float64(*order.Discount)
	}

	if req.Amount < expectedAmount {
		middleware.WriteJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("insufficient payment amount. Expected: %.2f, Received: %.2f", expectedAmount, req.Amount))
		return
	}

	// Process payment based on method
	// Only offline payments are supported through this endpoint
	// For M-Pesa and Card payments, use /api/payments/initiate endpoint with Intasend
	var paymentResult map[string]interface{}
	switch req.PaymentMethod {
	case "offline":
		paymentResult, err = h.processOfflinePayment(order, req)
	case "mpesa", "card":
		middleware.WriteJSONError(w, http.StatusBadRequest,
			"Please use /api/payments/initiate endpoint for M-Pesa and Card payments via Intasend")
		return
	default:
		middleware.WriteJSONError(w, http.StatusBadRequest,
			"unsupported payment method. Use 'offline' or initiate payment via /api/payments/initiate")
		return
	}

	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// ATOMIC TRANSACTION: Process payment + generate tickets together
	// This ensures both operations succeed or both fail - no partial state
	if err := h.ProcessPaymentWithTickets(order.ID, req.PaymentMethod, paymentResult); err != nil {
		// Transaction failed - payment and tickets were rolled back
		middleware.WriteJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("payment transaction failed: %v", err))
		return
	}

	// Reload order to get updated status with tickets
	if err := h.db.Preload("OrderItems.TicketClass.Event").First(&order, order.ID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to reload order")
		return
	}

	// Count generated tickets
	var ticketCount int64
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON tickets.order_item_id = order_items.id").
		Where("order_items.order_id = ?", order.ID).
		Count(&ticketCount)

	response := map[string]interface{}{
		"message":         "Payment processed and tickets generated successfully",
		"order":           convertToOrderResponse(order),
		"payment_result":  paymentResult,
		"tickets_created": ticketCount,
	}

	json.NewEncoder(w).Encode(response)
}

// VerifyPayment handles verifying a payment (e.g., webhook callback)
// DEPRECATED: Payment verification is now handled automatically via Intasend webhooks at /api/payments/webhook/intasend
// This endpoint is maintained for backward compatibility with offline payment verification only
func (h *OrderHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Parse verification data
	var verificationData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&verificationData); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get order with items
	var order models.Order
	if err := h.db.Preload("OrderItems.TicketClass.Event").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// NOTE: For Intasend payments (M-Pesa/Card), verification is automatic via webhooks
	// This endpoint is only for manual verification of offline payments

	// Extract payment method from verification data
	paymentMethod := "unknown"
	if method, ok := verificationData["payment_method"].(string); ok {
		paymentMethod = method
	}

	// ATOMIC TRANSACTION: Verify payment + generate tickets together
	if err := h.ProcessPaymentWithTickets(order.ID, paymentMethod, verificationData); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("payment verification and ticket generation failed: %v", err))
		return
	}

	// Count generated tickets
	var ticketCount int64
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON tickets.order_item_id = order_items.id").
		Where("order_items.order_id = ?", order.ID).
		Count(&ticketCount)

	response := map[string]interface{}{
		"message":          "Payment verified and tickets generated successfully",
		"payment_verified": true,
		"order_id":         order.ID,
		"tickets_created":  ticketCount,
	}

	json.NewEncoder(w).Encode(response)
}

// REMOVED: processStripePayment and processMpesaPayment
// All online payments (M-Pesa, Card) are now processed through Intasend API
// Use the /api/payments/initiate endpoint instead
// See internal/payments/intasend.go for implementation details

// processOfflinePayment processes offline payment
func (h *OrderHandler) processOfflinePayment(order models.Order, req ProcessPaymentRequest) (map[string]interface{}, error) {
	// Offline payments require manual verification by organizer/admin

	return map[string]interface{}{
		"payment_method":      "offline",
		"status":              "pending_verification",
		"amount":              req.Amount,
		"currency":            req.Currency,
		"verification_needed": true,
	}, nil
}
