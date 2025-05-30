<template>
    <q-page class="nations-page">
        <!-- Hero Section -->
        <div class="hero-section">
            <div class="hero-container">
                <div class="hero-content">
                    <div class="hero-badge">
                        <q-icon name="flag" size="1.2rem" />
                        <span>Global Insights</span>
                    </div>
                    <h1 class="hero-title">
                        Nations
                        <span class="gradient-text">Analysis</span>
                    </h1>
                    <p class="hero-subtitle">
                        Explore player distributions by nationality. Discover hidden gems and understand global talent pools across different countries.
                    </p>
                </div>
            </div>
        </div>
        
        <div class="q-pa-md">

            <q-banner
                v-if="pageLoadingError"
                class="text-white bg-negative q-mb-md"
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

            <!-- Share Button -->
            <div v-if="!pageLoadingError && currentDatasetId" class="share-button-container">
                <q-btn
                    unelevated
                    icon="share"
                    label="Share Dataset"
                    color="positive"
                    @click="shareDataset"
                    class="share-btn-enhanced"
                    size="md"
                >
                    <q-tooltip>Copy shareable link to clipboard</q-tooltip>
                </q-btn>
            </div>

            <div v-if="!pageLoadingError" class="modern-filter-section">
                <div class="filter-header">
                    <h2 class="filter-title">Nation Selection</h2>
                    <p class="filter-subtitle">Choose a nation to analyze player talents and distributions</p>
                </div>
                <div class="filter-card"
                     :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'">
                    <div class="filter-content">
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
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : 'bg-white text-dark'
                        "
                        clearable
                        @clear="clearNationSelection"
                        :disable="pageLoading || allPlayersData.length === 0"
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
                </div>
            </div>

            <div v-if="pageLoading" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="3em" />
                <div
                    class="q-mt-md text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading player data from server...
                </div>
            </div>
            <div v-else-if="loadingNation" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="2em" />
                <div
                    class="q-mt-sm text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading nation data...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <!-- Nations Overview Card -->
                <q-card
                    v-if="!selectedNationName && !loadingNation && allPlayersData.length > 0"
                    class="q-mb-md"
                    :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-md">Nations Overview</div>
                        <div class="nations-list">
                            <div
                                v-for="nation in nationsWithRatings"
                                :key="nation.name"
                                class="nation-row"
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
                                    <div 
                                        v-if="nation.bestFormationOverall > 0"
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
                                    <div 
                                        v-else
                                        class="no-rating-message"
                                    >
                                        Incomplete Squad
                                    </div>
                                </div>
                                <div class="nation-overall">
                                    <div class="nation-rating">
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

                <!-- Selected Nation Details -->
                <div v-if="selectedNationName && !loadingNation">
                    <q-card
                        :class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-9'
                                : 'bg-white'
                        "
                        class="q-mb-md"
                    >
                        <q-card-section>
                            <div class="text-h6 q-mb-md">
                                {{ selectedNationName }} - Best Formation & XI
                            </div>
                            <div class="row q-col-gutter-md items-start">
                                <div class="col-12 col-md-4">
                                    <q-select
                                        v-model="selectedFormationKey"
                                        :options="formationOptions"
                                        label="Select Formation"
                                        outlined
                                        dense
                                        emit-value
                                        map-options
                                        :label-color="
                                            quasarInstance.dark.isActive
                                                ? 'grey-4'
                                                : ''
                                        "
                                        :popup-content-class="
                                            quasarInstance.dark.isActive
                                                ? 'bg-grey-8 text-white'
                                                : 'bg-white text-dark'
                                        "
                                    />
                                    <div
                                        v-if="bestNationAverageOverall !== null"
                                        class="q-mt-md"
                                    >
                                        <div class="text-subtitle1 q-mb-sm"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-3'
                                                    : 'text-grey-8'
                                            "
                                        >
                                            Best XI Average Overall:
                                            <span
                                                class="text-weight-bold attribute-value"
                                                :class="
                                                    getOverallClass(
                                                        bestNationAverageOverall,
                                                    )
                                                "
                                            >
                                                {{ bestNationAverageOverall }}
                                            </span>
                                        </div>
                                        <div 
                                            v-if="currentNationSectionRatings.attRating > 0"
                                            class="section-ratings-detail"
                                        >
                                            <div class="section-rating-detail att">
                                                <span class="section-label-detail">ATT</span>
                                                <span 
                                                    class="section-value-detail"
                                                    :class="getOverallClass(currentNationSectionRatings.attRating)"
                                                >
                                                    {{ currentNationSectionRatings.attRating }}
                                                </span>
                                            </div>
                                            <div class="section-rating-detail mid">
                                                <span class="section-label-detail">MID</span>
                                                <span 
                                                    class="section-value-detail"
                                                    :class="getOverallClass(currentNationSectionRatings.midRating)"
                                                >
                                                    {{ currentNationSectionRatings.midRating }}
                                                </span>
                                            </div>
                                            <div class="section-rating-detail def">
                                                <span class="section-label-detail">DEF</span>
                                                <span 
                                                    class="section-value-detail"
                                                    :class="getOverallClass(currentNationSectionRatings.defRating)"
                                                >
                                                    {{ currentNationSectionRatings.defRating }}
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                    <q-banner
                                        v-if="calculationMessage"
                                        class="q-mt-sm"
                                        :class="calculationMessageClass"
                                    >
                                        {{ calculationMessage }}
                                    </q-banner>
                                    
                                    <!-- Compact Squad Depth -->
                                    <div 
                                        v-if="selectedFormationKey && Object.keys(squadComposition).length > 0"
                                        class="q-mt-md"
                                    >
                                        <div class="text-subtitle2 text-weight-bold q-mb-sm">Squad Depth</div>
                                        <div class="compact-squad-depth">
                                            <div
                                                v-for="slot in currentFormationLayout.flatMap(
                                                    (row) => row.positions,
                                                )"
                                                :key="slot.id"
                                                class="depth-position-compact"
                                            >
                                                <div class="position-label">
                                                    {{
                                                        getSlotDisplayName(
                                                            slot,
                                                            currentFormationLayout.flatMap(
                                                                (r) => r.positions,
                                                            ),
                                                        )
                                                    }}
                                                </div>
                                                <div 
                                                    v-if="squadComposition[slot.id] && squadComposition[slot.id].length > 0"
                                                    class="depth-players-compact"
                                                >
                                                    <div
                                                        v-for="(playerEntry, index) in squadComposition[slot.id].slice(0, 3)"
                                                        :key="playerEntry.player.name + '-' + slot.id + '-' + index"
                                                        class="depth-player-compact"
                                                        :class="{ 'starter': index === 0, 'backup': index > 0 }"
                                                        @click="handlePlayerSelectedFromNation(playerEntry.player)"
                                                    >
                                                        <span class="player-rank-compact">{{ index + 1 }}.</span>
                                                        <span class="player-name-compact">{{ playerEntry.player.name }}</span>
                                                        <span 
                                                            class="overall-compact"
                                                            :class="getOverallClass(playerEntry.overallInRole)"
                                                        >
                                                            {{ playerEntry.overallInRole }}
                                                        </span>
                                                    </div>
                                                </div>
                                                <div 
                                                    v-else
                                                    class="no-players-compact"
                                                >
                                                    No players
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div class="col-12 col-md-8">
                                    <PitchDisplay
                                        :formation="currentFormationLayout"
                                        :players="bestNationPlayersForPitch"
                                        @player-click="
                                            handlePlayerSelectedFromNation
                                        "
                                        @player-moved="handlePlayerMovedOnPitch"
                                    />
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>

                    <q-card
                        class="q-mb-md"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-9'
                                : 'bg-white'
                        "
                    >
                        <q-card-section>
                            <div class="text-h6 q-mb-sm">
                                Players from {{ selectedNationName }} ({{
                                    nationPlayers.length
                                }})
                            </div>
                            <PlayerDataTable
                                v-if="nationPlayers.length > 0"
                                :players="nationPlayers"
                                :loading="false"
                                @player-selected="handlePlayerSelectedFromNation"
                                :is-goalkeeper-view="nationIsGoalkeeperView"
                                :currency-symbol="detectedCurrencySymbol"
                                table-style="max-height: 400px;"
                                class="nation-player-table"
                            />
                            <q-banner
                                v-else
                                class="text-center"
                                :class="
                                    quasarInstance.dark.isActive
                                        ? 'bg-grey-8 text-grey-4'
                                        : 'bg-grey-2 text-grey-7'
                                "
                            >
                                No players found for this nation with the current
                                data.
                            </q-banner>
                        </q-card-section>
                    </q-card>

                </div>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingNation &&
                        allPlayersData.length > 0 &&
                        !selectedNationName
                    "
                    class="text-center q-mt-lg"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'bg-blue-grey-8 text-blue-grey-2'
                            : 'bg-blue-1 text-primary'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="info" />
                    </template>
                    Select a nation to view its best formation and players.
                </q-banner>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingNation &&
                        allPlayersData.length === 0 &&
                        !pageLoadingError
                    "
                    class="text-center q-mt-lg"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'bg-red-9 text-red-2'
                            : 'bg-red-1 text-negative'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="warning" />
                    </template>
                    No player data available. Please upload a player file on the
                    main page first.
                    <q-btn
                        flat
                        color="primary"
                        label="Go to Upload Page"
                        @click="router.push('/')"
                        class="q-ml-md"
                    />
                </q-banner>
            </div>
        </div>
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
import { debounce } from '../utils/debounce'
import { formationCache } from '../utils/formationCache'
import { formations, getFormationLayout } from '../utils/formations'

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

    const nationsWithRatings = computed(() => {
      if (!allPlayersData.value || allPlayersData.value.length === 0) return []

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
              bestFormationOverall: 0,
              players: []
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

      // Second pass: get top players per position for each nation and calculate best formation overall
      const nationsArray = Array.from(nationsMap.values())
      for (const nation of nationsArray) {
        // Get top 10 players per position to optimize performance
        const topPlayersByPosition = {}
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

        for (const position of allPositions) {
          const playersForPosition = nation.players.filter(player => {
            const playerPositions = player.shortPositions || []
            return playerPositions.includes(position)
          })

          // Sort by Overall and take top 10
          playersForPosition.sort((a, b) => (b.Overall || 0) - (a.Overall || 0))
          topPlayersByPosition[position] = playersForPosition.slice(0, 10)
        }

        let bestOverall = 0
        let hasFullSquad = false
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

          // Calculate player assignments for this formation using only top players
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

              if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
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

          // Check if we have a full squad (player in every position)
          const filledPositions = Object.values(tempSquadComposition).filter(
            slotPlayers => slotPlayers.length > 0
          ).length
          const isFullSquad = filledPositions === formationSlots.length

          if (isFullSquad) {
            hasFullSquad = true

            // Calculate average overall for this formation
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
        }

        // Only set overall if nation has at least one full squad possible
        nation.bestFormationOverall = hasFullSquad ? bestOverall : 0
        nation.attRating = hasFullSquad ? bestSectionRatings.attRating : 0
        nation.midRating = hasFullSquad ? bestSectionRatings.midRating : 0
        nation.defRating = hasFullSquad ? bestSectionRatings.defRating : 0
      }

      const sortedNations = nationsArray.sort(
        (a, b) => b.bestFormationOverall - a.bestFormationOverall
      )

      // Limit initial rendering for performance
      if (!showAllNations.value && sortedNations.length > INITIAL_NATIONS_LIMIT) {
        return sortedNations.slice(0, INITIAL_NATIONS_LIMIT)
      }

      return sortedNations
    })

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
      return nationPlayers.value.some(p => p.positionGroups?.includes('Goalkeepers'))
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
        return 0
      }

      const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls)
        ? player.roleSpecificOveralls.length > 0
        : Object.keys(player.roleSpecificOveralls).length > 0

      if (!hasRoleOveralls) {
        return 0
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
            bestScoreForRole = MIN_SUITABILITY_THRESHOLD
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
            bestScoreForRole = MIN_SUITABILITY_THRESHOLD - 10
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
      }

      return bestScoreForRole
    }

    const MIN_SUITABILITY_THRESHOLD = 40

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
      newVal => {
        if (pageLoading.value) return
        if (newVal && newVal.length > 0) {
          populateNationFilterOptions()
          if (selectedNationName.value) loadNationPlayers()
        } else if (!pageLoadingError.value) {
          clearNationSelection()
          allNationNamesCache.value = []
          nationOptions.value = []
        }
      },
      { deep: true }
    )

    watch(
      () => route.query.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          sessionStorage.setItem('currentDatasetId', newId)
          await fetchPlayersAndCurrency(newId)
          clearNationSelection()
          if (!pageLoadingError.value && allPlayersData.value.length > 0) {
            populateNationFilterOptions()
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
      showAllNations
    }
  }
}
</script>

<style lang="scss" scoped>
.nations-page {
    max-width: 1600px;
    margin: 0 auto;
}

// Hero Section
.hero-section {
    padding: 4rem 0;
    background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
    color: white;
    position: relative;
    overflow: hidden;
    margin: -1.5rem -1.5rem 2rem -1.5rem;
    
    &::before {
        content: "";
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: radial-gradient(
            circle at 30% 20%,
            rgba(255, 255, 255, 0.05) 0%,
            transparent 50%
        );
        pointer-events: none;
    }
    
    .hero-container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 0 2rem;
        display: flex;
        justify-content: center;
        align-items: center;
        position: relative;
        z-index: 1;
    }
    
    .hero-content {
        text-align: center;
        
        .hero-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background: rgba(255, 255, 255, 0.1);
            padding: 0.5rem 1rem;
            border-radius: 20px;
            font-size: 0.85rem;
            font-weight: 500;
            margin-bottom: 2rem;
            backdrop-filter: blur(10px);
        }
        
        .hero-title {
            font-size: 3.5rem;
            font-weight: 700;
            line-height: 1.1;
            margin: 0 0 1.5rem 0;
            
            @media (max-width: 768px) {
                font-size: 2.5rem;
            }
            
            .gradient-text {
                background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%);
                -webkit-background-clip: text;
                -webkit-text-fill-color: transparent;
                background-clip: text;
            }
        }
        
        .hero-subtitle {
            font-size: 1.2rem;
            line-height: 1.6;
            margin: 0;
            opacity: 0.9;
            font-weight: 300;
            max-width: 600px;
            margin: 0 auto;
            
            @media (max-width: 768px) {
                font-size: 1.1rem;
            }
        }
    }
}

.share-button-container {
    display: flex;
    justify-content: flex-end;
    margin: 2rem 0;
    padding: 0 2rem;
}

// Modern Filter Section
.modern-filter-section {
    margin: 3rem 0;
    
    .filter-header {
        text-align: center;
        margin-bottom: 2rem;
        
        .filter-title {
            font-size: 2rem;
            font-weight: 700;
            margin: 0 0 0.5rem 0;
            color: #1a237e;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.9);
            }
        }
        
        .filter-subtitle {
            font-size: 1rem;
            color: #666;
            margin: 0;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
    
    .filter-card {
        background: white;
        border-radius: 16px;
        padding: 2rem;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
        border: 1px solid rgba(0, 0, 0, 0.05);
        max-width: 600px;
        margin: 0 auto;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .filter-content {
            .q-field {
                .q-field__control {
                    border-radius: 12px;
                }
            }
        }
    }
}

.page-title {
    // Standard title styling
}

.filter-card,
.q-card {
    border-radius: $generic-border-radius;
}

.nation-player-table {
    :deep(.q-table__container) {
        max-height: 450px;
        overflow-y: auto;
    }
    :deep(th) {
        position: sticky;
        top: 0;
        z-index: 1;
    }
    .body--dark & :deep(th) {
        background-color: $grey-9 !important;
    }
    .body--light & :deep(th) {
        background-color: $grey-2 !important;
    }
}

.nations-list {
    max-height: 600px;
    overflow-y: auto;
}

.nation-row {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-radius: 6px;
    border: 1px solid rgba(0, 0, 0, 0.1);
    margin-bottom: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    
    .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.1);
        
        &:hover {
            background-color: rgba(255, 255, 255, 0.08);
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
        }
    }
    
    .body--light & {
        background-color: rgba(0, 0, 0, 0.02);
        border-color: rgba(0, 0, 0, 0.1);
        
        &:hover {
            background-color: rgba(0, 0, 0, 0.05);
        }
    }
}

.nation-flag-container {
    width: 32px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 16px;
    flex-shrink: 0;
}

.nationality-flag {
    border: 1px solid rgba(0, 0, 0, 0.1);
    border-radius: 2px;
    
    .body--dark & {
        border-color: rgba(255, 255, 255, 0.2);
    }
}

.nation-info {
    flex-shrink: 0;
    min-width: 120px;
}

.nation-name {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 2px;
    
    .body--dark & {
        color: $grey-2;
    }
    
    .body--light & {
        color: $grey-8;
    }
}

.player-count {
    font-size: 0.85rem;
    color: $grey-6;
    
    .body--dark & {
        color: $grey-4;
    }
}

.nation-section-ratings {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0 16px;
}

.section-ratings-large {
    display: flex;
    gap: 20px;
    align-items: center;
}

.section-rating-large {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    
    .section-label-large {
        font-size: 0.8rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.8px;
    }
    
    .section-value-large {
        font-size: 1.1rem;
        font-weight: bold;
        padding: 4px 8px;
        border-radius: 6px;
        min-width: 32px;
        text-align: center;
        border: 1px solid rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            border-color: rgba(255, 255, 255, 0.1);
        }
    }
    
    &.att .section-label-large {
        color: #F44336; // Red for attack
        
        .body--dark & {
            color: #FF5722;
        }
    }
    
    &.mid .section-label-large {
        color: #2196F3; // Blue for midfield
        
        .body--dark & {
            color: #03A9F4;
        }
    }
    
    &.def .section-label-large {
        color: #4CAF50; // Green for defense
        
        .body--dark & {
            color: #8BC34A;
        }
    }
}

.no-rating-message {
    font-size: 0.9rem;
    color: $grey-6;
    font-style: italic;
    text-align: center;
    
    .body--dark & {
        color: $grey-5;
    }
}

.nation-overall {
    flex-shrink: 0;
    margin-left: 16px;
}

.nation-rating {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    min-width: 120px;
}

.highest-overall-large {
    font-weight: bold;
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 1rem;
    text-align: center;
    min-width: 60px;
}

.star-rating-large {
    display: flex;
    gap: 2px;
    font-size: 1.2rem;
    line-height: 1;
}

.star-large {
    transition: color 0.2s ease;
    
    &.star-full {
        color: #FFD700; // Gold
    }
    
    &.star-half {
        background: linear-gradient(90deg, #FFD700 50%, transparent 50%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        
        // Fallback for browsers that don't support background-clip: text
        @supports not (-webkit-background-clip: text) {
            color: #FFD700;
        }
    }
    
    &.star-empty {
        color: #E0E0E0;
        
        .body--dark & {
            color: #424242;
        }
    }
}

// Keep the original smaller versions for compatibility
.highest-overall {
    font-weight: bold;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8rem;
    text-align: center;
    min-width: 50px;
}

.star-rating {
    display: flex;
    gap: 1px;
    font-size: 0.9rem;
    line-height: 1;
}

.star {
    transition: color 0.2s ease;
    
    &.star-full {
        color: #FFD700; // Gold
    }
    
    &.star-half {
        background: linear-gradient(90deg, #FFD700 50%, transparent 50%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        
        // Fallback for browsers that don't support background-clip: text
        @supports not (-webkit-background-clip: text) {
            color: #FFD700;
        }
    }
    
    &.star-empty {
        color: #E0E0E0;
        
        .body--dark & {
            color: #424242;
        }
    }
}

.section-ratings {
    display: flex;
    gap: 8px;
    margin-top: 4px;
}

.section-rating {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    
    .section-label {
        font-size: 0.65rem;
        font-weight: 600;
        color: $grey-6;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        
        .body--dark & {
            color: $grey-4;
        }
    }
    
    .section-value {
        font-size: 0.7rem;
        font-weight: bold;
        padding: 1px 4px;
        border-radius: 3px;
        min-width: 20px;
        text-align: center;
    }
    
    &.att .section-label {
        color: #F44336; // Red for attack
        
        .body--dark & {
            color: #FF5722;
        }
    }
    
    &.mid .section-label {
        color: #2196F3; // Blue for midfield
        
        .body--dark & {
            color: #03A9F4;
        }
    }
    
    &.def .section-label {
        color: #4CAF50; // Green for defense
        
        .body--dark & {
            color: #8BC34A;
        }
    }
}

.section-ratings-detail {
    display: flex;
    gap: 12px;
    margin-top: 8px;
}

.section-rating-detail {
    display: flex;
    align-items: center;
    gap: 6px;
    
    .section-label-detail {
        font-size: 0.75rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        min-width: 28px;
    }
    
    .section-value-detail {
        font-size: 0.8rem;
        font-weight: bold;
        padding: 2px 6px;
        border-radius: 4px;
        min-width: 28px;
        text-align: center;
    }
    
    &.att .section-label-detail {
        color: #F44336; // Red for attack
        
        .body--dark & {
            color: #FF5722;
        }
    }
    
    &.mid .section-label-detail {
        color: #2196F3; // Blue for midfield
        
        .body--dark & {
            color: #03A9F4;
        }
    }
    
    &.def .section-label-detail {
        color: #4CAF50; // Green for defense
        
        .body--dark & {
            color: #8BC34A;
        }
    }
}

.attribute-value {
    display: inline-block;
    min-width: 28px;
    text-align: center;
    font-weight: 600;
    padding: 2px 5px;
    border-radius: 4px;
    line-height: 1.3;
    font-size: 0.8em;
}

.compact-squad-depth {
    max-height: 500px;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 8px;
    width: 100%;
    
    .depth-position-compact {
        padding: 8px;
        border-radius: 6px;
        border: 1px solid rgba(0, 0, 0, 0.1);
        min-height: 100px;
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.08);
            border-color: rgba(255, 255, 255, 0.1);
        }
        
        .body--light & {
            background-color: rgba(0, 0, 0, 0.05);
            border-color: rgba(0, 0, 0, 0.1);
        }
    }
    
    .position-label {
        font-size: 0.75rem;
        font-weight: 700;
        margin-bottom: 6px;
        text-align: center;
        color: $grey-7;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        
        .body--dark & {
            color: $grey-3;
        }
    }
    
    .depth-players-compact {
        display: flex;
        flex-direction: column;
        gap: 3px;
    }
    
    .depth-player-compact {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 3px 6px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.75rem;
        min-height: 22px;
        
        &.starter {
            font-weight: 700;
            background-color: rgba($positive, 0.15);
            border: 1px solid rgba($positive, 0.3);
            
            .body--dark & {
                background-color: rgba($positive, 0.25);
                border-color: rgba($positive, 0.4);
            }
        }
        
        &.backup {
            font-weight: 500;
            background-color: rgba($grey-5, 0.1);
            
            .body--dark & {
                background-color: rgba($grey-5, 0.15);
            }
        }
        
        &:hover {
            background-color: rgba($primary, 0.15);
            transform: translateY(-1px);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            
            .body--dark & {
                background-color: rgba($primary, 0.25);
            }
        }
    }
    
    .player-rank-compact {
        font-size: 0.65rem;
        font-weight: bold;
        min-width: 14px;
        color: $grey-6;
        
        .body--dark & {
            color: $grey-4;
        }
    }
    
    .player-name-compact {
        flex: 1;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        font-size: 0.72rem;
    }
    
    .overall-compact {
        font-size: 0.7rem;
        font-weight: bold;
        padding: 2px 4px;
        border-radius: 3px;
        min-width: 24px;
        text-align: center;
        border: 1px solid rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            border-color: rgba(255, 255, 255, 0.1);
        }
    }
    
    .no-players-compact {
        font-size: 0.7rem;
        color: $grey-6;
        font-style: italic;
        text-align: center;
        padding: 8px;
        
        .body--dark & {
            color: $grey-5;
        }
    }
}

.share-btn-enhanced {
    font-weight: 600;
    border-radius: 6px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    }
    
    .body--dark & {
        box-shadow: 0 2px 4px rgba(255, 255, 255, 0.1);
        
        &:hover {
            box-shadow: 0 4px 8px rgba(255, 255, 255, 0.15);
        }
    }
}

.show-more-btn {
    font-weight: 500;
    border-radius: 8px;
    padding: 8px 24px;
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
        }
    }
}
</style>