package main

import (
	"context"
	"fmt"
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

	var result []Player
	for _, player := range players {
		switch filterType {
		case DivisionFilterSame:
			if player.Division == targetDivision {
				result = append(result, player)
			}
		case DivisionFilterTop5:
			if IsTop5League(player.Division, player.BasedIn) {
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
		attribute.String("filters.division_filter", fmt.Sprintf("%d", divisionFilter)),
	)

	return filteredPlayers
}

// FastFilterAndCopyPlayers filters the stored (shared) players slice by position/role/age/
// transfer value/salary and deep-copies only the survivors, instead of deep-copying the
// whole dataset up front and filtering afterward.
//
// Callers MUST guarantee divisionFilter == DivisionFilterAll for the request, since this
// intentionally skips division filtering (a no-op in that case) and, more importantly, skips
// percentile (re)calculation — percentiles must already be present on the stored data before
// calling this, since they're computed relative to the whole population and cannot be
// correctly computed from a pre-filtered subset. See docs/PERFORMANCE_FIXES_2026-07-08.md #3.
//
// This reads Player fields via pointer (players[i], never `for _, p := range players`) so it
// never copies the Player struct's mu sync.RWMutex by value — that mutex is actively used
// elsewhere (ca_calculation.go, handlers.go, performance_stats.go, player_processing.go) for
// concurrent-safe access to the stored data, so copying it by value would be unsafe. Only
// fastDeepCopyPlayer (pointer receiver, builds a fresh mutex) touches the original structs.
func FastFilterAndCopyPlayers(players []Player, filterPosition, filterRole, minAgeStr, maxAgeStr,
	minTransferValueStr, maxTransferValueStr, maxSalaryStr string) []Player {
	if len(players) == 0 {
		return nil
	}

	minAge, hasMinAge := 0, false
	if minAgeStr != "" {
		if v, err := strconv.Atoi(minAgeStr); err == nil {
			minAge, hasMinAge = v, true
		}
	}
	maxAge, hasMaxAge := 0, false
	if maxAgeStr != "" {
		if v, err := strconv.Atoi(maxAgeStr); err == nil {
			maxAge, hasMaxAge = v, true
		}
	}
	// NOTE: ParseMonetaryValueGo returns (originalDisplay string, numericValue int64,
	// detectedSymbol string). ApplyAllFilters (player_data_handler_utils.go, above) compares
	// the *first* (string) return value, not the numeric one — matched here verbatim for
	// behavioral parity with that function, even though a numeric comparison would arguably
	// be more correct. Not changed as part of this perf pass; see
	// docs/PERFORMANCE_FIXES_2026-07-08.md.
	minValue, hasMinValue := "", false
	if minTransferValueStr != "" {
		minValue, _, _ = ParseMonetaryValueGo(minTransferValueStr)
		hasMinValue = true
	}
	maxValue, hasMaxValue := "", false
	if maxTransferValueStr != "" {
		maxValue, _, _ = ParseMonetaryValueGo(maxTransferValueStr)
		hasMaxValue = true
	}
	maxSalary, hasMaxSalary := "", false
	if maxSalaryStr != "" {
		maxSalary, _, _ = ParseMonetaryValueGo(maxSalaryStr)
		hasMaxSalary = true
	}

	result := make([]Player, 0, len(players))
	for i := range players {
		p := &players[i]

		if filterPosition != "" && !strings.Contains(p.Position, filterPosition) {
			continue
		}

		if filterRole != "" {
			hasRole := false
			for _, role := range p.RoleSpecificOveralls {
				if strings.Contains(role.RoleName, filterRole) {
					hasRole = true
					break
				}
			}
			if !hasRole {
				continue
			}
		}

		if hasMinAge {
			playerAge, err := strconv.Atoi(p.Age)
			if err != nil || playerAge < minAge {
				continue
			}
		}
		if hasMaxAge {
			playerAge, err := strconv.Atoi(p.Age)
			if err != nil || playerAge > maxAge {
				continue
			}
		}

		if hasMinValue {
			playerValue, _, _ := ParseMonetaryValueGo(p.TransferValue)
			if playerValue < minValue {
				continue
			}
		}
		if hasMaxValue {
			playerValue, _, _ := ParseMonetaryValueGo(p.TransferValue)
			if playerValue > maxValue {
				continue
			}
		}

		if hasMaxSalary {
			playerSalary, _, _ := ParseMonetaryValueGo(p.Wage)
			if playerSalary > maxSalary {
				continue
			}
		}

		result = append(result, fastDeepCopyPlayer(p))
	}

	return result
}
