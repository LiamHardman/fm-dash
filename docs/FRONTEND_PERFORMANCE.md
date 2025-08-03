# Frontend Performance Guide

This document outlines the performance optimizations, monitoring, and best practices for the FM-Dash frontend application built with Vue 3, Quasar Framework, and Vite.

## Performance Overview

FM-Dash is optimized to handle large Football Manager datasets (50MB+ files, 10,000+ players) with smooth, responsive user interactions. Our performance strategy focuses on:

- **Bundle Optimization**: Advanced code splitting and tree shaking
- **Memory Management**: Efficient handling of large datasets
- **Virtual Scrolling**: Smooth navigation through thousands of players
- **Progressive Loading**: Optimized image and data loading
- **Caching Strategies**: Smart data and resource caching

## Architecture Performance Features

### Modern Build System

```javascript
// Vite Configuration Highlights
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-core': ['vue', 'vue-router', 'pinia'],
          'ui-framework': ['quasar'],
          'charts': ['chart.js', 'vue-chartjs'],
          'utils': ['@vueuse/core']
        }
      }
    },
    chunkSizeWarningLimit: 500,
    target: ['es2020', 'chrome80', 'firefox78', 'safari14']
  }
})
```

### Advanced State Management

**Memory-Optimized Pinia Stores**:
```javascript
// Using shallowRef for large arrays to avoid deep reactivity overhead
const players = shallowRef([])

// LRU Cache for filtered results
const filterCache = new LRUCache(100)

// Computed properties with intelligent caching
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
```

## Performance Optimizations

### 1. Bundle Optimization

**Code Splitting Strategy**:
- **Core Framework**: Vue, Router, Pinia in separate chunk
- **UI Components**: Quasar components tree-shaken
- **Feature-Based**: Charts, analytics in dedicated chunks
- **Page-Based**: Route-level code splitting for better caching

**Tree Shaking**:
```javascript
// Optimized imports - only load what's needed
import { QTable, QBtn, QInput } from 'quasar'
import { debounce } from '@vueuse/core'
```

**Bundle Analysis**:
```bash
# Generate bundle analysis
npm run build:analyze

# View detailed bundle composition
npm run analyze:bundle
```

### 2. Virtual Scrolling

**Implementation for Large Player Lists**:
```vue
<template>
  <q-virtual-scroll
    :items="players"
    :item-size="60"
    v-slot="{ item, index }"
    style="max-height: 400px;"
  >
    <PlayerRow :player="item" :key="index" />
  </q-virtual-scroll>
</template>
```

**Benefits**:
- Renders only visible items (~20-30 players at once)
- Handles 10,000+ players without performance degradation
- Smooth scrolling with momentum preservation
- Memory usage remains constant regardless of list size

### 3. Progressive Image Loading

**Smart Image Loading System**:
```javascript
// Priority-based image loading
class ImageLoadingSystem {
  loadImage(src, options = {}) {
    const {
      priority = 'normal',
      formats = ['avif', 'webp', 'jpg'],
      progressive = true,
      lazy = true
    } = options

    return this.processImageQueue(src, options)
  }
}
```

**Features**:
- **Format Detection**: AVIF → WebP → JPEG fallback
- **Lazy Loading**: Images load as they enter viewport
- **Priority Queue**: Important images load first
- **Progressive Enhancement**: Low-quality placeholder → high-quality image

### 4. Memory Management

**Object Pooling for Large Datasets**:
```javascript
class MemoryManager {
  createObjectPool(name, factory, resetFn, initialSize = 10) {
    const pool = new ObjectPool(factory, resetFn, initialSize)
    this.objectPools.set(name, pool)
    return pool
  }

  performCleanup() {
    this.caches.forEach(cache => cache.clear())
    this.objectPools.forEach(pool => pool.releaseAll())
    
    if (window.gc) window.gc() // Force GC if available
  }
}
```

**Memory Monitoring**:
```javascript
// Automatic memory monitoring
setInterval(() => {
  if (performance.memory?.usedJSHeapSize > MEMORY_THRESHOLD) {
    memoryManager.performCleanup()
  }
}, 30000)
```

### 5. Web Workers

**Background Processing for Heavy Calculations**:
```javascript
// Player calculation worker
self.onmessage = function(e) {
  const { players, calculations } = e.data
  
  const results = players.map(player => {
    return {
      ...player,
      fifaRating: calculateFifaRating(player),
      percentiles: calculatePercentiles(player),
      valueRatio: calculateValueRatio(player)
    }
  })
  
  self.postMessage(results)
}
```

**Worker Usage**:
```javascript
// Offload heavy calculations to worker thread
const worker = new Worker('/workers/playerCalculationWorker.js')

worker.postMessage({ players: rawPlayers, calculations: ['fifa', 'percentiles'] })

worker.onmessage = (e) => {
  players.value = e.data // Update UI with calculated data
}
```

## Performance Monitoring

### Core Web Vitals Tracking

**Largest Contentful Paint (LCP)**:
- Target: < 2.5 seconds
- Optimization: Critical resource prioritization, image optimization

**First Input Delay (FID)**:
- Target: < 100 milliseconds  
- Optimization: Web workers, event handler optimization

**Cumulative Layout Shift (CLS)**:
- Target: < 0.1
- Optimization: Reserved space for dynamic content

**Implementation**:
```javascript
// Core Web Vitals monitoring
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals'

getCLS(console.log)
getFID(console.log)
getFCP(console.log)
getLCP(console.log)
getTTFB(console.log)
```

### Performance Profiling

**Custom Performance Marks**:
```javascript
// Mark critical performance points
performance.mark('player-data-start')
await loadPlayerData()
performance.mark('player-data-end')

performance.measure('player-data-load', 'player-data-start', 'player-data-end')
```

**Memory Usage Tracking**:
```javascript
// Monitor memory usage
const observer = new PerformanceObserver((list) => {
  list.getEntries().forEach((entry) => {
    if (entry.entryType === 'measure') {
      console.log(`${entry.name}: ${entry.duration}ms`)
    }
  })
})

observer.observe({ entryTypes: ['measure'] })
```

## Configuration Options

### Environment Variables

```bash
# Bundle Optimization
VITE_CODE_SPLITTING=true
VITE_TREE_SHAKING=true
VITE_CHUNK_SIZE_WARNING=500

# Memory Management
VITE_VIRTUAL_SCROLL_BUFFER=5
VITE_OBJECT_POOL_SIZE=50
VITE_LRU_CACHE_SIZE=100
VITE_MEMORY_THRESHOLD=200

# Image Optimization
VITE_LAZY_LOADING=true
VITE_WEBP_SUPPORT=true
VITE_AVIF_SUPPORT=true
VITE_IMAGE_PRELOAD_COUNT=5

# Performance Monitoring
VITE_PERFORMANCE_TRACKING=true
VITE_CORE_WEB_VITALS=true
VITE_MEMORY_MONITORING=true
```

### Runtime Configuration

```javascript
// Quasar performance settings
const quasarOptions = {
  config: {
    loading: {
      delay: 0,
      message: 'Loading players...',
      spinnerSize: 80
    }
  },
  plugins: [
    'Loading',
    'LoadingBar',
    'Notify'
  ]
}
```

## Performance Best Practices

### Component Optimization

**Efficient Component Design**:
```vue
<script setup>
// Use shallowRef for large, static data
const players = shallowRef([])

// Memoize expensive computations
const expensiveCalculation = computed(() => {
  return useMemoize(() => {
    return heavyCalculation(players.value)
  }, [players.value.length])
})

// Debounce user input
const debouncedSearch = useDebounceFn((term) => {
  searchPlayers(term)
}, 300)
</script>

<template>
  <!-- Use v-memo for lists with complex items -->
  <div v-for="player in players" :key="player.id" v-memo="[player.overall, player.potential]">
    <PlayerCard :player="player" />
  </div>
</template>
```

**List Optimization**:
```vue
<!-- Prefer virtual scrolling for large lists -->
<q-virtual-scroll
  :items="filteredPlayers"
  :item-size="60"
  v-slot="{ item, index }"
>
  <PlayerRow :player="item" :key="item.id" />
</q-virtual-scroll>

<!-- Use pagination for non-virtual lists -->
<q-pagination
  v-model="currentPage"
  :max="totalPages"
  :max-pages="6"
  boundary-numbers
/>
```

### Data Loading Strategies

**Progressive Data Loading**:
```javascript
// Load essential data first
const essentialData = await loadEssentialPlayerData()
players.value = essentialData

// Load additional data in background
nextTick(() => {
  loadAdditionalPlayerData().then(additionalData => {
    players.value = mergePlayerData(essentialData, additionalData)
  })
})
```

**Intelligent Caching**:
```javascript
// Cache frequently accessed data
const playerCache = new Map()

function getPlayerDetails(playerId) {
  if (playerCache.has(playerId)) {
    return playerCache.get(playerId)
  }
  
  const details = fetchPlayerDetails(playerId)
  playerCache.set(playerId, details)
  return details
}
```

### Network Optimization

**Request Optimization**:
```javascript
// Batch multiple requests
const batchedRequests = [
  fetchPlayers(),
  fetchTeams(),
  fetchLeagues()
]

const [players, teams, leagues] = await Promise.all(batchedRequests)
```

**Resource Hints**:
```html
<!-- Preload critical resources -->
<link rel="preload" href="/api/players" as="fetch" crossorigin>

<!-- Prefetch likely next resources -->
<link rel="prefetch" href="/api/teams">

<!-- Preconnect to external domains -->
<link rel="preconnect" href="https://cdn.example.com">
```

## Performance Testing

### Automated Testing

**Bundle Size Testing**:
```javascript
// Bundle size assertions
describe('Bundle Size', () => {
  it('should not exceed size limits', () => {
    const stats = getBundleStats()
    expect(stats.chunks.vendor.size).toBeLessThan(500 * 1024) // 500KB
    expect(stats.chunks.app.size).toBeLessThan(300 * 1024)    // 300KB
  })
})
```

**Performance Testing**:
```javascript
// Performance regression tests
describe('Performance', () => {
  it('should load player list within time limit', async () => {
    const start = performance.now()
    await loadPlayerList(1000) // Load 1000 players
    const duration = performance.now() - start
    
    expect(duration).toBeLessThan(2000) // Under 2 seconds
  })
})
```

### Manual Testing

**Performance Checklist**:
- [ ] Initial page load under 3 seconds
- [ ] Player list scrolling is smooth (60fps)
- [ ] Search results appear within 500ms
- [ ] Memory usage stable over extended use
- [ ] No layout shifts during loading
- [ ] Images load progressively without blocking

**Load Testing Scenarios**:
1. **Small Dataset**: 100-500 players
2. **Medium Dataset**: 1,000-5,000 players  
3. **Large Dataset**: 10,000+ players
4. **Extreme Dataset**: 50,000+ players (stress test)

## Troubleshooting Performance Issues

### Common Issues

**Slow Initial Load**:
```bash
# Check bundle sizes
npm run build:analyze

# Optimize critical path
npm run lighthouse

# Review network timing
# Open DevTools → Network → Reload page
```

**Memory Leaks**:
```javascript
// Monitor memory usage
const observer = new PerformanceObserver((list) => {
  list.getEntries().forEach((entry) => {
    if (entry.name === 'usedJSHeapSize') {
      console.log('Memory usage:', entry.size)
    }
  })
})
```

**Scroll Performance**:
```javascript
// Debug scroll performance
let scrollTimeout
window.addEventListener('scroll', () => {
  if (scrollTimeout) return
  
  scrollTimeout = setTimeout(() => {
    console.log('Scroll performance check')
    scrollTimeout = null
  }, 16) // 60fps = 16ms
})
```

### Performance Debugging Tools

**Browser DevTools**:
- **Performance Tab**: Record and analyze runtime performance
- **Memory Tab**: Track memory usage and detect leaks
- **Network Tab**: Analyze loading times and resource sizes
- **Lighthouse**: Automated performance auditing

**Vue DevTools**:
- **Performance**: Component render timing
- **Timeline**: Reactive data changes
- **Components**: Memory usage by component

**Custom Performance Dashboard**:
```vue
<template>
  <div v-if="showPerformanceStats" class="performance-overlay">
    <div>FPS: {{ fps }}</div>
    <div>Memory: {{ memoryUsage }}MB</div>
    <div>Players: {{ playerCount }}</div>
    <div>Render Time: {{ renderTime }}ms</div>
  </div>
</template>
```

## Mobile Performance

### Mobile-Specific Optimizations

**Touch Optimization**:
```css
/* Optimize touch scrolling */
.player-list {
  -webkit-overflow-scrolling: touch;
  overscroll-behavior: contain;
}

/* Reduce touch delay */
.interactive-element {
  touch-action: manipulation;
}
```

**Responsive Performance**:
```javascript
// Adjust performance based on device capabilities
const deviceCapabilities = {
  isLowEnd: navigator.hardwareConcurrency <= 2,
  hasLowMemory: navigator.deviceMemory <= 4,
  isSlowConnection: navigator.connection?.effectiveType === '2g'
}

if (deviceCapabilities.isLowEnd) {
  // Reduce visual effects
  disableAnimations()
  reduceWorkerCount()
}
```

**Battery Optimization**:
```javascript
// Optimize for battery life
navigator.getBattery?.().then(battery => {
  if (battery.level < 0.2) {
    // Enable power saving mode
    reduceFPS()
    disableNonEssentialFeatures()
  }
})
```

## Performance Metrics

### Key Performance Indicators

| Metric | Target | Current | Notes |
|--------|--------|---------|-------|
| Initial Bundle Size | < 500KB | ~450KB | Main application bundle |
| Time to Interactive | < 3s | ~2.1s | On 3G connection |
| Player List Render | < 100ms | ~85ms | 1000 players |
| Search Response | < 200ms | ~150ms | Fuzzy search |
| Memory Usage (10k players) | < 200MB | ~180MB | Stable over time |

### Performance Budget

```javascript
// Performance budget configuration
const performanceBudget = {
  bundles: {
    main: 500 * 1024,      // 500KB
    vendor: 800 * 1024,    // 800KB
    chunks: 200 * 1024     // 200KB per chunk
  },
  assets: {
    images: 2 * 1024 * 1024, // 2MB total
    fonts: 100 * 1024        // 100KB
  },
  timing: {
    fcp: 1500,    // First Contentful Paint
    lcp: 2500,    // Largest Contentful Paint
    fid: 100,     // First Input Delay
    cls: 0.1      // Cumulative Layout Shift
  }
}
```

---

**Performance is a feature!** 🚀

For implementation details, see [Technical Details](TECHNICAL_DETAILS.md) and [Configuration Guide](CONFIGURATION.md).
