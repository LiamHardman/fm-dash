import { computed, onUnmounted, ref, shallowRef, triggerRef, watchEffect } from 'vue'

/**
 * Enhanced memoization composable for large datasets
 * Provides lazy evaluation, cache management, and memory optimization
 */

export function useMemoization(options = {}) {
  const {
    maxCacheSize = 1000,
    ttl = 300000, // 5 minutes default TTL
    enableStats = true
  } = options

  // Cache storage
  const cache = new Map()
  const cacheTimestamps = new Map()
  const cacheStats = ref({
    hits: 0,
    misses: 0,
    evictions: 0,
    size: 0
  })

  /**
   * Create a memoized computed property with lazy evaluation
   */
  const memoizedComputed = (fn, keyFn = null, options = {}) => {
    const { lazy = true, dependencies = [], ttl: computedTtl = ttl } = options

    let cachedValue = null
    let isValid = false
    let lastDependencyValues = []

    const computedRef = computed(() => {
      // Check if dependencies have changed
      const currentDependencyValues = dependencies.map(dep =>
        typeof dep === 'function' ? dep() : dep.value
      )

      const dependenciesChanged =
        !isValid ||
        currentDependencyValues.length !== lastDependencyValues.length ||
        currentDependencyValues.some((val, index) => val !== lastDependencyValues[index])

      if (dependenciesChanged || !isValid) {
        const key = keyFn
          ? keyFn(...currentDependencyValues)
          : JSON.stringify(currentDependencyValues)

        // Check cache first
        if (cache.has(key)) {
          const timestamp = cacheTimestamps.get(key)
          if (Date.now() - timestamp < computedTtl) {
            cachedValue = cache.get(key)
            isValid = true
            lastDependencyValues = currentDependencyValues

            if (enableStats) {
              cacheStats.value.hits++
            }

            return cachedValue
          } else {
            // Cache expired
            cache.delete(key)
            cacheTimestamps.delete(key)
          }
        }

        // Compute new value
        cachedValue = fn(...currentDependencyValues)
        isValid = true
        lastDependencyValues = currentDependencyValues

        // Store in cache
        _setCache(key, cachedValue)

        if (enableStats) {
          cacheStats.value.misses++
        }
      }

      return cachedValue
    })

    return computedRef
  }

  /**
   * Create a lazy computed property that only evaluates when accessed
   */
  const lazyComputed = (fn, dependencies = []) => {
    let cachedValue = null
    let isComputed = false
    let lastDependencyValues = []

    return computed(() => {
      const currentDependencyValues = dependencies.map(dep =>
        typeof dep === 'function' ? dep() : dep.value
      )

      const dependenciesChanged =
        currentDependencyValues.length !== lastDependencyValues.length ||
        currentDependencyValues.some((val, index) => val !== lastDependencyValues[index])

      if (!isComputed || dependenciesChanged) {
        cachedValue = fn()
        isComputed = true
        lastDependencyValues = currentDependencyValues
      }

      return cachedValue
    })
  }

  /**
   * Create a memoized function with automatic cache management
   */
  const memoize = (fn, keyFnOrOptions = null, options = {}) => {
    // Handle different parameter formats for backward compatibility
    let keyFn = null
    let actualOptions = options

    if (typeof keyFnOrOptions === 'function') {
      // Old format: memoize(fn, keyFn, options)
      keyFn = keyFnOrOptions
    } else if (typeof keyFnOrOptions === 'object' && keyFnOrOptions !== null) {
      // New format: memoize(fn, options) where options may contain keyGenerator
      actualOptions = keyFnOrOptions
      keyFn = actualOptions.keyGenerator || null
    }

    const { ttl: fnTtl = ttl, maxSize = maxCacheSize } = actualOptions

    const memoizedFn = (...args) => {
      const key = keyFn ? keyFn(...args) : JSON.stringify(args)

      // Check cache
      if (cache.has(key)) {
        const timestamp = cacheTimestamps.get(key)
        if (Date.now() - timestamp < fnTtl) {
          if (enableStats) {
            cacheStats.value.hits++
          }
          return cache.get(key)
        } else {
          // Cache expired
          cache.delete(key)
          cacheTimestamps.delete(key)
        }
      }

      // Compute new value
      const result = fn(...args)

      // Store in cache
      _setCache(key, result)

      if (enableStats) {
        cacheStats.value.misses++
      }

      return result
    }

    // Add cache management methods to the memoized function
    memoizedFn.clearCache = () => {
      // Clear only entries created by this specific memoized function
      // Since we're using a shared cache, we can't easily isolate entries
      // For now, we'll clear the entire cache
      const clearedCount = cache.size
      cache.clear()
      cacheTimestamps.clear()
      if (enableStats) {
        cacheStats.value.size = 0
      }
      return clearedCount
    }

    memoizedFn.getStats = () => {
      return {
        cacheSize: cache.size,
        hits: cacheStats.value.hits,
        misses: cacheStats.value.misses
      }
    }

    return memoizedFn
  }

  /**
   * Create a debounced memoized function
   */
  const debouncedMemoize = (fn, delay = 300, keyFn = null) => {
    const timeouts = new Map()

    return (...args) => {
      const key = keyFn ? keyFn(...args) : JSON.stringify(args)

      // Clear existing timeout
      if (timeouts.has(key)) {
        clearTimeout(timeouts.get(key))
      }

      return new Promise(resolve => {
        const timeout = setTimeout(() => {
          // Check cache first
          if (cache.has(key)) {
            const timestamp = cacheTimestamps.get(key)
            if (Date.now() - timestamp < ttl) {
              if (enableStats) {
                cacheStats.value.hits++
              }
              resolve(cache.get(key))
              return
            }
          }

          // Compute new value
          const result = fn(...args)
          _setCache(key, result)

          if (enableStats) {
            cacheStats.value.misses++
          }

          resolve(result)
          timeouts.delete(key)
        }, delay)

        timeouts.set(key, timeout)
      })
    }
  }

  /**
   * Create a batch processor with memoization
   */
  const batchMemoize = (fn, batchSize = 100, keyFn = null) => {
    const batches = new Map()

    return items => {
      const key = keyFn ? keyFn(items) : `batch_${items.length}_${Date.now()}`

      // Check if we can use cached batch results
      if (cache.has(key)) {
        const timestamp = cacheTimestamps.get(key)
        if (Date.now() - timestamp < ttl) {
          if (enableStats) {
            cacheStats.value.hits++
          }
          return cache.get(key)
        }
      }

      // Process in batches
      const results = []
      for (let i = 0; i < items.length; i += batchSize) {
        const batch = items.slice(i, i + batchSize)
        const batchResult = fn(batch)
        results.push(...(Array.isArray(batchResult) ? batchResult : [batchResult]))
      }

      _setCache(key, results)

      if (enableStats) {
        cacheStats.value.misses++
      }

      return results
    }
  }

  /**
   * Set cache with LRU eviction
   */
  const _setCache = (key, value) => {
    // Evict old entries if cache is full
    if (cache.size >= maxCacheSize) {
      _evictLRU()
    }

    cache.set(key, value)
    cacheTimestamps.set(key, Date.now())

    if (enableStats) {
      cacheStats.value.size = cache.size
    }
  }

  /**
   * Evict least recently used entries
   */
  const _evictLRU = () => {
    const entries = Array.from(cacheTimestamps.entries())
    entries.sort((a, b) => a[1] - b[1]) // Sort by timestamp

    const toEvict = Math.ceil(maxCacheSize * 0.2) // Evict 20%

    for (let i = 0; i < toEvict && i < entries.length; i++) {
      const [key] = entries[i]
      cache.delete(key)
      cacheTimestamps.delete(key)

      if (enableStats) {
        cacheStats.value.evictions++
      }
    }
  }

  /**
   * Clear expired entries
   */
  const clearExpired = () => {
    const now = Date.now()
    const expiredKeys = []

    for (const [key, timestamp] of cacheTimestamps) {
      if (now - timestamp > ttl) {
        expiredKeys.push(key)
      }
    }

    for (const key of expiredKeys) {
      cache.delete(key)
      cacheTimestamps.delete(key)
    }

    if (enableStats) {
      cacheStats.value.size = cache.size
    }

    return expiredKeys.length
  }

  /**
   * Clear all cache
   */
  const clearCache = () => {
    const size = cache.size
    cache.clear()
    cacheTimestamps.clear()

    if (enableStats) {
      cacheStats.value.size = 0
    }

    return size
  }

  /**
   * Get cache statistics
   */
  const getStats = () => {
    return {
      ...cacheStats.value,
      hitRate: cacheStats.value.hits / (cacheStats.value.hits + cacheStats.value.misses) || 0,
      memoryUsage: _estimateMemoryUsage()
    }
  }

  /**
   * Estimate memory usage (rough calculation)
   */
  const _estimateMemoryUsage = () => {
    let size = 0
    for (const [key, value] of cache) {
      size += JSON.stringify(key).length + JSON.stringify(value).length
    }
    return size
  }

  // Periodic cleanup
  const cleanupInterval = setInterval(clearExpired, ttl)

  onUnmounted(() => {
    clearInterval(cleanupInterval)
    clearCache()
  })

  return {
    // Core functions
    memoizedComputed,
    lazyComputed,
    memoize,
    debouncedMemoize,
    batchMemoize,

    // Cache management
    clearCache,
    clearExpired,
    getStats,

    // Stats
    cacheStats: cacheStats.value
  }
}

/**
 * Specialized memoization for large arrays
 */
export function useArrayMemoization(options = {}) {
  const { chunkSize = 1000, enableVirtualization = true } = options

  const memoization = useMemoization(options)

  /**
   * Memoized array processing with chunking
   */
  const processArray = memoization.memoize(
    (array, processFn, keyFn = null) => {
      if (!Array.isArray(array)) return []

      const key = keyFn ? keyFn(array) : `array_${array.length}_${Date.now()}`

      if (array.length <= chunkSize) {
        return processFn(array)
      }

      // Process in chunks
      const results = []
      for (let i = 0; i < array.length; i += chunkSize) {
        const chunk = array.slice(i, i + chunkSize)
        const chunkResult = processFn(chunk)
        results.push(...(Array.isArray(chunkResult) ? chunkResult : [chunkResult]))
      }

      return results
    },
    (array, processFn) => `processArray_${array.length}_${processFn.name || 'anonymous'}`
  )

  /**
   * Memoized array filtering
   */
  const filterArray = memoization.memoize(
    (array, filterFn) => {
      return processArray(array, chunk => chunk.filter(filterFn))
    },
    (array, filterFn) => `filter_${array.length}_${filterFn.toString().slice(0, 50)}`
  )

  /**
   * Memoized array mapping
   */
  const mapArray = memoization.memoize(
    (array, mapFn) => {
      return processArray(array, chunk => chunk.map(mapFn))
    },
    (array, mapFn) => `map_${array.length}_${mapFn.toString().slice(0, 50)}`
  )

  /**
   * Memoized array sorting
   */
  const sortArray = memoization.memoize(
    (array, compareFn) => {
      // For large arrays, use a more efficient sorting approach
      if (array.length > 10000) {
        return [...array].sort(compareFn)
      }
      return array.slice().sort(compareFn)
    },
    (array, compareFn) => `sort_${array.length}_${compareFn.toString().slice(0, 50)}`
  )

  return {
    ...memoization,
    processArray,
    filterArray,
    mapArray,
    sortArray
  }
}

// Create a lazy global memoization instance to avoid circular dependencies
let globalMemoizationInstance = null

const getGlobalMemoization = () => {
  if (!globalMemoizationInstance) {
    try {
      globalMemoizationInstance = useMemoization({
        maxCacheSize: 2000,
        ttl: 600000, // 10 minutes
        enableStats: true
      })
    } catch (error) {
      console.warn('Failed to create global memoization:', error)
      // Fallback implementation
      globalMemoizationInstance = {
        memoize: (fn, keyFnOrOptions = null, options = {}) => {
          return (...args) => fn(...args)
        }
      }
    }
  }
  return globalMemoizationInstance
}

export const memoize = (fn, keyFnOrOptions = null, options = {}) => {
  return getGlobalMemoization().memoize(fn, keyFnOrOptions, options)
}

export default useMemoization
