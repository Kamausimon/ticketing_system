package notifications

import (
	"fmt"
	"log"
	"time"

	"ticketing_system/internal/config"
)

// NotificationService handles all notification operations
type NotificationService struct {
	emailService *EmailService
	config       *config.Config
}

// NewNotificationService creates a new notification service
func NewNotificationService(cfg *config.Config) *NotificationService {
	return &NotificationService{
		emailService: NewEmailService(&cfg.Email),
		config:       cfg,
	}
}

// GetEmailService returns the underlying email service
func (s *NotificationService) GetEmailService() *EmailService {
	return s.emailService
}

// WelcomeData holds data for welcome emails
type WelcomeData struct {
	Name    string
	BaseURL string
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *NotificationService) SendWelcomeEmail(email, name string) error {
	data := WelcomeData{
		Name:    name,
		BaseURL: s.config.App.FrontendURL,
	}

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Welcome to Ticketing System!",
		"welcome",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send welcome email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Welcome email sent to %s", email)
	return nil
}

// VerificationData holds data for verification emails
type VerificationData struct {
	Name             string
	VerificationCode string
	VerificationURL  string
}

// SendVerificationEmail sends an email verification
func (s *NotificationService) SendVerificationEmail(email, name, code string) error {
	data := VerificationData{
		Name:             name,
		VerificationCode: code,
		VerificationURL:  fmt.Sprintf("%s/verify-email?code=%s", s.config.App.FrontendURL, code),
	}

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Verify Your Email Address",
		"verification",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Verification email sent to %s", email)
	return nil
}

// PasswordResetData holds data for password reset emails
type PasswordResetData struct {
	Name     string
	ResetURL string
}

// SendPasswordResetEmail sends a password reset email
func (s *NotificationService) SendPasswordResetEmail(email, name, token string) error {
	data := PasswordResetData{
		Name:     name,
		ResetURL: fmt.Sprintf("%s/reset-password?token=%s", s.config.App.FrontendURL, token),
	}

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Password Reset Request",
		"password_reset",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send password reset email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Password reset email sent to %s", email)
	return nil
}

// OrderItem represents an item in an order
type OrderItem struct {
	Name     string
	Quantity int
	Price    float64
	Currency string
}

// OrderConfirmationData holds data for order confirmation emails
type OrderConfirmationData struct {
	CustomerName string
	OrderNumber  string
	EventName    string
	EventDate    string
	VenueName    string
	Items        []OrderItem
	Currency     string
	Total        float64
	TicketsURL   string
}

// SendOrderConfirmationEmail sends an order confirmation email
func (s *NotificationService) SendOrderConfirmationEmail(email string, data OrderConfirmationData) error {
	data.TicketsURL = fmt.Sprintf("%s/my-tickets", s.config.App.FrontendURL)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		fmt.Sprintf("Order Confirmed - %s", data.EventName),
		"order_confirmation",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send order confirmation to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Order confirmation sent to %s for order %s", email, data.OrderNumber)
	return nil
}

// TicketData holds data for ticket generation emails
type TicketData struct {
	AttendeeName string
	EventName    string
	EventDate    string
	VenueName    string
	TicketType   string
	TicketNumber string
	QRCodeURL    string
	DownloadURL  string
}

// SendTicketGeneratedEmail sends a ticket generated email
func (s *NotificationService) SendTicketGeneratedEmail(email string, data TicketData) error {
	data.DownloadURL = fmt.Sprintf("%s/tickets/%s/download", s.config.App.BaseURL, data.TicketNumber)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		fmt.Sprintf("Your Ticket for %s", data.EventName),
		"ticket_generated",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send ticket email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Ticket email sent to %s for %s", email, data.EventName)
	return nil
}

// EventReminderData holds data for event reminder emails
type EventReminderData struct {
	AttendeeName string
	EventName    string
	EventDate    string
	EventTime    string
	VenueName    string
	VenueAddress string
	TimeUntil    string
	TicketsURL   string
}

// SendEventReminderEmail sends an event reminder email
func (s *NotificationService) SendEventReminderEmail(email string, data EventReminderData) error {
	data.TicketsURL = fmt.Sprintf("%s/my-tickets", s.config.App.FrontendURL)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		fmt.Sprintf("Reminder: %s is Coming Up!", data.EventName),
		"event_reminder",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send event reminder to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Event reminder sent to %s for %s", email, data.EventName)
	return nil
}

// PaymentConfirmationData holds data for payment confirmation emails
type PaymentConfirmationData struct {
	CustomerName  string
	TransactionID string
	Currency      string
	Amount        float64
	PaymentMethod string
	PaymentDate   string
	OrderNumber   string
	ReceiptURL    string
}

// SendPaymentConfirmationEmail sends a payment confirmation email
func (s *NotificationService) SendPaymentConfirmationEmail(email string, data PaymentConfirmationData) error {
	data.ReceiptURL = fmt.Sprintf("%s/receipts/%s", s.config.App.BaseURL, data.TransactionID)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Payment Confirmation",
		"payment_confirmation",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send payment confirmation to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Payment confirmation sent to %s for transaction %s", email, data.TransactionID)
	return nil
}

// RefundData holds data for refund processed emails
type RefundData struct {
	CustomerName   string
	RefundID       string
	OrderNumber    string
	Currency       string
	RefundAmount   float64
	ProcessedDate  string
	RefundMethod   string
	ProcessingDays int
}

// SendRefundProcessedEmail sends a refund processed email
func (s *NotificationService) SendRefundProcessedEmail(email string, data RefundData) error {
	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Refund Processed Successfully",
		"refund_processed",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send refund email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Refund confirmation sent to %s for refund %s", email, data.RefundID)
	return nil
}

// SendPlainEmail sends a plain text email
func (s *NotificationService) SendPlainEmail(to []string, subject, body string) error {
	emailData := EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	return s.emailService.Send(emailData)
}

// SendHTMLEmail sends an HTML email
func (s *NotificationService) SendHTMLEmail(to []string, subject, htmlBody string) error {
	emailData := EmailData{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
	}

	return s.emailService.Send(emailData)
}

// ScheduleEventReminders schedules reminders for upcoming events
func (s *NotificationService) ScheduleEventReminders() {
	// This would typically run as a background job
	// For now, it's a placeholder for the implementation
	log.Println("📅 Event reminder scheduler started")
}

// TestEmailConfiguration tests the email configuration
func (s *NotificationService) TestEmailConfiguration(testEmail string) error {
	testData := EmailData{
		To:      []string{testEmail},
		Subject: "Test Email - Ticketing System",
		Body:    "This is a test email to verify your email configuration is working correctly.",
		HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        .success { color: #10B981; font-size: 24px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="success">✓ Success!</div>
    <p>Your email configuration is working correctly.</p>
    <p>Sent at: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
</body>
</html>
`,
	}

	err := s.emailService.Send(testData)
	if err != nil {
		return fmt.Errorf("email test failed: %w", err)
	}

	log.Printf("✅ Test email sent successfully to %s", testEmail)
	return nil
}
