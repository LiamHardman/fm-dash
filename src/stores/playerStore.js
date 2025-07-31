import { defineStore } from 'pinia'
import { computed, ref, shallowRef, watch } from 'vue'
import playerService from '../services/playerService.js'
import { PerformanceTracker } from '../utils/performance.js'
import uploadFlowProfiler from '../utils/performanceProfiler.js'

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
    averagePayloadSize: 0,
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

  const uniqueDivisions = ref([])

  // Update uniqueDivisions when allPlayers changes
  watch(
    allPlayers,
    (newPlayers) => {
      if (!Array.isArray(newPlayers) || newPlayers.length === 0) {
        uniqueDivisions.value = []
        return
      }
      const divisions = new Set()
      for (const p of newPlayers) {
        if (p.division && p.division.trim() !== '') {
          divisions.add(p.division)
        }
      }
      uniqueDivisions.value = Array.from(divisions).sort()
    },
    { immediate: true }
  )

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
      // Handle both number and string wage amounts
      let wageAmount = player.wageAmount
      if (typeof wageAmount === 'string') {
        wageAmount = Number.parseInt(wageAmount, 10)
      }

      if (typeof wageAmount === 'number' && !Number.isNaN(wageAmount) && wageAmount > 0) {
        hasValidValue = true
        if (wageAmount < min) min = wageAmount
        if (wageAmount > max) max = wageAmount
      }
    }

    // Debug: Log some sample wage amounts
    if (allPlayers.value.length > 0) {
      const samplePlayers = allPlayers.value.slice(0, 5)
      console.log(
        'Sample wage amounts:',
        samplePlayers.map((p) => ({
          name: p.name,
          wageAmount: p.wageAmount,
          type: typeof p.wageAmount,
        }))
      )
    }

    if (!hasValidValue) {
      console.log('No valid wage amounts found in player data')
      return { min: 0, max: 1000000 }
    }

    min = Math.max(0, min) // Ensure min is not negative

    if (min >= max) {
      // Handles cases where all values are same, or only one value
      // Use a more reasonable range based on the salary value
      if (min <= 10000) {
        max = min + 5000 // For low salaries, add 5K
      } else if (min <= 100000) {
        max = min + 25000 // For medium salaries, add 25K
      } else if (min <= 1000000) {
        max = min + 100000 // For high salaries, add 100K
      } else {
        max = min + 500000 // For very high salaries, add 500K
      }
    }

    // Debug logging to help identify salary range issues
    if (min === max && min > 0) {
      console.log(
        `Salary range computation: All players have same salary ${min}, setting max to ${max}`
      )
    }

    return { min, max }
  })

  async function uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    // Start profiling the upload flow
    uploadFlowProfiler.startUploadFlow()

    const tracker = new PerformanceTracker('Upload Process')
    loading.value = true
    error.value = ''

    // Clean up old cache entries before upload
    uploadFlowProfiler.markStep('cache_cleanup_start')
    cleanupOldCache()
    uploadFlowProfiler.markStep('cache_cleanup_end')

    try {
      // Stage 1: Upload file (onProgress handles this)
      uploadFlowProfiler.markStep('file_upload_start')
      tracker.checkpoint('Starting upload')
      const response = await playerService.uploadPlayerFile(formData, maxSizeBytes, onProgress)
      uploadFlowProfiler.markStep('file_upload_end', {
        playerCount: response.playerCount,
        hasPlayers: !!response.players,
        hasRoles: !!response.roles,
      })
      tracker.checkpoint('Upload completed')

      uploadFlowProfiler.markStep('session_storage_start')
      currentDatasetId.value = response.datasetId
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '£'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      uploadFlowProfiler.markStep('session_storage_end')
      tracker.checkpoint('Session storage updated')

      // OPTIMIZATION: Use data from enhanced upload response if available
      if (response.players && response.roles) {
        uploadFlowProfiler.markStep('data_processing_start')

        // Store data immediately from upload response
        allPlayers.value = processPlayersFromAPI(response.players)
        allAvailableRoles.value = response.roles

        uploadFlowProfiler.markStep('data_processing_end', {
          playerCount: response.players.length,
          roleCount: response.roles.length,
        })

        // Cache the processed data for future use (with size limit)
        uploadFlowProfiler.markStep('cache_processing_start')
        const cacheKey = `dataset_${response.datasetId}`

        let cacheSize = 0
        let cacheType = 'none'
        let isLargeDataset = false

        try {
          // For large datasets, use lightweight cache
          isLargeDataset = response.players.length > 100
          const cacheData = isLargeDataset
            ? createLightweightCache(
                response.players,
                response.roles,
                response.detectedCurrencySymbol
              )
            : {
                players: response.players,
                roles: response.roles,
                currencySymbol: response.detectedCurrencySymbol,
                timestamp: Date.now(),
              }

          const cacheString = JSON.stringify(cacheData)
          cacheSize = Math.round(cacheString.length / 1024)
          cacheType = isLargeDataset ? 'lightweight' : 'full'

          // Check if cache would exceed reasonable size (1MB limit)
          if (cacheString.length < 1024 * 1024) {
            sessionStorage.setItem(cacheKey, cacheString)
            console.log('Data cached successfully', {
              datasetId: response.datasetId,
              cacheSize: `${cacheSize}KB`,
              cacheType: cacheType,
              playerCount: response.players.length,
            })
          } else {
            console.warn('Cache size too large, skipping cache', {
              datasetId: response.datasetId,
              cacheSize: `${cacheSize}KB`,
              playerCount: response.players.length,
            })
            cacheType = 'skipped'
          }
        } catch (error) {
          console.warn('Failed to cache data, continuing without cache', {
            datasetId: response.datasetId,
            error: error.message,
          })
          cacheType = 'failed'
        }

        uploadFlowProfiler.markStep('cache_processing_end', {
          cacheSize: `${cacheSize}KB`,
          cacheType: cacheType,
        })

        tracker.checkpoint('Data stored from upload response')

        // Stage 4: Complete (skip redundant API calls)
        if (onProgress) onProgress(100)
        tracker.finish()

        uploadFlowProfiler.endUploadFlow()
        return response
      }

      // Fallback to original flow for backward compatibility
      uploadFlowProfiler.markStep('fallback_api_calls_start')
      tracker.checkpoint('Using fallback data fetching')

      // Stage 2 & 3: Fetch processed data and available roles in parallel
      if (onProgress) onProgress(80)

      // Run both API calls in parallel to reduce total time
      const [_playerDataResponse] = await Promise.all([
        fetchPlayersByDatasetId(currentDatasetId.value),
        fetchAllAvailableRoles(), // This can run in parallel since it doesn't depend on player data
      ])
      uploadFlowProfiler.markStep('fallback_api_calls_end')
      tracker.checkpoint('Data fetching completed')

      // Stage 4: Complete
      if (onProgress) onProgress(100)
      tracker.finish()

      uploadFlowProfiler.endUploadFlow()
      return response
    } catch (e) {
      uploadFlowProfiler.markStep('error_handling', { error: e.message })
      tracker.checkpoint('Error occurred')
      if (e.status === 413 && e.message) {
        error.value = e.message
      } else {
        error.value = `Failed to process file: ${e.message || 'Unknown error'}`
      }
      resetState()
      uploadFlowProfiler.endUploadFlow()
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

    uploadFlowProfiler.markStep('fetch_players_start', { datasetId })
    const tracker = new PerformanceTracker('Fetch Players Data')
    loading.value = true
    error.value = ''
    try {
      tracker.checkpoint('Starting API call')
      uploadFlowProfiler.markStep('api_call_start')
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
      uploadFlowProfiler.markStep('api_call_end', {
        playerCount: response.players?.length || 0,
        hasProtobuf: !!response._protobuf,
      })
      tracker.checkpoint('API call completed')

      // Update protobuf metrics if available
      if (response._protobuf) {
        updateProtobufMetrics(response._protobuf)
      }

      // Extract players from response (handle both protobuf and JSON formats)
      uploadFlowProfiler.markStep('player_processing_start')
      const players = response.players || []
      allPlayers.value = processPlayersFromAPI(players)
      uploadFlowProfiler.markStep('player_processing_end', {
        processedPlayerCount: allPlayers.value.length,
      })
      tracker.checkpoint('Players processed')

      uploadFlowProfiler.markStep('state_update_start')
      currentDatasetId.value = datasetId
      detectedCurrencySymbol.value = response.currencySymbol || '£'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      uploadFlowProfiler.markStep('state_update_end')
      tracker.checkpoint('State updated')
      tracker.finish()

      return response
    } catch (e) {
      uploadFlowProfiler.markStep('fetch_players_error', { error: e.message })
      tracker.checkpoint('Error occurred')

      // Handle 404 errors specially - data might not be ready yet for large files
      if (e.message?.includes('404')) {
        console.log(
          'Dataset not ready yet (404), this is expected for large files being processed in background'
        )
        error.value =
          'Dataset is still being processed in the background. Please wait a moment and try again.'
      } else {
        error.value = `Failed to fetch player data: ${e.message || 'Unknown error'}`
      }

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

    const startTime = performance.now()
    const processed = playersData.map((p) => {
      // Ensure all required fields are present and properly formatted
      const processedPlayer = {
        ...p,
        // Ensure age is a number
        age: Number.parseInt(p.age, 10) || 0,

        // Ensure wage amount is a number
        wageAmount:
          typeof p.wageAmount === 'string'
            ? Number.parseInt(p.wageAmount, 10) || 0
            : p.wageAmount || 0,

        // Ensure arrays are properly initialized
        shortPositions: Array.isArray(p.shortPositions) ? p.shortPositions : [],
        short_positions: Array.isArray(p.shortPositions) ? p.shortPositions : [], // Keep both for compatibility
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

        // Ensure overall is a number (handle both field names)
        Overall: Number.parseInt(p.Overall || p.overall, 10) || 0,

        // Ensure numeric attributes are properly formatted
        numericAttributes: p.numericAttributes || {},

        // Ensure performance stats are properly formatted
        performanceStatsNumeric: p.performanceStatsNumeric || {},
        performancePercentiles: p.performancePercentiles || {},
      }

      return processedPlayer
    })

    const processingTime = performance.now() - startTime
    if (processingTime > 50) {
      // Log if processing takes more than 50ms
      console.log(
        `Player processing took ${processingTime.toFixed(2)}ms for ${playersData.length} players`
      )
    }

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
    // Clean up old cache entries before loading
    cleanupOldCache()

    const storedDatasetId = sessionStorage.getItem('currentDatasetId')
    const storedCurrencySymbol = sessionStorage.getItem('detectedCurrencySymbol')

    if (storedDatasetId) {
      currentDatasetId.value = storedDatasetId
      if (storedCurrencySymbol) {
        detectedCurrencySymbol.value = storedCurrencySymbol
      }

      // OPTIMIZATION: Try to load from cache first
      const cacheKey = `dataset_${storedDatasetId}`
      const cachedData = sessionStorage.getItem(cacheKey)

      if (cachedData) {
        try {
          const parsed = JSON.parse(cachedData)
          const cacheAge = Date.now() - parsed.timestamp
          const maxCacheAge = 30 * 60 * 1000 // 30 minutes

          if (cacheAge < maxCacheAge && validateCacheIntegrity(parsed, storedDatasetId)) {
            // Use cached data
            if (parsed.players) {
              // Full cache available
              allPlayers.value = processPlayersFromAPI(parsed.players)
              allAvailableRoles.value = parsed.roles
              detectedCurrencySymbol.value = parsed.currencySymbol || storedCurrencySymbol || '£'

              console.log('Loaded full data from cache', {
                datasetId: storedDatasetId,
                playerCount: parsed.players.length,
                roleCount: parsed.roles.length,
                cacheAgeMinutes: Math.round(cacheAge / 60000),
              })

              return // Skip API calls if cache is valid
            }
            if (parsed.playerCount && parsed.roles) {
              // Lightweight cache available - still need to fetch full data
              allAvailableRoles.value = parsed.roles
              detectedCurrencySymbol.value = parsed.currencySymbol || storedCurrencySymbol || '£'

              console.log('Lightweight cache found, fetching full data', {
                datasetId: storedDatasetId,
                expectedPlayerCount: parsed.playerCount,
                roleCount: parsed.roles.length,
                cacheAgeMinutes: Math.round(cacheAge / 60000),
              })

              // Still need to fetch full player data, but we have roles
              try {
                await fetchPlayersByDatasetId(storedDatasetId)
                return
              } catch (_e) {
                // Error handled in fetch functions
              }
            }
          }
        } catch (error) {
          console.warn('Failed to load from cache, falling back to API', error)
        }
      }

      // Fallback to API calls
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

  // Helper function to create a lightweight cache for large datasets
  function createLightweightCache(players, roles, currencySymbol) {
    // For large datasets, only cache essential metadata
    const essentialData = {
      playerCount: players.length,
      roles: roles,
      currencySymbol: currencySymbol,
      timestamp: Date.now(),
      // Only cache first few players as sample for validation
      samplePlayers: players.slice(0, 5).map((p) => ({
        uid: p.uid,
        name: p.name,
        position: p.position,
        overall: p.Overall,
      })),
    }

    return essentialData
  }

  // Helper function to validate cache integrity
  function validateCacheIntegrity(cachedData, datasetId) {
    if (!cachedData || !cachedData.playerCount || !cachedData.roles) {
      return false
    }

    // Check if we have sample players for validation
    if (cachedData.samplePlayers && cachedData.samplePlayers.length > 0) {
      console.log('Cache validation passed', {
        datasetId,
        playerCount: cachedData.playerCount,
        roleCount: cachedData.roles.length,
        samplePlayers: cachedData.samplePlayers.length,
      })
      return true
    }

    return false
  }

  // Helper function to clean up old cache entries
  function cleanupOldCache() {
    try {
      const keys = Object.keys(sessionStorage)
      const cacheKeys = keys.filter((key) => key.startsWith('dataset_'))
      const now = Date.now()
      const maxAge = 30 * 60 * 1000 // 30 minutes

      let cleanedCount = 0
      for (const key of cacheKeys) {
        try {
          const data = sessionStorage.getItem(key)
          if (data) {
            const parsed = JSON.parse(data)
            if (parsed.timestamp && now - parsed.timestamp > maxAge) {
              sessionStorage.removeItem(key)
              cleanedCount++
            }
          }
        } catch (_error) {
          // Remove invalid cache entries
          sessionStorage.removeItem(key)
          cleanedCount++
        }
      }

      if (cleanedCount > 0) {
        console.log('Cleaned up old cache entries', { cleanedCount })
      }
    } catch (error) {
      console.warn('Failed to cleanup cache', error)
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
          protobufInfo.payloadSize) /
        protobufMetrics.value.requestCount
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
      metrics: { ...protobufMetrics.value },
    }
  }

  // New methods for setting data directly
  function setPlayers(players) {
    allPlayers.value = processPlayersFromAPI(players)
  }

  function setCurrencySymbol(symbol) {
    detectedCurrencySymbol.value = symbol
    sessionStorage.setItem('detectedCurrencySymbol', symbol)
  }

  function setCurrentDatasetId(datasetId) {
    currentDatasetId.value = datasetId
    sessionStorage.setItem('currentDatasetId', datasetId)
  }

  return {
    allPlayers,
    currentDatasetId,
    detectedCurrencySymbol,
    loading,
    error,
    uniqueClubs,
    uniqueNationalities,
    uniqueDivisions,
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
    setPlayers,
    setCurrencySymbol,
    setCurrentDatasetId,
    AGE_SLIDER_MIN_DEFAULT,
    AGE_SLIDER_MAX_DEFAULT,
  }
})
