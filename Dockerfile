FROM node:18-alpine AS vue-builder
LABEL stage=vue-builder
WORKDIR /app-vue

COPY package*.json ./

RUN npm install

COPY . .

ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN echo "Building Vue app with VITE_API_BASE_URL=${VITE_API_BASE_URL}"
RUN npm run build

FROM golang:1.23-bullseye AS go-builder
LABEL stage=go-builder
WORKDIR /app-go

# It's good practice to install ca-certificates if you're making HTTPS requests during the build
# or if your Go application needs them.
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY src/api/go.mod src/api/go.sum ./
RUN go mod download
RUN go mod verify

COPY src/api/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app-go/v2fmdash-server .

FROM debian:bullseye-slim
LABEL stage=final-image
WORKDIR /app

# Corrected RUN command: removed non-breaking spaces
# Ensure packages are installed one after another or on separate lines if issues persist,
# but this combined command should work with standard spaces.
RUN apt-get update && apt-get install -y \
    nginx \
    supervisor \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY nginx.conf /etc/nginx/nginx.conf

# Remove the default Nginx site configuration
RUN rm -f /etc/nginx/sites-enabled/default

# Copy Vue.js build from the vue-builder stage
COPY --from=vue-builder /app-vue/dist /var/www/html/

# Copy Go binary from the go-builder stage
COPY --from=go-builder /app-go/v2fmdash-server .

# Copy static assets for Go (like JSON files) from the go-builder stage
COPY --from=go-builder /app-go/public ./public/

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create log directory for Supervisor
RUN mkdir -p /var/log/supervisor

EXPOSE 8080

# Environment variables for ports, used by supervisord.conf
ENV PORT_GO_API=8091
ENV PORT_NGINX=8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
