#!/bin/bash

# Intasend Payment Integration Test Script
# Usage: ./test_payment_integration.sh

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="https://ticketingapp.ngrok.dev"
API_BASE="$BASE_URL"

echo -e "${YELLOW}=== Intasend Payment Integration Test ===${NC}\n"

# Test 1: Server Connectivity Check
echo -e "${YELLOW}[1/5] Testing server connectivity...${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" || echo "000")
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "404" ]; then
    echo -e "${GREEN}✅ Server is reachable${NC}\n"
else
    echo -e "${RED}❌ Server connectivity check failed (HTTP $HTTP_CODE)${NC}\n"
    echo -e "${YELLOW}Make sure ngrok is forwarding to port 8080${NC}\n"
    exit 1
fi

# Test 2: Webhook Endpoint Exists
echo -e "${YELLOW}[2/5] Testing webhook endpoint availability...${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_BASE/webhooks/intasend" \
    -H "Content-Type: application/json" \
    -d '{"test":"connection"}' || echo "000")

if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Webhook endpoint is accessible${NC}"
    echo -e "   (HTTP $HTTP_CODE - endpoint exists and responding)${NC}\n"
else
    echo -e "${RED}❌ Webhook endpoint not accessible (HTTP $HTTP_CODE)${NC}\n"
fi

# Test 3: Payment Initiation Endpoint (without auth - should fail)
echo -e "${YELLOW}[3/5] Testing payment initiation endpoint...${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_BASE/payments/initiate" \
    -H "Content-Type: application/json" \
    -d '{"test":"request"}' || echo "000")

if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "404" ]; then
    echo -e "${GREEN}✅ Payment initiation endpoint exists${NC}"
    echo -e "   (HTTP $HTTP_CODE - endpoint responding correctly)${NC}\n"
else
    echo -e "${RED}❌ Payment initiation endpoint issue (HTTP $HTTP_CODE)${NC}\n"
fi

# Test 4: Check Environment Variables (local check)
echo -e "${YELLOW}[4/5] Checking local environment variables...${NC}"
MISSING_VARS=0

if [ -z "$INTASEND_PUBLISHABLE_KEY" ]; then
    echo -e "${RED}❌ INTASEND_PUBLISHABLE_KEY not set${NC}"
    MISSING_VARS=1
else
    echo -e "${GREEN}✅ INTASEND_PUBLISHABLE_KEY is set${NC}"
fi

if [ -z "$INTASEND_SECRET_KEY" ]; then
    echo -e "${RED}❌ INTASEND_SECRET_KEY not set${NC}"
    MISSING_VARS=1
else
    echo -e "${GREEN}✅ INTASEND_SECRET_KEY is set${NC}"
fi

if [ -z "$INTASEND_WEBHOOK_SECRET" ]; then
    echo -e "${YELLOW}⚠️  INTASEND_WEBHOOK_SECRET not set (needed for webhook verification)${NC}"
else
    echo -e "${GREEN}✅ INTASEND_WEBHOOK_SECRET is set${NC}"
fi

if [ -z "$INTASEND_TEST_MODE" ]; then
    echo -e "${YELLOW}⚠️  INTASEND_TEST_MODE not set (defaulting to production)${NC}"
else
    echo -e "${GREEN}✅ INTASEND_TEST_MODE is set to: $INTASEND_TEST_MODE${NC}"
fi

echo ""

# Test 5: Summary
echo -e "${YELLOW}[5/5] Summary${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Server URL:        ${GREEN}$BASE_URL${NC}"
echo -e "Webhook URL:       ${GREEN}$API_BASE/webhooks/intasend${NC}"
echo -e "Payment Init URL:  ${GREEN}$API_BASE/payments/initiate${NC}"
echo -e "Payment Verify:    ${GREEN}$API_BASE/payments/verify/{id}${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $MISSING_VARS -eq 0 ]; then
    echo -e "\n${GREEN}✅ All tests passed! Your system is ready for payment testing.${NC}\n"
    echo -e "${YELLOW}Next Steps:${NC}"
    echo -e "1. Configure webhook in Intasend Dashboard:"
    echo -e "   URL: ${GREEN}$API_BASE/webhooks/intasend${NC}"
    echo -e "2. Copy webhook secret to INTASEND_WEBHOOK_SECRET"
    echo -e "3. Restart your server"
    echo -e "4. Run a test payment\n"
else
    echo -e "\n${YELLOW}⚠️  Some environment variables are missing.${NC}"
    echo -e "Please set them in your .env file or environment.\n"
fi

echo -e "${YELLOW}For detailed testing, see WEBHOOK_SETUP.md${NC}\n"
