<template>
    <q-page class="nation-view-page">
        <div class="main-content">
            <!-- Error Banner -->
            <q-banner
                v-if="pageLoadingError"
                class="error-banner"
                rounded
            >
                <template v-slot:avatar>
                    <q-icon name="error" />
                </template>
                {{ pageLoadingError }}
                <q-btn
                    flat
                    color="white"
                    label="Go to Upload Page"
                    @click="router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>

            <!-- Share Button - Modern Design -->
            <div v-if="!pageLoadingError && currentDatasetId" class="share-section">
                <q-btn
                    unelevated
                    icon-right="share"
                    label="Share Dataset"
                    color="primary"
                    @click="shareDataset"
                    class="share-btn-modern"
                    size="md"
                >
                    <q-tooltip>Share this dataset with others</q-tooltip>
                </q-btn>
            </div>

            <!-- No Nation Selected State -->
            <div v-if="!pageLoadingError && !selectedNationName && !pageLoading" class="empty-state">
                <q-card class="empty-state-card">
                    <q-card-section class="empty-state-content">
                        <div class="empty-state-icon">
                            <q-icon name="flag" size="4rem" />
                        </div>
                        <h3 class="empty-state-title">Select a Nation to Begin</h3>
                        <p class="empty-state-description">
                            Choose a nation from the search above to unlock detailed tactical analysis, formation optimization, and squad insights.
                        </p>
                        <q-btn
                            color="primary"
                            unelevated
                            label="Browse Dataset"
                            @click="router.push(`/dataset/${currentDatasetId}`)"
                            v-if="currentDatasetId"
                            class="empty-state-btn"
                        />
                    </q-card-section>
                </q-card>
            </div>

            <!-- Loading States -->
            <div v-if="pageLoading" class="loading-state">
                <q-spinner-orbit color="primary" size="4em" />
                <div class="loading-text">Loading player database...</div>
            </div>
            
            <div v-else-if="cacheLoading" class="loading-state">
                <q-spinner-cube color="primary" size="3em" />
                <div class="loading-text">Loading cached nation ratings...</div>
            </div>
            
            <div v-else-if="loadingNation" class="loading-state">
                <q-spinner-dots color="primary" size="3em" />
                <div class="loading-text">Analyzing nation composition...</div>
            </div>

            <!-- Main Nation Content -->
            <div v-if="!pageLoading && !pageLoadingError && selectedNationName && !loadingNation" class="nation-dashboard">
                
                <!-- Nation Header Section (replaces hero) -->
                <div class="nation-hero-section">
                    <div class="nation-hero-content">
                        <div class="nation-primary-info">
                            <div class="nation-name-section">
                                <h1 class="nation-name-hero">{{ selectedNationName }}</h1>
                                <div class="nation-flag-hero">
                                    <img
                                        v-if="currentNationFlagISO"
                                        :src="`https://flagcdn.com/w20/${currentNationFlagISO.toLowerCase()}.png`"
                                        :alt="selectedNationName"
                                        width="24"
                                        height="16"
                                        class="nationality-flag"
                                    />
                                    <q-icon v-else name="flag" size="1.2rem" />
                                    <span>{{ nationPlayers.length }} Players</span>
                                </div>
                            </div>
                            
                            <!-- Star Rating Display -->
                            <div v-if="bestNationAverageOverall !== null" class="star-rating-display">
                                <div class="overall-score">{{ bestNationAverageOverall }}</div>
                                <div class="star-container">
                                    <span
                                        v-for="star in 5"
                                        :key="star"
                                        class="star-modern"
                                        :class="getStarClass(bestNationAverageOverall, star)"
                                    >
                                        ★
                                    </span>
                                </div>
                                <div class="rating-label">Overall Rating</div>
                            </div>
                        </div>
                        
                        <!-- Performance Metrics -->
                        <div v-if="currentNationSectionRatings.attRating > 0 || currentNationSectionRatings.midRating > 0 || currentNationSectionRatings.defRating > 0" class="performance-metrics">
                            <div class="metrics-grid">
                                <div v-if="currentNationSectionRatings.attRating > 0" class="metric-card attack">
                                    <div class="metric-icon">⚔️</div>
                                    <div class="metric-value" :class="getOverallClass(currentNationSectionRatings.attRating)">
                                        {{ currentNationSectionRatings.attRating }}
                                    </div>
                                    <div class="metric-label">Attack</div>
                                </div>
                                <div v-if="currentNationSectionRatings.midRating > 0" class="metric-card midfield">
                                    <div class="metric-icon">⚽</div>
                                    <div class="metric-value" :class="getOverallClass(currentNationSectionRatings.midRating)">
                                        {{ currentNationSectionRatings.midRating }}
                                    </div>
                                    <div class="metric-label">Midfield</div>
                                </div>
                                <div v-if="currentNationSectionRatings.defRating > 0" class="metric-card defense">
                                    <div class="metric-icon">🛡️</div>
                                    <div class="metric-value" :class="getOverallClass(currentNationSectionRatings.defRating)">
                                        {{ currentNationSectionRatings.defRating }}
                                    </div>
                                    <div class="metric-label">Defense</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Formation & Tactics Section - New Layout -->
                <div class="formation-tactics-layout">
                    <!-- Left Side - Formation Controls and Squad Depth -->
                    <div class="formation-controls-panel">
                        <!-- Formation Selection -->
                        <q-card class="formation-card">
                            <q-card-section>
                                <div class="card-header">
                                    <h3 class="card-title">
                                        <q-icon name="diagram" class="card-icon" />
                                        Tactical Setup
                                    </h3>
                                    <p class="card-subtitle">Optimize your formation and lineup</p>
                                </div>
                                
                                <div class="formation-controls">
                                    <q-select
                                        v-model="selectedFormationKey"
                                        :options="formationOptions"
                                        label="Select Formation"
                                        outlined
                                        emit-value
                                        map-options
                                        class="formation-select"
                                        :label-color="quasarInstance.dark.isActive ? 'grey-4' : ''"
                                    />
                                    
                                    <q-banner
                                        v-if="calculationMessage"
                                        class="calculation-banner"
                                        :class="calculationMessageClass"
                                    >
                                        {{ calculationMessage }}
                                    </q-banner>
                                </div>
                            </q-card-section>
                        </q-card>

                        <!-- Squad Depth Card -->
                        <q-card 
                            v-if="selectedFormationKey && Object.keys(squadComposition).length > 0"
                            class="squad-depth-card"
                        >
                            <q-card-section>
                                <div class="card-header">
                                    <h3 class="card-title">
                                        <q-icon name="groups_3" class="card-icon" />
                                        Squad Depth
                                    </h3>
                                    <p class="card-subtitle">Player availability by position</p>
                                </div>
                                
                                <div class="squad-depth-grid">
                                    <div
                                        v-for="slot in currentFormationLayout.flatMap(row => row.positions)"
                                        :key="slot.id"
                                        class="depth-position-modern"
                                    >
                                        <div class="position-header">
                                            <span class="position-name">
                                                {{ getSlotDisplayName(slot, currentFormationLayout.flatMap(r => r.positions)) }}
                                            </span>
                                            <span class="player-count">
                                                {{ squadComposition[slot.id]?.length || 0 }} players
                                            </span>
                                        </div>
                                        
                                        <div v-if="squadComposition[slot.id] && squadComposition[slot.id].length > 0" class="depth-players-modern">
                                            <div
                                                v-for="(playerEntry, index) in squadComposition[slot.id].slice(0, 3)"
                                                :key="playerEntry.player.name + '-' + slot.id + '-' + index"
                                                class="player-card-mini"
                                                :class="{ 'is-starter': index === 0 }"
                                                @click="handlePlayerSelectedFromNation(playerEntry.player)"
                                            >
                                                <div class="player-rank">{{ index + 1 }}</div>
                                                <div class="player-info">
                                                    <div class="player-name">{{ playerEntry.player.name }}</div>
                                                    <div class="player-positions">
                                                        {{ playerEntry.player.shortPositions?.slice(0, 2).join(', ') || 'N/A' }}
                                                    </div>
                                                </div>
                                                <div class="player-rating" :class="getOverallClass(playerEntry.overallInRole)">
                                                    {{ playerEntry.overallInRole }}
                                                </div>
                                            </div>
                                        </div>
                                        
                                        <div v-else class="no-players-state">
                                            <q-icon name="person_off" size="1.5rem" />
                                            <span>No suitable players</span>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>

                    <!-- Right Side - Pitch Visualization -->
                    <div class="formation-display-panel">
                        <q-card class="pitch-card" v-if="selectedFormationKey">
                            <q-card-section>
                                <div class="card-header">
                                    <h3 class="card-title">
                                        <q-icon name="stadium" class="card-icon" />
                                        Formation View
                                    </h3>
                                    <p class="card-subtitle">Interactive pitch with your starting XI</p>
                                </div>
                                
                                <div class="pitch-container">
                                    <PitchDisplay
                                        :formation="currentFormationLayout"
                                        :players="bestNationPlayersForPitch"
                                        @player-click="handlePlayerSelectedFromNation"
                                        @player-moved="handlePlayerMovedOnPitch"
                                    />
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <!-- Players Table -->
                <q-card class="players-table-card">
                    <q-card-section>
                        <div class="card-header">
                            <h3 class="card-title">
                                <q-icon name="group" class="card-icon" />
                                Squad Overview
                            </h3>
                            <p class="card-subtitle">
                                All {{ nationPlayers.length }} players from {{ selectedNationName }}
                            </p>
                        </div>
                        
                        <div class="table-container">
                            <PlayerDataTable
                                v-if="nationPlayers.length > 0"
                                :players="nationPlayers"
                                :loading="false"
                                @player-selected="handlePlayerSelectedFromNation"
                                @team-selected="handleTeamSelected"
                                :is-goalkeeper-view="nationIsGoalkeeperView"
                                :currency-symbol="detectedCurrencySymbol"
                                :dataset-id="currentDatasetId"
                                class="modern-table"
                            />
                            <div v-else class="no-players-banner">
                                <q-icon name="person_off" size="3rem" />
                                <h4>No Players Found</h4>
                                <p>No players found for this nation with the current data filters.</p>
                            </div>
                        </div>
                    </q-card-section>
                </q-card>
            </div>

            <!-- Nations Overview Card - When no nation is selected -->
            <q-card
                v-if="!pageLoading && !pageLoadingError && !selectedNationName && !loadingNation && allPlayersData.length > 0"
                class="nations-overview-card"
            >
                <q-card-section>
                    <div class="card-header">
                        <h3 class="card-title">
                            <q-icon name="public" class="card-icon" />
                            Nations Overview
                        </h3>
                        <p class="card-subtitle">Select a nation to analyze player talents and distributions</p>
                    </div>
                    
                    <!-- NEW: Calculation Progress Bar -->
                    <div v-if="isCalculatingRatings" class="calculation-progress-section">
                        <div class="progress-header">
                            <div class="progress-title">
                                <q-icon name="calculate" class="q-mr-sm" />
                                Calculating Nation Ratings...
                            </div>
                            <div class="progress-stats">
                                {{ calculationProgress.current }} / {{ calculationProgress.total }}
                            </div>
                        </div>
                        <q-linear-progress
                            :value="calculationProgress.total > 0 ? calculationProgress.current / calculationProgress.total : 0"
                            color="primary"
                            track-color="grey-3"
                            class="progress-bar"
                            :class="{ 'progress-bar-dark': quasarInstance.dark.isActive }"
                        />
                        <div class="progress-description">
                            Nations will update with their tactical ratings as calculations complete
                        </div>
                    </div>
                    
                    <div class="modern-filter-section">
                        <q-select
                            v-model="selectedNationName"
                            :options="nationOptions"
                            label="Search and Select Nation"
                            outlined
                            dense
                            use-input
                            hide-selected
                            fill-input
                            input-debounce="300"
                            @filter="filterNationOptions"
                            @update:model-value="loadNationPlayers"
                            :label-color="quasarInstance.dark.isActive ? 'grey-4' : ''"
                            :popup-content-class="quasarInstance.dark.isActive ? 'bg-grey-8 text-white' : 'bg-white text-dark'"
                            clearable
                            @clear="clearNationSelection"
                            :disable="pageLoading || allPlayersData.length === 0"
                            class="nation-select"
                        >
                            <template v-slot:no-option>
                                <q-item>
                                    <q-item-section class="text-grey">
                                        No nations found.
                                    </q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>
                    
                    <div class="nations-list">
                        <div
                            v-for="nation in (showAllNations ? nationsWithRatings : nationsWithRatings.slice(0, 10)).filter(n => n && n.name)"
                            :key="nation.name || 'unknown'"
                            class="nation-row"
                            :class="{ 'calculating': nation.isCalculating }"
                            @click="selectNation(nation.name)"
                        >
                            <div class="nation-flag-container">
                                <img
                                    v-if="nation.nationality_iso"
                                    :src="`https://flagcdn.com/w20/${nation.nationality_iso.toLowerCase()}.png`"
                                    :alt="nation.name"
                                    width="24"
                                    height="16"
                                    class="nationality-flag"
                                    @error="onFlagError($event, nation)"
                                />
                                <q-icon
                                    v-else
                                    name="flag"
                                    size="sm"
                                    :color="quasarInstance.dark.isActive ? 'grey-6' : 'grey-7'"
                                />
                            </div>
                            <div class="nation-info">
                                <div class="nation-name">{{ nation.name }}</div>
                                <div class="player-count">{{ nation.playerCount }} players</div>
                            </div>
                            <div class="nation-section-ratings">
                                <!-- NEW: Show calculating state -->
                                <div v-if="nation.isCalculating" class="calculating-ratings">
                                    <q-spinner-dots color="primary" size="1.5em" />
                                    <span class="calculating-text">Calculating...</span>
                                </div>
                                <!-- NEW: Show ratings when available -->
                                <div 
                                    v-else-if="nation.bestFormationOverall !== null && nation.bestFormationOverall > 0"
                                    class="section-ratings-large"
                                >
                                    <div class="section-rating-large att">
                                        <span class="section-label-large">ATT</span>
                                        <span 
                                            class="section-value-large"
                                            :class="getOverallClass(nation.attRating)"
                                        >
                                            {{ nation.attRating }}
                                        </span>
                                    </div>
                                    <div class="section-rating-large mid">
                                        <span class="section-label-large">MID</span>
                                        <span 
                                            class="section-value-large"
                                            :class="getOverallClass(nation.midRating)"
                                        >
                                            {{ nation.midRating }}
                                        </span>
                                    </div>
                                    <div class="section-rating-large def">
                                        <span class="section-label-large">DEF</span>
                                        <span 
                                            class="section-value-large"
                                            :class="getOverallClass(nation.defRating)"
                                        >
                                            {{ nation.defRating }}
                                        </span>
                                    </div>
                                </div>
                                <!-- NEW: Show pending state for nations not yet calculated -->
                                <div 
                                    v-else-if="nation.bestFormationOverall === null"
                                    class="pending-calculation"
                                >
                                    <q-icon name="schedule" size="1em" class="q-mr-xs" />
                                    <span class="pending-text">Pending...</span>
                                </div>
                                <!-- Show incomplete squad for nations with calculated 0 rating -->
                                <div 
                                    v-else
                                    class="no-rating-message"
                                >
                                    Incomplete Squad
                                </div>
                            </div>
                            <div class="nation-overall">
                                <!-- NEW: Show calculating state -->
                                <div v-if="nation.isCalculating" class="nation-rating">
                                    <div class="calculating-overall">
                                        <q-spinner-dots color="primary" size="1em" />
                                    </div>
                                </div>
                                <!-- NEW: Show ratings when available -->
                                <div v-else-if="nation.bestFormationOverall !== null" class="nation-rating">
                                    <div 
                                        class="highest-overall-large"
                                        :class="getOverallClass(nation.bestFormationOverall)"
                                    >
                                        {{ nation.bestFormationOverall > 0 ? nation.bestFormationOverall + ' AVG' : 'N/A' }}
                                    </div>
                                    <div class="star-rating-large">
                                        <span
                                            v-for="star in 5"
                                            :key="star"
                                            class="star-large"
                                            :class="getStarClass(nation.bestFormationOverall, star)"
                                        >
                                            ★
                                        </span>
                                    </div>
                                </div>
                                <!-- NEW: Show pending state -->
                                <div v-else class="nation-rating">
                                    <div class="pending-overall">
                                        <q-icon name="schedule" size="1em" />
                                        <div class="pending-stars">- - -</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Show More Button -->
                    <div v-if="!showAllNations && allPlayersData.length > 0" class="text-center q-mt-md">
                        <q-btn
                            flat
                            color="primary"
                            @click="showAllNations = true"
                            class="show-more-btn"
                        >
                            Show All Nations
                            <q-icon name="expand_more" class="q-ml-sm" />
                        </q-btn>
                    </div>
                </q-card-section>
            </q-card>

            <!-- Additional Banners -->
            <q-banner
                v-else-if="!pageLoading && !loadingNation && allPlayersData.length > 0 && !selectedNationName"
                class="info-banner"
            >
                <template v-slot:avatar>
                    <q-icon name="info" />
                </template>
                Please select a nation to view its players and analyze formations.
            </q-banner>
            
            <q-banner
                v-else-if="!pageLoading && !loadingNation && allPlayersData.length === 0 && !pageLoadingError"
                class="warning-banner"
            >
                <template v-slot:avatar>
                    <q-icon name="warning" />
                </template>
                No player data available. Please upload a player file on the main page first.
                <q-btn
                    flat
                    color="primary"
                    label="Go to Upload Page"
                    @click="router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>
        </div>

        <!-- Player Detail Dialog -->
        <PlayerDetailDialog
            :player="playerForDetailView"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
            :currency-symbol="detectedCurrencySymbol"
        />
    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PitchDisplay from '../components/PitchDisplay.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerStore } from '../stores/playerStore'
import { formationCache } from '../utils/formationCache'
import { formations, getFormationLayout } from '../utils/formations'
import { cacheLogger } from '../utils/logger'

const fmSlotRoleMatcher = {
  GK: ['Goalkeeper'],
  'D (R)': ['Defender (Right)', 'Right Back'],
  'D (L)': ['Defender (Left)', 'Left Back'],
  'D (C)': ['Defender (Centre)', 'Centre Back'],
  'WB (R)': ['Wing-Back (Right)', 'Right Wing-Back'],
  'WB (L)': ['Wing-Back (Left)', 'Left Wing-Back'],
  'DM (C)': ['Defensive Midfielder (Centre)', 'Centre Defensive Midfielder'],
  'M (R)': ['Midfielder (Right)', 'Right Midfielder'],
  'M (L)': ['Midfielder (Left)', 'Left Midfielder'],
  'M (C)': ['Midfielder (Centre)', 'Centre Midfielder'],
  'AM (R)': ['Attacking Midfielder (Right)', 'Right Attacking Midfielder', 'Winger (Right)'],
  'AM (L)': ['Attacking Midfielder (Left)', 'Left Attacking Midfielder', 'Winger (Left)'],
  'AM (C)': ['Attacking Midfielder (Centre)', 'Centre Attacking Midfielder'],
  'ST (C)': ['Striker (Centre)', 'Striker']
}

export default {
  name: 'NationsPage',
  components: { PlayerDataTable, PlayerDetailDialog, PitchDisplay },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()

    const selectedNationName = ref(null)
    const nationOptions = ref([])
    const allNationNamesCache = ref([])
    const nationPlayers = ref([])
    const loadingNation = ref(false)
    const pageLoading = ref(true)
    const pageLoadingError = ref('')

    // Computed properties from store
    const allPlayersData = computed(() => playerStore.allPlayers)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
    const currentDatasetId = computed(() => playerStore.currentDatasetId)

    const selectedFormationKey = ref(null)

    const squadComposition = ref({})

    const bestNationAverageOverall = ref(null)
    const calculationMessage = ref('')
    const calculationMessageClass = ref('')

    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)

    // NEW: Async nations system
    const nationsData = ref([])
    const isCalculatingRatings = ref(false)
    const calculationProgress = ref({ current: 0, total: 0 })
    const calculationQueue = ref([])
    const isProcessingQueue = ref(false)

    // NEW: Nation ratings cache system
    const CACHE_VERSION = '1.1' // Increment when calculation logic changes
    const cacheLoading = ref(false)
    const cacheKey = ref(null)

    // NEW: Generate cache key based on dataset and player data
    const generateCacheKey = () => {
      if (!currentDatasetId.value || !allPlayersData.value.length) {
        return null
      }

      // Create a simple hash based on dataset ID, player count, and key player data
      // This ensures cache invalidation when the underlying data changes
      const playerCount = allPlayersData.value.length
      const samplePlayerData = allPlayersData.value
        .slice(0, 10)
        .map(p => `${p.name}:${p.Overall || 0}:${p.nationality}`)
        .join('|')
      const cacheInput = `${currentDatasetId.value}:${playerCount}:${samplePlayerData}:${CACHE_VERSION}`

      // Simple hash function (could use crypto.subtle.digest for better hashing in production)
      let hash = 0
      for (let i = 0; i < cacheInput.length; i++) {
        const char = cacheInput.charCodeAt(i)
        hash = (hash << 5) - hash + char
        hash = hash & hash // Convert to 32-bit integer
      }

      return `nation_ratings_${Math.abs(hash).toString(36)}`
    }

    // NEW: Save nation ratings to cache
    const saveNationRatingsToCache = async () => {
      if (!cacheKey.value || nationsData.value.length === 0) {
        return
      }

      try {
        const cacheData = {
          version: CACHE_VERSION,
          datasetId: currentDatasetId.value,
          generatedAt: new Date().toISOString(),
          playerCount: allPlayersData.value.length,
          nationsData: nationsData.value.map(nation => ({
            name: nation.name,
            nationality_iso: nation.nationality_iso,
            playerCount: nation.playerCount,
            bestFormationOverall: nation.bestFormationOverall,
            attRating: nation.attRating,
            midRating: nation.midRating,
            defRating: nation.defRating
          }))
        }

        // Save to API endpoint which will handle local/S3 storage based on configuration
        const response = await fetch(`/api/cache/nation-ratings/${cacheKey.value}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(cacheData)
        })

        if (response.ok) {
          cacheLogger.cache('save', cacheKey.value, 'Nation ratings cached successfully')
        } else {
          cacheLogger.warn(`Failed to cache nation ratings: ${response.status}`)
        }
      } catch (error) {
        cacheLogger.warn('Error saving nation ratings to cache:', error)
      }
    }

    // NEW: Load nation ratings from cache
    const loadNationRatingsFromCache = async () => {
      if (!cacheKey.value) {
        return false
      }

      cacheLoading.value = true
      try {
        const response = await fetch(`/api/cache/nation-ratings/${cacheKey.value}`)

        if (!response.ok) {
          cacheLogger.cache('miss', cacheKey.value, 'No cached nation ratings found')
          return false
        }

        const cacheData = await response.json()

        // Validate cache data
        if (cacheData.version !== CACHE_VERSION) {
          cacheLogger.cache(
            'invalidate',
            cacheKey.value,
            `Cache version mismatch (${cacheData.version} vs ${CACHE_VERSION})`
          )
          return false
        }

        if (cacheData.datasetId !== currentDatasetId.value) {
          cacheLogger.cache('invalidate', cacheKey.value, 'Cache dataset mismatch')
          return false
        }

        if (cacheData.playerCount !== allPlayersData.value.length) {
          cacheLogger.cache(
            'invalidate',
            cacheKey.value,
            `Player count changed (${cacheData.playerCount} vs ${allPlayersData.value.length})`
          )
          return false
        }

        // Restore cached nation data
        const restoredNations = cacheData.nationsData.map(cachedNation => ({
          ...cachedNation,
          isCalculating: false,
          players: [], // Will be populated if needed
          topPlayersByPosition: {} // Will be populated if needed
        }))

        nationsData.value = restoredNations

        cacheLogger.cache(
          'hit',
          cacheKey.value,
          `Loaded ${restoredNations.length} nation ratings from cache (generated ${cacheData.generatedAt})`
        )

        // Show a brief message about cache usage
        calculationMessage.value = `Loaded nation ratings from cache (generated ${new Date(cacheData.generatedAt).toLocaleString()})`
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'

        setTimeout(() => {
          calculationMessage.value = ''
        }, 3000)

        return true
      } catch (error) {
        cacheLogger.warn('Error loading nation ratings from cache:', error)
        return false
      } finally {
        cacheLoading.value = false
      }
    }

    // NEW: Clear cache (useful for debugging or forcing recalculation)
    const _clearNationRatingsCache = async () => {
      if (!cacheKey.value) return

      try {
        const response = await fetch(`/api/cache/nation-ratings/${cacheKey.value}`, {
          method: 'DELETE'
        })

        if (response.ok) {
          cacheLogger.cache('clear', cacheKey.value, 'Cleared nation ratings cache')
        } else {
          cacheLogger.warn('Error clearing cache:', error)
        }
      } catch (_error) {}
    }

    // Map position names to their short codes, more specific for each side
    const fmMatcherToRoleKeyPrefix = {
      GOALKEEPER: 'GK',
      SWEEPER: 'DC',
      'DEFENDER (RIGHT)': 'DR',
      'RIGHT BACK': 'DR',
      'DEFENDER (LEFT)': 'DL',
      'LEFT BACK': 'DL',
      'DEFENDER (CENTRE)': 'DC',
      'CENTRE BACK': 'DC',
      'WING-BACK (RIGHT)': 'WBR',
      'RIGHT WING-BACK': 'WBR',
      'WING-BACK (LEFT)': 'WBL',
      'LEFT WING-BACK': 'WBL',
      'DEFENSIVE MIDFIELDER (CENTRE)': 'DM',
      'CENTRE DEFENSIVE MIDFIELDER': 'DM',
      'MIDFIELDER (RIGHT)': 'MR',
      'RIGHT MIDFIELDER': 'MR',
      'MIDFIELDER (LEFT)': 'ML',
      'LEFT MIDFIELDER': 'ML',
      'MIDFIELDER (CENTRE)': 'MC',
      'CENTRE MIDFIELDER': 'MC',
      'ATTACKING MIDFIELDER (RIGHT)': 'AMR',
      'RIGHT ATTACKING MIDFIELDER': 'AMR',
      'WINGER (RIGHT)': 'AMR',
      'ATTACKING MIDFIELDER (LEFT)': 'AML',
      'LEFT ATTACKING MIDFIELDER': 'AML',
      'WINGER (LEFT)': 'AML',
      'ATTACKING MIDFIELDER (CENTRE)': 'AMC',
      'CENTRE ATTACKING MIDFIELDER': 'AMC',
      'STRIKER (CENTRE)': 'ST',
      STRIKER: 'ST'
    }

    const positionSideMap = {
      'D (R)': ['DR'],
      'D (L)': ['DL'],
      'D (C)': ['DC'],
      'WB (R)': ['WBR'],
      'WB (L)': ['WBL'],
      'DM (C)': ['DM'],
      'M (R)': ['MR'],
      'M (L)': ['ML'],
      'M (C)': ['MC'],
      'AM (R)': ['AMR'],
      'AM (L)': ['AML'],
      'AM (C)': ['AMC'],
      'ST (C)': ['ST'],
      GK: ['GK']
    }

    const fallbackPositionMap = {
      'D (R)': ['DR', 'WBR', 'MR'],
      'D (L)': ['DL', 'WBL', 'ML'],
      'D (C)': ['DC', 'DM'],
      'WB (R)': ['WBR', 'DR', 'MR'],
      'WB (L)': ['WBL', 'DL', 'ML'],
      'DM (C)': ['DM', 'DC', 'MC'],
      'M (R)': ['MR', 'WBR', 'AMR'],
      'M (L)': ['ML', 'WBL', 'AML'],
      'M (C)': ['MC', 'DM'],
      'AM (R)': ['AMR', 'MR'],
      'AM (L)': ['AML', 'ML'],
      'AM (C)': ['AMC', 'MC'],
      'ST (C)': ['ST', 'AMC'],
      GK: ['GK']
    }

    // Reactive refs for pagination
    const showAllNations = ref(false)
    const INITIAL_NATIONS_LIMIT = 50

    // NEW: Replace the heavy computed property with reactive data and async calculation
    const nationsWithRatings = computed(() => {
      // During initial load, show nations without ratings
      if (isCalculatingRatings.value && nationsData.value.length === 0) {
        return []
      }

      // Return sorted nations data
      const sortedNations = [...nationsData.value].sort(
        (a, b) => (b.bestFormationOverall || 0) - (a.bestFormationOverall || 0)
      )

      // Limit initial rendering for performance
      if (!showAllNations.value && sortedNations.length > INITIAL_NATIONS_LIMIT) {
        return sortedNations.slice(0, INITIAL_NATIONS_LIMIT)
      }

      return sortedNations
    })

    // NEW: Initialize nations with basic data (no ratings yet)
    const initializeNationsData = () => {
      if (!allPlayersData.value || allPlayersData.value.length === 0) {
        nationsData.value = []
        return
      }

      const nationsMap = new Map()

      // First pass: collect all players by nationality
      for (const player of allPlayersData.value) {
        if (player.nationality && player.nationality.trim() !== '') {
          const nationality = player.nationality

          if (!nationsMap.has(nationality)) {
            nationsMap.set(nationality, {
              name: nationality,
              nationality_iso: player.nationality_iso || null,
              playerCount: 0,
              bestFormationOverall: null, // Will be calculated async
              isCalculating: false,
              players: [],
              topPlayersByPosition: {} // NEW: Pre-filtered top players per position
            })
          }

          const nation = nationsMap.get(nationality)
          nation.playerCount++
          nation.players.push(player)

          // Set nationality_iso if we don't have it yet
          if (!nation.nationality_iso && player.nationality_iso) {
            nation.nationality_iso = player.nationality_iso
          }
        }
      }

      // NEW: Second pass - Pre-filter to top 10 players per position for each nation
      const allPositions = [
        'GK',
        'DR',
        'DL',
        'DC',
        'WBR',
        'WBL',
        'DM',
        'MR',
        'ML',
        'MC',
        'AMR',
        'AML',
        'AMC',
        'ST'
      ]

      for (const [_nationalityName, nation] of nationsMap) {
        const topPlayersByPosition = {}
        let _totalOptimizedPlayers = 0

        for (const position of allPositions) {
          const playersForPosition = nation.players.filter(player => {
            const playerPositions = player.shortPositions || []
            return playerPositions.includes(position)
          })

          // Sort by Overall and take top 10
          playersForPosition.sort((a, b) => (b.Overall || 0) - (a.Overall || 0))
          const topPlayers = playersForPosition.slice(0, 10)

          if (topPlayers.length > 0) {
            topPlayersByPosition[position] = topPlayers
            _totalOptimizedPlayers += topPlayers.length
          }
        }

        // Replace the full players array with optimized data
        nation.topPlayersByPosition = topPlayersByPosition
      }

      // Convert to array and sort by player count initially
      nationsData.value = Array.from(nationsMap.values()).sort(
        (a, b) => b.playerCount - a.playerCount
      )
    }

    // NEW: Calculate ratings for a single nation
    const calculateNationRatings = async nation => {
      // Set calculating state
      const nationIndex = nationsData.value.findIndex(n => n.name === nation.name)
      if (nationIndex !== -1) {
        nationsData.value[nationIndex].isCalculating = true
      }

      // Simulate small delay to prevent UI blocking
      await new Promise(resolve => setTimeout(resolve, 10))

      // NEW: Use pre-filtered top players instead of filtering during calculation
      const topPlayersByPosition = nation.topPlayersByPosition

      if (!topPlayersByPosition || Object.keys(topPlayersByPosition).length === 0) {
        // Update with zero ratings
        if (nationIndex !== -1) {
          nationsData.value[nationIndex] = {
            ...nationsData.value[nationIndex],
            bestFormationOverall: 0,
            attRating: 0,
            midRating: 0,
            defRating: 0,
            isCalculating: false
          }
        }
        return
      }

      let bestOverall = 0
      let hasMinimumPlayers = false
      let bestSectionRatings = { attRating: 0, midRating: 0, defRating: 0 }

      // Test each formation to find the best average overall for this nation
      for (const formationKey of Object.keys(formations)) {
        const formationLayoutForCalc = getFormationLayout(formationKey)
        if (!formationLayoutForCalc) continue

        const formationSlots = formationLayoutForCalc.flatMap(row => row.positions)
        const tempSquadComposition = {}

        for (const slot of formationSlots) {
          tempSquadComposition[slot.id] = []
        }

        // Calculate player assignments for this formation using pre-filtered top players
        const allPotentialPlayerAssignments = []
        for (const slot of formationSlots) {
          const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
          const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []

          // Get relevant players for this slot (exact matches first, then fallbacks)
          let relevantPlayers = []

          // Add exact position matches first
          for (const position of slotPositions) {
            if (topPlayersByPosition[position]) {
              relevantPlayers = [...relevantPlayers, ...topPlayersByPosition[position]]
            }
          }

          // Add fallback positions if needed
          for (const position of fallbackPositions) {
            if (topPlayersByPosition[position]) {
              for (const player of topPlayersByPosition[position]) {
                // Only add if not already included from exact matches
                if (!relevantPlayers.some(p => p.name === player.name)) {
                  relevantPlayers.push(player)
                }
              }
            }
          }

          for (const player of relevantPlayers) {
            const overallInRole = getPlayerOverallForRole(player, slot.role)

            // Lowered threshold to be more inclusive
            if (overallInRole >= 20) {
              const playerPositions = player.shortPositions || []
              const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))

              const assignment = {
                player,
                slotId: slot.id,
                slotRole: slot.role,
                overallInRole: overallInRole,
                sortScore: overallInRole,
                exactMatch: isExactMatch
              }

              if (isExactMatch) {
                assignment.sortScore += 10000
              } else {
                assignment.sortScore -= 5000
              }

              allPotentialPlayerAssignments.push(assignment)
            }
          }
        }

        allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)
        const assignedPlayersToSlots = new Set()

        // Fill starting XI for this formation
        for (const slot of formationSlots) {
          for (const assignment of allPotentialPlayerAssignments) {
            if (
              assignment.slotId === slot.id &&
              !assignedPlayersToSlots.has(assignment.player.name)
            ) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }

        // Check if we have enough players to generate meaningful ratings
        const filledPositions = Object.values(tempSquadComposition).filter(
          slotPlayers => slotPlayers.length > 0
        ).length
        const hasEnoughPlayers = filledPositions >= 7

        if (hasEnoughPlayers) {
          hasMinimumPlayers = true

          // Calculate average overall for this formation based on filled positions only
          let sumOfStartersOverall = 0
          let startersCount = 0
          for (const slotPlayers of Object.values(tempSquadComposition)) {
            if (slotPlayers && slotPlayers.length > 0) {
              sumOfStartersOverall += slotPlayers[0].overallInRole
              startersCount++
            }
          }

          if (startersCount > 0) {
            const averageOverall = Math.round(sumOfStartersOverall / startersCount)
            if (averageOverall > bestOverall) {
              bestOverall = averageOverall
              // Calculate section ratings directly for this formation
              const defensivePositions = ['GK', 'D (R)', 'D (L)', 'D (C)', 'WB (R)', 'WB (L)']
              const midfielderPositions = ['DM (C)', 'M (R)', 'M (L)', 'M (C)', 'AM (C)']
              const attackingPositions = ['AM (R)', 'AM (L)', 'ST (C)']

              let attSum = 0
              let attCount = 0
              let midSum = 0
              let midCount = 0
              let defSum = 0
              let defCount = 0

              for (const slot of formationSlots) {
                const slotPlayers = tempSquadComposition[slot.id]
                if (slotPlayers && slotPlayers.length > 0) {
                  const starter = slotPlayers[0]
                  const rating = starter.overallInRole

                  if (attackingPositions.includes(slot.role)) {
                    attSum += rating
                    attCount++
                  } else if (midfielderPositions.includes(slot.role)) {
                    midSum += rating
                    midCount++
                  } else if (defensivePositions.includes(slot.role)) {
                    defSum += rating
                    defCount++
                  }
                }
              }

              bestSectionRatings = {
                attRating: attCount > 0 ? Math.round(attSum / attCount) : 0,
                midRating: midCount > 0 ? Math.round(midSum / midCount) : 0,
                defRating: defCount > 0 ? Math.round(defSum / defCount) : 0
              }
            }
          }
        }

        // Reduced delay frequency since we're working with much smaller datasets now
        if (Object.keys(formations).indexOf(formationKey) % 5 === 0) {
          await new Promise(resolve => setTimeout(resolve, 1))
        }
      }

      // Update the nation with calculated ratings
      if (nationIndex !== -1) {
        nationsData.value[nationIndex] = {
          ...nationsData.value[nationIndex],
          bestFormationOverall: hasMinimumPlayers ? bestOverall : 0,
          attRating: hasMinimumPlayers ? bestSectionRatings.attRating : 0,
          midRating: hasMinimumPlayers ? bestSectionRatings.midRating : 0,
          defRating: hasMinimumPlayers ? bestSectionRatings.defRating : 0,
          isCalculating: false
        }
      }
    }

    // NEW: Process calculation queue with caching
    const processCalculationQueue = async () => {
      if (isProcessingQueue.value || calculationQueue.value.length === 0) {
        return
      }

      isProcessingQueue.value = true
      isCalculatingRatings.value = true

      while (calculationQueue.value.length > 0) {
        const nation = calculationQueue.value.shift()
        calculationProgress.value.current =
          calculationProgress.value.total - calculationQueue.value.length

        try {
          await calculateNationRatings(nation)
        } catch (_error) {
          // Mark as failed and continue
          const nationIndex = nationsData.value.findIndex(n => n.name === nation.name)
          if (nationIndex !== -1) {
            nationsData.value[nationIndex].isCalculating = false
            nationsData.value[nationIndex].bestFormationOverall = 0
          }
        }

        // Small delay between nations to keep UI responsive
        await new Promise(resolve => setTimeout(resolve, 5))
      }

      isProcessingQueue.value = false
      isCalculatingRatings.value = false
      calculationProgress.value = { current: 0, total: 0 }
      await saveNationRatingsToCache()
    }

    // NEW: Start rating calculations
    const startRatingCalculations = () => {
      if (nationsData.value.length === 0) return

      // Queue all nations for calculation
      calculationQueue.value = [...nationsData.value]
      calculationProgress.value = { current: 0, total: nationsData.value.length }

      // Start processing
      processCalculationQueue()
    }

    const fetchPlayersAndCurrency = async datasetId => {
      pageLoading.value = true
      pageLoadingError.value = ''
      try {
        await playerStore.fetchPlayersByDatasetId(datasetId)
      } catch (err) {
        pageLoadingError.value = `Failed to load player data: ${err.message || 'Unknown server error'}. Please try uploading again.`
      } finally {
        pageLoading.value = false
      }
    }

    onMounted(async () => {
      const datasetIdFromQuery = route.query.datasetId
      const datasetIdFromRoute = route.params.datasetId
      const nationFromQuery = route.query.nation
      const finalDatasetId =
        datasetIdFromRoute || datasetIdFromQuery || sessionStorage.getItem('currentDatasetId')

      if (finalDatasetId) {
        if (
          datasetIdFromQuery &&
          datasetIdFromQuery !== sessionStorage.getItem('currentDatasetId')
        ) {
          sessionStorage.setItem('currentDatasetId', datasetIdFromQuery)
        } else if (!datasetIdFromQuery && sessionStorage.getItem('currentDatasetId')) {
          router.replace({ query: { datasetId: finalDatasetId } })
        }
        await fetchPlayersAndCurrency(finalDatasetId)
      } else {
        pageLoadingError.value =
          'No player dataset ID found. Please upload a file on the main page.'
        pageLoading.value = false
      }

      if (!pageLoadingError.value && allPlayersData.value.length > 0) {
        populateNationFilterOptions()

        // NEW: Generate cache key and try to load from cache first
        cacheKey.value = generateCacheKey()

        // Try to load cached nation ratings
        const cacheLoaded = await loadNationRatingsFromCache()

        if (!cacheLoaded) {
          initializeNationsData()

          // Start calculating ratings in the background
          setTimeout(() => {
            startRatingCalculations()
          }, 100) // Small delay to let UI render first
        } else {
        }

        if (nationFromQuery && nationFromQuery.trim() !== '') {
          selectedNationName.value = nationFromQuery
          loadNationPlayers()
        } else if (selectedNationName.value) {
          loadNationPlayers()
        }
      }
    })

    const populateNationFilterOptions = () => {
      if (!allPlayersData.value || allPlayersData.value.length === 0) {
        allNationNamesCache.value = []
        nationOptions.value = []
        return
      }
      const uniqueNations = new Set()
      for (const player of allPlayersData.value) {
        if (player.nationality && player.nationality.trim() !== '') {
          uniqueNations.add(player.nationality)
        }
      }
      allNationNamesCache.value = Array.from(uniqueNations).sort()
      nationOptions.value = allNationNamesCache.value
    }

    const filterNationOptions = (val, update) => {
      if (val === '') {
        update(() => {
          nationOptions.value = allNationNamesCache.value
        })
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        nationOptions.value = allNationNamesCache.value.filter(
          nation => nation.toLowerCase().indexOf(needle) > -1
        )
      })
    }

    const selectNation = nationName => {
      selectedNationName.value = nationName
      loadNationPlayers()
    }

    const loadNationPlayers = () => {
      if (!selectedNationName.value) {
        nationPlayers.value = []
        squadComposition.value = {}
        bestNationAverageOverall.value = null
        calculationMessage.value = ''
        selectedFormationKey.value = null
        return
      }
      loadingNation.value = true
      setTimeout(() => {
        if (Array.isArray(allPlayersData.value)) {
          nationPlayers.value = allPlayersData.value.filter(
            p => p.nationality === selectedNationName.value
          )
        } else {
          nationPlayers.value = []
        }

        if (nationPlayers.value.length > 0) {
          const bestFormation = calculateBestFormationForNation()
          if (bestFormation) {
            selectedFormationKey.value = bestFormation
            calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`
            calculationMessageClass.value = quasarInstance.dark.isActive
              ? 'bg-info text-white'
              : 'bg-blue-2 text-primary'
          } else {
            selectedFormationKey.value = null
            squadComposition.value = {}
            bestNationAverageOverall.value = null
            calculationMessage.value = 'No suitable formation found for this nation.'
            calculationMessageClass.value = quasarInstance.dark.isActive
              ? 'text-grey-5'
              : 'text-grey-7'
          }
        } else {
          selectedFormationKey.value = null
          squadComposition.value = {}
          bestNationAverageOverall.value = null
          calculationMessage.value = 'No players found for this nation.'
          calculationMessageClass.value = quasarInstance.dark.isActive
            ? 'text-grey-5'
            : 'text-grey-7'
        }

        loadingNation.value = false
      }, 200)
    }

    const clearNationSelection = () => {
      selectedNationName.value = null
      nationPlayers.value = []
      selectedFormationKey.value = null
      squadComposition.value = {}
      bestNationAverageOverall.value = null
      calculationMessage.value = ''
    }

    const formationOptions = computed(() => {
      return Object.keys(formations).map(key => ({
        label: formations[key].name,
        value: key
      }))
    })

    const currentFormationLayout = computed(() => {
      if (!selectedFormationKey.value) {
        return []
      }
      return getFormationLayout(selectedFormationKey.value) || []
    })

    const bestNationPlayersForPitch = computed(() => {
      const starters = {}
      if (!squadComposition.value || Object.keys(squadComposition.value).length === 0) {
        return starters
      }
      for (const slotId in squadComposition.value) {
        if (squadComposition.value[slotId] && squadComposition.value[slotId].length > 0) {
          const starterEntry = squadComposition.value[slotId][0]
          starters[slotId] = {
            ...starterEntry.player,
            Overall: starterEntry.overallInRole,
            exactPositionMatch: starterEntry.exactMatch
          }
        } else {
          starters[slotId] = null
        }
      }
      return starters
    })

    const currentNationSectionRatings = computed(() => {
      if (!squadComposition.value || !currentFormationLayout.value) {
        return { attRating: 0, midRating: 0, defRating: 0 }
      }
      return calculateSectionRatings(squadComposition.value, currentFormationLayout.value)
    })

    const nationIsGoalkeeperView = computed(() => {
      // Only show goalkeeper view if more than half the players are goalkeepers
      if (nationPlayers.value.length === 0) return false

      const goalkeeperCount = nationPlayers.value.filter(p =>
        p.positionGroups?.includes('Goalkeepers')
      ).length

      // Only show goalkeeper view if more than half the players are goalkeepers
      return goalkeeperCount > nationPlayers.value.length / 2
    })

    const handlePlayerSelectedFromNation = player => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const getOverallClass = overall => {
      if (overall === null || overall === undefined || overall === 0) return 'rating-na'
      const numericOverall = Number(overall)
      if (Number.isNaN(numericOverall)) return 'rating-na'

      if (numericOverall >= 90) return 'rating-tier-6'
      if (numericOverall >= 80) return 'rating-tier-5'
      if (numericOverall >= 70) return 'rating-tier-4'
      if (numericOverall >= 55) return 'rating-tier-3'
      if (numericOverall >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    const getStarClass = (overall, starPosition) => {
      if (!overall || overall === 0) return 'empty'

      const starRating = getStarRating(overall)

      if (starPosition <= Math.floor(starRating)) {
        return 'filled'
      }
      if (starPosition === Math.floor(starRating) + 1 && starRating % 1 === 0.5) {
        return 'half'
      }
      return 'empty'
    }

    const getStarRating = overall => {
      if (!overall || overall === 0) return 0

      if (overall >= 85) return 5
      if (overall >= 82) return 4.5
      if (overall >= 78) return 4
      if (overall >= 74) return 3.5
      if (overall >= 70) return 3
      if (overall >= 67) return 2.5
      if (overall >= 64) return 2
      if (overall >= 60) return 1.5
      if (overall >= 55) return 1
      if (overall >= 50) return 0.5
      return 0
    }

    const calculateSectionRatings = (squadComposition, formationLayout) => {
      if (!squadComposition || !formationLayout) {
        return { attRating: 0, midRating: 0, defRating: 0 }
      }

      const formationSlots = formationLayout.flatMap(row => row.positions)

      // Define position categories
      const defensivePositions = ['GK', 'D (R)', 'D (L)', 'D (C)', 'WB (R)', 'WB (L)']
      const midfielderPositions = ['DM (C)', 'M (R)', 'M (L)', 'M (C)', 'AM (C)']
      const attackingPositions = ['AM (R)', 'AM (L)', 'ST (C)']

      let attSum = 0
      let attCount = 0
      let midSum = 0
      let midCount = 0
      let defSum = 0
      let defCount = 0

      for (const slot of formationSlots) {
        const slotPlayers = squadComposition[slot.id]
        if (slotPlayers && slotPlayers.length > 0) {
          const starter = slotPlayers[0]
          const rating = starter.overallInRole

          if (attackingPositions.includes(slot.role)) {
            attSum += rating
            attCount++
          } else if (midfielderPositions.includes(slot.role)) {
            midSum += rating
            midCount++
          } else if (defensivePositions.includes(slot.role)) {
            defSum += rating
            defCount++
          }
        }
      }

      return {
        attRating: attCount > 0 ? Math.round(attSum / attCount) : 0,
        midRating: midCount > 0 ? Math.round(midSum / midCount) : 0,
        defRating: defCount > 0 ? Math.round(defSum / defCount) : 0
      }
    }

    const getPlayerOverallForRole = (player, slotFormationRole) => {
      if (!player || !slotFormationRole) return 0

      let bestScoreForRole = 0

      if (!player.roleSpecificOveralls) {
        // If no role-specific overalls, use player's general Overall as fallback
        return player.Overall || 0
      }

      const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls)
        ? player.roleSpecificOveralls.length > 0
        : Object.keys(player.roleSpecificOveralls).length > 0

      if (!hasRoleOveralls) {
        // If no role-specific overalls, use player's general Overall as fallback
        return player.Overall || 0
      }

      const upperSlotRoleOriginal = slotFormationRole.toUpperCase()
      const requiredPositions = positionSideMap[upperSlotRoleOriginal] || []

      if (player.shortPositions && player.shortPositions.length > 0) {
        const exactPositionMatches = player.shortPositions.filter(pos =>
          requiredPositions.includes(pos)
        )

        if (exactPositionMatches.length > 0) {
          if (Array.isArray(player.roleSpecificOveralls)) {
            for (const rso of player.roleSpecificOveralls) {
              const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

              if (exactPositionMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, rso.score)
              }
            }
          } else {
            for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
              const rsoBasePosition = roleName.split(' - ')[0].trim()

              if (exactPositionMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, score)
              }
            }
          }

          if (bestScoreForRole === 0) {
            bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD, player.Overall || 0)
          }
        }
      }

      if (bestScoreForRole > 0) {
        return bestScoreForRole
      }

      const fallbackPositions = fallbackPositionMap[upperSlotRoleOriginal] || []

      if (player.shortPositions && player.shortPositions.length > 0) {
        const fallbackMatches = player.shortPositions.filter(pos => fallbackPositions.includes(pos))

        if (fallbackMatches.length > 0) {
          if (Array.isArray(player.roleSpecificOveralls)) {
            for (const rso of player.roleSpecificOveralls) {
              const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

              if (fallbackMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, rso.score)
              }
            }
          } else {
            for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
              const rsoBasePosition = roleName.split(' - ')[0].trim()

              if (fallbackMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, score)
              }
            }
          }

          if (bestScoreForRole === 0) {
            bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD - 10, (player.Overall || 0) - 5)
          }
        }
      }

      if (bestScoreForRole === 0) {
        const upperSlotRole = slotFormationRole.toUpperCase()
        const fmPositionMatchers = fmSlotRoleMatcher[upperSlotRole] || [upperSlotRole]

        const targetRoleKeyPrefixes = fmPositionMatchers
          .map(matcher => fmMatcherToRoleKeyPrefix[matcher.toUpperCase()])
          .filter(prefix => !!prefix)
          .reduce((acc, val) => {
            if (!acc.includes(val)) {
              acc.push(val)
            }
            return acc
          }, [])

        if (Array.isArray(player.roleSpecificOveralls)) {
          for (const rso of player.roleSpecificOveralls) {
            const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

            if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
              bestScoreForRole = Math.max(bestScoreForRole, rso.score)
            }
          }
        } else if (player.roleSpecificOveralls) {
          for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
            const rsoBasePosition = roleName.split(' - ')[0].trim()

            if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
              bestScoreForRole = Math.max(bestScoreForRole, score)
            }
          }
        }

        // Final fallback to player's general Overall rating if still nothing found
        if (bestScoreForRole === 0) {
          bestScoreForRole = Math.max(0, (player.Overall || 0) - 10)
        }
      }

      return bestScoreForRole
    }

    const MIN_SUITABILITY_THRESHOLD = 20

    const getSlotDisplayName = (slot, allSlots) => {
      const roleCounts = allSlots.reduce((acc, s) => {
        acc[s.role] = (acc[s.role] || 0) + 1
        return acc
      }, {})

      if (roleCounts[slot.role] > 1) {
        return slot.id.split('_')[0]
      }
      return slot.role
    }

    const calculateBestFormationForNation = () => {
      if (nationPlayers.value.length === 0) {
        return null
      }

      // Check cache first
      const cacheKey = formationCache.generateKey(nationPlayers.value, 'nation-best')
      const cachedResult = formationCache.get(cacheKey)
      if (cachedResult) {
        return cachedResult.bestFormationKey
      }

      let bestFormationKey = null
      let bestAverageOverall = 0

      for (const formationKey of Object.keys(formations)) {
        const formationLayoutForCalc = getFormationLayout(formationKey)
        if (!formationLayoutForCalc) continue

        const formationSlots = formationLayoutForCalc.flatMap(row => row.positions)
        const tempSquadComposition = {}

        for (const slot of formationSlots) {
          tempSquadComposition[slot.id] = []
        }

        const allPotentialPlayerAssignments = []
        for (const slot of formationSlots) {
          for (const player of nationPlayers.value) {
            const overallInRole = getPlayerOverallForRole(player, slot.role)

            if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
              const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
              const playerPositions = player.shortPositions || []
              const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))

              const canPlayInPosition = isExactMatch

              if (canPlayInPosition && overallInRole >= MIN_SUITABILITY_THRESHOLD) {
                const assignment = {
                  player,
                  slotId: slot.id,
                  slotRole: slot.role,
                  overallInRole: overallInRole,
                  sortScore: overallInRole,
                  exactMatch: isExactMatch
                }

                if (isExactMatch) {
                  assignment.sortScore += 10000
                } else {
                  assignment.sortScore -= 5000
                }

                allPotentialPlayerAssignments.push(assignment)
              }
            }
          }
        }

        allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)

        const assignedPlayersToSlots = new Set()

        for (const slot of formationSlots) {
          for (const assignment of allPotentialPlayerAssignments) {
            if (
              assignment.slotId === slot.id &&
              !assignedPlayersToSlots.has(assignment.player.name)
            ) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }

        let sumOfStartersOverall = 0
        let startersCount = 0
        for (const slotPlayers of Object.values(tempSquadComposition)) {
          if (slotPlayers && slotPlayers.length > 0) {
            sumOfStartersOverall += slotPlayers[0].overallInRole
            startersCount++
          }
        }

        if (startersCount > 0) {
          const averageOverall = sumOfStartersOverall / startersCount
          if (averageOverall > bestAverageOverall) {
            bestAverageOverall = averageOverall
            bestFormationKey = formationKey
          }
        }
      }

      // Cache the result
      if (bestFormationKey) {
        formationCache.set(cacheKey, {
          bestFormationKey,
          bestAverageOverall,
          nationName: selectedNationName.value
        })
      }

      return bestFormationKey
    }

    const calculateBestNationAndDepth = () => {
      if (!selectedFormationKey.value || nationPlayers.value.length === 0) {
        squadComposition.value = {}
        bestNationAverageOverall.value = null
        calculationMessage.value = selectedFormationKey.value
          ? 'No players in the selected nation.'
          : 'Select a formation.'
        calculationMessageClass.value = 'bg-warning text-dark'
        return
      }

      calculationMessage.value = 'Calculating best nation team and depth...'
      calculationMessageClass.value = quasarInstance.dark.isActive
        ? 'bg-info text-white'
        : 'bg-blue-2 text-primary'

      const tempSquadComposition = {}
      const formationLayoutForCalc = getFormationLayout(selectedFormationKey.value)
      if (!formationLayoutForCalc) {
        calculationMessage.value = 'Invalid formation selected.'
        calculationMessageClass.value = 'bg-negative text-white'
        return
      }

      const formationSlots = formationLayoutForCalc.flatMap(row => row.positions)

      for (const slot of formationSlots) {
        tempSquadComposition[slot.id] = []
      }

      // NEW: For the detailed squad calculation, we still use all nation players
      // since we need full depth charts and the user expects to see all their players
      // This is only for the selected nation, so performance impact is minimal
      const allPotentialPlayerAssignments = []
      for (const slot of formationSlots) {
        for (const player of nationPlayers.value) {
          const overallInRole = getPlayerOverallForRole(player, slot.role)

          if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
            const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
            const playerPositions = player.shortPositions || []
            const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))

            const canPlayInPosition = isExactMatch

            if (canPlayInPosition && overallInRole >= MIN_SUITABILITY_THRESHOLD) {
              const assignment = {
                player,
                slotId: slot.id,
                slotRole: slot.role,
                overallInRole: overallInRole,
                sortScore: overallInRole,
                exactMatch: isExactMatch
              }

              if (isExactMatch) {
                assignment.sortScore += 10000
              } else {
                assignment.sortScore -= 5000
              }

              allPotentialPlayerAssignments.push(assignment)
            }
          }
        }
      }

      allPotentialPlayerAssignments.sort((a, b) => {
        return b.sortScore - a.sortScore
      })

      const assignedPlayersToSlots = new Set()

      for (let depthIndex = 0; depthIndex < 3; depthIndex++) {
        for (const slot of formationSlots) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                assignment.exactMatch &&
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  for (const sId in tempSquadComposition) {
                    if (
                      tempSquadComposition[sId].length > 0 &&
                      tempSquadComposition[sId][0].player.name === assignment.player.name
                    ) {
                      alreadyStarterElsewhere = true
                      break
                    }
                  }
                }

                if (!alreadyStarterElsewhere) {
                  tempSquadComposition[slot.id].push({
                    player: assignment.player,
                    overallInRole: assignment.overallInRole,
                    exactMatch: assignment.exactMatch
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break
                }
              }
            }
          }
        }

        for (const slot of formationSlots) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  for (const sId in tempSquadComposition) {
                    if (
                      tempSquadComposition[sId].length > 0 &&
                      tempSquadComposition[sId][0].player.name === assignment.player.name
                    ) {
                      alreadyStarterElsewhere = true
                      break
                    }
                  }
                }

                if (!alreadyStarterElsewhere) {
                  tempSquadComposition[slot.id].push({
                    player: assignment.player,
                    overallInRole: assignment.overallInRole,
                    exactMatch: assignment.exactMatch
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break
                }
              }
            }
          }
        }
      }

      for (const slotId in tempSquadComposition) {
        tempSquadComposition[slotId].sort((a, b) => b.overallInRole - a.overallInRole)
      }

      for (const slot of formationSlots) {
        if (tempSquadComposition[slot.id].length === 0) {
          const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []

          const fallbackAssignments = []

          for (const player of nationPlayers.value) {
            if (!assignedPlayersToSlots.has(player.name)) {
              const playerPositions = player.shortPositions || []

              const canPlayFallback = playerPositions.some(pos => fallbackPositions.includes(pos))

              if (canPlayFallback) {
                const overallInRole = getPlayerOverallForRole(player, slot.role)
                if (overallInRole >= MIN_SUITABILITY_THRESHOLD - 10) {
                  fallbackAssignments.push({
                    player,
                    overallInRole,
                    exactMatch: false
                  })
                }
              }
            }
          }

          fallbackAssignments.sort((a, b) => b.overallInRole - a.overallInRole)

          if (fallbackAssignments.length > 0) {
            const bestFallback = fallbackAssignments[0]
            tempSquadComposition[slot.id].push(bestFallback)
            assignedPlayersToSlots.add(bestFallback.player.name)
          }
        }
      }

      squadComposition.value = tempSquadComposition

      let sumOfStartersOverall = 0
      let startersCount = 0
      for (const slotPlayers of Object.values(squadComposition.value)) {
        if (slotPlayers && slotPlayers.length > 0) {
          sumOfStartersOverall += slotPlayers[0].overallInRole
          startersCount++
        }
      }

      if (startersCount > 0) {
        bestNationAverageOverall.value = Math.round(sumOfStartersOverall / startersCount)
        calculationMessage.value = `Best XI & Depth calculated. Average Overall: ${bestNationAverageOverall.value}.`
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'
      } else {
        bestNationAverageOverall.value = 0
        calculationMessage.value = 'Could not assign any suitable players to form a Best XI.'
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-negative text-white'
          : 'bg-red-2 text-negative'
      }
    }

    watch(selectedFormationKey, newKey => {
      if (newKey && selectedNationName.value) {
        calculateBestNationAndDepth()
      } else {
        squadComposition.value = {}
        bestNationAverageOverall.value = null
        calculationMessage.value = 'Select a nation and formation.'
        calculationMessageClass.value = quasarInstance.dark.isActive ? 'text-grey-5' : 'text-grey-7'
      }
    })

    const handlePlayerMovedOnPitch = moveData => {
      const { player, fromSlotId, toSlotId, toSlotRole } = moveData

      const currentStarters = JSON.parse(JSON.stringify(bestNationPlayersForPitch.value))
      const playerToMoveFullData = allPlayersData.value.find(p => p.name === player.name)

      if (!playerToMoveFullData) return

      const overallInNewRole = getPlayerOverallForRole(playerToMoveFullData, toSlotRole)

      const playerPositions = playerToMoveFullData.shortPositions || []
      const slotPositions = positionSideMap[toSlotRole.toUpperCase()] || []
      const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))

      const playerCurrentlyInTargetSlotFullData = currentStarters[toSlotId]
        ? allPlayersData.value.find(p => p.name === currentStarters[toSlotId].name)
        : null

      currentStarters[toSlotId] = {
        ...playerToMoveFullData,
        Overall: overallInNewRole,
        exactPositionMatch: isExactMatch
      }

      if (playerCurrentlyInTargetSlotFullData && fromSlotId) {
        const originalRoleOfFromSlot = currentFormationLayout.value
          .flatMap(r => r.positions)
          .find(p => p.id === fromSlotId)?.role

        if (originalRoleOfFromSlot) {
          const overallInOldRole = getPlayerOverallForRole(
            playerCurrentlyInTargetSlotFullData,
            originalRoleOfFromSlot
          )

          const playerPositions = playerCurrentlyInTargetSlotFullData.shortPositions || []
          const slotPositions = positionSideMap[originalRoleOfFromSlot.toUpperCase()] || []
          const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))

          currentStarters[fromSlotId] = {
            ...playerCurrentlyInTargetSlotFullData,
            Overall: overallInOldRole,
            exactPositionMatch: isExactMatch
          }
        } else {
          currentStarters[fromSlotId] = null
        }
      } else if (fromSlotId) {
        currentStarters[fromSlotId] = null
      }

      const newPitchState = { ...currentStarters }

      let sumOfDisplayedOveralls = 0
      let countOfDisplayedOveralls = 0
      for (const p of Object.values(newPitchState)) {
        if (p && typeof p.Overall === 'number') {
          sumOfDisplayedOveralls += p.Overall
          countOfDisplayedOveralls++
        }
      }
      bestNationAverageOverall.value =
        countOfDisplayedOveralls > 0
          ? Math.round(sumOfDisplayedOveralls / countOfDisplayedOveralls)
          : 0

      calculationMessage.value = `Nation team visually adjusted. New Avg Overall: ${bestNationAverageOverall.value}. (Depth chart not updated by drag & drop).`
      calculationMessageClass.value = quasarInstance.dark.isActive
        ? 'bg-info text-white'
        : 'bg-blue-2 text-primary'
    }

    watch(
      () => allPlayersData.value,
      async newVal => {
        if (pageLoading.value) return
        if (newVal && newVal.length > 0) {
          populateNationFilterOptions()

          // NEW: Generate new cache key and try to load from cache
          cacheKey.value = generateCacheKey()

          const cacheLoaded = await loadNationRatingsFromCache()

          if (!cacheLoaded) {
            // NEW: Initialize nations and start calculations
            initializeNationsData()
            setTimeout(() => {
              startRatingCalculations()
            }, 100)
          }

          if (selectedNationName.value) loadNationPlayers()
        } else if (!pageLoadingError.value) {
          clearNationSelection()
          allNationNamesCache.value = []
          nationOptions.value = []
          nationsData.value = [] // NEW: Clear nations data
          cacheKey.value = null // NEW: Clear cache key
        }
      },
      { immediate: false }
    )

    watch(
      () => route.query.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          sessionStorage.setItem('currentDatasetId', newId)
          await fetchPlayersAndCurrency(newId)
          clearNationSelection()
          nationsData.value = [] // NEW: Clear nations data
          cacheKey.value = null // NEW: Clear cache key
          if (!pageLoadingError.value && allPlayersData.value.length > 0) {
            populateNationFilterOptions()

            // NEW: Generate cache key and try to load from cache
            cacheKey.value = generateCacheKey()
            const cacheLoaded = await loadNationRatingsFromCache()

            if (!cacheLoaded) {
              initializeNationsData() // NEW: Initialize nations
              setTimeout(() => {
                startRatingCalculations() // NEW: Start calculations
              }, 100)
            }
          }
        }
      }
    )

    watch(
      () => route.query.nation,
      newNation => {
        if (newNation && newNation.trim() !== '' && newNation !== selectedNationName.value) {
          selectedNationName.value = newNation
          loadNationPlayers()
        }
      }
    )

    const onFlagError = (event, _nation) => {
      // Hide the broken image and fallback will show the icon
      event.target.style.display = 'none'
    }

    const shareDataset = async () => {
      if (!currentDatasetId.value) return

      const shareUrl = `${window.location.origin}/nations/${currentDatasetId.value}`

      try {
        await navigator.clipboard.writeText(shareUrl)
        quasarInstance.notify({
          message: 'Link copied to clipboard!',
          color: 'positive',
          icon: 'check_circle',
          position: 'top',
          timeout: 2000
        })
      } catch (_err) {
        const textArea = document.createElement('textarea')
        textArea.value = shareUrl
        document.body.appendChild(textArea)
        textArea.select()
        document.execCommand('copy')
        document.body.removeChild(textArea)

        quasarInstance.notify({
          message: 'Link copied to clipboard!',
          color: 'positive',
          icon: 'check_circle',
          position: 'top',
          timeout: 2000
        })
      }
    }

    // Add computed property for current nation flag ISO
    const currentNationFlagISO = computed(() => {
      if (!selectedNationName.value) return null
      const nation = nationsWithRatings.value.find(n => n.name === selectedNationName.value)
      return nation?.nationality_iso || null
    })

    // Add handleTeamSelected function (for compatibility with PlayerDataTable)
    const handleTeamSelected = teamName => {
      // Redirect to team view page when a team is selected from the player table
      router.push(`/team-view/${currentDatasetId.value}?team=${encodeURIComponent(teamName)}`)
    }

    return {
      allPlayersData,
      selectedNationName,
      nationOptions,
      filterNationOptions,
      loadNationPlayers,
      clearNationSelection,
      selectNation,
      nationPlayers,
      loadingNation,
      pageLoading,
      pageLoadingError,
      selectedFormationKey,
      formationOptions,
      currentFormationLayout,
      squadComposition,
      bestNationPlayersForPitch,
      bestNationAverageOverall,
      currentNationSectionRatings,
      calculationMessage,
      calculationMessageClass,
      playerForDetailView,
      showPlayerDetailDialog,
      handlePlayerSelectedFromNation,
      nationIsGoalkeeperView,
      getOverallClass,
      getStarClass,
      getStarRating,
      calculateSectionRatings,
      getSlotDisplayName,
      handlePlayerMovedOnPitch,
      quasarInstance,
      router,
      detectedCurrencySymbol,
      currentDatasetId,
      shareDataset,
      onFlagError,
      nationsWithRatings,
      showAllNations,
      currentNationFlagISO,
      handleTeamSelected,
      // NEW: Add new reactive properties
      nationsData,
      isCalculatingRatings,
      calculationProgress,
      isProcessingQueue,
      cacheLoading,
      cacheKey
    }
  }
}
</script>

<style lang="scss" scoped>
// Variables
$border-radius: 12px;
$border-radius-small: 8px;
$card-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
$card-shadow-hover: 0 4px 20px rgba(0, 0, 0, 0.12);

// Base Page Layout
.nation-view-page {
    min-height: 100vh;
    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
    
    .body--dark & {
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
}

.main-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
    
    @media (max-width: 768px) {
        padding: 1rem;
    }
}

// Banners
.error-banner {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

.info-banner {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

.warning-banner {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

// Share Section
.share-section {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 2rem;
    
    .share-btn-modern {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        border: none;
        box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        transition: all 0.3s ease;
        
        &:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
        }
    }
}

// Empty State
.empty-state {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 60vh;
    
    .empty-state-card {
        background: white;
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        border: 1px solid rgba(0, 0, 0, 0.06);
        max-width: 500px;
        width: 100%;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border-color: rgba(255, 255, 255, 0.1);
        }
        
        .empty-state-content {
            text-align: center;
            padding: 3rem;
            
            .empty-state-icon {
                margin-bottom: 2rem;
                color: #667eea;
                opacity: 0.7;
            }
            
            .empty-state-title {
                font-size: 1.75rem;
                font-weight: 700;
                margin: 0 0 1rem 0;
                color: #1e293b;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .empty-state-description {
                font-size: 1rem;
                color: #64748b;
                line-height: 1.6;
                margin: 0 0 2rem 0;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
            
            .empty-state-btn {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                border: none;
                padding: 0.75rem 2rem;
                font-weight: 600;
            }
        }
    }
}

// Loading States
.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 40vh;
    gap: 1.5rem;
    
    .loading-text {
        font-size: 1.1rem;
        color: #64748b;
        font-weight: 500;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
    }
}

// Nation Dashboard
.nation-dashboard {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

// Nation Hero Section
.nation-hero-section {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: $border-radius;
    padding: 3rem;
    color: white;
    position: relative;
    overflow: hidden;
    box-shadow: $card-shadow;
    
    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, transparent 100%);
        pointer-events: none;
    }
    
    .nation-hero-content {
        position: relative;
        z-index: 1;
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 2rem;
        
        @media (max-width: 1200px) {
            flex-direction: column;
            gap: 2rem;
        }
    }
    
    .nation-primary-info {
        display: flex;
        align-items: center;
        gap: 2rem;
        flex: 1;
        
        @media (max-width: 768px) {
            flex-direction: column;
            align-items: flex-start;
            gap: 1rem;
        }
    }
    
    .nation-name-section {
        .nation-name-hero {
            font-size: 3rem;
            font-weight: 800;
            margin: 0 0 0.5rem 0;
            line-height: 1.1;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            
            @media (max-width: 768px) {
                font-size: 2.2rem;
            }
        }
        
        .nation-flag-hero {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-size: 1.1rem;
            font-weight: 600;
            opacity: 0.9;
            
            .nationality-flag {
                border-radius: 4px;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
            }
        }
    }
    
    .star-rating-display {
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        background: rgba(255, 255, 255, 0.1);
        border-radius: $border-radius-small;
        padding: 1.5rem;
        backdrop-filter: blur(10px);
        
        .overall-score {
            font-size: 2.5rem;
            font-weight: 800;
            margin-bottom: 0.5rem;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        
        .star-container {
            margin-bottom: 0.5rem;
            
            .star-modern {
                font-size: 1.5rem;
                margin: 0 0.1rem;
                transition: all 0.2s ease;
                
                &.filled {
                    color: #fbbf24;
                    text-shadow: 0 0 10px rgba(251, 191, 36, 0.5);
                }
                
                &.half {
                    background: linear-gradient(90deg, #fbbf24 50%, rgba(255, 255, 255, 0.3) 50%);
                    -webkit-background-clip: text;
                    -webkit-text-fill-color: transparent;
                    background-clip: text;
                }
                
                &.empty {
                    color: rgba(255, 255, 255, 0.3);
                }
            }
        }
        
        .rating-label {
            font-size: 0.9rem;
            font-weight: 600;
            opacity: 0.8;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
    }
    
    .performance-metrics {
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 1rem;
            
            @media (max-width: 768px) {
                grid-template-columns: 1fr;
                gap: 0.75rem;
            }
            
            .metric-card {
                background: rgba(255, 255, 255, 0.1);
                border-radius: $border-radius-small;
                padding: 1.5rem;
                text-align: center;
                backdrop-filter: blur(10px);
                border: 1px solid rgba(255, 255, 255, 0.2);
                transition: transform 0.2s ease;
                
                &:hover {
                    transform: translateY(-2px);
                }
                
                .metric-icon {
                    font-size: 2rem;
                    margin-bottom: 0.5rem;
                }
                
                .metric-value {
                    font-size: 1.8rem;
                    font-weight: 800;
                    margin-bottom: 0.25rem;
                    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
                }
                
                .metric-label {
                    font-size: 0.9rem;
                    font-weight: 600;
                    opacity: 0.9;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                }
            }
        }
    }
}

// Formation & Tactics Layout
.formation-tactics-layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
    
    @media (max-width: 1200px) {
        grid-template-columns: 1fr;
        gap: 1.5rem;
    }
}

.formation-controls-panel {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

// Card Styles
.formation-card,
.squad-depth-card,
.pitch-card,
.players-table-card,
.nations-overview-card {
    background: white;
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    border: 1px solid rgba(0, 0, 0, 0.06);
    transition: box-shadow 0.3s ease;
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.1);
    }
    
    &:hover {
        box-shadow: $card-shadow-hover;
    }
}

.card-header {
    margin-bottom: 1.5rem;
    
    .card-title {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        font-size: 1.3rem;
        font-weight: 700;
        margin: 0 0 0.5rem 0;
        color: #1e293b;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.9);
        }
        
        .card-icon {
            color: #667eea;
            font-size: 1.5rem;
        }
    }
    
    .card-subtitle {
        font-size: 0.95rem;
        color: #64748b;
        margin: 0;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
    }
}

// Formation Controls
.formation-controls {
    .formation-select {
        margin-bottom: 1rem;
        
        .q-field__control {
            border-radius: $border-radius-small;
        }
    }
    
    .calculation-banner {
        border-radius: $border-radius-small;
        border: none;
        font-weight: 500;
    }
}

// Squad Depth Grid
.squad-depth-card {
    .squad-depth-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 0.75rem;
        
        @media (max-width: 768px) {
            grid-template-columns: 1fr;
            gap: 0.75rem;
        }
        
        .depth-position-modern {
            background: rgba(255, 255, 255, 0.6);
            border-radius: $border-radius-small;
            padding: 0.75rem;
            border: 1px solid rgba(0, 0, 0, 0.08);
            transition: all 0.3s ease;
            
            .body--dark & {
                background: rgba(255, 255, 255, 0.05);
                border-color: rgba(255, 255, 255, 0.1);
            }
            
            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
            }
            
            .position-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 0.5rem;
                
                .position-name {
                    font-weight: 700;
                    font-size: 0.8rem;
                    color: #2d3436;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.9);
                    }
                }
                
                .player-count {
                    font-size: 0.7rem;
                    color: #636e72;
                    background: rgba(0, 0, 0, 0.05);
                    padding: 0.15rem 0.4rem;
                    border-radius: 10px;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.7);
                        background: rgba(255, 255, 255, 0.1);
                    }
                }
            }
            
            .depth-players-modern {
                display: flex;
                flex-direction: column;
                gap: 0.4rem;
                
                .player-card-mini {
                    display: grid;
                    grid-template-columns: auto 1fr auto;
                    gap: 0.4rem;
                    align-items: center;
                    padding: 0.4rem;
                    background: rgba(255, 255, 255, 0.8);
                    border-radius: $border-radius-small;
                    cursor: pointer;
                    transition: all 0.2s ease;
                    border: 1px solid rgba(0, 0, 0, 0.05);
                    
                    .body--dark & {
                        background: rgba(255, 255, 255, 0.08);
                        border-color: rgba(255, 255, 255, 0.1);
                    }
                    
                    &.is-starter {
                        background: linear-gradient(135deg, rgba(0, 184, 148, 0.1) 0%, rgba(0, 206, 201, 0.05) 100%);
                        border-color: rgba(0, 184, 148, 0.2);
                        font-weight: 600;
                    }
                    
                    &:hover {
                        transform: translateX(2px);
                        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                        background: rgba(103, 126, 234, 0.1);
                        
                        .body--dark & {
                            background: rgba(103, 126, 234, 0.15);
                        }
                    }
                    
                    .player-rank {
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        width: 16px;
                        height: 16px;
                        background: rgba(103, 126, 234, 0.1);
                        color: #667eea;
                        border-radius: 50%;
                        font-size: 0.65rem;
                        font-weight: 700;
                    }
                    
                    .player-info {
                        min-width: 0;
                        
                        .player-name {
                            font-size: 0.75rem;
                            font-weight: 600;
                            white-space: nowrap;
                            overflow: hidden;
                            text-overflow: ellipsis;
                            color: #2d3436;
                            
                            .body--dark & {
                                color: rgba(255, 255, 255, 0.9);
                            }
                        }
                        
                        .player-positions {
                            font-size: 0.6rem;
                            color: #636e72;
                            margin-top: 0.1rem;
                            
                            .body--dark & {
                                color: rgba(255, 255, 255, 0.6);
                            }
                        }
                    }
                    
                    .player-rating {
                        font-size: 0.7rem;
                        font-weight: 700;
                        padding: 0.15rem 0.3rem;
                        border-radius: 3px;
                        min-width: 24px;
                        text-align: center;
                        border: 1px solid rgba(0, 0, 0, 0.1);
                        
                        .body--dark & {
                            border-color: rgba(255, 255, 255, 0.2);
                        }
                    }
                }
            }
            
            .no-players-state {
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 0.2rem;
                padding: 0.75rem;
                color: #636e72;
                font-style: italic;
                font-size: 0.75rem;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.5);
                }
                
                .q-icon {
                    font-size: 1rem;
                }
            }
        }
    }
}

// Pitch Container
.pitch-container {
    background: linear-gradient(135deg, #00b894 0%, #00cec9 100%);
    border-radius: $border-radius;
    padding: 2rem;
    margin-top: 1rem;
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
            linear-gradient(90deg, rgba(255, 255, 255, 0.1) 50%, transparent 50%),
            linear-gradient(0deg, rgba(255, 255, 255, 0.1) 50%, transparent 50%);
        background-size: 20px 20px;
        opacity: 0.3;
    }
}

// Table Container
.table-container {
    .modern-table {
        border-radius: $border-radius-small;
        overflow: hidden;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    
    .no-players-banner {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        padding: 3rem 2rem;
        text-align: center;
        color: #636e72;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
        
        h4 {
            margin: 0;
            font-size: 1.5rem;
            font-weight: 600;
        }
        
        p {
            margin: 0;
            font-size: 1rem;
        }
    }
}

// Nations Overview Card specific styles
.nations-overview-card {
    .modern-filter-section {
        margin-bottom: 2rem;
        
        .nation-select {
            .q-field__control {
                border-radius: $border-radius-small;
            }
        }
    }
    
    // NEW: Calculation Progress Section
    .calculation-progress-section {
        margin-bottom: 2rem;
        padding: 1.5rem;
        background: linear-gradient(135deg, rgba(103, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.05) 100%);
        border-radius: $border-radius-small;
        border: 1px solid rgba(103, 126, 234, 0.2);
        
        .body--dark & {
            background: linear-gradient(135deg, rgba(103, 126, 234, 0.15) 0%, rgba(118, 75, 162, 0.1) 100%);
            border-color: rgba(103, 126, 234, 0.3);
        }
        
        .progress-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
            
            .progress-title {
                display: flex;
                align-items: center;
                font-weight: 600;
                color: #667eea;
                font-size: 1rem;
                
                .body--dark & {
                    color: rgba(103, 126, 234, 0.9);
                }
            }
            
            .progress-stats {
                font-size: 0.9rem;
                font-weight: 500;
                color: #64748b;
                background: rgba(255, 255, 255, 0.7);
                padding: 0.25rem 0.75rem;
                border-radius: 15px;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.8);
                    background: rgba(255, 255, 255, 0.1);
                }
            }
        }
        
        .progress-bar {
            height: 8px;
            border-radius: 4px;
            margin-bottom: 0.75rem;
            
            &.progress-bar-dark {
                .q-linear-progress__track {
                    background: rgba(255, 255, 255, 0.1);
                }
            }
        }
        
        .progress-description {
            font-size: 0.85rem;
            color: #64748b;
            text-align: center;
            font-style: italic;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
    
    .nations-list {
        .nation-row {
            display: grid;
            grid-template-columns: auto 1fr auto auto;
            gap: 1rem;
            align-items: center;
            padding: 1rem;
            border-radius: $border-radius-small;
            cursor: pointer;
            transition: all 0.2s ease;
            border: 1px solid rgba(0, 0, 0, 0.06);
            margin-bottom: 0.5rem;
            
            .body--dark & {
                border-color: rgba(255, 255, 255, 0.1);
            }
            
            &.calculating {
                background: linear-gradient(135deg, rgba(103, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.04) 100%);
                border-color: rgba(103, 126, 234, 0.2);
                
                .body--dark & {
                    background: linear-gradient(135deg, rgba(103, 126, 234, 0.12) 0%, rgba(118, 75, 162, 0.08) 100%);
                    border-color: rgba(103, 126, 234, 0.3);
                }
            }
            
            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
                background: rgba(103, 126, 234, 0.05);
                
                .body--dark & {
                    background: rgba(103, 126, 234, 0.1);
                }
            }
            
            .nation-flag-container {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 32px;
                
                .nationality-flag {
                    border-radius: 4px;
                    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
                }
            }
            
            .nation-info {
                .nation-name {
                    font-weight: 700;
                    font-size: 1rem;
                    color: #1e293b;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.9);
                    }
                }
                
                .player-count {
                    font-size: 0.85rem;
                    color: #64748b;
                    margin-top: 0.2rem;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.6);
                    }
                }
            }
            
            .nation-section-ratings {
                // NEW: Calculating state styles
                .calculating-ratings {
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                    
                    .calculating-text {
                        font-size: 0.85rem;
                        color: #667eea;
                        font-weight: 500;
                        
                        .body--dark & {
                            color: rgba(103, 126, 234, 0.9);
                        }
                    }
                }
                
                // NEW: Pending state styles
                .pending-calculation {
                    display: flex;
                    align-items: center;
                    color: #9ca3af;
                    font-size: 0.85rem;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.5);
                    }
                    
                    .pending-text {
                        font-style: italic;
                    }
                }
                
                .section-ratings-large {
                    display: flex;
                    gap: 0.5rem;
                    
                    .section-rating-large {
                        display: flex;
                        flex-direction: column;
                        align-items: center;
                        
                        .section-label-large {
                            font-size: 0.7rem;
                            font-weight: 600;
                            text-transform: uppercase;
                            letter-spacing: 0.5px;
                            margin-bottom: 0.2rem;
                        }
                        
                        .section-value-large {
                            font-size: 0.8rem;
                            font-weight: 700;
                            padding: 0.2rem 0.4rem;
                            border-radius: 4px;
                            min-width: 24px;
                            text-align: center;
                        }
                    }
                }
                
                .no-rating-message {
                    font-size: 0.8rem;
                    color: #9ca3af;
                    font-style: italic;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.5);
                    }
                }
            }
            
            .nation-overall {
                text-align: center;
                
                .nation-rating {
                    // NEW: Calculating state styles
                    .calculating-overall {
                        display: flex;
                        justify-content: center;
                        align-items: center;
                        min-height: 2rem;
                    }
                    
                    // NEW: Pending state styles
                    .pending-overall {
                        display: flex;
                        flex-direction: column;
                        align-items: center;
                        gap: 0.25rem;
                        color: #9ca3af;
                        
                        .body--dark & {
                            color: rgba(255, 255, 255, 0.5);
                        }
                        
                        .pending-stars {
                            font-size: 0.8rem;
                            font-style: italic;
                        }
                    }
                    
                    .highest-overall-large {
                        font-size: 0.9rem;
                        font-weight: 700;
                        margin-bottom: 0.3rem;
                    }
                    
                    .star-rating-large {
                        .star-large {
                            font-size: 0.8rem;
                            margin: 0 0.05rem;
                            
                            &.filled {
                                color: #fbbf24;
                            }
                            
                            &.half {
                                color: #fbbf24;
                                opacity: 0.5;
                            }
                            
                            &.empty {
                                color: #d1d5db;
                                
                                .body--dark & {
                                    color: rgba(255, 255, 255, 0.3);
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    .show-more-btn {
        font-weight: 500;
        border-radius: $border-radius-small;
        padding: 0.75rem 2rem;
        transition: all 0.2s ease;
        
        &:hover {
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            
            .body--dark & {
                box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
            }
        }
    }
}

// Overall Class Colors
.excellent {
    background-color: #10b981 !important;
    color: white !important;
}

.very-good {
    background-color: #34d399 !important;
    color: white !important;
}

.good {
    background-color: #fbbf24 !important;
    color: white !important;
}

.average {
    background-color: #f59e0b !important;
    color: white !important;
}

.below-average {
    background-color: #ef4444 !important;
    color: white !important;
}

.poor {
    background-color: #991b1b !important;
    color: white !important;
}

// Responsive Design
@media (max-width: 1200px) {
    .formation-tactics-layout {
        grid-template-columns: 1fr;
        gap: 1.5rem;
    }
    
    .formation-controls-panel {
        gap: 1.5rem;
    }
}

@media (max-width: 768px) {
    .main-content {
        padding: 1rem;
    }
    
    .nation-dashboard {
        gap: 1.5rem;
    }
    
    .nation-hero-section {
        padding: 2rem;
    }
    
    .performance-metrics .metrics-grid {
        grid-template-columns: 1fr;
        gap: 0.75rem;
    }
    
    .squad-depth-grid {
        grid-template-columns: 1fr !important;
        gap: 0.75rem !important;
    }
    
    .nations-list .nation-row {
        grid-template-columns: auto 1fr auto;
        gap: 0.75rem;
        
        .nation-section-ratings {
            display: none;
        }
    }
}
</style>