#!/bin/bash

echo "🔄 Refreshing OAuth Token for Expense Tracker"
echo "=============================================="

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Creating one..."
    echo "GOOGLE_CLOUD_PROJECT=your-project-id" > .env
    echo "✅ Created .env file with your project ID"
fi

# Remove expired token
if [ -f credentials/token.json ]; then
    echo "🗑️  Removing expired token..."
    rm credentials/token.json
    echo "✅ Expired token removed"
else
    echo "ℹ️  No existing token found"
fi

echo ""
echo "🚀 Starting OAuth flow in Docker..."
echo "📝 Steps to follow:"
echo "   1. Container will start and show an OAuth URL"
echo "   2. Copy the URL and open it in your browser"
echo "   3. Complete the Google OAuth flow"
echo "   4. The app will automatically save the new token"
echo ""
echo "Press Enter to continue..."
read

# Start the container for OAuth
docker-compose up --build

echo ""
echo "🎉 Token refresh complete!"
echo "💡 You can now run 'docker-compose up' for regular operations" 