# Bulk Operations Implementation Guide

## Overview

This document describes the **Bulk Operations** module for the ticketing system, providing efficient ways to manage multiple attendees, refunds, and tickets simultaneously.

---

## Table of Contents

1. [Bulk Email to Attendees](#1-bulk-email-to-attendees)
2. [Bulk Refund Processing](#2-bulk-refund-processing)
3. [Bulk Ticket Exports](#3-bulk-ticket-exports)
4. [API Reference](#4-api-reference)
5. [Code Examples](#5-code-examples)
6. [Testing Guide](#6-testing-guide)

---

## 1. Bulk Email to Attendees

### Features
- Send custom emails to multiple attendees at once
- Filter attendees by event, arrival status, or ticket class
- Support both plain text and HTML email formats
- Track success/failure rates for each bulk operation
- Send event update notifications to all attendees

### Use Cases
- Event updates and announcements
- Important information to specific attendee groups
- Pre-event reminders to those who haven't arrived
- Post-event thank you messages

### Implementation Details

**File:** `/internal/attendees/bulk.go`

**Key Functions:**
- `SendBulkEmail()` - Send custom emails to filtered attendees
- `SendEventUpdateEmail()` - Quick event update to all attendees
- `ExportAttendeesData()` - Export attendee data in CSV or JSON format

**Request Structure:**
```go
type BulkEmailRequest struct {
    EventID     uint                 `json:"event_id"`
    AttendeeIDs []uint               `json:"attendee_ids,omitempty"`
    Subject     string               `json:"subject"`
    Message     string               `json:"message"`
    HTMLMessage string               `json:"html_message,omitempty"`
    Filters     *BulkEmailFilters    `json:"filters,omitempty"`
}

type BulkEmailFilters struct {
    HasArrived     *bool  `json:"has_arrived,omitempty"`
    IsRefunded     *bool  `json:"is_refunded,omitempty"`
    TicketClassIDs []uint `json:"ticket_class_ids,omitempty"`
}
```

**Response Structure:**
```go
type BulkEmailResponse struct {
    TotalSent    int      `json:"total_sent"`
    TotalFailed  int      `json:"total_failed"`
    FailedEmails []string `json:"failed_emails,omitempty"`
    Message      string   `json:"message"`
}
```

---

## 2. Bulk Refund Processing

### Features
- Process multiple refunds simultaneously
- Bulk approve or reject pending refunds
- Auto-approve eligible refunds based on criteria
- Get refund statistics for events
- Track processing results for each refund

### Use Cases
- Event cancellation requiring mass refunds
- Batch approval of small-amount refunds
- End-of-day refund processing
- Emergency refund situations

### Implementation Details

**File:** `/internal/refunds/bulk.go`

**Key Functions:**
- `ProcessBulkRefunds()` - Process multiple refunds (approve/reject)
- `AutoApproveBulkRefunds()` - Auto-approve eligible refunds
- `GetBulkRefundStats()` - Get refund statistics for an event

**Request Structure:**
```go
type BulkRefundRequest struct {
    RefundIDs []uint `json:"refund_ids"`
    Action    string `json:"action"` // "approve" or "reject"
    Reason    string `json:"reason,omitempty"`
}

type BulkAutoApproveRequest struct {
    EventID         uint    `json:"event_id"`
    MaxRefundAmount float64 `json:"max_refund_amount,omitempty"`
    DaysBeforeEvent int     `json:"days_before_event,omitempty"`
}
```

**Response Structure:**
```go
type BulkRefundResponse struct {
    TotalProcessed int                `json:"total_processed"`
    TotalSucceeded int                `json:"total_succeeded"`
    TotalFailed    int                `json:"total_failed"`
    Results        []RefundBulkResult `json:"results"`
    Message        string             `json:"message"`
}

type RefundBulkResult struct {
    RefundID     uint    `json:"refund_id"`
    OrderID      uint    `json:"order_id"`
    Status       string  `json:"status"` // "success" or "failed"
    Error        string  `json:"error,omitempty"`
    RefundAmount float64 `json:"refund_amount,omitempty"`
}
```

**Statistics Structure:**
```go
type BulkRefundStats struct {
    EventID             uint    `json:"event_id"`
    TotalRefunds        int     `json:"total_refunds"`
    PendingRefunds      int     `json:"pending_refunds"`
    ApprovedRefunds     int     `json:"approved_refunds"`
    RejectedRefunds     int     `json:"rejected_refunds"`
    CompletedRefunds    int     `json:"completed_refunds"`
    TotalRefundAmount   float64 `json:"total_refund_amount"`
    PendingRefundAmount float64 `json:"pending_refund_amount"`
}
```

---

## 3. Bulk Ticket Exports

### Features
- Export tickets to CSV or JSON formats
- Filter tickets by status, class, check-in state, or date
- Get comprehensive ticket statistics
- Bulk update ticket statuses
- Include QR codes and attendee information

### Use Cases
- Event attendance reports
- Financial reconciliation
- Ticket distribution analysis
- Compliance and audit requirements
- Integration with external systems

### Implementation Details

**File:** `/internal/tickets/bulk.go`

**Key Functions:**
- `BulkExportTickets()` - Export tickets in CSV or JSON
- `GetBulkTicketStats()` - Get detailed ticket statistics
- `BulkUpdateTicketStatus()` - Update status of multiple tickets

**Request Structure:**
```go
type BulkExportRequest struct {
    EventID   uint                  `json:"event_id"`
    TicketIDs []uint                `json:"ticket_ids,omitempty"`
    Format    string                `json:"format"` // "csv" or "json"
    IncludeQR bool                  `json:"include_qr,omitempty"`
    Filters   *TicketExportFilters  `json:"filters,omitempty"`
}

type TicketExportFilters struct {
    Status         string `json:"status,omitempty"`
    TicketClassIDs []uint `json:"ticket_class_ids,omitempty"`
    IsCheckedIn    *bool  `json:"is_checked_in,omitempty"`
    IsTransferred  *bool  `json:"is_transferred,omitempty"`
    IsRefunded     *bool  `json:"is_refunded,omitempty"`
    DateFrom       string `json:"date_from,omitempty"` // YYYY-MM-DD
    DateTo         string `json:"date_to,omitempty"`   // YYYY-MM-DD
}
```

**CSV Export Columns:**
- Ticket ID, Ticket Number, Ticket Class, Price
- Status, Owner Email, Owner Name
- Attendee Name, Attendee Email
- Is Checked In, Check-in Time
- Is Transferred, Is Refunded
- QR Code, Created At, Updated At

**Statistics Structure:**
```go
type BulkTicketStats struct {
    EventID            uint               `json:"event_id"`
    TotalTickets       int                `json:"total_tickets"`
    ActiveTickets      int                `json:"active_tickets"`
    UsedTickets        int                `json:"used_tickets"`
    TransferredTickets int                `json:"transferred_tickets"`
    RefundedTickets    int                `json:"refunded_tickets"`
    CheckedInTickets   int                `json:"checked_in_tickets"`
    TotalRevenue       float64            `json:"total_revenue"`
    TicketsByClass     []TicketClassStats `json:"tickets_by_class"`
}

type TicketClassStats struct {
    TicketClassID   uint    `json:"ticket_class_id"`
    TicketClassName string  `json:"ticket_class_name"`
    TotalSold       int     `json:"total_sold"`
    CheckedIn       int     `json:"checked_in"`
    Revenue         float64 `json:"revenue"`
}
```

---

## 4. API Reference

### Attendee Bulk Operations

#### Send Bulk Email
```http
POST /attendees/bulk/email
Authorization: Bearer {token}
Content-Type: application/json

{
  "event_id": 1,
  "subject": "Event Update",
  "message": "Important update about your event...",
  "filters": {
    "has_arrived": false
  }
}
```

#### Export Attendees Data
```http
POST /attendees/bulk/export
Authorization: Bearer {token}
Content-Type: application/json

{
  "event_id": 1,
  "format": "csv",
  "filters": {
    "has_arrived": true
  }
}
```

#### Send Event Update Email
```http
POST /attendees/event/update-email?event_id=1
Authorization: Bearer {token}
Content-Type: application/json

{
  "subject": "Event Time Change",
  "message": "The event start time has been changed...",
  "only_non_arrived": true
}
```

---

### Refund Bulk Operations

#### Process Bulk Refunds
```http
POST /refunds/bulk/process
Authorization: Bearer {token}
Content-Type: application/json

{
  "refund_ids": [1, 2, 3],
  "action": "approve",
  "reason": "Batch approval"
}
```

#### Auto-Approve Refunds
```http
POST /refunds/bulk/auto-approve
Authorization: Bearer {token}
Content-Type: application/json

{
  "event_id": 1,
  "max_refund_amount": 100.00,
  "days_before_event": 7
}
```

#### Get Refund Statistics
```http
GET /refunds/bulk/stats?event_id=1
Authorization: Bearer {token}
```

---

### Ticket Bulk Operations

#### Export Tickets
```http
POST /tickets/bulk/export
Authorization: Bearer {token}
Content-Type: application/json

{
  "event_id": 1,
  "format": "csv",
  "filters": {
    "is_checked_in": true,
    "date_from": "2024-01-01"
  }
}
```

#### Get Ticket Statistics
```http
GET /tickets/bulk/stats?event_id=1
Authorization: Bearer {token}
```

#### Bulk Update Ticket Status
```http
POST /tickets/bulk/status
Authorization: Bearer {token}
Content-Type: application/json

{
  "ticket_ids": [1, 2, 3],
  "status": "cancelled"
}
```

---

## 5. Code Examples

### Example 1: Send Email to Non-Arrived Attendees

```bash
curl -X POST http://localhost:8080/attendees/bulk/email \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": 1,
    "subject": "Reminder: Event Tomorrow",
    "message": "This is a friendly reminder that the event is tomorrow at 6 PM.",
    "filters": {
      "has_arrived": false
    }
  }'
```

**Expected Response:**
```json
{
  "total_sent": 150,
  "total_failed": 2,
  "failed_emails": ["invalid@email.com", "bounced@email.com"],
  "message": "Successfully sent 150 emails, 2 failed"
}
```

---

### Example 2: Auto-Approve Small Refunds

```bash
curl -X POST http://localhost:8080/refunds/bulk/auto-approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": 1,
    "max_refund_amount": 50.00,
    "days_before_event": 14
  }'
```

**Expected Response:**
```json
{
  "total_processed": 25,
  "total_succeeded": 23,
  "total_failed": 2,
  "results": [
    {
      "refund_id": 1,
      "order_id": 123,
      "status": "success",
      "refund_amount": 25.00
    },
    {
      "refund_id": 2,
      "order_id": 124,
      "status": "failed",
      "error": "refund status is 'completed', not pending"
    }
  ],
  "message": "Auto-approved 23 refunds, 2 failed"
}
```

---

### Example 3: Export Tickets to CSV

```bash
curl -X POST http://localhost:8080/tickets/bulk/export \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": 1,
    "format": "csv",
    "filters": {
      "is_checked_in": true,
      "ticket_class_ids": [1, 2]
    }
  }' \
  --output tickets_export.csv
```

**CSV Output:**
```csv
Ticket ID,Ticket Number,Ticket Class,Price,Status,Owner Email,...
1,TKT-2024-001,VIP,100.00,used,john@example.com,...
2,TKT-2024-002,Regular,50.00,used,jane@example.com,...
```

---

### Example 4: Get Ticket Statistics

```bash
curl -X GET http://localhost:8080/tickets/bulk/stats?event_id=1 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "event_id": 1,
  "total_tickets": 500,
  "active_tickets": 300,
  "used_tickets": 180,
  "transferred_tickets": 10,
  "refunded_tickets": 10,
  "checked_in_tickets": 180,
  "total_revenue": 25000.00,
  "tickets_by_class": [
    {
      "ticket_class_id": 1,
      "ticket_class_name": "VIP",
      "total_sold": 100,
      "checked_in": 85,
      "revenue": 10000.00
    },
    {
      "ticket_class_id": 2,
      "ticket_class_name": "Regular",
      "total_sold": 400,
      "checked_in": 95,
      "revenue": 15000.00
    }
  ]
}
```

---

## 6. Testing Guide

### Test Scenarios

#### Bulk Email Tests
1. **Send to all attendees** - Verify all attendees receive the email
2. **Send to filtered attendees** - Test has_arrived filter
3. **Send to specific ticket classes** - Test ticket_class_ids filter
4. **Handle email failures** - Verify failed emails are tracked
5. **HTML vs Plain Text** - Test both email formats

#### Bulk Refund Tests
1. **Approve multiple refunds** - Verify all refunds are approved
2. **Reject multiple refunds** - Verify rejection with reason
3. **Auto-approve with amount limit** - Test max_refund_amount filter
4. **Auto-approve with date limit** - Test days_before_event filter
5. **Mixed results handling** - Verify some succeed, some fail
6. **Statistics accuracy** - Verify refund counts and amounts

#### Bulk Export Tests
1. **CSV export** - Verify correct format and data
2. **JSON export** - Verify correct structure
3. **Filter by status** - Test status filter
4. **Filter by check-in** - Test is_checked_in filter
5. **Filter by date range** - Test date_from and date_to
6. **Statistics calculation** - Verify all stats are correct

### Test Script

See `/test-bulk-operations.sh` for automated testing.

---

## Security Considerations

1. **Authorization**: All bulk operations require valid JWT tokens
2. **Ownership Verification**: Operations only work on events owned by the authenticated user
3. **Rate Limiting**: Consider implementing rate limits for bulk operations
4. **Email Throttling**: Bulk emails should respect service provider limits
5. **Audit Logging**: Track all bulk operations for compliance

---

## Performance Notes

1. **Batch Processing**: Operations process items sequentially to avoid overwhelming services
2. **Database Optimization**: Use proper indexes on event_id, status fields
3. **Email Queuing**: Consider implementing queue for large email batches
4. **Export Pagination**: Large exports should be paginated or streamed
5. **Statistics Caching**: Consider caching frequently accessed statistics

---

## Error Handling

All bulk operations follow a consistent error handling pattern:

1. **Validation Errors**: Return 400 Bad Request with details
2. **Authentication Errors**: Return 401 Unauthorized
3. **Authorization Errors**: Return 403 Forbidden
4. **Not Found Errors**: Return 404 Not Found
5. **Server Errors**: Return 500 Internal Server Error

Bulk operations that process multiple items return detailed results showing success/failure for each item.

---

## Future Enhancements

1. **Async Processing**: Move bulk operations to background jobs
2. **Progress Tracking**: Real-time progress updates via WebSocket
3. **Scheduled Operations**: Schedule bulk operations for future execution
4. **Template Library**: Pre-built email templates for common scenarios
5. **Advanced Filtering**: More complex filtering options
6. **Export Formats**: Support for Excel, PDF exports
7. **Batch Limits**: Configurable limits per operation type

---

## Support

For issues or questions about bulk operations:
- Check the API documentation
- Review error responses for details
- Contact support with operation details and timestamps
