# Production Deployment Best Practices

## 🎯 Core Principle: Never Test in Production

This guide helps you deploy safely without crashing your production app.

---

## 1. Staging Environment Setup (Railway)

### Create Staging Environment
1. Go to Railway Dashboard → Your Project
2. Click "New Environment" → Name it "Staging"
3. Deploy the same code to staging
4. Use a **separate staging database**

### Configure Staging
```bash
# In Railway Staging environment, set these variables:
ENVIRONMENT=staging
DATABASE_URL=<staging-database-url>
FRONTEND_URL=https://staging.yourapp.com

# Keep production safe with:
ENVIRONMENT=production
DATABASE_URL=<production-database-url>
FRONTEND_URL=https://yourapp.com
```

### Deployment Flow
```
Local Dev → Push to GitHub → Deploy to Staging → Test → Deploy to Production
```

**Golden Rule:** Every change goes through staging first!

---

## 2. Health Checks & Auto-Rollback

### Health Check Endpoint
Your app now has `/health` endpoint that checks:
- ✅ Database connectivity
- ✅ Server responsiveness

```bash
# Test health check
curl https://ticketingsystem-production-4a1d.up.railway.app/health
```

### Railway Auto-Rollback Setup
1. Railway → Settings → Health Checks
2. Set path: `/health`
3. Enable auto-rollback on failure
4. Set timeout: 30 seconds

**What happens:** If deployment fails health checks, Railway automatically reverts to previous version.

---

## 3. Database Migration Safety

### ⚠️ Current Issue: AutoMigrate in Production
`AutoMigrate` is dangerous because it can:
- Drop columns with data
- Change types incorrectly
- No rollback capability

### Solution: Use Versioned Migrations

Install migration tool:
```bash
go get -u github.com/golang-migrate/migrate/v4
```

Create migrations:
```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq add_event_capacity

# This creates:
# migrations/000001_add_event_capacity.up.sql
# migrations/000001_add_event_capacity.down.sql
```

Example migration:
```sql
-- 000001_add_event_capacity.up.sql
ALTER TABLE events ADD COLUMN capacity INTEGER DEFAULT 0;
CREATE INDEX idx_events_capacity ON events(capacity);

-- 000001_add_event_capacity.down.sql
DROP INDEX IF EXISTS idx_events_capacity;
ALTER TABLE events DROP COLUMN IF EXISTS capacity;
```

Run migrations:
```bash
# On staging first!
migrate -path migrations -database "$STAGING_DATABASE_URL" up

# Test thoroughly, then production
migrate -path migrations -database "$DATABASE_URL" up

# Rollback if needed
migrate -path migrations -database "$DATABASE_URL" down 1
```

### Update main.go for Production
```go
// In main.go - only run AutoMigrate in development
env := os.Getenv("ENVIRONMENT")
if env == "development" || env == "" {
    // Auto-migrate in dev only
    DB.AutoMigrate(/* all models */)
} else {
    // In production, use explicit migrations
    fmt.Println("⚠️  Production mode - run migrations manually")
}
```

---

## 4. Database Backups

### Automatic Backups (Railway)
Railway automatically backs up PostgreSQL daily. Configure:
1. Railway → Database → Settings
2. Enable continuous backups
3. Set retention: 7 days minimum

### Manual Backup Before Risky Changes
```bash
# Backup production database
railway run pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore if needed
railway run psql $DATABASE_URL < backup_20260113_143000.sql
```

### Test Restore on Staging
```bash
# Restore production backup to staging to test
railway run --environment staging psql $DATABASE_URL < backup_production.sql
```

---

## 5. Monitoring & Alerts

### Built-in Monitoring
You already have:
- ✅ Prometheus metrics at `/metrics`
- ✅ Railway logs (auto-captured)
- ✅ Health checks

### Add Error Tracking (Sentry)
```bash
go get github.com/getsentry/sentry-go
```

```go
// In main.go
import "github.com/getsentry/sentry-go"

func main() {
    // Initialize Sentry
    err := sentry.Init(sentry.ClientOptions{
        Dsn: os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENVIRONMENT"),
        TracesSampleRate: 1.0,
    })
    defer sentry.Flush(2 * time.Second)
    
    // Capture panics automatically
    defer sentry.Recover()
}
```

Get free Sentry account at: https://sentry.io

### Set Up Alerts
1. **Railway Alerts:**
   - Project → Settings → Notifications
   - Enable Discord/Slack/Email alerts
   - Alert on: Deployment failures, crashes

2. **Sentry Alerts:**
   - Alert on: Error spikes, new errors
   - Notify: Email, Slack, PagerDuty

---

## 6. Feature Flags

Control feature rollout without redeploying:

```go
// In config/config.go
type FeatureFlags struct {
    EnableNewCheckout bool
    EnableAIRecommendations bool
    MaxConcurrentOrders int
}

func LoadFeatureFlags() FeatureFlags {
    return FeatureFlags{
        EnableNewCheckout: getEnvBool("FEATURE_NEW_CHECKOUT", false),
        EnableAIRecommendations: getEnvBool("FEATURE_AI_RECOMMENDATIONS", false),
        MaxConcurrentOrders: getEnvInt("MAX_CONCURRENT_ORDERS", 100),
    }
}
```

Usage:
```go
// In handler
if featureFlags.EnableNewCheckout {
    // New code
    return handleNewCheckout(w, r)
}
// Old stable code
return handleOldCheckout(w, r)
```

**Benefit:** Enable features for staging, keep disabled in production until tested.

---

## 7. Deployment Checklist

### Before Every Production Deployment:

- [ ] Changes tested locally
- [ ] All tests pass (`go test ./...`)
- [ ] Deployed to staging
- [ ] Tested on staging (manual + automated)
- [ ] Database migrations tested on staging
- [ ] Backup production database
- [ ] Check Railway health checks are enabled
- [ ] Monitor logs during deployment
- [ ] Verify health check passes
- [ ] Test critical user flows
- [ ] Monitor error rates for 15 minutes

### During Deployment:

```bash
# Watch Railway logs in real-time
railway logs --follow

# Or use Railway CLI to watch deployment
railway status

# Test health immediately after deploy
curl https://your-app.railway.app/health
```

### Rollback Plan:

```bash
# Railway auto-rollbacks on health check failure
# Manual rollback if needed:
# 1. Railway Dashboard → Deployments
# 2. Click previous deployment
# 3. Click "Redeploy"

# Or via CLI:
railway rollback
```

---

## 8. Code Quality Guards

### Prevent log.Fatal in Handlers
Add a pre-commit hook:

```bash
# .git/hooks/pre-commit
#!/bin/bash
if git diff --cached | grep -q "log.Fatal"; then
    echo "❌ Found log.Fatal() - use error returns instead!"
    exit 1
fi
```

### Required Error Handling Pattern
```go
// ❌ BAD: Crashes server
func handler(w http.ResponseWriter, r *http.Request) {
    if err != nil {
        log.Fatal(err)  // NEVER DO THIS
    }
}

// ✅ GOOD: Returns error
func handler(w http.ResponseWriter, r *http.Request) {
    if err != nil {
        log.Printf("Error: %v", err)  // Log for debugging
        middleware.WriteJSONError(w, http.StatusInternalServerError, "error")
        return
    }
}
```

---

## 9. Load Testing Before Launch

Test your app handles traffic:

```bash
# Install k6
brew install k6  # macOS
# or download from k6.io

# Create load test
cat > load_test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
};

export default function () {
  let res = http.get('https://your-app.railway.app/events');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
EOF

# Run test on STAGING first!
k6 run load_test.js
```

---

## 10. Security Checklist

- [ ] Never commit `.env` file (add to `.gitignore`)
- [ ] Use Railway secrets for sensitive data
- [ ] Enable HTTPS only (Railway does this automatically)
- [ ] Rate limiting enabled (you have this ✅)
- [ ] SQL injection protection (GORM handles this ✅)
- [ ] Authentication on protected routes (fixed ✅)
- [ ] CORS configured properly (fixed ✅)
- [ ] Input validation on all endpoints
- [ ] Update dependencies regularly

---

## Quick Reference: Railway Commands

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Link to project
railway link

# View logs
railway logs --follow

# Run command in production
railway run <command>

# Create staging environment
railway environment --create staging

# Switch environments
railway environment staging

# Deploy to specific environment
railway up --environment staging
```

---

## Emergency Procedures

### App Won't Start
```bash
# Check logs
railway logs --tail 100

# Check health
curl https://your-app.railway.app/health

# Rollback immediately
railway rollback
```

### Database Issue
```bash
# Check connection from Railway
railway run psql $DATABASE_URL -c "SELECT 1"

# Restore from backup
railway run psql $DATABASE_URL < backup_file.sql
```

### High Error Rate
```bash
# Scale down traffic if needed
# Railway → Settings → Scale to 0 replicas temporarily
# Fix issue, test on staging, then scale back up
```

---

## Success Metrics

Track these to know your deployment is healthy:
- Response times < 200ms
- Error rate < 0.1%
- Health check success rate > 99.9%
- Database query time < 50ms
- Zero 502/503 errors

Monitor at: `/metrics` (Prometheus format)

---

## Next Steps

1. ✅ Set up staging environment today
2. ✅ Enable health checks and auto-rollback
3. ✅ Set up Sentry for error tracking
4. 📅 Next week: Implement versioned migrations
5. 📅 Before launch: Load testing on staging

**Remember:** Production is sacred. Always test in staging first! 🛡️
