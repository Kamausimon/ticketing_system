# Email System Quick Start Guide

This guide will help you quickly set up and test the email system in the Ticketing System application.

## Quick Setup (5 minutes)

### Step 1: Get Mailtrap Credentials

1. Go to [https://mailtrap.io](https://mailtrap.io)
2. Sign up for a free account
3. Click on "Email Testing" → "Inboxes"
4. Create or select an inbox
5. Go to "SMTP Settings" tab
6. Copy your credentials

### Step 2: Configure Environment

Create or update your `.env` file:

```bash
# Copy the example file
cp .env.example .env

# Edit the .env file and update these values:
EMAIL_PROVIDER=mailtrap
EMAIL_HOST=sandbox.smtp.mailtrap.io
EMAIL_PORT=2525
EMAIL_USERNAME=your_username_from_mailtrap
EMAIL_PASSWORD=your_password_from_mailtrap
EMAIL_FROM=noreply@ticketing.com
EMAIL_FROM_NAME=Ticketing System
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false
EMAIL_TEST_MODE=false
```

### Step 3: Test the Configuration

#### Option A: Using the Test Endpoint

1. Start your server:
```bash
go run cmd/api-server/main.go
```

2. Send a test email:
```bash
curl -X POST http://localhost:8080/notifications/test \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

3. Check your Mailtrap inbox for the test email

#### Option B: Using Go Code

Create a test file `test_email.go`:

```go
package main

import (
	"log"
	"ticketing_system/internal/config"
	"ticketing_system/internal/notifications"
)

func main() {
	// Load configuration
	cfg := config.LoadOrPanic()

	// Create notification service
	notifService := notifications.NewNotificationService(cfg)

	// Test email configuration
	err := notifService.TestEmailConfiguration("test@example.com")
	if err != nil {
		log.Fatalf("Email test failed: %v", err)
	}

	log.Println("✅ Email test successful! Check your Mailtrap inbox.")
}
```

Run it:
```bash
go run test_email.go
```

## Testing Different Email Types

### Welcome Email

```bash
curl -X POST http://localhost:8080/notifications/welcome \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "name": "John Doe"
  }'
```

### Verification Email

```bash
curl -X POST http://localhost:8080/notifications/verification \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "code": "ABC123"
  }'
```

### Password Reset Email

```bash
curl -X POST http://localhost:8080/notifications/password-reset \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "token": "reset_token_here"
  }'
```

## Integrate with Your Code

### In User Registration

```go
// In your auth handler
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// ... user registration logic ...

	// Send welcome email
	go h.notificationService.SendWelcomeEmail(user.Email, user.FirstName)

	// ... rest of the handler ...
}
```

### In Password Reset

```go
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// ... generate reset token ...

	// Send password reset email
	go h.notificationService.SendPasswordResetEmail(
		user.Email,
		user.FirstName,
		resetToken,
	)

	// ... rest of the handler ...
}
```

### In Order Processing

```go
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// ... order creation logic ...

	// Send order confirmation
	orderData := notifications.OrderConfirmationData{
		CustomerName: customer.Name,
		OrderNumber:  order.OrderNumber,
		EventName:    event.Name,
		// ... fill in other fields ...
	}
	
	go h.notificationService.SendOrderConfirmationEmail(
		customer.Email,
		orderData,
	)

	// ... rest of the handler ...
}
```

## Moving to Production (Zoho)

When you're ready to go live:

### Step 1: Set up Zoho Mail

1. Sign up at [https://www.zoho.com/mail/](https://www.zoho.com/mail/)
2. Add and verify your domain
3. Create email account: `noreply@yourdomain.com`
4. Generate app password:
   - Account Settings → Security → App Passwords
   - Create password for "Mail"

### Step 2: Update Production Environment

```bash
# Production .env
EMAIL_PROVIDER=zoho
EMAIL_HOST=smtp.zoho.com
EMAIL_PORT=465
EMAIL_USERNAME=noreply@yourdomain.com
EMAIL_PASSWORD=your_zoho_app_password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your Company Name
EMAIL_USE_TLS=false
EMAIL_USE_SSL=true
EMAIL_TEST_MODE=false
```

### Step 3: Configure DNS Records

Add these DNS records for your domain:

**SPF Record:**
```
TXT @ "v=spf1 include:zoho.com ~all"
```

**DKIM Record:**
```
TXT zoho._domainkey "v=DKIM1; k=rsa; p=YOUR_PUBLIC_KEY_FROM_ZOHO"
```

**DMARC Record:**
```
TXT _dmarc "v=DMARC1; p=none; rua=mailto:postmaster@yourdomain.com"
```

### Step 4: Test Production Configuration

```bash
# Test with production config
curl -X POST https://your-domain.com/notifications/test \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-real-email@example.com"
  }'
```

## Troubleshooting

### Issue: "Authentication failed"
- ✅ Double-check username and password
- ✅ For Zoho, use app password, not account password
- ✅ Ensure no extra spaces in credentials

### Issue: "Connection timeout"
- ✅ Check `EMAIL_HOST` and `EMAIL_PORT`
- ✅ Verify firewall isn't blocking SMTP ports
- ✅ Try different ports (587, 465, 2525)

### Issue: "TLS handshake failed"
- ✅ Check `EMAIL_USE_TLS` and `EMAIL_USE_SSL` settings
- ✅ Port 587 usually uses TLS
- ✅ Port 465 usually uses SSL

### Issue: Emails not arriving
- ✅ Check Mailtrap inbox (it catches all test emails)
- ✅ Verify `EMAIL_TEST_MODE` is set to `false`
- ✅ Check application logs for errors
- ✅ Test with `TestEmailConfiguration()`

## Common Patterns

### Async Email Sending

Always send emails asynchronously to avoid blocking:

```go
go notificationService.SendWelcomeEmail(email, name)
```

### Error Handling

Log errors but don't fail the main operation:

```go
if err := notificationService.SendWelcomeEmail(email, name); err != nil {
    log.Printf("Failed to send welcome email: %v", err)
    // Continue with the operation
}
```

### Retry Logic

The service automatically retries failed sends up to `EMAIL_MAX_RETRIES` times.

## Next Steps

- [ ] Set up Mailtrap and test all email templates
- [ ] Integrate email sending into your handlers
- [ ] Set up production email provider (Zoho)
- [ ] Configure DNS records for production domain
- [ ] Set up email monitoring and alerts
- [ ] Consider implementing email queue for high volume

## Need Help?

- Check the full documentation in `internal/notifications/README.md`
- Review email templates in `internal/notifications/templates.go`
- Check application logs for detailed error messages
- Test configuration with `TestEmailConfiguration()`

## Pro Tips

💡 **Development**: Keep `EMAIL_TEST_MODE=false` and use Mailtrap to see actual emails

💡 **Testing**: Set `EMAIL_TEST_MODE=true` for unit tests to prevent actual sends

💡 **Production**: Always use SSL/TLS and strong authentication

💡 **Performance**: Send emails asynchronously with goroutines

💡 **Monitoring**: Log all email operations for debugging

💡 **Security**: Never commit real credentials to version control
