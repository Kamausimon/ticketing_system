package main

import (
	"fmt"
	"log"

	"ticketing_system/internal/config"
	"ticketing_system/internal/notifications"
)

/*
Example: Using Go's Built-in SMTP (No Third-Party Services)

This example shows how to use the email system with ANY SMTP server
using only Go's standard library `net/smtp` package.
*/

func main() {
	fmt.Println("📧 Go Built-in SMTP Email Example")
	fmt.Println("===================================\n")

	// Example 1: Gmail SMTP
	fmt.Println("Example 1: Gmail SMTP Configuration")
	gmailConfig := &config.EmailConfig{
		Provider:   "gmail",
		Host:       "smtp.gmail.com",
		Port:       587,
		Username:   "your-email@gmail.com",
		Password:   "your-app-password", // Use App Password, not regular password
		FromEmail:  "your-email@gmail.com",
		FromName:   "Your App Name",
		UseTLS:     true,
		UseSSL:     false,
		MaxRetries: 3,
		TestMode:   false,
	}
	fmt.Printf("✅ Gmail: %s:%d (TLS)\n\n", gmailConfig.Host, gmailConfig.Port)

	// Example 2: Outlook/Hotmail SMTP
	fmt.Println("Example 2: Outlook SMTP Configuration")
	outlookConfig := &config.EmailConfig{
		Provider:   "outlook",
		Host:       "smtp-mail.outlook.com",
		Port:       587,
		Username:   "your-email@outlook.com",
		Password:   "your-password",
		FromEmail:  "your-email@outlook.com",
		FromName:   "Your App Name",
		UseTLS:     true,
		UseSSL:     false,
		MaxRetries: 3,
		TestMode:   false,
	}
	fmt.Printf("✅ Outlook: %s:%d (TLS)\n\n", outlookConfig.Host, outlookConfig.Port)

	// Example 3: Custom/Local SMTP Server (no auth)
	fmt.Println("Example 3: Local SMTP Server (No Authentication)")
	localConfig := &config.EmailConfig{
		Provider:   "local",
		Host:       "localhost",
		Port:       25,
		Username:   "", // No auth required
		Password:   "", // No auth required
		FromEmail:  "noreply@localhost",
		FromName:   "Local Ticketing System",
		UseTLS:     false,
		UseSSL:     false,
		MaxRetries: 1,
		TestMode:   false,
	}
	fmt.Printf("✅ Local: %s:%d (No encryption, no auth)\n\n", localConfig.Host, localConfig.Port)

	// Example 4: Mailtrap (for testing only)
	fmt.Println("Example 4: Mailtrap SMTP Configuration (Testing)")
	mailtrapConfig := &config.EmailConfig{
		Provider:   "mailtrap",
		Host:       "smtp.mailtrap.io",
		Port:       587, // or 2525
		Username:   "your-mailtrap-username",
		Password:   "your-mailtrap-password",
		FromEmail:  "test@example.com",
		FromName:   "Test Sender",
		UseTLS:     true,
		UseSSL:     false,
		MaxRetries: 3,
		TestMode:   false,
	}
	fmt.Printf("✅ Mailtrap: %s:%d (TLS)\n\n", mailtrapConfig.Host, mailtrapConfig.Port)

	// Use the configuration you want
	cfg := &config.Config{
		Email: *gmailConfig, // Change this to use different config
	}

	// Create notification service
	notifService := notifications.NewNotificationService(cfg)

	// Send a simple test email
	fmt.Println("📤 Sending test email...")
	err := notifService.SendPlainEmail(
		[]string{"recipient@example.com"},
		"Test Email from Go SMTP",
		"This email was sent using Go's built-in net/smtp package!",
	)

	if err != nil {
		log.Printf("❌ Failed to send email: %v\n", err)
	} else {
		fmt.Println("✅ Email sent successfully!\n")
	}

	// Send an HTML email
	fmt.Println("📤 Sending HTML email...")
	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body { font-family: Arial, sans-serif; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			h1 { color: #4F46E5; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Hello from Go!</h1>
			<p>This is a <strong>beautiful HTML email</strong> sent using:</p>
			<ul>
				<li>Go's standard library (net/smtp)</li>
				<li>No third-party dependencies</li>
				<li>Any SMTP server you want</li>
			</ul>
			<p>Simple and powerful! 🚀</p>
		</div>
	</body>
	</html>
	`

	err = notifService.SendHTMLEmail(
		[]string{"recipient@example.com"},
		"HTML Email from Go SMTP",
		htmlBody,
	)

	if err != nil {
		log.Printf("❌ Failed to send HTML email: %v\n", err)
	} else {
		fmt.Println("✅ HTML email sent successfully!\n")
	}

	// Send using template
	fmt.Println("📤 Sending welcome email using template...")
	err = notifService.SendWelcomeEmail("newuser@example.com", "John Doe")
	if err != nil {
		log.Printf("❌ Failed to send welcome email: %v\n", err)
	} else {
		fmt.Println("✅ Welcome email sent!\n")
	}

	fmt.Println("🎉 Done! All emails sent using Go's built-in SMTP.")
}

/*
ENVIRONMENT VARIABLES (.env file):
===================================

# For Gmail
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=your-email@gmail.com
EMAIL_FROM_NAME=Your App
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false

# For Outlook/Hotmail
EMAIL_HOST=smtp-mail.outlook.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@outlook.com
EMAIL_PASSWORD=your-password
EMAIL_FROM=your-email@outlook.com
EMAIL_USE_TLS=true

# For Local SMTP Server (no auth)
EMAIL_HOST=localhost
EMAIL_PORT=25
EMAIL_FROM=noreply@localhost
EMAIL_USE_TLS=false
EMAIL_USE_SSL=false
EMAIL_USERNAME=
EMAIL_PASSWORD=

# For Any Custom SMTP Server
EMAIL_HOST=mail.yourserver.com
EMAIL_PORT=587
EMAIL_USERNAME=smtp_user
EMAIL_PASSWORD=smtp_pass
EMAIL_FROM=noreply@yourdomain.com
EMAIL_USE_TLS=true

NOTES:
======

1. Gmail requires "App Passwords" - not your regular password
   - Go to: https://myaccount.google.com/apppasswords
   - Generate an app password and use that

2. Common SMTP Ports:
   - 587: STARTTLS (most common, recommended)
   - 465: SSL/TLS
   - 25:  Plain (no encryption, local only)

3. Authentication is OPTIONAL
   - Leave USERNAME and PASSWORD empty for servers that don't need auth
   - Perfect for local development servers

4. The system uses Go's standard library only:
   - net/smtp for sending
   - html/template for rendering
   - crypto/tls for encryption
   - No external dependencies!

5. Easy to switch providers:
   - Just change the environment variables
   - No code changes needed
   - Works with ANY SMTP server

*/
