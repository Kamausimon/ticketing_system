    # Backend Architecture

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │   Web App    │    │  Mobile App  │    │  Admin Panel │                  │
│  │  (React/JS)  │    │   (Future)   │    │   (Future)   │                  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                  │
│         │                   │                    │                           │
│         └───────────────────┴────────────────────┘                           │
│                             │                                                │
│                       REST API / HTTPS                                       │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                              ▼                                                │
│                    ┌──────────────────┐                                      │
│                    │   API Gateway    │                                      │
│                    │  (Port 8080)     │                                      │
│                    │   Gorilla Mux    │                                      │
│                    └────────┬─────────┘                                      │
│                             │                                                │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                   MIDDLEWARE LAYER                                           │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌──────────────┬───────────┼───────────┬──────────────┐                    │
│  │              │           │           │              │                    │
│  ▼              ▼           ▼           ▼              ▼                    │
│ ┌────┐    ┌─────────┐  ┌──────┐  ┌──────────┐  ┌──────────┐               │
│ │CORS│    │  Auth   │  │ Rate │  │  Email   │  │Prometheus│               │
│ │    │    │  JWT    │  │Limit │  │Verify    │  │ Metrics  │               │
│ └────┘    └─────────┘  └──────┘  └──────────┘  └──────────┘               │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                   APPLICATION LAYER (Go Handlers)                            │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌──────────┬───────────────┼─────────────┬──────────────┬──────────┐      │
│  │          │               │             │              │          │      │
│  ▼          ▼               ▼             ▼              ▼          ▼      │
│ ┌────┐  ┌───────┐  ┌────────────┐  ┌──────────┐  ┌────────┐  ┌────────┐  │
│ │Auth│  │Events │  │   Orders   │  │ Tickets  │  │Payments│  │Organiz.│  │
│ │    │  │       │  │            │  │          │  │        │  │        │  │
│ └────┘  └───────┘  └────────────┘  └──────────┘  └────────┘  └────────┘  │
│                                                                              │
│ ┌─────────┐  ┌────────┐  ┌──────────┐  ┌────────┐  ┌──────────┐          │
│ │Analytics│  │Refunds │  │Inventory │  │Support │  │Settlement│          │
│ │         │  │        │  │          │  │ (AI)   │  │          │          │
│ └─────────┘  └────────┘  └──────────┘  └────────┘  └──────────┘          │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                      BUSINESS LOGIC LAYER                                    │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌───────────────────────────┼──────────────────────────────┐               │
│  │  Core Services            │                              │               │
│  │                           ▼                              │               │
│  │  ┌──────────────────────────────────────────────┐       │               │
│  │  │  • Transaction Management (ACID)             │       │               │
│  │  │  • Concurrency Control (Optimistic Locking)  │       │               │
│  │  │  • Reservation System (30-min timeout)       │       │               │
│  │  │  • Capacity Management                        │       │               │
│  │  │  • Bank Account Encryption (AES-256)         │       │               │
│  │  │  • 2FA (TOTP)                                │       │               │
│  │  │  • Email Verification                         │       │               │
│  │  └──────────────────────────────────────────────┘       │               │
│  └──────────────────────────────────────────────────────────┘               │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                      DATA ACCESS LAYER (GORM)                                │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌──────────────────────────▼─────────────────────────────┐                 │
│  │  Database Models & Repositories                        │                 │
│  │  • Users, Events, Tickets, Orders, Payments            │                 │
│  │  • Organizers, Venues, Promotions                      │                 │
│  │  • Analytics, Support Tickets, AI Context             │                 │
│  └────────────────────────────────────────────────────────┘                 │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                      STORAGE & CACHE LAYER                                   │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌──────────────┬───────────┴──────────┬──────────────┐                     │
│  │              │                      │              │                     │
│  ▼              ▼                      ▼              ▼                     │
│ ┌──────────┐  ┌─────────┐       ┌─────────┐    ┌──────────┐               │
│ │PostgreSQL│  │  Redis  │       │  AWS S3 │    │  Local   │               │
│ │          │  │         │       │         │    │  Storage │               │
│ │ Primary  │  │ Session │       │ Images  │    │ Fallback │               │
│ │   DB     │  │ Cache   │       │  Files  │    │          │               │
│ └──────────┘  └─────────┘       └─────────┘    └──────────┘               │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                      EXTERNAL SERVICES                                       │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│  ┌──────────────┬───────────┴──────────┬──────────────┬─────────────┐      │
│  │              │                      │              │             │      │
│  ▼              ▼                      ▼              ▼             ▼      │
│ ┌────────┐  ┌─────────┐        ┌─────────┐    ┌──────────┐  ┌─────────┐  │
│ │IntaSend│  │  SMTP   │        │ OpenAI  │    │Prometheus│  │ Grafana │  │
│ │        │  │         │        │         │    │          │  │         │  │
│ │Payment │  │ Email   │        │AI Supp. │    │ Metrics  │  │Dashboard│  │
│ │Gateway │  │Service  │        │         │    │          │  │         │  │
│ └────────┘  └─────────┘        └─────────┘    └──────────┘  └─────────┘  │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                      BACKGROUND JOBS                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌──────────────────────────────────────────────────────────┐               │
│  │  • Reservation Cleanup (Every 5 minutes)                 │               │
│  │  • Expired Reservation Release                           │               │
│  │  • Notification Queue Processing                         │               │
│  └──────────────────────────────────────────────────────────┘               │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

## Key Architecture Features

### 1. **Layered Architecture**
- **Separation of Concerns**: Each layer has specific responsibilities
- **Maintainability**: Easy to modify individual layers
- **Testability**: Each layer can be tested independently

### 2. **Middleware Stack**
- **CORS**: Cross-origin resource sharing for web clients
- **Authentication**: JWT-based token validation
- **Rate Limiting**: IP-based request throttling (Redis-backed)
- **Email Verification**: Enforce verified email for sensitive operations
- **Metrics**: Prometheus instrumentation for all operations

### 3. **Core Business Features**

#### Security & Authentication
- JWT authentication with secure token management
- 2FA with TOTP (Time-based One-Time Password)
- Email verification workflow
- Bank account encryption (AES-256-GCM)
- Password reset with secure tokens

#### Event Management
- Event creation and management
- Image uploads (S3 + local fallback)
- Venue management
- Category-based organization
- Search and filtering

#### Ticketing System
- Real-time inventory management
- Optimistic locking for concurrency
- 30-minute reservation system
- QR code generation
- PDF ticket generation
- Email delivery

#### Payment Processing
- IntaSend integration (M-Pesa, Card)
- Webhook handling
- Transaction atomicity
- Refund processing
- Settlement calculations

#### Analytics & Monitoring
- Prometheus metrics collection
- Grafana dashboards
- Event performance tracking
- Sales analytics
- System health monitoring

#### AI-Powered Support
- OpenAI integration
- Context-aware responses
- Ticket management
- Conversation history

### 4. **Data Storage Strategy**

#### PostgreSQL (Primary Database)
- All transactional data
- ACID compliance
- Relational integrity
- GORM ORM for type-safe queries

#### Redis (Session & Cache)
- Session management
- Rate limiting counters
- Hot data caching
- In-memory fallback available

#### AWS S3 (File Storage)
- Event images
- Organizer logos
- Generated PDFs
- Local filesystem fallback

### 5. **Reliability Features**

#### Concurrency Control
- Optimistic locking with version field
- Transaction isolation
- Race condition prevention
- Inventory depletion protection

#### Error Handling
- Graceful degradation (S3 → Local storage)
- Comprehensive error responses
- Rollback mechanisms
- Retry logic for external services

#### Background Jobs
- Automatic reservation cleanup
- Expired ticket release
- System maintenance tasks

### 6. **Scalability Considerations**

#### Current Implementation
- Stateless API design
- Horizontal scaling ready
- Session management via Redis
- Database connection pooling

#### Future Enhancements
- Load balancer integration
- Database read replicas
- CDN for static assets
- Message queue (RabbitMQ/Kafka)
- Microservices migration path

### 7. **Observability**

#### Metrics (Prometheus)
- Request rates and latencies
- Error rates
- Database query performance
- Cache hit/miss ratios
- Business metrics (sales, bookings)

#### Logging
- Structured logging
- Request tracing
- Error tracking
- Audit trails

#### Monitoring (Grafana)
- Real-time dashboards
- Alert configuration
- Performance visualization
- Business intelligence

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gorilla Mux (HTTP router)
- **ORM**: GORM
- **Authentication**: JWT-Go
- **2FA**: pquerna/otp

### Storage
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **File Storage**: AWS S3
- **Session Store**: Redis

### External Services
- **Payment**: IntaSend API
- **Email**: SMTP (Gmail, SendGrid, etc.)
- **AI**: OpenAI GPT-4
- **Monitoring**: Prometheus + Grafana

### DevOps
- **Containerization**: Docker
- **Container Orchestration**: Docker Compose
- **CI/CD**: GitHub Actions (ready)
- **Deployment**: Railway, AWS, or VPS

## API Design

### RESTful Principles
- Resource-based URLs
- HTTP methods (GET, POST, PUT, DELETE)
- JSON request/response bodies
- Proper status codes
- Pagination for list endpoints

### Rate Limiting
- Per-endpoint rate limits
- IP-based throttling
- Configurable limits
- Graceful error responses

### Versioning
- URL-based (future: `/v1/`, `/v2/`)
- Backward compatibility
- Deprecation notices

## Security Measures

1. **Authentication & Authorization**
   - JWT tokens with expiration
   - Role-based access control (User, Organizer, Admin)
   - Email verification required
   - 2FA optional for enhanced security

2. **Data Protection**
   - Password hashing (bcrypt)
   - Bank account encryption (AES-256)
   - SQL injection prevention (GORM parameterization)
   - XSS protection
   - CORS configuration

3. **Rate Limiting**
   - Login attempts: 5 per minute
   - API calls: 100 per minute
   - Payment operations: 10 per minute

4. **Audit & Compliance**
   - Soft deletes (deleted_at)
   - Created/Updated timestamps
   - Transaction logging

## Development Workflow

```
Developer → Git Push → Build → Test → Deploy
                         ↓
                  Docker Build
                         ↓
                   API Server
                         ↓
              Database Migration
                         ↓
                    Production
```

## Deployment Architecture

```
┌──────────────────────────────────────────────┐
│            Load Balancer (Future)            │
└──────────────┬───────────────────────────────┘
               │
        ┌──────┴──────┐
        ▼             ▼
    ┌──────┐      ┌──────┐
    │ App  │      │ App  │  (Horizontal Scaling)
    │Server│      │Server│
    └──┬───┘      └──┬───┘
       │             │
       └──────┬──────┘
              │
    ┌─────────┴─────────┐
    │                   │
    ▼                   ▼
┌────────┐        ┌─────────┐
│Database│        │  Redis  │
│Primary │        │ Cluster │
└────────┘        └─────────┘
```

## Performance Characteristics

- **Response Time**: < 100ms for most endpoints
- **Concurrency**: Handles 1000+ concurrent users
- **Throughput**: 10,000+ requests per minute
- **Database**: Connection pooling (max 100 connections)
- **Caching**: Redis for session and hot data

## Future Enhancements

1. **Microservices Architecture**
   - Event Service
   - Payment Service
   - Notification Service
   - Analytics Service

2. **Enhanced Features**
   - Real-time notifications (WebSocket)
   - Mobile app (React Native/Flutter)
   - Admin dashboard
   - Advanced analytics

3. **Infrastructure**
   - Kubernetes orchestration
   - CI/CD pipeline
   - Blue-green deployment
   - Multi-region support

---

**Version**: 1.0  
**Last Updated**: January 2026  
**Maintained By**: Development Team
