#!/bin/sh
# Script to inject Railway environment variables into the demo app

# Create a config.js file with the API URL from environment variable
cat > /usr/share/nginx/html/config.js << EOF
window.RAILWAY_API_URL = '${API_BASE_URL:-http://localhost:8080}';
EOF

echo "API URL configured: ${API_BASE_URL:-http://localhost:8080}"

# Start nginx
nginx -g 'daemon off;'
