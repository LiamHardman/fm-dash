import { computed, reactive, ref } from 'vue'

/**
 * Backend-powered composable for team logo handling
 * Uses the Go API for team name to ID matching instead of loading full teams data
 */
export function useTeamLogosBackend(options = {}) {
  const {
    similarityThreshold = 0.7,
    maxResults: _maxResults = 10,
    cacheTimeout = 3600000 // 1 hour cache
  } = options

  // Cache for team matches and logos
  const matchCache = reactive(new Map())
  const logoCache = reactive(new Map())
  const isLoading = ref(false)
  const lastError = ref(null)

  /**
   * Get team matches from backend API
   * @param {string} teamName - Team name to search for
   * @returns {Promise<Array>} - Array of team matches
   */
  const getTeamMatches = async teamName => {
    if (!teamName || teamName.trim() === '') {
      return []
    }

    const normalizedName = teamName.trim()
    const cacheKey = `matches_${normalizedName.toLowerCase()}`

    // Check cache first
    const cached = matchCache.get(cacheKey)
    if (cached && Date.now() - cached.timestamp < cacheTimeout) {
      return cached.data
    }

    try {
      isLoading.value = true
      lastError.value = null

      const response = await fetch(`/api/team-match?name=${encodeURIComponent(normalizedName)}`, {
        headers: {
          'Accept': 'application/x-protobuf'
        }
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const matches = await response.json()

      // Cache the results
      matchCache.set(cacheKey, {
        data: matches,
        timestamp: Date.now()
      })

      return matches
    } catch (error) {
      lastError.value = error.message
      return []
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Get the best team match for a given name
   * @param {string} teamName - Team name to search for
   * @returns {Promise<Object|null>} - Best match or null
   */
  const getBestTeamMatch = async teamName => {
    const matches = await getTeamMatches(teamName)

    if (matches.length === 0) {
      return null
    }

    // Return the first match (highest score) that meets the threshold
    const bestMatch = matches[0]
    if (bestMatch.score >= similarityThreshold) {
      return bestMatch
    }

    return null
  }

  /**
   * Get team ID for a given team name
   * @param {string} teamName - Team name to search for
   * @returns {Promise<string|null>} - Team ID or null
   */
  const getTeamId = async teamName => {
    const match = await getBestTeamMatch(teamName)
    return match ? match.id : null
  }

  /**
   * Get team logo URL for a given team name
   * @param {string} teamName - Team name to search for
   * @returns {Promise<string|null>} - Logo URL or null
   */
  const getTeamLogoUrl = async teamName => {
    if (!teamName) return null

    const cacheKey = `logo_${teamName.toLowerCase()}`
    const cached = logoCache.get(cacheKey)

    if (cached && Date.now() - cached.timestamp < cacheTimeout) {
      return cached.url
    }

    const teamId = await getTeamId(teamName)
    if (!teamId) {
      // Cache negative result
      logoCache.set(cacheKey, {
        url: null,
        timestamp: Date.now()
      })
      return null
    }

    const logoUrl = `/api/logos?teamId=${encodeURIComponent(teamId)}`

    // Cache the logo URL
    logoCache.set(cacheKey, {
      url: logoUrl,
      timestamp: Date.now()
    })

    return logoUrl
  }

  /**
   * Create a reactive team logo URL that updates when the team name changes
   * @param {import('vue').Ref<string>} teamNameRef - Reactive team name reference
   * @returns {import('vue').ComputedRef<string|null>} - Reactive logo URL
   */
  const createReactiveLogoUrl = teamNameRef => {
    const logoUrl = ref(null)

    // Watch for changes in team name and update logo URL
    const updateLogoUrl = async () => {
      if (teamNameRef.value) {
        logoUrl.value = await getTeamLogoUrl(teamNameRef.value)
      } else {
        logoUrl.value = null
      }
    }

    // Initial load
    updateLogoUrl()

    return computed(() => logoUrl.value)
  }

  /**
   * Batch process multiple team names for logo URLs
   * @param {Array<string>} teamNames - Array of team names
   * @param {Function} onProgress - Progress callback (optional)
   * @returns {Promise<Map<string, string|null>>} - Map of team names to logo URLs
   */
  const batchGetTeamLogos = async (teamNames, onProgress = null) => {
    const results = new Map()
    const total = teamNames.length

    for (let i = 0; i < teamNames.length; i++) {
      const teamName = teamNames[i]
      const logoUrl = await getTeamLogoUrl(teamName)
      results.set(teamName, logoUrl)

      if (onProgress) {
        onProgress({
          processed: i + 1,
          total,
          current: teamName,
          percentage: Math.round(((i + 1) / total) * 100)
        })
      }
    }

    return results
  }

  /**
   * Check if a team has a logo available
   * @param {string} teamName - Team name to check
   * @returns {Promise<boolean>} - True if logo is available
   */
  const hasTeamLogo = async teamName => {
    const logoUrl = await getTeamLogoUrl(teamName)
    return logoUrl !== null
  }

  /**
   * Clear all caches
   */
  const clearCache = () => {
    matchCache.clear()
    logoCache.clear()
  }

  /**
   * Get cache statistics
   * @returns {Object} - Cache statistics
   */
  const getCacheStats = () => {
    return {
      matchCacheSize: matchCache.size,
      logoCacheSize: logoCache.size,
      totalCacheSize: matchCache.size + logoCache.size
    }
  }

  /**
   * Preload team matches for common searches
   * @param {Array<string>} teamNames - Array of team names to preload
   */
  const preloadTeamMatches = async teamNames => {
    const promises = teamNames.map(name => getTeamMatches(name))
    await Promise.allSettled(promises)
  }

  return {
    // Data
    isLoading: computed(() => isLoading.value),
    lastError: computed(() => lastError.value),
    cacheStats: computed(() => getCacheStats()),

    // Methods
    getTeamMatches,
    getBestTeamMatch,
    getTeamId,
    getTeamLogoUrl,
    createReactiveLogoUrl,
    batchGetTeamLogos,
    hasTeamLogo,
    clearCache,
    preloadTeamMatches,

    // Utilities
    getCacheStats
  }
}
