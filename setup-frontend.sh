#!/bin/bash

# Frontend Setup Script
# Run this from WSL terminal: bash setup-frontend.sh

echo "🎨 Setting up Ticketing System Frontend..."

cd frontend

echo "📦 Installing dependencies..."
npm install

echo "✅ Frontend setup complete!"
echo ""
echo "To start the development server:"
echo "  cd frontend"
echo "  npm run dev"
echo ""
echo "The frontend will be available at http://localhost:3000"
