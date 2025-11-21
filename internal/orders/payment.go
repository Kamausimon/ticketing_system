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

// ProcessPayment handles processing payment for an order
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
	var paymentResult map[string]interface{}
	switch req.PaymentMethod {
	case "stripe":
		paymentResult, err = h.processStripePayment(order, req)
	case "mpesa":
		paymentResult, err = h.processMpesaPayment(order, req)
	case "offline":
		paymentResult, err = h.processOfflinePayment(order, req)
	default:
		middleware.WriteJSONError(w, http.StatusBadRequest, "unsupported payment method")
		return
	}

	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update order payment status
	order.PaymentStatus = models.PaymentCompleted
	order.IsPaymentReceived = true
	order.Status = models.OrderPaid

	if err := h.db.Save(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update order")
		return
	}

	response := map[string]interface{}{
		"message":        "Payment processed successfully",
		"order":          convertToOrderResponse(order),
		"payment_result": paymentResult,
	}

	json.NewEncoder(w).Encode(response)
}

// VerifyPayment handles verifying a payment (e.g., webhook callback)
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

	// Get order
	var order models.Order
	if err := h.db.First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// TODO: Implement actual payment verification with payment gateway
	// This would verify the payment with Stripe, M-Pesa, etc.

	response := map[string]interface{}{
		"message":          "Payment verification completed",
		"payment_verified": true,
		"order_id":         order.ID,
	}

	json.NewEncoder(w).Encode(response)
}

// processStripePayment processes payment through Stripe
func (h *OrderHandler) processStripePayment(order models.Order, req ProcessPaymentRequest) (map[string]interface{}, error) {
	// TODO: Implement actual Stripe payment processing
	// This is a mock implementation

	// Simulate payment processing
	time.Sleep(500 * time.Millisecond)

	return map[string]interface{}{
		"payment_method": "stripe",
		"transaction_id": fmt.Sprintf("stripe_%d_%d", order.ID, time.Now().Unix()),
		"status":         "success",
		"amount":         req.Amount,
		"currency":       req.Currency,
	}, nil
}

// processMpesaPayment processes payment through M-Pesa
func (h *OrderHandler) processMpesaPayment(order models.Order, req ProcessPaymentRequest) (map[string]interface{}, error) {
	// TODO: Implement actual M-Pesa payment processing
	// This is a mock implementation

	// Simulate payment processing
	time.Sleep(500 * time.Millisecond)

	return map[string]interface{}{
		"payment_method": "mpesa",
		"transaction_id": fmt.Sprintf("mpesa_%d_%d", order.ID, time.Now().Unix()),
		"status":         "success",
		"amount":         req.Amount,
		"currency":       req.Currency,
	}, nil
}

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
