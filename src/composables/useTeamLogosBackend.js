import { computed, reactive, ref } from 'vue'

/**
 * Backend-powered composable for team logo handling
 * Uses the Go API for team name to ID matching instead of loading full teams data
 */

// Module-level singleton so all callers share the same cache. This ensures
// that when submitLogoOverride invalidates a cache entry, the next fetch
// from any component (e.g. TeamLogo) re-requests from the API.
let _singleton = null

export function useTeamLogosBackend(options = {}) {
  if (_singleton) return _singleton

  const {
    similarityThreshold = 0.7,
    maxResults: _maxResults = 10,
    cacheTimeout = 3600000, // 1 hour cache
  } = options

  // Cache for team matches and logos
  const matchCache = reactive(new Map())
  const resolutionCache = reactive(new Map())
  const logoCache = reactive(new Map())
  const isLoading = ref(false)
  const lastError = ref(null)

  // In-flight request deduplication: maps cache key → pending Promise
  const pendingResolutions = new Map()
  const pendingMatches = new Map()

  /**
   * Get team matches from backend API
   * @param {string} teamName - Team name to search for
   * @returns {Promise<Array>} - Array of team matches
   */
  const getTeamMatches = async (teamName) => {
    if (!teamName || teamName.trim() === '') {
      return []
    }

    const normalizedName = teamName.trim()
    const cacheKey = `matches_${normalizedName.toLowerCase()}`

    const cached = matchCache.get(cacheKey)
    if (cached && Date.now() - cached.timestamp < cacheTimeout) {
      return cached.data
    }

    if (pendingMatches.has(cacheKey)) {
      return pendingMatches.get(cacheKey)
    }

    const promise = (async () => {
      try {
        isLoading.value = true
        lastError.value = null

        const response = await fetch(`/api/team-match?name=${encodeURIComponent(normalizedName)}`)

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        const matches = await response.json()

        matchCache.set(cacheKey, {
          data: matches,
          timestamp: Date.now(),
        })

        return matches
      } catch (error) {
        lastError.value = error.message
        return []
      } finally {
        isLoading.value = false
        pendingMatches.delete(cacheKey)
      }
    })()

    pendingMatches.set(cacheKey, promise)
    return promise
  }

  /**
   * Resolve a team name to structured logo display metadata.
   * @param {string} teamName - Team name to resolve
   * @returns {Promise<Object|null>} - Resolution payload or null
   */
  const getClubLogoResolution = async (teamName) => {
    if (!teamName || teamName.trim() === '') {
      return null
    }

    const normalizedName = teamName.trim()
    const cacheKey = `resolution_${normalizedName.toLowerCase()}`

    const cached = resolutionCache.get(cacheKey)
    if (cached && Date.now() - cached.timestamp < cacheTimeout) {
      return cached.data
    }

    if (pendingResolutions.has(cacheKey)) {
      return pendingResolutions.get(cacheKey)
    }

    const promise = (async () => {
      try {
        isLoading.value = true
        lastError.value = null

        const response = await fetch(
          `/api/club-logo/resolve?name=${encodeURIComponent(normalizedName)}`
        )

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        const resolution = await response.json()
        resolutionCache.set(cacheKey, {
          data: resolution,
          timestamp: Date.now(),
        })

        return resolution
      } catch (error) {
        lastError.value = error.message
        return null
      } finally {
        isLoading.value = false
        pendingResolutions.delete(cacheKey)
      }
    })()

    pendingResolutions.set(cacheKey, promise)
    return promise
  }

  /**
   * Get the best team match for a given name
   * @param {string} teamName - Team name to search for
   * @returns {Promise<Object|null>} - Best match or null
   */
  const getBestTeamMatch = async (teamName) => {
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
  const getTeamId = async (teamName) => {
    const resolution = await getClubLogoResolution(teamName)
    if (resolution && ['exact', 'confident'].includes(resolution.status)) {
      return resolution.teamId || null
    }
    return null
  }

  /**
   * Get team logo URL for a given team name
   * @param {string} teamName - Team name to search for
   * @returns {Promise<string|null>} - Logo URL or null
   */
  const getTeamLogoUrl = async (teamName) => {
    if (!teamName) return null

    const cacheKey = `logo_${teamName.toLowerCase()}`
    const cached = logoCache.get(cacheKey)

    if (cached && Date.now() - cached.timestamp < cacheTimeout) {
      return cached.url
    }

    const resolution = await getClubLogoResolution(teamName)
    const isDisplayable =
      resolution && ['exact', 'confident'].includes(resolution.status) && resolution.logoUrl

    if (!isDisplayable) {
      // Cache negative result
      logoCache.set(cacheKey, {
        url: null,
        timestamp: Date.now(),
      })
      return null
    }

    const logoUrl = resolution.logoUrl

    // Cache the logo URL
    logoCache.set(cacheKey, {
      url: logoUrl,
      timestamp: Date.now(),
    })

    return logoUrl
  }

  /**
   * Create a reactive team logo URL that updates when the team name changes
   * @param {import('vue').Ref<string>} teamNameRef - Reactive team name reference
   * @returns {import('vue').ComputedRef<string|null>} - Reactive logo URL
   */
  const createReactiveLogoUrl = (teamNameRef) => {
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
          percentage: Math.round(((i + 1) / total) * 100),
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
  const hasTeamLogo = async (teamName) => {
    const resolution = await getClubLogoResolution(teamName)
    return Boolean(
      resolution && ['exact', 'confident'].includes(resolution.status) && resolution.logoUrl
    )
  }

  /**
   * Submit a logo override for a club name.
   * @param {string} teamName - The raw club name as seen in the FM export
   * @param {string|null} teamId - The confirmed team ID, or null/empty to reject any logo for this name
   */
  const submitLogoOverride = async (teamName, teamId) => {
    if (!teamName) return

    try {
      await fetch('/api/club-logo/overrides', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: teamName, teamId: teamId ?? '' }),
      })

      // Invalidate cached resolution and logo URL for this name so the next
      // call re-fetches with the new override applied
      const lower = teamName.toLowerCase()
      resolutionCache.delete(`resolution_${lower}`)
      logoCache.delete(`logo_${lower}`)
    } catch (error) {
      lastError.value = error.message
    }
  }

  /**
   * Remove a previously-saved override so normal matching resumes.
   * @param {string} teamName - The raw club name to clear
   */
  const deleteLogoOverride = async (teamName) => {
    if (!teamName) return

    try {
      await fetch(`/api/club-logo/overrides?name=${encodeURIComponent(teamName)}`, {
        method: 'DELETE',
      })

      const lower = teamName.toLowerCase()
      resolutionCache.delete(`resolution_${lower}`)
      logoCache.delete(`logo_${lower}`)
    } catch (error) {
      lastError.value = error.message
    }
  }

  /**
   * Clear all caches
   */
  const clearCache = () => {
    matchCache.clear()
    resolutionCache.clear()
    logoCache.clear()
  }

  /**
   * Get cache statistics
   * @returns {Object} - Cache statistics
   */
  const getCacheStats = () => {
    return {
      matchCacheSize: matchCache.size,
      resolutionCacheSize: resolutionCache.size,
      logoCacheSize: logoCache.size,
      totalCacheSize: matchCache.size + resolutionCache.size + logoCache.size,
    }
  }

  /**
   * Preload team matches for common searches
   * @param {Array<string>} teamNames - Array of team names to preload
   */
  const preloadTeamMatches = async (teamNames) => {
    const promises = teamNames.map((name) => getTeamMatches(name))
    await Promise.allSettled(promises)
  }

  _singleton = {
    // Data
    isLoading: computed(() => isLoading.value),
    lastError: computed(() => lastError.value),
    cacheStats: computed(() => getCacheStats()),

    // Methods
    getTeamMatches,
    getClubLogoResolution,
    getBestTeamMatch,
    getTeamId,
    getTeamLogoUrl,
    createReactiveLogoUrl,
    batchGetTeamLogos,
    hasTeamLogo,
    clearCache,
    preloadTeamMatches,
    submitLogoOverride,
    deleteLogoOverride,

    // Utilities
    getCacheStats,
  }
  return _singleton
}
