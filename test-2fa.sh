#!/bin/bash

# Two-Factor Authentication Test Script
# This script helps test the 2FA implementation

set -e

API_URL="${API_URL:-http://localhost:8080}"
EMAIL="${TEST_EMAIL:-test@example.com}"
PASSWORD="${TEST_PASSWORD:-password123}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔐 Two-Factor Authentication Test Suite"
echo "========================================"
echo ""

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Step 1: Register a test user
echo "Step 1: Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST "${API_URL}/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"first_name\": \"Test\",
        \"last_name\": \"User\",
        \"username\": \"testuser_2fa\",
        \"phone\": \"+254700000000\",
        \"email\": \"${EMAIL}\",
        \"password\": \"${PASSWORD}\"
    }")

if echo "$REGISTER_RESPONSE" | grep -q "user_id"; then
    print_success "User registered successfully"
    USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"user_id":[0-9]*' | grep -o '[0-9]*')
    print_info "User ID: $USER_ID"
else
    print_info "User might already exist, continuing with login..."
fi

echo ""

# Step 2: Login to get token
echo "Step 2: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"${EMAIL}\",
        \"password\": \"${PASSWORD}\"
    }")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    print_success "Login successful"
    print_info "Token: ${TOKEN:0:50}..."
else
    print_error "Login failed"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo ""

# Step 3: Check initial 2FA status
echo "Step 3: Checking 2FA status..."
STATUS_RESPONSE=$(curl -s -X GET "${API_URL}/2fa/status" \
    -H "Authorization: Bearer $TOKEN")

print_info "Current status: $STATUS_RESPONSE"

if echo "$STATUS_RESPONSE" | grep -q '"enabled":true'; then
    print_info "2FA is already enabled. Testing login flow..."
    
    # Test login with 2FA
    echo ""
    echo "Testing 2FA login flow..."
    LOGIN_2FA=$(curl -s -X POST "${API_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${EMAIL}\",
            \"password\": \"${PASSWORD}\"
        }")
    
    if echo "$LOGIN_2FA" | grep -q "requires_2fa"; then
        print_success "2FA is properly enforced on login"
        TEMP_TOKEN=$(echo "$LOGIN_2FA" | grep -o '"temp_token":"[^"]*"' | cut -d'"' -f4)
        print_info "Temp token received: ${TEMP_TOKEN:0:50}..."
        
        echo ""
        print_info "Now you need to:"
        echo "1. Open your authenticator app"
        echo "2. Get the current TOTP code"
        echo "3. Run the following command:"
        echo ""
        echo "curl -X POST ${API_URL}/2fa/verify-login \\"
        echo "  -H \"Authorization: Bearer $TEMP_TOKEN\" \\"
        echo "  -H \"Content-Type: application/json\" \\"
        echo "  -d '{\"code\":\"YOUR_CODE_HERE\"}'"
    else
        print_error "Expected requires_2fa in response"
        echo "$LOGIN_2FA"
    fi
    
    exit 0
fi

echo ""

# Step 4: Setup 2FA
echo "Step 4: Setting up 2FA..."
print_info "This will return a QR code and recovery codes"

SETUP_RESPONSE=$(curl -s -X POST "${API_URL}/2fa/setup" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"password\": \"${PASSWORD}\"}")

if echo "$SETUP_RESPONSE" | grep -q "secret"; then
    print_success "2FA setup initiated successfully"
    
    # Extract secret and QR URL
    SECRET=$(echo "$SETUP_RESPONSE" | grep -o '"secret":"[^"]*"' | cut -d'"' -f4)
    QR_URL=$(echo "$SETUP_RESPONSE" | grep -o '"qr_code_url":"[^"]*"' | sed 's/"qr_code_url":"//g' | sed 's/"//g')
    
    print_info "Secret: $SECRET"
    echo ""
    print_info "Provisioning URL:"
    echo "$QR_URL"
    echo ""
    
    # Extract recovery codes
    print_info "Recovery Codes (save these!):"
    echo "$SETUP_RESPONSE" | grep -o '"[A-F0-9]\{8\}-[A-F0-9]\{8\}"' | sed 's/"//g' | nl
    
    echo ""
    print_info "Next steps:"
    echo "1. Add this account to your authenticator app using:"
    echo "   - Scan the QR code (if displayed), OR"
    echo "   - Manually enter the secret: $SECRET"
    echo ""
    echo "2. After adding, get the 6-digit code from your app"
    echo ""
    echo "3. Verify the setup by running:"
    echo "   curl -X POST ${API_URL}/2fa/verify-setup \\"
    echo "     -H \"Authorization: Bearer $TOKEN\" \\"
    echo "     -H \"Content-Type: application/json\" \\"
    echo "     -d '{\"code\":\"YOUR_CODE\"}'"
    echo ""
    echo "4. Test the complete login flow:"
    echo "   a. Login: curl -X POST ${API_URL}/login -d '{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}'"
    echo "   b. Get temp_token from response"
    echo "   c. Verify: curl -X POST ${API_URL}/2fa/verify-login -H \"Authorization: Bearer TEMP_TOKEN\" -d '{\"code\":\"YOUR_CODE\"}'"
    
else
    print_error "2FA setup failed"
    echo "$SETUP_RESPONSE"
    exit 1
fi

echo ""
echo "========================================"
print_success "2FA Test Suite Completed"
echo ""
print_info "Summary:"
echo "- User registered/logged in ✅"
echo "- 2FA setup initiated ✅"
echo "- QR code and recovery codes generated ✅"
echo ""
print_info "Manual steps required to complete testing:"
echo "1. Scan QR code with authenticator app"
echo "2. Verify setup with TOTP code"
echo "3. Test login flow with 2FA"
echo "4. (Optional) Test recovery codes"
