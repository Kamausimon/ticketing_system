package ratelimit

import (
	"net/http"
	"strings"
)

// Middleware is a rate limiting middleware for HTTP handlers
type Middleware struct {
	limiter Limiter
	keyFunc KeyFunc
}

// KeyFunc extracts the rate limit key from a request
type KeyFunc func(*http.Request) string

// NewMiddleware creates a new rate limiting middleware
func NewMiddleware(limiter Limiter, keyFunc KeyFunc) *Middleware {
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}
	return &Middleware{
		limiter: limiter,
		keyFunc: keyFunc,
	}
}

// Handler wraps an HTTP handler with rate limiting
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := m.keyFunc(r)

		if !m.limiter.Allow(key) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandlerFunc wraps an HTTP handler function with rate limiting
func (m *Middleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := m.keyFunc(r)

		if !m.limiter.Allow(key) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// defaultKeyFunc extracts the client IP address from the request
func defaultKeyFunc(r *http.Request) string {
	// Try X-Forwarded-For header first (for proxies)
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		ip := strings.Split(r.RemoteAddr, ":")[0]
		return ip
	}

	return "unknown"
}

// KeyFuncs provides common key extraction functions
var KeyFuncs = struct {
	// ByIP uses the client's IP address as the rate limit key
	ByIP KeyFunc
	// ByUserID uses the authenticated user ID as the rate limit key
	ByUserID KeyFunc
	// ByEndpoint uses the request endpoint as the rate limit key
	ByEndpoint KeyFunc
	// ByIPAndEndpoint combines IP and endpoint
	ByIPAndEndpoint KeyFunc
}{
	ByIP: defaultKeyFunc,
	ByUserID: func(r *http.Request) string {
		// Extract user ID from context or header
		// This is a placeholder - adjust based on your auth implementation
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			return "user:" + userID
		}
		return defaultKeyFunc(r)
	},
	ByEndpoint: func(r *http.Request) string {
		return r.Method + ":" + r.URL.Path
	},
	ByIPAndEndpoint: func(r *http.Request) string {
		ip := defaultKeyFunc(r)
		return ip + ":" + r.Method + ":" + r.URL.Path
	},
}
