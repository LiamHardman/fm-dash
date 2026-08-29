<template>
  <q-page class="shortlists-page">
    <div class="page-container">
      <PageHeader
        title="Shortlists"
        subtitle="Plan recruitment across datasets. Player details are available only while their source dataset is loaded."
        icon="playlist_add_check"
      >
        <template #actions>
          <q-btn flat icon="download" label="Export" @click="exportList" />
          <q-btn unelevated color="primary" icon="add" label="New shortlist" @click="openCreateDialog" />
        </template>
      </PageHeader>

      <q-card flat bordered class="list-toolbar q-mb-md">
        <q-card-section class="row items-center q-col-gutter-md">
          <div class="col-12 col-sm">
            <q-select
              v-model="shortlistStore.activeListId"
              :options="listOptions"
              emit-value
              map-options
              outlined
              dense
              label="Active shortlist"
            />
          </div>
          <div class="col-auto row q-gutter-sm">
            <q-btn flat dense icon="edit" label="Rename" @click="openRenameDialog" />
            <q-btn
              flat
              dense
              color="negative"
              icon="delete_outline"
              label="Delete"
              :disable="shortlistStore.document.lists.length === 1"
              @click="confirmDelete"
            />
          </div>
        </q-card-section>
      </q-card>

      <div class="summary-grid q-mb-md">
        <StatTile icon="groups" label="Targets" :value="activeItems.length" />
        <StatTile icon="account_balance_wallet" label="Target fees" :value="formatMoney(targetFeeTotal)" />
        <StatTile icon="payments" label="Target wages" :value="formatMoney(targetWageTotal)" />
        <StatTile icon="warning_amber" label="Unavailable references" :value="unavailableCount" />
      </div>

      <SectionCard v-if="coverage.length" title="Current position coverage" icon="shield">
        <div class="coverage-row">
          <q-chip v-for="entry in coverage" :key="entry.position" outline color="primary">
            {{ entry.position }}: {{ entry.count }}
          </q-chip>
        </div>
      </SectionCard>

      <q-banner v-if="unavailableCount" rounded class="q-mt-md bg-orange-1 text-orange-10">
        {{ unavailableCount }} player reference{{ unavailableCount === 1 ? '' : 's' }} cannot be resolved because its source Dataset is not currently loaded. Planning metadata remains available.
      </q-banner>

      <EmptyState
        v-if="!activeItems.length"
        class="q-mt-md"
        icon="playlist_add"
        title="This Shortlist is empty"
        description="Add players from a player table using the Add to Shortlist action."
      >
        <template #actions>
          <q-btn
            unelevated
            color="primary"
            icon="search"
            label="Browse players"
            :to="currentDatasetId ? `/dataset/${currentDatasetId}` : '/upload'"
          />
        </template>
      </EmptyState>

      <div v-else class="q-mt-md">
        <section v-for="group in groupedItems" :key="group.status" class="status-section">
          <div class="status-heading">
            <h2>{{ group.label }}</h2>
            <q-badge color="primary" :label="group.items.length" />
          </div>
          <q-card v-for="entry in group.items" :key="entry.key" flat bordered class="target-card">
            <q-card-section class="target-grid">
              <div class="target-player">
                <template v-if="entry.player">
                  <div class="text-subtitle1">{{ entry.player.name }}</div>
                  <div class="text-caption text-grey-7">
                    {{ entry.player.club || 'No club' }} · {{ entry.player.pos || entry.player.Position || 'Position unavailable' }}
                  </div>
                </template>
                <template v-else>
                  <div class="text-subtitle1">Player unavailable</div>
                  <div class="text-caption text-grey-7">Source Dataset is not loaded</div>
                </template>
              </div>
              <q-select
                v-model="entry.item.status"
                :options="statusOptions"
                emit-value
                map-options
                outlined
                dense
                label="Status"
                @update:model-value="saveItem(entry.item)"
              />
              <q-select
                v-model="entry.item.priority"
                :options="priorityOptions"
                emit-value
                map-options
                outlined
                dense
                label="Priority"
                @update:model-value="saveItem(entry.item)"
              />
              <q-input
                v-model.number="entry.item.targetFee"
                outlined
                dense
                type="number"
                min="0"
                label="Target fee"
                @update:model-value="saveItem(entry.item)"
              />
              <q-input
                v-model.number="entry.item.targetWage"
                outlined
                dense
                type="number"
                min="0"
                label="Target wage"
                @update:model-value="saveItem(entry.item)"
              />
              <q-btn flat round color="negative" icon="close" aria-label="Remove target" @click="removeItem(entry.item)" />
            </q-card-section>
            <q-card-section class="target-notes q-pt-none">
              <q-input
                :model-value="entry.item.tags.join(', ')"
                outlined
                dense
                label="Tags"
                hint="Separate tags with commas"
                @update:model-value="updateTags(entry.item, $event)"
              />
              <q-input
                v-model="entry.item.notes"
                outlined
                dense
                autogrow
                label="Notes"
                @blur="saveItem(entry.item)"
              />
            </q-card-section>
          </q-card>
        </section>
      </div>
    </div>

    <q-dialog v-model="showListDialog">
      <q-card style="min-width: 320px">
        <q-card-section><div class="text-h6">{{ listDialogMode === 'create' ? 'New shortlist' : 'Rename shortlist' }}</div></q-card-section>
        <q-card-section class="q-pt-none"><q-input v-model="listName" outlined autofocus label="Name" @keyup.enter="submitListDialog" /></q-card-section>
        <q-card-actions align="right"><q-btn flat label="Cancel" v-close-popup /><q-btn unelevated color="primary" label="Save" :disable="!listName.trim()" @click="submitListDialog" /></q-card-actions>
      </q-card>
    </q-dialog>

    <q-dialog v-model="showDeleteDialog">
      <q-card style="min-width: 320px">
        <q-card-section><div class="text-h6">Delete shortlist?</div></q-card-section>
        <q-card-section class="q-pt-none">This removes the list and its recruitment planning metadata. It cannot be undone.</q-card-section>
        <q-card-actions align="right"><q-btn flat label="Cancel" v-close-popup /><q-btn flat color="negative" label="Delete" @click="deleteList" /></q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script>
import { computed, defineComponent, onMounted, ref } from 'vue'
import EmptyState from '../components/layout/EmptyState.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import StatTile from '../components/layout/StatTile.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useShortlistStore } from '../stores/shortlistStore'

const statusOptions = [
  { label: 'Watching', value: 'watching' },
  { label: 'Scouting', value: 'scouting' },
  { label: 'Bid planned', value: 'bid_planned' },
  { label: 'Signed', value: 'signed' },
  { label: 'Not pursuing', value: 'not_pursuing' },
]
const priorityOptions = [
  { label: 'High', value: 'high' },
  { label: 'Medium', value: 'medium' },
  { label: 'Low', value: 'low' },
]
const includedInBudget = new Set(['watching', 'scouting', 'bid_planned'])

export default defineComponent({
  name: 'ShortlistsPage',
  components: { EmptyState, PageHeader, SectionCard, StatTile },
  setup() {
    const playerStore = usePlayerStore()
    const shortlistStore = useShortlistStore()
    const showListDialog = ref(false)
    const showDeleteDialog = ref(false)
    const listDialogMode = ref('create')
    const listName = ref('')
    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const activeItems = computed(() => shortlistStore.activeList?.items || [])
    const playerByReference = computed(() => {
      const datasetId = currentDatasetId.value
      const players = playerStore.allPlayers || []
      return new Map(players.map((player) => [`${datasetId}:${player.uid ?? player.UID}`, player]))
    })
    const entries = computed(() =>
      activeItems.value.map((item) => {
        const key = `${item.playerRef.datasetId}:${item.playerRef.playerUid}`
        return { item, key, player: playerByReference.value.get(key) || null }
      })
    )
    const groupedItems = computed(() =>
      statusOptions
        .map((status) => ({
          ...status,
          items: entries.value.filter((entry) => entry.item.status === status.value),
        }))
        .filter((group) => group.items.length)
    )
    const unavailableCount = computed(() => entries.value.filter((entry) => !entry.player).length)
    const budgetItems = computed(() =>
      activeItems.value.filter((item) => includedInBudget.has(item.status))
    )
    const targetFeeTotal = computed(() =>
      budgetItems.value.reduce((total, item) => total + Number(item.targetFee || 0), 0)
    )
    const targetWageTotal = computed(() =>
      budgetItems.value.reduce((total, item) => total + Number(item.targetWage || 0), 0)
    )
    const coverage = computed(() => {
      const counts = new Map()
      for (const entry of entries.value) {
        if (!entry.player || !includedInBudget.has(entry.item.status)) continue
        const position = entry.player.pos || entry.player.Position
        if (position) counts.set(position, (counts.get(position) || 0) + 1)
      }
      return [...counts.entries()].map(([position, count]) => ({ position, count }))
    })
    const listOptions = computed(() =>
      shortlistStore.document.lists.map((list) => ({ label: list.name, value: list.id }))
    )

    onMounted(() => shortlistStore.load())

    const saveItem = (item) => shortlistStore.updateItem(item, {})
    const updateTags = (item, value) =>
      shortlistStore.updateItem(item, {
        tags: String(value)
          .split(',')
          .map((tag) => tag.trim())
          .filter(Boolean),
      })
    const removeItem = (item) => shortlistStore.removeItem(item)
    const openCreateDialog = () => {
      listDialogMode.value = 'create'
      listName.value = ''
      showListDialog.value = true
    }
    const openRenameDialog = () => {
      listDialogMode.value = 'rename'
      listName.value = shortlistStore.activeList?.name || ''
      showListDialog.value = true
    }
    const submitListDialog = async () => {
      if (listDialogMode.value === 'create') await shortlistStore.createList(listName.value)
      else await shortlistStore.renameActiveList(listName.value)
      showListDialog.value = false
    }
    const confirmDelete = () => {
      showDeleteDialog.value = true
    }
    const deleteList = async () => {
      await shortlistStore.deleteActiveList()
      showDeleteDialog.value = false
    }
    const formatMoney = (value) =>
      new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency: 'USD',
        maximumFractionDigits: 0,
      }).format(value || 0)
    const csvCell = (value) => `"${String(value ?? '').replaceAll('"', '""')}"`
    const exportList = () => {
      const rows = [['Player', 'Status', 'Priority', 'Target fee', 'Target wage', 'Tags', 'Notes']]
      for (const entry of entries.value) {
        rows.push([
          entry.player?.name || 'Player unavailable',
          entry.item.status,
          entry.item.priority,
          entry.item.targetFee,
          entry.item.targetWage,
          entry.item.tags.join(', '),
          entry.item.notes,
        ])
      }
      const link = document.createElement('a')
      link.href = URL.createObjectURL(
        new Blob([rows.map((row) => row.map(csvCell).join(',')).join('\n')], { type: 'text/csv' })
      )
      link.download = `${shortlistStore.activeList?.name || 'shortlist'}.csv`
      link.click()
      URL.revokeObjectURL(link.href)
    }

    return {
      shortlistStore,
      currentDatasetId,
      activeItems,
      groupedItems,
      unavailableCount,
      targetFeeTotal,
      targetWageTotal,
      coverage,
      listOptions,
      statusOptions,
      priorityOptions,
      showListDialog,
      showDeleteDialog,
      listDialogMode,
      listName,
      saveItem,
      updateTags,
      removeItem,
      openCreateDialog,
      openRenameDialog,
      submitListDialog,
      confirmDelete,
      deleteList,
      formatMoney,
      exportList,
    }
  },
})
</script>

<style lang="scss" scoped>
.shortlists-page { min-height: 100vh; }
.page-container { max-width: var(--content-max-width); margin: 0 auto; padding: var(--page-gutter); }
.list-toolbar, .target-card { background: var(--surface-card); }
.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: var(--section-gap); }
.coverage-row { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.status-section { margin-top: 1.5rem; }
.status-heading { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.status-heading h2 { font-size: 1rem; margin: 0; }
.target-card + .target-card { margin-top: 0.75rem; }
.target-grid { display: grid; grid-template-columns: minmax(180px, 2fr) repeat(4, minmax(120px, 1fr)) auto; gap: 0.75rem; align-items: center; }
.target-notes { display: grid; grid-template-columns: minmax(160px, 1fr) minmax(220px, 2fr); gap: 0.75rem; }
@media (max-width: 900px) { .target-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .target-player { grid-column: 1 / -1; } }
@media (max-width: 600px) { .page-container { padding: var(--page-gutter-sm); } .target-grid, .target-notes { grid-template-columns: 1fr; } }
</style>
