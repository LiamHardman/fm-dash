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

- **Vue.js 3**: Progressive framework with Composition API and performance optimizations
- **Quasar Framework**: Material Design component library with tree-shaking support
- **Vite**: Modern build tool with advanced code splitting and optimization
- **Pinia**: State management with memory-efficient patterns and caching
- **Vitest**: Unit testing framework with performance testing capabilities
- **Performance Tools**: Core Web Vitals monitoring, bundle analysis, and memory profiling

### Component Structure

```
src/
├── components/          # Reusable UI components
│   ├── layout/         # Layout components (header, sidebar, etc.)
│   ├── data/           # Data display components (tables, charts)
│   ├── forms/          # Form components (upload, search, filters)
│   ├── common/         # Common UI elements (buttons, modals)
│   ├── filters/        # Advanced filtering components
│   └── player-details/ # Player detail components
├── pages/              # Page-level components (lazy-loaded)
│   ├── LandingPage.vue      # Landing page with upload
│   ├── DatasetPage.vue      # Player browser and search
│   ├── PerformancePage.vue  # Performance monitoring
│   ├── TeamViewPage.vue     # Team analysis
│   └── PlayerUploadPage.vue # File upload interface
├── composables/        # Reusable composition functions
│   ├── useApi.js            # API client logic
│   ├── usePlayerFilters.js  # Advanced filtering
│   ├── useVirtualScrolling.js # Virtual scrolling optimization
│   ├── usePerformanceOptimizations.js # Performance utilities
│   ├── useMemoization.js    # Caching and memoization
│   └── useWebWorkers.js     # Web worker management
├── stores/             # Pinia state stores
│   ├── playerStore.js  # Player data with memory optimization
│   ├── uiStore.js      # UI state management
│   └── wishlistStore.js # Wishlist functionality
├── services/           # Business logic services
│   ├── api.js          # HTTP client configuration
│   ├── playerService.js # Player data operations
│   ├── wishlistService.js # Wishlist management
│   └── analytics.js    # Performance analytics
├── utils/              # Utility functions
│   ├── performance.js  # Performance monitoring utilities
│   ├── imageOptimization.js # Image loading optimization
│   ├── formationCache.js # Formation caching
│   └── security.js     # Security utilities
└── workers/            # Web workers for background processing
    └── playerCalculationWorker.js # Player calculations
```

### Performance-Optimized State Management

FM-Dash uses Pinia with advanced performance optimizations for handling large datasets. The store implementation includes comprehensive memoization and caching strategies:

**Current Implementation Status:**
- ✅ Basic shallowRef optimization for large arrays
- ✅ LRU cache for filtered results
- ✅ Advanced memoization with standalone utility functions
- 🔄 Enhanced virtual scrolling integration
- 🔄 Web worker data processing
- 🔄 Advanced memory management with object pooling

```javascript
// Performance-optimized Player Store Pattern (Enhanced)
export const usePlayersStore = defineStore('players', () => {
  // State - using shallowRef for large arrays to avoid deep reactivity overhead
  const players = shallowRef([])
  const loading = ref(false)
  const filters = ref({})
  
  // Enhanced LRU Cache for filtered results with better memory management
  const filterCache = new LRUCache(100)
  const virtualScrollCache = new Map()
  
  // Memory-efficient getters with advanced caching
  const filteredPlayers = computed(() => {
    const cacheKey = JSON.stringify(filters.value)
    
    const cached = filterCache.get(cacheKey)
    if (cached) return cached
    
    const filtered = players.value.filter(player => 
      applyFilters(player, filters.value)
    )
    
    filterCache.set(cacheKey, filtered)
    return filtered
  })
  
  // Enhanced virtual scrolling support
  const getVirtualScrollData = (startIndex, endIndex) => {
    const cacheKey = `${startIndex}-${endIndex}`
    
    if (virtualScrollCache.has(cacheKey)) {
      return virtualScrollCache.get(cacheKey)
    }
    
    const slice = filteredPlayers.value.slice(startIndex, endIndex)
    virtualScrollCache.set(cacheKey, slice)
    
    // Cleanup old cache entries
    if (virtualScrollCache.size > 50) {
      const firstKey = virtualScrollCache.keys().next().value
      virtualScrollCache.delete(firstKey)
    }
    
    return slice
  }
  
  // Web worker integration for heavy operations
  const processDataInWorker = async (operation, data) => {
    const worker = new Worker('/src/workers/playerCalculationWorker.js')
    
    return new Promise((resolve, reject) => {
      worker.postMessage({ operation, data })
      
      worker.onmessage = (e) => {
        resolve(e.data)
        worker.terminate()
      }
      
      worker.onerror = (error) => {
        reject(error)
        worker.terminate()
      }
    })
  }
  
  // Batch operations to minimize reactivity triggers
  async function fetchPlayers(params = {}) {
    loading.value = true
    try {
      const response = await api.get('/players', { params })
      // Single assignment to trigger reactivity once
      players.value = response.data.players
      
      // Clear caches when new data is loaded
      filterCache.clear()
      virtualScrollCache.clear()
    } finally {
      loading.value = false
    }
  }
  
  // Enhanced memory cleanup for component unmounting
  const cleanup = () => {
    filterCache.clear()
    virtualScrollCache.clear()
  }
  
  return {
    players: readonly(players),
    loading: readonly(loading),
    filteredPlayers,
    getVirtualScrollData,
    processDataInWorker,
    fetchPlayers,
    setFilters,
    cleanup
  }
})
```

### Frontend Performance Architecture

#### Bundle Optimization Strategy

```
Build Pipeline:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Source Code   │───►│   Vite Build    │───►│  Optimized      │
│                 │    │                 │    │  Bundles        │
│ • Vue SFCs      │    │ • Code Splitting│    │                 │
│ • TypeScript    │    │ • Tree Shaking  │    │ • vue-core.js   │
│ • SCSS          │    │ • Minification  │    │ • ui-framework  │
│ • Assets        │    │ • Compression   │    │ • charts.js     │
└─────────────────┘    └─────────────────┘    │ • utils.js      │
                                              └─────────────────┘

Dependency Optimization:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Third-Party     │───►│ Pre-bundled     │───►│ Optimized       │
│ Libraries       │    │ Dependencies    │    │ Loading         │
│                 │    │                 │    │                 │
│ • Chart.js      │    │ • chart.js/     │    │ • Faster dev    │
│ • VueUse        │    │   helpers       │    │   server        │
│ • Quasar        │    │ • Specific      │    │ • Better        │
│ • Dev Tools     │    │   imports       │    │   caching       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

#### Virtual Scrolling Architecture

```javascript
// Advanced Virtual Scrolling Implementation
class VirtualScrollManager {
  constructor(container, itemHeight, bufferSize = 5) {
    this.container = container
    this.itemHeight = itemHeight
    this.bufferSize = bufferSize
    this.viewportHeight = container.clientHeight
    this.visibleCount = Math.ceil(this.viewportHeight / itemHeight)
    
    // Object pool for DOM elements
    this.elementPool = new ObjectPool(() => this.createElement())
    
    // Intersection Observer for efficient visibility detection
    this.observer = new IntersectionObserver(this.handleIntersection.bind(this))
  }

  updateVisibleItems(items, scrollTop) {
    const startIndex = Math.floor(scrollTop / this.itemHeight)
    const endIndex = Math.min(
      startIndex + this.visibleCount + this.bufferSize,
      items.length
    )
    
    // Efficiently update only changed items
    this.renderVisibleRange(items.slice(startIndex, endIndex), startIndex)
  }

  renderVisibleRange(items, startIndex) {
    // Use document fragment for efficient DOM updates
    const fragment = document.createDocumentFragment()
    
    items.forEach((item, index) => {
      const element = this.elementPool.acquire()
      this.updateElement(element, item, startIndex + index)
      fragment.appendChild(element)
    })
    
    // Single DOM update
    this.container.appendChild(fragment)
  }
}
```

#### Memory Management Architecture

```javascript
// Memory Management System
class MemoryManager {
  constructor() {
    this.objectPools = new Map()
    this.caches = new Map()
    this.memoryThreshold = 200 * 1024 * 1024 // 200MB
    this.monitoringInterval = null
  }

  createObjectPool(name, factory, resetFn, initialSize = 10) {
    const pool = new ObjectPool(factory, resetFn, initialSize)
    this.objectPools.set(name, pool)
    return pool
  }

  createLRUCache(name, maxSize, ttl) {
    const cache = new LRUCache(maxSize, ttl)
    this.caches.set(name, cache)
    return cache
  }

  startMemoryMonitoring() {
    this.monitoringInterval = setInterval(() => {
      if (performance.memory) {
        const used = performance.memory.usedJSHeapSize
        
        if (used > this.memoryThreshold) {
          this.performCleanup()
        }
      }
    }, 30000) // Check every 30 seconds
  }

  performCleanup() {
    // Clear caches
    this.caches.forEach(cache => cache.clear())
    
    // Return objects to pools
    this.objectPools.forEach(pool => pool.releaseAll())
    
    // Force garbage collection if available
    if (window.gc) {
      window.gc()
    }
  }
}
```

#### Image Loading Architecture

```javascript
// Progressive Image Loading System
class ImageLoadingSystem {
  constructor() {
    this.loadingQueue = new PriorityQueue()
    this.cache = new Map()
    this.observers = new Map()
    this.maxConcurrent = 6
    this.currentLoading = 0
  }

  loadImage(src, options = {}) {
    const {
      priority = 'normal',
      formats = ['avif', 'webp', 'jpg'],
      progressive = true,
      lazy = true
    } = options

    if (this.cache.has(src)) {
      return Promise.resolve(this.cache.get(src))
    }

    const loadPromise = new Promise((resolve, reject) => {
      const task = {
        src,
        formats,
        progressive,
        priority: this.getPriorityValue(priority),
        resolve,
        reject
      }

      if (lazy) {
        this.queueForLazyLoading(task)
      } else {
        this.loadingQueue.enqueue(task)
        this.processQueue()
      }
    })

    this.cache.set(src, loadPromise)
    return loadPromise
  }

  async processQueue() {
    if (this.currentLoading >= this.maxConcurrent || this.loadingQueue.isEmpty()) {
      return
    }

    this.currentLoading++
    const task = this.loadingQueue.dequeue()

    try {
      const result = await this.loadSingleImage(task)
      task.resolve(result)
    } catch (error) {
      task.reject(error)
    } finally {
      this.currentLoading--
      this.processQueue() // Process next task
    }
  }
}
```

## Backend Architecture

### Technology Stack

- **Go 1.24+**: High-performance, concurrent language with latest optimizations
- **Standard Library**: Minimal dependencies, native HTTP server
- **HTML Parser**: golang.org/x/net/html for Football Manager data
- **Apache Arrow**: Columnar data processing for high-performance analytics
- **Parquet**: Efficient columnar storage format for large datasets
- **S3 Storage**: AWS S3 or MinIO for file storage
- **Compression**: klauspost/compress for optimized data compression
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