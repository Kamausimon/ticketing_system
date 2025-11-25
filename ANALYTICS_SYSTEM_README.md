# Analytics System - Prometheus & Grafana Setup

## 🎯 Overview

Complete monitoring and analytics solution for the ticketing system using **Prometheus** (metrics collection), **Grafana** (visualization), and **AlertManager** (alerting).

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Ticketing System API                         │
│                   (Go Application on :8080)                      │
│                                                                   │
│  • Exposes /metrics endpoint with business & system metrics      │
│  • Prometheus middleware tracks all HTTP requests               │
│  • Custom metrics for tickets, orders, payments, events         │
└─────────────────┬───────────────────────────────────────────────┘
                  │
                  │ HTTP Scrape every 15s
                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Prometheus (:9090)                           │
│                                                                   │
│  • Scrapes metrics from API server                              │
│  • Stores time-series data (30 days retention)                  │
│  • Evaluates alerting rules                                     │
│  • Pre-computes recording rules                                 │
└─────────┬───────────────────────────┬───────────────────────────┘
          │                           │
          │ PromQL Queries            │ Alerts
          ▼                           ▼
┌───────────────────────┐   ┌─────────────────────────┐
│   Grafana (:3000)     │   │  AlertManager (:9093)   │
│                       │   │                         │
│  • Visualizations     │   │  • Route alerts         │
│  • Dashboards         │   │  • Send notifications   │
│  • Custom queries     │   │  • Silence alerts       │
└───────────────────────┘   └─────────────────────────┘
```

## 📦 Components

### 1. **Prometheus** (Port 9090)
- Time-series metrics database
- Scrapes metrics from the API server every 15s
- 30-day data retention
- Built-in query language (PromQL)

### 2. **Grafana** (Port 3000)
- Visualization platform
- Pre-configured dashboards
- Real-time monitoring
- Custom alerts and notifications

### 3. **AlertManager** (Port 9093)
- Alert routing and management
- Notification channels (Slack, Email, PagerDuty)
- Alert grouping and silencing

### 4. **Node Exporter** (Port 9100)
- System-level metrics (CPU, memory, disk, network)
- OS-level monitoring

### 5. **cAdvisor** (Port 8082)
- Container metrics
- Resource usage by container

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- 2GB free RAM
- Ports available: 3000, 8080, 8082, 9090, 9093, 9100

### Installation

1. **Run the setup script:**
   ```bash
   ./setup-monitoring.sh
   ```

2. **Start your API server:**
   ```bash
   cd cmd/api-server
   go run main.go
   ```

3. **Access Grafana:**
   - URL: http://localhost:3000
   - Username: `admin`
   - Password: `admin123`

4. **View metrics:**
   - Your app metrics: http://localhost:8080/metrics
   - Prometheus UI: http://localhost:9090
   - AlertManager: http://localhost:9093

## 📊 Available Dashboards

### 1. Business Overview Dashboard
**Focus:** Revenue, sales, and business KPIs

**Key Metrics:**
- Revenue today (USD)
- Tickets sold (real-time)
- Active events count
- Orders completed
- Revenue trend by hour
- Ticket sales trend
- Revenue by organizer
- Payment methods distribution
- Low inventory alerts
- Conversion rate

**Use Cases:**
- Daily business monitoring
- Sales performance tracking
- Revenue forecasting
- Organizer performance comparison

### 2. Performance & System Metrics Dashboard
**Focus:** Application performance and health

**Key Metrics:**
- Request rate (req/sec)
- P95/P99 latency
- Error rates (4xx, 5xx)
- Database query duration
- Memory usage
- Goroutines count
- DB connection pool status
- Payment processing duration
- Cache hit rate

**Use Cases:**
- Performance optimization
- Capacity planning
- Troubleshooting slow requests
- Resource monitoring

### 3. Payment Analytics Dashboard
**Focus:** Payment processing and transactions

**Key Metrics:**
- Payment success rate
- Payment failures
- Total transactions
- Average payment duration
- Success vs failures trend
- Payment methods distribution
- Gateway performance comparison
- Failure reasons breakdown
- Refunds issued
- Platform fees collected

**Use Cases:**
- Payment gateway monitoring
- Transaction success optimization
- Gateway performance comparison
- Revenue reconciliation

## 📈 Key Metrics Explained

### Business Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `ticketing_revenue_total` | Counter | Total revenue generated | currency, event_id, organizer_id |
| `ticketing_tickets_sold_total` | Counter | Total tickets sold | event_id, ticket_class, organizer_id |
| `ticketing_orders_completed_total` | Counter | Completed orders | payment_method |
| `ticketing_events_active` | Gauge | Currently active events | - |
| `ticketing_inventory_available` | Gauge | Available tickets | event_id, ticket_class |

### Performance Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `ticketing_http_request_duration_seconds` | Histogram | Request latency | method, endpoint |
| `ticketing_http_requests_total` | Counter | Total HTTP requests | method, endpoint, status |
| `ticketing_db_query_duration_seconds` | Histogram | Database query time | operation, table |
| `ticketing_payment_duration_seconds` | Histogram | Payment processing time | gateway |

### System Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `ticketing_goroutines` | Gauge | Active goroutines | - |
| `ticketing_memory_usage_bytes` | Gauge | Memory usage | - |
| `ticketing_db_connections` | Gauge | DB connection pool | state |

## 🔔 Alerting

### Configured Alerts

1. **Business Alerts**
   - No ticket sales (30 min)
   - Low inventory (< 10 tickets)
   - High payment failure rate (> 10%)
   - Significant revenue drop (> 50% vs yesterday)

2. **Performance Alerts**
   - High error rate (> 5%)
   - High latency (P95 > 2s)
   - Slow database queries (P95 > 1s)
   - High database error rate

3. **System Alerts**
   - High memory usage (> 1GB)
   - High goroutine count (> 10,000)
   - DB connection pool exhausted (> 90%)
   - Service down

### Alert Configuration

Edit alert rules in `prometheus/alerts.yml`:
```yaml
- alert: HighPaymentFailureRate
  expr: |
    sum(rate(ticketing_payment_failures_total[5m])) 
    / 
    sum(rate(ticketing_payment_attempts_total[5m])) > 0.1
  for: 3m
  labels:
    severity: critical
  annotations:
    summary: "High payment failure rate"
```

Configure notifications in `prometheus/alertmanager.yml`:
```yaml
receivers:
  - name: 'slack-business'
    slack_configs:
      - channel: '#business-alerts'
        api_url: 'YOUR_SLACK_WEBHOOK_URL'
```

## 🔍 Useful PromQL Queries

### Revenue Analysis
```promql
# Total revenue today (USD)
sum(increase(ticketing_revenue_total{currency="USD"}[1d]))

# Revenue per hour
sum(rate(ticketing_revenue_total{currency="USD"}[5m])) * 3600

# Revenue by organizer
sum by (organizer_id) (ticketing_revenue_total{currency="USD"})
```

### Sales Performance
```promql
# Tickets sold per hour
sum(rate(ticketing_tickets_sold_total[5m])) * 3600

# Sales by event
sum by (event_id) (increase(ticketing_tickets_sold_total[1h]))

# Conversion rate
sum(ticketing_orders_completed_total) / sum(ticketing_http_requests_total{endpoint="/events/*"})
```

### Performance Analysis
```promql
# P95 latency
histogram_quantile(0.95, rate(ticketing_http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(ticketing_http_requests_total{status=~"5.."}[5m])) / sum(rate(ticketing_http_requests_total[5m]))

# Requests per second
sum(rate(ticketing_http_requests_total[5m]))
```

## 🛠️ Management Commands

### Start monitoring stack
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### Stop monitoring stack
```bash
docker-compose -f docker-compose.monitoring.yml down
```

### View logs
```bash
# All services
docker-compose -f docker-compose.monitoring.yml logs -f

# Specific service
docker logs -f ticketing_prometheus
docker logs -f ticketing_grafana
```

### Restart services
```bash
docker-compose -f docker-compose.monitoring.yml restart
```

### Reload Prometheus configuration
```bash
curl -X POST http://localhost:9090/-/reload
```

## 🔧 Configuration Files

```
ticketing_system/
├── prometheus/
│   ├── prometheus.yml        # Prometheus configuration
│   ├── alerts.yml            # Alert rules
│   ├── recording_rules.yml   # Pre-computed queries
│   └── alertmanager.yml      # Alert routing
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml
│   │   └── dashboards/
│   │       └── dashboards.yml
│   └── dashboards/
│       ├── business_overview.json
│       ├── performance_system.json
│       └── payment_analytics.json
├── docker-compose.monitoring.yml
└── setup-monitoring.sh
```

## 📚 Integration with Your Application

The API server (`cmd/api-server/main.go`) is already configured with:

1. **Prometheus middleware** - Automatically tracks all HTTP requests
2. **Metrics endpoint** - Exposes `/metrics` for Prometheus scraping
3. **System metrics collector** - Collects Go runtime and DB metrics every 15s

### Adding Custom Metrics

Example: Track a custom business event
```go
// In your handler
metrics.TrackTicketSale(eventID, ticketClass, organizerID)
metrics.TrackRevenue(amount, currency, eventID, organizerID)
```

All available tracking functions are in `internal/analytics/middleware.go`.

## 🐛 Troubleshooting

### Prometheus not scraping metrics
```bash
# Check if API server is exposing metrics
curl http://localhost:8080/metrics

# Check Prometheus targets
# Visit http://localhost:9090/targets
```

### Grafana dashboards not loading
```bash
# Verify datasource connection
# Grafana → Configuration → Data Sources → Prometheus → Test

# Check dashboard files exist
ls -la grafana/dashboards/
```

### High memory usage
```yaml
# Adjust retention in docker-compose.monitoring.yml
command:
  - "--storage.tsdb.retention.time=15d"  # Reduce from 30d
```

### Missing metrics
```bash
# Check if Go dependencies are installed
go list -m github.com/prometheus/client_golang

# Verify metrics are being created
grep "promauto.New" internal/analytics/metrics.go
```

## 📊 Performance Considerations

- **Scrape interval:** 15s (configurable in prometheus.yml)
- **Data retention:** 30 days (configurable)
- **Memory usage:** ~500MB-1GB for Prometheus
- **Disk space:** ~1-2GB for 30 days of metrics
- **Cardinality:** Keep label values low (< 10,000 unique combinations)

## 🔐 Security Best Practices

1. **Change default passwords:**
   ```yaml
   # In docker-compose.monitoring.yml
   - GF_SECURITY_ADMIN_PASSWORD=your_secure_password
   ```

2. **Restrict network access:**
   - Use firewall rules to limit access to monitoring ports
   - Consider running behind a VPN or reverse proxy

3. **Enable authentication:**
   - Configure Prometheus with basic auth
   - Use Grafana's built-in authentication

4. **Sensitive data:**
   - Don't include PII in metric labels
   - Use aggregated metrics for sensitive information

## 🚀 Advanced Features

### Recording Rules
Pre-computed queries for faster dashboard loading (already configured in `prometheus/recording_rules.yml`).

### Custom Alerts
Add your own alerts to `prometheus/alerts.yml` and reload Prometheus.

### Long-term Storage
For data retention > 30 days, consider:
- Thanos
- Cortex
- VictoriaMetrics
- Periodic export to database

### High Availability
For production, run multiple instances:
- Prometheus with federation
- Grafana with shared database
- AlertManager cluster

## 📖 Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Tutorial](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Best Practices Guide](PROMETHEUS_GRAFANA_GUIDE.md)

## 🆘 Support

For issues or questions:
1. Check logs: `docker-compose -f docker-compose.monitoring.yml logs`
2. Verify configuration files
3. Review the full guide: `PROMETHEUS_GRAFANA_GUIDE.md`

---

**Status:** ✅ Production Ready  
**Last Updated:** November 24, 2025  
**Version:** 1.0.0
