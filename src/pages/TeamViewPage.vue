<template>
    <q-page class="team-view-page">
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

            <!-- No Team Selected State -->
            <div v-if="!pageLoadingError && !selectedTeamName && !pageLoading" class="empty-state">
                <q-card class="empty-state-card">
                    <q-card-section class="empty-state-content">
                        <div class="empty-state-icon">
                            <q-icon name="groups" size="4rem" />
                        </div>
                        <h3 class="empty-state-title">Select a Team to Begin</h3>
                        <p class="empty-state-description">
                            Choose a team from the search above to unlock detailed tactical analysis, formation optimization, and squad insights.
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
            
            <div v-else-if="loadingTeam" class="loading-state">
                <q-spinner-dots color="primary" size="3em" />
                <div class="loading-text">Analyzing team composition...</div>
            </div>

            <!-- Main Team Content -->
            <div v-if="!pageLoading && !pageLoadingError && selectedTeamName && !loadingTeam" class="team-dashboard">
                
                <!-- Team Header Section (replaces hero) -->
                <GradientBackground 
                    :image-url="teamLogoUrl"
                    :fallback-gradient="'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'"
                    :intensity="0.4"
                    class="team-hero-section"
                >
                    <div class="team-hero-content">
                        <div class="team-primary-info">
                            <div class="team-name-section">
                                <div class="team-title-with-logo">
                                    <TeamLogo 
                                        :team-name="selectedTeamName"
                                        :size="48"
                                        class="team-logo-header"
                                    />
                                    <h1 class="team-name-hero">{{ selectedTeamName }}</h1>
                                </div>
                                <div v-if="teamDivision" class="team-division-hero">
                                    <q-icon name="military_tech" size="1.2rem" />
                                    <span>{{ teamDivision }}</span>
                                </div>
                            </div>
                            
                            <!-- Star Rating Display -->
                            <div v-if="bestTeamAverageOverall !== null" class="star-rating-display">
                                <div class="overall-score">{{ bestTeamAverageOverall }}</div>
                                <div class="star-container">
                                    <span
                                        v-for="star in 5"
                                        :key="star"
                                        class="star-modern"
                                        :class="getStarClass(bestTeamAverageOverall, star)"
                                    >
                                        ★
                                    </span>
                                </div>
                                <div class="rating-label">Overall Rating</div>
                            </div>
                        </div>
                        
                        <!-- Performance Metrics -->
                        <div v-if="currentTeamSectionRatings.attRating > 0 || currentTeamSectionRatings.midRating > 0 || currentTeamSectionRatings.defRating > 0" class="performance-metrics">
                            <div class="metrics-grid">
                                <div v-if="currentTeamSectionRatings.attRating > 0" class="metric-card attack">
                                    <div class="metric-icon">⚔️</div>
                                    <div class="metric-value" :class="getOverallClass(currentTeamSectionRatings.attRating)">
                                        {{ currentTeamSectionRatings.attRating }}
                                    </div>
                                    <div class="metric-label">Attack</div>
                                </div>
                                <div v-if="currentTeamSectionRatings.midRating > 0" class="metric-card midfield">
                                    <div class="metric-icon">⚽</div>
                                    <div class="metric-value" :class="getOverallClass(currentTeamSectionRatings.midRating)">
                                        {{ currentTeamSectionRatings.midRating }}
                                    </div>
                                    <div class="metric-label">Midfield</div>
                                </div>
                                <div v-if="currentTeamSectionRatings.defRating > 0" class="metric-card defense">
                                    <div class="metric-icon">🛡️</div>
                                    <div class="metric-value" :class="getOverallClass(currentTeamSectionRatings.defRating)">
                                        {{ currentTeamSectionRatings.defRating }}
                                    </div>
                                    <div class="metric-label">Defense</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </GradientBackground>

                <!-- Prestige Player Cards -->
                <div v-if="teamPlayers.length > 0" class="prestige-cards-section">
                    <div class="prestige-cards-row">
                        <div v-if="iconCardPlayer" class="prestige-card-wrapper">
                            <PlayerCards
                                :player="iconCardPlayer"
                                cardDesignOverride="card-design--parchment-frame"
                                :positionOverride="getDisplayPosition(iconCardPlayer)"
                                :datasetId="currentDatasetId"
                                :currency-symbol="detectedCurrencySymbol"
                            />
                            <div class="prestige-card-label icon-label">Highest Rating: {{ iconCardPlayer.originalOverall }}</div>
                        </div>
                        <div v-if="heroCardPlayer" class="prestige-card-wrapper">
                            <PlayerCards
                                :player="heroCardPlayer"
                                cardDesignOverride="card-design--badge-poster"
                                :positionOverride="getDisplayPosition(heroCardPlayer)"
                                :datasetId="currentDatasetId"
                                :currency-symbol="detectedCurrencySymbol"
                            />
                            <div class="prestige-card-label hero-label">Old but gold, {{ heroCardPlayer.originalOverall }} rating and {{ heroCardPlayer.originalAge }} age</div>
                        </div>
                        <div v-if="motmCardPlayer" class="prestige-card-wrapper">
                            <PlayerCards
                                :player="motmCardPlayer"
                                cardDesignOverride="card-design--match-night-flare"
                                :positionOverride="getDisplayPosition(motmCardPlayer)"
                                :datasetId="currentDatasetId"
                                :currency-symbol="detectedCurrencySymbol"
                            />
                            <div class="prestige-card-label motm-label">Top Performer: {{ motmCardPlayer.originalAvgRating }}</div>
                        </div>
                    </div>
                </div>

                <!-- Formation & Tactics Section -->
                <div class="formation-tactics-layout">
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

                    <q-card class="pitch-card" v-if="selectedFormationKey">
                        <q-card-section>
                            <div class="card-header">
                                <h3 class="card-title">
                                    <q-icon name="stadium" class="card-icon" />
                                    Formation View
                                </h3>
                                <p class="card-subtitle">Interactive pitch with your starting XI</p>
                            </div>
                            
                            <div class="pitch-stage">
                                <PitchDisplay
                                    :formation="currentFormationLayout"
                                    :players="bestTeamPlayersForPitch"
                                    display-mode="cards"
                                    :currency-symbol="detectedCurrencySymbol"
                                    :dataset-id="currentDatasetId"
                                    @player-click="handlePlayerSelectedFromTeam"
                                    @player-moved="handlePlayerMovedOnPitch"
                                />
                            </div>
                        </q-card-section>
                    </q-card>

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
                                            @click="handlePlayerSelectedFromTeam(playerEntry.player)"
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

                <!-- Players Table -->
                <q-card class="players-table-card">
                    <q-card-section>
                        <div class="card-header">
                            <h3 class="card-title">
                                <q-icon name="group" class="card-icon" />
                                Squad Overview
                            </h3>
                            <p class="card-subtitle">
                                All {{ teamPlayers.length }} players in {{ selectedTeamName }}
                            </p>
                        </div>
                        
                        <div class="table-container">
                            <PlayerDataTable
                                v-if="teamPlayers.length > 0"
                                :players="teamPlayers"
                                :loading="false"
                                @player-selected="handlePlayerSelectedFromTeam"
                                @team-selected="handleTeamSelected"
                                :is-goalkeeper-view="teamIsGoalkeeperView"
                                :currency-symbol="detectedCurrencySymbol"
                                :dataset-id="currentDatasetId"
                                class="modern-table"
                            />
                            <div v-else class="no-players-banner">
                                <q-icon name="person_off" size="3rem" />
                                <h4>No Players Found</h4>
                                <p>No players found for this team with the current data filters.</p>
                            </div>
                        </div>
                    </q-card-section>
                </q-card>
            </div>

            <!-- Additional Banners -->
            <q-banner
                v-else-if="!pageLoading && !loadingTeam && currentDatasetId && !selectedTeamName"
                class="info-banner"
            >
                <template v-slot:avatar>
                    <q-icon name="info" />
                </template>
                Please select a team to view its players and analyze formations.
            </q-banner>
            
            <q-banner
                v-else-if="!pageLoading && !loadingTeam && !currentDatasetId && !pageLoadingError"
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
import GradientBackground from '../components/GradientBackground.vue'
import PitchDisplay from '../components/PitchDisplay.vue'
import PlayerCards from '../components/PlayerCards.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import TeamLogo from '../components/TeamLogo.vue'
import { fetchTeamData } from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'
import { debounce } from '../utils/debounce'
import { formationCache } from '../utils/formationCache'
import { formations, getFormationLayout } from '../utils/formations'

// Currency utils are not directly used here for formatting,
// but PlayerDataTable and PlayerDetailDialog will use them with the passed symbol.

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
  'ST (C)': ['Striker (Centre)', 'Striker'],
}

export default {
  name: 'TeamViewPage',
  components: {
    PlayerDataTable,
    PlayerDetailDialog,
    PitchDisplay,
    PlayerCards,
    TeamLogo,
    GradientBackground,
  },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()

    const selectedTeamName = ref(null)
    const teamPlayers = ref([])
    const loadingTeam = ref(false)
    const pageLoading = ref(true)
    const pageLoadingError = ref('')

    const allPlayersData = computed(() => playerStore.allPlayers)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
    const currentDatasetId = computed(() => playerStore.currentDatasetId)

    // Computed property for team logo URL for gradient background
    const teamLogoUrl = computed(() => {
      if (!selectedTeamName.value) return null
      // We'll use the same API endpoint that TeamLogo component uses
      return `/api/logos?teamId=${encodeURIComponent(selectedTeamName.value)}&size=256`
    })

    const selectedFormationKey = ref(null)

    const squadComposition = ref({})

    const bestTeamAverageOverall = ref(null)
    const calculationMessage = ref('')
    const calculationMessageClass = ref('')

    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)

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
      STRIKER: 'ST',
    }

    // For handling combined positions like D/WB(R)
    // The first position is the PREFERRED position, others are fallbacks
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
      GK: ['GK'],
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
      GK: ['GK'],
    }

    const LINEUP_SELECTION_VERSION = 'v2-strict-positions'

    const getPlayerShortPositions = (player) => {
      if (Array.isArray(player?.shortPositions)) return player.shortPositions
      if (Array.isArray(player?.short_positions)) return player.short_positions
      return []
    }

    const getSlotExactPositions = (slotRole) => positionSideMap[slotRole.toUpperCase()] || []

    const getSlotFallbackPositions = (slotRole) => {
      const exactPositions = getSlotExactPositions(slotRole)
      const fallbackPositions = fallbackPositionMap[slotRole.toUpperCase()] || []
      return fallbackPositions.filter((position) => !exactPositions.includes(position))
    }

    const getExactCandidateCountForSlot = (slot) => {
      const slotPositions = getSlotExactPositions(slot.role)
      return teamPlayers.value.filter((player) =>
        getPlayerShortPositions(player).some((position) => slotPositions.includes(position))
      ).length
    }

    const getSlotSelectionOrder = (formationSlots) => {
      return [...formationSlots].sort((a, b) => {
        const exactCandidateDifference =
          getExactCandidateCountForSlot(a) - getExactCandidateCountForSlot(b)
        if (exactCandidateDifference !== 0) return exactCandidateDifference
        return formationSlots.indexOf(a) - formationSlots.indexOf(b)
      })
    }

    const fetchDatasetMetadata = async (datasetId) => {
      pageLoading.value = true
      pageLoadingError.value = ''
      try {
        // Only set the ID in the store, don't fetch all players
        playerStore.setCurrentDatasetId(datasetId)

        // If we don't have a currency symbol yet, we'll get it from team data
        // or we could potentially fetch just the config/metadata here if needed.
      } catch (err) {
        pageLoadingError.value = `Failed to initialize dataset: ${err.message || 'Unknown error'}.`
      } finally {
        pageLoading.value = false
      }
    }

    onMounted(async () => {
      const datasetIdFromQuery = route.query.datasetId
      const datasetIdFromRoute = route.params.datasetId
      const teamFromQuery = route.query.team
      const finalDatasetId =
        datasetIdFromRoute || datasetIdFromQuery || sessionStorage.getItem('currentDatasetId')

      if (finalDatasetId) {
        if (
          datasetIdFromQuery &&
          datasetIdFromQuery !== sessionStorage.getItem('currentDatasetId')
        ) {
          sessionStorage.setItem('currentDatasetId', datasetIdFromQuery)
        } else if (!datasetIdFromQuery && sessionStorage.getItem('currentDatasetId')) {
          // If loading from session, ensure query param is updated for consistency/bookmarking
          router.replace({ query: { datasetId: finalDatasetId } })
        }
        await fetchDatasetMetadata(finalDatasetId)
      } else {
        pageLoadingError.value =
          'No player dataset ID found. Please upload a file on the main page.'
        pageLoading.value = false
      }

      if (!pageLoadingError.value) {
        // If a team was specified in the query params, select it
        if (teamFromQuery && teamFromQuery.trim() !== '') {
          selectedTeamName.value = teamFromQuery
          loadTeamPlayers()
        }
      }
    })

    const loadTeamPlayersImmediate = async () => {
      if (!selectedTeamName.value) {
        teamPlayers.value = []
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
        calculationMessage.value = ''
        selectedFormationKey.value = null
        return
      }

      loadingTeam.value = true

      try {
        // Use the new team data API to get all detailed player data in one request
        const teamData = await fetchTeamData(currentDatasetId.value, 'team', selectedTeamName.value)

        if (teamData.data?.players) {
          teamPlayers.value = teamData.data.players

          // Update currency symbol if returned by the API
          if (teamData.data.currency_symbol) {
            playerStore.setCurrencySymbol(teamData.data.currency_symbol)
          }

          console.log('Team players loaded via API:', {
            teamName: selectedTeamName.value,
            playerCount: teamData.data.players.length,
            currencySymbol: teamData.data.currency_symbol,
          })

          // Auto-select the best formation for this team
          if (teamData.data.players.length > 0) {
            const bestFormation = calculateBestFormationForTeam()
            if (bestFormation) {
              selectedFormationKey.value = bestFormation
              calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`
              calculationMessageClass.value = quasarInstance.dark.isActive
                ? 'bg-info text-white'
                : 'bg-blue-2 text-primary'
            } else {
              selectedFormationKey.value = null
              squadComposition.value = {}
              bestTeamAverageOverall.value = null
              calculationMessage.value = 'No suitable formation found for this team.'
              calculationMessageClass.value = quasarInstance.dark.isActive
                ? 'text-grey-5'
                : 'text-grey-7'
            }
          } else {
            selectedFormationKey.value = null
            squadComposition.value = {}
            bestTeamAverageOverall.value = null
            calculationMessage.value = 'No players found for this team.'
            calculationMessageClass.value = quasarInstance.dark.isActive
              ? 'text-grey-5'
              : 'text-grey-7'
          }
        } else {
          throw new Error('Invalid team data response')
        }
      } catch (error) {
        console.error('Error loading team players:', error)
        calculationMessage.value = `Failed to load team players: ${error.message}`
        calculationMessageClass.value = quasarInstance.dark.isActive ? 'text-red-5' : 'text-red-7'

        // Fallback to empty state
        teamPlayers.value = []
        selectedFormationKey.value = null
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
      }

      loadingTeam.value = false
    }

    // Debounced version for better performance
    const loadTeamPlayers = debounce(async () => {
      await loadTeamPlayersImmediate()
    }, 300)

    // Simplified team clearing - now just resets the team name since navigation is handled by URL
    const clearTeamSelection = () => {
      selectedTeamName.value = null
      teamPlayers.value = []
      selectedFormationKey.value = null
      squadComposition.value = {}
      bestTeamAverageOverall.value = null
      calculationMessage.value = ''
    }

    const formationOptions = computed(() => {
      return Object.keys(formations).map((key) => ({
        label: formations[key].name,
        value: key,
      }))
    })

    const currentFormationLayout = computed(() => {
      if (!selectedFormationKey.value) {
        return []
      }
      return getFormationLayout(selectedFormationKey.value) || []
    })

    const bestTeamPlayersForPitch = computed(() => {
      const starters = {}
      if (!squadComposition.value || Object.keys(squadComposition.value).length === 0) {
        return starters
      }
      for (const slotId in squadComposition.value) {
        if (squadComposition.value[slotId] && squadComposition.value[slotId].length > 0) {
          const starterEntry = squadComposition.value[slotId][0]
          // Use the role-specific score for this position, not their global Overall
          // Add the exactMatch flag to display if the player is in their natural position
          starters[slotId] = {
            ...starterEntry.player, // Spread all player properties
            Overall: starterEntry.overallInRole, // Use position-specific rating
            exactPositionMatch: starterEntry.exactMatch, // Pass this to the pitch display
          }
        } else {
          starters[slotId] = null // No player for this slot
        }
      }
      return starters
    })

    const currentTeamSectionRatings = computed(() => {
      if (!squadComposition.value || !currentFormationLayout.value) {
        return { attRating: 0, midRating: 0, defRating: 0 }
      }
      return calculateSectionRatings(squadComposition.value, currentFormationLayout.value)
    })

    const teamIsGoalkeeperView = computed(() => {
      // This computed property is for the PlayerDataTable on this page.
      // It should only show goalkeeper view if the majority of players are goalkeepers,
      // otherwise default to outfield player view which is what users typically want to see.
      if (teamPlayers.value.length === 0) return false

      const goalkeeperCount = teamPlayers.value.filter((p) =>
        p.position_groups?.includes('Goalkeepers')
      ).length

      // Only show goalkeeper view if more than half the players are goalkeepers
      return goalkeeperCount > teamPlayers.value.length / 2
    })

    const teamDivision = computed(() => {
      // Get the division from the first player in the team
      if (teamPlayers.value.length > 0) {
        return teamPlayers.value[0].division
      }
      return null
    })

    const getStarRating = (overall) => {
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

    const getStarClass = (overall, starPosition) => {
      if (!overall || overall === 0) return 'star-empty'

      const starRating = getStarRating(overall)

      if (starPosition <= Math.floor(starRating)) {
        return 'star-full'
      }
      if (starPosition === Math.floor(starRating) + 1 && starRating % 1 === 0.5) {
        return 'star-half'
      }
      return 'star-empty'
    }

    const handlePlayerSelectedFromTeam = (player) => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const handleTeamSelected = (teamName) => {
      // Navigate to team view page with the selected team
      const datasetId = currentDatasetId.value || sessionStorage.getItem('currentDatasetId')
      if (datasetId && teamName && teamName !== selectedTeamName.value) {
        // Update the URL to reflect the new team selection
        router.push({
          path: '/team-view',
          query: {
            datasetId: datasetId,
            team: teamName,
          },
        })
      }
    }

    const getOverallClass = (overall) => {
      if (overall === null || overall === undefined) return 'rating-na'
      const numericOverall = Number(overall)
      if (Number.isNaN(numericOverall)) return 'rating-na'

      if (numericOverall >= 90) return 'rating-tier-6'
      if (numericOverall >= 80) return 'rating-tier-5'
      if (numericOverall >= 70) return 'rating-tier-4'
      if (numericOverall >= 55) return 'rating-tier-3'
      if (numericOverall >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    const calculateSectionRatings = (squadComposition, formationLayout) => {
      if (!squadComposition || !formationLayout) {
        return { attRating: 0, midRating: 0, defRating: 0 }
      }

      const formationSlots = formationLayout.flatMap((row) => row.positions)

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
        defRating: defCount > 0 ? Math.round(defSum / defCount) : 0,
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
        const exactPositionMatches = player.shortPositions.filter((pos) =>
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
        const fallbackMatches = player.shortPositions.filter((pos) =>
          fallbackPositions.includes(pos)
        )

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
          .map((matcher) => fmMatcherToRoleKeyPrefix[matcher.toUpperCase()])
          .filter((prefix) => !!prefix)
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

    const MIN_SUITABILITY_THRESHOLD = 10

    const getSlotDisplayName = (slot, allSlots) => {
      const roleCounts = allSlots.reduce((acc, s) => {
        acc[s.role] = (acc[s.role] || 0) + 1
        return acc
      }, {})

      if (roleCounts[slot.role] > 1) {
        // If multiple slots have the same base role (e.g., two "ST (C)"),
        // use the more specific ID (like "STCL", "STCR").
        // Extract the prefix from ID, e.g., "STCL" from "STCL_41212N"
        return slot.id.split('_')[0]
      }
      return slot.role // Otherwise, use the general role name like "AM (C)"
    }

    const calculateBestFormationForTeam = () => {
      if (teamPlayers.value.length === 0) {
        // console.log('No team players available for formation calculation')
        return null
      }

      console.log('Calculating best formation for team:', {
        teamName: selectedTeamName.value,
        playerCount: teamPlayers.value.length,
        samplePlayerData: teamPlayers.value[0]
          ? {
              name: teamPlayers.value[0].name,
              shortPositions: teamPlayers.value[0].shortPositions,
              roleSpecificOveralls: teamPlayers.value[0].roleSpecificOveralls?.length || 0,
              Overall: teamPlayers.value[0].Overall,
            }
          : null,
      })

      // Debug: Check position mappings
      console.log('Position mappings for GK:', positionSideMap.GK)
      console.log('Position mappings for ST (C):', positionSideMap['ST (C)'])
      console.log('Fallback mappings for GK:', fallbackPositionMap.GK)

      // Check cache first
      const cacheKey = formationCache.generateKey(
        teamPlayers.value,
        `team-best-${LINEUP_SELECTION_VERSION}`
      )
      const cachedResult = formationCache.get(cacheKey)
      if (cachedResult) {
        // console.log('Using cached formation result:', cachedResult.bestFormationKey)
        return cachedResult.bestFormationKey
      }

      let bestFormationKey = null
      let bestAverageOverall = 0

      // Test each formation to find the one with highest average overall
      for (const formationKey of Object.keys(formations)) {
        const formationLayoutForCalc = getFormationLayout(formationKey)
        if (!formationLayoutForCalc) continue

        const formationSlots = formationLayoutForCalc.flatMap((row) => row.positions)
        const tempSquadComposition = {}

        // Initialize slots
        for (const slot of formationSlots) {
          tempSquadComposition[slot.id] = []
        }

        // Calculate player scores for each position in this formation
        const allPotentialPlayerAssignments = []
        for (const slot of formationSlots) {
          // Use the same logic as calculateBestTeamAndDepth for consistency
          for (const player of teamPlayers.value) {
            const overallInRole = getPlayerOverallForRole(player, slot.role)

            if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
              const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
              const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []
              const playerPositions = player.shortPositions || []

              const isExactMatch = playerPositions.some((pos) => slotPositions.includes(pos))
              const isFallbackMatch = playerPositions.some((pos) => fallbackPositions.includes(pos))

              // Include both exact matches and fallback matches
              if (isExactMatch || isFallbackMatch) {
                const assignment = {
                  player,
                  slotId: slot.id,
                  slotRole: slot.role,
                  overallInRole: overallInRole,
                  sortScore: overallInRole,
                  exactMatch: isExactMatch,
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

          // Debug: Log the first few players and their position data
          if (formationKey === '442_classic' && slot.role === 'GK') {
            console.log('Debug - Sample players for GK position:')
            for (let i = 0; i < Math.min(3, teamPlayers.value.length); i++) {
              const player = teamPlayers.value[i]
              const overallInRole = getPlayerOverallForRole(player, slot.role)
              console.log(
                `Player ${i}: ${player.name}, positions: ${player.shortPositions}, overall: ${overallInRole}, threshold: ${MIN_SUITABILITY_THRESHOLD}`
              )
            }
          }
        }

        // Sort assignments by sort score
        allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)

        const assignedPlayersToSlots = new Set()

        // Fill scarce natural positions first so flexible players do not block wing-backs/full-backs.
        for (const slot of getSlotSelectionOrder(formationSlots)) {
          for (const assignment of allPotentialPlayerAssignments) {
            if (
              assignment.slotId === slot.id &&
              !assignedPlayersToSlots.has(assignment.player.name)
            ) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch,
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }

        // Calculate average overall for this formation
        let sumOfStartersOverall = 0
        let startersCount = 0
        let filledPositions = 0
        for (const slotPlayers of Object.values(tempSquadComposition)) {
          if (slotPlayers && slotPlayers.length > 0) {
            sumOfStartersOverall += slotPlayers[0].overallInRole
            startersCount++
            filledPositions++
          }
        }

        // Only consider formations with at least 5 filled positions (minimum viable team)
        const hasEnoughPlayers = filledPositions >= 5

        console.log(
          `Formation ${formationKey}: filled=${filledPositions}, avg=${startersCount > 0 ? (sumOfStartersOverall / startersCount).toFixed(1) : 0}, viable=${hasEnoughPlayers}`
        )

        if (startersCount > 0 && hasEnoughPlayers) {
          const averageOverall = sumOfStartersOverall / startersCount
          if (averageOverall > bestAverageOverall) {
            bestAverageOverall = averageOverall
            bestFormationKey = formationKey
          }
        }
      }

      console.log('Formation calculation result:', {
        bestFormationKey,
        bestAverageOverall,
        totalFormationsTested: Object.keys(formations).length,
      })

      // Cache the result
      if (bestFormationKey) {
        formationCache.set(cacheKey, {
          bestFormationKey,
          bestAverageOverall,
          teamName: selectedTeamName.value,
        })
      }

      return bestFormationKey
    }

    const calculateBestTeamAndDepth = () => {
      if (!selectedFormationKey.value || teamPlayers.value.length === 0) {
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
        calculationMessage.value = selectedFormationKey.value
          ? 'No players in the selected team.'
          : 'Select a formation.'
        calculationMessageClass.value = 'bg-warning text-dark'
        return
      }

      // Check cache first for squad composition
      const cacheKey = formationCache.generateKey(
        teamPlayers.value,
        `team-depth-${LINEUP_SELECTION_VERSION}-${selectedFormationKey.value}`
      )
      const cachedResult = formationCache.get(cacheKey)
      if (cachedResult) {
        squadComposition.value = cachedResult.squadComposition
        bestTeamAverageOverall.value = cachedResult.bestTeamAverageOverall
        calculationMessage.value = `Best XI & Depth calculated (cached). Average Overall: ${cachedResult.bestTeamAverageOverall}.`
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'
        return
      }

      calculationMessage.value = 'Calculating best team and depth...'
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

      const formationSlots = formationLayoutForCalc.flatMap((row) => row.positions)

      // Initialize slots
      for (const slot of formationSlots) {
        tempSquadComposition[slot.id] = []
      }

      // ENHANCEMENT: First, compute all player scores for all positions
      // and check which players can play in which positions
      const playerPositionMap = new Map() // Maps player name to positions they can play

      for (const player of teamPlayers.value) {
        const playablePositions = [...getPlayerShortPositions(player)]
        playerPositionMap.set(player.name, playablePositions)
      }

      // Calculate player scores for each position
      const allPotentialPlayerAssignments = []
      for (const slot of formationSlots) {
        for (const player of teamPlayers.value) {
          const overallInRole = getPlayerOverallForRole(
            player,
            slot.role // Use the general role from formation (e.g., "ST (C)")
          )

          // Only include players who meet the threshold and are properly positioned
          if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
            // Get the compatible positions for this slot
            const slotPositions = getSlotExactPositions(slot.role)
            const fallbackPositions = getSlotFallbackPositions(slot.role)

            // STRICT POSITION CHECKING: Check if player can play in this position
            // For this to be true, the player MUST have one of the required positions
            // in their shortPositions array

            const playerPositions = playerPositionMap.get(player.name) || []

            // For first XI and depth chart, we ONLY want players who can ACTUALLY play the position
            // isExactMatch means player has the EXACT position for this slot
            const isExactMatch = playerPositions.some((pos) => slotPositions.includes(pos))

            // Include both exact matches and fallback positions for more options
            const isFallbackMatch = playerPositions.some((pos) => fallbackPositions.includes(pos))
            const canPlayInPosition = isExactMatch || isFallbackMatch

            if (canPlayInPosition) {
              // Strict position filtering:
              // 1. For first team selection, we want EXACT position matches only unless
              //    there are no players for a position
              // 2. For depth, we can be more flexible

              // Store the original role score and position match info
              const assignment = {
                player,
                slotId: slot.id,
                slotRole: slot.role,
                overallInRole: overallInRole, // Store original score for display
                sortScore: overallInRole, // Will be used for sorting
                exactMatch: isExactMatch, // Flag for UI display
              }

              // Adjust sort score based on position match
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

      // Sort assignments by the sort score, which already includes position match bonus
      allPotentialPlayerAssignments.sort((a, b) => {
        return b.sortScore - a.sortScore
      })

      const assignedPlayersToSlots = new Set()
      const slotsByNaturalScarcity = getSlotSelectionOrder(formationSlots)

      for (let depthIndex = 0; depthIndex < 3; depthIndex++) {
        // First pass: fill positions with exact matches
        for (const slot of slotsByNaturalScarcity) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            // If this slot needs a player at current depth
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                assignment.exactMatch && // Only use exact matches in first pass
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                // Check if this player is already a starter in *another* slot if we are filling backups
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  // Only check for backups
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
                    exactMatch: assignment.exactMatch,
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break // Move to next slot for this depth level
                }
              }
            }
          }
        }

        // Second pass: fill remaining positions with fallback matches
        for (const slot of slotsByNaturalScarcity) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            // If this slot still needs a player after the first pass
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                !assignment.exactMatch &&
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                // Check if this player is already a starter in *another* slot if we are filling backups
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  // Only check for backups
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
                    exactMatch: assignment.exactMatch,
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break // Move to next slot for this depth level
                }
              }
            }
          }
        }
      }

      // Ensure each slot in tempSquadComposition is sorted by overallInRole descending
      for (const slotId in tempSquadComposition) {
        tempSquadComposition[slotId].sort((a, b) => {
          if (a.exactMatch !== b.exactMatch) return a.exactMatch ? -1 : 1
          return b.overallInRole - a.overallInRole
        })
      }

      // Check if any positions have no players assigned at all
      // In that case, try to find any player who can play there as a fallback
      for (const slot of formationSlots) {
        if (tempSquadComposition[slot.id].length === 0) {
          // Get fallback positions for this slot
          const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []

          // Find any players who can play in fallback positions
          const fallbackAssignments = []

          for (const player of teamPlayers.value) {
            if (!assignedPlayersToSlots.has(player.name)) {
              const playerPositions = getPlayerShortPositions(player)

              // Check if player can play any fallback position
              const canPlayFallback = playerPositions.some((pos) => fallbackPositions.includes(pos))

              if (canPlayFallback) {
                const overallInRole = getPlayerOverallForRole(player, slot.role)
                if (overallInRole >= MIN_SUITABILITY_THRESHOLD - 10) {
                  fallbackAssignments.push({
                    player,
                    overallInRole,
                    exactMatch: false,
                  })
                }
              }
            }
          }

          // Sort fallbacks by score
          fallbackAssignments.sort((a, b) => b.overallInRole - a.overallInRole)

          // Add best fallback if available
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
        bestTeamAverageOverall.value = Math.round(sumOfStartersOverall / startersCount)
        calculationMessage.value = `Best XI & Depth calculated. Average Overall: ${bestTeamAverageOverall.value}.`
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'
      } else {
        bestTeamAverageOverall.value = 0
        calculationMessage.value = 'Could not assign any suitable players to form a Best XI.'
        calculationMessageClass.value = quasarInstance.dark.isActive
          ? 'bg-negative text-white'
          : 'bg-red-2 text-negative'
      }

      // Cache the result
      if (bestTeamAverageOverall.value > 0) {
        formationCache.set(cacheKey, {
          squadComposition: squadComposition.value,
          bestTeamAverageOverall: bestTeamAverageOverall.value,
          teamName: selectedTeamName.value,
          formation: selectedFormationKey.value,
        })
      }
    }

    watch(selectedFormationKey, (newKey) => {
      if (newKey && selectedTeamName.value) {
        calculateBestTeamAndDepth()
      } else {
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
        calculationMessage.value = 'Select a team and formation.'
        calculationMessageClass.value = quasarInstance.dark.isActive ? 'text-grey-5' : 'text-grey-7'
      }
    })

    const handlePlayerMovedOnPitch = (moveData) => {
      const { player, fromSlotId, toSlotId, toSlotRole } = moveData

      const currentStarters = JSON.parse(JSON.stringify(bestTeamPlayersForPitch.value))
      const playerToMoveFullData = teamPlayers.value.find((p) => p.name === player.name)

      if (!playerToMoveFullData) return

      // Calculate the role-specific rating for this player in the new position
      const overallInNewRole = getPlayerOverallForRole(playerToMoveFullData, toSlotRole)

      // Check if player is in their natural position in the new slot
      const playerPositions = playerToMoveFullData.shortPositions || []
      const slotPositions = positionSideMap[toSlotRole.toUpperCase()] || []
      const isExactMatch = playerPositions.some((pos) => slotPositions.includes(pos))

      const playerCurrentlyInTargetSlotFullData = currentStarters[toSlotId]
        ? teamPlayers.value.find((p) => p.name === currentStarters[toSlotId].name)
        : null

      // Update target slot with role-specific rating and position match info
      currentStarters[toSlotId] = {
        ...playerToMoveFullData,
        Overall: overallInNewRole, // Role-specific rating for the position
        exactPositionMatch: isExactMatch, // Position match flag for UI
      }

      // Update original slot
      if (playerCurrentlyInTargetSlotFullData && fromSlotId) {
        const originalRoleOfFromSlot = currentFormationLayout.value
          .flatMap((r) => r.positions)
          .find((p) => p.id === fromSlotId)?.role

        if (originalRoleOfFromSlot) {
          // Calculate role-specific rating for the player in the original slot
          const overallInOldRole = getPlayerOverallForRole(
            playerCurrentlyInTargetSlotFullData,
            originalRoleOfFromSlot
          )

          // Check if player is in their natural position in the original slot
          const playerPositions = playerCurrentlyInTargetSlotFullData.shortPositions || []
          const slotPositions = positionSideMap[originalRoleOfFromSlot.toUpperCase()] || []
          const isExactMatch = playerPositions.some((pos) => slotPositions.includes(pos))

          currentStarters[fromSlotId] = {
            ...playerCurrentlyInTargetSlotFullData,
            Overall: overallInOldRole, // Role-specific rating
            exactPositionMatch: isExactMatch, // Position match flag
          }
        } else {
          currentStarters[fromSlotId] = null
        }
      } else if (fromSlotId) {
        currentStarters[fromSlotId] = null
      }

      // To make PitchDisplay update, we need to change the object reference
      // or ensure its internal properties are reactive.
      // This is a simplified visual swap; it doesn't formally update squadComposition.
      // For a temporary visual update of the pitch:
      const newPitchState = { ...currentStarters }

      let sumOfDisplayedOveralls = 0
      let countOfDisplayedOveralls = 0
      for (const p of Object.values(newPitchState)) {
        if (p && typeof p.Overall === 'number') {
          // p.Overall is now the position-specific rating
          sumOfDisplayedOveralls += p.Overall
          countOfDisplayedOveralls++
        }
      }
      bestTeamAverageOverall.value =
        countOfDisplayedOveralls > 0
          ? Math.round(sumOfDisplayedOveralls / countOfDisplayedOveralls)
          : 0

      calculationMessage.value = `Team visually adjusted. New Avg Overall: ${bestTeamAverageOverall.value}. (Depth chart not updated by drag & drop).`
      calculationMessageClass.value = quasarInstance.dark.isActive
        ? 'bg-info text-white'
        : 'bg-blue-2 text-primary'

      // To actually make PitchDisplay update from this drag-drop,
      // bestTeamPlayersForPitch would need to be made writable or a separate ref used.
      // For now, this is a visual indication of the swap's effect on average overall.
      // The actual `bestTeamPlayersForPitch` computed will still be based on `squadComposition`.
      // To truly reflect the drag-drop, `squadComposition` itself would need to be modified.
    }

    watch(
      () => route.query.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          sessionStorage.setItem('currentDatasetId', newId)
          await fetchDatasetMetadata(newId)
          clearTeamSelection() // Reset team selection as data has changed
          if (!pageLoadingError.value && selectedTeamName.value) {
            loadTeamPlayers() // Reload team if already selected
          }
        }
      }
    )

    watch(
      () => route.query.team,
      (newTeam) => {
        if (newTeam && newTeam !== selectedTeamName.value) {
          selectedTeamName.value = newTeam
          loadTeamPlayers()
        } else if (!newTeam && selectedTeamName.value) {
          clearTeamSelection()
        }
      }
    )

    const shareDataset = async () => {
      if (!currentDatasetId.value) return

      const shareUrl = `${window.location.origin}/team-view/${currentDatasetId.value}`

      try {
        await navigator.clipboard.writeText(shareUrl)
        quasarInstance.notify({
          message: 'Link copied to clipboard!',
          color: 'positive',
          icon: 'check_circle',
          position: 'top',
          timeout: 2000,
        })
      } catch (_err) {
        // Fallback for older browsers
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
          timeout: 2000,
        })
      }
    }

    // --- Prestige Player Card Helpers ---
    const clampNumber = (value, min, max) => Math.min(max, Math.max(min, value))

    const getOverallValueLocal = (player) => {
      const overall = Number(player?.Overall ?? player?.overall ?? 0)
      return Number.isNaN(overall) ? 0 : overall
    }

    const outfieldStatKeys = ['pac', 'dri', 'sho', 'def', 'pas', 'phy']
    const goalkeeperStatKeys = ['div', 'han', 'kic', 'ref', 'spd', 'pos']

    const statRelevanceByPosition = {
      GK: { div: 1.18, han: 1.08, kic: 0.72, ref: 1.2, spd: 0.55, pos: 1.12 },
      LB: { pac: 0.9, dri: 0.65, sho: 0.22, def: 1.2, pas: 0.76, phy: 1.05 },
      RB: { pac: 0.9, dri: 0.65, sho: 0.22, def: 1.2, pas: 0.76, phy: 1.05 },
      CB: { pac: 0.58, dri: 0.28, sho: 0.18, def: 1.24, pas: 0.64, phy: 1.14 },
      CDM: { pac: 0.5, dri: 0.68, sho: 0.36, def: 1.12, pas: 1.1, phy: 0.96 },
      CM: { pac: 0.48, dri: 0.9, sho: 0.58, def: 0.68, pas: 1.18, phy: 0.78 },
      CAM: { pac: 0.72, dri: 1.12, sho: 0.95, def: 0.24, pas: 1.16, phy: 0.52 },
      LW: { pac: 1.12, dri: 1.14, sho: 0.82, def: 0.22, pas: 0.82, phy: 0.52 },
      LM: { pac: 1.02, dri: 1.04, sho: 0.62, def: 0.48, pas: 0.92, phy: 0.62 },
      ST: { pac: 0.84, dri: 0.72, sho: 1.22, def: 0.16, pas: 0.58, phy: 0.92 },
      RW: { pac: 1.12, dri: 1.14, sho: 0.82, def: 0.22, pas: 0.82, phy: 0.52 },
      RM: { pac: 1.02, dri: 1.04, sho: 0.62, def: 0.48, pas: 0.92, phy: 0.62 },
    }

    const getRelevantStatFloor = (targetOverall, relevance) => {
      if (relevance >= 1.18) return targetOverall - 2
      if (relevance >= 1) return targetOverall - 7
      if (relevance >= 0.82) return targetOverall - 8
      if (relevance >= 0.62) return targetOverall - 10
      return 0
    }

    const getUpgradedStat = (
      statValue,
      baseOverall,
      targetOverall,
      relevance,
      displayPosition,
      positionBonus
    ) => {
      const numericStat = Number(statValue)
      if (Number.isNaN(numericStat) || numericStat <= 0) return statValue
      const overallBoost = Math.max(0, targetOverall - baseOverall)
      let statBoost = overallBoost * (0.65 + relevance * 0.7) + positionBonus
      if (numericStat >= targetOverall + 2) statBoost *= 0.35
      else if (numericStat >= targetOverall) statBoost *= 0.5
      const floor = getRelevantStatFloor(targetOverall, relevance)
      const isAttacker = ['LW', 'LM', 'ST', 'RW', 'RM'].includes(displayPosition)
      const positionCap =
        relevance >= 0.62 ? targetOverall + 8 : targetOverall - (isAttacker ? 5 : 8)
      const boosted = Math.max(numericStat + statBoost, floor)
      return Math.round(clampNumber(boosted, numericStat, Math.min(99, positionCap)))
    }

    const shortToDisplayPosition = {
      GK: 'GK',
      DC: 'CB',
      DR: 'RB',
      DL: 'LB',
      WBR: 'RB',
      WBL: 'LB',
      DM: 'CDM',
      MR: 'RM',
      ML: 'LM',
      MC: 'CM',
      AMR: 'RW',
      AML: 'LW',
      AMC: 'CAM',
      ST: 'ST',
    }

    const getDisplayPosition = (player) => {
      if (!player) return 'CM'
      const positions = player.shortPositions || player.short_positions || []
      for (const pos of positions) {
        const mapped = shortToDisplayPosition[pos]
        if (mapped) return mapped
      }
      return 'CM'
    }

    const createPrestigePlayer = (player, boostAmount, minOverall = 0) => {
      const baseOverall = getOverallValueLocal(player)
      const displayPosition = getDisplayPosition(player)
      const isGK = displayPosition === 'GK'
      const targetOverall = Math.min(99, Math.max(baseOverall + boostAmount, minOverall))
      const relevance = statRelevanceByPosition[displayPosition] || statRelevanceByPosition.CM
      const statKeys = isGK ? goalkeeperStatKeys : outfieldStatKeys
      const positionBonus = isGK ? 0 : 1

      const boostedPlayer = { ...player, Overall: targetOverall, overall: targetOverall }
      for (const key of statKeys) {
        if (player[key] !== undefined) {
          boostedPlayer[key] = getUpgradedStat(
            player[key],
            baseOverall,
            targetOverall,
            relevance[key] || 0.5,
            displayPosition,
            positionBonus
          )
        }
      }
      return boostedPlayer
    }

    const iconCardPlayer = computed(() => {
      if (!teamPlayers.value.length) return null
      const best = teamPlayers.value.reduce((prev, curr) =>
        getOverallValueLocal(curr) > getOverallValueLocal(prev) ? curr : prev
      )
      return {
        ...createPrestigePlayer(best, 10, 85),
        isIcon: true,
        originalOverall: getOverallValueLocal(best),
      }
    })

    const heroCardPlayer = computed(() => {
      if (!teamPlayers.value.length) return null
      const withApps = teamPlayers.value.filter((p) => {
        const apps = Number(p.Apps ?? p.apps ?? p['Ap'] ?? 0)
        return !Number.isNaN(apps) && apps >= 10
      })
      const pool = withApps.length > 0 ? withApps : teamPlayers.value
      const oldest = pool.reduce((prev, curr) => {
        const prevAge = Number(prev.age ?? prev.Age ?? 0)
        const currAge = Number(curr.age ?? curr.Age ?? 0)
        return currAge > prevAge ? curr : prev
      })
      return {
        ...createPrestigePlayer(oldest, 8),
        originalOverall: getOverallValueLocal(oldest),
        originalAge: Number(oldest.age ?? oldest.Age ?? 0),
      }
    })

    const motmCardPlayer = computed(() => {
      if (!teamPlayers.value.length) return null
      const withRating = teamPlayers.value.filter((p) => {
        const rating = Number(p['Av Rat'] ?? p.avRat ?? 0)
        return !Number.isNaN(rating) && rating > 0
      })
      let chosen
      if (withRating.length > 0) {
        chosen = withRating.reduce((prev, curr) => {
          const prevRat = Number(prev['Av Rat'] ?? prev.avRat ?? 0)
          const currRat = Number(curr['Av Rat'] ?? curr.avRat ?? 0)
          return currRat > prevRat ? curr : prev
        })
      } else {
        const sorted = [...teamPlayers.value].sort(
          (a, b) => getOverallValueLocal(b) - getOverallValueLocal(a)
        )
        chosen = sorted[1] ?? sorted[0]
      }
      const avRat = Number(chosen['Av Rat'] ?? chosen.avRat ?? 0)
      return {
        ...createPrestigePlayer(chosen, 5),
        originalAvgRating: avRat > 0 ? avRat : getOverallValueLocal(chosen),
      }
    })

    const _startersWithRoleRatings = computed(() => {
      if (!squadComposition.value || Object.keys(squadComposition.value).length === 0) {
        return {}
      }
      const starters = {}
      for (const [slotId, starterEntry] of Object.entries(squadComposition.value)) {
        if (starterEntry?.player) {
          starters[slotId] = {
            ...starterEntry.player,
            Overall: starterEntry.overallInRole,
            exactPositionMatch: starterEntry.exactMatch,
          }
        } else {
          starters[slotId] = null
        }
      }
      return starters
    })

    return {
      allPlayersData,
      selectedTeamName,
      loadTeamPlayers,
      clearTeamSelection,
      teamPlayers,
      loadingTeam,
      pageLoading,
      pageLoadingError,
      selectedFormationKey,
      formationOptions,
      currentFormationLayout,
      squadComposition,
      bestTeamPlayersForPitch,
      bestTeamAverageOverall,
      currentTeamSectionRatings,
      calculateSectionRatings,
      calculationMessage,
      calculationMessageClass,
      playerForDetailView,
      showPlayerDetailDialog,
      handlePlayerSelectedFromTeam,
      handleTeamSelected,
      teamIsGoalkeeperView,
      teamDivision,
      getStarRating,
      getStarClass,
      getOverallClass,
      getSlotDisplayName,
      handlePlayerMovedOnPitch,
      quasarInstance,
      router,
      detectedCurrencySymbol, // Expose currency symbol
      currentDatasetId,
      shareDataset,
      teamLogoUrl,
      iconCardPlayer,
      heroCardPlayer,
      motmCardPlayer,
      getDisplayPosition,
    }
  },
}
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

// Main Content Layout
.main-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
    
    @media (max-width: 768px) {
        padding: 1rem;
    }
}

// Modern Banners
.error-banner {
    background: $danger-gradient;
    color: white;
    margin-bottom: 2rem;
    border-radius: $border-radius;
    box-shadow: $card-shadow;
}

.info-banner {
    background: linear-gradient(135deg, #74b9ff 0%, #0984e3 100%);
    color: white;
    margin: 2rem 0;
    border-radius: $border-radius;
    box-shadow: $card-shadow;
}

.warning-banner {
    background: $warning-gradient;
    color: #2d3436;
    margin: 2rem 0;
    border-radius: $border-radius;
    box-shadow: $card-shadow;
}

// Share Section
.share-section {
    display: flex;
    justify-content: flex-end;
    margin: 2rem 0;
    
    .share-btn-modern {
        font-weight: 600;
        border-radius: $border-radius-small;
        box-shadow: $card-shadow;
        transition: all 0.3s ease;
        padding: 0.75rem 1.5rem;
        
        &:hover {
            transform: translateY(-2px);
            box-shadow: $card-shadow-hover;
        }
    }
}

// Empty State
.empty-state {
    margin: 4rem 0;
    
    .empty-state-card {
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        border: 1px solid rgba(0, 0, 0, 0.05);
        
        .body--dark & {
            border: 1px solid rgba(255, 255, 255, 0.1);
            background: rgba(255, 255, 255, 0.02);
        }
        
        .empty-state-content {
            text-align: center;
            padding: 4rem 2rem;
            
            .empty-state-icon {
                margin-bottom: 2rem;
                opacity: 0.6;
            }
            
            .empty-state-title {
                font-size: 2rem;
                font-weight: 700;
                margin: 0 0 1rem 0;
                color: #2d3436;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .empty-state-description {
                font-size: 1.1rem;
                line-height: 1.6;
                color: #636e72;
                margin: 0 0 2rem 0;
                max-width: 500px;
                margin-left: auto;
                margin-right: auto;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
            
            .empty-state-btn {
                font-weight: 600;
                padding: 0.75rem 2rem;
                border-radius: $border-radius-small;
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
    padding: 4rem 2rem;
    
    .loading-text {
        margin-top: 1.5rem;
        font-size: 1.1rem;
        color: #636e72;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
    }
}

// Team Dashboard
.team-dashboard {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

// Team Hero Section (replaces the hero section)
.team-hero-section {
    color: white;
    border-radius: $border-radius;
    padding: 3rem;
    margin-bottom: 2rem;
    position: relative;
    overflow: hidden;
    
    .team-hero-content {
        position: relative;
        z-index: 2;
        display: grid;
        grid-template-columns: 1fr auto;
        gap: 3rem;
        align-items: center;
        
        @media (max-width: 968px) {
            grid-template-columns: 1fr;
            gap: 2rem;
            text-align: center;
        }
        
        .team-primary-info {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            
            .team-name-section {
                .team-title-with-logo {
                    display: flex;
                    align-items: center;
                    gap: 1rem;
                    justify-content: center;
                    
                    @media (max-width: 768px) {
                        gap: 0.75rem;
                    }
                    
                    .team-logo-header {
                        width: 48px;
                        height: 48px;
                        flex-shrink: 0;
                        
                        @media (max-width: 768px) {
                            width: 36px;
                            height: 36px;
                        }
                    }
                }
                
                .team-name-hero {
                    font-size: 3.5rem;
                    font-weight: 800;
                    margin: 0;
                    color: white;
                    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
                    
                    @media (max-width: 768px) {
                        font-size: 2.5rem;
                    }
                }
                
                .team-division-hero {
                    display: flex;
                    align-items: center;
                    gap: 0.75rem;
                    font-size: 1.2rem;
                    font-weight: 500;
                    color: rgba(255, 255, 255, 0.9);
                    
                    .q-icon {
                        color: #ffd700;
                    }
                    
                    @media (max-width: 968px) {
                        justify-content: center;
                    }
                }
            }
            
            .star-rating-display {
                display: flex;
                align-items: center;
                gap: 1rem;
                
                @media (max-width: 968px) {
                    justify-content: center;
                }
                
                .overall-score {
                    font-size: 3rem;
                    font-weight: 800;
                    color: #ffd700;
                    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
                }
                
                .star-container {
                    display: flex;
                    gap: 4px;
                    
                    .star-modern {
                        font-size: 2rem;
                        transition: all 0.2s ease;
                        
                        &.star-full {
                            color: #ffd700;
                            text-shadow: 0 0 8px rgba(255, 215, 0, 0.6);
                        }
                        
                        &.star-half {
                            color: #ffd700;
                            opacity: 0.6;
                        }
                        
                        &.star-empty {
                            color: rgba(255, 255, 255, 0.3);
                        }
                    }
                }
                
                .rating-label {
                    font-size: 0.9rem;
                    color: rgba(255, 255, 255, 0.8);
                    font-weight: 500;
                }
            }
        }
        
        .performance-metrics {
            .metrics-grid {
                display: grid;
                grid-template-columns: repeat(3, 1fr);
                gap: 1rem;
                
                @media (max-width: 768px) {
                    grid-template-columns: repeat(3, 1fr);
                    gap: 0.75rem;
                }
                
                .metric-card {
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    gap: 0.5rem;
                    padding: 1.5rem 1rem;
                    border-radius: $border-radius-small;
                    background: rgba(255, 255, 255, 0.15);
                    border: 2px solid rgba(255, 255, 255, 0.2);
                    transition: all 0.3s ease;
                    backdrop-filter: blur(10px);
                    
                    &:hover {
                        transform: translateY(-2px);
                        background: rgba(255, 255, 255, 0.2);
                    }
                    
                    .metric-icon {
                        font-size: 2rem;
                        margin-bottom: 0.25rem;
                    }
                    
                    .metric-value {
                        font-size: 2rem;
                        font-weight: 800;
                        margin: 0.25rem 0;
                        color: white;
                        text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
                    }
                    
                    .metric-label {
                        font-size: 0.85rem;
                        font-weight: 600;
                        text-transform: uppercase;
                        letter-spacing: 0.5px;
                        color: rgba(255, 255, 255, 0.9);
                    }
                    
                    @media (max-width: 768px) {
                        padding: 1rem 0.75rem;
                        
                        .metric-icon {
                            font-size: 1.5rem;
                        }
                        
                        .metric-value {
                            font-size: 1.5rem;
                        }
                        
                        .metric-label {
                            font-size: 0.75rem;
                        }
                    }
                }
            }
        }
    }
}

@keyframes float {
    0%, 100% { transform: translateY(0px) rotate(0deg); }
    33% { transform: translateY(-20px) rotate(1deg); }
    66% { transform: translateY(-10px) rotate(-1deg); }
}

.formation-tactics-layout {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    margin-bottom: 2rem;
}

.formation-controls-panel {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.formation-display-panel {
    display: flex;
    flex-direction: column;
}

// Modern Card Styles
.formation-card,
.squad-depth-card,
.pitch-card,
.players-table-card {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    border: 1px solid rgba(0, 0, 0, 0.05);
    transition: all 0.3s ease;
    
    .body--dark & {
        border: 1px solid rgba(255, 255, 255, 0.1);
        background: rgba(255, 255, 255, 0.02);
    }
    
    &:hover {
        box-shadow: $card-shadow-hover;
        transform: translateY(-2px);
    }
    
    .card-header {
        margin-bottom: 2rem;
        
        .card-title {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-size: 1.4rem;
            font-weight: 700;
            margin: 0 0 0.5rem 0;
            color: #2d3436;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.9);
            }
            
            .card-icon {
                color: #667eea;
            }
        }
        
        .card-subtitle {
            font-size: 1rem;
            color: #636e72;
            margin: 0;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
}

// Formation Controls
.formation-controls {
    display: grid;
    grid-template-columns: minmax(260px, 380px) minmax(0, 1fr);
    gap: 1rem;
    align-items: start;

    .formation-select {
        margin-bottom: 0;
    }
    
    .calculation-banner {
        border-radius: $border-radius-small;
        font-weight: 500;
    }

    @media (max-width: 768px) {
        grid-template-columns: 1fr;
    }
}

.squad-depth-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
    
    @media (max-width: 768px) {
        grid-template-columns: 1fr;
        gap: 0.75rem;
    }
    
    .depth-position-modern {
        background: rgba(255, 255, 255, 0.6);
        border-radius: $border-radius-small;
        padding: 1rem;
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
            margin-bottom: 0.75rem;
            
            .position-name {
                font-weight: 700;
                font-size: 0.9rem;
                color: #2d3436;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .player-count {
                font-size: 0.75rem;
                color: #636e72;
                background: rgba(0, 0, 0, 0.05);
                padding: 0.2rem 0.5rem;
                border-radius: 12px;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                    background: rgba(255, 255, 255, 0.1);
                }
            }
        }
        
        .depth-players-modern {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            
            .player-card-mini {
                display: grid;
                grid-template-columns: auto 1fr auto;
                gap: 0.5rem;
                align-items: center;
                padding: 0.5rem;
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
                    transform: translateX(4px);
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
                    width: 20px;
                    height: 20px;
                    background: rgba(103, 126, 234, 0.1);
                    color: #667eea;
                    border-radius: 50%;
                    font-size: 0.7rem;
                    font-weight: 700;
                }
                
                .player-info {
                    min-width: 0;
                    
                    .player-name {
                        font-size: 0.8rem;
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
                        font-size: 0.65rem;
                        color: #636e72;
                        margin-top: 0.2rem;
                        
                        .body--dark & {
                            color: rgba(255, 255, 255, 0.6);
                        }
                    }
                }
                
                .player-rating {
                    font-size: 0.75rem;
                    font-weight: 700;
                    padding: 0.2rem 0.4rem;
                    border-radius: 4px;
                    min-width: 28px;
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
            gap: 0.3rem;
            padding: 1rem;
            color: #636e72;
            font-style: italic;
            font-size: 0.8rem;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.5);
            }
            
            .q-icon {
                font-size: 1.2rem;
            }
        }
    }
}

// Pitch Container
.pitch-stage {
    background: linear-gradient(135deg, #00b894 0%, #00cec9 100%);
    border-radius: $border-radius;
    padding: 0.75rem;
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

    :deep(.pitch-container--cards) {
        max-width: min(1280px, 100%);
        min-height: max(820px, calc(var(--formation-rows) * 176px));
        margin: 0 auto;
        padding: 16px 10px;
    }

    :deep(.pitch-container--cards .formation-row) {
        min-height: 166px;
    }

    :deep(.pitch-container--cards .player-slot) {
        min-height: 166px;
    }

    :deep(.player-card-slot) {
        width: 122px;
        height: 183px;
    }

    :deep(.pitch-player-card) {
        --card-scale: 0.43;
        --card-hover-lift: -10px;
        --card-transform-origin: top left;
    }

    :deep(.empty-card-slot) {
        width: 122px;
        height: 170px;
    }

    @media (max-width: 900px) {
        padding: 0.5rem;

        :deep(.pitch-container--cards) {
            min-height: max(700px, calc(var(--formation-rows) * 146px));
            padding: 12px 6px;
        }

        :deep(.pitch-container--cards .formation-row),
        :deep(.pitch-container--cards .player-slot) {
            min-height: 140px;
        }

        :deep(.player-card-slot) {
            width: 102px;
            height: 153px;
        }

        :deep(.pitch-player-card) {
            --card-scale: 0.36;
            --card-hover-lift: -10px;
            --card-transform-origin: top left;
        }

        :deep(.empty-card-slot) {
            width: 102px;
            height: 142px;
        }
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

// Responsive Design
@media (max-width: 1200px) {
    .formation-tactics-layout {
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
    
    .team-dashboard {
        gap: 1.5rem;
    }
    
    .team-hero-section {
        padding: 2rem;
    }
    
    .formation-tactics-layout {
        gap: 1rem;
    }
    
    .formation-controls-panel {
        gap: 1rem;
    }
}

// Dark mode enhancements
.body--dark {
    .team-hero-section {
        background: linear-gradient(135deg, #2d3436 0%, #636e72 100%);
    }
}

// Utility classes for rating colors (ensure these exist)
.rating-tier-6 { color: #8e24aa; background-color: rgba(142, 36, 170, 0.1); }
.rating-tier-5 { color: #1976d2; background-color: rgba(25, 118, 210, 0.1); }
.rating-tier-4 { color: #388e3c; background-color: rgba(56, 142, 60, 0.1); }
.rating-tier-3 { color: #f57c00; background-color: rgba(245, 124, 0, 0.1); }
.rating-tier-2 { color: #d32f2f; background-color: rgba(211, 47, 47, 0.1); }
.rating-tier-1 { color: #616161; background-color: rgba(97, 97, 97, 0.1); }

// Prestige Player Cards Section
.prestige-cards-section {
    margin: 2.5rem 0;

    .prestige-cards-row {
        display: flex;
        justify-content: center;
        align-items: flex-start;
        gap: 3rem;
        flex-wrap: wrap;
    }

    .prestige-card-wrapper {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.75rem;
    }

    .prestige-card-label {
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.12em;
        text-transform: uppercase;
        padding: 0.3rem 1rem;
        border-radius: 20px;

        &.icon-label {
            background: linear-gradient(135deg, #e9d8b4 0%, #ad8432 100%);
            color: #35240f;
        }

        &.hero-label {
            background: linear-gradient(135deg, #26204b 0%, #f2bf4d 100%);
            color: #fff2df;
        }

        &.motm-label {
            background: linear-gradient(135deg, #2d0d08 0%, #ff7f26 100%);
            color: #fff2e8;
        }
    }

    @media (max-width: 768px) {
        margin: 1.5rem 0;

        .prestige-cards-row {
            gap: 1.5rem;
        }
    }
}
</style>
