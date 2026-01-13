#!/bin/bash

# Load Testing Script for Ticketing System
# Backend: https://ticketingsystem-production-4a1d.up.railway.app

API_URL="https://ticketingsystem-production-4a1d.up.railway.app"
RESULTS_DIR="./load-test-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

echo -e "${GREEN}=== Ticketing System Load Test ===${NC}"
echo "Timestamp: $TIMESTAMP"
echo "Target: $API_URL"
echo ""

# Function to check if tool is installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Install with: sudo apt-get install $1  (or brew install $1 on macOS)"
        exit 1
    fi
}

# Test 1: Health Check Endpoint (Light Load)
test_health_check() {
    echo -e "${YELLOW}Test 1: Health Check (100 requests, 10 concurrent)${NC}"
    hey -n 100 -c 10 \
        "$API_URL/health" \
        2>&1 | tee "$RESULTS_DIR/health_check_$TIMESTAMP.txt"
    echo ""
}

# Test 2: Events List (Medium Load)
test_events_list() {
    echo -e "${YELLOW}Test 2: Events List (500 requests, 50 concurrent)${NC}"
    hey -n 500 -c 50 \
        "$API_URL/events" \
        2>&1 | tee "$RESULTS_DIR/events_list_$TIMESTAMP.txt"
    echo ""
}

# Test 3: Search Events (Heavy Load)
test_search_events() {
    echo -e "${YELLOW}Test 3: Search Events (1000 requests, 100 concurrent)${NC}"
    hey -n 1000 -c 100 \
        "$API_URL/events/search?query=concert" \
        2>&1 | tee "$RESULTS_DIR/search_events_$TIMESTAMP.txt"
    echo ""
}

# Test 4: Spike Test (Sudden Traffic Burst)
test_spike() {
    echo -e "${YELLOW}Test 4: Spike Test (2000 requests, 200 concurrent)${NC}"
    hey -n 2000 -c 200 \
        "$API_URL/events" \
        2>&1 | tee "$RESULTS_DIR/spike_test_$TIMESTAMP.txt"
    echo ""
}

# Test 5: Sustained Load (Duration-based)
test_sustained() {
    echo -e "${YELLOW}Test 5: Sustained Load (30 seconds, 50 concurrent)${NC}"
    hey -z 30s -c 50 \
        "$API_URL/events" \
        2>&1 | tee "$RESULTS_DIR/sustained_load_$TIMESTAMP.txt"
    echo ""
}

# Test 6: Metrics Endpoint (to verify monitoring)
test_metrics() {
    echo -e "${YELLOW}Test 6: Metrics Endpoint (50 requests, 5 concurrent)${NC}"
    hey -n 50 -c 5 \
        "$API_URL/metrics" \
        2>&1 | tee "$RESULTS_DIR/metrics_$TIMESTAMP.txt"
    echo ""
}

# Main execution
main() {
    # Check if hey is installed
    check_tool hey
    
    echo -e "${GREEN}Starting load tests...${NC}"
    echo ""
    
    # Run tests based on user selection
    case "${1:-all}" in
        light)
            test_health_check
            test_metrics
            ;;
        medium)
            test_health_check
            test_events_list
            test_metrics
            ;;
        heavy)
            test_events_list
            test_search_events
            test_spike
            ;;
        sustained)
            test_sustained
            ;;
        all)
            test_health_check
            sleep 5
            test_events_list
            sleep 5
            test_search_events
            sleep 5
            test_spike
            sleep 5
            test_sustained
            sleep 5
            test_metrics
            ;;
        *)
            echo "Usage: $0 {light|medium|heavy|sustained|all}"
            echo ""
            echo "  light     - Basic health checks"
            echo "  medium    - Normal traffic simulation"
            echo "  heavy     - High traffic with spike test"
            echo "  sustained - Long-running load test"
            echo "  all       - Run all tests (default)"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}=== Load Test Complete ===${NC}"
    echo "Results saved to: $RESULTS_DIR"
    echo ""
    echo -e "${YELLOW}View metrics at:${NC}"
    echo "  Grafana: http://localhost:3001"
    echo "  Prometheus: http://localhost:9090"
    echo ""
}

main "$@"
