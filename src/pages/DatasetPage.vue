<template>
    <q-page class="dataset-page">
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
                    Loading dataset...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <!-- Hero Section -->
                <div class="hero-section">
                    <div class="hero-container">
                        <div class="hero-content">
                            <div class="hero-badge">
                                <q-icon name="analytics" size="1.2rem" />
                                <span>Dataset Analysis</span>
                            </div>
                            <h1 class="hero-title">
                                Your Football Manager
                                <span class="gradient-text">Data Hub</span>
                            </h1>
                            <p class="hero-subtitle">
                                Explore comprehensive player analysis, team formations, and strategic insights 
                                from your Football Manager dataset.
                            </p>
                        </div>
                        <div class="hero-stats">
                            <div class="stat-card">
                                <div class="stat-number">{{ formatNumber(allPlayersData.length) }}</div>
                                <div class="stat-label">Players</div>
                            </div>
                            <div class="stat-card">
                                <div class="stat-number">{{ formatNumber(uniqueClubs.length) }}</div>
                                <div class="stat-label">Clubs</div>
                            </div>
                            <div class="stat-card">
                                <div class="stat-number">{{ formatNumber(uniqueNationalities.length) }}</div>
                                <div class="stat-label">Nations</div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Share Button positioned after hero -->
                <div v-if="currentDatasetId" class="share-button-container">
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

                <!-- Enhanced Quick Actions -->
                <div class="quick-actions-section">
                    <div class="actions-header">
                        <h2 class="actions-title">Quick Actions</h2>
                        <p class="actions-subtitle">Explore your dataset with these powerful analysis tools</p>
                    </div>
                    
                    <div class="actions-grid">
                        <div class="action-feature-card" @click="showUpgradeFinder = true" :class="{ 'disabled': allPlayersData.length === 0 }">
                            <div class="feature-icon-container accent-gradient">
                                <q-icon name="find_replace" size="2.5rem" />
                            </div>
                            <div class="feature-content">
                                <h3 class="feature-title">Upgrade Finder</h3>
                                <p class="feature-description">
                                    Find potential player upgrades and compare similar profiles. Build your dream squad strategically.
                                </p>
                            </div>
                        </div>

                        <div class="action-feature-card" @click="showWonderkids = true" :class="{ 'disabled': allPlayersData.length === 0 }">
                            <div class="feature-icon-container wonderkids-gradient">
                                <q-icon name="stars" size="2.5rem" />
                            </div>
                            <div class="feature-content">
                                <h3 class="feature-title">Find Wonderkids</h3>
                                <p class="feature-description">
                                    Discover the best young talents aged 15-21. Filter by transfer value and salary to find your next star.
                                </p>
                            </div>
                        </div>

                        <div class="action-feature-card" @click="showBargainHunter = true" :class="{ 'disabled': allPlayersData.length === 0 }">
                            <div class="feature-icon-container bargain-gradient">
                                <q-icon name="local_offer" size="2.5rem" />
                            </div>
                            <div class="feature-content">
                                <h3 class="feature-title">Bargain Hunter</h3>
                                <p class="feature-description">
                                    Find the best value players within your budget. Ranked by value score: overall rating divided by transfer value.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                <q-card
                    v-if="allPlayersData.length > 0"
                    class="q-mb-md"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <PlayerFilters
                            @filter-changed="handleFiltersChanged"
                            :all-available-roles="allAvailableRoles"
                            :unique-clubs="uniqueClubs"
                            :unique-nationalities="uniqueNationalities"
                            :unique-media-handlings="uniqueMediaHandlings"
                            :unique-personalities="uniquePersonalities"
                            :transfer-value-range="transferValueRangeForFilters"
                            :initial-dataset-range="
                                initialDatasetTransferValueRangeForFilters
                            "
                            :salary-range="salaryRangeForFilters"
                            :currency-symbol="detectedCurrencySymbol"
                            :age-slider-min-default="AGE_SLIDER_MIN_DEFAULT"
                            :age-slider-max-default="AGE_SLIDER_MAX_DEFAULT"
                            :is-loading="loading"
                        />
                    </q-card-section>
                </q-card>

                <q-card
                    v-if="allPlayersData.length > 0"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-sm">
                            Players ({{ filteredPlayers.length }})
                        </div>
                        <PlayerDataTable
                            :players="filteredPlayers"
                            :loading="loading"
                            @player-selected="handlePlayerSelected"
                            @team-selected="handleTeamSelected"
                            :is-goalkeeper-view="isGoalkeeperView"
                            :currency-symbol="detectedCurrencySymbol"
                            :dataset-id="currentDatasetId"
                        />
                    </q-card-section>
                </q-card>

                <q-banner
                    v-else-if="!pageLoading && !pageLoadingError"
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
                    No player data found for this dataset.
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
            :dataset-id="currentDatasetId"
        />
        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayersData"
            @close="showUpgradeFinder = false"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
        />
        <WonderkidsDialog
            :show="showWonderkids"
            :players="allPlayersData"
            @close="showWonderkids = false"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
        />
        <BargainHunterDialog
            :show="showBargainHunter"
            :players="allPlayersData"
            @close="showBargainHunter = false"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
        />
    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BargainHunterDialog from '../components/BargainHunterDialog.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import UpgradeFinderDialog from '../components/UpgradeFinderDialog.vue'
import WonderkidsDialog from '../components/WonderkidsDialog.vue'
import PlayerFilters from '../components/filters/PlayerFilters.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'

// Define FM attribute keys for filtering (raw keys as used in player.attributes)
const rawTechnicalAttributeKeysConst = [
  'Cor',
  'Cro',
  'Dri',
  'Fin',
  'Fir',
  'Fre',
  'Hea',
  'Lon',
  'L Th',
  'Mar',
  'Pas',
  'Pen',
  'Tck',
  'Tec'
]
const rawMentalAttributeKeysConst = [
  'Agg',
  'Ant',
  'Bra',
  'Cmp',
  'Cnt',
  'Dec',
  'Det',
  'Fla',
  'Ldr',
  'OtB',
  'Pos',
  'Tea',
  'Vis',
  'Wor'
]
const rawPhysicalAttributeKeysConst = ['Acc', 'Agi', 'Bal', 'Jum', 'Nat', 'Pac', 'Sta', 'Str']
const rawGoalkeeperAttributeKeysConst = [
  'Aer',
  'Cmd',
  'Com',
  'Ecc',
  'Han',
  'Kic',
  '1v1',
  'Pun',
  'Ref',
  'TRO',
  'Thr'
]

const allRawFmAttributeKeys = [
  ...rawTechnicalAttributeKeysConst,
  ...rawMentalAttributeKeysConst,
  ...rawPhysicalAttributeKeysConst,
  ...rawGoalkeeperAttributeKeysConst
]

// Helper to create filter keys like 'minCor', 'minLTh' (matching PlayerFilters.vue's formatAttrKey for consistency)
const formatFilterKeyPrefix = attrKey => {
  return attrKey.replace(/\s+/g, '').replace(/\(|\)/g, '')
}

export default {
  name: 'DatasetPage',
  components: {
    PlayerDataTable,
    PlayerDetailDialog,
    PlayerFilters,
    UpgradeFinderDialog,
    WonderkidsDialog,
    BargainHunterDialog
  },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()

    const pageLoading = ref(true)
    const pageLoadingError = ref('')
    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)
    const showUpgradeFinder = ref(false)
    const showWonderkids = ref(false)
    const showBargainHunter = ref(false)

    // Centralized filter state for this page
    const currentFilters = ref({
      name: '',
      club: null,
      position: null,
      role: null,
      nationality: null,
      mediaHandling: [],
      personality: [],
      ageRange: {
        min: playerStore.AGE_SLIDER_MIN_DEFAULT,
        max: playerStore.AGE_SLIDER_MAX_DEFAULT
      },
      transferValueRangeLocal: {
        min: 0,
        max: 100000000,
        userSet: false
      },
      maxSalary: 1000000,
      minOverall: 0,
      minPAC: 0,
      minSHO: 0,
      minPAS: 0,
      minDRI: 0,
      minDEF: 0,
      minPHY: 0,
      minGK: 0,
      minDIV: 0,
      minHAN: 0,
      minREF: 0,
      minKIC: 0,
      minSPD: 0,
      minPOS: 0
    })

    // Initialize FM attribute filters in currentFilters
    for (const attrKey of allRawFmAttributeKeys) {
      currentFilters.value[`min${formatFilterKeyPrefix(attrKey)}`] = 0
    }

    // Computed properties from store
    const allPlayersData = computed(() => playerStore.allPlayers)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const loading = computed(() => playerStore.loading)
    const uniqueClubs = computed(() => playerStore.uniqueClubs)
    const uniqueNationalities = computed(() => playerStore.uniqueNationalities)
    const uniqueMediaHandlings = computed(() => playerStore.uniqueMediaHandlings)
    const uniquePersonalities = computed(() => playerStore.uniquePersonalities)

    // For PlayerFilters component props - ensure safe access with fallbacks
    const transferValueRangeForFilters = computed(() => {
      const range = playerStore.currentDatasetTransferValueRange || {
        min: 0,
        max: 100000000
      }
      return range
    })
    const initialDatasetTransferValueRangeForFilters = computed(() => {
      const range = playerStore.initialDatasetTransferValueRange || {
        min: 0,
        max: 100000000
      }
      return range
    })
    const salaryRangeForFilters = computed(() => {
      const range = playerStore.salaryRange || { min: 0, max: 1000000 }
      return range
    })

    const allAvailableRoles = computed(() => playerStore.allAvailableRoles)
    const AGE_SLIDER_MIN_DEFAULT = computed(() => playerStore.AGE_SLIDER_MIN_DEFAULT)
    const AGE_SLIDER_MAX_DEFAULT = computed(() => playerStore.AGE_SLIDER_MAX_DEFAULT)

    const isGoalkeeperView = computed(() => {
      return (
        currentFilters.value.position === 'GK' || currentFilters.value.role?.includes('Goalkeeper')
      )
    })

    const filteredPlayers = computed(() => {
      if (!Array.isArray(allPlayersData.value)) return []

      return allPlayersData.value
        .filter(player => {
          // Name filter
          if (
            currentFilters.value.name &&
            !player.name.toLowerCase().includes(currentFilters.value.name.toLowerCase())
          ) {
            return false
          }

          // Club filter
          if (currentFilters.value.club && player.club !== currentFilters.value.club) {
            return false
          }

          // Position filter
          if (currentFilters.value.position) {
            const hasPosition = player.shortPositions?.includes(currentFilters.value.position)
            if (!hasPosition) return false
          }

          // Role filter
          if (currentFilters.value.role) {
            const hasRole = player.roleSpecificOveralls?.some(
              role => role.roleName === currentFilters.value.role
            )
            if (!hasRole) return false
          }

          // Nationality filter
          if (
            currentFilters.value.nationality &&
            player.nationality !== currentFilters.value.nationality
          ) {
            return false
          }

          // Media handling filter
          if (currentFilters.value.mediaHandling && currentFilters.value.mediaHandling.length > 0) {
            if (!player.media_handling) return false
            const playerMediaHandlings = player.media_handling.split(',').map(s => s.trim())
            const hasMediaHandling = currentFilters.value.mediaHandling.some(filter =>
              playerMediaHandlings.includes(filter)
            )
            if (!hasMediaHandling) return false
          }

          // Personality filter
          if (currentFilters.value.personality && currentFilters.value.personality.length > 0) {
            if (!player.personality) return false
            const hasPersonality = currentFilters.value.personality.includes(player.personality)
            if (!hasPersonality) return false
          }

          // Age range filter
          const playerAge = Number.parseInt(player.age, 10) || 0
          if (
            playerAge < currentFilters.value.ageRange.min ||
            playerAge > currentFilters.value.ageRange.max
          ) {
            return false
          }

          // Transfer value range filter
          if (
            player.transferValueAmount < currentFilters.value.transferValueRangeLocal.min ||
            player.transferValueAmount > currentFilters.value.transferValueRangeLocal.max
          ) {
            return false
          }

          // Max salary filter
          if (
            currentFilters.value.maxSalary !== null &&
            player.wageAmount > currentFilters.value.maxSalary
          ) {
            return false
          }

          // FIFA-style stat minimum filters
          if (
            currentFilters.value.minOverall > 0 &&
            (player.Overall || 0) < currentFilters.value.minOverall
          )
            return false
          if (currentFilters.value.minPAC > 0 && (player.PAC || 0) < currentFilters.value.minPAC)
            return false
          if (currentFilters.value.minSHO > 0 && (player.SHO || 0) < currentFilters.value.minSHO)
            return false
          if (currentFilters.value.minPAS > 0 && (player.PAS || 0) < currentFilters.value.minPAS)
            return false
          if (currentFilters.value.minDRI > 0 && (player.DRI || 0) < currentFilters.value.minDRI)
            return false
          if (currentFilters.value.minDEF > 0 && (player.DEF || 0) < currentFilters.value.minDEF)
            return false
          if (currentFilters.value.minPHY > 0 && (player.PHY || 0) < currentFilters.value.minPHY)
            return false
          if (currentFilters.value.minGK > 0 && (player.GK || 0) < currentFilters.value.minGK)
            return false
          if (currentFilters.value.minDIV > 0 && (player.DIV || 0) < currentFilters.value.minDIV)
            return false
          if (currentFilters.value.minHAN > 0 && (player.HAN || 0) < currentFilters.value.minHAN)
            return false
          if (currentFilters.value.minREF > 0 && (player.REF || 0) < currentFilters.value.minREF)
            return false
          if (currentFilters.value.minKIC > 0 && (player.KIC || 0) < currentFilters.value.minKIC)
            return false
          if (currentFilters.value.minSPD > 0 && (player.SPD || 0) < currentFilters.value.minSPD)
            return false
          if (currentFilters.value.minPOS > 0 && (player.POS || 0) < currentFilters.value.minPOS)
            return false

          // FM Attribute minimum filters
          for (const rawAttrKey of allRawFmAttributeKeys) {
            const filterKeyForVal = `min${formatFilterKeyPrefix(rawAttrKey)}`
            const minVal = currentFilters.value[filterKeyForVal]

            if (minVal > 0) {
              const playerAttrStr = player.attributes[rawAttrKey] // FM attributes are in player.attributes as strings
              const playerAttrVal = Number.parseInt(playerAttrStr, 10) || 0
              if (playerAttrVal < minVal) {
                return false
              }
            }
          }

          return true
        })
        .map(player => {
          if (currentFilters.value.role && player.roleSpecificOveralls) {
            let roleSpecificOverall = null
            if (Array.isArray(player.roleSpecificOveralls)) {
              const roleMatch = player.roleSpecificOveralls.find(
                rso => rso.roleName === currentFilters.value.role
              )
              if (roleMatch) roleSpecificOverall = roleMatch.score
            } else if (typeof player.roleSpecificOveralls === 'object') {
              roleSpecificOverall = player.roleSpecificOveralls[currentFilters.value.role]
            }

            if (roleSpecificOverall !== null && roleSpecificOverall !== undefined) {
              return { ...player, Overall: roleSpecificOverall }
            }
          }
          return player
        })
    })

    const fetchDataset = async datasetId => {
      pageLoading.value = true
      pageLoadingError.value = ''
      try {
        await playerStore.fetchPlayersByDatasetId(datasetId)
        await playerStore.fetchAllAvailableRoles()

        // Initialize wishlist for this dataset
        await wishlistStore.initializeWishlistForDataset(datasetId)

        // Safely access store values and provide defaults
        const initTvRange = playerStore.initialDatasetTransferValueRange
        const initSalaryRange = playerStore.salaryRange

        currentFilters.value.transferValueRangeLocal = {
          min: initTvRange && initTvRange.min !== undefined ? initTvRange.min : 0,
          max: initTvRange && initTvRange.max !== undefined ? initTvRange.max : 100000000,
          userSet: false
        }
        currentFilters.value.maxSalary =
          initSalaryRange && initSalaryRange.max !== undefined ? initSalaryRange.max : 1000000
      } catch (err) {
        pageLoadingError.value = `Failed to load dataset: ${err.message || 'Unknown server error'}.`
        playerStore.resetState()
      } finally {
        pageLoading.value = false
      }
    }

    onMounted(async () => {
      const datasetIdFromRoute = route.params.datasetId
      if (datasetIdFromRoute) {
        await fetchDataset(datasetIdFromRoute)
      } else {
        pageLoadingError.value = 'No dataset ID provided in URL.'
        pageLoading.value = false
      }
    })

    const shareDataset = async () => {
      if (!currentDatasetId.value) return
      const shareUrl = `${window.location.origin}/dataset/${currentDatasetId.value}`
      try {
        await navigator.clipboard.writeText(shareUrl)
        quasarInstance.notify({
          message: 'Dataset link copied to clipboard!',
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
          message: 'Dataset link copied to clipboard!',
          color: 'positive',
          icon: 'check_circle',
          position: 'top',
          timeout: 2000
        })
      }
      try {
        if (window.gtag) {
          window.gtag('event', 'share_dataset', {
            dataset_id: currentDatasetId.value
          })
        }
      } catch (e) {
        console.error('GA Event failed', e)
      }
    }

    const handlePlayerSelected = player => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const handleTeamSelected = teamName => {
      if (currentDatasetId.value) {
        const url = router.resolve({
          path: '/team-view',
          query: {
            datasetId: currentDatasetId.value,
            team: teamName
          }
        }).href
        const newWindow = window.open(url, '_blank')
        if (!newWindow) {
          console.error('Failed to open new window - likely blocked by popup blocker')
        } else {
        }
      } else {
      }
    }

    const handleFiltersChanged = filtersFromChild => {
      const newTransferRange = filtersFromChild.transferValueRangeLocal
      const oldTransferRange = currentFilters.value.transferValueRangeLocal

      currentFilters.value = {
        ...currentFilters.value,
        ...filtersFromChild
      }

      if (
        newTransferRange &&
        oldTransferRange &&
        (newTransferRange.min !== oldTransferRange.min ||
          newTransferRange.max !== oldTransferRange.max)
      ) {
        currentFilters.value.transferValueRangeLocal.userSet = true
      }
    }

    watch(
      () => route.params.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          await fetchDataset(newId)
        }
      }
    )

    watch(
      () => playerStore.initialDatasetTransferValueRange,
      newRange => {
        if (newRange && !currentFilters.value.transferValueRangeLocal.userSet) {
          // Check userSet flag
          currentFilters.value.transferValueRangeLocal = {
            min: newRange.min !== undefined ? newRange.min : 0,
            max: newRange.max !== undefined ? newRange.max : 100000000,
            userSet: false // Keep userSet as false when programmatically updating
          }
        }
      },
      { immediate: true, deep: true }
    )

    watch(
      () => playerStore.salaryRange,
      newRange => {
        const currentMaxSalary = currentFilters.value.maxSalary
        const storeMaxSalary = playerStore.salaryRange?.max

        // Only update if maxSalary hasn't been manually set by user OR is still at its initial large default OR matches the current store max
        if (
          newRange &&
          (currentMaxSalary === 1000000 ||
            currentMaxSalary === null ||
            currentMaxSalary === storeMaxSalary)
        ) {
          if (newRange.max !== undefined) {
            currentFilters.value.maxSalary = newRange.max
          } else {
            currentFilters.value.maxSalary = 1000000
          }
        }
      },
      { immediate: true, deep: true }
    )

    // Utility function to format large numbers
    const formatNumber = num => {
      if (num >= 1000000) {
        return `${(num / 1000000).toFixed(1).replace(/\.0$/, '')}M`
      }
      if (num >= 1000) {
        return `${(num / 1000).toFixed(1).replace(/\.0$/, '')}K`
      }
      return num?.toString() || '0'
    }

    return {
      pageLoading,
      pageLoadingError,
      allPlayersData,
      detectedCurrencySymbol,
      currentDatasetId,
      loading,
      uniqueClubs,
      uniqueNationalities,
      uniqueMediaHandlings,
      uniquePersonalities,
      transferValueRangeForFilters,
      initialDatasetTransferValueRangeForFilters,
      salaryRangeForFilters,
      allAvailableRoles,
      AGE_SLIDER_MIN_DEFAULT,
      AGE_SLIDER_MAX_DEFAULT,
      isGoalkeeperView,
      filteredPlayers,
      playerForDetailView,
      showPlayerDetailDialog,
      showUpgradeFinder,
      showWonderkids,
      showBargainHunter,
      shareDataset,
      handlePlayerSelected,
      handleTeamSelected,
      handleFiltersChanged,
      quasarInstance,
      router,
      currentFilters,
      formatNumber
    }
  }
}
</script>

<style lang="scss" scoped>
.dataset-page {
    max-width: 1600px;
    margin: 0 auto;
}

.page-title {
    margin: 0;
}

.share-btn-enhanced {
    font-weight: 600;
    border-radius: 6px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    min-width: 140px;

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

.stats-card {
    height: 100%;
    transition: transform 0.2s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);

        &:hover {
            box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
        }
    }
}

.q-card {
    border-radius: $generic-border-radius;
}

.quick-actions-card {
    .action-card {
        height: 100%;
        transition: all 0.3s ease;
        border-radius: 12px;

        &:hover {
            transform: translateY(-4px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.12);
        }

        .body--dark & {
            background-color: rgba(255, 255, 255, 0.03);

            &:hover {
                background-color: rgba(255, 255, 255, 0.06);
                box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
            }
        }

        .body--light & {
            background-color: rgba(0, 0, 0, 0.02);

            &:hover {
                background-color: rgba(0, 0, 0, 0.04);
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
        display: grid;
        grid-template-columns: 1fr auto;
        gap: 4rem;
        align-items: center;
        position: relative;
        z-index: 1;
        
        @media (max-width: 768px) {
            grid-template-columns: 1fr;
            gap: 2rem;
            text-align: center;
        }
    }
    
    .hero-content {
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
            max-width: 500px;
            
            @media (max-width: 768px) {
                font-size: 1.1rem;
                max-width: none;
            }
        }
    }
    
    .hero-stats {
        display: flex;
        gap: 1.5rem;
        
        @media (max-width: 768px) {
            justify-content: center;
        }
        
        @media (max-width: 480px) {
            flex-direction: column;
            gap: 1rem;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 1.5rem;
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            min-width: 100px;
            
            .stat-number {
                font-size: 2.2rem;
                font-weight: 700;
                color: #64b5f6;
                margin-bottom: 0.5rem;
                line-height: 1;
            }
            
            .stat-label {
                font-size: 0.9rem;
                opacity: 0.8;
                font-weight: 500;
            }
        }
    }
}

// Enhanced Quick Actions Section
.quick-actions-section {
    margin: 3rem 0;
    padding: 0 1rem;
    
    .actions-header {
        text-align: center;
        margin-bottom: 3rem;
        
        .actions-title {
            font-size: 2.5rem;
            font-weight: 700;
            margin: 0 0 1rem 0;
            color: #1a237e;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.9);
            }
        }
        
        .actions-subtitle {
            font-size: 1.1rem;
            color: #666;
            margin: 0;
            max-width: 600px;
            margin: 0 auto;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
    
    .actions-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 2rem;
        max-width: 1200px;
        margin: 0 auto;
        
        @media (max-width: 1200px) {
            grid-template-columns: repeat(2, 1fr);
        }
        
        @media (max-width: 768px) {
            grid-template-columns: 1fr;
            gap: 1.5rem;
        }
    }
    
    .action-feature-card {
        background: white;
        border-radius: 20px;
        padding: 2rem;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
        border: 1px solid rgba(0, 0, 0, 0.05);
        cursor: pointer;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        gap: 1.5rem;
        position: relative;
        overflow: hidden;
        
        &::before {
            content: "";
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(135deg, rgba(26, 35, 126, 0.02) 0%, transparent 100%);
            opacity: 0;
            transition: opacity 0.3s ease;
        }
        
        &:hover {
            transform: translateY(-8px);
            box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
            
            &::before {
                opacity: 1;
            }
            
            .feature-icon-container {
                transform: scale(1.1) rotate(5deg);
            }
        }
        
        &.disabled {
            opacity: 0.5;
            cursor: not-allowed;
            
            &:hover {
                transform: none;
                box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
                
                .feature-icon-container {
                    transform: none;
                }
            }
        }
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            
            &:hover {
                background: rgba(255, 255, 255, 0.08);
                box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
            }
        }
        
        .feature-icon-container {
            width: 80px;
            height: 80px;
            border-radius: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            flex-shrink: 0;
            transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            
            &.primary-gradient {
                background: linear-gradient(135deg, #1976d2 0%, #42a5f5 100%);
            }
            
            &.secondary-gradient {
                background: linear-gradient(135deg, #7b1fa2 0%, #ba68c8 100%);
            }
            
            &.info-gradient {
                background: linear-gradient(135deg, #0288d1 0%, #4fc3f7 100%);
            }
            
            &.accent-gradient {
                background: linear-gradient(135deg, #c2185b 0%, #f06292 100%);
            }
            
            &.wonderkids-gradient {
                background: linear-gradient(135deg, #ffd700 0%, #ffa500 100%);
            }
            
            &.bargain-gradient {
                background: linear-gradient(135deg, #4caf50 0%, #8bc34a 100%);
            }
        }
        
        .feature-content {
            .feature-title {
                font-size: 1.4rem;
                font-weight: 600;
                margin: 0 0 0.5rem 0;
                color: #1a237e;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .feature-description {
                font-size: 0.9rem;
                line-height: 1.5;
                color: #666;
                margin: 0;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }
        
        @media (max-width: 768px) {
            padding: 1.5rem;
            
            .feature-icon-container {
                width: 60px;
                height: 60px;
                
                .q-icon {
                    font-size: 2rem !important;
                }
            }
            
            .feature-content {
                .feature-title {
                    font-size: 1.2rem;
                }
                
                .feature-description {
                    font-size: 0.85rem;
                }
            }
        }
    }
}
</style>
