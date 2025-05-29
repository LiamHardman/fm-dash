# Football Manager Player Parser

A tool for parsing and displaying Football Manager HTML player data with a Vue.js + Quasar UI interface and Go backend.

## Project Setup

### Backend (Go)

```bash
# Initialize Go module
go mod init fm24golang

# Install dependencies
go get golang.org/x/net/html
go mod tidy
```

### Frontend (Vue + Quasar)

```bash
# Install Node.js dependencies
npm install
```

## Running the Application

### Start the Backend API

```bash
# Start the Go backend API
go run main.go
```

The backend will run on http://localhost:8080.

### Start the Frontend Development Server

```bash
# Start the Vue.js frontend with hot-reload
npm run dev
```

The frontend will run on http://localhost:3000.

## How to Use

1. Start both the backend API and frontend development server
2. Access the frontend at http://localhost:3000
3. Choose one of the following options:
   - **Demo**: Click the "Demo" button on the homepage to view pre-loaded sample data (dataset ID: demo)
   - **Upload**: Select an HTML file exported from Football Manager containing player data and click "Upload and Parse"
4. Use search, sort, and pagination to explore the data

## Building for Production

```bash
# Build the frontend for production
npm run build
```

This will generate static files in the `dist` directory that can be served by any static file server.


# Querying LLM:

Use this -> codeweaver --ignore "node_modules,\.git"
or ./CodeWeaver --ignore "node_modules,testdata.html,.git,CodeWeaver"


## Configuration

### Environment Variables

#### Server Configuration
- `SERVER_PORT=8091` - Port for the API server (default: 8091)
- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed CORS origins (default: localhost:3000,localhost:8080)

#### Storage Configuration  
- `S3_ENDPOINT` - S3-compatible server endpoint (e.g., localhost:9000)
- `S3_ACCESS_KEY` - S3 access key
- `S3_SECRET_KEY` - S3 secret key
- `S3_BUCKET_NAME=v2fmdash` - S3 bucket name (default: v2fmdash)
- `S3_USE_SSL=false` - Whether to use SSL for S3 connection (default: false)
- `DATASETS_DIR=./datasets` - Directory for local dataset storage when S3 is disabled (default: ./datasets)

#### Observability Configuration
- `OTEL_ENABLED=true` - Enable OpenTelemetry tracing and metrics collection (default: false)
- `ENABLE_METRICS=true` - Enable detailed metrics collection (requires OTEL_ENABLED=true)
- `SERVICE_NAME=v2fmdash-api` - Service name for OpenTelemetry
- `OTEL_EXPORTER_OTLP_ENDPOINT=signoz.signoz:4317` - OTLP exporter endpoint
- `INSECURE_MODE=true` - Whether to use insecure gRPC connection for OTLP

### Storage Backend

The application supports three storage backends with automatic fallback:

1. **S3 Storage** (preferred): Data is persisted to S3-compatible object storage with in-memory fallback
2. **Hybrid Storage** (fallback when S3 is disabled): Combines in-memory storage with local file persistence  
3. **In-Memory Storage** (last resort): Data is stored in memory only and lost when the server restarts

#### Storage Selection Logic

- **S3 Available**: Uses S3 storage with in-memory fallback if S3 operations fail
- **S3 Disabled**: Uses hybrid storage (in-memory + local files) for persistence without S3
- **Fallback**: Uses in-memory only if local file storage fails to initialize

#### Hybrid Storage Behavior (when S3 is disabled)

When S3 is not configured or credentials are missing, the application uses hybrid storage:

- **Write Operations**: Data is stored in both memory (for fast access) and local files (for persistence)
- **Read Operations**: Checks memory first, then falls back to local files if not found in memory
- **Persistence**: Datasets are saved as compressed JSON files in the `DATASETS_DIR` directory
- **Performance**: Memory access for frequently used datasets, disk fallback for others
- **Caching**: Datasets loaded from disk are automatically cached in memory for faster subsequent access

To configure local dataset storage:

```bash
# Set custom datasets directory (optional)
export DATASETS_DIR=/path/to/your/datasets

# S3 disabled - will use hybrid storage
# (no S3_ENDPOINT configured)
go run main.go
```

The S3 bucket name defaults to `v2fmdash` but can be configured via `S3_BUCKET_NAME`.

### Observability

The application supports optional OpenTelemetry integration for **tracing**, **metrics**, and **log streaming**:

#### Enabling Observability

1. **Runtime Control**: Set `OTEL_ENABLED=true` to enable OpenTelemetry features
2. **Build-time Control**: Use build tags to completely exclude OTEL libraries:
   - Default build: `go build` (includes OTEL libraries, controlled by OTEL_ENABLED)
   - Minimal build: `go build -tags no_otel` (excludes OTEL libraries entirely)

#### Log Streaming to OTLP

When `OTEL_ENABLED=true`, all application logs are automatically streamed to your OTLP endpoint using structured logging:

```bash
# Enable log streaming to your observability backend
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=your-otlp-endpoint:4317
export SERVICE_NAME=v2fmdash-api

# Start the application - logs will stream to your OTLP endpoint
go run main.go
```

**Log Features:**
- Structured logging with key-value pairs for better searchability
- Automatic conversion from Go's `slog` to OTLP log format
- Dual output: logs appear both locally (console) and remotely (OTLP)
- Graceful fallback if OTLP endpoint is unavailable

#### Benefits of the OTEL_ENABLED Flag

- **Performance**: When disabled, OTEL libraries are still imported but no telemetry overhead occurs
- **Resource Usage**: Reduces memory and CPU usage when observability isn't needed
- **Deployment Flexibility**: Same binary can run with or without observability
- **Build Size**: Use `-tags no_otel` for smaller binaries without OTEL dependencies
- **Log Streaming**: Centralized log collection when enabled, local-only when disabled
