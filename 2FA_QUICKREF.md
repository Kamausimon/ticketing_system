# 2FA Quick Reference

## Setup Flow (First Time)

```
1. POST /2fa/setup + password → Get QR code + recovery codes
2. Scan QR with authenticator app
3. POST /2fa/verify-setup + TOTP code → 2FA enabled
```

## Login Flow (With 2FA)

```
1. POST /login + credentials → temp_token if 2FA enabled
2. POST /2fa/verify-login + temp_token + TOTP → full access token
```

## API Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| POST | `/2fa/setup` | Start 2FA setup | ✅ JWT |
| POST | `/2fa/verify-setup` | Complete setup | ✅ JWT |
| POST | `/2fa/verify-login` | Verify during login | ✅ Temp Token |
| POST | `/2fa/disable` | Disable 2FA | ✅ JWT |
| GET | `/2fa/status` | Check status | ✅ JWT |
| POST | `/2fa/recovery-codes` | Regenerate codes | ✅ JWT |
| GET | `/2fa/attempts` | View attempts | ✅ JWT |

## Middleware

```go
// Require 2FA for route
middleware.Require2FA(DB)

// Require organizer + 2FA
middleware.RequireOrganizerWith2FA(DB)

// Recommend 2FA (non-blocking)
middleware.Recommend2FA(DB)
```

## Recovery

**Lost Authenticator:** Use recovery code during login
**Lost Both:** Admin must disable 2FA in database

## Security Specs

- **Algorithm:** TOTP (RFC 6238)
- **Code Length:** 6 digits
- **Time Step:** 30 seconds
- **Clock Skew:** ±30 seconds
- **Secret:** 160 bits (Base32)
- **Recovery Codes:** 10 codes, single-use, bcrypt hashed

## Rate Limits

- Setup: 10 req/min per IP
- Login verification: 5 req/min per IP
- General: 10 req/min per IP

## Code Examples

### Enable 2FA (cURL)
```bash
curl -X POST http://localhost:8080/2fa/setup \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"password":"user_pass"}'
```

### Login with 2FA (JavaScript)
```javascript
// Step 1
const login = await fetch('/login', {
  method: 'POST',
  body: JSON.stringify({email, password})
});
const {requires_2fa, temp_token} = await login.json();

// Step 2 (if requires_2fa)
const verify = await fetch('/2fa/verify-login', {
  method: 'POST',
  headers: {'Authorization': `Bearer ${temp_token}`},
  body: JSON.stringify({code: '123456'})
});
const {token} = await verify.json();
```

## Recommended Use Cases

| Account Type | Recommendation |
|--------------|----------------|
| Organizers | ✅ **Mandatory** |
| Admins | ✅ **Mandatory** |
| Customers (with purchases) | ⚠️ **Recommended** |
| Basic customers | 📝 Optional |

## Files

- `/internal/auth/totp.go` - TOTP implementation
- `/internal/auth/twofa_handler.go` - HTTP handlers
- `/internal/models/twofa.go` - Database models
- `/internal/middleware/twofa.go` - Route protection
- `/cmd/api-server/main.go` - Routes registration

## Compatible Apps

✅ Google Authenticator
✅ Authy
✅ Microsoft Authenticator
✅ 1Password
✅ Bitwarden
✅ Any TOTP-compliant app

---

**Status:** ✅ Production Ready | **Last Updated:** Nov 30, 2025
