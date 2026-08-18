<template>
    <div class="scout-report-tab">
        <div v-if="managedTeamMissing" class="managed-team-missing">
            <q-icon name="shield" size="2em" class="q-mb-sm" />
            <div class="text-subtitle1">Set your managed team first</div>
            <div class="text-caption text-grey-6">
                AI Scout Report compares this player against your own squad — use the
                "Set managed team" link on the dataset page above.
            </div>
        </div>

        <div v-else-if="loading" class="scout-loading">
            <q-spinner-dots color="primary" size="2.5em" />
            <div class="funny-message">{{ loadingMessage }}</div>
        </div>

        <div v-else-if="error" class="scout-error">
            <q-icon name="error_outline" size="2em" class="q-mb-sm text-negative" />
            <div class="text-body2">{{ error }}</div>
            <q-btn flat color="primary" label="Try again" icon="refresh" class="q-mt-sm" @click="generateReport" />
        </div>

        <div v-else class="dash-grid">
            <div class="cell verdict-cell">
                <div v-if="report" class="grade-badge" :style="{ background: gradeColor(report.subjectGrade) }">
                    {{ report.subjectGrade }}
                </div>
                <div class="verdict-name">{{ player.name }}</div>
                <div class="verdict-stars">
                    <StarRating :stars="stars.squadStars" label="Squad" />
                    <StarRating :stars="stars.divisionStars" label="Division" />
                </div>
                <div v-if="report" class="verdict-rationale">{{ report.subjectGradeRationale }}</div>
            </div>

            <div class="cell controls-cell">
                <div class="controls-heading">Compare As</div>
                <q-btn-toggle
                    v-model="position"
                    :options="positionOptions"
                    dense
                    unelevated
                    toggle-color="primary"
                    class="position-toggle"
                    @update:model-value="onPositionChange"
                />
                <q-btn
                    label="Regenerate Report"
                    icon="refresh"
                    color="primary"
                    outline
                    class="full-width q-mt-md"
                    size="sm"
                    @click="generateReport"
                />
                <div class="quick-facts">
                    <div><span>Value</span><strong>{{ formatCurrency(player.transferValueAmount, currencySymbol, player.transfer_value) }}</strong></div>
                    <div><span>Wage</span><strong>{{ formatCurrency(player.wageAmount, currencySymbol, player.wage) }}/wk</strong></div>
                    <div><span>Age</span><strong>{{ player.age }}</strong></div>
                    <div><span>Best Role</span><strong>{{ bestRoleLabel }}</strong></div>
                </div>
            </div>

            <template v-if="report">
                <div class="cell pros-cell">
                    <div class="pc-heading pros"><q-icon name="add_circle" size="16px" /> Pros</div>
                    <ul>
                        <li v-for="(p, i) in report.pros" :key="i">{{ p }}</li>
                    </ul>
                </div>
                <div class="cell cons-cell">
                    <div class="pc-heading cons"><q-icon name="remove_circle" size="16px" /> Cons</div>
                    <ul>
                        <li v-for="(c, i) in report.cons" :key="i">{{ c }}</li>
                    </ul>
                </div>

                <div class="cell table-cell">
                    <div class="section-heading">Comparable Players</div>
                    <div v-if="report.comparablePlayers.length === 0" class="text-caption text-grey-6">
                        No comparable players found for this search.
                    </div>
                    <table v-else class="comparable-table">
                        <thead>
                            <tr>
                                <th>Player</th>
                                <th>Club</th>
                                <th>Age</th>
                                <th>Value</th>
                                <th>Wage</th>
                                <th>Squad</th>
                                <th>Division</th>
                                <th>Grade</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="p in report.comparablePlayers" :key="p.uid">
                                <td class="name-cell name-cell-link" @click="openComparablePlayer(p.uid)">{{ p.name }}</td>
                                <td>{{ p.club }}</td>
                                <td>{{ p.age }}</td>
                                <td>{{ formatCurrency(p.transferValueAmount, currencySymbol) }}</td>
                                <td>{{ formatCurrency(p.wageAmount, currencySymbol) }}</td>
                                <td><StarRating :stars="p.squadStars" :show-label="false" /></td>
                                <td><StarRating :stars="p.divisionStars" :show-label="false" /></td>
                                <td><span class="grade-badge sm" :style="{ background: gradeColor(p.grade) }">{{ p.grade }}</span></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </template>
        </div>

        <!-- Own independent PlayerDetailDialog instance for drilling into a comparable player,
             mirroring UniversalSearch.vue / ChatWidget.vue's pattern rather than mutating the
             dialog this tab is already nested inside. Opens on a fresh instance, which defaults
             to the Simple tab, so drilling in never auto-fires another scout report request. -->
        <PlayerDetailDialog
            v-if="comparablePlayerForDetailView"
            :player="comparablePlayerForDetailView"
            :show="showComparablePlayerDialog"
            :currency-symbol="currencySymbol"
            :dataset-id="datasetId"
            @close="showComparablePlayerDialog = false"
        />
    </div>
</template>

<script>
import { computed, defineComponent, onMounted, onUnmounted, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { useScoutReportStore } from '../stores/scoutReportStore'
import { useUiStore } from '../stores/uiStore'
import { formatCurrency } from '../utils/currencyUtils'
import { llmRequestHeaders } from '../utils/llmHeaders'
import { consumeSSEStream } from '../utils/sseClient'
import PlayerDetailDialog from './PlayerDetailDialog.vue'
import StarRating from './StarRating.vue'

const FUNNY_LOADING_MESSAGES = [
  'Finding the next robotic striker...',
  'Checking the wage demands...',
  'Asking around the boot room...',
  'Cross-referencing scouting reports...',
  'Arguing with the chairman about the budget...',
  'Timing his sprint over 30 metres...',
  'Watching one more clip, just to be sure...',
]

const GRADE_COLORS = {
  'A+': '#16a34a',
  A: '#22c55e',
  'B+': '#3b82f6',
  B: '#60a5fa',
  'C+': '#f59e0b',
  C: '#f97316',
  D: '#ef4444',
}

export default defineComponent({
  name: 'ScoutReportTab',
  components: { StarRating, PlayerDetailDialog },
  props: {
    player: { type: Object, required: true },
    datasetId: { type: String, required: true },
    currencySymbol: { type: String, default: '$' },
  },
  setup(props) {
    const uiStore = useUiStore()
    const scoutReportStore = useScoutReportStore()
    const playerStore = usePlayerStore()

    const playerUid = computed(() => props.player.uid ?? props.player.UID)
    const positionOptions = computed(() => {
      const shorts = props.player.shortPositions || props.player.ShortPositions || []
      return shorts.map((p) => ({ label: p, value: p }))
    })

    function defaultPosition() {
      const shorts = props.player.shortPositions || props.player.ShortPositions || []
      return shorts[0] || ''
    }

    const position = ref(defaultPosition())
    const stars = ref({ squadStars: 0, divisionStars: 0 })
    const report = ref(null)
    const loading = ref(false)
    const error = ref(null)
    const managedTeamMissing = ref(false)
    const loadingMessage = ref(FUNNY_LOADING_MESSAGES[0])
    let loadingMessageTimer = null

    const bestRoleLabel = computed(() => {
      const role = props.player.bestRoleOverall || props.player.BestRoleOverall || ''
      const parts = role.split(' - ')
      return parts.length > 1 ? parts[1] : role || position.value
    })

    function gradeColor(grade) {
      return GRADE_COLORS[grade] || '#64748b'
    }

    const comparablePlayerForDetailView = ref(null)
    const showComparablePlayerDialog = ref(false)

    function openComparablePlayer(uid) {
      const target = String(uid)
      const found = (playerStore.allPlayers || []).find((p) => String(p.uid ?? p.UID) === target)
      if (!found) return
      comparablePlayerForDetailView.value = found
      showComparablePlayerDialog.value = true
    }

    function startLoadingMessages() {
      let idx = 0
      loadingMessageTimer = setInterval(() => {
        idx = (idx + 1) % FUNNY_LOADING_MESSAGES.length
        loadingMessage.value = FUNNY_LOADING_MESSAGES[idx]
      }, 1400)
    }
    function stopLoadingMessages() {
      clearInterval(loadingMessageTimer)
    }

    async function fetchStars() {
      if (!playerUid.value || !position.value) return
      try {
        const params = new URLSearchParams({ playerUid: playerUid.value, position: position.value })
        const res = await fetch(`/api/scout-report-stars/${props.datasetId}?${params}`)
        if (res.status === 400) {
          managedTeamMissing.value = true
          return
        }
        if (!res.ok) return
        managedTeamMissing.value = false
        stars.value = await res.json()
      } catch (e) {
        console.error('[ScoutReportTab] failed to fetch stars', e)
      }
    }

    async function generateReport() {
      if (!playerUid.value || !position.value || managedTeamMissing.value) return
      loading.value = true
      error.value = null
      startLoadingMessages()
      try {
        const res = await fetch(`/api/scout-report/${props.datasetId}`, {
          method: 'POST',
          headers: llmRequestHeaders(uiStore),
          body: JSON.stringify({ playerUid: playerUid.value, position: position.value }),
        })
        if (!res.ok || !res.body) {
          let message = `Unexpected error (${res.status}).`
          try {
            const body = await res.json()
            if (body?.error) message = body.error
          } catch (_e) {}
          error.value = message
          return
        }
        await consumeSSEStream(res, {
          done: (data) => {
            report.value = data
            scoutReportStore.set(props.datasetId, playerUid.value, position.value, data)
          },
          error: (payload) => {
            error.value = payload.message || 'Something went wrong.'
          },
        })
      } catch (e) {
        console.error('[ScoutReportTab] request threw', e)
        error.value = 'Could not reach the server. Check your connection and try again.'
      } finally {
        loading.value = false
        stopLoadingMessages()
      }
    }

    function onPositionChange() {
      fetchStars()
      const cached = scoutReportStore.get(props.datasetId, playerUid.value, position.value)
      report.value = cached
      error.value = null
      // No auto-refetch on position change (decided: explicit Regenerate only) — if nothing
      // is cached for this position yet, the dashboard just shows the controls until the
      // user presses Regenerate.
    }

    onMounted(async () => {
      await fetchStars()
      if (managedTeamMissing.value) return
      const cached = scoutReportStore.get(props.datasetId, playerUid.value, position.value)
      if (cached) {
        report.value = cached
      } else {
        await generateReport()
      }
    })
    onUnmounted(stopLoadingMessages)

    watch(
      () => props.player,
      () => {
        position.value = defaultPosition()
        report.value = null
        onPositionChange()
      }
    )

    return {
      position,
      positionOptions,
      stars,
      report,
      loading,
      error,
      managedTeamMissing,
      loadingMessage,
      bestRoleLabel,
      gradeColor,
      generateReport,
      onPositionChange,
      formatCurrency,
      comparablePlayerForDetailView,
      showComparablePlayerDialog,
      openComparablePlayer,
    }
  },
})
</script>

<style scoped>
.scout-report-tab {
    min-height: 200px;
}
.managed-team-missing,
.scout-loading,
.scout-error {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 48px 24px;
    gap: 6px;
}
.funny-message {
    color: var(--text-secondary, #64748b);
    font-style: italic;
    font-size: 14px;
    margin-top: 8px;
}

.dash-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-template-rows: auto auto auto;
    gap: 14px;
}
.table-cell {
    grid-column: 1 / -1;
}
.cell {
    background: var(--surface-card, #fff);
    border: 1px solid var(--surface-border, rgba(0, 0, 0, 0.08));
    border-radius: 10px;
    padding: 16px;
}
.verdict-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 6px;
    background: var(--surface-raised, #f8fafc);
}
.verdict-name {
    font-weight: 700;
    color: var(--text-primary, #1e293b);
}
.verdict-stars {
    display: flex;
    gap: 16px;
    margin: 6px 0;
}
.verdict-rationale {
    font-size: 12px;
    color: var(--text-secondary, #64748b);
    line-height: 1.5;
}
.controls-cell {
    display: flex;
    flex-direction: column;
}
.controls-heading,
.section-heading {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted, #94a3b8);
    margin-bottom: 10px;
}
.position-toggle {
    align-self: flex-start;
}
.quick-facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-top: 14px;
}
.quick-facts div {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: var(--text-secondary, #64748b);
    border-bottom: 1px dashed var(--surface-border, rgba(0, 0, 0, 0.08));
    padding-bottom: 4px;
}
.quick-facts strong {
    color: var(--text-primary, #1e293b);
}
.pc-heading {
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 600;
    margin-bottom: 8px;
}
.pc-heading.pros {
    color: #16a34a;
}
.pc-heading.cons {
    color: #dc2626;
}
.pros-cell ul,
.cons-cell ul {
    margin: 0;
    padding-left: 20px;
    font-size: 13px;
    color: var(--text-secondary, #64748b);
    line-height: 1.6;
}
.comparable-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
}
.comparable-table th {
    text-align: left;
    color: var(--text-muted, #94a3b8);
    font-weight: 600;
    padding: 6px 8px;
    border-bottom: 1px solid var(--surface-border, rgba(0, 0, 0, 0.08));
}
.comparable-table td {
    padding: 8px;
    border-bottom: 1px solid var(--surface-border, rgba(0, 0, 0, 0.08));
    color: var(--text-secondary, #64748b);
    vertical-align: middle;
}
.name-cell {
    color: var(--text-primary, #1e293b);
    font-weight: 600;
}
.name-cell-link {
    cursor: pointer;
}
.name-cell-link:hover {
    text-decoration: underline;
    color: var(--q-primary, #1976d2);
}
.grade-badge {
    display: inline-flex !important;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    font-weight: 700;
    line-height: 1;
    letter-spacing: 0.02em;
    color: #fff;
    font-size: 22px;
    padding: 10px 16px;
}
.grade-badge.sm {
    font-size: 11px;
    padding: 3px 7px;
    border-radius: 6px;
}
</style>
