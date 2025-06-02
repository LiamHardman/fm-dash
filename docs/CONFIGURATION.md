# Configuration Guide

This document explains how to configure the Football Manager Data Browser (FM-Dash) for different environments and use cases.

## Environment Variables

### Backend Configuration

#### Server Settings

```bash
# Server Configuration
SERVER_PORT=8091                    # Port for the API server (default: 8091)
SERVER_HOST=localhost               # Host to bind to (default: localhost)
SERVER_READ_TIMEOUT=30s             # HTTP read timeout
SERVER_WRITE_TIMEOUT=30s            # HTTP write timeout
SERVER_IDLE_TIMEOUT=120s            # HTTP idle timeout
```

#### CORS Settings

```bash
# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
CORS_MAX_AGE=3600                   # Preflight cache duration in seconds
```

#### File Upload Configuration

```bash
# File Upload Settings
MAX_UPLOAD_SIZE=50MB                # Maximum file size for uploads
UPLOAD_TIMEOUT=300s                 # Upload timeout duration
ALLOWED_FILE_TYPES=.html,.htm       # Allowed file extensions
TEMP_DIR=/tmp/uploads               # Temporary directory for uploads
```

#### Storage Configuration (S3/MinIO)

```bash
# S3/MinIO Configuration
S3_ENDPOINT=localhost:9000          # S3 endpoint URL
S3_ACCESS_KEY=minioadmin            # S3 access key
S3_SECRET_KEY=minioadmin            # S3 secret key
S3_BUCKET=fm-dash-data                 # S3 bucket name
S3_REGION=us-east-1                 # S3 region
S3_USE_SSL=false                    # Use SSL for S3 connection (true/false)
S3_FORCE_PATH_STYLE=true            # Force path-style URLs (for MinIO)
```

#### Processing Configuration

```bash
# Data Processing Settings
WORKER_COUNT=4                      # Number of worker goroutines (default: CPU count)
BATCH_SIZE=100                      # Players processed per batch
PROCESSING_TIMEOUT=600s             # Maximum processing time
MEMORY_LIMIT=512MB                  # Memory limit for processing
```

#### Cache Configuration

```bash
# Caching Settings
CACHE_ENABLED=true                  # Enable in-memory caching
CACHE_TTL=3600                      # Cache TTL in seconds (1 hour)
CACHE_MAX_SIZE=100MB                # Maximum cache size
```

#### Observability Settings

```bash
# OpenTelemetry Configuration
OTEL_ENABLED=true                   # Enable OpenTelemetry
OTEL_SERVICE_NAME=fm-dash-backend      # Service name for traces
OTEL_SERVICE_VERSION=1.0.0          # Service version
OTEL_JAEGER_ENDPOINT=http://localhost:14268/api/traces
OTEL_COLLECTOR_ENDPOINT=http://localhost:4317

# Metrics Configuration
METRICS_ENABLED=true                # Enable metrics collection
METRICS_PORT=9090                   # Prometheus metrics port
```

#### Security Settings

```bash
# Security Configuration
API_RATE_LIMIT=100                  # Requests per minute per IP
UPLOAD_RATE_LIMIT=5                 # Uploads per hour per IP
EXPORT_RATE_LIMIT=10                # Exports per hour per IP
SECURITY_HEADERS=true               # Enable security headers
```

### Frontend Configuration

#### Development Settings

```bash
# Development Configuration
VITE_API_URL=http://localhost:8091  # Backend API URL
VITE_APP_TITLE=FM-Dash                 # Application title
VITE_APP_VERSION=1.0.0              # Application version
VITE_DEBUG=true                     # Enable debug mode
```

#### Production Settings

```bash
# Production Configuration
VITE_API_URL=https://api.yourdomain.com
VITE_APP_TITLE=Football Manager Data Browser
VITE_APP_VERSION=1.0.0
VITE_DEBUG=false
VITE_SENTRY_DSN=https://your-sentry-dsn  # Error tracking
```

#### Feature Flags

```bash
# Feature Configuration
VITE_ENABLE_ANALYTICS=true          # Enable analytics features
VITE_ENABLE_EXPORT=true             # Enable data export
VITE_ENABLE_COMPARISON=true         # Enable player comparison
VITE_ENABLE_SHARING=false           # Enable data sharing features
VITE_MAX_PLAYERS_DISPLAY=1000       # Maximum players to display at once
```

## Configuration Files

### Backend Configuration (`config.yaml`)

```yaml
server:
  port: 8091
  host: "localhost"
  timeouts:
    read: "30s"
    write: "30s"
    idle: "120s"

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "https://yourdomain.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Requested-With"
  max_age: 3600

storage:
  type: "s3"  # s3, minio, or local
  s3:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "fm-dash-data"
    region: "us-east-1"
    use_ssl: false
    force_path_style: true

processing:
  worker_count: 4
  batch_size: 100
  timeout: "600s"
  memory_limit: "512MB"

cache:
  enabled: true
  ttl: "3600s"
  max_size: "100MB"

observability:
  otel:
    enabled: true
    service_name: "fm-dash-backend"
    service_version: "1.0.0"
    jaeger_endpoint: "http://localhost:14268/api/traces"
  metrics:
    enabled: true
    port: 9090

security:
  rate_limits:
    api: 100        # requests per minute
    upload: 5       # uploads per hour
    export: 10      # exports per hour
  security_headers: true

upload:
  max_size: "50MB"
  timeout: "300s"
  allowed_types:
    - ".html"
    - ".htm"
  temp_dir: "/tmp/uploads"
```

### Frontend Configuration (`quasar.config.js`)

```javascript
module.exports = function (ctx) {
  return {
    // Boot files
    boot: [
      'axios',
      'pinia'
    ],

    // CSS framework
    css: [
      'app.scss'
    ],

    // Quasar plugins
    framework: {
      config: {
        brand: {
          primary: '#1976D2',
          secondary: '#26A69A',
          accent: '#9C27B0',
          dark: '#1d1d1d',
          positive: '#21BA45',
          negative: '#C10015',
          info: '#31CCEC',
          warning: '#F2C037'
        }
      },
      plugins: [
        'Notify',
        'Dialog',
        'Loading',
        'LoadingBar'
      ]
    },

    // Build configuration
    build: {
      vueRouterMode: 'hash',
      publicPath: '/',
      
      env: {
        API_URL: process.env.VITE_API_URL || 'http://localhost:8091',
        APP_TITLE: process.env.VITE_APP_TITLE || 'FM-Dash',
        DEBUG: process.env.VITE_DEBUG === 'true',
        VERSION: process.env.VITE_APP_VERSION || '1.0.0'
      },

      // Vite configuration
      vitePlugins: [
        ['@intlify/vite-plugin-vue-i18n', {
          include: path.resolve(__dirname, './src/i18n/**')
        }]
      ]
    },

    // Development server
    devServer: {
      port: 3000,
      open: true,
      proxy: {
        '/api': {
          target: process.env.VITE_API_URL || 'http://localhost:8091',
          changeOrigin: true,
          secure: false
        }
      }
    }
  }
}
```

## Docker Configuration

### Environment File (`.env`)

```bash
# Application Configuration
APP_ENV=production
LOG_LEVEL=info

# Backend Configuration
SERVER_PORT=8091
WORKER_COUNT=4
PROCESSING_TIMEOUT=600s

# Storage Configuration
S3_ENDPOINT=minio:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=fm-dash-data
S3_USE_SSL=false

# Security
CORS_ALLOWED_ORIGINS=https://yourdomain.com
API_RATE_LIMIT=100

# Observability
OTEL_ENABLED=true
OTEL_JAEGER_ENDPOINT=http://jaeger:14268/api/traces
METRICS_ENABLED=true
```

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER_PORT=8091
      - S3_ENDPOINT=minio:9000
      - S3_ACCESS_KEY=${MINIO_ACCESS_KEY}
      - S3_SECRET_KEY=${MINIO_SECRET_KEY}
      - S3_BUCKET=fm-dash-data
      - S3_USE_SSL=false
    depends_on:
      - minio
    networks:
      - fm-dash-network

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
      - MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    networks:
      - fm-dash-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - app
    networks:
      - fm-dash-network

volumes:
  minio_data:

networks:
  fm-dash-network:
    driver: bridge
```

## Kubernetes Configuration

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fm-dash-config
data:
  config.yaml: |
    server:
      port: 8091
      host: "0.0.0.0"
    storage:
      type: "s3"
      s3:
        bucket: "fm-dash-data"
        region: "us-east-1"
        use_ssl: true
    processing:
      worker_count: 4
      batch_size: 100
    cache:
      enabled: true
      ttl: "3600s"
    observability:
      otel:
        enabled: true
        service_name: "fm-dash-backend"
      metrics:
        enabled: true
        port: 9090
```

### Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: fm-dash-secrets
type: Opaque
stringData:
  s3-access-key: "your-access-key"
  s3-secret-key: "your-secret-key"
  s3-endpoint: "your-s3-endpoint"
```

### Deployment with Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fm-dash-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fm-dash
  template:
    metadata:
      labels:
        app: fm-dash
    spec:
      containers:
      - name: fm-dash
        image: fm-dash:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090  # Metrics port
        
        env:
        - name: CONFIG_FILE
          value: "/etc/config/config.yaml"
        - name: S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: fm-dash-secrets
              key: s3-access-key
        - name: S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: fm-dash-secrets
              key: s3-secret-key
        
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

      volumes:
      - name: config-volume
        configMap:
          name: fm-dash-config
```

## Environment-Specific Configurations

### Development Environment

```bash
# Development Settings
APP_ENV=development
LOG_LEVEL=debug
SERVER_HOST=localhost
SERVER_PORT=8091

# Frontend Development
VITE_API_URL=http://localhost:8091
VITE_DEBUG=true
VITE_HOT_RELOAD=true

# Storage (Local MinIO)
S3_ENDPOINT=localhost:9000
S3_USE_SSL=false
S3_FORCE_PATH_STYLE=true

# Disable features for development
CORS_ALLOWED_ORIGINS=http://localhost:3000
OTEL_ENABLED=false
RATE_LIMITING_ENABLED=false
```

### Staging Environment

```bash
# Staging Settings
APP_ENV=staging
LOG_LEVEL=info
SERVER_HOST=0.0.0.0
SERVER_PORT=8091

# Frontend Staging
VITE_API_URL=https://api-staging.yourdomain.com
VITE_DEBUG=false
VITE_SENTRY_DSN=https://your-staging-sentry-dsn

# Storage (AWS S3)
S3_ENDPOINT=s3.amazonaws.com
S3_USE_SSL=true
S3_FORCE_PATH_STYLE=false

# Enable observability
OTEL_ENABLED=true
METRICS_ENABLED=true
RATE_LIMITING_ENABLED=true
```

### Production Environment

```bash
# Production Settings
APP_ENV=production
LOG_LEVEL=warn
SERVER_HOST=0.0.0.0
SERVER_PORT=8091

# Frontend Production
VITE_API_URL=https://api.yourdomain.com
VITE_DEBUG=false
VITE_SENTRY_DSN=https://your-production-sentry-dsn

# Storage (AWS S3 with encryption)
S3_ENDPOINT=s3.amazonaws.com
S3_USE_SSL=true
S3_FORCE_PATH_STYLE=false
S3_SERVER_SIDE_ENCRYPTION=true

# Security and monitoring
SECURITY_HEADERS=true
RATE_LIMITING_ENABLED=true
OTEL_ENABLED=true
METRICS_ENABLED=true

# Performance tuning
WORKER_COUNT=8
BATCH_SIZE=200
CACHE_ENABLED=true
```

## Performance Tuning

### Backend Performance Configuration

```bash
# Goroutine Configuration
WORKER_COUNT=8                      # Number of processing workers
MAX_GOROUTINES=1000                 # Maximum concurrent goroutines
GOROUTINE_POOL_SIZE=100             # Pool size for worker goroutines

# Memory Configuration
MEMORY_LIMIT=1GB                    # Memory limit for the application
GC_TARGET_PERCENTAGE=100            # Go garbage collection target
MAX_HEAP_SIZE=800MB                 # Maximum heap size

# HTTP Configuration
MAX_CONCURRENT_CONNECTIONS=1000     # Maximum concurrent HTTP connections
KEEP_ALIVE_TIMEOUT=30s              # Keep-alive timeout
READ_HEADER_TIMEOUT=10s             # Read header timeout

# Cache Configuration
CACHE_ENABLED=true
CACHE_TTL=3600                      # Cache TTL in seconds
CACHE_MAX_SIZE=200MB                # Maximum cache size
CACHE_CLEANUP_INTERVAL=300s         # Cache cleanup interval
```

### Frontend Performance Configuration

```javascript
// Quasar performance configuration
module.exports = {
  build: {
    // Code splitting
    analyze: false,
    minify: true,
    
    // Chunk optimization
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          quasar: ['quasar'],
          charts: ['chart.js', 'vue-chartjs']
        }
      }
    },

    // Compression
    gzip: true,
    brotli: true,

    // Source maps (disable in production)
    sourcemap: process.env.NODE_ENV !== 'production'
  },

  // PWA configuration
  pwa: {
    workboxPluginMode: 'InjectManifest',
    workboxOptions: {
      maximumFileSizeToCacheInBytes: 5000000
    }
  }
}
```

## Monitoring Configuration

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'fm-dash-backend'
    static_configs:
      - targets: ['fm-dash-app:9090']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'fm-dash-nginx'
    static_configs:
      - targets: ['nginx-exporter:9113']
```

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "FM-Dash Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{ method }} {{ path }}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph", 
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration_seconds_bucket)",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## Troubleshooting Configuration Issues

### Common Configuration Problems

1. **CORS Issues**
   ```bash
   # Check CORS configuration
   CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
   ```

2. **S3 Connection Issues**
   ```bash
   # Verify S3 settings
   S3_ENDPOINT=your-endpoint
   S3_USE_SSL=true
   S3_FORCE_PATH_STYLE=false  # Set to true for MinIO
   ```

3. **Memory Issues**
   ```bash
   # Adjust memory limits
   MEMORY_LIMIT=1GB
   WORKER_COUNT=4  # Reduce if memory constrained
   ```

### Configuration Validation

```bash
# Validate configuration
npm run config:validate

# Check environment variables
npm run env:check

# Test S3 connectivity
npm run test:s3

# Verify all services
npm run health:check
```

---

For additional configuration options and advanced setups, refer to the specific component documentation or contact the development team. 