package auth

import (
	"net/http"
	"strings"
)

// AddSecurityHeaders adds security headers to password reset responses
func AddSecurityHeaders(w http.ResponseWriter) {
	// Prevent clickjacking
	w.Header().Set("X-Frame-Options", "DENY")
	// Prevent content type sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Enable XSS protection
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	// Content Security Policy
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	// Referrer Policy
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// Permissions Policy
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// GetClientIP extracts the client IP from the request
// Handles X-Forwarded-For, X-Real-IP headers for proxy support
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Use remote address
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// stringPtr returns a pointer to a string value
func stringPtr(s string) *string {
	return &s
}
