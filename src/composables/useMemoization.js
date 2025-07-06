import { computed, isRef, ref, shallowRef } from 'vue'

// Cache for memoized functions
const memoCache = new Map()

/**
 * Creates a memoized version of a function with configurable cache
 * @param {Function} fn - Function to memoize
 * @param {Object} options - Configuration options
 * @returns {Function} Memoized function
 */
export function memoize(fn, options = {}) {
  const {
    maxSize = 100,
    keyGenerator = (...args) => JSON.stringify(args),
    ttl = null, // Time to live in milliseconds
    cacheKey = null
  } = options

  const cache = cacheKey ? memoCache.get(cacheKey) || new Map() : new Map()

  if (cacheKey && !memoCache.has(cacheKey)) {
    memoCache.set(cacheKey, cache)
  }

  const memoizedFn = (...args) => {
    const key = keyGenerator(...args)

    if (cache.has(key)) {
      const cached = cache.get(key)

      // Check TTL if specified
      if (ttl && Date.now() - cached.timestamp > ttl) {
        cache.delete(key)
      } else {
        return cached.value
      }
    }

    // Compute new result
    const result = fn(...args)

    // Store in cache
    const cacheEntry = {
      value: result,
      timestamp: Date.now()
    }

    cache.set(key, cacheEntry)

    // Implement LRU eviction if cache is too large
    if (cache.size > maxSize) {
      const firstKey = cache.keys().next().value
      cache.delete(firstKey)
    }

    return result
  }

  // Add method to clear cache
  memoizedFn.clearCache = () => cache.clear()
  memoizedFn.cache = cache

  return memoizedFn
}

/**
 * Creates a memoized computed property with dependency tracking
 * @param {Function} getter - Computed getter function
 * @param {Object} options - Configuration options
 * @returns {ComputedRef} Memoized computed ref
 */
export function memoizedComputed(getter, options = {}) {
  const { equality = (a, b) => a === b, dependencies = [], debug = false } = options

  const cache = ref(null)
  const hasCachedValue = ref(false)
  const lastDepsValues = ref([])

  return computed(() => {
    // Track dependencies if provided
    const currentDepsValues = dependencies.map(dep => (isRef(dep) ? dep.value : dep))

    // Check if dependencies have changed
    const depsChanged =
      !hasCachedValue.value ||
      currentDepsValues.length !== lastDepsValues.value.length ||
      currentDepsValues.some((val, i) => !equality(val, lastDepsValues.value[i]))

    if (depsChanged || !hasCachedValue.value) {
      if (debug) {
      }

      cache.value = getter()
      hasCachedValue.value = true
      lastDepsValues.value = currentDepsValues
    }

    return cache.value
  })
}

/**
 * Creates a reactive memoized function that works with Vue reactivity
 * @param {Function} fn - Function to memoize
 * @param {Object} options - Configuration options
 * @returns {Function} Reactive memoized function
 */
export function reactiveMemoize(fn, options = {}) {
  const { keyGenerator = (...args) => JSON.stringify(args) } = options

  const cache = shallowRef(new Map())

  return (...args) => {
    const key = keyGenerator(...args)
    const currentCache = cache.value

    if (currentCache.has(key)) {
      return currentCache.get(key)
    }

    const result = fn(...args)

    // Create new map to trigger reactivity
    const newCache = new Map(currentCache)
    newCache.set(key, result)
    cache.value = newCache

    return result
  }
}

/**
 * Creates a memoized getter for object properties
 * @param {Function} getter - Property getter function
 * @param {Object} options - Configuration options
 * @returns {Function} Memoized property getter
 */
export function memoizedGetter(getter, options = {}) {
  const memoizedFn = memoize(getter, options)

  return (obj, prop) => {
    return memoizedFn(obj, prop)
  }
}

/**
 * Performance monitoring for memoized functions
 */
export class MemoizationProfiler {
  constructor() {
    this.stats = new Map()
  }

  profile(name, fn, options = {}) {
    const originalFn = memoize(fn, options)

    // Initialize stats
    if (!this.stats.has(name)) {
      this.stats.set(name, {
        calls: 0,
        cacheHits: 0,
        cacheMisses: 0,
        totalTime: 0,
        averageTime: 0
      })
    }

    return (...args) => {
      const stats = this.stats.get(name)
      const startTime = performance.now()

      const hadCachedValue = originalFn.cache.has(
        options.keyGenerator ? options.keyGenerator(...args) : JSON.stringify(args)
      )

      const result = originalFn(...args)

      const endTime = performance.now()
      const executionTime = endTime - startTime

      stats.calls++
      stats.totalTime += executionTime
      stats.averageTime = stats.totalTime / stats.calls

      if (hadCachedValue) {
        stats.cacheHits++
      } else {
        stats.cacheMisses++
      }

      return result
    }
  }

  getStats(name) {
    return this.stats.get(name)
  }

  getAllStats() {
    return Object.fromEntries(this.stats)
  }

  clearStats() {
    this.stats.clear()
  }
}

// Global profiler instance
export const globalProfiler = new MemoizationProfiler()

/**
 * Composable for managing multiple memoized computeds
 */
export function useMemoizedComputeds(computedDefinitions) {
  const memoizedRefs = {}

  for (const [key, definition] of Object.entries(computedDefinitions)) {
    const { getter, dependencies = [], options = {} } = definition

    memoizedRefs[key] = memoizedComputed(getter, {
      dependencies,
      ...options
    })
  }

  return memoizedRefs
}

/**
 * Deep equality function for complex objects
 */
export function deepEqual(a, b) {
  if (a === b) return true

  if (a == null || b == null) return a === b

  if (typeof a !== 'object' || typeof b !== 'object') return a === b

  const keysA = Object.keys(a)
  const keysB = Object.keys(b)

  if (keysA.length !== keysB.length) return false

  for (const key of keysA) {
    if (!keysB.includes(key)) return false
    if (!deepEqual(a[key], b[key])) return false
  }

  return true
}

/**
 * Shallow equality function for arrays and objects
 */
export function shallowEqual(a, b) {
  if (a === b) return true

  if (a == null || b == null) return a === b

  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    return a.every((item, index) => item === b[index])
  }

  if (typeof a === 'object' && typeof b === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)

    if (keysA.length !== keysB.length) return false

    return keysA.every(key => a[key] === b[key])
  }

  return false
}
