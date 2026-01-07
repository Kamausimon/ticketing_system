# Redis & S3 Integration Guide

This guide explains how to integrate Redis for session caching and AWS S3 for image storage in the ticketing system.

## Features

### Redis Session Manager
- **Primary storage**: Redis for distributed, fast session caching
- **Automatic fallback**: In-memory cache when Redis is unavailable
- **Health monitoring**: Continuous health checks with automatic failover
- **Zero downtime**: Seamless switching between Redis and in-memory storage

### S3 Storage Service
- **Primary storage**: AWS S3 for scalable image storage
- **Automatic fallback**: Local filesystem when S3 is unavailable
- **Supported operations**: Upload, delete, presigned URLs
- **File validation**: Size and type validation for images

## Configuration

### 1. Environment Variables

Add these variables to your `.env` file:

```env
# Redis Configuration
REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# AWS S3 Configuration
S3_ENABLED=true
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
S3_BUCKET=your-bucket-name
S3_PUBLIC_URL=https://your-bucket-name.s3.us-east-1.amazonaws.com

# Local Storage Fallback
LOCAL_STORAGE_PATH=./uploads
```

### 2. Redis Setup

#### Local Development (Docker)
```bash
docker run -d --name redis \
  -p 6379:6379 \
  redis:7-alpine
```

#### Production (AWS ElastiCache)
```bash
# Create ElastiCache Redis cluster
aws elasticache create-cache-cluster \
  --cache-cluster-id my-redis-cluster \
  --cache-node-type cache.t3.micro \
  --engine redis \
  --num-cache-nodes 1
```

### 3. S3 Setup

#### Create S3 Bucket
```bash
aws s3 mb s3://your-bucket-name --region us-east-1
```

#### Set Bucket Policy for Public Read
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-bucket-name/*"
    }
  ]
}
```

#### Enable CORS (if needed)
```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
    "AllowedOrigins": ["*"],
    "ExposeHeaders": []
  }
]
```

## Usage Examples

### Session Management

The session manager is automatically initialized in `main.go` and can be used throughout the application:

```go
// Store session data
sessionManager.Set("user_123", userData, 30*time.Minute)

// Retrieve session data
var userData UserData
err := sessionManager.Get("user_123", &userData)

// Delete session
sessionManager.Delete("user_123")

// Check if key exists
exists := sessionManager.Exists("user_123")

// Check Redis health
isHealthy := sessionManager.IsRedisHealthy()
```

### File Storage

The storage service is integrated into organizer and event handlers:

#### Uploading Files
```go
// Upload file (automatically uses S3 or local fallback)
result, err := storageService.UploadFile(file, fileHeader, "events")
if err != nil {
    // Handle error
}

// Access upload result
fmt.Println("URL:", result.URL)
fmt.Println("Key:", result.Key)
fmt.Println("Backend:", result.Backend) // "s3" or "local"
```

#### Deleting Files
```go
err := storageService.DeleteFile(fileKey)
```

#### Generating Presigned URLs
```go
// Generate temporary access URL (1 hour expiration)
url, err := storageService.GeneratePresignedURL(fileKey, 1*time.Hour)
```

## Fallback Behavior

### Redis Fallback
1. **Startup**: System attempts to connect to Redis
2. **Health checks**: Runs every 10 seconds
3. **Failure detection**: If Redis becomes unavailable, automatically switches to in-memory cache
4. **Recovery**: Automatically switches back to Redis when available
5. **Seamless operation**: Application continues without interruption

### S3 Fallback
1. **Startup**: System attempts to connect to S3
2. **Failure handling**: If S3 is unavailable, uses local filesystem
3. **Directory creation**: Automatically creates `./uploads` directory
4. **Consistent API**: Same interface regardless of backend

## Monitoring

### Check Backend Status

Add a status endpoint to monitor storage backends:

```go
router.HandleFunc("/status/storage", func(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "redis": map[string]interface{}{
            "enabled": sessionManager.IsRedisHealthy(),
            "backend": "redis",
        },
        "storage": storageService.GetBackendInfo(),
    }
    json.NewEncoder(w).Encode(status)
}).Methods("GET")
```

### Log Output

The system provides clear status messages:

```
✅ Redis session manager initialized (with in-memory fallback)
✅ S3 storage initialized successfully
```

Or if services are unavailable:

```
⚠️  Redis health check failed: connection refused (using fallback cache)
⚠️  S3 bucket not accessible: NoSuchBucket (using local storage)
```

## Best Practices

### Redis
1. **Use password authentication** in production
2. **Configure persistence** (RDB/AOF) for data durability
3. **Set appropriate TTL** for session data
4. **Monitor memory usage** and configure maxmemory policies
5. **Use Redis Sentinel** or ElastiCache for high availability

### S3
1. **Use IAM roles** instead of access keys when running on AWS
2. **Enable versioning** for important buckets
3. **Configure lifecycle policies** to manage old files
4. **Use CloudFront CDN** for better performance
5. **Enable server-side encryption** for sensitive data
6. **Set up proper CORS** for browser uploads

### Security
1. **Never commit credentials** to version control
2. **Use environment variables** for all secrets
3. **Rotate credentials regularly**
4. **Implement rate limiting** on upload endpoints
5. **Validate file types and sizes** before upload

## Troubleshooting

### Redis Connection Issues
```bash
# Test Redis connection
redis-cli -h localhost -p 6379 ping

# Check Redis logs
docker logs redis
```

### S3 Permission Issues
```bash
# Test AWS credentials
aws sts get-caller-identity

# List buckets
aws s3 ls

# Test bucket access
aws s3 ls s3://your-bucket-name
```

### Storage Path Issues
```bash
# Ensure upload directory exists and is writable
mkdir -p ./uploads
chmod 755 ./uploads
```

## Development vs Production

### Development
- Redis: Optional, use in-memory fallback
- S3: Optional, use local storage
- Credentials: Use `.env` file

### Production
- Redis: Required (ElastiCache recommended)
- S3: Required for scalability
- Credentials: Use IAM roles or AWS Secrets Manager

## Performance Considerations

### Redis
- **Latency**: < 1ms for local, < 5ms for ElastiCache
- **Throughput**: 100K+ ops/sec
- **Memory**: Plan for ~1-10MB per session

### S3
- **Latency**: 100-200ms for uploads
- **Throughput**: Unlimited (scales automatically)
- **Cost**: $0.023 per GB stored, $0.09 per GB transferred

### Local Storage
- **Latency**: < 1ms
- **Throughput**: Limited by disk I/O
- **Cost**: Storage hardware only

## Migration Guide

### Existing Local Files to S3
```bash
# Sync local uploads to S3
aws s3 sync ./uploads s3://your-bucket-name/

# Update database records
# Run migration script to update image URLs
```

### From No Cache to Redis
No migration needed - sessions start fresh with Redis enabled.

## Additional Resources

- [Redis Documentation](https://redis.io/documentation)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/)
- [Redis Go Client](https://redis.uptrace.dev/)
