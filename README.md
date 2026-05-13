# Event Ticketing System

A production-grade event ticketing platform built in Go, designed and rebuilt three times to demonstrate architectural progression — from a monolith to an event-driven system to a fully cloud-native AWS deployment.

Each version is a complete, runnable application that solves the same domain problem with a meaningfully different architecture. They are not toy examples — each one includes authentication, payment processing, PDF ticket generation, real-time inventory management, refund workflows, analytics, and email delivery.

---

## Versions

| Version | Folder | Architecture | Key Technologies |
|---|---|---|---|
| v1 | [v1-monolithic-architecture/](v1-monolithic-architecture/) | Monolithic REST API | Go, PostgreSQL, Redis, S3 |
| v2 | [v2-event-driven-arch/](v2-event-driven-arch/) | Event-driven microservices | Go, Kafka, PostgreSQL, Redis, S3 |
| v3 | [v3-aws-native-arch/](v3-aws-native-arch/) | Cloud-native on AWS | Go, SNS/SQS, Aurora PostgreSQL, ElastiCache, S3, Terraform |

---

## Architectural Evolution

### v1 — Monolithic

A single deployable binary. The API server handles every request synchronously: ticket purchase, payment confirmation, email delivery, inventory decrement, and PDF generation all happen within the same HTTP request lifecycle.

Simple to understand and deploy, but tightly coupled — a slow email provider or a payment gateway timeout directly affects API response time.

```
Client → API Server → PostgreSQL
                    → Redis (cache)
                    → S3 (uploads)
                    → SMTP (emails, in-request)
```

**Single entry point:** `cmd/api-server`

---

### v2 — Event-Driven

The same domain, redesigned around events. The API server writes domain events to an outbox table, and a fleet of independent workers consume those events from Kafka topics. Each worker owns a single responsibility.

A slow email delivery no longer blocks ticket generation. Workers can be scaled, restarted, and deployed independently. The outbox pattern guarantees no events are lost even under failure.

```
Client → API Server → PostgreSQL (+ outbox table)
                        ↓
                 Outbox Publisher
                        ↓
                      Kafka
                 ┌──────┼──────┐
                 ↓      ↓      ↓
           payment   ticket   email
           worker    worker   worker  ...
```

**Entry points:** `cmd/api-server` + 11 worker binaries

**Kafka topics:**
`order.created` · `order.payment_confirmed` · `order.confirmed` · `order.cancelled` · `order.completed` · `reservation.expired` · `inventory.released` · `event.cancelled` · `payment.completed` · `tickets.generated` · `notification.email.requested` · `refund.approved` · `refund.completed`

---

### v3 — AWS-Native

The same event-driven design from v2, with every self-managed Docker service replaced by a fully managed AWS equivalent. No Kafka cluster to size and operate — SNS/SQS handles fan-out and queuing. No Postgres container to back up — Aurora does it automatically. No Redis instance to patch — ElastiCache manages it.

The application code changes are minimal (a new `internal/messaging/` package replaces `internal/kafka/`). The meaningful shift is in the infrastructure: Terraform provisions the entire AWS environment in one command.

```
Client → API Server → Aurora PostgreSQL (+ outbox table)
                             ↓
                      Outbox Publisher
                             ↓
                          Amazon SNS
                    ┌────────┼────────┐
                    ↓        ↓        ↓
                 SQS      SQS      SQS    (14 queues total)
                  ↓        ↓        ↓
              payment   ticket   email
              worker    worker   worker  ...
```

**Entry points:** Same binaries as v2, wired to SQS instead of Kafka

**AWS services:** SNS · SQS · RDS Aurora PostgreSQL · ElastiCache · S3 · (Terraform-managed)

---

## Feature Set

All three versions implement the same full feature set:

### Attendees
- Browse, search, and filter events
- Reserve and purchase tickets (Stripe, Intasend/M-Pesa, offline)
- Receive order confirmation and PDF ticket emails
- Download PDF tickets with QR codes
- Transfer tickets to other attendees
- Request and track refunds
- Waitlist for sold-out events
- Two-factor authentication (TOTP)

### Organizers
- Create and publish events with venue and capacity management
- Define multiple ticket classes with pricing and availability windows
- Real-time sales dashboard — revenue, tickets sold, check-in rate
- Bulk attendee export
- Approve or reject refund requests
- Cancel events with automatic refund cascade
- Promotional codes and discount management
- Payout account management

### Admins
- User and organizer management
- Platform-wide analytics
- AI-assisted support ticket classification
- Rate limiting and security monitoring
- Bulk operations

---

## Technology Comparison

| | v1 | v2 | v3 |
|---|---|---|---|
| **Language** | Go | Go | Go |
| **API** | gorilla/mux | gorilla/mux | gorilla/mux |
| **ORM** | GORM | GORM | GORM |
| **Database** | PostgreSQL (Docker) | PostgreSQL (Docker) | RDS Aurora PostgreSQL |
| **Messaging** | — | Kafka (kafka-go) | Amazon SNS + SQS |
| **Cache** | Redis (Docker) | Redis (Docker) | ElastiCache Redis |
| **Storage** | AWS S3 | AWS S3 | AWS S3 |
| **Payments** | Intasend, Stripe | Intasend, Stripe | Intasend, Stripe |
| **Email** | Brevo API | Brevo API | Brevo API |
| **PDF** | gofpdf | gofpdf | gofpdf |
| **QR Codes** | go-qrcode | go-qrcode | go-qrcode |
| **Monitoring** | Prometheus + Grafana | Prometheus + Grafana | Prometheus + Grafana |
| **Infrastructure** | Docker Compose | Docker Compose | Terraform + AWS |
| **Local dev** | Docker Compose | Docker Compose | LocalStack + Docker Compose |

---

## Quick Start

### v1 — Monolithic

```bash
cd v1-monolithic-architecture

# Start dependencies
docker compose -f docker-compose.monitoring.yml up -d

# Configure
cp .env.example .env   # fill in DB, Redis, S3, payment keys

# Run
go run cmd/api-server/main.go
```

### v2 — Event-Driven

```bash
cd v2-event-driven-arch

# Start Kafka, PostgreSQL, Redis
docker compose up -d

# Configure
cp .env.example .env   # add KAFKA_BROKERS=localhost:9092

# Run the API server
go run cmd/api-server/main.go

# Run workers (each in its own terminal)
go run cmd/outbox-publisher/main.go
go run cmd/payment-worker/main.go
go run cmd/ticket-worker/main.go
go run cmd/ticket-email-worker/main.go
go run cmd/notification-worker/main.go
go run cmd/email-worker/main.go
go run cmd/event-cancellation-worker/main.go
go run cmd/refund-processor-worker/main.go
go run cmd/reservation-expiry-worker/main.go
go run cmd/waitlist-worker/main.go
go run cmd/analytics-worker/main.go
```

### v3 — AWS-Native

**Local development with LocalStack (no AWS account needed):**

```bash
cd v3-aws-native-arch

# Start LocalStack, PostgreSQL, Redis
docker compose -f docker-compose.localstack.yml up -d

# Create all SNS topics and SQS queues in LocalStack
bash scripts/localstack-init.sh
# The script prints all SNS_ARN_* and SQS_URL_* values — paste them into .env

# Configure
cp .env.example .env
# Set AWS_ENDPOINT_URL=http://localhost:4566
# Set AWS_ACCESS_KEY_ID=test / AWS_SECRET_ACCESS_KEY=test
# Paste the ARNs and URLs from the init script

# Run (same commands as v2)
go run cmd/api-server/main.go
go run cmd/outbox-publisher/main.go
# ... other workers
```

**Deploy to real AWS:**

```bash
cd v3-aws-native-arch/terraform

terraform init
terraform apply -var="db_password=<your-password>"

# Export outputs to .env
terraform output -json sns_arns
terraform output -json sqs_queue_urls
terraform output db_endpoint
terraform output redis_endpoint
```

---

## Project Structure (per version)

```
<version>/
├── cmd/
│   ├── api-server/                  # HTTP API — routes, handlers, middleware
│   ├── outbox-publisher/            # Polls outbox table, publishes events     (v2, v3)
│   ├── payment-worker/              # Confirms offline payments                (v2, v3)
│   ├── ticket-worker/               # Generates tickets after payment          (v2, v3)
│   ├── ticket-email-worker/         # Emails PDF tickets to attendees          (v2, v3)
│   ├── notification-worker/         # Sends order confirmation emails          (v2, v3)
│   ├── email-worker/                # Delivers outbox-queued emails            (v2, v3)
│   ├── event-cancellation-worker/   # Cascades event cancellation              (v2, v3)
│   ├── refund-processor-worker/     # Executes approved refunds                (v2, v3)
│   ├── reservation-expiry-worker/   # Expires stale reservations               (v2, v3)
│   ├── waitlist-worker/             # Notifies waitlisted users                (v2, v3)
│   └── analytics-worker/            # Updates per-event daily stats            (v2, v3)
│
├── internal/
│   ├── kafka/                       # Kafka producer + consumer                (v2 only)
│   ├── messaging/                   # SNS publisher + SQS consumer             (v3 only)
│   ├── outbox/                      # Transactional outbox pattern             (v2, v3)
│   ├── config/                      # Environment-based config loader
│   ├── database/                    # GORM connection setup
│   ├── cache/                       # Redis client
│   ├── storage/                     # S3 client
│   ├── auth/                        # JWT, 2FA, password reset
│   ├── events/                      # Event CRUD, publishing, image uploads
│   ├── orders/                      # Order lifecycle, calculations
│   ├── tickets/                     # Ticket generation, PDF, QR, check-in
│   ├── payments/                    # Stripe, Intasend, webhook handlers
│   ├── refunds/                     # Refund requests, approvals, processing
│   ├── inventory/                   # Capacity, reservations, waitlist
│   ├── notifications/               # Email templates and delivery
│   ├── analytics/                   # Sales metrics, reports, dashboards
│   ├── organizers/                  # Organizer onboarding, profiles, payouts
│   ├── attendees/                   # Attendee management, bulk operations
│   ├── promotions/                  # Promo codes, discounts
│   └── settlement/                  # Organizer payout calculations
│
├── terraform/                       # AWS infrastructure as code              (v3 only)
│   ├── vpc.tf                       # VPC, subnets, security groups
│   ├── sns.tf                       # 13 SNS topics
│   ├── sqs.tf                       # 14 SQS queues + DLQs + subscriptions
│   ├── rds.tf                       # Aurora PostgreSQL cluster
│   └── elasticache.tf               # ElastiCache Redis
│
├── migrations/                      # SQL schema migrations
├── docker-compose.yml               # Kafka + PostgreSQL + Redis               (v2)
├── docker-compose.localstack.yml    # LocalStack + PostgreSQL + Redis          (v3)
├── docker-compose.monitoring.yml    # Prometheus + Grafana
└── .env.example                     # All required environment variables
```

---

## API Reference

All versions expose the same REST API on port `8080`.

### Auth
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
POST   /api/auth/verify-email
POST   /api/auth/reset-password
POST   /api/auth/2fa/enable
POST   /api/auth/2fa/verify
```

### Events
```
GET    /api/events
GET    /api/events/:id
POST   /api/events                         # organizer
PUT    /api/events/:id                     # organizer
DELETE /api/events/:id                     # organizer
POST   /api/organizers/events/:id/publish  # organizer
POST   /api/organizers/events/:id/cancel   # organizer
```

### Orders & Tickets
```
POST   /api/orders
GET    /api/orders/:id
POST   /api/orders/:id/payment
GET    /api/tickets
GET    /api/tickets/:id
GET    /api/tickets/:id/pdf
POST   /api/tickets/:id/checkin
POST   /api/tickets/:id/transfer
```

### Refunds
```
POST   /api/refunds
GET    /api/refunds/:id
POST   /api/refunds/:id/approve            # organizer/admin
POST   /api/refunds/:id/reject             # organizer/admin
```

### Analytics
```
GET    /api/analytics/events/:id/summary
GET    /api/analytics/events/:id/sales
GET    /api/analytics/events/:id/audience
GET    /api/organizers/:id/dashboard
GET    /api/admin/analytics
```

---

## Monitoring

All three versions ship the same Prometheus + Grafana stack:

```bash
docker compose -f docker-compose.monitoring.yml up -d
```

- Prometheus: [http://localhost:9090](http://localhost:9090)
- Grafana: [http://localhost:3001](http://localhost:3001) — `admin / admin123`

---

## Design Decisions Across Versions

**Why Kafka in v2, not HTTP calls between services?**
HTTP-based service calls create synchronous coupling — if the email service is slow, ticket generation blocks. Kafka decouples producers from consumers completely. Each worker can fail and restart independently without affecting the API response.

**Why SNS/SQS in v3 instead of managed Kafka (MSK)?**
MSK preserves the Kafka API but still requires cluster sizing decisions and costs significantly more at low throughput. SNS/SQS is fully serverless and demonstrates a meaningfully different technology from v2 — a better portfolio differentiator. The code change is also a useful demonstration of how the outbox pattern's `Sender` interface makes the messaging backend swappable with minimal impact.

**Why the outbox pattern in v2 and v3?**
Publishing directly to Kafka/SNS from within an HTTP handler creates a distributed transaction: if the handler commits the DB write but the publish fails, the event is silently lost. The outbox table makes event publishing atomic with the business data — the publisher can crash and restart, and it will simply resume draining the table.

**Why the same domain across all three versions?**
Changing both the problem and the solution at the same time makes it hard to see what the architecture is actually contributing. Keeping the domain constant makes the architectural trade-offs legible.

---

## License

MIT

---

Built with Go — [kamausimon217@gmail.com](mailto:kamausimon217@gmail.com)
