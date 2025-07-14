#!/bin/bash

echo "🔍 Debug: Container Environment Variables"
echo "=========================================="

echo "API-related environment variables:"
env | grep -E "(API_)" | sort

echo ""
echo "🔍 Debug: Configuration Injection Process"
echo "=========================================="

echo "Current working directory: $(pwd)"
echo "Contents of /var/www/html/:"
ls -la /var/www/html/ 2>/dev/null || echo "Directory /var/www/html/ does not exist"

echo ""
echo "Checking if config.js exists:"
if [ -f "/var/www/html/config.js" ]; then
    echo "✅ config.js exists"
    echo "Contents of config.js:"
    cat /var/www/html/config.js
else
    echo "❌ config.js does not exist"
fi

echo ""
echo "🔧 Manual Configuration Injection Test"
echo "======================================="

# Simulate the configuration injection manually
API_ENDPOINT="${API_ENDPOINT:-}"

echo "Environment variable values:"
echo "  API_ENDPOINT='${API_ENDPOINT}'"

echo ""
echo "Creating test config.js:"
cat > /tmp/test-config.js << EOF
window.APP_CONFIG = {
  API_ENDPOINT: '${API_ENDPOINT}'
};
EOF

echo "Generated test config.js:"
cat /tmp/test-config.js

echo ""
echo "🔍 Container Process Check"
echo "========================="
echo "Running processes:"
ps aux | grep -E "(nginx|v2fmdash)" || echo "No nginx or v2fmdash processes found"

echo ""
echo "📁 File System Check"
echo "==================="
echo "Contents of application directory:"
ls -la /app/ 2>/dev/null || echo "Directory /app/ does not exist"

echo "Nginx configuration:"
ls -la /etc/nginx/ 2>/dev/null || echo "Nginx config directory not found"

echo ""
echo "🎯 Recommendations"
echo "=================="
echo "1. Verify config.js is created in the correct location (/var/www/html/config.js)"
echo "2. Check that nginx is serving files from /var/www/html/"
echo "3. Verify the frontend can access /config.js URL"
echo "4. Check browser network tab for config.js load status" 
