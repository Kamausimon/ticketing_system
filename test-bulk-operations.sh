#!/bin/bash

# Bulk Operations Test Script
# Tests all bulk operations endpoints

BASE_URL="http://localhost:8080"
TOKEN=""
EVENT_ID=1
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "   Bulk Operations Test Suite"
echo "========================================="
echo ""

# Check if server is running
echo "Checking if API server is running..."
if ! curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ API server is not running at ${BASE_URL}${NC}"
    echo "Please start the server first: cd cmd/api-server && go run main.go"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Login to get token
echo "Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "organizer@example.com",
    "password": "password123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Failed to authenticate${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo -e "${GREEN}✓ Authentication successful${NC}"
echo ""

# Test counter
PASSED=0
FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5

    echo -n "Testing $name... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}${endpoint}" \
            -H "Authorization: Bearer ${TOKEN}")
    else
        response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}${endpoint}" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
        if [ ! -z "$body" ]; then
            echo "   Response: $(echo $body | head -c 100)..."
        fi
    else
        echo -e "${RED}✗ FAILED${NC} (Expected $expected_code, got $http_code)"
        FAILED=$((FAILED + 1))
        echo "   Response: $body"
    fi
    echo ""
}

echo "========================================="
echo "   ATTENDEE BULK OPERATIONS"
echo "========================================="
echo ""

# Test 1: Send bulk email to attendees
test_endpoint \
    "Send Bulk Email" \
    "POST" \
    "/attendees/bulk/email" \
    '{
        "event_id": 1,
        "subject": "Test Bulk Email",
        "message": "This is a test message to all attendees.",
        "filters": {
            "has_arrived": false
        }
    }' \
    "200"

# Test 2: Send event update email
test_endpoint \
    "Send Event Update Email" \
    "POST" \
    "/attendees/event/update-email?event_id=1" \
    '{
        "subject": "Event Update",
        "message": "Important event update for all attendees.",
        "only_non_arrived": true
    }' \
    "200"

# Test 3: Export attendees to CSV
test_endpoint \
    "Export Attendees (CSV)" \
    "POST" \
    "/attendees/bulk/export" \
    '{
        "event_id": 1,
        "format": "csv",
        "filters": {
            "has_arrived": true
        }
    }' \
    "200"

# Test 4: Export attendees to JSON
test_endpoint \
    "Export Attendees (JSON)" \
    "POST" \
    "/attendees/bulk/export" \
    '{
        "event_id": 1,
        "format": "json"
    }' \
    "200"

# Test 5: Send email with HTML content
test_endpoint \
    "Send HTML Email" \
    "POST" \
    "/attendees/bulk/email" \
    '{
        "event_id": 1,
        "subject": "HTML Test",
        "message": "Plain text fallback",
        "html_message": "<h1>HTML Test</h1><p>This is an HTML email.</p>"
    }' \
    "200"

# Test 6: Send email to specific ticket classes
test_endpoint \
    "Send Email to Specific Ticket Classes" \
    "POST" \
    "/attendees/bulk/email" \
    '{
        "event_id": 1,
        "subject": "VIP Update",
        "message": "Special message for VIP attendees",
        "filters": {
            "ticket_class_ids": [1]
        }
    }' \
    "200"

echo "========================================="
echo "   REFUND BULK OPERATIONS"
echo "========================================="
echo ""

# Test 7: Get refund statistics
test_endpoint \
    "Get Refund Statistics" \
    "GET" \
    "/refunds/bulk/stats?event_id=1" \
    "" \
    "200"

# Test 8: Process bulk refunds (approve)
test_endpoint \
    "Bulk Approve Refunds" \
    "POST" \
    "/refunds/bulk/process" \
    '{
        "refund_ids": [1, 2],
        "action": "approve"
    }' \
    "200"

# Test 9: Process bulk refunds (reject)
test_endpoint \
    "Bulk Reject Refunds" \
    "POST" \
    "/refunds/bulk/process" \
    '{
        "refund_ids": [3],
        "action": "reject",
        "reason": "Testing bulk rejection"
    }' \
    "200"

# Test 10: Auto-approve eligible refunds
test_endpoint \
    "Auto-Approve Small Refunds" \
    "POST" \
    "/refunds/bulk/auto-approve" \
    '{
        "event_id": 1,
        "max_refund_amount": 50.00,
        "days_before_event": 14
    }' \
    "200"

# Test 11: Auto-approve with date constraint
test_endpoint \
    "Auto-Approve with Date Constraint" \
    "POST" \
    "/refunds/bulk/auto-approve" \
    '{
        "event_id": 1,
        "max_refund_amount": 100.00,
        "days_before_event": 30
    }' \
    "200"

echo "========================================="
echo "   TICKET BULK OPERATIONS"
echo "========================================="
echo ""

# Test 12: Get ticket statistics
test_endpoint \
    "Get Ticket Statistics" \
    "GET" \
    "/tickets/bulk/stats?event_id=1" \
    "" \
    "200"

# Test 13: Export tickets to CSV
test_endpoint \
    "Export Tickets (CSV)" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "csv"
    }' \
    "200"

# Test 14: Export tickets to JSON
test_endpoint \
    "Export Tickets (JSON)" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "json"
    }' \
    "200"

# Test 15: Export with status filter
test_endpoint \
    "Export Active Tickets Only" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "csv",
        "filters": {
            "status": "active"
        }
    }' \
    "200"

# Test 16: Export with check-in filter
test_endpoint \
    "Export Checked-In Tickets" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "json",
        "filters": {
            "is_checked_in": true
        }
    }' \
    "200"

# Test 17: Export with date range
test_endpoint \
    "Export Tickets with Date Range" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "csv",
        "filters": {
            "date_from": "2024-01-01",
            "date_to": "2024-12-31"
        }
    }' \
    "200"

# Test 18: Export with ticket class filter
test_endpoint \
    "Export VIP Tickets" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "json",
        "filters": {
            "ticket_class_ids": [1]
        }
    }' \
    "200"

# Test 19: Bulk update ticket status
test_endpoint \
    "Bulk Update Ticket Status" \
    "POST" \
    "/tickets/bulk/status" \
    '{
        "ticket_ids": [1, 2, 3],
        "status": "cancelled"
    }' \
    "200"

# Test 20: Bulk update to active
test_endpoint \
    "Bulk Update to Active" \
    "POST" \
    "/tickets/bulk/status" \
    '{
        "ticket_ids": [4, 5],
        "status": "active"
    }' \
    "200"

echo "========================================="
echo "   ERROR HANDLING TESTS"
echo "========================================="
echo ""

# Test 21: Missing event_id
test_endpoint \
    "Bulk Email - Missing Event ID" \
    "POST" \
    "/attendees/bulk/email" \
    '{
        "subject": "Test",
        "message": "Test message"
    }' \
    "400"

# Test 22: Invalid format
test_endpoint \
    "Export - Invalid Format" \
    "POST" \
    "/tickets/bulk/export" \
    '{
        "event_id": 1,
        "format": "invalid_format"
    }' \
    "400"

# Test 23: Invalid refund action
test_endpoint \
    "Bulk Refund - Invalid Action" \
    "POST" \
    "/refunds/bulk/process" \
    '{
        "refund_ids": [1],
        "action": "invalid"
    }' \
    "400"

# Test 24: Empty refund IDs
test_endpoint \
    "Bulk Refund - Empty IDs" \
    "POST" \
    "/refunds/bulk/process" \
    '{
        "refund_ids": [],
        "action": "approve"
    }' \
    "400"

# Test 25: Invalid ticket status
test_endpoint \
    "Bulk Update - Invalid Status" \
    "POST" \
    "/tickets/bulk/status" \
    '{
        "ticket_ids": [1],
        "status": "invalid_status"
    }' \
    "400"

echo "========================================="
echo "   TEST SUMMARY"
echo "========================================="
echo ""
echo "Total Tests: $((PASSED + FAILED))"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
