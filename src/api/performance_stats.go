package main

import (
	"log"
	"math"
	"sort"
	"sync"
)

// Object pools for reducing allocations during percentile calculations
var (
	float64SlicePool = sync.Pool{
		New: func() interface{} {
			return make([]float64, 0, 1000) // Pre-allocate with reasonable capacity
		},
	}
	
	percentileMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]float64, len(PerformanceStatKeys))
		},
	}
)

// calculatePercentileValue computes the percentile rank of a specific value within a sorted list of values.
// Optimized version with better performance for large datasets.
func calculatePercentileValue(value float64, sortedValues []float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return -1
	}

	// Use binary search for better performance on large datasets
	index := sort.SearchFloat64s(sortedValues, value)
	
	// Count equal values
	countEqual := 0
	for i := index; i < n && sortedValues[i] == value; i++ {
		countEqual++
	}

	percentile := (float64(index) + (0.5 * float64(countEqual))) / float64(n) * 100.0
	return math.Round(percentile)
}

// DivisionFilter represents the different division filtering options
type DivisionFilter int

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

// CalculatePlayerPerformancePercentiles computes and populates percentile ranks for various performance statistics
// for a list of players. It calculates global percentiles and percentiles within broad and detailed position groups.
// Modifies the Player objects in the players slice directly.
func CalculatePlayerPerformancePercentiles(players []Player) {
	CalculatePlayerPerformancePercentilesWithDivisionFilter(players, DivisionFilterAll, "")
}

// CalculatePlayerPerformancePercentilesWithDivisionFilter computes and populates percentile ranks with division filtering
func CalculatePlayerPerformancePercentilesWithDivisionFilter(players []Player, divisionFilter DivisionFilter, targetDivision string) {
	if len(players) == 0 {
		return
	}

	log.Printf("🔄 Calculating percentiles with division filter: %d, target: %s, player count: %d", divisionFilter, targetDivision, len(players))

	// Count players that will be included in the filter
	includedCount := 0
	for i := range players {
		if isPlayerInTargetDivision(&players[i], divisionFilter, targetDivision) {
			includedCount++
		}
	}
	log.Printf("📊 Division filter will include %d out of %d players", includedCount, len(players))

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

	// --- Global Percentiles (apply division filtering when not "all") ---
	for _, statKey := range PerformanceStatKeys {
		// Get slice from pool
		allStatValues := float64SlicePool.Get().([]float64)
		allStatValues = allStatValues[:0] // Reset slice
		defer float64SlicePool.Put(allStatValues)
		
		// Pre-allocate with estimated capacity
		if cap(allStatValues) < includedCount {
			allStatValues = make([]float64, 0, includedCount)
		}
		
		for i := range players {
			player := &players[i]
			val, ok := player.PerformanceStatsNumeric[statKey]
			// Apply division filter for Global percentiles too (except when filter is "all")
			if ok && !math.IsNaN(val) && isPlayerInTargetDivision(player, divisionFilter, targetDivision) {
				allStatValues = append(allStatValues, val)
			}
		}

		if len(allStatValues) == 0 {
			for i := range players {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
			continue
		}
		sort.Float64s(allStatValues)

		for i := range players {
			val, ok := players[i].PerformanceStatsNumeric[statKey]
			if ok && !math.IsNaN(val) {
				players[i].PerformancePercentiles["Global"][statKey] = calculatePercentileValue(val, allStatValues)
			} else {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
		}
	}

	// --- Broad Positional Group Percentiles with division filtering ---
	// Pre-collect stat values for each broad group to avoid repeated filtering
	groupStatValueLists := make(map[string]map[string][]float64)

	for _, groupName := range PositionGroupsForPercentiles {
		groupStatValueLists[groupName] = make(map[string][]float64)
		for _, statKey := range PerformanceStatKeys {
			// Estimate capacity based on player count and group distribution
			estimatedCapacity := includedCount / len(PositionGroupsForPercentiles)
			if estimatedCapacity < 10 {
				estimatedCapacity = 10
			}
			groupStatValueLists[groupName][statKey] = make([]float64, 0, estimatedCapacity)
		}
	}

	// Populate the stat lists for each group (with division filtering)
	for i := range players {
		player := &players[i]
		if !isPlayerInTargetDivision(player, divisionFilter, targetDivision) {
			continue
		}
		for _, pg := range player.PositionGroups {
			if statMap, ok := groupStatValueLists[pg]; ok {
				for _, statKey := range PerformanceStatKeys {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						statMap[statKey] = append(statMap[statKey], val)
					}
				}
			}
		}
	}

	// Calculate percentiles for broad groups
	for _, groupName := range PositionGroupsForPercentiles {
		// Ensure percentile map for this group is initialized for all players
		for i := range players {
			if players[i].PerformancePercentiles[groupName] == nil {
				players[i].PerformancePercentiles[groupName] = make(map[string]float64)
			}
		}

		for _, statKey := range PerformanceStatKeys {
			groupValues := groupStatValueLists[groupName][statKey]

			if len(groupValues) == 0 {
				for i := range players {
					isPlayerInGroup := false
					for _, pg := range players[i].PositionGroups {
						if pg == groupName {
							isPlayerInGroup = true
							break
						}
					}
					if isPlayerInGroup {
						players[i].PerformancePercentiles[groupName][statKey] = -1
					}
				}
				continue
			}

			sort.Float64s(groupValues)

			for i := range players {
				isPlayerInGroup := false
				for _, pg := range players[i].PositionGroups {
					if pg == groupName {
						isPlayerInGroup = true
						break
					}
				}
				if isPlayerInGroup {
					val, ok := players[i].PerformanceStatsNumeric[statKey]
					if ok && !math.IsNaN(val) {
						players[i].PerformancePercentiles[groupName][statKey] = calculatePercentileValue(val, groupValues)
					} else {
						players[i].PerformancePercentiles[groupName][statKey] = -1
					}
				}
			}
		}
	}

	// --- Detailed Position Group Percentiles ---
	for detailedGroupName, shortPositions := range DetailedPositionGroupsForPercentiles {
		// Ensure percentile map for this detailed group is initialized for all players
		for i := range players {
			if players[i].PerformancePercentiles[detailedGroupName] == nil {
				players[i].PerformancePercentiles[detailedGroupName] = make(map[string]float64)
			}
		}

		// Create a set for O(1) lookup
		shortPositionsSet := make(map[string]bool, len(shortPositions))
		for _, pos := range shortPositions {
			shortPositionsSet[pos] = true
		}

		for _, statKey := range PerformanceStatKeys {
			// Get slice from pool for detailed group values
			detailedGroupValues := float64SlicePool.Get().([]float64)
			detailedGroupValues = detailedGroupValues[:0]
			
			// Collect values for this detailed group
			for i := range players {
				player := &players[i]
				if !isPlayerInTargetDivision(player, divisionFilter, targetDivision) {
					continue
				}
				
				// Check if player has any position in this detailed group
				hasDetailedPosition := false
				for _, shortPos := range player.ShortPositions {
					if shortPositionsSet[shortPos] {
						hasDetailedPosition = true
						break
					}
				}
				
				if hasDetailedPosition {
					val, ok := player.PerformanceStatsNumeric[statKey]
					if ok && !math.IsNaN(val) {
						detailedGroupValues = append(detailedGroupValues, val)
					}
				}
			}

			if len(detailedGroupValues) == 0 {
				// Mark all players in this detailed group as undefined for this stat
				for i := range players {
					hasDetailedPosition := false
					for _, shortPos := range players[i].ShortPositions {
						if shortPositionsSet[shortPos] {
							hasDetailedPosition = true
							break
						}
					}
					if hasDetailedPosition {
						players[i].PerformancePercentiles[detailedGroupName][statKey] = -1
					}
				}
				float64SlicePool.Put(detailedGroupValues)
				continue
			}

			sort.Float64s(detailedGroupValues)

			// Calculate percentiles for players in this detailed group
			for i := range players {
				hasDetailedPosition := false
				for _, shortPos := range players[i].ShortPositions {
					if shortPositionsSet[shortPos] {
						hasDetailedPosition = true
						break
					}
				}
				if hasDetailedPosition {
					val, ok := players[i].PerformanceStatsNumeric[statKey]
					if ok && !math.IsNaN(val) {
						players[i].PerformancePercentiles[detailedGroupName][statKey] = calculatePercentileValue(val, detailedGroupValues)
					} else {
						players[i].PerformancePercentiles[detailedGroupName][statKey] = -1
					}
				}
			}
			
			float64SlicePool.Put(detailedGroupValues)
		}
	}

	log.Printf("✅ Percentile calculation completed for %d players", len(players))
}
