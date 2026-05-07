<template>
    <q-page class="leagues-page">
        <!-- Hero Section -->
        <div class="hero-section">
            <div class="hero-container">
                <div class="hero-content">
                    <div class="hero-badge">
                        <q-icon name="sports" size="1.2rem" />
                        <span>League Explorer</span>
                    </div>
                    <h1 class="hero-title">
                        League
                        <span class="gradient-text">Analytics</span>
                    </h1>
                    <p class="hero-subtitle">
                        Dive deep into leagues and competitions. Compare performance across different tournaments and discover emerging talents.
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
                    <h2 class="filter-title">League Selection</h2>
                    <p class="filter-subtitle">Choose a league to analyze teams and player distributions</p>
                </div>
                <div class="filter-card"
                     :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'">
                    <div class="filter-content">
                    <q-select
                        v-model="selectedLeagueName"
                        :options="leagueOptions"
                        label="Search and Select League"
                        outlined
                        dense
                        use-input
                        hide-selected
                        fill-input
                        input-debounce="300"
                        @filter="filterLeagueOptions"
                        @update:model-value="loadLeagueTeams"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : 'bg-white text-dark'
                        "
                        clearable
                        @clear="clearLeagueSelection"
                        :disable="pageLoading || allLeaguesData.length === 0"
                    >
                        <template v-slot:no-option>
                            <q-item>
                                <q-item-section class="text-grey">
                                    No leagues found.
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
                    Loading leagues data from server...
                </div>
            </div>
            <div v-else-if="loadingLeague" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="2em" />
                <div
                    class="q-mt-sm text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading league teams...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <!-- Leagues Overview Card -->
                <q-card
                    v-if="!selectedLeagueName && !loadingLeague && allLeaguesData.length > 0"
                    class="q-mb-md"
                    :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-md">Leagues Overview</div>
                        <div class="leagues-list">
                            <div
                                v-for="league in displayedLeagues"
                                :key="league.name"
                                class="league-row"
                                @click="selectLeague(league.name)"
                            >
                                <div class="league-info">
                                    <div class="league-name">{{ league.name }}</div>
                                    <div class="team-count">{{ league.teamCount }} teams</div>
                                    <div class="player-count">{{ league.playerCount }} players</div>
                                </div>
                                <div class="league-overall">
                                    <div class="star-rating-large">
                                        <span
                                            v-for="star in 5"
                                            :key="star"
                                            class="star-large"
                                            :class="getStarClass(league.bestOverall, star)"
                                        >
                                            ★
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Show More Button -->
                        <div v-if="!showAllLeagues && allLeaguesData.length > INITIAL_LEAGUES_LIMIT" class="text-center q-mt-md">
                            <q-btn
                                flat
                                color="primary"
                                @click="showAllLeagues = true"
                                class="show-more-btn"
                            >
                                Show All Leagues
                                <q-icon name="expand_more" class="q-ml-sm" />
                            </q-btn>
                        </div>
                    </q-card-section>
                </q-card>

                <!-- Selected League Teams -->
                <div v-if="selectedLeagueName && !loadingLeague && leagueTeams.length > 0">
                    <section
                        class="tots-lineup q-mb-md"
                        :class="quasarInstance.dark.isActive ? 'tots-lineup--dark' : 'tots-lineup--light'"
                    >
                        <div class="tots-lineup__header">
                            <div>
                                <div class="tots-lineup__kicker">Team of the Season</div>
                                <h2>{{ selectedLeagueName }} XI</h2>
                            </div>
                            <p>Top rated eligible players by position group, minimum 5 appearances.</p>
                        </div>

                        <div v-if="teamOfSeasonLineup.hasPlayers" class="tots-lineup__pitch">
                            <div
                                v-for="row in teamOfSeasonLineup.rows"
                                :key="row.group"
                                class="tots-lineup__row"
                                :class="`tots-lineup__row--${row.group.toLowerCase()}`"
                            >
                                <div
                                    v-for="slot in row.slots"
                                    :key="`${row.group}-${slot.index}`"
                                    class="tots-card-slot"
                                    :class="{ 'tots-card-slot--empty': !slot.player }"
                                >
                                    <PlayerCards
                                        v-if="slot.player"
                                        :player="slot.player"
                                        :currency-symbol="detectedCurrencySymbol"
                                        :dataset-id="currentDatasetId"
                                        :card-design-override="getTotsCardDesign(slot.player)"
                                        :position-override="slot.player.totsDisplayPosition"
                                        @click="openPlayerDetail(slot.player)"
                                    />
                                    <template v-else>
                                        <span>{{ row.group }}</span>
                                    </template>
                                </div>
                            </div>
                        </div>
                        <div v-else class="tots-lineup__empty">
                            <div>No eligible players found for this league.</div>
                            <div class="tots-lineup__empty-details">
                                {{ teamOfSeasonLineup.totalPlayers }} players,
                                {{ teamOfSeasonLineup.playersWithRating }} with rating,
                                {{ teamOfSeasonLineup.playersWithFiveApps }} with 5+ appearances.
                            </div>
                        </div>
                    </section>

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
                                {{ selectedLeagueName }} - Teams ({{ leagueTeams.length }})
                            </div>
                            <div class="teams-list">
                                <div
                                    v-for="team in leagueTeams"
                                    :key="team.name"
                                    class="team-row"
                                    @click="handleTeamSelected(team)"
                                >
                                    <div class="team-info">
                                        <div class="team-name">{{ team.name }}</div>
                                        <div class="team-player-count">{{ team.playerCount }} players</div>
                                    </div>
                                    <div class="team-overall">
                                        <div class="star-rating-large">
                                            <span
                                                v-for="star in 5"
                                                :key="star"
                                                class="star-large"
                                                :class="getStarClass(team.bestOverall, star)"
                                            >
                                                ★
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>
                </div>

                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingLeague &&
                        allLeaguesData.length > 0 &&
                        !selectedLeagueName
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
                    Select a league to view its teams and their ratings.
                </q-banner>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingLeague &&
                        allLeaguesData.length === 0 &&
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
                    No league data available. Please upload a player file with Division information on the main page first.
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
import PlayerCards from '../components/PlayerCards.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerUtils } from '../composables/usePlayerUtils'
import { usePlayerStore } from '../stores/playerStore'

export default {
  name: 'LeaguesPage',
  components: { PlayerDetailDialog, PlayerCards },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()
    const { positionGroups } = usePlayerUtils()

    const selectedLeagueName = ref(null)
    const leagueOptions = ref([])
    const allLeagueNamesCache = ref([])
    const allLeaguesData = ref([])
    const leagueTeams = ref([])
    const loadingLeague = ref(false)
    const pageLoading = ref(true)
    const pageLoadingError = ref('')

    // Pagination for leagues
    const showAllLeagues = ref(false)
    const INITIAL_LEAGUES_LIMIT = 30

    // Computed properties from store
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
    const currentDatasetId = computed(() => playerStore.currentDatasetId)

    // Limited leagues for initial rendering
    const displayedLeagues = computed(() => {
      if (!allLeaguesData.value || allLeaguesData.value.length === 0) return []

      if (!showAllLeagues.value && allLeaguesData.value.length > INITIAL_LEAGUES_LIMIT) {
        return allLeaguesData.value.slice(0, INITIAL_LEAGUES_LIMIT)
      }

      return allLeaguesData.value
    })

    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)

    const totsShape = [
      {
        group: 'ATT',
        slots: [
          {
            key: 'att-left',
            displayPreference: ['LW', 'LM', 'ST', 'RW', 'RM'],
            eligible: ['LW', 'AML', 'LM', 'ML', 'ST', 'STC', 'CF', 'RW', 'AMR', 'RM', 'MR'],
          },
          {
            key: 'att-centre',
            displayPreference: ['ST', 'LW', 'RW', 'LM', 'RM'],
            eligible: ['ST', 'STC', 'CF', 'LW', 'AML', 'LM', 'ML', 'RW', 'AMR', 'RM', 'MR'],
          },
          {
            key: 'att-right',
            displayPreference: ['RW', 'RM', 'ST', 'LW', 'LM'],
            eligible: ['RW', 'AMR', 'RM', 'MR', 'ST', 'STC', 'CF', 'LW', 'AML', 'LM', 'ML'],
          },
        ],
      },
      {
        group: 'MID',
        slots: Array.from({ length: 3 }, (_, index) => ({
          key: `mid-${index}`,
          displayPreference: ['CM', 'CDM', 'CAM'],
          eligible: ['CM', 'MC', 'CDM', 'DM', 'CAM', 'AMC'],
        })),
      },
      {
        group: 'DEF',
        slots: [
          {
            key: 'def-lb',
            displayPosition: 'LB',
            displayPreference: ['LB'],
            eligible: ['LB', 'DL', 'LWB', 'WBL'],
          },
          {
            key: 'def-lcb',
            displayPosition: 'CB',
            displayPreference: ['CB'],
            eligible: ['CB', 'DC', 'D', 'SW'],
          },
          {
            key: 'def-rcb',
            displayPosition: 'CB',
            displayPreference: ['CB'],
            eligible: ['CB', 'DC', 'D', 'SW'],
          },
          {
            key: 'def-rb',
            displayPosition: 'RB',
            displayPreference: ['RB'],
            eligible: ['RB', 'DR', 'RWB', 'WBR'],
          },
        ],
      },
      {
        group: 'GK',
        slots: [
          {
            key: 'gk',
            displayPosition: 'GK',
            displayPreference: ['GK'],
            eligible: ['GK', 'Goalkeeper'],
          },
        ],
      },
    ]

    const displayPositionAliases = {
      GK: ['GK', 'Goalkeeper'],
      LB: ['LB', 'DL', 'LWB', 'WBL'],
      CB: ['CB', 'DC', 'D', 'SW'],
      RB: ['RB', 'DR', 'RWB', 'WBR'],
      CM: ['CM', 'MC', 'M'],
      CDM: ['CDM', 'DM'],
      CAM: ['CAM', 'AMC', 'AM'],
      LW: ['LW', 'AML'],
      LM: ['LM', 'ML'],
      ST: ['ST', 'STC', 'CF'],
      RW: ['RW', 'AMR'],
      RM: ['RM', 'MR'],
    }

    const knownPositionCodes = Array.from(
      new Set([
        ...Object.values(displayPositionAliases).flat(),
        ...Object.values(positionGroups).flat(),
        'D',
        'M',
        'AM',
      ])
    )
      .map((position) => String(position).toUpperCase())
      .filter(Boolean)
      .sort((a, b) => b.length - a.length)

    const getPerformanceStat = (player, keys) => {
      const stats = player?.performanceStatsNumeric || player?.PerformanceStatsNumeric || {}
      const attrs = player?.attributes || {}

      for (const key of keys) {
        const raw = stats[key] ?? attrs[key]
        if (raw === null || raw === undefined || raw === '' || raw === '-') continue
        const value = Number(String(raw).replace('%', '').replace(',', '').trim())
        if (!Number.isNaN(value)) return value
      }

      return null
    }

    const getPlayerPositionCodes = (player) => {
      const codes = new Set()

      for (const pos of player?.shortPositions || player?.short_positions || []) {
        if (pos) codes.add(String(pos).toUpperCase())
      }

      const rawPosition = String(player?.position || '').toUpperCase()
      for (const code of knownPositionCodes) {
        const pattern = new RegExp(`(^|[^A-Z])${code}([^A-Z]|$)`)
        if (pattern.test(rawPosition)) codes.add(code)
      }

      const sideGroups = rawPosition.matchAll(/\b(D|WB|M|AM)\s*\(([^)]+)\)/g)
      for (const match of sideGroups) {
        const family = match[1]
        const sides = match[2]

        if (family === 'D') {
          codes.add('D')
          if (sides.includes('C')) codes.add('DC')
          if (sides.includes('R')) codes.add('DR')
          if (sides.includes('L')) codes.add('DL')
        }
        if (family === 'WB') {
          if (sides.includes('R')) codes.add('WBR')
          if (sides.includes('L')) codes.add('WBL')
        }
        if (family === 'M') {
          codes.add('M')
          if (sides.includes('C')) codes.add('MC')
          if (sides.includes('R')) codes.add('MR')
          if (sides.includes('L')) codes.add('ML')
        }
        if (family === 'AM') {
          codes.add('AM')
          if (sides.includes('C')) codes.add('AMC')
          if (sides.includes('R')) codes.add('AMR')
          if (sides.includes('L')) codes.add('AML')
        }
      }

      for (const parsedPosition of player?.parsedPositions || []) {
        const normalized = String(parsedPosition).toLowerCase()
        if (normalized.includes('goalkeeper')) codes.add('GK')
        if (
          normalized.includes('back') ||
          normalized.includes('defender') ||
          normalized.includes('sweeper')
        ) {
          codes.add('D')
        }
        if (normalized.includes('midfielder')) codes.add('M')
        if (normalized.includes('striker') || normalized.includes('winger')) codes.add('ST')
      }

      return codes
    }

    const canPlayAny = (codes, positions) => {
      return positions.some((position) => codes.has(String(position).toUpperCase()))
    }

    const canPlayDisplayPosition = (codes, position) => {
      return canPlayAny(codes, displayPositionAliases[position] || [position])
    }

    const getSlotDisplayPosition = (player, slot) => {
      if (slot.displayPosition) return slot.displayPosition
      const codes = getPlayerPositionCodes(player)

      for (const displayPosition of slot.displayPreference || []) {
        if (canPlayDisplayPosition(codes, displayPosition)) return displayPosition
      }

      return slot.displayPreference?.[0] || ''
    }

    const getSlotFitScore = (player, slot) => {
      const codes = getPlayerPositionCodes(player)
      const index = (slot.displayPreference || []).findIndex((displayPosition) =>
        canPlayDisplayPosition(codes, displayPosition)
      )

      return index === -1 ? 99 : index
    }

    const getOverallValue = (player) => {
      const overall = Number(player?.Overall ?? player?.overall ?? 0)
      return Number.isNaN(overall) ? 0 : overall
    }

    const getPlayerSelectionKey = (player) =>
      player?.uid || player?.UID || player?.id || player?.name

    const clampNumber = (value, min, max) => Math.min(max, Math.max(min, value))

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

    const createTotsPlayer = (player, displayPosition, leagueAverages) => {
      const baseOverall = getOverallValue(player)
      const leagueAverageOverall = leagueAverages.averageOverall || baseOverall || 65
      const ratingBonus = player.totsRating
        ? clampNumber(
            (player.totsRating - leagueAverages.averagePerformanceRating) * 1.3,
            -1.5,
            2.5
          )
        : 0
      const belowLeagueBoost = Math.max(0, leagueAverageOverall - baseOverall) * 0.68
      const aboveLeagueDiscount = Math.max(0, baseOverall - leagueAverageOverall) * 0.08
      const rawBoost = 8 + belowLeagueBoost - aboveLeagueDiscount + ratingBonus
      const targetOverall = Math.round(
        clampNumber(baseOverall + rawBoost, baseOverall + 4, displayPosition === 'GK' ? 94 : 96)
      )
      const relevance = statRelevanceByPosition[displayPosition] || statRelevanceByPosition.CM
      const statKeys = displayPosition === 'GK' ? goalkeeperStatKeys : outfieldStatKeys
      const positionBonus =
        displayPosition === 'GK'
          ? 0
          : ['LW', 'LM', 'RW', 'RM'].includes(displayPosition)
            ? 3
            : displayPosition === 'ST'
              ? 2
              : 1
      const boostedPlayer = {
        ...player,
        Overall: targetOverall,
        overall: targetOverall,
        totsBaseOverall: baseOverall,
        totsCardDesign: getTotsCardDesignForOverall(targetOverall),
        totsDisplayPosition: displayPosition,
      }

      for (const key of statKeys) {
        boostedPlayer[key] = getUpgradedStat(
          player[key],
          baseOverall,
          targetOverall,
          relevance[key] || 0.25,
          displayPosition,
          positionBonus
        )
      }

      return boostedPlayer
    }

    const teamOfSeasonLineup = computed(() => {
      const selectedLeague = selectedLeagueName.value
      const players = Array.isArray(playerStore.allPlayers) ? playerStore.allPlayers : []

      const eligiblePlayers = []
      let totalPlayers = 0
      let playersWithRating = 0
      let playersWithFiveApps = 0
      let overallTotal = 0
      let overallCount = 0
      let performanceRatingTotal = 0
      let performanceRatingCount = 0

      for (const player of players) {
        if (!player || player.division !== selectedLeague) continue
        totalPlayers++

        const rating = getPerformanceStat(player, ['Av Rat', 'Rating', 'Average Rating'])
        const appearances = getPerformanceStat(player, ['Apps', 'Appearances', 'Aps'])
        const overall = getOverallValue(player)

        if (rating !== null) playersWithRating++
        if (appearances !== null && appearances >= 5) playersWithFiveApps++
        if (overall > 0) {
          overallTotal += overall
          overallCount++
        }
        if (rating !== null) {
          performanceRatingTotal += rating
          performanceRatingCount++
        }

        if (rating === null || appearances === null || appearances < 5) continue
        eligiblePlayers.push({
          ...player,
          totsRating: rating,
          totsAppearances: appearances,
        })
      }

      const leagueAverages = {
        averageOverall: overallCount > 0 ? overallTotal / overallCount : 65,
        averagePerformanceRating:
          performanceRatingCount > 0 ? performanceRatingTotal / performanceRatingCount : 6.8,
      }
      const selectedIds = new Set()
      const rows = totsShape.map((row) => {
        return {
          group: row.group,
          slots: row.slots.map((slot, index) => {
            const candidates = eligiblePlayers
              .filter((player) => !selectedIds.has(getPlayerSelectionKey(player)))
              .filter((player) => canPlayAny(getPlayerPositionCodes(player), slot.eligible))
              .sort((a, b) => {
                const fitDifference = getSlotFitScore(a, slot) - getSlotFitScore(b, slot)
                if (fitDifference !== 0) return fitDifference
                if (b.totsRating !== a.totsRating) return b.totsRating - a.totsRating
                if (b.totsAppearances !== a.totsAppearances)
                  return b.totsAppearances - a.totsAppearances
                return getOverallValue(b) - getOverallValue(a)
              })

            const selectedPlayer = candidates[0] || null
            if (selectedPlayer) selectedIds.add(getPlayerSelectionKey(selectedPlayer))
            const displayPosition = selectedPlayer
              ? getSlotDisplayPosition(selectedPlayer, slot)
              : slot.displayPosition || slot.displayPreference?.[0] || row.group

            return {
              index,
              player: selectedPlayer
                ? createTotsPlayer(selectedPlayer, displayPosition, leagueAverages)
                : null,
            }
          }),
        }
      })

      return {
        rows,
        totalPlayers,
        playersWithRating,
        playersWithFiveApps,
        hasPlayers: rows.some((row) => row.slots.some((slot) => slot.player)),
      }
    })

    const fetchLeaguesAndCurrency = async (datasetId) => {
      pageLoading.value = true
      pageLoadingError.value = ''
      try {
        const response = await fetch(`/api/leagues/${datasetId}`)
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        const leaguesData = await response.json()
        allLeaguesData.value = leaguesData || []

        // Also fetch currency symbol from player store
        await playerStore.fetchPlayersByDatasetId(datasetId)
      } catch (err) {
        pageLoadingError.value = `Failed to load leagues data: ${err.message || 'Unknown server error'}. Please try uploading again.`
      } finally {
        pageLoading.value = false
      }
    }

    const fetchTeamsForLeague = async (datasetId, leagueName) => {
      loadingLeague.value = true
      try {
        const response = await fetch(`/api/teams/${datasetId}/${encodeURIComponent(leagueName)}`)
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        const teamsData = await response.json()
        leagueTeams.value = teamsData || []
      } catch (_err) {
        leagueTeams.value = []
      } finally {
        loadingLeague.value = false
      }
    }

    onMounted(async () => {
      const datasetIdFromQuery = route.query.datasetId
      const datasetIdFromRoute = route.params.datasetId
      const leagueFromQuery = route.query.league
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
        await fetchLeaguesAndCurrency(finalDatasetId)
      } else {
        pageLoadingError.value =
          'No player dataset ID found. Please upload a file on the main page.'
        pageLoading.value = false
      }

      if (!pageLoadingError.value && allLeaguesData.value.length > 0) {
        populateLeagueFilterOptions()

        if (leagueFromQuery && leagueFromQuery.trim() !== '') {
          selectedLeagueName.value = leagueFromQuery
          loadLeagueTeams()
        }
      }
    })

    const populateLeagueFilterOptions = () => {
      if (!allLeaguesData.value || allLeaguesData.value.length === 0) {
        allLeagueNamesCache.value = []
        leagueOptions.value = []
        return
      }
      const leagueNames = allLeaguesData.value.map((league) => league.name).sort()
      allLeagueNamesCache.value = leagueNames
      leagueOptions.value = leagueNames
    }

    const filterLeagueOptions = (val, update) => {
      if (val === '') {
        update(() => {
          leagueOptions.value = allLeagueNamesCache.value
        })
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        leagueOptions.value = allLeagueNamesCache.value.filter(
          (league) => league.toLowerCase().indexOf(needle) > -1
        )
      })
    }

    const selectLeague = (leagueName) => {
      selectedLeagueName.value = leagueName
      loadLeagueTeams()
    }

    const loadLeagueTeams = async () => {
      if (!selectedLeagueName.value) {
        leagueTeams.value = []
        return
      }

      const datasetId = currentDatasetId.value || sessionStorage.getItem('currentDatasetId')
      if (datasetId) {
        await fetchTeamsForLeague(datasetId, selectedLeagueName.value)
      }
    }

    const clearLeagueSelection = () => {
      selectedLeagueName.value = null
      leagueTeams.value = []
    }

    const handleTeamSelected = (team) => {
      // Navigate to the team-view page with team filtering
      const datasetId = currentDatasetId.value || sessionStorage.getItem('currentDatasetId')
      if (datasetId) {
        router.push({
          path: '/team-view',
          query: {
            datasetId: datasetId,
            team: team.name,
          },
        })
      }
    }

    const openPlayerDetail = (player) => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const getTotsCardDesignForOverall = (overall) => {
      if (overall >= 75) return 'card-design--gold-season-trophy'
      if (overall >= 65) return 'card-design--silver-season-laurel'
      return 'card-design--bronze-season-ribbon'
    }

    const getTotsCardDesign = (player) => {
      const overall = Number(player?.Overall ?? player?.overall ?? 0)
      return getTotsCardDesignForOverall(overall)
    }

    const getOverallClass = (overall) => {
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

    const shareDataset = async () => {
      if (!currentDatasetId.value) return

      const shareUrl = `${window.location.origin}/leagues/${currentDatasetId.value}`

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

    watch(
      () => route.query.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          sessionStorage.setItem('currentDatasetId', newId)
          await fetchLeaguesAndCurrency(newId)
          clearLeagueSelection()
          if (!pageLoadingError.value && allLeaguesData.value.length > 0) {
            populateLeagueFilterOptions()
          }
        }
      }
    )

    watch(
      () => route.query.league,
      (newLeague) => {
        if (newLeague && newLeague.trim() !== '' && newLeague !== selectedLeagueName.value) {
          selectedLeagueName.value = newLeague
          loadLeagueTeams()
        }
      }
    )

    return {
      allLeaguesData,
      selectedLeagueName,
      leagueOptions,
      filterLeagueOptions,
      loadLeagueTeams,
      clearLeagueSelection,
      selectLeague,
      leagueTeams,
      loadingLeague,
      pageLoading,
      pageLoadingError,
      playerForDetailView,
      showPlayerDetailDialog,
      handleTeamSelected,
      openPlayerDetail,
      teamOfSeasonLineup,
      getTotsCardDesign,
      getOverallClass,
      getStarClass,
      getStarRating,
      quasarInstance,
      router,
      detectedCurrencySymbol,
      currentDatasetId,
      shareDataset,
      displayedLeagues,
      showAllLeagues,
      INITIAL_LEAGUES_LIMIT,
    }
  },
}
</script>

<style lang="scss" scoped>
.leagues-page {
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

.tots-lineup {
    border-radius: 8px;
    padding: 1.25rem;
    border: 1px solid rgba(25, 90, 120, 0.22);
    background:
        linear-gradient(145deg, rgba(10, 48, 70, 0.96), rgba(17, 89, 94, 0.94) 48%, rgba(132, 105, 38, 0.9)),
        repeating-linear-gradient(90deg, rgba(255, 255, 255, 0.06) 0 1px, transparent 1px 18px);
    box-shadow: 0 18px 42px rgba(0, 0, 0, 0.18);
    color: white;

    &--dark {
        border-color: rgba(255, 215, 128, 0.28);
    }
}

.tots-lineup__header {
    display: flex;
    justify-content: space-between;
    gap: 1.5rem;
    align-items: flex-start;
    margin-bottom: 1rem;

    h2 {
        margin: 0;
        font-size: 1.35rem;
        font-weight: 800;
        letter-spacing: 0;
    }

    p {
        max-width: 360px;
        margin: 0;
        color: rgba(255, 255, 255, 0.78);
        font-size: 0.9rem;
        line-height: 1.45;
        text-align: right;
    }
}

.tots-lineup__kicker {
    color: #f4d06f;
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0;
    text-transform: uppercase;
}

.tots-lineup__pitch {
    display: grid;
    gap: 0.85rem;
    padding: 1rem;
    border-radius: 8px;
    background:
        linear-gradient(90deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px),
        linear-gradient(0deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px),
        rgba(1, 28, 39, 0.42);
    background-size: 52px 52px;
    border: 1px solid rgba(255, 255, 255, 0.12);
}

.tots-lineup__empty {
    padding: 1.25rem;
    border-radius: 8px;
    border: 1px dashed rgba(255, 255, 255, 0.28);
    background: rgba(1, 28, 39, 0.32);
    color: rgba(255, 255, 255, 0.82);
    font-size: 0.92rem;
    font-weight: 700;
    line-height: 1.4;
    text-align: center;
}

.tots-lineup__empty-details {
    margin-top: 0.35rem;
    color: rgba(255, 255, 255, 0.68);
    font-size: 0.82rem;
}

.tots-lineup__row {
    display: grid;
    justify-content: center;
    gap: clamp(0.65rem, 2vw, 1.25rem);

    &--att,
    &--mid {
        grid-template-columns: repeat(3, 154px);
    }

    &--def {
        grid-template-columns: repeat(4, 154px);
    }

    &--gk {
        grid-template-columns: 154px;
    }
}

.tots-card-slot {
    position: relative;
    width: 154px;
    height: 231px;
    display: flex;
    justify-content: center;
    align-items: flex-start;

    :deep(.fifa-card) {
        flex: 0 0 auto;
        transform: scale(0.55);
        transform-origin: top center;
    }

    :deep(.fifa-card:hover),
    :deep(.fifa-card:focus-within) {
        transform: scale(0.55) translateY(-6px);
    }

    &--empty {
        min-height: 124px;
        display: grid;
        place-items: center;
        border: 1px dashed rgba(255, 255, 255, 0.24);
        border-radius: 8px;
        color: rgba(255, 255, 255, 0.72);
        background: rgba(255, 255, 255, 0.08);
        font-size: 0.78rem;
        font-weight: 800;
    }
}

.tots-card-slot--empty {
    height: auto;
}

.tots-card-slot--empty span {
    display: inline-grid;
    min-width: 2.8rem;
    min-height: 2.8rem;
    place-items: center;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
}

.leagues-list,
.teams-list {
    max-height: 600px;
    overflow-y: auto;
}

.league-row,
.team-row {
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

.league-info,
.team-info {
    flex-shrink: 0;
    min-width: 150px;
}

.league-name,
.team-name {
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

.team-count,
.player-count,
.team-player-count {
    font-size: 0.85rem;
    color: $grey-6;
    
    .body--dark & {
        color: $grey-4;
    }
}

.league-section-ratings,
.team-section-ratings {
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

.league-overall,
.team-overall {
    flex-shrink: 0;
    margin-left: auto;
    display: flex;
    justify-content: flex-end;
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
    gap: 4px;
    font-size: 2rem;
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

@media (max-width: 760px) {
    .tots-lineup {
        padding: 0.85rem;
    }

    .tots-lineup__header {
        display: grid;
        gap: 0.5rem;

        p {
            max-width: none;
            text-align: left;
        }
    }

    .tots-lineup__pitch {
        gap: 0.7rem;
        padding: 0.65rem;
    }

    .tots-lineup__row {
        gap: 0.45rem;

        &--att,
        &--mid {
            grid-template-columns: repeat(3, 92px);
        }

        &--def {
            grid-template-columns: repeat(4, 92px);
        }

        &--gk {
            grid-template-columns: 92px;
        }
    }

    .tots-card-slot {
        width: 92px;
        height: 138px;

        :deep(.fifa-card) {
            transform: scale(0.329);
        }

        :deep(.fifa-card:hover),
        :deep(.fifa-card:focus-within) {
            transform: scale(0.329) translateY(-6px);
        }

        &--empty {
            min-height: 74px;
        }
    }
}

@media (max-width: 420px) {
    .tots-lineup__row {
        gap: 0.25rem;

        &--att,
        &--mid {
            grid-template-columns: repeat(3, 68px);
        }

        &--def {
            grid-template-columns: repeat(4, 68px);
        }

        &--gk {
            grid-template-columns: 68px;
        }
    }

    .tots-card-slot {
        width: 68px;
        height: 102px;

        :deep(.fifa-card) {
            transform: scale(0.243);
        }

        :deep(.fifa-card:hover),
        :deep(.fifa-card:focus-within) {
            transform: scale(0.243) translateY(-6px);
        }
    }
}
</style>
