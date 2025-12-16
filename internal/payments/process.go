package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"ticketing_system/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// getUserIDFromToken extracts user ID from JWT token
func getUserIDFromToken(r *http.Request) uint {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading env variables: %v", err)
		return 0
	}
	tokenSecret := os.Getenv("JWTSECRET")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return 0
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			return 0
		}
		return uint(userID)
	}

	return 0
}

// InitiatePayment initiates a payment for an order
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate order exists
	var order models.Order
	if err := h.db.First(&order, req.OrderID).Error; err != nil {
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

	if err := h.db.Create(&paymentRecord).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create payment record")
		return
	}

	// Process based on payment method
	var response PaymentResponse

	// Track payment attempt
	if h.metrics != nil {
		h.metrics.TrackPaymentAttempt("intasend", req.PaymentMethod)
	}

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
		// Get redirect URL from environment or use callback URL
		redirectURL := os.Getenv("PAYMENT_CALLBACK_URL")
		if redirectURL == "" {
			redirectURL = "https://yourdomain.com/payment/callback"
		}
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
		TransactionID: status.Invoice.ID,
		Status:        status.Invoice.State,
		Amount:        int64(status.Invoice.Value * 100),
		Currency:      status.Invoice.Currency,
		PaymentMethod: status.Invoice.Provider,
	}

	if status.Invoice.State == "COMPLETE" {
		completedAt := status.Invoice.UpdatedAt
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
	if err := h.db.Where("order_id = ?", orderID).
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

// GetPaymentHistory gets payment history for the authenticated user
func (h *PaymentHandler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := getUserIDFromToken(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get account_id for this user
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	if user.AccountID == 0 {
		writeError(w, http.StatusNotFound, "No account associated with this user")
		return
	}

	var payments []models.PaymentRecord
	if err := h.db.Where("account_id = ?", user.AccountID).
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

// GetAllPayments (Admin) - gets all payments or for a specific account
func (h *PaymentHandler) GetAllPayments(w http.ResponseWriter, r *http.Request) {
	accountIDStr := r.URL.Query().Get("account_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	query := h.db.Order("created_at DESC").Limit(limit)

	// Filter by account_id if provided
	if accountIDStr != "" {
		accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid account ID format")
			return
		}
		query = query.Where("account_id = ?", accountID)
	}

	var payments []models.PaymentRecord
	if err := query.Find(&payments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}

	// Get total count
	var total int64
	countQuery := h.db.Model(&models.PaymentRecord{})
	if accountIDStr != "" {
		accountID, _ := strconv.ParseUint(accountIDStr, 10, 64)
		countQuery = countQuery.Where("account_id = ?", accountID)
	}
	countQuery.Count(&total)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"total":    total,
		"limit":    limit,
	})
}
