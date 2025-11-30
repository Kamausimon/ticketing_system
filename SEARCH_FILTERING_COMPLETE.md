# Search & Filtering Features - Complete Documentation

This document provides comprehensive information about the search and filtering capabilities added to the ticketing system.

## Overview

The following search and filtering features have been implemented:

1. **Event Search** - Dedicated search endpoints for public events and organizer events
2. **Advanced Ticket Filtering** - Comprehensive filtering options for organizers to manage tickets
3. **Order History Search** - Search functionality for user orders and organizer orders
4. **Attendee Search & Filtering** - Advanced filtering and search for event attendees

---

## 1. Event Search

### Public Event Search
Search for publicly available (live) events across multiple fields.

**Endpoint:** `GET /events/search`

**Query Parameters:**
- `q` (required) - Search query string
- `page` (optional, default: 1) - Page number
- `limit` (optional, default: 20, max: 100) - Results per page
- `category` (optional) - Filter by event category
- `location` (optional) - Filter by location
- `start_date` (optional) - Filter events starting from this date (YYYY-MM-DD)
- `end_date` (optional) - Filter events ending before this date (YYYY-MM-DD)
- `sort_by` (optional) - Sort criteria: `date`, `popularity`, `created` (default: popularity)
- `sort_order` (optional) - Sort order: `asc`, `desc`

**Search Fields:**
- Event title
- Event description
- Location
- Tags

**Example Request:**
```bash
curl -X GET "http://localhost:8080/events/search?q=concert&category=music&location=nairobi&sort_by=date"
```

**Response:**
```json
{
  "query": "concert",
  "events": [
    {
      "id": 1,
      "title": "Summer Concert Festival",
      "location": "Nairobi",
      "category": "music",
      "start_date": "2025-12-15T18:00:00Z",
      ...
    }
  ],
  "total_count": 25,
  "page": 1,
  "limit": 20,
  "total_pages": 2
}
```

### Organizer Event Search
Search within organizer's own events (includes all statuses: draft, live, cancelled).

**Endpoint:** `GET /organizers/events/search`

**Authentication:** Required (Organizer role)

**Query Parameters:**
Same as public event search, plus:
- `status` (optional) - Filter by event status: `draft`, `live`, `cancelled`, `completed`

**Example Request:**
```bash
curl -X GET "http://localhost:8080/organizers/events/search?q=workshop&status=live" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 2. Advanced Ticket Filtering

### Filter Event Tickets (Advanced)
Comprehensive filtering for organizers to manage event tickets with statistics.

**Endpoint:** `GET /organizers/tickets/filter`

**Authentication:** Required (Organizer role)

**Query Parameters:**

**Basic Filters:**
- `event_id` (required) - Event ID
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)
- `status` (optional) - Ticket status: `active`, `used`, `cancelled`, `refunded`
- `search` (optional) - Search by ticket number, holder name, or email

**Advanced Filters:**
- `ticket_class_id` (optional) - Filter by specific ticket class
- `ticket_class_names` (optional) - Comma-separated list of ticket class names
- `min_price` (optional) - Minimum ticket price
- `max_price` (optional) - Maximum ticket price
- `is_checked_in` (optional) - Boolean: true/false
- `checked_in_before` (optional) - ISO 8601 datetime
- `checked_in_after` (optional) - ISO 8601 datetime
- `transfer_status` (optional) - `original`, `transferred`, `received`
- `refund_status` (optional) - `refunded`, `not_refunded`
- `order_status` (optional) - Order status: `pending`, `paid`, `cancelled`, etc.
- `payment_status` (optional) - Payment status: `pending`, `paid`, `failed`, etc.
- `start_date` (optional) - Filter by ticket creation date (YYYY-MM-DD)
- `end_date` (optional) - Filter by ticket creation date (YYYY-MM-DD)

**Example Request:**
```bash
curl -X GET "http://localhost:8080/organizers/tickets/filter?event_id=123&is_checked_in=false&min_price=50&max_price=200" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "tickets": [
    {
      "id": 1001,
      "ticket_number": "TKT-1001-ABCD",
      "event_title": "Summer Concert",
      "ticket_class": "VIP",
      "holder_name": "John Doe",
      "holder_email": "john@example.com",
      "status": "active",
      "price": 150.00,
      "checked_in_at": null,
      ...
    }
  ],
  "total_count": 45,
  "page": 1,
  "limit": 20,
  "total_pages": 3,
  "stats": {
    "total_count": 45,
    "active_count": 40,
    "used_count": 3,
    "cancelled_count": 1,
    "refunded_count": 1,
    "check_in_rate": 6.67,
    "total_revenue": 6750.00
  }
}
```

### Search Event Tickets
Search tickets within a specific event.

**Endpoint:** `GET /organizers/tickets/search`

**Authentication:** Required (Organizer role)

**Query Parameters:**
- `event_id` (required) - Event ID
- `q` (required) - Search query
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)
- `status` (optional) - Filter by ticket status

**Search Fields:**
- Ticket number
- Holder name
- Holder email
- Order first name
- Order last name

**Example Request:**
```bash
curl -X GET "http://localhost:8080/organizers/tickets/search?event_id=123&q=john" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 3. Order History Search

### Search User Orders
Search within user's own order history.

**Endpoint:** `GET /orders/search`

**Authentication:** Required

**Query Parameters:**
- `q` (required) - Search query
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)
- `status` (optional) - Order status filter
- `payment_status` (optional) - Payment status filter
- `event_id` (optional) - Filter by event
- `start_date` (optional) - Filter by creation date (YYYY-MM-DD)
- `end_date` (optional) - Filter by creation date (YYYY-MM-DD)

**Search Fields:**
- Email
- First name
- Last name
- Order ID
- Phone number

**Example Request:**
```bash
curl -X GET "http://localhost:8080/orders/search?q=john&status=paid" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "query": "john",
  "orders": [
    {
      "id": 5001,
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "status": "paid",
      "payment_status": "paid",
      "amount": 300.00,
      "event": {
        "id": 123,
        "title": "Summer Concert"
      },
      "created_at": "2025-11-15T10:30:00Z",
      ...
    }
  ],
  "total_count": 8,
  "page": 1,
  "limit": 20,
  "total_pages": 1
}
```

### Search Organizer Orders
Search orders for organizer's events.

**Endpoint:** `GET /organizers/orders/search`

**Authentication:** Required (Organizer role)

**Query Parameters:**
Same as user order search, with additional search field:
- Event title (searches across organizer's events)

**Example Request:**
```bash
curl -X GET "http://localhost:8080/organizers/orders/search?q=concert&payment_status=paid" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 4. Attendee Search & Filtering

### Advanced Attendee Filtering
Comprehensive filtering for attendees with statistics.

**Endpoint:** `GET /attendees/filter`

**Authentication:** Required (Organizer role for event-specific filtering)

**Query Parameters:**

**Basic Filters:**
- `page` (optional, default: 1)
- `limit` (optional, default: 50, max: 100)
- `event_id` (optional) - Filter by event
- `search` (optional) - Search term
- `sort_by` (optional) - `name`, `email`, `arrival_time`, `registration_time` (default: registration_time)
- `sort_order` (optional) - `asc`, `desc` (default: desc)

**Advanced Filters:**
- `ticket_class_id` (optional) - Filter by ticket class
- `has_arrived` (optional) - Boolean: true/false
- `is_refunded` (optional) - Boolean: true/false
- `checked_in_before` (optional) - ISO 8601 datetime
- `checked_in_after` (optional) - ISO 8601 datetime
- `registration_before` (optional) - ISO 8601 datetime
- `registration_after` (optional) - ISO 8601 datetime
- `order_status` (optional) - Filter by associated order status
- `ticket_status` (optional) - Filter by associated ticket status

**Example Request:**
```bash
curl -X GET "http://localhost:8080/attendees/filter?event_id=123&has_arrived=false&sort_by=name" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "attendees": [
    {
      "id": 2001,
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane@example.com",
      "has_arrived": false,
      "is_refunded": false,
      "event": {
        "id": 123,
        "title": "Summer Concert"
      },
      "created_at": "2025-11-20T14:30:00Z",
      ...
    }
  ],
  "total_count": 150,
  "page": 1,
  "limit": 50,
  "total_pages": 3,
  "stats": {
    "total_count": 150,
    "arrived_count": 120,
    "refunded_count": 5,
    "arrival_rate": 80.0
  }
}
```

### Search Attendees (Legacy)
Simple search across all attendees.

**Endpoint:** `GET /attendees/search`

**Query Parameters:**
- `q` (required) - Search query
- `event_id` (optional) - Filter by event

**Search Fields:**
- First name
- Last name
- Email

**Example Request:**
```bash
curl -X GET "http://localhost:8080/attendees/search?q=john&event_id=123" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Search Attendees by Event
Dedicated search within a specific event.

**Endpoint:** `GET /attendees/search/event`

**Authentication:** Required (Organizer role)

**Query Parameters:**
- `event_id` (required) - Event ID
- `q` (required) - Search query
- `page` (optional, default: 1)
- `limit` (optional, default: 50, max: 100)

**Search Fields:**
- First name
- Last name
- Email
- Phone number

**Example Request:**
```bash
curl -X GET "http://localhost:8080/attendees/search/event?event_id=123&q=john" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Common Response Patterns

### Pagination
All list/search endpoints return paginated results with the following structure:

```json
{
  "items": [...],
  "total_count": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5
}
```

### Error Responses

**400 Bad Request:**
```json
{
  "error": "search query (q) is required"
}
```

**401 Unauthorized:**
```json
{
  "error": "authentication required"
}
```

**403 Forbidden:**
```json
{
  "error": "access denied"
}
```

**500 Internal Server Error:**
```json
{
  "error": "failed to fetch events"
}
```

---

## Usage Examples

### 1. Search for concerts in Nairobi
```bash
curl -X GET "http://localhost:8080/events/search?q=concert&location=nairobi"
```

### 2. Find all unchecked-in VIP tickets
```bash
curl -X GET "http://localhost:8080/organizers/tickets/filter?event_id=123&ticket_class_names=VIP&is_checked_in=false" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. Search orders by customer email
```bash
curl -X GET "http://localhost:8080/orders/search?q=customer@example.com" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Filter attendees who haven't arrived
```bash
curl -X GET "http://localhost:8080/attendees/filter?event_id=123&has_arrived=false" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Find transferred tickets
```bash
curl -X GET "http://localhost:8080/organizers/tickets/filter?event_id=123&transfer_status=transferred" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Performance Considerations

1. **Pagination:** Use appropriate `limit` values to avoid large result sets
2. **Indexing:** Database indexes are applied to frequently searched fields:
   - Event: title, location, tags, start_date
   - Ticket: ticket_number, holder_email, status
   - Order: email, first_name, last_name, status
   - Attendee: first_name, last_name, email

3. **Search Optimization:** Search queries use `ILIKE` for case-insensitive matching with wildcard patterns

---

## Summary of New Endpoints

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/events/search` | GET | Public event search | No |
| `/organizers/events/search` | GET | Organizer event search | Yes (Organizer) |
| `/organizers/tickets/filter` | GET | Advanced ticket filtering | Yes (Organizer) |
| `/organizers/tickets/search` | GET | Ticket search | Yes (Organizer) |
| `/orders/search` | GET | User order search | Yes |
| `/organizers/orders/search` | GET | Organizer order search | Yes (Organizer) |
| `/attendees/filter` | GET | Advanced attendee filtering | Yes |
| `/attendees/search/event` | GET | Event-specific attendee search | Yes (Organizer) |

---

## Next Steps

1. **Frontend Integration:** Update UI components to use these new endpoints
2. **Caching:** Consider implementing Redis caching for frequently searched queries
3. **Analytics:** Track search patterns to improve relevance
4. **Export:** Add CSV/Excel export functionality for filtered results
5. **Saved Filters:** Allow users to save commonly used filter combinations

---

## Files Modified/Created

### New Files:
- `internal/events/search.go` - Event search implementation
- `internal/tickets/filter.go` - Advanced ticket filtering
- `internal/orders/search.go` - Order search functionality
- `internal/attendees/filter.go` - Attendee filtering and search

### Modified Files:
- `cmd/api-server/main.go` - Added new route handlers

---

## Testing

Run the following tests to verify functionality:

```bash
# Test event search
curl -X GET "http://localhost:8080/events/search?q=test"

# Test ticket filtering (requires auth and event_id)
curl -X GET "http://localhost:8080/organizers/tickets/filter?event_id=1&is_checked_in=false" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Test order search (requires auth)
curl -X GET "http://localhost:8080/orders/search?q=john" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Test attendee filtering (requires auth)
curl -X GET "http://localhost:8080/attendees/filter?event_id=1&has_arrived=false" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

**Version:** 1.0  
**Last Updated:** November 30, 2025  
**Status:** ✅ Complete & Ready for Use
