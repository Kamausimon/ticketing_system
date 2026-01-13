## Railway Deployment - Visual Guide

### Current Issue ❌
```
Railway trying to use docker-compose.monitoring.yml
         ↓
    Parse Error
(Not a Dockerfile!)
```

### Solution ✅
```
Create separate Dockerfiles:
    /Dockerfile           (Backend API)
    /demo-app/Dockerfile  (Demo Frontend)
```

---

## File Structure

```
ticketing_system/
│
├── Dockerfile ← Backend uses this
├── .dockerignore ← Excludes demo-app/
├── railway.json
├── cmd/
│   └── api-server/
│       └── main.go ← Backend entry
├── internal/ ← Backend code
│
└── demo-app/ ← Separate demo application
    ├── Dockerfile ← Demo uses this
    ├── railway.json
    ├── docker-entrypoint.sh
    ├── nginx.conf
    ├── index.html
    ├── app.js
    └── styles.css
```

---

## Deployment Flow

### Option 1: Both on Railway (2 Projects)

```
┌─────────────────────────────────────────┐
│         Railway Account                 │
└───────────┬─────────────────────────────┘
            │
    ┌───────┴───────┐
    │               │
    ▼               ▼
┌───────┐      ┌───────┐
│Project│      │Project│
│   1   │      │   2   │
└───┬───┘      └───┬───┘
    │              │
    │              │
    ▼              ▼
┌────────────┐  ┌──────────┐
│  Backend   │  │   Demo   │
│    API     │  │ Frontend │
│            │  │          │
│Root: /     │  │Root:     │
│File:       │  │/demo-app │
│Dockerfile  │  │File:     │
│            │  │Dockerfile│
│Port: 8080  │  │Port: 80  │
└────────────┘  └──────────┘
     │
     ├── PostgreSQL (added)
     └── Redis (added)
```

**Steps:**
1. Create Project 1 → Deploy from repo → Root: `/`
2. Create Project 2 → Deploy from same repo → Root: `/demo-app`
3. Project 1: Add PostgreSQL + Redis
4. Project 2: Set `API_BASE_URL` to Project 1's domain

---

### Option 2: Backend on Railway, Demo on Vercel (Recommended)

```
┌─────────────┐         ┌──────────────┐
│   Railway   │         │    Vercel    │
│  (Backend)  │         │    (Demo)    │
└──────┬──────┘         └──────┬───────┘
       │                       │
       ▼                       ▼
┌──────────────┐        ┌─────────────┐
│   Backend    │        │    Demo     │
│     API      │◄───────│  Frontend   │
│              │  Calls │             │
│  Go Server   │        │ Static HTML │
│  Port: 8080  │        │  Free Tier  │
└──────────────┘        └─────────────┘
       │
       ├── PostgreSQL
       └── Redis
```

**Benefits:**
- ✅ Demo is FREE on Vercel
- ✅ Demo is faster (CDN)
- ✅ Backend isolated
- ✅ Easy updates

---

## Railway Web UI Steps

### Backend Deployment

```
Step 1: New Project
https://railway.app/new
   │
   ├─→ "Deploy from GitHub repo"
   │
   └─→ Select: ticketing_system
         │
         └─→ Railway detects /Dockerfile ✅

Step 2: Add Services
   │
   ├─→ Click "+ New" → PostgreSQL
   │    └─→ Auto-creates DATABASE_URL
   │
   └─→ Click "+ New" → Redis
        └─→ Auto-creates REDIS_URL

Step 3: Environment Variables
   │
   └─→ Service → Variables → Add All

Step 4: Generate Domain
   │
   └─→ Settings → Generate Domain
        │
        └─→ https://xxx.railway.app
```

### Demo Deployment (Separate Project)

```
Step 1: New Project
https://railway.app/new
   │
   └─→ Select: same ticketing_system repo

Step 2: Set Root Directory
   │
   └─→ Service → Settings
        │
        └─→ Root Directory: /demo-app
             │
             └─→ Railway uses /demo-app/Dockerfile ✅

Step 3: Set Backend URL
   │
   └─→ Variables
        │
        └─→ API_BASE_URL=https://xxx.railway.app
                              ↑
                    (From backend project)

Step 4: Generate Domain
   │
   └─→ Settings → Generate Domain
        │
        └─→ https://yyy.railway.app
             │
             └─→ Share this URL! 🎉
```

---

## Communication Flow

```
User visits Demo
       │
       ├─→ https://demo.railway.app
       │   (Static HTML loaded)
       │
       └─→ User clicks "Browse Events"
              │
              ├─→ app.js makes API call
              │   GET https://api.railway.app/api/events
              │
              └─→ Backend processes request
                     │
                     ├─→ Queries PostgreSQL
                     ├─→ Checks Redis cache
                     └─→ Returns JSON
                            │
                            └─→ Demo displays events ✅
```

---

## Environment Variables Mapping

### Backend (.env)
```bash
DATABASE_URL=postgres://...     # Auto from Railway PostgreSQL
REDIS_URL=redis://...           # Auto from Railway Redis
PORT=8080                       # Railway auto-detects
JWT_SECRET=xxx                  # You add manually
SMTP_HOST=smtp.gmail.com        # You add manually
CORS_ALLOWED_ORIGINS=https://demo.railway.app
```

### Demo
```bash
API_BASE_URL=https://backend.railway.app
```

---

## Build Process Visualization

### Backend Build
```
Railway starts build
       │
       ├─→ Reads /Dockerfile
       │      │
       │      ├─→ Stage 1: golang:1.25.3-alpine
       │      │   ├─→ Copy go.mod, go.sum
       │      │   ├─→ go mod download
       │      │   ├─→ Copy source code
       │      │   └─→ Build binary: api-server
       │      │
       │      └─→ Stage 2: alpine:latest
       │          ├─→ Copy binary from Stage 1
       │          ├─→ Copy migrations/
       │          └─→ CMD ["./api-server"]
       │
       └─→ Container starts
              │
              └─→ Listening on :8080 ✅
```

### Demo Build
```
Railway starts build (in /demo-app)
       │
       ├─→ Reads demo-app/Dockerfile
       │      │
       │      ├─→ Base: nginx:alpine
       │      ├─→ Copy index.html, app.js, styles.css
       │      ├─→ Copy nginx.conf
       │      └─→ Copy docker-entrypoint.sh
       │
       └─→ Container starts
              │
              ├─→ Runs docker-entrypoint.sh
              │   └─→ Creates config.js with API_BASE_URL
              │
              └─→ nginx serves on :80 ✅
```

---

## Quick Test Commands

```bash
# Test if backend Dockerfile works locally
docker build -t test-backend .
docker run -p 8080:8080 --env-file .env test-backend
curl http://localhost:8080/health

# Test if demo Dockerfile works locally
cd demo-app
docker build -t test-demo .
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 test-demo
# Open http://localhost:3000 in browser

# Verify demo-app is excluded from backend
docker build -t check .
docker run check ls -la
# Should NOT see demo-app/ folder
```

---

## Common Errors Explained

### Error: "dockerfile parse error on line 1: unknown instruction: services"
```
❌ Railway is reading: docker-compose.monitoring.yml
   This starts with: services:
   Docker expects: FROM

✅ Fix: Railway needs Dockerfile
   This starts with: FROM golang:1.25.3-alpine
```

### Error: "CORS policy: No 'Access-Control-Allow-Origin'"
```
Demo: https://demo.railway.app
  │
  └─→ API Call to: https://api.railway.app
         │
         └─→ Backend checks CORS_ALLOWED_ORIGINS
                │
                ├─→ Not found: demo.railway.app
                └─→ ❌ Blocks request

✅ Fix: Add to backend environment variables:
   CORS_ALLOWED_ORIGINS=https://demo.railway.app
```

### Error: "Demo shows 'localhost:8080' error"
```
Demo loads in browser
  │
  └─→ app.js: API_URL = 'http://localhost:8080'
         │
         └─→ Browser tries localhost ❌
              (No backend running locally!)

✅ Fix: Set API_BASE_URL env var in Railway
   Or update app.js with your backend URL
```

---

## Success Checklist

Backend Deployment:
- ✅ Railway project created
- ✅ /Dockerfile detected
- ✅ PostgreSQL added
- ✅ Redis added
- ✅ Environment variables set
- ✅ Domain generated
- ✅ Health check passes: https://xxx.railway.app/health

Demo Deployment:
- ✅ Separate Railway project (or Vercel)
- ✅ Root directory set to /demo-app
- ✅ API_BASE_URL points to backend
- ✅ Domain generated
- ✅ Can load demo in browser
- ✅ Can register and login
- ✅ Can browse events

---

## Monitoring After Deployment

```bash
# View Railway logs
railway logs

# Or in web UI:
Project → Service → Deployments → Click latest → View Logs
```

Look for:
```
✅ Environment variables loaded
✅ Database connection successful
✅ Redis connection successful
✅ Server listening on :8080
```

---

## Summary

**Problem**: docker-compose.monitoring.yml is not a Dockerfile
**Solution**: Use the created /Dockerfile for backend

**Problem**: Demo app mixed with backend
**Solution**: Separate deployments using /demo-app/Dockerfile

**Result**: 
- Backend API: Professional, scalable, independent
- Demo Frontend: Shareable, fast, free hosting option
