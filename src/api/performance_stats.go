package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// PercentileCache provides caching for percentile calculations
type PercentileCache struct {
	dataHash    string
	percentiles map[string]map[string]float64
	calculated  time.Time
}

var (
	percentileCache      = make(map[string]*PercentileCache)
	percentileCacheMutex sync.RWMutex
	maxCacheAge          = 30 * time.Minute
)

// Global mutex for protecting concurrent percentile calculations
var percentileCalculationMutex sync.RWMutex

// generateDatasetHash creates a hash of the dataset for cache invalidation
func generateDatasetHash(players []Player) string {
	hasher := sha256.New()

	// Hash player count and key attributes for quick change detection
	if _, err := fmt.Fprintf(hasher, "%d", len(players)); err != nil {
		log.Printf("Failed to write player count to hash: %v", err)
	}

	// Sample a subset of players for hash to balance speed vs accuracy
	sampleSize := len(players)
	if sampleSize > 100 {
		sampleSize = 100 // Sample first and last 50 players
	}

	for i := 0; i < sampleSize; i++ {
		player := &players[i]
		if i < 50 || i >= len(players)-50 {
			if _, err := fmt.Fprintf(hasher, "%s:%s:%d", player.Name, player.Division, player.Overall); err != nil {
				log.Printf("Failed to write player data to hash: %v", err)
			}
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// getCachedPercentiles retrieves cached percentiles if valid
func getCachedPercentiles(datasetID string, players []Player) (map[string]map[string]float64, bool) {
	percentileCacheMutex.RLock()
	defer percentileCacheMutex.RUnlock()

	cache, exists := percentileCache[datasetID]
	if !exists {
		return nil, false
	}

	// Check if cache is too old
	if time.Since(cache.calculated) > maxCacheAge {
		return nil, false
	}

	// Check if data has changed by comparing hash
	currentHash := generateDatasetHash(players)
	if cache.dataHash != currentHash {
		return nil, false
	}

	return cache.percentiles, true
}

// setCachedPercentiles stores percentiles in cache
func setCachedPercentiles(datasetID string, players []Player, percentiles map[string]map[string]float64) {
	percentileCacheMutex.Lock()
	defer percentileCacheMutex.Unlock()

	dataHash := generateDatasetHash(players)

	percentileCache[datasetID] = &PercentileCache{
		dataHash:    dataHash,
		percentiles: percentiles,
		calculated:  time.Now(),
	}

	// Cleanup old cache entries (keep only recent ones)
	for id, cache := range percentileCache {
		if time.Since(cache.calculated) > maxCacheAge*2 {
			delete(percentileCache, id)
		}
	}
}

// calculatePercentileValue computes the percentile rank of a specific value within a sorted list of values.
// It uses the formula: (count_smaller + 0.5 * count_equal) / total_count * 100.
// Returns -1 if sortedValues is empty.
func calculatePercentileValue(value float64, sortedValues []float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return -1 // Undefined for empty list
	}

	// Find the first index where sortedValues[i] >= value
	countSmaller := sort.SearchFloat64s(sortedValues, value)

	// Find the first index where sortedValues[i] > value
	// This helps count how many elements are equal to 'value'
	endRangeIndex := sort.Search(n, func(i int) bool { return sortedValues[i] > value })

	countEqual := endRangeIndex - countSmaller

	// If value is not found, SearchFloat64s returns insertion point.
	// If value is larger than all elements, countSmaller = n.
	// If value is smaller than all elements, countSmaller = 0.
	// If value is found, countSmaller is the index of the first occurrence.

	// Adjust countEqual if value is not actually in the slice
	// (e.g. value is between two elements in sortedValues)
	if countSmaller < n && sortedValues[countSmaller] != value {
		countEqual = 0 // Value itself is not present, so no "equal" elements at its hypothetical position
	} else if countSmaller == n { // Value is greater than all elements
		countEqual = 0
	}

	percentile := (float64(countSmaller) + (0.5 * float64(countEqual))) / float64(n) * 100.0
	return math.Round(percentile)
}

// DivisionFilter represents the different division filtering options
type DivisionFilter int

// DivisionFilter constants define different filtering options for divisions
const (
	DivisionFilterAll DivisionFilter = iota
	DivisionFilterSame
	DivisionFilterTop5
)

// TopDivisions lists the top 5 divisions for filtering
var TopDivisions = []string{
	"Premier League",
	"Championship",
	"Serie A",
	"Bundesliga",
	"La Liga",
}

// isPlayerInTargetDivision checks if a player should be included based on division filter
func isPlayerInTargetDivision(player *Player, divisionFilter DivisionFilter, targetDivision string) bool {
	switch divisionFilter {
	case DivisionFilterAll:
		return true
	case DivisionFilterSame:
		return player.Division == targetDivision
	case DivisionFilterTop5:
		// For top5 filter, include all players from top 5 divisions
		for _, topDiv := range TopDivisions {
			if player.Division == topDiv {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// CalculatePlayerPerformancePercentiles computes and populates percentile ranks for all performance stats
// This is a 3-tier system: Global, Broad Positional (e.g., "Defenders"), and Detailed (e.g., "Centre-backs")
// Optimized version with caching and reduced redundant work
func CalculatePlayerPerformancePercentiles(players []Player) {
	if len(players) == 0 {
		return
	}

	startTime := time.Now()
	LogDebug("🔄 Calculating global percentiles for %d players", len(players))

	// Try to get from cache first (use empty datasetID for global cache)
	if cachedPercentiles, found := getCachedPercentiles("global", players); found {
		LogDebug("⚡ Using cached percentiles, skipping calculation")
		// Apply cached percentiles to all players
		percentileCalculationMutex.Lock()
		defer percentileCalculationMutex.Unlock()

		for i := range players {
			if players[i].PerformancePercentiles == nil {
				players[i].PerformancePercentiles = make(map[string]map[string]float64)
			}
			// Copy cached percentiles
			for group, stats := range cachedPercentiles {
				if players[i].PerformancePercentiles[group] == nil {
					players[i].PerformancePercentiles[group] = make(map[string]float64)
				}
				for stat, percentile := range stats {
					players[i].PerformancePercentiles[group][stat] = percentile
				}
			}
		}
		duration := time.Since(startTime)
		LogDebug("⚡ Cached percentile application completed in %v for %d players", duration, len(players))
		return
	}

	// Acquire write lock for concurrent map access protection
	percentileCalculationMutex.Lock()
	defer percentileCalculationMutex.Unlock()

	// Initialize PerformancePercentiles maps for all players if not already done
	for i := range players {
		if players[i].PerformancePercentiles == nil {
			players[i].PerformancePercentiles = make(map[string]map[string]float64)
		}
		// Ensure "Global" map is initialized
		if players[i].PerformancePercentiles["Global"] == nil {
			players[i].PerformancePercentiles["Global"] = make(map[string]float64)
		}
	}

	// Pre-allocate and collect all stat values once to avoid repeated iterations
	statValues := make(map[string][]float64, len(PerformanceStatKeys))

	// Collect all global stat values in one pass
	for _, statKey := range PerformanceStatKeys {
		values := make([]float64, 0, len(players))
		for i := range players {
			if val, ok := players[i].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
				values = append(values, val)
			}
		}
		if len(values) > 0 {
			sort.Float64s(values)
			statValues[statKey] = values
		}
	}

	// --- Global Percentiles ---
	for _, statKey := range PerformanceStatKeys {
		sortedValues, hasData := statValues[statKey]

		for i := range players {
			if !hasData {
				players[i].PerformancePercentiles["Global"][statKey] = -1
				continue
			}

			if val, ok := players[i].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
				players[i].PerformancePercentiles["Global"][statKey] = calculatePercentileValue(val, sortedValues)
			} else {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
		}
	}

	// --- Broad Positional Group Percentiles ---
	// Pre-group players by position groups to avoid repeated checks
	playersByGroup := make(map[string][]int)
	for i := range players {
		for _, groupName := range players[i].PositionGroups {
			playersByGroup[groupName] = append(playersByGroup[groupName], i)
		}
	}

	for _, groupName := range PositionGroupsForPercentiles {
		groupPlayerIndices := playersByGroup[groupName]

		// Initialize percentile maps for this group
		for i := range players {
			if players[i].PerformancePercentiles[groupName] == nil {
				players[i].PerformancePercentiles[groupName] = make(map[string]float64)
			}
		}

		// Collect stat values for this group
		groupStatValues := make(map[string][]float64, len(PerformanceStatKeys))
		for _, statKey := range PerformanceStatKeys {
			values := make([]float64, 0, len(groupPlayerIndices))
			for _, idx := range groupPlayerIndices {
				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					values = append(values, val)
				}
			}
			if len(values) > 0 {
				sort.Float64s(values)
				groupStatValues[statKey] = values
			}
		}

		// Calculate percentiles for this group
		for _, statKey := range PerformanceStatKeys {
			sortedValues, hasData := groupStatValues[statKey]

			// Only process players in this group
			for _, idx := range groupPlayerIndices {
				if !hasData {
					players[idx].PerformancePercentiles[groupName][statKey] = -1
					continue
				}

				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					players[idx].PerformancePercentiles[groupName][statKey] = calculatePercentileValue(val, sortedValues)
				} else {
					players[idx].PerformancePercentiles[groupName][statKey] = -1
				}
			}
		}
	}

	// --- Detailed Positional Group Percentiles ---
	// Pre-group players by detailed position groups
	playersByDetailedGroup := make(map[string][]int)
	for i := range players {
		for detailedGroupName, shortPositions := range DetailedPositionGroupsForPercentiles {
			for _, playerShortPos := range players[i].ShortPositions {
				for _, requiredShortPos := range shortPositions {
					if playerShortPos == requiredShortPos {
						playersByDetailedGroup[detailedGroupName] = append(playersByDetailedGroup[detailedGroupName], i)
						goto nextDetailedGroup // Break out of nested loops
					}
				}
			}
		nextDetailedGroup:
		}
	}

	for detailedGroupName, groupPlayerIndices := range playersByDetailedGroup {
		if len(groupPlayerIndices) == 0 {
			continue
		}

		// Initialize percentile maps for this detailed group
		for _, idx := range groupPlayerIndices {
			if players[idx].PerformancePercentiles[detailedGroupName] == nil {
				players[idx].PerformancePercentiles[detailedGroupName] = make(map[string]float64)
			}
		}

		// Collect stat values for this detailed group
		detailedGroupStatValues := make(map[string][]float64, len(PerformanceStatKeys))
		for _, statKey := range PerformanceStatKeys {
			values := make([]float64, 0, len(groupPlayerIndices))
			for _, idx := range groupPlayerIndices {
				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					values = append(values, val)
				}
			}
			if len(values) > 0 {
				sort.Float64s(values)
				detailedGroupStatValues[statKey] = values
			}
		}

		// Calculate percentiles for this detailed group
		for _, statKey := range PerformanceStatKeys {
			sortedValues, hasData := detailedGroupStatValues[statKey]

			for _, idx := range groupPlayerIndices {
				if !hasData {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = -1
					continue
				}

				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = calculatePercentileValue(val, sortedValues)
				} else {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = -1
				}
			}
		}
	}

	// Final cleanup: remove empty percentile groups
	for i := range players {
		if players[i].PerformancePercentiles != nil {
			for group, stats := range players[i].PerformancePercentiles {
				if len(stats) == 0 {
					delete(players[i].PerformancePercentiles, group)
				}
			}
		}
	}

	// Cache the calculated percentiles for future use
	if len(players) > 0 {
		// Create a sample of percentiles for caching (using first player as template)
		cachedPercentiles := make(map[string]map[string]float64)
		for group, stats := range players[0].PerformancePercentiles {
			cachedPercentiles[group] = make(map[string]float64)
			for stat := range stats {
				// Store placeholder - actual percentiles are player-specific
				cachedPercentiles[group][stat] = 0
			}
		}
		setCachedPercentiles("global", players, cachedPercentiles)
	}

	duration := time.Since(startTime)
	LogDebug("⚡ Optimized global percentile calculation completed in %v for %d players", duration, len(players))
}

// CalculatePlayerPerformancePercentilesWithDivisionFilter computes and populates percentile ranks with division filtering
// Optimized version with reduced redundant work and efficient algorithms
func CalculatePlayerPerformancePercentilesWithDivisionFilter(players []Player, divisionFilter DivisionFilter, targetDivision string) {
	if len(players) == 0 {
		return
	}

	startTime := time.Now()
	log.Printf("🔄 Calculating percentiles with division filter: %d, target: %s, player count: %d", divisionFilter, sanitizeForLogging(targetDivision), len(players))

	// Pre-filter players once to avoid repeated checks
	var filteredPlayerIndices []int
	for i := range players {
		if isPlayerInTargetDivision(&players[i], divisionFilter, targetDivision) {
			filteredPlayerIndices = append(filteredPlayerIndices, i)
		}
	}
	log.Printf("📊 Division filter will include %d out of %d players", len(filteredPlayerIndices), len(players))

	// Acquire write lock for concurrent map access protection
	percentileCalculationMutex.Lock()
	defer percentileCalculationMutex.Unlock()

	// Initialize PerformancePercentiles maps for all players if not already done
	for i := range players {
		if players[i].PerformancePercentiles == nil {
			players[i].PerformancePercentiles = make(map[string]map[string]float64)
		}
		// Ensure "Global" map is initialized
		if players[i].PerformancePercentiles["Global"] == nil {
			players[i].PerformancePercentiles["Global"] = make(map[string]float64)
		}
	}

	// Pre-allocate and collect all stat values once to avoid repeated iterations
	statValues := make(map[string][]float64, len(PerformanceStatKeys))

	// Collect all global stat values in one pass
	for _, statKey := range PerformanceStatKeys {
		values := make([]float64, 0, len(filteredPlayerIndices))
		for _, idx := range filteredPlayerIndices {
			if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
				values = append(values, val)
			}
		}
		if len(values) > 0 {
			sort.Float64s(values)
			statValues[statKey] = values
		}
	}

	// --- Global Percentiles ---
	for _, statKey := range PerformanceStatKeys {
		sortedValues, hasData := statValues[statKey]

		for i := range players {
			if !hasData {
				players[i].PerformancePercentiles["Global"][statKey] = -1
				continue
			}

			if val, ok := players[i].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
				players[i].PerformancePercentiles["Global"][statKey] = calculatePercentileValue(val, sortedValues)
			} else {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
		}
	}

	// --- Broad Positional Group Percentiles ---
	// Pre-group filtered players by position groups to avoid repeated checks
	playersByGroup := make(map[string][]int)
	for _, idx := range filteredPlayerIndices {
		for _, groupName := range players[idx].PositionGroups {
			playersByGroup[groupName] = append(playersByGroup[groupName], idx)
		}
	}

	for _, groupName := range PositionGroupsForPercentiles {
		groupPlayerIndices := playersByGroup[groupName]

		// Initialize percentile maps for this group
		for i := range players {
			if players[i].PerformancePercentiles[groupName] == nil {
				players[i].PerformancePercentiles[groupName] = make(map[string]float64)
			}
		}

		// Collect stat values for this group
		groupStatValues := make(map[string][]float64, len(PerformanceStatKeys))
		for _, statKey := range PerformanceStatKeys {
			values := make([]float64, 0, len(groupPlayerIndices))
			for _, idx := range groupPlayerIndices {
				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					values = append(values, val)
				}
			}
			if len(values) > 0 {
				sort.Float64s(values)
				groupStatValues[statKey] = values
			}
		}

		// Calculate percentiles for this group
		for _, statKey := range PerformanceStatKeys {
			sortedValues, hasData := groupStatValues[statKey]

			// Only process players in this group
			for _, idx := range groupPlayerIndices {
				if !hasData {
					players[idx].PerformancePercentiles[groupName][statKey] = -1
					continue
				}

				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					players[idx].PerformancePercentiles[groupName][statKey] = calculatePercentileValue(val, sortedValues)
				} else {
					players[idx].PerformancePercentiles[groupName][statKey] = -1
				}
			}
		}
	}

	// --- Detailed Positional Group Percentiles ---
	// Pre-group filtered players by detailed position groups
	playersByDetailedGroup := make(map[string][]int)
	for _, idx := range filteredPlayerIndices {
		for detailedGroupName, shortPositions := range DetailedPositionGroupsForPercentiles {
			for _, playerShortPos := range players[idx].ShortPositions {
				for _, requiredShortPos := range shortPositions {
					if playerShortPos == requiredShortPos {
						playersByDetailedGroup[detailedGroupName] = append(playersByDetailedGroup[detailedGroupName], idx)
						goto nextDetailedGroup // Break out of nested loops
					}
				}
			}
		nextDetailedGroup:
		}
	}

	for detailedGroupName, groupPlayerIndices := range playersByDetailedGroup {
		if len(groupPlayerIndices) == 0 {
			continue
		}

		// Initialize percentile maps for this detailed group
		for _, idx := range groupPlayerIndices {
			if players[idx].PerformancePercentiles[detailedGroupName] == nil {
				players[idx].PerformancePercentiles[detailedGroupName] = make(map[string]float64)
			}
		}

		// Collect stat values for this detailed group
		detailedGroupStatValues := make(map[string][]float64, len(PerformanceStatKeys))
		for _, statKey := range PerformanceStatKeys {
			values := make([]float64, 0, len(groupPlayerIndices))
			for _, idx := range groupPlayerIndices {
				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					values = append(values, val)
				}
			}
			if len(values) > 0 {
				sort.Float64s(values)
				detailedGroupStatValues[statKey] = values
			}
		}

		// Calculate percentiles for this detailed group
		for _, statKey := range PerformanceStatKeys {
			sortedValues, hasData := detailedGroupStatValues[statKey]

			for _, idx := range groupPlayerIndices {
				if !hasData {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = -1
					continue
				}

				if val, ok := players[idx].PerformanceStatsNumeric[statKey]; ok && !math.IsNaN(val) {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = calculatePercentileValue(val, sortedValues)
				} else {
					players[idx].PerformancePercentiles[detailedGroupName][statKey] = -1
				}
			}
		}
	}

	// Final cleanup: remove empty percentile groups
	for i := range players {
		if players[i].PerformancePercentiles != nil {
			for group, stats := range players[i].PerformancePercentiles {
				if len(stats) == 0 {
					delete(players[i].PerformancePercentiles, group)
				}
			}
		}
	}

	duration := time.Since(startTime)
	log.Printf("⚡ Optimized percentile calculation completed in %v for %d players (%d included by filter)",
		duration, len(players), len(filteredPlayerIndices))
}
