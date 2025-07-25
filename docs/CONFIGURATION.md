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

# Dataset Retention Configuration
DATASET_RETENTION_DAYS=30           # Number of days to retain datasets before automatic cleanup (default: 30)

# Image Storage Configuration
IMAGE_API_URL=https://sortitoutsi.b-cdn.net/uploads  # External image API URL for faces/logos
FACES_DIR=./faces                   # Local faces directory (fallback)
LOGOS_DIR=./logos                   # Local logos directory (fallback)
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
VITE_HOT_RELOAD=true                # Enable hot module replacement
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

#### Performance Configuration

```bash
# Bundle Optimization (In Progress)
VITE_BUNDLE_ANALYZER=false          # Enable bundle analysis
VITE_CODE_SPLITTING=true            # Enable code splitting
VITE_TREE_SHAKING=true              # Enable tree shaking
VITE_MINIFY=true                    # Enable minification
VITE_CHUNK_SIZE_WARNING=500         # Chunk size warning threshold (KB) - reduced for better optimization
VITE_THIRD_PARTY_OPTIMIZATION=true  # Enable third-party library optimization (in progress)

# Memory Management
VITE_VIRTUAL_SCROLL_BUFFER=5        # Virtual scroll buffer size
VITE_OBJECT_POOL_SIZE=50            # Object pool initial size
VITE_LRU_CACHE_SIZE=100             # LRU cache maximum size
VITE_MEMORY_THRESHOLD=200           # Memory threshold in MB

# Image Optimization
VITE_LAZY_LOADING=true              # Enable lazy loading
VITE_WEBP_SUPPORT=true              # Enable WebP format support
VITE_AVIF_SUPPORT=true              # Enable AVIF format support
VITE_IMAGE_PRELOAD_COUNT=5          # Number of images to preload
VITE_IMAGE_CACHE_SIZE=50            # Image cache size (MB)

# Mobile Optimization
VITE_MOBILE_OPTIMIZATIONS=true      # Enable mobile optimizations
VITE_TOUCH_OPTIMIZATION=true        # Optimize touch interactions
VITE_REDUCED_MOTION_SUPPORT=true    # Support reduced motion preferences
VITE_BATTERY_OPTIMIZATION=true      # Enable battery-aware optimizations

# Performance Monitoring
VITE_PERFORMANCE_TRACKING=true      # Enable performance tracking
VITE_CORE_WEB_VITALS=true          # Track Core Web Vitals
VITE_MEMORY_MONITORING=true         # Monitor memory usage
VITE_BUNDLE_MONITORING=true         # Monitor bundle loading
```

#### Feature Flags

```bash
# Feature Configuration
VITE_ENABLE_ANALYTICS=true          # Enable analytics features
VITE_ENABLE_EXPORT=true             # Enable data export
VITE_ENABLE_COMPARISON=true         # Enable player comparison
VITE_ENABLE_SHARING=false           # Enable data sharing features
VITE_MAX_PLAYERS_DISPLAY=1000       # Maximum players to display at once

# Performance Features
VITE_ENABLE_WEB_WORKERS=true        # Enable web workers for calculations
VITE_ENABLE_SERVICE_WORKER=true     # Enable service worker for caching
VITE_ENABLE_VIRTUAL_SCROLLING=true  # Enable virtual scrolling
VITE_ENABLE_PROGRESSIVE_LOADING=true # Enable progressive image loading
```

#### Advanced Performance Settings

```bash
# Network Optimization
VITE_REQUEST_TIMEOUT=30000          # API request timeout (ms)
VITE_RETRY_ATTEMPTS=3               # Number of retry attempts
VITE_RETRY_DELAY=1000               # Retry delay (ms)
VITE_CONCURRENT_REQUESTS=6          # Maximum concurrent requests

# Animation Configuration
VITE_ANIMATION_DURATION=300         # Default animation duration (ms)
VITE_ANIMATION_EASING=ease-out      # Default animation easing
VITE_DISABLE_ANIMATIONS_LOW_BATTERY=true # Disable animations on low battery

# Development Performance
VITE_DEV_OVERLAY=true               # Show performance overlay in development
VITE_DEV_MEMORY_WARNINGS=true       # Show memory warnings in development
VITE_DEV_BUNDLE_WARNINGS=true       # Show bundle size warnings
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
  images:
    api_url: "https://sortitoutsi.b-cdn.net/uploads"  # External image API URL
    faces_dir: "./faces"                              # Local faces directory
    logos_dir: "./logos"                              # Local logos directory

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

### Frontend Configuration (`vite.config.js`)

```javascript
// Enhanced Vite configuration with performance optimizations
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },

  plugins: [
    vue({
      template: { transformAssetUrls }
    }),
    quasar({
      sassVariables: '@/quasar-variables.scss'
    }),
    // Bundle analyzer for performance monitoring
    (process.env.ANALYZE === 'true' || process.env.NODE_ENV === 'development') &&
      visualizer({
        filename: 'dist/stats.html',
        open: process.env.ANALYZE === 'true',
        gzipSize: true,
        brotliSize: true,
        template: 'treemap'
      })
  ].filter(Boolean),

  build: {
    // Advanced code splitting and optimization
    rollupOptions: {
      output: {
        manualChunks: {
          // Core framework chunks
          'vue-core': ['vue', 'vue-router', 'pinia'],
          'ui-framework': ['quasar'],
          
          // Feature-based chunks
          'charts': ['chart.js', 'vue-chartjs', 'chartjs-plugin-annotation'],
          'utils': ['@vueuse/core'],
          
          // Page-based chunks for better caching
          'pages-main': [
            './src/pages/LandingPage.vue',
            './src/pages/DatasetPage.vue'
          ],
          'pages-secondary': [
            './src/pages/PerformancePage.vue',
            './src/pages/TeamViewPage.vue'
          ]
        },
        
        // Optimized file naming for better caching
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId
            ? chunkInfo.facadeModuleId.split('/').pop().replace('.vue', '')
            : 'chunk'
          return `js/${facadeModuleId}-[hash].js`
        }
      }
    },

    // Performance optimization settings
    chunkSizeWarningLimit: parseInt(process.env.VITE_CHUNK_SIZE_WARNING) || 800,
    sourcemap: process.env.NODE_ENV === 'development',
    cssCodeSplit: true,
    assetsInlineLimit: 2048,
    
    // Advanced minification
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: process.env.NODE_ENV === 'production',
        drop_debugger: true,
        passes: 2
      }
    },
    
    // Target modern browsers for smaller bundles
    target: ['es2020', 'chrome80', 'firefox78', 'safari14']
  },

  // Enhanced dependency optimization
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'pinia',
      'quasar',
      '@vueuse/core',
      // Pre-bundle commonly used utilities with specific imports
      'chart.js/helpers',
      'chart.js/auto',
      'vue-chartjs',
      'chartjs-plugin-annotation'
    ],
    exclude: [
      '@vitejs/plugin-vue',
      // Exclude development-only dependencies
      'rollup-plugin-visualizer',
      '@vue/devtools-api'
    ],
    esbuildOptions: {
      target: 'es2020',
      format: 'esm',
      treeShaking: true,
      // Optimize for production builds
      minify: process.env.NODE_ENV === 'production',
      // Define globals for better optimization
      define: {
        'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development')
      }
    },
    // Force optimization of specific packages
    force: process.env.NODE_ENV === 'production'
  },

  // Development server with performance monitoring
  server: {
    port: 3000,
    proxy: {
      '/upload': {
        target: 'http://localhost:8091',
        changeOrigin: true
      },
      '/api': {
        target: 'http://localhost:8091',
        changeOrigin: true
      }
    }
  },

  // Performance monitoring definitions
  define: {
    __PERFORMANCE_TRACKING__: JSON.stringify(
      process.env.VITE_PERFORMANCE_TRACKING === 'true'
    ),
    __CORE_WEB_VITALS__: JSON.stringify(
      process.env.VITE_CORE_WEB_VITALS === 'true'
    ),
    __MEMORY_MONITORING__: JSON.stringify(
      process.env.VITE_MEMORY_MONITORING === 'true'
    )
  }
})
```

### Performance-Optimized Quasar Configuration

```javascript
// quasar.config.js with performance enhancements
module.exports = function (ctx) {
  return {
    // Optimized boot files loading
    boot: [
      { path: 'axios', server: false },
      { path: 'pinia', server: false },
      { path: 'performance', server: false } // Performance monitoring
    ],

    // CSS framework with tree-shaking
    css: [
      'app.scss'
    ],

    // Quasar plugins (only load what's needed)
    framework: {
      config: {
        brand: {
          primary: '#1976D2',
          secondary: '#26A69A'
        }
      },
      plugins: [
        'Notify',
        'Dialog',
        'Loading',
        'LoadingBar'
      ],
      // Tree-shake Quasar components
      components: [
        'QLayout',
        'QHeader',
        'QDrawer',
        'QPageContainer',
        'QPage',
        'QBtn',
        'QTable',
        'QVirtualScroll'
      ],
      directives: [
        'Ripple',
        'ClosePopup'
      ]
    },

    // Build configuration with performance optimizations
    build: {
      vueRouterMode: 'hash',
      publicPath: '/',
      
      // Environment variables
      env: {
        API_URL: process.env.VITE_API_URL || 'http://localhost:8091',
        PERFORMANCE_TRACKING: process.env.VITE_PERFORMANCE_TRACKING || 'false',
        MEMORY_MONITORING: process.env.VITE_MEMORY_MONITORING || 'false'
      },

      // Advanced build optimizations
      minify: true,
      sourcemap: ctx.dev,
      
      // Webpack optimizations (if using webpack mode)
      chainWebpack(chain) {
        // Optimize chunks
        chain.optimization.splitChunks({
          chunks: 'all',
          cacheGroups: {
            vendor: {
              test: /[\\/]node_modules[\\/]/,
              name: 'vendors',
              chunks: 'all'
            },
            quasar: {
              test: /[\\/]node_modules[\\/]quasar[\\/]/,
              name: 'quasar',
              chunks: 'all'
            }
          }
        })

        // Preload important chunks
        chain.plugin('preload').tap(options => {
          options[0] = {
            rel: 'preload',
            include: 'initial',
            fileBlacklist: [/\.map$/, /hot-update\.js$/]
          }
          return options
        })
      }
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
    },

    // PWA configuration for performance
    pwa: {
      workboxPluginMode: 'InjectManifest',
      workboxOptions: {
        maximumFileSizeToCacheInBytes: 5000000,
        skipWaiting: true,
        clientsClaim: true
      },
      manifest: {
        name: 'FM-Dash',
        short_name: 'FM-Dash',
        description: 'Football Manager Data Analysis Platform',
        display: 'standalone',
        orientation: 'portrait',
        background_color: '#ffffff',
        theme_color: '#1976D2'
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

## Image Storage Configuration

The application supports multiple methods for serving player faces and team logos, with automatic fallback mechanisms for maximum reliability.

### Image Loading Priority

When requesting player faces or team logos, the system follows this priority order:

1. **External Image API** (if `IMAGE_API_URL` is configured)
2. **S3/MinIO Storage** (if S3 is configured and enabled)
3. **Local File Storage** (fallback to local directories)

### External Image API Configuration

To use an external CDN or image API service (recommended for production):

```bash
# Set the base URL for your external image service
IMAGE_API_URL=https://sortitoutsi.b-cdn.net/uploads
```

When configured, the system will redirect image requests to:
- **Player Faces**: `{IMAGE_API_URL}/face/{playerUID}.png?width=256`
- **Team Logos**: `{IMAGE_API_URL}/team/{teamId}.png?width=256`

#### Example External URLs

For the CDN URL `https://sortitoutsi.b-cdn.net/uploads`:
- Player face: `https://sortitoutsi.b-cdn.net/uploads/face/123456.png?width=256`
- Team logo: `https://sortitoutsi.b-cdn.net/uploads/team/5000.png?width=256`

### Local Storage Configuration

For local development or when external APIs are unavailable:

```bash
# Local directory paths (relative to application root)
FACES_DIR=./faces                   # Directory containing player face images
LOGOS_DIR=./logos                   # Directory containing team logo images
```

Expected directory structure:
```
faces/
├── 123456.png                     # Player UID as filename
├── 789012.png
└── ...

logos/
└── Clubs/
    └── Normal/
        └── Normal/
            ├── 5000.png            # Team ID as filename
            ├── 5001.png
            └── ...
```

### S3/MinIO Storage Configuration

For cloud storage with S3-compatible services:

```bash
# S3 bucket configuration (uses same bucket as datasets by default)
S3_FACES_BUCKET=fm-dash-faces       # Optional: separate bucket for faces
S3_LOGOS_BUCKET=fm-dash-logos       # Optional: separate bucket for logos
```

Expected S3 object structure:
```
faces/
├── 123456.png
├── 789012.png
└── ...

logos/Clubs/Normal/Normal/
├── 5000.png
├── 5001.png
└── ...
```

### Configuration Examples

#### Production with External CDN

```bash
# Recommended for production - fastest performance
IMAGE_API_URL=https://your-cdn.com/fm24-images

# Fallback configuration (optional)
S3_FACES_BUCKET=fm-dash-faces
S3_LOGOS_BUCKET=fm-dash-logos
FACES_DIR=./faces
LOGOS_DIR=./logos
```

#### Development with Local Images

```bash
# For local development
FACES_DIR=./assets/faces
LOGOS_DIR=./assets/logos
```

#### Hybrid Configuration

```bash
# Use CDN for most images, S3 for fallback
IMAGE_API_URL=https://your-cdn.com/fm24-images
S3_ENDPOINT=s3.amazonaws.com
S3_BUCKET_NAME=fm-dash-data
S3_FACES_BUCKET=fm-dash-faces
S3_LOGOS_BUCKET=fm-dash-logos
```

### Image Format Requirements

- **Format**: PNG (recommended) or JPG
- **Naming**: 
  - Player faces: `{playerUID}.png`
  - Team logos: `{teamId}.png`
- **Size**: Any size (application will add `?width=256` parameter for optimization)

### Troubleshooting Image Issues

1. **Images not loading**: Check the browser developer console for 404 errors
2. **Slow image loading**: Consider using a CDN with the `IMAGE_API_URL` option
3. **CORS issues with external images**: Ensure your CDN supports CORS headers

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