#!/bin/bash

# Prometheus & Grafana Setup Script for Ticketing System
# This script sets up a complete monitoring stack with Prometheus, Grafana, and AlertManager

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up Prometheus & Grafana monitoring stack...${NC}\n"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
if ! command_exists docker; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

if ! command_exists go; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ All prerequisites met${NC}\n"

# Check if directories exist
echo -e "${YELLOW}📁 Checking directory structure...${NC}"
if [ ! -d "prometheus" ] || [ ! -d "grafana/provisioning" ] || [ ! -d "grafana/dashboards" ]; then
    echo -e "${RED}❌ Required directories not found. Make sure you're in the project root.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Directory structure verified${NC}\n"

# Add Prometheus Go client dependencies
echo -e "${YELLOW}📦 Installing Prometheus Go client libraries...${NC}"
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
echo -e "${GREEN}✅ Go dependencies installed${NC}\n"

# Stop existing containers if running
echo -e "${YELLOW}🛑 Stopping any existing monitoring containers...${NC}"
docker-compose -f docker-compose.monitoring.yml down 2>/dev/null || true
echo -e "${GREEN}✅ Existing containers stopped${NC}\n"

# Start monitoring stack
echo -e "${YELLOW}🐳 Starting monitoring stack...${NC}"
docker-compose -f docker-compose.monitoring.yml up -d

# Wait for services to be ready
echo -e "${YELLOW}⏳ Waiting for services to start (this may take 30-60 seconds)...${NC}"
sleep 15

# Check Prometheus
echo -e "\n${YELLOW}Checking Prometheus...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:9090/-/healthy > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Prometheus is running at http://localhost:9090${NC}"
        break
    else
        if [ $i -eq 10 ]; then
            echo -e "${RED}❌ Prometheus failed to start. Check logs: docker logs ticketing_prometheus${NC}"
        else
            sleep 3
        fi
    fi
done

# Check Grafana
echo -e "${YELLOW}Checking Grafana...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:3001/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Grafana is running at http://localhost:3001${NC}"
        echo -e "   ${BLUE}Username: admin${NC}"
        echo -e "   ${BLUE}Password: admin123${NC}"
        break
    else
        if [ $i -eq 10 ]; then
            echo -e "${RED}❌ Grafana failed to start. Check logs: docker logs ticketing_grafana${NC}"
        else
            sleep 3
        fi
    fi
done

# Check AlertManager
echo -e "${YELLOW}Checking AlertManager...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:9093/-/healthy > /dev/null 2>&1; then
        echo -e "${GREEN}✅ AlertManager is running at http://localhost:9093${NC}"
        break
    else
        if [ $i -eq 10 ]; then
            echo -e "${RED}❌ AlertManager failed to start. Check logs: docker logs ticketing_alertmanager${NC}"
        else
            sleep 3
        fi
    fi
done

# Check Node Exporter
echo -e "${YELLOW}Checking Node Exporter...${NC}"
if curl -s http://localhost:9100/metrics > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Node Exporter is running at http://localhost:9100${NC}"
else
    echo -e "${YELLOW}⚠️  Node Exporter may not be running${NC}"
fi

# Check cAdvisor
echo -e "${YELLOW}Checking cAdvisor...${NC}"
if curl -s http://localhost:8082/metrics > /dev/null 2>&1; then
    echo -e "${GREEN}✅ cAdvisor is running at http://localhost:8082${NC}"
else
    echo -e "${YELLOW}⚠️  cAdvisor may not be running${NC}"
fi

# Display summary
echo -e "\n${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Monitoring stack setup complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

echo -e "${BLUE}📊 Access Points:${NC}"
echo -e "   • Grafana:      ${YELLOW}http://localhost:3001${NC} (admin/admin123)"
echo -e "   • Prometheus:   ${YELLOW}http://localhost:9090${NC}"
echo -e "   • AlertManager: ${YELLOW}http://localhost:9093${NC}"
echo -e "   • Node Exporter: ${YELLOW}http://localhost:9100${NC}"
echo -e "   • cAdvisor:     ${YELLOW}http://localhost:8082${NC}\n"

echo -e "${BLUE}📈 Available Dashboards:${NC}"
echo -e "   • Business Overview"
echo -e "   • Performance & System Metrics"
echo -e "   • Payment Analytics\n"

echo -e "${BLUE}🔧 Next Steps:${NC}"
echo -e "   1. Start your API server: ${YELLOW}cd cmd/api-server && go run main.go${NC}"
echo -e "   2. Your app will expose metrics at: ${YELLOW}http://localhost:8080/metrics${NC}"
echo -e "   3. Visit Grafana dashboards to see live metrics"
echo -e "   4. Configure alerts in AlertManager (prometheus/alertmanager.yml)\n"

echo -e "${BLUE}📖 Documentation:${NC}"
echo -e "   • Full guide: ${YELLOW}PROMETHEUS_GRAFANA_GUIDE.md${NC}"
echo -e "   • View logs: ${YELLOW}docker-compose -f docker-compose.monitoring.yml logs -f${NC}"
echo -e "   • Stop stack: ${YELLOW}docker-compose -f docker-compose.monitoring.yml down${NC}\n"

echo -e "${GREEN}Happy monitoring! 🎉${NC}"
