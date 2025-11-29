# Rate Limiting Implementation - Protected Routes

This document outlines the rate limiting configuration applied to critical API routes in the ticketing system to prevent abuse and ensure system stability.

## Overview

Rate limiting has been integrated into the API server using the `pkg/ratelimit` package with a token bucket algorithm. Different endpoints have different rate limits based on their criticality and resource consumption.

## Rate Limit Tiers

### 1. **Login Tier** (Strictest - 5 attempts/min per IP)
**Purpose**: Prevent brute force attacks on login endpoints

**Protected Routes**:
- `POST /login` - User login

**Configuration**:
- Requests per second: ~0.083 (5 per minute)
- Burst size: 5
- Used by: IP address

---

### 2. **Auth Tier** (10 requests/min per IP)
**Purpose**: Protect authentication endpoints from abuse

**Protected Routes**:
- `POST /register` - User registration
- `POST /logout` - User logout
- `POST /forgot-password` - Password recovery
- `POST /resetPassword` - Password reset

**Configuration**:
- Requests per second: ~0.167 (10 per minute)
- Burst size: 10
- Used by: IP address

---

### 3. **Payment Tier** (5 requests/min per IP)
**Purpose**: Prevent duplicate payment submissions and accidental charges

**Protected Routes**:
- `POST /payments/initiate` - Initiate payment
- `GET /payments/verify/{id}` - Verify payment
- `POST /payments/methods` - Save payment method
- `DELETE /payments/methods/{id}` - Delete payment method
- `POST /payments/methods/{id}/default` - Set default payment method
- `PUT /payments/methods/{id}/expiry` - Update payment method expiry
- `POST /payments/refunds` - Initiate refund
- `POST /payments/refunds/{id}/approve` - Approve refund
- `PUT /orders/{id}/status` - Update order status
- `POST /orders/{id}/cancel` - Cancel order
- `POST /orders/{id}/refund` - Refund order
- `POST /orders/{id}/payment` - Process payment
- `POST /orders/{id}/payment/verify` - Verify order payment
- `POST /tickets/{id}/transfer` - Transfer ticket
- `POST /refunds` - Request refund
- `POST /refunds/{id}/cancel` - Cancel refund request

**Configuration**:
- Requests per second: ~0.083 (5 per minute)
- Burst size: 5
- Used by: IP address

---

### 4. **API Tier** (100 requests/sec per IP, burst 200)
**Purpose**: Standard rate limiting for regular API operations

**Protected Routes**:
- `POST /orders` - Create order
- `POST /orders/calculate` - Calculate order
- `GET /orders` - List orders
- `GET /orders/{id}` - Get order details
- `GET /orders/{id}/summary` - Get order summary
- `GET /orders/stats` - Get order statistics
- `GET /orders/{id}/payment/status` - Get payment status
- `GET /payments/history` - Get payment history
- `GET /payments/methods` - Get payment methods
- `GET /payments/refunds/{id}/status` - Get refund status
- `GET /payments/refunds` - List refunds
- `GET /tickets/{id}/transfer-history` - Get ticket transfer history
- `GET /inventory/reservations/{id}` - Get reservation
- `GET /inventory/reservations` - List user reservations
- `GET /inventory/events/{id}/reservations` - Get event reservations

**Configuration**:
- Requests per second: 100
- Burst size: 200
- Used by: IP address

---

### 5. **Download Tier** (3 requests/sec per IP)
**Purpose**: Control bandwidth usage for file downloads

**Protected Routes**:
- `GET /tickets/{id}/pdf` - Download ticket PDF

**Configuration**:
- Requests per second: 3
- Burst size: 5
- Used by: IP address

---

### 6. **Inventory Tier** (50 requests/sec, burst 100)
**Purpose**: Prevent inventory abuse and overselling

**Protected Routes**:
- `GET /inventory/tickets/{id}` - Get ticket availability
- `GET /inventory/events/{id}` - Get event inventory
- `GET /inventory/status/{id}` - Get inventory status
- `POST /inventory/bulk-check` - Bulk check availability
- `POST /inventory/reservations` - Create reservation
- `GET /inventory/reservations/{id}/validate` - Validate reservation
- `POST /inventory/reservations/{id}/extend` - Extend reservation
- `DELETE /inventory/reservations/{id}/release` - Release reservation
- `POST /inventory/reservations/expired` - Release expired reservations
- `POST /inventory/reservations/convert` - Convert reservation to order
- `DELETE /inventory/reservations/session` - Release session reservations

**Configuration**:
- Requests per second: 50
- Burst size: 100
- Cleanup interval: 5 hours
- Used by: IP address

---

## Unprotected Routes

The following routes are NOT rate limited to ensure system accessibility:

- `GET /metrics` - Prometheus metrics endpoint
- `GET /events` - List all events (public)
- `GET /events/{id}` - Get event details (public)
- `GET /events/{id}/images` - Get event images (public)
- `GET /account/countries` - Get supported countries
- `GET /account/timezones` - Get available timezones
- `GET /account/currencies` - Get available currencies
- `GET /account/date-formats` - Get date formats
- All other GET requests for read-only data

---

## HTTP Response Headers

When a rate limit is enforced, the API returns:

```
HTTP/1.1 429 Too Many Requests

X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1234567890
Retry-After: 45
```

**Header meanings**:
- `X-RateLimit-Remaining`: Number of requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets
- `Retry-After`: Number of seconds to wait before retrying

---

## Error Response

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded",
  "status": 429,
  "retry_after": 45
}
```

---

## Usage Statistics

The rate limiter automatically:
- Tracks requests per client IP address
- Refills tokens based on the configured rate
- Allows bursts up to the configured burst size
- Cleans up stale entries periodically
- Is fully thread-safe for concurrent requests

---

## Configuration Reference

| Tier | Requests/Sec | Burst | Window | Cleanup | Use Case |
|------|--------------|-------|--------|---------|----------|
| Login | 0.083 | 5 | N/A | 5 min | Brute force protection |
| Auth | 0.167 | 10 | N/A | 5 min | Authentication operations |
| Payment | 0.083 | 5 | N/A | 5 min | Financial operations |
| API | 100 | 200 | N/A | 5 min | Standard API calls |
| Download | 3 | 5 | N/A | 5 min | File downloads |
| Inventory | 50 | 100 | N/A | 5 hours | Inventory operations |

---

## Integration Details

### How It Works

1. **Request arrives** → IP address is extracted
2. **Rate limit check** → Token bucket checked for the IP
3. **Token available?**
   - YES → Token consumed, request proceeds
   - NO → 429 Too Many Requests returned
4. **Token refill** → Tokens are automatically refilled at configured rate

### Middleware Stack

```
HTTP Request
    ↓
Rate Limit Middleware (per-route)
    ↓
Prometheus Middleware
    ↓
Handler Function
```

### Extracting Client IP

The rate limiter extracts client IP in this order:
1. `X-Forwarded-For` header (first IP if multiple)
2. `X-Real-IP` header
3. Request `RemoteAddr`

This ensures correct IP identification even behind proxies.

---

## Testing Rate Limits

### Using curl

```bash
# Test login rate limit (5 attempts per minute)
for i in {1..6}; do
  curl -X POST http://localhost:8080/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"test"}' \
    -w "\nStatus: %{http_code}\n\n"
done
```

### Using load testing tool (ab - ApacheBench)

```bash
# Generate 100 requests
ab -n 100 -c 10 http://localhost:8080/orders
```

### Using k6 (load testing)

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '10s', target: 50 },
  ],
};

export default function() {
  let response = http.get('http://localhost:8080/orders');
  check(response, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });
}
```

---

## Monitoring Rate Limit Violations

### Logs to Monitor

```
Rate limit exceeded for IP: 192.168.1.1
Endpoint: POST /login
Requests: 6/5 allowed
```

### Metrics to Track

If extended with metrics middleware:
- `ratelimit_exceeded_total` - Total rate limit violations
- `ratelimit_remaining_avg` - Average remaining requests
- `ratelimit_response_time_ms` - Rate limit check latency

---

## Future Enhancements

1. **User-based rate limiting** - Track by user ID instead of IP
2. **Endpoint-specific tuning** - Adjust rates per endpoint
3. **Tiered access** - Different limits for premium users
4. **Geographic rate limiting** - Different limits by region
5. **Circuit breaker** - Temporarily block abusive IPs
6. **Analytics dashboard** - Visual rate limit monitoring

---

## Questions & Troubleshooting

### Q: My legitimate app is getting rate limited
**A**: Contact support to whitelist your IP or implement exponential backoff

### Q: Can I increase the rate limits?
**A**: Update configuration in `cmd/api-server/main.go` and redeploy

### Q: How do I disable rate limiting?
**A**: Remove the middleware wrappers (not recommended for production)

### Q: What if I'm behind a proxy?
**A**: Ensure `X-Forwarded-For` or `X-Real-IP` headers are set correctly

---

## Security Best Practices

1. ✅ Always rate limit authentication endpoints
2. ✅ Always rate limit payment endpoints
3. ✅ Monitor for patterns of abuse
4. ✅ Adjust limits based on traffic patterns
5. ✅ Keep rate limit logic in core server
6. ✅ Log all rate limit violations
7. ✅ Use stricter limits during high-traffic periods

---

## Deployment Checklist

- [ ] Rate limiting package tested
- [ ] All critical routes protected
- [ ] Rate limit configuration reviewed
- [ ] Monitoring alerts configured
- [ ] Documentation updated
- [ ] Team trained on rate limits
- [ ] Deployment completed
- [ ] Monitoring active

