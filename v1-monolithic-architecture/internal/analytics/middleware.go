package analytics

import (
	"net/http"
	"strconv"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// PrometheusMiddleware instruments HTTP requests with Prometheus metrics
func PrometheusMiddleware(metrics *PrometheusMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			endpoint := r.URL.Path
			method := r.Method
			status := strconv.Itoa(rw.statusCode)

			// HTTP metrics
			metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
			metrics.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(rw.size))
		})
	}
}

// TrackTicketSale records a ticket sale
func (m *PrometheusMetrics) TrackTicketSale(eventID, ticketClass, organizerID string) {
	m.TicketsSold.WithLabelValues(eventID, ticketClass, organizerID).Inc()
}

// TrackOrderCreated records an order creation
func (m *PrometheusMetrics) TrackOrderCreated(status string) {
	m.OrdersCreated.WithLabelValues(status).Inc()
}

// TrackOrderCompleted records a completed order
func (m *PrometheusMetrics) TrackOrderCompleted(paymentMethod string, value float64, currency string, duration time.Duration) {
	m.OrdersCompleted.WithLabelValues(paymentMethod).Inc()
	m.OrderValue.WithLabelValues(currency).Observe(value)
	m.OrderProcessingTime.WithLabelValues(paymentMethod).Observe(duration.Seconds())
}

// TrackRevenue records revenue
func (m *PrometheusMetrics) TrackRevenue(amount float64, currency, eventID, organizerID string) {
	m.RevenueTotal.WithLabelValues(currency, eventID, organizerID).Add(amount)
}

// TrackPaymentAttempt records a payment attempt
func (m *PrometheusMetrics) TrackPaymentAttempt(gateway, method string) {
	m.PaymentAttempts.WithLabelValues(gateway, method).Inc()
}

// TrackPaymentSuccess records a successful payment
func (m *PrometheusMetrics) TrackPaymentSuccess(gateway, method string, duration time.Duration) {
	m.PaymentSuccess.WithLabelValues(gateway, method).Inc()
	m.PaymentDuration.WithLabelValues(gateway).Observe(duration.Seconds())
}

// TrackPaymentFailure records a failed payment
func (m *PrometheusMetrics) TrackPaymentFailure(gateway, method, errorType string) {
	m.PaymentFailures.WithLabelValues(gateway, method, errorType).Inc()
}

// TrackEventCreated records an event creation
func (m *PrometheusMetrics) TrackEventCreated(category, organizerID string) {
	m.EventsCreated.WithLabelValues(category, organizerID).Inc()
}

// TrackEventView records an event page view
func (m *PrometheusMetrics) TrackEventView(eventID string) {
	m.EventViews.WithLabelValues(eventID).Inc()
}

// UpdateInventory updates inventory gauges
func (m *PrometheusMetrics) UpdateInventory(eventID, ticketClass string, available, reserved float64) {
	m.InventoryAvailable.WithLabelValues(eventID, ticketClass).Set(available)
	m.InventoryReserved.WithLabelValues(eventID, ticketClass).Set(reserved)
}

// TrackInventoryRelease records inventory release
func (m *PrometheusMetrics) TrackInventoryRelease(eventID, reason string) {
	m.InventoryReleased.WithLabelValues(eventID, reason).Inc()
}

// TrackLoginAttempt records a login attempt
func (m *PrometheusMetrics) TrackLoginAttempt(status string) {
	m.LoginAttempts.WithLabelValues(status).Inc()
}

// TrackPromotionUsage records promotion usage
func (m *PrometheusMetrics) TrackPromotionUsage(promotionID, code string, discount float64, currency string) {
	m.PromotionUsage.WithLabelValues(promotionID, code).Inc()
	m.PromotionDiscount.WithLabelValues(promotionID, currency).Add(discount)
}

// TrackDBQuery records database query metrics
func (m *PrometheusMetrics) TrackDBQuery(operation, table string, duration time.Duration) {
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// TrackDBError records database errors
func (m *PrometheusMetrics) TrackDBError(operation, errorType string) {
	m.DBErrors.WithLabelValues(operation, errorType).Inc()
}

// UpdateDBConnections updates database connection metrics
func (m *PrometheusMetrics) UpdateDBConnections(idle, inUse, max int) {
	m.DBConnections.WithLabelValues("idle").Set(float64(idle))
	m.DBConnections.WithLabelValues("in_use").Set(float64(inUse))
	m.DBConnections.WithLabelValues("max").Set(float64(max))
}

// TrackCacheHit records a cache hit
func (m *PrometheusMetrics) TrackCacheHit(cacheName string) {
	m.CacheHits.WithLabelValues(cacheName).Inc()
}

// TrackCacheMiss records a cache miss
func (m *PrometheusMetrics) TrackCacheMiss(cacheName string) {
	m.CacheMisses.WithLabelValues(cacheName).Inc()
}
