package ratelimit

import (
	"testing"
	"time"
)

// TestTokenBucketAllow tests basic token bucket allowing
func TestTokenBucketAllow(t *testing.T) {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}
	limiter := NewTokenBucket(config)

	for i := 0; i < 20; i++ {
		if !limiter.Allow("client-1") {
			t.Fatalf("Expected allow at request %d, got rejected", i+1)
		}
	}

	if limiter.Allow("client-1") {
		t.Fatal("Expected rejection after burst, got allowed")
	}
}

// TestTokenBucketRefill tests token refill
func TestTokenBucketRefill(t *testing.T) {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         10,
	}
	limiter := NewTokenBucket(config)

	for i := 0; i < 10; i++ {
		if !limiter.Allow("client-1") {
			t.Fatalf("Expected allow at request %d, got rejected", i+1)
		}
	}

	if limiter.Allow("client-1") {
		t.Fatal("Expected rejection when out of tokens")
	}

	time.Sleep(150 * time.Millisecond)

	if !limiter.Allow("client-1") {
		t.Fatal("Expected allow after refill")
	}
}

// TestTokenBucketAllowN tests allowing N tokens
func TestTokenBucketAllowN(t *testing.T) {
	config := Config{
		RequestsPerSecond: 100,
		BurstSize:         100,
	}
	limiter := NewTokenBucket(config)

	if !limiter.AllowN("client-1", 50) {
		t.Fatal("Expected allow for 50 tokens")
	}

	if !limiter.AllowN("client-1", 50) {
		t.Fatal("Expected allow for another 50 tokens")
	}

	if limiter.AllowN("client-1", 1) {
		t.Fatal("Expected rejection when out of tokens")
	}
}

// TestTokenBucketResult tests detailed result information
func TestTokenBucketResult(t *testing.T) {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         10,
	}
	limiter := NewTokenBucket(config)

	result := limiter.AllowWithResult("client-1")
	if !result.Allowed {
		t.Fatal("Expected first request to be allowed")
	}

	if result.Remaining <= 0 {
		t.Fatal("Expected remaining tokens to be positive")
	}

	if result.ResetAfter <= 0 {
		t.Fatal("Expected reset after duration to be positive")
	}
}

// TestTokenBucketReset tests reset functionality
func TestTokenBucketReset(t *testing.T) {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         10,
	}
	limiter := NewTokenBucket(config)

	for i := 0; i < 10; i++ {
		limiter.Allow("client-1")
	}

	if limiter.Allow("client-1") {
		t.Fatal("Expected rejection when out of tokens")
	}

	limiter.Reset("client-1")

	if !limiter.Allow("client-1") {
		t.Fatal("Expected allow after reset")
	}
}

// TestSlidingWindowAllow tests basic sliding window allowing
func TestSlidingWindowAllow(t *testing.T) {
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         5,
		Window:            time.Second,
	}
	limiter := NewSlidingWindow(config)

	for i := 0; i < 5; i++ {
		if !limiter.Allow("client-1") {
			t.Fatalf("Expected allow at request %d, got rejected", i+1)
		}
	}

	if limiter.Allow("client-1") {
		t.Fatal("Expected rejection beyond burst size")
	}
}

// TestSlidingWindowWindow tests sliding window time window
func TestSlidingWindowWindow(t *testing.T) {
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         5,
		Window:            time.Second,
	}
	limiter := NewSlidingWindow(config)

	for i := 0; i < 5; i++ {
		limiter.Allow("client-1")
	}

	if limiter.Allow("client-1") {
		t.Fatal("Expected rejection when window full")
	}

	time.Sleep(1100 * time.Millisecond)

	if !limiter.Allow("client-1") {
		t.Fatal("Expected allow after window passed")
	}
}

// TestSlidingWindowAllowN tests allowing N requests
func TestSlidingWindowAllowN(t *testing.T) {
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         10,
		Window:            time.Second,
	}
	limiter := NewSlidingWindow(config)

	if !limiter.AllowN("client-1", 5) {
		t.Fatal("Expected allow for 5 requests")
	}

	// Should allow another 5 requests
	if !limiter.AllowN("client-1", 5) {
		t.Fatal("Expected allow for another 5 requests")
	}

	// Should reject 1 more request
	if limiter.AllowN("client-1", 1) {
		t.Fatal("Expected rejection when window full")
	}
}

// TestSlidingWindowResult tests detailed result information
func TestSlidingWindowResult(t *testing.T) {
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         5,
		Window:            time.Second,
	}
	limiter := NewSlidingWindow(config)

	result := limiter.AllowWithResult("client-1")
	if !result.Allowed {
		t.Fatal("Expected first request to be allowed")
	}

	if result.Remaining <= 0 {
		t.Fatal("Expected remaining to be positive")
	}

	if result.ResetAfter <= 0 {
		t.Fatal("Expected reset after to be positive")
	}
}

// TestGovernor tests governor functionality
func TestGovernor(t *testing.T) {
	gov := NewTokenBucketGovernor()

	// Create limiters with presets
	gov.GetOrCreate("api", Presets.API)
	gov.GetOrCreate("auth", Presets.Auth)

	// Should allow requests
	if !gov.Allow("api", "client-1") {
		t.Fatal("Expected API request to be allowed")
	}

	if !gov.Allow("auth", "client-1") {
		t.Fatal("Expected auth request to be allowed")
	}

	// Get limiter
	limiter := gov.Get("api")
	if limiter == nil {
		t.Fatal("Expected to retrieve API limiter")
	}

	// Get non-existent limiter
	limiter = gov.Get("nonexistent")
	if limiter != nil {
		t.Fatal("Expected nil for non-existent limiter")
	}
}

// TestGovernorAllowWithResult tests governor result reporting
func TestGovernorAllowWithResult(t *testing.T) {
	gov := NewSlidingWindowGovernor()
	gov.GetOrCreate("api", Presets.API)

	result := gov.AllowWithResult("api", "client-1")
	if !result.Allowed {
		t.Fatal("Expected request to be allowed")
	}

	if result.Remaining < 0 {
		t.Fatal("Expected remaining to be >= 0")
	}
}

// TestGovernorReset tests governor reset
func TestGovernorReset(t *testing.T) {
	gov := NewTokenBucketGovernor()
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         5,
	}
	gov.GetOrCreate("test", config)

	// Consume requests
	for i := 0; i < 5; i++ {
		gov.Allow("test", "client-1")
	}

	// Should be rejected
	if gov.Allow("test", "client-1") {
		t.Fatal("Expected rejection when out of tokens")
	}

	// Reset
	gov.Reset("test", "client-1")

	// Should be allowed
	if !gov.Allow("test", "client-1") {
		t.Fatal("Expected allow after reset")
	}
}

// TestConcurrency tests concurrent access
func TestConcurrency(t *testing.T) {
	limiter := NewTokenBucket(Config{
		RequestsPerSecond: 1000,
		BurstSize:         10000,
	})

	// Create channels for synchronization
	done := make(chan bool, 100)
	errors := make(chan error, 100)

	// Launch concurrent requests from 100 clients
	for i := 0; i < 100; i++ {
		go func(clientID int) {
			defer func() { done <- true }()
			key := "client-" + string(rune(clientID))
			for j := 0; j < 100; j++ {
				_ = limiter.Allow(key)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	if len(errors) > 0 {
		t.Fatalf("Expected no errors, got %d", len(errors))
	}
}

// BenchmarkTokenBucketAllow benchmarks token bucket allow
func BenchmarkTokenBucketAllow(b *testing.B) {
	limiter := NewTokenBucket(Config{
		RequestsPerSecond: 1000,
		BurstSize:         10000,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("client-1")
	}
}

// BenchmarkSlidingWindowAllow benchmarks sliding window allow
func BenchmarkSlidingWindowAllow(b *testing.B) {
	limiter := NewSlidingWindow(Config{
		RequestsPerSecond: 1000,
		BurstSize:         10000,
		Window:            time.Second,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("client-1")
	}
}

// BenchmarkTokenBucketAllowN benchmarks token bucket allowN
func BenchmarkTokenBucketAllowN(b *testing.B) {
	limiter := NewTokenBucket(Config{
		RequestsPerSecond: 1000,
		BurstSize:         10000,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AllowN("client-1", 1)
	}
}

// BenchmarkTokenBucketConcurrent benchmarks token bucket under concurrent load
func BenchmarkTokenBucketConcurrent(b *testing.B) {
	limiter := NewTokenBucket(Config{
		RequestsPerSecond: 1000,
		BurstSize:         10000,
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			limiter.Allow("client-" + string(rune(i%100)))
			i++
		}
	})
}
