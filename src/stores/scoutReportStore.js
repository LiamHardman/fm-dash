import { defineStore } from 'pinia'
import { ref } from 'vue'

// In-memory session cache for AI Scout Report. Originally the source of truth (ticket 04
// of the original Scout Report map), now demoted to a pure request-dedup layer in front
// of the backend (Scout Report v2 map ticket 04, .scratch/scout-report-v2/issues/
// 04-player-dialog-load-regenerate-states.md): reports are backend-persisted, so a fresh
// mount or reload always trusts a GET /api/scout-report/{datasetId} read rather than this
// cache alone — a report generated via chat in a different session/tab is picked up
// correctly instead of being shadowed by a stale local entry. This cache only avoids
// redundant network round-trips on rapid position toggling within one browser session.
// Keyed by `${datasetId}:${playerUid}:${position}`, never persisted (a plain Map held in
// a ref) — a full page reload starts fresh. Not explicitly cleared on dataset switch: a
// stale entry under a different datasetId simply never gets hit again since the key
// includes it.
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
