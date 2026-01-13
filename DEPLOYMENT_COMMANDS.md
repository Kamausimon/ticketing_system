# Railway Deployment - Command Reference

## ✅ Problem Solved

**Error**: `dockerfile parse error on line 1: unknown instruction: services:`

**Cause**: Railway tried to use `docker-compose.monitoring.yml` as Dockerfile

**Fix**: Created proper `/Dockerfile` for backend and `/demo-app/Dockerfile` for demo

---

## 🚀 Quick Deploy Commands

### Test Locally First

```bash
# Build and test backend
docker build -t backend-test .
docker run -p 8080:8080 --env-file .env backend-test

# Test backend health
curl http://localhost:8080/health

# Build and test demo
cd demo-app
docker build -t demo-test .
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 demo-test

# Open http://localhost:3000 in browser
```

---

## 🌐 Deploy with Railway CLI

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Deploy backend
railway init
railway up

# Add services
railway add postgresql
railway add redis

# Set environment variables
railway variables set JWT_SECRET=your-secret-here
railway variables set ENVIRONMENT=production
railway variables set SMTP_HOST=smtp.gmail.com
railway variables set SMTP_PORT=587
railway variables set SMTP_USER=your-email@gmail.com
railway variables set SMTP_PASSWORD=your-app-password
railway variables set CORS_ALLOWED_ORIGINS=*

# Generate domain
railway domain

# View logs
railway logs
```

---

## 🎨 Deploy Demo Separately

### Option 1: Railway (in /demo-app)

```bash
cd demo-app

# Create new Railway project
railway init

# Set backend URL
railway variables set API_BASE_URL=https://your-backend.railway.app

# Deploy
railway up

# Generate domain
railway domain
```

### Option 2: Vercel (Recommended - FREE)

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy demo
cd demo-app
vercel

# For production
vercel --prod
```

### Option 3: Netlify

```bash
# Install Netlify CLI
npm install -g netlify-cli

# Deploy demo
cd demo-app
netlify deploy

# For production
netlify deploy --prod
```

---

## 🔍 Verify Deployment

```bash
# Check backend health
curl https://your-backend.railway.app/health

# Should return:
# {"status":"healthy"}

# Check backend API
curl https://your-backend.railway.app/api/events

# Test demo (open in browser)
open https://your-demo.railway.app
```

---

## 📋 Environment Variables

### Backend (Railway)

```bash
# Auto-injected by Railway
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
PORT=8080

# You must add these
JWT_SECRET=your-strong-secret-change-this
ENVIRONMENT=production
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
CORS_ALLOWED_ORIGINS=https://your-demo.railway.app

# Optional
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
```

### Demo (Railway/Vercel)

```bash
API_BASE_URL=https://your-backend.railway.app
```

---

## 🐛 Debugging Commands

```bash
# Check if Dockerfile exists
ls -la Dockerfile
ls -la demo-app/Dockerfile

# Build locally to test
docker build -t test .

# Run and check logs
docker run -p 8080:8080 test

# Check running containers
docker ps

# View container logs
docker logs <container-id>

# SSH into Railway deployment
railway run bash

# Check Railway deployment status
railway status

# View Railway logs
railway logs --tail 100
```

---

## 🔄 Update Deployments

```bash
# Update backend
git add .
git commit -m "Update backend"
git push origin main
# Railway auto-deploys

# Or manually
railway up

# Update demo
cd demo-app
git add .
git commit -m "Update demo"
git push origin main
# Railway/Vercel auto-deploys

# Or manually
railway up  # for Railway
vercel --prod  # for Vercel
```

---

## 📊 Monitoring Commands

```bash
# View Railway logs
railway logs

# Follow logs in real-time
railway logs --follow

# View last 100 lines
railway logs --tail 100

# Check resource usage
railway status

# View deployment history
railway deployments
```

---

## 🗑️ Cleanup Commands

```bash
# Remove Railway project
railway delete

# Remove local Docker images
docker rmi backend-test demo-test

# Remove containers
docker container prune

# Remove unused images
docker image prune -a
```

---

## 🔐 Security Commands

```bash
# Generate strong JWT secret
openssl rand -base64 32

# Generate secure password
openssl rand -base64 24

# Check for secrets in code (before committing)
git diff | grep -i "password\|secret\|key"
```

---

## 📦 Quick Setup Script

Save as `quick-deploy.sh`:

```bash
#!/bin/bash

echo "🚀 Quick Railway Deployment"

# Check requirements
command -v railway >/dev/null 2>&1 || { echo "Install Railway CLI first"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Install Docker first"; exit 1; }

# Test locally first
echo "Testing backend locally..."
docker build -t backend-test . || exit 1

echo "Testing demo locally..."
cd demo-app && docker build -t demo-test . || exit 1
cd ..

echo "✅ Local tests passed"

# Deploy backend
echo "Deploying backend..."
railway init
railway add postgresql
railway add redis
railway up

# Get backend URL
BACKEND_URL=$(railway domain)
echo "Backend deployed at: $BACKEND_URL"

# Deploy demo
echo "Deploying demo..."
cd demo-app
railway init
railway variables set API_BASE_URL="https://$BACKEND_URL"
railway up
cd ..

echo "✅ Deployment complete!"
railway status
```

---

## 📝 Common Tasks

### Update API URL in Demo

```bash
# If using environment variable
railway variables set API_BASE_URL=https://new-backend-url.railway.app

# If hardcoded in app.js
cd demo-app
# Edit app.js line 3
# const API_URL = 'https://new-backend-url.railway.app';
git commit -am "Update API URL"
git push
```

### Add CORS Origin

```bash
railway variables set CORS_ALLOWED_ORIGINS=https://demo1.com,https://demo2.com
```

### View Database URL

```bash
railway variables get DATABASE_URL
```

### Run Migrations

```bash
# Via Railway CLI
railway run go run cmd/migrate/main.go up

# Or add to Dockerfile startup
# See Dockerfile for example
```

---

## 🎯 Quick Reference

| Task | Command |
|------|---------|
| Login to Railway | `railway login` |
| Create project | `railway init` |
| Deploy | `railway up` |
| View logs | `railway logs` |
| Add PostgreSQL | `railway add postgresql` |
| Add Redis | `railway add redis` |
| Set env var | `railway variables set KEY=value` |
| Generate domain | `railway domain` |
| Check status | `railway status` |
| Delete project | `railway delete` |

---

## 📚 Documentation Links

- Railway Docs: https://docs.railway.app
- Railway CLI: https://docs.railway.app/develop/cli
- Vercel Docs: https://vercel.com/docs
- Docker Docs: https://docs.docker.com
- This Project: See `/RAILWAY_QUICKFIX.md`

---

## ✅ Success Checklist

- [ ] `/Dockerfile` exists
- [ ] `/demo-app/Dockerfile` exists
- [ ] Railway CLI installed
- [ ] Docker installed and running
- [ ] Local tests pass
- [ ] Backend deployed to Railway
- [ ] PostgreSQL added to Railway
- [ ] Redis added to Railway
- [ ] Environment variables configured
- [ ] Backend domain generated
- [ ] Demo deployed (Railway or Vercel)
- [ ] Demo API_BASE_URL set
- [ ] Demo domain generated
- [ ] Health check passes
- [ ] Can register and login in demo
- [ ] Can browse and purchase tickets

---

**Quick Start**: Copy commands from this file and run them in order!
