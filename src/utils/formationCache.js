class FormationCache {
  constructor(maxSize = 100) {
    this.cache = new Map()
    this.maxSize = maxSize
  }

  // Generate cache key based on players and formation parameters
  generateKey(players, formationType = 'team') {
    if (!players || !Array.isArray(players)) return null

    // Create a hash based on player UIDs and their key attributes
    const playerHash = players
      .sort((a, b) => {
        // Create unique identifiers for sorting
        const getUniqueId = p => {
          let id = p.UID || p.uid
          if (!id || id === '') {
            id = `${p.name || 'unknown'}-${p.club || 'unknown'}-${p.age || 'unknown'}-${p.position || 'unknown'}`
          }
          // Ensure id is always a string for localeCompare
          return String(id)
        }
        return getUniqueId(a).localeCompare(getUniqueId(b))
      })
      .map(p => {
        let playerUID = p.UID || p.uid
        if (!playerUID || playerUID === '') {
          playerUID = `${p.name || 'unknown'}-${p.club || 'unknown'}-${p.age || 'unknown'}-${p.position || 'unknown'}`
        }
        // Ensure playerUID is always a string
        return `${String(playerUID)}-${p.Overall || 0}-${(p.shortPositions || []).join(',')}`
      })
      .join('|')

    return `${formationType}-${this.simpleHash(playerHash)}`
  }

  // Simple hash function for cache keys
  simpleHash(str) {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = (hash << 5) - hash + char
      hash = hash & hash // Convert to 32-bit integer
    }
    return Math.abs(hash).toString(36)
  }

  get(key) {
    if (!key) return null

    const cached = this.cache.get(key)
    if (cached) {
      // Move to end (LRU behavior)
      this.cache.delete(key)
      this.cache.set(key, cached)
      return cached
    }
    return null
  }

  set(key, value) {
    if (!key) return

    // Remove oldest entries if cache is full
    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value
      this.cache.delete(firstKey)
    }

    this.cache.set(key, {
      ...value,
      timestamp: Date.now()
    })
  }

  has(key) {
    return key && this.cache.has(key)
  }

  clear() {
    this.cache.clear()
  }

  // Remove expired entries (older than 10 minutes)
  cleanup(maxAge = 10 * 60 * 1000) {
    const now = Date.now()
    for (const [key, value] of this.cache.entries()) {
      if (now - value.timestamp > maxAge) {
        this.cache.delete(key)
      }
    }
  }

  size() {
    return this.cache.size
  }
}

// Export singleton instance
export const formationCache = new FormationCache()

// Auto-cleanup every 5 minutes
setInterval(
  () => {
    formationCache.cleanup()
  },
  5 * 60 * 1000
)
