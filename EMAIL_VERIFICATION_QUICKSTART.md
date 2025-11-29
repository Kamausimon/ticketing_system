# Email Verification - Quick Start for Developers

## What's New? 🚀

Email verification is now required before users can:
- ✅ Download ticket PDFs
- ✅ Transfer tickets to others
- ✅ Generate tickets

## Key Changes at a Glance

### Registration (No Changes to Frontend)
```
User registers → Email sent automatically → Must verify email
```

### Ticket Operations (Now Protected)
```
Ticket download/transfer → Check: Email verified? → ✅ Success or ❌ 403 Error
```

### New API Endpoints
```
POST /verify-email          - Verify with token
POST /resend-verification   - Resend email
GET /verify-email/status    - Check status
```

---

## For Frontend Developers

### 1. Registration Page (Minimal Changes)
After successful registration (201):
```javascript
// Show message
showMessage("Registration successful! Check your email to verify.")
// Redirect to verification page
redirectTo('/verify-email')
```

### 2. Email Verification Page (New Component)
```javascript
// Get token from URL: ?code=abc123...
const token = new URLSearchParams(window.location.search).get('code')

// Verify email
fetch('/verify-email', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ token })
})
.then(r => r.json())
.then(data => {
  if (data.message) {
    showSuccess('Email verified! You can now download tickets.')
    redirectTo('/dashboard')
  } else {
    showError('Invalid or expired token.')
    // Show resend button
  }
})
```

### 3. Resend Verification (New Feature)
```javascript
// User clicks "Resend Email"
fetch('/resend-verification', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: userEmail })
})
.then(r => r.json())
.then(data => {
  if (r.ok) showSuccess('Email resent!')
  else showError(data.error)
})
```

### 4. Ticket Operations (Still the Same)
```javascript
// Download ticket - same as before
// If user not verified: 403 Forbidden + clear error message
// User sees: "Please verify your email to download tickets"
```

### 5. Account Status Page (New Info)
```javascript
// Check verification status
fetch('/verify-email/status', {
  headers: { 'Authorization': `Bearer ${token}` }
})
.then(r => r.json())
.then(data => {
  if (data.email_verified) {
    showBadge('✅ Email Verified', 'green')
  } else {
    showBanner('⚠️ Verify your email to access all features', 'yellow')
  }
})
```

---

## For Backend Developers

### Database Migration
```go
// Already handled in main.go
DB.AutoMigrate(&models.User{}, &models.EmailVerification{})
```

### Using Email Verification Middleware
```go
// Protect an endpoint
router.Handle("/tickets/download", 
  emailVerificationMiddleware(http.HandlerFunc(handler)))
```

### Getting User Verification Status
```go
var user models.User
DB.First(&user, userID)

if !user.EmailVerified {
  // User not verified
}
```

### Checking Verification Status in Code
```go
// In your handler
var verification models.EmailVerification
DB.Where("user_id = ? AND status = ?", userID, models.VerificationVerified).
   First(&verification)
```

---

## Testing Locally

### 1. Enable Test Mode (Don't Send Real Emails)
```bash
# In .env
SMTP_TEST_MODE=true
```

### 2. Test Registration
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe",
    "phone": "+254712345678",
    "email": "test@example.com",
    "password": "test123"
  }'

# Response: 201 Created
# {"message":"user registered successfully...", "user_id":1, "email":"test@example.com"}
```

### 3. Verify Email
```bash
# From the console logs, get the token
# Then verify:
curl -X POST http://localhost:8080/verify-email \
  -H "Content-Type: application/json" \
  -d '{"token":"abc123def456..."}'

# Response: 200 OK
# {"message":"email verified successfully", "email":"test@example.com"}
```

### 4. Try Protected Endpoint (Unverified)
```bash
# Without verifying, try to download
curl -X GET http://localhost:8080/tickets/1/pdf \
  -H "Authorization: Bearer {jwt_token}"

# Response: 403 Forbidden
# {"error":"email verification required. Please verify..."}
```

### 5. Try After Verification
```bash
# After verifying email, same request works
curl -X GET http://localhost:8080/tickets/1/pdf \
  -H "Authorization: Bearer {jwt_token}"

# Response: 200 OK (with PDF file)
```

---

## Error Messages Users Will See

| Scenario | Error Message |
|----------|---------------|
| Trying to download unverified | "Email verification required. Please verify your email address to perform this action" |
| Invalid token | "Invalid or expired verification token" |
| Expired token (24hrs) | "Verification token has expired" |
| Already verified | "Email already verified" |
| Email doesn't exist | "User not found" |
| Resend too soon | "Please wait before requesting a new verification email" (5-min cooldown) |
| Too many resends | "Maximum resend attempts reached. Please contact support" |

---

## Configuration (.env)

### Required for Email
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM_EMAIL=noreply@ticketing.com
SMTP_FROM_NAME=Ticketing System
SMTP_USE_TLS=true
SMTP_TEST_MODE=false          # Set true to prevent sending
```

### Frontend URL
```env
FRONTEND_URL=http://localhost:3000  # Local dev
# OR
FRONTEND_URL=https://ticketing.com   # Production
```

---

## Common Issues & Fixes

### ❌ "Email not received"
- ✅ Check SMTP_TEST_MODE=true in dev
- ✅ Check spam folder
- ✅ Verify SMTP credentials
- ✅ Check email logs

### ❌ "Token expired immediately"
- ✅ Tokens valid 24 hours - user must request new one
- ✅ Show resend button if expired

### ❌ "Can't download tickets after verification"
- ✅ Check GET /verify-email/status returns true
- ✅ Verify database shows email_verified=true
- ✅ Check JWT token is valid

### ❌ "Build fails - models not found"
- ✅ Run: `go mod tidy`
- ✅ Run: `go mod download`
- ✅ Rebuild with `go build`

---

## File Locations

### New Files Created
```
EMAIL_VERIFICATION_IMPLEMENTATION.md  - Full technical docs
EMAIL_VERIFICATION_API.md             - API reference
EMAIL_VERIFICATION_SUMMARY.md         - This implementation summary
internal/models/emailVerification.go  - New database model
```

### Modified Files
```
internal/auth/main.go                 - Registration + verification
internal/models/user.go               - Added verification fields
internal/middleware/main.go           - Added verification middleware
cmd/api-server/main.go                - Route registration
```

---

## Quick Reference: New Endpoints

```
POST /register
  Input: {first_name, last_name, username, phone, email, password}
  Output: {message, user_id, email}
  Auto-action: Sends verification email ✉️

POST /verify-email
  Input: {token}
  Output: {message, email}
  Effect: Marks email as verified

POST /resend-verification
  Input: {email}
  Output: {message, email}
  Rate-limit: 5-minute cooldown, max 3 attempts

GET /verify-email/status
  Output: {email_verified: bool, email, verified_at}
  Requires: Valid JWT token
```

---

## Enforcement Map

| Operation | Public | Auth Required | Email Verified |
|-----------|--------|---------------|-----------------|
| Register | ✅ | ❌ | ❌ |
| Login | ✅ | ❌ | ❌ |
| View tickets | ✅ | ✅ | ❌ |
| Download ticket | ✅ | ✅ | ✅ ⚠️ |
| Transfer ticket | ✅ | ✅ | ✅ ⚠️ |
| Generate ticket | ✅ | ✅ | ✅ ⚠️ |

⚠️ = NEW REQUIREMENT

---

## Next Steps

1. ✅ Code review (ready to review)
2. ⏳ Set up test email account
3. ⏳ Update frontend components
4. ⏳ Test full user journey
5. ⏳ Deploy to staging
6. ⏳ Deploy to production

---

**Questions?** Check full docs in `EMAIL_VERIFICATION_API.md` or `EMAIL_VERIFICATION_IMPLEMENTATION.md`

**Status:** ✅ Ready for Testing
**Version:** 1.0
**Date:** November 29, 2025
