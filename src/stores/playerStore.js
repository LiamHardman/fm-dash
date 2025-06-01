import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'
import playerService from '../services/playerService'

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
    for (const p of allPlayers.value) {
      if (p.club) clubs.add(p.club)
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

  // Renamed for clarity: this reflects the range of the currently loaded dataset in the store
  const currentDatasetTransferValueRange = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 100000000 } // Default
    }
    const transferValuesNumeric = allPlayers.value
      .filter(p => typeof p.transferValueAmount === 'number')
      .map(p => p.transferValueAmount)

    if (transferValuesNumeric.length === 0) {
      return { min: 0, max: 100000000 }
    }

    let min = Math.min(...transferValuesNumeric)
    let max = Math.max(...transferValuesNumeric)

    min = Math.max(0, min) // Ensure min is not negative

    if (min >= max) {
      // Handles cases where all values are same, or only one value
      max = min + 50000 // Ensure max is greater for range slider
    }
    if (min === 0 && max === 50000 && transferValuesNumeric.every(v => v === 0)) {
      // If all values were 0, max was set to 50000. This is fine.
    }
    if (transferValuesNumeric.length === 1 && min === max) {
      // If only one player, ensure range
      max = min + 50000
    }

    const result = { min, max }
    return result
  })

  // This can be used by PlayerFilters as the true initial range of the whole dataset
  const initialDatasetTransferValueRange = computed(() => {
    // For now, this will be the same as currentDatasetTransferValueRange,
    // as allPlayers.value represents the full dataset fetched.
    // If server-side pagination/filtering were implemented, this might differ.
    return currentDatasetTransferValueRange.value
  })

  const salaryRange = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 1000000 } // Default
    }
    const salaryAmountsNumeric = allPlayers.value
      .filter(p => typeof p.wageAmount === 'number')
      .map(p => p.wageAmount)

    if (salaryAmountsNumeric.length === 0) {
      return { min: 0, max: 1000000 }
    }

    let min = Math.min(...salaryAmountsNumeric)
    let max = Math.max(...salaryAmountsNumeric)

    min = Math.max(0, min) // Ensure min is not negative

    if (min >= max) {
      // Handles cases where all values are same, or only one value
      max = min + 10000 // Ensure max is greater for range slider
    }
    if (salaryAmountsNumeric.length === 1 && min === max) {
      // If only one player, ensure range
      max = min + 10000
    }
    if (min === 0 && max === 10000 && salaryAmountsNumeric.every(v => v === 0)) {
      // If all values were 0, max was set to 10000. This is fine.
    }
    const result = { min, max }
    return result
  })

  async function uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    loading.value = true
    error.value = ''
    try {
      // Stage 1: Upload file (onProgress handles this)
      const response = await playerService.uploadPlayerFile(formData, maxSizeBytes, onProgress)
      currentDatasetId.value = response.datasetId
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '$'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      
      // Stage 2: Fetch processed data
      if (onProgress) onProgress(80)
      await fetchPlayersByDatasetId(currentDatasetId.value)
      
      // Stage 3: Fetch available roles
      if (onProgress) onProgress(95)
      await fetchAllAvailableRoles()
      
      // Stage 4: Complete
      if (onProgress) onProgress(100)
      
      return response
    } catch (e) {
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
    loading.value = true
    error.value = ''
    try {
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
      allPlayers.value = processPlayersFromAPI(response.players)
      currentDatasetId.value = datasetId
      detectedCurrencySymbol.value = response.currencySymbol || '$'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      return response
    } catch (e) {
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
    } catch (e) {
      console.error('playerStore: Error fetching available roles:', e)
      allAvailableRoles.value = []
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) return []
    return playersData.map(p => ({
      ...p,
      age: Number.parseInt(p.age, 10) || 0
    }))
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
    currentDatasetTransferValueRange, // Use this for current range
    initialDatasetTransferValueRange, // Use this for the full dataset's initial range
    salaryRange, // Added salaryRange
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
