# Email Verification API Quick Reference

## Overview
Email verification ensures users provide valid email addresses during registration and prevents unverified users from downloading or transferring tickets.

## Endpoints

### 1. Register User
```http
POST /register
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "phone": "+254712345678",
  "email": "john@example.com",
  "password": "SecurePassword123"
}

Response 201 Created:
{
  "message": "user registered successfully. Please check your email to verify your account",
  "user_id": 1,
  "email": "john@example.com"
}
```

**What Happens:**
1. User account created with `email_verified = false`
2. 32-character verification token generated
3. EmailVerification record created with 24-hour expiry
4. Verification email automatically sent to user
5. User can click link in email or copy token to verify

---

### 2. Verify Email
```http
POST /verify-email
Content-Type: application/json

{
  "token": "abc123def456abc123def456abc123de"
}

Response 200 OK:
{
  "message": "email verified successfully",
  "email": "john@example.com"
}

Response 400 Bad Request (Expired Token):
{
  "error": "verification token has expired"
}

Response 400 Bad Request (Invalid Token):
{
  "error": "invalid or expired verification token"
}

Response 409 Conflict (Already Verified):
{
  "error": "email already verified"
}
```

**What This Does:**
- Validates the token
- Checks expiration (24 hours)
- Marks email as verified in User record
- Updates verification status in EmailVerification record
- User can now download tickets and perform transfers

---

### 3. Resend Verification Email
```http
POST /resend-verification
Content-Type: application/json

{
  "email": "john@example.com"
}

Response 200 OK:
{
  "message": "verification email resent successfully",
  "email": "john@example.com"
}

Response 400 Bad Request (No Pending Verification):
{
  "error": "no pending verification found"
}

Response 404 Not Found (User Not Found):
{
  "error": "user not found"
}

Response 409 Conflict (Already Verified):
{
  "error": "email already verified"
}

Response 429 Too Many Requests (Rate Limited):
{
  "error": "please wait before requesting a new verification email"
}
(Wait 5 minutes before requesting another)

Response 429 Too Many Requests (Max Attempts):
{
  "error": "maximum resend attempts reached. Please contact support"
}
(Max 3 resend attempts allowed)
```

**What This Does:**
- Generates new verification token
- Sends new verification email
- Enforces 5-minute cooldown between resends
- Tracks resend count (max 3 attempts)
- If max attempts exceeded, user must contact support

---

### 4. Check Email Verification Status
```http
GET /verify-email/status
Authorization: Bearer {jwt_token}

Response 200 OK:
{
  "email_verified": true,
  "email": "john@example.com",
  "verified_at": "2025-11-29T10:30:45Z"
}

Response 200 OK (Not Verified):
{
  "email_verified": false,
  "email": "john@example.com",
  "verified_at": null
}

Response 401 Unauthorized:
{
  "error": "unauthorized"
}
```

**What This Does:**
- Returns current email verification status
- Shows when email was verified (if applicable)
- Available only to authenticated users

---

## Protected Endpoints (Require Email Verification)

These endpoints will return `403 Forbidden` if the user's email is not verified:

### Ticket Operations
- `POST /tickets/generate` - Generate tickets for event
- `POST /tickets/regenerate-qr` - Regenerate QR code
- `GET /tickets/{id}/pdf` - Download ticket PDF
- `POST /tickets/{id}/transfer` - Transfer ticket to another user

**Error Response (403 Forbidden):**
```json
{
  "error": "email verification required. Please verify your email address to perform this action"
}
```

---

## Email Content

### Verification Email Subject
```
Verify Your Email Address
```

### Verification Email Body (HTML)
```
Dear [First Name],

Please verify your email address to complete your registration.

Verification Link:
https://ticketing.yoursite.com/verify-email?code=[TOKEN]

Or copy this code to verify:
[TOKEN]

This link will expire in 24 hours.

If you didn't create this account, please ignore this email.

Best regards,
Ticketing System Team
```

---

## Implementation Flow

### User Registration Flow
```
1. User fills registration form
   ↓
2. POST /register
   ↓
3. System creates user (unverified)
   ↓
4. System generates verification token
   ↓
5. System sends verification email
   ↓
6. Return 201 with message asking to check email
   ↓
7. User receives email with verification link
   ↓
8. User clicks link or copies token
   ↓
9. POST /verify-email with token
   ↓
10. System marks email as verified
   ↓
11. User can now download tickets and transfer
```

### User Needs Resend Flow
```
1. User clicks "Resend verification" link
   ↓
2. Enters email address
   ↓
3. POST /resend-verification
   ↓
4. System checks:
   - Email exists? ✓
   - Not already verified? ✓
   - 5 minutes since last send? ✓
   - Less than 3 resends? ✓
   ↓
5. Generate new token
   ↓
6. Send new email
   ↓
7. Return 200 OK
   ↓
8. User receives new verification email
```

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/register` | 10 requests | per minute per IP |
| `/verify-email` | 10 requests | per minute per IP |
| `/resend-verification` | 5 minute cooldown | between requests |
| `/verify-email/status` | 10 requests | per minute per IP |

---

## Security Details

✅ **Token Generation**
- Cryptographically secure random tokens
- 32-character hex-encoded strings
- Unique per user registration

✅ **Token Validity**
- Valid for 24 hours from issue
- One-time use (marked as verified)
- Expired tokens logged and marked invalid

✅ **Rate Limiting**
- Per-IP rate limiting on registration
- 5-minute cooldown on resends
- Maximum 3 resend attempts
- After 3 attempts, must contact support

✅ **Audit Trail**
- IP address stored with each token
- User agent stored for fraud detection
- Timestamp of each verification attempt
- Status changes tracked in database

---

## Error Codes Reference

| Status | Code | Meaning | Action |
|--------|------|---------|--------|
| 200 | OK | Success | Proceed |
| 201 | Created | User created | Check email for verification |
| 400 | Bad Request | Invalid input/expired token | Resend verification or fix input |
| 401 | Unauthorized | Not authenticated | Login first |
| 403 | Forbidden | Email not verified | Verify email to proceed |
| 404 | Not Found | User/email not found | Check email address |
| 409 | Conflict | Email already verified/exists | Use different email or login |
| 429 | Too Many Requests | Rate limited | Wait 5 minutes or contact support |
| 500 | Server Error | System error | Retry or contact support |

---

## Frontend Integration Examples

### Register Component
```javascript
async function registerUser(formData) {
  const response = await fetch('/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(formData)
  });
  
  if (response.ok) {
    showMessage('Registration successful! Check your email to verify.');
    redirectTo('/verify-email');
  }
}
```

### Verify Email Component
```javascript
async function verifyEmail(token) {
  const response = await fetch('/verify-email', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
  
  if (response.ok) {
    showMessage('Email verified! You can now download tickets.');
    redirectTo('/dashboard');
  } else {
    showError('Invalid or expired token. Request a new one.');
  }
}
```

### Check Status Component
```javascript
async function checkEmailStatus() {
  const token = getAuthToken(); // Get JWT from localStorage
  
  const response = await fetch('/verify-email/status', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const data = await response.json();
  
  if (!data.email_verified) {
    showBanner('Please verify your email to access all features');
  }
}
```

---

## Troubleshooting

### Q: User says they didn't receive verification email
**A:** 
1. Check spam folder
2. Verify email address is correct
3. Call `POST /resend-verification` to send again
4. Check email service logs for bounces

### Q: "Token expired" error
**A:** Tokens valid for 24 hours. User must request new verification via resend endpoint.

### Q: "Maximum resend attempts reached"
**A:** User exceeded 3 resends. Direct them to support team for manual verification.

### Q: User can't download tickets after verification
**A:**
1. Check status with `GET /verify-email/status`
2. Verify response shows `email_verified: true`
3. If still failing, check user record in database

---

**Last Updated:** November 29, 2025  
**Version:** 1.0
