# Demo Deployment Guide - Quick Start 🚀

This guide will help you deploy your ticketing system backend as a demo in **less than 10 minutes**.

## Prerequisites

1. **AWS Account** with CLI credentials configured
2. **Linux/macOS** terminal (or WSL on Windows)
3. **Internet connection**

## Quick Deploy (3 Steps)

### Step 1: Configure AWS CLI (if not already done)

```bash
aws configure
# Enter your:
# - AWS Access Key ID
# - AWS Secret Access Key  
# - Default region (e.g., us-east-1)
# - Output format: json
```

### Step 2: Run the Deployment Script

```bash
# Make scripts executable
chmod +x demo-deploy.sh deploy-app.sh cleanup-demo.sh

# Deploy infrastructure (takes ~3 minutes)
./demo-deploy.sh
```

This script will:
- ✅ Create VPC with networking
- ✅ Launch EC2 instance (t3.medium)
- ✅ Set up PostgreSQL and Redis via Docker
- ✅ Configure security groups
- ✅ Save deployment info to `deployment-info.txt`

**Output Example:**
```
========================================
  Deployment Successful!
========================================

Instance is initializing (takes ~3 minutes)

Public IP: 54.123.45.67
SSH Command: ssh -i ticketing-demo-key.pem ec2-user@54.123.45.67

Deployment details saved to: deployment-info.txt
```

### Step 3: Deploy Your Application

```bash
# Wait 3 minutes for instance initialization, then:
./deploy-app.sh
```

This will:
- ✅ Deploy a demo web server
- ✅ Set up systemd service
- ✅ Start the application on port 8080

**Your demo is now live!** 🎉

## Access Your Demo

### Via Web Browser
```
http://YOUR_PUBLIC_IP:8080
```

### Via Command Line
```bash
# Health check
curl http://YOUR_PUBLIC_IP:8080/health

# Status endpoint
curl http://YOUR_PUBLIC_IP:8080/api/status
```

### SSH Access
```bash
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP
```

## Deploying Your Actual Application

The demo script creates a placeholder server. To deploy your real ticketing system:

### Option 1: Deploy from Repository

```bash
# SSH into the server
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP

# Navigate to app directory
cd /opt/ticketing-system

# Backup demo files
mv main.go main.go.backup

# Clone your repository
git clone https://github.com/YOUR_USERNAME/ticketing_system.git .
# Or use the repo URL with your code

# Install dependencies
/usr/local/go/bin/go mod download

# Run migrations
cd migrations
/usr/local/go/bin/go run main.go
cd ..

# Build your application
/usr/local/go/bin/go build -o api-server ./cmd/api-server

# Update the systemd service
sudo nano /etc/systemd/system/ticketing-demo.service
# Change ExecStart to: /opt/ticketing-system/api-server

# Restart service
sudo systemctl daemon-reload
sudo systemctl restart ticketing-demo

# Check status
sudo systemctl status ticketing-demo

# View logs
sudo journalctl -u ticketing-demo -f
```

### Option 2: Deploy from Local Machine

```bash
# From your local project directory
tar czf app.tar.gz cmd/ internal/ migrations/ go.mod go.sum

# Copy to server
scp -i ticketing-demo-key.pem app.tar.gz ec2-user@YOUR_PUBLIC_IP:/tmp/

# SSH and deploy
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP << 'EOF'
cd /opt/ticketing-system
tar xzf /tmp/app.tar.gz
/usr/local/go/bin/go build -o api-server ./cmd/api-server
sudo systemctl restart ticketing-demo
EOF
```

## Environment Variables

The application has access to these environment variables:

```bash
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ticketing_db
DB_USER=ticketing_user
DB_PASSWORD=ChangeMe123!
REDIS_HOST=localhost
REDIS_PORT=6379
```

To add more variables:
```bash
sudo nano /etc/systemd/system/ticketing-demo.service
# Add lines like: Environment="VAR_NAME=value"
sudo systemctl daemon-reload
sudo systemctl restart ticketing-demo
```

## Monitoring & Troubleshooting

### Check Application Status
```bash
# Service status
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo systemctl status ticketing-demo"

# View live logs
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo journalctl -u ticketing-demo -f"

# Check setup status
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "cat /tmp/setup-status"
```

### Check Database
```bash
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP << 'EOF'
docker exec -it $(docker ps -qf "name=postgres") psql -U ticketing_user -d ticketing_db
# Inside psql:
# \dt - list tables
# \q - quit
EOF
```

### Check Redis
```bash
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP << 'EOF'
docker exec -it $(docker ps -qf "name=redis") redis-cli ping
EOF
```

### Restart Services
```bash
# Restart application
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo systemctl restart ticketing-demo"

# Restart database containers
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "cd /opt/ticketing-system && docker-compose restart"
```

## Security Notes

### For Production Use

⚠️ **This is a demo setup**. For production, you should:

1. **Change default passwords**
   ```bash
   # Edit docker-compose.yml with strong passwords
   ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP
   sudo nano /opt/ticketing-system/docker-compose.yml
   ```

2. **Use AWS RDS instead of Docker PostgreSQL**
   - Better backup and recovery
   - Automatic updates
   - High availability

3. **Use AWS ElastiCache instead of Docker Redis**
   - Managed service
   - Better performance
   - Automatic failover

4. **Add HTTPS with SSL/TLS**
   - Use AWS Certificate Manager
   - Add Application Load Balancer
   - See full guide: AWS_FIREWALL_DEPLOYMENT_GUIDE.md

5. **Restrict SSH access**
   - Current setup allows SSH from your IP only
   - Use AWS Systems Manager Session Manager for better security

6. **Enable CloudWatch monitoring**
   - Set up log aggregation
   - Create alarms for errors
   - Monitor resource usage

## Cost Estimate

### Demo Setup
- **EC2 t3.medium**: ~$0.0416/hour = ~$30/month
- **EBS Storage**: ~$0.10/GB = ~$2/month (20GB)
- **Data Transfer**: First 1GB free, then ~$0.09/GB

**Estimated Demo Cost**: ~$32-40/month

### Free Tier Eligible
If you're in your first 12 months:
- 750 hours/month of t2.micro (use t2.micro in script)
- 30GB EBS storage
- **Potential cost**: $0-5/month with free tier

### To Reduce Costs
```bash
# Stop instance when not in use
aws ec2 stop-instances --instance-ids YOUR_INSTANCE_ID

# Start when needed
aws ec2 start-instances --instance-ids YOUR_INSTANCE_ID

# Or delete completely
./cleanup-demo.sh
```

## Cleanup (Delete Everything)

When you're done with the demo:

```bash
./cleanup-demo.sh
```

This will:
- ✅ Terminate EC2 instance
- ✅ Delete VPC and all networking
- ✅ Remove security groups
- ✅ Clean up local files

**Note**: SSH key pair is NOT deleted automatically. To remove:
```bash
aws ec2 delete-key-pair --key-name ticketing-demo-key
rm -f ticketing-demo-key.pem
```

## Common Issues & Solutions

### Issue: "Connection refused" when accessing app
```bash
# Check if service is running
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo systemctl status ticketing-demo"

# Check logs
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo journalctl -u ticketing-demo -n 50"
```

### Issue: Can't SSH to instance
```bash
# Check security group allows your IP
aws ec2 describe-security-groups --group-ids YOUR_SG_ID

# Your IP may have changed - update security group
NEW_IP=$(curl -s ifconfig.me)
aws ec2 authorize-security-group-ingress \
  --group-id YOUR_SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr ${NEW_IP}/32
```

### Issue: Database connection errors
```bash
# Check if containers are running
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "docker ps"

# Check container logs
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "docker logs postgres"
```

### Issue: Application won't start
```bash
# Check for port conflicts
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo lsof -i :8080"

# Check application errors
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP "sudo journalctl -u ticketing-demo -f"
```

## Next Steps

1. ✅ **Test your endpoints**: Use curl or Postman to test API endpoints
2. ✅ **Run migrations**: Ensure database schema is set up
3. ✅ **Load test data**: Use seed scripts to populate database
4. ✅ **Performance testing**: Test with real load
5. ✅ **Add monitoring**: Set up CloudWatch dashboards
6. ✅ **Production deployment**: Follow AWS_FIREWALL_DEPLOYMENT_GUIDE.md

## Support Files

- `demo-deploy.sh` - Infrastructure deployment
- `deploy-app.sh` - Application deployment
- `cleanup-demo.sh` - Resource cleanup
- `deployment-info.txt` - Generated deployment details
- `AWS_FIREWALL_DEPLOYMENT_GUIDE.md` - Full production guide

## Getting Help

If you encounter issues:

1. Check `deployment-info.txt` for your configuration
2. Review logs: `sudo journalctl -u ticketing-demo -f`
3. Check AWS CloudWatch for instance metrics
4. Verify security groups allow necessary traffic

---

**Your demo is ready to go! 🚀**

Access it at: `http://YOUR_PUBLIC_IP:8080`
