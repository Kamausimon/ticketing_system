# Rate Limiting Implementation using Golang

Imagine a coffee shop where everyone can get unlimited drinks simultaneously? Pure chaos, right? The same logic applies to your API. Without proper mechanisms in place to control requests being made to your API, several problems arise:

- Attackers could brute force your login endpoint
- Someone could spam your order creation routes, crashing your database
- One client could hoard resources and end up starving others
- Worse, your payment endpoint could end up being hammered, causing duplicate payment inconsistencies

## Choosing the Right Algorithm

There are many different ways to implement rate limiting: sliding window log, fixed window counter, leaky bucket, and more. In this article, we'll focus on the most widespread one—the **Token Bucket algorithm**—and how to implement it using Golang.

## Understanding Token Bucket

Think of a bucket that holds tokens, where each token equals one request. The bucket fills back up at a steady rate that you can customize, and it has a maximum capacity which we define as a burst limit.

### A Practical Example

Picture a bucket with 5 tokens set to be used at the login endpoint:

1. The user requests to login → one token is used → 4 tokens remaining
2. They request to login again → 3 tokens remaining
3. This continues until there are no more tokens in the bucket
4. If they request to login again after using all 5 tokens, the API will reject the request since there are no more tokens to utilize

The bucket refills gradually at your configured rate, allowing the user to make requests again once tokens are available.

## Implementation

Now that we understand how the flow works, let's get into the implementation.

### Step 1: Essential Imports

Let's start by importing the crucial packages:

```go
import (
    "sync"
    "time"
)
```

### Step 2: Create the Bucket Structure

```go
type bucketState struct {
    tokens     float64
    lastRefill time.Time
}

type TokenBucket struct {
    mu               sync.RWMutex
    tokensPerSecond  float64
    maxTokens        float64
    tokens           map[string]*bucketState
    cleanupInterval  time.Duration
}
```

**Why `sync.RWMutex`?**

`sync.RWMutex` is a great choice since it reduces lock contention frequency by only blocking calls with write operations, making it favorable especially when trying to control access to a resource. This is important in our case since we're trying to regulate the frequency of access to our endpoints. By ensuring that only one thread can access the shared resource at a time while making others wait, it ensures we don't have race conditions which could cause inconsistencies.

### Step 3: Create a Rate Limiter Factory

This function creates a rate limiter, which is crucial for customizing rate limit rules according to our needs:

```go
func NewTokenBucket(requestsPerSecond float64, maxTokens float64) *TokenBucket {
    tb := &TokenBucket{
        tokensPerSecond: requestsPerSecond,
        maxTokens:       maxTokens,
        tokens:          make(map[string]*bucketState),
        cleanupInterval: 5 * time.Minute,
    }
    
    go tb.cleanup()
    return tb
}
```

By using this function, we can create a custom limiter like this:

```go
limiter := NewTokenBucket(10, 20)
```

This creates a limiter that allows 10 requests per second with a maximum burst of 20.

### Step 4: Implement Allow and Refill Logic

The core functionality checks whether a client is allowed to access the resource while automatically refilling the bucket with tokens at a steady rate:

```go
func (tb *TokenBucket) Allow(clientID string) bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    now := time.Now()
    state, exists := tb.tokens[clientID]
    
    if !exists {
        // First request from this client
        tb.tokens[clientID] = &bucketState{
            tokens:     tb.maxTokens - 1,
            lastRefill: now,
        }
        return true
    }
    
    // Calculate tokens to add based on elapsed time
    elapsed := now.Sub(state.lastRefill).Seconds()
    tokensToAdd := elapsed * tb.tokensPerSecond
    state.tokens = min(state.tokens+tokensToAdd, tb.maxTokens)
    state.lastRefill = now
    
    // Check if we can allow this request
    if state.tokens >= 1 {
        state.tokens--
        return true
    }
    
    return false
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}
```

The function takes a client identifier (which can be an IP address, a user ID, or an API key). It then:

1. Checks if a bucket already exists for this client
2. If not, creates one and uses one token to allow the client access
3. If the bucket exists, refills tokens based on elapsed time since the last refill
4. Consumes one token if available

The bucket refills at a steady rate (`tokensPerSecond`) up to a maximum (`maxToken`). The `Allow` function returns `true` if there's a token that can be consumed (allowing the client to access the resource) and `false` if there are no more tokens (which can then be customized to return an error to the client, like "too many requests").

### Step 5: Extract Client IP Address

Suppose we use an IP address as the client ID (which is the best option). We need a way to extract the client's real IP address. This is important because if your server is sitting behind a load balancer or a reverse proxy like Nginx, you need to extract the client's real IP address and not the proxy's IP address. Otherwise, all requests will be rate-limited because they will seem to originate from the same source.

```go
func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header (set by load balancers/proxies)
    if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
        ips := strings.Split(xForwardedFor, ",")
        return strings.TrimSpace(ips[0])
    }
    
    // Check X-Real-IP header (alternative)
    if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
        return xRealIP
    }
    
    // Fallback to RemoteAddr
    if r.RemoteAddr != "" {
        ip := strings.Split(r.RemoteAddr, ":")[0]
        return ip
    }
    
    return "unknown"
}
```

This function extracts the IP address from the headers. If you're using the client's user ID instead, you can add it to the JWT token, extract it, and use it in your `Allow` function. But for this example, we're using the client's IP address.

### Step 6: Create the Rate Limiting Middleware

Now all that's remaining is creating a middleware that will sit between our route and the request—a rate limiting middleware that acts as our guard and redirects traffic accordingly:

```go
func RateLimitMiddleware(limiter *TokenBucket) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)
            
            if !limiter.Allow(clientIP) {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

Our rate limiting middleware wraps HTTP handlers (your `loginHandler`, `paymentHandler`, etc.) and brings everything together:

1. Extracts the IP address by calling the IP extraction function we created
2. Passes that IP address to the `Allow` function to check if the client is allowed to access the resource
3. If allowed (`true`), we let the request through to the handler
4. If not allowed (`false`), we return a 429 status code with a custom error message to the client

### Step 7: Applying the Middleware

```go
func main() {
    // Create a rate limiter: 5 requests per minute
    loginLimiter := NewTokenBucket(5.0/60.0, 5)
    
    // Create router
    router := http.NewServeMux()
    
    // Apply rate limiting middleware to login handler
    router.Handle("/login", RateLimitMiddleware(loginLimiter)(http.HandlerFunc(loginHandler)))
    
    // Start server
    http.ListenAndServe(":8080", router)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    // Your login logic here
    w.Write([]byte("Login successful"))
}
```

## Understanding the Math

Let's go over the math implementation using the login handler as an example. You can set the `requestsPerSecond` to 5 per minute, which translates to a steady rate of:

```
5 requests / 60 seconds = 0.083 tokens/second
```

With a maximum token amount of 5.

### Example Scenario

Let's say the user last requested to login 30 seconds ago with 2 tokens remaining in their bucket:

1. **Time elapsed**: 30 seconds
2. **Tokens to add**: 30 seconds × 0.083 tokens/second = 2.5 tokens
3. **New total**: 2 + 2.5 = 4.5 tokens (still less than max of 5)

If they request to login again:
- Check: Is 4.5 >= 1? **Yes** ✓
- **Allow** the request
- **Remaining tokens**: 4.5 - 1 = 3.5 tokens

## Complete Request Flow

```
Client Request
    ↓
Middleware
    ↓
Extract Client IP (getClientIP)
    ↓
Check Token Bucket (Allow)
    ↓
    ├─→ Tokens >= 1?
    │       ↓
    │      YES → Consume token → Pass to Handler → Process Request
    │
    └─→ Tokens < 1?
            ↓
           NO → Return 429 Error (Rate limit exceeded)
```

## Memory Management

Add a cleanup function to prevent memory leaks from abandoned IP addresses:

```go
func (tb *TokenBucket) cleanup() {
    ticker := time.NewTicker(tb.cleanupInterval)
    for range ticker.C {
        tb.mu.Lock()
        now := time.Now()
        for key, state := range tb.tokens {
            // Delete entries not accessed in the last 5 minutes
            if now.Sub(state.lastRefill) > tb.cleanupInterval {
                delete(tb.tokens, key)
            }
        }
        tb.mu.Unlock()
    }
}
```

This goroutine runs every 5 minutes and removes entries that haven't been accessed recently, keeping memory usage under control.

## Different Rate Limits for Different Endpoints

You can create different rate limiters for different security needs:

```go
// Strict: Login endpoint (prevent brute force)
loginLimiter := NewTokenBucket(5.0/60.0, 5)  // 5 per minute

// Moderate: API endpoints
apiLimiter := NewTokenBucket(100, 200)  // 100 per second, burst 200

// Strict: Payment endpoints (prevent duplicate charges)
paymentLimiter := NewTokenBucket(5.0/60.0, 5)  // 5 per minute

// Apply to routes
router.Handle("/login", RateLimitMiddleware(loginLimiter)(loginHandler))
router.Handle("/api/tickets", RateLimitMiddleware(apiLimiter)(ticketsHandler))
router.Handle("/payments", RateLimitMiddleware(paymentLimiter)(paymentHandler))
```

## Testing Your Rate Limiter

Test your implementation with curl:

```bash
# Test 6 rapid requests (should allow 5, deny the 6th)
for i in {1..6}; do
    echo "Request $i:"
    curl -X POST http://localhost:8080/login \
         -d '{"email":"test@example.com","password":"test"}' \
         -i | grep "HTTP"
    sleep 0.5
done
```

Expected output:
```
Request 1: HTTP/1.1 200 OK
Request 2: HTTP/1.1 200 OK
Request 3: HTTP/1.1 200 OK
Request 4: HTTP/1.1 200 OK
Request 5: HTTP/1.1 200 OK
Request 6: HTTP/1.1 429 Too Many Requests
```

## Important Considerations

### In-Memory Limitation

This implementation stores rate limit data in memory, which means:

- **Single server**: Works perfectly ✓
- **Multiple servers** (load balanced): Each server tracks its own limits independently
  - Client hitting Server A and Server B could make double the requests
  - **Solution**: Use Redis or another shared datastore for distributed rate limiting

### Production Enhancements

For production systems, consider:

1. **Redis-backed storage** for distributed rate limiting across multiple servers
2. **Informative response headers** so clients know their quota:
   ```go
   w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int(state.tokens)))
   w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
   ```

3. **User-based rate limiting** instead of just IP-based (for authenticated requests)
4. **Monitoring and metrics** to track rate limit hits and adjust thresholds
5. **Graceful configuration** that allows runtime adjustments without restarts

## Conclusion

Every API needs a rate limiting mechanism for security and, most importantly, to prevent someone from crashing your server or abusing your endpoints. The Token Bucket algorithm provides a flexible, burst-friendly approach that balances security with user experience.

However, rate limiting alone does not protect against all attacks. Distributed DDoS attacks, for example, require additional measures such as:

- Web Application Firewalls (WAF)
- Intrusion Prevention Systems (IPS)
- Content Delivery Networks (CDNs) like Cloudflare
- Constant monitoring and alerting

But rate limiting is a great start toward building secure, reliable systems. It's a fundamental building block that every production API should have in place from day one.

---

## Key Takeaways

✓ **Token Bucket** allows smooth traffic handling with burst capability  
✓ **IP-based** rate limiting is simple and effective for public endpoints  
✓ **Middleware pattern** keeps your handlers clean and testable  
✓ **Thread-safe** implementation with `sync.RWMutex` prevents race conditions  
✓ **Memory cleanup** prevents unbounded growth  
✓ **Different tiers** for different endpoints based on security needs  
✓ **Proper IP extraction** handles load balancers and proxies correctly  

Rate limiting is not just a nice-to-have—it's essential infrastructure that protects your API, your database, and your users.
