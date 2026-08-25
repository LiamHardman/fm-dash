<template>
    <q-page class="save-analysis-page">
        <div class="page-container">
            <!-- Loading State -->
            <div v-if="pageLoading" class="loading-state">
                <q-spinner-orbit color="primary" size="4em" />
                <div class="loading-text">Loading player database...</div>
            </div>

            <!-- Error State -->
            <EmptyState
                v-else-if="pageLoadingError"
                icon="error"
                title="Couldn't load save analysis"
                :description="pageLoadingError"
            >
                <template #actions>
                    <q-btn unelevated color="primary" label="Go to Upload Page" @click="router.push('/')" />
                </template>
            </EmptyState>

            <!-- Main Content -->
            <div v-else>
                <PageHeader
                    title="Save Analysis"
                    :subtitle="`CA vs. Overall across ${formatNumber(filteredPlayers.length)} of ${formatNumber(allPlayersData.length)} players`"
                    icon="insights"
                >
                    <template #actions>
                        <q-select
                            v-model="selectedPosition"
                            :options="positionOptions"
                            label="Position"
                            dense
                            outlined
                            style="min-width: 180px"
                        />
                        <q-input
                            v-model.number="minOverall"
                            type="number"
                            label="Min Overall"
                            dense
                            outlined
                            clearable
                            style="width: 130px"
                        />
                        <q-input
                            v-model.number="maxOverall"
                            type="number"
                            label="Max Overall"
                            dense
                            outlined
                            clearable
                            style="width: 130px"
                        />
                    </template>
                </PageHeader>

                <SectionCard class="q-mb-md">
                    <SaveAnalysisScatter
                        :players="filteredPlayers"
                        :is-dark-mode="isDarkMode"
                        @player-click="openPlayerDetail"
                    />
                </SectionCard>

                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-6">
                        <SectionCard title="Rating Outliers" icon="swap_vert" class="full-height">
                            <div class="text-caption text-grey q-mb-sm">
                                CA rescaled to Overall's 0-99 scale, then compared. Click a tab to see which direction.
                            </div>
                            <q-tabs v-model="outlierTab" dense align="justify" class="text-grey" active-color="primary" indicator-color="primary">
                                <q-tab name="overrated" label="Overall > CA" />
                                <q-tab name="underrated" label="CA > Overall" />
                            </q-tabs>
                            <q-separator />
                            <q-table
                                :rows="outlierTab === 'overrated' ? overratedOutliers : underratedOutliers"
                                :columns="outlierColumns"
                                row-key="uid"
                                flat
                                dense
                                hide-pagination
                                :rows-per-page-options="[0]"
                                @row-click="(_evt, row) => openPlayerDetail(row)"
                                class="save-analysis-table"
                            />
                        </SectionCard>
                    </div>
                    <div class="col-12 col-md-6">
                        <SectionCard title="Best Role Distribution" icon="workspaces" class="full-height">
                            <div class="text-caption text-grey q-mb-sm">
                                How many filtered players have each role as their single best role.
                            </div>
                            <q-separator />
                            <q-table
                                :rows="roleDistribution"
                                :columns="roleDistributionColumns"
                                row-key="role"
                                flat
                                dense
                                :pagination="{ rowsPerPage: 10, sortBy: 'count', descending: true }"
                                class="save-analysis-table"
                            />
                        </SectionCard>
                    </div>
                </div>

                <SectionCard title="All filtered players" icon="table_rows">
                    <q-table
                        :rows="filteredPlayers"
                        :columns="tableColumns"
                        row-key="uid"
                        flat
                        dense
                        :pagination="{ rowsPerPage: 20, sortBy: 'overall', descending: true }"
                        @row-click="(_evt, row) => openPlayerDetail(row)"
                        class="save-analysis-table"
                    />
                </SectionCard>
            </div>
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
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
// biome-ignore lint/correctness/noUnusedImports: used in template
import EmptyState from '../components/layout/EmptyState.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PageHeader from '../components/layout/PageHeader.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import SectionCard from '../components/layout/SectionCard.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import SaveAnalysisScatter from '../components/SaveAnalysisScatter.vue'
import { useDynamicComponents } from '../composables/useDynamicComponents.js'
import { fetchPerformanceData } from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
// biome-ignore lint/correctness/noUnusedImports: used in template
import { formatNumber } from '../utils/currencyUtils.js'

// biome-ignore lint/correctness/noUnusedVariables: used in template
const router = useRouter()
const route = useRoute()
const playerStore = usePlayerStore()
const uiStore = useUiStore()

const {
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  DynamicPlayerDetailDialog,
} = useDynamicComponents()

// biome-ignore lint/correctness/noUnusedVariables: used in template
const isDarkMode = computed(() => uiStore.isDarkModeActive)

const pageLoading = ref(true)
const pageLoadingError = ref('')
const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)

const allPlayersData = computed(() => playerStore.allPlayers)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const currentDatasetId = computed(() => playerStore.currentDatasetId)

const POSITION_ORDER = [
  'GK',
  'DC',
  'DL',
  'DR',
  'WBL',
  'WBR',
  'DM',
  'MC',
  'ML',
  'MR',
  'AMC',
  'AML',
  'AMR',
  'ST',
]

const selectedPosition = ref('All')
const minOverall = ref(null)
const maxOverall = ref(null)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const outlierTab = ref('overrated')

// biome-ignore lint/correctness/noUnusedVariables: used in template
const positionOptions = computed(() => {
  const present = new Set()
  for (const player of allPlayersData.value) {
    const positions = player.shortPositions || player.short_positions || []
    for (const pos of positions) present.add(pos)
  }
  return ['All', ...POSITION_ORDER.filter((pos) => present.has(pos))]
})

const getOverall = (row) => row.overall ?? row.Overall
const getCA = (row) => row.ca ?? row.CA
// CA is 0-200, Overall is 0-99; rescale CA onto Overall's range so the two are comparable.
const getCAScaled = (row) => {
  const ca = getCA(row)
  return ca === null || ca === undefined ? null : Math.round((ca / 200) * 99)
}
const getRatingDiff = (row) => {
  const caScaled = getCAScaled(row)
  const overall = getOverall(row)
  return caScaled === null || overall === null || overall === undefined ? null : overall - caScaled
}

const isBoundSet = (v) => v !== null && v !== undefined && v !== ''

const filteredPlayers = computed(() => {
  return allPlayersData.value.filter((player) => {
    const positions = player.shortPositions || player.short_positions || []
    if (selectedPosition.value !== 'All' && !positions.includes(selectedPosition.value))
      return false

    const overall = getOverall(player)
    if (
      isBoundSet(minOverall.value) &&
      (overall === null || overall === undefined || overall < minOverall.value)
    )
      return false
    if (
      isBoundSet(maxOverall.value) &&
      (overall === null || overall === undefined || overall > maxOverall.value)
    )
      return false

    return true
  })
})

const playersWithRatingDiff = computed(() =>
  filteredPlayers.value.filter((player) => getRatingDiff(player) !== null)
)

// Overall higher than CA suggests — the "Murodov-shape" overrated direction.
// biome-ignore lint/correctness/noUnusedVariables: used in template
const overratedOutliers = computed(() =>
  [...playersWithRatingDiff.value].sort((a, b) => getRatingDiff(b) - getRatingDiff(a)).slice(0, 15)
)

// CA higher than Overall suggests — the "Yamal-shape" underrated direction.
// biome-ignore lint/correctness/noUnusedVariables: used in template
const underratedOutliers = computed(() =>
  [...playersWithRatingDiff.value].sort((a, b) => getRatingDiff(a) - getRatingDiff(b)).slice(0, 15)
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const roleDistribution = computed(() => {
  const counts = new Map()
  for (const player of filteredPlayers.value) {
    const role = player.bestRoleOverall || player.BestRoleOverall || 'Unknown'
    counts.set(role, (counts.get(role) || 0) + 1)
  }
  const total = filteredPlayers.value.length
  return [...counts.entries()]
    .map(([role, count]) => ({
      role,
      count,
      share: total ? Math.round((count / total) * 1000) / 10 : 0,
    }))
    .sort((a, b) => b.count - a.count)
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const tableColumns = [
  { name: 'name', label: 'Name', field: 'name', align: 'left', sortable: true },
  { name: 'club', label: 'Club', field: 'club', align: 'left', sortable: true },
  {
    name: 'positions',
    label: 'Positions',
    field: (row) => (row.shortPositions || row.short_positions || []).join(', '),
    align: 'left',
    sortable: true,
  },
  {
    name: 'bestRoleOverall',
    label: 'Best Role',
    field: (row) => row.bestRoleOverall || row.BestRoleOverall || '—',
    align: 'left',
    sortable: true,
  },
  {
    name: 'ca',
    label: 'CA',
    field: (row) => row.ca ?? row.CA,
    align: 'right',
    sortable: true,
  },
  {
    name: 'overall',
    label: 'Overall',
    field: (row) => row.overall ?? row.Overall,
    align: 'right',
    sortable: true,
  },
  {
    name: 'caScaled',
    label: 'CA → 99',
    field: getCAScaled,
    align: 'right',
    sortable: true,
  },
  {
    name: 'ratingDiff',
    label: 'Δ (Ovr−CA)',
    field: getRatingDiff,
    align: 'right',
    sortable: true,
  },
]

// biome-ignore lint/correctness/noUnusedVariables: used in template
const outlierColumns = [
  { name: 'name', label: 'Name', field: 'name', align: 'left', sortable: true },
  {
    name: 'positions',
    label: 'Pos',
    field: (row) => (row.shortPositions || row.short_positions || []).join(', '),
    align: 'left',
  },
  { name: 'ca', label: 'CA', field: getCA, align: 'right', sortable: true },
  { name: 'overall', label: 'Overall', field: getOverall, align: 'right', sortable: true },
  { name: 'ratingDiff', label: 'Δ', field: getRatingDiff, align: 'right', sortable: true },
]

// biome-ignore lint/correctness/noUnusedVariables: used in template
const roleDistributionColumns = [
  { name: 'role', label: 'Best Role', field: 'role', align: 'left', sortable: true },
  { name: 'count', label: 'Count', field: 'count', align: 'right', sortable: true },
  {
    name: 'share',
    label: '% of filtered',
    field: 'share',
    align: 'right',
    sortable: true,
    format: (v) => `${v}%`,
  },
]

// biome-ignore lint/correctness/noUnusedVariables: used in template
const openPlayerDetail = (player) => {
  playerForDetailView.value = player
  showPlayerDetailDialog.value = true
}

const fetchPlayers = async (datasetId) => {
  pageLoading.value = true
  pageLoadingError.value = ''
  try {
    const performanceResponse = await fetchPerformanceData(datasetId)
    playerStore.setPlayers(performanceResponse.data.players)
    playerStore.setCurrencySymbol(performanceResponse.data.currencySymbol)
    playerStore.setCurrentDatasetId(datasetId)
  } catch (err) {
    pageLoadingError.value = `Failed to load save data: ${err.message || 'Unknown server error'}.`
  } finally {
    pageLoading.value = false
  }
}

onMounted(async () => {
  const datasetId =
    route.params.datasetId || route.query.datasetId || sessionStorage.getItem('currentDatasetId')

  if (!datasetId) {
    pageLoadingError.value = 'No dataset available. Please upload a save first.'
    pageLoading.value = false
    return
  }

  await fetchPlayers(datasetId)
})
</script>

<style scoped>
.page-container {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);
}

.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 300px;
    gap: 1rem;
}

.save-analysis-table :deep(tbody tr) {
    cursor: pointer;
}

@media (max-width: 768px) {
    .page-container {
        padding: var(--page-gutter-sm);
    }
}
</style>
