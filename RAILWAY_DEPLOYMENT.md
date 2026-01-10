# Deploy to Railway - Quick & Easy 🚂

Railway is a modern platform that makes deployment incredibly simple. No complex VPC setup needed!

## Why Railway?

✅ **Simple**: Deploy in minutes, not hours
✅ **Affordable**: $5/month starter plan (or free tier)
✅ **No DevOps**: Automatic SSL, monitoring, logging
✅ **Git Integration**: Deploy on every push
✅ **Built-in DB**: PostgreSQL and Redis included

## Quick Deploy (5 Minutes)

### Step 1: Create Railway Account

1. Go to [railway.app](https://railway.app)
2. Sign up with GitHub
3. That's it!

### Step 2: Create New Project

```bash
# Install Railway CLI
npm install -g @railway/cli

# Or using curl
curl -fsSL https://railway.app/install.sh | sh

# Login
railway login
```

### Step 3: Initialize Project

```bash
cd /home/kamau/projects/ticketing_system

# Initialize Railway project
railway init

# Link to your project
railway link
```

### Step 4: Add Services

#### Add PostgreSQL
```bash
railway add --database postgresql
```

#### Add Redis
```bash
railway add --database redis
```

### Step 5: Configure Environment Variables

Create `railway.json`:
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS",
    "buildCommand": "go build -o api-server ./cmd/api-server"
  },
  "deploy": {
    "startCommand": "./api-server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

Create `nixpacks.toml`:
```toml
[phases.setup]
nixPkgs = ["go_1_22"]

[phases.build]
cmds = ["go mod download", "go build -o api-server ./cmd/api-server"]

[start]
cmd = "./api-server"
```

### Step 6: Set Environment Variables

```bash
# Set via CLI
railway variables set PORT=8080
railway variables set DB_HOST=${{POSTGRES_HOST}}
railway variables set DB_PORT=${{POSTGRES_PORT}}
railway variables set DB_NAME=${{POSTGRES_DB}}
railway variables set DB_USER=${{POSTGRES_USER}}
railway variables set DB_PASSWORD=${{POSTGRES_PASSWORD}}
railway variables set REDIS_HOST=${{REDIS_HOST}}
railway variables set REDIS_PORT=${{REDIS_PORT}}

# Or set in Railway dashboard (easier)
```

### Step 7: Deploy!

```bash
railway up
```

That's it! Your app is now live! 🎉

## Get Your URL

```bash
# Generate domain
railway domain

# Your app will be at: https://your-app.railway.app
```

## Using Railway Dashboard (Easiest Method)

### 1. Create Project from GitHub

1. Go to [railway.app/new](https://railway.app/new)
2. Click "Deploy from GitHub repo"
3. Select your ticketing_system repository
4. Railway auto-detects it's a Go app

### 2. Add Database Services

1. Click "New" → "Database" → "Add PostgreSQL"
2. Click "New" → "Database" → "Add Redis"
3. Railway automatically creates and links them

### 3. Configure Environment Variables

In your service settings, add:
```
PORT=8080
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_NAME=${{Postgres.PGDATABASE}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
REDIS_HOST=${{Redis.REDIS_HOST}}
REDIS_PORT=${{Redis.REDIS_PORT}}
```

Railway provides these automatically via service references!

### 4. Deploy

Push to GitHub, and Railway auto-deploys!

## Dockerfile (Optional, for more control)

Create `Dockerfile`:
```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/api-server

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]
```

Then deploy:
```bash
railway up
```

## Run Migrations

```bash
# Connect to your Railway project
railway run go run migrations/main.go

# Or SSH into a one-off container
railway run bash
go run migrations/main.go
exit
```

## View Logs

```bash
# Real-time logs
railway logs

# Or view in dashboard: Settings → Logs
```

## Custom Domain

```bash
# Add your domain
railway domain add yourdomain.com

# Railway provides SSL automatically!
```

## Environment-specific Deploys

```bash
# Create staging environment
railway environment

# Deploy to production
railway up --environment production

# Deploy to staging
railway up --environment staging
```

## Cost Breakdown

### Hobby Plan (Free)
- $5 credit/month
- 500 hours execution
- 100GB outbound bandwidth
- Good for demos!

### Starter Plan ($5/month)
- $5 credit included
- Usage-based after that
- ~$0.000463/min for execution
- ~$0.10/GB bandwidth

### Typical Monthly Cost
- **Small demo**: $0-5 (free tier)
- **Light production**: $10-20
- **Medium traffic**: $20-50

Much cheaper than AWS for small apps!

## Monitoring

Railway provides built-in:
- ✅ Metrics dashboard
- ✅ Real-time logs
- ✅ Resource usage
- ✅ Deployment history
- ✅ Automatic health checks

## Database Backups

PostgreSQL backups are automatic on paid plans!

```bash
# Download backup
railway backup download

# Restore backup
railway backup restore <backup-id>
```

## Scaling

```bash
# Scale up (via dashboard)
# Settings → Resources → Increase memory/CPU

# Or use replicas (Pro plan)
railway service scale --replicas 3
```

## CI/CD with GitHub Actions

Create `.github/workflows/deploy.yml`:
```yaml
name: Deploy to Railway

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Railway
        run: npm i -g @railway/cli
        
      - name: Deploy
        run: railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

Get token: `railway token` → Save as GitHub secret

## Comparison: Railway vs AWS

| Feature | Railway | AWS (Demo Setup) |
|---------|---------|------------------|
| Setup Time | 5 minutes | 30+ minutes |
| Cost (demo) | $0-10/month | $30-40/month |
| SSL/HTTPS | Automatic | Manual setup |
| Database | Click to add | Manual RDS setup |
| Monitoring | Built-in | CloudWatch setup |
| Scaling | Click to scale | ASG configuration |
| Learning Curve | Easy | Steep |

**For demos and MVPs**: Railway wins!
**For enterprise/large scale**: AWS provides more control

## Complete Setup Script

```bash
#!/bin/bash
# deploy-railway.sh

set -e

echo "🚂 Deploying to Railway..."

# Install Railway CLI if not present
if ! command -v railway &> /dev/null; then
    echo "Installing Railway CLI..."
    npm install -g @railway/cli
fi

# Login (opens browser)
railway login

# Create project
railway init

# Add databases
railway add --database postgresql
railway add --database redis

# Deploy
railway up

# Generate domain
railway domain

echo "✅ Deployment complete!"
echo "Check your dashboard: https://railway.app/dashboard"
```

Make it executable:
```bash
chmod +x deploy-railway.sh
./deploy-railway.sh
```

## Testing Your Deployment

```bash
# Get your URL
RAILWAY_URL=$(railway status --json | jq -r '.domain')

# Test health endpoint
curl https://$RAILWAY_URL/health

# Test API
curl https://$RAILWAY_URL/api/events
```

## Troubleshooting

### Build Fails
```bash
# Check logs
railway logs --build

# Common fix: ensure go.mod is committed
git add go.mod go.sum
git commit -m "Add go modules"
git push
```

### Database Connection Issues
```bash
# Check database status
railway status

# Verify environment variables
railway variables

# Test connection
railway run -- go run test-db-connection.go
```

### App Crashes
```bash
# View crash logs
railway logs --service api

# Check resource usage
railway service metrics
```

## Pro Tips

1. **Use Railway's service references**: `${{Postgres.PGHOST}}` automatically links services
2. **Enable PR deployments**: Every PR gets its own URL for testing
3. **Use templates**: Railway has pre-built templates for common stacks
4. **Monitor costs**: Set up billing alerts in dashboard
5. **Use staging environments**: Test before deploying to production

## Migration from AWS to Railway

If you already deployed on AWS:

1. Export your database:
```bash
pg_dump -h aws-db-host -U user dbname > backup.sql
```

2. Import to Railway:
```bash
railway run psql $DATABASE_URL < backup.sql
```

3. Update DNS to point to Railway
4. Delete AWS resources

## Resources

- [Railway Docs](https://docs.railway.app/)
- [Railway Discord](https://discord.gg/railway)
- [Railway Templates](https://railway.app/templates)
- [Railway Status](https://status.railway.app/)

---

## Quick Command Reference

```bash
# Login
railway login

# Create project
railway init

# Add database
railway add

# Deploy
railway up

# View logs
railway logs

# Open dashboard
railway open

# Get status
railway status

# Environment variables
railway variables set KEY=value
railway variables list

# Generate domain
railway domain

# Connect to database
railway connect postgres

# Run migrations
railway run go run migrations/main.go
```

---

**Deploy in 5 minutes with Railway! Much easier than AWS for demos! 🚂**
