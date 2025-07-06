#!/bin/sh
# entrypoint.sh

# Inject runtime configuration for frontend
echo "Injecting runtime configuration..."

# Default to an empty string if API_ENDPOINT is not set.
API_ENDPOINT="${API_ENDPOINT:-}"

# Create the config object to be loaded by the frontend.
cat > /var/www/html/config.js << EOF
window.APP_CONFIG = {
  API_ENDPOINT: '${API_ENDPOINT}'
};
EOF

echo "Configuration injected:"
echo "  API_ENDPOINT='${API_ENDPOINT}'"

# Start Nginx in the background.
echo "Starting Nginx..."
nginx -g 'daemon on; master_process on;'

# Start the Go API server in the foreground.
# This will be the main process that keeps the container running.
echo "Starting Go API server on port ${PORT_GO_API}..."
# PORT_GO_API is an environment variable set in the Dockerfile.
/app/v2fmdash-server
