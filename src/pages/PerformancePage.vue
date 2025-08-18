<template>
    <q-page class="performance-page">
        <!-- Loading State -->
        <div v-if="pageLoading" class="loading-state">
            <q-spinner-orbit color="primary" size="4em" />
            <div class="loading-text">Loading player database...</div>
        </div>

        <!-- Error State -->
        <div v-else-if="pageLoadingError" class="error-container">
            <q-banner class="error-banner" rounded>
                <template v-slot:avatar>
                    <q-icon name="error" />
                </template>
                {{ pageLoadingError }}
                <q-btn flat color="white" label="Go to Upload Page" @click="router.push('/')" class="q-ml-md" />
            </q-banner>
        </div>

        <!-- No Data State -->
        <div v-else-if="!currentDatasetId || allPlayersData.length === 0" class="no-data-container">
             <q-banner class="no-data-banner">
                <template v-slot:avatar>
                    <q-icon name="warning" />
                </template>
                No player data found. Please upload a dataset first.
                <q-btn flat color="primary" label="Go to Upload Page" @click="router.push('/')" class="q-ml-md"/>
            </q-banner>
        </div>

        <!-- Main Content -->
        <div v-else class="main-content">
            <!-- New Styled Hero Header -->
            <div class="performance-hero-section">
                <div class="hero-content">
                    <div class="hero-left">
                        <div class="hero-title-line">
                            <q-icon name="trending_up" size="2.5rem" />
                            <h1 class="hero-title">Performance Leaders</h1>
                        </div>
                        <p class="hero-subtitle">
                            {{ formatNumber(filteredPlayers.length) }} players matching filters from {{ formatNumber(allPlayersData.length) }} total
                        </p>
                    </div>
                    <div class="hero-right">
                        <q-btn 
                            unelevated 
                            icon="download" 
                            label="Export" 
                            @click="_openExportOptions" 
                            :disable="filteredPlayers.length === 0"
                            class="export-btn-modern q-mr-sm"
                        >
                            <q-tooltip v-if="filteredPlayers.length > 0">
                                Export {{ filteredPlayers.length }} filtered players
                            </q-tooltip>
                            <q-tooltip v-else>
                                No players to export
                            </q-tooltip>
                        </q-btn>
                        <q-btn unelevated icon="share" label="Share" @click="shareDataset" class="share-btn-modern"/>
                    </div>
                </div>

                <!-- Filter Bar Integrated into Hero -->
                <PerformanceFilters
                  :selected-divisions="selectedDivisions"
                  :selected-positions="selectedPositions"
                  :slider-value="sliderValue"
                  :overall-slider-value="overallSliderValue"
                  :division-options="divisionOptions"
                  :position-options="positionOptions"
                  :selected-divisions-display-text="selectedDivisionsDisplayText"
                  :selected-positions-display-text="selectedPositionsDisplayText"
                  :max-minutes="maxMinutes"
                  :max-overall="maxOverall"
                  :filter-divisions-fn="filterDivisionsFn"
                  :filter-positions-fn="filterPositionsFn"
                  :select-all-divisions="selectAllDivisions"
                  :clear-all-divisions="clearAllDivisions"
                  :select-all-positions="selectAllPositions"
                  :clear-all-positions="clearAllPositions"
                  @update:selectedDivisions="(v) => (selectedDivisions = v)"
                  @update:selectedPositions="(v) => (selectedPositions = v)"
                  @update:sliderValue="(v) => (sliderValue = v)"
                  @update:overallSliderValue="(v) => (overallSliderValue = v)"
                />
            </div>

             <!-- Tabbed Content Section -->
            <q-card class="tabs-card">
                <q-tabs
                    v-model="currentTab"
                    dense
                    class="text-grey"
                    active-color="primary"
                    indicator-color="primary"
                    align="justify"
                    narrow-indicator
                >
                    <q-tab name="attacking" icon="sports_soccer" label="Attacking" />
                    <q-tab name="passing" icon="swap_horiz" label="Passing" />
                    <q-tab name="defending" icon="shield" label="Defending" />
                    <q-tab name="goalkeeping" icon="sports_hockey" label="Goalkeeping" />
                </q-tabs>

                <q-separator />

                <q-tab-panels v-model="currentTab" animated>
                    <q-tab-panel name="attacking">
                    <AttackingPanel
                      :filtered-players="filteredPlayers"
                                            :is-dark-mode="isDarkMode" 
                      :top-players-by-stat="topPlayersByStat"
                      :attacking-stats="attackingStats"
                      :attacking-charts="attackingCharts"
                                            @player-click="openPlayerDetail"
                                        />
                                </q-tab-panel>
                    <q-tab-panel name="passing">
                    <PassingPanel
                      :filtered-players="filteredPlayers"
                                            :is-dark-mode="isDarkMode" 
                      :top-players-by-stat="topPlayersByStat"
                      :passing-stats="passingStats"
                      :passing-charts="passingCharts"
                                            @player-click="openPlayerDetail"
                                        />
                                </q-tab-panel>
                    <q-tab-panel name="defending">
                    <DefendingPanel
                      :filtered-players="filteredPlayers"
                                            :is-dark-mode="isDarkMode" 
                      :top-players-by-stat="topPlayersByStat"
                      :defending-stats="defendingStats"
                      :defending-charts="defendingCharts"
                                            @player-click="openPlayerDetail"
                                        />
                                </q-tab-panel>
                    <q-tab-panel name="goalkeeping">
                    <GoalkeepingPanel
                      :filtered-players="filteredPlayers"
                                            :is-dark-mode="isDarkMode" 
                      :top-players-by-stat="topPlayersByStat"
                      :goalkeeping-stats="goalkeepingStats"
                      :goalkeeping-charts="goalkeepingCharts"
                                            @player-click="openPlayerDetail"
                                        />
                    </q-tab-panel>
                </q-tab-panels>
            </q-card>
        </div>

        <!-- Player Detail Dialog -->
                <DynamicPlayerDetailDialog 
            :player="playerForDetailView" 
            :show="showPlayerDetailDialog" 
            @close="showPlayerDetailDialog = false" 
            :currency-symbol="detectedCurrencySymbol" 
            :dataset-id="currentDatasetId" 
        />

        <!-- Export Options Dialog -->
        <DynamicExportOptionsDialog
            :show="showExportOptions"
            :player-count="filteredPlayers.length"
            :export-type="_exportFormat"
            @close="showExportOptions = false"
            @export="_handleExportWithOptions"
        />
    </q-page>
</template>

<script setup>
import { debounce, useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
// Dynamic imports for better performance
import { useDynamicComponents } from '../composables/useDynamicComponents.js'
import { fetchPerformanceData } from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { exportPlayersToCSV } from '../utils/csvExport.js'
import { formatNumber } from '../utils/currencyUtils.js'
import { getNumericValue, getPlayerDivision } from '../utils/playerUtils'
import { useFiltering } from './performance/composables/useFiltering'
import { usePerformanceFilters } from './performance/composables/usePerformanceFilters'
import { useThresholds } from './performance/composables/useThresholds'
import { useTopPlayers } from './performance/composables/useTopPlayers'
import PerformanceFilters from './performance/PerformanceFilters.vue'
import AttackingPanel from './performance/panels/AttackingPanel.vue'
import DefendingPanel from './performance/panels/DefendingPanel.vue'
import GoalkeepingPanel from './performance/panels/GoalkeepingPanel.vue'
import PassingPanel from './performance/panels/PassingPanel.vue'

const router = useRouter()
const route = useRoute()
const $q = useQuasar()
const playerStore = usePlayerStore()
const uiStore = useUiStore()

// Initialize dynamic components
const {
  DynamicPlayerDetailDialog,
  DynamicExportOptionsDialog,
  DynamicScatterPlotCard,
  initializePreloading,
} = useDynamicComponents()

// Initialize preloading for performance page
initializePreloading(route)

// --- Reactive Data ---
const pageLoadingError = ref('')
const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)
const topPlayersByStat = ref({})

// Export functionality
const showExportOptions = ref(false)
const _exportFormat = ref('csv')
// biome-ignore lint/correctness/noUnusedVariables: used in template
const currentTab = ref('attacking')
const pageLoading = ref(true)

// Add computed property for dark mode detection using UI store
const isDarkMode = computed(() => uiStore.isDarkModeActive)

// --- Computed Properties from Store (declare before composables that depend on it) ---
const allPlayersData = computed(() => playerStore.allPlayers)

// --- Filter State with new defaults (via composable) ---
const {
  sliderValue,
  selectedMinutes,
  overallSliderValue,
  selectedOverall,
  selectedDivisions,
  divisionOptions,
  selectedPositions,
  positionOptions,
  availableDivisions,
  availablePositions,
  selectedDivisionsDisplayText,
  selectedPositionsDisplayText,
  maxMinutes,
  maxOverall,
  filterDivisionsFn,
  filterPositionsFn,
  selectAllDivisions,
  clearAllDivisions,
  selectAllPositions,
  clearAllPositions,
} = usePerformanceFilters(allPlayersData)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
const currentDatasetId = computed(() => playerStore.currentDatasetId)

// maxMinutes and maxOverall provided by usePerformanceFilters

// Function to calculate thresholds for ~100 players
const calculateThresholds = () => {
  if (!allPlayersData.value.length) return { minutes: 0, overall: 0 }

  // Get all available divisions for comparison
  const allAvailableDivisions = [
    ...new Set(allPlayersData.value.map((p) => getPlayerDivision(p)).filter(Boolean)),
  ]

  // First filter by selected divisions and positions
  const filteredByDivisionAndPosition = allPlayersData.value.filter((player) => {
    const division = getPlayerDivision(player)
    // Consider "all divisions" if empty array OR if all available divisions are selected
    const divisionMatch =
      selectedDivisions.value.length === 0 ||
      selectedDivisions.value.length === allAvailableDivisions.length ||
      selectedDivisions.value.includes(division)

    const positionMatch =
      selectedPositions.value.length === 0 ||
      selectedPositions.value.some((selectedPos) => {
        const playerPositions = player.short_positions || player.shortPositions || []
        // Handle position groups
        if (selectedPos === 'Goalkeeper') return playerPositions.includes('GK')
        if (selectedPos === 'Defender')
          return playerPositions.some((pos) => ['DC', 'DR', 'DL', 'WBR', 'WBL'].includes(pos))
        if (selectedPos === 'Midfielder')
          return playerPositions.some((pos) =>
            ['DM', 'MC', 'MR', 'ML', 'AMR', 'AMC', 'AML'].includes(pos)
          )
        if (selectedPos === 'Forward') return playerPositions.includes('ST')

        // Handle specific positions
        if (
          [
            'GK',
            'DC',
            'DR',
            'DL',
            'WBR',
            'WBL',
            'DM',
            'MC',
            'MR',
            'ML',
            'AMC',
            'AMR',
            'AML',
            'ST',
          ].includes(selectedPos)
        ) {
          return playerPositions.includes(selectedPos)
        }

        return false
      })

    return divisionMatch && positionMatch
  })

  // Step 1: Calculate overall threshold to get ~250 players
  const sortedOverall = [...filteredByDivisionAndPosition]
    .map((p) => p.Overall || 0)
    .sort((a, b) => b - a)

  const overallTargetIndex = Math.min(249, sortedOverall.length - 1) // Target 250 players
  const overallThreshold = sortedOverall[overallTargetIndex] || 0

  // Step 2: Filter by overall rating first, then calculate minutes threshold
  const playersAfterOverallFilter = filteredByDivisionAndPosition.filter(
    (player) => (player.Overall || 0) >= overallThreshold
  )

  // Step 3: Calculate minutes threshold from the overall-filtered players
  const sortedMinutes = [...playersAfterOverallFilter]
    .map((p) => getNumericValue(p.attributes?.Mins) || 0)
    .sort((a, b) => b - a)

  const minutesTargetIndex = Math.min(99, sortedMinutes.length - 1) // Target 100 players
  const rawMinutesThreshold = sortedMinutes[minutesTargetIndex] || 0
  const minutesThreshold = Math.round(rawMinutesThreshold / 50) * 50

  return { minutes: minutesThreshold, overall: overallThreshold }
}

// availableDivisions and availablePositions moved to composable

const { filteredPlayers } = useFiltering(allPlayersData, {
  selectedMinutes,
  selectedOverall,
  selectedDivisions,
  selectedPositions,
  getNumericValue,
  getPlayerDivision,
})

// display text moved to composable

// --- Data Definitions (Charts & Stats) ---
import { scatterPlotConfigs as scatterPlotConfigsConst } from './performance/config/chartConfigs'

const scatterPlotConfigs = ref(scatterPlotConfigsConst)

import { statCategories as statCategoriesConst } from './performance/config/statConfigs'

const statCategories = statCategoriesConst

// --- Computed properties for each tab ---
const attackingCharts = computed(() =>
  scatterPlotConfigs.value.filter((c) => c.category === 'attacking')
)
const passingCharts = computed(() =>
  scatterPlotConfigs.value.filter((c) => c.category === 'passing')
)
const defendingCharts = computed(() =>
  scatterPlotConfigs.value.filter((c) => c.category === 'defending')
)
const goalkeepingCharts = computed(() =>
  scatterPlotConfigs.value.filter((c) => c.category === 'goalkeeping')
)

// Add computed properties for each subcategory
// biome-ignore lint/correctness/noUnusedVariables: used in template
const attackingShootingCharts = computed(() =>
  attackingCharts.value.filter((c) => c.group === 'shooting')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingCreativeCharts = computed(() =>
  passingCharts.value.filter((c) => c.group === 'creative')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingProgressionCharts = computed(() =>
  passingCharts.value.filter((c) => c.group === 'progression')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingCrossingCharts = computed(() =>
  passingCharts.value.filter((c) => c.group === 'crossing')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingDuelsCharts = computed(() =>
  defendingCharts.value.filter((c) => c.group === 'duels')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingPressingCharts = computed(() =>
  defendingCharts.value.filter((c) => c.group === 'pressing')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingAerialCharts = computed(() =>
  defendingCharts.value.filter((c) => c.group === 'aerial')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingWorkrateCharts = computed(() =>
  defendingCharts.value.filter((c) => c.group === 'workrate')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const goalkeepingShotstoppingCharts = computed(() =>
  goalkeepingCharts.value.filter((c) => c.group === 'shotstopping')
)

const attackingStats = computed(() => statCategories.offensive)
const passingStats = computed(() => statCategories.passing)
const defendingStats = computed(() => statCategories.defensive)
const goalkeepingStats = computed(() => statCategories.goalkeeping)

const allStatsForCalculation = computed(() => [
  ...attackingStats.value,
  ...passingStats.value,
  ...defendingStats.value,
  ...goalkeepingStats.value,
])

// --- Helper Methods ---

// --- Core Methods ---
const { topPlayersByStat: _topPlayersByStat, calculateTopPerformers } = useTopPlayers(
  filteredPlayers,
  computed(() => allStatsForCalculation.value),
  getNumericValue
)
watch(
  _topPlayersByStat,
  (v) => {
    topPlayersByStat.value = v
  },
  { immediate: true, deep: true }
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const openPlayerDetail = (player) => {
  // Find the full player object from allPlayersData
  const fullPlayer =
    allPlayersData.value.find((p) => p.name === player.name && p.club === player.club) || player

  // Debug: Log the player data to see what fields are available
  console.log('Player selected:', {
    name: fullPlayer.name,
    uid: fullPlayer.uid,
    UID: fullPlayer.UID,
    hasAttributes: !!fullPlayer.attributes,
    hasPerformancePercentiles: !!fullPlayer.performancePercentiles,
    attributesCount: Object.keys(fullPlayer.attributes || {}).length,
  })

  playerForDetailView.value = fullPlayer
  showPlayerDetailDialog.value = true
}

// Export functionality
const _openExportOptions = () => {
  showExportOptions.value = true
}

const _handleExportWithOptions = async (exportOptions) => {
  try {
    if (!currentDatasetId.value) {
      $q.notify({
        type: 'negative',
        message: 'No dataset available for export',
        position: 'top',
      })
      return
    }

    console.log(`Exporting dataset ${currentDatasetId.value} with options:`, exportOptions)

    // Export using the backend API
    await exportPlayersToCSV(currentDatasetId.value, exportOptions.format)

    // Show success message
    $q.notify({
      type: 'positive',
      message: `Successfully exported dataset as ${exportOptions.format.toUpperCase()}`,
      position: 'top',
      actions: [
        {
          label: 'Dismiss',
          color: 'white',
        },
      ],
    })

    // Close the dialog
    showExportOptions.value = false
  } catch (error) {
    console.error('Export failed:', error)
    $q.notify({
      type: 'negative',
      message: `Export failed: ${error.message}`,
      position: 'top',
    })
  }
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
const shareDataset = () => {
  if (!currentDatasetId.value) return
  const shareUrl = `${window.location.origin}/performance/${currentDatasetId.value}`
  navigator.clipboard.writeText(shareUrl).then(() => {
    $q.notify({
      message: 'Link copied to clipboard!',
      color: 'positive',
      icon: 'content_copy',
      position: 'top',
    })
  })
}

const fetchPlayersAndCurrency = async (datasetId) => {
  pageLoading.value = true
  pageLoadingError.value = ''
  try {
    // Use the new performance API to get detailed player data with all attributes
    const performanceResponse = await fetchPerformanceData(datasetId)

    // Update the player store with the performance data
    playerStore.setPlayers(performanceResponse.data.players)
    playerStore.setCurrencySymbol(performanceResponse.data.currencySymbol)
    playerStore.setCurrentDatasetId(datasetId)

    // Set default to all divisions and all positions for proper threshold calculation
    const allDivisions = [
      ...new Set(allPlayersData.value.map((p) => getPlayerDivision(p)).filter(Boolean)),
    ]
    selectedDivisions.value = [...allDivisions]
    selectedPositions.value = [] // Empty means all positions

    // Set both thresholds after data is loaded to ensure at least 100 players
    const { calculateThresholds } = useThresholds(
      allPlayersData,
      { selectedDivisions, selectedPositions },
      getNumericValue,
      getPlayerDivision
    )
    const thresholds = calculateThresholds()

    sliderValue.value = thresholds.minutes
    selectedMinutes.value = thresholds.minutes
    overallSliderValue.value = thresholds.overall
    selectedOverall.value = thresholds.overall
  } catch (err) {
    pageLoadingError.value = `Failed to load performance data: ${err.message || 'Unknown server error'}. Please try uploading again.`
  } finally {
    pageLoading.value = false
  }
}

const initializeData = async () => {
  const datasetIdFromQuery = route.query.datasetId
  const datasetIdFromRoute = route.params.datasetId
  const finalDatasetId =
    datasetIdFromRoute || datasetIdFromQuery || sessionStorage.getItem('currentDatasetId')

  if (finalDatasetId) {
    if (datasetIdFromQuery && datasetIdFromQuery !== sessionStorage.getItem('currentDatasetId')) {
      sessionStorage.setItem('currentDatasetId', datasetIdFromQuery)
    } else if (!datasetIdFromQuery && sessionStorage.getItem('currentDatasetId')) {
      // If loading from session, ensure query param is updated for consistency/bookmarking
      router.replace({ query: { datasetId: finalDatasetId } })
    }
    await fetchPlayersAndCurrency(finalDatasetId)
  } else {
    pageLoadingError.value = 'No dataset available. Please upload a dataset first.'
    pageLoading.value = false
  }
}

// filter handlers moved to usePerformanceFilters composable

// --- Watchers & Lifecycle ---
watch(
  () => route.query.datasetId,
  async (newId, oldId) => {
    if (newId && newId !== oldId) {
      sessionStorage.setItem('currentDatasetId', newId)
      await fetchPlayersAndCurrency(newId)
    }
  }
)

watch(
  sliderValue,
  debounce((newValue) => {
    selectedMinutes.value = newValue
  }, 300)
)

watch(
  overallSliderValue,
  debounce((newValue) => {
    selectedOverall.value = newValue
  }, 300)
)

// Watch for changes that should recalculate thresholds
watch(
  [selectedDivisions, selectedPositions],
  () => {
    if (allPlayersData.value.length > 0) {
      const _thresholds = calculateThresholds()
    }
  },
  { deep: true }
)

watch(
  filteredPlayers,
  () => {
    calculateTopPerformers()
  },
  { deep: true, immediate: true }
)

watch(
  availableDivisions,
  (newDivisions) => {
    divisionOptions.value = newDivisions
  },
  { immediate: true }
)

watch(
  availablePositions,
  (newPositions) => {
    positionOptions.value = newPositions
  },
  { immediate: true }
)

// Add these refs after the existing refs
// biome-ignore lint/correctness/noUnusedVariables: used in template
const attackingPlotTab = ref('shooting')
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingPlotTab = ref('creative')
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingPlotTab = ref('duels')
// biome-ignore lint/correctness/noUnusedVariables: used in template
const goalkeepingPlotTab = ref('shotstopping')

onMounted(() => {
  initializeData()
})
</script>

<style lang="scss" scoped>
// Modern Design System Variables
$primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
$success-gradient: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
$warning-gradient: linear-gradient(135deg, #ffeaa7 0%, #fab1a0 100%);
$danger-gradient: linear-gradient(135deg, #fd79a8 0%, #fdcb6e 100%);
$card-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
$card-shadow-hover: 0 12px 40px rgba(0, 0, 0, 0.15);
$border-radius: 16px;
$border-radius-small: 8px;

.performance-page {
    background-color: #f4f6f8;
    .body--dark & {
        background-color: #121212;
    }
}

.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 50vh;
    gap: 1rem;

    .loading-text {
        font-size: 1.2rem;
        color: #666;
        .body--dark & {
            color: #999;
        }
    }
}

.main-content {
    max-width: 1600px;
    margin: 0 auto;
    padding: 2rem;
}

// Hero Section Styling
.performance-hero-section {
    background: $primary-gradient;
    color: white;
    border-radius: $border-radius;
    padding: 2rem 2.5rem;
    margin-bottom: 2rem;
    box-shadow: $card-shadow;
    position: relative;
    overflow: hidden;
    
    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-image: 
            radial-gradient(circle at 20% 50%, rgba(255, 255, 255, 0.1) 0%, transparent 50%),
            radial-gradient(circle at 80% 20%, rgba(255, 255, 255, 0.1) 0%, transparent 50%);
        animation: float 20s ease-in-out infinite;
    }
    
    .hero-content {
        position: relative;
        z-index: 2;
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 1.5rem;

        .hero-left {
            .hero-title-line {
                display: flex;
                align-items: center;
                gap: 1rem;
            }
            .hero-title {
                font-size: 2.25rem;
                font-weight: 700;
                margin: 0;
                line-height: 1.2;
                text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
            }
            .hero-subtitle {
                font-size: 1rem;
                margin: 0.5rem 0 0 0;
                color: rgba(255, 255, 255, 0.8);
            }
        }
        .share-btn-modern {
            background: rgba(255,255,255,0.15);
            color: white;
            font-weight: 600;
            border-radius: $border-radius-small;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
            &:hover {
                background: rgba(255,255,255,0.25);
                transform: translateY(-2px);
            }
        }

        .export-btn-modern {
            background: rgba(255,255,255,0.15);
            color: white;
            font-weight: 600;
            border-radius: $border-radius-small;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
            &:hover {
                background: rgba(255,255,255,0.25);
                transform: translateY(-2px);
            }
        }
    }

    .filter-bar {
        position: relative;
        z-index: 2;
        display: grid;
        grid-template-columns: 1fr 1fr 1fr 1fr;
        gap: 1.5rem;
        align-items: center;
        background: rgba(255,255,255,0.1);
        padding: 1.5rem;
        border-radius: $border-radius-small;
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255,255,255,0.2);

        .division-filter-container {
            width: 100%;
            
            .division-filter-header {
                display: flex;
                gap: 0.5rem;
                margin-bottom: 0.75rem;
                
                .division-action-btn {
                    background: rgba(255,255,255,0.1);
                    border-color: rgba(255,255,255,0.3);
                    color: white;
                    font-size: 0.8rem;
                    
                    &:hover {
                        background: rgba(255,255,255,0.2);
                        transform: translateY(-1px);
                    }
                    
                    &.q-btn--outline {
                        border-width: 1px;
                    }
                    
                    // Override Quasar's color classes for dark theme
                    &.text-primary {
                        color: #90caf9 !important;
                        border-color: rgba(144, 202, 249, 0.5) !important;
                        
                        &:hover {
                            background: rgba(144, 202, 249, 0.1) !important;
                        }
                    }
                    
                    &.text-negative {
                        color: #f48fb1 !important;
                        border-color: rgba(244, 143, 177, 0.5) !important;
                        
                        &:hover {
                            background: rgba(244, 143, 177, 0.1) !important;
                        }
                    }
                }
            }
            
            .division-filter {
                :deep(.q-field__control) {
                    background: rgba(255,255,255,0.15);
                    border: 1px solid rgba(255,255,255,0.2);
                    border-radius: $border-radius-small;
                    color: white;
                }
                :deep(.q-field__label) {
                    color: rgba(255,255,255,0.8);
                }
            }
        }

        .position-filter-container {
            width: 100%;
            
            .position-filter-header {
                display: flex;
                gap: 0.5rem;
                margin-bottom: 0.75rem;
                
                .position-action-btn {
                    background: rgba(255,255,255,0.1);
                    border-color: rgba(255,255,255,0.3);
                    color: white;
                    font-size: 0.8rem;
                    
                    &:hover {
                        background: rgba(255,255,255,0.2);
                        transform: translateY(-1px);
                    }
                    
                    &.q-btn--outline {
                        border-width: 1px;
                    }
                    
                    // Override Quasar's color classes for dark theme
                    &.text-primary {
                        color: #90caf9 !important;
                        border-color: rgba(144, 202, 249, 0.5) !important;
                        
                        &:hover {
                            background: rgba(144, 202, 249, 0.1) !important;
                        }
                    }
                    
                    &.text-negative {
                        color: #f48fb1 !important;
                        border-color: rgba(244, 143, 177, 0.5) !important;
                        
                        &:hover {
                            background: rgba(244, 143, 177, 0.1) !important;
                        }
                    }
                }
            }
            
            .position-filter {
                :deep(.q-field__control) {
                    background: rgba(255,255,255,0.15);
                    border: 1px solid rgba(255,255,255,0.2);
                    border-radius: $border-radius-small;
                    color: white;
                }
                :deep(.q-field__label) {
                    color: rgba(255,255,255,0.8);
                }
                
                // Style for position group separators
                :deep(.q-item--disabled) {
                    font-weight: 600;
                    color: rgba(255,255,255,0.6) !important;
                    background: rgba(255,255,255,0.05);
                    text-align: center;
                    font-size: 0.8rem;
                    pointer-events: none;
                }
            }
        }

        .minutes-filter {
            .slider-label {
                font-size: 0.9rem;
                font-weight: 500;
                color: rgba(255,255,255,0.9);
                margin-bottom: 0.5rem;
            }
            :deep(.q-slider__track) {
                background: rgba(255,255,255,0.2);
            }
            :deep(.q-slider__track--active) {
                background: white;
            }
            :deep(.q-slider__thumb) {
                background: white;
                border: 2px solid rgba(255,255,255,0.5);
            }
        }

        .overall-filter {
            .slider-label {
                font-size: 0.9rem;
                font-weight: 500;
                color: rgba(255,255,255,0.9);
                margin-bottom: 0.5rem;
            }
            :deep(.q-slider__track) {
                background: rgba(255,255,255,0.2);
            }
            :deep(.q-slider__track--active) {
                background: white;
            }
            :deep(.q-slider__thumb) {
                background: white;
                border: 2px solid rgba(255,255,255,0.5);
            }
        }
    }
}

// Tabs Styling
.tabs-card {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    overflow: hidden;
    background: white;
    transition: all 0.3s ease;

    .body--dark & {
        background: #1e1e1e;
        border: 1px solid rgba(255,255,255,0.1);
    }

    &:hover {
        box-shadow: $card-shadow-hover;
        transform: translateY(-2px);
    }

    .q-tabs {
        padding: 0 1rem;
        background: rgba(0,0,0,0.02);
        border-bottom: 1px solid rgba(0,0,0,0.05);

        .body--dark & {
            background: rgba(255,255,255,0.02);
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }

        .q-tab {
            font-weight: 600;
            padding: 1rem 1.5rem;
            transition: all 0.3s ease;

            &:hover {
                background: rgba(0,0,0,0.02);
                .body--dark & {
                    background: rgba(255,255,255,0.05);
                }
            }

            &--active {
                color: #667eea;
                .body--dark & {
                    color: #90caf9;
                }
            }
        }
    }

    .q-tab-panel {
        padding: 2rem;
    }
}

.tab-content-layout {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.category-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #333;
    padding-bottom: 0.5rem;
    border-bottom: 2px solid #667eea;
    align-self: flex-start;
    margin-bottom: 1rem;

    .body--dark & {
        color: #f5f5f5;
        border-bottom-color: #90caf9;
    }
}

.charts-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-bottom: 2rem;
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
}

// Banners for error/no-data states
.error-container, .no-data-container {
    padding: 2rem;
}

.error-banner {
    background: $danger-gradient;
    color: white;
    border-radius: $border-radius;
    box-shadow: $card-shadow;
}

.no-data-banner {
    background: white;
    border: 1px solid rgba(0,0,0,0.1);
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    
    .body--dark & {
        background: #1e1e1e;
        color: #f5f5f5;
        border-color: rgba(255,255,255,0.1);
    }
}

@keyframes float {
    0%, 100% { transform: translateY(0px) rotate(0deg); }
    33% { transform: translateY(-20px) rotate(1deg); }
    66% { transform: translateY(-10px) rotate(-1deg); }
}

// Responsive Design
@media (max-width: 1200px) {
    .main-content {
        padding: 1.5rem;
    }
    
    .performance-hero-section {
        padding: 1.5rem;
    }
    
    .filter-bar {
        grid-template-columns: 1fr 1fr;
        gap: 1.5rem;
    }
}

@media (max-width: 768px) {
    .main-content {
        padding: 1rem;
    }
    
    .performance-hero-section {
        padding: 1.25rem;
        
        .hero-content {
            flex-direction: column;
            gap: 1rem;
            
            .hero-left {
                text-align: center;
                
                .hero-title-line {
                    justify-content: center;
                }
            }
        }
    }
    
    .filter-bar {
        grid-template-columns: 1fr;
        gap: 1rem;
    }
    
    .charts-grid {
        grid-template-columns: 1fr;
    }
    
    .stats-grid {
        grid-template-columns: 1fr;
    }
}

// For goalkeeping, we want single column
.goalkeeping .charts-grid {
    grid-template-columns: 1fr;
}
</style>
