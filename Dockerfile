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

FROM golang:1.24-bullseye AS go-builder
LABEL stage=go-builder
WORKDIR /app-go
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY src/api/go.mod src/api/go.sum ./
RUN go mod download
RUN go mod verify
COPY src/api/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app-go/v2fmdash-server .

FROM debian:bullseye-slim
LABEL stage=final-image
WORKDIR /app

RUN apt-get update && apt-get install -y \
    nginx \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY nginx.conf /etc/nginx/nginx.conf

RUN rm -f /etc/nginx/sites-enabled/default

COPY --from=vue-builder /app-vue/dist /var/www/html/

COPY --from=go-builder /app-go/v2fmdash-server .

COPY --from=go-builder /app-go/public ./public/

# Copy the utils directory containing teams_data.json
COPY --from=go-builder /app-go/utils ./utils/

RUN groupadd -r appgroup && useradd -r -g appgroup -d /app -s /sbin/nologin -c "Application User" appuser
RUN chown -R appuser:appgroup /app/public
RUN chown appuser:appgroup /app/v2fmdash-server
RUN chown -R www-data:www-data /var/www/html
RUN chown -R www-data:www-data /var/log/nginx
RUN chown -R www-data:www-data /var/lib/nginx
RUN touch /run/nginx.pid && chown www-data:www-data /run/nginx.pid

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

ENV PORT_GO_API=8091
ENV PORT_NGINX=8080

CMD ["/usr/local/bin/entrypoint.sh"]
