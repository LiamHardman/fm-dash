package main

import (
	"context"
	"strconv"
	"strings"
	
	"go.opentelemetry.io/otel/attribute"
)

// DeletePlayerData removes a dataset from storage and cleans up caches
func DeletePlayerData(datasetID string) {
	// Call the existing DeleteDataset function
	_ = DeleteDataset(datasetID)
	
	// Also clean up any format-aware cache entries
	cacheKey := "players:" + datasetID
	DeleteAllFormatVariants(cacheKey)
}

// We'll use the existing DivisionFilter type from performance_stats.go

// We'll use the existing OptimizedDeepCopyPlayers function from copy_optimization.go

// ApplyDivisionFilter filters players based on division filter type
func ApplyDivisionFilter(players []Player, filterType DivisionFilter, targetDivision string) []Player {
	if filterType == DivisionFilterAll || len(players) == 0 {
		return players
	}
	
	// If no target division specified, use the first player's division
	if targetDivision == "" && len(players) > 0 {
		targetDivision = players[0].Division
	}
	
	// Top 5 divisions (hardcoded for simplicity)
	top5Divisions := map[string]bool{
		"Premier Division": true,
		"Championship": true,
		"League One": true,
		"La Liga": true,
		"Serie A": true,
	}
	
	var result []Player
	for _, player := range players {
		switch filterType {
		case DivisionFilterSame:
			if player.Division == targetDivision {
				result = append(result, player)
			}
		case DivisionFilterTop5:
			if top5Divisions[player.Division] {
				result = append(result, player)
			}
		}
	}
	
	return result
}

// ApplyAllFilters applies all filters to the player data
func ApplyAllFilters(ctx context.Context, players []Player, 
	filterPosition, filterRole, minAgeStr, maxAgeStr, 
	minTransferValueStr, maxTransferValueStr, maxSalaryStr string,
	divisionFilter DivisionFilter, targetDivision, positionCompare string) []Player {
	
	// Apply division filter first
	filteredPlayers := ApplyDivisionFilter(players, divisionFilter, targetDivision)
	
	// Apply position filter
	if filterPosition != "" {
		var positionFiltered []Player
		for _, player := range filteredPlayers {
			if strings.Contains(player.Position, filterPosition) {
				positionFiltered = append(positionFiltered, player)
			}
		}
		filteredPlayers = positionFiltered
	}
	
	// Apply role filter
	if filterRole != "" {
		var roleFiltered []Player
		for _, player := range filteredPlayers {
			hasRole := false
			for _, role := range player.RoleSpecificOveralls {
				if strings.Contains(role.RoleName, filterRole) {
					hasRole = true
					break
				}
			}
			if hasRole {
				roleFiltered = append(roleFiltered, player)
			}
		}
		filteredPlayers = roleFiltered
	}
	
	// Apply age filters
	if minAgeStr != "" {
		minAge, err := strconv.Atoi(minAgeStr)
		if err == nil {
			var ageFiltered []Player
			for _, player := range filteredPlayers {
				playerAge, err := strconv.Atoi(player.Age)
				if err == nil && playerAge >= minAge {
					ageFiltered = append(ageFiltered, player)
				}
			}
			filteredPlayers = ageFiltered
		}
	}
	
	if maxAgeStr != "" {
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err == nil {
			var ageFiltered []Player
			for _, player := range filteredPlayers {
				playerAge, err := strconv.Atoi(player.Age)
				if err == nil && playerAge <= maxAge {
					ageFiltered = append(ageFiltered, player)
				}
			}
			filteredPlayers = ageFiltered
		}
	}
	
	// Apply transfer value filters
	if minTransferValueStr != "" {
		minValue, _, _ := ParseMonetaryValueGo(minTransferValueStr)
		var valueFiltered []Player
		for _, player := range filteredPlayers {
			playerValue, _, _ := ParseMonetaryValueGo(player.TransferValue)
			if playerValue >= minValue {
				valueFiltered = append(valueFiltered, player)
			}
		}
		filteredPlayers = valueFiltered
	}
	
	if maxTransferValueStr != "" {
		maxValue, _, _ := ParseMonetaryValueGo(maxTransferValueStr)
		var valueFiltered []Player
		for _, player := range filteredPlayers {
			playerValue, _, _ := ParseMonetaryValueGo(player.TransferValue)
			if playerValue <= maxValue {
				valueFiltered = append(valueFiltered, player)
			}
		}
		filteredPlayers = valueFiltered
	}
	
	// Apply salary filter
	if maxSalaryStr != "" {
		maxSalary, _, _ := ParseMonetaryValueGo(maxSalaryStr)
		var salaryFiltered []Player
		for _, player := range filteredPlayers {
			playerSalary, _, _ := ParseMonetaryValueGo(player.Wage)
			if playerSalary <= maxSalary {
				salaryFiltered = append(salaryFiltered, player)
			}
		}
		filteredPlayers = salaryFiltered
	}
	
	// Log filter results
	SetSpanAttributes(ctx,
		attribute.Int("filters.result_count", len(filteredPlayers)),
		attribute.String("filters.position", filterPosition),
		attribute.String("filters.role", filterRole),
		attribute.String("filters.min_age", minAgeStr),
		attribute.String("filters.max_age", maxAgeStr),
		attribute.String("filters.division_filter", string(divisionFilter)),
	)
	
	return filteredPlayers
}