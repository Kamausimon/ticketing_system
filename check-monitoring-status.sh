#!/bin/bash

# Monitoring Stack Status Checker
# Quickly check the health of all monitoring components

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}    Ticketing System Monitoring Stack Status${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

# Function to check service health
check_service() {
    local name=$1
    local url=$2
    local container=$3
    
    echo -ne "${YELLOW}Checking $name...${NC} "
    
    # Check if container is running
    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        # Check HTTP endpoint
        if curl -s "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Running${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️  Container running but endpoint not responding${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Not running${NC}"
        return 1
    fi
}

# Function to check container status
check_container() {
    local name=$1
    local container=$2
    
    echo -ne "${YELLOW}Checking $name...${NC} "
    
    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        echo -e "${GREEN}✅ Running${NC}"
        return 0
    else
        echo -e "${RED}❌ Not running${NC}"
        return 1
    fi
}

# Check core services
echo -e "${BLUE}Core Monitoring Services:${NC}"
check_service "Prometheus" "http://localhost:9090/-/healthy" "ticketing_prometheus"
check_service "Grafana" "http://localhost:3001/api/health" "ticketing_grafana"
check_service "AlertManager" "http://localhost:9093/-/healthy" "ticketing_alertmanager"

echo ""
echo -e "${BLUE}System Exporters:${NC}"
check_container "Node Exporter" "ticketing_node_exporter"
check_container "cAdvisor" "ticketing_cadvisor"

# Check API server
echo ""
echo -e "${BLUE}Application:${NC}"
echo -ne "${YELLOW}Checking API Server...${NC} "
if curl -s http://localhost:8080/metrics > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Running & exposing metrics${NC}"
    API_RUNNING=true
else
    echo -e "${RED}❌ Not running or not exposing metrics${NC}"
    API_RUNNING=false
fi

# Show metrics count if API is running
if [ "$API_RUNNING" = true ]; then
    METRICS_COUNT=$(curl -s http://localhost:8080/metrics | grep -c "^ticketing_")
    echo -e "   ${BLUE}Metrics exposed: ${YELLOW}${METRICS_COUNT}${NC}"
fi

# Show Prometheus targets status
echo ""
echo -e "${BLUE}Prometheus Targets:${NC}"
if command -v jq >/dev/null 2>&1; then
    TARGETS=$(curl -s http://localhost:9090/api/v1/targets 2>/dev/null | jq -r '.data.activeTargets[] | "\(.labels.job): \(.health)"' 2>/dev/null)
    if [ -n "$TARGETS" ]; then
        echo "$TARGETS" | while IFS=: read -r job health; do
            if [ "$health" = " up" ]; then
                echo -e "   ${GREEN}✅${NC} $job"
            else
                echo -e "   ${RED}❌${NC} $job"
            fi
        done
    else
        echo -e "   ${YELLOW}⚠️  Could not fetch target status${NC}"
    fi
else
    echo -e "   ${YELLOW}⚠️  Install 'jq' for detailed target status${NC}"
fi

# Show active alerts
echo ""
echo -e "${BLUE}Active Alerts:${NC}"
if command -v jq >/dev/null 2>&1; then
    ALERTS=$(curl -s http://localhost:9090/api/v1/alerts 2>/dev/null | jq -r '.data.alerts[] | select(.state=="firing") | .labels.alertname' 2>/dev/null)
    if [ -n "$ALERTS" ]; then
        echo "$ALERTS" | while read -r alert; do
            echo -e "   ${RED}🚨${NC} $alert"
        done
    else
        echo -e "   ${GREEN}✅ No alerts firing${NC}"
    fi
else
    echo -e "   ${YELLOW}⚠️  Install 'jq' for alert status${NC}"
fi

# Show disk usage
echo ""
echo -e "${BLUE}Storage Usage:${NC}"
PROMETHEUS_SIZE=$(docker exec ticketing_prometheus du -sh /prometheus 2>/dev/null | awk '{print $1}')
if [ -n "$PROMETHEUS_SIZE" ]; then
    echo -e "   Prometheus data: ${YELLOW}${PROMETHEUS_SIZE}${NC}"
fi

GRAFANA_SIZE=$(docker exec ticketing_grafana du -sh /var/lib/grafana 2>/dev/null | awk '{print $1}')
if [ -n "$GRAFANA_SIZE" ]; then
    echo -e "   Grafana data: ${YELLOW}${GRAFANA_SIZE}${NC}"
fi

# Show quick stats
echo ""
echo -e "${BLUE}Quick Stats (Last 5 minutes):${NC}"
if [ "$API_RUNNING" = true ]; then
    REQUESTS=$(curl -s 'http://localhost:9090/api/v1/query?query=sum(increase(ticketing_http_requests_total[5m]))' 2>/dev/null | jq -r '.data.result[0].value[1]' 2>/dev/null)
    if [ -n "$REQUESTS" ] && [ "$REQUESTS" != "null" ]; then
        echo -e "   HTTP Requests: ${YELLOW}$(printf "%.0f" $REQUESTS)${NC}"
    fi
    
    TICKETS=$(curl -s 'http://localhost:9090/api/v1/query?query=sum(increase(ticketing_tickets_sold_total[5m]))' 2>/dev/null | jq -r '.data.result[0].value[1]' 2>/dev/null)
    if [ -n "$TICKETS" ] && [ "$TICKETS" != "null" ]; then
        echo -e "   Tickets Sold: ${YELLOW}$(printf "%.0f" $TICKETS)${NC}"
    fi
    
    ORDERS=$(curl -s 'http://localhost:9090/api/v1/query?query=sum(increase(ticketing_orders_completed_total[5m]))' 2>/dev/null | jq -r '.data.result[0].value[1]' 2>/dev/null)
    if [ -n "$ORDERS" ] && [ "$ORDERS" != "null" ]; then
        echo -e "   Orders Completed: ${YELLOW}$(printf "%.0f" $ORDERS)${NC}"
    fi
fi

# Summary and links
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Access Links:${NC}"
echo -e "   Grafana:      ${YELLOW}http://localhost:3001${NC}"
echo -e "   Prometheus:   ${YELLOW}http://localhost:9090${NC}"
echo -e "   AlertManager: ${YELLOW}http://localhost:9093${NC}"
echo -e "   API Metrics:  ${YELLOW}http://localhost:8080/metrics${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
