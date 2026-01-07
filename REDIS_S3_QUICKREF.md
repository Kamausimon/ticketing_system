# Redis & S3 Quick Reference Card

## 🚀 Quick Start

### Enable Redis (Optional)
```bash
# Start Redis with Docker
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Update .env
REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
```

### Enable S3 (Optional)
```bash
# Update .env
S3_ENABLED=true
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
AWS_REGION=us-east-1
S3_BUCKET=your-bucket
```

### Run Application
```bash
go run cmd/api-server/main.go
```

## 📊 Status Indicators

### Successful Initialization
```
✅ Redis session manager initialized (with in-memory fallback)
✅ S3 storage initialized successfully
```

### Fallback Mode
```
⚠️  Redis disabled - using in-memory sessions only
⚠️  S3 bucket not accessible (using local storage)
```

## 🔧 Configuration Matrix

| Feature | Development | Production |
|---------|-------------|------------|
| Redis | Optional | Recommended |
| S3 | Optional | Recommended |
| Local Storage | Default | Fallback |
| In-memory Cache | Default | Fallback |

## 📁 File Upload Endpoints

### Upload Organizer Logo
```bash
POST /organizers/logo
Content-Type: multipart/form-data
Authorization: Bearer {token}

Form Data:
  logo: (image file)
```

### Upload Event Image
```bash
POST /organizers/events/{id}/images
Content-Type: multipart/form-data
Authorization: Bearer {token}

Form Data:
  image: (image file)
```

## 🎯 Features

### Redis Session Manager
- ✅ Distributed session caching
- ✅ Auto-fallback to in-memory
- ✅ Health monitoring (10s intervals)
- ✅ Seamless failover
- ✅ Zero-downtime recovery

### S3 Storage Service
- ✅ Scalable cloud storage
- ✅ Auto-fallback to local files
- ✅ Image validation
- ✅ Presigned URLs
- ✅ Automatic cleanup

## ⚙️ Environment Variables

```env
# Redis Configuration
REDIS_ENABLED=false              # Enable Redis
REDIS_ADDR=localhost:6379        # Redis address
REDIS_PASSWORD=                  # Redis password
REDIS_DB=0                       # Redis database number

# S3 Configuration
S3_ENABLED=false                 # Enable S3
AWS_ACCESS_KEY_ID=               # AWS access key
AWS_SECRET_ACCESS_KEY=           # AWS secret key
AWS_REGION=us-east-1            # AWS region
S3_BUCKET=                       # S3 bucket name
S3_PUBLIC_URL=                   # Public URL base
LOCAL_STORAGE_PATH=./uploads     # Local fallback path
```

## 🏗️ Architecture

```
Request Flow:
┌─────────────┐
│   Client    │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│  API Server │
└─────┬───────┘
      │
      ├────────────────────┐
      │                    │
      ▼                    ▼
┌─────────────┐    ┌──────────────┐
│   Storage   │    │   Session    │
│   Service   │    │   Manager    │
└─────┬───────┘    └──────┬───────┘
      │                   │
      ├──────┬            ├──────┬
      │      │            │      │
      ▼      ▼            ▼      ▼
   ┌───┐  ┌────┐      ┌─────┐ ┌────┐
   │S3 │  │Local│     │Redis│ │Mem │
   └───┘  └────┘      └─────┘ └────┘
 Primary Fallback   Primary Fallback
```

## 🔍 Troubleshooting

### Redis Issues
```bash
# Test connection
redis-cli -h localhost -p 6379 ping

# Check logs
docker logs redis

# Application will auto-fallback to in-memory
```

### S3 Issues
```bash
# Test AWS credentials
aws sts get-caller-identity

# Test bucket access
aws s3 ls s3://your-bucket-name

# Application will auto-fallback to local storage
```

### Upload Issues
```bash
# Check permissions
ls -la uploads/
chmod 755 uploads/

# Check disk space
df -h
```

## 📈 Performance

| Operation | Redis | In-Memory | S3 | Local |
|-----------|-------|-----------|----|----|
| Read | < 1ms | < 0.1ms | 50-100ms | < 10ms |
| Write | < 2ms | < 0.1ms | 100-200ms | < 10ms |
| Throughput | 100K/s | Unlimited | Unlimited | Disk I/O |

## 💰 Cost Estimates (Monthly)

### Redis (AWS ElastiCache)
- t3.micro: ~$15/month
- t3.small: ~$30/month

### S3
- Storage: $0.023/GB
- Transfer: $0.09/GB
- Requests: $0.005/1000

### Development
- Redis (Docker): Free
- Local Storage: Free

## 📚 Related Documentation

- [`REDIS_S3_INTEGRATION.md`](REDIS_S3_INTEGRATION.md) - Full integration guide
- [`APPLICATION_COMPLETE.md`](APPLICATION_COMPLETE.md) - Complete feature list
- [`API_ROUTES.md`](API_ROUTES.md) - API reference

## 🎉 Success Indicators

Application is working correctly when you see:
1. ✅ Build completes without errors
2. ✅ Server starts on port 8080
3. ✅ Services initialize (Redis/S3 or fallbacks)
4. ✅ File uploads return valid URLs
5. ✅ Metrics available at /metrics

## 🚦 Quick Health Check

```bash
# Check if server is running
curl http://localhost:8080/metrics

# Should see Prometheus metrics output
```

---

**Remember**: Both Redis and S3 are **optional**. The application works perfectly with in-memory cache and local file storage!
