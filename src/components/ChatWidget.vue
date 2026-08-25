<!-- FM-Dash Chatbot widget. Implements Variant B (full-height right-side drawer), selected
     from the three prototypes built for Wayfinder ticket "Frontend Chat Widget UX"
     (.scratch/chatbot/issues/07-frontend-chat-widget-ux.md; prototype set on branch
     prototype/chatbot-widget-ux, commit b39e05c — not promoted directly, rebuilt here as
     production code per that ticket's Answer).

     Mounted once at app-shell level in App.vue, mirroring UniversalSearch.vue's pattern of
     owning its own independent PlayerDetailDialog instance (ticket 05) rather than a global
     singleton. Dataset-scoped: the FAB only renders when a dataset is loaded and a managed
     team is set (map's widget-scope decision). -->
<template>
  <div v-if="chatStore.managedTeam" class="chat-widget-root">
    <transition name="chat-drawer">
      <div v-if="open" class="chat-drawer">
        <!-- Dialog chrome: header (icon/title/close), the same convention used by
             PlayerDetailDialog/SettingsModal — icon, title, actions, close, all in normal
             flow. Uses a q-avatar in place of the usual q-icon since the widget wants a
             filled "bot" glyph here. -->
        <div class="dialog-chrome chat-header">
          <div class="dialog-chrome__header">
            <q-avatar size="32px" color="primary" text-color="white" icon="smart_toy" class="q-mr-sm" />
            <div class="dialog-chrome__title-group">
              <div class="dialog-chrome__title">Scout Assistant</div>
              <div class="chat-title-sub">{{ chatStore.statusLabel || 'Ask about your squad, targets, tactics' }}</div>
            </div>
            <q-space />
            <div class="dialog-chrome__actions">
              <q-btn flat dense no-caps label="New Chat" icon="add_comment" :disable="chatStore.messages.length === 0" @click="confirmNewChatOpen = true" />
              <q-btn flat dense round icon="close" class="dialog-chrome__close" @click="open = false" />
            </div>
          </div>
        </div>

        <div class="chat-body" ref="scrollEl">
          <div v-if="chatStore.messages.length === 0" class="chat-empty">
            <div class="chat-empty-heading">What do you want to know?</div>
            <div class="chat-preset-grid">
              <q-card v-for="q in presetQuestions" :key="q" flat bordered class="chat-preset-card" @click="ask(q)">
                <q-card-section class="chat-preset-card-body">{{ q }}</q-card-section>
              </q-card>
            </div>
          </div>

          <template v-else>
            <div v-for="(m, i) in chatStore.messages" :key="i" class="chat-row" :class="`chat-row--${m.role}`">
              <q-avatar
                v-if="m.role !== 'user'"
                size="26px"
                :color="m.role === 'error' ? 'negative' : 'primary'"
                text-color="white"
                :icon="m.role === 'error' ? 'error_outline' : 'smart_toy'"
                class="chat-avatar"
              />
              <div class="chat-bubble" :class="`chat-bubble--${m.role}`">
                <template v-if="m.role === 'assistant'">
                  <div class="chat-markdown" v-html="renderMessageHtml(m)" @click="onBubbleClick"></div>
                  <div v-if="m.chart" class="chat-chart" :class="{ 'chat-chart--tall': m.chart.template === 'tactic_formation' }">
                    <Radar v-if="m.chart.template === 'player_radar'" :data="m.chart.data" :options="radarOptions" style="max-height: 280px" />
                    <Bar v-else-if="m.chart.template === 'team_bar'" :data="m.chart.data" :options="barOptions" style="max-height: 280px" />
                    <div v-else-if="m.chart.template === 'player_comparison_table'" class="chat-table-wrap">
                      <table class="chat-comparison-table">
                        <thead>
                          <tr>
                            <th></th>
                            <th v-for="name in m.chart.data.players" :key="name">{{ name }}</th>
                          </tr>
                        </thead>
                        <tbody v-for="cat in m.chart.data.categories" :key="cat.title">
                          <tr class="chat-comparison-table-category">
                            <td :colspan="m.chart.data.players.length + 1">{{ cat.title }}</td>
                          </tr>
                          <tr v-for="row in cat.rows" :key="row.label">
                            <td>{{ row.label }}</td>
                            <td v-for="(v, vi) in row.values" :key="vi">{{ v }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <div v-else-if="m.chart.template === 'team_comparison_table'" class="chat-table-wrap">
                      <table class="chat-comparison-table">
                        <thead>
                          <tr>
                            <th>Position</th>
                            <th>{{ m.chart.data.clubs[0] }}</th>
                            <th>{{ m.chart.data.clubs[1] }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="row in m.chart.data.rows" :key="row.position">
                            <td>{{ row.position }}</td>
                            <td>{{ row.aAvgOvr }} <span class="chat-table-depth">({{ row.aDepth }})</span></td>
                            <td>{{ row.bAvgOvr }} <span class="chat-table-depth">({{ row.bDepth }})</span></td>
                          </tr>
                          <tr class="chat-comparison-table-category">
                            <td>Best XI</td>
                            <td>{{ m.chart.data.bestXI[0] }}</td>
                            <td>{{ m.chart.data.bestXI[1] }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <div v-else-if="m.chart.template === 'tactic_formation'" class="chat-pitch-wrap">
                      <PitchDisplay
                        v-if="tacticFormationData(m.chart.data.formationKey)"
                        :formation="tacticFormationData(m.chart.data.formationKey).formation"
                        :players="tacticFormationData(m.chart.data.formationKey).players"
                        disable-player-clicks
                        :currency-symbol="detectedCurrencySymbol"
                      />
                      <div v-else class="chat-pitch-unavailable">Formation data unavailable.</div>
                    </div>
                  </div>
                </template>
                <template v-else>{{ m.text }}</template>
              </div>
            </div>

            <div v-if="chatStore.isStreaming" class="chat-row chat-row--assistant">
              <q-avatar size="26px" color="primary" text-color="white" icon="smart_toy" class="chat-avatar" />
              <div class="chat-bubble chat-bubble--status">
                <q-spinner-dots size="1rem" class="q-mr-xs" />{{ chatStore.statusLabel || 'Thinking…' }}
              </div>
            </div>
          </template>
        </div>

        <div class="chat-input-row">
          <div v-if="chatStore.isTurnLimitReached" class="chat-limit-notice">
            This conversation has reached its turn limit — start a New Chat to continue.
          </div>
          <template v-else>
            <div class="chat-chip-row" v-if="chatStore.messages.length">
              <q-chip v-for="q in presetQuestions" :key="q" dense clickable outline color="primary" size="sm" @click="ask(q)">{{ q }}</q-chip>
            </div>
            <div class="chat-input-inner">
              <q-input
                v-model="draft"
                dense
                outlined
                placeholder="Ask a question…"
                class="chat-input"
                :disable="chatStore.isStreaming"
                @keyup.enter="send"
              />
              <q-btn round color="primary" icon="send" :disable="!draft || chatStore.isStreaming" @click="send" />
            </div>
          </template>
        </div>
      </div>
    </transition>

    <q-btn round color="primary" size="lg" icon="smart_toy" class="chat-fab" @click="open = !open">
      <q-tooltip v-if="!open">Scout Assistant</q-tooltip>
    </q-btn>

    <q-dialog v-model="confirmNewChatOpen">
      <q-card class="chat-confirm-card">
        <q-card-section class="text-h6">Start a new chat?</q-card-section>
        <q-card-section class="q-pt-none">This conversation will be lost.</q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Cancel" @click="confirmNewChatOpen = false" />
          <q-btn flat color="negative" label="Start New Chat" @click="doNewChat" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <PlayerDetailDialog
      :player="playerForDetailView"
      :show="showPlayerDetailDialog"
      @close="showPlayerDetailDialog = false"
      :currency-symbol="detectedCurrencySymbol"
      :dataset-id="playerStore.currentDatasetId"
    />
  </div>
</template>

<script setup>
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  RadialLinearScale,
  Tooltip,
} from 'chart.js'
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import { Bar, Radar } from 'vue-chartjs'
import { useRouter } from 'vue-router'
import { computeSquadComposition } from '../composables/useBestXI'
import { useChatStore } from '../stores/chatStore'
import { usePlayerStore } from '../stores/playerStore'
import { renderChatMessageHtml } from '../utils/chatCitations'
import { getFormationLayout } from '../utils/formations'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PitchDisplay from './PitchDisplay.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PlayerDetailDialog from './PlayerDetailDialog.vue'

ChartJS.register(
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale
)

const PRESET_QUESTIONS = [
  'Who should I sign next?',
  'Find me the best wonderkids',
  'Find me some homegrown talent',
  'What tactic would fit this squad best?',
]

const chatStore = useChatStore()
const playerStore = usePlayerStore()
const router = useRouter()

// biome-ignore lint/correctness/noUnusedVariables: used in template
const open = ref(false)
const draft = ref('')
const confirmNewChatOpen = ref(false)
const scrollEl = ref(null)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const presetQuestions = PRESET_QUESTIONS

const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol || '$')

// player_radar plots raw FM attributes, which are on a 1-20 scale — not Overall
// (0-100-ish), which is what barOptions below is correctly scaled for.
// biome-ignore lint/correctness/noUnusedVariables: used in template
const radarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: { r: { beginAtZero: true, max: 20 } },
}
// biome-ignore lint/correctness/noUnusedVariables: used in template
const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: { y: { beginAtZero: true, max: 100 } },
}

// Managed team isn't a Pinia store (it's local state on DatasetPage.vue, set via
// ManagedTeamDialog) — this widget deliberately doesn't reach into that page's state
// (same "own independent lookup" precedent as ticket 05's PlayerDetailDialog decision).
// So a freshly-saved managed team wouldn't otherwise be noticed until something else
// re-triggers a check; poll briefly whenever a dataset is loaded but no managed team has
// been found yet, so the FAB appears right after the setup dialog is saved rather than
// only after a reload. Stops as soon as one is found.
let managedTeamPoll = null

function stopManagedTeamPoll() {
  if (managedTeamPoll) {
    clearInterval(managedTeamPoll)
    managedTeamPoll = null
  }
}

watch(
  () => playerStore.currentDatasetId,
  async (datasetId) => {
    stopManagedTeamPoll()
    await chatStore.checkManagedTeam(datasetId)
    if (datasetId && !chatStore.managedTeam) {
      managedTeamPoll = setInterval(async () => {
        await chatStore.checkManagedTeam(datasetId)
        if (chatStore.managedTeam) stopManagedTeamPoll()
      }, 3000)
    }
  },
  { immediate: true }
)

onUnmounted(stopManagedTeamPoll)

function findPlayerByUid(uid) {
  const target = String(uid)
  return (playerStore.allPlayers || []).find((p) => String(p.uid ?? p.UID) === target)
}

function openPlayer(uid) {
  const player = findPlayerByUid(uid)
  if (!player) return
  playerForDetailView.value = player
  showPlayerDetailDialog.value = true
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function renderMessageHtml(m) {
  return renderChatMessageHtml(m.text, m.referencedPlayers, (uid) => Boolean(findPlayerByUid(uid)))
}

// Computes the tactic_formation pitch data client-side from the live squad (ticket 03,
// .scratch/llm-refinements/issues/03-chat-tactics-formation-display.md) — the model
// only ever supplies a formationKey; computeSquadComposition (shared with
// TeamViewPage.vue's own Best XI) decides who actually plays where, so the two
// surfaces can never disagree about who best fits a given formation.
// biome-ignore lint/correctness/noUnusedVariables: used in template
function tacticFormationData(formationKey) {
  const formation = getFormationLayout(formationKey)
  if (!formation.length || !chatStore.managedTeam?.club) return null

  const squad = (playerStore.allPlayers || []).filter((p) => p.club === chatStore.managedTeam.club)
  if (squad.length === 0) return null

  const { squadComposition } = computeSquadComposition(squad, formation)
  const players = {}
  for (const row of formation) {
    for (const slot of row.positions) {
      const starter = squadComposition[slot.id]?.[0]
      players[slot.id] = starter
        ? {
            ...starter.player,
            Overall: starter.overallInRole,
            exactPositionMatch: starter.exactMatch,
          }
        : null
    }
  }
  return { formation, players }
}

// Player-link spans live inside v-html, so they can't carry a Vue @click binding —
// this delegated handler reads the uid off whichever chip was actually clicked.
// biome-ignore lint/correctness/noUnusedVariables: used in template
function onBubbleClick(event) {
  const uid = event.target?.dataset?.playerUid
  if (uid) openPlayer(uid)
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function ask(question) {
  draft.value = question
  send()
}

async function send() {
  const text = draft.value
  if (!text || chatStore.isStreaming || chatStore.isTurnLimitReached) return
  draft.value = ''
  await scrollToBottom()
  await chatStore.sendMessage(playerStore.currentDatasetId, text)
  await scrollToBottom()
  navigateIfRequested()
}

// Performs the route change a navigate_to_page tool call requested (ticket 04,
// .scratch/llm-refinements/issues/04-chat-query-rewrite-and-navigation.md). The
// drawer stays open across the navigation since the widget is mounted once at the
// app-shell level (App.vue) — no action needed here to keep it open.
function navigateIfRequested() {
  const last = chatStore.messages[chatStore.messages.length - 1]
  const nav = last?.navigate
  if (!nav) return
  const isDataset = nav.page === 'dataset'
  router.push({
    name: nav.page,
    params: isDataset ? { datasetId: playerStore.currentDatasetId } : {},
    query: isDataset && nav.search ? { search: nav.search } : {},
  })
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function doNewChat() {
  chatStore.newChat()
  confirmNewChatOpen.value = false
}

async function scrollToBottom() {
  await nextTick()
  if (scrollEl.value) scrollEl.value.scrollTop = scrollEl.value.scrollHeight
}

watch(
  () => chatStore.messages.length,
  () => scrollToBottom()
)
</script>

<style lang="scss" scoped>
.chat-widget-root {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 4000;
}
.chat-fab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  pointer-events: auto;
  box-shadow: var(--shadow-3);
}
.chat-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 440px;
  max-width: 92vw;
  background: var(--surface-card);
  box-shadow: var(--shadow-3);
  display: flex;
  flex-direction: column;
  pointer-events: auto;
}
.chat-drawer-enter-active,
.chat-drawer-leave-active {
  transition: transform 0.25s ease;
}
.chat-drawer-enter-from,
.chat-drawer-leave-to {
  transform: translateX(100%);
}

// Dialog chrome: unified header convention shared with PlayerDetailDialog /
// SettingsModal — icon, title, actions, close, all in normal flow.
.dialog-chrome {
  flex: 0 0 auto;
  background: var(--surface-raised);
  border-bottom: 1px solid var(--surface-border);
}
.dialog-chrome__header {
  display: flex;
  align-items: center;
  padding: 14px 16px;
}
.dialog-chrome__title-group {
  min-width: 0;
}
.dialog-chrome__title {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
}
.dialog-chrome__actions {
  display: flex;
  gap: 4px;
  align-items: center;
}
.dialog-chrome__close {
  transition: transform 0.15s ease;

  &:hover {
    transform: scale(1.08);
  }
}
.chat-title-sub {
  font-size: 0.72rem;
  color: var(--text-secondary);
}

.chat-body {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 16px;
}
.chat-empty {
  padding: 12px 4px;
}
.chat-empty-heading {
  font-size: 1.1rem;
  font-weight: 500;
  margin-bottom: 14px;
}
.chat-preset-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.chat-preset-card {
  cursor: pointer;
  transition: box-shadow 0.15s ease;
  background: var(--surface-card);
  border-color: var(--surface-border);

  &:hover {
    box-shadow: var(--shadow-2);
  }
}
.chat-preset-card-body {
  font-size: 0.82rem;
  padding: 14px;
}

.chat-row {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  align-items: flex-start;
}
.chat-row--user {
  justify-content: flex-end;
}
.chat-avatar {
  flex: 0 0 auto;
  margin-top: 2px;
}
.chat-bubble {
  max-width: 78%;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 0.86rem;
  line-height: 1.5;
}
.chat-bubble--user {
  background: var(--accent);
  color: var(--text-on-brand);
  white-space: pre-wrap;
}
.chat-bubble--assistant {
  background: var(--surface-raised);
  color: var(--text-primary);
}
.chat-bubble--error {
  // Error-tier color, kept semantic/hardcoded per established precedent.
  background: #fdecea;
  color: #a33;
  white-space: pre-wrap;
}
.chat-bubble--status {
  background: var(--surface-raised);
  color: var(--text-secondary);
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  white-space: pre-wrap;
}
.chat-markdown {
  :deep(p) {
    margin: 0 0 0.6em;

    &:last-child {
      margin-bottom: 0;
    }
  }
  :deep(h1),
  :deep(h2),
  :deep(h3),
  :deep(h4),
  :deep(h5),
  :deep(h6) {
    margin: 0.8em 0 0.4em;
    font-weight: 600;
    line-height: 1.3;

    &:first-child {
      margin-top: 0;
    }
  }
  :deep(h1) {
    font-size: 1.05em;
  }
  :deep(h2) {
    font-size: 1em;
  }
  :deep(h3),
  :deep(h4),
  :deep(h5),
  :deep(h6) {
    font-size: 0.95em;
  }
  :deep(ul),
  :deep(ol) {
    margin: 0 0 0.6em;
    padding-left: 1.3em;

    &:last-child {
      margin-bottom: 0;
    }
  }
  :deep(li) {
    margin-bottom: 0.2em;
  }
  :deep(code) {
    background: var(--surface-border-strong);
    border-radius: 4px;
    padding: 0.1em 0.3em;
    font-size: 0.85em;
  }
  :deep(pre) {
    background: var(--surface-border-strong);
    border-radius: 6px;
    padding: 0.6em 0.8em;
    overflow-x: auto;
    margin: 0 0 0.6em;

    code {
      background: none;
      padding: 0;
    }
  }
  :deep(strong) {
    font-weight: 700;
  }
  :deep(a) {
    color: var(--accent);
  }
  :deep(.chat-player-link) {
    color: var(--accent);
    font-weight: 600;
    text-decoration: underline;
    cursor: pointer;
  }
  :deep(.chat-player-link--unresolved) {
    color: inherit;
    font-weight: inherit;
    text-decoration: none;
    cursor: default;
  }
}
.chat-chart {
  margin-top: 10px;
  height: 280px;
}
.chat-chart--tall {
  height: auto;
}
.chat-pitch-wrap {
  max-height: 480px;
  overflow: auto;
  border-radius: 6px;
  transform: scale(0.85);
  transform-origin: top left;
  width: 117.6%;
}
.chat-pitch-unavailable {
  padding: 12px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}
.chat-table-wrap {
  height: 100%;
  overflow: auto;
  border: 1px solid var(--surface-border);
  border-radius: 6px;
}
.chat-comparison-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.78rem;
  white-space: nowrap;

  th,
  td {
    padding: 4px 8px;
    text-align: right;
    border-bottom: 1px solid var(--surface-border);
  }
  th:first-child,
  td:first-child {
    text-align: left;
  }
  thead th {
    position: sticky;
    top: 0;
    background: var(--surface-raised);
    font-weight: 600;
  }
  .chat-comparison-table-category td {
    font-weight: 600;
    background: var(--accent-soft);
  }
  .chat-table-depth {
    color: var(--text-secondary);
    font-size: 0.85em;
  }
}

.chat-input-row {
  padding: 12px 16px;
  border-top: 1px solid var(--surface-border);
  flex: 0 0 auto;
}
.chat-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}
.chat-input-inner {
  display: flex;
  gap: 8px;
}
.chat-input {
  flex: 1;
}
.chat-limit-notice {
  font-size: 0.82rem;
  color: var(--text-secondary);
  text-align: center;
  padding: 8px 0;
}
.chat-confirm-card {
  min-width: 320px;
}
</style>
