# Event Ticketing System — v1 Monolithic Architecture

This is the first version of the event ticketing platform. It is a single deployable Go binary that handles the entire system: authentication, event management, ticket purchasing, payment processing, PDF generation, email delivery, refunds, settlements, analytics, and admin operations — all within one HTTP server process.

This version establishes the full domain model and feature set that v2 and v3 build upon architecturally.

---

## Architecture

Everything runs in a single process. The API server initialises all service handlers at startup and wires them to routes. There are no background workers, no message queues, and no event bus — all operations complete synchronously within the HTTP request lifecycle.

```
Client
  │
  ▼
┌─────────────────────────────────────────────────────┐
│                    API Server (:8080)                │
│                                                     │
│  auth · events · orders · tickets · payments        │
│  refunds · settlements · inventory · promotions     │
│  analytics · notifications · support · admin        │
│  organizers · attendees · venues · accounts         │
│                                                     │
│  [reservation cleanup goroutine — runs in-process]  │
└───────────┬─────────────────┬───────────────────────┘
            │                 │
            ▼                 ▼
      PostgreSQL            Redis
                              │
                              ▼
                             S3
```

**Key characteristic:** a slow email provider, a large PDF generation job, or a payment gateway timeout directly affects the response time seen by the user making that request. This is the trade-off v2 solves with workers.

---

## What's in the Box

### Attendees
- Browse, search, and filter public events
- Reserve tickets with a timed hold (auto-expired by an in-process goroutine)
- Purchase tickets — Intasend (M-Pesa), offline payment
- Receive order confirmation and PDF ticket emails
- Download PDF tickets with embedded QR codes
- Transfer tickets to another attendee
- Request refunds and track their status
- Join waitlists for sold-out events
- Email verification required before ticket downloads and transfers
- Two-factor authentication (TOTP) with recovery codes

### Organizers
- Apply to become an organiser (KYC / verification flow)
- Create and publish events with venue assignment, images, and capacity
- Define multiple ticket classes with pricing, sale windows, and capacity limits
- Pause and resume individual ticket classes
- Real-time sales dashboard — revenue, tickets sold, check-in rate, capacity usage
- Manage orders, process and approve refund requests
- Cancel events (triggers refund cascade for all paid orders)
- Manage promotional codes — percentage and fixed discounts, usage limits, date windows
- View and export attendee lists, badge data, check-in reports
- Bank details management (AES-256 encrypted at rest) for payouts
- Settlement calculation, preview, batch processing, and export

### Admins
- Full user management — list, search, update roles and status
- Organiser verification and KYC approval
- Platform-wide analytics — sales volume, revenue, engagement metrics
- Support ticket management with AI-assisted classification
- Webhook log inspection and manual retry
- Bulk refund operations

---

## Project Structure

```
v1-monolithic-architecture/
├── cmd/
│   ├── api-server/
│   │   ├── main.go                 # Entry point — all routes, middleware, startup
│   │   └── ratelimit_integration.go
│   └── check-migration/
│       └── main.go                 # Utility to verify schema is up to date
│
├── internal/
│   ├── auth/                       # JWT auth, registration, login, password reset
│   ├── accounts/                   # Profile, address, preferences, security settings
│   ├── events/                     # Event CRUD, image upload, search, publishing
│   ├── ticketclasses/              # Ticket class CRUD, pause/resume
│   ├── orders/                     # Order creation, calculation, cancellation
│   ├── tickets/                    # Ticket generation, PDF, QR, check-in, transfer
│   ├── payments/                   # Intasend, Stripe, webhook handlers, refund initiation
│   ├── refunds/                    # Refund lifecycle, bulk operations, notifications
│   ├── settlement/                 # Organiser payout calculation and processing
│   ├── inventory/                  # Reservations, availability, waitlist
│   ├── promotions/                 # Promo codes, eligibility, usage tracking
│   ├── organizers/                 # Organiser onboarding, profile, bank details
│   ├── attendees/                  # Attendee management, bulk check-in, export
│   ├── venues/                     # Venue CRUD, availability, calendar
│   ├── analytics/                  # Prometheus metrics, sales/engagement reports
│   ├── notifications/              # Email templates, Brevo API, SMTP delivery
│   ├── support/                    # Support tickets, comments, AI classification
│   ├── admin/                      # Admin user management routes
│   ├── ai/                         # AI ticket classifier
│   ├── cache/                      # Redis session manager and events cache
│   ├── config/                     # Environment-based configuration loader
│   ├── database/                   # GORM connection and auto-migration
│   ├── middleware/                  # JWT auth, email verification, JSON error helpers
│   ├── models/                     # All GORM model definitions
│   ├── security/                   # AES-256 encryption service
│   ├── seed/                       # Reference data seeding (timezones, currencies)
│   └── storage/                    # AWS S3 client with local fallback
│
├── pkg/
│   ├── pdf/                        # PDF ticket generator (gofpdf)
│   ├── qrcode/                     # QR code generator
│   └── ratelimit/                  # Token bucket rate limiter
│
├── migrations/                     # SQL migration files
├── grafana/                        # Grafana dashboard config
├── prometheus/                     # Prometheus scrape config
├── docker-compose.monitoring.yml   # Prometheus + Grafana stack
└── .env                            # Environment variables (see below)
```

---

## Rate Limiting

Routes are protected by a token bucket rate limiter, grouped by sensitivity:

| Limiter | Limit | Applied to |
|---|---|---|
| `login` | 5 req / min per IP | Login, 2FA verify |
| `auth` | 10 req / min per IP | Register, forgot password, 2FA setup |
| `payment` | 5 req / min per IP | Payment initiation, order status updates, refunds |
| `download` | 3 req / sec per user | PDF ticket downloads |
| `inventory` | 50 req / sec per IP | Availability checks, reservations |
| `api` | 100 req / sec per IP | Everything else |

---

## Local Development

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis (optional — falls back to in-memory sessions if disabled)
- Docker (for the monitoring stack)

### 1. Start dependencies

```bash
# PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_DB=ticketing_system \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres:15-alpine

# Redis (optional)
docker run -d -p 6379:6379 redis:7-alpine
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your credentials
```

### 3. Run the server

```bash
go run cmd/api-server/main.go
```

The server auto-migrates all tables on startup and seeds reference data (timezones, currencies). No separate migration step is required.

### 4. Verify

```bash
curl http://localhost:8080/health
# {"status":"healthy","checks":{"database":{"healthy":true}},...}
```

---

## Docker

```bash
# Build
docker build -t ticketing-v1 .

# Run
docker run -p 8080:8080 --env-file .env ticketing-v1
```

---

## Monitoring

```bash
docker compose -f docker-compose.monitoring.yml up -d
```

- Prometheus: [http://localhost:9090](http://localhost:9090)
- Grafana: [http://localhost:3001](http://localhost:3001) — `admin / admin123`

Metrics are exposed at `GET /metrics`.

---

## Environment Variables

```bash
# ── Database ──────────────────────────────────────────
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ticketing_system
DB_SSL_MODE=disable

# ── Server ────────────────────────────────────────────
SERVER_PORT=8080
APP_ENV=development
APP_NAME=Ticketing System
APP_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000

# ── Auth ─────────────────────────────────────────────
JWTSECRET=your-jwt-secret

# ── Security ──────────────────────────────────────────
ENCRYPTION_KEY=              # exactly 16, 24, or 32 bytes (AES key for bank details)

# ── Redis (optional) ──────────────────────────────────
REDIS_ENABLED=false
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# ── AWS S3 ────────────────────────────────────────────
S3_ENABLED=true
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=us-east-1
S3_BUCKET=
S3_PUBLIC_URL=

# ── Email ─────────────────────────────────────────────
# Priority: Brevo API → SMTP → test mode (log only)
#
# Option 1: Brevo API (recommended for cloud deployments)
BREVO_API_KEY=               # set this and SMTP vars below are ignored

# Option 2: SMTP (works with Gmail, Outlook, SendGrid, Mailgun, or any relay)
EMAIL_HOST=smtp.gmail.com    # or smtp-relay.brevo.com, smtp.sendgrid.net, localhost
EMAIL_PORT=587               # 587 = STARTTLS, 465 = SSL, 25 = plain
EMAIL_USERNAME=
EMAIL_PASSWORD=
EMAIL_USE_TLS=true           # STARTTLS on port 587
EMAIL_USE_SSL=false          # SSL on port 465
EMAIL_TIMEOUT=30

EMAIL_FROM=
EMAIL_FROM_NAME=Ticketing System
EMAIL_TEST_MODE=true          # set false to actually send emails

# ── Payments ──────────────────────────────────────────
INTASEND_PUBLISHABLE_KEY=
INTASEND_SECRET_KEY=
INTASEND_WEBHOOK_SECRET=
INTASEND_TEST_MODE=true
```

---

## API Reference

All endpoints run on `http://localhost:8080`.

### Health
```
GET    /health
GET    /metrics
```

### Auth & 2FA
```
POST   /register
POST   /login
POST   /logout
POST   /forgot-password
POST   /resetPassword
GET    /verify-email                 # link from email
POST   /verify-email
POST   /resend-verification
GET    /verify-email/status
POST   /2fa/setup
POST   /2fa/verify-setup
POST   /2fa/verify-login
POST   /2fa/disable
GET    /2fa/status
POST   /2fa/recovery-codes
GET    /2fa/attempts
```

### Account
```
GET    /account/profile
PUT    /account/profile
DELETE /account
GET    /account/address
PUT    /account/address
DELETE /account/address
GET    /account/countries
GET    /account/preferences
PUT    /account/preferences
GET    /account/settings
PUT    /account/settings
GET    /account/timezones
GET    /account/currencies
GET    /account/date-formats
GET    /account/datetime-formats
GET    /account/security
POST   /account/security/password
GET    /account/security/login-history
POST   /account/security/lock
POST   /account/security/unlock
GET    /account/activity
GET    /account/activity/types
POST   /account/activity/log
DELETE /account/activity/clear
GET    /account/stats
GET    /account/payment-methods
GET    /account/payment-gateway
GET    /account/payment-gateway/info
```

### Events (Public)
```
GET    /events
GET    /events/search
GET    /events/{id}
GET    /events/{id}/images
```

### Events (Organizer)
```
GET    /organizers/events
GET    /organizers/events/search
POST   /organizers/events
PUT    /organizers/events/{id}
DELETE /organizers/events/{id}
POST   /organizers/events/{id}/publish
POST   /organizers/events/{id}/images
DELETE /organizers/events/{id}/images/{imageId}
```

### Ticket Classes
```
POST   /organizers/events/{eventId}/ticket-classes
GET    /organizers/events/{eventId}/ticket-classes
GET    /organizers/events/{eventId}/ticket-classes/{id}
PUT    /organizers/events/{eventId}/ticket-classes/{id}
DELETE /organizers/events/{eventId}/ticket-classes/{id}
POST   /organizers/events/{eventId}/ticket-classes/{id}/pause
POST   /organizers/events/{eventId}/ticket-classes/{id}/resume
```

### Orders
```
POST   /orders
POST   /orders/calculate
GET    /orders
GET    /orders/search
GET    /orders/stats
GET    /orders/{id}
GET    /orders/{id}/summary
PUT    /orders/{id}/status
POST   /orders/{id}/cancel
POST   /orders/{id}/refund
GET    /organizers/orders
GET    /organizers/orders/search
```

### Tickets
```
POST   /tickets/generate             # requires email verification
POST   /tickets/regenerate-qr        # requires email verification
GET    /tickets
GET    /tickets/{id}
GET    /tickets/number
GET    /tickets/stats
GET    /tickets/{id}/pdf             # requires email verification + download rate limit
POST   /tickets/{id}/transfer        # requires email verification
GET    /tickets/{id}/transfer-history
POST   /tickets/validate
POST   /tickets/validate/qr
POST   /tickets/checkin
POST   /tickets/checkin/bulk
POST   /tickets/checkin/undo
GET    /tickets/checkin/stats
POST   /tickets/bulk/export
GET    /tickets/bulk/stats
POST   /tickets/bulk/status
GET    /organizers/tickets
GET    /organizers/tickets/filter
GET    /organizers/tickets/search
```

### Inventory & Reservations
```
GET    /inventory/tickets/{id}
GET    /inventory/events/{id}
GET    /inventory/status/{id}
POST   /inventory/bulk-check
GET    /inventory/capacity/tickets/{id}
GET    /inventory/capacity/events/{id}
GET    /inventory/capacity/events/{id}/monitor
POST   /inventory/reservations
GET    /inventory/reservations
GET    /inventory/reservations/{id}
GET    /inventory/reservations/{id}/validate
POST   /inventory/reservations/{id}/extend
POST   /inventory/reservations/expired
DELETE /inventory/reservations/session
DELETE /inventory/reservations/{id}/release
GET    /inventory/events/{id}/reservations
POST   /inventory/waitlist
GET    /inventory/waitlist
GET    /inventory/waitlist/{id}
POST   /inventory/waitlist/notify
DELETE /inventory/waitlist/{id}/leave
GET    /inventory/waitlist/events/{id}/stats
```

### Payments
```
POST   /payments/initiate
GET    /payments/history
POST   /payments/verify/{id}
GET    /payments/orders/{id}/status
GET    /payments/gateways
POST   /payments/methods
GET    /payments/methods
DELETE /payments/methods/{id}
POST   /payments/methods/{id}/default
PUT    /payments/methods/{id}/expiry
POST   /payments/refunds
GET    /payments/refunds
GET    /payments/refunds/{id}/status
POST   /payments/refunds/{id}/approve
POST   /webhooks/intasend
GET    /webhooks/logs
POST   /webhooks/logs/{id}/retry
```

### Refunds
```
POST   /refunds
GET    /refunds
GET    /refunds/{id}
POST   /refunds/{id}/cancel
GET    /organizers/refunds
GET    /admin/refunds/pending
GET    /admin/refunds/{id}
POST   /admin/refunds/{id}/approve
POST   /admin/refunds/{id}/process
POST   /admin/refunds/{id}/retry
GET    /admin/refunds/statistics
POST   /refunds/bulk/process
POST   /refunds/bulk/auto-approve
GET    /refunds/bulk/stats
```

### Settlements
```
GET    /settlements/calculate/event/{id}
GET    /settlements/preview
GET    /settlements/eligibility/event/{id}
POST   /settlements/batch
GET    /settlements/{id}
GET    /settlements
POST   /settlements/{id}/approve
POST   /settlements/{id}/process
POST   /settlements/{id}/cancel
POST   /settlements/{id}/withhold
GET    /settlements/{id}/report
GET    /settlements/summary/organizer/{id}
GET    /settlements/summary/platform
GET    /settlements/export
GET    /settlements/history/organizer/{id}
GET    /settlements/pending
GET    /settlements/failed
POST   /settlements/{id}/retry
POST   /settlements/items/{id}/complete
POST   /settlements/items/{id}/fail
GET    /organizers/settlements
GET    /organizers/settlements/summary
POST   /webhooks/settlements/complete
```

### Promotions
```
GET    /promotions
GET    /promotions/active
GET    /promotions/search
POST   /promotions
GET    /promotions/{id}
GET    /promotions/code/{code}
PUT    /promotions/{id}
DELETE /promotions/{id}
POST   /promotions/{id}/clone
POST   /promotions/{id}/activate
POST   /promotions/{id}/pause
POST   /promotions/{id}/deactivate
POST   /promotions/{id}/extend
POST   /promotions/validate
POST   /promotions/eligibility
POST   /promotions/usage/revoke
GET    /promotions/{id}/usage
POST   /promotions/{id}/usage
GET    /promotions/{id}/usage/details
GET    /promotions/{id}/stats
GET    /promotions/{id}/analytics
GET    /promotions/{id}/roi
GET    /promotions/{id}/conversion
GET    /promotions/{id}/revenue
GET    /organizers/promotions
GET    /organizers/promotions/stats
```

### Attendees
```
GET    /attendees
GET    /attendees/filter
GET    /attendees/search
GET    /attendees/search/event
GET    /attendees/count
GET    /attendees/{id}
GET    /attendees/ticket
GET    /attendees/order/{id}
POST   /attendees/checkin
POST   /attendees/checkin/bulk
POST   /attendees/checkin/undo
PUT    /attendees/{id}
POST   /attendees/{id}/no-show
POST   /attendees/{id}/transfer
POST   /attendees/bulk/email
POST   /attendees/bulk/export
POST   /attendees/event/update-email
GET    /attendees/export
GET    /attendees/badges
GET    /attendees/stats
GET    /attendees/report/checkin
GET    /attendees/timeline
GET    /attendees/no-shows
GET    /organizers/attendees
```

### Venues
```
POST   /venues
GET    /venues
GET    /venues/{id}
PUT    /venues/{id}
DELETE /venues/{id}
GET    /venues/search/location
GET    /venues/type
GET    /venues/{id}/stats
GET    /venues/{id}/events
GET    /venues/{id}/availability
GET    /venues/{id}/calendar
GET    /venues/available
POST   /venues/{id}/restore
DELETE /venues/{id}/permanent
```

### Organizers
```
POST   /organizers/apply
GET    /organizers/profile
PUT    /organizers/profile
GET    /organizers/onboarding/status
GET    /organizers/dashboard
GET    /organizers/dashboard/stats
POST   /organizers/logo
POST   /organizers/verification/email
POST   /organizers/verify-email
GET    /organizers/bank-details
PUT    /organizers/bank-details
GET    /admin/organizers/pending
POST   /admin/organizers/{id}/verify
PUT    /organizers/kyc/update
```

### Admin
```
GET    /admin/users
GET    /admin/users/search
GET    /admin/users/stats
GET    /admin/users/{id}
PUT    /admin/users/{id}/role
PUT    /admin/users/{id}/status
GET    /admin/payments
```

### Support
```
POST   /support/tickets
GET    /support/tickets
GET    /support/tickets/stats
GET    /support/tickets/{id}
POST   /support/tickets/{id}/comments
PUT    /support/tickets/{id}
```

### Notifications
```
POST   /notifications/test
POST   /notifications/welcome
POST   /notifications/verification
POST   /notifications/password-reset
```

---

## Technology Stack

| Category | Choice |
|---|---|
| Language | Go 1.21+ |
| HTTP router | gorilla/mux |
| ORM | GORM |
| Database | PostgreSQL |
| Cache / Sessions | Redis (go-redis) with in-memory fallback |
| Storage | AWS S3 with local filesystem fallback |
| PDF generation | gofpdf |
| QR codes | go-qrcode |
| Payments | Intasend (M-Pesa), Stripe |
| Email | Brevo API or any SMTP server |
| Password hashing | argon2id |
| Monitoring | Prometheus + Grafana |
| Rate limiting | Custom token bucket |

---

## Limitations (addressed in v2 and v3)

- **Synchronous email delivery** — a slow Brevo API call or SMTP timeout blocks the HTTP response
- **In-process reservation cleanup** — runs as a goroutine inside the API server; if the server restarts, reservations that expired during the downtime are not cleaned up until the next poll cycle
- **No retry on notification failure** — if an email send fails mid-request, it is lost with no retry mechanism
- **Single point of failure** — all functionality is in one process; a panic in any handler affects the entire system
- **Vertical scaling only** — there is no way to scale the email-sending capacity independently of the API capacity

These are addressed in v2 by moving to an event-driven architecture with independent workers and an outbox pattern.

---

## License

MIT

---

Built with Go — [kamausimon217@gmail.com](mailto:kamausimon217@gmail.com)
