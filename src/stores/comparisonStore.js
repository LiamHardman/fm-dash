import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

const MAX_COMPARISON_PLAYERS = 4
const STORAGE_KEY = 'fm_dash_comparison'

export const useComparisonStore = defineStore('comparison', () => {
  const comparisonByDataset = ref(loadFromStorage())

  function loadFromStorage() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      return raw ? JSON.parse(raw) : {}
    } catch {
      return {}
    }
  }

  function persist() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(comparisonByDataset.value))
    } catch {}
  }

  const getPlayersForDataset = (datasetId) => {
    if (!datasetId) return []
    return comparisonByDataset.value[datasetId] || []
  }

  const getCount = (datasetId) => getPlayersForDataset(datasetId).length

  const isInComparison = (datasetId, player) => {
    if (!datasetId || !player) return false
    return getPlayersForDataset(datasetId).some(
      (p) => p.name === player.name && p.club === player.club
    )
  }

  const addToComparison = (datasetId, player) => {
    if (!datasetId || !player) return false
    if (!comparisonByDataset.value[datasetId]) {
      comparisonByDataset.value[datasetId] = []
    }
    const list = comparisonByDataset.value[datasetId]
    if (list.length >= MAX_COMPARISON_PLAYERS) return false
    if (list.some((p) => p.name === player.name && p.club === player.club)) return false
    list.push(player)
    persist()
    return true
  }

  const removeFromComparison = (datasetId, player) => {
    if (!datasetId || !player || !comparisonByDataset.value[datasetId]) return
    const list = comparisonByDataset.value[datasetId]
    const idx = list.findIndex((p) => p.name === player.name && p.club === player.club)
    if (idx !== -1) {
      list.splice(idx, 1)
      persist()
    }
  }

  const clearComparison = (datasetId) => {
    if (datasetId) {
      comparisonByDataset.value[datasetId] = []
      persist()
    }
  }

  // Convenience computed for components that pass datasetId reactively
  const makeComputed = (datasetId) => computed(() => getPlayersForDataset(datasetId))

  return {
    comparisonByDataset,
    getPlayersForDataset,
    getCount,
    isInComparison,
    addToComparison,
    removeFromComparison,
    clearComparison,
    makeComputed,
    MAX_COMPARISON_PLAYERS,
  }
})
