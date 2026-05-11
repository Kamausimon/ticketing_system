# Ticketing System — v2 Event-Driven Architecture

This is the second major version of the ticketing system, rearchitected from a monolithic REST API into an event-driven, multi-service Go application. The core design shift is replacing synchronous in-process calls with Kafka events and an outbox pattern, allowing each concern (payments, ticketing, notifications, refunds, analytics) to scale and fail independently.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Services](#services)
- [Event Flow](#event-flow)
- [Outbox Pattern](#outbox-pattern)
- [Internal Packages](#internal-packages)
- [Auth & Security](#auth--security)
- [Payment & Refund Flow](#payment--refund-flow)
- [Monitoring & Observability](#monitoring--observability)
- [Infrastructure](#infrastructure)
- [Configuration](#configuration)
- [Running Locally](#running-locally)
- [API Overview](#api-overview)

---

## Architecture Overview

```
                          ┌─────────────────┐
                          │   API Server    │  <- REST (port 8080)
                          │  (cmd/api-server)│
                          └────────┬────────┘
                                   │ writes DB + outbox atomically
                          ┌────────▼────────┐
                          │  PostgreSQL DB  │
                          │  outbox_events  │
                          └────────┬────────┘
                                   │ polls every 1s
                          ┌────────▼────────┐
                          │ Outbox Publisher │
                          └────────┬────────┘
                                   │ publishes to
                          ┌────────▼────────┐
                          │     Kafka       │  <- event bus
                          └──┬──┬──┬──┬────┘
                             │  │  │  │
              ┌──────────────┘  │  │  └──────────────────┐
              │                 │  │                      │
   ┌──────────▼──────┐  ┌──────▼──────┐  ┌──────────────▼──────┐
   │ Payment Worker  │  │Ticket Worker│  │ Notification Worker  │
   └─────────────────┘  └─────────────┘  └──────────────────────┘
         (+ 7 more workers)
```

Every state change in the api-server is written to the `outbox_events` table in the same DB transaction as the business data. The outbox publisher picks these up and forwards them to Kafka. Workers consume Kafka topics and act independently — no shared in-process state.

---

## Services

There are 13 binaries under `cmd/`:

| Service | Description |
|---|---|
| `api-server` | Main REST API. Handles all HTTP requests: events, orders, payments, tickets, organizers, admin, settlements, refunds, analytics. |
| `outbox-publisher` | Polls `outbox_events` table every second and publishes pending events to Kafka (50-event batches, 5 retries, marks `published` on success). |
| `payment-worker` | Consumes `order.created` → validates orders → initiates payment via IntaSend → emits `payment.completed`. |
| `ticket-worker` | Consumes `payment.completed` → creates `Ticket` records → generates QR codes → reserves inventory → emits `tickets.generated`. |
| `ticket-email-worker` | Consumes `tickets.generated` → generates PDF ticket → sends email with attachment. |
| `notification-worker` | Consumes `order.confirmed` → sends order confirmation email to customer. |
| `email-worker` | Consumes `notification.email.requested` → generic SMTP relay with retry logic. |
| `analytics-worker` | Consumes 6 topics → aggregates into `system_metrics`, `event_metrics`, and `user_engagement_metrics` tables. |
| `refund-processor-worker` | Consumes `refund.approved` → calls IntaSend refund API → marks refund `processed` → emits `refund.completed`. |
| `event-cancellation-worker` | Consumes `event.cancelled` → auto-refunds all tickets → marks orders cancelled → notifies attendees. |
| `reservation-expiry-worker` | Polls `reserved_tickets` table every minute → releases expired reservations → emits `reservation.expired` → triggers waitlist. |
| `waitlist-worker` | Consumes `inventory.released` → notifies next waitlist user → reserves ticket with 15-minute timeout. |
| `check-migration` | Utility that validates DB connection and verifies expected tables exist. |

---

## Event Flow

### Kafka Topics

| Topic | Published by | Consumed by |
|---|---|---|
| `order.created` | api-server (via outbox) | payment-worker |
| `order.confirmed` | payment-worker | notification-worker, analytics-worker |
| `order.cancelled` | event-cancellation-worker | analytics-worker |
| `order.completed` | api-server | analytics-worker |
| `payment.completed` | payment-worker (via outbox) | ticket-worker, analytics-worker |
| `tickets.generated` | ticket-worker | ticket-email-worker, analytics-worker |
| `notification.email.requested` | various workers | email-worker |
| `reservation.expired` | reservation-expiry-worker | analytics-worker |
| `inventory.released` | reservation-expiry-worker | waitlist-worker |
| `event.cancelled` | api-server (via outbox) | event-cancellation-worker |
| `refund.approved` | api-server / event-cancellation-worker | refund-processor-worker |
| `refund.completed` | refund-processor-worker | analytics-worker |

### Happy-Path Order Sequence

```
Customer places order
        │
        ▼
api-server → Order(pending) + OutboxEvent(order.created)   [single transaction]
        │
        ▼
outbox-publisher → Kafka[order.created]
        │
        ▼
payment-worker → IntaSend API → PaymentRecord(pending)
        │ (webhook callback)
        ▼
payment-worker → PaymentRecord(success) + OutboxEvent(payment.completed)
        │
        ▼
ticket-worker → Ticket records + QR codes + OutboxEvent(tickets.generated)
        │
        ├──▶ ticket-email-worker → PDF + email to customer
        ├──▶ notification-worker → confirmation email
        └──▶ analytics-worker   → metrics aggregation
```

All workers use Kafka consumer groups and only commit offsets after successful processing, giving at-least-once delivery semantics with idempotent consumers.

---

## Outbox Pattern

**Problem:** Publishing a Kafka event after a DB write is a two-phase operation. If the process crashes between the write and the publish, the event is lost. If Kafka is temporarily unavailable, the write either fails unnecessarily or the event is silently dropped.

**Solution:** Write the event to an `outbox_events` table in the same DB transaction as the business data. A separate `outbox-publisher` process polls the table and forwards to Kafka.

```
┌─────────────────────────────────────┐
│  DB transaction (api-server)        │
│  INSERT INTO orders ...             │
│  INSERT INTO outbox_events          │
│    (topic, payload, status=pending) │
└─────────────────────────────────────┘
           | (1-second poll)
┌─────────────────────────────────────┐
│  outbox-publisher                   │
│  SELECT * FROM outbox_events        │
│    WHERE status='pending' LIMIT 50  │
│  → kafka.Publish(topic, payload)    │
│  → UPDATE status='published'        │
└─────────────────────────────────────┘
```

**Guarantees:**
- Event is never lost: if the api-server crashes, the outbox row survives in Postgres.
- Kafka and DB are never out of sync.
- Failed publishes are retried up to 5 times before being marked `failed` for manual review.

Implementation: `internal/outbox/`, `models.OutboxEvent`.

---

## Internal Packages

| Package | Responsibilities |
|---|---|
| `accounts` | User profile, settings (timezone, currency, date formats), account lock/unlock, login history |
| `admin` | User listing, role management, organizer approvals, system oversight |
| `analytics` | Prometheus metric registration, HTTP middleware instrumentation, background system metric collector |
| `api_events` | Kafka message DTOs — serialization structs for all events |
| `attendees` | Check-in/undo, no-show tracking, bulk email, badge export, attendance reports |
| `auth` | Registration, login/logout, email verification, password reset, JWT issuance, 2FA |
| `cache` | Redis SessionManager with in-memory fallback, event caching, connection health monitoring |
| `config` | Loads `.env` into typed `Config` struct |
| `database` | GORM init, connection pooling, auto-migration |
| `events` | Event CRUD, publish/unpublish, image management, organizer permissions, search |
| `handlers` | HTTP router wiring, dependency injection for all handlers |
| `inventory` | Ticket availability, reservations (15-min timeout), waitlist (FIFO), bulk availability |
| `kafka` | Consumer/producer wrappers, topic constants, broker management |
| `middleware` | JWT auth, email verification gate, CORS, Prometheus instrumentation |
| `models` | 35+ GORM models: User, Order, Ticket, Payment, Refund, Settlement, Event, Organizer, etc. |
| `notifications` | SMTP abstraction (Gmail/SendGrid/SES/Mailtrap/Zoho), template rendering, retry |
| `orders` | Order creation, calculations, status transitions, cancellation, payment linking |
| `organizers` | Application/approval workflow, KYC, bank detail storage (AES-256 encrypted), logo upload |
| `outbox` | OutboxEvent model, repository (fetch/mark/retry), publisher loop |
| `payments` | IntaSend integration, payment initiation/verification, webhook handling, payment method CRUD |
| `promotions` | Promo code CRUD, eligibility, usage tracking, ROI analytics |
| `refunds` | Request creation, approval workflow, bulk operations, state machine |
| `security` | AES-256-GCM encryption/decryption for sensitive fields |
| `seed` | DB seeding: timezones, currencies, date formats, notification preferences |
| `settlement` | Payout calculation, batch creation, approval/processing, dispute withholding |
| `storage` | S3 abstraction with local fallback, signed URLs, public URL generation |
| `support` | Support ticket CRUD, comments, AI-assisted responses, admin assignment |
| `ticketclasses` | Ticket type CRUD (price, quantity, sales window), pause/resume, per-class inventory |
| `tickets` | Ticket generation, QR regeneration, PDF download (rate-limited), transfer, bulk check-in, export |
| `venues` | Venue CRUD, capacity management, location search, availability calendar, soft delete |

---

## Auth & Security

### Authentication

- **JWT (HS256)** with claims: `user_id`, `email`, `exp`, `iat`
- Refresh tokens stored in DB with `token_version` for revocation on logout
- `LoginHistory` table records every login (IP, user-agent, timestamp)

### Two-Factor Authentication (TOTP)

- 30-second TOTP windows, Base32-encoded 160-bit secrets
- QR code generation for authenticator apps (Google Authenticator, Authy)
- 10 bcrypt-hashed recovery codes stored per user
- Attempt logging via `TwoFactorAttempt` table
- Rate limited: **3 attempts/minute per user** — prevents brute force even with IP rotation

Endpoints: `POST /2fa/setup`, `/2fa/verify-setup`, `/2fa/verify-login`, `/2fa/disable`, `/2fa/recovery-codes`

### Rate Limiting

Token bucket + sliding window hybrid implemented in `pkg/ratelimit`:

| Scope | Limit |
|---|---|
| General API | 100 req/s per IP |
| Auth endpoints | 10 req/min per IP |
| Login | 5 attempts/min per IP |
| Payment initiation | 5 req/min per IP |
| PDF download | 3 req/s per user (JWT-keyed) |
| Inventory checks | 50 req/s per IP (burst: 100) |
| 2FA verification | 3 attempts/min per user |

### Email Verification Gate

Middleware blocks payment initiation and PDF downloads if the user's email is unverified. Verification supports both link-based and OTP-code flows.

### Encryption at Rest

Organizer bank details are encrypted with **AES-256-GCM** before storage. The nonce is prepended to the ciphertext and the result is base64-encoded. Key must be exactly 16, 24, or 32 bytes (`ENCRYPTION_KEY` env var). Implementation: `internal/security`.

---

## Payment & Refund Flow

### Payment

```
POST /orders                    → Order(pending) + outbox event
POST /payments/initiate         → IntaSend API → returns payment URL
POST /webhooks/intasend         → webhook from IntaSend → PaymentRecord(success)
                                  → outbox event(payment.completed)
payment-worker consumes         → Order(confirmed)
ticket-worker consumes          → Tickets created, inventory reserved
```

Webhook requests are verified against `INTASEND_WEBHOOK_SECRET` and logged to `webhook_logs` for debugging and manual retry.

### Refund

```
POST /refunds                          → RefundRecord(pending_approval)
POST /admin/refunds/{id}/approve       → RefundRecord(approved) + outbox event
refund-processor-worker consumes       → IntaSend refund API call
                                         → RefundRecord(processed)
                                         → outbox event(refund.completed)
```

Bulk operations: `POST /refunds/bulk/process`, `POST /refunds/bulk/auto-approve` (auto-triggers on event cancellation).

---

## Monitoring & Observability

### Prometheus Metrics (api-server `/metrics`)

**HTTP:** `ticketing_http_requests_total`, `ticketing_http_request_duration_seconds`

**Business:** `ticketing_tickets_sold_total`, `ticketing_revenue_total`, `ticketing_orders_created`, `ticketing_payment_attempts`, `ticketing_inventory_available`

**System:** `ticketing_goroutines_count`, `ticketing_memory_usage_bytes`, `ticketing_cpu_usage_percent`

**Cache:** `ticketing_cache_hits_total`, `ticketing_cache_misses_total`

**Database:** `ticketing_db_query_duration_seconds`, `ticketing_db_connections`, `ticketing_db_errors_total`

### Alert Rules (`prometheus/alerts.yml`)

| Alert | Severity | Condition |
|---|---|---|
| No sales in 30 min | warning | zero ticket sales over 30-minute window |
| Low inventory | warning | available tickets < 10 |
| High payment failure rate | critical | failure rate > 10% |
| Revenue drop | warning | revenue < 50% of previous day |
| High HTTP error rate | critical | error rate > 5% |
| High P95 latency | warning | P95 > 2s |
| Kafka consumer lag | warning | lag exceeds threshold |

### Grafana Dashboards

Three auto-provisioned dashboards (port 3001, `admin/admin123`):

- **Business Overview** — revenue, tickets sold, active events, event creation rate
- **Payment Analytics** — payment success rate, refund volume, transaction throughput
- **System Performance** — response times, error rates, DB connection pool, goroutine count

### Monitoring Stack

```
docker-compose.monitoring.yml
├── Prometheus    (port 9090, 30-day retention)
├── Grafana       (port 3001)
├── AlertManager  (port 9093) → Slack / email
├── Node Exporter (port 9100) → host CPU/memory/disk
└── cAdvisor      (port 8080) → per-container resource usage
```

---

## Infrastructure

### docker-compose.yml (development stack)

| Service | Image | Port |
|---|---|---|
| PostgreSQL | `postgres:16-alpine` | 5432 |
| Kafka | `bitnami/kafka:3.7` (KRaft, no Zookeeper) | 9092 |
| Redis | `redis:alpine` | 6379 |

Kafka runs in KRaft mode — no Zookeeper dependency. Topics are auto-created on first publish.

### Database

35+ tables managed by GORM AutoMigrate. Migration order: custom types → core user/auth → business (events, venues, organizers) → transactional (orders, tickets, payments, refunds) → analytics. Run automatically on api-server startup or via `cmd/check-migration`.

### Redis

Used for session storage with automatic in-memory fallback if Redis is unavailable. The `SessionManager` pings Redis every 10 seconds and switches backends transparently — a Redis outage does not take down auth.

### S3 / Storage

File storage (ticket PDFs, event images, organizer logos) goes to S3 in production and falls back to local filesystem in development. Configured via `AWS_*` and `S3_BUCKET_NAME` env vars.

---

## Configuration

Copy `.env.example` to `.env` and populate:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=ticketing_system

# JWT
JWT_SECRET=your-jwt-secret-min-32-chars

# Encryption (must be 16, 24, or 32 bytes)
ENCRYPTION_KEY=your-32-byte-encryption-key-here

# Email (example: Mailtrap for dev)
EMAIL_HOST=smtp.mailtrap.io
EMAIL_PORT=587
EMAIL_USERNAME=your-mailtrap-user
EMAIL_PASSWORD=your-mailtrap-pass
EMAIL_FROM=noreply@yourdomain.com

# Payments (IntaSend)
INTASEND_PUBLISHABLE_KEY=your-publishable-key
INTASEND_SECRET_KEY=your-secret-key
INTASEND_WEBHOOK_SECRET=your-webhook-secret
INTASEND_TEST_MODE=true

# Kafka
KAFKA_BROKERS=localhost:9092

# Redis (optional — falls back to in-memory if not set)
REDIS_URL=redis://localhost:6379

# S3 (optional — falls back to local storage if not set)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
S3_BUCKET_NAME=your-bucket
```

---

## Running Locally

**Prerequisites:** Go 1.25+, Docker, Docker Compose

```bash
# 1. Start infrastructure (PostgreSQL, Kafka, Redis)
docker compose up -d

# 2. Configure environment
cp .env.example .env
# edit .env with your values

# 3. Start the API server (runs DB migrations automatically)
go run ./cmd/api-server

# 4. Start workers (each in a separate terminal)
go run ./cmd/outbox-publisher
go run ./cmd/payment-worker
go run ./cmd/ticket-worker
go run ./cmd/ticket-email-worker
go run ./cmd/notification-worker
go run ./cmd/email-worker
go run ./cmd/analytics-worker
go run ./cmd/refund-processor-worker
go run ./cmd/event-cancellation-worker
go run ./cmd/reservation-expiry-worker
go run ./cmd/waitlist-worker

# 5. (Optional) Start monitoring stack
docker compose -f docker-compose.monitoring.yml up -d
```

| Endpoint | URL |
|---|---|
| API | http://localhost:8080 |
| Grafana | http://localhost:3001 (admin / admin123) |
| Prometheus | http://localhost:9090 |

---

## API Overview

The api-server exposes 200+ routes. Key groups:

| Group | Sample Endpoints |
|---|---|
| Auth | `POST /register`, `POST /login`, `POST /logout`, `POST /forgot-password`, `POST /resetPassword`, `GET /verify-email` |
| 2FA | `POST /2fa/setup`, `POST /2fa/verify-setup`, `POST /2fa/verify-login`, `POST /2fa/disable` |
| Events | `GET /events`, `POST /events`, `PUT /organizers/events/{id}`, `POST /organizers/events/{id}/publish` |
| Tickets | `GET /tickets/{id}`, `GET /tickets/{id}/pdf`, `POST /tickets/{id}/transfer`, `POST /tickets/checkin` |
| Orders | `POST /orders`, `GET /orders`, `PUT /orders/{id}/status`, `POST /orders/{id}/cancel` |
| Payments | `POST /payments/initiate`, `POST /payments/verify/{id}`, `POST /webhooks/intasend` |
| Refunds | `POST /refunds`, `GET /admin/refunds/pending`, `POST /admin/refunds/{id}/approve` |
| Settlements | `POST /settlements/batch`, `GET /settlements/{id}/report` |
| Organizers | `POST /organizers/apply`, `GET /organizers/profile`, `PUT /organizers/bank-details` |
| Admin | `GET /admin/users`, `GET /admin/organizers/pending`, `POST /admin/users/{id}/role` |
| Promotions | `GET /promotions`, `POST /promotions`, `POST /promotions/validate` |
| Venues | `GET /venues`, `POST /venues`, `DELETE /venues/{id}` |
| Metrics | `GET /metrics` (Prometheus scrape endpoint) |

---

## Key Design Decisions

**Outbox over dual-write** — Guarantees that a DB write and its corresponding Kafka event are always in sync. No event is ever silently dropped, even if Kafka is temporarily unavailable.

**One binary per concern** — Each worker has a single job. Payment processing, ticket generation, notifications, and analytics can be scaled, redeployed, or debugged independently.

**At-least-once delivery with idempotent consumers** — Workers only commit Kafka offsets after successful processing. All consumers are designed to handle duplicate messages safely.

**AES-256-GCM for sensitive fields** — Bank account details are encrypted before hitting the database. The encryption key never leaves the application process.

**Redis optional** — The session layer degrades gracefully to in-memory storage if Redis is unavailable, so a Redis outage does not take down auth.

**Rate limiting at the application layer** — Token bucket + sliding window protects expensive operations (payment initiation, PDF generation, 2FA verification) without requiring an API gateway.
