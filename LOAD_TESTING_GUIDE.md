# Load Testing Guide for Ticketing System

## Overview
Test your Railway-hosted ticketing system with realistic traffic patterns.

**Backend:** https://ticketingsystem-production-4a1d.up.railway.app
**Monitoring:** Grafana (http://localhost:3001), Prometheus (http://localhost:9090)

---

## Quick Start

### 1. Install Load Testing Tools

**Option A: hey (Simple, recommended for quick tests)**
```bash
# Ubuntu/Debian
sudo apt-get install hey

# macOS
brew install hey

# Or download binary
wget https://hey-release.s3.us-east-2.amazonaws.com/hey_linux_amd64
chmod +x hey_linux_amd64
sudo mv hey_linux_amd64 /usr/local/bin/hey
```

**Option B: k6 (Advanced, better for complex scenarios)**
```bash
# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# macOS
brew install k6
```

### 2. Make Scripts Executable
```bash
chmod +x load-test.sh
chmod +x realistic-load-test.sh
```

---

## Running Tests

### Test 1: Quick Load Test (Using hey)

**Light load:**
```bash
./load-test.sh light
```

**Medium load:**
```bash
./load-test.sh medium
```

**Heavy load with spike:**
```bash
./load-test.sh heavy
```

**All tests:**
```bash
./load-test.sh all
```

### Test 2: Advanced Load Test (Using k6)

**Run default scenario (gradual ramp-up):**
```bash
k6 run load-test-k6.js
```

**Run with custom VUs and duration:**
```bash
k6 run --vus 100 --duration 5m load-test-k6.js
```

**Generate HTML report:**
```bash
k6 run --out json=results.json load-test-k6.js
```

### Test 3: Realistic User Journey

**Simulate 20 users browsing for 5 minutes:**
```bash
./realistic-load-test.sh
```

**Customize:**
Edit the script and change:
- `CONCURRENT_USERS=20` (number of simultaneous users)
- `DURATION=300` (test duration in seconds)

---

## Test Scenarios Explained

### 1. **Light Load** (Development/Staging)
- 100 requests, 10 concurrent users
- Good for: Testing basic functionality, finding obvious bugs

### 2. **Medium Load** (Expected Traffic)
- 500 requests, 50 concurrent users
- Good for: Simulating normal business hours

### 3. **Heavy Load** (Peak Traffic)
- 1000-2000 requests, 100-200 concurrent users
- Good for: Testing system limits, event launches

### 4. **Spike Test** (Sudden Traffic Burst)
- Sudden increase from 10 to 200 users
- Good for: Testing auto-scaling, when event tickets go on sale

### 5. **Sustained Load** (Soak Test)
- Constant load for extended period (30+ minutes)
- Good for: Finding memory leaks, connection pool issues

---

## Quick Manual Tests

**Single endpoint test:**
```bash
hey -n 1000 -c 50 https://ticketingsystem-production-4a1d.up.railway.app/events
```

**With custom headers:**
```bash
hey -n 500 -c 25 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  https://ticketingsystem-production-4a1d.up.railway.app/api/orders
```

**POST request:**
```bash
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}' \
  https://ticketingsystem-production-4a1d.up.railway.app/login
```

---

## Monitoring During Tests

### 1. **Open Grafana Dashboard**
```bash
# In browser
http://localhost:3001

# Login: admin / admin123
```

### 2. **Watch Key Metrics**
- **Request Rate:** Requests per second
- **Response Time:** P95, P99 latency
- **Error Rate:** 4xx, 5xx errors
- **Database Connections:** Active connections
- **Memory Usage:** Application memory
- **CPU Usage:** CPU utilization

### 3. **Prometheus Queries**
```bash
# Open Prometheus
http://localhost:9090

# Useful queries:
rate(http_requests_total[1m])                    # Request rate
histogram_quantile(0.95, http_request_duration)  # P95 latency
rate(http_requests_total{status=~"5.."}[1m])    # Error rate
```

---

## What to Look For

### ✅ **Good Signs**
- Response times < 500ms for most requests
- Error rate < 1%
- Consistent performance as load increases
- No memory leaks during sustained tests
- Database connections stay within limits

### ⚠️ **Warning Signs**
- Response times > 2s
- Error rate > 5%
- Increasing latency over time
- Memory continuously growing
- Database connection exhaustion

### 🚨 **Critical Issues**
- Service crashes or restarts
- Error rate > 10%
- Complete request failures
- Database connection errors
- Out of memory errors

---

## Interpreting Results

### hey Output:
```
Summary:
  Total:        10.5234 secs
  Slowest:      2.1234 secs
  Fastest:      0.0234 secs
  Average:      0.5234 secs
  
Status code distribution:
  [200] 980 responses   ✅ Good - 98% success rate
  [500] 20 responses    ⚠️  Need investigation
```

### k6 Output:
```
✓ http_req_duration.......avg=123ms p(95)=234ms  ✅ Good
✓ http_req_failed.........rate=0.5%             ✅ Low error rate
✗ errors..................rate=2.3%             ⚠️  Check errors
```

---

## Railway-Specific Considerations

### 1. **Rate Limiting**
Railway may have rate limits. Monitor for 429 responses.

### 2. **Auto-scaling**
Watch if Railway auto-scales your service during tests.

### 3. **Database Connections**
Railway Postgres has connection limits. Monitor active connections.

### 4. **Costs**
Load testing will increase usage. Monitor Railway billing dashboard.

---

## Best Practices

1. **Start Small:** Begin with light tests, gradually increase load
2. **Test Off-Peak:** Don't test during actual user traffic
3. **Monitor Everything:** Keep Grafana open during tests
4. **Document Results:** Save test results for comparison
5. **Test Realistic Scenarios:** Use user journey scripts
6. **Check Database:** Monitor database performance separately
7. **Test Recovery:** See how system recovers after spike

---

## Example Testing Schedule

**Week 1 - Baseline:**
```bash
./load-test.sh light
# Document: Response times, error rate, resource usage
```

**Week 2 - Normal Load:**
```bash
./load-test.sh medium
# Compare: Are metrics still good?
```

**Week 3 - Peak Load:**
```bash
./load-test.sh heavy
# Identify: Breaking point, bottlenecks
```

**Week 4 - Soak Test:**
```bash
# Run for 1 hour
hey -z 1h -c 50 https://ticketingsystem-production-4a1d.up.railway.app/events
# Look for: Memory leaks, connection leaks
```

---

## Troubleshooting

**Problem: "hey: command not found"**
```bash
# Install hey first (see Installation section)
```

**Problem: "Connection refused"**
```bash
# Check if your Railway app is running
curl https://ticketingsystem-production-4a1d.up.railway.app/health
```

**Problem: "Too many 5xx errors"**
```bash
# Check Railway logs
# Check database connections in Grafana
# Reduce concurrent users
```

**Problem: "Metrics not showing in Grafana"**
```bash
# Verify Prometheus is scraping
curl http://localhost:9090/api/v1/targets
# Check prometheus.yml configuration
```

---

## Next Steps After Testing

1. **Optimize Slow Endpoints:** Focus on highest latency
2. **Add Caching:** For frequently accessed data
3. **Database Indexing:** For slow queries
4. **Connection Pooling:** Optimize database connections
5. **CDN for Static Assets:** If serving files
6. **Rate Limiting:** Protect against abuse
7. **Auto-scaling:** Configure Railway scaling rules

---

## Resources

- [hey GitHub](https://github.com/rakyll/hey)
- [k6 Documentation](https://k6.io/docs/)
- [Railway Docs](https://docs.railway.app/)
- [Grafana Tutorials](https://grafana.com/tutorials/)
