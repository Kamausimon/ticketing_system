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
		VerificationURL:  fmt.Sprintf("%s/verify-email?code=%s", s.config.App.BaseURL, code),
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

// OrganizerApprovalData holds data for organizer approval emails
type OrganizerApprovalData struct {
	OrganizerName  string
	OrganizerEmail string
	DashboardURL   string
}

// SendOrganizerApprovalEmail sends an approval email to an organizer
func (s *NotificationService) SendOrganizerApprovalEmail(email string, data OrganizerApprovalData) error {
	data.DashboardURL = fmt.Sprintf("%s/organizer/dashboard", s.config.App.FrontendURL)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Your Organizer Account Has Been Approved!",
		"organizer_approval",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send organizer approval email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Organizer approval email sent to %s", email)
	return nil
}

// OrganizerRejectionData holds data for organizer rejection emails
type OrganizerRejectionData struct {
	OrganizerName   string
	OrganizerEmail  string
	RejectionReason string
	ReapplyURL      string
	SupportEmail    string
}

// SendOrganizerRejectionEmail sends a rejection email to an organizer
func (s *NotificationService) SendOrganizerRejectionEmail(email string, data OrganizerRejectionData) error {
	data.ReapplyURL = fmt.Sprintf("%s/organizer/apply", s.config.App.FrontendURL)
	if data.SupportEmail == "" {
		data.SupportEmail = s.config.Email.FromEmail
	}

	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Organizer Account Application Status",
		"organizer_rejection",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send organizer rejection email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Organizer rejection email sent to %s", email)
	return nil
}

// WaitlistNotificationData holds data for waitlist notification emails
type WaitlistNotificationData struct {
	Name            string
	EventName       string
	EventDate       string
	VenueName       string
	TicketClassName string
	Quantity        int
	Price           float64
	Currency        string
	ExpiresAt       string
	PurchaseURL     string
}

// SendWaitlistNotificationEmail sends a notification when tickets become available
func (s *NotificationService) SendWaitlistNotificationEmail(email string, data WaitlistNotificationData) error {
	err := s.emailService.SendWithTemplate(
		[]string{email},
		fmt.Sprintf("Tickets Available: %s", data.EventName),
		"waitlist_notification",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send waitlist notification to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Waitlist notification sent to %s for %s", email, data.EventName)
	return nil
}

// OrganizerApplicationConfirmationData holds data for organizer application confirmation emails
type OrganizerApplicationConfirmationData struct {
	Name  string
	Email string
}

// SendOrganizerApplicationConfirmation sends a confirmation email to organizer after application
func (s *NotificationService) SendOrganizerApplicationConfirmation(email string, data OrganizerApplicationConfirmationData) error {
	err := s.emailService.SendWithTemplate(
		[]string{email},
		"Organizer Application Received",
		"organizer_application_confirmation",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send organizer application confirmation to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Organizer application confirmation sent to %s", email)
	return nil
}

// AdminOrganizerNotificationData holds data for admin notification about new organizer
type AdminOrganizerNotificationData struct {
	AdminName      string
	OrganizerName  string
	OrganizerEmail string
	OrganizerPhone string
	TaxName        string
	TaxPin         string
	AppliedDate    string
	ReviewURL      string
}

// SendAdminOrganizerNotification sends notification to admins about new organizer application
func (s *NotificationService) SendAdminOrganizerNotification(email string, data AdminOrganizerNotificationData) error {
	err := s.emailService.SendWithTemplate(
		[]string{email},
		"New Organizer Application - Action Required",
		"admin_organizer_notification",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send admin organizer notification to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Admin organizer notification sent to %s", email)
	return nil
}

// GetAdminReviewURL returns the URL for admin to review organizer application
func (s *NotificationService) GetAdminReviewURL(organizerID uint) string {
	return fmt.Sprintf("%s/admin/organizers/pending/%d", s.config.App.FrontendURL, organizerID)
}

// EmailVerificationData holds data for organizer email verification
type EmailVerificationData struct {
	Name             string
	Email            string
	VerificationLink string
	ExpiresAt        string
}

// SendOrganizerEmailVerification sends email verification to organizer
func (s *NotificationService) SendOrganizerEmailVerification(email string, data EmailVerificationData) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .button:hover { background: #45a049; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📧 Verify Your Email Address</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Thank you for applying to become an organizer! Please verify your email address to continue with the approval process.</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <a href="%s" class="button">Verify Email Address</a>
            </div>
            
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; background: #f0f0f0; padding: 10px; border-radius: 3px;">%s</p>
            
            <div class="warning">
                <strong>⚠️ Important:</strong>
                <ul>
                    <li>This link will expire on <strong>%s</strong></li>
                    <li>If you didn't apply to become an organizer, please ignore this email</li>
                    <li>For security, never share this verification link with anyone</li>
                </ul>
            </div>
            
            <p>After verification, our admin team will review your application and contact you via email.</p>
            
            <p style="margin-top: 30px;">Best regards,<br>The Ticketing System Team</p>
        </div>
        <div class="footer">
            <p>This is an automated email. Please do not reply.</p>
            <p>&copy; 2025 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, data.Name, data.VerificationLink, data.VerificationLink, data.ExpiresAt)

	emailData := EmailData{
		To:       []string{email},
		Subject:  "📧 Verify Your Organizer Email Address",
		HTMLBody: htmlBody,
		Body: fmt.Sprintf(`Verify Your Email Address

Hi %s,

Thank you for applying to become an organizer! Please verify your email address to continue with the approval process.

Verification Link: %s

This link will expire on %s.

If you didn't apply to become an organizer, please ignore this email.

Best regards,
The Ticketing System Team
`, data.Name, data.VerificationLink, data.ExpiresAt),
	}

	if err := s.emailService.Send(emailData); err != nil {
		log.Printf("❌ Failed to send email verification to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Email verification sent to %s", email)
	return nil
}

// OrganizerWelcomeData holds data for organizer welcome email
type OrganizerWelcomeData struct {
	OrganizerName  string
	OrganizerEmail string
}

// SendOrganizerWelcome sends a welcome email after email verification
func (s *NotificationService) SendOrganizerWelcome(email string, data OrganizerWelcomeData) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; padding: 15px; margin: 20px 0; border-radius: 5px; color: #155724; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Email Verified Successfully!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            
            <div class="success">
                <strong>✓ Your email has been verified successfully!</strong>
            </div>
            
            <p>Thank you for verifying your email address. Your organizer application is now under review by our admin team.</p>
            
            <h3>What's Next?</h3>
            <ul>
                <li><strong>Review Process:</strong> Our team will review your application within 2-3 business days</li>
                <li><strong>KYC Verification:</strong> You may be contacted for additional verification if needed</li>
                <li><strong>Approval Notification:</strong> You'll receive an email once your application is approved</li>
                <li><strong>Get Started:</strong> After approval, you can start creating and managing events</li>
            </ul>
            
            <p>If you have any questions, feel free to contact our support team.</p>
            
            <p style="margin-top: 30px;">Best regards,<br>The Ticketing System Team</p>
        </div>
        <div class="footer">
            <p>This is an automated notification.</p>
            <p>&copy; 2025 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, data.OrganizerName)

	emailData := EmailData{
		To:       []string{email},
		Subject:  "✅ Email Verified - Application Under Review",
		HTMLBody: htmlBody,
		Body: fmt.Sprintf(`Email Verified Successfully!

Hi %s,

Your email has been verified successfully!

Your organizer application is now under review by our admin team.

What's Next?
- Review Process: Our team will review your application within 2-3 business days
- KYC Verification: You may be contacted for additional verification if needed
- Approval Notification: You'll receive an email once your application is approved
- Get Started: After approval, you can start creating and managing events

If you have any questions, feel free to contact our support team.

Best regards,
The Ticketing System Team
`, data.OrganizerName),
	}

	if err := s.emailService.Send(emailData); err != nil {
		log.Printf("❌ Failed to send welcome email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Welcome email sent to %s", email)
	return nil
}

// SendBankDetailsChangeNotification sends notification when bank details are changed
func (s *NotificationService) SendBankDetailsChangeNotification(email string, data map[string]interface{}) error {
	name := data["Name"].(string)
	organizerName := data["OrganizerName"].(string)
	changedBy := data["ChangedBy"].(string)
	changedByEmail := data["ChangedByEmail"].(string)
	ipAddress := data["IPAddress"].(string)
	timestamp := time.Now().Format("January 2, 2006 at 3:04 PM MST")
	supportEmail := s.GetSupportEmail()

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #ff9800; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .alert { background: #fff3cd; border-left: 4px solid #ff9800; padding: 15px; margin: 20px 0; }
        .info-box { background: #fff; border: 1px solid #ddd; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .info-row { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-row:last-child { border-bottom: none; }
        .label { font-weight: bold; color: #666; display: inline-block; width: 150px; }
        .warning { background: #f8d7da; border: 1px solid #f5c2c7; padding: 15px; margin: 20px 0; border-radius: 5px; color: #842029; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏦 Bank Details Updated</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            
            <div class="alert">
                <strong>⚠️ Security Alert:</strong> The bank account details for <strong>%s</strong> have been updated.
            </div>
            
            <div class="info-box">
                <h3 style="margin-top: 0;">Change Details:</h3>
                <div class="info-row">
                    <span class="label">Changed By:</span> %s (%s)
                </div>
                <div class="info-row">
                    <span class="label">Timestamp:</span> %s
                </div>
                <div class="info-row">
                    <span class="label">IP Address:</span> %s
                </div>
            </div>
            
            <div class="warning">
                <strong>⚠️ Didn't make this change?</strong><br><br>
                If you did not authorize this change, please take immediate action:
                <ul style="margin: 10px 0;">
                    <li>Contact support immediately at <strong>%s</strong></li>
                    <li>Change your account password</li>
                    <li>Review your account activity</li>
                    <li>Check for any unauthorized access</li>
                </ul>
            </div>
            
            <p><strong>Why did I receive this?</strong></p>
            <p>All team members with access to the organizer account receive this notification for security purposes. This helps detect unauthorized changes and protects your payment information.</p>
            
            <p style="margin-top: 30px;">If you made this change, no action is required.</p>
            
            <p style="margin-top: 30px;">Best regards,<br>The Ticketing System Security Team</p>
        </div>
        <div class="footer">
            <p>This is a security notification. For your protection, we cannot disable these alerts.</p>
            <p>&copy; 2025 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, name, organizerName, changedBy, changedByEmail, timestamp, ipAddress, supportEmail)

	emailData := EmailData{
		To:       []string{email},
		Subject:  "🏦 Security Alert: Bank Details Updated",
		HTMLBody: htmlBody,
		Body: fmt.Sprintf(`Bank Details Updated - Security Alert

Hi %s,

SECURITY ALERT: The bank account details for %s have been updated.

Change Details:
- Changed By: %s (%s)
- Timestamp: %s
- IP Address: %s

DIDN'T MAKE THIS CHANGE?
If you did not authorize this change, please take immediate action:
1. Contact support immediately at %s
2. Change your account password
3. Review your account activity
4. Check for any unauthorized access

Why did I receive this?
All team members with access to the organizer account receive this notification for security purposes.

If you made this change, no action is required.

Best regards,
The Ticketing System Security Team
`, name, organizerName, changedBy, changedByEmail, timestamp, ipAddress, supportEmail),
	}

	if err := s.emailService.Send(emailData); err != nil {
		log.Printf("❌ Failed to send bank details change notification to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Bank details change notification sent to %s", email)
	return nil
}

// returns the support email address
func (s *NotificationService) GetSupportEmail() string {
	if s.config.Email.FromEmail != "" {
		return s.config.Email.FromEmail
	}
	return "support@ticketingsystem.com"
}

// SupportTicketCreatedData holds data for support ticket creation emails
type SupportTicketCreatedData struct {
	TicketID      uint
	TicketNumber  string
	Subject       string
	Description   string
	Category      string
	Priority      string
	CustomerName  string
	CustomerEmail string
	OrderID       *uint
	EventID       *uint
	CreatedAt     string
	DashboardURL  string
	AIClassified  bool
	AIPriority    string
	AIConfidence  int
	AIReasoning   string
}

// SendSupportTicketCreated sends an email to support team when a new ticket is created
func (s *NotificationService) SendSupportTicketCreated(data SupportTicketCreatedData) error {
	// Send to support email
	supportEmail := s.GetSupportEmail()

	err := s.emailService.SendWithTemplate(
		[]string{supportEmail},
		fmt.Sprintf("New Support Ticket #%s - %s", data.TicketNumber, data.Priority),
		"support_ticket_created",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send support ticket creation email: %v", err)
		return err
	}

	log.Printf("✅ Support ticket creation email sent for ticket #%s", data.TicketNumber)
	return nil
}

// SupportTicketStatusUpdateData holds data for ticket status update emails
type SupportTicketStatusUpdateData struct {
	TicketNumber    string
	Subject         string
	CustomerName    string
	OldStatus       string
	NewStatus       string
	Priority        string
	AssignedTo      string
	ResolutionNotes string
	ResolvedAt      string
	UpdatedAt       string
	TicketURL       string
	SupportEmail    string
}

// SendTicketStatusUpdate sends an email to the customer when ticket status is updated
func (s *NotificationService) SendTicketStatusUpdate(email string, data SupportTicketStatusUpdateData) error {
	data.SupportEmail = s.GetSupportEmail()

	subject := fmt.Sprintf("Ticket #%s Updated - %s", data.TicketNumber, data.NewStatus)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		subject,
		"support_ticket_status_update",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send ticket status update email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Ticket status update email sent to %s for ticket #%s", email, data.TicketNumber)
	return nil
}

// SupportTicketCommentData holds data for comment notification emails
type SupportTicketCommentData struct {
	TicketNumber  string
	Subject       string
	CustomerName  string
	Status        string
	CommentAuthor string
	Comment       string
	CommentTime   string
	TicketURL     string
	SupportEmail  string
}

// SendTicketCommentAdded sends an email when a comment is added to a ticket
func (s *NotificationService) SendTicketCommentAdded(email string, data SupportTicketCommentData) error {
	data.SupportEmail = s.GetSupportEmail()

	subject := fmt.Sprintf("New comment on Ticket #%s", data.TicketNumber)

	err := s.emailService.SendWithTemplate(
		[]string{email},
		subject,
		"support_ticket_comment_added",
		data,
	)

	if err != nil {
		log.Printf("❌ Failed to send comment notification email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ Comment notification email sent to %s for ticket #%s", email, data.TicketNumber)
	return nil
}
