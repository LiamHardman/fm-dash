import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'
import playerService from '../services/playerService.js'
import { PerformanceTracker } from '../utils/performance.js'

export const usePlayerStore = defineStore('player', () => {
  const allPlayers = shallowRef([])
  const currentDatasetId = ref(sessionStorage.getItem('currentDatasetId') || null)
  const detectedCurrencySymbol = ref(sessionStorage.getItem('detectedCurrencySymbol') || '£')
  const loading = ref(false)
  const error = ref('')
  const allAvailableRoles = ref([])
  const protobufMetrics = ref({
    enabled: false,
    compressionRatio: 0,
    requestCount: 0,
    averagePayloadSize: 0
  })

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
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '£'
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

      // Update protobuf metrics if available
      if (response._protobuf) {
        updateProtobufMetrics(response._protobuf)
      }

      // Extract players from response (handle both protobuf and JSON formats)
      const players = response.players || []
      allPlayers.value = processPlayersFromAPI(players)
      tracker.checkpoint('Players processed')

      currentDatasetId.value = datasetId
      detectedCurrencySymbol.value = response.currencySymbol || '£'
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
      const response = await playerService.getAvailableRoles()
      
      // Update protobuf metrics if available
      if (response._protobuf) {
        updateProtobufMetrics(response._protobuf)
      }
      
      // Extract roles from response (handle both protobuf and JSON formats)
      const roles = response.roles || response
      allAvailableRoles.value = Array.isArray(roles) ? roles.sort() : []
    } catch (_e) {
      allAvailableRoles.value = []
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) {
      return []
    }

    const processed = playersData.map(p => {
      // Ensure all required fields are present and properly formatted
      const processedPlayer = {
        ...p,
        // Ensure age is a number
        age: Number.parseInt(p.age, 10) || 0,
        
        // Ensure arrays are properly initialized
        short_positions: Array.isArray(p.short_positions) ? p.short_positions : [],
        parsedPositions: Array.isArray(p.parsedPositions) ? p.parsedPositions : [],
        positionGroups: Array.isArray(p.positionGroups) ? p.positionGroups : [],
        
        // Ensure role-specific overalls are properly formatted
        roleSpecificOveralls: Array.isArray(p.roleSpecificOveralls) ? p.roleSpecificOveralls : [],
        
        // Ensure FIFA-style stats are numbers
        PAC: Number.parseInt(p.PAC, 10) || 0,
        SHO: Number.parseInt(p.SHO, 10) || 0,
        PAS: Number.parseInt(p.PAS, 10) || 0,
        DRI: Number.parseInt(p.DRI, 10) || 0,
        DEF: Number.parseInt(p.DEF, 10) || 0,
        PHY: Number.parseInt(p.PHY, 10) || 0,
        
        // Ensure goalkeeper stats are numbers
        GK: Number.parseInt(p.GK, 10) || 0,
        DIV: Number.parseInt(p.DIV, 10) || 0,
        HAN: Number.parseInt(p.HAN, 10) || 0,
        REF: Number.parseInt(p.REF, 10) || 0,
        KIC: Number.parseInt(p.KIC, 10) || 0,
        SPD: Number.parseInt(p.SPD, 10) || 0,
        POS: Number.parseInt(p.POS, 10) || 0,
        
        // Ensure overall is a number
        Overall: Number.parseInt(p.Overall, 10) || 0,
        
        // Ensure numeric attributes are properly formatted
        numericAttributes: p.numericAttributes || {},
        
        // Ensure performance stats are properly formatted
        performanceStatsNumeric: p.performanceStatsNumeric || {},
        performancePercentiles: p.performancePercentiles || {}
      }
      
      return processedPlayer
    })

    return processed
  }

  function resetState() {
    allPlayers.value = []
    currentDatasetId.value = null
    detectedCurrencySymbol.value = '£'
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
  
  /**
   * Update protobuf metrics based on API response
   * @param {Object} protobufInfo - Protobuf metadata from response
   */
  function updateProtobufMetrics(protobufInfo) {
    protobufMetrics.value.enabled = protobufInfo.format === 'protobuf'
    protobufMetrics.value.requestCount++
    
    if (protobufInfo.payloadSize) {
      // Update average payload size
      protobufMetrics.value.averagePayloadSize = 
        (protobufMetrics.value.averagePayloadSize * (protobufMetrics.value.requestCount - 1) + 
         protobufInfo.payloadSize) / protobufMetrics.value.requestCount
    }
    
    if (protobufInfo.compressionRatio) {
      protobufMetrics.value.compressionRatio = protobufInfo.compressionRatio
    }
  }
  
  /**
   * Toggle protobuf support
   * @param {boolean} enabled - Whether protobuf should be enabled
   */
  function setProtobufEnabled(enabled) {
    playerService.setProtobufEnabled(enabled)
  }
  
  /**
   * Get protobuf client status
   */
  function getProtobufStatus() {
    return {
      ...playerService.getClientStatus(),
      metrics: { ...protobufMetrics.value }
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
    protobufMetrics,
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    fetchAllAvailableRoles,
    resetState,
    loadFromSessionStorage,
    setProtobufEnabled,
    getProtobufStatus,
    AGE_SLIDER_MIN_DEFAULT,
    AGE_SLIDER_MAX_DEFAULT
  }
})