#!/bin/bash
# Railway Staging Environment Setup Script

echo "🚀 Railway Staging Environment Setup"
echo "======================================"
echo ""
echo "This script helps you set up a staging environment on Railway."
echo ""

# Check if railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing..."
    npm install -g @railway/cli
    echo "✅ Railway CLI installed"
fi

echo ""
echo "📋 Setup Steps:"
echo ""
echo "1. Login to Railway:"
railway login

echo ""
echo "2. Link to your project:"
railway link

echo ""
echo "3. Create staging environment:"
echo "   Running: railway environment create staging"
read -p "Press Enter to continue..."
railway environment --create staging

echo ""
echo "4. Switch to staging environment:"
railway environment staging

echo ""
echo "5. Set staging environment variables:"
echo ""
read -p "Enter your staging database URL (or press Enter to skip): " staging_db
if [ ! -z "$staging_db" ]; then
    railway variables set DATABASE_URL="$staging_db"
fi

read -p "Enter your staging frontend URL (or press Enter to skip): " staging_frontend
if [ ! -z "$staging_frontend" ]; then
    railway variables set FRONTEND_URL="$staging_frontend"
fi

railway variables set ENVIRONMENT="staging"

echo ""
echo "6. Deploy to staging:"
read -p "Deploy now? (y/n): " deploy_now
if [ "$deploy_now" = "y" ]; then
    railway up
fi

echo ""
echo "✅ Staging environment setup complete!"
echo ""
echo "📚 Next steps:"
echo "  - Switch environments: railway environment [production|staging]"
echo "  - View logs: railway logs --follow"
echo "  - Deploy: railway up"
echo ""
echo "🔗 Useful Commands:"
echo "  railway environment              # List environments"
echo "  railway environment staging      # Switch to staging"
echo "  railway environment production   # Switch to production"
echo "  railway logs --follow            # Watch logs"
echo "  railway status                   # Check deployment status"
echo ""
echo "📖 Full guide: PRODUCTION_DEPLOYMENT_GUIDE.md"
