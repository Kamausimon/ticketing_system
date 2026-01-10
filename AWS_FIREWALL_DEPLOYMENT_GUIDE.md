# AWS Firewall & Backend Deployment Guide

## Overview
This guide provides comprehensive steps to deploy your Go backend ticketing system to AWS with proper firewall configuration using Security Groups, Network ACLs, and AWS WAF.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Prerequisites](#prerequisites)
3. [VPC Setup](#vpc-setup)
4. [Security Groups Configuration](#security-groups-configuration)
5. [EC2 Instance Setup](#ec2-instance-setup)
6. [Application Load Balancer (ALB) Setup](#application-load-balancer-setup)
7. [AWS WAF Configuration](#aws-waf-configuration)
8. [Database Security (RDS)](#database-security)
9. [Redis Cache Security](#redis-cache-security)
10. [Demo Setup](#demo-setup)
11. [Monitoring & Logging](#monitoring--logging)
12. [Best Practices](#best-practices)

---

## Architecture Overview

```
Internet
    ↓
 AWS WAF (Web Application Firewall)
    ↓
Application Load Balancer (Public Subnet)
    ↓
EC2 Instances (Private Subnet) - Go Backend
    ↓
RDS PostgreSQL (Private Subnet)
Redis ElastiCache (Private Subnet)
```

---

## Prerequisites

### Required Tools
```bash
# AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Configure AWS credentials
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Default region: us-east-1
# Default output format: json

# Verify installation
aws --version
```

### Required AWS Permissions
- EC2 (Full Access)
- VPC (Full Access)
- RDS (Full Access)
- ElastiCache (Full Access)
- WAF (Full Access)
- CloudWatch (Full Access)
- IAM (Create roles and policies)

---

## VPC Setup

### Step 1: Create VPC

```bash
# Create VPC
VPC_ID=$(aws ec2 create-vpc \
  --cidr-block 10.0.0.0/16 \
  --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=ticketing-system-vpc}]' \
  --query 'Vpc.VpcId' \
  --output text)

echo "VPC ID: $VPC_ID"

# Enable DNS hostname
aws ec2 modify-vpc-attribute \
  --vpc-id $VPC_ID \
  --enable-dns-hostnames
```

### Step 2: Create Subnets

```bash
# Public Subnet 1 (for ALB)
PUBLIC_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.1.0/24 \
  --availability-zone us-east-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=public-subnet-1}]' \
  --query 'Subnet.SubnetId' \
  --output text)

# Public Subnet 2 (for ALB - required for high availability)
PUBLIC_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.2.0/24 \
  --availability-zone us-east-1b \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=public-subnet-2}]' \
  --query 'Subnet.SubnetId' \
  --output text)

# Private Subnet 1 (for application servers)
PRIVATE_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.10.0/24 \
  --availability-zone us-east-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=private-subnet-1}]' \
  --query 'Subnet.SubnetId' \
  --output text)

# Private Subnet 2 (for application servers)
PRIVATE_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.11.0/24 \
  --availability-zone us-east-1b \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=private-subnet-2}]' \
  --query 'Subnet.SubnetId' \
  --output text)

# Database Subnet 1
DB_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.20.0/24 \
  --availability-zone us-east-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=db-subnet-1}]' \
  --query 'Subnet.SubnetId' \
  --output text)

# Database Subnet 2
DB_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.21.0/24 \
  --availability-zone us-east-1b \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=db-subnet-2}]' \
  --query 'Subnet.SubnetId' \
  --output text)

echo "Public Subnet 1: $PUBLIC_SUBNET_1"
echo "Public Subnet 2: $PUBLIC_SUBNET_2"
echo "Private Subnet 1: $PRIVATE_SUBNET_1"
echo "Private Subnet 2: $PRIVATE_SUBNET_2"
```

### Step 3: Internet Gateway & NAT Gateway

```bash
# Create Internet Gateway
IGW_ID=$(aws ec2 create-internet-gateway \
  --tag-specifications 'ResourceType=internet-gateway,Tags=[{Key=Name,Value=ticketing-igw}]' \
  --query 'InternetGateway.InternetGatewayId' \
  --output text)

# Attach to VPC
aws ec2 attach-internet-gateway \
  --internet-gateway-id $IGW_ID \
  --vpc-id $VPC_ID

# Allocate Elastic IP for NAT Gateway
EIP_ALLOC=$(aws ec2 allocate-address \
  --domain vpc \
  --query 'AllocationId' \
  --output text)

# Create NAT Gateway in Public Subnet
NAT_GW=$(aws ec2 create-nat-gateway \
  --subnet-id $PUBLIC_SUBNET_1 \
  --allocation-id $EIP_ALLOC \
  --tag-specifications 'ResourceType=nat-gateway,Tags=[{Key=Name,Value=ticketing-nat}]' \
  --query 'NatGateway.NatGatewayId' \
  --output text)

echo "NAT Gateway: $NAT_GW (wait for it to become available)"
aws ec2 wait nat-gateway-available --nat-gateway-ids $NAT_GW
```

### Step 4: Route Tables

```bash
# Public Route Table
PUBLIC_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=route-table,Tags=[{Key=Name,Value=public-rt}]' \
  --query 'RouteTable.RouteTableId' \
  --output text)

# Add route to Internet Gateway
aws ec2 create-route \
  --route-table-id $PUBLIC_RT \
  --destination-cidr-block 0.0.0.0/0 \
  --gateway-id $IGW_ID

# Associate public subnets
aws ec2 associate-route-table --route-table-id $PUBLIC_RT --subnet-id $PUBLIC_SUBNET_1
aws ec2 associate-route-table --route-table-id $PUBLIC_RT --subnet-id $PUBLIC_SUBNET_2

# Private Route Table
PRIVATE_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=route-table,Tags=[{Key=Name,Value=private-rt}]' \
  --query 'RouteTable.RouteTableId' \
  --output text)

# Add route to NAT Gateway
aws ec2 create-route \
  --route-table-id $PRIVATE_RT \
  --destination-cidr-block 0.0.0.0/0 \
  --nat-gateway-id $NAT_GW

# Associate private subnets
aws ec2 associate-route-table --route-table-id $PRIVATE_RT --subnet-id $PRIVATE_SUBNET_1
aws ec2 associate-route-table --route-table-id $PRIVATE_RT --subnet-id $PRIVATE_SUBNET_2
```

---

## Security Groups Configuration

### Step 5: Create Security Groups

```bash
# ALB Security Group (allows HTTP/HTTPS from internet)
ALB_SG=$(aws ec2 create-security-group \
  --group-name alb-security-group \
  --description "Security group for Application Load Balancer" \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=security-group,Tags=[{Key=Name,Value=alb-sg}]' \
  --query 'GroupId' \
  --output text)

# Allow HTTP from anywhere
aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0

# Allow HTTPS from anywhere
aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0

# Application Security Group (allows traffic from ALB only)
APP_SG=$(aws ec2 create-security-group \
  --group-name app-security-group \
  --description "Security group for application servers" \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=security-group,Tags=[{Key=Name,Value=app-sg}]' \
  --query 'GroupId' \
  --output text)

# Allow traffic from ALB on port 8080 (your Go app port)
aws ec2 authorize-security-group-ingress \
  --group-id $APP_SG \
  --protocol tcp \
  --port 8080 \
  --source-group $ALB_SG

# Allow SSH from your IP (for management)
# Replace YOUR_IP with your actual IP address
aws ec2 authorize-security-group-ingress \
  --group-id $APP_SG \
  --protocol tcp \
  --port 22 \
  --cidr YOUR_IP/32

# Database Security Group (allows traffic from application only)
DB_SG=$(aws ec2 create-security-group \
  --group-name db-security-group \
  --description "Security group for RDS PostgreSQL" \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=security-group,Tags=[{Key=Name,Value=db-sg}]' \
  --query 'GroupId' \
  --output text)

# Allow PostgreSQL from application servers
aws ec2 authorize-security-group-ingress \
  --group-id $DB_SG \
  --protocol tcp \
  --port 5432 \
  --source-group $APP_SG

# Redis Security Group
REDIS_SG=$(aws ec2 create-security-group \
  --group-name redis-security-group \
  --description "Security group for Redis ElastiCache" \
  --vpc-id $VPC_ID \
  --tag-specifications 'ResourceType=security-group,Tags=[{Key=Name,Value=redis-sg}]' \
  --query 'GroupId' \
  --output text)

# Allow Redis from application servers
aws ec2 authorize-security-group-ingress \
  --group-id $REDIS_SG \
  --protocol tcp \
  --port 6379 \
  --source-group $APP_SG

echo "Security Groups created:"
echo "ALB SG: $ALB_SG"
echo "APP SG: $APP_SG"
echo "DB SG: $DB_SG"
echo "Redis SG: $REDIS_SG"
```

---

## EC2 Instance Setup

### Step 6: Create IAM Role for EC2

```bash
# Create trust policy
cat > ec2-trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Create IAM role
aws iam create-role \
  --role-name ticketing-ec2-role \
  --assume-role-policy-document file://ec2-trust-policy.json

# Attach policies
aws iam attach-role-policy \
  --role-name ticketing-ec2-role \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore

aws iam attach-role-policy \
  --role-name ticketing-ec2-role \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy

# Create instance profile
aws iam create-instance-profile \
  --instance-profile-name ticketing-ec2-profile

aws iam add-role-to-instance-profile \
  --instance-profile-name ticketing-ec2-profile \
  --role-name ticketing-ec2-role
```

### Step 7: Launch EC2 Instance

```bash
# Create user data script
cat > user-data.sh << 'EOF'
#!/bin/bash
# Update system
yum update -y

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile

# Install Docker
yum install -y docker
systemctl start docker
systemctl enable docker

# Install Git
yum install -y git

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Create application directory
mkdir -p /opt/ticketing-system
chown ec2-user:ec2-user /opt/ticketing-system

echo "EC2 instance setup complete" > /tmp/setup-complete
EOF

# Launch EC2 instance
INSTANCE_ID=$(aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.medium \
  --key-name YOUR_KEY_PAIR_NAME \
  --security-group-ids $APP_SG \
  --subnet-id $PRIVATE_SUBNET_1 \
  --iam-instance-profile Name=ticketing-ec2-profile \
  --user-data file://user-data.sh \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=ticketing-app-server}]' \
  --query 'Instances[0].InstanceId' \
  --output text)

echo "EC2 Instance ID: $INSTANCE_ID"
aws ec2 wait instance-running --instance-ids $INSTANCE_ID
```

---

## Application Load Balancer Setup

### Step 8: Create Target Group

```bash
# Create target group
TG_ARN=$(aws elbv2 create-target-group \
  --name ticketing-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id $VPC_ID \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3 \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text)

# Register EC2 instance
aws elbv2 register-targets \
  --target-group-arn $TG_ARN \
  --targets Id=$INSTANCE_ID

echo "Target Group ARN: $TG_ARN"
```

### Step 9: Create Application Load Balancer

```bash
# Create ALB
ALB_ARN=$(aws elbv2 create-load-balancer \
  --name ticketing-alb \
  --subnets $PUBLIC_SUBNET_1 $PUBLIC_SUBNET_2 \
  --security-groups $ALB_SG \
  --scheme internet-facing \
  --type application \
  --ip-address-type ipv4 \
  --tags Key=Name,Value=ticketing-alb \
  --query 'LoadBalancers[0].LoadBalancerArn' \
  --output text)

# Get ALB DNS name
ALB_DNS=$(aws elbv2 describe-load-balancers \
  --load-balancer-arns $ALB_ARN \
  --query 'LoadBalancers[0].DNSName' \
  --output text)

echo "ALB DNS: $ALB_DNS"

# Create HTTP listener
LISTENER_ARN=$(aws elbv2 create-listener \
  --load-balancer-arn $ALB_ARN \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=$TG_ARN \
  --query 'Listeners[0].ListenerArn' \
  --output text)

echo "Listener ARN: $LISTENER_ARN"
```

---

## AWS WAF Configuration

### Step 10: Create WAF Web ACL

```bash
# Create WAF Web ACL
cat > waf-rules.json << 'EOF'
{
  "Name": "ticketing-waf",
  "Scope": "REGIONAL",
  "DefaultAction": {
    "Allow": {}
  },
  "Description": "WAF for ticketing system",
  "Rules": [
    {
      "Name": "RateLimitRule",
      "Priority": 1,
      "Statement": {
        "RateBasedStatement": {
          "Limit": 2000,
          "AggregateKeyType": "IP"
        }
      },
      "Action": {
        "Block": {}
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "RateLimitRule"
      }
    },
    {
      "Name": "AWSManagedRulesCommonRuleSet",
      "Priority": 2,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet"
        }
      },
      "OverrideAction": {
        "None": {}
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AWSManagedRulesCommonRuleSetMetric"
      }
    },
    {
      "Name": "AWSManagedRulesKnownBadInputsRuleSet",
      "Priority": 3,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesKnownBadInputsRuleSet"
        }
      },
      "OverrideAction": {
        "None": {}
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AWSManagedRulesKnownBadInputsRuleSetMetric"
      }
    },
    {
      "Name": "SQLInjectionProtection",
      "Priority": 4,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesSQLiRuleSet"
        }
      },
      "OverrideAction": {
        "None": {}
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLInjectionProtectionMetric"
      }
    }
  ],
  "VisibilityConfig": {
    "SampledRequestsEnabled": true,
    "CloudWatchMetricsEnabled": true,
    "MetricName": "ticketing-waf"
  }
}
EOF

# Create Web ACL
WAF_ARN=$(aws wafv2 create-web-acl \
  --name ticketing-waf \
  --scope REGIONAL \
  --region us-east-1 \
  --default-action Allow={} \
  --description "WAF for ticketing system" \
  --rules file://waf-rules.json \
  --visibility-config SampledRequestsEnabled=true,CloudWatchMetricsEnabled=true,MetricName=ticketing-waf \
  --query 'Summary.ARN' \
  --output text)

# Associate WAF with ALB
aws wafv2 associate-web-acl \
  --web-acl-arn $WAF_ARN \
  --resource-arn $ALB_ARN \
  --region us-east-1

echo "WAF ARN: $WAF_ARN"
```

---

## Database Security

### Step 11: Create RDS PostgreSQL Instance

```bash
# Create DB subnet group
aws rds create-db-subnet-group \
  --db-subnet-group-name ticketing-db-subnet-group \
  --db-subnet-group-description "Subnet group for ticketing DB" \
  --subnet-ids $DB_SUBNET_1 $DB_SUBNET_2

# Create RDS instance
aws rds create-db-instance \
  --db-instance-identifier ticketing-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 15.4 \
  --master-username dbadmin \
  --master-user-password 'YOUR_STRONG_PASSWORD' \
  --allocated-storage 20 \
  --vpc-security-group-ids $DB_SG \
  --db-subnet-group-name ticketing-db-subnet-group \
  --backup-retention-period 7 \
  --storage-encrypted \
  --no-publicly-accessible \
  --enable-cloudwatch-logs-exports '["postgresql"]'

# Wait for DB to be available
aws rds wait db-instance-available --db-instance-identifier ticketing-db

# Get DB endpoint
DB_ENDPOINT=$(aws rds describe-db-instances \
  --db-instance-identifier ticketing-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text)

echo "Database Endpoint: $DB_ENDPOINT"
```

---

## Redis Cache Security

### Step 12: Create ElastiCache Redis Cluster

```bash
# Create cache subnet group
aws elasticache create-cache-subnet-group \
  --cache-subnet-group-name ticketing-redis-subnet-group \
  --cache-subnet-group-description "Subnet group for Redis" \
  --subnet-ids $PRIVATE_SUBNET_1 $PRIVATE_SUBNET_2

# Create Redis cluster
aws elasticache create-cache-cluster \
  --cache-cluster-id ticketing-redis \
  --cache-node-type cache.t3.micro \
  --engine redis \
  --engine-version 7.0 \
  --num-cache-nodes 1 \
  --cache-subnet-group-name ticketing-redis-subnet-group \
  --security-group-ids $REDIS_SG \
  --snapshot-retention-limit 5

# Wait for cache cluster
aws elasticache wait cache-cluster-available --cache-cluster-id ticketing-redis

# Get Redis endpoint
REDIS_ENDPOINT=$(aws elasticache describe-cache-clusters \
  --cache-cluster-id ticketing-redis \
  --show-cache-node-info \
  --query 'CacheClusters[0].CacheNodes[0].Endpoint.Address' \
  --output text)

echo "Redis Endpoint: $REDIS_ENDPOINT"
```

---

## Demo Setup

### Step 13: Deploy Application

Create a deployment script on your EC2 instance:

```bash
#!/bin/bash
# deploy-app.sh

# Set environment variables
export DB_HOST="$DB_ENDPOINT"
export DB_PORT="5432"
export DB_USER="dbadmin"
export DB_PASSWORD="YOUR_STRONG_PASSWORD"
export DB_NAME="ticketing_db"
export REDIS_HOST="$REDIS_ENDPOINT"
export REDIS_PORT="6379"
export PORT="8080"

# Clone repository
cd /opt/ticketing-system
git clone YOUR_REPO_URL .

# Build application
go build -o api-server ./cmd/api-server

# Create systemd service
cat > /etc/systemd/system/ticketing-app.service << 'EOF'
[Unit]
Description=Ticketing System API Server
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/ticketing-system
ExecStart=/opt/ticketing-system/api-server
Restart=on-failure
RestartSec=5s
Environment="DB_HOST=$DB_ENDPOINT"
Environment="DB_PORT=5432"
Environment="REDIS_HOST=$REDIS_ENDPOINT"
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl start ticketing-app
systemctl enable ticketing-app
```

### Step 14: Test the Setup

```bash
# Test ALB health
curl http://$ALB_DNS/health

# Test API endpoint
curl http://$ALB_DNS/api/events

# Monitor logs
journalctl -u ticketing-app -f
```

---

## Monitoring & Logging

### Step 15: CloudWatch Setup

```bash
# Create log group
aws logs create-log-group --log-group-name /aws/ticketing-system

# Create CloudWatch alarms
aws cloudwatch put-metric-alarm \
  --alarm-name high-cpu-utilization \
  --alarm-description "Alert when CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/EC2 \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --dimensions Name=InstanceId,Value=$INSTANCE_ID

# Monitor ALB
aws cloudwatch put-metric-alarm \
  --alarm-name alb-unhealthy-hosts \
  --alarm-description "Alert when unhealthy host count > 0" \
  --metric-name UnHealthyHostCount \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 60 \
  --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --evaluation-periods 2
```

---

## Best Practices

### Security Best Practices
1. **Least Privilege**: Only allow necessary ports and IPs
2. **Encryption**: Enable encryption at rest and in transit
3. **Secrets Management**: Use AWS Secrets Manager for sensitive data
4. **Regular Updates**: Keep OS and dependencies updated
5. **MFA**: Enable MFA for AWS account access
6. **IAM Roles**: Use IAM roles instead of access keys

### Network Security
```bash
# Example: Using AWS Secrets Manager
aws secretsmanager create-secret \
  --name ticketing/db/password \
  --secret-string '{"username":"dbadmin","password":"YOUR_PASSWORD"}'

# Retrieve in application
aws secretsmanager get-secret-value \
  --secret-id ticketing/db/password
```

### WAF Advanced Configuration
```bash
# Add geo-blocking (example: block all except US)
cat > geo-rule.json << 'EOF'
{
  "Name": "GeoBlockingRule",
  "Priority": 0,
  "Statement": {
    "NotStatement": {
      "Statement": {
        "GeoMatchStatement": {
          "CountryCodes": ["US"]
        }
      }
    }
  },
  "Action": {
    "Block": {}
  },
  "VisibilityConfig": {
    "SampledRequestsEnabled": true,
    "CloudWatchMetricsEnabled": true,
    "MetricName": "GeoBlockingRule"
  }
}
EOF
```

### Auto Scaling Configuration
```bash
# Create launch template
aws ec2 create-launch-template \
  --launch-template-name ticketing-lt \
  --version-description "Version 1" \
  --launch-template-data file://launch-template.json

# Create auto scaling group
aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name ticketing-asg \
  --launch-template LaunchTemplateName=ticketing-lt \
  --min-size 2 \
  --max-size 6 \
  --desired-capacity 2 \
  --vpc-zone-identifier "$PRIVATE_SUBNET_1,$PRIVATE_SUBNET_2" \
  --target-group-arns $TG_ARN \
  --health-check-type ELB \
  --health-check-grace-period 300
```

---

## Cleanup (Demo Teardown)

```bash
# Delete Auto Scaling Group
aws autoscaling delete-auto-scaling-group \
  --auto-scaling-group-name ticketing-asg \
  --force-delete

# Terminate EC2 instances
aws ec2 terminate-instances --instance-ids $INSTANCE_ID

# Delete ALB
aws elbv2 delete-load-balancer --load-balancer-arn $ALB_ARN
aws elbv2 delete-target-group --target-group-arn $TG_ARN

# Delete RDS
aws rds delete-db-instance \
  --db-instance-identifier ticketing-db \
  --skip-final-snapshot

# Delete ElastiCache
aws elasticache delete-cache-cluster --cache-cluster-id ticketing-redis

# Delete WAF
aws wafv2 disassociate-web-acl --resource-arn $ALB_ARN
aws wafv2 delete-web-acl --name ticketing-waf --scope REGIONAL --id WAF_ID

# Delete NAT Gateway
aws ec2 delete-nat-gateway --nat-gateway-id $NAT_GW
aws ec2 release-address --allocation-id $EIP_ALLOC

# Delete Internet Gateway
aws ec2 detach-internet-gateway --internet-gateway-id $IGW_ID --vpc-id $VPC_ID
aws ec2 delete-internet-gateway --internet-gateway-id $IGW_ID

# Delete Subnets
aws ec2 delete-subnet --subnet-id $PUBLIC_SUBNET_1
aws ec2 delete-subnet --subnet-id $PUBLIC_SUBNET_2
aws ec2 delete-subnet --subnet-id $PRIVATE_SUBNET_1
aws ec2 delete-subnet --subnet-id $PRIVATE_SUBNET_2

# Delete Security Groups
aws ec2 delete-security-group --group-id $ALB_SG
aws ec2 delete-security-group --group-id $APP_SG
aws ec2 delete-security-group --group-id $DB_SG
aws ec2 delete-security-group --group-id $REDIS_SG

# Delete VPC
aws ec2 delete-vpc --vpc-id $VPC_ID
```

---

## Cost Estimation (Demo)

### Monthly Cost Breakdown
- **EC2 t3.medium**: ~$30/month
- **RDS db.t3.micro**: ~$15/month
- **ElastiCache cache.t3.micro**: ~$12/month
- **ALB**: ~$20/month
- **NAT Gateway**: ~$32/month
- **Data Transfer**: ~$10/month (varies)
- **WAF**: ~$5/month + $1 per million requests

**Total**: ~$124/month

### Free Tier Eligible (First 12 months)
- 750 hours EC2 t2.micro
- 750 hours RDS db.t2.micro
- 50GB data transfer out

---

## Troubleshooting

### Common Issues

1. **Can't connect to database**
   ```bash
   # Check security group rules
   aws ec2 describe-security-groups --group-ids $DB_SG
   
   # Test from EC2
   psql -h $DB_ENDPOINT -U dbadmin -d postgres
   ```

2. **ALB health checks failing**
   ```bash
   # Check target health
   aws elbv2 describe-target-health --target-group-arn $TG_ARN
   
   # Check application logs
   journalctl -u ticketing-app -n 100
   ```

3. **WAF blocking legitimate traffic**
   ```bash
   # Check WAF logs
   aws wafv2 list-logging-configurations --scope REGIONAL
   
   # Disable specific rules temporarily
   aws wafv2 update-web-acl ...
   ```

---

## Additional Resources

- [AWS WAF Developer Guide](https://docs.aws.amazon.com/waf/)
- [AWS VPC Documentation](https://docs.aws.amazon.com/vpc/)
- [AWS Security Best Practices](https://aws.amazon.com/architecture/security-identity-compliance/)
- [Go Deployment Best Practices](https://golang.org/doc/)

---

## Summary

This guide provided a complete walkthrough for deploying your Go backend with AWS firewall protection. Key security features include:

✅ VPC isolation with public/private subnets  
✅ Security Groups for fine-grained access control  
✅ AWS WAF for application-layer protection  
✅ Rate limiting and DDoS protection  
✅ Encrypted data at rest and in transit  
✅ Private database and cache instances  
✅ CloudWatch monitoring and alerting  

Your application is now production-ready with enterprise-grade security! 🚀
