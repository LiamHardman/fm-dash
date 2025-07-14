import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'
import playerService from '../services/playerService.js'
import { PerformanceTracker } from '../utils/performance.js'

export const usePlayerStore = defineStore('player', () => {
  const allPlayers = shallowRef([])
  const currentDatasetId = ref(sessionStorage.getItem('currentDatasetId') || null)
  const detectedCurrencySymbol = ref(sessionStorage.getItem('detectedCurrencySymbol') || '$')
  const loading = ref(false)
  const error = ref('')
  const allAvailableRoles = ref([])

  const AGE_SLIDER_MIN_DEFAULT = 15
  const AGE_SLIDER_MAX_DEFAULT = 50

  const uniqueClubs = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []
    const clubs = new Set()
    for (const player of allPlayers.value) {
      if (player.club) clubs.add(player.club)
    }
    return Array.from(clubs).sort()
  })

  const uniqueNationalities = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []
    const nationalities = new Set()
    for (const p of allPlayers.value) {
      if (p.nationality) nationalities.add(p.nationality)
    }
    return Array.from(nationalities).sort()
  })

  const uniqueMediaHandlings = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []
    const mediaHandlingsIndividual = new Set()
    for (const p of allPlayers.value) {
      if (p.media_handling) {
        for (const style of p.media_handling.split(',')) {
          const trimmedStyle = style.trim()
          if (trimmedStyle) mediaHandlingsIndividual.add(trimmedStyle)
        }
      }
    }
    return Array.from(mediaHandlingsIndividual).sort()
  })

  const uniquePersonalities = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []
    const personalities = new Set()
    for (const p of allPlayers.value) {
      if (p.personality) personalities.add(p.personality)
    }
    return Array.from(personalities).sort()
  })

  const uniquePositionsCount = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return 0
    const s = new Set()
    for (const player of allPlayers.value) {
      if (player.parsedPositions) {
        for (const pos of player.parsedPositions) {
          s.add(pos)
        }
      }
    }
    return s.size
  })

  const currentDatasetTransferValueRange = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 100000000 } // Default
    }

    let min = Number.MAX_SAFE_INTEGER
    let max = Number.MIN_SAFE_INTEGER
    let hasValidValue = false
    for (const player of allPlayers.value) {
      if (typeof player.transferValueAmount === 'number') {
        hasValidValue = true
        if (player.transferValueAmount < min) min = player.transferValueAmount
        if (player.transferValueAmount > max) max = player.transferValueAmount
      }
    }

    if (!hasValidValue) {
      return { min: 0, max: 100000000 }
    }

    min = Math.max(0, min) // Ensure min is not negative

    if (min >= max) {
      // Handles cases where all values are same, or only one value
      max = min + 50000 // Ensure max is greater for range slider
    }

    const result = { min, max }
    return result
  })

  const initialDatasetTransferValueRange = computed(() => {
    return currentDatasetTransferValueRange.value
  })

  const salaryRange = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 1000000 } // Default
    }

    let min = Number.MAX_SAFE_INTEGER
    let max = Number.MIN_SAFE_INTEGER
    let hasValidValue = false

    for (const player of allPlayers.value) {
      if (typeof player.wageAmount === 'number') {
        hasValidValue = true
        if (player.wageAmount < min) min = player.wageAmount
        if (player.wageAmount > max) max = player.wageAmount
      }
    }

    if (!hasValidValue) {
      return { min: 0, max: 1000000 }
    }

    min = Math.max(0, min) // Ensure min is not negative

    if (min >= max) {
      // Handles cases where all values are same, or only one value
      max = min + 10000 // Ensure max is greater for range slider
    }

    return { min, max }
  })

  async function uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    const tracker = new PerformanceTracker('Upload Process')
    loading.value = true
    error.value = ''
    try {
      // Stage 1: Upload file (onProgress handles this)
      tracker.checkpoint('Starting upload')
      const response = await playerService.uploadPlayerFile(formData, maxSizeBytes, onProgress)
      tracker.checkpoint('Upload completed')

      currentDatasetId.value = response.datasetId
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '$'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      tracker.checkpoint('Session storage updated')

      // Stage 2 & 3: Fetch processed data and available roles in parallel
      if (onProgress) onProgress(80)

      // Run both API calls in parallel to reduce total time
      const [_playerDataResponse] = await Promise.all([
        fetchPlayersByDatasetId(currentDatasetId.value),
        fetchAllAvailableRoles() // This can run in parallel since it doesn't depend on player data
      ])
      tracker.checkpoint('Data fetching completed')

      // Stage 4: Complete
      if (onProgress) onProgress(100)
      tracker.finish()

      return response
    } catch (e) {
      tracker.checkpoint('Error occurred')
      if (e.status === 413 && e.message) {
        error.value = e.message
      } else {
        error.value = `Failed to process file: ${e.message || 'Unknown error'}`
      }
      resetState()
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchPlayersByDatasetId(
    datasetId,
    positionFilter = null,
    roleFilter = null,
    ageRangeFilter = null,
    transferValueRangeFilter = null,
    maxSalaryFilter = null,
    divisionFilter = 'all',
    targetDivision = null,
    positionCompare = 'all'
  ) {
    if (!datasetId) {
      resetState()
      return
    }
    const tracker = new PerformanceTracker('Fetch Players Data')
    loading.value = true
    error.value = ''
    try {
      tracker.checkpoint('Starting API call')
      const response = await playerService.getPlayersByDatasetId(
        datasetId,
        positionFilter,
        roleFilter,
        ageRangeFilter,
        transferValueRangeFilter,
        maxSalaryFilter,
        divisionFilter,
        targetDivision,
        positionCompare
      )
      tracker.checkpoint('API call completed')

      allPlayers.value = processPlayersFromAPI(response.players)
      tracker.checkpoint('Players processed')

      currentDatasetId.value = datasetId
      detectedCurrencySymbol.value = response.currencySymbol || '$'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      tracker.checkpoint('State updated')
      tracker.finish()

      return response
    } catch (e) {
      tracker.checkpoint('Error occurred')
      error.value = `Failed to fetch player data: ${e.message || 'Unknown error'}`
      resetState()
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchAllAvailableRoles(force = false) {
    if (allAvailableRoles.value.length > 0 && !force) return
    try {
      const roles = await playerService.getAvailableRoles()
      allAvailableRoles.value = roles.sort()
    } catch (_e) {
      allAvailableRoles.value = []
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) {
      return []
    }

    const processed = playersData.map(p => ({
      ...p,
      age: Number.parseInt(p.age, 10) || 0
    }))

    return processed
  }

  function resetState() {
    allPlayers.value = []
    currentDatasetId.value = null
    detectedCurrencySymbol.value = '$'
    allAvailableRoles.value = []
    sessionStorage.removeItem('currentDatasetId')
    sessionStorage.removeItem('detectedCurrencySymbol')
  }

  async function loadFromSessionStorage() {
    const storedDatasetId = sessionStorage.getItem('currentDatasetId')
    const storedCurrencySymbol = sessionStorage.getItem('detectedCurrencySymbol')
    if (storedDatasetId) {
      currentDatasetId.value = storedDatasetId
      if (storedCurrencySymbol) {
        detectedCurrencySymbol.value = storedCurrencySymbol
      }
      try {
        await fetchPlayersByDatasetId(storedDatasetId)
        await fetchAllAvailableRoles()
      } catch (_e) {
        // Error handled in fetch functions
      }
    } else {
      resetState()
    }
  }

  return {
    allPlayers,
    currentDatasetId,
    detectedCurrencySymbol,
    loading,
    error,
    uniqueClubs,
    uniqueNationalities,
    uniqueMediaHandlings,
    uniquePersonalities,
    uniquePositionsCount,
    currentDatasetTransferValueRange,
    initialDatasetTransferValueRange,
    salaryRange,
    allAvailableRoles,
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    fetchAllAvailableRoles,
    resetState,
    loadFromSessionStorage,
    AGE_SLIDER_MIN_DEFAULT,
    AGE_SLIDER_MAX_DEFAULT
  }
})
