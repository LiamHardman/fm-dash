package main

import (
	"math"
	"sort"
)

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

// CalculatePlayerPerformancePercentiles computes and populates percentile ranks for various performance statistics
// for a list of players. It calculates global percentiles and percentiles within broad and detailed position groups.
// Modifies the Player objects in the players slice directly.
func CalculatePlayerPerformancePercentiles(players []Player) {
	if len(players) == 0 {
		return
	}

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

	// --- Global Percentiles ---
	for _, statKey := range PerformanceStatKeys { // PerformanceStatKeys from config.go
		allStatValues := make([]float64, 0, len(players))
		for i := range players { // Iterate by index to modify original slice elements
			val, ok := players[i].PerformanceStatsNumeric[statKey]
			if ok && !math.IsNaN(val) { // Only include valid, non-NaN numbers
				allStatValues = append(allStatValues, val)
			}
		}

		if len(allStatValues) == 0 { // No valid data for this stat across all players
			for i := range players {
				players[i].PerformancePercentiles["Global"][statKey] = -1 // Mark as undefined
			}
			continue // Move to the next statKey
		}
		sort.Float64s(allStatValues) // Sort all values for this stat

		for i := range players {
			val, ok := players[i].PerformanceStatsNumeric[statKey]
			if ok && !math.IsNaN(val) {
				players[i].PerformancePercentiles["Global"][statKey] = calculatePercentileValue(val, allStatValues)
			} else {
				players[i].PerformancePercentiles["Global"][statKey] = -1 // Undefined if player has no valid stat
			}
		}
	}

	// --- Broad Positional Group Percentiles (e.g., "Defenders", "Midfielders") ---
	// Pre-collect stat values for each broad group to avoid repeated filtering
	groupStatValueLists := make(map[string]map[string][]float64) // groupName -> statKey -> []values

	for _, groupName := range PositionGroupsForPercentiles { // From config.go
		groupStatValueLists[groupName] = make(map[string][]float64)
		for _, statKey := range PerformanceStatKeys {
			groupStatValueLists[groupName][statKey] = make([]float64, 0, len(players)/len(PositionGroupsForPercentiles)) // Estimate capacity
		}
	}

	// Populate the stat lists for each group
	for i := range players {
		player := &players[i]                      // Work with a pointer to the player
		for _, pg := range player.PositionGroups { // Player's broad position groups
			if _, ok := groupStatValueLists[pg]; ok { // If this is a group we're tracking
				for _, statKey := range PerformanceStatKeys {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						groupStatValueLists[pg][statKey] = append(groupStatValueLists[pg][statKey], val)
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

			if len(groupValues) == 0 { // No data for this stat in this group
				for i := range players { // Iterate all players
					isPlayerInGroup := false
					for _, pg := range players[i].PositionGroups {
						if pg == groupName {
							isPlayerInGroup = true
							break
						}
					}
					if isPlayerInGroup { // Only mark for players in this group
						players[i].PerformancePercentiles[groupName][statKey] = -1
					}
				}
				continue
			}
			sort.Float64s(groupValues)

			for i := range players {
				player := &players[i]
				isPlayerInGroup := false
				for _, pg := range player.PositionGroups {
					if pg == groupName {
						isPlayerInGroup = true
						break
					}
				}

				if isPlayerInGroup {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						player.PerformancePercentiles[groupName][statKey] = calculatePercentileValue(val, groupValues)
					} else {
						player.PerformancePercentiles[groupName][statKey] = -1
					}
				}
			}
		}
	}

	// --- Detailed Positional Group Percentiles (e.g., "Full-backs", "Centre-backs") ---
	// DetailedPositionGroupsForPercentiles from config.go
	for detailedGroupName, shortPositionsInGroup := range DetailedPositionGroupsForPercentiles {
		// Ensure percentile map for this detailed group is initialized for all players
		for i := range players {
			if players[i].PerformancePercentiles[detailedGroupName] == nil {
				players[i].PerformancePercentiles[detailedGroupName] = make(map[string]float64)
			}
		}

		currentDetailedGroupStatValues := make(map[string][]float64) // statKey -> []values
		for _, statKey := range PerformanceStatKeys {
			currentDetailedGroupStatValues[statKey] = make([]float64, 0, len(players)/10) // Estimate capacity
		}

		playerIndicesInDetailedGroup := []int{} // Store indices of players belonging to this detailed group

		// Collect stat values and player indices for the current detailed group
		for i := range players {
			player := &players[i]
			isPlayerInThisDetailedGroup := false
			for _, playerShortPos := range player.ShortPositions { // Check against player's short codes
				for _, requiredShortPos := range shortPositionsInGroup {
					if playerShortPos == requiredShortPos {
						isPlayerInThisDetailedGroup = true
						break
					}
				}
				if isPlayerInThisDetailedGroup {
					break
				}
			}

			if isPlayerInThisDetailedGroup {
				playerIndicesInDetailedGroup = append(playerIndicesInDetailedGroup, i)
				for _, statKey := range PerformanceStatKeys {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						currentDetailedGroupStatValues[statKey] = append(currentDetailedGroupStatValues[statKey], val)
					}
				}
			}
		}

		// Calculate percentiles for this detailed group
		for _, statKey := range PerformanceStatKeys {
			statValuesForCalc := currentDetailedGroupStatValues[statKey]

			if len(statValuesForCalc) == 0 { // No data for this stat in this detailed group
				for _, playerIndex := range playerIndicesInDetailedGroup {
					players[playerIndex].PerformancePercentiles[detailedGroupName][statKey] = -1
				}
				continue
			}
			sort.Float64s(statValuesForCalc)

			for _, playerIndex := range playerIndicesInDetailedGroup {
				player := &players[playerIndex]
				val, statOk := player.PerformanceStatsNumeric[statKey]
				if statOk && !math.IsNaN(val) {
					player.PerformancePercentiles[detailedGroupName][statKey] = calculatePercentileValue(val, statValuesForCalc)
				} else {
					player.PerformancePercentiles[detailedGroupName][statKey] = -1
				}
			}
		}
	}
}
