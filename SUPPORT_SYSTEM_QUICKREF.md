# Support Ticket System - Quick Reference

## Overview
Complete support ticket system for customers and organizers to raise issues. Ready for AI-based classification and prioritization.

## Database Models

### SupportTicket
- **Ticket Number**: Auto-generated unique identifier (TKT-YYYYMMDD-XXXX)
- **Subject & Description**: Issue details
- **Category**: payment, booking, account, event, technical, refund, general, feature_request
- **Priority**: critical, high, medium, low
- **Status**: open, in_progress, resolved, closed
- **User Relations**: Links to User, Order, Event, Organizer
- **Assignment**: Can be assigned to support staff
- **AI Fields**: ai_classified, ai_priority, ai_confidence_score, ai_reasoning (ready for AI integration)

### SupportTicketComment
- Comments/updates on tickets
- Can be internal (staff only) or public
- Tracks author information

## API Endpoints

### Public Endpoints (No auth required)
```
POST   /support/tickets              Create new support ticket
```

### Authenticated User Endpoints
```
GET    /support/tickets              List tickets (filtered by role)
GET    /support/tickets/{id}         Get ticket details
POST   /support/tickets/{id}/comments Add comment to ticket
```

### Admin/Support Staff Only
```
PUT    /support/tickets/{id}         Update ticket (status, priority, assignment)
GET    /support/tickets/stats        Get ticket statistics
```

## Access Control

### Customers
- Can create tickets (with or without authentication)
- Can view their own tickets
- Can comment on their tickets

### Organizers
- Can create tickets
- Can view tickets related to their events
- Can view their own tickets
- Can comment on tickets they have access to

### Support Staff (Role: "support")
- Can view all tickets
- Can update ticket status, priority, assignment
- Can add internal notes
- Can view statistics

### Admins
- Full access to all support operations

## Creating a Ticket

### Request Example
```json
POST /support/tickets
{
  "subject": "Payment not received after purchase",
  "description": "I completed payment for order #1234 but haven't received tickets",
  "category": "payment",
  "email": "customer@example.com",
  "name": "John Doe",
  "phone_number": "+254700000000",
  "order_id": 1234,
  "event_id": 56
}
```

### Response
```json
{
  "message": "Support ticket created successfully",
  "ticket": {
    "id": 1,
    "ticket_number": "TKT-20260107-0001",
    "subject": "Payment not received after purchase",
    "description": "...",
    "category": "payment",
    "priority": "medium",
    "status": "open",
    "email": "customer@example.com",
    "name": "John Doe",
    "created_at": "2026-01-07T10:30:00Z"
  }
}
```

## Listing Tickets

### Query Parameters
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 20, max: 100)
- `status`: Filter by status (open, in_progress, resolved, closed)
- `priority`: Filter by priority (critical, high, medium, low)
- `category`: Filter by category
- `search`: Search in ticket number, subject, description, email

### Example
```bash
GET /support/tickets?status=open&priority=high&page=1&per_page=20
```

### Response
```json
{
  "tickets": [...],
  "total": 150,
  "page": 1,
  "per_page": 20,
  "total_pages": 8
}
```

## Updating a Ticket (Admin/Support Only)

```json
PUT /support/tickets/{id}
{
  "status": "in_progress",
  "priority": "high",
  "assigned_to_id": 5,
  "resolution_notes": "Investigating payment gateway logs"
}
```

## Adding Comments

```json
POST /support/tickets/{id}/comments
{
  "comment": "We've identified the issue and are working on a fix",
  "is_internal": false
}
```

## Ticket Statistics (Admin/Support Only)

```bash
GET /support/tickets/stats
```

Returns:
- Total tickets
- Breakdown by status
- Breakdown by priority
- Breakdown by category
- Average resolution time (in hours)

## AI Integration Fields

The ticket model includes fields ready for AI classification:
- `ai_classified`: Boolean flag
- `ai_priority`: AI-suggested priority
- `ai_confidence_score`: Confidence level (0-1)
- `ai_reasoning`: Explanation for classification

These fields can be populated by an AI service during ticket creation or as a background job.

## User Roles

Added new role: **"support"** for support staff members
- Has access to all tickets
- Can update and manage tickets
- Can view statistics
- Cannot perform admin operations (user management, etc.)

## Next Steps for AI Integration

1. **Create AI Classifier Service** (`internal/ai/classifier.go`)
   - Integrate OpenAI API or similar
   - Define classification prompt
   - Handle rate limiting and errors

2. **Add Background Worker**
   - Process tickets asynchronously
   - Update AI fields after classification
   - Retry failed classifications

3. **Add Feedback Loop**
   - Track when staff override AI priority
   - Log corrections for model improvement
   - Generate accuracy metrics

4. **Email Notifications**
   - New ticket created → Support team
   - Status updated → Customer
   - Comment added → Relevant parties
   - High priority detected → Immediate alert

## Database Migration

Models are automatically migrated on server startup:
- `support_tickets` table
- `support_ticket_comments` table

## Testing the System

### Create a test ticket (as guest)
```bash
curl -X POST http://localhost:8080/support/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Test ticket",
    "description": "Testing support system",
    "category": "technical",
    "email": "test@example.com",
    "name": "Test User"
  }'
```

### List tickets (requires auth)
```bash
curl http://localhost:8080/support/tickets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update ticket (admin/support only)
```bash
curl -X PUT http://localhost:8080/support/tickets/1 \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "priority": "high"
  }'
```

## Implementation Summary

✅ **Completed:**
- Database models for tickets and comments
- Full CRUD operations
- Role-based access control
- Filtering and search
- Comment system
- Statistics endpoint
- AI-ready fields
- Automatic ticket number generation

🔜 **Ready for AI Layer:**
- AI classifier service
- Background processing
- Priority prediction
- Category classification
- Automated responses
- Smart routing

---

**Files Created/Modified:**
- `internal/models/support_tickets.go` - Database models
- `internal/support/handler.go` - API handlers and business logic
- `internal/models/user.go` - Added RoleSupport
- `cmd/api-server/main.go` - Routes and initialization
