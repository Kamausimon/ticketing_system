# Rate Limiting - Quick Reference Guide

## ⚡ Quick Start

Rate limiting is now integrated into your ticketing system API. Critical routes are automatically protected from abuse.

---

## 🚦 Rate Limits at a Glance

| Endpoint Type | Limit | Burst | Protection |
|---|---|---|---|
| Login | 5/min | 1 | Brute force |
| Auth (register, logout, etc) | 10/min | 10 | Account takeover |
| Payments | 5/min | 5 | Duplicate charges |
| Orders | 100/sec | 200 | Abuse |
| Downloads (PDFs) | 3/sec | 5 | Bandwidth |
| Inventory | 50/sec | 100 | Overselling |

---

## 🔴 What Happens When Rate Limited?

**Response Status**: `429 Too Many Requests`

**Response Headers**:
```
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1701259800
Retry-After: 45
```

**Body**: `Rate limit exceeded`

---

## ✅ Protected Routes (Partial List)

### Authentication
- `POST /login` ← **STRICTEST** (5 attempts/min)
- `POST /register`
- `POST /logout`
- `POST /forgot-password`
- `POST /resetPassword`

### Payments
- `POST /payments/initiate` ← **STRICT** (5 requests/min)
- `POST /payments/methods`
- `DELETE /payments/methods/{id}`
- `POST /payments/refunds`
- `POST /orders/{id}/payment`

### Inventory
- `POST /inventory/bulk-check` ← **MEDIUM** (50 req/sec)
- `POST /inventory/reservations`
- `GET /inventory/tickets/{id}`

### Downloads
- `GET /tickets/{id}/pdf` ← **SLOW** (3 req/sec)

---

## 💻 Client Implementation

### JavaScript - Handling 429

```javascript
async function apiCall(url, options) {
  const response = await fetch(url, options);
  
  if (response.status === 429) {
    const retryAfter = response.headers.get('Retry-After');
    console.warn(`Rate limited. Wait ${retryAfter}s`);
    // Wait before retrying
    await new Promise(r => setTimeout(r, retryAfter * 1000));
    return apiCall(url, options); // Retry
  }
  
  return response;
}
```

### Python - Automatic Retry

```python
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

session = requests.Session()
retry = Retry(total=3, status_forcelist=[429], backoff_factor=1)
adapter = HTTPAdapter(max_retries=retry)
session.mount('http://', adapter)
session.mount('https://', adapter)

response = session.post('http://api.example.com/orders', json={...})
```

### cURL - Check Headers

```bash
curl -i -X POST http://localhost:8080/login \
  -d '{"email":"user@example.com","password":"pass"}'
  
# Look for:
# X-RateLimit-Remaining: 4
# X-RateLimit-Reset: 1701259800
```

---

## 🔧 Adjusting Limits

**Location**: `cmd/api-server/main.go`

**Example**: Increase payment limit from 5 to 10 requests/min

```go
gov.GetOrCreate("payment", ratelimit.Config{
  RequestsPerSecond: 10.0 / 60,  // Changed from 5
  BurstSize:         10,
  CleanupInterval:   5 * time.Minute,
})
```

---

## 📊 Understanding Headers

### X-RateLimit-Remaining
**Meaning**: How many requests you have left this minute/second

```
X-RateLimit-Remaining: 3
```
= 3 more requests allowed before hitting limit

### X-RateLimit-Reset
**Meaning**: Unix timestamp when limit resets

```
X-RateLimit-Reset: 1701259800
```
= Limit resets at Jan 15, 2024 10:30:00 UTC

### Retry-After
**Meaning**: Seconds to wait before retrying (on 429 only)

```
Retry-After: 45
```
= Wait 45 seconds, then retry

---

## ⚠️ Common Mistakes

### ❌ Not Handling 429
```javascript
// BAD - Will fail repeatedly
fetch('/api/orders').then(r => r.json());
```

### ✅ Proper Error Handling
```javascript
// GOOD - Respects rate limits
fetch('/api/orders')
  .then(r => {
    if (r.status === 429) {
      const wait = r.headers.get('Retry-After');
      throw new Error(`Wait ${wait}s`);
    }
    return r.json();
  });
```

### ❌ Hammering Endpoint
```javascript
// BAD - 100 requests instantly
for (let i = 0; i < 100; i++) {
  fetch('/api/inventory/bulk-check', {...});
}
```

### ✅ Properly Spaced
```javascript
// GOOD - Spaced over time
for (let i = 0; i < 100; i++) {
  await delay(100); // 100ms between requests
  fetch('/api/inventory/bulk-check', {...});
}
```

---

## 📈 Monitoring

### Check Rate Limit Status

```bash
# See current remaining requests
curl -I http://localhost:8080/orders | grep X-RateLimit
```

### Log Rate Limit Violations

Violations are logged with:
- IP address
- Endpoint
- Timestamp
- Limiter tier used

---

## 🆘 Troubleshooting

### Q: Getting 429 errors immediately?
**A**: 
- Check if you're behind a proxy (set `X-Forwarded-For` header)
- Check your IP isn't already rate limited from previous abuse
- Wait for rate limit window to reset

### Q: Need higher limits?
**A**: Contact your admin to adjust configuration in `main.go`

### Q: Can I whitelist my IP?
**A**: Not currently, but can be added. Contact maintainers.

### Q: Retry limit too high?
**A**: Implement exponential backoff to avoid overwhelming the server

---

## 🔐 Security Notes

Rate limiting protects:
1. **Login**: 5 attempts/min - stops brute force attacks
2. **Payments**: 5/min - prevents accidental/malicious double-charging
3. **Inventory**: 50/sec - prevents bot-driven overselling
4. **Downloads**: 3/sec - prevents bandwidth abuse

**These are intentionally strict to prevent abuse.**

---

## 📞 Support

- Package docs: `pkg/ratelimit/README.md`
- Implementation details: `RATELIMIT_IMPLEMENTATION.md`
- Code examples: `RATELIMIT_EXAMPLES.md`
- Integration code: `cmd/api-server/ratelimit_integration.go`

---

## ✅ Checklist for Integration

- [ ] Read this guide
- [ ] Implement 429 error handling in your client
- [ ] Add exponential backoff retry logic
- [ ] Monitor `X-RateLimit-Remaining` header
- [ ] Set `Retry-After` in user-facing messages
- [ ] Test with high concurrency
- [ ] Monitor logs for patterns

---

**Last Updated**: January 2024  
**Status**: Production Ready ✅
