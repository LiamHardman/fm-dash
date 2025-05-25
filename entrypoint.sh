#!/bin/sh
# entrypoint.sh

# Start Nginx in the background.
# The '-g "daemon off;"' directive makes Nginx run in the foreground relative to its own process management,
# but we'll background it here with '&' so the script can then start the Go app.
# Systemd would handle this differently, but for a simple container CMD, this is common.
echo "Starting Nginx..."
nginx -g 'daemon on; master_process on;'

# Start the Go API server in the foreground.
# This will be the main process that keeps the container running.
echo "Starting Go API server on port ${PORT_GO_API}..."
# PORT_GO_API is an environment variable set in the Dockerfile.
# The Go application should listen on the port specified by this ENV var.
/app/v2fmdash-server
