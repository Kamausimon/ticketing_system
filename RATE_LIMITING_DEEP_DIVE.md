# 🧠 Rate Limiting Deep Dive - Understanding Your Implementation

## **The Core Problem**

Imagine a coffee shop where everyone can order unlimited drinks simultaneously. Chaos, right? Your API is the same - without rate limiting:
- Attackers can brute-force login attempts
- Someone could spam order creation, crashing your DB
- One client could monopolize resources, starving others
- Payment endpoints could be hammered, causing duplicate charges

**Rate limiting = Traffic control for your API**

---

## **How Your System Works: The Token Bucket Analogy**

Think of a bucket that:
1. **Holds tokens** (each token = 1 request you can make)
2. **Refills automatically** at a steady rate
3. **Has a maximum capacity** (burst limit)

### Example: Login Endpoint
```
Bucket capacity: 5 tokens
Refill rate: 5 tokens per minute (0.083/second)

User makes 5 login attempts rapidly → Uses all 5 tokens
Try a 6th attempt immediately → DENIED (no tokens left)
Wait 12 seconds → 1 token refills → Can try again
```

**Visual Representation:**
```
Time 0s:  [🪙🪙🪙🪙🪙] 5 tokens available
Request:  [🪙🪙🪙🪙_] 4 tokens (1 used)
Request:  [🪙🪙🪙__] 3 tokens (1 used)
Request:  [🪙🪙___] 2 tokens (1 used)
Request:  [🪙____] 1 token (1 used)
Request:  [_____] 0 tokens (1 used)
Request:  ❌ DENIED - No tokens!

After 12s: [🪙____] 1 token refilled → Can try again
```

---

## **Your Implementation - Key Components**

### **1. The Core Algorithm** ([pkg/ratelimit/token_bucket.go](pkg/ratelimit/token_bucket.go))

```go
// When a request comes in:
func (tb *TokenBucket) Allow(key string) bool {
    // Step 1: Calculate how many tokens to refill
    elapsed := now.Sub(state.lastRefil).Seconds()
    tokensToAdd := elapsed * tb.tokensPerSecond
    
    // Step 2: Add tokens (but don't exceed max)
    state.tokens = min(state.tokens + tokensToAdd, tb.maxTokens)
    
    // Step 3: Check if we have at least 1 token
    if state.tokens >= 1 {
        state.tokens--  // Take a token
        return true     // Allow request
    }
    return false        // Deny request
}
```

**What's happening:**
- Every call checks: "Has enough time passed to add more tokens?"
- If yes, refill the bucket (up to max)
- If a token is available, take it and allow the request
- If bucket is empty, deny the request

**Math Example:**
```go
// Login endpoint: 5 requests per minute
RequestsPerSecond = 5/60 = 0.083
BurstSize = 5

// User last requested 30 seconds ago with 2 tokens remaining
elapsed = 30 seconds
tokensToAdd = 30 * 0.083 = 2.5 tokens
newTotal = 2 + 2.5 = 4.5 tokens (still under max of 5)

// Current request needs 1 token
if 4.5 >= 1:  ✓ Allow
    tokens = 4.5 - 1 = 3.5 remaining
```

---

### **2. The Middleware** ([pkg/ratelimit/middleware.go](pkg/ratelimit/middleware.go))

This wraps your HTTP handlers:

```go
func (m *Middleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := m.keyFunc(r)  // Extract IP address
        
        if !m.limiter.Allow(key) {  // Check the bucket
            http.Error(w, "Rate limit exceeded", 429)
            return
        }
        
        next.ServeHTTP(w, r)  // Allow request through
    })
}
```

**Request Flow:**
```
Client Request → Middleware → Extract Key (IP) → Check Token Bucket
                                                         ↓
                                            Yes: tokens available?
                                                         ↓
                                            ┌────────────┴────────────┐
                                           YES                       NO
                                            ↓                         ↓
                                    Pass to Handler           Return 429 Error
                                            ↓
                                    Process Request
                                            ↓
                                    Return Response
```

**Key Extraction Logic:**
```go
func defaultKeyFunc(r *http.Request) string {
    // 1. Check if behind proxy: X-Forwarded-For
    if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
        ips := strings.Split(xForwardedFor, ",")
        return strings.TrimSpace(ips[0])  // "192.168.1.1"
    }
    
    // 2. Check alternative: X-Real-IP  
    if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
        return xRealIP
    }
    
    // 3. Fallback: RemoteAddr
    if r.RemoteAddr != "" {
        ip := strings.Split(r.RemoteAddr, ":")[0]
        return ip
    }
    
    return "unknown"
}
```

**Why this matters:** 
If you're behind a load balancer or reverse proxy (like Nginx), you need to extract the *real* client IP, not the proxy's IP. Otherwise, all requests would be rate-limited as if they came from a single source!

**Scenario:**
```
Without X-Forwarded-For:
User A (192.168.1.1) → Load Balancer (10.0.0.5) → Your API
                                        ↑
                          API sees: 10.0.0.5 (wrong!)

With X-Forwarded-For:
User A (192.168.1.1) → Load Balancer → Your API
                       Header: X-Forwarded-For: 192.168.1.1
                                        ↑
                          API sees: 192.168.1.1 (correct!)
```

---

### **3. The Governor Pattern** ([pkg/ratelimit/governor.go](pkg/ratelimit/governor.go))

The Governor manages multiple rate limiters for different endpoints:

```go
func InitializeRateLimiting() *ratelimit.Governor {
    gov := ratelimit.NewTokenBucketGovernor()

    gov.GetOrCreate("api", ratelimit.Presets.API)         // 100 req/s, burst 200
    gov.GetOrCreate("auth", ratelimit.Presets.Auth)       // 10 req/min
    gov.GetOrCreate("login", ratelimit.Presets.Login)     // 5 attempts/min
    gov.GetOrCreate("payment", ratelimit.Presets.Payment) // 5 req/min
    gov.GetOrCreate("download", ratelimit.Presets.Download) // 3 req/s

    return gov
}
```

**Why use a Governor?**
Without it, you'd need separate limiters for each endpoint type, making management messy. The Governor acts as a centralized registry.

**Conceptual View:**
```
Governor
├── "api" limiter → TokenBucket(100 req/s, burst 200)
├── "auth" limiter → TokenBucket(10 req/min, burst 10)
├── "login" limiter → TokenBucket(5 req/min, burst 5)
├── "payment" limiter → TokenBucket(5 req/min, burst 5)
└── "download" limiter → TokenBucket(3 req/s, burst 6)

Each limiter tracks tokens per IP:
"api" limiter:
  ├── "192.168.1.1" → 150 tokens
  ├── "192.168.1.2" → 180 tokens
  └── "10.0.0.5" → 200 tokens

"login" limiter:
  ├── "192.168.1.1" → 3 tokens
  └── "192.168.1.2" → 5 tokens
```

---

### **4. Rate Limit Tiers (Your Strategy)**

You've created **5 different tiers** for different security needs:

| Tier | Rate | Burst | Why? | Example Endpoints |
|------|------|-------|------|-------------------|
| **Login** | 5/min | 5 | Prevent brute force attacks | `POST /login` |
| **Auth** | 10/min | 10 | Protect registration/password reset | `POST /register`, `POST /forgot-password` |
| **Payment** | 5/min | 5 | Prevent duplicate charges/errors | `POST /payments/initiate`, `POST /refunds` |
| **API** | 100/sec | 200 | Normal operations (browsing, reading) | `GET /orders`, `GET /events` |
| **Download** | 3/sec | 6 | Control bandwidth for PDFs | `GET /tickets/{id}/pdf` |

**The thinking behind each tier:**

1. **Login (Strictest)** - 5 attempts per minute
   - **Threat:** Brute force password attacks
   - **Logic:** Legitimate users rarely fail login 5+ times in a minute
   - **Protection:** Attacker would need 20 minutes to try 100 passwords

2. **Auth** - 10 per minute
   - **Threat:** Account enumeration, spam registrations
   - **Logic:** Registration is a one-time action
   - **Protection:** Limits bot account creation

3. **Payment** - 5 per minute
   - **Threat:** Duplicate charges, payment system abuse
   - **Logic:** Real users don't initiate 5+ payments per minute
   - **Protection:** Prevents accidental double-clicks causing duplicate charges

4. **API** - 100 per second
   - **Threat:** Resource exhaustion, scraping
   - **Logic:** Normal browsing involves many requests
   - **Protection:** Allows legitimate use while stopping aggressive bots
   - **Math:** 100 req/s = 6,000 req/min = 360,000 req/hour per IP

5. **Download** - 3 per second
   - **Threat:** Bandwidth abuse
   - **Logic:** PDF generation is CPU-intensive
   - **Protection:** Prevents one user from monopolizing server resources
   - **Math:** 3 req/s = 180 downloads/min = reasonable for legitimate use

---

### **5. Response Headers** ([cmd/api-server/ratelimit_integration.go](cmd/api-server/ratelimit_integration.go))

Your implementation adds helpful headers:

```go
// Add rate limit headers
w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(result.ResetAfter).Unix()))

if !result.Allowed {
    // Add Retry-After header
    w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
    http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

**Example Response:**
```http
HTTP/1.1 200 OK
X-RateLimit-Remaining: 147
X-RateLimit-Reset: 1735574400
Content-Type: application/json

{"status": "success"}
```

**When rate limited:**
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1735574412
Retry-After: 12

Rate limit exceeded
```

**Benefits:**
- **Transparency:** Clients know how many requests they have left
- **Graceful degradation:** Clients can back off intelligently
- **Better UX:** Mobile apps can show "Try again in 12 seconds"

---

## **The Thread-Safety Magic**

```go
type TokenBucket struct {
    mu sync.RWMutex  // This is critical!
    tokens map[string]*bucketState
}

func (tb *TokenBucket) Allow(key string) bool {
    tb.mu.Lock()         // Lock before accessing shared data
    defer tb.mu.Unlock() // Unlock when done
    // ... rest of logic
}
```

**Why this matters:**

**Without Mutex (Race Condition):**
```
Time    Goroutine 1                 Goroutine 2
────────────────────────────────────────────────────
0ms     Read tokens: 5              
1ms                                 Read tokens: 5
2ms     Subtract 1 → 4              
3ms                                 Subtract 1 → 4
4ms     Write back: 4               
5ms                                 Write back: 4

Result: Both requests allowed, but tokens = 4 (should be 3!)
```

**With Mutex (Correct):**
```
Time    Goroutine 1                 Goroutine 2
────────────────────────────────────────────────────
0ms     Lock acquired               
1ms     Read tokens: 5              Waiting...
2ms     Subtract 1 → 4              Waiting...
3ms     Write back: 4               Waiting...
4ms     Unlock                      Waiting...
5ms                                 Lock acquired
6ms                                 Read tokens: 4
7ms                                 Subtract 1 → 3
8ms                                 Write back: 3
9ms                                 Unlock

Result: Both requests allowed, tokens = 3 ✓
```

**Real-world scenario:**
- 1000 users hit `/login` simultaneously
- They all try to check/modify the same token bucket for their IP
- Without locks: Corrupted data, wrong token counts, security breach
- With `sync.RWMutex`: Only one goroutine modifies at a time = accurate counting

**RWMutex vs Mutex:**
```go
// RWMutex allows multiple readers OR one writer
mu.RLock()   // Multiple goroutines can read simultaneously
mu.RUnlock() // Release read lock

mu.Lock()    // Only one goroutine can write
mu.Unlock()  // Release write lock
```

Your implementation uses full `Lock()` because you're both reading and writing (checking and decrementing tokens).

---

## **Memory Management: The Cleanup**

```go
func (tb *TokenBucket) cleanup() {
    ticker := time.NewTicker(tb.cleanupInterval)  // Every 5 minutes
    for range ticker.C {
        tb.mu.Lock()
        now := time.Now()
        for key, state := range tb.tokens {
            // Delete entries not accessed recently
            if now.Sub(state.lastRefil) > tb.cleanupInterval {
                delete(tb.tokens, key)
            }
        }
        tb.mu.Unlock()
    }
}
```

**Why this matters:**

**Without Cleanup:**
```
Hour 0: 100 unique IPs visit → 100 entries in map
Hour 1: 200 more IPs visit → 300 entries in map
Hour 2: 150 more IPs visit → 450 entries in map
Hour 24: 5000 entries in map (most are abandoned)
Day 30: 150,000 entries → Gigabytes of memory!
```

**With Cleanup:**
```
Every 5 minutes:
- Check each entry's last access time
- If not accessed in 5+ minutes → Delete
- Keeps only active users in memory

Result: Memory stays constant (only active users tracked)
```

**Trade-offs:**
- **Cleanup interval too short:** Waste CPU checking frequently
- **Cleanup interval too long:** Waste memory on abandoned entries
- **Your choice (5 minutes):** Good balance for typical web traffic

---

## **Real-World Example: Tracing a Login Attempt**

Let's trace a complete request through your system:

### **Scenario:** User tries to log in

**Request:**
```http
POST /auth/login HTTP/1.1
Host: api.ticketing.com
X-Forwarded-For: 102.168.1.50
Content-Type: application/json

{"email": "user@example.com", "password": "wrong"}
```

**Step-by-Step Processing:**

1. **Request arrives at router**
   ```go
   // Gorilla mux routes to auth handler
   authRouter.Use(RateLimitingMiddleware(gov, "login"))
   ```

2. **Middleware extracts key**
   ```go
   keyFunc := ratelimit.KeyFuncs.ByIP
   key := keyFunc(r)  // Returns: "102.168.1.50"
   ```

3. **Governor looks up limiter**
   ```go
   result := gov.AllowWithResult("login", "102.168.1.50")
   // Uses the "login" limiter (5 req/min, burst 5)
   ```

4. **Token bucket calculation**
   ```go
   // First visit? Create new bucket
   state := &bucketState{
       tokens:    5,      // Start with full bucket
       lastRefil: now,    // Current time
   }
   
   // OR returning visit? Calculate refill
   elapsed := now.Sub(state.lastRefil).Seconds()  // 10 seconds
   tokensToAdd := 10 * 0.083  // = 0.83 tokens
   state.tokens = min(3 + 0.83, 5)  // = 3.83 tokens
   ```

5. **Decision**
   ```go
   if state.tokens >= 1 {  // 3.83 >= 1? YES
       state.tokens--      // 3.83 - 1 = 2.83 tokens
       return true         // Allow request
   }
   ```

6. **Add response headers**
   ```go
   w.Header().Set("X-RateLimit-Remaining", "2")     // Floor of 2.83
   w.Header().Set("X-RateLimit-Reset", "1735574412")
   ```

7. **Pass to login handler**
   ```go
   next.ServeHTTP(w, r)  // Actual login logic executes
   ```

**Response:**
```http
HTTP/1.1 401 Unauthorized
X-RateLimit-Remaining: 2
X-RateLimit-Reset: 1735574412
Content-Type: application/json

{"error": "Invalid credentials"}
```

### **User tries 3 more times (all wrong password):**

```
Attempt 1: 2.83 tokens → Allow → 1.83 remaining
Attempt 2: 1.83 tokens → Allow → 0.83 remaining
Attempt 3: 0.83 tokens → DENY → 0 remaining
```

**Final Response:**
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1735574460
Retry-After: 12

Rate limit exceeded
```

User must wait 12 seconds before trying again.

---

## **Advanced Patterns in Your Code**

### **1. Custom Key Functions**

You can rate-limit by user instead of IP:

```go
func CustomKeyFunction(r *http.Request) string {
    // Try to get authenticated user ID from JWT or session
    if userID := r.Header.Get("X-User-ID"); userID != "" {
        return "user:" + userID
    }
    
    // Fall back to IP-based rate limiting for unauthenticated requests
    return ratelimit.KeyFuncs.ByIP(r)
}
```

**Use case:**
- Authenticated users: Rate limit per account (prevents sharing IPs like office NAT)
- Anonymous users: Rate limit per IP (no other identifier available)

**Example:**
```
Office with 50 employees (same IP: 192.168.1.1):
- Without user-based: All 50 share same rate limit (bad UX)
- With user-based: Each employee gets their own bucket ✓
```

### **2. Dynamic Rate Limit Adjustment**

```go
func AdjustRateLimits(gov *ratelimit.Governor, limiterName string, newConfig ratelimit.Config) {
    gov.GetOrCreate(limiterName, newConfig)
    fmt.Printf("Updated rate limit for %s: %.0f req/s, burst %d\n",
        limiterName, newConfig.RequestsPerSecond, newConfig.BurstSize)
}
```

**Use cases:**
- Under attack? Tighten limits temporarily
- Low traffic period? Relax limits
- Premium users? Create "premium" limiter with higher limits

### **3. Per-Route Fine-Tuning**

```go
// Very sensitive operation gets its own strict limit
customLimiter := ratelimit.NewTokenBucket(ratelimit.Config{
    RequestsPerSecond: 1,  // Very strict: 1 req/s
    BurstSize:         2,  // Allow burst of 2
})

router.HandleFunc("/admin/delete-all-data", 
    customMiddleware.HandlerFunc(dangerousOperation))
```

---

## **Comparison with Other Algorithms**

### **Token Bucket vs Fixed Window**

**Fixed Window:**
```
Minute 1: [|||||||||| ] 10 requests → All allowed
Minute 2: [          ] Reset → Fresh 10 requests

Problem: Burst at boundaries
- 00:59 → 10 requests
- 01:00 → 10 requests (reset)
= 20 requests in 2 seconds!
```

**Token Bucket (Your Choice):**
```
Any 60-second period: Max 10 requests
Smooth refilling prevents boundary bursts
✓ Better traffic shaping
```

### **Token Bucket vs Leaky Bucket**

**Leaky Bucket:**
- Requests queue up
- Process at constant rate
- Requests wait in queue

**Token Bucket:**
- No queue
- Allow/deny immediately
- Burst capability

**Why you chose Token Bucket:**
- Better for APIs (immediate response)
- Allows legitimate bursts
- Simpler implementation
- No need to manage queues

---

## **Performance Characteristics**

### **Time Complexity**
- **Allow()**: O(1) - Map lookup + simple math
- **Cleanup()**: O(n) - Iterate all keys, but runs infrequently

### **Space Complexity**
- O(n) where n = number of unique IPs/keys
- With cleanup: n = active users only
- Typical: 1000 active IPs = ~50-100 KB memory

### **Concurrency**
- Thread-safe with mutex
- Read-heavy optimization possible with RWMutex (future improvement)
- Current: Safe for production under load

---

## **Testing Your Rate Limiter**

### **Manual Test with curl:**

```bash
# Test login rate limit (5 per minute)
for i in {1..6}; do
    echo "Request $i:"
    curl -X POST http://localhost:8080/auth/login \
         -H "Content-Type: application/json" \
         -d '{"email":"test@example.com","password":"test"}' \
         -i | grep -E "HTTP|X-RateLimit|Retry-After"
    echo ""
done
```

**Expected output:**
```
Request 1: HTTP/1.1 200 OK, X-RateLimit-Remaining: 4
Request 2: HTTP/1.1 200 OK, X-RateLimit-Remaining: 3
Request 3: HTTP/1.1 200 OK, X-RateLimit-Remaining: 2
Request 4: HTTP/1.1 200 OK, X-RateLimit-Remaining: 1
Request 5: HTTP/1.1 200 OK, X-RateLimit-Remaining: 0
Request 6: HTTP/1.1 429 Too Many Requests, Retry-After: 12
```

### **Load Test with Apache Bench:**

```bash
# Simulate 100 concurrent users making 1000 requests
ab -n 1000 -c 100 http://localhost:8080/api/tickets

# Look for 429 responses in output
```

---

## **Common Issues & Solutions**

### **Issue 1: All requests get same rate limit**
**Cause:** Behind load balancer but not reading X-Forwarded-For
**Solution:** Your middleware already handles this! ✓

### **Issue 2: Memory grows over time**
**Cause:** No cleanup of old entries
**Solution:** Your cleanup goroutine handles this! ✓

### **Issue 3: Distributed systems**
**Problem:** Multiple API servers = separate token buckets
**Current:** Each server tracks its own limits (IP hitting server A and server B gets double the requests)
**Future Solution:** Use Redis for shared state (on your roadmap!)

### **Issue 4: Testing in development**
**Problem:** Hitting rate limits during testing
**Solution:** 
```go
if os.Getenv("ENV") == "development" {
    // Use very generous limits in dev
    config.RequestsPerSecond = 1000
}
```

---

## **Future Enhancements (Your Roadmap)**

### **1. Redis-Based Rate Limiting**
```go
// Current: In-memory (per server)
type TokenBucket struct {
    tokens map[string]*bucketState  // Lost on restart
}

// Future: Redis-backed (shared across servers)
func (tb *RedisBucket) Allow(key string) bool {
    // Use Redis INCR with TTL
    // Survives restarts
    // Shared across all API servers
}
```

**Benefits:**
- Distributed rate limiting
- Survives server restarts
- Accurate limits across load balancers

### **2. Rate Limit by User Tier**
```go
func GetLimitForUser(userID string) ratelimit.Config {
    user := db.GetUser(userID)
    switch user.Tier {
    case "premium":
        return ratelimit.Config{RequestsPerSecond: 1000}
    case "basic":
        return ratelimit.Config{RequestsPerSecond: 100}
    default:
        return ratelimit.Presets.API
    }
}
```

### **3. Adaptive Rate Limiting**
```go
// Monitor system load
if cpuUsage > 80% {
    // Automatically tighten rate limits
    governor.AdjustAll(0.5)  // Reduce all limits by 50%
}
```

---

## **Key Takeaways for Your Blog Post**

### **What Makes Your Implementation Production-Ready:**

1. **Token Bucket = Smooth & Flexible**
   - Allows bursts (good UX for legitimate users)
   - Recovers gracefully (refills over time)
   - Better than rigid fixed-window approaches

2. **IP-Based = Simple but Effective**
   - No authentication needed for rate limiting
   - Works for public endpoints
   - Proxy-aware (X-Forwarded-For handling)
   - Downside: Shared IPs (offices, NAT) get same limit

3. **Tiered Approach = Smart Security**
   - Not all endpoints are equal
   - Payment/Auth need stricter control
   - Browse/Read can be generous
   - Matches threat model to protection level

4. **Governor Pattern = Maintainable**
   - Centralized rate limit management
   - Easy to add new limiters
   - Clean separation of concerns

5. **Production Features:**
   - Thread-safe (handles concurrency correctly)
   - Memory-efficient (automatic cleanup)
   - Observable (response headers show remaining quota)
   - Transparent (clients know when to retry)

---

## **Code Snippet for Your Blog**

```go
// Production-ready rate limiting setup
func setupRateLimiting() {
    // Create governor to manage multiple rate limiters
    gov := ratelimit.NewTokenBucketGovernor()
    
    // Define limits for different endpoint types
    gov.GetOrCreate("api", ratelimit.Config{
        RequestsPerSecond: 100,  // Normal API operations
        BurstSize:         200,  // Allow bursts
    })
    
    gov.GetOrCreate("payment", ratelimit.Config{
        RequestsPerSecond: 0.083,  // 5 per minute
        BurstSize:         5,       // Strict limit
    })
    
    // Apply to routes
    router.Use(RateLimitingMiddleware(gov, "api"))
    
    // Payment routes get stricter limits
    paymentRouter := router.PathPrefix("/payments").Subrouter()
    paymentRouter.Use(RateLimitingMiddleware(gov, "payment"))
}

// Middleware with informative headers
func RateLimitingMiddleware(gov *ratelimit.Governor, limiterName string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := extractClientIP(r)
            result := gov.AllowWithResult(limiterName, key)
            
            // Add headers so clients know their quota
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
            w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(result.ResetAfter).Unix()))
            
            if !result.Allowed {
                w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## **Related Files to Study**

- [pkg/ratelimit/token_bucket.go](pkg/ratelimit/token_bucket.go) - Core algorithm
- [pkg/ratelimit/middleware.go](pkg/ratelimit/middleware.go) - HTTP integration
- [pkg/ratelimit/governor.go](pkg/ratelimit/governor.go) - Multi-limiter management
- [cmd/api-server/ratelimit_integration.go](cmd/api-server/ratelimit_integration.go) - Production usage
- [RATELIMIT_IMPLEMENTATION.md](RATELIMIT_IMPLEMENTATION.md) - Route configuration
- [RATELIMIT_QUICKREF.md](RATELIMIT_QUICKREF.md) - Quick reference guide

---

## **Questions to Explore Further**

1. How would you modify this for multi-datacenter deployment?
2. What happens when Redis goes down (if you implement Redis-backed limiting)?
3. How would you handle rate limiting for GraphQL APIs (single endpoint, many operations)?
4. What metrics would you track to tune rate limits over time?
5. How would you implement "circuit breakers" alongside rate limiting?

---

**Next Steps:**
- Test your understanding by explaining token bucket to a friend
- Draw a diagram of the request flow through your middleware
- Experiment with different rate limits and observe behavior
- Consider writing a blog post section on rate limiting using this knowledge
