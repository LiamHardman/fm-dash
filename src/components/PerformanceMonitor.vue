<template>
    <q-card 
        v-if="showMonitor" 
        class="performance-monitor fixed-bottom-right q-ma-md"
        style="width: 280px; z-index: 9999;"
        :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
    >
        <q-card-section class="q-pa-sm">
            <div class="row items-center q-mb-xs">
                <q-icon name="speed" size="sm" class="q-mr-xs" />
                <span class="text-caption text-weight-bold">Performance Monitor</span>
                <q-space />
                <q-btn 
                    flat 
                    round 
                    dense 
                    size="xs" 
                    icon="close" 
                    @click="showMonitor = false"
                />
            </div>
            
            <!-- Frame Rate -->
            <div class="performance-metric">
                <div class="row items-center justify-between">
                    <span class="text-caption">FPS:</span>
                    <span 
                        class="text-caption text-weight-bold"
                        :class="fps >= 50 ? 'text-positive' : fps >= 30 ? 'text-warning' : 'text-negative'"
                    >
                        {{ Math.round(fps) }}
                    </span>
                </div>
                <q-linear-progress 
                    :value="fps / 60" 
                    color="positive" 
                    size="4px" 
                    class="q-mt-xs"
                />
            </div>

            <!-- Memory Usage -->
            <div class="performance-metric">
                <div class="row items-center justify-between">
                    <span class="text-caption">Memory:</span>
                    <span class="text-caption text-weight-bold">
                        {{ Math.round(memoryUsage.used / 1024 / 1024) }}MB
                    </span>
                </div>
                <q-linear-progress 
                    :value="memoryUsage.used / memoryUsage.limit" 
                    :color="memoryUsage.used / memoryUsage.limit > 0.8 ? 'negative' : 'info'" 
                    size="4px" 
                    class="q-mt-xs"
                />
            </div>

            <!-- Worker Status -->
            <div class="performance-metric">
                <div class="row items-center justify-between">
                    <span class="text-caption">Workers:</span>
                    <span class="text-caption text-weight-bold">
                        {{ workerStatus.active }}/{{ workerStatus.total }}
                    </span>
                </div>
                <div class="text-caption text-grey-6">
                    {{ workerStatus.pendingTasks }} pending tasks
                </div>
            </div>

            <!-- Cache Statistics -->
            <div class="performance-metric">
                <div class="row items-center justify-between">
                    <span class="text-caption">Cache Hit Rate:</span>
                    <span class="text-caption text-weight-bold text-positive">
                        {{ Math.round(cacheStats.hitRate * 100) }}%
                    </span>
                </div>
                <div class="text-caption text-grey-6">
                    {{ cacheStats.hits }}/{{ cacheStats.total }} hits
                </div>
            </div>

            <!-- Rendering Performance -->
            <div class="performance-metric">
                <div class="row items-center justify-between">
                    <span class="text-caption">Render Time:</span>
                    <span class="text-caption text-weight-bold">
                        {{ Math.round(renderStats.averageTime) }}ms
                    </span>
                </div>
                <div class="text-caption text-grey-6">
                    {{ renderStats.operations }} operations
                </div>
            </div>

            <!-- Toggle Debug Mode -->
            <div class="q-mt-sm">
                <q-btn 
                    flat 
                    dense 
                    size="xs" 
                    :label="debugMode ? 'Hide Debug' : 'Show Debug'"
                    @click="debugMode = !debugMode"
                />
            </div>

            <!-- Debug Information -->
            <div v-if="debugMode" class="q-mt-sm">
                <div class="text-caption">
                    <div>Virtual Items: {{ virtualScrollStats.visibleItems }}</div>
                    <div>Scroll Position: {{ Math.round(virtualScrollStats.scrollTop) }}px</div>
                    <div>Memoization Savings: {{ Math.round(memoizationSavings) }}%</div>
                </div>
            </div>
        </q-card-section>
    </q-card>

    <!-- Performance Toggle Button -->
    <q-btn
        v-if="!showMonitor"
        fab-mini
        icon="speed"
        color="primary"
        class="fixed-bottom-right q-ma-md"
        style="z-index: 9998;"
        @click="showMonitor = true"
    />
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, onUnmounted, ref } from 'vue'

export default {
  name: 'PerformanceMonitor',
  setup() {
    const qInstance = useQuasar()

    // State
    const showMonitor = ref(false)
    const debugMode = ref(false)
    const fps = ref(0)
    const memoryUsage = ref({ used: 0, limit: 0 })
    const workerStatus = ref({ active: 0, total: 0, pendingTasks: 0 })
    const cacheStats = ref({ hits: 0, total: 0, hitRate: 0 })
    const renderStats = ref({ averageTime: 0, operations: 0 })
    const virtualScrollStats = ref({ visibleItems: 0, scrollTop: 0 })
    const memoizationSavings = ref(0)

    // Performance monitoring
    let frameCount = 0
    let lastTime = performance.now()
    let animationFrameId = null
    let performanceObserver = null

    // FPS calculation
    const calculateFPS = () => {
      frameCount++
      const currentTime = performance.now()

      if (currentTime >= lastTime + 1000) {
        fps.value = (frameCount * 1000) / (currentTime - lastTime)
        frameCount = 0
        lastTime = currentTime
      }

      if (showMonitor.value) {
        animationFrameId = requestAnimationFrame(calculateFPS)
      }
    }

    // Memory monitoring
    const updateMemoryUsage = () => {
      if (performance.memory) {
        memoryUsage.value = {
          used: performance.memory.usedJSHeapSize,
          limit: performance.memory.jsHeapSizeLimit
        }
      }
    }

    // Worker status monitoring
    const updateWorkerStatus = () => {
      // This would be connected to your worker management system
      // For now, we'll simulate some data
      workerStatus.value = {
        active: 1, // Number of active workers
        total: 1, // Total workers created
        pendingTasks: 0 // Tasks waiting to be processed
      }
    }

    // Cache statistics
    const updateCacheStats = () => {
      // This would be connected to your memoization cache system
      // For demonstration, we'll use simulated data
      const totalOperations = 1000
      const cacheHits = 850

      cacheStats.value = {
        hits: cacheHits,
        total: totalOperations,
        hitRate: cacheHits / totalOperations
      }
    }

    // Render performance monitoring
    const updateRenderStats = () => {
      // Monitor render performance using Performance Observer
      if (window.PerformanceObserver) {
        if (performanceObserver) {
          performanceObserver.disconnect()
        }

        performanceObserver = new PerformanceObserver(list => {
          const entries = list.getEntries()
          const renderTimes = entries
            .filter(entry => entry.entryType === 'measure')
            .map(entry => entry.duration)

          if (renderTimes.length > 0) {
            const avgTime = renderTimes.reduce((a, b) => a + b, 0) / renderTimes.length
            renderStats.value = {
              averageTime: avgTime,
              operations: renderTimes.length
            }
          }
        })

        performanceObserver.observe({ entryTypes: ['measure', 'navigation', 'paint'] })
      }
    }

    // Virtual scroll monitoring
    const updateVirtualScrollStats = () => {
      // This would be connected to your virtual scroll implementation
      virtualScrollStats.value = {
        visibleItems: 20, // Number of visible items
        scrollTop: 0 // Current scroll position
      }
    }

    // Memoization savings calculation
    const updateMemoizationSavings = () => {
      // Calculate the percentage of operations saved by memoization
      const totalOperations = 1000
      const memoizedOperations = 750
      memoizationSavings.value = (memoizedOperations / totalOperations) * 100
    }

    // Periodic updates
    let updateInterval = null

    const startMonitoring = () => {
      if (animationFrameId) return

      calculateFPS()
      updateInterval = setInterval(() => {
        updateMemoryUsage()
        updateWorkerStatus()
        updateCacheStats()
        updateRenderStats()
        updateVirtualScrollStats()
        updateMemoizationSavings()
      }, 1000)
    }

    const stopMonitoring = () => {
      if (animationFrameId) {
        cancelAnimationFrame(animationFrameId)
        animationFrameId = null
      }

      if (updateInterval) {
        clearInterval(updateInterval)
        updateInterval = null
      }

      if (performanceObserver) {
        performanceObserver.disconnect()
        performanceObserver = null
      }
    }

    // Watch for monitor visibility
    const startMonitoringWatcher = computed(() => showMonitor.value)
    const _unwatchMonitor = () => {
      if (startMonitoringWatcher.value) {
        startMonitoring()
      } else {
        stopMonitoring()
      }
    }

    onMounted(() => {
      // Enable performance monitoring in development
      if (process.env.NODE_ENV === 'development') {
        showMonitor.value = false // Start hidden
      }
    })

    onUnmounted(() => {
      stopMonitoring()
    })

    // Watch for show/hide changes
    const _unwatchShowMonitor = computed(() => {
      if (showMonitor.value) {
        startMonitoring()
      } else {
        stopMonitoring()
      }
    })

    return {
      qInstance,
      showMonitor,
      debugMode,
      fps,
      memoryUsage,
      workerStatus,
      cacheStats,
      renderStats,
      virtualScrollStats,
      memoizationSavings
    }
  }
}
</script>

<style scoped>
.performance-monitor {
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
}

.performance-metric {
    margin-bottom: 8px;
}

.performance-metric:last-child {
    margin-bottom: 0;
}

.fixed-bottom-right {
    position: fixed;
    bottom: 0;
    right: 0;
}
</style>