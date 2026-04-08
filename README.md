# 🎫 Event Ticketing System

A comprehensive event ticketing platform built with Go, featuring user authentication, event management, ticket purchasing, payment processing, and analytics.

## 🚀 Quick Start - Railway Deployment



### Deploy Backend

1. Go to https://railway.app/new
2. Deploy from GitHub → Select this repo
3. Add PostgreSQL and Redis
4. Configure environment variables
5. Generate domain



### Deploy Demo App

The demo app in `/demo-app` can be deployed separately:

**Option 1 - Railway**: Create new project → Root directory: `/demo-app`  
**Option 2 - Vercel**: `cd demo-app && vercel` (FREE)



---



## 🏗️ Project Structure

```
ticketing_system/
├── Dockerfile                    # Backend Docker build
├── .dockerignore                 # Excludes demo from backend
├── railway.json                  # Railway configuration
│
├── cmd/
│   └── api-server/              # Backend entry point
│       └── main.go
│
├── internal/                     # Backend business logic
│   ├── auth/                    # Authentication
│   ├── events/                  # Event management
│   ├── tickets/                 # Ticket operations
│   ├── payments/                # Payment processing
│   ├── analytics/               # Analytics engine
│   └── ...
│
├── demo-app/                     # Demo frontend (separate)
│   ├── Dockerfile               # Demo Docker build
│   ├── index.html               # Demo UI
│   ├── app.js                   # Demo logic
│   └── styles.css               # Demo styles
│
├── migrations/                   # Database migrations
├── prometheus/                   # Monitoring config
└── grafana/                      # Dashboard config
```

---

## 🎨 Features

### For Attendees
- ✅ Browse and search events
- ✅ Purchase tickets with payment integration
- ✅ Receive email confirmations
- ✅ Download PDF tickets with QR codes
- ✅ View order history
- ✅ Email verification
- ✅ Two-factor authentication

### For Organizers
- ✅ Create and manage events
- ✅ Set ticket classes and pricing
- ✅ Track sales and analytics
- ✅ Manage capacity
- ✅ Process refunds
- ✅ View revenue reports
- ✅ Export data

### For Admins
- ✅ User management
- ✅ System analytics
- ✅ Bulk operations
- ✅ Support system with AI
- ✅ Rate limiting
- ✅ Security monitoring

---

## 🛠️ Technology Stack

- **Backend**: Go 1.25.3
- **Database**: PostgreSQL
- **Cache**: Redis
- **Storage**: AWS S3
- **Payments**: IntaSend
- **Monitoring**: Prometheus + Grafana
- **Email**: SMTP (Gmail/SendGrid)
- **PDF**: gofpdf
- **QR Codes**: go-qrcode

---

## 🚀 Local Development

### Prerequisites
- Go 1.25.3+
- PostgreSQL
- Redis
- Docker (optional)

### Setup

1. **Clone and install dependencies**
   ```bash
   git clone <your-repo>
   cd ticketing_system
   go mod download
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Run database**
   ```bash
   docker run -d -p 5432:5432 \
     -e POSTGRES_DB=ticketing_system \
     -e POSTGRES_PASSWORD=postgres \
     postgres:latest
   ```

4. **Run Redis**
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   ```

5. **Start backend**
   ```bash
   go run cmd/api-server/main.go
   ```

6. **Run demo (in another terminal)**
   ```bash
   cd demo-app
   python3 -m http.server 3000
   ```

7. **Access**
   - Backend API: http://localhost:8080
   - Demo App: http://localhost:3000

---

## 🐳 Docker Deployment

### Backend
```bash
docker build -t ticketing-backend .
docker run -p 8080:8080 --env-file .env ticketing-backend
```

### Demo
```bash
cd demo-app
docker build -t ticketing-demo .
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 ticketing-demo
```

### Full Stack with Docker Compose (Local Only)
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/events

# Test with coverage
go test -cover ./...

# Integration tests
./test-email-setup.sh
./test-pdf-system.sh
./test-bulk-operations.sh
```

---

## 📊 Monitoring

The system includes Prometheus and Grafana for monitoring:

```bash
# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# Access dashboards
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin123)
```



---

## 🔐 Environment Variables

Key environment variables for deployment:

```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/db

# Redis
REDIS_URL=redis://host:6379

# Server
PORT=8080
ENVIRONMENT=production

# Auth
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
S3_BUCKET_NAME=your-bucket

# Payments
INTASEND_PUBLISHABLE_KEY=your-key
INTASEND_SECRET_KEY=your-secret

# CORS
CORS_ALLOWED_ORIGINS=https://your-demo.com
```

See [.env.example](.env.example) for complete list.

---

## 📖 API Documentation

### Authentication
```bash
POST /api/auth/register       # Register user
POST /api/auth/login          # Login
POST /api/auth/logout         # Logout
POST /api/auth/reset-password # Reset password
POST /api/auth/2fa/enable     # Enable 2FA
```

### Events
```bash
GET    /api/events            # List events
GET    /api/events/:id        # Get event
POST   /api/events            # Create event (organizer)
PUT    /api/events/:id        # Update event (organizer)
DELETE /api/events/:id        # Delete event (organizer)
```

### Tickets
```bash
POST   /api/tickets/purchase  # Purchase tickets
GET    /api/tickets/my        # My tickets
GET    /api/tickets/:id/pdf   # Download PDF
POST   /api/tickets/:id/verify # Verify ticket
```



---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

## 📝 License

This project is licensed under the MIT License.

---

## 🆘 Support


- **Issues**: Open a GitHub issue
- **Email**: kamausimon217@gmail.com

---



---

**Built with ❤️ using Go**
