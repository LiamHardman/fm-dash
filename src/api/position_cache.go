package main

import (
	"sync"
)

// PositionMatchCache caches position matching results for better performance
type PositionMatchCache struct {
	cache map[string]bool
	mutex sync.RWMutex
}

// Global position match cache
var positionMatchCache = &PositionMatchCache{
	cache: make(map[string]bool),
}

// CachedMatchesPositionForUpgrade checks if a player matches the required position with caching
func CachedMatchesPositionForUpgrade(player Player, position string) bool {
	// Create cache key
	key := player.Position + "|" + position

	// Try to get from cache
	positionMatchCache.mutex.RLock()
	if cached, exists := positionMatchCache.cache[key]; exists {
		positionMatchCache.mutex.RUnlock()
		return cached
	}
	positionMatchCache.mutex.RUnlock()

	// Calculate and cache
	matches := matchesPositionForUpgrade(player, position)

	positionMatchCache.mutex.Lock()
	// Prevent cache from growing too large
	if len(positionMatchCache.cache) > 5000 {
		// Clear cache if it gets too large (simple eviction)
		positionMatchCache.cache = make(map[string]bool)
	}
	positionMatchCache.cache[key] = matches
	positionMatchCache.mutex.Unlock()

	return matches
}

// ClearPositionMatchCache clears the position matching cache
func ClearPositionMatchCache() {
	positionMatchCache.mutex.Lock()
	positionMatchCache.cache = make(map[string]bool)
	positionMatchCache.mutex.Unlock()
}

// GetPositionMatchCacheStats returns cache statistics
func GetPositionMatchCacheStats() (size int, capacity int) {
	positionMatchCache.mutex.RLock()
	defer positionMatchCache.mutex.RUnlock()
	return len(positionMatchCache.cache), 5000
}
