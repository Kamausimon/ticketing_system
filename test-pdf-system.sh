#!/bin/bash

# Test Ticket PDF Generation System
# This script tests the complete PDF generation flow

set -e  # Exit on error

echo "🎫 Testing Ticket PDF Generation System"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test storage directory
echo "📁 Checking storage directory..."
if [ ! -d "storage/tickets" ]; then
    echo -e "${YELLOW}Creating storage/tickets directory...${NC}"
    mkdir -p storage/tickets
fi
echo -e "${GREEN}✓ Storage directory ready${NC}"
echo ""

# Test QR code package
echo "📷 Testing QR code generation..."
cd pkg/qrcode
if go test -v > /tmp/qr_test.log 2>&1; then
    echo -e "${GREEN}✓ QR code tests passed${NC}"
else
    echo -e "${RED}✗ QR code tests failed${NC}"
    cat /tmp/qr_test.log
    exit 1
fi
cd ../..
echo ""

# Test PDF package
echo "📄 Testing PDF generation..."
cd pkg/pdf
if go test -v > /tmp/pdf_test.log 2>&1; then
    echo -e "${GREEN}✓ PDF tests passed${NC}"
else
    echo -e "${RED}✗ PDF tests failed${NC}"
    cat /tmp/pdf_test.log
    exit 1
fi
cd ../..
echo ""

# Build the server
echo "🔨 Building API server..."
if go build -o bin/api-server ./cmd/api-server > /tmp/build.log 2>&1; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    cat /tmp/build.log
    exit 1
fi
echo ""

# Run the standalone example
echo "🎨 Testing standalone PDF example..."
cd examples/ticket_pdf
if go run main.go > /tmp/example.log 2>&1; then
    if [ -f "ticket_TKT-2024-VIP-001234.pdf" ]; then
        echo -e "${GREEN}✓ Example PDF generated successfully${NC}"
        ls -lh ticket_TKT-2024-VIP-001234.pdf
        
        # Verify it's actually a PDF
        if file ticket_TKT-2024-VIP-001234.pdf | grep -q "PDF"; then
            echo -e "${GREEN}✓ File is valid PDF${NC}"
        else
            echo -e "${RED}✗ File is not a valid PDF${NC}"
            exit 1
        fi
        
        # Clean up
        rm -f ticket_TKT-2024-VIP-001234.pdf
    else
        echo -e "${RED}✗ PDF file not created${NC}"
        cat /tmp/example.log
        exit 1
    fi
else
    echo -e "${RED}✗ Example failed${NC}"
    cat /tmp/example.log
    exit 1
fi
cd ../..
echo ""

# Check for required dependencies
echo "📦 Checking dependencies..."
if grep -q "github.com/jung-kurt/gofpdf" go.mod && \
   grep -q "github.com/skip2/go-qrcode" go.mod; then
    echo -e "${GREEN}✓ All dependencies present${NC}"
else
    echo -e "${RED}✗ Missing dependencies${NC}"
    exit 1
fi
echo ""

# Verify metrics are defined
echo "📊 Checking metrics..."
if grep -q "TicketDownloads" internal/analytics/metrics.go; then
    echo -e "${GREEN}✓ TicketDownloads metric defined${NC}"
else
    echo -e "${RED}✗ TicketDownloads metric missing${NC}"
    exit 1
fi
echo ""

# Summary
echo "========================================"
echo -e "${GREEN}✅ All tests passed!${NC}"
echo ""
echo "📋 Summary:"
echo "  ✓ Storage directory created"
echo "  ✓ QR code generation working"
echo "  ✓ PDF generation working"
echo "  ✓ API server builds successfully"
echo "  ✓ Standalone example works"
echo "  ✓ Dependencies installed"
echo "  ✓ Metrics configured"
echo ""
echo "🚀 System ready for production!"
echo ""
echo "Next steps:"
echo "  1. Start the API server: ./bin/api-server"
echo "  2. Create an order and generate tickets"
echo "  3. Download PDF: GET /api/tickets/{id}/pdf"
echo ""
