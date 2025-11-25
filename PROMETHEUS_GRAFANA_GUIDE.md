# Prometheus & Grafana Integration Guide

## Overview

This guide explains how to integrate Prometheus and Grafana with the ticketing system for real-time monitoring and analytics.

## Architecture

```
┌─────────────────┐
│   Application   │  Exposes /metrics endpoint
│   (Go API)      │  Instruments business logic
└────────┬────────┘
         │
         │ HTTP Scrape (every 15s)
         ▼
┌─────────────────┐
│   Prometheus    │  Stores time-series data
│   Server        │  Executes queries
└────────┬────────┘
         │
         │ PromQL queries
         ▼
┌─────────────────┐
│    Grafana      │  Visualizes data
│   Dashboards    │  Creates alerts
└─────────────────┘
```

## Data Flow

### 1. Application Instrumentation

The Go application exposes metrics through:
- **Counters**: Monotonically increasing values (tickets sold, revenue)
- **Gauges**: Values that can go up/down (active users, inventory)
- **Histograms**: Distribution of values (request duration, order value)
- **Summaries**: Similar to histograms with quantiles

### 2. Prometheus Scraping

Prometheus scrapes the `/metrics` endpoint at regular intervals (default 15s):

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'ticketing_system'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### 3. Data Storage

Prometheus stores all metrics as time-series data:
```
metric_name{label1="value1", label2="value2"} value timestamp
```

Example:
```
ticketing_tickets_sold_total{event_id="123", ticket_class="VIP", organizer_id="45"} 150 1732435200
```

### 4. Querying with PromQL

Grafana uses PromQL to query Prometheus:

**Rate of ticket sales (tickets/sec):**
```promql
rate(ticketing_tickets_sold_total[5m])
```

**Total revenue by organizer:**
```promql
sum by (organizer_id) (ticketing_revenue_total)
```

**95th percentile response time:**
```promql
histogram_quantile(0.95, rate(ticketing_http_request_duration_seconds_bucket[5m]))
```

## Metrics Exposed

### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ticketing_tickets_sold_total` | Counter | event_id, ticket_class, organizer_id | Total tickets sold |
| `ticketing_revenue_total` | Counter | currency, event_id, organizer_id | Total revenue |
| `ticketing_orders_completed_total` | Counter | payment_method | Completed orders |
| `ticketing_events_active` | Gauge | - | Currently active events |
| `ticketing_inventory_available` | Gauge | event_id, ticket_class | Available tickets |

### Performance Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ticketing_http_request_duration_seconds` | Histogram | method, endpoint | Request latency |
| `ticketing_http_requests_total` | Counter | method, endpoint, status | Total HTTP requests |
| `ticketing_db_query_duration_seconds` | Histogram | operation, table | Database query time |
| `ticketing_payment_duration_seconds` | Histogram | gateway | Payment processing time |

### System Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ticketing_goroutines` | Gauge | - | Active goroutines |
| `ticketing_memory_usage_bytes` | Gauge | - | Memory usage |
| `ticketing_db_connections` | Gauge | state | DB connection pool |

## Setup Instructions

### 1. Install Dependencies

Add to `go.mod`:
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### 2. Update main.go

```go
import (
    "ticketing_system/internal/analytics"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Initialize metrics
    metrics := analytics.NewPrometheusMetrics()
    
    // Add middleware
    router.Use(analytics.PrometheusMiddleware(metrics))
    
    // Expose /metrics endpoint
    router.Handle("/metrics", promhttp.Handler())
    
    // Pass metrics to handlers that need it
    eventHandler := events.NewEventHandler(DB, metrics)
    orderHandler := orders.NewOrderHandler(DB, metrics)
    // ... etc
}
```

### 3. Install Prometheus

**Using Docker:**
```bash
docker run -d -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

**prometheus.yml:**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'ticketing_system'
    static_configs:
      - targets: ['host.docker.internal:8080']
```

### 4. Install Grafana

**Using Docker:**
```bash
docker run -d -p 3000:3000 \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana
```

Access at: http://localhost:3000 (admin/admin)

### 5. Configure Grafana

1. **Add Prometheus Data Source:**
   - Configuration → Data Sources → Add data source
   - Select Prometheus
   - URL: `http://prometheus:9090` (or `http://localhost:9090`)
   - Save & Test

2. **Import Dashboard:**
   - Create → Import
   - Upload the provided JSON dashboard file
   - Select Prometheus data source

## Example Dashboards

### 1. Business Overview Dashboard

**Panels:**
- Total Revenue (Today/Week/Month)
- Tickets Sold (Real-time counter)
- Active Events (Gauge)
- Conversion Rate (Funnel chart)
- Revenue by Organizer (Pie chart)
- Sales Timeline (Time series)

**PromQL Queries:**
```promql
# Total revenue today
sum(increase(ticketing_revenue_total[1d]))

# Tickets sold per hour
sum(rate(ticketing_tickets_sold_total[1h])) * 3600

# Conversion rate
sum(ticketing_orders_completed_total) / sum(ticketing_http_requests_total{endpoint="/events/*"})
```

### 2. Performance Dashboard

**Panels:**
- Request Rate (requests/sec)
- P95 Latency (milliseconds)
- Error Rate (%)
- Database Query Time
- Active Connections

**PromQL Queries:**
```promql
# Request rate
rate(ticketing_http_requests_total[5m])

# P95 latency
histogram_quantile(0.95, rate(ticketing_http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(ticketing_http_requests_total{status=~"5.."}[5m])) / sum(rate(ticketing_http_requests_total[5m]))
```

### 3. Event-Specific Dashboard

Variables: `$event_id`

**Panels:**
- Tickets Available
- Sales Rate
- Revenue Generated
- Page Views
- Conversion Funnel

**PromQL Queries:**
```promql
# Tickets available for event
ticketing_inventory_available{event_id="$event_id"}

# Event page views
rate(ticketing_event_views_total{event_id="$event_id"}[5m])
```

## Alerting Rules

Configure alerts in Prometheus or Grafana:

### Prometheus alerts.yml:
```yaml
groups:
  - name: ticketing_alerts
    interval: 30s
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: rate(ticketing_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/sec"

      # Low inventory
      - alert: LowInventory
        expr: ticketing_inventory_available < 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Low inventory for event {{ $labels.event_id }}"

      # Payment failures
      - alert: HighPaymentFailureRate
        expr: rate(ticketing_payment_failures_total[5m]) / rate(ticketing_payment_attempts_total[5m]) > 0.1
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High payment failure rate: {{ $value }}%"
```

## Data Translation Example

### Raw Application Event → Prometheus Metric

**Application Code:**
```go
// When order is completed
startTime := time.Now()
order := processOrder(orderData)
duration := time.Since(startTime)

// Record metrics
metrics.TrackOrderCompleted(
    order.PaymentMethod,    // "stripe"
    float64(order.Total),   // 150.50
    order.Currency,         // "USD"
    duration,               // 2.3s
)
metrics.TrackRevenue(
    float64(order.Total),
    order.Currency,
    strconv.Itoa(order.EventID),
    strconv.Itoa(order.OrganizerID),
)
```

**Resulting Prometheus Metrics:**
```
ticketing_orders_completed_total{payment_method="stripe"} 1
ticketing_order_value{currency="USD"} 150.5
ticketing_order_processing_duration_seconds_bucket{payment_method="stripe",le="2.5"} 1
ticketing_revenue_total{currency="USD",event_id="123",organizer_id="45"} 150.5
```

**Grafana Query:**
```promql
sum(rate(ticketing_revenue_total{organizer_id="45"}[1h])) * 3600
```

**Displayed as:**
"$5,450/hour" on the dashboard

## Best Practices

### 1. Metric Naming
- Use descriptive names: `ticketing_tickets_sold_total` not `tickets`
- Include units: `_seconds`, `_bytes`, `_total`
- Use consistent prefixes: All metrics start with `ticketing_`

### 2. Label Usage
- Keep cardinality low (< 10,000 unique combinations)
- Don't use user IDs or order IDs as labels
- Use labels for grouping: event_id, organizer_id, status

### 3. Performance
- Counters are cheapest (just increment)
- Histograms are expensive (multiple buckets)
- Don't create metrics in hot paths
- Pre-initialize metrics at startup

### 4. Retention
- Prometheus default: 15 days
- For long-term storage, use:
  - Thanos
  - Cortex
  - VictoriaMetrics
  - Or periodically export to database

## Troubleshooting

### Metrics not appearing
1. Check `/metrics` endpoint returns data
2. Verify Prometheus can reach the application
3. Check Prometheus targets page: http://localhost:9090/targets

### High cardinality warnings
- Reduce label values
- Aggregate at query time instead of metric time
- Use recording rules for expensive queries

### Slow queries
- Use recording rules for complex calculations
- Increase scrape intervals
- Reduce time ranges in dashboards

## Next Steps

1. Implement metrics in your handlers
2. Set up Prometheus and Grafana with Docker Compose
3. Create custom dashboards for your needs
4. Configure alerts for critical metrics
5. Set up long-term storage if needed
