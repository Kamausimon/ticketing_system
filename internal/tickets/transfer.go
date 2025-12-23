package tickets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"github.com/gorilla/mux"
)

// TransferTicket handles transferring a ticket to another person
func (h *TicketHandler) TransferTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Parse request
	var req TransferTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewHolderName == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_holder_name is required")
		return
	}

	if req.NewHolderEmail == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_holder_email is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check ticket status
	if ticket.Status != models.TicketActive {
		middleware.WriteJSONError(w, http.StatusBadRequest, "only active tickets can be transferred")
		return
	}

	// Check if event allows transfers (in production, add this to Event model)
	// For now, we'll allow all transfers

	// Store original holder info for history
	originalHolderName := ticket.HolderName
	originalHolderEmail := ticket.HolderEmail

	// Update ticket holder
	ticket.HolderName = req.NewHolderName
	ticket.HolderEmail = req.NewHolderEmail

	// Begin transaction to save ticket and log transfer history
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to transfer ticket")
		return
	}

	// Log transfer in history
	transferHistory := models.TicketTransferHistory{
		TicketID:        uint(ticketID),
		FromHolderName:  originalHolderName,
		FromHolderEmail: originalHolderEmail,
		ToHolderName:    req.NewHolderName,
		ToHolderEmail:   req.NewHolderEmail,
		TransferredBy:   userID,
		TransferReason:  req.TransferReason,
		IPAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
	}

	if err := tx.Create(&transferHistory).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to log transfer history")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to complete transfer")
		return
	}

	// Log activity for audit
	activity := models.AccountActivity{
		AccountID:   user.AccountID,
		UserID:      &userID,
		Action:      models.ActionTicketTransferred,
		Category:    models.ActivityCategoryTicket,
		Description: fmt.Sprintf("Ticket %s transferred from %s to %s", ticket.TicketNumber, originalHolderEmail, req.NewHolderEmail),
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Success:     true,
		Severity:    models.SeverityInfo,
		Resource:    "ticket",
		Timestamp:   time.Now(),
	}
	h.db.Create(&activity)

	// Send email notifications
	go h.sendTransferEmailToNewHolder(&ticket, originalHolderName)
	go h.sendTransferConfirmationToOriginalHolder(originalHolderName, originalHolderEmail, &ticket)

	response := map[string]interface{}{
		"message":          "Ticket transferred successfully",
		"ticket_number":    ticket.TicketNumber,
		"new_holder_name":  ticket.HolderName,
		"new_holder_email": ticket.HolderEmail,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTransferHistory handles getting the transfer history of a ticket
func (h *TicketHandler) GetTransferHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership or organizer access
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		// Check if user is the event organizer
		var event models.Event
		h.db.Preload("OrderItem.TicketClass").First(&ticket)
		if err := h.db.Where("id = ? AND account_id = ?",
			ticket.OrderItem.TicketClass.EventID, user.AccountID).First(&event).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Fetch transfer history from database
	var transferHistory []models.TicketTransferHistory
	if err := h.db.Where("ticket_id = ?", ticketID).Order("transferred_at DESC").Find(&transferHistory).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch transfer history")
		return
	}

	// Build transfer history response
	type TransferRecord struct {
		ID              uint      `json:"id"`
		FromHolderName  string    `json:"from_holder_name"`
		FromHolderEmail string    `json:"from_holder_email"`
		ToHolderName    string    `json:"to_holder_name"`
		ToHolderEmail   string    `json:"to_holder_email"`
		TransferredAt   time.Time `json:"transferred_at"`
		TransferReason  string    `json:"transfer_reason,omitempty"`
	}

	transfers := make([]TransferRecord, len(transferHistory))
	for i, th := range transferHistory {
		transfers[i] = TransferRecord{
			ID:              th.ID,
			FromHolderName:  th.FromHolderName,
			FromHolderEmail: th.FromHolderEmail,
			ToHolderName:    th.ToHolderName,
			ToHolderEmail:   th.ToHolderEmail,
			TransferredAt:   th.TransferredAt,
			TransferReason:  th.TransferReason,
		}
	}

	response := map[string]interface{}{
		"ticket_number":    ticket.TicketNumber,
		"current_holder":   ticket.HolderName,
		"current_email":    ticket.HolderEmail,
		"transfer_count":   len(transfers),
		"transfer_history": transfers,
	}

	json.NewEncoder(w).Encode(response)
}

// sendTransferEmailToNewHolder sends an email to the new ticket holder
func (h *TicketHandler) sendTransferEmailToNewHolder(ticket *models.Ticket, fromHolder string) {
	if h.notificationService == nil {
		fmt.Printf("⚠️ Notification service not available, skipping transfer email for ticket %s\n", ticket.TicketNumber)
		return
	}

	// Load full ticket data with event details
	var fullTicket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticket.ID).
		First(&fullTicket).Error; err != nil {
		fmt.Printf("⚠️ Failed to load ticket details for transfer email: %v\n", err)
		return
	}

	event := fullTicket.OrderItem.TicketClass.Event

	emailData := notifications.EmailData{
		To:      []string{fullTicket.HolderEmail},
		Subject: fmt.Sprintf("🎫 Ticket Transferred to You - %s", event.Title),
		HTMLBody: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #10B981; }
        .info-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .info-label { font-weight: bold; color: #6b7280; }
        .info-value { color: #111827; }
        .button { display: inline-block; background: #10B981; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
        .notice { background: #fef3c7; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #f59e0b; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎫 A Ticket Has Been Transferred to You!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Great news! <strong>%s</strong> has transferred a ticket to you.</p>
            
            <div class="ticket-box">
                <h3>Event Details</h3>
                <div class="info-row">
                    <span class="info-label">Event:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Date:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Location:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Ticket Type:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Ticket Number:</span>
                    <span class="info-value">%s</span>
                </div>
            </div>

            <div class="notice">
                <strong>⚠️ Important:</strong> This ticket is now registered in your name. You can download your ticket PDF by logging into your account or by contacting support.
            </div>

            <p><strong>What's Next?</strong></p>
            <ul>
                <li>Download your ticket PDF from your account</li>
                <li>Save the QR code for event entry</li>
                <li>Arrive early on the event day</li>
                <li>Show your QR code at the entrance</li>
            </ul>

            <p>If you didn't expect this transfer or have any questions, please contact us immediately.</p>

            <div class="footer">
                <p>Questions? Contact us at support@ticketing.local</p>
                <p>&copy; 2024 Ticketing System. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, fullTicket.HolderName, fromHolder, event.Title, event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"), event.Location, fullTicket.OrderItem.TicketClass.Name, fullTicket.TicketNumber),
	}

	if err := h.notificationService.GetEmailService().Send(emailData); err != nil {
		fmt.Printf("❌ Failed to send transfer email to new holder %s for ticket %s: %v\n", fullTicket.HolderEmail, fullTicket.TicketNumber, err)
		return
	}

	fmt.Printf("✅ Transfer notification email sent to new holder %s for ticket %s\n", fullTicket.HolderEmail, fullTicket.TicketNumber)
}

// sendTransferConfirmationToOriginalHolder sends a confirmation email to the original holder
func (h *TicketHandler) sendTransferConfirmationToOriginalHolder(originalName, originalEmail string, ticket *models.Ticket) {
	if h.notificationService == nil {
		fmt.Printf("⚠️ Notification service not available, skipping confirmation email\n")
		return
	}

	// Load full ticket data with event details
	var fullTicket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticket.ID).
		First(&fullTicket).Error; err != nil {
		fmt.Printf("⚠️ Failed to load ticket details for confirmation email: %v\n", err)
		return
	}

	event := fullTicket.OrderItem.TicketClass.Event

	emailData := notifications.EmailData{
		To:      []string{originalEmail},
		Subject: fmt.Sprintf("Ticket Transfer Confirmed - %s", event.Title),
		HTMLBody: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3B82F6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #3B82F6; }
        .info-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .info-label { font-weight: bold; color: #6b7280; }
        .info-value { color: #111827; }
        .success { background: #d1fae5; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #10B981; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Ticket Transfer Confirmed</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>This confirms that you have successfully transferred your ticket to <strong>%s</strong> (%s).</p>
            
            <div class="success">
                <strong>✓ Transfer Complete:</strong> The ticket is no longer valid in your name and cannot be used by you.
            </div>

            <div class="ticket-box">
                <h3>Transfer Details</h3>
                <div class="info-row">
                    <span class="info-label">Event:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Date:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Ticket Number:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Transferred To:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">New Holder Email:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Transfer Date:</span>
                    <span class="info-value">%s</span>
                </div>
            </div>

            <p><strong>Important:</strong></p>
            <ul>
                <li>This ticket is now registered to the new holder</li>
                <li>You will no longer be able to access or use this ticket</li>
                <li>The new holder has been notified via email</li>
                <li>If this was done in error, please contact support immediately</li>
            </ul>

            <p>If you did not authorize this transfer, please contact us immediately at support@ticketing.local</p>

            <div class="footer">
                <p>Questions? Contact us at support@ticketing.local</p>
                <p>&copy; 2024 Ticketing System. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, originalName, fullTicket.HolderName, fullTicket.HolderEmail, event.Title, event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"), fullTicket.TicketNumber, fullTicket.HolderName, fullTicket.HolderEmail, time.Now().Format("Monday, January 2, 2006 at 3:04 PM")),
	}

	if err := h.notificationService.GetEmailService().Send(emailData); err != nil {
		fmt.Printf("❌ Failed to send confirmation email to original holder %s: %v\n", originalEmail, err)
		return
	}

	fmt.Printf("✅ Transfer confirmation email sent to original holder %s\n", originalEmail)
}
