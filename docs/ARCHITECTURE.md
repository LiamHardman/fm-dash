# Architecture Guide

This document provides a comprehensive overview of the Football Manager Data Browser (FM-Dash) architecture, design decisions, and technical implementation details.

## System Overview

FM-Dash is a modern, cloud-native web application designed for processing and analyzing Football Manager player data. The architecture follows a microservices-inspired design with clear separation of concerns.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Storage       │
│   (Vue.js)      │◄──►│   (Go)          │◄──►│   (S3/MinIO)    │
│                 │    │                 │    │                 │
│ • Vue 3         │    │ • REST API      │    │ • File Storage  │
│ • Quasar UI     │    │ • HTML Parser   │    │ • Binary Data   │
│ • Vite Build    │    │ • Data Process  │    │ • Processed     │
│ • Pinia Store   │    │ • Analytics     │    │   Results       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Deployment    │
                    │                 │
                    │ • Docker        │
                    │ • Kubernetes    │
                    │ • Nginx         │
                    │ • OpenTelemetry │
                    └─────────────────┘
```

## Frontend Architecture

### Technology Stack

- **Vue.js 3**: Progressive framework with Composition API
- **Quasar Framework**: Material Design component library
- **Vite**: Modern build tool for fast development
- **Pinia**: State management with TypeScript support
- **Vitest**: Unit testing framework

### Component Structure

```
src/
├── components/          # Reusable UI components
│   ├── layout/         # Layout components (header, sidebar, etc.)
│   ├── data/           # Data display components (tables, charts)
│   ├── forms/          # Form components (upload, search, filters)
│   └── common/         # Common UI elements (buttons, modals)
├── pages/              # Page-level components
│   ├── Home.vue        # Landing page with upload
│   ├── Players.vue     # Player browser and search
│   ├── Analytics.vue   # Data visualization
│   └── Compare.vue     # Player comparison
├── composables/        # Reusable composition functions
│   ├── useApi.js       # API client logic
│   ├── useSearch.js    # Search functionality
│   └── useFilters.js   # Filter management
├── stores/             # Pinia state stores
│   ├── players.js      # Player data management
│   ├── ui.js           # UI state (modals, loading, etc.)
│   └── search.js       # Search state and history
├── services/           # Business logic services
│   ├── api.js          # HTTP client configuration
│   ├── upload.js       # File upload handling
│   └── export.js       # Data export functionality
└── utils/              # Utility functions
    ├── formatters.js   # Data formatting
    ├── validators.js   # Input validation
    └── constants.js    # Application constants
```

### State Management Pattern

FM-Dash uses Pinia for centralized state management with a clear data flow:

```javascript
// Example: Player Store Pattern
export const usePlayersStore = defineStore('players', () => {
  // State
  const players = ref([])
  const loading = ref(false)
  const filters = ref({})
  
  // Getters
  const filteredPlayers = computed(() => {
    return players.value.filter(player => 
      applyFilters(player, filters.value)
    )
  })
  
  // Actions
  async function fetchPlayers(params = {}) {
    loading.value = true
    try {
      const response = await api.get('/players', { params })
      players.value = response.data.players
    } finally {
      loading.value = false
    }
  }
  
  return {
    players: readonly(players),
    loading: readonly(loading),
    filteredPlayers,
    fetchPlayers,
    setFilters
  }
})
```

## Backend Architecture

### Technology Stack

- **Go 1.21+**: High-performance, concurrent language
- **Standard Library**: Minimal dependencies, native HTTP server
- **HTML Parser**: golang.org/x/net/html for Football Manager data
- **S3 Storage**: AWS S3 or MinIO for file storage
- **OpenTelemetry**: Observability and monitoring

### Service Layer Architecture

The backend follows a service-oriented architecture with clear separation of concerns:

```
src/api/
├── main.go                    # Application entry point
├── handlers.go                # HTTP route handlers
├── middleware.go              # HTTP middleware (CORS, auth, etc.)
├── services/                  # Business logic layer
│   ├── service_manager.go     # Service coordination
│   ├── processing_service.go  # File processing logic
│   ├── player_service.go      # Player data operations
│   └── search_service.go      # Search and filtering
├── errors/                    # Error handling
│   ├── errors.go              # Custom error types
│   └── middleware.go          # Error middleware
├── workers/                   # Background processing
└── utils/                     # Shared utilities
```

### Data Processing Pipeline

The core of FM-Dash is its sophisticated data processing pipeline:

```
HTML File Upload
       │
       ▼
┌─────────────────┐
│   File Upload   │ ← Validation, size limits
│   Handler       │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│   HTML Parser   │ ← Tokenization, DOM parsing
│                 │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│   Data          │ ← Cell extraction, validation
│   Extraction    │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│   Player        │ ← Attribute mapping, calculations
│   Processing    │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│   Enhancement   │ ← FIFA ratings, position groups
│                 │
└─────────────────┘
       │
       ▼
┌─────────────────┐
│   Storage       │ ← S3 upload, metadata storage
│                 │
└─────────────────┘
```

### Concurrency Model

Go's goroutines are used strategically for performance:

```go
// Example: Concurrent player processing
func ProcessPlayersInBatches(players []RawPlayer) []Player {
    const batchSize = 100
    const workers = runtime.NumCPU()
    
    jobs := make(chan []RawPlayer, len(players)/batchSize+1)
    results := make(chan []Player, len(players)/batchSize+1)
    
    // Start workers
    for w := 0; w < workers; w++ {
        go func() {
            for batch := range jobs {
                processed := make([]Player, len(batch))
                for i, raw := range batch {
                    processed[i] = ProcessPlayer(raw)
                }
                results <- processed
            }
        }()
    }
    
    // Distribute work
    for i := 0; i < len(players); i += batchSize {
        end := min(i+batchSize, len(players))
        jobs <- players[i:end]
    }
    close(jobs)
    
    // Collect results
    var allProcessed []Player
    for i := 0; i < (len(players)/batchSize + 1); i++ {
        allProcessed = append(allProcessed, <-results...)
    }
    
    return allProcessed
}
```

## Data Models

### Player Data Structure

The core data model represents a Football Manager player with all attributes:

```go
type Player struct {
    // Basic Information
    ID           int    `json:"id"`
    Name         string `json:"name"`
    Age          int    `json:"age"`
    Position     string `json:"position"`
    Club         string `json:"club"`
    Nationality  string `json:"nationality"`
    
    // Ratings
    Overall      int `json:"overall"`
    Potential    int `json:"potential"`
    FifaOverall  int `json:"fifa_overall"`
    
    // Financial
    TransferValue    string `json:"transfer_value"`
    TransferValueRaw int64  `json:"transfer_value_raw"`
    Wage            string `json:"wage"`
    WageRaw         int64  `json:"wage_raw"`
    
    // Processed Fields
    ParsedPositions []string            `json:"parsed_positions"`
    PositionGroups  []string            `json:"position_groups"`
    ShortPositions  []string            `json:"short_positions"`
    
    // Attributes (50+ technical, mental, physical)
    NumericAttributes map[string]int `json:"numeric_attributes"`
    
    // FIFA-style ratings
    PAC int `json:"pac"` // Pace
    SHO int `json:"sho"` // Shooting
    PAS int `json:"pas"` // Passing
    DRI int `json:"dri"` // Dribbling
    DEF int `json:"def"` // Defending
    PHY int `json:"phy"` // Physical
    
    // Performance Statistics
    PerformanceStats map[string]interface{} `json:"performance_stats"`
    
    // Metadata
    ProcessedAt time.Time `json:"processed_at"`
    Source      string    `json:"source"`
}
```

### Search and Filter Models

```go
type SearchRequest struct {
    SearchTerm string              `json:"search_term"`
    Filters    PlayerFilters       `json:"filters"`
    Sort       SortOptions         `json:"sort"`
    Pagination PaginationOptions   `json:"pagination"`
}

type PlayerFilters struct {
    Positions      []string      `json:"positions"`
    Clubs          []string      `json:"clubs"`
    Nationalities  []string      `json:"nationalities"`
    AgeRange       *RangeFilter  `json:"age_range"`
    OverallRange   *RangeFilter  `json:"overall_range"`
    ValueRange     *ValueRange   `json:"value_range"`
    Attributes     map[string]*RangeFilter `json:"attributes"`
}
```

## Storage Architecture

### File Storage Strategy

FM-Dash uses a tiered storage approach:

```
Storage Hierarchy:
├── uploads/              # Original HTML files
│   └── {uuid}/
│       ├── original.html
│       └── metadata.json
├── processed/            # Processed data
│   └── {uuid}/
│       ├── players.json
│       ├── statistics.json
│       └── index.json
└── exports/              # Generated exports
    └── {uuid}/
        ├── export.csv
        ├── export.xlsx
        └── expires_at.txt
```

### S3 Integration

```go
type StorageService struct {
    client   *s3.Client
    bucket   string
    region   string
    useSSL   bool
}

func (s *StorageService) StoreProcessedData(uploadID string, data *ProcessedData) error {
    key := fmt.Sprintf("processed/%s/players.json", uploadID)
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal data: %w", err)
    }
    
    _, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(jsonData),
        ContentType: aws.String("application/json"),
        Metadata: map[string]string{
            "upload-id":     uploadID,
            "players-count": strconv.Itoa(len(data.Players)),
            "processed-at":  time.Now().ISO8601(),
        },
    })
    
    return err
}
```

## Performance Optimizations

### Frontend Optimizations

1. **Virtual Scrolling**: For large player lists
```javascript
// Using Quasar's virtual scroll for 10k+ players
<q-virtual-scroll
  :items="players"
  :item-size="60"
  v-slot="{ item, index }"
>
  <PlayerRow :player="item" :key="index" />
</q-virtual-scroll>
```

2. **Lazy Loading**: Components load on demand
```javascript
// Route-level code splitting
const Players = () => import('../pages/Players.vue')
const Analytics = () => import('../pages/Analytics.vue')
```

3. **Debounced Search**: Reduce API calls
```javascript
const debouncedSearch = debounce(async (term) => {
  await searchPlayers(term)
}, 300)
```

### Backend Optimizations

1. **Memory-Efficient Parsing**: Stream-based HTML processing
2. **Goroutine Pools**: Controlled concurrency for CPU-intensive tasks
3. **Response Caching**: Cache frequently requested data
4. **Connection Pooling**: Reuse HTTP connections

### Caching Strategy

```
Cache Layers:
├── Browser Cache     # Static assets (24h)
├── CDN Cache        # API responses (5min)
├── Application Cache # Processed data (1h)
└── Database Cache   # Query results (15min)
```

## Security Considerations

### Input Validation

```go
func ValidateUploadFile(file multipart.File, header *multipart.FileHeader) error {
    // File size validation
    if header.Size > MaxFileSize {
        return ErrFileTooLarge
    }
    
    // File type validation
    if !strings.HasSuffix(header.Filename, ".html") {
        return ErrInvalidFileType
    }
    
    // Content validation - check for HTML markers
    buffer := make([]byte, 512)
    n, err := file.Read(buffer)
    if err != nil {
        return err
    }
    
    contentType := http.DetectContentType(buffer[:n])
    if !strings.Contains(contentType, "text/html") {
        return ErrInvalidContent
    }
    
    // Reset file pointer
    file.Seek(0, 0)
    return nil
}
```

### CORS Configuration

```go
func EnableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if isAllowedOrigin(origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }
        
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Max-Age", "3600")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Deployment Architecture

### Containerization

```dockerfile
# Multi-stage build for optimal image size
FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY src/ ./src/
COPY public/ ./public/
COPY *.config.js ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM nginx:alpine
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY --from=backend-builder /app/main /usr/local/bin/
COPY nginx.conf /etc/nginx/nginx.conf
COPY supervisord.conf /etc/supervisord.conf
COPY entrypoint.sh /entrypoint.sh

RUN apk add --no-cache supervisor
EXPOSE 8080
CMD ["/entrypoint.sh"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fmdb-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fmdb
  template:
    metadata:
      labels:
        app: fmdb
    spec:
      containers:
      - name: fmdb
        image: fmdb:latest
        ports:
        - containerPort: 8080
        env:
        - name: S3_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: s3-secret
              key: endpoint
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
```

## Monitoring and Observability

### OpenTelemetry Integration

```go
func initTracing() {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
    if err != nil {
        log.Fatal(err)
    }
    
    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("fmdb-backend"),
            semconv.ServiceVersionKey.String("v1.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
}

// Usage in handlers
func (h *Handler) GetPlayers(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer("fmdb").Start(r.Context(), "get_players")
    defer span.End()
    
    players, err := h.playerService.GetPlayers(ctx, filters)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    span.SetAttributes(
        attribute.Int("players.count", len(players)),
        attribute.String("filters.position", filters.Position),
    )
    
    json.NewEncoder(w).Encode(players)
}
```

### Health Checks

```go
type HealthCheck struct {
    Status      string            `json:"status"`
    Checks      map[string]string `json:"checks"`
    Uptime      string            `json:"uptime"`
    MemoryUsage string            `json:"memory_usage"`
    Goroutines  int               `json:"goroutines"`
    Timestamp   time.Time         `json:"timestamp"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    checks := make(map[string]string)
    
    // Check S3 connectivity
    if err := h.storage.Ping(); err != nil {
        checks["s3"] = "unhealthy: " + err.Error()
    } else {
        checks["s3"] = "healthy"
    }
    
    // Memory stats
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    health := HealthCheck{
        Status:      "healthy",
        Checks:      checks,
        Uptime:      time.Since(startTime).String(),
        MemoryUsage: fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
        Goroutines:  runtime.NumGoroutine(),
        Timestamp:   time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

## Design Decisions

### Why Go for Backend?

1. **Performance**: Native compilation, efficient memory usage
2. **Concurrency**: Goroutines for parallel processing
3. **Standard Library**: Rich HTTP and HTML parsing capabilities
4. **Deployment**: Single binary, easy containerization
5. **Maintenance**: Strong typing, excellent tooling

### Why Vue.js + Quasar?

1. **Developer Experience**: Composition API, reactive system
2. **Component Library**: Material Design, comprehensive components
3. **Performance**: Virtual DOM, efficient updates
4. **Build System**: Vite for fast development
5. **Ecosystem**: Rich plugin ecosystem, TypeScript support

### Why S3 Storage?

1. **Scalability**: Handles large files efficiently
2. **Durability**: 99.999999999% (11 9's) durability
3. **Cost**: Pay-as-you-use pricing model
4. **Integration**: Works with MinIO for local development
5. **Features**: Versioning, lifecycle policies, encryption

---

This architecture ensures FM-Dash is scalable, maintainable, and performant while providing an excellent user experience for Football Manager data analysis. 