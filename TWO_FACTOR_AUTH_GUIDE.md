# Two-Factor Authentication (2FA) Implementation Guide

## Overview

This ticketing system now includes a complete **Two-Factor Authentication (2FA)** implementation using **TOTP (Time-based One-Time Password)** protocol. This is the same standard used by Google Authenticator, Authy, Microsoft Authenticator, and other authenticator apps.

### Key Features

✅ **TOTP-based authentication** (RFC 6238 compliant)
✅ **QR code generation** for easy setup
✅ **Recovery codes** for account recovery
✅ **Rate limiting** on all 2FA endpoints
✅ **Comprehensive attempt logging** for security auditing
✅ **Middleware support** for protecting high-security routes
✅ **Graceful 2FA requirement** during login
✅ **Recovery code regeneration** capability

---

## Architecture

### Database Models

#### 1. `TwoFactorAuth` Table
Stores the main 2FA configuration for users:
- `user_id` - Links to the user
- `enabled` - Whether 2FA is active
- `secret` - Encrypted TOTP secret
- `verified_at` - When 2FA was first enabled
- `last_used_at` - Last successful 2FA verification
- `method` - Authentication method (currently only 'totp')

#### 2. `RecoveryCode` Table
Stores hashed recovery codes:
- `two_factor_auth_id` - Links to 2FA config
- `code_hash` - Bcrypt hashed recovery code
- `used` - Whether the code has been used
- `used_at` - When the code was used
- `used_from_ip` - IP address where code was used

#### 3. `TwoFactorAttempt` Table
Logs all 2FA verification attempts:
- `user_id` - User attempting verification
- `success` - Whether attempt succeeded
- `ip_address` - Source IP
- `failure_type` - Reason for failure
- `attempted_at` - Timestamp

#### 4. `TwoFactorSession` Table
Temporary storage during 2FA setup:
- `user_id` - User setting up 2FA
- `secret` - Temporary secret (until verified)
- `expires_at` - Session expiration
- `verified` - Whether setup is complete

---

## API Endpoints

### 1. Setup 2FA

**POST** `/2fa/setup`

Initiates 2FA setup for an authenticated user.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "password": "user_password"
}
```

**Response (200 OK):**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code_data": "data:image/png;base64,...",
  "qr_code_url": "otpauth://totp/Ticketing%20System:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Ticketing%20System&algorithm=SHA1&digits=6&period=30",
  "backup_codes": [
    "A1B2C3D4-E5F6G7H8",
    "I9J0K1L2-M3N4O5P6",
    ...
  ]
}
```

**Security:**
- Requires current password confirmation
- Returns secret and QR code **only once**
- Recovery codes shown **only once** - user must save them
- Rate limited (10 req/min per IP)

---

### 2. Verify Setup

**POST** `/2fa/verify-setup`

Verifies the TOTP code to complete 2FA setup.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "Two-factor authentication has been successfully enabled",
  "enabled": true
}
```

**Security:**
- Must be called within 15 minutes of setup
- Validates TOTP code with clock skew tolerance
- Enables 2FA only after successful verification

---

### 3. Login (Updated Flow)

**POST** `/login`

Standard login now checks for 2FA.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response without 2FA (200 OK):**
```json
{
  "message": "Login successful",
  "user_id": 123,
  "role": "organizer",
  "token": "eyJhbGc..."
}
```

**Response with 2FA enabled (200 OK):**
```json
{
  "message": "Two-factor authentication required",
  "requires_2fa": true,
  "temp_token": "eyJhbGc...",
  "user_id": 123
}
```

**Flow:**
1. User submits email + password
2. If 2FA is disabled → return full access token
3. If 2FA is enabled → return temporary token (15 min expiry)
4. Client must call `/2fa/verify-login` with TOTP code

---

### 4. Verify Login (2FA)

**POST** `/2fa/verify-login`

Verifies 2FA code during login and issues full access token.

**Headers:**
```
Authorization: Bearer <TEMP_TOKEN>
```

**Request Body:**
```json
{
  "code": "123456",
  "is_recovery_code": false,
  "trust_device": false
}
```

**Response (200 OK):**
```json
{
  "message": "Two-factor authentication verified successfully",
  "verified": true,
  "token": "eyJhbGc...",
  "user_id": 123,
  "role": "organizer"
}
```

**Using Recovery Code:**
```json
{
  "code": "A1B2C3D4-E5F6G7H8",
  "is_recovery_code": true
}
```

**Security:**
- Rate limited (5 req/min per IP)
- Accepts TOTP codes or recovery codes
- Recovery codes are single-use and marked as used
- Failed attempts are logged for security monitoring

---

### 5. Check 2FA Status

**GET** `/2fa/status`

Returns current 2FA status for authenticated user.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200 OK):**
```json
{
  "enabled": true,
  "verified_at": "2025-11-30T10:30:00Z",
  "last_used_at": "2025-11-30T15:45:00Z",
  "method": "totp",
  "recovery_codes_remaining": 8
}
```

---

### 6. Disable 2FA

**POST** `/2fa/disable`

Disables 2FA for the user.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "password": "user_password",
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "Two-factor authentication has been disabled",
  "enabled": false
}
```

**Security:**
- Requires both password and current TOTP code
- Deletes all recovery codes
- Rate limited

---

### 7. Regenerate Recovery Codes

**POST** `/2fa/recovery-codes`

Generates new recovery codes (invalidates old ones).

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "password": "user_password",
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "Recovery codes regenerated successfully",
  "recovery_codes": [
    "X1Y2Z3A4-B5C6D7E8",
    "F9G0H1I2-J3K4L5M6",
    ...
  ]
}
```

**Security:**
- Requires password and TOTP code
- Old recovery codes are deleted
- New codes shown only once

---

### 8. View Recent Attempts

**GET** `/2fa/attempts`

Returns recent 2FA verification attempts (for debugging/security).

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200 OK):**
```json
{
  "attempts": [
    {
      "id": 1,
      "user_id": 123,
      "success": true,
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "failure_type": "",
      "attempted_at": "2025-11-30T15:45:00Z"
    },
    {
      "id": 2,
      "user_id": 123,
      "success": false,
      "ip_address": "192.168.1.100",
      "failure_type": "invalid_code",
      "attempted_at": "2025-11-30T15:44:30Z"
    }
  ],
  "count": 2
}
```

---

## Client Implementation Guide

### Step 1: Enable 2FA

```javascript
// 1. Initiate setup
const setupResponse = await fetch('/2fa/setup', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    password: userPassword
  })
});

const { secret, qr_code_data, backup_codes } = await setupResponse.json();

// 2. Display QR code for user to scan
displayQRCode(qr_code_data);

// 3. Display backup codes and warn user to save them
displayBackupCodes(backup_codes);

// 4. User enters code from authenticator app
const verifyResponse = await fetch('/2fa/verify-setup', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    code: userEnteredCode
  })
});

// 2FA is now enabled!
```

### Step 2: Login with 2FA

```javascript
// 1. Standard login
const loginResponse = await fetch('/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: userEmail,
    password: userPassword
  })
});

const loginData = await loginResponse.json();

// 2. Check if 2FA is required
if (loginData.requires_2fa) {
  // Store temporary token
  const tempToken = loginData.temp_token;
  
  // Show 2FA input form
  const code = await prompt2FACode();
  
  // 3. Verify 2FA
  const verifyResponse = await fetch('/2fa/verify-login', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${tempToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      code: code,
      is_recovery_code: false
    })
  });
  
  const verifyData = await verifyResponse.json();
  
  // Store full access token
  storeToken(verifyData.token);
} else {
  // No 2FA, use token directly
  storeToken(loginData.token);
}
```

### Step 3: Using Recovery Codes

```javascript
// If user lost their authenticator app
const verifyResponse = await fetch('/2fa/verify-login', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${tempToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    code: "A1B2C3D4-E5F6G7H8",
    is_recovery_code: true
  })
});
```

---

## Middleware Usage

### Require 2FA for Specific Routes

Use the `Require2FA` middleware to enforce 2FA for high-security operations:

```go
// Example: Protect settlement processing
router.Handle("/settlements/{id}/process", 
    middleware.Require2FA(DB)(
        http.HandlerFunc(settlementHandler.ProcessSettlement)
    )
).Methods(http.MethodPost)
```

### Require Organizer with 2FA

For organizer-specific operations that need 2FA:

```go
// Example: Protect payout operations
router.Handle("/organizers/payouts/initiate",
    middleware.RequireOrganizerWith2FA(DB)(
        http.HandlerFunc(payoutHandler.InitiatePayout)
    )
).Methods(http.MethodPost)
```

### Recommend 2FA (Non-blocking)

Add a header suggesting 2FA without blocking access:

```go
router.Handle("/account/settings",
    middleware.Recommend2FA(DB)(
        http.HandlerFunc(accountHandler.GetAccountSettings)
    )
).Methods(http.MethodGet)
```

---

## Security Features

### 1. TOTP Algorithm
- **Standard:** RFC 6238 compliant
- **Hash:** HMAC-SHA1
- **Time Step:** 30 seconds
- **Code Length:** 6 digits
- **Clock Skew:** ±1 time step (30 seconds tolerance)

### 2. Secret Generation
- **Length:** 160 bits (20 bytes)
- **Encoding:** Base32 without padding
- **Entropy:** Cryptographically secure random generation

### 3. Recovery Codes
- **Count:** 10 codes generated per setup
- **Format:** XXXX-XXXX (8 hex characters)
- **Storage:** Bcrypt hashed (cost 12)
- **Single-use:** Automatically marked as used
- **Regeneration:** User can generate new codes anytime

### 4. Rate Limiting
- **Setup endpoints:** 10 req/min per IP
- **Login verification:** 5 req/min per IP
- **Standard operations:** 10 req/min per IP

### 5. Attempt Logging
All verification attempts are logged with:
- Timestamp
- IP address
- User agent
- Success/failure status
- Failure reason
- Used for security monitoring and fraud detection

### 6. Session Security
- **Temporary tokens:** 15-minute expiry during 2FA setup
- **Setup sessions:** Auto-expire after 15 minutes
- **Password confirmation:** Required for setup and disable operations

---

## Testing

### Manual Testing with Authenticator App

1. **Download an authenticator app:**
   - Google Authenticator
   - Authy
   - Microsoft Authenticator
   - 1Password

2. **Enable 2FA:**
   ```bash
   curl -X POST http://localhost:8080/2fa/setup \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"password": "your_password"}'
   ```

3. **Scan QR code** from response with authenticator app

4. **Verify setup:**
   ```bash
   curl -X POST http://localhost:8080/2fa/verify-setup \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"code": "123456"}'
   ```

5. **Test login:**
   ```bash
   # Step 1: Login
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "password": "password"}'
   
   # Step 2: Verify 2FA
   curl -X POST http://localhost:8080/2fa/verify-login \
     -H "Authorization: Bearer TEMP_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"code": "123456"}'
   ```

### Test Recovery Codes

```bash
# Use a recovery code instead of TOTP
curl -X POST http://localhost:8080/2fa/verify-login \
  -H "Authorization: Bearer TEMP_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "A1B2C3D4-E5F6G7H8", "is_recovery_code": true}'
```

---

## Business Value

### For High-Value Accounts (Organizers)

1. **Financial Protection:**
   - Protect settlement withdrawals
   - Secure payout operations
   - Prevent unauthorized refunds

2. **Account Security:**
   - Prevent account takeover
   - Protect sensitive organizer data
   - Secure API key management

3. **Compliance:**
   - Meet security requirements for payment processing
   - Demonstrate security best practices
   - Reduce liability

### Recommended 2FA Enforcement

**Mandatory for:**
- ✅ Organizer accounts handling payments
- ✅ Settlement processing
- ✅ Payout initiation
- ✅ Refund approvals over threshold

**Recommended for:**
- ⚠️ Admin accounts
- ⚠️ Customer accounts with order history
- ⚠️ API access

**Optional for:**
- 📝 Basic customer accounts
- 📝 Read-only operations

---

## Troubleshooting

### Common Issues

**1. "Invalid verification code"**
- Check device time is synchronized
- Verify clock skew tolerance (±30 seconds)
- Ensure code hasn't been used already

**2. "Token expired"**
- Setup sessions expire after 15 minutes
- Restart setup process from beginning

**3. "Maximum reset attempts exceeded"**
- Wait for rate limit window to reset
- Check attempt logs for security issues

**4. Lost authenticator device**
- Use recovery codes
- Contact support if all recovery codes used
- Admin can disable 2FA for the user

### Admin Recovery

If a user loses both authenticator and recovery codes:

```sql
-- Disable 2FA for user (admin operation)
DELETE FROM two_factor_auths WHERE user_id = <USER_ID>;
DELETE FROM recovery_codes WHERE two_factor_auth_id IN 
  (SELECT id FROM two_factor_auths WHERE user_id = <USER_ID>);
```

---

## Future Enhancements

### Potential Improvements

1. **SMS 2FA:** Add SMS-based verification as alternative
2. **Trusted Devices:** Remember devices for X days
3. **Backup Methods:** Email-based backup codes
4. **Risk-Based Auth:** Require 2FA only for suspicious logins
5. **Passkeys/WebAuthn:** Modern passwordless authentication
6. **Push Notifications:** Approve login via app notification
7. **Admin Dashboard:** View 2FA adoption rates
8. **Force 2FA:** Admin-enforced 2FA for specific roles

---

## Support & Resources

### TOTP Standards
- [RFC 6238 - TOTP](https://tools.ietf.org/html/rfc6238)
- [RFC 4226 - HOTP](https://tools.ietf.org/html/rfc4226)

### Compatible Apps
- Google Authenticator
- Authy
- Microsoft Authenticator
- 1Password
- Bitwarden
- LastPass Authenticator

### Key Files
- Implementation: `/internal/auth/totp.go`
- Handlers: `/internal/auth/twofa_handler.go`
- Models: `/internal/models/twofa.go`
- Middleware: `/internal/middleware/twofa.go`
- Routes: `/cmd/api-server/main.go`

---

## Quick Reference

```
POST   /2fa/setup              - Initiate 2FA setup
POST   /2fa/verify-setup       - Complete 2FA setup
POST   /2fa/verify-login       - Verify 2FA during login
POST   /2fa/disable            - Disable 2FA
GET    /2fa/status             - Check 2FA status
POST   /2fa/recovery-codes     - Regenerate recovery codes
GET    /2fa/attempts           - View recent attempts
```

**Rate Limits:**
- Auth endpoints: 10 req/min
- Login endpoints: 5 req/min

**Key Security Points:**
- Password required for setup/disable
- TOTP code required for sensitive operations
- Recovery codes are single-use
- All attempts logged
- 15-minute temporary tokens during setup
- ±30 second clock skew tolerance

---

**Implementation Status:** ✅ **Complete and Production-Ready**

Last Updated: November 30, 2025
