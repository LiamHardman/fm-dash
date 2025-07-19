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
                         <q-btn unelevated icon="share" label="Share" @click="shareDataset" class="share-btn-modern"/>
                    </div>
                </div>

                <!-- Filter Bar Integrated into Hero -->
                <div class="filter-bar">
                    <div class="division-filter-container">
                        <div class="division-filter-header">
                            <q-btn
                                @click="selectAllDivisions"
                                size="sm"
                                color="primary"
                                icon="select_all"
                                label="Select All"
                                dense
                                outline
                                class="division-action-btn"
                            />
                            <q-btn
                                @click="clearAllDivisions"
                                size="sm"
                                color="negative"
                                icon="clear_all"
                                label="Clear All"
                                dense
                                outline
                                class="division-action-btn"
                            />
                        </div>
                        <q-select
                            v-model="selectedDivisions"
                            :options="divisionOptions"
                            label="Filter by Division"
                            dense
                            outlined
                            multiple
                            :use-chips="selectedDivisions.length <= 5"
                            use-input
                            @filter="filterDivisionsFn"
                            class="division-filter"
                            dark
                            popup-content-class="bg-grey-10"
                            :display-value="selectedDivisionsDisplayText"
                        >
                             <template v-slot:no-option>
                                <q-item>
                                    <q-item-section class="text-grey">No divisions found</q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>
                    <div class="position-filter-container">
                        <div class="position-filter-header">
                            <q-btn
                                @click="selectAllPositions"
                                size="sm"
                                color="primary"
                                icon="select_all"
                                label="Select All"
                                dense
                                outline
                                class="position-action-btn"
                            />
                            <q-btn
                                @click="clearAllPositions"
                                size="sm"
                                color="negative"
                                icon="clear_all"
                                label="Clear All"
                                dense
                                outline
                                class="position-action-btn"
                            />
                        </div>
                        <q-select
                            v-model="selectedPositions"
                            :options="positionOptions"
                            label="Filter by Position"
                            dense
                            outlined
                            multiple
                            :use-chips="selectedPositions.length <= 5"
                            use-input
                            @filter="filterPositionsFn"
                            class="position-filter"
                            dark
                            popup-content-class="bg-grey-10"
                            :display-value="selectedPositionsDisplayText"
                            emit-value
                            map-options
                        >
                             <template v-slot:no-option>
                                <q-item>
                                    <q-item-section class="text-grey">No positions found</q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>
                    <div class="minutes-filter">
                        <div class="slider-label">Minimum Minutes Played</div>
                        <q-slider
                            v-model="sliderValue"
                            :min="0"
                            :max="maxMinutes"
                            :step="50"
                            label
                            :label-value="`${sliderValue}+ mins`"
                            label-always
                            class="q-mt-sm"
                            dark
                            color="light-blue-4"
                        />
                    </div>
                    <div class="overall-filter">
                        <div class="slider-label">Minimum Overall Rating</div>
                        <q-slider
                            v-model="overallSliderValue"
                            :min="0"
                            :max="maxOverall"
                            :step="1"
                            label
                            :label-value="`${overallSliderValue}+ OVR`"
                            label-always
                            class="q-mt-sm"
                            dark
                            color="light-green-4"
                        />
                    </div>
                </div>
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
                        <div class="tab-content-layout">
                            <q-tabs
                                v-model="attackingPlotTab"
                                dense
                                class="text-grey"
                                active-color="primary"
                                indicator-color="primary"
                                align="justify"
                                narrow-indicator
                            >
                                <q-tab name="shooting" icon="sports_soccer" label="Shooting" />
                            </q-tabs>

                            <q-tab-panels v-model="attackingPlotTab" animated>
                                <q-tab-panel name="shooting">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in attackingShootingCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in attackingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                            </q-tab-panels>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="passing">
                        <div class="tab-content-layout">
                            <q-tabs
                                v-model="passingPlotTab"
                                dense
                                class="text-grey"
                                active-color="primary"
                                indicator-color="primary"
                                align="justify"
                                narrow-indicator
                            >
                                <q-tab name="creative" icon="lightbulb" label="Creative" />
                                <q-tab name="progression" icon="trending_up" label="Progression" />
                                <q-tab name="crossing" icon="swap_horiz" label="Crossing" />
                            </q-tabs>

                            <q-tab-panels v-model="passingPlotTab" animated>
                                <q-tab-panel name="creative">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in passingCreativeCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                                <q-tab-panel name="progression">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in passingProgressionCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                                <q-tab-panel name="crossing">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in passingCrossingCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                            </q-tab-panels>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="defending">
                        <div class="tab-content-layout">
                            <q-tabs
                                v-model="defendingPlotTab"
                                dense
                                class="text-grey"
                                active-color="primary"
                                indicator-color="primary"
                                align="justify"
                                narrow-indicator
                            >
                                <q-tab name="duels" icon="sports_martial_arts" label="Duels" />
                                <q-tab name="pressing" icon="speed" label="Pressing" />
                                <q-tab name="aerial" icon="vertical_align_top" label="Aerial" />
                                <q-tab name="workrate" icon="directions_run" label="Work Rate" />
                            </q-tabs>

                            <q-tab-panels v-model="defendingPlotTab" animated>
                                <q-tab-panel name="duels">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in defendingDuelsCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                                <q-tab-panel name="pressing">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in defendingPressingCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                                <q-tab-panel name="aerial">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in defendingAerialCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                                <q-tab-panel name="workrate">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in defendingWorkrateCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                            </q-tab-panels>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="goalkeeping">
                        <div class="tab-content-layout">
                            <q-tabs
                                v-model="goalkeepingPlotTab"
                                dense
                                class="text-grey"
                                active-color="primary"
                                indicator-color="primary"
                                align="justify"
                                narrow-indicator
                            >
                                <q-tab name="shotstopping" icon="sports_hockey" label="Shot-Stopping" />
                            </q-tabs>

                            <q-tab-panels v-model="goalkeepingPlotTab" animated>
                                <q-tab-panel name="shotstopping">
                                    <div class="charts-grid">
                                        <DynamicScatterPlotCard 
                                            v-for="config in goalkeepingShotstoppingCharts" 
                                            :key="config.title" 
                                            v-bind="config" 
                                            :is-dark-mode="$q.dark.isActive" 
                                            :all-players-data="filteredPlayers"
                                            @player-click="openPlayerDetail"
                                        />
                                    </div>
                                    <div class="stats-grid">
                                        <StatCard v-for="stat in goalkeepingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                                    </div>
                                </q-tab-panel>
                            </q-tab-panels>
                        </div>
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
    </q-page>
</template>

<script setup>
import { debounce, useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
// biome-ignore lint/correctness/noUnusedImports: used in template
import StatCard from '../components/StatCard.vue'
// Dynamic imports for better performance
import { useDynamicComponents } from '../composables/useDynamicComponents.js'
import { usePlayerStore } from '../stores/playerStore'
import { fetchPerformanceData } from '../services/playerService'
import { getNumericValue, getPlayerDivision } from '../utils/playerUtils'
import { formatNumber } from '../utils/currencyUtils'

const router = useRouter()
const route = useRoute()
const $q = useQuasar()
const playerStore = usePlayerStore()

// Initialize dynamic components
const { DynamicPlayerDetailDialog, DynamicScatterPlotCard, initializePreloading } =
  useDynamicComponents()

// Initialize preloading for performance page
initializePreloading(route)

// --- Reactive Data ---
const pageLoadingError = ref('')
const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)
const topPlayersByStat = ref({})
// biome-ignore lint/correctness/noUnusedVariables: used in template
const currentTab = ref('attacking')
const pageLoading = ref(true)

// --- Filter State with new defaults ---
const sliderValue = ref(0)
const selectedMinutes = ref(0)
const overallSliderValue = ref(0)
const selectedOverall = ref(0)
const selectedDivisions = ref([]) // Changed to empty array - default to all divisions
const divisionOptions = ref([])
const selectedPositions = ref([])
const positionOptions = ref([])

// --- Computed Properties from Store ---
const allPlayersData = computed(() => playerStore.allPlayers)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
const currentDatasetId = computed(() => playerStore.currentDatasetId)

// --- Computed Properties for Filtering ---
// biome-ignore lint/correctness/noUnusedVariables: used in template
const maxMinutes = computed(() =>
  Math.max(2000, ...allPlayersData.value.map(p => getNumericValue(p.attributes?.Mins) || 0))
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const maxOverall = computed(() => Math.max(100, ...allPlayersData.value.map(p => p.Overall || 0)))

// Function to calculate thresholds for ~100 players
const calculateThresholds = () => {
  if (!allPlayersData.value.length) return { minutes: 0, overall: 0 }

  // Get all available divisions for comparison
  const allAvailableDivisions = [
    ...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))
  ]

  // First filter by selected divisions and positions
  const filteredByDivisionAndPosition = allPlayersData.value.filter(player => {
    const division = getPlayerDivision(player)
    // Consider "all divisions" if empty array OR if all available divisions are selected
    const divisionMatch =
      selectedDivisions.value.length === 0 ||
      selectedDivisions.value.length === allAvailableDivisions.length ||
      selectedDivisions.value.includes(division)

    const positionMatch =
      selectedPositions.value.length === 0 ||
      selectedPositions.value.some(selectedPos => {
        // Handle position groups
        if (selectedPos === 'Goalkeeper') return player.shortPositions?.includes('GK')
        if (selectedPos === 'Defender')
          return player.shortPositions?.some(pos => ['DC', 'DR', 'DL', 'WBR', 'WBL'].includes(pos))
        if (selectedPos === 'Midfielder')
          return player.shortPositions?.some(pos =>
            ['DM', 'MC', 'MR', 'ML', 'AMR', 'AMC', 'AML'].includes(pos)
          )
        if (selectedPos === 'Forward') return player.shortPositions?.includes('ST')

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
            'ST'
          ].includes(selectedPos)
        ) {
          return player.shortPositions?.includes(selectedPos)
        }

        return false
      })

    return divisionMatch && positionMatch
  })

  // Step 1: Calculate overall threshold to get ~250 players
  const sortedOverall = [...filteredByDivisionAndPosition]
    .map(p => p.Overall || 0)
    .sort((a, b) => b - a)

  const overallTargetIndex = Math.min(249, sortedOverall.length - 1) // Target 250 players
  const overallThreshold = sortedOverall[overallTargetIndex] || 0

  // Step 2: Filter by overall rating first, then calculate minutes threshold
  const playersAfterOverallFilter = filteredByDivisionAndPosition.filter(
    player => (player.Overall || 0) >= overallThreshold
  )

  // Step 3: Calculate minutes threshold from the overall-filtered players
  const sortedMinutes = [...playersAfterOverallFilter]
    .map(p => getNumericValue(p.attributes?.Mins) || 0)
    .sort((a, b) => b - a)

  const minutesTargetIndex = Math.min(99, sortedMinutes.length - 1) // Target 100 players
  const rawMinutesThreshold = sortedMinutes[minutesTargetIndex] || 0
  const minutesThreshold = Math.round(rawMinutesThreshold / 50) * 50

  return { minutes: minutesThreshold, overall: overallThreshold }
}

const availableDivisions = computed(() => {
  const divisions = [
    ...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))
  ].sort()
  // Ensure default selections are included if they exist in the data
  selectedDivisions.value = selectedDivisions.value.filter(d => divisions.includes(d))
  return divisions
})

const availablePositions = computed(() => {
  return [
    // Position Groups
    { label: '--- Position Groups ---', value: null, disable: true },
    { label: 'Goalkeeper', value: 'Goalkeeper', group: true },
    { label: 'Defender', value: 'Defender', group: true },
    { label: 'Midfielder', value: 'Midfielder', group: true },
    { label: 'Forward', value: 'Forward', group: true },

    // Specific Positions
    { label: '--- Specific Positions ---', value: null, disable: true },

    // Goalkeeper
    { label: 'GK - Goalkeeper', value: 'GK' },

    // Defenders
    { label: 'DC - Centre Back', value: 'DC' },
    { label: 'DR - Right Back', value: 'DR' },
    { label: 'DL - Left Back', value: 'DL' },
    { label: 'WBR - Right Wing-Back', value: 'WBR' },
    { label: 'WBL - Left Wing-Back', value: 'WBL' },

    // Midfielders
    { label: 'DM - Defensive Midfielder', value: 'DM' },
    { label: 'MC - Centre Midfielder', value: 'MC' },
    { label: 'MR - Right Midfielder', value: 'MR' },
    { label: 'ML - Left Midfielder', value: 'ML' },
    { label: 'AMC - Attacking Mid (Centre)', value: 'AMC' },
    { label: 'AMR - Attacking Mid (Right)', value: 'AMR' },
    { label: 'AML - Attacking Mid (Left)', value: 'AML' },

    // Forwards
    { label: 'ST - Striker', value: 'ST' }
  ]
})

const filteredPlayers = computed(() => {
  // Get all available divisions for comparison
  const allAvailableDivisions = [
    ...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))
  ]

  return allPlayersData.value.filter(player => {
    const minutesPlayed = getNumericValue(player.attributes?.Mins) || 0
    const overall = player.Overall || 0
    const division = getPlayerDivision(player)

    const matchesMinutes = minutesPlayed >= selectedMinutes.value
    const matchesOverall = overall >= selectedOverall.value
    // Consider "all divisions" if empty array OR if all available divisions are selected
    const matchesDivision =
      selectedDivisions.value.length === 0 ||
      selectedDivisions.value.length === allAvailableDivisions.length ||
      selectedDivisions.value.includes(division)

    const matchesPosition =
      selectedPositions.value.length === 0 ||
      selectedPositions.value.some(selectedPos => {
        // Handle position groups
        if (selectedPos === 'Goalkeeper') return player.shortPositions?.includes('GK')
        if (selectedPos === 'Defender')
          return player.shortPositions?.some(pos => ['DC', 'DR', 'DL', 'WBR', 'WBL'].includes(pos))
        if (selectedPos === 'Midfielder')
          return player.shortPositions?.some(pos =>
            ['DM', 'MC', 'MR', 'ML', 'AMR', 'AMC', 'AML'].includes(pos)
          )
        if (selectedPos === 'Forward') return player.shortPositions?.includes('ST')

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
            'ST'
          ].includes(selectedPos)
        ) {
          return player.shortPositions?.includes(selectedPos)
        }

        return false
      })

    return matchesMinutes && matchesOverall && matchesDivision && matchesPosition
  })
})

// Computed property for custom display text in division filter
// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectedDivisionsDisplayText = computed(() => {
  const count = selectedDivisions.value.length
  const allAvailableDivisions = [
    ...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))
  ]

  if (count === 0 || count === allAvailableDivisions.length) {
    return 'All divisions'
  } else if (count <= 5) {
    return selectedDivisions.value.join(', ')
  } else {
    return `${count} divisions selected`
  }
})

// Computed property for custom display text in position filter
// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectedPositionsDisplayText = computed(() => {
  const count = selectedPositions.value.length
  if (count === 0) {
    return ''
  } else if (count <= 3) {
    // Show labels for the selected positions
    const labels = selectedPositions.value.map(value => {
      const option = availablePositions.value.find(opt => opt.value === value)
      return option ? option.label : value
    })
    return labels.join(', ')
  } else {
    return `${count} positions selected`
  }
})

// --- Data Definitions (Charts & Stats) ---
const scatterPlotConfigs = ref([
  // Attacking plots
  {
    category: 'attacking',
    group: 'shooting',
    title: 'Shooting Performance',
    xAxisKey: 'xG/90',
    yAxisKey: 'Gls/90',
    xAxisLabel: 'Expected Goals per 90',
    yAxisLabel: 'Goals per 90',
    quadrantLabels: {
      topRight: ['Elite', 'Over-performing'],
      topLeft: ['Clinical', 'Over-performing'],
      bottomRight: ['Wasteful', 'Under-performing'],
      bottomLeft: ['Low Threat', 'Under-performing']
    }
  },
  {
    category: 'attacking',
    group: 'shooting',
    title: 'Shooting Efficiency',
    xAxisKey: 'Shot/90',
    yAxisKey: 'Conv %',
    xAxisLabel: 'Shots per 90',
    yAxisLabel: 'Conversion %',
    quadrantLabels: {
      topRight: ['Elite Attacker', ''],
      topLeft: ['Selective Shooter', ''],
      bottomRight: ['Inefficient Volume', ''],
      bottomLeft: ['Limited Threat', '']
    }
  },

  // Passing plots
  {
    category: 'passing',
    group: 'creative',
    title: 'Creative Passing',
    xAxisKey: 'xA/90',
    yAxisKey: 'Asts/90',
    xAxisLabel: 'Expected Assists per 90',
    yAxisLabel: 'Assists per 90',
    quadrantLabels: {
      topRight: ['Elite Creator', ''],
      topLeft: ['Fortunate Creator', ''],
      bottomRight: ['Unlucky Creator', ''],
      bottomLeft: ['Limited Creator', '']
    }
  },
  {
    category: 'passing',
    group: 'creative',
    title: 'Key Passes',
    xAxisKey: 'K Ps/90',
    yAxisKey: 'Asts/90',
    xAxisLabel: 'Key Passes per 90',
    yAxisLabel: 'Assists per 90',
    quadrantLabels: {
      topRight: ['Elite Creator', ''],
      topLeft: ['Clinical Passer', ''],
      bottomRight: ['Wasteful Creator', ''],
      bottomLeft: ['Limited Creator', '']
    }
  },

  {
    category: 'passing',
    group: 'progression',
    title: 'Passing Progression',
    xAxisKey: 'Pr passes/90',
    yAxisKey: 'Pas %',
    xAxisLabel: 'Progressive Passes per 90',
    yAxisLabel: 'Pass Completion %',
    quadrantLabels: {
      topRight: ['Accurate Progressive', ''],
      topLeft: ['Selective Progressive', ''],
      bottomRight: ['Risky Progressive', ''],
      bottomLeft: ['Limited Progressive', '']
    }
  },
  {
    category: 'passing',
    group: 'progression',
    title: 'Passing Volume',
    xAxisKey: 'Ps C/90',
    yAxisKey: 'Pas %',
    xAxisLabel: 'Passes Completed per 90',
    yAxisLabel: 'Pass Completion %',
    quadrantLabels: {
      topRight: ['Elite Passer', ''],
      topLeft: ['Accurate Passer', ''],
      bottomRight: ['Volume Passer', ''],
      bottomLeft: ['Limited Passer', '']
    }
  },

  {
    category: 'passing',
    group: 'crossing',
    title: 'Crossing Efficiency',
    xAxisKey: 'Crs A/90',
    yAxisKey: 'Cr C/90',
    xAxisLabel: 'Crosses Attempted per 90',
    yAxisLabel: 'Crosses Completed per 90',
    quadrantLabels: {
      topRight: ['Elite Crosser', ''],
      topLeft: ['Selective Crosser', ''],
      bottomRight: ['Volume Crosser', ''],
      bottomLeft: ['Limited Crosser', '']
    }
  },
  {
    category: 'passing',
    group: 'crossing',
    title: 'Crossing Impact',
    xAxisKey: 'Crs A/90',
    yAxisKey: 'xA/90',
    xAxisLabel: 'Crosses Attempted per 90',
    yAxisLabel: 'Expected Assists per 90',
    quadrantLabels: {
      topRight: ['Elite Crosser', ''],
      topLeft: ['Clinical Crosser', ''],
      bottomRight: ['Volume Crosser', ''],
      bottomLeft: ['Limited Crosser', '']
    }
  },

  // Defending plots
  {
    category: 'defending',
    group: 'duels',
    title: 'Defensive Duels',
    xAxisKey: 'Tck/90',
    yAxisKey: 'Tck R',
    xAxisLabel: 'Tackles per 90',
    yAxisLabel: 'Tackle Success %',
    quadrantLabels: {
      topRight: ['Elite Ball-Winner', ''],
      topLeft: ['Conservative', ''],
      bottomRight: ['Reckless', ''],
      bottomLeft: ['Passive', '']
    }
  },
  {
    category: 'defending',
    group: 'duels',
    title: 'Defensive Actions',
    xAxisKey: 'Tck/90',
    yAxisKey: 'Int/90',
    xAxisLabel: 'Tackles per 90',
    yAxisLabel: 'Interceptions per 90',
    quadrantLabels: {
      topRight: ['Elite Defender', ''],
      topLeft: ['Tackle Specialist', ''],
      bottomRight: ['Interception Specialist', ''],
      bottomLeft: ['Limited Defender', '']
    }
  },

  {
    category: 'defending',
    group: 'pressing',
    title: 'Pressing Efficiency',
    xAxisKey: 'Pres C/90',
    yAxisKey: 'Poss Won/90',
    xAxisLabel: 'Pressures Completed per 90',
    yAxisLabel: 'Possession Won per 90',
    quadrantLabels: {
      topRight: ['Effective Presser', ''],
      topLeft: ['Positional Winner', ''],
      bottomRight: ['Ineffective Presser', ''],
      bottomLeft: ['Low Activity', '']
    }
  },
  {
    category: 'defending',
    group: 'pressing',
    title: 'Pressing Impact',
    xAxisKey: 'Pres A/90',
    yAxisKey: 'Poss Won/90',
    xAxisLabel: 'Pressures Attempted per 90',
    yAxisLabel: 'Possession Won per 90',
    quadrantLabels: {
      topRight: ['Elite Presser', ''],
      topLeft: ['Selective Presser', ''],
      bottomRight: ['Ineffective Presser', ''],
      bottomLeft: ['Limited Presser', '']
    }
  },

  {
    category: 'defending',
    group: 'aerial',
    title: 'Aerial Duels',
    xAxisKey: 'Aer A/90',
    yAxisKey: 'Hdrs W/90',
    xAxisLabel: 'Aerial Challenges per 90',
    yAxisLabel: 'Headers Won per 90',
    quadrantLabels: {
      topRight: ['Elite Aerial', ''],
      topLeft: ['Selective Winner', ''],
      bottomRight: ['Ineffective Challenger', ''],
      bottomLeft: ['Limited Aerial', '']
    }
  },
  {
    category: 'defending',
    group: 'aerial',
    title: 'Aerial & Clearance Impact',
    xAxisKey: 'K Hdrs/90',
    yAxisKey: 'Clr/90',
    xAxisLabel: 'Key Headers per 90',
    yAxisLabel: 'Clearances per 90',
    quadrantLabels: {
      topRight: ['Elite Aerial Defender', ''],
      topLeft: ['Selective Header', ''],
      bottomRight: ['Clearance Specialist', ''],
      bottomLeft: ['Limited Aerial', '']
    }
  },

  {
    category: 'defending',
    group: 'workrate',
    title: 'Work Rate',
    xAxisKey: 'Dist/90',
    yAxisKey: 'Sprints/90',
    xAxisLabel: 'Distance Covered per 90',
    yAxisLabel: 'Sprints per 90',
    quadrantLabels: {
      topRight: ['Elite Work Rate', ''],
      topLeft: ['Endurance Runner', ''],
      bottomRight: ['Sprint Specialist', ''],
      bottomLeft: ['Limited Movement', '']
    }
  },
  {
    category: 'defending',
    group: 'workrate',
    title: 'Defensive Intensity',
    xAxisKey: 'Dist/90',
    yAxisKey: 'Pres A/90',
    xAxisLabel: 'Distance Covered per 90',
    yAxisLabel: 'Pressures Attempted per 90',
    quadrantLabels: {
      topRight: ['Elite Intensity', ''],
      topLeft: ['Endurance Presser', ''],
      bottomRight: ['Sprint Presser', ''],
      bottomLeft: ['Limited Intensity', '']
    }
  },

  // Goalkeeping plots
  {
    category: 'goalkeeping',
    group: 'shotstopping',
    title: 'Shot-Stopping',
    xAxisKey: 'Con/90',
    yAxisKey: 'Sv %',
    xAxisLabel: 'Goals Conceded per 90',
    yAxisLabel: 'Save Percentage',
    quadrantLabels: {
      topRight: ['Busy & Effective', ''],
      topLeft: ['Elite Goalkeeper', ''],
      bottomRight: ['Struggling', ''],
      bottomLeft: ['Protected', '']
    }
  },
  {
    category: 'goalkeeping',
    group: 'shotstopping',
    title: 'Shot Prevention',
    xAxisKey: 'xGP/90',
    yAxisKey: 'Con/90',
    xAxisLabel: 'Expected Goals Prevented per 90',
    yAxisLabel: 'Goals Conceded per 90',
    quadrantLabels: {
      topRight: ['Elite Shot-Stopper', ''],
      topLeft: ['Protected Keeper', ''],
      bottomRight: ['Exposed Keeper', ''],
      bottomLeft: ['Limited Impact', '']
    }
  }
])

const statCategories = {
  offensive: [
    { key: 'Gls/90', name: 'Goals per 90' },
    { key: 'xG/90', name: 'xG per 90' },
    { key: 'Shot/90', name: 'Shots per 90' },
    { key: 'Conv %', name: 'Conversion %' }
  ],
  passing: [
    { key: 'Asts/90', name: 'Assists per 90' },
    { key: 'xA/90', name: 'xA per 90' },
    { key: 'K Ps/90', name: 'Key Passes per 90' },
    { key: 'Pas %', name: 'Pass Completion %' }
  ],
  defensive: [
    { key: 'Tck/90', name: 'Tackles per 90' },
    { key: 'Int/90', name: 'Interceptions per 90' },
    { key: 'Hdrs W/90', name: 'Headers Won per 90' },
    { key: 'Pres C/90', name: 'Pressures Completed p90' }
  ],
  goalkeeping: [
    { key: 'Con/90', name: 'Goals Conceded p90' },
    { key: 'xGP/90', name: 'xG Prevented p90' },
    { key: 'Sv %', name: 'Save Percentage' },
    { key: 'Clean Sheets', name: 'Clean Sheets' }
  ]
}

// --- Computed properties for each tab ---
const attackingCharts = computed(() =>
  scatterPlotConfigs.value.filter(c => c.category === 'attacking')
)
const passingCharts = computed(() => scatterPlotConfigs.value.filter(c => c.category === 'passing'))
const defendingCharts = computed(() =>
  scatterPlotConfigs.value.filter(c => c.category === 'defending')
)
const goalkeepingCharts = computed(() =>
  scatterPlotConfigs.value.filter(c => c.category === 'goalkeeping')
)

// Add computed properties for each subcategory
// biome-ignore lint/correctness/noUnusedVariables: used in template
const attackingShootingCharts = computed(() =>
  attackingCharts.value.filter(c => c.group === 'shooting')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingCreativeCharts = computed(() =>
  passingCharts.value.filter(c => c.group === 'creative')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingProgressionCharts = computed(() =>
  passingCharts.value.filter(c => c.group === 'progression')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const passingCrossingCharts = computed(() =>
  passingCharts.value.filter(c => c.group === 'crossing')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingDuelsCharts = computed(() => defendingCharts.value.filter(c => c.group === 'duels'))
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingPressingCharts = computed(() =>
  defendingCharts.value.filter(c => c.group === 'pressing')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingAerialCharts = computed(() =>
  defendingCharts.value.filter(c => c.group === 'aerial')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const defendingWorkrateCharts = computed(() =>
  defendingCharts.value.filter(c => c.group === 'workrate')
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const goalkeepingShotstoppingCharts = computed(() =>
  goalkeepingCharts.value.filter(c => c.group === 'shotstopping')
)

const attackingStats = computed(() => statCategories.offensive)
const passingStats = computed(() => statCategories.passing)
const defendingStats = computed(() => statCategories.defensive)
const goalkeepingStats = computed(() => statCategories.goalkeeping)

const allStatsForCalculation = computed(() => [
  ...attackingStats.value,
  ...passingStats.value,
  ...defendingStats.value,
  ...goalkeepingStats.value
])

// --- Helper Methods ---

// --- Core Methods ---
const calculateTopPerformers = () => {
  const playersToProcess = filteredPlayers.value
  const results = {}
  const uniqueStats = [
    ...new Map(allStatsForCalculation.value.map(item => [item.key, item])).values()
  ]

  uniqueStats.forEach(stat => {
    const playersWithStat = playersToProcess.filter(player => {
      const numValue = getNumericValue(player.attributes?.[stat.key])
      return numValue !== null && (stat.key === 'Con/90' ? numValue >= 0 : numValue > 0)
    })
    results[stat.key] = playersWithStat
      .sort((a, b) => {
        const valA = getNumericValue(a.attributes[stat.key])
        const valB = getNumericValue(b.attributes[stat.key])
        return stat.key === 'Con/90' ? valA - valB : valB - valA
      })
      .slice(0, 10)
  })
  topPlayersByStat.value = results
}


// biome-ignore lint/correctness/noUnusedVariables: used in template
const openPlayerDetail = player => {
  // Find the full player object from allPlayersData
  const fullPlayer =
    allPlayersData.value.find(p => p.name === player.name && p.club === player.club) || player

  // Debug: Log the player data to see what fields are available
  console.log('Player selected:', {
    name: fullPlayer.name,
    uid: fullPlayer.uid,
    UID: fullPlayer.UID,
    hasAttributes: !!fullPlayer.attributes,
    hasPerformancePercentiles: !!fullPlayer.performancePercentiles,
    attributesCount: Object.keys(fullPlayer.attributes || {}).length
  })

  playerForDetailView.value = fullPlayer
  showPlayerDetailDialog.value = true
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
      position: 'top'
    })
  })
}

const fetchPlayersAndCurrency = async datasetId => {
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
      ...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))
    ]
    selectedDivisions.value = [...allDivisions]
    selectedPositions.value = [] // Empty means all positions

    // Set both thresholds after data is loaded to ensure at least 100 players
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

// biome-ignore lint/correctness/noUnusedVariables: used in template
const filterDivisionsFn = (val, update) => {
  update(() => {
    const needle = val.toLowerCase()
    divisionOptions.value = availableDivisions.value.filter(
      v => v.toLowerCase().indexOf(needle) > -1
    )
  })
}

// Functions to handle select all and clear all divisions
// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectAllDivisions = () => {
  selectedDivisions.value = [...availableDivisions.value]
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
const clearAllDivisions = () => {
  selectedDivisions.value = []
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
const filterPositionsFn = (val, update) => {
  update(() => {
    const needle = val.toLowerCase()
    positionOptions.value = availablePositions.value.filter(
      option => option.label.toLowerCase().indexOf(needle) > -1
    )
  })
}

// Functions to handle select all and clear all positions
// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectAllPositions = () => {
  selectedPositions.value = availablePositions.value
    .filter(option => option.value !== null && !option.disable)
    .map(option => option.value)
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
const clearAllPositions = () => {
  selectedPositions.value = []
}

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
  debounce(newValue => {
    selectedMinutes.value = newValue
  }, 300)
)

watch(
  overallSliderValue,
  debounce(newValue => {
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
  newDivisions => {
    divisionOptions.value = newDivisions
  },
  { immediate: true }
)

watch(
  availablePositions,
  newPositions => {
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
