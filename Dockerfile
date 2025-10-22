FROM node:24-alpine AS vue-builder
LABEL stage=vue-builder
WORKDIR /app-vue

# Copy package files for better caching and consistent dependency versions
COPY package.json package-lock.json ./

# Use npm ci with BuildKit cache mount for faster, reproducible builds
RUN --mount=type=cache,target=/root/.npm \
    npm ci --legacy-peer-deps

# Copy only necessary source files (dockerignore filters out the rest)
COPY src ./src
COPY public ./public
COPY index.html vite.config.js ./

ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN echo "Building Vue app with VITE_API_BASE_URL=${VITE_API_BASE_URL}"

# Build the Vue application
RUN npm run build

FROM golang:1.24-alpine AS go-builder
LABEL stage=go-builder
WORKDIR /app-go
RUN apk add --no-cache ca-certificates git

COPY src/api/go.mod src/api/go.sum ./

# Use BuildKit cache mount for Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    go mod verify

COPY src/api/ ./

# Build with cache mounts for faster compilation
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app-go/v2fmdash-server .

# Use Alpine for the final image instead of Debian
FROM alpine:3.20
LABEL stage=final-image
WORKDIR /app

# Install only essential packages, use nginx from Alpine repos
RUN apk add --no-cache \
    nginx \
    ca-certificates \
    && rm -rf /var/cache/apk/*

COPY nginx.conf /etc/nginx/nginx.conf

RUN rm -f /etc/nginx/sites-enabled/default

COPY --from=vue-builder /app-vue/dist /var/www/html/

COPY --from=go-builder /app-go/v2fmdash-server .

COPY --from=go-builder /app-go/public ./public/

# Copy the utils directory containing teams_data.json
COPY --from=go-builder /app-go/utils ./utils/

# Create application user with Alpine-compatible commands
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /app -s /sbin/nologin
RUN chown -R appuser:appgroup /app/public
RUN chown appuser:appgroup /app/v2fmdash-server
RUN chown -R nginx:nginx /var/www/html
RUN chown -R nginx:nginx /var/log/nginx
RUN chown -R nginx:nginx /var/lib/nginx
RUN touch /run/nginx.pid && chown nginx:nginx /run/nginx.pid

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

ENV PORT_GO_API=8091
ENV PORT_NGINX=8080

CMD ["/usr/local/bin/entrypoint.sh"]
