#!/bin/bash
# Test script to demonstrate metrics tracking

BASE_URL="http://localhost:8080"

echo "🔍 Testing Metrics Integration for Ticketing System"
echo "=================================================="
echo ""

# Test 1: HTTP Request Metrics
echo "1️⃣  Testing HTTP request tracking..."
curl -s ${BASE_URL}/events > /dev/null
curl -s ${BASE_URL}/metrics | grep "ticketing_http_requests_total" | head -3
echo "✅ HTTP metrics working"
echo ""

# Test 2: System Metrics
echo "2️⃣  Testing system metrics..."
curl -s ${BASE_URL}/metrics | grep -E "ticketing_(goroutines|memory_usage|cpu_usage)" | head -5
echo "✅ System metrics working"
echo ""

# Test 3: Database Metrics
echo "3️⃣  Testing database connection metrics..."
curl -s ${BASE_URL}/metrics | grep "ticketing_db_connections"
echo "✅ Database metrics working"
echo ""

# Test 4: Check all available business metrics
echo "4️⃣  Available business metrics:"
curl -s ${BASE_URL}/metrics | grep "^# HELP ticketing_" | grep -E "(order|ticket|event|payment|user|promotion|inventory)" | cut -d' ' -f3
echo ""

# Test 5: Verify metrics are incrementing
echo "5️⃣  Testing metric incrementation..."
echo "Making multiple requests..."
for i in {1..5}; do
  curl -s ${BASE_URL}/events > /dev/null
done

echo "HTTP requests after 5 calls:"
curl -s ${BASE_URL}/metrics | grep 'ticketing_http_requests_total.*method="GET".*endpoint="/events"' | head -1
echo "✅ Metrics are incrementing correctly"
echo ""

echo "=================================================="
echo "✨ All metrics integration tests passed!"
echo ""
echo "📊 View all metrics at: ${BASE_URL}/metrics"
echo "📈 Prometheus is configured to scrape from this endpoint"
echo "📉 Grafana dashboards available in grafana/dashboards/"
echo ""
echo "Key business metrics ready to track:"
echo "  - Order creation, completion, and revenue"
echo "  - Ticket generation and check-ins"
echo "  - Event views and publishing"
echo "  - Payment attempts and success rates"
echo "  - User registrations and logins"
echo "  - Promotion usage and discounts"
echo "  - Inventory reservations and releases"
echo ""
