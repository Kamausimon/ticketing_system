#!/bin/bash

# Email System Setup and Test Script
# This script helps you set up and test the email system

set -e

echo "📧 Email System Setup & Test"
echo "============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo "Creating .env from .env.example..."
    
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✅ .env file created${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  Please update the following in your .env file:${NC}"
        echo "   - EMAIL_USERNAME (your Mailtrap username)"
        echo "   - EMAIL_PASSWORD (your Mailtrap password)"
        echo ""
        echo "Get your credentials from: https://mailtrap.io"
        echo ""
        read -p "Press Enter after updating .env file..."
    else
        echo -e "${RED}❌ .env.example not found${NC}"
        exit 1
    fi
fi

# Check if required email variables are set
echo "🔍 Checking email configuration..."
source .env

if [ -z "$EMAIL_USERNAME" ] || [ -z "$EMAIL_PASSWORD" ]; then
    echo -e "${RED}❌ EMAIL_USERNAME or EMAIL_PASSWORD not set in .env${NC}"
    echo ""
    echo "Please add these to your .env file:"
    echo "EMAIL_USERNAME=your_mailtrap_username"
    echo "EMAIL_PASSWORD=your_mailtrap_password"
    exit 1
fi

echo -e "${GREEN}✅ Email configuration found${NC}"
echo "   Provider: $EMAIL_PROVIDER"
echo "   Host: $EMAIL_HOST"
echo "   Port: $EMAIL_PORT"
echo "   From: $EMAIL_FROM"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed${NC}"
go version
echo ""

# Run the integration example
echo "🚀 Running email integration example..."
echo ""

if [ -f "examples/email_integration.go" ]; then
    go run examples/email_integration.go
    
    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✅ Email test completed successfully!${NC}"
        echo ""
        echo "📬 Check your Mailtrap inbox at: https://mailtrap.io/inboxes"
        echo ""
        echo "You should see:"
        echo "  - Test email"
        echo "  - Welcome email"
        echo "  - Verification email"
        echo "  - Password reset email"
        echo "  - Order confirmation"
        echo "  - Ticket generated email"
    else
        echo ""
        echo -e "${RED}❌ Email test failed${NC}"
        echo ""
        echo "Troubleshooting:"
        echo "  1. Check your .env file has correct credentials"
        echo "  2. Verify your Mailtrap account is active"
        echo "  3. Check the error messages above"
        exit 1
    fi
else
    echo -e "${RED}❌ examples/email_integration.go not found${NC}"
    exit 1
fi

echo ""
echo "📚 Next Steps:"
echo "  1. Integrate email service into your handlers"
echo "  2. See EMAIL_QUICKSTART.md for integration guide"
echo "  3. For production, switch to Zoho or another provider"
echo ""
echo "🎉 Setup complete!"
