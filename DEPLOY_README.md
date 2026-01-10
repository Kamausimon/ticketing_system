# 🚀 Deploy Your Backend Demo in 3 Commands

## Quick Start

```bash
# 1. Configure AWS (if not already done)
aws configure

# 2. Deploy infrastructure (~3 minutes)
./demo-deploy.sh

# 3. Deploy application (~2 minutes)
./deploy-app.sh
```

**That's it!** Your backend is now live at `http://YOUR_IP:8080`

## What You Get

✅ **EC2 Instance** (t3.medium) with:
- Go 1.22 pre-installed
- Docker & Docker Compose
- PostgreSQL 15 database
- Redis 7 cache
- Systemd service for auto-restart

✅ **Complete Infrastructure**:
- VPC with public subnet
- Security groups (SSH restricted to your IP, HTTP/HTTPS open)
- Internet gateway
- All networking configured

✅ **Demo Web Server**:
- Health check endpoint: `/health`
- Status API: `/api/status`
- Web interface with documentation

## Deployment Info

After running `demo-deploy.sh`, check `deployment-info.txt` for:
- Public IP address
- SSH command
- Database credentials
- All resource IDs

## Example Output

```bash
$ ./demo-deploy.sh
========================================
Ticketing System - Demo Deployment
========================================

✓ AWS credentials configured
✓ Your IP: 203.0.113.42
✓ Created key pair: ticketing-demo-key.pem
✓ VPC ID: vpc-0123456789abcdef
✓ Internet Gateway: igw-0123456789abcdef
✓ Public Subnet: subnet-0123456789abcdef
✓ Security Group: sg-0123456789abcdef
✓ AMI ID: ami-0abcdef1234567890
✓ Instance ID: i-0123456789abcdef
✓ Instance is running!
✓ Public IP: 54.123.45.67

========================================
  Deployment Successful!
========================================

Public IP: 54.123.45.67
SSH Command: ssh -i ticketing-demo-key.pem ec2-user@54.123.45.67

$ ./deploy-app.sh
========================================
Deploy Application to EC2
========================================

✓ Instance is ready
Deploying application to server...
Application deployed successfully!

========================================
  Deployment Complete!
========================================

Your demo server is now running at:
http://54.123.45.67:8080
```

## Access Your Demo

### Web Browser
```
http://YOUR_PUBLIC_IP:8080
```

### Command Line
```bash
# Health check
curl http://YOUR_PUBLIC_IP:8080/health

# Status
curl http://YOUR_PUBLIC_IP:8080/api/status
```

### SSH Access
```bash
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP
```

## Deploy Your Actual Application

The scripts deploy a placeholder demo server. To deploy your real code:

```bash
# SSH into server
ssh -i ticketing-demo-key.pem ec2-user@YOUR_PUBLIC_IP

# Clone your repo
cd /opt/ticketing-system
git clone YOUR_REPO_URL .

# Build
/usr/local/go/bin/go build -o api-server ./cmd/api-server

# Update systemd service
sudo nano /etc/systemd/system/ticketing-demo.service
# Change ExecStart to: /opt/ticketing-system/api-server

# Restart
sudo systemctl restart ticketing-demo
```

## Monitoring

```bash
# View logs
ssh -i KEY.pem ec2-user@IP "sudo journalctl -u ticketing-demo -f"

# Check status
ssh -i KEY.pem ec2-user@IP "sudo systemctl status ticketing-demo"

# Database access
ssh -i KEY.pem ec2-user@IP
docker exec -it $(docker ps -qf "name=postgres") psql -U ticketing_user -d ticketing_db
```

## Environment Variables (Pre-configured)

```
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ticketing_db
DB_USER=ticketing_user
DB_PASSWORD=ChangeMe123!
REDIS_HOST=localhost
REDIS_PORT=6379
```

## Cost

**Demo**: ~$30-40/month
**Free Tier**: $0-5/month (if eligible)

Stop when not in use:
```bash
aws ec2 stop-instances --instance-ids YOUR_INSTANCE_ID
```

## Cleanup

Delete everything:
```bash
./cleanup-demo.sh
```

## Troubleshooting

**Can't connect?**
```bash
# Check service
ssh -i KEY.pem ec2-user@IP "sudo systemctl status ticketing-demo"

# View logs
ssh -i KEY.pem ec2-user@IP "sudo journalctl -u ticketing-demo -n 50"
```

**Database issues?**
```bash
# Check containers
ssh -i KEY.pem ec2-user@IP "docker ps"

# Restart
ssh -i KEY.pem ec2-user@IP "cd /opt/ticketing-system && docker-compose restart"
```

## Files Created

- `demo-deploy.sh` - Deploy infrastructure
- `deploy-app.sh` - Deploy application
- `cleanup-demo.sh` - Delete everything
- `deployment-info.txt` - Your deployment details (generated)
- `ticketing-demo-key.pem` - SSH key (generated)

## Full Documentation

For production deployment with AWS WAF, load balancers, and enterprise security:
- See [AWS_FIREWALL_DEPLOYMENT_GUIDE.md](AWS_FIREWALL_DEPLOYMENT_GUIDE.md)
- See [DEMO_DEPLOYMENT_GUIDE.md](DEMO_DEPLOYMENT_GUIDE.md)

## Security Notes

⚠️ **For demo/development only**

For production:
1. Change database password
2. Use AWS RDS instead of Docker
3. Add HTTPS/SSL
4. Enable CloudWatch monitoring
5. Follow AWS_FIREWALL_DEPLOYMENT_GUIDE.md

---

**Ready to deploy? Run `./demo-deploy.sh` now! 🚀**
