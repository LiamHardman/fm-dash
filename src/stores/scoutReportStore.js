import { defineStore } from 'pinia'
import { ref } from 'vue'

// In-memory session cache for AI Scout Report — ticket 04 of the Scout Report map
// (.scratch/scout-report/issues/04-frontend-request-contract-and-session-cache.md).
// Keyed by `${datasetId}:${playerUid}:${position}`, never persisted (a plain Map held in
// a ref) — reopening the same player/position within the browser session is instant/free;
// a full page reload starts fresh. Not explicitly cleared on dataset switch: a stale entry
// under a different datasetId simply never gets hit again since the key includes it.
export const useScoutReportStore = defineStore('scoutReport', () => {
  const cache = ref(new Map())

  function key(datasetId, playerUid, position) {
    return `${datasetId}:${playerUid}:${position}`
  }

  function get(datasetId, playerUid, position) {
    return cache.value.get(key(datasetId, playerUid, position)) || null
  }

  function set(datasetId, playerUid, position, report) {
    cache.value.set(key(datasetId, playerUid, position), report)
  }

  return { get, set }
})
