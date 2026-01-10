# Quick Start Guide 🚀

## Running Tests

### Run All Tests
```bash
cd /home/kamau/projects/ticketing_system
go test -v ./internal/attendees/ ./internal/venues/
```

### With Coverage
```bash
go test -cover ./internal/attendees/ ./internal/venues/
```

### Generate HTML Report
```bash
go test -coverprofile=coverage.out ./internal/attendees/ ./internal/venues/
go tool cover -html=coverage.out -o coverage.html
xdg-open coverage.html
```

## AWS Deployment Quick Start

### 1. Install AWS CLI
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
aws configure
```

### 2. Set Environment Variables
```bash
export AWS_REGION="us-east-1"
export KEY_PAIR_NAME="your-key-pair"
export YOUR_IP="$(curl -s ifconfig.me)"
```

### 3. Create VPC
```bash
VPC_ID=$(aws ec2 create-vpc --cidr-block 10.0.0.0/16 --query 'Vpc.VpcId' --output text)
echo "VPC ID: $VPC_ID"
```

### 4. Follow Complete Guide
See [AWS_FIREWALL_DEPLOYMENT_GUIDE.md](AWS_FIREWALL_DEPLOYMENT_GUIDE.md) for detailed steps.

## Test Files

### Attendees Tests
- **File**: `internal/attendees/attendees_test.go`
- **Tests**: 30+ comprehensive tests
- **Status**: ✅ Passing

### Venues Tests
- **File**: `internal/venues/venues_test.go`
- **Tests**: 30+ comprehensive tests
- **Status**: ✅ Passing

## Documentation

1. **TESTS_AND_DEPLOYMENT_SUMMARY.md** - Complete overview
2. **AWS_FIREWALL_DEPLOYMENT_GUIDE.md** - AWS deployment steps
3. **TESTING_DEPLOYMENT_QUICKREF.md** - Testing commands
4. **This file** - Quick start

## Common Commands

### Test Specific Module
```bash
go test -v ./internal/attendees/ -run TestCheckInAttendee
go test -v ./internal/venues/ -run TestCreateVenue
```

### Check Test Coverage
```bash
go test -cover ./internal/attendees/
go test -cover ./internal/venues/
```

### Run Benchmarks
```bash
go test -bench=. ./internal/attendees/
```

## AWS Cost Estimate

- **Development**: ~$8/month (t3.micro)
- **Production**: ~$124/month (t3.medium + RDS + Redis + WAF)
- **Free Tier**: First 12 months eligible

## Support

For issues or questions:
1. Check TESTS_AND_DEPLOYMENT_SUMMARY.md
2. Review AWS_FIREWALL_DEPLOYMENT_GUIDE.md troubleshooting section
3. Review test error messages for specific failures

---

**Everything is ready to go! 🎉**
