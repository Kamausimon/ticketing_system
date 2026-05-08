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
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"ticketing_system/pkg/pdf"
	"ticketing_system/pkg/qrcode"
	"time"

	"github.com/gorilla/mux"
)

// FlexibleFloat handles both string and float64 from JSON
type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as float64 first
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		*f = FlexibleFloat(floatVal)
		return nil
	}

	// If that fails, try as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}

	// Parse the string as float
	floatVal, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return err
	}
	*f = FlexibleFloat(floatVal)
	return nil
}

// Intasend Webhook Event Structure
type IntasendWebhookEvent struct {
	ID           string        `json:"id"`
	InvoiceID    string        `json:"invoice_id"`
	State        string        `json:"state"` // PENDING, PROCESSING, COMPLETE, FAILED
	Provider     string        `json:"provider"`
	Charges      FlexibleFloat `json:"charges"`
	NetAmount    FlexibleFloat `json:"net_amount"`
	Currency     string        `json:"currency"`
	Value        FlexibleFloat `json:"value"`
	Account      string        `json:"account"`
	APIRef       string        `json:"api_ref"`
	Host         string        `json:"host"`
	RetryCount   int           `json:"retry_count"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
	FailedReason string        `json:"failed_reason,omitempty"`
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

	// Verify webhook signature (skip in test mode if signature is not provided)
	signature := r.Header.Get("X-IntaSend-Signature")
	if signature != "" {
		// Signature provided - verify it
		if !h.verifyIntasendSignature(body, signature) {
			// Log suspicious webhook
			log.Printf("⚠️ Invalid webhook signature from %s", r.RemoteAddr)
			h.logWebhook(models.WebhookIntasend, "unknown", string(body), r.Header, false, "Invalid signature", r.RemoteAddr, r.UserAgent())
			writeError(w, http.StatusUnauthorized, "Invalid webhook signature")
			return
		}
		log.Printf("✅ Webhook signature verified")
	} else {
		// No signature provided
		if h.IntasendTestMode {
			log.Printf("⚠️ No signature provided - allowing in TEST MODE")
		} else {
			log.Printf("❌ No signature provided in PRODUCTION MODE - rejecting")
			h.logWebhook(models.WebhookIntasend, "unknown", string(body), r.Header, false, "Missing signature in production", r.RemoteAddr, r.UserAgent())
			writeError(w, http.StatusUnauthorized, "Missing webhook signature")
			return
		}
	}

	// Parse webhook event
	var event IntasendWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Failed to parse webhook JSON: %v", err)
		h.logWebhook(models.WebhookIntasend, "unknown", string(body), r.Header, false, fmt.Sprintf("Failed to parse JSON: %v", err), r.RemoteAddr, r.UserAgent())
		writeError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	log.Printf("📨 Received webhook event: ID=%s, State=%s, APIRef=%s, InvoiceID=%s", event.ID, event.State, event.APIRef, event.InvoiceID)

	// Log webhook FIRST to prevent duplicate processing (acts as a lock)
	now := time.Now()
	headersJSON, _ := json.Marshal(r.Header)
	userAgent := r.UserAgent()
	webhookLog := models.WebhookLog{
		Provider:          models.WebhookIntasend,
		EventID:           event.ID,
		EventType:         "payment",
		Status:            models.WebhookReceived,
		Payload:           string(body),
		Headers:           string(headersJSON),
		Success:           false,
		IPAddress:         r.RemoteAddr,
		UserAgent:         &userAgent,
		SignatureValid:    true,
		ProcessedAt:       &now,
		ExternalReference: &event.InvoiceID,
	}

	// Try to create webhook log - will fail if duplicate exists due to unique constraint
	if err := h.db.Create(&webhookLog).Error; err != nil {
		// Check if it's a duplicate
		if h.checkDuplicateIntasendWebhook(event.InvoiceID, event.State) {
			log.Printf("⚠️ Duplicate webhook event detected: InvoiceID=%s, State=%s", event.InvoiceID, event.State)
			writeJSON(w, http.StatusOK, WebhookEventResponse{
				Received:  true,
				Processed: false,
				Message:   "Duplicate event ignored",
			})
			return
		}
		// Not a duplicate, some other error
		log.Printf("❌ Failed to log webhook: %v", err)
	}

	// Process the webhook based on state
	success, err := h.processIntasendWebhookSafe(&event)

	// Update webhook log status
	if success {
		webhookLog.Status = models.WebhookProcessed
		webhookLog.Success = true
		log.Printf("✅ Webhook processed successfully: InvoiceID=%s, State=%s", event.InvoiceID, event.State)
	} else {
		webhookLog.Status = models.WebhookFailed
		if err != nil {
			errMsg := err.Error()
			webhookLog.ErrorMessage = &errMsg
			log.Printf("❌ Failed to process webhook (InvoiceID: %s, State: %s): %v", event.InvoiceID, event.State, err)
		}
	}
	h.db.Save(&webhookLog)

	writeJSON(w, http.StatusOK, WebhookEventResponse{
		Received:  true,
		Processed: success,
		Message:   "Webhook processed",
	})
}

// processIntasendWebhook processes the Intasend webhook event
func (h *PaymentHandler) processIntasendWebhook(event *IntasendWebhookEvent) (bool, error) {
	// Find the payment record by API reference
	log.Printf("🔍 Looking for payment record with API ref: %s", event.APIRef)

	var paymentRecord models.PaymentRecord
	if err := h.db.Where("external_reference = ?", event.APIRef).First(&paymentRecord).Error; err != nil {
		log.Printf("❌ Payment record not found for API ref %s: %v", event.APIRef, err)

		// Debug: Try to find any payment records for this order
		var count int64
		h.db.Model(&models.PaymentRecord{}).Where("external_reference LIKE ?", "%"+event.APIRef+"%").Count(&count)
		log.Printf("📊 Found %d payment records with similar API ref", count)

		return false, fmt.Errorf("payment record not found for API ref %s", event.APIRef)
	}

	log.Printf("✅ Found payment record: ID=%d, OrderID=%v, Status=%s", paymentRecord.ID, paymentRecord.OrderID, paymentRecord.Status)

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
// ATOMIC TRANSACTION: Payment verification + ticket generation
func (h *PaymentHandler) handleIntasendComplete(event *IntasendWebhookEvent, paymentRecord *models.PaymentRecord) (bool, error) {
	if paymentRecord.Status == models.RecordCompleted {
		return true, nil // Already processed
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return false, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Always ensure transaction is closed (either committed or rolled back)
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Panic in handleIntasendComplete, transaction rolled back: %v", r)
		} else if !committed {
			tx.Rollback()
			log.Printf("⚠️ Transaction not committed, rolling back")
		}
	}()

	// Update payment record
	now := time.Now()
	paymentRecord.Status = models.RecordCompleted
	paymentRecord.CompletedAt = &now
	paymentRecord.ExternalTransactionID = &event.ID

	// Calculate fees
	charges := int64(float64(event.Charges) * 100) // Convert to cents
	netAmount := int64(float64(event.NetAmount) * 100)
	paymentRecord.GatewayFeeAmount = models.Money(charges)
	paymentRecord.NetAmount = models.Money(netAmount)

	if err := tx.Save(paymentRecord).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to update payment record: %w", err)
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
		return false, fmt.Errorf("failed to create transaction: %w", err)
	}

	// CRITICAL: Update order status AND generate tickets atomically
	if paymentRecord.OrderID != nil {
		var order models.Order
		if err := tx.Preload("OrderItems.TicketClass.Event").
			First(&order, *paymentRecord.OrderID).Error; err != nil {
			tx.Rollback()
			return false, fmt.Errorf("failed to load order: %w", err)
		}

		// Check if already paid
		if order.Status == models.OrderPaid || order.Status == models.OrderFulfilled {
			log.Printf("ℹ️ Order %d already processed (Status: %s) - skipping duplicate webhook", order.ID, order.Status)
			if err := tx.Commit().Error; err != nil {
				return false, fmt.Errorf("failed to commit transaction: %w", err)
			}
			committed = true // Mark as committed to prevent rollback in defer
			return true, nil
		}

		// Update order status
		order.Status = models.OrderPaid
		order.PaymentStatus = models.PaymentCompleted
		order.IsPaymentReceived = true
		order.CompletedAt = &now

		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			return false, fmt.Errorf("failed to update order: %w", err)
		}

		// Generate tickets for each order item (within same transaction)
		for _, item := range order.OrderItems {
			// Check if tickets already exist
			var existingCount int64
			if err := tx.Model(&models.Ticket{}).
				Where("order_item_id = ?", item.ID).
				Count(&existingCount).Error; err != nil {
				tx.Rollback()
				return false, fmt.Errorf("failed to check existing tickets: %w", err)
			}

			if existingCount > 0 {
				continue // Already generated
			}

			// Create tickets
			for i := 0; i < item.Quantity; i++ {
				ticket := models.Ticket{
					OrderItemID:  item.ID,
					TicketNumber: generateTicketNumber(item.TicketClass.EventID, order.ID, item.ID, i),
					HolderName:   fmt.Sprintf("%s %s", order.FirstName, order.LastName),
					HolderEmail:  order.Email,
					QRCode:       generateQRCode(item.TicketClass.EventID, order.ID, i),
					Status:       models.TicketActive,
				}

				if err := tx.Create(&ticket).Error; err != nil {
					tx.Rollback()
					return false, fmt.Errorf("failed to create ticket: %w", err)
				}
			}
		}

		// Mark order as fulfilled after tickets created
		order.Status = models.OrderFulfilled
		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			return false, fmt.Errorf("failed to mark order as fulfilled: %w", err)
		}

		log.Printf("✅ Payment verified and tickets generated for order %d", order.ID)
	}

	// Commit transaction - payment + tickets succeed together
	if err := tx.Commit().Error; err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true // Mark as committed to prevent rollback in defer

	// Track metrics (after successful commit)
	if h.metrics != nil && paymentRecord.OrderID != nil {
		// Calculate duration from payment initiation to completion
		duration := time.Since(paymentRecord.InitiatedAt)
		h.metrics.TrackPaymentSuccess(event.Provider, "intasend", duration)
	}

	// Trigger PDF generation and ticket emails after successful commit
	if paymentRecord.OrderID != nil {
		go h.generateAndEmailTickets(*paymentRecord.OrderID)
	}

	return true, nil
}

// Helper functions for ticket generation (duplicated for independence)
func generateTicketNumber(eventID, orderID, itemID uint, index int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TKT-%d-%d-%d-%d-%d", eventID, orderID, itemID, index, timestamp)
}

func generateQRCode(eventID, orderID uint, index int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TICKET:EVENT%d:ORDER%d:IDX%d:TIME%d", eventID, orderID, index, timestamp)
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
		log.Printf("⚠️ Webhook secret not configured")
		return false // No secret configured
	}

	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(h.IntasendWebhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Debug logging
	log.Printf("🔐 Signature verification:")
	log.Printf("   Received signature: %s", signature)
	log.Printf("   Expected signature: %s", expectedSignature)
	log.Printf("   Secret length: %d", len(h.IntasendWebhookSecret))
	log.Printf("   Payload length: %d", len(payload))

	isValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !isValid {
		log.Printf("❌ Signature mismatch!")
	} else {
		log.Printf("✅ Signature valid!")
	}

	return isValid
}

// checkDuplicateIntasendWebhook checks for duplicate Intasend webhooks using invoice_id + state
func (h *PaymentHandler) checkDuplicateIntasendWebhook(invoiceID, state string) bool {
	var count int64
	h.db.Model(&models.WebhookLog{}).
		Where("provider = ? AND external_reference = ? AND payload LIKE ?",
			models.WebhookIntasend, invoiceID, "%\"state\": \""+state+"\"%").
		Count(&count)
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

// generateAndEmailTickets generates PDFs and sends ticket emails with the ticketgenerated template
func (h *PaymentHandler) generateAndEmailTickets(orderID uint) {
	log.Printf("🎫 Starting ticket generation and email sending for order %d", orderID)

	if h.notificationService == nil {
		log.Printf("⚠️ Notification service not available, skipping ticket emails for order %d", orderID)
		return
	}

	// Load all tickets for this order
	var tickets []models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Preload("OrderItem.Order").
		Joins("JOIN order_items ON tickets.order_item_id = order_items.id").
		Where("order_items.order_id = ?", orderID).
		Find(&tickets).Error; err != nil {
		log.Printf("❌ Failed to load tickets for order %d: %v", orderID, err)
		return
	}

	if len(tickets) == 0 {
		log.Printf("⚠️ No tickets found for order %d", orderID)
		return
	}

	log.Printf("📧 Found %d tickets for order %d. Generating PDFs and sending emails...", len(tickets), orderID)

	// Import PDF and QR code generators
	pdfGenerator := h.createPDFGenerator()
	qrGenerator := h.createQRGenerator()

	// Generate PDFs and send emails for each ticket
	for i := range tickets {
		ticket := &tickets[i]
		event := &ticket.OrderItem.TicketClass.Event
		order := &ticket.OrderItem.Order

		// Generate PDF with QR code
		log.Printf("🔄 Generating PDF for ticket %s (Event: %s, Holder: %s)", ticket.TicketNumber, event.Title, ticket.HolderEmail)
		pdfPath, pdfData, err := h.generateTicketPDFWithQR(ticket, event, order, pdfGenerator, qrGenerator)
		if err != nil {
			log.Printf("❌ Failed to generate PDF for ticket %s: %v", ticket.TicketNumber, err)
		} else {
			// Update ticket with PDF path
			h.db.Model(ticket).Update("pdf_path", pdfPath)
			log.Printf("✅ Generated PDF for ticket %s at %s (size: %d bytes)", ticket.TicketNumber, pdfPath, len(pdfData))
		}

		// Prepare email with PDF attachment
		emailData := notifications.EmailData{
			To:      []string{ticket.HolderEmail},
			Subject: fmt.Sprintf("Your Ticket for %s", event.Title),
		}

		// Add PDF attachment if generated successfully
		if pdfData != nil {
			emailData.Attachments = []notifications.Attachment{
				{
					Filename: fmt.Sprintf("ticket_%s.pdf", ticket.TicketNumber),
					Content:  pdfData,
					MimeType: "application/pdf",
				},
			}
			log.Printf("📎 Added PDF attachment (%d bytes) to email", len(pdfData))
		} else {
			log.Printf("⚠️ No PDF data available for attachment")
		}

		emailData.HTMLBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .ticket-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #10B981; }
        .button { display: inline-block; background: #10B981; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎫 Your Ticket is Confirmed!</h1>
        </div>
        <div class="content">
            <p>Hi <strong>%s</strong>,</p>
            <p>Great news! Your ticket for <strong>%s</strong> has been confirmed.</p>
            
            <div class="ticket-box">
                <h3>Ticket Details</h3>
                <p><strong>Event:</strong> %s</p>
                <p><strong>Date:</strong> %s</p>
                <p><strong>Location:</strong> %s</p>
                <p><strong>Ticket Type:</strong> %s</p>
                <p><strong>Ticket Number:</strong> <code>%s</code></p>
            </div>

            <p><strong>What's Next?</strong></p>
            <ul>
                <li>Save your ticket number: <code>%s</code></li>
                <li>You can view your ticket anytime in your account</li>
                <li>Show your ticket at the event entrance</li>
                <li>Arrive early on event day</li>
            </ul>

            <p><strong>Important:</strong> Keep your ticket number safe. You'll need it to check in at the event.</p>

            <div class="footer">
                <p>Questions? Contact support</p>
                <p>&copy; 2025 Ticketing System. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>`,
			ticket.HolderName,
			event.Title,
			event.Title,
			event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"),
			event.Location,
			ticket.OrderItem.TicketClass.Name,
			ticket.TicketNumber,
			ticket.TicketNumber,
		)

		log.Printf("📤 Sending email to %s with subject: %s", ticket.HolderEmail, emailData.Subject)
		if err := h.notificationService.GetEmailService().Send(emailData); err != nil {
			log.Printf("❌ Failed to send ticket email to %s: %v", ticket.HolderEmail, err)
		} else {
			log.Printf("✅ Ticket confirmation email sent to %s for ticket %s", ticket.HolderEmail, ticket.TicketNumber)
		}
	}

	log.Printf("🎉 Completed ticket generation and email sending for order %d", orderID)
}

// Helper functions for PDF generation

func (h *PaymentHandler) createPDFGenerator() *pdf.TicketGenerator {
	return pdf.NewTicketGenerator()
}

func (h *PaymentHandler) createQRGenerator() *qrcode.Generator {
	return qrcode.NewGenerator().WithSize(512)
}

func (h *PaymentHandler) generateTicketPDFWithQR(ticket *models.Ticket, event *models.Event, order *models.Order, pdfGen *pdf.TicketGenerator, qrGen *qrcode.Generator) (string, []byte, error) {
	// Generate QR code content
	qrContent := fmt.Sprintf("TICKET:%s|EVENT:%d|ATTENDEE:%s",
		ticket.TicketNumber,
		event.ID,
		ticket.HolderName,
	)

	// Generate QR code bytes
	qrBytes, err := qrGen.GenerateBytes(qrContent)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Prepare ticket data for PDF
	ticketData := pdf.TicketData{
		TicketNumber:  ticket.TicketNumber,
		EventName:     event.Title,
		EventDate:     event.StartDate,
		EventTime:     event.StartDate.Format("3:04 PM"),
		VenueName:     event.Location,
		VenueAddress:  event.Location,
		AttendeeName:  ticket.HolderName,
		AttendeeEmail: ticket.HolderEmail,
		TicketType:    ticket.OrderItem.TicketClass.Name,
		Price:         float64(ticket.OrderItem.UnitPrice) / 100.0,
		Currency:      string(order.Currency),
		QRCode:        qrBytes,
		OrderNumber:   fmt.Sprintf("ORD-%d", order.ID),
		PurchaseDate:  order.CreatedAt,
	}

	// Create storage directory
	storageDir := filepath.Join("storage", "tickets", fmt.Sprintf("%d", order.ID))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate PDF to file
	pdfFileName := fmt.Sprintf("ticket_%s.pdf", ticket.TicketNumber)
	pdfPath := filepath.Join(storageDir, pdfFileName)

	if err := pdfGen.GenerateToFile(ticketData, pdfPath); err != nil {
		return "", nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Read PDF bytes for email attachment
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return pdfPath, nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	return pdfPath, pdfBytes, nil
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
