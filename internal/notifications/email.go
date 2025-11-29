package notifications

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"

	"ticketing_system/internal/config"
)

// EmailService handles email operations
type EmailService struct {
	config *config.EmailConfig
	auth   smtp.Auth
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.EmailConfig) *EmailService {
	var auth smtp.Auth

	// Set up authentication using Go's built-in smtp.PlainAuth
	// This works with any SMTP server (Gmail, Outlook, custom servers, etc.)
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	// If no credentials provided, auth will be nil (for servers that don't require auth)

	return &EmailService{
		config: cfg,
		auth:   auth,
	}
}

// EmailData represents email content
type EmailData struct {
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Content  []byte
	MimeType string
}

// Send sends an email
func (s *EmailService) Send(data EmailData) error {
	if s.config.TestMode {
		log.Printf("📧 [TEST MODE] Email would be sent to: %v, Subject: %s", data.To, data.Subject)
		return nil
	}

	var lastErr error
	for i := 0; i < s.config.MaxRetries; i++ {
		err := s.sendWithRetry(data)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("⚠️ Email send attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", s.config.MaxRetries, lastErr)
}

// sendWithRetry sends an email with a single attempt
func (s *EmailService) sendWithRetry(data EmailData) error {
	// Build message
	msg := s.buildMessage(data)

	// Get server address
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Send email based on TLS/SSL configuration
	if s.config.UseSSL {
		return s.sendWithSSL(addr, data.To, msg)
	} else if s.config.UseTLS {
		return s.sendWithTLS(addr, data.To, msg)
	}

	// Send without encryption (not recommended for production)
	return smtp.SendMail(addr, s.auth, s.config.FromEmail, data.To, msg)
}

// sendWithTLS sends email using STARTTLS
func (s *EmailService) sendWithTLS(addr string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(s.auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	// Send data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendWithSSL sends email using SSL/TLS
func (s *EmailService) sendWithSSL(addr string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSL connection: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(s.auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	// Send data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// buildMessage builds the email message
func (s *EmailService) buildMessage(data EmailData) []byte {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", joinEmails(data.To)))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// If HTML body is provided, use multipart
	if data.HTMLBody != "" {
		boundary := "boundary-ticketing-system"
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		// Plain text part
		if data.Body != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(data.Body)
			buf.WriteString("\r\n")
		}

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(data.HTMLBody)
		buf.WriteString("\r\n")
		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Plain text only
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(data.Body)
	}

	return buf.Bytes()
}

// SendWithTemplate sends an email using a template
func (s *EmailService) SendWithTemplate(to []string, subject string, templateName string, data interface{}) error {
	tmpl, err := s.getTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	emailData := EmailData{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBuf.String(),
	}

	return s.Send(emailData)
}

// getTemplate retrieves an email template
func (s *EmailService) getTemplate(name string) (*template.Template, error) {
	templates := map[string]string{
		"welcome":              welcomeTemplate,
		"verification":         verificationTemplate,
		"password_reset":       passwordResetTemplate,
		"order_confirmation":   orderConfirmationTemplate,
		"ticket_generated":     ticketGeneratedTemplate,
		"event_reminder":       eventReminderTemplate,
		"payment_confirmation": paymentConfirmationTemplate,
		"refund_processed":     refundProcessedTemplate,
		"organizer_approval":   organizerApprovalTemplate,
		"organizer_rejection":  organizerRejectionTemplate,
	}

	tmplStr, exists := templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}

	return template.New(name).Parse(tmplStr)
}

// Helper function to join email addresses
func joinEmails(emails []string) string {
	if len(emails) == 0 {
		return ""
	}
	result := emails[0]
	for i := 1; i < len(emails); i++ {
		result += ", " + emails[i]
	}
	return result
}
