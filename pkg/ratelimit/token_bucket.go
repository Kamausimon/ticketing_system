package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu              sync.RWMutex
	tokensPerSecond float64
	maxTokens       float64
	tokens          map[string]*bucketState
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

type bucketState struct {
	tokens    float64
	lastRefil time.Time
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(config Config) *TokenBucket {
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	tb := &TokenBucket{
		tokensPerSecond: config.RequestsPerSecond,
		maxTokens:       float64(config.BurstSize),
		tokens:          make(map[string]*bucketState),
		lastCleanup:     time.Now(),
		cleanupInterval: config.CleanupInterval,
	}

	if tb.maxTokens == 0 {
		tb.maxTokens = config.RequestsPerSecond * 2 // Default burst size is 2x rate
	}

	// Start background cleanup
	go tb.cleanup()

	return tb
}

// Allow checks if a request is allowed under the token bucket algorithm
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	state, exists := tb.tokens[key]

	if !exists {
		// New key: start with full bucket
		tb.tokens[key] = &bucketState{
			tokens:    tb.maxTokens - 1,
			lastRefil: now,
		}
		return true
	}

	// Calculate elapsed time and refill tokens
	elapsed := now.Sub(state.lastRefil).Seconds()
	tokensToAdd := elapsed * tb.tokensPerSecond
	state.tokens = min(state.tokens+tokensToAdd, tb.maxTokens)
	state.lastRefil = now

	// Check if we have at least 1 token
	if state.tokens >= 1 {
		state.tokens--
		return true
	}

	return false
}

// AllowN checks if n requests are allowed
func (tb *TokenBucket) AllowN(key string, n int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	state, exists := tb.tokens[key]

	if !exists {
		// New key: start with full bucket
		state = &bucketState{
			tokens:    tb.maxTokens,
			lastRefil: now,
		}
		tb.tokens[key] = state
	}

	// Calculate elapsed time and refill tokens
	elapsed := now.Sub(state.lastRefil).Seconds()
	tokensToAdd := elapsed * tb.tokensPerSecond
	state.tokens = min(state.tokens+tokensToAdd, tb.maxTokens)
	state.lastRefil = now

	// Check if we have enough tokens
	if state.tokens >= float64(n) {
		state.tokens -= float64(n)
		return true
	}

	return false
}

// AllowWithResult checks if a request is allowed and returns detailed result
func (tb *TokenBucket) AllowWithResult(key string) Result {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	state, exists := tb.tokens[key]

	if !exists {
		state = &bucketState{
			tokens:    tb.maxTokens - 1,
			lastRefil: now,
		}
		tb.tokens[key] = state
		return Result{
			Allowed:    true,
			Remaining:  int64(state.tokens),
			ResetAfter: tb.cleanupInterval,
			RetryAfter: 0,
		}
	}

	// Calculate elapsed time and refill tokens
	elapsed := now.Sub(state.lastRefil).Seconds()
	tokensToAdd := elapsed * tb.tokensPerSecond
	state.tokens = min(state.tokens+tokensToAdd, tb.maxTokens)
	state.lastRefil = now

	allowed := state.tokens >= 1
	if allowed {
		state.tokens--
	}

	// Calculate retry after if not allowed
	retryAfter := time.Duration(0)
	if !allowed && tb.tokensPerSecond > 0 {
		retryAfter = time.Duration((1 - state.tokens) / tb.tokensPerSecond * float64(time.Second))
	}

	return Result{
		Allowed:    allowed,
		Remaining:  int64(max(0, state.tokens)),
		ResetAfter: tb.cleanupInterval,
		RetryAfter: retryAfter,
	}
}

// Reset clears the rate limit state for the given key
func (tb *TokenBucket) Reset(key string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	delete(tb.tokens, key)
}

// cleanup periodically removes old entries
func (tb *TokenBucket) cleanup() {
	ticker := time.NewTicker(tb.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tb.mu.Lock()
		now := time.Now()
		for key, state := range tb.tokens {
			// Remove entries that haven't been accessed in 2x cleanup interval
			if now.Sub(state.lastRefil) > 2*tb.cleanupInterval {
				delete(tb.tokens, key)
			}
		}
		tb.mu.Unlock()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
