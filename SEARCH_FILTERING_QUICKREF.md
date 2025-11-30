# Search & Filtering - Quick Reference Guide

Quick reference for all search and filtering endpoints in the ticketing system.

## Event Search

### Public Event Search
```
GET /events/search?q={query}&category={category}&location={location}
```
**Auth:** Not required  
**Search Fields:** title, description, location, tags

### Organizer Event Search
```
GET /organizers/events/search?q={query}&status={status}
```
**Auth:** Required (Organizer)  
**Search Fields:** title, description, location, tags

---

## Ticket Filtering & Search

### Advanced Ticket Filtering
```
GET /organizers/tickets/filter?event_id={id}&is_checked_in={bool}&min_price={amount}
```
**Auth:** Required (Organizer)  
**Filters:** status, ticket_class_id, price range, check-in status, transfer status, order/payment status  
**Returns:** Tickets + Statistics

### Ticket Search
```
GET /organizers/tickets/search?event_id={id}&q={query}
```
**Auth:** Required (Organizer)  
**Search Fields:** ticket_number, holder_name, holder_email, order names

---

## Order Search

### User Order Search
```
GET /orders/search?q={query}&status={status}&payment_status={status}
```
**Auth:** Required  
**Search Fields:** email, first_name, last_name, order_id, phone_number

### Organizer Order Search
```
GET /organizers/orders/search?q={query}&event_id={id}
```
**Auth:** Required (Organizer)  
**Search Fields:** customer info + event_title

---

## Attendee Filtering & Search

### Advanced Attendee Filtering
```
GET /attendees/filter?event_id={id}&has_arrived={bool}&sort_by={field}
```
**Auth:** Required  
**Filters:** event_id, ticket_class_id, arrival status, refund status, registration times  
**Returns:** Attendees + Statistics

### Event-Specific Attendee Search
```
GET /attendees/search/event?event_id={id}&q={query}
```
**Auth:** Required (Organizer)  
**Search Fields:** first_name, last_name, email, phone_number

### Legacy Attendee Search
```
GET /attendees/search?q={query}&event_id={id}
```
**Auth:** Required  
**Search Fields:** first_name, last_name, email

---

## Common Parameters

### Pagination
- `page` - Page number (default: 1)
- `limit` - Results per page (default: 20-50, max: 100)

### Sorting
- `sort_by` - Sort field
- `sort_order` - `asc` or `desc`

### Date Filters
- Use `YYYY-MM-DD` format for dates
- Use ISO 8601 format for datetimes

---

## Response Format

All endpoints return paginated results:

```json
{
  "query": "search term",      // For search endpoints
  "items": [...],               // Array of results
  "total_count": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5,
  "stats": {...}                // Optional statistics
}
```

---

## Quick Examples

### 1. Search Events
```bash
curl "http://localhost:8080/events/search?q=concert&location=nairobi"
```

### 2. Filter Unchecked Tickets
```bash
curl "http://localhost:8080/organizers/tickets/filter?event_id=123&is_checked_in=false" \
  -H "Authorization: Bearer TOKEN"
```

### 3. Search Orders
```bash
curl "http://localhost:8080/orders/search?q=john@example.com" \
  -H "Authorization: Bearer TOKEN"
```

### 4. Filter Attendees
```bash
curl "http://localhost:8080/attendees/filter?event_id=123&has_arrived=false" \
  -H "Authorization: Bearer TOKEN"
```

---

## Filter Combinations

### Example: VIP Tickets Not Checked In, Price Over $100
```
GET /organizers/tickets/filter?event_id=123&ticket_class_names=VIP&is_checked_in=false&min_price=100
```

### Example: Attendees Registered Last 7 Days, Not Arrived
```
GET /attendees/filter?event_id=123&registration_after=2025-11-23T00:00:00Z&has_arrived=false
```

### Example: Paid Orders with Customer Name
```
GET /orders/search?q=john&status=paid&payment_status=paid
```

---

## Advanced Ticket Filters

**Status Filters:**
- `status` - active, used, cancelled, refunded
- `is_checked_in` - true/false
- `transfer_status` - original, transferred, received
- `refund_status` - refunded, not_refunded

**Price Filters:**
- `min_price` - Minimum ticket price
- `max_price` - Maximum ticket price

**Time Filters:**
- `checked_in_before` / `checked_in_after`
- `start_date` / `end_date`

**Order Filters:**
- `order_status` - pending, paid, cancelled, etc.
- `payment_status` - pending, paid, failed, etc.

---

## Advanced Attendee Filters

**Status Filters:**
- `has_arrived` - true/false
- `is_refunded` - true/false
- `ticket_status` - active, used, cancelled, refunded
- `order_status` - pending, paid, cancelled, etc.

**Time Filters:**
- `checked_in_before` / `checked_in_after`
- `registration_before` / `registration_after`

**Sorting Options:**
- `name` - Sort by last name, first name
- `email` - Sort by email
- `arrival_time` - Sort by check-in time
- `registration_time` - Sort by registration date

---

## Statistics Included

### Ticket Filter Stats
```json
{
  "total_count": 100,
  "active_count": 85,
  "used_count": 10,
  "cancelled_count": 3,
  "refunded_count": 2,
  "check_in_rate": 10.0,
  "total_revenue": 15000.00
}
```

### Attendee Filter Stats
```json
{
  "total_count": 200,
  "arrived_count": 180,
  "refunded_count": 5,
  "arrival_rate": 90.0
}
```

---

## Error Codes

- `400` - Bad Request (missing required parameters)
- `401` - Unauthorized (authentication required)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

---

## Performance Tips

1. Use appropriate `limit` values (20-50 for UI, higher for exports)
2. Combine filters to narrow results before searching
3. Use specific event_id filters when possible
4. Sort by indexed fields for better performance
5. Cache frequently used filter combinations

---

## API Summary Table

| Feature | Endpoint | Auth | Key Params |
|---------|----------|------|------------|
| Event Search (Public) | `/events/search` | No | `q`, `category`, `location` |
| Event Search (Organizer) | `/organizers/events/search` | Yes | `q`, `status` |
| Ticket Filter | `/organizers/tickets/filter` | Yes | `event_id`, `is_checked_in`, price |
| Ticket Search | `/organizers/tickets/search` | Yes | `event_id`, `q` |
| Order Search | `/orders/search` | Yes | `q`, `status` |
| Order Search (Organizer) | `/organizers/orders/search` | Yes | `q`, `event_id` |
| Attendee Filter | `/attendees/filter` | Yes | `event_id`, `has_arrived` |
| Attendee Search | `/attendees/search/event` | Yes | `event_id`, `q` |

---

**Version:** 1.0  
**Last Updated:** November 30, 2025
