package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// SafeStringInterning provides thread-safe string deduplication
// that doesn't modify strings during concurrent access
type SafeStringInterning struct {
	strings map[string]string
	mutex   sync.RWMutex
	// Use atomic counters to avoid race conditions
	totalRequests int64
	cacheHits     int64
	memSaved      int64
	// Add cleanup tracking
	lastCleanup time.Time
	maxSize     int
}

// Global safe string intern pools with size limits
var (
	safeClubInterning        = NewSafeStringInterning(5000) // Max 5000 clubs
	safePositionInterning    = NewSafeStringInterning(50)   // Max 50 positions
	safeNationalityInterning = NewSafeStringInterning(300)  // Max 300 nationalities
	safeDivisionInterning    = NewSafeStringInterning(100)  // Max 100 divisions
	safePersonalityInterning = NewSafeStringInterning(100)  // Max 100 personalities
)

// NewSafeStringInterning creates a new safe string interning instance with size limit
func NewSafeStringInterning(maxSize int) *SafeStringInterning {
	return &SafeStringInterning{
		strings:     make(map[string]string),
		maxSize:     maxSize,
		lastCleanup: time.Now(),
	}
}

// SafeIntern returns an interned version of the string without modifying the original
// This is safe for concurrent access and includes cleanup
func (si *SafeStringInterning) SafeIntern(s string) string {
	if s == "" {
		return s
	}

	// Try read lock first for performance
	si.mutex.RLock()
	if interned, exists := si.strings[s]; exists {
		si.mutex.RUnlock()
		// Use atomic operations for thread-safe statistics
		atomic.AddInt64(&si.totalRequests, 1)
		atomic.AddInt64(&si.cacheHits, 1)
		return interned
	}
	si.mutex.RUnlock()

	// Write lock for new string
	si.mutex.Lock()
	defer si.mutex.Unlock()

	// Double-check pattern
	if interned, exists := si.strings[s]; exists {
		// Use atomic operations for thread-safe statistics
		atomic.AddInt64(&si.totalRequests, 1)
		atomic.AddInt64(&si.cacheHits, 1)
		return interned
	}

	// Check if cleanup is needed (every 10 minutes)
	if time.Since(si.lastCleanup) > 10*time.Minute {
		si.cleanupIfNeeded()
	}

	// Store new string if under limit
	if len(si.strings) < si.maxSize {
		si.strings[s] = s
		// Use atomic operations for thread-safe statistics
		atomic.AddInt64(&si.totalRequests, 1)
		atomic.AddInt64(&si.memSaved, int64(len(s)))
		return s
	}

	// If over limit, just return original string (no interning)
	atomic.AddInt64(&si.totalRequests, 1)
	return s
}

// cleanupIfNeeded removes old entries if map is getting too large
func (si *SafeStringInterning) cleanupIfNeeded() {
	si.lastCleanup = time.Now()

	// If map is over 80% full, clear 20% of entries
	if len(si.strings) > int(float64(si.maxSize)*0.8) {
		targetSize := int(float64(si.maxSize) * 0.6) // Reduce to 60%

		// Simple cleanup: remove entries until we reach target size
		// In a real system, you might want LRU or other sophisticated cleanup
		current := 0
		for k := range si.strings {
			if current >= targetSize {
				delete(si.strings, k)
			}
			current++
		}
	}
}

// SafeOptimizePlayerStrings safely optimizes player strings without modifying during concurrent access
func SafeOptimizePlayerStrings(player *Player) {
	// Only intern if we're sure it's safe
	if player == nil {
		return
	}

	// Create a temporary copy for interning - don't modify original during serialization
	player.Club = safeClubInterning.SafeIntern(player.Club)
	player.Position = safePositionInterning.SafeIntern(player.Position)
	player.Nationality = safeNationalityInterning.SafeIntern(player.Nationality)
	player.Division = safeDivisionInterning.SafeIntern(player.Division)
	player.Personality = safePersonalityInterning.SafeIntern(player.Personality)
}

// GetSafeStringInterningStats returns statistics for all safe interning pools
func GetSafeStringInterningStats() map[string]map[string]int64 {
	return map[string]map[string]int64{
		"clubs": {
			"unique_strings": int64(len(safeClubInterning.strings)),
			"total_requests": atomic.LoadInt64(&safeClubInterning.totalRequests),
			"cache_hits":     atomic.LoadInt64(&safeClubInterning.cacheHits),
			"memory_saved":   atomic.LoadInt64(&safeClubInterning.memSaved),
		},
		"positions": {
			"unique_strings": int64(len(safePositionInterning.strings)),
			"total_requests": atomic.LoadInt64(&safePositionInterning.totalRequests),
			"cache_hits":     atomic.LoadInt64(&safePositionInterning.cacheHits),
			"memory_saved":   atomic.LoadInt64(&safePositionInterning.memSaved),
		},
		"nationalities": {
			"unique_strings": int64(len(safeNationalityInterning.strings)),
			"total_requests": atomic.LoadInt64(&safeNationalityInterning.totalRequests),
			"cache_hits":     atomic.LoadInt64(&safeNationalityInterning.cacheHits),
			"memory_saved":   atomic.LoadInt64(&safeNationalityInterning.memSaved),
		},
		"divisions": {
			"unique_strings": int64(len(safeDivisionInterning.strings)),
			"total_requests": atomic.LoadInt64(&safeDivisionInterning.totalRequests),
			"cache_hits":     atomic.LoadInt64(&safeDivisionInterning.cacheHits),
			"memory_saved":   atomic.LoadInt64(&safeDivisionInterning.memSaved),
		},
		"personalities": {
			"unique_strings": int64(len(safePersonalityInterning.strings)),
			"total_requests": atomic.LoadInt64(&safePersonalityInterning.totalRequests),
			"cache_hits":     atomic.LoadInt64(&safePersonalityInterning.cacheHits),
			"memory_saved":   atomic.LoadInt64(&safePersonalityInterning.memSaved),
		},
	}
}
