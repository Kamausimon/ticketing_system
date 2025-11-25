# Metrics Tracking Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Ticketing System API                          │
│                   (Port 8080)                                    │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ All HTTP Requests
            ▼
┌───────────────────────────────────────────────────────────────────┐
│               PrometheusMiddleware                                │
│  Tracks: Request Count, Duration, Response Size                  │
└───────────┬───────────────────────────────────────────────────────┘
            │
            ▼
┌────────────────────────────────────────────────────────────────────┐
│                      Request Handlers                              │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │   Orders     │  │   Tickets    │  │   Events     │           │
│  │ • Created    │  │ • Generated  │  │ • Created    │           │
│  │ • Completed  │  │ • Checked-in │  │ • Published  │           │
│  │ • Revenue    │  │ • Refunded   │  │ • Viewed     │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  Payments    │  │    Auth      │  │  Promotions  │           │
│  │ • Attempts   │  │ • Register   │  │ • Usage      │           │
│  │ • Success    │  │ • Login      │  │ • Discounts  │           │
│  │ • Failures   │  │ • Failed     │  │              │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  Inventory   │  │  Attendees   │  │   Refunds    │           │
│  │ • Reserved   │  │ • Check-ins  │  │ • Processed  │           │
│  │ • Available  │  │ • No-shows   │  │ • Amounts    │           │
│  │ • Released   │  │              │  │              │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
└────────────────┬───────────────────────────────────────────────────┘
                 │
                 │ Metrics recorded via
                 │ h.metrics.Track*()
                 ▼
┌────────────────────────────────────────────────────────────────────┐
│              PrometheusMetrics (Analytics Module)                  │
│                                                                    │
│  • 40+ Business Metrics (Counters, Gauges, Histograms)           │
│  • HTTP Metrics (Request tracking)                                │
│  • System Metrics (CPU, Memory, Goroutines)                       │
│  • Database Metrics (Connections, Query duration)                 │
│                                                                    │
└────────────────┬───────────────────────────────────────────────────┘
                 │
                 │ Exposed via /metrics endpoint
                 │ (Prometheus format)
                 ▼
┌────────────────────────────────────────────────────────────────────┐
│                      Prometheus                                    │
│                     (Port 9090)                                    │
│                                                                    │
│  • Scrapes metrics every 15s                                      │
│  • Stores time-series data                                        │
│  • Evaluates alert rules                                          │
│  • Provides PromQL query interface                                │
└────────────────┬───────────────────────────────────────────────────┘
                 │
                 ├──────────────────┬─────────────────┐
                 ▼                  ▼                 ▼
    ┌─────────────────┐  ┌──────────────────┐  ┌────────────────┐
    │    Grafana      │  │  Alertmanager    │  │  Ad-hoc        │
    │   (Port 3000)   │  │   (Port 9093)    │  │  Queries       │
    │                 │  │                  │  │                │
    │ • Dashboards    │  │ • Routes alerts  │  │ • CLI queries  │
    │ • Visualizations│  │ • Notifications  │  │ • API access   │
    │ • Alerts        │  │ • Grouping       │  │                │
    └─────────────────┘  └──────────────────┘  └────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                    System Metrics Collector                        │
│                   (Background Goroutine)                           │
│                                                                    │
│  • Collects metrics every 10s:                                    │
│    - Goroutine count                                              │
│    - Memory usage (heap, stack)                                   │
│    - CPU percentage                                               │
│    - Database connection pool stats                               │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

## Metric Flow Example: Order Creation

```
User creates order
        ↓
[POST /orders] → PrometheusMiddleware
        ↓           ↓
        ↓      HTTP metrics recorded
        ↓
OrderHandler.CreateOrder()
        ↓
┌───────┴─────────────────────────────────────────┐
│                                                  │
│ Business Logic:                                  │
│ 1. Validate request                             │
│ 2. Check inventory                              │
│ 3. Calculate total                              │
│ 4. Create order                                 │
│ 5. Save to database ✓                           │
│                                                  │
│ Metrics Tracking:                               │
│ ✓ h.metrics.TrackOrderCreated("pending")        │
│ ✓ h.metrics.OrderValue.Observe(totalAmount)     │
│                                                  │
└──────────────────────────────────────────────────┘
        ↓
Response sent to user
        ↓
Metrics available at /metrics
        ↓
Prometheus scrapes
        ↓
Visible in Grafana dashboard
```

## Key Metric Types

### Counters (Always Increasing)
```
ticketing_orders_created_total{status="pending"} 1523
ticketing_tickets_generated_total{event_id="42"} 856
ticketing_users_registered_total 2048
```

### Gauges (Can Go Up or Down)
```
ticketing_inventory_available{event_id="42",ticket_class="vip"} 25
ticketing_events_active 142
ticketing_users_active 89
```

### Histograms (Distribution of Values)
```
ticketing_order_processing_duration_seconds_bucket{le="1"} 1234
ticketing_order_processing_duration_seconds_bucket{le="5"} 1456
ticketing_order_value{currency="USD"} 125000.00
```

## Real-time Monitoring Queries

### Business KPIs
```promql
# Total revenue today
sum(increase(ticketing_revenue_total[24h]))

# Orders per minute
rate(ticketing_orders_created_total[1m]) * 60

# Payment success rate
rate(ticketing_payment_success_total[5m]) / 
rate(ticketing_payment_attempts_total[5m]) * 100

# Active events
ticketing_events_active

# Check-in rate
sum(ticketing_tickets_checked_in_total) / 
sum(ticketing_tickets_generated_total) * 100
```

### Performance Metrics
```promql
# 95th percentile order processing time
histogram_quantile(0.95, 
  rate(ticketing_order_processing_duration_seconds_bucket[5m]))

# Requests per second by endpoint
sum by (endpoint) (rate(ticketing_http_requests_total[1m]))

# Error rate
sum(rate(ticketing_http_requests_total{status=~"5.."}[5m])) / 
sum(rate(ticketing_http_requests_total[5m])) * 100
```

### System Health
```promql
# Memory usage
ticketing_memory_usage_bytes

# CPU usage
ticketing_cpu_usage_percent

# Database connections
ticketing_db_connections{state="in_use"}

# Active goroutines
ticketing_goroutines
```

---

**All metrics are now live and tracking!** 🎉
