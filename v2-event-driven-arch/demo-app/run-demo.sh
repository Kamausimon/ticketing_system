#!/bin/bash

# Demo App Launcher for Event Ticketing System
# This script helps you run the demo frontend application

set -e

echo "🎫 Event Ticketing System - Demo App Launcher"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if backend is running
echo "Checking if backend is running..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Backend is running on http://localhost:8080${NC}"
else
    echo -e "${YELLOW}⚠ Backend doesn't seem to be running on http://localhost:8080${NC}"
    echo "Please start your backend first with: ./deploy-app.sh"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""
echo "Choose how to run the demo app:"
echo "1) Python HTTP Server (recommended)"
echo "2) Node.js (npx serve)"
echo "3) PHP built-in server"
echo "4) Just open in browser (may have limitations)"
echo ""
read -p "Enter your choice (1-4): " choice

cd demo-app

case $choice in
    1)
        echo -e "${GREEN}Starting Python HTTP Server...${NC}"
        echo "Demo app will be available at: http://localhost:3000"
        echo "Press Ctrl+C to stop"
        echo ""
        python3 -m http.server 3000
        ;;
    2)
        echo -e "${GREEN}Starting Node.js server...${NC}"
        if ! command -v npx &> /dev/null; then
            echo -e "${RED}npx (Node.js) not found. Please install Node.js first.${NC}"
            exit 1
        fi
        echo "Demo app will be available at: http://localhost:3000"
        echo "Press Ctrl+C to stop"
        echo ""
        npx serve -p 3000 .
        ;;
    3)
        echo -e "${GREEN}Starting PHP server...${NC}"
        if ! command -v php &> /dev/null; then
            echo -e "${RED}PHP not found. Please install PHP first.${NC}"
            exit 1
        fi
        echo "Demo app will be available at: http://localhost:3000"
        echo "Press Ctrl+C to stop"
        echo ""
        php -S localhost:3000
        ;;
    4)
        echo -e "${GREEN}Opening in browser...${NC}"
        echo -e "${YELLOW}Note: Some features may not work with file:// protocol${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            open index.html
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            xdg-open index.html
        else
            echo "Please open demo-app/index.html in your browser manually"
        fi
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac
