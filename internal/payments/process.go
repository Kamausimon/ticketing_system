package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// InitiatePayment initiates a payment for an order
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate order exists
	var order models.Order
	if err := h.DB.First(&order, req.OrderID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Generate unique API reference
	apiRef := fmt.Sprintf("ORD-%d-%d", req.OrderID, time.Now().Unix())

	// Create payment record
	paymentRecord := models.PaymentRecord{
		Amount:            models.Money(req.Amount),
		Currency:          req.Currency,
		Type:              models.RecordCustomerPayment,
		Status:            models.RecordPending,
		OrderID:           &req.OrderID,
		EventID:           &order.EventID,
		AccountID:         &order.AccountID,
		Description:       fmt.Sprintf("Payment for Order #%d", req.OrderID),
		InitiatedAt:       time.Now(),
		ExternalReference: &apiRef,
	}

	if err := h.DB.Create(&paymentRecord).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create payment record")
		return
	}

	// Process based on payment method
	var response PaymentResponse

	switch req.PaymentMethod {
	case "mpesa":
		if req.PhoneNumber == nil || *req.PhoneNumber == "" {
			writeError(w, http.StatusBadRequest, "Phone number required for M-Pesa")
			return
		}
		resp, err := h.InitiateMpesaPayment(req.OrderID, req.Amount, *req.PhoneNumber, req.Email, apiRef)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("M-Pesa initiation failed: %v", err))
			return
		}
		checkoutID := resp.CheckoutRequest
		response = PaymentResponse{
			Success:           true,
			TransactionID:     resp.ID,
			Status:            resp.State,
			CheckoutRequestID: &checkoutID,
			Message:           "M-Pesa STK Push sent. Enter your PIN on your phone.",
		}

	case "card":
		redirectURL := "https://yourdomain.com/payment/callback" // TODO: Make configurable
		if req.CallbackURL != nil {
			redirectURL = *req.CallbackURL
		}
		resp, err := h.InitiateCardPayment(req.OrderID, req.Amount, req.Email, order.FirstName, order.LastName, apiRef, redirectURL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Card payment initiation failed: %v", err))
			return
		}
		response = PaymentResponse{
			Success:       true,
			TransactionID: resp.ID,
			Status:        "pending",
			PaymentURL:    &resp.URL,
			Message:       "Redirect to payment page to complete card payment.",
		}

	default:
		writeError(w, http.StatusBadRequest, "Unsupported payment method")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// VerifyPayment verifies payment status
func (h *PaymentHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Get transaction status from Intasend
	status, err := h.GetIntasendTransactionStatus(transactionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to verify payment: %v", err))
		return
	}

	response := PaymentStatusResponse{
		TransactionID: status.ID,
		Status:        status.State,
		Amount:        int64(status.Value * 100),
		Currency:      status.Currency,
		PaymentMethod: status.Provider,
	}

	if status.State == "COMPLETE" {
		completedAt := status.UpdatedAt
		response.CompletedAt = &completedAt
	}

	writeJSON(w, http.StatusOK, response)
}

// GetPaymentStatus gets payment status for an order
func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var paymentRecord models.PaymentRecord
	if err := h.DB.Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&paymentRecord).Error; err != nil {
		writeError(w, http.StatusNotFound, "No payment found for this order")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order_id":     orderID,
		"status":       paymentRecord.Status,
		"amount":       paymentRecord.Amount,
		"currency":     paymentRecord.Currency,
		"initiated_at": paymentRecord.InitiatedAt,
		"completed_at": paymentRecord.CompletedAt,
	})
}

// GetPaymentHistory gets payment history for an account
func (h *PaymentHandler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	accountIDStr := r.URL.Query().Get("account_id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var payments []models.PaymentRecord
	if err := h.DB.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(50).
		Find(&payments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch payment history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"total":    len(payments),
	})
}
