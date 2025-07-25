<template>
  <div v-if="showMonitor" class="performance-monitor-overlay">
    <q-card class="performance-monitor-card">
      <q-card-section class="q-pa-sm">
        <div class="row items-center q-mb-xs">
          <q-icon name="speed" size="sm" class="q-mr-xs" />
          <span class="text-caption text-weight-bold">Player Detail Performance</span>
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
        
        <!-- Cache Performance -->
        <div class="performance-metric">
          <div class="row items-center justify-between">
            <span class="text-caption">Cache Hit Rate:</span>
            <span class="text-caption text-weight-bold text-positive">
              {{ Math.round(metrics.cacheHitRate * 100) }}%
            </span>
          </div>
          <div class="text-caption text-grey-6">
            {{ metrics.cacheHits }}/{{ metrics.cacheHits + metrics.cacheMisses }} hits
          </div>
        </div>

        <!-- Average Load Time -->
        <div class="performance-metric">
          <div class="row items-center justify-between">
            <span class="text-caption">Avg Load Time:</span>
            <span class="text-caption text-weight-bold">
              {{ Math.round(metrics.averageLoadTime) }}ms
            </span>
          </div>
        </div>

        <!-- Total Requests -->
        <div class="performance-metric">
          <div class="row items-center justify-between">
            <span class="text-caption">Total Requests:</span>
            <span class="text-caption text-weight-bold">
              {{ metrics.totalRequests }}
            </span>
          </div>
        </div>

        <!-- Current Dialog State -->
        <div class="performance-metric">
          <div class="row items-center justify-between">
            <span class="text-caption">Dialog State:</span>
            <span class="text-caption text-weight-bold">
              {{ dialogState }}
            </span>
          </div>
        </div>

        <!-- Performance Actions -->
        <div class="q-mt-sm">
          <q-btn 
            flat 
            dense 
            size="xs" 
            label="Clear Cache" 
            @click="clearCache"
            color="warning"
          />
          <q-btn 
            flat 
            dense 
            size="xs" 
            label="Refresh" 
            @click="refreshMetrics"
            class="q-ml-xs"
            color="primary"
          />
        </div>
      </q-card-section>
    </q-card>
  </div>

  <!-- Performance Toggle Button -->
  <q-btn
    v-if="!showMonitor && isDevelopment"
    fab-mini
    icon="speed"
    color="primary"
    class="fixed-bottom-left q-ma-md"
    style="z-index: 9998;"
    @click="showMonitor = true"
  />
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getPerformanceMetrics, clearAllCaches } from '../utils/playerDetailOptimizer.js'

export default {
  name: 'PlayerDetailPerformanceMonitor',
  props: {
    isLoading: {
      type: Boolean,
      default: false
    },
    hasError: {
      type: Boolean,
      default: false
    },
    playerName: {
      type: String,
      default: ''
    }
  },
  setup(props) {
    const showMonitor = ref(false)
    const isDevelopment = import.meta.env.DEV
    const metrics = ref({
      cacheHits: 0,
      cacheMisses: 0,
      cacheHitRate: 0,
      averageLoadTime: 0,
      totalRequests: 0
    })

    // Update metrics periodically
    let metricsInterval = null

    const updateMetrics = () => {
      const currentMetrics = getPerformanceMetrics()
      metrics.value = currentMetrics
    }

    const refreshMetrics = () => {
      updateMetrics()
    }

    const clearCache = () => {
      clearAllCaches()
      updateMetrics()
    }

    const dialogState = computed(() => {
      if (props.hasError) return 'Error'
      if (props.isLoading) return 'Loading'
      if (props.playerName) return 'Loaded'
      return 'Idle'
    })

    onMounted(() => {
      updateMetrics()
      metricsInterval = setInterval(updateMetrics, 2000) // Update every 2 seconds
    })

    onUnmounted(() => {
      if (metricsInterval) {
        clearInterval(metricsInterval)
      }
    })

    return {
      showMonitor,
      metrics,
      dialogState,
      refreshMetrics,
      clearCache,
      isDevelopment
    }
  }
}
</script>

<style lang="scss" scoped>
.performance-monitor-overlay {
  position: fixed;
  top: 20px;
  left: 20px;
  z-index: 9999;
}

.performance-monitor-card {
  width: 280px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(0, 0, 0, 0.1);
  
  .body--dark & {
    background: rgba(30, 41, 59, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
}

.performance-metric {
  margin-bottom: 8px;
  
  &:last-child {
    margin-bottom: 0;
  }
}

.fixed-bottom-left {
  position: fixed;
  bottom: 0;
  left: 0;
}
</style> 