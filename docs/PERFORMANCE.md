# Performance Guide

This document covers optimization techniques and performance considerations for the Football Manager Data Browser (FM-Dash).

## Performance Overview

FM-Dash is designed to handle large Football Manager datasets efficiently while providing responsive user interactions. Performance optimization focuses on several key areas:

- **File Processing**: Efficiently parsing large HTML exports (50MB+)
- **Memory Management**: Handling thousands of players in memory
- **Search Performance**: Fast filtering and search across large datasets
- **UI Responsiveness**: Smooth interactions with large data tables
- **Network Optimization**: Efficient API responses and caching

## Backend Performance

### CPU Optimization

#### Goroutine Configuration

```bash
# Optimal settings for different CPU configurations

# 4-core systems
export WORKER_COUNT=4
export BATCH_SIZE=100

# 8-core systems
export WORKER_COUNT=8
export BATCH_SIZE=200

# 16-core systems
export WORKER_COUNT=12
export BATCH_SIZE=300
```

#### Processing Optimization

```go
// Example: Optimized player processing
func ProcessPlayersOptimized(players []RawPlayer) []Player {
    const (
        batchSize = 200
        workerCount = runtime.NumCPU()
        bufferSize = workerCount * 2
    )
    
    jobs := make(chan []RawPlayer, bufferSize)
    results := make(chan []Player, bufferSize)
    
    // Worker pool with optimal sizing
    var wg sync.WaitGroup
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for batch := range jobs {
                processed := processPlayerBatch(batch)
                results <- processed
            }
        }()
    }
    
    // Distribute work efficiently
    go func() {
        defer close(jobs)
        for i := 0; i < len(players); i += batchSize {
            end := min(i+batchSize, len(players))
            jobs <- players[i:end]
        }
    }()
    
    // Collect results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    var allProcessed []Player
    for batch := range results {
        allProcessed = append(allProcessed, batch...)
    }
    
    return allProcessed
}
```

### Memory Optimization

#### Memory Configuration

```bash
# Memory tuning for different scenarios

# Low memory systems (< 2GB)
export MEMORY_LIMIT=1GB
export WORKER_COUNT=2
export BATCH_SIZE=50
export CACHE_MAX_SIZE=50MB
export GOGC=50  # Aggressive garbage collection

# Standard systems (2-8GB)
export MEMORY_LIMIT=4GB
export WORKER_COUNT=4
export BATCH_SIZE=100
export CACHE_MAX_SIZE=200MB
export GOGC=100

# High memory systems (8GB+)
export MEMORY_LIMIT=8GB
export WORKER_COUNT=8
export BATCH_SIZE=300
export CACHE_MAX_SIZE=500MB
export GOGC=200  # Less frequent GC
```

#### Memory Profiling

```bash
# Monitor memory usage
watch -n 5 'curl -s http://localhost:8091/api/health | jq .memory_usage'

# Generate memory profile
curl http://localhost:8091/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Analyze memory allocations
curl http://localhost:8091/debug/pprof/allocs > allocs.prof
go tool pprof allocs.prof

# Check for memory leaks
curl http://localhost:8091/debug/pprof/heap?debug=1 | grep -E "(heap|stack)"
```

#### Memory-Efficient Patterns

```go
// Efficient string handling
var stringPool = sync.Pool{
    New: func() interface{} {
        return make([]string, 0, 100)
    },
}

func processPlayerStrings(data []string) {
    buffer := stringPool.Get().([]string)
    defer func() {
        buffer = buffer[:0]
        stringPool.Put(buffer)
    }()
    
    // Process strings using pooled buffer
}

// Stream processing for large files
func ProcessLargeFile(reader io.Reader) error {
    scanner := bufio.NewScanner(reader)
    scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 1MB buffer
    
    for scanner.Scan() {
        line := scanner.Text()
        // Process line without storing all in memory
        processLine(line)
    }
    
    return scanner.Err()
}
```

### I/O Optimization

#### File Processing

```bash
# Optimize file I/O
export READ_BUFFER_SIZE=64KB
export WRITE_BUFFER_SIZE=64KB
export MAX_CONCURRENT_UPLOADS=5

# Stream processing for large files
export STREAM_PROCESSING=true
export CHUNK_SIZE=1MB
```

#### S3 Performance

```bash
# S3 optimization settings
export S3_MAX_RETRIES=3
export S3_RETRY_DELAY=1s
export S3_TIMEOUT=30s
export S3_MAX_CONCURRENT_REQUESTS=10

# Connection pooling
export S3_MAX_IDLE_CONNS=20
export S3_MAX_IDLE_CONNS_PER_HOST=10
export S3_IDLE_CONN_TIMEOUT=90s
```

### Caching Strategy

#### Application-Level Caching

```go
// LRU cache implementation
type PlayerCache struct {
    cache    *lru.Cache
    mutex    sync.RWMutex
    ttl      time.Duration
    cleanup  time.Duration
}

func NewPlayerCache(size int, ttl time.Duration) *PlayerCache {
    cache, _ := lru.New(size)
    
    pc := &PlayerCache{
        cache:   cache,
        ttl:     ttl,
        cleanup: ttl / 4,
    }
    
    // Start cleanup goroutine
    go pc.cleanupExpired()
    
    return pc
}

func (pc *PlayerCache) Get(key string) (*Player, bool) {
    pc.mutex.RLock()
    defer pc.mutex.RUnlock()
    
    if item, ok := pc.cache.Get(key); ok {
        if entry := item.(*cacheEntry); time.Since(entry.timestamp) < pc.ttl {
            return entry.player, true
        }
        pc.cache.Remove(key)
    }
    
    return nil, false
}
```

#### Cache Configuration

```yaml
cache:
  enabled: true
  
  # Player data cache
  players:
    max_size: 10000      # Number of players
    ttl: "1h"            # Time to live
    
  # Search results cache
  search:
    max_size: 1000       # Number of search queries
    ttl: "15m"
    
  # API response cache
  responses:
    max_size: "100MB"    # Memory size
    ttl: "5m"
    
  # Static data cache
  static:
    enabled: true
    ttl: "24h"
```

## Frontend Performance

### Bundle Optimization

#### Code Splitting

```javascript
// Route-level code splitting
const router = createRouter({
  routes: [
    {
      path: '/',
      component: () => import('../pages/Home.vue')
    },
    {
      path: '/players',
      component: () => import('../pages/Players.vue')
    },
    {
      path: '/analytics',
      component: () => import('../pages/Analytics.vue')
    },
    {
      path: '/compare',
      component: () => import('../pages/Compare.vue')
    }
  ]
})

// Component-level code splitting
export default {
  components: {
    PlayerTable: () => import('./PlayerTable.vue'),
    PlayerCard: () => import('./PlayerCard.vue'),
    PlayerFilters: () => import('./PlayerFilters.vue')
  }
}
```

#### Build Optimization

```javascript
// vite.config.js optimization
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['quasar'],
          charts: ['chart.js', 'vue-chartjs'],
          utils: ['lodash', 'date-fns'],
          
          // Feature chunks
          player: ['./src/components/player/*'],
          search: ['./src/components/search/*'],
          analytics: ['./src/components/analytics/*']
        }
      }
    },
    
    // Compression
    cssCodeSplit: true,
    sourcemap: false, // Disable in production
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    }
  },
  
  // Asset optimization
  assetsInclude: ['**/*.woff2'],
  
  // Dependency optimization
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'quasar']
  }
})
```

### Data Handling Optimization

#### Virtual Scrolling

```vue
<template>
  <!-- Large dataset handling with virtual scrolling -->
  <q-virtual-scroll
    :items="filteredPlayers"
    :item-size="playerRowHeight"
    v-slot="{ item, index }"
    style="max-height: 600px"
    @virtual-scroll="onVirtualScroll"
  >
    <PlayerRow 
      :key="item.id"
      :player="item" 
      :index="index"
      @click="selectPlayer(item)"
    />
  </q-virtual-scroll>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { usePlayersStore } from '../stores/players'

const playersStore = usePlayersStore()
const playerRowHeight = 60

// Efficient filtering with memoization
const filteredPlayers = computed(() => {
  return playersStore.players.filter(player => {
    return applyFilters(player, playersStore.filters)
  })
})

// Virtual scroll optimization
const onVirtualScroll = (info) => {
  // Load more data if needed
  if (info.index > filteredPlayers.value.length - 50) {
    playersStore.loadMorePlayers()
  }
}
</script>
```

#### Debounced Search

```javascript
// Optimized search with debouncing
import { debounce } from 'lodash-es'

export const usePlayerSearch = () => {
  const searchTerm = ref('')
  const searchResults = ref([])
  const isSearching = ref(false)
  
  // Debounced search function
  const debouncedSearch = debounce(async (term) => {
    if (!term.trim()) {
      searchResults.value = []
      return
    }
    
    isSearching.value = true
    try {
      const results = await api.searchPlayers({
        search: term,
        limit: 50 // Limit initial results
      })
      searchResults.value = results
    } catch (error) {
      console.error('Search failed:', error)
    } finally {
      isSearching.value = false
    }
  }, 300)
  
  // Watch for search term changes
  watch(searchTerm, (newTerm) => {
    debouncedSearch(newTerm)
  })
  
  return {
    searchTerm,
    searchResults,
    isSearching
  }
}
```

#### State Management Optimization

```javascript
// Optimized Pinia store
export const usePlayersStore = defineStore('players', () => {
  // State
  const players = ref([])
  const searchCache = new Map()
  const filterCache = new Map()
  
  // Getters with caching
  const filteredPlayers = computed(() => {
    const cacheKey = JSON.stringify(filters.value)
    
    if (filterCache.has(cacheKey)) {
      return filterCache.get(cacheKey)
    }
    
    const filtered = players.value.filter(player => 
      applyFilters(player, filters.value)
    )
    
    // Cache results (with size limit)
    if (filterCache.size > 100) {
      const firstKey = filterCache.keys().next().value
      filterCache.delete(firstKey)
    }
    filterCache.set(cacheKey, filtered)
    
    return filtered
  })
  
  // Batch operations
  const updateMultiplePlayers = (updates) => {
    const newPlayers = [...players.value]
    
    updates.forEach(({ id, data }) => {
      const index = newPlayers.findIndex(p => p.id === id)
      if (index !== -1) {
        newPlayers[index] = { ...newPlayers[index], ...data }
      }
    })
    
    players.value = newPlayers
  }
  
  return {
    players: readonly(players),
    filteredPlayers,
    updateMultiplePlayers
  }
})
```

## Database/Search Performance

### Search Optimization

#### Indexed Search

```go
// Search index for fast filtering
type SearchIndex struct {
    byName        map[string][]*Player
    byPosition    map[string][]*Player
    byClub        map[string][]*Player
    byNationality map[string][]*Player
    byOverall     map[int][]*Player
    mutex         sync.RWMutex
}

func (si *SearchIndex) BuildIndex(players []*Player) {
    si.mutex.Lock()
    defer si.mutex.Unlock()
    
    // Clear existing indices
    si.byName = make(map[string][]*Player)
    si.byPosition = make(map[string][]*Player)
    si.byClub = make(map[string][]*Player)
    si.byNationality = make(map[string][]*Player)
    si.byOverall = make(map[int][]*Player)
    
    // Build indices
    for _, player := range players {
        // Name index (lowercase for case-insensitive search)
        nameKey := strings.ToLower(player.Name)
        si.byName[nameKey] = append(si.byName[nameKey], player)
        
        // Position index
        for _, pos := range player.ParsedPositions {
            si.byPosition[pos] = append(si.byPosition[pos], player)
        }
        
        // Club index
        si.byClub[player.Club] = append(si.byClub[player.Club], player)
        
        // Nationality index
        si.byNationality[player.Nationality] = append(si.byNationality[player.Nationality], player)
        
        // Overall rating index
        si.byOverall[player.Overall] = append(si.byOverall[player.Overall], player)
    }
}

func (si *SearchIndex) Search(query SearchQuery) []*Player {
    si.mutex.RLock()
    defer si.mutex.RUnlock()
    
    var candidates []*Player
    
    // Use most selective filter first
    if query.Name != "" {
        nameKey := strings.ToLower(query.Name)
        candidates = si.byName[nameKey]
    } else if query.Position != "" {
        candidates = si.byPosition[query.Position]
    } else if query.Club != "" {
        candidates = si.byClub[query.Club]
    } else {
        // No specific filter, return all players
        for _, players := range si.byName {
            candidates = append(candidates, players...)
        }
    }
    
    // Apply additional filters
    return si.applyFilters(candidates, query)
}
```

#### Pagination Optimization

```go
// Efficient pagination
type PaginatedResult struct {
    Players    []*Player `json:"players"`
    TotalCount int       `json:"total_count"`
    Page       int       `json:"page"`
    Limit      int       `json:"limit"`
    HasMore    bool      `json:"has_more"`
}

func (s *PlayerService) GetPlayersPaginated(filters PlayerFilters, page, limit int) (*PaginatedResult, error) {
    // Use cursor-based pagination for better performance
    offset := (page - 1) * limit
    
    // Get total count efficiently
    totalCount := s.searchIndex.Count(filters)
    
    // Get players for current page
    allFiltered := s.searchIndex.Search(filters)
    
    // Apply pagination
    start := offset
    if start > len(allFiltered) {
        start = len(allFiltered)
    }
    
    end := start + limit
    if end > len(allFiltered) {
        end = len(allFiltered)
    }
    
    players := allFiltered[start:end]
    
    return &PaginatedResult{
        Players:    players,
        TotalCount: totalCount,
        Page:       page,
        Limit:      limit,
        HasMore:    end < len(allFiltered),
    }, nil
}
```

## Network Performance

### API Optimization

#### Response Compression

```go
// Gzip compression middleware
func GzipMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Set("Vary", "Accept-Encoding")
        
        gz := gzip.NewWriter(w)
        defer gz.Close()
        
        gzw := &gzipResponseWriter{
            ResponseWriter: w,
            Writer:         gz,
        }
        
        next.ServeHTTP(gzw, r)
    })
}
```

#### Efficient JSON Responses

```go
// Optimized JSON encoding
func (h *Handler) GetPlayersOptimized(w http.ResponseWriter, r *http.Request) {
    players, err := h.playerService.GetPlayers(r.Context(), filters)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Set headers for optimization
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Cache-Control", "public, max-age=300") // 5 min cache
    
    // Use streaming JSON encoder for large responses
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "") // No indentation for smaller size
    
    if err := encoder.Encode(players); err != nil {
        log.Printf("Failed to encode response: %v", err)
    }
}
```

#### HTTP/2 and Connection Optimization

```bash
# Nginx HTTP/2 configuration
server {
    listen 443 ssl http2;
    
    # Connection optimization
    keepalive_timeout 65;
    keepalive_requests 1000;
    
    # Compression
    gzip on;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml;
    
    # Caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    location /api/ {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
```

### CDN and Caching

#### Browser Caching

```javascript
// Service Worker for aggressive caching
self.addEventListener('fetch', event => {
  if (event.request.url.includes('/api/players')) {
    event.respondWith(
      caches.open('api-cache').then(cache => {
        return cache.match(event.request).then(response => {
          if (response) {
            // Return cached response
            return response
          }
          
          // Fetch and cache
          return fetch(event.request).then(response => {
            if (response.status === 200) {
              cache.put(event.request, response.clone())
            }
            return response
          })
        })
      })
    )
  }
})
```

## Monitoring and Benchmarking

### Performance Metrics

#### Backend Metrics

```go
// Custom metrics collection
var (
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )
    
    processingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "file_processing_duration_seconds",
            Help:    "Duration of file processing",
            Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
        },
        []string{"file_size_mb"},
    )
    
    memoryUsage = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Current memory usage",
        },
    )
)

// Metrics middleware
func MetricsMiddleware(next http.Handler) http.Handler {
    return promhttp.InstrumentHandlerDuration(
        httpRequestDuration,
        next,
    )
}
```

#### Frontend Performance Monitoring

```javascript
// Performance monitoring
export const performanceMonitor = {
  // Measure page load time
  measurePageLoad() {
    const navigation = performance.getEntriesByType('navigation')[0]
    return {
      domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
      loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
      totalTime: navigation.loadEventEnd - navigation.fetchStart
    }
  },
  
  // Measure API response time
  measureApiCall(apiCall) {
    const start = performance.now()
    
    return apiCall().then(result => {
      const duration = performance.now() - start
      
      // Send to analytics
      this.sendMetric('api_call_duration', duration)
      
      return result
    })
  },
  
  // Measure component render time
  measureComponentRender(componentName, renderFn) {
    const start = performance.now()
    const result = renderFn()
    const duration = performance.now() - start
    
    if (duration > 16) { // > 1 frame at 60fps
      console.warn(`Slow render detected for ${componentName}: ${duration}ms`)
    }
    
    return result
  }
}
```

### Benchmarking

#### Load Testing

```bash
# Apache Bench testing
ab -n 1000 -c 10 http://localhost:8091/api/players

# Artillery.js load testing
artillery quick --count 100 --num 10 http://localhost:8091/api/health

# Custom load test script
cat > load-test.yml << EOF
config:
  target: 'http://localhost:8091'
  phases:
    - duration: 60
      arrivalRate: 10
    - duration: 120
      arrivalRate: 20
    - duration: 60
      arrivalRate: 10

scenarios:
  - name: "Upload and search"
    flow:
      - post:
          url: "/api/upload"
          formData:
            file: "@test-data.html"
      - get:
          url: "/api/players"
          qs:
            search: "messi"
            limit: 50
EOF

artillery run load-test.yml
```

#### Memory Benchmarking

```go
// Benchmark tests
func BenchmarkPlayerProcessing(b *testing.B) {
    players := generateTestPlayers(10000)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        ProcessPlayers(players)
    }
}

func BenchmarkSearchIndex(b *testing.B) {
    index := NewSearchIndex()
    players := generateTestPlayers(50000)
    index.BuildIndex(players)
    
    query := SearchQuery{
        Position: "ST",
        MinOverall: 80,
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        index.Search(query)
    }
}
```

## Performance Best Practices

### Development Best Practices

1. **Use Profiling Early**: Profile during development, not just in production
2. **Measure Everything**: Don't optimize without measuring
3. **Optimize Bottlenecks**: Focus on the slowest parts first
4. **Test with Real Data**: Use realistic data sizes for testing
5. **Monitor Continuously**: Set up monitoring from day one

### Code Optimization

1. **Minimize Allocations**: Reuse objects and use object pools
2. **Efficient Data Structures**: Choose the right data structure for each use case
3. **Lazy Loading**: Load data only when needed
4. **Batch Operations**: Group multiple operations together
5. **Cache Wisely**: Cache expensive computations and data

### Infrastructure Optimization

1. **Resource Allocation**: Right-size your containers and VMs
2. **Network Optimization**: Use CDNs and edge caching
3. **Database Optimization**: Index frequently queried fields
4. **Load Balancing**: Distribute load across multiple instances
5. **Auto-scaling**: Scale resources based on demand

---

Regular performance monitoring and optimization ensure FM-Dash remains fast and responsive as datasets grow and user load increases. 