import { defineStore } from 'pinia'
import { computed, ref, shallowRef, triggerRef } from 'vue'
import { useArrayMemoization, useMemoization } from '../composables/useMemoization.js'
import playerService from '../services/playerService.js'
import IndexedMap from '../utils/IndexedMap.js'
import { PerformanceTracker } from '../utils/performance.js'

export const useOptimizedPlayerStore = defineStore('optimizedPlayer', () => {
  // Use shallowRef for large datasets to avoid deep reactivity overhead
  const allPlayers = shallowRef([])
  const playersIndexedMap = shallowRef(null)
  const currentDatasetId = ref(sessionStorage.getItem('currentDatasetId') || null)
  const detectedCurrencySymbol = ref(sessionStorage.getItem('detectedCurrencySymbol') || '€')
  const loading = ref(false)
  const error = ref('')
  const allAvailableRoles = ref([])

  // Initialize memoization for expensive computations with error handling
  let memoization = null
  let arrayMemoization = null

  const initializeMemoization = () => {
    try {
      if (!memoization) {
        memoization = useMemoization({
          maxCacheSize: 500,
          ttl: 600000, // 10 minutes
          enableStats: true
        })
      }

      if (!arrayMemoization) {
        arrayMemoization = useArrayMemoization({
          chunkSize: 2000,
          enableVirtualization: true
        })
      }
    } catch (error) {
      console.warn('Failed to initialize memoization:', error)
      // Fallback to basic computed properties
      memoization = {
        lazyComputed: (fn, deps) => computed(fn),
        memoizedComputed: (fn, keyFn, options) => computed(fn)
      }
      arrayMemoization = {
        processArray: (arr, fn) => [fn(arr)]
      }
    }
  }

  // Initialize memoization immediately
  initializeMemoization()

  // Constants
  const AGE_SLIDER_MIN_DEFAULT = 15
  const AGE_SLIDER_MAX_DEFAULT = 50

  // Initialize IndexedMap for O(1) lookups
  const initializeIndexedMap = players => {
    if (!Array.isArray(players) || players.length === 0) {
      playersIndexedMap.value = null
      return
    }

    const indexedMap = new IndexedMap({
      primaryKey: 'name',
      indexes: ['club', 'nationality', 'position', 'personality', 'media_handling']
    })

    // Add unique ID if not present and bulk insert
    const playersWithId = players.map((player, index) => ({
      ...player,
      _id: player._id || `player_${index}_${player.name?.replace(/\s+/g, '_') || index}`
    }))

    const bulkResult = indexedMap.bulkSet(playersWithId)
    console.log(
      `IndexedMap initialized: ${bulkResult.inserted} players, ${bulkResult.itemsPerSecond} items/sec`
    )

    playersIndexedMap.value = indexedMap
    triggerRef(playersIndexedMap) // Manually trigger reactivity
  }

  // Optimized computed properties with lazy evaluation and memoization
  const uniqueClubs = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (playersIndexedMap.value) {
      return playersIndexedMap.value.getUniqueValues('club').sort()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          const clubs = new Set()
          for (const player of chunk) {
            if (player.club) clubs.add(player.club)
          }
          return Array.from(clubs)
        })
        .flat()
        .sort()
    }

    // Fallback implementation
    const clubs = new Set()
    for (const player of allPlayers.value) {
      if (player.club) clubs.add(player.club)
    }
    return Array.from(clubs).sort()
  })

  const uniqueNationalities = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (playersIndexedMap.value) {
      return playersIndexedMap.value.getUniqueValues('nationality').sort()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          const nationalities = new Set()
          for (const player of chunk) {
            if (player.nationality) nationalities.add(player.nationality)
          }
          return Array.from(nationalities)
        })
        .flat()
        .sort()
    }

    // Fallback implementation
    const nationalities = new Set()
    for (const player of allPlayers.value) {
      if (player.nationality) nationalities.add(player.nationality)
    }
    return Array.from(nationalities).sort()
  })

  const uniqueMediaHandlings = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          const mediaHandlings = new Set()
          for (const player of chunk) {
            if (player.media_handling) {
              for (const style of player.media_handling.split(',')) {
                const trimmedStyle = style.trim()
                if (trimmedStyle) mediaHandlings.add(trimmedStyle)
              }
            }
          }
          return Array.from(mediaHandlings)
        })
        .flat()
        .sort()
    }

    // Fallback implementation
    const mediaHandlings = new Set()
    for (const player of allPlayers.value) {
      if (player.media_handling) {
        for (const style of player.media_handling.split(',')) {
          const trimmedStyle = style.trim()
          if (trimmedStyle) mediaHandlings.add(trimmedStyle)
        }
      }
    }
    return Array.from(mediaHandlings).sort()
  })

  const uniquePersonalities = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (playersIndexedMap.value) {
      return playersIndexedMap.value.getUniqueValues('personality').sort()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return []

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          const personalities = new Set()
          for (const player of chunk) {
            if (player.personality) personalities.add(player.personality)
          }
          return Array.from(personalities)
        })
        .flat()
        .sort()
    }

    // Fallback implementation
    const personalities = new Set()
    for (const player of allPlayers.value) {
      if (player.personality) personalities.add(player.personality)
    }
    return Array.from(personalities).sort()
  })

  const uniquePositionsCount = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return 0

    const positions = new Set()
    for (const player of allPlayers.value) {
      if (player.parsedPositions) {
        for (const pos of player.parsedPositions) {
          positions.add(pos)
        }
      }
    }
    return positions.size
  })

  // Optimized transfer value range calculation with memoization
  const currentDatasetTransferValueRange = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 100000000 }
    }

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          let min = Number.MAX_SAFE_INTEGER
          let max = Number.MIN_SAFE_INTEGER
          let hasValidValue = false

          for (const player of chunk) {
            if (typeof player.transferValueAmount === 'number') {
              hasValidValue = true
              if (player.transferValueAmount < min) min = player.transferValueAmount
              if (player.transferValueAmount > max) max = player.transferValueAmount
            }
          }

          return hasValidValue ? { min, max, hasValidValue } : { hasValidValue: false }
        })
        .reduce(
          (acc, chunkResult) => {
            if (!chunkResult.hasValidValue) return acc

            return {
              min: Math.min(acc.min, chunkResult.min),
              max: Math.max(acc.max, chunkResult.max),
              hasValidValue: true
            }
          },
          { min: Number.MAX_SAFE_INTEGER, max: Number.MIN_SAFE_INTEGER, hasValidValue: false }
        )
    }

    // Fallback implementation
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

    return { min, max }
  })

  const initialDatasetTransferValueRange = computed(() => {
    return currentDatasetTransferValueRange.value
  })

  // Optimized salary range calculation
  const salaryRange = computed(() => {
    // Ensure memoization is initialized
    if (!memoization) {
      initializeMemoization()
    }

    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 1000000 }
    }

    if (arrayMemoization) {
      return arrayMemoization
        .processArray(allPlayers.value, chunk => {
          let min = Number.MAX_SAFE_INTEGER
          let max = Number.MIN_SAFE_INTEGER
          let hasValidValue = false

          for (const player of chunk) {
            if (typeof player.wageAmount === 'number') {
              hasValidValue = true
              if (player.wageAmount < min) min = player.wageAmount
              if (player.wageAmount > max) max = player.wageAmount
            }
          }

          return hasValidValue ? { min, max, hasValidValue } : { hasValidValue: false }
        })
        .reduce(
          (acc, chunkResult) => {
            if (!chunkResult.hasValidValue) return acc

            return {
              min: Math.min(acc.min, chunkResult.min),
              max: Math.max(acc.max, chunkResult.max),
              hasValidValue: true
            }
          },
          { min: Number.MAX_SAFE_INTEGER, max: Number.MIN_SAFE_INTEGER, hasValidValue: false }
        )
    }

    // Fallback implementation
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

  // Efficient player lookup functions using IndexedMap
  const findPlayersByClub = club => {
    if (playersIndexedMap.value) {
      return playersIndexedMap.value.findByIndex('club', club)
    }
    return allPlayers.value.filter(player => player.club === club)
  }

  const findPlayersByNationality = nationality => {
    if (playersIndexedMap.value) {
      return playersIndexedMap.value.findByIndex('nationality', nationality)
    }
    return allPlayers.value.filter(player => player.nationality === nationality)
  }

  const findPlayersByPosition = position => {
    if (playersIndexedMap.value) {
      return playersIndexedMap.value.findByIndex('position', position)
    }
    return allPlayers.value.filter(player => player.position?.includes(position))
  }

  const findPlayersWhere = criteria => {
    if (playersIndexedMap.value) {
      return playersIndexedMap.value.findWhere(criteria)
    }
    return allPlayers.value.filter(player => {
      return Object.entries(criteria).every(([key, value]) => player[key] === value)
    })
  }

  // Optimized data processing functions
  async function uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    const tracker = new PerformanceTracker('Upload Process')
    loading.value = true
    error.value = ''

    try {
      tracker.checkpoint('Starting upload')
      const response = await playerService.uploadPlayerFile(formData, maxSizeBytes, onProgress)
      tracker.checkpoint('Upload completed')

      currentDatasetId.value = response.datasetId
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '€'
      sessionStorage.setItem('currentDatasetId', currentDatasetId.value)
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value)
      tracker.checkpoint('Session storage updated')

      if (onProgress) onProgress(80)

      const [_playerDataResponse] = await Promise.all([
        fetchPlayersByDatasetId(currentDatasetId.value),
        fetchAllAvailableRoles()
      ])
      tracker.checkpoint('Data fetching completed')

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

      const processedPlayers = processPlayersFromAPI(response.players)
      allPlayers.value = processedPlayers
      triggerRef(allPlayers) // Manually trigger reactivity for shallowRef

      // Initialize IndexedMap for efficient lookups
      initializeIndexedMap(processedPlayers)
      tracker.checkpoint('Players processed and indexed')

      currentDatasetId.value = datasetId
      detectedCurrencySymbol.value = response.currencySymbol || '€'
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

    // Use batch processing for large datasets
    return arrayMemoization
      .processArray(playersData, chunk =>
        chunk.map(p => ({
          ...p,
          age: Number.parseInt(p.age, 10) || 0
        }))
      )
      .flat()
  }

  function resetState() {
    allPlayers.value = []
    playersIndexedMap.value = null
    currentDatasetId.value = null
    detectedCurrencySymbol.value = '€'
    allAvailableRoles.value = []
    sessionStorage.removeItem('currentDatasetId')
    sessionStorage.removeItem('detectedCurrencySymbol')

    // Clear memoization cache
    memoization.clearCache()
    arrayMemoization.clearCache()

    // Trigger reactivity
    triggerRef(allPlayers)
    triggerRef(playersIndexedMap)
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

  // Performance monitoring
  const getPerformanceStats = () => {
    return {
      memoization: memoization.getStats(),
      arrayMemoization: arrayMemoization.getStats(),
      indexedMap: playersIndexedMap.value?.getStats() || null,
      playersCount: allPlayers.value.length
    }
  }

  // Memory optimization
  const optimizeMemory = () => {
    // Clear expired cache entries
    const expiredMemo = memoization.clearExpired()
    const expiredArray = arrayMemoization.clearExpired()

    // Optimize IndexedMap
    let optimizedIndexes = 0
    if (playersIndexedMap.value) {
      const result = playersIndexedMap.value.optimize()
      optimizedIndexes = result.removedEntries
    }

    return {
      expiredMemoEntries: expiredMemo,
      expiredArrayEntries: expiredArray,
      optimizedIndexes
    }
  }

  return {
    // Data
    allPlayers,
    playersIndexedMap,
    currentDatasetId,
    detectedCurrencySymbol,
    loading,
    error,
    allAvailableRoles,

    // Computed properties
    uniqueClubs,
    uniqueNationalities,
    uniqueMediaHandlings,
    uniquePersonalities,
    uniquePositionsCount,
    currentDatasetTransferValueRange,
    initialDatasetTransferValueRange,
    salaryRange,

    // Lookup functions
    findPlayersByClub,
    findPlayersByNationality,
    findPlayersByPosition,
    findPlayersWhere,

    // Actions
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    fetchAllAvailableRoles,
    resetState,
    loadFromSessionStorage,

    // Performance
    getPerformanceStats,
    optimizeMemory,

    // Constants
    AGE_SLIDER_MIN_DEFAULT,
    AGE_SLIDER_MAX_DEFAULT
  }
})

export default useOptimizedPlayerStore
