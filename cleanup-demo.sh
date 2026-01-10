#!/bin/bash
set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Cleanup Demo Deployment${NC}"
echo -e "${YELLOW}========================================${NC}\n"

# Check if deployment-info.txt exists
if [ ! -f "deployment-info.txt" ]; then
    echo -e "${RED}Error: deployment-info.txt not found${NC}"
    echo "Nothing to cleanup."
    exit 0
fi

# Extract deployment info
AWS_REGION=$(grep "Region:" deployment-info.txt | awk '{print $2}')
VPC_ID=$(grep "VPC ID:" deployment-info.txt | awk '{print $3}')
INSTANCE_ID=$(grep "Instance ID:" deployment-info.txt | awk '{print $3}')
SG_ID=$(grep "Security Group:" deployment-info.txt | awk '{print $3}')

echo -e "${YELLOW}This will delete:${NC}"
echo "  - EC2 Instance: $INSTANCE_ID"
echo "  - VPC: $VPC_ID"
echo "  - All associated resources"
echo ""
read -p "Are you sure? (type 'yes' to confirm): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cleanup cancelled."
    exit 0
fi

echo ""
echo -e "${YELLOW}Starting cleanup...${NC}\n"

# Terminate EC2 instance
if [ ! -z "$INSTANCE_ID" ]; then
    echo -e "${YELLOW}Terminating EC2 instance...${NC}"
    aws ec2 terminate-instances --instance-ids "$INSTANCE_ID" --region "$AWS_REGION" &> /dev/null || true
    echo "Waiting for instance to terminate..."
    aws ec2 wait instance-terminated --instance-ids "$INSTANCE_ID" --region "$AWS_REGION" 2>/dev/null || true
    echo -e "${GREEN}✓ Instance terminated${NC}"
fi

# Wait a bit for resources to detach
sleep 5

# Delete security group
if [ ! -z "$SG_ID" ]; then
    echo -e "${YELLOW}Deleting security group...${NC}"
    aws ec2 delete-security-group --group-id "$SG_ID" --region "$AWS_REGION" 2>/dev/null || true
    echo -e "${GREEN}✓ Security group deleted${NC}"
fi

# Get and delete all subnets in VPC
echo -e "${YELLOW}Deleting subnets...${NC}"
SUBNETS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --region "$AWS_REGION" --query 'Subnets[].SubnetId' --output text 2>/dev/null || echo "")
for subnet in $SUBNETS; do
    aws ec2 delete-subnet --subnet-id "$subnet" --region "$AWS_REGION" 2>/dev/null || true
done
echo -e "${GREEN}✓ Subnets deleted${NC}"

# Detach and delete internet gateway
echo -e "${YELLOW}Deleting internet gateway...${NC}"
IGW_ID=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$VPC_ID" --region "$AWS_REGION" --query 'InternetGateways[0].InternetGatewayId' --output text 2>/dev/null || echo "")
if [ ! -z "$IGW_ID" ] && [ "$IGW_ID" != "None" ]; then
    aws ec2 detach-internet-gateway --internet-gateway-id "$IGW_ID" --vpc-id "$VPC_ID" --region "$AWS_REGION" 2>/dev/null || true
    aws ec2 delete-internet-gateway --internet-gateway-id "$IGW_ID" --region "$AWS_REGION" 2>/dev/null || true
fi
echo -e "${GREEN}✓ Internet gateway deleted${NC}"

# Delete route tables (except main)
echo -e "${YELLOW}Deleting route tables...${NC}"
ROUTE_TABLES=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --region "$AWS_REGION" --query 'RouteTables[?Associations[0].Main!=`true`].RouteTableId' --output text 2>/dev/null || echo "")
for rt in $ROUTE_TABLES; do
    # Disassociate first
    ASSOC_IDS=$(aws ec2 describe-route-tables --route-table-ids "$rt" --region "$AWS_REGION" --query 'RouteTables[0].Associations[].RouteTableAssociationId' --output text 2>/dev/null || echo "")
    for assoc in $ASSOC_IDS; do
        aws ec2 disassociate-route-table --association-id "$assoc" --region "$AWS_REGION" 2>/dev/null || true
    done
    aws ec2 delete-route-table --route-table-id "$rt" --region "$AWS_REGION" 2>/dev/null || true
done
echo -e "${GREEN}✓ Route tables deleted${NC}"

# Delete VPC
echo -e "${YELLOW}Deleting VPC...${NC}"
aws ec2 delete-vpc --vpc-id "$VPC_ID" --region "$AWS_REGION" 2>/dev/null || true
echo -e "${GREEN}✓ VPC deleted${NC}"

# Remove deployment files
echo -e "${YELLOW}Cleaning up local files...${NC}"
rm -f deployment-info.txt user-data.sh
echo -e "${GREEN}✓ Local files cleaned${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Cleanup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "All demo resources have been deleted."
echo ""
echo "Note: SSH key pair was NOT deleted. To remove it:"
echo -e "${YELLOW}aws ec2 delete-key-pair --key-name ticketing-demo-key --region $AWS_REGION${NC}"
echo -e "${YELLOW}rm -f ticketing-demo-key.pem${NC}"
echo ""
