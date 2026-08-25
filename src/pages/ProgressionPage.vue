<template>
  <q-page class="progression-page">
    <div class="page-container">
      <PageHeader
        title="Player Progression"
        subtitle="Upload 2 or more exports from the same save to see how your players have developed between snapshots. Order is detected automatically from average squad age."
        icon="trending_up"
      />

      <div v-if="progressionStore.status === 'analyzed'" class="progression-stats">
        <StatTile icon="groups" label="Shared players" :value="progressionStore.players.length" />
        <StatTile icon="layers" label="Snapshots" :value="progressionStore.order.length" />
      </div>

      <!-- Upload -->
      <SectionCard :title="`Snapshots (${slots.length})`" icon="upload_file" class="q-mb-md">
        <template #actions>
          <q-btn
            unelevated
            dense
            color="primary"
            icon="add"
            label="Add another snapshot"
            @click="triggerFilePicker"
            :loading="anyUploading"
          />
          <input
            ref="fileInput"
            type="file"
            accept=".html,.csv"
            class="hidden-file-input"
            @change="onFileSelected"
          />
        </template>

        <q-list v-if="slots.length" bordered separator class="snapshot-list">
          <q-item v-for="slot in slots" :key="slot.id">
            <q-item-section avatar>
              <q-spinner v-if="slot.status === 'uploading'" color="primary" size="20px" />
              <q-icon
                v-else
                :name="slot.status === 'parsed' ? 'check_circle' : 'error'"
                :color="slot.status === 'parsed' ? 'positive' : 'negative'"
              />
            </q-item-section>
            <q-item-section>
              <div>{{ slot.filename }}</div>
              <div v-if="slot.status === 'error'" class="text-caption text-negative">
                {{ slot.errorMessage }}
              </div>
            </q-item-section>
            <q-item-section side>
              <q-chip
                dense
                :color="slot.status === 'parsed' ? 'positive' : slot.status === 'error' ? 'negative' : 'grey-6'"
                text-color="white"
              >
                {{ slot.status }}
              </q-chip>
            </q-item-section>
            <q-item-section side>
              <q-btn dense flat round icon="close" size="sm" @click="progressionStore.removeSlot(slot.id)" />
            </q-item-section>
          </q-item>
        </q-list>

        <EmptyState
          v-else
          icon="upload_file"
          title="No snapshots yet"
          description="Add at least 2 snapshots to analyze."
        />

        <div class="row items-center q-mt-md q-gutter-sm">
          <q-btn
            unelevated
            color="primary"
            label="Analyze"
            :disable="parsedCount < 2 || progressionStore.status === 'analyzing'"
            :loading="progressionStore.status === 'analyzing'"
            @click="progressionStore.analyze()"
          />
          <q-btn
            v-if="progressionStore.status !== 'idle'"
            flat
            dense
            label="Start over"
            @click="startOver"
          />
        </div>

        <q-banner v-if="progressionStore.status === 'ambiguous-order'" class="bg-warning text-dark q-mt-md" rounded>
          <div class="text-weight-bold q-mb-xs">
            Order ambiguous — these snapshots have the same average squad age
          </div>
          <div class="text-caption q-mb-sm">Drag isn't available yet — use the arrows to reorder.</div>
          <q-list bordered separator class="surface-card">
            <q-item v-for="(id, i) in reorderList" :key="id">
              <q-item-section>{{ i + 1 }}. {{ filenameFor(id) }}</q-item-section>
              <q-item-section side>
                <q-btn dense flat icon="arrow_upward" :disable="i === 0" @click="moveReorder(i, -1)" />
                <q-btn
                  dense
                  flat
                  icon="arrow_downward"
                  :disable="i === reorderList.length - 1"
                  @click="moveReorder(i, 1)"
                />
              </q-item-section>
            </q-item>
          </q-list>
          <q-btn class="q-mt-sm" unelevated color="primary" label="Confirm order" @click="confirmReorder" />
        </q-banner>

        <q-banner v-if="progressionStore.status === 'error'" class="bg-negative text-white q-mt-md" rounded>
          {{ progressionStore.errorMessage }}
        </q-banner>
      </SectionCard>

      <!-- Empty intersection -->
      <q-banner
        v-if="progressionStore.status === 'analyzed' && progressionStore.emptyIntersection"
        class="bg-info text-white q-mb-md"
        rounded
      >
        <template #avatar><q-icon name="info" /></template>
        No players found in every uploaded save — check these are from the same game.
      </q-banner>

      <!-- Results -->
      <div v-if="progressionStore.status === 'analyzed' && !progressionStore.emptyIntersection">
        <div class="row items-center q-gutter-md q-mb-sm results-toolbar">
          <q-select
            dense
            outlined
            style="min-width: 220px"
            :options="sortFieldOptions"
            :model-value="progressionStore.sortField"
            label="Sort by change in…"
            emit-value
            map-options
            @update:model-value="(field) => progressionStore.setSort(field, progressionStore.sortDir)"
          />
          <q-btn
            dense
            flat
            :icon="progressionStore.sortDir === 'desc' ? 'arrow_downward' : 'arrow_upward'"
            :label="progressionStore.sortDir === 'desc' ? 'Highest first' : 'Lowest first'"
            @click="toggleSortDir"
          />
          <q-space />
          <div class="text-caption text-grey-6">
            {{ filteredPlayers.length }} of {{ progressionStore.players.length }} players
          </div>
          <q-btn
            dense
            flat
            :icon="showFilters ? 'filter_list_off' : 'filter_list'"
            :label="showFilters ? 'Hide filters' : 'Filters'"
            @click="showFilters = !showFilters"
          />
        </div>

        <q-slide-transition>
          <SectionCard v-show="showFilters" class="q-mb-md">
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-4">
                <q-input dense outlined v-model="filters.name" label="Name" clearable />
              </div>
              <div class="col-12 col-sm-4">
                <q-select dense outlined clearable v-model="filters.club" :options="clubOptions" label="Club" />
              </div>
              <div class="col-12 col-sm-4">
                <q-select
                  dense
                  outlined
                  multiple
                  clearable
                  use-chips
                  v-model="filters.position"
                  :options="positionOptions"
                  label="Position"
                />
              </div>
              <div class="col-12 col-sm-4">
                <q-select
                  dense
                  outlined
                  multiple
                  clearable
                  use-chips
                  v-model="filters.nationality"
                  :options="nationalityOptions"
                  label="Nationality"
                />
              </div>
              <div class="col-12 col-sm-4">
                <q-select
                  dense
                  outlined
                  multiple
                  clearable
                  use-chips
                  v-model="filters.division"
                  :options="divisionOptions"
                  label="Division"
                />
              </div>
              <div class="col-6 col-sm-2">
                <q-input dense outlined type="number" v-model.number="filters.ageRange.min" label="Min age" />
              </div>
              <div class="col-6 col-sm-2">
                <q-input dense outlined type="number" v-model.number="filters.ageRange.max" label="Max age" />
              </div>
              <div class="col-12 col-sm-4">
                <q-input dense outlined type="number" v-model.number="filters.minOverall" label="Min overall" />
              </div>
            </div>
          </SectionCard>
        </q-slide-transition>

        <SectionCard>
          <q-markup-table flat class="progression-table">
            <thead>
              <tr>
                <th class="expand-col"></th>
                <th class="text-left">Name</th>
                <th class="text-left">Position</th>
                <th class="text-left">Club</th>
                <th class="text-center">Overall</th>
                <th class="text-right">Overall Δ</th>
                <th class="text-right">Value</th>
                <th class="text-right">Value Δ</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="p in filteredPlayers" :key="p.uid">
                <tr class="row-clickable" @click="toggleExpand(p.uid)">
                  <td class="expand-col">
                    <q-icon :name="expandedUid === p.uid ? 'expand_less' : 'expand_more'" color="grey-6" />
                  </td>
                  <td>
                    <div class="name-cell">
                      <img
                        v-if="latestOf(p).nationalityIso"
                        :src="`https://flagcdn.com/w20/${latestOf(p).nationalityIso.toLowerCase()}.png`"
                        :alt="latestOf(p).nationality || 'Flag'"
                        width="16"
                        height="11"
                        class="name-flag"
                      />
                      <span class="text-weight-medium player-name-link" @click.stop="openPlayerDetail(p)">
                        {{ latestOf(p).name }}
                      </span>
                    </div>
                  </td>
                  <td>{{ latestOf(p).position }}</td>
                  <td>
                    <div class="club-cell">
                      <TeamLogo
                        v-if="latestOf(p).club && latestOf(p).club !== '-'"
                        :team-name="latestOf(p).club"
                        :size="22"
                      />
                      <span>{{ latestOf(p).club }}</span>
                    </div>
                  </td>
                  <td class="text-center">
                    <StatGauge :value="latestOf(p).overall" :max="100" :size="38" />
                  </td>
                  <td class="text-right">
                    <span :class="deltaClass(p, overallField)">{{ deltaLabel(p, overallField) }}</span>
                  </td>
                  <td class="text-right money-value" :class="latestOf(p).transferValueAmount ? 'money-uniform' : 'money-na'">
                    {{ formatCurrency(latestOf(p).transferValueAmount, progressionStore.currencySymbol) }}
                  </td>
                  <td class="text-right">
                    <span :class="deltaClass(p, valueField)">{{ deltaLabel(p, valueField) }}</span>
                  </td>
                </tr>
                <tr v-if="expandedUid === p.uid" class="expand-row">
                  <td colspan="8">
                    <div class="row q-gutter-lg q-pa-md">
                      <div v-for="f in ALL_FIELDS" :key="f.key" class="col-auto trend-item">
                        <div class="row items-center justify-between q-mb-xs trend-item-header">
                          <span class="text-caption text-grey-6">{{ f.label }}</span>
                          <span class="text-caption text-weight-medium" :class="deltaClass(p, f)">
                            {{ deltaLabel(p, f) }}
                          </span>
                        </div>
                        <ProgressionTrendChart
                          :labels="snapshotDateLabels"
                          :values="p.snapshots.map((s) => f.accessor(s))"
                          :color="trendColor(p, f)"
                          :width="150"
                          :height="46"
                        />
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </q-markup-table>
        </SectionCard>
      </div>
    </div>

    <PlayerDetailDialog
      :player="detailDialogPlayer"
      :show="showPlayerDetail"
      :currency-symbol="progressionStore.currencySymbol"
      :dataset-id="detailDialogDatasetId"
      :snapshot-tabs="snapshotTabOptions"
      :active-snapshot-index="activeSnapshotIndex"
      @update:active-snapshot-index="activeSnapshotIndex = $event"
      @close="showPlayerDetail = false"
    />
  </q-page>
</template>

<script setup>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import EmptyState from '../components/layout/EmptyState.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PageHeader from '../components/layout/PageHeader.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import SectionCard from '../components/layout/SectionCard.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import StatTile from '../components/layout/StatTile.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import ProgressionTrendChart from '../components/ProgressionTrendChart.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import StatGauge from '../components/player-table/StatGauge.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import TeamLogo from '../components/TeamLogo.vue'
import { filterLatestSnapshotPlayers, uniqueValues } from '../composables/useLatestSnapshotFilters'
import playerService from '../services/playerService'
import { useProgressionStore } from '../stores/progressionStore'
import { formatCurrency } from '../utils/currencyUtils'

const $q = useQuasar()
const progressionStore = useProgressionStore()
const fileInput = ref(null)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const showFilters = ref(false)
const expandedUid = ref(null)
const reorderList = ref([])
const detailPlayer = ref(null)
const activeSnapshotIndex = ref(0)
const showPlayerDetail = ref(false)

const slots = computed(() => progressionStore.slots)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const parsedCount = computed(() => slots.value.filter((s) => s.status === 'parsed').length)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const anyUploading = computed(() => slots.value.some((s) => s.status === 'uploading'))

const ALL_FIELDS = [
  { key: 'Overall', label: 'Overall', accessor: (p) => p.overall, format: (v) => `${v}` },
  {
    key: 'Value',
    label: 'Value',
    accessor: (p) => p.transferValueAmount,
    format: (v) => formatCurrency(v, progressionStore.currencySymbol),
  },
  {
    key: 'Wage',
    label: 'Wage',
    accessor: (p) => p.wageAmount,
    format: (v) => formatCurrency(v, progressionStore.currencySymbol),
  },
  {
    key: 'Age',
    label: 'Age',
    accessor: (p) => Number.parseInt(p.age, 10) || 0,
    format: (v) => `${v}`,
  },
]

// biome-ignore lint/correctness/noUnusedVariables: used in template
const overallField = ALL_FIELDS[0]
// biome-ignore lint/correctness/noUnusedVariables: used in template
const valueField = ALL_FIELDS[1]

// biome-ignore lint/correctness/noUnusedVariables: used in template
const sortFieldOptions = computed(() => {
  const attrKeys = new Set()
  for (const p of progressionStore.players) {
    const latest = p.snapshots[p.snapshots.length - 1]
    for (const key of Object.keys(latest?.numericAttributes || {})) attrKeys.add(key)
  }
  return [
    ...ALL_FIELDS.map((f) => ({ label: f.label, value: f.key })),
    ...Array.from(attrKeys)
      .sort()
      .map((key) => ({ label: key, value: key })),
  ]
})

const filters = ref({
  name: '',
  club: null,
  position: [],
  nationality: [],
  division: [],
  ageRange: { min: 0, max: 0 },
  transferValueRange: { min: 0, max: 0 },
  maxSalary: 0,
  minOverall: 0,
  minCA: 0,
  minMBR: 0,
})

const latestSnapshotPlayers = computed(() =>
  progressionStore.players.map((p) => ({
    ...p.snapshots[p.snapshots.length - 1],
    __progressionUid: p.uid,
  }))
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const filteredPlayers = computed(() => {
  const allowedUids = new Set(
    filterLatestSnapshotPlayers(latestSnapshotPlayers.value, filters.value).map(
      (p) => p.__progressionUid
    )
  )
  return progressionStore.players.filter((p) => allowedUids.has(p.uid))
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const clubOptions = computed(() => uniqueValues(latestSnapshotPlayers.value, 'club'))
// biome-ignore lint/correctness/noUnusedVariables: used in template
const nationalityOptions = computed(() => uniqueValues(latestSnapshotPlayers.value, 'nationality'))
// biome-ignore lint/correctness/noUnusedVariables: used in template
const divisionOptions = computed(() => uniqueValues(latestSnapshotPlayers.value, 'division'))
// biome-ignore lint/correctness/noUnusedVariables: used in template
const positionOptions = computed(() => {
  const set = new Set()
  for (const p of latestSnapshotPlayers.value) {
    for (const pos of p.shortPositions || []) set.add(pos)
  }
  return Array.from(set).sort()
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const snapshotDateLabels = computed(() => progressionStore.order.map((_, i) => `#${i + 1}`))

// Snapshot-tab labels for PlayerDetailDialog: one per snapshot dataset, in chronological
// order, labeled by the original upload filename so the tabs read like "barcajuly25.csv".
// biome-ignore lint/correctness/noUnusedVariables: used in template
const snapshotTabOptions = computed(() =>
  progressionStore.order.map((datasetId) => ({ label: filenameFor(datasetId) }))
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const detailDialogPlayer = computed(
  () => detailPlayer.value?.snapshots[activeSnapshotIndex.value] ?? null
)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const detailDialogDatasetId = computed(
  () => progressionStore.order[activeSnapshotIndex.value] ?? null
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
function openPlayerDetail(progressionPlayer) {
  detailPlayer.value = progressionPlayer
  activeSnapshotIndex.value = progressionPlayer.snapshots.length - 1 // default to the latest snapshot
  showPlayerDetail.value = true
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function latestOf(progressionPlayer) {
  return progressionPlayer.snapshots[progressionPlayer.snapshots.length - 1]
}

function fieldDelta(progressionPlayer, field) {
  const first = field.accessor(progressionPlayer.snapshots[0])
  const last = field.accessor(progressionPlayer.snapshots[progressionPlayer.snapshots.length - 1])
  return last - first
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function deltaLabel(progressionPlayer, field) {
  const delta = fieldDelta(progressionPlayer, field)
  const sign = delta > 0 ? '+' : ''
  return `${sign}${field.format(delta)}`
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function deltaClass(progressionPlayer, field) {
  const delta = fieldDelta(progressionPlayer, field)
  return delta > 0 ? 'text-positive' : delta < 0 ? 'text-negative' : 'text-grey-6'
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function trendColor(progressionPlayer, field) {
  const delta = fieldDelta(progressionPlayer, field)
  if (delta > 0) return $q.dark.isActive ? '#4ade80' : '#21ba45'
  if (delta < 0) return $q.dark.isActive ? '#f87171' : '#c10015'
  return $q.dark.isActive ? '#94a3b8' : '#9e9e9e'
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function toggleExpand(uid) {
  expandedUid.value = expandedUid.value === uid ? null : uid
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function triggerFilePicker() {
  fileInput.value?.click()
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
async function onFileSelected(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return

  const slotId = `slot-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  progressionStore.addSlot({
    id: slotId,
    filename: file.name,
    status: 'uploading',
    datasetId: null,
  })

  try {
    const formData = new FormData()
    formData.append('playerFile', file)
    const response = await playerService.uploadPlayerFile(formData)
    progressionStore.updateSlot(slotId, { status: 'parsed', datasetId: response.datasetId })
  } catch (error) {
    progressionStore.updateSlot(slotId, {
      status: 'error',
      errorMessage: error.message || 'Upload failed',
    })
  }
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function startOver() {
  progressionStore.reset()
  expandedUid.value = null
}

function filenameFor(datasetId) {
  return slots.value.find((s) => s.datasetId === datasetId)?.filename || datasetId
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function moveReorder(i, dir) {
  const j = i + dir
  ;[reorderList.value[i], reorderList.value[j]] = [reorderList.value[j], reorderList.value[i]]
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
async function confirmReorder() {
  await progressionStore.confirmOrder(reorderList.value)
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function toggleSortDir() {
  progressionStore.setSort(
    progressionStore.sortField,
    progressionStore.sortDir === 'desc' ? 'asc' : 'desc'
  )
}

onMounted(() => {
  progressionStore.reset()
})

// Keep the reorder list in sync whenever we enter the ambiguous-order state
watch(
  () => progressionStore.status,
  (status) => {
    if (status === 'ambiguous-order') {
      reorderList.value = [...progressionStore.ambiguousDatasetIds]
    }
  }
)
</script>

<style scoped>
.page-container {
  max-width: var(--content-max-width);
  margin: 0 auto;
  padding: var(--page-gutter);
}

.progression-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--section-gap);
  margin-bottom: var(--section-gap);
}

.surface-card {
  background: var(--surface-card);
  border-color: var(--surface-border);
}

.hidden-file-input {
  display: none;
}

.snapshot-list {
  border-color: var(--surface-border);
}

.results-toolbar {
  flex-wrap: wrap;
}

.row-clickable {
  cursor: pointer;
}
.row-clickable:hover {
  background: var(--accent-soft);
}

.expand-col {
  width: 36px;
}

.expand-row td {
  background: var(--surface-raised);
  padding: 0 !important;
}

.trend-item {
  min-width: 150px;
}

.trend-item-header {
  min-width: 150px;
}

.progression-table :deep(th) {
  font-weight: 600;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.progression-table :deep(td) {
  padding-top: 10px;
  padding-bottom: 10px;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.name-flag {
  border-radius: 2px;
  flex-shrink: 0;
}

.player-name-link {
  cursor: pointer;
}
.player-name-link:hover {
  color: var(--accent);
  text-decoration: underline;
}

.club-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.money-value {
  font-weight: 500;
}

.money-uniform {
  color: var(--text-primary);
}

.money-na {
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .page-container {
    padding: var(--page-gutter-sm);
  }
}
</style>
