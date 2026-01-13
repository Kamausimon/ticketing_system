# Railway Deployment Troubleshooting

## ✅ Environment Variables - Working!

The warning about `.env` file is **normal and expected**:
```
⚠️  Warning: Error loading .env file: open .env: no such file or directory
⚠️  Using system environment variables instead
```

**This is correct behavior!** Railway injects variables as system environment variables.

---

## 🔍 Find the Real Crash Cause

### Step 1: Check Railway Logs

In Railway dashboard:
1. Click on your **ticketing_system** service
2. Click **"Deployments"** tab
3. Click the latest deployment
4. Look for error messages **after** the .env warning

**Look for these common errors:**

### Database Connection Error
```
❌ failed to connect to database
❌ could not connect to server
❌ connection refused
```

**Fix**: Make sure PostgreSQL plugin is added and connected
1. Railway → Your Project → Click "+ New"
2. Select "Database" → "PostgreSQL"
3. It will auto-create `DATABASE_URL` variable

---

### Port Binding Error
```
❌ bind: address already in use
❌ listen tcp: address already in use
```

**Fix**: Railway automatically sets PORT. Check your code uses:
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
```

---

### Missing Required Variables
```
❌ DB_HOST is not set
❌ SMTP configuration missing
```

**Fix**: Check you have these critical variables set:
- `DATABASE_URL` (from PostgreSQL plugin)
- `REDIS_URL` (from Redis plugin)
- `PORT` (Railway sets automatically)
- `JWT_SECRET`

---

### Redis Connection Error
```
❌ failed to connect to redis
❌ dial tcp: connection refused
```

**Fix**: Add Redis plugin
1. Railway → Your Project → Click "+ New"
2. Select "Database" → "Redis"
3. It will auto-create `REDIS_URL` variable

---

## 🎯 Quick Diagnostic Checklist

Check these in Railway dashboard:

### Services Tab
- [ ] PostgreSQL plugin added and running
- [ ] Redis plugin added and running
- [ ] Backend service shows "Active" (not "Crashed")

### Variables Tab (you have 44 ✅)
- [ ] `DATABASE_URL` present (from PostgreSQL)
- [ ] `REDIS_URL` present (from Redis)
- [ ] `JWT_SECRET` set
- [ ] `PORT` either set or blank (Railway auto-sets)

### Deployment Logs
Look for actual error after the `.env` warnings:
```
✅ Environment variables loaded from .env file    <- Ignore this
✅ Using system environment variables              <- This is what Railway does
❌ [ACTUAL ERROR HERE]                             <- This is what to fix!
```

---

## 🔧 Most Common Fixes

### Fix 1: Database Not Connected
```bash
# In Railway, make sure:
1. PostgreSQL plugin exists
2. It's in the same project
3. DATABASE_URL variable is auto-created
4. Restart backend service after adding database
```

### Fix 2: Missing Redis
```bash
# In Railway:
1. Add Redis plugin
2. REDIS_URL auto-created
3. Restart backend service
```

### Fix 3: Database SSL Mode
Your app might need:
```bash
# Add this variable in Railway:
DB_SSLMODE=disable
```

Or update your connection string to include `sslmode=require`

### Fix 4: Port Configuration
Remove `PORT` variable if you set it manually. Let Railway set it automatically.

Or ensure your code reads it correctly:
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
log.Printf("Starting server on port %s", port)
http.ListenAndServe("0.0.0.0:"+port, router)
```

---

## 📋 What to Send Me

To help you further, share:

1. **Last 20 lines of Railway logs** (after the .env warning)
2. **Plugins added**: PostgreSQL? Redis? Both?
3. **Error message**: The actual error (not the .env warning)

---

## ✅ The .env File is Fine!

**You don't need a .env file on Railway!** Your 44 variables are injected as system environment variables, which is perfect.

The app's code does this:
```go
// Try to load .env (for local development)
if err := godotenv.Load(".env"); err != nil {
    // File not found - this is OK on Railway!
    fmt.Printf("⚠️  Warning: Error loading .env file: %v\n", err)
    fmt.Println("⚠️  Using system environment variables instead")  // <- Railway uses this!
}
```

This is **correct behavior** for production deployments.
