#!/bin/sh

# Configuration injection script for runtime environment variables
# This script creates a config file for the frontend based on container env vars.

echo "Injecting runtime configuration..."

# Default to an empty string if API_ENDPOINT is not set.
# An empty string means the frontend will use relative paths for API calls.
API_ENDPOINT="${API_ENDPOINT:-}"

# Create the config object to be loaded by the frontend.
cat > /usr/share/nginx/html/config.js << EOF
window.APP_CONFIG = {
  API_ENDPOINT: '${API_ENDPOINT}'
};
EOF

echo "Configuration injected:"
echo "  API_ENDPOINT='${API_ENDPOINT}'"

# Start nginx as the main process
exec nginx -g "daemon off;" 