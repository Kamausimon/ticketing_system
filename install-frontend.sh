#!/bin/bash

echo "🧹 Cleaning up old files..."
cd /home/kamau/projects/ticketing_system/frontend

# Force remove any remaining files
sudo rm -rf node_modules package-lock.json 2>/dev/null || true

echo "📦 Installing dependencies (this may take a few minutes)..."
npm install

if [ $? -eq 0 ]; then
    echo "✅ Installation successful!"
    echo ""
    echo "🚀 To start the development server, run:"
    echo "   cd /home/kamau/projects/ticketing_system/frontend"
    echo "   npm run dev"
else
    echo "❌ Installation failed. Please manually delete node_modules folder from File Explorer and try again."
fi
