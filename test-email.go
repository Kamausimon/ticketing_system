package main

import (
	"log"
	"os"
	"ticketing_system/internal/config"
	"ticketing_system/internal/notifications"
)

func main() {
	// Load email configuration from environment
	// Try SSL on port 465 if TLS on 587 is timing out
	useSSL := os.Getenv("TEST_USE_SSL") == "true"
	port := 587
	if useSSL {
		port = 465
	}

	emailConfig := &config.EmailConfig{
		Host:       os.Getenv("EMAIL_HOST"),
		Port:       port,
		Username:   os.Getenv("EMAIL_USERNAME"),
		Password:   os.Getenv("EMAIL_PASSWORD"),
		FromEmail:  os.Getenv("EMAIL_FROM"),
		FromName:   os.Getenv("EMAIL_FROM_NAME"),
		UseTLS:     !useSSL,
		UseSSL:     useSSL,
		Timeout:    60, // Increased timeout for slow connections
		MaxRetries: 3,
		TestMode:   false,
	}

	log.Printf("Testing email with config: Host=%s, Port=%d, Username=%s, UseTLS=%v",
		emailConfig.Host, emailConfig.Port, emailConfig.Username, emailConfig.UseTLS)

	// Create email service
	emailService := notifications.NewEmailService(emailConfig)

	// Test email data
	emailData := notifications.EmailData{
		To:       []string{os.Getenv("TEST_EMAIL_TO")},
		Subject:  "Test Email - Connection Fix",
		Body:     "This is a test email to verify the timeout fix is working correctly.",
		HTMLBody: "<h1>Test Email</h1><p>This is a test email to verify the timeout fix is working correctly.</p>",
	}

	log.Printf("Sending test email to: %v", emailData.To)

	// Send email
	if err := emailService.Send(emailData); err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
	}

	log.Println("✅ Email sent successfully!")
}
