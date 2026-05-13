package ratelimit

import "time"

// Limiter defines the interface for rate limiting strategies
type Limiter interface {
	// Allow checks if a request from the given key is allowed
	Allow(key string) bool
	// Reset clears the rate limit state for the given key
	Reset(key string)
}

// Config holds rate limiting configuration
type Config struct {
	// RequestsPerSecond is the number of requests allowed per second
	RequestsPerSecond float64
	// BurstSize is the maximum number of requests allowed in a burst
	BurstSize int64
	// Window is the time window for sliding window algorithm
	Window time.Duration
	// CleanupInterval is how often to clean up expired entries
	CleanupInterval time.Duration
}

// Result holds the result of a rate limit check
type Result struct {
	Allowed    bool
	Remaining  int64
	ResetAfter time.Duration
	RetryAfter time.Duration
}
