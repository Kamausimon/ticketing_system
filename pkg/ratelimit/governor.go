package ratelimit

import (
	"sync"
	"time"
)

// Governor manages rate limiters for different endpoints or resources
type Governor struct {
	mu       sync.RWMutex
	limiters map[string]Limiter
	factory  LimiterFactory
}

// LimiterFactory creates limiters with the given configuration
type LimiterFactory func(Config) Limiter

// NewGovernor creates a new governor with a limiter factory
func NewGovernor(factory LimiterFactory) *Governor {
	return &Governor{
		limiters: make(map[string]Limiter),
		factory:  factory,
	}
}

// NewTokenBucketGovernor creates a governor that uses token bucket limiters
func NewTokenBucketGovernor() *Governor {
	return NewGovernor(func(config Config) Limiter {
		return NewTokenBucket(config)
	})
}

// NewSlidingWindowGovernor creates a governor that uses sliding window limiters
func NewSlidingWindowGovernor() *Governor {
	return NewGovernor(func(config Config) Limiter {
		return NewSlidingWindow(config)
	})
}

// GetOrCreate gets an existing limiter or creates a new one
func (g *Governor) GetOrCreate(name string, config Config) Limiter {
	g.mu.RLock()
	if limiter, exists := g.limiters[name]; exists {
		g.mu.RUnlock()
		return limiter
	}
	g.mu.RUnlock()

	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check in case another goroutine created it
	if limiter, exists := g.limiters[name]; exists {
		return limiter
	}

	limiter := g.factory(config)
	g.limiters[name] = limiter
	return limiter
}

// Get retrieves an existing limiter or returns nil
func (g *Governor) Get(name string) Limiter {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.limiters[name]
}

// Register registers a new limiter
func (g *Governor) Register(name string, limiter Limiter) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.limiters[name] = limiter
}

// Allow checks if a request is allowed using a named limiter
func (g *Governor) Allow(limiterName, key string) bool {
	limiter := g.Get(limiterName)
	if limiter == nil {
		return true // Allow if limiter doesn't exist
	}
	return limiter.Allow(key)
}

// AllowWithResult checks if a request is allowed and returns detailed result
func (g *Governor) AllowWithResult(limiterName, key string) Result {
	limiter := g.Get(limiterName)
	if limiter == nil {
		return Result{Allowed: true}
	}

	// Type assert to get detailed results if available
	switch l := limiter.(type) {
	case *TokenBucket:
		return l.AllowWithResult(key)
	case *SlidingWindow:
		return l.AllowWithResult(key)
	default:
		return Result{Allowed: limiter.Allow(key)}
	}
}

// Reset clears the rate limit state for a key across a named limiter
func (g *Governor) Reset(limiterName, key string) {
	limiter := g.Get(limiterName)
	if limiter != nil {
		limiter.Reset(key)
	}
}

// Preset configuration constants for common rate limit scenarios
var Presets = struct {
	// API rate limits - 100 requests per second with burst of 200
	API Config
	// Auth rate limits - 10 requests per minute per IP
	Auth Config
	// Login rate limits - 5 attempts per minute per IP
	Login Config
	// Payment rate limits - 5 requests per minute
	Payment Config
	// Download rate limits - 3 requests per second per user
	Download Config
}{
	API: Config{
		RequestsPerSecond: 100,
		BurstSize:         200,
		CleanupInterval:   5 * time.Minute,
	},
	Auth: Config{
		RequestsPerSecond: 10.0 / 60,
		BurstSize:         10,
		CleanupInterval:   5 * time.Minute,
	},
	Login: Config{
		RequestsPerSecond: 5.0 / 60,
		BurstSize:         5,
		CleanupInterval:   5 * time.Minute,
	},
	Payment: Config{
		RequestsPerSecond: 5.0 / 60,
		BurstSize:         5,
		CleanupInterval:   5 * time.Minute,
	},
	Download: Config{
		RequestsPerSecond: 3,
		BurstSize:         5,
		CleanupInterval:   5 * time.Minute,
	},
}
