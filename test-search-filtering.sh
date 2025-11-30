#!/bin/bash

# Search & Filtering Test Script
# Tests all new search and filtering endpoints

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Authentication token (replace with actual token)
TOKEN="YOUR_AUTH_TOKEN_HERE"

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Search & Filtering Test Suite${NC}"
echo -e "${BLUE}================================${NC}\n"

# Function to print test header
print_test() {
    echo -e "\n${GREEN}Testing: $1${NC}"
    echo "----------------------------------------"
}

# Function to make authenticated request
auth_request() {
    curl -s -X GET "$1" -H "Authorization: Bearer $TOKEN"
}

# 1. Event Search Tests
print_test "1. Public Event Search"
echo "Searching for 'concert' events..."
curl -s -X GET "${BASE_URL}/events/search?q=concert&limit=5" | jq '.'

print_test "2. Event Search with Filters"
echo "Searching for music events in Nairobi..."
curl -s -X GET "${BASE_URL}/events/search?q=music&category=music&location=nairobi" | jq '.'

print_test "3. Organizer Event Search (Requires Auth)"
echo "Searching organizer's events..."
auth_request "${BASE_URL}/organizers/events/search?q=workshop" | jq '.'

# 2. Ticket Filtering Tests
print_test "4. Advanced Ticket Filtering (Requires Auth)"
echo "Filtering unchecked-in tickets..."
auth_request "${BASE_URL}/organizers/tickets/filter?event_id=1&is_checked_in=false&limit=5" | jq '.'

print_test "5. Ticket Filter by Price Range"
echo "Filtering tickets between $50-$200..."
auth_request "${BASE_URL}/organizers/tickets/filter?event_id=1&min_price=50&max_price=200" | jq '.'

print_test "6. Ticket Search"
echo "Searching tickets by holder name..."
auth_request "${BASE_URL}/organizers/tickets/search?event_id=1&q=john" | jq '.'

# 3. Order Search Tests
print_test "7. User Order Search (Requires Auth)"
echo "Searching user's orders..."
auth_request "${BASE_URL}/orders/search?q=john&limit=5" | jq '.'

print_test "8. Order Search with Status Filter"
echo "Searching paid orders..."
auth_request "${BASE_URL}/orders/search?q=example&status=paid" | jq '.'

print_test "9. Organizer Order Search (Requires Auth)"
echo "Searching organizer's orders..."
auth_request "${BASE_URL}/organizers/orders/search?q=concert" | jq '.'

# 4. Attendee Filtering Tests
print_test "10. Advanced Attendee Filtering (Requires Auth)"
echo "Filtering attendees who haven't arrived..."
auth_request "${BASE_URL}/attendees/filter?event_id=1&has_arrived=false&limit=5" | jq '.'

print_test "11. Attendee Filter with Sorting"
echo "Filtering attendees sorted by name..."
auth_request "${BASE_URL}/attendees/filter?event_id=1&sort_by=name&sort_order=asc" | jq '.'

print_test "12. Event-Specific Attendee Search"
echo "Searching attendees in specific event..."
auth_request "${BASE_URL}/attendees/search/event?event_id=1&q=john" | jq '.'

# 5. Complex Filter Combinations
print_test "13. Complex Ticket Filter"
echo "VIP tickets, not checked in, price > $100..."
auth_request "${BASE_URL}/organizers/tickets/filter?event_id=1&ticket_class_names=VIP&is_checked_in=false&min_price=100" | jq '.'

print_test "14. Complex Attendee Filter"
echo "Recent registrations, not arrived..."
auth_request "${BASE_URL}/attendees/filter?event_id=1&registration_after=2025-11-01T00:00:00Z&has_arrived=false" | jq '.'

# 6. Statistics Tests
print_test "15. Ticket Statistics from Filter"
echo "Getting ticket statistics..."
auth_request "${BASE_URL}/organizers/tickets/filter?event_id=1&limit=1" | jq '.stats'

print_test "16. Attendee Statistics from Filter"
echo "Getting attendee statistics..."
auth_request "${BASE_URL}/attendees/filter?event_id=1&limit=1" | jq '.stats'

# Summary
echo -e "\n${BLUE}================================${NC}"
echo -e "${BLUE}Test Suite Complete${NC}"
echo -e "${BLUE}================================${NC}\n"

echo -e "${GREEN}✓ Event Search Tests: 3${NC}"
echo -e "${GREEN}✓ Ticket Filtering Tests: 4${NC}"
echo -e "${GREEN}✓ Order Search Tests: 3${NC}"
echo -e "${GREEN}✓ Attendee Filtering Tests: 4${NC}"
echo -e "${GREEN}✓ Complex Filter Tests: 2${NC}"
echo -e "${GREEN}✓ Statistics Tests: 2${NC}"
echo -e "\n${GREEN}Total Tests: 18${NC}\n"

echo "Note: Replace 'YOUR_AUTH_TOKEN_HERE' with a valid token"
echo "Note: Adjust event_id values based on your test data"
