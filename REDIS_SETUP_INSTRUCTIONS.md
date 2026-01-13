# Redis Setup Instructions for Railway

## Step 1: Get Railway Redis Connection String

1. Go to your Railway dashboard: https://railway.app/
2. Click on your **Redis** service
3. Click on the **Variables** tab
4. Copy the value of `REDIS_PRIVATE_URL` or `REDIS_URL`

It should look like one of these formats:
```
redis://default:PASSWORD@redis.railway.internal:6379
redis://default:PASSWORD@redis-production.up.railway.app:6379
```

## Step 2: Update Your .env File

Replace the Redis configuration in your `.env` file:

### OLD (Current - Wrong):
```
REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_DB=0
REDIS_PASSWORD=your_redis_password
```

### NEW (Replace with your actual Railway Redis URL):
```
REDIS_ENABLED=true
REDIS_URL=redis://default:YOUR_PASSWORD@redis.railway.internal:6379
REDIS_ADDR=redis.railway.internal:6379
REDIS_DB=0
REDIS_PASSWORD=YOUR_ACTUAL_PASSWORD
```

**IMPORTANT:** Use the **Private Network URL** (redis.railway.internal) not the public URL.
This is faster and more secure since it stays within Railway's private network.

## Step 3: Update Railway Environment Variables

In your Railway **ticketing_system** service:
1. Go to Variables tab
2. Add/Update these variables:
   ```
   REDIS_ENABLED=true
   REDIS_URL=<paste your Redis private URL here>
   REDIS_ADDR=redis.railway.internal:6379
   REDIS_DB=0
   REDIS_PASSWORD=<your Redis password>
   ```

## Step 4: Verify Connection

After deploying, check the logs. You should see:
```
✅ Redis session manager initialized (with in-memory fallback)
✅ Redis health check: connected
```

If you see errors, the connection string might be wrong.

## What This Will Fix

Once Redis is properly connected, we'll implement caching for:

1. **Events List** (Cache for 5 minutes)
   - Current: 270ms - 2s response time
   - With cache: < 50ms response time ⚡

2. **Search Results** (Cache for 2 minutes)
   - Current: Slow database queries
   - With cache: Instant results ⚡

3. **Event Details** (Cache for 10 minutes)
   - Frequently viewed events stay in memory

4. **Metrics Endpoint** (Cache for 30 seconds)
   - Current: 8 seconds! 😱
   - With cache: < 100ms ⚡

## Testing After Setup

```bash
# 1. Deploy your changes to Railway
git add .
git commit -m "Configure Railway Redis connection"
git push

# 2. Wait for deployment, then test
./load-test.sh light

# 3. Check Railway Redis metrics
# You should now see:
# - Memory usage > 0 B
# - Network traffic activity
# - CPU usage when requests come in
```

## Expected Improvements

Before Redis:
- P95 latency: 1.77s
- Throughput: 14 req/sec

After Redis caching:
- P95 latency: < 300ms (5-6x faster!)
- Throughput: 100+ req/sec (7x more!)
- Database load: 80% reduction

## Troubleshooting

**Can't find Redis URL?**
- Make sure Redis service is deployed in Railway
- Check the service is in the same project as your app

**Connection timeout?**
- Verify you're using the PRIVATE URL (redis.railway.internal)
- Check both services are in the same Railway environment

**Still showing 0 B memory?**
- Check application logs for Redis connection errors
- Verify REDIS_ENABLED=true
- Restart the application after setting variables
