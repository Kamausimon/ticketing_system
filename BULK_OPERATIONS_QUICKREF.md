# Bulk Operations - Quick Reference

## Endpoints

### Attendees
- `POST /attendees/bulk/email` - Send bulk emails
- `POST /attendees/bulk/export` - Export attendees
- `POST /attendees/event/update-email` - Send event updates

### Refunds
- `POST /refunds/bulk/process` - Process multiple refunds
- `POST /refunds/bulk/auto-approve` - Auto-approve eligible refunds
- `GET /refunds/bulk/stats` - Get refund statistics

### Tickets
- `POST /tickets/bulk/export` - Export tickets
- `GET /tickets/bulk/stats` - Get ticket statistics
- `POST /tickets/bulk/status` - Update ticket statuses

---

## Quick Examples

### Send Email to Non-Arrived Attendees
```bash
curl -X POST http://localhost:8080/attendees/bulk/email \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event_id": 1,
    "subject": "Reminder",
    "message": "Event is tomorrow!",
    "filters": {"has_arrived": false}
  }'
```

### Auto-Approve Small Refunds
```bash
curl -X POST http://localhost:8080/refunds/bulk/auto-approve \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event_id": 1,
    "max_refund_amount": 50.00,
    "days_before_event": 7
  }'
```

### Export Tickets to CSV
```bash
curl -X POST http://localhost:8080/tickets/bulk/export \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event_id": 1, "format": "csv"}' \
  --output tickets.csv
```

### Get Ticket Stats
```bash
curl http://localhost:8080/tickets/bulk/stats?event_id=1 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Common Filters

### Email Filters
```json
{
  "has_arrived": false,
  "is_refunded": false,
  "ticket_class_ids": [1, 2, 3]
}
```

### Ticket Filters
```json
{
  "status": "active",
  "is_checked_in": true,
  "is_refunded": false,
  "date_from": "2024-01-01",
  "date_to": "2024-12-31"
}
```

### Refund Criteria
```json
{
  "max_refund_amount": 100.00,
  "days_before_event": 14
}
```

---

## Response Patterns

### Bulk Email Response
```json
{
  "total_sent": 150,
  "total_failed": 2,
  "failed_emails": ["bad@email.com"],
  "message": "Successfully sent 150 emails, 2 failed"
}
```

### Bulk Refund Response
```json
{
  "total_processed": 10,
  "total_succeeded": 8,
  "total_failed": 2,
  "results": [
    {
      "refund_id": 1,
      "order_id": 123,
      "status": "success",
      "refund_amount": 50.00
    }
  ],
  "message": "Processed 10 refunds: 8 succeeded, 2 failed"
}
```

### Ticket Stats Response
```json
{
  "event_id": 1,
  "total_tickets": 500,
  "active_tickets": 300,
  "checked_in_tickets": 180,
  "total_revenue": 25000.00,
  "tickets_by_class": [...]
}
```

---

## Error Codes

- `400` - Invalid request (missing fields, bad format)
- `401` - Unauthorized (invalid/missing token)
- `403` - Forbidden (not event owner)
- `404` - Not found (event/resource doesn't exist)
- `500` - Server error

---

## Files

- `/internal/attendees/bulk.go` - Attendee bulk operations (536 lines)
- `/internal/refunds/bulk.go` - Refund bulk operations (392 lines)
- `/internal/tickets/bulk.go` - Ticket bulk operations (502 lines)
- Total: **1,430 lines of new code**

---

## Routes Added (12 new endpoints)

### Attendee Routes (3)
```go
router.HandleFunc("/attendees/bulk/email", handler.SendBulkEmail)
router.HandleFunc("/attendees/bulk/export", handler.ExportAttendeesData)
router.HandleFunc("/attendees/event/update-email", handler.SendEventUpdateEmail)
```

### Refund Routes (3)
```go
router.HandleFunc("/refunds/bulk/process", handler.ProcessBulkRefunds)
router.HandleFunc("/refunds/bulk/auto-approve", handler.AutoApproveBulkRefunds)
router.HandleFunc("/refunds/bulk/stats", handler.GetBulkRefundStats)
```

### Ticket Routes (3)
```go
router.HandleFunc("/tickets/bulk/export", handler.BulkExportTickets)
router.HandleFunc("/tickets/bulk/stats", handler.GetBulkTicketStats)
router.HandleFunc("/tickets/bulk/status", handler.BulkUpdateTicketStatus)
```

---

## Dependencies

- Existing notification service for emails
- RefundHandler for refund processing
- GORM for database queries
- Models: RefundRecord, Ticket, Attendee, Event, Order

---

## Testing

Run comprehensive tests:
```bash
./test-bulk-operations.sh
```

Test individual endpoints:
```bash
# Test bulk email
./test-bulk-operations.sh email

# Test bulk refunds
./test-bulk-operations.sh refunds

# Test bulk exports
./test-bulk-operations.sh exports
```
