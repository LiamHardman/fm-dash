# Performance Guide

This document covers optimization techniques and performance considerations for the Football Manager Data Browser (FM-Dash).

## Performance Overview

FM-Dash is designed to handle large Football Manager datasets efficiently while providing responsive user interactions. Performance optimization focuses on several key areas:

- **File Processing**: Efficiently parsing large HTML exports (50MB+)
- **Memory Management**: Handling thousands of players in memory
- **Search Performance**: Fast filtering and search across large datasets
- **UI Responsiveness**: Smooth interactions with large data tables
- **Network Optimization**: Efficient API responses and caching
- **Frontend Bundle Optimization**: Minimal initial load times and efficient code splitting
- **Mobile Performance**: Optimized touch interactions and battery usage
- **Image Loading**: Modern formats with lazy loading and caching strategies

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

### Bundle Optimization and Code Splitting

FM-Dash implements advanced bundle optimization strategies to achieve fast initial load times and efficient caching.

#### Current Bundle Configuration

The application uses Vite with optimized chunking strategy:

```javascript
// vite.config.js - Current optimized configuration
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Core Vue framework
          'vue-core': ['vue', 'vue-router', 'pinia'],
          // UI framework
          'ui-framework': ['quasar'],
          // Charts and visualization
          charts: ['chart.js', 'vue-chartjs', 'chartjs-plugin-annotation'],
          // Utilities and composables
          utils: ['@vueuse/core']
        }
      }
    },
    chunkSizeWarningLimit: 800, // Encourage smaller chunks
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: process.env.NODE_ENV === 'production',
        drop_debugger: true,
        passes: 2 // Multiple compression passes
      }
    }
  }
})
```

#### Route-Level Code Splitting

```javascript
// Router configuration with lazy loading
const router = createRouter({
  routes: [
    {
      path: '/',
      component: () => import('../pages/LandingPage.vue')
    },
    {
      path: '/upload',
      component: () => import('../pages/PlayerUploadPage.vue')
    },
    {
      path: '/dataset/:id',
      component: () => import('../pages/DatasetPage.vue')
    },
    {
      path: '/performance',
      component: () => import('../pages/PerformancePage.vue')
    },
    {
      path: '/teams/:id',
      component: () => import('../pages/TeamViewPage.vue')
    }
  ]
})

// Component-level code splitting for heavy components
export default {
  components: {
    PlayerDataTable: () => import('./PlayerDataTable.vue'),
    PlayerDetailDialog: () => import('./PlayerDetailDialog.vue'),
    ScatterPlotCard: () => import('./ScatterPlotCard.vue'),
    ExportOptionsDialog: () => import('./ExportOptionsDialog.vue')
  }
}
```

#### Dynamic Component Loading

```javascript
// Dynamic component loader utility
export class DynamicComponentLoader {
  constructor() {
    this.loadingComponents = new Map()
    this.loadedComponents = new Map()
  }

  async loadComponent(componentName, importFn) {
    if (this.loadedComponents.has(componentName)) {
      return this.loadedComponents.get(componentName)
    }

    if (this.loadingComponents.has(componentName)) {
      return this.loadingComponents.get(componentName)
    }

    const loadPromise = importFn().then(module => {
      const component = module.default || module
      this.loadedComponents.set(componentName, component)
      this.loadingComponents.delete(componentName)
      return component
    })

    this.loadingComponents.set(componentName, loadPromise)
    return loadPromise
  }

  preloadComponent(componentName, importFn) {
    // Preload without blocking
    requestIdleCallback(() => {
      this.loadComponent(componentName, importFn)
    })
  }
}

// Usage in components
const componentLoader = new DynamicComponentLoader()

export default {
  async created() {
    // Preload likely-to-be-used components
    componentLoader.preloadComponent('PlayerDetailDialog', 
      () => import('./PlayerDetailDialog.vue'))
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
// Optimized Pinia store with advanced caching and memory management
export const usePlayersStore = defineStore('players', () => {
  // State - using shallowRef for large arrays to avoid deep reactivity
  const players = shallowRef([])
  const searchCache = new Map()
  const filterCache = new Map()
  const maxCacheSize = 100
  
  // LRU Cache implementation for filtered results
  class LRUCache {
    constructor(maxSize) {
      this.maxSize = maxSize
      this.cache = new Map()
    }
    
    get(key) {
      if (this.cache.has(key)) {
        const value = this.cache.get(key)
        this.cache.delete(key)
        this.cache.set(key, value) // Move to end
        return value
      }
      return null
    }
    
    set(key, value) {
      if (this.cache.has(key)) {
        this.cache.delete(key)
      } else if (this.cache.size >= this.maxSize) {
        const firstKey = this.cache.keys().next().value
        this.cache.delete(firstKey)
      }
      this.cache.set(key, value)
    }
  }
  
  const lruCache = new LRUCache(maxCacheSize)
  
  // Getters with advanced caching
  const filteredPlayers = computed(() => {
    const cacheKey = JSON.stringify(filters.value)
    
    const cached = lruCache.get(cacheKey)
    if (cached) {
      return cached
    }
    
    const filtered = players.value.filter(player => 
      applyFilters(player, filters.value)
    )
    
    lruCache.set(cacheKey, filtered)
    return filtered
  })
  
  // Batch operations with minimal reactivity triggers
  const updateMultiplePlayers = (updates) => {
    const newPlayers = [...players.value]
    
    updates.forEach(({ id, data }) => {
      const index = newPlayers.findIndex(p => p.id === id)
      if (index !== -1) {
        newPlayers[index] = { ...newPlayers[index], ...data }
      }
    })
    
    // Single reactivity trigger for all updates
    players.value = newPlayers
    
    // Clear related caches
    lruCache.cache.clear()
  }
  
  // Memory cleanup on component unmount
  const cleanup = () => {
    searchCache.clear()
    filterCache.clear()
    lruCache.cache.clear()
  }
  
  return {
    players: readonly(players),
    filteredPlayers,
    updateMultiplePlayers,
    cleanup
  }
})
```

### Memory Management and Object Pooling

#### Object Pool Implementation

```javascript
// Object pool manager for frequently created objects
export class ObjectPoolManager {
  constructor() {
    this.pools = new Map()
    this.stats = new Map()
  }

  createPool(name, factory, resetFn, initialSize = 10) {
    const pool = {
      objects: [],
      factory,
      resetFn,
      created: 0,
      reused: 0
    }

    // Pre-populate pool
    for (let i = 0; i < initialSize; i++) {
      pool.objects.push(factory())
      pool.created++
    }

    this.pools.set(name, pool)
    this.stats.set(name, { created: pool.created, reused: 0 })
  }

  acquire(poolName) {
    const pool = this.pools.get(poolName)
    if (!pool) {
      throw new Error(`Pool ${poolName} not found`)
    }

    let obj
    if (pool.objects.length > 0) {
      obj = pool.objects.pop()
      pool.reused++
      this.stats.get(poolName).reused++
    } else {
      obj = pool.factory()
      pool.created++
      this.stats.get(poolName).created++
    }

    return obj
  }

  release(poolName, obj) {
    const pool = this.pools.get(poolName)
    if (!pool) return

    if (pool.resetFn) {
      pool.resetFn(obj)
    }

    pool.objects.push(obj)
  }

  getStats() {
    return Object.fromEntries(this.stats)
  }
}

// Usage example for player table rows
const objectPool = new ObjectPoolManager()

// Create pool for table row objects
objectPool.createPool('tableRow', 
  () => ({ 
    id: null, 
    name: '', 
    position: '', 
    club: '', 
    overall: 0,
    selected: false 
  }),
  (obj) => {
    obj.id = null
    obj.name = ''
    obj.position = ''
    obj.club = ''
    obj.overall = 0
    obj.selected = false
  },
  50 // Initial pool size
)

// In component
export default {
  setup() {
    const tableRows = ref([])
    
    const createTableRow = (player) => {
      const row = objectPool.acquire('tableRow')
      row.id = player.id
      row.name = player.name
      row.position = player.position
      row.club = player.club
      row.overall = player.overall
      return row
    }
    
    const releaseTableRow = (row) => {
      objectPool.release('tableRow', row)
    }
    
    onUnmounted(() => {
      // Release all objects back to pool
      tableRows.value.forEach(releaseTableRow)
    })
    
    return { createTableRow, releaseTableRow }
  }
}
```

### Image Optimization and Asset Management

#### Advanced Image Loading System

```javascript
// Modern image loader with WebP/AVIF support and progressive loading
export class AdvancedImageLoader {
  constructor() {
    this.cache = new Map()
    this.loadingQueue = []
    this.preloadQueue = []
    this.maxConcurrent = 6
    this.currentLoading = 0
  }

  async loadImage(src, options = {}) {
    const {
      formats = ['avif', 'webp', 'jpg'],
      sizes = [],
      priority = 'normal',
      placeholder = true,
      progressive = true
    } = options

    // Check cache first
    if (this.cache.has(src)) {
      return this.cache.get(src)
    }

    // Create loading promise
    const loadPromise = this.createLoadPromise(src, formats, sizes, {
      placeholder,
      progressive,
      priority
    })

    this.cache.set(src, loadPromise)
    return loadPromise
  }

  async createLoadPromise(src, formats, sizes, options) {
    // Try modern formats first
    for (const format of formats) {
      const modernSrc = this.getModernSrc(src, format)
      if (await this.canLoadFormat(format)) {
        try {
          return await this.loadSingleImage(modernSrc, options)
        } catch (error) {
          console.warn(`Failed to load ${format} format, trying next...`)
        }
      }
    }

    // Fallback to original
    return this.loadSingleImage(src, options)
  }

  async loadSingleImage(src, options) {
    return new Promise((resolve, reject) => {
      const img = new Image()
      
      if (options.progressive) {
        // Load low-quality placeholder first
        const placeholderSrc = this.getPlaceholderSrc(src)
        const placeholder = new Image()
        
        placeholder.onload = () => {
          resolve({
            element: placeholder,
            src: placeholderSrc,
            isPlaceholder: true,
            upgrade: () => this.upgradeImage(img, src)
          })
        }
        
        placeholder.src = placeholderSrc
      }

      img.onload = () => {
        resolve({
          element: img,
          src,
          isPlaceholder: false
        })
      }

      img.onerror = reject
      img.src = src
    })
  }

  getModernSrc(src, format) {
    const ext = src.split('.').pop()
    return src.replace(`.${ext}`, `.${format}`)
  }

  getPlaceholderSrc(src) {
    // Generate low-quality placeholder URL
    return src.replace(/\.(jpg|jpeg|png|webp)$/i, '_placeholder.$1')
  }

  async canLoadFormat(format) {
    if (format === 'avif') {
      return this.supportsAVIF()
    }
    if (format === 'webp') {
      return this.supportsWebP()
    }
    return true
  }

  supportsWebP() {
    const canvas = document.createElement('canvas')
    canvas.width = 1
    canvas.height = 1
    return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0
  }

  supportsAVIF() {
    const canvas = document.createElement('canvas')
    canvas.width = 1
    canvas.height = 1
    return canvas.toDataURL('image/avif').indexOf('data:image/avif') === 0
  }
}

// Lazy loading composable with Intersection Observer
export function useLazyLoading(options = {}) {
  const {
    rootMargin = '50px',
    threshold = 0.1,
    unobserveOnLoad = true
  } = options

  const imageLoader = new AdvancedImageLoader()
  const observedElements = new WeakMap()

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(async (entry) => {
      if (entry.isIntersecting) {
        const element = entry.target
        const src = element.dataset.src
        const options = JSON.parse(element.dataset.options || '{}')

        try {
          const result = await imageLoader.loadImage(src, options)
          
          if (result.isPlaceholder) {
            element.src = result.src
            element.classList.add('placeholder')
            
            // Upgrade to full quality
            const fullImage = await result.upgrade()
            element.src = fullImage.src
            element.classList.remove('placeholder')
            element.classList.add('loaded')
          } else {
            element.src = result.src
            element.classList.add('loaded')
          }

          if (unobserveOnLoad) {
            observer.unobserve(element)
          }
        } catch (error) {
          element.classList.add('error')
          console.error('Failed to load image:', error)
        }
      }
    })
  }, { rootMargin, threshold })

  const observe = (element, src, loadOptions = {}) => {
    element.dataset.src = src
    element.dataset.options = JSON.stringify(loadOptions)
    observer.observe(element)
    observedElements.set(element, { src, options: loadOptions })
  }

  const unobserve = (element) => {
    observer.unobserve(element)
    observedElements.delete(element)
  }

  const cleanup = () => {
    observer.disconnect()
  }

  return {
    observe,
    unobserve,
    cleanup
  }
}
```

### Mobile Performance Optimization

#### Touch Optimization and Battery Efficiency

```javascript
// Mobile-optimized touch handling
export class MobileOptimizer {
  constructor() {
    this.isLowEndDevice = this.detectLowEndDevice()
    this.reducedMotion = this.prefersReducedMotion()
    this.touchHandlers = new Map()
  }

  detectLowEndDevice() {
    // Detect low-end devices based on various factors
    const memory = navigator.deviceMemory || 4
    const cores = navigator.hardwareConcurrency || 4
    const connection = navigator.connection
    
    return memory <= 2 || cores <= 2 || 
           (connection && connection.effectiveType === 'slow-2g')
  }

  prefersReducedMotion() {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }

  optimizeTouchEvents(element, handlers) {
    const optimizedHandlers = {}

    // Use passive listeners for better scroll performance
    if (handlers.touchstart) {
      optimizedHandlers.touchstart = this.debounce(handlers.touchstart, 16)
      element.addEventListener('touchstart', optimizedHandlers.touchstart, { passive: true })
    }

    if (handlers.touchmove) {
      optimizedHandlers.touchmove = this.throttle(handlers.touchmove, 16)
      element.addEventListener('touchmove', optimizedHandlers.touchmove, { passive: true })
    }

    if (handlers.touchend) {
      optimizedHandlers.touchend = handlers.touchend
      element.addEventListener('touchend', optimizedHandlers.touchend, { passive: true })
    }

    this.touchHandlers.set(element, optimizedHandlers)
    return optimizedHandlers
  }

  createMobileVirtualScroll(container, items, itemHeight) {
    const viewportHeight = container.clientHeight
    const visibleCount = Math.ceil(viewportHeight / itemHeight) + 2 // Buffer
    const totalHeight = items.length * itemHeight

    let scrollTop = 0
    let startIndex = 0
    let endIndex = Math.min(visibleCount, items.length)

    const updateVisibleItems = () => {
      startIndex = Math.floor(scrollTop / itemHeight)
      endIndex = Math.min(startIndex + visibleCount, items.length)
      
      // Update DOM efficiently
      this.updateVirtualScrollDOM(container, items.slice(startIndex, endIndex), 
                                  startIndex, itemHeight, totalHeight)
    }

    // Optimized scroll handler for mobile
    const scrollHandler = this.throttle((e) => {
      scrollTop = e.target.scrollTop
      
      // Use requestAnimationFrame for smooth updates
      requestAnimationFrame(updateVisibleItems)
    }, 16)

    container.addEventListener('scroll', scrollHandler, { passive: true })
    
    // Initial render
    updateVisibleItems()

    return {
      updateItems: (newItems) => {
        items = newItems
        updateVisibleItems()
      },
      destroy: () => {
        container.removeEventListener('scroll', scrollHandler)
      }
    }
  }

  throttle(func, limit) {
    let inThrottle
    return function() {
      const args = arguments
      const context = this
      if (!inThrottle) {
        func.apply(context, args)
        inThrottle = true
        setTimeout(() => inThrottle = false, limit)
      }
    }
  }

  debounce(func, wait) {
    let timeout
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout)
        func(...args)
      }
      clearTimeout(timeout)
      timeout = setTimeout(later, wait)
    }
  }
}

// Battery-efficient animation manager
export class BatteryEfficientAnimations {
  constructor() {
    this.animationQueue = []
    this.isAnimating = false
    this.batteryLevel = this.getBatteryLevel()
    this.reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }

  async getBatteryLevel() {
    if ('getBattery' in navigator) {
      const battery = await navigator.getBattery()
      return battery.level
    }
    return 1 // Assume full battery if API not available
  }

  shouldReduceAnimations() {
    return this.reducedMotion || this.batteryLevel < 0.2
  }

  queueAnimation(element, keyframes, options = {}) {
    if (this.shouldReduceAnimations()) {
      // Skip animation, apply final state immediately
      const finalFrame = keyframes[keyframes.length - 1]
      Object.assign(element.style, finalFrame)
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      this.animationQueue.push({
        element,
        keyframes,
        options: {
          duration: 300,
          easing: 'ease-out',
          ...options
        },
        resolve
      })

      this.processQueue()
    })
  }

  processQueue() {
    if (this.isAnimating || this.animationQueue.length === 0) {
      return
    }

    this.isAnimating = true
    const animation = this.animationQueue.shift()

    const animationInstance = animation.element.animate(
      animation.keyframes,
      animation.options
    )

    animationInstance.addEventListener('finish', () => {
      animation.resolve()
      this.isAnimating = false
      this.processQueue() // Process next animation
    })
  }
}
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

### Performance Monitoring and Core Web Vitals

#### Frontend Performance Tracking

```javascript
// Core Web Vitals monitoring implementation
export class PerformanceTracker {
  constructor() {
    this.metrics = new Map()
    this.observers = []
    this.initializeTracking()
  }

  initializeTracking() {
    // Largest Contentful Paint (LCP)
    this.trackLCP()
    
    // First Input Delay (FID)
    this.trackFID()
    
    // Cumulative Layout Shift (CLS)
    this.trackCLS()
    
    // First Contentful Paint (FCP)
    this.trackFCP()
    
    // Time to First Byte (TTFB)
    this.trackTTFB()
  }

  trackLCP() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      const lastEntry = entries[entries.length - 1]
      
      this.recordMetric('LCP', lastEntry.startTime, {
        element: lastEntry.element?.tagName,
        url: lastEntry.url
      })
    })
    
    observer.observe({ entryTypes: ['largest-contentful-paint'] })
    this.observers.push(observer)
  }

  trackFID() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        this.recordMetric('FID', entry.processingStart - entry.startTime, {
          eventType: entry.name
        })
      })
    })
    
    observer.observe({ entryTypes: ['first-input'] })
    this.observers.push(observer)
  }

  trackCLS() {
    let clsValue = 0
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        if (!entry.hadRecentInput) {
          clsValue += entry.value
        }
      })
      
      this.recordMetric('CLS', clsValue)
    })
    
    observer.observe({ entryTypes: ['layout-shift'] })
    this.observers.push(observer)
  }

  trackCustomMetrics() {
    // Table rendering time
    this.measureTableRender = (playerCount) => {
      const start = performance.now()
      return () => {
        const duration = performance.now() - start
        this.recordMetric('table_render_time', duration, { playerCount })
      }
    }

    // Memory usage tracking
    this.trackMemoryUsage = () => {
      if ('memory' in performance) {
        const memory = performance.memory
        this.recordMetric('memory_usage', {
          used: memory.usedJSHeapSize,
          total: memory.totalJSHeapSize,
          limit: memory.jsHeapSizeLimit
        })
      }
    }

    // Bundle loading time
    this.trackBundleLoad = (chunkName) => {
      const start = performance.now()
      return () => {
        const duration = performance.now() - start
        this.recordMetric('bundle_load_time', duration, { chunkName })
      }
    }
  }

  recordMetric(name, value, metadata = {}) {
    const metric = {
      name,
      value,
      timestamp: Date.now(),
      url: window.location.href,
      userAgent: navigator.userAgent,
      ...metadata
    }

    this.metrics.set(`${name}_${Date.now()}`, metric)
    
    // Send to analytics service
    this.sendToAnalytics(metric)
    
    // Check thresholds and alert if needed
    this.checkThresholds(name, value)
  }

  checkThresholds(metricName, value) {
    const thresholds = {
      LCP: 2500, // 2.5 seconds
      FID: 100,  // 100ms
      CLS: 0.1,  // 0.1
      FCP: 1800, // 1.8 seconds
      TTFB: 800, // 800ms
      table_render_time: 500, // 500ms
      bundle_load_time: 3000  // 3 seconds
    }

    if (thresholds[metricName] && value > thresholds[metricName]) {
      console.warn(`Performance threshold exceeded for ${metricName}: ${value}`)
      
      // Send alert to monitoring service
      this.sendAlert(metricName, value, thresholds[metricName])
    }
  }

  async sendToAnalytics(metric) {
    try {
      await fetch('/api/analytics/performance', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(metric)
      })
    } catch (error) {
      console.error('Failed to send performance metric:', error)
    }
  }

  getMetricsSummary() {
    const summary = {}
    
    for (const [key, metric] of this.metrics) {
      if (!summary[metric.name]) {
        summary[metric.name] = {
          count: 0,
          total: 0,
          min: Infinity,
          max: -Infinity,
          values: []
        }
      }
      
      const stat = summary[metric.name]
      stat.count++
      stat.total += metric.value
      stat.min = Math.min(stat.min, metric.value)
      stat.max = Math.max(stat.max, metric.value)
      stat.values.push(metric.value)
    }
    
    // Calculate percentiles
    for (const stat of Object.values(summary)) {
      stat.average = stat.total / stat.count
      stat.values.sort((a, b) => a - b)
      stat.p50 = this.percentile(stat.values, 50)
      stat.p75 = this.percentile(stat.values, 75)
      stat.p95 = this.percentile(stat.values, 95)
    }
    
    return summary
  }

  percentile(values, p) {
    const index = Math.ceil((p / 100) * values.length) - 1
    return values[index]
  }

  cleanup() {
    this.observers.forEach(observer => observer.disconnect())
    this.metrics.clear()
  }
}

// Usage in main application
const performanceTracker = new PerformanceTracker()

// Track custom events
export const trackTableRender = performanceTracker.measureTableRender
export const trackBundleLoad = performanceTracker.trackBundleLoad
export const trackMemoryUsage = performanceTracker.trackMemoryUsage
```

#### Bundle Analysis and Monitoring

```javascript
// Bundle analysis utilities
export class BundleAnalyzer {
  constructor() {
    this.chunkSizes = new Map()
    this.loadTimes = new Map()
    this.dependencies = new Map()
  }

  trackChunkLoad(chunkName, size) {
    const start = performance.now()
    
    return () => {
      const loadTime = performance.now() - start
      this.chunkSizes.set(chunkName, size)
      this.loadTimes.set(chunkName, loadTime)
      
      // Alert if chunk is too large
      if (size > 800 * 1024) { // 800KB threshold
        console.warn(`Large chunk detected: ${chunkName} (${(size / 1024).toFixed(1)}KB)`)
      }
      
      // Alert if load time is too slow
      if (loadTime > 3000) { // 3 second threshold
        console.warn(`Slow chunk load: ${chunkName} (${loadTime.toFixed(0)}ms)`)
      }
    }
  }

  analyzeBundleComposition() {
    const analysis = {
      totalSize: 0,
      chunkCount: this.chunkSizes.size,
      largestChunk: { name: '', size: 0 },
      slowestChunk: { name: '', time: 0 },
      recommendations: []
    }

    // Analyze chunk sizes
    for (const [name, size] of this.chunkSizes) {
      analysis.totalSize += size
      
      if (size > analysis.largestChunk.size) {
        analysis.largestChunk = { name, size }
      }
    }

    // Analyze load times
    for (const [name, time] of this.loadTimes) {
      if (time > analysis.slowestChunk.time) {
        analysis.slowestChunk = { name, time }
      }
    }

    // Generate recommendations
    if (analysis.largestChunk.size > 1024 * 1024) { // 1MB
      analysis.recommendations.push(
        `Consider splitting ${analysis.largestChunk.name} chunk (${(analysis.largestChunk.size / 1024 / 1024).toFixed(1)}MB)`
      )
    }

    if (analysis.slowestChunk.time > 5000) { // 5 seconds
      analysis.recommendations.push(
        `Optimize loading for ${analysis.slowestChunk.name} chunk (${(analysis.slowestChunk.time / 1000).toFixed(1)}s)`
      )
    }

    return analysis
  }

  generateReport() {
    const analysis = this.analyzeBundleComposition()
    
    return {
      timestamp: new Date().toISOString(),
      bundleAnalysis: analysis,
      chunkDetails: Array.from(this.chunkSizes.entries()).map(([name, size]) => ({
        name,
        size,
        sizeFormatted: `${(size / 1024).toFixed(1)}KB`,
        loadTime: this.loadTimes.get(name) || 0
      })).sort((a, b) => b.size - a.size)
    }
  }
}
```

## Performance Best Practices

### Frontend Performance Best Practices

1. **Bundle Optimization**
   - Keep initial bundle under 200KB gzipped
   - Use route-based code splitting for major features
   - Implement dynamic imports for heavy components
   - Monitor bundle size in CI/CD pipeline

2. **Memory Management**
   - Use `shallowRef` for large arrays to avoid deep reactivity
   - Implement object pooling for frequently created objects
   - Clear caches and event listeners on component unmount
   - Monitor memory usage and detect leaks early

3. **Image and Asset Optimization**
   - Use modern image formats (WebP/AVIF) with fallbacks
   - Implement lazy loading for off-screen images
   - Optimize image sizes and use responsive images
   - Cache images aggressively with proper invalidation

4. **Mobile Performance**
   - Use passive event listeners for touch events
   - Implement reduced motion preferences
   - Optimize for low-end devices and slow networks
   - Monitor battery usage and adapt accordingly

5. **Performance Monitoring**
   - Track Core Web Vitals continuously
   - Monitor custom metrics relevant to your application
   - Set up performance budgets and alerts
   - Analyze performance regressions in CI/CD

### Backend Performance Best Practices

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

## Performance Testing and Validation

### Automated Performance Testing

```bash
# Lighthouse CI for automated performance testing
npm install -g @lhci/cli

# Run Lighthouse CI
lhci autorun --config=.lighthouserc.js

# Performance budget configuration
# .lighthouserc.js
module.exports = {
  ci: {
    collect: {
      url: ['http://localhost:3000'],
      numberOfRuns: 3
    },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'first-contentful-paint': ['error', { maxNumericValue: 2000 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }]
      }
    }
  }
}
```

### Load Testing for Frontend

```javascript
// Frontend load testing with realistic user scenarios
import { test, expect } from '@playwright/test'

test.describe('Performance Tests', () => {
  test('Large dataset rendering performance', async ({ page }) => {
    await page.goto('/dataset/large-test-dataset')
    
    // Measure table rendering time
    const startTime = Date.now()
    await page.waitForSelector('[data-testid="player-table"]')
    const renderTime = Date.now() - startTime
    
    expect(renderTime).toBeLessThan(2000) // 2 second threshold
    
    // Check for memory leaks
    const initialMemory = await page.evaluate(() => performance.memory?.usedJSHeapSize || 0)
    
    // Simulate user interactions
    await page.click('[data-testid="filter-position"]')
    await page.selectOption('[data-testid="position-select"]', 'ST')
    await page.waitForTimeout(1000)
    
    const finalMemory = await page.evaluate(() => performance.memory?.usedJSHeapSize || 0)
    const memoryIncrease = finalMemory - initialMemory
    
    // Memory increase should be reasonable
    expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024) // 50MB threshold
  })

  test('Mobile performance', async ({ page, browserName }) => {
    // Simulate mobile device
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/dataset/test-dataset')
    
    // Measure touch interaction responsiveness
    const startTime = Date.now()
    await page.tap('[data-testid="player-row-1"]')
    await page.waitForSelector('[data-testid="player-detail-dialog"]')
    const interactionTime = Date.now() - startTime
    
    expect(interactionTime).toBeLessThan(300) // 300ms threshold for mobile
  })
})
```

---

Regular performance monitoring and optimization ensure FM-Dash remains fast and responsive as datasets grow and user load increases. The comprehensive performance optimization strategy covers both frontend and backend concerns, with particular emphasis on the new frontend performance requirements including bundle optimization, memory management, image loading, mobile performance, and continuous monitoring. 