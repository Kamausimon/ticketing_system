#!/bin/bash
set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Ticketing System - Demo Deployment${NC}"
echo -e "${GREEN}========================================${NC}\n"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI not found. Installing...${NC}"
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip -q awscliv2.zip
    sudo ./aws/install
    rm -rf aws awscliv2.zip
fi

# Check AWS credentials
echo -e "${YELLOW}Checking AWS credentials...${NC}"
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}AWS credentials not configured. Please run 'aws configure' first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ AWS credentials configured${NC}\n"

# Set variables
export AWS_REGION="${AWS_REGION:-us-east-1}"
export PROJECT_NAME="ticketing-demo"
export KEY_PAIR_NAME="${KEY_PAIR_NAME:-ticketing-demo-key}"

echo -e "${YELLOW}Configuration:${NC}"
echo "  AWS Region: $AWS_REGION"
echo "  Project: $PROJECT_NAME"
echo "  Key Pair: $KEY_PAIR_NAME"
echo ""

# Get your current IP for SSH access
echo -e "${YELLOW}Getting your IP address...${NC}"
YOUR_IP=$(curl -s ifconfig.me)
echo -e "${GREEN}✓ Your IP: $YOUR_IP${NC}\n"

# Create key pair if it doesn't exist
echo -e "${YELLOW}Setting up SSH key pair...${NC}"
if ! aws ec2 describe-key-pairs --key-names "$KEY_PAIR_NAME" --region "$AWS_REGION" &> /dev/null; then
    aws ec2 create-key-pair \
        --key-name "$KEY_PAIR_NAME" \
        --region "$AWS_REGION" \
        --query 'KeyMaterial' \
        --output text > ${KEY_PAIR_NAME}.pem
    chmod 400 ${KEY_PAIR_NAME}.pem
    echo -e "${GREEN}✓ Created key pair: ${KEY_PAIR_NAME}.pem${NC}"
else
    echo -e "${GREEN}✓ Key pair already exists${NC}"
fi

# Create VPC
echo -e "\n${YELLOW}Creating VPC...${NC}"
VPC_ID=$(aws ec2 create-vpc \
    --cidr-block 10.0.0.0/16 \
    --region "$AWS_REGION" \
    --tag-specifications "ResourceType=vpc,Tags=[{Key=Name,Value=$PROJECT_NAME-vpc}]" \
    --query 'Vpc.VpcId' \
    --output text 2>/dev/null || aws ec2 describe-vpcs \
    --filters "Name=tag:Name,Values=$PROJECT_NAME-vpc" \
    --query 'Vpcs[0].VpcId' \
    --output text)

aws ec2 modify-vpc-attribute --vpc-id "$VPC_ID" --enable-dns-hostnames --region "$AWS_REGION"
echo -e "${GREEN}✓ VPC ID: $VPC_ID${NC}"

# Create Internet Gateway
echo -e "${YELLOW}Creating Internet Gateway...${NC}"
IGW_ID=$(aws ec2 create-internet-gateway \
    --region "$AWS_REGION" \
    --tag-specifications "ResourceType=internet-gateway,Tags=[{Key=Name,Value=$PROJECT_NAME-igw}]" \
    --query 'InternetGateway.InternetGatewayId' \
    --output text 2>/dev/null || aws ec2 describe-internet-gateways \
    --filters "Name=tag:Name,Values=$PROJECT_NAME-igw" \
    --query 'InternetGateways[0].InternetGatewayId' \
    --output text)

aws ec2 attach-internet-gateway --internet-gateway-id "$IGW_ID" --vpc-id "$VPC_ID" --region "$AWS_REGION" 2>/dev/null || true
echo -e "${GREEN}✓ Internet Gateway: $IGW_ID${NC}"

# Create Public Subnet
echo -e "${YELLOW}Creating Public Subnet...${NC}"
PUBLIC_SUBNET=$(aws ec2 create-subnet \
    --vpc-id "$VPC_ID" \
    --cidr-block 10.0.1.0/24 \
    --availability-zone ${AWS_REGION}a \
    --region "$AWS_REGION" \
    --tag-specifications "ResourceType=subnet,Tags=[{Key=Name,Value=$PROJECT_NAME-public-subnet}]" \
    --query 'Subnet.SubnetId' \
    --output text 2>/dev/null || aws ec2 describe-subnets \
    --filters "Name=tag:Name,Values=$PROJECT_NAME-public-subnet" \
    --query 'Subnets[0].SubnetId' \
    --output text)

aws ec2 modify-subnet-attribute --subnet-id "$PUBLIC_SUBNET" --map-public-ip-on-launch --region "$AWS_REGION"
echo -e "${GREEN}✓ Public Subnet: $PUBLIC_SUBNET${NC}"

# Create Route Table
echo -e "${YELLOW}Creating Route Table...${NC}"
ROUTE_TABLE=$(aws ec2 create-route-table \
    --vpc-id "$VPC_ID" \
    --region "$AWS_REGION" \
    --tag-specifications "ResourceType=route-table,Tags=[{Key=Name,Value=$PROJECT_NAME-public-rt}]" \
    --query 'RouteTable.RouteTableId' \
    --output text 2>/dev/null || aws ec2 describe-route-tables \
    --filters "Name=tag:Name,Values=$PROJECT_NAME-public-rt" \
    --query 'RouteTables[0].RouteTableId' \
    --output text)

aws ec2 create-route --route-table-id "$ROUTE_TABLE" --destination-cidr-block 0.0.0.0/0 --gateway-id "$IGW_ID" --region "$AWS_REGION" 2>/dev/null || true
aws ec2 associate-route-table --route-table-id "$ROUTE_TABLE" --subnet-id "$PUBLIC_SUBNET" --region "$AWS_REGION" 2>/dev/null || true
echo -e "${GREEN}✓ Route Table: $ROUTE_TABLE${NC}"

# Create Security Group
echo -e "${YELLOW}Creating Security Group...${NC}"
SG_ID=$(aws ec2 create-security-group \
    --group-name "$PROJECT_NAME-sg" \
    --description "Security group for $PROJECT_NAME" \
    --vpc-id "$VPC_ID" \
    --region "$AWS_REGION" \
    --query 'GroupId' \
    --output text 2>/dev/null || aws ec2 describe-security-groups \
    --filters "Name=group-name,Values=$PROJECT_NAME-sg" "Name=vpc-id,Values=$VPC_ID" \
    --query 'SecurityGroups[0].GroupId' \
    --output text)

# Add security group rules
aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 22 --cidr ${YOUR_IP}/32 --region "$AWS_REGION" 2>/dev/null || true
aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 80 --cidr 0.0.0.0/0 --region "$AWS_REGION" 2>/dev/null || true
aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 443 --cidr 0.0.0.0/0 --region "$AWS_REGION" 2>/dev/null || true
aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 8080 --cidr 0.0.0.0/0 --region "$AWS_REGION" 2>/dev/null || true
echo -e "${GREEN}✓ Security Group: $SG_ID${NC}"

# Get latest Amazon Linux 2 AMI
echo -e "${YELLOW}Finding latest Amazon Linux 2 AMI...${NC}"
AMI_ID=$(aws ec2 describe-images \
    --owners amazon \
    --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" "Name=state,Values=available" \
    --region "$AWS_REGION" \
    --query 'sort_by(Images, &CreationDate)[-1].ImageId' \
    --output text)
echo -e "${GREEN}✓ AMI ID: $AMI_ID${NC}"

# Create user data script
cat > user-data.sh << 'USERDATA'
#!/bin/bash
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "Starting setup at $(date)"

# Update system
yum update -y

# Install Go
cd /tmp
wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/ec2-user/.bashrc
source /etc/profile

# Install Git and other tools
yum install -y git postgresql15

# Install Docker
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create application directory
mkdir -p /opt/ticketing-system
chown ec2-user:ec2-user /opt/ticketing-system

# Create demo docker-compose file
cat > /opt/ticketing-system/docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ticketing_db
      POSTGRES_USER: ticketing_user
      POSTGRES_PASSWORD: ChangeMe123!
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ticketing_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
EOF

# Start database services
cd /opt/ticketing-system
docker-compose up -d

# Wait for services to be ready
sleep 30

echo "Setup completed at $(date)"
echo "Status: ready" > /tmp/setup-status
USERDATA

# Launch EC2 Instance
echo -e "\n${YELLOW}Launching EC2 Instance...${NC}"
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id "$AMI_ID" \
    --instance-type t3.medium \
    --key-name "$KEY_PAIR_NAME" \
    --security-group-ids "$SG_ID" \
    --subnet-id "$PUBLIC_SUBNET" \
    --user-data file://user-data.sh \
    --region "$AWS_REGION" \
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$PROJECT_NAME-server}]" \
    --query 'Instances[0].InstanceId' \
    --output text)

echo -e "${GREEN}✓ Instance ID: $INSTANCE_ID${NC}"
echo -e "${YELLOW}Waiting for instance to be running...${NC}"

aws ec2 wait instance-running --instance-ids "$INSTANCE_ID" --region "$AWS_REGION"

# Get instance public IP
PUBLIC_IP=$(aws ec2 describe-instances \
    --instance-ids "$INSTANCE_ID" \
    --region "$AWS_REGION" \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text)

echo -e "${GREEN}✓ Instance is running!${NC}"
echo -e "${GREEN}✓ Public IP: $PUBLIC_IP${NC}"

# Save deployment info
cat > deployment-info.txt << EOF
Deployment Information
=====================
Date: $(date)
Region: $AWS_REGION
VPC ID: $VPC_ID
Subnet ID: $PUBLIC_SUBNET
Security Group: $SG_ID
Instance ID: $INSTANCE_ID
Public IP: $PUBLIC_IP
SSH Key: ${KEY_PAIR_NAME}.pem

SSH Access:
-----------
ssh -i ${KEY_PAIR_NAME}.pem ec2-user@$PUBLIC_IP

Database Connection (from instance):
------------------------------------
Host: localhost
Port: 5432
Database: ticketing_db
User: ticketing_user
Password: ChangeMe123!

Redis Connection (from instance):
---------------------------------
Host: localhost
Port: 6379

API Endpoint (after deployment):
--------------------------------
http://$PUBLIC_IP:8080

Next Steps:
-----------
1. Wait 2-3 minutes for setup to complete
2. SSH into the instance
3. Clone your repository in /opt/ticketing-system
4. Build and run your application
5. Access it at http://$PUBLIC_IP:8080

To check setup status:
ssh -i ${KEY_PAIR_NAME}.pem ec2-user@$PUBLIC_IP "cat /tmp/setup-status"
EOF

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Deployment Successful!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Instance is initializing (takes ~3 minutes)${NC}"
echo ""
echo -e "Public IP: ${GREEN}$PUBLIC_IP${NC}"
echo -e "SSH Command: ${GREEN}ssh -i ${KEY_PAIR_NAME}.pem ec2-user@$PUBLIC_IP${NC}"
echo ""
echo -e "Deployment details saved to: ${GREEN}deployment-info.txt${NC}"
echo ""
echo -e "${YELLOW}Next: Deploy your application (see deploy-app.sh)${NC}"
