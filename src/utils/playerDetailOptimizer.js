/**
 * PlayerDetailOptimizer - Performance optimization utilities for PlayerDetailDialog
 */

import logger from './logger.js'

// Simple in-memory cache for player data
const playerDataCache = new Map()
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes

/**
 * Get cached player data
 */
export function getCachedPlayerData(datasetId, playerUID) {
  const cacheKey = `${datasetId}-${playerUID}`
  const cached = playerDataCache.get(cacheKey)
  
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.data
  }
  
  return null
}

/**
 * Set cached player data
 */
export function setCachedPlayerData(datasetId, playerUID, data) {
  const cacheKey = `${datasetId}-${playerUID}`
  playerDataCache.set(cacheKey, {
    data,
    timestamp: Date.now()
  })
  
  // Clean up old entries if cache gets too large
  if (playerDataCache.size > 100) {
    const entries = Array.from(playerDataCache.entries())
    entries.sort((a, b) => a[1].timestamp - b[1].timestamp)
    
    // Remove oldest 20 entries
    for (let i = 0; i < 20 && i < entries.length; i++) {
      playerDataCache.delete(entries[i][0])
    }
  }
}

/**
 * Clear all cached data
 */
export function clearPlayerDataCache() {
  playerDataCache.clear()
}

/**
 * Get cache statistics
 */
export function getCacheStats() {
  const now = Date.now()
  let validEntries = 0
  let expiredEntries = 0
  
  for (const [key, value] of playerDataCache.entries()) {
    if (now - value.timestamp < CACHE_TTL) {
      validEntries++
    } else {
      expiredEntries++
    }
  }
  
  return {
    totalEntries: playerDataCache.size,
    validEntries,
    expiredEntries,
    hitRate: 0 // Would need to track hits/misses to calculate this
  }
}

/**
 * Optimized percentile cache
 */
class PercentileCache {
  constructor() {
    this.cache = new Map()
    this.timestamps = new Map()
  }

  /**
   * Generate cache key for percentile data
   */
  generateKey(playerUID, division, position) {
    return `${playerUID}:${division}:${position}`
  }

  /**
   * Get cached percentile data
   */
  get(playerUID, division, position) {
    const key = this.generateKey(playerUID, division, position)
    const cached = this.cache.get(key)
    const timestamp = this.timestamps.get(key)

    if (cached && timestamp && Date.now() - timestamp < PERCENTILE_CACHE_TTL) {
      logger.debug('Percentile cache hit', { playerUID, division, position })
      return cached
    }

    // Remove expired entry
    if (cached) {
      this.cache.delete(key)
      this.timestamps.delete(key)
    }

    logger.debug('Percentile cache miss', { playerUID, division, position })
    return null
  }

  /**
   * Set percentile data in cache
   */
  set(playerUID, division, position, data) {
    const key = this.generateKey(playerUID, division, position)
    this.cache.set(key, data)
    this.timestamps.set(key, Date.now())
    logger.debug('Percentile data cached', { playerUID, division, position })
  }

  /**
   * Clear expired entries
   */
  cleanup() {
    const now = Date.now()
    for (const [key, timestamp] of this.timestamps.entries()) {
      if (now - timestamp > PERCENTILE_CACHE_TTL) {
        this.cache.delete(key)
        this.timestamps.delete(key)
      }
    }
  }
}

// Global cache instances
const percentileCache = new PercentileCache()

/**
 * Optimized player data fetcher with caching
 */
export async function fetchPlayerDataOptimized(datasetId, playerUID, options = {}) {
  const startTime = performance.now()
  const { forceRefresh = false, includePercentiles = true } = options

  try {
    // Check cache first
    if (!forceRefresh) {
      const cached = getCachedPlayerData(datasetId, playerUID)
      if (cached) {
        return {
          data: cached,
          fromCache: true,
          loadTime: performance.now() - startTime
        }
      }
    }

    // Import the service dynamically to avoid circular dependencies
    const { fetchFullPlayerStats } = await import('../services/playerService.js')
    
    // Fetch from API
    const result = await fetchFullPlayerStats(datasetId, playerUID)
    
    if (result.format === 'json' && result.data.player) {
      const playerData = result.data.player
      
      // Cache the result
      setCachedPlayerData(datasetId, playerUID, playerData)
      
      const loadTime = performance.now() - startTime
      logger.info('Player data fetched successfully', {
        datasetId,
        playerUID,
        playerName: playerData.name,
        loadTime: `${loadTime.toFixed(2)}ms`,
        fromCache: false
      })

      return {
        data: playerData,
        fromCache: false,
        loadTime
      }
    } else {
      throw new Error('Invalid response format')
    }
  } catch (error) {
    logger.error('Failed to fetch player data', {
      datasetId,
      playerUID,
      error: error.message
    })
    throw error
  }
}

/**
 * Optimized percentile fetcher with caching
 */
export async function fetchPercentilesOptimized(playerUID, division, position) {
  const startTime = performance.now()

  try {
    // Check cache first
    const cached = percentileCache.get(playerUID, division, position)
    if (cached) {
      return {
        data: cached,
        fromCache: true,
        loadTime: performance.now() - startTime
      }
    }

    // Fetch from API
    const response = await fetch(`/api/player-percentiles/${playerUID}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        playerUID: playerUID.toString(),
        compareDivision: division,
        comparePosition: position
      })
    })

    if (response.ok) {
      const percentiles = await response.json()
      
      // Cache the result
      percentileCache.set(playerUID, division, position, percentiles)
      
      const loadTime = performance.now() - startTime
      logger.info('Percentiles fetched successfully', {
        playerUID,
        division,
        position,
        loadTime: `${loadTime.toFixed(2)}ms`,
        fromCache: false
      })

      return {
        data: percentiles,
        fromCache: false,
        loadTime
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
  } catch (error) {
    logger.error('Failed to fetch percentiles', {
      playerUID,
      division,
      position,
      error: error.message
    })
    throw error
  }
}

/**
 * Preload player data for better UX
 */
export function preloadPlayerData(datasetId, playerUID) {
  // Preload in background without blocking UI
  setTimeout(async () => {
    try {
      await fetchPlayerDataOptimized(datasetId, playerUID, { includePercentiles: false })
      logger.debug('Player data preloaded', { datasetId, playerUID })
    } catch (error) {
      logger.debug('Player data preload failed', { datasetId, playerUID, error: error.message })
    }
  }, 100)
}

/**
 * Preload percentiles for a player
 */
export function preloadPercentiles(playerUID, division = 'same', position = 'Global') {
  setTimeout(async () => {
    try {
      await fetchPercentilesOptimized(playerUID, division, position)
      logger.debug('Percentiles preloaded', { playerUID, division, position })
    } catch (error) {
      logger.debug('Percentiles preload failed', { playerUID, division, position, error: error.message })
    }
  }, 200)
}

/**
 * Get performance metrics
 */
export function getPerformanceMetrics() {
  return {
    ...performanceMetrics,
    cacheHitRate: performanceMetrics.totalRequests > 0 
      ? performanceMetrics.cacheHits / (performanceMetrics.cacheHits + performanceMetrics.cacheMisses)
      : 0,
    averageLoadTime: performanceMetrics.averageLoadTime
  }
}

/**
 * Clear all caches
 */
export function clearAllCaches() {
  playerDataCache.clear()
  percentileCache.clear()
  logger.info('All player detail caches cleared')
}

/**
 * Cleanup expired cache entries
 */
export function cleanupCaches() {
  // playerDataCache is a simple Map, so we need to manually clean up expired entries
  const now = Date.now()
  for (const [key, value] of playerDataCache.entries()) {
    if (now - value.timestamp > CACHE_TTL) {
      playerDataCache.delete(key)
    }
  }
  percentileCache.cleanup()
}

/**
 * Optimized data fetcher that combines player data and percentiles
 */
export async function fetchPlayerDataWithPercentiles(datasetId, playerUID, division = 'same', position = 'Global') {
  const startTime = performance.now()

  try {
    // Fetch player data and percentiles in parallel
    const [playerResult, percentileResult] = await Promise.allSettled([
      fetchPlayerDataOptimized(datasetId, playerUID),
      fetchPercentilesOptimized(playerUID, division, position)
    ])

    const playerData = playerResult.status === 'fulfilled' ? playerResult.value.data : null
    const percentiles = percentileResult.status === 'fulfilled' ? percentileResult.value.data : null

    // Combine the data
    if (playerData && percentiles) {
      if (!playerData.performancePercentiles) {
        playerData.performancePercentiles = {}
      }
      Object.assign(playerData.performancePercentiles, percentiles)
    }

    const totalTime = performance.now() - startTime
    logger.info('Player data with percentiles fetched', {
      datasetId,
      playerUID,
      totalTime: `${totalTime.toFixed(2)}ms`,
      playerSuccess: playerResult.status === 'fulfilled',
      percentileSuccess: percentileResult.status === 'fulfilled'
    })

    return {
      data: playerData,
      percentiles,
      loadTime: totalTime,
      fromCache: {
        player: playerResult.status === 'fulfilled' ? playerResult.value.fromCache : false,
        percentiles: percentileResult.status === 'fulfilled' ? percentileResult.value.fromCache : false
      }
    }
  } catch (error) {
    logger.error('Failed to fetch player data with percentiles', {
      datasetId,
      playerUID,
      error: error.message
    })
    throw error
  }
}

// Cleanup caches periodically
setInterval(cleanupCaches, 60000) // Every minute

export default {
  fetchPlayerDataOptimized,
  fetchPercentilesOptimized,
  preloadPlayerData,
  preloadPercentiles,
  getPerformanceMetrics,
  clearAllCaches,
  cleanupCaches,
  fetchPlayerDataWithPercentiles
} 