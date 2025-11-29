# Rate Limiting - Example Requests & Responses

This document shows practical examples of how rate limiting works in the ticketing system API.

## 1. Successful Request (Within Limit)

### Request
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Response (First Request)
```
HTTP/1.1 200 OK
X-RateLimit-Remaining: 4
X-RateLimit-Reset: 1701259800
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

**Explanation**:
- Request succeeded with status 200
- 4 more login attempts allowed in this minute (5 total, used 1)
- Reset time: 1701259800 (Unix timestamp)

---

## 2. Rate Limit Exceeded Response

### Request (6th login attempt within 1 minute)
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "attacker@example.com",
    "password": "guessedpassword"
  }'
```

### Response
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1701259800
Retry-After: 45
Content-Type: text/plain

Rate limit exceeded
```

**Explanation**:
- Request blocked with status 429
- No more requests allowed (0 remaining)
- Must wait 45 seconds before retrying
- Reset happens at timestamp 1701259800

---

## 3. Payment Operation Rate Limiting

### Request 1-5: Successful Payments
```bash
# First payment - succeeds
curl -X POST http://localhost:8080/payments/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token123" \
  -d '{
    "order_id": "order-123",
    "amount": 5000,
    "currency": "KES"
  }'
```

### Response (Request 1)
```
HTTP/1.1 200 OK
X-RateLimit-Remaining: 4
X-RateLimit-Reset: 1701259800
Retry-After: 0

{
  "payment_id": "pay-001",
  "status": "initiated",
  "amount": 5000
}
```

### Request 6: Duplicate Payment (Rate Limited)
```bash
# Sixth payment attempt within 1 minute
curl -X POST http://localhost:8080/payments/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token123" \
  -d '{
    "order_id": "order-456",
    "amount": 3000,
    "currency": "KES"
  }'
```

### Response (Request 6)
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1701259860
Retry-After: 60

Rate limit exceeded
```

---

## 4. Inventory Check Rate Limiting

### Rapid Inventory Checks (Within Limit)

```bash
# Check 1 - succeeds
curl -X POST http://localhost:8080/inventory/bulk-check \
  -H "Content-Type: application/json" \
  -d '{
    "ticket_ids": ["t1", "t2", "t3", "t4", "t5"]
  }'
```

**Response Headers**:
```
X-RateLimit-Remaining: 49
X-RateLimit-Reset: 1701259800
```

### Multiple Rapid Checks

```bash
# Perform 50 inventory checks in quick succession
for i in {1..50}; do
  curl -X POST http://localhost:8080/inventory/bulk-check \
    -H "Content-Type: application/json" \
    -d "{\"ticket_ids\": [\"t$i\"]}" \
    -w "Request $i: %{http_code}\n"
done
```

**Results**:
```
Request 1: 200 (49 remaining)
Request 2: 200 (48 remaining)
...
Request 50: 200 (0 remaining)
Request 51: 429 (Too Many Requests)
```

---

## 5. Ticket PDF Download Rate Limiting

### Legitimate User Downloads

```bash
# User downloads their own ticket PDF
curl -X GET http://localhost:8080/tickets/ticket-123/pdf \
  -H "Authorization: Bearer user-token" \
  -o ticket.pdf
```

**Response**:
```
HTTP/1.1 200 OK
X-RateLimit-Remaining: 4
Retry-After: 0
Content-Type: application/pdf

[PDF Binary Content]
```

### Attempted Abuse (Multiple Downloads)

```bash
# Download attempts in quick succession
for i in {1..4}; do
  curl -X GET http://localhost:8080/tickets/ticket-$i/pdf \
    -H "Authorization: Bearer token" \
    -w "Download $i: %{http_code}\n"
done

# Fifth attempt exceeds limit
curl -X GET http://localhost:8080/tickets/ticket-5/pdf \
  -w "Download 5: %{http_code}\n"
```

**Results**:
```
Download 1: 200
Download 2: 200
Download 3: 200
Download 4: 200
Download 5: 429 (Too Many Requests)
Retry-After: 20
```

---

## 6. Order Creation Rate Limiting

### Single Order Creation

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer user-token" \
  -d '{
    "event_id": "event-123",
    "tickets": [
      {
        "ticket_type_id": "tt-1",
        "quantity": 2
      }
    ],
    "attendees": [
      {
        "first_name": "John",
        "last_name": "Doe",
        "email": "john@example.com"
      }
    ]
  }'
```

**Response**:
```
HTTP/1.1 200 OK
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1701259800

{
  "order_id": "order-789",
  "status": "pending",
  "total_amount": 10000
}
```

---

## 7. Handling Rate Limits - Client-Side Implementation

### Exponential Backoff Strategy

```javascript
// JavaScript example with exponential backoff
async function makeRequest(url, options, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After');
        const waitTime = (retryAfter ? parseInt(retryAfter) : Math.pow(2, attempt)) * 1000;
        
        console.log(`Rate limited. Waiting ${waitTime}ms before retry...`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
        continue;
      }
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      
      return await response.json();
    } catch (error) {
      if (attempt === maxRetries - 1) throw error;
    }
  }
}

// Usage
makeRequest('/api/orders', { method: 'POST', body: JSON.stringify(orderData) })
  .then(result => console.log('Success:', result))
  .catch(error => console.error('Failed:', error));
```

### Python Example with Retry Logic

```python
import requests
import time
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

def create_session_with_retries():
    session = requests.Session()
    
    # Define retry strategy
    retry_strategy = Retry(
        total=3,
        status_forcelist=[429],
        allowed_methods=["POST", "GET"],
        backoff_factor=1
    )
    
    adapter = HTTPAdapter(max_retries=retry_strategy)
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    
    return session

# Usage
session = create_session_with_retries()

response = session.post(
    'http://localhost:8080/orders',
    json={
        'event_id': 'event-123',
        'tickets': [{'ticket_type_id': 'tt-1', 'quantity': 2}]
    },
    headers={'Authorization': 'Bearer token'}
)

print(response.status_code)
print(response.headers.get('X-RateLimit-Remaining'))
```

---

## 8. Rate Limit Information from Response Headers

### Parsing Rate Limit Headers

```javascript
function parseRateLimitHeaders(response) {
  return {
    remaining: parseInt(response.headers.get('X-RateLimit-Remaining') || '0'),
    resetTime: parseInt(response.headers.get('X-RateLimit-Reset') || '0'),
    retryAfter: parseInt(response.headers.get('Retry-After') || '0'),
    isRateLimited: response.status === 429
  };
}

// Usage
async function fetchWithRateLimitInfo(url) {
  const response = await fetch(url);
  const rateLimitInfo = parseRateLimitHeaders(response);
  
  if (rateLimitInfo.isRateLimited) {
    console.log(`Rate limited. Retry after ${rateLimitInfo.retryAfter} seconds`);
  } else {
    console.log(`${rateLimitInfo.remaining} requests remaining`);
  }
  
  return response;
}
```

---

## 9. Monitoring Rate Limit Violations

### Server-Side Logging

```go
// Log when rate limit is exceeded
func logRateLimitViolation(ip string, endpoint string, limiterName string) {
  log.Warnf(
    "Rate limit exceeded: IP=%s, Endpoint=%s, Limiter=%s, Time=%v",
    ip, endpoint, limiterName, time.Now(),
  )
  // Send to monitoring system
  metrics.RecordRateLimitViolation(limiterName)
}
```

### Sample Logs

```
2024-01-15 10:23:45 WARN: Rate limit exceeded: IP=192.168.1.100, Endpoint=/login, Limiter=login
2024-01-15 10:24:12 WARN: Rate limit exceeded: IP=203.45.67.89, Endpoint=/payments/initiate, Limiter=payment
2024-01-15 10:25:01 WARN: Rate limit exceeded: IP=192.168.1.101, Endpoint=/inventory/bulk-check, Limiter=inventory
```

---

## 10. Adjusting Rate Limits for Different Scenarios

### Low Traffic Period Configuration

```go
// Off-peak: More generous limits
gov.GetOrCreate("api", ratelimit.Config{
  RequestsPerSecond: 200,
  BurstSize: 400,
})
```

### High Traffic Period Configuration

```go
// Peak hours: Stricter limits
gov.GetOrCreate("api", ratelimit.Config{
  RequestsPerSecond: 50,
  BurstSize: 100,
})
```

### Premium User Configuration

```go
// Premium users: Higher limits
premiumLimiter := ratelimit.NewTokenBucket(ratelimit.Config{
  RequestsPerSecond: 200,
  BurstSize: 400,
})
```

---

## Common Scenarios & Solutions

### Scenario 1: Bot Attempting Brute Force

```
Requests: POST /login from IP 203.45.67.89
Attempt 1: 200 OK
Attempt 2: 200 OK
Attempt 3: 200 OK
Attempt 4: 200 OK
Attempt 5: 200 OK
Attempt 6: 429 Too Many Requests
Attempt 7: 429 Too Many Requests
```

**Solution**: Rate limiter blocks after 5 attempts, IP must wait 1 minute to retry

---

### Scenario 2: User Rapid-Clicking Checkout

```
Click 1: POST /orders → 200 OK
Click 2: POST /orders → 200 OK
Click 3: POST /orders → 200 OK
Click 4: POST /orders → 200 OK
Click 5: POST /orders → 200 OK
Click 6: POST /orders → 429 Too Many Requests (Retry-After: 60)
```

**Solution**: Client receives Retry-After header, shows "Please wait" message

---

### Scenario 3: Aggressive Scraping

```
Request 1: GET /inventory/bulk-check → 200 OK (49 remaining)
Request 2: GET /inventory/bulk-check → 200 OK (48 remaining)
...
Request 50: GET /inventory/bulk-check → 200 OK (0 remaining)
Request 51: GET /inventory/bulk-check → 429 Too Many Requests
```

**Solution**: Scraper blocked from making more requests for ~5 seconds

---

## Best Practices for API Consumers

1. **Implement Retry Logic**: Use exponential backoff when you get 429
2. **Cache Results**: Avoid repeated requests for the same data
3. **Respect Rate Limit Headers**: Check `Retry-After` before retrying
4. **Batch Operations**: Use bulk endpoints when available
5. **Stagger Requests**: Spread requests over time rather than all at once
6. **Monitor Headers**: Log `X-RateLimit-Remaining` to track usage
7. **Handle Gracefully**: Show users meaningful messages when rate limited

