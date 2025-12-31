package main

import (
	"fmt"
	"log"
	"net/http"
	"ticketing_system/pkg/ratelimit"
	"time"

	"github.com/gorilla/mux"
)

func InitializeRateLimiting() *ratelimit.Governor {
	gov := ratelimit.NewTokenBucketGovernor()

	gov.GetOrCreate("api", ratelimit.Presets.API)           // 100 req/s, burst 200
	gov.GetOrCreate("auth", ratelimit.Presets.Auth)         // 10 req/min
	gov.GetOrCreate("login", ratelimit.Presets.Login)       // 5 attempts/min
	gov.GetOrCreate("payment", ratelimit.Presets.Payment)   // 5 req/min
	gov.GetOrCreate("download", ratelimit.Presets.Download) // 3 req/s

	return gov
}

func RateLimitingMiddleware(gov *ratelimit.Governor, limiterName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keyFunc := ratelimit.KeyFuncs.ByIP
			key := keyFunc(r)

			result := gov.AllowWithResult(limiterName, key)

			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(result.ResetAfter).Unix()))

			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Example handlers
func getTicketsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Get tickets endpoint"}`)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Login endpoint"}`)
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Payment endpoint"}`)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File content")
}

// SetupRoutesWithRateLimiting demonstrates how to set up routes with rate limiting
func SetupRoutesWithRateLimiting() *mux.Router {
	gov := InitializeRateLimiting()
	r := mux.NewRouter()

	// Public API routes - standard rate limit
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(RateLimitingMiddleware(gov, "api"))
	apiRouter.HandleFunc("/tickets", getTicketsHandler).Methods("GET")

	// Authentication routes - stricter rate limit
	authRouter := r.PathPrefix("/auth").Subrouter()
	authRouter.Use(RateLimitingMiddleware(gov, "login"))
	authRouter.HandleFunc("/login", loginHandler).Methods("POST")

	// Payment routes - strict rate limit
	paymentRouter := r.PathPrefix("/payments").Subrouter()
	paymentRouter.Use(RateLimitingMiddleware(gov, "payment"))
	paymentRouter.HandleFunc("/process", paymentHandler).Methods("POST")

	// Download routes - bandwidth control
	downloadRouter := r.PathPrefix("/downloads").Subrouter()
	downloadRouter.Use(RateLimitingMiddleware(gov, "download"))
	downloadRouter.HandleFunc("/ticket/{id}", downloadHandler).Methods("GET")

	return r
}

// Example: Custom key function for per-user rate limiting
func CustomKeyFunction(r *http.Request) string {
	// Try to get authenticated user ID
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return "user:" + userID
	}

	// Fall back to IP-based rate limiting
	return ratelimit.KeyFuncs.ByIP(r)
}

// Example: Per-endpoint rate limiting
func ExampleMain() {
	// Initialize rate limiter governor
	gov := InitializeRateLimiting()

	// Create router with rate limiting
	router := SetupRoutesWithRateLimiting()

	// You can also use custom middleware for specific use cases
	customLimiter := ratelimit.NewTokenBucket(ratelimit.Config{
		RequestsPerSecond: 1, // Very strict: 1 req/s
		BurstSize:         2, // Allow burst of 2
		CleanupInterval:   5 * time.Minute,
	})

	customMiddleware := ratelimit.NewMiddleware(customLimiter, CustomKeyFunction)

	// Apply to specific handler
	router.HandleFunc("/admin/sensitive", customMiddleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Sensitive admin operation")
	})).Methods("POST")

	// Start server
	log.Println("Starting server on :8080 with rate limiting")
	_ = router
	_ = gov
	// http.ListenAndServe(":8080", router)
}

// Example: Dynamic rate limit adjustment
func AdjustRateLimits(gov *ratelimit.Governor, limiterName string, newConfig ratelimit.Config) {
	// This demonstrates how you could potentially reload rate limits
	// In a production system, you might read from a config file or database
	gov.GetOrCreate(limiterName, newConfig)
	fmt.Printf("Updated rate limit for %s: %.0f req/s, burst %d\n",
		limiterName, newConfig.RequestsPerSecond, newConfig.BurstSize)
}
