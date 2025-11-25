package analytics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics holds all Prometheus metrics for the application
type PrometheusMetrics struct {
	// HTTP Metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Business Metrics - Tickets
	TicketsSold        *prometheus.CounterVec
	TicketsGenerated   *prometheus.CounterVec
	TicketsCheckedIn   *prometheus.CounterVec
	TicketsRefunded    *prometheus.CounterVec
	TicketsTransferred prometheus.Counter

	// Business Metrics - Orders
	OrdersCreated       *prometheus.CounterVec
	OrdersCompleted     *prometheus.CounterVec
	OrdersFailed        *prometheus.CounterVec
	OrderValue          *prometheus.HistogramVec
	OrderProcessingTime *prometheus.HistogramVec

	// Business Metrics - Revenue
	RevenueTotal      *prometheus.CounterVec
	PlatformFeesTotal *prometheus.CounterVec
	RefundsTotal      *prometheus.CounterVec

	// Business Metrics - Events
	EventsCreated   *prometheus.CounterVec
	EventsPublished prometheus.Counter
	EventsCancelled prometheus.Counter
	ActiveEvents    prometheus.Gauge
	EventViews      *prometheus.CounterVec

	// Business Metrics - Users
	UsersRegistered prometheus.Counter
	UsersActive     prometheus.Gauge
	LoginAttempts   *prometheus.CounterVec
	SessionDuration prometheus.Histogram

	// Business Metrics - Inventory
	InventoryAvailable *prometheus.GaugeVec
	InventoryReserved  *prometheus.GaugeVec
	InventoryReleased  *prometheus.CounterVec

	// Payment Metrics
	PaymentAttempts *prometheus.CounterVec
	PaymentSuccess  *prometheus.CounterVec
	PaymentFailures *prometheus.CounterVec
	PaymentDuration *prometheus.HistogramVec

	// Promotions Metrics
	PromotionUsage    *prometheus.CounterVec
	PromotionDiscount *prometheus.CounterVec

	// Database Metrics
	DBQueryDuration *prometheus.HistogramVec
	DBConnections   *prometheus.GaugeVec
	DBErrors        *prometheus.CounterVec

	// System Metrics
	GoroutinesCount prometheus.Gauge
	MemoryUsage     prometheus.Gauge
	CPUUsage        prometheus.Gauge

	// Cache Metrics
	CacheHits      *prometheus.CounterVec
	CacheMisses    *prometheus.CounterVec
	CacheEvictions *prometheus.CounterVec
}

// NewPrometheusMetrics initializes and registers all Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint"},
		),

		// Tickets
		TicketsSold: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_tickets_sold_total",
				Help: "Total number of tickets sold",
			},
			[]string{"event_id", "ticket_class", "organizer_id"},
		),
		TicketsGenerated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_tickets_generated_total",
				Help: "Total number of tickets generated",
			},
			[]string{"event_id", "order_id"},
		),
		TicketsCheckedIn: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_tickets_checked_in_total",
				Help: "Total number of tickets checked in",
			},
			[]string{"event_id"},
		),
		TicketsRefunded: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_tickets_refunded_total",
				Help: "Total number of tickets refunded",
			},
			[]string{"event_id", "reason"},
		),
		TicketsTransferred: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ticketing_tickets_transferred_total",
				Help: "Total number of tickets transferred",
			},
		),

		// Orders
		OrdersCreated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_orders_created_total",
				Help: "Total number of orders created",
			},
			[]string{"status"},
		),
		OrdersCompleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_orders_completed_total",
				Help: "Total number of orders completed",
			},
			[]string{"payment_method"},
		),
		OrdersFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_orders_failed_total",
				Help: "Total number of failed orders",
			},
			[]string{"reason"},
		),
		OrderValue: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_order_value",
				Help:    "Order value distribution",
				Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000},
			},
			[]string{"currency"},
		),
		OrderProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_order_processing_duration_seconds",
				Help:    "Order processing time in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"payment_method"},
		),

		// Revenue
		RevenueTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_revenue_total",
				Help: "Total revenue generated",
			},
			[]string{"currency", "event_id", "organizer_id"},
		),
		PlatformFeesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_platform_fees_total",
				Help: "Total platform fees collected",
			},
			[]string{"currency"},
		),
		RefundsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_refunds_total",
				Help: "Total refunds issued",
			},
			[]string{"currency", "reason"},
		),

		// Events
		EventsCreated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_events_created_total",
				Help: "Total number of events created",
			},
			[]string{"category", "organizer_id"},
		),
		EventsPublished: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ticketing_events_published_total",
				Help: "Total number of events published",
			},
		),
		EventsCancelled: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ticketing_events_cancelled_total",
				Help: "Total number of events cancelled",
			},
		),
		ActiveEvents: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ticketing_events_active",
				Help: "Number of currently active events",
			},
		),
		EventViews: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_event_views_total",
				Help: "Total number of event page views",
			},
			[]string{"event_id"},
		),

		// Users
		UsersRegistered: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ticketing_users_registered_total",
				Help: "Total number of registered users",
			},
		),
		UsersActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ticketing_users_active",
				Help: "Number of currently active users",
			},
		),
		LoginAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_login_attempts_total",
				Help: "Total number of login attempts",
			},
			[]string{"status"},
		),
		SessionDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ticketing_session_duration_seconds",
				Help:    "User session duration in seconds",
				Buckets: []float64{60, 300, 600, 1800, 3600, 7200},
			},
		),

		// Inventory
		InventoryAvailable: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ticketing_inventory_available",
				Help: "Current available inventory",
			},
			[]string{"event_id", "ticket_class"},
		),
		InventoryReserved: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ticketing_inventory_reserved",
				Help: "Current reserved inventory",
			},
			[]string{"event_id", "ticket_class"},
		),
		InventoryReleased: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_inventory_released_total",
				Help: "Total inventory released from reservations",
			},
			[]string{"event_id", "reason"},
		),

		// Payments
		PaymentAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_payment_attempts_total",
				Help: "Total payment attempts",
			},
			[]string{"gateway", "method"},
		),
		PaymentSuccess: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_payment_success_total",
				Help: "Total successful payments",
			},
			[]string{"gateway", "method"},
		),
		PaymentFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_payment_failures_total",
				Help: "Total failed payments",
			},
			[]string{"gateway", "method", "error_type"},
		),
		PaymentDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_payment_duration_seconds",
				Help:    "Payment processing duration",
				Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"gateway"},
		),

		// Promotions
		PromotionUsage: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_promotion_usage_total",
				Help: "Total promotion code usage",
			},
			[]string{"promotion_id", "code"},
		),
		PromotionDiscount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_promotion_discount_total",
				Help: "Total discount amount from promotions",
			},
			[]string{"promotion_id", "currency"},
		),

		// Database
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ticketing_db_query_duration_seconds",
				Help:    "Database query duration",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
			},
			[]string{"operation", "table"},
		),
		DBConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ticketing_db_connections",
				Help: "Database connection pool status",
			},
			[]string{"state"}, // idle, in_use, max
		),
		DBErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_db_errors_total",
				Help: "Total database errors",
			},
			[]string{"operation", "error_type"},
		),

		// System
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ticketing_goroutines",
				Help: "Current number of goroutines",
			},
		),
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ticketing_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ticketing_cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),

		// Cache
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_cache_hits_total",
				Help: "Total cache hits",
			},
			[]string{"cache_name"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_cache_misses_total",
				Help: "Total cache misses",
			},
			[]string{"cache_name"},
		),
		CacheEvictions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ticketing_cache_evictions_total",
				Help: "Total cache evictions",
			},
			[]string{"cache_name", "reason"},
		),
	}

	return metrics
}
