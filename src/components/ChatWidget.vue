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
        <div class="chat-header">
          <div class="chat-header-title">
            <q-avatar size="32px" color="primary" text-color="white" icon="smart_toy" class="q-mr-sm" />
            <div>
              <div class="chat-title-main">Scout Assistant</div>
              <div class="chat-title-sub">{{ chatStore.statusLabel || 'Ask about your squad, targets, tactics' }}</div>
            </div>
          </div>
          <div class="chat-header-actions">
            <q-btn flat dense no-caps label="New Chat" icon="add_comment" :disable="chatStore.messages.length === 0" @click="confirmNewChatOpen = true" />
            <q-btn flat dense round icon="close" @click="open = false" />
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
                  <span v-for="(part, pi) in parseChatCitations(m.text, m.referencedPlayers)" :key="pi">
                    <span v-if="part.type === 'text'">{{ part.content }}</span>
                    <span
                      v-else
                      class="chat-player-link"
                      :class="{ 'chat-player-link--unresolved': !findPlayerByUid(part.uid) }"
                      @click="openPlayer(part.uid)"
                    >{{ part.name }}</span>
                  </span>
                  <div v-if="m.chart" class="chat-chart">
                    <Radar v-if="m.chart.template === 'player_radar'" :data="m.chart.data" :options="radarOptions" style="max-height: 280px" />
                    <Bar v-else-if="m.chart.template === 'team_bar'" :data="m.chart.data" :options="barOptions" style="max-height: 280px" />
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
import { Bar, Radar } from 'vue-chartjs'
import { useChatStore } from '../stores/chatStore'
import { usePlayerStore } from '../stores/playerStore'
import { parseChatCitations } from '../utils/chatCitations'
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

const open = ref(false)
const draft = ref('')
const confirmNewChatOpen = ref(false)
const scrollEl = ref(null)
const presetQuestions = PRESET_QUESTIONS

const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol || '$')

// player_radar plots raw FM attributes, which are on a 1-20 scale — not Overall
// (0-100-ish), which is what barOptions below is correctly scaled for.
const radarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: { r: { beginAtZero: true, max: 20 } },
}
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
  { immediate: true },
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
}

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
  box-shadow: 0 4px 18px rgba(26, 35, 126, 0.35);
}
.chat-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 440px;
  max-width: 92vw;
  background: white;
  box-shadow: -6px 0 40px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  pointer-events: auto;

  .body--dark & {
    background: #16161f;
  }
}
.chat-drawer-enter-active,
.chat-drawer-leave-active {
  transition: transform 0.25s ease;
}
.chat-drawer-enter-from,
.chat-drawer-leave-to {
  transform: translateX(100%);
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
  flex: 0 0 auto;
}
.chat-header-title {
  display: flex;
  align-items: center;
}
.chat-title-main {
  font-weight: 600;
  font-size: 0.95rem;
}
.chat-title-sub {
  font-size: 0.72rem;
  color: #888;
}
.chat-header-actions {
  display: flex;
  gap: 4px;
  align-items: center;
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

  &:hover {
    box-shadow: 0 2px 10px rgba(26, 35, 126, 0.15);
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
  white-space: pre-wrap;
}
.chat-bubble--user {
  background: #1a237e;
  color: white;
}
.chat-bubble--assistant {
  background: #f3f4fa;
  color: #222;

  .body--dark & {
    background: #24242f;
    color: #eee;
  }
}
.chat-bubble--error {
  background: #fdecea;
  color: #a33;
}
.chat-bubble--status {
  background: #f3f4fa;
  color: #888;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
}
.chat-player-link {
  color: #1a237e;
  font-weight: 600;
  text-decoration: underline;
  cursor: pointer;

  .body--dark & {
    color: #7c8fff;
  }
}
.chat-player-link--unresolved {
  color: inherit;
  font-weight: inherit;
  text-decoration: none;
  cursor: default;
}
.chat-chart {
  margin-top: 10px;
  height: 280px;
}

.chat-input-row {
  padding: 12px 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
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
  color: #888;
  text-align: center;
  padding: 8px 0;
}
.chat-confirm-card {
  min-width: 320px;
}
</style>
