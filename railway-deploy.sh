#!/bin/bash

echo "🚀 Railway Deployment Quick Start"
echo "=================================="
echo ""

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found"
    echo "📦 Install it with: npm install -g @railway/cli"
    echo "Or deploy via web: https://railway.app"
    exit 1
fi

echo "✅ Railway CLI found"
echo ""

# Login check
echo "🔐 Checking Railway authentication..."
if ! railway whoami &> /dev/null; then
    echo "Please login to Railway:"
    railway login
fi

echo ""
echo "Choose deployment option:"
echo "1. Deploy Backend only"
echo "2. Deploy Demo only"
echo "3. Deploy both (separate projects)"
echo ""
read -p "Enter choice (1-3): " choice

case $choice in
    1)
        echo ""
        echo "📦 Deploying Backend..."
        railway init
        railway up
        echo ""
        echo "🎉 Backend deployed!"
        echo "Don't forget to:"
        echo "  - Add PostgreSQL plugin"
        echo "  - Add Redis plugin"
        echo "  - Configure environment variables"
        echo "  - Generate domain"
        ;;
    2)
        echo ""
        echo "📦 Deploying Demo App..."
        cd demo-app || exit
        railway init
        railway up
        echo ""
        echo "🎉 Demo deployed!"
        echo "Don't forget to:"
        echo "  - Set API_BASE_URL environment variable"
        echo "  - Generate domain"
        cd ..
        ;;
    3)
        echo ""
        echo "📦 Deploying Backend first..."
        railway init
        railway up
        backend_url=$(railway domain)
        
        echo ""
        echo "✅ Backend deployed at: $backend_url"
        echo ""
        echo "📦 Now deploying Demo in a new project..."
        echo "Press Enter to continue..."
        read
        
        cd demo-app || exit
        railway init
        railway variables set API_BASE_URL="https://$backend_url"
        railway up
        
        echo ""
        echo "🎉 Both deployed!"
        echo "Backend: https://$backend_url"
        echo "Demo: Check Railway dashboard for URL"
        cd ..
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "✅ Deployment complete!"
echo "Visit Railway dashboard: https://railway.app/dashboard"
