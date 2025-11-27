# Rate Limiter

A comprehensive rate limiting package for Go that provides multiple rate limiting strategies and middleware support for HTTP servers.

## Features

- **Token Bucket Algorithm**: Smooth burst capacity with configurable token refill
- **Sliding Window Algorithm**: Counter-based rate limiting with time windows
- **HTTP Middleware**: Ready-to-use middleware for Gorilla Mux and standard HTTP handlers
- **Governor Pattern**: Manage multiple rate limiters for different endpoints/resources
- **Flexible Key Functions**: Extract rate limit keys from IP, user ID, endpoint, or custom logic
- **Preset Configurations**: Pre-configured rate limits for common scenarios
- **Thread-Safe**: Safe concurrent access with mutex protection
- **Automatic Cleanup**: Background cleanup of stale entries

## Installation

This package is part of the ticketing_system module. Import it as:

```go
import "ticketing_system/pkg/ratelimit"
```

## Quick Start

### Basic Token Bucket

```go
package main

import (
	"ticketing_system/pkg/ratelimit"
	"time"
)

func main() {
	config := ratelimit.Config{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}

	limiter := ratelimit.NewTokenBucket(config)

	// Check if request is allowed
	if limiter.Allow("client-ip:192.168.1.1") {
		// Process request
	} else {
		// Rate limit exceeded
	}
}
```

### HTTP Middleware

```go
import (
	"net/http"
	"ticketing_system/pkg/ratelimit"
)

func setupRoutes(mux *http.ServeMux) {
	// Create limiter
	limiter := ratelimit.NewTokenBucket(ratelimit.Presets.API)

	// Create middleware
	middleware := ratelimit.NewMiddleware(limiter, ratelimit.KeyFuncs.ByIP)

	// Wrap your handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	mux.HandleFunc("/api/endpoint", middleware.HandlerFunc(handler))
}
```

### Governor for Multiple Endpoints

```go
import "ticketing_system/pkg/ratelimit"

func main() {
	// Create a governor
	gov := ratelimit.NewTokenBucketGovernor()

	// Different rate limits for different endpoints
	gov.GetOrCreate("api", ratelimit.Presets.API)
	gov.GetOrCreate("auth", ratelimit.Presets.Auth)
	gov.GetOrCreate("payment", ratelimit.Presets.Payment)

	// Check rate limits
	if gov.Allow("api", "client:192.168.1.1") {
		// Process API request
	}

	if gov.Allow("auth", "client:192.168.1.1") {
		// Process auth request
	}

	result := gov.AllowWithResult("payment", "user-123")
	if !result.Allowed {
		// Too many requests, inform client about retry
		// Use result.RetryAfter for Retry-After header
	}
}
```

## Algorithms

### Token Bucket

The token bucket algorithm allows smooth bursting of traffic while maintaining a consistent long-term rate.

- **How it works**: 
  - Tokens are added at a constant rate (RequestsPerSecond)
  - Each request consumes 1 token
  - Up to BurstSize tokens can be accumulated
  - If no tokens available, request is rejected

- **Best for**: APIs with variable traffic patterns that need to allow bursts

### Sliding Window

The sliding window algorithm maintains a strict request count within a time window.

- **How it works**:
  - Requests are counted within a rolling time window
  - Old requests outside the window are discarded
  - New requests are allowed until the limit is reached
  - Window rolls forward automatically

- **Best for**: Strict rate limiting requirements with no burst allowance

## Preset Configurations

The package includes common rate limit presets:

- `Presets.API`: 100 req/s with burst of 200
- `Presets.Auth`: 10 req/min per IP (prevents brute force)
- `Presets.Login`: 5 attempts per minute (strict login protection)
- `Presets.Payment`: 5 requests per minute (prevents accidental duplicate payments)
- `Presets.Download`: 3 req/s per user (controls bandwidth usage)

## Key Functions

### Built-in Key Functions

- `KeyFuncs.ByIP`: Rate limit by client IP address (default)
- `KeyFuncs.ByUserID`: Rate limit by authenticated user ID
- `KeyFuncs.ByEndpoint`: Rate limit by API endpoint
- `KeyFuncs.ByIPAndEndpoint`: Combine IP and endpoint

### Custom Key Function

```go
customKey := func(r *http.Request) string {
	// Your custom logic
	return r.Header.Get("X-API-Key")
}

middleware := ratelimit.NewMiddleware(limiter, customKey)
```

## Detailed Results

Get more information about rate limit status:

```go
result := limiter.AllowWithResult("client-key")

if !result.Allowed {
	// Set Retry-After header
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
}

// Log remaining requests
log.Printf("Requests remaining: %d", result.Remaining)
```

## Configuration

```go
config := ratelimit.Config{
	RequestsPerSecond: 10,      // Allow 10 requests per second
	BurstSize:         20,      // Allow bursts up to 20 requests
	Window:            time.Second, // Time window for sliding window (optional)
	CleanupInterval:   5 * time.Minute, // Cleanup interval for old entries
}
```

## Integration with Gorilla Mux

```go
import (
	"github.com/gorilla/mux"
	"ticketing_system/pkg/ratelimit"
)

func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Create rate limiter
	limiter := ratelimit.NewTokenBucket(ratelimit.Presets.API)
	middleware := ratelimit.NewMiddleware(limiter, ratelimit.KeyFuncs.ByIP)

	// Apply middleware to specific routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.Handler)

	api.HandleFunc("/users", getUsersHandler).Methods("GET")
	api.HandleFunc("/tickets", getTicketsHandler).Methods("GET")

	return r
}
```

## Thread Safety

All rate limiters are thread-safe and can be safely shared across goroutines.

## Performance Considerations

- Token Bucket: O(1) per request, minimal memory overhead
- Sliding Window: O(n) per request where n is requests within the window, more memory for tracking
- Governor: O(1) lookup, stores multiple limiters

For high-traffic scenarios, Token Bucket is recommended.

## License

Part of the ticketing_system project.
