package refunds

import (
	"fmt"
	"log"

	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
)

// sendRefundRequestedEmail sends an email to the customer when refund is requested
func (h *RefundHandler) sendRefundRequestedEmail(refund *models.RefundRecord, order *models.Order) {
	if h.notificationService == nil {
		log.Println("⚠️ Notification service not configured")
		return
	}

	// Fetch account (customer) information
	var account models.Account
	if err := h.db.First(&account, refund.AccountID).Error; err != nil {
		log.Printf("❌ Failed to fetch account for refund notification: %v", err)
		return
	}

	// Prepare email data
	data := notifications.RefundData{
		CustomerName:   account.FirstName + " " + account.LastName,
		RefundID:       refund.RefundNumber,
		OrderNumber:    fmt.Sprintf("#%d", order.ID),
		Currency:       refund.Currency,
		RefundAmount:   float64(refund.RefundAmount) / 100.0,
		ProcessedDate:  refund.RequestedAt.Format("2006-01-02 15:04:05"),
		RefundMethod:   "Original Payment Method",
		ProcessingDays: 3,
	}

	// Send email
	if err := h.notificationService.SendPlainEmail(
		[]string{account.Email},
		fmt.Sprintf("Refund Request Received - %s", refund.RefundNumber),
		generateRefundRequestEmailBody(&data),
	); err != nil {
		log.Printf("❌ Failed to send refund requested email: %v", err)
		return
	}

	log.Printf("✅ Refund requested email sent to %s for refund %s", account.Email, refund.RefundNumber)
}

// sendRefundApprovedEmail sends an email to the customer when refund is approved
func (h *RefundHandler) sendRefundApprovedEmail(refund *models.RefundRecord) {
	if h.notificationService == nil {
		log.Println("⚠️ Notification service not configured")
		return
	}

	// Fetch account (customer) information
	var account models.Account
	if err := h.db.First(&account, refund.AccountID).Error; err != nil {
		log.Printf("❌ Failed to fetch account for refund notification: %v", err)
		return
	}

	// Fetch order for reference
	var order models.Order
	if err := h.db.First(&order, refund.OrderID).Error; err != nil {
		log.Printf("❌ Failed to fetch order for refund notification: %v", err)
		return
	}

	// Prepare email data
	data := notifications.RefundData{
		CustomerName:   account.FirstName + " " + account.LastName,
		RefundID:       refund.RefundNumber,
		OrderNumber:    fmt.Sprintf("#%d", order.ID),
		Currency:       refund.Currency,
		RefundAmount:   float64(refund.RefundAmount) / 100.0,
		ProcessedDate:  refund.ApprovedAt.Format("2006-01-02 15:04:05"),
		RefundMethod:   "Original Payment Method",
		ProcessingDays: 3,
	}

	// Send email
	if err := h.notificationService.SendPlainEmail(
		[]string{account.Email},
		fmt.Sprintf("Refund Approved - %s", refund.RefundNumber),
		generateRefundApprovedEmailBody(&data),
	); err != nil {
		log.Printf("❌ Failed to send refund approved email: %v", err)
		return
	}

	log.Printf("✅ Refund approved email sent to %s for refund %s", account.Email, refund.RefundNumber)
}

// sendRefundRejectedEmail sends an email to the customer when refund is rejected
func (h *RefundHandler) sendRefundRejectedEmail(refund *models.RefundRecord) {
	if h.notificationService == nil {
		log.Println("⚠️ Notification service not configured")
		return
	}

	// Fetch account (customer) information
	var account models.Account
	if err := h.db.First(&account, refund.AccountID).Error; err != nil {
		log.Printf("❌ Failed to fetch account for refund notification: %v", err)
		return
	}

	// Fetch order for reference
	var order models.Order
	if err := h.db.First(&order, refund.OrderID).Error; err != nil {
		log.Printf("❌ Failed to fetch order for refund notification: %v", err)
		return
	}

	// Build rejection reason message
	rejectionReason := "Not specified"
	if refund.RejectionReason != nil {
		rejectionReason = *refund.RejectionReason
	}

	// Send email
	body := fmt.Sprintf(`
Dear %s,

We regret to inform you that your refund request has been rejected.

Refund Details:
- Refund ID: %s
- Order Number: #%d
- Amount: %s %.2f
- Reason for Rejection: %s

If you believe this is in error, please contact our support team for further assistance.

Best regards,
Ticketing System Support Team
`, account.FirstName+" "+account.LastName, refund.RefundNumber, order.ID, refund.Currency, float64(refund.RefundAmount)/100.0, rejectionReason)

	if err := h.notificationService.SendPlainEmail(
		[]string{account.Email},
		fmt.Sprintf("Refund Request Rejected - %s", refund.RefundNumber),
		body,
	); err != nil {
		log.Printf("❌ Failed to send refund rejected email: %v", err)
		return
	}

	log.Printf("✅ Refund rejected email sent to %s for refund %s", account.Email, refund.RefundNumber)
}

// sendRefundCompletedEmail sends an email to the customer when refund is completed
func (h *RefundHandler) sendRefundCompletedEmail(refund *models.RefundRecord) {
	if h.notificationService == nil {
		log.Println("⚠️ Notification service not configured")
		return
	}

	// Fetch account (customer) information
	var account models.Account
	if err := h.db.First(&account, refund.AccountID).Error; err != nil {
		log.Printf("❌ Failed to fetch account for refund notification: %v", err)
		return
	}

	// Fetch order for reference
	var order models.Order
	if err := h.db.First(&order, refund.OrderID).Error; err != nil {
		log.Printf("❌ Failed to fetch order for refund notification: %v", err)
		return
	}

	// Prepare email data
	data := notifications.RefundData{
		CustomerName:   account.FirstName + " " + account.LastName,
		RefundID:       refund.RefundNumber,
		OrderNumber:    fmt.Sprintf("#%d", order.ID),
		Currency:       refund.Currency,
		RefundAmount:   float64(refund.RefundAmount) / 100.0,
		ProcessedDate:  refund.CompletedAt.Format("2006-01-02 15:04:05"),
		RefundMethod:   "Original Payment Method",
		ProcessingDays: 3,
	}

	// Send email using the notification service template
	if err := h.notificationService.SendRefundProcessedEmail(account.Email, data); err != nil {
		log.Printf("❌ Failed to send refund completed email: %v", err)
		return
	}

	log.Printf("✅ Refund completed email sent to %s for refund %s", account.Email, refund.RefundNumber)
}

// sendOrganizerRefundPendingEmail sends an email to the organizer when a refund is pending approval
func (h *RefundHandler) sendOrganizerRefundPendingEmail(refund *models.RefundRecord, order *models.Order) {
	if h.notificationService == nil {
		log.Println("⚠️ Notification service not configured")
		return
	}

	// Fetch organizer information
	var organizer models.Organizer
	if err := h.db.First(&organizer, refund.OrganizerID).Error; err != nil {
		log.Printf("❌ Failed to fetch organizer for refund notification: %v", err)
		return
	}

	// Fetch organizer's account for email
	var account models.Account
	if err := h.db.Where("id = ?", organizer.AccountID).First(&account).Error; err != nil {
		log.Printf("❌ Failed to fetch organizer account for refund notification: %v", err)
		return
	}

	// Fetch customer account for reference
	var customerAccount models.Account
	if err := h.db.First(&customerAccount, refund.AccountID).Error; err != nil {
		log.Printf("❌ Failed to fetch customer account: %v", err)
		return
	}

	// Fetch event for context
	var event models.Event
	if err := h.db.First(&event, order.EventID).Error; err != nil {
		log.Printf("❌ Failed to fetch event for refund notification: %v", err)
		return
	}

	// Build notification email body
	body := fmt.Sprintf(`
Dear %s,

A new refund request has been submitted for your event and requires your review.

Refund Details:
- Refund ID: %s
- Order Number: #%d
- Customer: %s (%s)
- Event: %s
- Amount: %s %.2f
- Refund Type: %s
- Reason: %s
- Request Date: %s

Please log into your dashboard to review and approve/reject this refund request.

Best regards,
Ticketing System
`, account.FirstName+" "+account.LastName, refund.RefundNumber, order.ID, customerAccount.FirstName+" "+customerAccount.LastName, customerAccount.Email, event.Title, refund.Currency, float64(refund.RefundAmount)/100.0, refund.RefundType, refund.RefundReason, refund.RequestedAt.Format("2006-01-02 15:04:05"))

	if err := h.notificationService.SendPlainEmail(
		[]string{account.Email},
		fmt.Sprintf("New Refund Request - %s", refund.RefundNumber),
		body,
	); err != nil {
		log.Printf("❌ Failed to send organizer refund notification: %v", err)
		return
	}

	log.Printf("✅ Organizer refund notification sent to %s for refund %s", account.Email, refund.RefundNumber)
}

// Helper functions to generate email bodies

func generateRefundRequestEmailBody(data *notifications.RefundData) string {
	return fmt.Sprintf(`
Dear %s,

We have received your refund request. Thank you for providing the details.

Refund Details:
- Refund ID: %s
- Order Number: %s
- Amount: %s %.2f
- Request Date: %s

Your refund request is now being reviewed by our team. You will receive another email once it has been approved or rejected.

We appreciate your patience.

Best regards,
Ticketing System Support Team
`, data.CustomerName, data.RefundID, data.OrderNumber, data.Currency, data.RefundAmount, data.ProcessedDate)
}

func generateRefundApprovedEmailBody(data *notifications.RefundData) string {
	return fmt.Sprintf(`
Dear %s,

Great news! Your refund request has been approved.

Refund Details:
- Refund ID: %s
- Order Number: %s
- Amount: %s %.2f
- Approval Date: %s
- Processing Method: %s

Your refund will be credited to your original payment method within %d business days. Please note that it may take an additional 1-3 business days for the credit to appear in your account, depending on your bank.

If you have any questions, please don't hesitate to contact us.

Best regards,
Ticketing System Support Team
`, data.CustomerName, data.RefundID, data.OrderNumber, data.Currency, data.RefundAmount, data.ProcessedDate, data.RefundMethod, data.ProcessingDays)
}
