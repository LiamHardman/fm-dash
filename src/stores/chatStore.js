import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { llmRequestHeaders } from '../utils/llmHeaders'
import { consumeSSEStream } from '../utils/sseClient'
import { useUiStore } from './uiStore'

// FM-Dash Chatbot conversation state. In-memory only (ticket 01 / map decision) — no
// persistence beyond this Pinia store's lifetime, cleared by page reload or "New Chat".
// Mounted once at app-shell level (App.vue), so this store is the single source of
// truth the SSE client resends as history on every turn (ticket 01: stateless backend).

const CHATBOT_MAX_TURNS = 50

export const useChatStore = defineStore('chat', () => {
  const uiStore = useUiStore()

  const messages = ref([]) // { role: 'user'|'assistant'|'error', text, referencedPlayers?, chart? }
  const statusLabel = ref('')
  const isStreaming = ref(false)
  const managedTeam = ref(null) // { club, division } | null
  const managedTeamChecked = ref(false)

  // conversationId is telemetry-only (tracing map ticket 04) — sent as a header so the
  // backend can group this chat's requests together in observability spans/metrics
  // (messages/searches per chat). Never read by any backend behavior/routing logic, so
  // regenerating it here on newChat() is safe with no effect on conversation state.
  const conversationId = ref(crypto.randomUUID())

  const userTurnCount = computed(() => messages.value.filter((m) => m.role === 'user').length)
  const isTurnLimitReached = computed(() => userTurnCount.value >= CHATBOT_MAX_TURNS)

  async function checkManagedTeam(datasetId) {
    if (!datasetId) {
      managedTeam.value = null
      managedTeamChecked.value = true
      return
    }
    try {
      const res = await fetch(`/api/managed-team/${datasetId}`)
      managedTeam.value = res.ok ? await res.json() : null
    } catch (_e) {
      managedTeam.value = null
    } finally {
      managedTeamChecked.value = true
    }
  }

  function newChat() {
    messages.value = []
    statusLabel.value = ''
    conversationId.value = crypto.randomUUID()
  }

  const sseHandlers = {
    status: (payload) => {
      statusLabel.value = payload.label || ''
    },
    done: (payload) => {
      statusLabel.value = ''
      messages.value.push({
        role: 'assistant',
        text: payload.text,
        referencedPlayers: payload.referencedPlayers || [],
        chart: payload.chart || null,
        navigate: payload.navigate || null,
      })
    },
    error: (payload) => {
      statusLabel.value = ''
      messages.value.push({ role: 'error', text: payload.message || 'Something went wrong.' })
    },
  }

  async function sendMessage(datasetId, text) {
    const trimmed = (text || '').trim()
    if (!trimmed || isStreaming.value || isTurnLimitReached.value || !datasetId) return

    // The failed-turn recovery rule (ticket 10): the user's own message is pushed here
    // and stays regardless of outcome. Only a successful turn appends an assistant
    // entry; a failure appends an 'error' entry instead. Since the history sent to the
    // backend is rebuilt from user/assistant messages only (see below), a failed turn
    // never becomes part of what's resent — nothing to roll back, conversation stays
    // intact, and retyping the same message is a normal retry.
    messages.value.push({ role: 'user', text: trimmed })
    isStreaming.value = true
    statusLabel.value = 'Thinking…'

    const history = messages.value
      .filter((m) => m.role === 'user' || m.role === 'assistant')
      .map((m) => ({ role: m.role, text: m.text }))

    try {
      const res = await fetch(`/api/chatbot/${datasetId}`, {
        method: 'POST',
        headers: {
          ...llmRequestHeaders(uiStore),
          'X-Chat-Conversation-Id': conversationId.value,
        },
        body: JSON.stringify({ history }),
      })

      if (!res.ok || !res.body) {
        let message = `Unexpected error (${res.status}).`
        try {
          const body = await res.json()
          if (body?.error) message = body.error
        } catch (_e) {
          // non-JSON error body — fall back to the generic message above
        }
        messages.value.push({ role: 'error', text: message })
        return
      }

      await consumeSSEStream(res, sseHandlers)
    } catch (_e) {
      messages.value.push({
        role: 'error',
        text: 'Connection to the assistant was lost — try sending that again.',
      })
    } finally {
      isStreaming.value = false
      statusLabel.value = ''
    }
  }

  return {
    messages,
    statusLabel,
    isStreaming,
    managedTeam,
    managedTeamChecked,
    isTurnLimitReached,
    conversationId,
    checkManagedTeam,
    newChat,
    sendMessage,
  }
})
