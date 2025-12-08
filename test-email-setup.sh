#!/bin/bash

# Email System Setup and Test Script - Gmail Diagnostic Version
# This script helps diagnose Gmail SMTP issues

echo "📧 Gmail Email Diagnostic Test"
echo "==============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ .env file not found${NC}"
    exit 1
fi

# Load environment variables
echo "🔍 Loading email configuration from .env..."
set -a
source .env
set +a

# Display current configuration
echo -e "${BLUE}Current Configuration:${NC}"
echo "   Provider: ${EMAIL_PROVIDER:-not set}"
echo "   Host: ${EMAIL_HOST:-not set}"
echo "   Port: ${EMAIL_PORT:-not set}"
echo "   Username: ${EMAIL_USERNAME:-not set}"
echo "   From: ${EMAIL_FROM:-not set}"
echo "   Use TLS: ${EMAIL_USE_TLS:-not set}"
echo "   Test Mode: ${EMAIL_TEST_MODE:-not set (defaults to true)}"
echo ""

# Critical checks
ERRORS=0

if [ -z "$EMAIL_USERNAME" ] || [ -z "$EMAIL_PASSWORD" ]; then
    echo -e "${RED}❌ EMAIL_USERNAME or EMAIL_PASSWORD not set${NC}"
    ERRORS=$((ERRORS + 1))
fi

if [ "$EMAIL_TEST_MODE" != "false" ]; then
    echo -e "${YELLOW}⚠️  WARNING: EMAIL_TEST_MODE is not set to 'false'${NC}"
    echo "   Emails will be logged but NOT actually sent!"
    echo "   Add this to your .env: EMAIL_TEST_MODE=false"
    echo ""
    ERRORS=$((ERRORS + 1))
fi

if [ "$EMAIL_HOST" != "smtp.gmail.com" ]; then
    echo -e "${YELLOW}⚠️  WARNING: EMAIL_HOST is not smtp.gmail.com${NC}"
    echo "   Current: $EMAIL_HOST"
    echo ""
fi

if [ "$EMAIL_PORT" != "587" ]; then
    echo -e "${YELLOW}⚠️  WARNING: EMAIL_PORT should be 587 for Gmail TLS${NC}"
    echo "   Current: $EMAIL_PORT"
    echo ""
fi

# Check for spaces in password (common issue)
if [[ "$EMAIL_PASSWORD" == *" "* ]]; then
    echo -e "${RED}❌ EMAIL_PASSWORD contains spaces!${NC}"
    echo "   Gmail app passwords should be 16 characters without spaces"
    echo "   Example: binorbjfnkqxkrty (not 'bino rbjf nkqx krty')"
    ERRORS=$((ERRORS + 1))
fi

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Found $ERRORS critical issue(s). Please fix them in .env before testing.${NC}"
    echo ""
    exit 1
fi

echo -e "${GREEN}✅ Configuration checks passed${NC}"
echo ""

# Check if server is running
echo "🔍 Checking if API server is running on port 8080..."
if ! curl -s http://localhost:8080/metrics > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  API server is not running${NC}"
    echo "   Starting server in background..."
    go run cmd/api-server/main.go > /tmp/email-test-server.log 2>&1 &
    SERVER_PID=$!
    echo "   Server PID: $SERVER_PID"
    
    # Wait for server to start
    echo "   Waiting for server to start..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/metrics > /dev/null 2>&1; then
            echo -e "${GREEN}   ✅ Server started${NC}"
            break
        fi
        sleep 1
    done
    
    if ! curl -s http://localhost:8080/metrics > /dev/null 2>&1; then
        echo -e "${RED}❌ Failed to start server. Check logs:${NC}"
        cat /tmp/email-test-server.log
        exit 1
    fi
    STARTED_SERVER=1
else
    echo -e "${GREEN}✅ Server is already running${NC}"
    STARTED_SERVER=0
fi
echo ""

# Test email endpoint with detailed output
echo "📧 Sending test email via API..."
echo ""

# Get recipient email or use default
RECIPIENT="${EMAIL_USERNAME:-test@example.com}"
echo "   Recipient: $RECIPIENT"
echo "   Sending via: POST http://localhost:8080/notifications/test"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/notifications/test \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$RECIPIENT\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

echo "   HTTP Status: $HTTP_CODE"
echo "   Response: $BODY"
echo ""

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ API call successful${NC}"
    echo ""
    echo -e "${BLUE}📬 Check your Gmail inbox:${NC}"
    echo "   1. Go to: https://mail.google.com"
    echo "   2. Check inbox for: $RECIPIENT"
    echo "   3. Check spam/junk folder if not in inbox"
    echo "   4. Subject should be: 'Email Configuration Test'"
    echo ""
    echo -e "${YELLOW}⏰ Note: Gmail may take 30-60 seconds to deliver${NC}"
else
    echo -e "${RED}❌ API call failed with status $HTTP_CODE${NC}"
    echo ""
    echo "Server logs (last 20 lines):"
    if [ -f /tmp/email-test-server.log ]; then
        tail -n 20 /tmp/email-test-server.log
    fi
fi

# Cleanup if we started the server
if [ $STARTED_SERVER -eq 1 ]; then
    echo ""
    echo "🧹 Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
    echo "   Server stopped"
fi

echo ""
echo "📚 Troubleshooting Guide:"
echo "   1. No email received?"
echo "      - Check Gmail spam/junk folder"
echo "      - Verify EMAIL_TEST_MODE=false in .env"
echo "      - Ensure app password has no spaces"
echo "      - Check 2FA is enabled on Gmail account"
echo ""
echo "   2. Authentication errors?"
echo "      - Generate new app password at: https://myaccount.google.com/apppasswords"
echo "      - Remove spaces from password in .env"
echo "      - Ensure EMAIL_USERNAME matches FROM address"
echo ""
echo "   3. Still not working?"
echo "      - Check server logs for detailed SMTP errors"
echo "      - Verify Gmail hasn't blocked the login"
echo "      - Try sending from a different Gmail account"
echo ""
echo "🎉 Test complete!"
