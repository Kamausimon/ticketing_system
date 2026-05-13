# Event Ticketing System — v3 AWS-Native Architecture

This is the third iteration of the event ticketing platform. Each version lives on its own branch and represents a distinct architectural evolution:

| Branch | Architecture | Messaging | Database | Cache |
|---|---|---|---|---|
| `main` (v1) | Monolithic | — | PostgreSQL (Docker) | Redis (Docker) |
| `v2-event-driven-arch` | Event-driven microservices | Kafka (Docker) | PostgreSQL (Docker) | Redis (Docker) |
| `v3-aws-native-arch` (this branch) | Event-driven, cloud-native | **SNS + SQS** | **RDS Aurora PostgreSQL** | **ElastiCache** |

v3 keeps the same domain logic and event-driven patterns from v2 but replaces every self-managed Docker service with a fully managed AWS equivalent — eliminating operational overhead and enabling production-grade scaling without running your own infrastructure.

---

## Architecture Overview

```
                        ┌─────────────────────────┐
                        │       API Server         │
                        │   (cmd/api-server)       │
                        └────────────┬────────────┘
                                     │ writes events to
                                     ▼
                        ┌─────────────────────────┐
                        │    Outbox Table          │
                        │   (RDS Aurora)           │
                        └────────────┬────────────┘
                                     │ polled by
                                     ▼
                        ┌─────────────────────────┐
                        │   Outbox Publisher       │
                        │ (cmd/outbox-publisher)   │
                        └────────────┬────────────┘
                                     │ publishes to
                                     ▼
                    ┌────────────────────────────────┐
                    │          Amazon SNS             │
                    │   (13 topics, one per event)    │
                    └────┬─────┬──────┬──────┬───────┘
                         │     │      │      │  fan-out subscriptions
              ┌──────────┘     │      │      └──────────┐
              ▼                ▼      ▼                  ▼
        ┌──────────┐    ┌──────────┐ ┌──────────┐ ┌──────────┐
        │  SQS     │    │  SQS     │ │  SQS     │ │  SQS     │
        │ Queue    │    │ Queue    │ │ Queue    │ │ Queue    │
        └────┬─────┘    └────┬─────┘ └────┬─────┘ └────┬─────┘
             │               │             │             │
             ▼               ▼             ▼             ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐ ┌──────────┐
        │ payment  │  │  ticket  │  │  email   │ │analytics │
        │ worker   │  │  worker  │  │  worker  │ │ worker   │
        └──────────┘  └──────────┘  └──────────┘ └──────────┘
```

### Outbox Pattern

Rather than publishing directly to SNS from within HTTP handlers (which would create a distributed transaction), the API server writes events to an `outbox` table in the same database transaction as the business data. A dedicated `outbox-publisher` process polls the table and forwards pending events to SNS. This guarantees **at-least-once delivery** — no event is lost even if the publisher crashes mid-flight.

### SNS Fan-out

Each event type maps to one SNS topic. Multiple SQS queues can subscribe to the same topic independently, so adding a new consumer (e.g. a fraud-detection worker) requires zero changes to the publisher — just a new queue and subscription.

### SQS Per-Worker Queues

Each worker has its own SQS queue. This means:
- Worker failures don't affect other workers
- Each worker scales independently
- Unprocessed messages land in a per-worker dead-letter queue (DLQ) after 5 failed attempts

---

## Event Catalogue

| Event | SNS Topic | Consumed by |
|---|---|---|
| `order.created` | `order-created` | payment-worker, analytics-worker |
| `order.payment_confirmed` | `order-payment-confirmed` | ticket-worker |
| `order.confirmed` | `order-confirmed` | notification-worker, analytics-worker |
| `order.cancelled` | `order-cancelled` | analytics-worker |
| `order.completed` | `order-completed` | — |
| `reservation.expired` | `reservation-expired` | — |
| `inventory.released` | `inventory-released` | waitlist-worker |
| `event.cancelled` | `event-cancelled` | event-cancellation-worker, analytics-worker |
| `payment.completed` | `payment-completed` | ticket-worker, analytics-worker |
| `tickets.generated` | `tickets-generated` | ticket-email-worker, analytics-worker |
| `notification.email.requested` | `notification-email-requested` | email-worker |
| `refund.approved` | `refund-approved` | refund-processor-worker |
| `refund.completed` | `refund-completed` | analytics-worker |

---

## Workers

| Worker | Queue | Responsibility |
|---|---|---|
| `outbox-publisher` | — | Polls outbox table, publishes to SNS |
| `payment-worker` | `payment-worker` | Confirms offline orders immediately |
| `ticket-worker` | `ticket-worker` | Generates tickets after payment |
| `ticket-email-worker` | `ticket-email-worker` | Generates PDF tickets and emails them |
| `email-worker` | `email-worker` | Sends transactional emails from outbox |
| `notification-worker` | `notification-worker` | Sends order confirmation emails |
| `event-cancellation-worker` | `event-cancellation-worker` | Cascades event cancellation to orders and tickets |
| `refund-processor-worker` | `refund-processor-worker` | Calls Intasend to execute approved refunds |
| `reservation-expiry-worker` | — | Timed poll — expires stale reservations |
| `waitlist-worker` | `waitlist-worker` | Notifies waitlisted users when inventory is released |
| `analytics-worker` | 6 queues (one per topic) | Updates per-event daily stats |

---

## AWS Infrastructure

Provisioned by Terraform in `terraform/`:

| Resource | Type | Purpose |
|---|---|---|
| VPC, subnets, SGs | `vpc.tf` | Network isolation |
| 13 SNS topics | `sns.tf` | Event bus |
| 14 SQS queues + 14 DLQs | `sqs.tf` | Per-worker message queues |
| Aurora PostgreSQL cluster | `rds.tf` | Primary datastore |
| ElastiCache Redis | `elasticache.tf` | Session cache, rate limiting |

---

## Project Structure

```
ticketing_system/
├── cmd/
│   ├── api-server/                  # HTTP API entry point
│   ├── outbox-publisher/            # SNS event publisher
│   ├── payment-worker/              # Processes order.created
│   ├── ticket-worker/               # Generates tickets after payment
│   ├── ticket-email-worker/         # Emails PDF tickets
│   ├── email-worker/                # Sends outbox emails
│   ├── notification-worker/         # Order confirmation emails
│   ├── event-cancellation-worker/   # Cascades event cancellations
│   ├── refund-processor-worker/     # Executes refunds via Intasend
│   ├── reservation-expiry-worker/   # Expires stale reservations
│   ├── waitlist-worker/             # Waitlist notifications
│   └── analytics-worker/            # Event stats updates
│
├── internal/
│   ├── messaging/                   # SNS publisher + SQS consumer (replaces kafka/)
│   │   ├── client.go                # AWS client factory (supports LocalStack)
│   │   ├── sns_publisher.go         # Implements outbox.Sender for SNS
│   │   ├── sqs_consumer.go          # Long-polling consumer with SNS envelope unwrap
│   │   └── topics.go                # Topic name constants
│   ├── config/                      # Config loader (includes MessagingConfig)
│   ├── outbox/                      # Transactional outbox (unchanged from v2)
│   ├── cache/                       # Redis client (points at ElastiCache)
│   ├── database/                    # GORM + Aurora connection
│   ├── auth/                        # JWT authentication, 2FA
│   ├── events/                      # Event management
│   ├── orders/                      # Order lifecycle
│   ├── tickets/                     # Ticket generation, PDF, QR codes
│   ├── payments/                    # Stripe, Intasend, webhooks
│   ├── refunds/                     # Refund workflows
│   ├── inventory/                   # Capacity and reservations
│   ├── notifications/               # Email templates and delivery
│   ├── analytics/                   # Sales and attendance metrics
│   └── ...
│
├── terraform/                       # AWS infrastructure as code
│   ├── main.tf                      # Provider and backend config
│   ├── variables.tf                 # All input variables
│   ├── outputs.tf                   # ARNs and URLs for .env
│   ├── vpc.tf                       # VPC, subnets, security groups
│   ├── sns.tf                       # SNS topics
│   ├── sqs.tf                       # SQS queues, DLQs, subscriptions
│   ├── rds.tf                       # Aurora PostgreSQL cluster
│   └── elasticache.tf               # ElastiCache Redis
│
├── scripts/
│   └── localstack-init.sh           # Creates SNS/SQS resources in LocalStack
│
├── migrations/                      # SQL schema migrations
├── docker-compose.yml               # v2-style local stack (Kafka + Postgres + Redis)
├── docker-compose.localstack.yml    # v3 local stack (LocalStack + Postgres + Redis)
├── docker-compose.monitoring.yml    # Prometheus + Grafana
├── .env.example                     # All required environment variables
└── Dockerfile                       # Container build
```

---

## Local Development

Local development uses [LocalStack](https://localstack.cloud/) to emulate SNS and SQS without an AWS account. Postgres and Redis run as plain Docker containers.

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- AWS CLI (`brew install awscli` or equivalent)

### 1. Start the local stack

```bash
docker compose -f docker-compose.localstack.yml up -d
```

This starts LocalStack (SNS + SQS + S3), PostgreSQL 15, and Redis 7.

### 2. Create SNS topics and SQS queues

```bash
bash scripts/localstack-init.sh
```

The script prints all the `SNS_ARN_*` and `SQS_URL_*` values you need for your `.env`.

### 3. Configure environment

```bash
cp .env.example .env
# Paste the output from localstack-init.sh
# Set AWS_ENDPOINT_URL=http://localhost:4566
# Set AWS_ACCESS_KEY_ID=test
# Set AWS_SECRET_ACCESS_KEY=test
```

### 4. Run the API server

```bash
go run cmd/api-server/main.go
```

### 5. Run workers (each in a separate terminal)

```bash
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

---

## AWS Deployment

### Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.6
- AWS CLI configured with sufficient IAM permissions
- An S3 bucket for Terraform state (optional but recommended)

### 1. Provision infrastructure

```bash
cd terraform

terraform init
terraform plan -var="db_password=<your-secure-password>"
terraform apply -var="db_password=<your-secure-password>"
```

### 2. Export environment variables

```bash
terraform output -json sns_arns
terraform output -json sqs_queue_urls
terraform output db_endpoint
terraform output redis_endpoint
```

Copy these into your `.env` (or your deployment platform's secret store). The `outputs.tf` file documents each output and which env var it maps to.

### 3. Run database migrations

```bash
go run cmd/check-migration/main.go
```

### 4. Deploy workers

Each worker is a stateless Go binary. Deploy them as ECS tasks, Lambda functions, EC2 instances, or any container runtime — they only need the `.env` values and outbound access to AWS endpoints.

For a quick start, build and push to ECR:

```bash
docker build -t ticketing-api .
docker tag ticketing-api <account>.dkr.ecr.<region>.amazonaws.com/ticketing-api:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/ticketing-api:latest
```

---

## Environment Variables

Full reference is in [.env.example](.env.example). Key groups:

### AWS Credentials
```bash
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=us-east-1
AWS_ENDPOINT_URL=           # LocalStack only: http://localhost:4566
```

### Database (RDS Aurora)
```bash
DB_HOST=                    # terraform output db_endpoint
DB_PORT=5432
DB_USER=ticketing_admin
DB_PASSWORD=
DB_NAME=ticketing_system
DB_SSL_MODE=require         # use "disable" for LocalStack
```

### Cache (ElastiCache)
```bash
REDIS_ADDR=                 # terraform output redis_endpoint
REDIS_PASSWORD=
REDIS_ENABLED=true
```

### SNS Topics (one per event type)
```bash
SNS_ARN_ORDER_CREATED=
SNS_ARN_ORDER_PAYMENT_CONFIRMED=
SNS_ARN_ORDER_CONFIRMED=
SNS_ARN_ORDER_CANCELLED=
SNS_ARN_ORDER_COMPLETED=
SNS_ARN_RESERVATION_EXPIRED=
SNS_ARN_INVENTORY_RELEASED=
SNS_ARN_EVENT_CANCELLED=
SNS_ARN_PAYMENT_COMPLETED=
SNS_ARN_TICKETS_GENERATED=
SNS_ARN_NOTIFICATION_EMAIL=
SNS_ARN_REFUND_APPROVED=
SNS_ARN_REFUND_COMPLETED=
```

### SQS Queue URLs (one per worker)
```bash
SQS_URL_PAYMENT_WORKER=
SQS_URL_TICKET_WORKER=
SQS_URL_TICKET_EMAIL_WORKER=
SQS_URL_EMAIL_WORKER=
SQS_URL_NOTIFICATION_WORKER=
SQS_URL_EVENT_CANCELLATION_WORKER=
SQS_URL_REFUND_PROCESSOR_WORKER=
SQS_URL_WAITLIST_WORKER=
SQS_URL_ANALYTICS_ORDER_CREATED=
SQS_URL_ANALYTICS_ORDER_CONFIRMED=
SQS_URL_ANALYTICS_PAYMENT_COMPLETED=
SQS_URL_ANALYTICS_ORDER_CANCELLED=
SQS_URL_ANALYTICS_EVENT_CANCELLED=
SQS_URL_ANALYTICS_TICKETS_GENERATED=
```

---

## API Reference

### Authentication
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
POST   /api/auth/reset-password
POST   /api/auth/2fa/enable
POST   /api/auth/2fa/verify
```

### Events
```
GET    /api/events
GET    /api/events/:id
POST   /api/events                       # organizer
PUT    /api/events/:id                   # organizer
DELETE /api/events/:id                   # organizer
POST   /api/organizers/events/:id/publish
```

### Orders & Tickets
```
POST   /api/orders                       # create order
GET    /api/orders/:id
POST   /api/orders/:id/payment
GET    /api/tickets/my
GET    /api/tickets/:id/pdf
POST   /api/tickets/:id/checkin
POST   /api/tickets/:id/transfer
```

### Refunds & Payments
```
POST   /api/refunds
GET    /api/refunds/:id
POST   /api/payments/webhook/intasend
POST   /api/payments/webhook/stripe
```

### Analytics (organizer)
```
GET    /api/analytics/events/:id/summary
GET    /api/analytics/events/:id/sales
GET    /api/analytics/events/:id/audience
GET    /api/organizers/:id/dashboard
```

---

## Monitoring

Prometheus metrics and Grafana dashboards are unchanged from v2:

```bash
docker compose -f docker-compose.monitoring.yml up -d
# Prometheus: http://localhost:9090
# Grafana:    http://localhost:3001  (admin / admin123)
```

---

## Key Design Decisions

**SNS over MSK (managed Kafka)**
MSK would preserve the exact Kafka API from v2 but costs significantly more and still requires cluster sizing decisions. SNS/SQS is fully serverless, cheaper at low-to-medium throughput, and demonstrates a meaningfully different technology than v2.

**Outbox pattern retained**
The outbox table remains the source of truth for event delivery. SNS is called only from the `outbox-publisher`, never directly from HTTP handlers. This means a SNS outage or a publisher crash never corrupts business data — the outbox is simply drained when the publisher recovers.

**SQS envelope unwrapping**
When SNS delivers to SQS, it wraps the payload in a JSON envelope. The `SQSConsumer.ReceiveMessage` method unwraps this transparently, so worker handlers receive the raw event payload identically to how they received Kafka message values in v2.

**Per-worker dead-letter queues**
Each SQS queue has a DLQ with a 14-day retention. Messages that fail 5 times are moved there automatically, giving operators time to inspect and replay them without losing data.

---

## Technology Stack

| Category | v2 | v3 |
|---|---|---|
| Language | Go | Go |
| HTTP router | gorilla/mux | gorilla/mux |
| ORM | GORM + PostgreSQL | GORM + Aurora PostgreSQL |
| Messaging | Kafka (kafka-go) | SNS + SQS (aws-sdk-go-v2) |
| Cache | Redis (go-redis) | ElastiCache (go-redis) |
| Storage | AWS S3 | AWS S3 |
| PDF | gofpdf | gofpdf |
| QR codes | go-qrcode | go-qrcode |
| Payments | Intasend, Stripe | Intasend, Stripe |
| Email | Brevo API | Brevo API |
| Monitoring | Prometheus + Grafana | Prometheus + Grafana |
| Infrastructure | Docker Compose | Terraform + AWS |

---

## License

MIT

---

Built with Go — [kamausimon217@gmail.com](mailto:kamausimon217@gmail.com)
