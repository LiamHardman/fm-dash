# Dockerfile
# Location: Root of your project
# This Dockerfile builds both the Vue.js frontend and the Go backend,
# and sets up Nginx to serve the frontend and proxy API requests to the Go backend.
# Supervisord is used to manage both processes.

# ---- Stage 1: Build Vue.js Frontend ----
FROM node:18-alpine AS vue-builder
LABEL stage=vue-builder
WORKDIR /app-vue

# Copy package.json and package-lock.json (or yarn.lock)
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application code (Vue frontend source)
COPY . .

# Build the application
# VITE_API_BASE_URL should be set to the path Nginx will proxy to the API.
# Example: "/api" if Nginx proxies /api/* to the Go backend.
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN echo "Building Vue app with VITE_API_BASE_URL=${VITE_API_BASE_URL}"
RUN npm run build
# Frontend is built into /app-vue/dist

# ---- Stage 2: Build Go Backend ----
FROM golang:1.23-bullseye AS go-builder
LABEL stage=go-builder
WORKDIR /app-go

# Install ca-certificates (might be needed if your Go app makes HTTPS calls)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy Go module files from src/api directory
COPY src/api/go.mod src/api/go.sum ./
# Download dependencies
RUN go mod download
RUN go mod verify

# Copy the entire Go API source code (from src/api)
COPY src/api/ ./

# Build the Go application
# Output the binary to /app-go/v2fmdash-server (updated name)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app-go/v2fmdash-server .
# Go binary is at /app-go/v2fmdash-server
# Runtime JSON files are in /app-go/public/

# ---- Stage 3: Final Production Image ----
FROM debian:bullseye-slim
LABEL stage=final-image
WORKDIR /app

# Install Nginx, supervisord, and ca-certificates
RUN apt-get update && apt-get install -y \
    nginx \
    supervisor \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# --- Nginx Setup ---
# Copy custom Nginx configuration
# This nginx.conf should be in the root of your project, next to this Dockerfile.
COPY nginx.conf /etc/nginx/nginx.conf

# Remove default Nginx site if it exists to avoid conflicts
RUN rm -f /etc/nginx/sites-enabled/default

# Copy built Vue.js frontend static assets from the vue-builder stage to Nginx webroot
COPY --from=vue-builder /app-vue/dist /var/www/html/

# --- Go Application Setup ---
# Copy the Go binary from the go-builder stage
COPY --from=go-builder /app-go/v2fmdash-server .

# Copy the Go application's public assets (e.g., JSON config files)
# These are expected by the Go app to be in a 'public' directory relative to its execution.
COPY --from=go-builder /app-go/public ./public/

# --- Supervisord Setup ---
# Copy supervisord configuration file
# This supervisord.conf should be in the root of your project.
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create log directory for supervisord (and potentially for Nginx/Go app if configured)
RUN mkdir -p /var/log/supervisor

# Nginx typically listens on 80. Go API will listen on 8091 (internally).
# The main exposed port will be Nginx's port.
EXPOSE 8080 # Nginx will be configured to listen on this port (matches PORT_NGINX)

# Set default environment variables for Go application (can be overridden)
ENV PORT_GO_API=8091
ENV PORT_NGINX=8080
# ENV GIN_MODE=release # If using Gin framework for Go

# Command to run supervisord, which will manage Nginx and the Go app
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
