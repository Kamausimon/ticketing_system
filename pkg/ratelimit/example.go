package ratelimit

import (
	"fmt"
	"net/http"
	"time"
)

// ExampleTokenBucket demonstrates token bucket rate limiting
func ExampleTokenBucket() {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}

	limiter := NewTokenBucket(config)

	// Simulate requests from different clients
	for i := 0; i < 25; i++ {
		if limiter.Allow("client-1") {
			fmt.Printf("Request %d: allowed\n", i+1)
		} else {
			fmt.Printf("Request %d: blocked\n", i+1)
		}
	}
}

// ExampleSlidingWindow demonstrates sliding window rate limiting
func ExampleSlidingWindow() {
	config := Config{
		RequestsPerSecond: 5,
		BurstSize:         5,
		Window:            time.Second,
	}

	limiter := NewSlidingWindow(config)

	// Simulate requests
	for i := 0; i < 10; i++ {
		if limiter.Allow("user-123") {
			fmt.Printf("Request %d: allowed\n", i+1)
		} else {
			fmt.Printf("Request %d: blocked\n", i+1)
		}
	}

	// Wait and try again
	time.Sleep(time.Second)
	if limiter.Allow("user-123") {
		fmt.Println("After 1 second: allowed")
	}
}

// ExampleMiddleware demonstrates HTTP middleware usage
func ExampleMiddleware() {
	// Create a rate limiter
	config := Config{
		RequestsPerSecond: 100,
		BurstSize:         200,
	}
	limiter := NewTokenBucket(config)

	// Create middleware with custom key function
	middleware := NewMiddleware(limiter, KeyFuncs.ByIP)

	// Wrap your handlers
	myHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	limitedHandler := middleware.HandlerFunc(myHandler)

	// Use in your mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/endpoint", limitedHandler)

	_ = mux // Use in http.ListenAndServe
}

// ExampleGovernor demonstrates using a governor for multiple endpoints
func ExampleGovernor() {
	gov := NewTokenBucketGovernor()

	// Create limiters for different endpoints
	apiLimiter := gov.GetOrCreate("api", Presets.API)
	authLimiter := gov.GetOrCreate("auth", Presets.Auth)
	paymentLimiter := gov.GetOrCreate("payment", Presets.Payment)

	_, _, _ = apiLimiter, authLimiter, paymentLimiter

	// Check rate limits
	if gov.Allow("api", "client-ip:192.168.1.1") {
		fmt.Println("API request allowed")
	}

	if gov.Allow("auth", "client-ip:192.168.1.1") {
		fmt.Println("Auth request allowed")
	}

	result := gov.AllowWithResult("payment", "user-123")
	fmt.Printf("Payment request - Allowed: %v, Remaining: %d, Retry after: %v\n",
		result.Allowed, result.Remaining, result.RetryAfter)
}

// ExampleDetailedResult demonstrates getting detailed rate limit information
func ExampleDetailedResult() {
	config := Config{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}

	limiter := NewTokenBucket(config)

	result := limiter.AllowWithResult("client-1")
	fmt.Printf("Allowed: %v\n", result.Allowed)
	fmt.Printf("Remaining: %d\n", result.Remaining)
	fmt.Printf("Reset after: %v\n", result.ResetAfter)
	fmt.Printf("Retry after: %v\n", result.RetryAfter)
}
