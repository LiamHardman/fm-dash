<template>
    <q-page class="scouting-book-page">
        <div class="page-container">
            <EmptyState
                v-if="!currentDatasetId"
                icon="folder_off"
                title="No dataset loaded"
                description="Please upload a dataset first to use the Scouting Book."
            >
                <template #actions>
                    <q-btn unelevated color="primary" icon="upload" label="Go to Upload Page" @click="router.push('/upload')" />
                </template>
            </EmptyState>

            <div v-if="currentDatasetId">
                <PageHeader
                    title="Scouting Book"
                    subtitle="Every player you've generated an AI Scout Report for, in one place."
                    icon="menu_book"
                />

                <div v-if="entries.length > 0" class="book-stats">
                    <StatTile icon="menu_book" label="Scouted" :value="entries.length" />
                    <StatTile icon="military_tech" label="A-grade or better" :value="aGradeCount" />
                    <StatTile icon="place" label="Positions covered" :value="positionCount" />
                </div>

                <SectionCard v-if="entries.length > 0" title="Reports" icon="list">
                    <div class="toolbar">
                        <q-btn-toggle
                            v-model="positionFilter"
                            :options="positionOptions"
                            dense
                            unelevated
                            toggle-color="primary"
                            clearable
                        />
                        <q-space />
                        <q-select
                            v-model="sortBy"
                            :options="sortOptions"
                            dense
                            outlined
                            emit-value
                            map-options
                            style="width: 200px"
                            label="Sort by"
                        />
                    </div>

                    <table class="ledger-table">
                        <thead>
                            <tr>
                                <th>Player</th>
                                <th>Position</th>
                                <th>Club</th>
                                <th>Grade</th>
                                <th>Squad</th>
                                <th>Division</th>
                                <th>Scouted</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="row in sortedRows" :key="row.playerUid + row.position">
                                <td class="name-cell" @click="openReport(row)">{{ row.playerName }}</td>
                                <td><span class="pos-chip">{{ row.position }}</span></td>
                                <td class="text-secondary">{{ row.club }}</td>
                                <td><span class="grade-badge" :style="{ background: gradeColor(row.grade) }">{{ row.grade }}</span></td>
                                <td><StarRating :stars="row.squadStars" :show-label="false" /></td>
                                <td><StarRating :stars="row.divisionStars" :show-label="false" /></td>
                                <td class="text-muted">{{ relativeTime(row.generatedAt) }}</td>
                                <td class="actions-cell">
                                    <q-btn dense flat round icon="favorite_border" size="sm" @click="addToWishlist(row)">
                                        <q-tooltip>Add to wishlist</q-tooltip>
                                    </q-btn>
                                    <q-btn dense flat round icon="close" size="sm" @click="removeReport(row)">
                                        <q-tooltip>Remove</q-tooltip>
                                    </q-btn>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </SectionCard>

                <EmptyState
                    v-else
                    icon="menu_book"
                    title="Your Scouting Book is empty"
                    description="Generate an AI Scout Report for a player — from their player page, or by asking the chat assistant — and it'll show up here."
                >
                    <template #actions>
                        <q-btn unelevated icon="search" label="Browse Players" color="primary" size="lg" @click="goToDataset" />
                    </template>
                </EmptyState>
            </div>
        </div>

        <PlayerDetailDialog
            v-if="playerForDetailView"
            :key="(playerForDetailView.uid ?? playerForDetailView.UID) + '-' + openPosition"
            :player="playerForDetailView"
            :show="showPlayerDetailDialog"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
            initial-tab="scoutReport"
            :initial-scout-position="openPosition"
            @close="showPlayerDetailDialog = false"
        />
    </q-page>
</template>

<script>
import { Notify } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import EmptyState from '../components/layout/EmptyState.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import StatTile from '../components/layout/StatTile.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import StarRating from '../components/StarRating.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'

const GRADE_COLORS = {
  'A+': '#16a34a',
  A: '#22c55e',
  'B+': '#3b82f6',
  B: '#60a5fa',
  'C+': '#f59e0b',
  C: '#f97316',
  D: '#ef4444',
}
const GRADE_ORDER = ['A+', 'A', 'B+', 'B', 'C+', 'C', 'D']

export default defineComponent({
  name: 'ScoutingBookPage',
  components: { PageHeader, SectionCard, StatTile, EmptyState, StarRating, PlayerDetailDialog },
  setup() {
    const router = useRouter()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()

    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)

    const entries = ref([])
    const positionFilter = ref(null)
    const sortBy = ref('date_desc')

    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)
    const openPosition = ref('')

    const positionOptions = computed(() => {
      const set = [...new Set(entries.value.map((e) => e.position))]
      return set.map((p) => ({ label: p, value: p }))
    })
    const sortOptions = [
      { label: 'Most recently scouted', value: 'date_desc' },
      { label: 'Grade (best first)', value: 'grade_desc' },
      { label: 'Squad stars', value: 'squad_desc' },
    ]

    const sortedRows = computed(() => {
      let list = entries.value
      if (positionFilter.value) list = list.filter((e) => e.position === positionFilter.value)
      list = [...list]
      if (sortBy.value === 'grade_desc') {
        list.sort((a, b) => GRADE_ORDER.indexOf(a.grade) - GRADE_ORDER.indexOf(b.grade))
      } else if (sortBy.value === 'squad_desc') {
        list.sort((a, b) => b.squadStars - a.squadStars)
      } else {
        list.sort((a, b) => new Date(b.generatedAt) - new Date(a.generatedAt))
      }
      return list
    })

    const aGradeCount = computed(
      () => entries.value.filter((e) => e.grade === 'A+' || e.grade === 'A').length
    )
    const positionCount = computed(() => new Set(entries.value.map((e) => e.position)).size)

    function gradeColor(grade) {
      return GRADE_COLORS[grade] || '#64748b'
    }

    function relativeTime(iso) {
      const diffMs = Date.now() - new Date(iso).getTime()
      const days = Math.floor(diffMs / 86400000)
      if (days <= 0) {
        const hours = Math.floor(diffMs / 3600000)
        if (hours <= 0) return 'just now'
        return hours === 1 ? '1 hour ago' : `${hours} hours ago`
      }
      if (days === 1) return 'yesterday'
      return `${days} days ago`
    }

    async function loadEntries() {
      if (!currentDatasetId.value) return
      try {
        const res = await fetch(`/api/scout-reports/${currentDatasetId.value}`)
        if (!res.ok) return
        entries.value = (await res.json()) || []
      } catch (e) {
        console.error('[ScoutingBookPage] failed to load Scouting Book', e)
      }
    }

    function resolvePlayer(uid) {
      const target = String(uid)
      return (playerStore.allPlayers || []).find((p) => String(p.uid ?? p.UID) === target) || null
    }

    function openReport(row) {
      const player = resolvePlayer(row.playerUid)
      if (!player) {
        Notify.create({
          type: 'warning',
          message: 'That player is no longer in the loaded dataset.',
        })
        return
      }
      playerForDetailView.value = player
      openPosition.value = row.position
      showPlayerDetailDialog.value = true
    }

    async function addToWishlist(row) {
      const player = resolvePlayer(row.playerUid)
      if (!player) {
        Notify.create({
          type: 'warning',
          message: 'That player is no longer in the loaded dataset.',
        })
        return
      }
      const added = await wishlistStore.addToWishlist(currentDatasetId.value, player)
      Notify.create({
        type: added ? 'positive' : 'info',
        message: added
          ? `${player.name} added to wishlist`
          : `${player.name} is already on your wishlist`,
        position: 'top',
      })
    }

    async function removeReport(row) {
      try {
        const params = new URLSearchParams({ playerUid: row.playerUid, position: row.position })
        await fetch(`/api/scout-report/${currentDatasetId.value}?${params}`, { method: 'DELETE' })
      } catch (e) {
        console.error('[ScoutingBookPage] failed to remove report', e)
      } finally {
        // Self-healing regardless of the DELETE's outcome (Scout Report v2 map ticket 01:
        // a 404 here means "already removed", not a real failure) — drop the row locally.
        entries.value = entries.value.filter(
          (e) => !(e.playerUid === row.playerUid && e.position === row.position)
        )
      }
    }

    function goToDataset() {
      router.push(currentDatasetId.value ? `/dataset/${currentDatasetId.value}` : '/upload')
    }

    onMounted(loadEntries)

    return {
      router,
      currentDatasetId,
      detectedCurrencySymbol,
      entries,
      positionFilter,
      positionOptions,
      sortBy,
      sortOptions,
      sortedRows,
      aGradeCount,
      positionCount,
      gradeColor,
      relativeTime,
      playerForDetailView,
      showPlayerDetailDialog,
      openPosition,
      openReport,
      addToWishlist,
      removeReport,
      goToDataset,
    }
  },
})
</script>

<style lang="scss" scoped>
.scouting-book-page {
    min-height: 100vh;
}
.page-container {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);
}
.book-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--section-gap);
    margin-bottom: var(--section-gap);
}
.toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
}
.ledger-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
}
.ledger-table th {
    text-align: left;
    color: var(--text-muted);
    font-weight: 600;
    padding: 8px 10px;
    border-bottom: 2px solid var(--surface-border);
    text-transform: uppercase;
    font-size: 11px;
    letter-spacing: 0.04em;
}
.ledger-table td {
    padding: 10px;
    border-bottom: 1px solid var(--surface-border);
    vertical-align: middle;
}
.name-cell {
    color: var(--text-primary);
    font-weight: 600;
    cursor: pointer;
}
.name-cell:hover {
    text-decoration: underline;
    color: var(--accent);
}
.pos-chip {
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
    border-radius: 6px;
    padding: 2px 6px;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
}
.text-secondary {
    color: var(--text-secondary);
}
.text-muted {
    color: var(--text-muted);
    font-size: 12px;
}
.grade-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    font-weight: 700;
    color: #fff;
    font-size: 11px;
    padding: 3px 8px;
}
.actions-cell {
    white-space: nowrap;
}

@media (max-width: 768px) {
    .page-container {
        padding: var(--page-gutter-sm);
    }
}
</style>
