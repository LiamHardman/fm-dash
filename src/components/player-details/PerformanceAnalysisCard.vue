<template>
  <q-card
    flat
    bordered
    class="performance-percentiles-card modern-stats-card"
  >
    <q-card-section class="performance-card-header">
      <div class="performance-header-title">
        <q-icon name="analytics" class="q-mr-sm" />
        Performance Analysis
      </div>
    </q-card-section>
    
    <q-card-section class="q-pa-md">
      <!-- Comparison Controls -->
      <div class="row q-col-gutter-sm q-mb-md">
        <div class="col-6">
          <q-select
            v-if="performanceComparisonOptions.length > 0"
            :disable="performanceComparisonOptions.length <= 1"
            v-model="selectedComparisonGroup"
            :options="performanceComparisonOptions"
            label="Compare Position"
            dense
            outlined
            emit-value
            map-options
            class="modern-select"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="$q.dark.isActive ? 'bg-grey-8 text-white' : 'bg-white text-dark'"
          />
          <q-tooltip
            v-if="performanceComparisonOptions.length <= 1 && performanceComparisonOptions.length > 0"
          >
            Only global comparison available for this player.
          </q-tooltip>
        </div>
        <div class="col-6">
          <q-select
            v-model="selectedDivisionFilter"
            :options="divisionFilterOptions"
            label="Compare Division"
            dense
            outlined
            emit-value
            map-options
            class="modern-select"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="$q.dark.isActive ? 'bg-grey-8 text-white' : 'bg-white text-dark'"
            @update:model-value="$emit('divisionFilterChange', $event)"
          />
        </div>
      </div>

      <!-- Loading State for Percentiles -->
      <div v-if="showLoadingState" class="percentile-loading-area">
        <div class="loading-content">
          <q-spinner-dots color="primary" size="2em" />
          <div class="loading-text">
            <div class="text-subtitle2">Calculating Performance Percentiles...</div>
            <div class="text-caption text-grey-6">
              {{ isLoadingPercentiles ? 'Fetching data...' : `Retry ${percentilesRetryCount + 1}/${maxRetries}` }}
            </div>
          </div>
          <q-btn 
            v-if="percentilesRetryCount > 0" 
            flat 
            size="sm" 
            color="primary" 
            label="Retry Now" 
            @click="manualRetry"
            class="q-mt-sm"
          />
        </div>
      </div>

      <!-- Percentile Content -->
      <div v-else-if="hasAnyPerformanceData" class="percentile-content-area">
        <div
          v-for="(stats, category, index) in categorizedPerformanceStats"
          :key="`perf-${category}-${selectedComparisonGroup}`"
          class="performance-category"
        >
          <div class="performance-category-header q-mb-sm">
            <span class="performance-category-title">{{ category }}</span>
          </div>
          
          <q-list separator dense class="performance-stats-list">
            <q-item
              v-for="statItem in stats"
              :key="`${statItem.key}-${selectedComparisonGroup}-${selectedDivisionFilter}`"
              class="performance-stat-item modern-stat-item"
            >
              <q-item-section class="stat-name-section">
                <q-item-label
                  lines="1"
                  class="stat-name-label"
                  :title="statItem.name"
                >
                  {{ statItem.name }}
                </q-item-label>
              </q-item-section>
              <q-item-section class="stat-bar-section">
                <div class="stat-bar-container">
                  <div class="stat-bar-track">
                    <div
                      class="stat-bar-fill"
                      :style="getBarFillStyle(statItem.percentile)"
                    ></div>
                  </div>
                  <span
                    v-if="statItem.percentile !== null && statItem.percentile >= 0"
                    class="stat-percentile-text"
                  >
                    {{ Math.round(statItem.percentile) }}
                  </span>
                  <span
                    v-else
                    class="stat-percentile-text text-caption text-grey-6"
                  >
                    N/A
                  </span>
                </div>
              </q-item-section>
              <q-item-section side class="stat-value-section">
                <span class="performance-stat-value">
                  {{ statItem.value !== "-" ? statItem.value : "N/A" }}
                </span>
              </q-item-section>
            </q-item>
          </q-list>
          
          <q-separator
            v-if="index < Object.keys(categorizedPerformanceStats).length - 1"
            class="q-my-md performance-separator"
          />
        </div>
      </div>
      
      <!-- No Data State -->
      <div v-else class="no-performance-data">
        <q-icon name="analytics" size="3em" class="text-grey-4 q-mb-md" />
        <div class="text-subtitle1 text-grey-6">Performance data unavailable</div>
        <div class="text-caption text-grey-6 q-mb-md">
          {{ percentilesRetryCount >= maxRetries 
            ? 'Could not load performance percentiles after multiple attempts.' 
            : 'Performance percentiles are not available for this player.' }}
        </div>
        <q-btn 
          v-if="percentilesRetryCount >= maxRetries" 
          flat 
          color="primary" 
          label="Try Again" 
          @click="manualRetry"
        />
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { computed, defineComponent, toRef } from 'vue'
import { usePercentileRetry } from '../../composables/usePercentileRetry'

export default defineComponent({
  name: 'PerformanceAnalysisCard',
  emits: ['divisionFilterChange'],
  props: {
    player: {
      type: Object,
      required: true
    },
    selectedComparisonGroup: {
      type: String,
      default: 'global'
    },
    selectedDivisionFilter: {
      type: String,
      default: 'all'
    },
    performanceComparisonOptions: {
      type: Array,
      default: () => []
    },
    divisionFilterOptions: {
      type: Array,
      default: () => []
    },
    datasetId: {
      type: String,
      required: false
    }
  },
  setup(props) {
    // Convert props to refs for the composable
    const playerRef = toRef(props, 'player')
    const datasetIdRef = toRef(props, 'datasetId')
    const selectedComparisonGroupRef = toRef(props, 'selectedComparisonGroup')

    // Use the percentile retry composable
    const {
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry
    } = usePercentileRetry(playerRef, datasetIdRef, selectedComparisonGroupRef)
    // Check if player has any performance data
    const hasAnyPerformanceData = computed(() => {
      if (!props.player?.performancePercentiles) return false

      const percentiles = props.player.performancePercentiles[props.selectedComparisonGroup]
      if (!percentiles) return false

      return Object.values(percentiles).some(
        value => value !== null && value !== undefined && value >= 0
      )
    })

    // Categorize performance stats
    const categorizedPerformanceStats = computed(() => {
      if (!props.player?.performancePercentiles) return {}

      const percentiles = props.player.performancePercentiles[props.selectedComparisonGroup]
      if (!percentiles) return {}

      // Define stat categories
      const categories = {
        Attacking: ['goals', 'assists', 'key_passes', 'shots_per_game', 'shots_on_target_per_game'],
        Defensive: [
          'tackles_per_game',
          'interceptions_per_game',
          'clearances_per_game',
          'blocks_per_game'
        ],
        Passing: ['passes_per_game', 'pass_accuracy', 'crosses_per_game', 'long_balls_per_game'],
        Physical: ['fouls_per_game', 'was_fouled_per_game', 'cards_per_game', 'distance_covered']
      }

      const result = {}

      for (const [category, statKeys] of Object.entries(categories)) {
        const categoryStats = []

        for (const statKey of statKeys) {
          if (percentiles[statKey] !== undefined) {
            categoryStats.push({
              key: statKey,
              name: formatStatName(statKey),
              percentile: percentiles[statKey],
              value: getStatValue(statKey)
            })
          }
        }

        if (categoryStats.length > 0) {
          result[category] = categoryStats
        }
      }

      return result
    })

    // Format stat names for display
    const formatStatName = statKey => {
      const nameMap = {
        goals: 'Goals',
        assists: 'Assists',
        key_passes: 'Key Passes',
        shots_per_game: 'Shots/Game',
        shots_on_target_per_game: 'Shots on Target/Game',
        tackles_per_game: 'Tackles/Game',
        interceptions_per_game: 'Interceptions/Game',
        clearances_per_game: 'Clearances/Game',
        blocks_per_game: 'Blocks/Game',
        passes_per_game: 'Passes/Game',
        pass_accuracy: 'Pass Accuracy %',
        crosses_per_game: 'Crosses/Game',
        long_balls_per_game: 'Long Balls/Game',
        fouls_per_game: 'Fouls/Game',
        was_fouled_per_game: 'Was Fouled/Game',
        cards_per_game: 'Cards/Game',
        distance_covered: 'Distance Covered'
      }

      return nameMap[statKey] || statKey.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
    }

    // Get actual stat value from player data
    const getStatValue = statKey => {
      if (!props.player?.performance) return '-'
      return props.player.performance[statKey] || '-'
    }

    // Generate bar fill style based on percentile
    const getBarFillStyle = percentile => {
      if (percentile === null || percentile === undefined || percentile < 0) {
        return { width: '0%', backgroundColor: '#e0e0e0' }
      }

      const width = Math.min(100, Math.max(0, percentile))
      let color = '#f44336' // Red for low percentiles

      if (percentile >= 80) {
        color = '#4caf50' // Green for high percentiles
      } else if (percentile >= 60) {
        color = '#8bc34a' // Light green
      } else if (percentile >= 40) {
        color = '#ff9800' // Orange
      } else if (percentile >= 20) {
        color = '#ff5722' // Orange-red
      }

      return {
        width: `${width}%`,
        backgroundColor: color,
        transition: 'all 0.3s ease'
      }
    }

    return {
      // Existing functionality
      hasAnyPerformanceData,
      categorizedPerformanceStats,
      getBarFillStyle,

      // Percentile retry functionality
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry
    }
  }
})
</script>

<style lang="scss" scoped>
.performance-percentiles-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
}

.performance-card-header {
  background: linear-gradient(135deg, #1976d2 0%, #1565c0 100%);
  color: white;
  padding: 1rem;
  
  .body--dark & {
    background: linear-gradient(135deg, #424242 0%, #303030 100%);
  }
}

.performance-header-title {
  font-weight: 600;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
}

.performance-category-header {
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 0.5rem;
  
  .body--dark & {
    border-color: rgba(255, 255, 255, 0.1);
  }
}

.performance-category-title {
  font-weight: 600;
  font-size: 1rem;
  color: #1976d2;
  
  .body--dark & {
    color: #64b5f6;
  }
}

.performance-stat-item {
  padding: 0.5rem 0;
  
  &:hover {
    background: rgba(25, 118, 210, 0.05);
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.03);
    }
  }
}

.stat-name-label {
  font-size: 0.875rem;
  font-weight: 500;
}

.stat-bar-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 100px;
}

.stat-bar-track {
  flex: 1;
  height: 8px;
  background: #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.1);
  }
}

.stat-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.stat-percentile-text {
  font-size: 0.75rem;
  font-weight: 600;
  min-width: 32px;
  text-align: right;
}

.performance-stat-value {
  font-weight: 600;
  color: #1976d2;
  
  .body--dark & {
    color: #64b5f6;
  }
}

.percentile-loading-area {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem 2rem;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 1rem;
}

.loading-text {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.no-performance-data {
  text-align: center;
  padding: 2rem;
  color: rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.6);
  }
}

.modern-select {
  .q-field__control {
    border-radius: 8px;
  }
}

.performance-separator {
  background: rgba(25, 118, 210, 0.2);
  height: 2px;
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.1);
  }
}
</style> 