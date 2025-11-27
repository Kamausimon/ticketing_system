# Email & Notifications System

This directory contains the email and notification system for the Ticketing System application. It supports both Mailtrap (for development) and Zoho (for production) email providers.

## Features

- 📧 **Multi-Provider Support**: Easily switch between Mailtrap, Zoho, or other SMTP providers
- 🎨 **HTML Email Templates**: Beautiful, responsive email templates for various notifications
- 🔄 **Auto-Retry**: Automatic retry mechanism for failed email sends
- 🔒 **Secure**: Support for TLS/SSL encrypted connections
- ✅ **Template System**: Reusable email templates with data binding
- 📊 **Logging**: Comprehensive logging for debugging and monitoring

## Email Templates

The system includes the following pre-built templates:

1. **Welcome Email** - Sent when a new user registers
2. **Email Verification** - For email address verification
3. **Password Reset** - For password reset requests
4. **Order Confirmation** - Sent after successful order placement
5. **Ticket Generated** - When tickets are generated for an order
6. **Event Reminder** - Reminder emails before events
7. **Payment Confirmation** - After successful payment processing
8. **Refund Processed** - When refunds are completed

## Setup

### 1. Mailtrap (Development)

Mailtrap is perfect for testing emails during development without sending real emails.

1. Sign up at [https://mailtrap.io](https://mailtrap.io)
2. Create an inbox
3. Get your SMTP credentials from the inbox settings
4. Update your `.env` file:

```env
EMAIL_PROVIDER=mailtrap
EMAIL_HOST=sandbox.smtp.mailtrap.io
EMAIL_PORT=2525
EMAIL_USERNAME=your_mailtrap_username
EMAIL_PASSWORD=your_mailtrap_password
EMAIL_FROM=noreply@ticketing.com
EMAIL_FROM_NAME=Ticketing System
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false
EMAIL_TEST_MODE=false
```

### 2. Zoho Mail (Production)

For production, you can use Zoho Mail or any other SMTP provider.

#### Zoho Setup Steps:

1. Sign up for Zoho Mail at [https://www.zoho.com/mail/](https://www.zoho.com/mail/)
2. Add and verify your domain
3. Create an email account (e.g., noreply@yourdomain.com)
4. Generate an App Password:
   - Go to Zoho Account Settings
   - Security > App Passwords
   - Generate a new app password for "Mail"
5. Update your `.env` file:

```env
EMAIL_PROVIDER=zoho
EMAIL_HOST=smtp.zoho.com
EMAIL_PORT=465
EMAIL_USERNAME=noreply@yourdomain.com
EMAIL_PASSWORD=your_zoho_app_password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Ticketing System
EMAIL_USE_TLS=false
EMAIL_USE_SSL=true
EMAIL_TEST_MODE=false
```

### 3. Other SMTP Providers

The system supports any SMTP provider. Common options:

**Gmail:**
```env
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false
```

**SendGrid:**
```env
EMAIL_HOST=smtp.sendgrid.net
EMAIL_PORT=587
EMAIL_USERNAME=apikey
EMAIL_PASSWORD=your_sendgrid_api_key
EMAIL_USE_TLS=true
```

**Amazon SES:**
```env
EMAIL_HOST=email-smtp.us-east-1.amazonaws.com
EMAIL_PORT=587
EMAIL_USE_TLS=true
```

## Usage

### Initialize the Notification Service

```go
import (
    "ticketing_system/internal/config"
    "ticketing_system/internal/notifications"
)

// Load configuration
cfg := config.LoadOrPanic()

// Create notification service
notifService := notifications.NewNotificationService(cfg)
```

### Send Welcome Email

```go
err := notifService.SendWelcomeEmail(
    "user@example.com",
    "John Doe",
)
if err != nil {
    log.Printf("Failed to send welcome email: %v", err)
}
```

### Send Order Confirmation

```go
orderData := notifications.OrderConfirmationData{
    CustomerName: "John Doe",
    OrderNumber:  "ORD-12345",
    EventName:    "Summer Music Festival",
    EventDate:    "2024-07-15",
    VenueName:    "Central Park",
    Items: []notifications.OrderItem{
        {
            Name:     "VIP Ticket",
            Quantity: 2,
            Price:    150.00,
            Currency: "USD",
        },
    },
    Currency: "USD",
    Total:    300.00,
}

err := notifService.SendOrderConfirmationEmail(
    "user@example.com",
    orderData,
)
```

### Send Custom Email

```go
// Plain text email
err := notifService.SendPlainEmail(
    []string{"user@example.com"},
    "Subject Here",
    "Email body here",
)

// HTML email
err := notifService.SendHTMLEmail(
    []string{"user@example.com"},
    "Subject Here",
    "<h1>Email body</h1><p>With HTML formatting</p>",
)
```

### Test Email Configuration

```go
err := notifService.TestEmailConfiguration("your-test-email@example.com")
if err != nil {
    log.Printf("Email configuration test failed: %v", err)
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `EMAIL_PROVIDER` | Email provider name (mailtrap, zoho, etc.) | mailtrap |
| `EMAIL_HOST` | SMTP host address | sandbox.smtp.mailtrap.io |
| `EMAIL_PORT` | SMTP port number | 2525 |
| `EMAIL_USERNAME` | SMTP username | - |
| `EMAIL_PASSWORD` | SMTP password | - |
| `EMAIL_FROM` | Default sender email address | noreply@ticketing.com |
| `EMAIL_FROM_NAME` | Default sender name | Ticketing System |
| `EMAIL_USE_TLS` | Enable STARTTLS | true |
| `EMAIL_USE_SSL` | Enable SSL/TLS | false |
| `EMAIL_TIMEOUT` | Connection timeout in seconds | 30 |
| `EMAIL_MAX_RETRIES` | Maximum retry attempts | 3 |
| `EMAIL_TEST_MODE` | Enable test mode (logs only) | false |

## Architecture

```
notifications/
├── email.go           # Core email service with SMTP handling
├── templates.go       # HTML email templates
├── notifications.go   # High-level notification service
└── README.md         # This file
```

### email.go
- SMTP connection handling
- TLS/SSL support
- Email sending with retry logic
- Template rendering

### templates.go
- HTML email templates
- Responsive design
- Consistent branding

### notifications.go
- High-level notification methods
- Business logic for different notification types
- Integration with config

## Best Practices

### Development
- Use Mailtrap to test all email functionality
- Enable `EMAIL_TEST_MODE` for unit tests
- Check spam scores before going to production

### Production
- Use a dedicated SMTP service (Zoho, SendGrid, SES)
- Set up SPF, DKIM, and DMARC records for your domain
- Monitor email delivery rates
- Set `EMAIL_TEST_MODE=false`
- Use strong passwords and app-specific passwords

### Security
- Never commit real credentials to version control
- Use environment variables for all sensitive data
- Rotate SMTP passwords regularly
- Use TLS/SSL for all connections
- Implement rate limiting for email sends

## Troubleshooting

### Emails Not Sending

1. **Check credentials**: Verify `EMAIL_USERNAME` and `EMAIL_PASSWORD`
2. **Check host and port**: Ensure correct SMTP server details
3. **Check TLS/SSL settings**: Match provider requirements
4. **Check logs**: Look for error messages in application logs
5. **Test configuration**: Use `TestEmailConfiguration()` method

### Emails Going to Spam

1. **Set up SPF record**: Add SPF record to DNS
2. **Configure DKIM**: Enable DKIM signing
3. **Add DMARC policy**: Set up DMARC record
4. **Verify domain**: Complete domain verification with email provider
5. **Monitor reputation**: Check sender reputation regularly

### Connection Timeouts

1. **Increase timeout**: Adjust `EMAIL_TIMEOUT` value
2. **Check firewall**: Ensure SMTP ports are not blocked
3. **Try different ports**: Test port 587, 465, or 2525
4. **Check provider status**: Verify SMTP service is operational

## API Reference

### NotificationService Methods

```go
// User onboarding
SendWelcomeEmail(email, name string) error
SendVerificationEmail(email, name, code string) error
SendPasswordResetEmail(email, name, token string) error

// Orders and tickets
SendOrderConfirmationEmail(email string, data OrderConfirmationData) error
SendTicketGeneratedEmail(email string, data TicketData) error

// Events
SendEventReminderEmail(email string, data EventReminderData) error

// Payments and refunds
SendPaymentConfirmationEmail(email string, data PaymentConfirmationData) error
SendRefundProcessedEmail(email string, data RefundData) error

// Custom emails
SendPlainEmail(to []string, subject, body string) error
SendHTMLEmail(to []string, subject, htmlBody string) error

// Testing
TestEmailConfiguration(testEmail string) error
```

## Future Enhancements

- [ ] Email queue for background processing
- [ ] Email templates in database
- [ ] A/B testing for email content
- [ ] Email analytics and tracking
- [ ] Unsubscribe management
- [ ] Multi-language support
- [ ] SMS notifications
- [ ] Push notifications
- [ ] Webhooks for delivery status

## Support

For issues or questions:
1. Check this documentation
2. Review application logs
3. Test with `TestEmailConfiguration()`
4. Check provider documentation (Mailtrap/Zoho)

## License

Part of the Ticketing System project.
