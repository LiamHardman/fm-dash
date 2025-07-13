package main

import (
	"sync"
)

// StringInterning provides memory-efficient string deduplication
type StringInterning struct {
	strings map[string]string
	mutex   sync.RWMutex
	stats   struct {
		totalRequests int64
		cacheHits     int64
		memSaved      int64
	}
}

// Global string intern pools for different categories
var (
	clubInterning        = NewStringInterning()
	positionInterning    = NewStringInterning()
	nationalityInterning = NewStringInterning()
	divisionInterning    = NewStringInterning()
	personalityInterning = NewStringInterning()
)

// NewStringInterning creates a new string interning instance
func NewStringInterning() *StringInterning {
	return &StringInterning{
		strings: make(map[string]string),
	}
}

// Intern returns the canonical version of the string, reducing memory usage
func (si *StringInterning) Intern(s string) string {
	if s == "" {
		return s
	}

	si.mutex.RLock()
	if interned, exists := si.strings[s]; exists {
		si.mutex.RUnlock()
		si.stats.cacheHits++
		return interned
	}
	si.mutex.RUnlock()

	si.mutex.Lock()
	defer si.mutex.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := si.strings[s]; exists {
		si.stats.cacheHits++
		return interned
	}

	// Store the string and return it
	si.strings[s] = s
	si.stats.totalRequests++
	si.stats.memSaved += int64(len(s)) // Rough estimate of memory saved
	return s
}

// Stats returns interning statistics
func (si *StringInterning) Stats() (totalRequests, cacheHits, memSaved int64) {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.stats.totalRequests, si.stats.cacheHits, si.stats.memSaved
}

// Clear removes all interned strings (useful for testing or memory cleanup)
func (si *StringInterning) Clear() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.strings = make(map[string]string)
	si.stats.totalRequests = 0
	si.stats.cacheHits = 0
	si.stats.memSaved = 0
}

// Size returns the number of unique strings interned
func (si *StringInterning) Size() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return len(si.strings)
}

// InternClub interns a club string for memory efficiency
func InternClub(club string) string {
	return clubInterning.Intern(club)
}

// InternPosition interns a position string for memory efficiency
func InternPosition(position string) string {
	return positionInterning.Intern(position)
}

// InternNationality interns a nationality string for memory efficiency
func InternNationality(nationality string) string {
	return nationalityInterning.Intern(nationality)
}

// InternDivision interns a division string for memory efficiency
func InternDivision(division string) string {
	return divisionInterning.Intern(division)
}

// InternPersonality interns a personality string for memory efficiency
func InternPersonality(personality string) string {
	return personalityInterning.Intern(personality)
}

// GetStringInterningStats returns statistics for all interning pools
func GetStringInterningStats() map[string]map[string]int64 {
	stats := make(map[string]map[string]int64)

	pools := map[string]*StringInterning{
		"clubs":         clubInterning,
		"positions":     positionInterning,
		"nationalities": nationalityInterning,
		"divisions":     divisionInterning,
		"personalities": personalityInterning,
	}

	for name, pool := range pools {
		total, hits, saved := pool.Stats()
		stats[name] = map[string]int64{
			"total_requests": total,
			"cache_hits":     hits,
			"memory_saved":   saved,
			"unique_strings": int64(pool.Size()),
		}
	}

	return stats
}

// OptimizePlayerStrings applies string interning to a player
func OptimizePlayerStrings(player *Player) {
	player.Club = InternClub(player.Club)
	player.Position = InternPosition(player.Position)
	player.Division = InternDivision(player.Division)
	player.Nationality = InternNationality(player.Nationality)
	player.NationalityISO = InternNationality(player.NationalityISO)
	player.NationalityFIFACode = InternNationality(player.NationalityFIFACode)

	if player.Personality != "" {
		player.Personality = InternPersonality(player.Personality)
	}
	if player.MediaHandling != "" {
		player.MediaHandling = InternPersonality(player.MediaHandling) // Similar to personality
	}

	// Intern position-related slices
	for i, pos := range player.ParsedPositions {
		player.ParsedPositions[i] = InternPosition(pos)
	}
	for i, pos := range player.ShortPositions {
		player.ShortPositions[i] = InternPosition(pos)
	}
	for i, group := range player.PositionGroups {
		player.PositionGroups[i] = InternPosition(group)
	}

	player.BestRoleOverall = InternPosition(player.BestRoleOverall)
}
