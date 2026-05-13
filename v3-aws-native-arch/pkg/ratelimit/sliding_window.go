package ratelimit

import (
	"sync"
	"time"
)

// SlidingWindow implements a sliding window rate limiter
type SlidingWindow struct {
	mu              sync.RWMutex
	maxRequests     int64
	window          time.Duration
	requests        map[string][]time.Time
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// NewSlidingWindow creates a new sliding window rate limiter
func NewSlidingWindow(config Config) *SlidingWindow {
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	if config.Window == 0 {
		config.Window = time.Second
	}

	sw := &SlidingWindow{
		maxRequests:     config.BurstSize,
		window:          config.Window,
		requests:        make(map[string][]time.Time),
		lastCleanup:     time.Now(),
		cleanupInterval: config.CleanupInterval,
	}

	if sw.maxRequests == 0 {
		sw.maxRequests = int64(config.RequestsPerSecond)
	}

	// Start background cleanup
	go sw.cleanup()

	return sw
}

// Allow checks if a request is allowed under the sliding window algorithm
func (sw *SlidingWindow) Allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.window)

	requests, exists := sw.requests[key]
	if !exists {
		sw.requests[key] = []time.Time{now}
		return true
	}

	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if int64(len(validRequests)) < sw.maxRequests {
		validRequests = append(validRequests, now)
		sw.requests[key] = validRequests
		return true
	}

	sw.requests[key] = validRequests
	return false
}

// AllowN checks if n requests are allowed
func (sw *SlidingWindow) AllowN(key string, n int64) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.window)

	requests, exists := sw.requests[key]
	if !exists {
		if n <= sw.maxRequests {
			times := make([]time.Time, n)
			for i := int64(0); i < n; i++ {
				times[i] = now
			}
			sw.requests[key] = times
			return true
		}
		return false
	}

	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	availableSlots := sw.maxRequests - int64(len(validRequests))
	if n <= availableSlots {
		for i := int64(0); i < n; i++ {
			validRequests = append(validRequests, now)
		}
		sw.requests[key] = validRequests
		return true
	}

	sw.requests[key] = validRequests
	return false
}

// AllowWithResult checks if a request is allowed and returns detailed result
func (sw *SlidingWindow) AllowWithResult(key string) Result {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.window)

	requests, exists := sw.requests[key]
	if !exists {
		sw.requests[key] = []time.Time{now}
		return Result{
			Allowed:    true,
			Remaining:  sw.maxRequests - 1,
			ResetAfter: sw.window,
			RetryAfter: 0,
		}
	}

	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	allowed := int64(len(validRequests)) < sw.maxRequests
	remaining := sw.maxRequests - int64(len(validRequests)) - 1

	if allowed {
		validRequests = append(validRequests, now)
	} else {
		remaining = 0
	}

	sw.requests[key] = validRequests

	retryAfter := time.Duration(0)
	if !allowed && len(validRequests) > 0 {
		oldestRequest := validRequests[0]
		retryAfter = oldestRequest.Add(sw.window).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
	}

	return Result{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetAfter: sw.window,
		RetryAfter: retryAfter,
	}
}

// Reset clears the rate limit state for the given key
func (sw *SlidingWindow) Reset(key string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	delete(sw.requests, key)
}

// cleanup periodically removes old entries
func (sw *SlidingWindow) cleanup() {
	ticker := time.NewTicker(sw.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sw.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-sw.window)

		for key, requests := range sw.requests {

			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(sw.requests, key)
			} else {
				sw.requests[key] = validRequests
			}
		}
		sw.mu.Unlock()
	}
}
