# API Routes Documentation

Complete list of all available API endpoints in the ticketing system.

## Table of Contents

- [Authentication](#authentication)
- [Organizers](#organizers)
- [Events](#events)
- [Accounts](#accounts)
- [Orders](#orders)
- [Tickets](#tickets)
- [Promotions](#promotions)
- [Inventory](#inventory)
- [Payments](#payments)
- [Refunds](#refunds)
- [Settlements](#settlements)
- [Attendees](#attendees)
- [Venues](#venues)
- [Notifications](#notifications)
- [Monitoring](#monitoring)

---

## Authentication

### User Registration & Login
- `POST /register` - Register a new user
- `POST /login` - User login
- `POST /logout` - User logout
- `POST /forgot-password` - Request password reset
- `POST /resetPassword` - Reset password with token

---

## Organizers

### Organizer Management
- `POST /organizers/apply` - Apply to become an organizer
- `GET /organizers/profile` - Get organizer profile
- `PUT /organizers/profile` - Update organizer profile
- `GET /organizers/onboarding/status` - Get onboarding status
- `GET /organizers/dashboard` - Get organizer dashboard
- `GET /organizers/dashboard/stats` - Get quick statistics
- `POST /organizers/logo` - Upload organizer logo
- `POST /organizers/verification/email` - Send verification email

### Admin - Organizer Verification
- `GET /admin/organizers/pending` - List pending organizer applications
- `POST /admin/organizers/{id}/verify` - Verify an organizer

---

## Events

### Public Event Routes
- `GET /events` - List all public events
- `GET /events/{id}` - Get event details
- `GET /events/{id}/images` - Get event images

### Organizer Event Management
- `GET /organizers/events` - List organizer's events
- `POST /organizers/events` - Create new event
- `PUT /organizers/events/{id}` - Update event
- `DELETE /organizers/events/{id}` - Delete event
- `POST /organizers/events/{id}/publish` - Publish event
- `POST /organizers/events/{id}/images` - Upload event image
- `DELETE /organizers/events/{id}/images/{imageId}` - Delete event image

---

## Accounts

### Profile Management
- `GET /account/profile` - Get account profile
- `PUT /account/profile` - Update account profile
- `DELETE /account` - Delete account

### Address Management
- `GET /account/address` - Get account address
- `PUT /account/address` - Update account address
- `DELETE /account/address` - Clear account address
- `GET /account/countries` - Get supported countries

### Settings & Preferences
- `GET /account/preferences` - Get account preferences
- `PUT /account/preferences` - Update account preferences
- `GET /account/settings` - Get account settings
- `PUT /account/settings` - Update account settings
- `GET /account/timezones` - Get available timezones
- `GET /account/currencies` - Get available currencies
- `GET /account/date-formats` - Get date format options

### Security
- `GET /account/security` - Get security settings
- `POST /account/security/password` - Change password
- `GET /account/security/login-history` - Get login history
- `POST /account/security/lock` - Lock account
- `POST /account/security/unlock` - Unlock account

### Activity & Stats
- `GET /account/activity` - Get account activity log
- `GET /account/activity/types` - Get activity types
- `POST /account/activity/log` - Log activity
- `DELETE /account/activity/clear` - Clear activity log
- `GET /account/stats` - Get account statistics

### Payment Methods
- `GET /account/payment-methods` - Get saved payment methods
- `GET /account/payment-gateway` - Get payment gateway settings
- `GET /account/payment-gateway/info` - Get payment gateway info

### Stripe Integration (Organizers)
- `POST /account/stripe/setup` - Setup Stripe integration
- `POST /account/stripe/connect` - Setup Stripe Connect
- `POST /account/stripe/complete` - Complete Stripe setup
- `DELETE /account/stripe/disconnect` - Disconnect Stripe

---

## Orders

### Order Creation & Calculation
- `POST /orders` - Create new order
- `POST /orders/calculate` - Calculate order total

### Order Viewing
- `GET /orders` - List user's orders
- `GET /orders/{id}` - Get order details
- `GET /orders/{id}/summary` - Get order summary
- `GET /orders/stats` - Get order statistics

### Order Management
- `PUT /orders/{id}/status` - Update order status
- `POST /orders/{id}/cancel` - Cancel order
- `POST /orders/{id}/refund` - Request refund

### Payment Processing
- `POST /orders/{id}/payment` - Process payment
- `POST /orders/{id}/payment/verify` - Verify payment

### Organizer View
- `GET /organizers/orders` - List organizer's orders

---

## Tickets

### Ticket Generation
- `POST /tickets/generate` - Generate tickets for an order
- `POST /tickets/regenerate-qr` - Regenerate ticket QR code

### Ticket Viewing
- `GET /tickets` - List user's tickets
- `GET /tickets/{id}` - Get ticket details
- `GET /tickets/number` - Get ticket by ticket number
- `GET /tickets/stats` - Get ticket statistics

### PDF Download
- `GET /tickets/{id}/pdf` - Download ticket as PDF ✨

### Ticket Transfer
- `POST /tickets/{id}/transfer` - Transfer ticket to another person
- `GET /tickets/{id}/transfer-history` - Get transfer history

### Validation (Organizer)
- `POST /tickets/validate` - Validate a ticket
- `POST /tickets/validate/qr` - Validate ticket by QR code

### Check-in (Organizer)
- `POST /tickets/checkin` - Check in a ticket
- `POST /tickets/checkin/bulk` - Bulk check-in tickets
- `POST /tickets/checkin/undo` - Undo check-in
- `GET /tickets/checkin/stats` - Get check-in statistics

### Event Tickets (Organizer)
- `GET /organizers/tickets` - List tickets for organizer's events

---

## Promotions

### Promotion Management
- `POST /promotions` - Create promotion
- `GET /promotions/{id}` - Get promotion details
- `GET /promotions/code/{code}` - Get promotion by code
- `PUT /promotions/{id}` - Update promotion
- `DELETE /promotions/{id}` - Delete promotion
- `POST /promotions/{id}/clone` - Clone promotion

### Status Management
- `POST /promotions/{id}/activate` - Activate promotion
- `POST /promotions/{id}/pause` - Pause promotion
- `POST /promotions/{id}/deactivate` - Deactivate promotion
- `POST /promotions/{id}/extend` - Extend promotion date

### Listing & Search
- `GET /promotions` - List promotions
- `GET /promotions/active` - List active promotions
- `GET /promotions/search` - Search promotions

### Validation & Usage
- `POST /promotions/validate` - Validate promotion code
- `POST /promotions/eligibility` - Check promotion eligibility
- `GET /promotions/{id}/usage` - Get promotion usage
- `POST /promotions/{id}/usage` - Record promotion usage
- `POST /promotions/usage/revoke` - Revoke promotion usage
- `GET /promotions/{id}/usage/details` - Get usage details

### Analytics
- `GET /promotions/{id}/stats` - Get promotion statistics
- `GET /promotions/{id}/analytics` - Get promotion analytics
- `GET /promotions/{id}/roi` - Get ROI metrics
- `GET /promotions/{id}/conversion` - Get conversion metrics
- `GET /promotions/{id}/revenue` - Get revenue impact

### Organizer View
- `GET /organizers/promotions` - List organizer's promotions
- `GET /organizers/promotions/stats` - Get organizer promotion stats

---

## Inventory

### Availability
- `GET /inventory/tickets/{id}` - Get ticket availability
- `GET /inventory/events/{id}` - Get event inventory
- `GET /inventory/status/{id}` - Get inventory status
- `POST /inventory/bulk-check` - Bulk check availability

### Reservations
- `POST /inventory/reservations` - Create reservation
- `GET /inventory/reservations/{id}` - Get reservation details
- `GET /inventory/reservations` - List user's reservations
- `GET /inventory/reservations/{id}/validate` - Validate reservation
- `POST /inventory/reservations/{id}/extend` - Extend reservation

### Release
- `DELETE /inventory/reservations/{id}/release` - Release reservation
- `POST /inventory/reservations/expired` - Release expired reservations
- `POST /inventory/reservations/convert` - Convert reservation to order
- `DELETE /inventory/reservations/session` - Release session reservations
- `GET /inventory/events/{id}/reservations` - Get event reservations

---

## Payments

### Payment Processing
- `POST /payments/initiate` - Initiate payment
- `GET /payments/verify/{id}` - Verify payment
- `GET /payments/orders/{id}/status` - Get payment status
- `GET /payments/history` - Get payment history

### Payment Methods
- `POST /payments/methods` - Save payment method
- `GET /payments/methods` - Get saved payment methods
- `DELETE /payments/methods/{id}` - Delete payment method
- `POST /payments/methods/{id}/default` - Set default payment method
- `PUT /payments/methods/{id}/expiry` - Update payment method expiry

### Refunds
- `POST /payments/refunds` - Initiate refund
- `GET /payments/refunds/{id}/status` - Get refund status
- `GET /payments/refunds` - List refunds
- `POST /payments/refunds/{id}/approve` - Approve refund

### Webhooks
- `POST /webhooks/intasend` - Intasend webhook handler
- `GET /webhooks/logs` - Get webhook logs
- `POST /webhooks/logs/{id}/retry` - Retry failed webhook

### Gateways
- `GET /payments/gateways` - Get available payment gateways

---

## Refunds

### Customer Refunds
- `POST /refunds` - Request refund
- `GET /refunds` - List user's refunds
- `GET /refunds/{id}` - Get refund status
- `POST /refunds/{id}/cancel` - Cancel refund request

### Admin/Organizer Refunds
- `GET /admin/refunds/pending` - List pending refunds
- `GET /admin/refunds/{id}` - Get refund details
- `POST /admin/refunds/{id}/approve` - Approve refund
- `POST /admin/refunds/{id}/process` - Process refund
- `POST /admin/refunds/{id}/retry` - Retry failed refund
- `GET /admin/refunds/statistics` - Get refund statistics

### Organizer View
- `GET /organizers/refunds` - List organizer's refunds

---

## Settlements

### Calculation & Preview
- `GET /settlements/calculate/event/{id}` - Calculate event settlement
- `GET /settlements/preview` - Get settlement preview
- `GET /settlements/eligibility/event/{id}` - Validate settlement eligibility

### Batch Creation & Processing
- `POST /settlements/batch` - Create settlement batch
- `GET /settlements/{id}` - Get settlement details
- `GET /settlements` - List settlements
- `POST /settlements/{id}/approve` - Approve settlement
- `POST /settlements/{id}/process` - Process settlement
- `POST /settlements/{id}/cancel` - Cancel settlement
- `POST /settlements/{id}/withhold` - Withhold settlement

### Reports & Analytics
- `GET /settlements/{id}/report` - Generate settlement report
- `GET /settlements/summary/organizer/{id}` - Get organizer summary
- `GET /settlements/summary/platform` - Get platform summary
- `GET /settlements/export` - Export settlements
- `GET /settlements/history/organizer/{id}` - Get settlement history

### Status & Management
- `GET /settlements/pending` - Get pending settlements
- `GET /settlements/failed` - Get failed settlements
- `POST /settlements/{id}/retry` - Retry failed settlement
- `POST /settlements/items/{id}/complete` - Complete settlement item
- `POST /settlements/items/{id}/fail` - Fail settlement item

### Organizer View
- `GET /organizers/settlements` - List organizer's settlements
- `GET /organizers/settlements/summary` - Get organizer summary

### Webhooks
- `POST /webhooks/settlements/complete` - Settlement webhook handler

---

## Attendees

### Listing & Search
- `GET /attendees` - List attendees
- `GET /attendees/search` - Search attendees
- `GET /attendees/count` - Get attendee count
- `GET /attendees/{id}` - Get attendee details
- `GET /attendees/ticket` - Get attendee by ticket
- `GET /attendees/order/{id}` - Get attendees by order

### Check-in Management
- `POST /attendees/checkin` - Check in attendee
- `POST /attendees/checkin/bulk` - Bulk check-in
- `POST /attendees/checkin/undo` - Undo check-in

### Update & Management
- `PUT /attendees/{id}` - Update attendee info
- `POST /attendees/{id}/no-show` - Mark as no-show
- `POST /attendees/{id}/transfer` - Transfer attendee

### Export & Reports
- `GET /attendees/export` - Export attendee list
- `GET /attendees/badges` - Export badge data

### Analytics
- `GET /attendees/stats` - Get attendance statistics
- `GET /attendees/report/checkin` - Get check-in report
- `GET /attendees/timeline` - Get attendance timeline
- `GET /attendees/no-shows` - Get no-show list

### Organizer View
- `GET /organizers/attendees` - List event attendees

---

## Venues

### CRUD Operations
- `POST /venues` - Create venue
- `GET /venues` - List venues
- `GET /venues/{id}` - Get venue details
- `PUT /venues/{id}` - Update venue
- `DELETE /venues/{id}` - Delete venue

### Search & Discovery
- `GET /venues/search/location` - Search venues by location
- `GET /venues/type` - Get venues by type

### Statistics & Information
- `GET /venues/{id}/stats` - Get venue statistics
- `GET /venues/{id}/events` - Get venue events

### Availability Management
- `GET /venues/{id}/availability` - Check venue availability
- `GET /venues/{id}/calendar` - Get venue calendar
- `GET /venues/available` - Find available venues

### Advanced Operations
- `POST /venues/{id}/restore` - Restore deleted venue
- `DELETE /venues/{id}/permanent` - Permanently delete venue

---

## Notifications

### Email Notifications ✨
- `POST /notifications/test` - Test email configuration
- `POST /notifications/welcome` - Send welcome email
- `POST /notifications/verification` - Send verification email
- `POST /notifications/password-reset` - Send password reset email

**Request Body Examples:**

**Test Email:**
```json
{
  "email": "user@example.com"
}
```

**Welcome Email:**
```json
{
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Verification Email:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "code": "123456"
}
```

**Password Reset:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "token": "reset-token-123"
}
```

---

## Monitoring

### Metrics
- `GET /metrics` - Prometheus metrics endpoint

---

## Route Summary

| Module | Routes Count |
|--------|--------------|
| Authentication | 5 |
| Organizers | 10 |
| Events | 9 |
| Accounts | 29 |
| Orders | 14 |
| Tickets | 19 |
| Promotions | 26 |
| Inventory | 15 |
| Payments | 19 |
| Refunds | 11 |
| Settlements | 26 |
| Attendees | 19 |
| Venues | 14 |
| Notifications | 4 ✨ |
| Monitoring | 1 |
| **Total** | **221** |

---

## Authentication & Authorization

Most endpoints require authentication via Bearer token:

```bash
Authorization: Bearer YOUR_JWT_TOKEN
```

### Public Endpoints (No Auth Required)
- `POST /register`
- `POST /login`
- `GET /events`
- `GET /events/{id}`
- `GET /events/{id}/images`
- `GET /venues`
- `GET /venues/{id}`

### Organizer-Only Endpoints
All endpoints under `/organizers/*` and `/admin/*` require organizer role.

---

## Common Response Formats

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

### Paginated Response
```json
{
  "data": [...],
  "total_count": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5
}
```

---

## New Features ✨

### Ticket PDF Generation
- **Route**: `GET /tickets/{id}/pdf`
- **Feature**: Download professional PDF tickets with QR codes
- **Authentication**: Required
- **Authorization**: Must own the ticket
- **Response**: PDF file download

### Email Notifications
- **Routes**: `/notifications/*`
- **Feature**: Send transactional emails via SMTP
- **Providers Supported**: Gmail, Outlook, Zoho, Mailgun, SendGrid, etc.
- **Configuration**: Via environment variables

---

## Environment Configuration

### Email Service (Notifications)
```env
# Email Configuration
EMAIL_PROVIDER=gmail
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM_ADDRESS=your-email@gmail.com
EMAIL_FROM_NAME=Your App Name
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false
```

### Database
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=ticketing
```

### Server
```env
SERVER_PORT=8080
JWT_SECRET=your-secret-key
```

---

## Testing Routes

### Using cURL

**Test Email Configuration:**
```bash
curl -X POST http://localhost:8080/notifications/test \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

**Download Ticket PDF:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/tickets/123/pdf \
  -o ticket.pdf
```

**Create Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": 1,
    "ticket_classes": [{"id": 1, "quantity": 2}]
  }'
```

---

## API Documentation Tools

For interactive API documentation, consider integrating:
- **Swagger/OpenAPI**: Generate interactive API docs
- **Postman Collection**: Import routes for testing
- **Insomnia**: REST client with collections

---

## Notes

- All timestamps are in UTC
- All monetary values are in cents/smallest currency unit
- File uploads use multipart/form-data
- JSON responses use snake_case for field names
- Date format: ISO 8601 (YYYY-MM-DDTHH:mm:ssZ)

---

**Version**: 1.0  
**Last Updated**: November 2024  
**Total Routes**: 221 ✅
