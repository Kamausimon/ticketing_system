package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// Intasend Webhook Event Structure
type IntasendWebhookEvent struct {
	ID           string  `json:"id"`
	InvoiceID    string  `json:"invoice_id"`
	State        string  `json:"state"` // PENDING, PROCESSING, COMPLETE, FAILED
	Provider     string  `json:"provider"`
	Charges      float64 `json:"charges"`
	NetAmount    float64 `json:"net_amount"`
	Currency     string  `json:"currency"`
	Value        float64 `json:"value"`
	Account      string  `json:"account"`
	APIRef       string  `json:"api_ref"`
	Host         string  `json:"host"`
	RetryCount   int     `json:"retry_count"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	FailedReason string  `json:"failed_reason,omitempty"`
}

// HandleIntasendWebhook processes incoming webhooks from Intasend with enhanced error handling
func (h *PaymentHandler) HandleIntasendWebhook(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			stackTrace := string(debug.Stack())
			log.Printf("❌ PANIC in webhook handler: %v\nStack: %s", rec, stackTrace)
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Read the raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read webhook body: %v", err)
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Verify webhook signature
	signature := r.Header.Get("X-IntaSend-Signature")
	if !h.verifyIntasendSignature(body, signature) {
		// Log suspicious webhook
		log.Printf("⚠️ Invalid webhook signature from %s", r.RemoteAddr)
		h.logWebhook(models.WebhookIntasend, "unknown", string(body), r.Header, false, "Invalid signature", r.RemoteAddr, r.UserAgent())
		writeError(w, http.StatusUnauthorized, "Invalid webhook signature")
		return
	}

	// Parse webhook event
	var event IntasendWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Failed to parse webhook JSON: %v", err)
		h.logWebhook(models.WebhookIntasend, "unknown", string(body), r.Header, false, fmt.Sprintf("Failed to parse JSON: %v", err), r.RemoteAddr, r.UserAgent())
		writeError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	log.Printf("📨 Received webhook event: ID=%s, State=%s, APIRef=%s", event.ID, event.State, event.APIRef)

	// Check for duplicate events
	isDuplicate := h.checkDuplicateWebhook(event.ID)
	if isDuplicate {
		log.Printf("⚠️ Duplicate webhook event: %s", event.ID)
		h.logWebhook(models.WebhookIntasend, event.ID, string(body), r.Header, true, "Duplicate event", r.RemoteAddr, r.UserAgent())
		writeJSON(w, http.StatusOK, WebhookEventResponse{
			Received:  true,
			Processed: false,
			Message:   "Duplicate event ignored",
		})
		return
	}

	// Process the webhook based on state
	success, err := h.processIntasendWebhookSafe(&event)
	if err != nil {
		log.Printf("❌ Failed to process webhook (Event: %s): %v", event.ID, err)
		h.logWebhook(models.WebhookIntasend, event.ID, string(body), r.Header, false, fmt.Sprintf("Processing failed: %v", err), r.RemoteAddr, r.UserAgent())
		writeJSON(w, http.StatusOK, WebhookEventResponse{
			Received:  true,
			Processed: false,
			Message:   "Webhook received but processing failed - will be retried",
		})
		return
	}

	// Log successful webhook
	log.Printf("✅ Webhook processed successfully: %s", event.ID)
	h.logWebhook(models.WebhookIntasend, event.ID, string(body), r.Header, true, "", r.RemoteAddr, r.UserAgent())

	writeJSON(w, http.StatusOK, WebhookEventResponse{
		Received:  success,
		Processed: success,
		Message:   "Webhook processed successfully",
	})
}

// processIntasendWebhook processes the Intasend webhook event
func (h *PaymentHandler) processIntasendWebhook(event *IntasendWebhookEvent) (bool, error) {
	// Find the payment record by API reference
	var paymentRecord models.PaymentRecord
	if err := h.db.Where("external_reference = ?", event.APIRef).First(&paymentRecord).Error; err != nil {
		return false, fmt.Errorf("payment record not found for API ref %s", event.APIRef)
	}

	// Update payment record based on state
	switch event.State {
	case "COMPLETE":
		return h.handleIntasendComplete(event, &paymentRecord)
	case "FAILED":
		return h.handleIntasendFailed(event, &paymentRecord)
	case "PROCESSING":
		return h.handleIntasendProcessing(event, &paymentRecord)
	default:
		return true, nil // Ignore other states for now
	}
}

// handleIntasendComplete handles successful payment
func (h *PaymentHandler) handleIntasendComplete(event *IntasendWebhookEvent, paymentRecord *models.PaymentRecord) (bool, error) {
	if paymentRecord.Status == models.RecordCompleted {
		return true, nil // Already processed
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update payment record
	now := time.Now()
	paymentRecord.Status = models.RecordCompleted
	paymentRecord.CompletedAt = &now
	paymentRecord.ExternalTransactionID = &event.ID

	// Calculate fees
	charges := int64(event.Charges * 100) // Convert to cents
	netAmount := int64(event.NetAmount * 100)
	paymentRecord.GatewayFeeAmount = models.Money(charges)
	paymentRecord.NetAmount = models.Money(netAmount)

	if err := tx.Save(paymentRecord).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// Update order status
	if paymentRecord.OrderID != nil {
		var order models.Order
		if err := tx.First(&order, *paymentRecord.OrderID).Error; err == nil {
			order.Status = models.OrderPaid
			order.PaymentStatus = models.PaymentCompleted
			order.IsPaymentReceived = true
			order.CompletedAt = &now
			tx.Save(&order)
		}
	}

	// Create payment transaction record
	transaction := models.PaymentTransaction{
		Amount:                paymentRecord.Amount,
		Currency:              paymentRecord.Currency,
		Type:                  models.TransactionPayment,
		Status:                models.TransactionCompleted,
		OrderID:               paymentRecord.OrderID,
		PaymentGatewayID:      paymentRecord.PaymentGatewayID,
		ExternalTransactionID: &event.ID,
		ExternalReference:     &event.APIRef,
		ProcessedAt:           &now,
		Description:           fmt.Sprintf("Payment via %s", event.Provider),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()
	return true, nil
}

// handleIntasendFailed handles failed payment
func (h *PaymentHandler) handleIntasendFailed(event *IntasendWebhookEvent, paymentRecord *models.PaymentRecord) (bool, error) {
	now := time.Now()
	paymentRecord.Status = models.RecordFailed
	paymentRecord.FailedAt = &now
	paymentRecord.ExternalTransactionID = &event.ID

	if event.FailedReason != "" {
		paymentRecord.Notes = &event.FailedReason
	}

	if err := h.db.Save(paymentRecord).Error; err != nil {
		return false, err
	}

	// Update order status
	if paymentRecord.OrderID != nil {
		var order models.Order
		if err := h.db.First(&order, *paymentRecord.OrderID).Error; err == nil {
			order.PaymentStatus = models.PaymentFailed
			h.db.Save(&order)
		}
	}

	return true, nil
}

// handleIntasendProcessing handles processing payment
func (h *PaymentHandler) handleIntasendProcessing(event *IntasendWebhookEvent, paymentRecord *models.PaymentRecord) (bool, error) {
	now := time.Now()
	paymentRecord.ProcessedAt = &now
	paymentRecord.ExternalTransactionID = &event.ID

	if err := h.db.Save(paymentRecord).Error; err != nil {
		return false, err
	}

	return true, nil
}

// verifyIntasendSignature verifies the webhook signature from Intasend
func (h *PaymentHandler) verifyIntasendSignature(payload []byte, signature string) bool {
	if h.IntasendWebhookSecret == "" {
		return false // No secret configured
	}

	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(h.IntasendWebhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// checkDuplicateWebhook checks if we've already processed this webhook event
func (h *PaymentHandler) checkDuplicateWebhook(eventID string) bool {
	var count int64
	h.db.Model(&models.WebhookLog{}).Where("event_id = ?", eventID).Count(&count)
	return count > 0
}

// logWebhook logs the webhook event
func (h *PaymentHandler) logWebhook(provider models.WebhookProvider, eventID, payload string, headers http.Header, success bool, errorMsg, ipAddress, userAgent string) {
	now := time.Now()

	headersJSON, _ := json.Marshal(headers)

	webhookLog := models.WebhookLog{
		Provider:       provider,
		EventID:        eventID,
		EventType:      "payment",
		Status:         models.WebhookReceived,
		Payload:        payload,
		Headers:        string(headersJSON),
		Success:        success,
		IPAddress:      ipAddress,
		UserAgent:      &userAgent,
		SignatureValid: success,
		ProcessedAt:    &now,
	}

	if success {
		webhookLog.Status = models.WebhookProcessed
	} else {
		webhookLog.Status = models.WebhookFailed
		webhookLog.ErrorMessage = &errorMsg
	}

	h.db.Create(&webhookLog)
}

// GetWebhookLogs returns webhook logs with filtering
func (h *PaymentHandler) GetWebhookLogs(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	query := h.db.Model(&models.WebhookLog{}).Order("created_at DESC").Limit(limit)

	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var logs []models.WebhookLog
	if err := query.Find(&logs).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch webhook logs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
	})
}

// RetryFailedWebhook retries a failed webhook
func (h *PaymentHandler) RetryFailedWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	var webhookLog models.WebhookLog
	if err := h.db.First(&webhookLog, webhookID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Webhook log not found")
		return
	}

	if !webhookLog.IsRetryable() {
		writeError(w, http.StatusBadRequest, "Webhook is not retryable")
		return
	}

	// Parse and reprocess the webhook
	var event IntasendWebhookEvent
	if err := json.Unmarshal([]byte(webhookLog.Payload), &event); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	success, err := h.processIntasendWebhook(&event)

	now := time.Now()
	webhookLog.RetryCount++
	webhookLog.LastRetryAt = &now

	if success {
		webhookLog.Status = models.WebhookProcessed
		webhookLog.Success = true
		webhookLog.ErrorMessage = nil
	} else {
		webhookLog.Status = models.WebhookFailed
		errMsg := err.Error()
		webhookLog.ErrorMessage = &errMsg
	}

	h.db.Save(&webhookLog)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     success,
		"retry_count": webhookLog.RetryCount,
		"message":     "Webhook retry completed",
	})
}

// processIntasendWebhookSafe wraps webhook processing with panic recovery
func (h *PaymentHandler) processIntasendWebhookSafe(event *IntasendWebhookEvent) (bool, error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("❌ Panic during webhook processing (Event: %s): %v\nStack: %s", event.ID, rec, string(debug.Stack()))
		}
	}()

	return h.processIntasendWebhook(event)
}

// STRIPE WEBHOOK HANDLER - COMMENTED OUT FOR FUTURE USE
/*
func (h *PaymentHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Verify signature
	signature := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, signature, h.StripeWebhookSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid webhook signature")
		return
	}

	// Process based on event type
	switch event.Type {
	case "payment_intent.succeeded":
		// Handle successful payment
	case "payment_intent.payment_failed":
		// Handle failed payment
	case "charge.refunded":
		// Handle refund
	}

	writeJSON(w, http.StatusOK, WebhookEventResponse{
		Received: true,
		Processed: true,
		Message: "Webhook processed",
	})
}
*/
