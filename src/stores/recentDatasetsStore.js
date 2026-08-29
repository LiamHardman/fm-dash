import { defineStore } from 'pinia'
import { ref } from 'vue'

// Small localStorage-backed history of previously-uploaded saves, letting the
// dashboard home (ticket 03 of the UI redesign map) offer "jump back into a
// past save" -- something the app has never tracked before (playerStore only
// keeps a single currentDatasetId in sessionStorage, cleared per tab).
const STORAGE_KEY = 'recentDatasets'
const MAX_ENTRIES = 8

export const useRecentDatasetsStore = defineStore('recentDatasets', () => {
  const entries = ref([]) // [{ datasetId, label, playerCount, uploadedAt }]

  function load() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      entries.value = stored ? JSON.parse(stored) : []
    } catch (_e) {
      entries.value = []
    }
  }

  function persist() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(entries.value))
    } catch (_e) {}
  }

  function recordDataset({ datasetId, label, playerCount }) {
    if (!datasetId) return
    const withoutExisting = entries.value.filter((e) => e.datasetId !== datasetId)
    withoutExisting.unshift({
      datasetId,
      label: label || 'FM Save',
      playerCount: playerCount || 0,
      uploadedAt: Date.now(),
    })
    entries.value = withoutExisting.slice(0, MAX_ENTRIES)
    persist()
  }

  function removeDataset(datasetId) {
    entries.value = entries.value.filter((e) => e.datasetId !== datasetId)
    persist()
  }

  function renameDataset(datasetId, label) {
    const entry = entries.value.find((item) => item.datasetId === datasetId)
    if (!entry || !label?.trim()) return
    entry.label = label.trim()
    persist()
  }

  return { entries, load, recordDataset, removeDataset, renameDataset }
})
