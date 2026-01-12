# Database Schema Documentation

## Overview
The ticketing system uses **PostgreSQL** as the primary database with **GORM** as the ORM. The database follows a normalized relational design with proper foreign keys and indexes for optimal performance.

## Entity Relationship Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         CORE ENTITIES                                         │
└──────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐
│     Users       │
├─────────────────┤
│ id (PK)         │◄──────────────┐
│ email           │                │
│ password_hash   │                │
│ first_name      │                │
│ last_name       │                │
│ phone_number    │                │
│ role (enum)     │                │         ┌────────────────┐
│ is_active       │                │         │  Organizers    │
│ email_verified  │                │         ├────────────────┤
│ last_login_at   │                │         │ id (PK)        │
│ created_at      │                ├─────────┤ user_id (FK)   │
│ updated_at      │                │         │ account_id(FK) │
│ deleted_at      │                │         │ name           │
└─────────────────┘                │         │ about          │
         │                         │         │ logo_path      │
         │                         │         │ website        │
         │                         │         │ status         │
         │                         │         │ created_at     │
         │                         │         └────────────────┘
         │                         │                  │
         │                         │                  │
         │                         │                  │
         │                         │                  ▼
         │                         │         ┌────────────────┐
         │                         │         │    Events      │
         │                         │         ├────────────────┤
         │                         │         │ id (PK)        │
         │                         │         │ organizer_id(FK)│
         │                         │         │ account_id(FK) │
         │                         │         │ title          │
         │                         │         │ description    │
         │                         │         │ location       │
         │                         │         │ start_date     │
         │                         │         │ end_date       │
         │                         │         │ on_sale_date   │
         │                         │         │ status (enum)  │
         │                         │         │ category (enum)│
         │                         │         │ currency       │
         │                         │         │ max_capacity   │
         │                         │         │ is_live        │
         │                         │         │ is_private     │
         │                         │         │ version (lock) │
         │                         │         │ created_at     │
         │                         │         └────────────────┘
         │                         │                  │
         │                         │                  │
         │                         │         ┌────────┴──────────┐
         │                         │         │                   │
         │                         │         ▼                   ▼
         │                         │  ┌──────────────┐   ┌──────────────┐
         │                         │  │EventImages   │   │TicketClasses │
         │                         │  ├──────────────┤   ├──────────────┤
         │                         │  │ id (PK)      │   │ id (PK)      │
         │                         │  │ event_id(FK) │   │ event_id(FK) │
         │                         │  │ image_path   │   │ name         │
         │                         │  │ display_order│   │ description  │
         │                         │  │ created_at   │   │ price        │
         │                         │  └──────────────┘   │ quantity     │
         │                         │                     │ sold         │
         │                         │                     │ version(lock)│
         │                         │                     │ created_at   │
         │                         │                     └──────────────┘
         │                         │                            │
         │                         │                            │
         ▼                         │                            │
┌─────────────────┐               │                            │
│     Orders      │◄──────────────┘                            │
├─────────────────┤                                            │
│ id (PK)         │                                            │
│ user_id (FK)    │                                            │
│ event_id (FK)   │                                            │
│ order_number    │                                            │
│ total_amount    │                                            │
│ status (enum)   │                                            │
│ payment_status  │                                            │
│ payment_method  │                                            │
│ currency        │                                            │
│ created_at      │                                            │
└─────────────────┘                                            │
         │                                                     │
         │                                                     │
         ▼                                                     │
┌─────────────────┐                                           │
│   OrderItems    │◄──────────────────────────────────────────┘
├─────────────────┤
│ id (PK)         │
│ order_id (FK)   │
│ ticket_class_id │
│ quantity        │
│ unit_price      │
│ subtotal        │
│ created_at      │
└─────────────────┘
         │
         │
         ▼
┌─────────────────┐
│    Tickets      │
├─────────────────┤
│ id (PK)         │
│ order_item_id(FK)│
│ ticket_number   │
│ qr_code         │
│ holder_name     │
│ holder_email    │
│ status (enum)   │
│ checked_in_at   │
│ used_at         │
│ created_at      │
└─────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                      PAYMENT & FINANCIAL                                      │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│ PaymentTransactions│
├───────────────────┤
│ id (PK)           │
│ order_id (FK)     │
│ transaction_id    │
│ provider          │
│ amount            │
│ currency          │
│ status            │
│ payment_method    │
│ response_data     │
│ created_at        │
└───────────────────┘
         │
         │
         ▼
┌───────────────────┐
│  RefundRecords    │
├───────────────────┤
│ id (PK)           │
│ order_id (FK)     │
│ transaction_id(FK)│
│ refund_amount     │
│ refund_reason     │
│ status            │
│ processed_by (FK) │
│ processed_at      │
│ created_at        │
└───────────────────┘
         │
         │
         ▼
┌───────────────────┐
│ RefundLineItems   │
├───────────────────┤
│ id (PK)           │
│ refund_record_id  │
│ order_item_id (FK)│
│ ticket_id (FK)    │
│ amount            │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│ SettlementRecords │
├───────────────────┤
│ id (PK)           │
│ organizer_id (FK) │
│ account_id (FK)   │
│ period_start      │
│ period_end        │
│ total_sales       │
│ platform_fee      │
│ net_amount        │
│ status            │
│ payout_date       │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│  PayoutAccounts   │
├───────────────────┤
│ id (PK)           │
│ organizer_id (FK) │
│ account_id (FK)   │
│ bank_name         │
│ account_number(enc)│
│ account_name      │
│ swift_code        │
│ is_default        │
│ is_verified       │
│ created_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY & AUTH                                       │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│ EmailVerification │
├───────────────────┤
│ id (PK)           │
│ user_id (FK)      │
│ token             │
│ expires_at        │
│ verified_at       │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│  PasswordResets   │
├───────────────────┤
│ id (PK)           │
│ user_id (FK)      │
│ token             │
│ expires_at        │
│ used_at           │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│   TwoFactorAuth   │
├───────────────────┤
│ id (PK)           │
│ user_id (FK)      │
│ secret            │
│ is_enabled        │
│ backup_codes      │
│ last_used_at      │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│ AccountActivity   │
├───────────────────┤
│ id (PK)           │
│ account_id (FK)   │
│ user_id (FK)      │
│ action            │
│ category          │
│ description       │
│ ip_address        │
│ user_agent        │
│ metadata          │
│ created_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                      INVENTORY & RESERVATIONS                                 │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│ ReservedTickets   │
├───────────────────┤
│ id (PK)           │
│ event_id (FK)     │
│ ticket_class_id   │
│ session_id        │
│ user_id (FK)      │
│ quantity          │
│ expires_at        │
│ status            │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│  WaitlistEntries  │
├───────────────────┤
│ id (PK)           │
│ event_id (FK)     │
│ ticket_class_id   │
│ email             │
│ name              │
│ quantity          │
│ status            │
│ priority          │
│ notified_at       │
│ created_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         NOTIFICATIONS & COMMS                                 │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐
│NotificationPreferences│
├──────────────────────┤
│ id (PK)              │
│ user_id (FK)         │
│ email_marketing      │
│ email_order_updates  │
│ email_event_updates  │
│ sms_notifications    │
│ push_notifications   │
│ created_at           │
└──────────────────────┘

┌───────────────────┐
│   WebhookLogs     │
├───────────────────┤
│ id (PK)           │
│ event_type        │
│ payload           │
│ status            │
│ attempts          │
│ last_attempt_at   │
│ created_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         SUPPORT & AI                                          │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│  SupportTickets   │
├───────────────────┤
│ id (PK)           │
│ user_id (FK)      │
│ ticket_number     │
│ subject           │
│ category          │
│ priority          │
│ status            │
│ assigned_to (FK)  │
│ conversation_hist │
│ created_at        │
│ updated_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                      ANALYTICS & METRICS                                      │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│   EventStats      │
├───────────────────┤
│ id (PK)           │
│ event_id (FK)     │
│ views             │
│ unique_visitors   │
│ tickets_sold      │
│ revenue           │
│ conversion_rate   │
│ date              │
│ created_at        │
└───────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                      CONFIGURATION & SETTINGS                                 │
└──────────────────────────────────────────────────────────────────────────────┘

┌───────────────────┐
│    Accounts       │
├───────────────────┤
│ id (PK)           │
│ name              │
│ subdomain         │
│ custom_domain     │
│ settings (JSON)   │
│ timezone_id (FK)  │
│ currency_id (FK)  │
│ date_format_id    │
│ is_active         │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│   Currencies      │
├───────────────────┤
│ id (PK)           │
│ code (USD, KSH)   │
│ name              │
│ symbol            │
│ exchange_rate     │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│   Timezones       │
├───────────────────┤
│ id (PK)           │
│ name              │
│ offset            │
│ display_name      │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│  DateTimeFormats  │
├───────────────────┤
│ id (PK)           │
│ format            │
│ display_name      │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│   Promotions      │
├───────────────────┤
│ id (PK)           │
│ event_id (FK)     │
│ code              │
│ discount_type     │
│ discount_value    │
│ max_uses          │
│ used_count        │
│ valid_from        │
│ valid_until       │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│     Venues        │
├───────────────────┤
│ id (PK)           │
│ name              │
│ address           │
│ city              │
│ country           │
│ capacity          │
│ created_at        │
└───────────────────┘

┌───────────────────┐
│   EventVenues     │
├───────────────────┤
│ id (PK)           │
│ event_id (FK)     │
│ venue_id (FK)     │
│ created_at        │
└───────────────────┘
```

## Table Descriptions

### Core Tables

#### **users**
Primary user table for authentication and profile management.
- **Key Fields**: email (unique), password_hash, role (customer/organizer/admin/support)
- **Indexes**: email, role, created_at
- **Soft Deletes**: Yes

#### **organizers**
Extended profile for users with organizer role.
- **Relationships**: 
  - Belongs to User (one-to-one)
  - Has many Events
  - Belongs to Account
- **Key Fields**: name, about, logo_path, status, website

#### **accounts**
Multi-tenancy support for white-label ticketing platforms.
- **Features**: Custom domain, subdomain, timezone, currency
- **Settings**: JSON field for flexible configuration
- **Relationships**: Has many Users, Organizers, Events

#### **events**
Core event entity with all event details.
- **Key Fields**: 
  - title, description, location
  - start_date, end_date, on_sale_date
  - status (draft/live/cancelled/completed)
  - category (music/conference/sports/art/etc)
  - version (for optimistic locking)
- **Indexes**: organizer_id, status, start_date, category
- **Relationships**:
  - Belongs to Organizer
  - Belongs to Account
  - Has many EventImages
  - Has many TicketClasses
  - Has many Orders

#### **event_images**
Multiple images per event with ordering.
- **Key Fields**: image_path (S3 URL or local path), display_order
- **Relationships**: Belongs to Event

#### **ticket_classes**
Different ticket types per event (VIP, General, Early Bird, etc).
- **Key Fields**: 
  - name, description, price
  - quantity (total), sold (current)
  - version (optimistic locking for concurrency)
- **Concurrency**: Protected by version field to prevent overselling
- **Relationships**: 
  - Belongs to Event
  - Has many OrderItems
  - Has many ReservedTickets

### Order & Ticketing Tables

#### **orders**
Purchase orders from users.
- **Key Fields**: 
  - order_number (unique)
  - total_amount, currency
  - status (pending/paid/fulfilled/cancelled/refunded)
  - payment_status, payment_method
- **Indexes**: user_id, event_id, status, order_number
- **Relationships**:
  - Belongs to User
  - Belongs to Event
  - Has many OrderItems
  - Has many PaymentTransactions

#### **order_items**
Line items within an order (different ticket classes).
- **Key Fields**: quantity, unit_price, subtotal
- **Relationships**:
  - Belongs to Order
  - Belongs to TicketClass
  - Has many Tickets

#### **tickets**
Individual tickets issued to attendees.
- **Key Fields**:
  - ticket_number (unique identifier)
  - qr_code (for check-in)
  - holder_name, holder_email
  - status (active/used/cancelled/refunded)
  - checked_in_at, used_at
- **Indexes**: ticket_number, qr_code, status
- **Relationships**: Belongs to OrderItem

### Payment Tables

#### **payment_transactions**
Payment gateway transactions (IntaSend, Stripe, etc).
- **Key Fields**:
  - transaction_id (from gateway)
  - provider (intasend/stripe)
  - amount, currency
  - status (pending/completed/failed)
  - payment_method (mpesa/card)
  - response_data (JSON - webhook payload)
- **Relationships**: Belongs to Order

#### **refund_records**
Refund requests and processing.
- **Key Fields**:
  - refund_amount, refund_reason
  - status (pending/approved/rejected/processed)
  - processed_by (admin user_id)
  - processed_at
- **Relationships**:
  - Belongs to Order
  - Has many RefundLineItems

#### **refund_line_items**
Individual items being refunded.
- **Relationships**:
  - Belongs to RefundRecord
  - Belongs to OrderItem
  - Belongs to Ticket

#### **settlement_records**
Payout calculations for organizers.
- **Key Fields**:
  - period_start, period_end
  - total_sales, platform_fee, net_amount
  - status (pending/processing/completed)
  - payout_date
- **Relationships**: 
  - Belongs to Organizer
  - Belongs to Account

#### **payout_accounts**
Organizer bank account details for settlements.
- **Key Fields**:
  - bank_name, account_number (encrypted)
  - account_name, swift_code
  - is_default, is_verified
- **Security**: account_number is encrypted with AES-256-GCM
- **Relationships**: Belongs to Organizer

### Inventory Management

#### **reserved_tickets**
Temporary reservations (30-minute holds).
- **Key Fields**:
  - session_id (anonymous users)
  - user_id (logged-in users)
  - quantity
  - expires_at
  - status (active/expired/converted/cancelled)
- **Background Job**: Cleanup job runs every 5 minutes to release expired reservations
- **Relationships**:
  - Belongs to Event
  - Belongs to TicketClass

#### **waitlist_entries**
Users waiting for sold-out events.
- **Key Fields**:
  - email, name, quantity
  - status (waiting/notified/converted/expired)
  - priority (higher = notified first)
  - notified_at
- **Relationships**: Belongs to Event

### Security Tables

#### **email_verification**
Email verification tokens for new users.
- **Key Fields**: token (unique), expires_at, verified_at
- **Lifecycle**: Created on registration, deleted after verification
- **Relationships**: Belongs to User

#### **password_resets**
Password reset tokens.
- **Key Fields**: token (unique), expires_at, used_at
- **Security**: Tokens expire after 1 hour
- **Relationships**: Belongs to User

#### **two_factor_auth**
2FA configuration for users.
- **Key Fields**:
  - secret (TOTP secret)
  - is_enabled
  - backup_codes (encrypted array)
  - last_used_at
- **Relationships**: Belongs to User

#### **account_activity**
Audit log of user actions.
- **Key Fields**:
  - action, category
  - description
  - ip_address, user_agent
  - metadata (JSON)
- **Indexes**: account_id, user_id, action, created_at
- **Relationships**: Belongs to Account and User

### Configuration Tables

#### **currencies**
Supported currencies with exchange rates.
- **Key Fields**: code (USD, KSH, EUR), symbol, exchange_rate

#### **timezones**
Available timezones for accounts.
- **Key Fields**: name (Africa/Nairobi), offset, display_name

#### **date_time_formats**
Date/time display formats.
- **Key Fields**: format (YYYY-MM-DD HH:mm), display_name

#### **promotions**
Discount codes and promotions.
- **Key Fields**:
  - code (unique)
  - discount_type (percentage/fixed)
  - discount_value
  - max_uses, used_count
  - valid_from, valid_until
- **Relationships**: Belongs to Event

#### **venues**
Physical venue database.
- **Key Fields**: name, address, city, country, capacity
- **Relationships**: Many-to-many with Events through EventVenues

### Support & AI

#### **support_tickets**
Customer support requests.
- **Key Fields**:
  - ticket_number (unique)
  - subject, category, priority
  - status (open/in_progress/resolved/closed)
  - assigned_to (support agent)
  - conversation_history (JSON - AI context)
- **AI Integration**: Uses OpenAI for automated responses
- **Relationships**: Belongs to User

### Analytics

#### **event_stats**
Daily statistics per event.
- **Key Fields**:
  - views, unique_visitors
  - tickets_sold, revenue
  - conversion_rate
  - date (daily aggregation)
- **Relationships**: Belongs to Event

### Notification Tables

#### **notification_preferences**
User communication preferences (GDPR compliant).
- **Key Fields**:
  - email_marketing, email_order_updates, email_event_updates
  - sms_notifications, push_notifications
- **Relationships**: Belongs to User

#### **webhook_logs**
Webhook delivery tracking (payment gateways).
- **Key Fields**:
  - event_type, payload (JSON)
  - status (pending/success/failed)
  - attempts, last_attempt_at

## Indexes & Performance

### Primary Indexes
- All tables have primary key (id) with auto-increment
- UUID-based identifiers for: order_number, ticket_number, transaction_id

### Foreign Key Indexes
- All foreign keys are indexed for join performance
- Composite indexes on frequently queried combinations:
  - (event_id, status) on orders
  - (user_id, created_at) on orders
  - (organizer_id, start_date) on events

### Unique Constraints
- users.email
- tickets.ticket_number
- tickets.qr_code
- orders.order_number
- promotions.code

## Concurrency Control

### Optimistic Locking
Tables with `version` field:
- **events**: Prevents conflicting updates
- **ticket_classes**: Critical for inventory management (prevents overselling)

### Transaction Isolation
- All order creation uses database transactions
- Inventory checks happen within transaction boundaries
- Rollback on any failure

## Soft Deletes
Most tables use GORM soft deletes (deleted_at column):
- Records are not physically deleted
- Enables audit trails and data recovery
- Filtered automatically by GORM

Tables without soft deletes:
- reservation_tickets (temporary data)
- webhook_logs (append-only)
- metrics tables (time-series data)

## Data Encryption

### Encrypted Fields
- **payout_accounts.account_number**: AES-256-GCM encryption
- **two_factor_auth.backup_codes**: Encrypted before storage
- **users.password_hash**: bcrypt hashing

## Database Migrations

Migrations are managed through GORM AutoMigrate:
```go
DB.AutoMigrate(
    &User{},
    &Organizer{},
    &Event{},
    &TicketClass{},
    &Order{},
    &Ticket{},
    // ... all models
)
```

Manual migrations available in: `/migrations/` directory

## Backup & Recovery

### Backup Strategy
- Daily automated backups (PostgreSQL pg_dump)
- Point-in-time recovery enabled
- 30-day retention policy

### Critical Tables (Priority 1)
- users, orders, tickets, payment_transactions
- payout_accounts, settlement_records

## Future Enhancements

1. **Partitioning**: Partition large tables (orders, tickets) by date
2. **Read Replicas**: Separate read/write databases
3. **Caching Layer**: Redis cache for frequently accessed data
4. **Archive Strategy**: Move old data to archive tables
5. **Full-Text Search**: PostgreSQL full-text search or Elasticsearch

---

**Version**: 1.0  
**Last Updated**: January 2026  
**Total Tables**: 38+
