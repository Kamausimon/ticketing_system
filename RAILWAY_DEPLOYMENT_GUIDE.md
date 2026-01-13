# Railway Deployment Guide

## Overview

This guide covers deploying both the backend API and demo app separately on Railway.

## Prerequisites

- Railway account (sign up at https://railway.app)
- Railway CLI (optional): `npm install -g @railway/cli`
- PostgreSQL database (Railway provides this)
- Redis instance (Railway provides this)

## Part 1: Deploy Backend API

### Step 1: Create a New Project on Railway

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Authorize Railway to access your repository
5. Select `ticketing_system` repository

### Step 2: Configure the Backend Service

1. Railway will detect the `Dockerfile` and automatically use it
2. Add the following environment variables in Railway dashboard:

```env
# Database (Railway will auto-inject DATABASE_URL)
DB_HOST=<provided by Railway postgres service>
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<provided by Railway>
DB_NAME=ticketing_system
DB_SSLMODE=require

# Server
PORT=8080
SERVER_ADDRESS=0.0.0.0:8080
ENVIRONMENT=production

# JWT
JWT_SECRET=<generate-strong-secret-key>
JWT_EXPIRY=24h

# Email (use your SMTP provider)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=<your-email>
SMTP_PASSWORD=<your-app-password>
SMTP_FROM=<your-email>

# Redis (Railway will auto-inject REDIS_URL)
REDIS_URL=<provided by Railway redis service>

# AWS S3 (for file uploads)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<your-aws-key>
AWS_SECRET_ACCESS_KEY=<your-aws-secret>
S3_BUCKET_NAME=<your-bucket-name>

# Payment Gateway (IntaSend)
INTASEND_PUBLISHABLE_KEY=<your-key>
INTASEND_SECRET_KEY=<your-secret>
INTASEND_WEBHOOK_SECRET=<your-webhook-secret>

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s

# CORS
CORS_ALLOWED_ORIGINS=https://your-demo-app-url.railway.app,https://your-frontend-url.com

# 2FA
TWO_FACTOR_ISSUER=YourTicketingSystem
```

### Step 3: Add PostgreSQL Database

1. In your Railway project, click "+ New"
2. Select "Database" → "PostgreSQL"
3. Railway will automatically create a DATABASE_URL environment variable
4. The backend will use this for database connections

### Step 4: Add Redis

1. In your Railway project, click "+ New"
2. Select "Database" → "Redis"
3. Railway will automatically create a REDIS_URL environment variable

### Step 5: Deploy

1. Railway will automatically deploy on push to main branch
2. Click on the backend service to see deployment logs
3. Once deployed, click "Settings" → "Generate Domain" to get your API URL

### Step 6: Run Migrations

Option A: Via Railway CLI
```bash
railway login
railway link <your-project-id>
railway run go run cmd/migrate/main.go up
```

Option B: Add migration to Dockerfile startup (recommended)
Create a startup script in the Dockerfile that runs migrations automatically.

## Part 2: Deploy Demo App (Separate Project)

### Step 1: Create a New Railway Project for Demo

1. Create a new project in Railway
2. Select "Empty Project"

### Step 2: Configure Demo Service

1. In Railway dashboard, click "+ New" → "GitHub Repo"
2. Select your repository
3. Set "Root Directory" to `/demo-app`
4. Railway will detect the Dockerfile in demo-app folder

### Step 3: Configure Environment Variables

Add this environment variable to tell the demo where the backend is:

```env
API_BASE_URL=https://your-backend-url.railway.app
```

You'll need to update `app.js` to use this environment variable or hardcode your backend URL.

### Step 4: Deploy Demo

1. Click "Deploy"
2. Once deployed, click "Settings" → "Generate Domain"
3. Share this URL with others to access your demo

## Part 3: Alternative - Vercel for Demo App

For static demos, Vercel is often simpler:

1. Install Vercel CLI: `npm install -g vercel`
2. Navigate to demo-app: `cd demo-app`
3. Run: `vercel`
4. Follow prompts
5. Update API URL in app.js to point to your Railway backend

## Testing the Deployment

### Test Backend API
```bash
curl https://your-backend-url.railway.app/health
```

### Test Demo App
1. Open `https://your-demo-url.railway.app` in browser
2. Try registering a new user
3. Browse events
4. Test ticket purchase flow

## Monitoring

Railway provides built-in monitoring:
- View logs in the deployment dashboard
- Set up alerts for errors
- Monitor resource usage

## Updating Deployments

### Backend
```bash
git add .
git commit -m "Update backend"
git push origin main
```
Railway automatically redeploys on push.

### Demo App
Same process - push to git and Railway redeploys automatically.

## Troubleshooting

### Build Fails
- Check Railway build logs
- Verify Dockerfile syntax
- Ensure all dependencies in go.mod

### Environment Variables
- Verify all required env vars are set
- Check for typos in variable names
- Ensure secrets are properly formatted

### Database Connection
- Verify DATABASE_URL is available
- Check if PostgreSQL plugin is properly connected
- Review connection string format

### CORS Errors in Demo
- Add demo URL to CORS_ALLOWED_ORIGINS in backend
- Verify API URL in demo app.js is correct
- Check browser console for specific CORS errors

## Cost Optimization

Railway free tier includes:
- $5 free credit per month
- Hobby plan: $5/month
- Pay for what you use

To optimize costs:
1. Use Railway's PostgreSQL (included)
2. Use Railway's Redis (included)
3. Scale down during development
4. Use Vercel for demo (free tier)

## Security Checklist

- ✅ Use strong JWT_SECRET
- ✅ Enable HTTPS (Railway provides this)
- ✅ Set proper CORS origins
- ✅ Use environment variables for secrets
- ✅ Enable rate limiting
- ✅ Use DATABASE_URL with SSL
- ✅ Secure SMTP credentials

## Next Steps

1. Set up custom domain (Railway Settings → Domains)
2. Configure health checks
3. Set up monitoring alerts
4. Configure backup strategy
5. Set up staging environment
