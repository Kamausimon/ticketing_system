# 🚀 Production Safety Quick Reference

## ✅ What's Now Protected

### 1. Health Check Endpoint
```bash
# Test it now:
curl https://ticketingsystem-production-4a1d.up.railway.app/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2026-01-13T14:30:00Z",
  "checks": {
    "database": {
      "healthy": true
    }
  }
}
```

### 2. Pre-commit Safety Hook
Automatically prevents:
- ❌ `log.Fatal()` in handlers (crashes server)
- ❌ Committing `.env` file (exposes secrets)
- ⚠️  Warns about hardcoded secrets

### 3. All Database Tables Now Created
Events, Tickets, Orders, and 60+ other tables will be created on next deploy.

---

## 🎯 Immediate Action Items

### Today - Enable Railway Auto-Rollback
1. Go to: https://railway.app → Your Project → Settings
2. Find "Health Checks" section
3. Set:
   - **Path:** `/health`
   - **Timeout:** 30 seconds
   - **Enabled:** ✅ ON
4. Save

**What it does:** If deployment fails health check, Railway automatically reverts to previous working version.

### This Week - Set Up Staging
Run this script:
```bash
./setup-staging.sh
```

Or manually:
```bash
railway environment --create staging
railway environment staging
railway up
```

### Before Next Change
Read: [PRODUCTION_DEPLOYMENT_GUIDE.md](PRODUCTION_DEPLOYMENT_GUIDE.md)

---

## 🔥 Emergency: How to Rollback

### Option 1: Automatic (Recommended)
Railway auto-rollbacks if health check fails ✅

### Option 2: Manual Rollback
```bash
# Via Dashboard:
Railway → Deployments → Previous Deployment → Redeploy

# Via CLI:
railway rollback
```

### Option 3: Nuclear Option
```bash
# Scale to 0 replicas (stops app temporarily)
Railway Dashboard → Settings → Scale → 0 replicas

# Fix issue, test on staging, then scale back up
```

---

## 📊 Monitor Your App

### Check Health
```bash
curl https://ticketingsystem-production-4a1d.up.railway.app/health
```

### View Logs
```bash
railway logs --follow
```

### Check Metrics
```bash
curl https://ticketingsystem-production-4a1d.up.railway.app/metrics
```

---

## ⚡ Quick Test: Is My Deployment Safe?

Run this checklist after deploying:

```bash
# 1. Health check passes
curl -s https://ticketingsystem-production-4a1d.up.railway.app/health | grep "healthy"

# 2. Critical endpoints work
curl -s https://ticketingsystem-production-4a1d.up.railway.app/events | grep "events"

# 3. No errors in logs
railway logs --tail 50 | grep -i error

# 4. Database connected
curl -s https://ticketingsystem-production-4a1d.up.railway.app/health | grep '"database":{"healthy":true}'
```

All pass? ✅ Deployment is safe!

---

## 🛡️ Safety Rules

### NEVER do in Production:
- ❌ Deploy without testing on staging
- ❌ Run `AutoMigrate` after initial setup (use versioned migrations)
- ❌ Use `log.Fatal()` in handlers
- ❌ Deploy on Friday afternoon
- ❌ Make database changes without backup

### ALWAYS do:
- ✅ Test on staging first
- ✅ Monitor logs during deployment
- ✅ Backup database before migrations
- ✅ Have rollback plan ready
- ✅ Watch health checks for 15 minutes after deploy

---

## 📚 Full Documentation

- **Complete Guide:** [PRODUCTION_DEPLOYMENT_GUIDE.md](PRODUCTION_DEPLOYMENT_GUIDE.md)
- **Railway Docs:** https://docs.railway.app
- **Health Checks:** https://docs.railway.app/deploy/healthchecks

---

## 🆘 Need Help?

1. **Check logs:** `railway logs --follow`
2. **Check health:** `curl .../health`
3. **Rollback if needed:** `railway rollback`
4. **Read guide:** PRODUCTION_DEPLOYMENT_GUIDE.md

**Pro Tip:** Test in staging = Sleep well at night 😴
