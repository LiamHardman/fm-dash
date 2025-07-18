<template>
    <q-page class="dataset-page full-height">
        <!-- Compact Top Bar with Filters and Actions -->
        <div class="top-bar" v-if="!pageLoading && !pageLoadingError && allPlayersData.length > 0">
            <div class="top-bar-content">
                <!-- Left section: Dataset info and stats -->
                <div class="dataset-info">
                    <div class="dataset-title">
                        <q-icon name="analytics" size="1.2rem" class="q-mr-xs" />
                        FM Data Hub
                    </div>
                    <div class="dataset-stats">
                        <span class="stat-item">{{ formatNumber(allPlayersData.length) }} Players</span>
                        <span class="stat-separator">•</span>
                        <span class="stat-item">{{ formatNumber(uniqueClubs.length) }} Clubs</span>
                        <span class="stat-separator">•</span>
                        <span class="stat-item">{{ formatNumber(uniqueNationalities.length) }} Nations</span>
                    </div>
                </div>

                <!-- Center section: Quick Actions -->
                <div class="quick-actions">
                    <q-btn
                        unelevated
                        dense
                        icon="find_replace"
                        label="Upgrade Finder"
                        color="primary"
                        @click="openUpgradeFinder"
                        :disable="allPlayersData.length === 0"
                        class="action-btn"
                        size="sm"
                    />
                    <q-btn
                        unelevated
                        dense
                        icon="stars"
                        label="Wonderkids"
                        color="secondary"
                        @click="openWonderkids"
                        :disable="allPlayersData.length === 0"
                        class="action-btn"
                        size="sm"
                    />
                    <q-btn
                        unelevated
                        dense
                        icon="local_offer"
                        label="Bargains"
                        color="positive"
                        @click="openBargainHunter"
                        :disable="allPlayersData.length === 0"
                        class="action-btn"
                        size="sm"
                    />
                    <q-btn
                        unelevated
                        dense
                        icon="person_off"
                        label="Free Agents"
                        color="deep-orange"
                        @click="openFreeAgents"
                        :disable="allPlayersData.length === 0"
                        class="action-btn"
                        size="sm"
                    />
                    <q-btn
                        unelevated
                        dense
                        icon="download"
                        label="Export"
                        color="accent"
                        @click="openExportOptions"
                        :disable="loading || !filteredPlayers || filteredPlayers.length === 0"
                        class="action-btn"
                        size="sm"
                    >
                        <q-tooltip v-if="filteredPlayers && filteredPlayers.length > 0">
                            Export {{ filteredPlayers.length }} filtered players
                        </q-tooltip>
                        <q-tooltip v-else>
                            No players to export
                        </q-tooltip>
                    </q-btn>
                </div>

                <!-- Right section: Share and filters toggle -->
                <div class="top-bar-controls">
                    <q-btn
                        v-if="currentDatasetId"
                        flat
                        dense
                        icon="share"
                        @click="shareDataset"
                        class="share-btn"
                        size="sm"
                    >
                        <q-tooltip>Share Dataset</q-tooltip>
                    </q-btn>
                    <q-btn
                        flat
                        dense
                        :icon="showFilters ? 'filter_list_off' : 'filter_list'"
                        @click="showFilters = !showFilters"
                        class="filter-toggle-btn"
                        size="sm"
                        :color="showFilters ? 'primary' : 'grey-6'"
                    >
                        <q-tooltip>{{ showFilters ? 'Hide' : 'Show' }} Filters</q-tooltip>
                    </q-btn>
                </div>
            </div>

            <!-- Collapsible Filters Section -->
            <q-slide-transition>
                <div v-show="showFilters" class="filters-section">
                    <PlayerFilters
                        @filter-changed="handleFiltersChanged"
                        :all-available-roles="allAvailableRoles"
                        :unique-clubs="uniqueClubs"
                        :unique-nationalities="uniqueNationalities"
                        :unique-media-handlings="uniqueMediaHandlings"
                        :unique-personalities="uniquePersonalities"
                        :transfer-value-range="transferValueRangeForFilters"
                        :initial-dataset-range="initialDatasetTransferValueRangeForFilters"
                        :salary-range="salaryRangeForFilters"
                        :currency-symbol="detectedCurrencySymbol"
                        :age-slider-min-default="AGE_SLIDER_MIN_DEFAULT"
                        :age-slider-max-default="AGE_SLIDER_MAX_DEFAULT"
                        :is-loading="loading"
                    />
                </div>
            </q-slide-transition>
        </div>

        <!-- Loading State -->
        <div v-if="pageLoading" class="loading-container">
            <q-spinner-dots color="primary" size="3em" />
            <div class="loading-text">Loading dataset...</div>
        </div>

        <!-- Error State -->
        <div v-if="pageLoadingError" class="error-container">
            <q-banner class="error-banner" rounded>
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
        </div>

        <!-- No Data State -->
        <div v-if="!pageLoading && !pageLoadingError && allPlayersData.length === 0" class="no-data-container">
            <q-banner class="no-data-banner">
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

        <!-- Full Screen Player Data Table -->
        <div
            v-if="!pageLoading && !pageLoadingError && allPlayersData.length > 0"
            class="table-container"
        >
            <div class="table-header">
                <span class="table-title">Players ({{ filteredPlayers.length }})</span>
            </div>
            <div class="table-wrapper">
                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @player-selected="handlePlayerSelected"
                    @team-selected="handleTeamSelected"
                    :is-goalkeeper-view="isGoalkeeperView"
                    :currency-symbol="detectedCurrencySymbol"
                    :dataset-id="currentDatasetId"
                    class="full-screen-table"
                />
            </div>
        </div>

        <!-- Dialogs -->
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
        <FreeAgentsDialog
            :show="showFreeAgents"
            :players="allPlayersData"
            @close="showFreeAgents = false"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
        />
        <ExportOptionsDialog
            :show="showExportOptions"
            :player-count="filteredPlayers ? filteredPlayers.length : 0"
            @close="showExportOptions = false"
            @export="handleExportWithOptions"
        />
    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BargainHunterDialog from '../components/BargainHunterDialog.vue'
import ExportOptionsDialog from '../components/ExportOptionsDialog.vue'
import FreeAgentsDialog from '../components/FreeAgentsDialog.vue'
import PlayerFilters from '../components/filters/PlayerFilters.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import UpgradeFinderDialog from '../components/UpgradeFinderDialog.vue'
import WonderkidsDialog from '../components/WonderkidsDialog.vue'
import { useAnalytics } from '../composables/useAnalytics'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'
import { exportPlayersWithOptions, validateExportData } from '../utils/csvExport'

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
    BargainHunterDialog,
    FreeAgentsDialog,
    ExportOptionsDialog
  },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()
    const analytics = useAnalytics()

    const pageLoading = ref(true)
    const pageLoadingError = ref('')
    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)
    const showUpgradeFinder = ref(false)
    const showWonderkids = ref(false)
    const showBargainHunter = ref(false)
    const showFreeAgents = ref(false)
    const showExportOptions = ref(false)

    const currentFilters = ref({
      name: '',
      position: '',
      role: '',
      club: '',
      nationality: '',
      mediaHandling: [],
      personality: [],
      ageRange: {
        min: computed(() => AGE_SLIDER_MIN_DEFAULT.value),
        max: computed(() => AGE_SLIDER_MAX_DEFAULT.value)
      },
      transferValueRangeLocal: {
        min: 0,
        max: 100000000
      },
      maxSalary: null,
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
      minPOS: 0,
      continentNationalities: []
    })

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

    const filteredPlayers = computed(() => {
      if (!Array.isArray(allPlayersData.value)) return []

      return allPlayersData.value
        .filter(player => {
          if (
            currentFilters.value.name &&
            !player.name.toLowerCase().includes(currentFilters.value.name.toLowerCase())
          ) {
            return false
          }

          if (currentFilters.value.club && player.club !== currentFilters.value.club) {
            return false
          }

          if (currentFilters.value.position) {
            const hasPosition = player.short_positions?.includes(currentFilters.value.position)
            if (!hasPosition) return false
          }

          if (currentFilters.value.role) {
            const hasRole = player.roleSpecificOveralls?.some(
              role => role.roleName === currentFilters.value.role
            )
            if (!hasRole) return false
          }

          if (
            currentFilters.value.nationality &&
            player.nationality !== currentFilters.value.nationality
          ) {
            return false
          }

          // Continent-based nationality filter (for preset filters like EU, Europe, etc.)
          if (
            currentFilters.value.continentNationalities &&
            currentFilters.value.continentNationalities.length > 0 &&
            !currentFilters.value.continentNationalities.includes(player.nationality)
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

    const isGoalkeeperView = computed(() => {
      // First check explicit position or role filters
      if (
        currentFilters.value.position === 'GK' ||
        currentFilters.value.role?.includes('Goalkeeper')
      ) {
        return true
      }

      // If we have filtered players, check if majority are goalkeepers
      if (filteredPlayers.value && filteredPlayers.value.length > 0) {
        const goalkeeperCount = filteredPlayers.value.filter(player => {
          const isGK =
            player.position?.includes('GK') ||
                          player.short_positions?.includes('GK') ||
            player.position_groups?.includes('Goalkeepers')
          return isGK
        }).length

        // Show goalkeeper view if more than 50% of filtered players are goalkeepers
        // and we have at least one goalkeeper
        return goalkeeperCount > 0 && goalkeeperCount > filteredPlayers.value.length / 2
      }

      return false
    })

    const fetchDataset = async datasetId => {
      pageLoading.value = true
      pageLoadingError.value = ''
      try {
        await playerStore.fetchPlayersByDatasetId(datasetId)
        await playerStore.fetchAllAvailableRoles()

        await wishlistStore.initializeWishlistForDataset(datasetId)

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

      // Track the share event using our analytics service
      analytics.shareDataset(currentDatasetId.value)
    }

    const handlePlayerSelected = player => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true

      // Track player detail view
      analytics.viewPlayerDetails(player.id || player.name, player.name)
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
        } else {
          // Track team navigation
          analytics.navigateToTeamView(teamName, currentDatasetId.value)
        }
      } else {
      }
    }

    const handleFiltersChanged = filtersFromChild => {
      const newTransferRange = filtersFromChild.transferValueRangeLocal
      const oldTransferRange = currentFilters.value.transferValueRangeLocal

      // Track filter usage for analytics
      Object.keys(filtersFromChild).forEach(filterKey => {
        const newValue = filtersFromChild[filterKey]
        const oldValue = currentFilters.value[filterKey]

        // Only track if the value actually changed and is meaningful
        if (
          newValue !== oldValue &&
          newValue !== '' &&
          newValue !== null &&
          newValue !== undefined
        ) {
          analytics.useFilter(
            filterKey,
            typeof newValue === 'object' ? JSON.stringify(newValue) : newValue
          )
        }
      })

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

    const openUpgradeFinder = () => {
      showUpgradeFinder.value = true
      analytics.trackButtonClick('Upgrade Finder', { feature_type: 'quick_action' })
    }

    const openWonderkids = () => {
      showWonderkids.value = true
      analytics.trackButtonClick('Find Wonderkids', { feature_type: 'quick_action' })
    }

    const openBargainHunter = () => {
      showBargainHunter.value = true
      analytics.trackButtonClick('Bargain Hunter', { feature_type: 'quick_action' })
    }

    const openFreeAgents = () => {
      showFreeAgents.value = true
      analytics.trackButtonClick('Free Agents', { feature_type: 'quick_action' })
    }

    const openExportOptions = () => {
      showExportOptions.value = true
      analytics.trackButtonClick('Export Options', { feature_type: 'export' })
    }

    const handleExportWithOptions = async exportOptions => {
      try {
        // Validate the export data
        const validation = validateExportData(filteredPlayers.value)

        if (!validation.valid) {
          quasarInstance.notify({
            type: 'negative',
            message: `Export failed: ${validation.errors.join(', ')}`,
            position: 'top'
          })
          return
        }

        // Show warnings if any
        if (validation.warnings.length > 0) {
          validation.warnings.forEach(warning => {
            quasarInstance.notify({
              type: 'warning',
              message: warning,
              position: 'top'
            })
          })
        }

        // Export with options
        await exportPlayersWithOptions(filteredPlayers.value, exportOptions)

        // Show success message
        quasarInstance.notify({
          type: 'positive',
          message: `Successfully exported ${filteredPlayers.value.length} players as ${exportOptions.format.toUpperCase()}`,
          position: 'top',
          actions: [
            {
              label: 'Dismiss',
              color: 'white'
            }
          ]
        })

        // Track export event
        analytics.downloadData('players', exportOptions.format)
        analytics.trackButtonClick(`Export ${exportOptions.format.toUpperCase()}`, {
          feature_type: 'export',
          player_count: filteredPlayers.value.length,
          preset: exportOptions.preset
        })

        // Close the dialog
        showExportOptions.value = false
      } catch (error) {
        quasarInstance.notify({
          type: 'negative',
          message: `Export failed: ${error.message}`,
          position: 'top'
        })
      }
    }

    const showFilters = ref(false)

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
      showFreeAgents,
      showExportOptions,
      shareDataset,
      handlePlayerSelected,
      handleTeamSelected,
      handleFiltersChanged,
      quasarInstance,
      router,
      currentFilters,
      formatNumber,
      openUpgradeFinder,
      openWonderkids,
      openBargainHunter,
      openFreeAgents,
      openExportOptions,
      handleExportWithOptions,
      showFilters
    }
  }
}
</script>

<style lang="scss" scoped>
.dataset-page {
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

// Top Bar Styles
.top-bar {
    background: white;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
    z-index: 10;
    flex-shrink: 0;
    
    .body--dark & {
        background: #1d1d1d;
        border-bottom-color: rgba(255, 255, 255, 0.1);
    }
}

.top-bar-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.5rem;
    gap: 1rem;
    min-height: 60px;
}

.dataset-info {
    flex: 1;
    min-width: 0;
    
    .dataset-title {
        font-size: 1.1rem;
        font-weight: 600;
        color: #1976d2;
        display: flex;
        align-items: center;
        margin-bottom: 0.25rem;
        
        .body--dark & {
            color: #64b5f6;
        }
    }
    
    .dataset-stats {
        font-size: 0.8rem;
        color: #666;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
        
        .stat-item {
            font-weight: 500;
        }
        
        .stat-separator {
            opacity: 0.5;
        }
    }
}

.quick-actions {
    display: flex;
    gap: 0.5rem;
    
    .action-btn {
        border-radius: 8px;
        padding: 0 12px;
        height: 32px;
        font-size: 0.8rem;
        font-weight: 500;
        
        @media (max-width: 768px) {
            padding: 0 8px;
            font-size: 0.75rem;
            
            .q-icon {
                font-size: 1rem;
            }
        }
    }
    
    @media (max-width: 1024px) {
        .action-btn .q-btn__content span:not(.q-icon) {
            display: none;
        }
    }
}

.top-bar-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    
    .share-btn,
    .filter-toggle-btn {
        border-radius: 8px;
        width: 36px;
        height: 36px;
    }
}

.filters-section {
    border-top: 1px solid rgba(0, 0, 0, 0.1);
    padding: 1rem 1.5rem;
    background: rgba(0, 0, 0, 0.02);
    
    .body--dark & {
        border-top-color: rgba(255, 255, 255, 0.1);
        background: rgba(255, 255, 255, 0.02);
    }
}

// Loading, Error, and No Data States
.loading-container,
.error-container,
.no-data-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
}

.loading-text {
    margin-top: 1rem;
    color: #666;
    font-size: 0.9rem;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.error-banner,
.no-data-banner {
    max-width: 600px;
    width: 100%;
}

.error-banner {
    background: #f44336;
    color: white;
}

.no-data-banner {
    background: rgba(255, 152, 0, 0.1);
    color: #f57c00;
    
    .body--dark & {
        background: rgba(255, 152, 0, 0.2);
        color: #ffb74d;
    }
}

// Table Container Styles
.table-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    margin: 0 1rem 1rem 1rem;
}

.table-header {
    padding: 0.75rem 0 0.5rem 0;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
    
    .body--dark & {
        border-bottom-color: rgba(255, 255, 255, 0.1);
    }
    
    .table-title {
        font-size: 1.1rem;
        font-weight: 600;
        color: #1976d2;
        
        .body--dark & {
            color: #64b5f6;
        }
    }
}

.table-wrapper {
    flex: 1;
    overflow: hidden;
    background: white;
    border-radius: 8px;
    border: 1px solid rgba(0, 0, 0, 0.1);
    
    .body--dark & {
        background: #1d1d1d;
        border-color: rgba(255, 255, 255, 0.1);
    }
}

// Global utility classes
.full-height {
    height: 100vh !important;
}

// Media queries for responsiveness
@media (max-width: 768px) {
    .top-bar-content {
        padding: 0.5rem 1rem;
        flex-wrap: wrap;
        min-height: auto;
    }
    
    .dataset-info {
        order: 1;
        width: 100%;
        margin-bottom: 0.5rem;
    }
    
    .quick-actions {
        order: 2;
        flex: 1;
    }
    
    .top-bar-controls {
        order: 3;
        margin-left: auto;
    }
    
    .filters-section {
        padding: 0.75rem 1rem;
    }
    
    .table-container {
        margin: 0 0.5rem 0.5rem 0.5rem;
    }
}

@media (max-width: 480px) {
    .quick-actions {
        gap: 0.25rem;
        
        .action-btn {
            padding: 0 6px;
            height: 28px;
            font-size: 0.7rem;
        }
    }
}
</style> 