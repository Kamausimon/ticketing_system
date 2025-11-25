# 🚀 Quick Start Guide - Analytics System

## What's Been Built

✅ **Complete monitoring stack with:**
- Prometheus (metrics collection)
- Grafana (visualization with 3 pre-built dashboards)
- AlertManager (alerting)
- Node Exporter (system metrics)
- cAdvisor (container metrics)

✅ **Application instrumentation:**
- Prometheus middleware in main.go
- 40+ business and system metrics
- Automatic HTTP request tracking
- Database and runtime metrics collection

✅ **Pre-configured dashboards:**
- Business Overview (revenue, sales, conversions)
- Performance & System Metrics
- Payment Analytics

✅ **Alert rules for:**
- Business issues (low sales, inventory)
- Performance problems (errors, latency)
- System issues (memory, connections)

## 🎯 Start in 3 Steps

### 1. Start Monitoring Stack
```bash
./setup-monitoring.sh
```

### 2. Start API Server
```bash
cd cmd/api-server
go run main.go
```

### 3. Open Grafana
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin123`

## 📊 What You'll See

**Dashboards automatically show:**
- Real-time revenue and sales
- Active events count
- Payment success rates
- Request latency and error rates
- System resource usage
- And much more!

## 🔍 Check System Status
```bash
./check-monitoring-status.sh
```

## 📍 Access Points

| Service | URL | Purpose |
|---------|-----|---------|
| Grafana | http://localhost:3000 | Dashboards & visualization |
| Prometheus | http://localhost:9090 | Metrics & queries |
| AlertManager | http://localhost:9093 | Alert management |
| API Metrics | http://localhost:8080/metrics | Raw metrics endpoint |

## 📈 Available Metrics

**Business Metrics:**
- `ticketing_revenue_total` - Total revenue
- `ticketing_tickets_sold_total` - Tickets sold
- `ticketing_orders_completed_total` - Completed orders
- `ticketing_events_active` - Active events
- `ticketing_inventory_available` - Available tickets

**Performance Metrics:**
- `ticketing_http_request_duration_seconds` - Request latency
- `ticketing_http_requests_total` - Total requests
- `ticketing_db_query_duration_seconds` - Database query time
- `ticketing_payment_duration_seconds` - Payment processing time

**System Metrics:**
- `ticketing_memory_usage_bytes` - Memory usage
- `ticketing_goroutines` - Active goroutines
- `ticketing_db_connections` - DB connection pool

## 🛠️ Useful Commands

```bash
# Start monitoring
./setup-monitoring.sh

# Check status
./check-monitoring-status.sh

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f

# Stop monitoring
docker-compose -f docker-compose.monitoring.yml down

# Restart services
docker-compose -f docker-compose.monitoring.yml restart
```

## 📚 Documentation

- **Comprehensive Guide:** `ANALYTICS_SYSTEM_README.md`
- **Technical Details:** `PROMETHEUS_GRAFANA_GUIDE.md`

## 🎓 Example Queries

Try these in Prometheus (http://localhost:9090):

```promql
# Revenue in last hour
sum(increase(ticketing_revenue_total{currency="USD"}[1h]))

# Tickets sold per minute
sum(rate(ticketing_tickets_sold_total[5m])) * 60

# Error rate percentage
sum(rate(ticketing_http_requests_total{status=~"5.."}[5m])) / sum(rate(ticketing_http_requests_total[5m])) * 100

# P95 latency
histogram_quantile(0.95, rate(ticketing_http_request_duration_seconds_bucket[5m]))
```

## 🔔 Alerts

Configured alerts will fire for:
- ❌ No sales in 30 minutes
- ⚠️ Low inventory (< 10 tickets)
- ❌ High payment failure rate (> 10%)
- ❌ High error rate (> 5%)
- ⚠️ High latency (P95 > 2s)
- ❌ Service down

Configure notifications in `prometheus/alertmanager.yml`

## ✅ Verification Checklist

After starting:
- [ ] Grafana loads at http://localhost:3000
- [ ] 3 dashboards are visible in Grafana
- [ ] API server shows metrics at http://localhost:8080/metrics
- [ ] Prometheus shows targets as "UP" at http://localhost:9090/targets
- [ ] Dashboard panels show data (might need to generate some traffic first)

## 🐛 Troubleshooting

**No data in dashboards?**
- Ensure API server is running
- Check http://localhost:8080/metrics shows metrics
- Verify Prometheus targets are UP: http://localhost:9090/targets

**Services not starting?**
- Check Docker is running
- Ensure ports 3000, 8080, 9090, 9093 are available
- View logs: `docker-compose -f docker-compose.monitoring.yml logs`

**Build errors?**
```bash
cd /home/kamau/projects/ticketing_system
go mod tidy
cd cmd/api-server
go build
```

## 🎉 Next Steps

1. ✅ Start monitoring stack
2. ✅ Start API server
3. ✅ Open Grafana dashboards
4. 📊 Generate some test traffic to your API
5. 🔔 Configure alert notifications (Slack/Email)
6. 📈 Customize dashboards for your needs
7. 🚀 Deploy to production

---

**Status:** ✅ Production Ready  
**Components:** 6 Docker containers + Instrumented Go API  
**Dashboards:** 3 pre-configured  
**Metrics:** 40+ tracked  
**Alerts:** 15+ configured
