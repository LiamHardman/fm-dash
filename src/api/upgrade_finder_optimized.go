package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// OptimizedFindPlayerUpgrades uses parallel processing and early filtering for much better performance
func OptimizedFindPlayerUpgrades(players []Player, teamPlayers []Player, req UpgradeFinderRequest) []Player {
	// Debug logging
	LogDebug("OptimizedFindPlayerUpgrades starting - players: %d, teamPlayers: %d, team: %s, position: %s, minOverall: %d",
		len(players), len(teamPlayers), req.Team, req.Position, req.MinOverall)

	// Early filtering - create team exclusion set
	teamPlayerMap := make(map[string]bool)
	if req.TeamDatasetID != "" && req.TeamDatasetID != req.DatasetID {
		LogDebug("Using separate team dataset for filtering")
		for _, teamPlayer := range teamPlayers {
			if teamPlayer.Club == req.Team {
				playerKey := teamPlayer.Name + "|" + teamPlayer.Position
				teamPlayerMap[playerKey] = true
			}
		}
		LogDebug("Found %d team players to exclude", len(teamPlayerMap))
	}

	// Pre-filter players in parallel chunks for better performance
	candidatesChannel := make(chan []Player, 10)

	// Calculate optimal chunk size for parallel processing
	numWorkers := Min(8, len(players)/1000+1)
	chunkSize := (len(players) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup

	// Launch parallel filtering workers
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := min(start+chunkSize, len(players))

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var candidates []Player
			for j := start; j < end; j++ {
				player := &players[j]

				// Early rejection filters (cheapest first)
				if !passesEarlyFilters(player, req, teamPlayerMap) {
					continue
				}

				// More expensive filters
				if !passesAdvancedFilters(player, req) {
					continue
				}

				candidates = append(candidates, *player)
			}

			if len(candidates) > 0 {
				candidatesChannel <- candidates
			}
		}(start, end)
	}

	// Close channel when all workers are done
	go func() {
		wg.Wait()
		close(candidatesChannel)
	}()

	// Collect results from all workers
	var allCandidates []Player
	for candidates := range candidatesChannel {
		allCandidates = append(allCandidates, candidates...)
	}

	// Sort by role-specific overall (descending) with cached calculations
	sortPlayersByRole(allCandidates, req.Role, req.Position)

	LogDebug("OptimizedFindPlayerUpgrades completed - found %d upgrades", len(allCandidates))
	if len(allCandidates) > 0 {
		// Log top 5 upgrades with their role-specific ratings for verification
		for i := 0; i < Min(5, len(allCandidates)); i++ {
			roleOverall := getCachedPlayerOverallForRole(allCandidates[i], req.Role, req.Position)
			LogDebug("Upgrade %d: %s (Club: %s, Role Overall: %d, Main Overall: %d)",
				i+1, allCandidates[i].Name, allCandidates[i].Club, roleOverall, allCandidates[i].Overall)
		}
	}

	return allCandidates
}

// passesEarlyFilters applies the cheapest filters first to quickly eliminate players
func passesEarlyFilters(player *Player, req UpgradeFinderRequest, teamPlayerMap map[string]bool) bool {
	// Skip players from the same team (fastest check)
	if req.Team != "" && player.Club == req.Team {
		return false
	}

	// Skip players already in team (if using separate datasets)
	if len(teamPlayerMap) > 0 {
		playerKey := player.Name + "|" + player.Position
		if teamPlayerMap[playerKey] {
			return false
		}
	}

	// Quick position check with caching
	if !CachedMatchesPositionForUpgrade(*player, req.Position) {
		return false
	}

	// Quick overall check (use main overall as approximation for early filtering)
	if player.Overall < req.MinOverall-5 { // Allow some tolerance for role-specific calculations
		return false
	}

	return true
}

// passesAdvancedFilters applies more expensive filters after early filtering
func passesAdvancedFilters(player *Player, req UpgradeFinderRequest) bool {
	// Precise role-based overall check
	playerOverall := getPlayerOverallForRole(*player, req.Role, req.Position)
	if playerOverall < req.MinOverall {
		return false
	}

	// Age filter
	if req.MaxAge > 0 {
		if playerAge, err := strconv.Atoi(player.Age); err == nil {
			if playerAge > req.MaxAge {
				return false
			}
		}
	}

	// Transfer value filter
	if req.MaxTransferValue > 0 {
		if player.TransferValue == "Not for Sale" ||
			strings.Contains(strings.ToLower(player.TransferValue), "not for sale") {
			return false
		}
		if player.TransferValueAmount > req.MaxTransferValue {
			return false
		}
	}

	// Salary filter
	if req.MaxSalary > 0 && player.WageAmount > req.MaxSalary {
		return false
	}

	// Attribute filters (batch check for efficiency)
	return passesAttributeFilters(player, req)
}

// passesAttributeFilters checks all attribute filters efficiently
func passesAttributeFilters(player *Player, req UpgradeFinderRequest) bool {
	// Use a slice of checks for better CPU cache performance
	attributeChecks := []struct {
		required int
		actual   int
	}{
		{req.MinPAC, player.PAC},
		{req.MinDRI, player.DRI},
		{req.MinSHO, player.SHO},
		{req.MinPAS, player.PAS},
		{req.MinDEF, player.DEF},
		{req.MinPHY, player.PHY},
		{req.MinGK, player.GK},
		{req.MinDIV, player.DIV},
		{req.MinHAN, player.HAN},
		{req.MinREF, player.REF},
		{req.MinKIC, player.KIC},
		{req.MinSPD, player.SPD},
		{req.MinPOS, player.POS},
	}

	for _, check := range attributeChecks {
		if check.required > 0 && check.actual < check.required {
			return false
		}
	}

	return true
}

// roleOverallCache caches role calculations to avoid repeated computation
var roleOverallCache = struct {
	sync.RWMutex
	cache map[string]int
}{cache: make(map[string]int)}

// getCachedPlayerOverallForRole gets player overall with caching
func getCachedPlayerOverallForRole(player Player, role, position string) int {
	// Create cache key
	key := strconv.FormatInt(player.UID, 10) + "|" + role + "|" + position

	// Try to get from cache
	roleOverallCache.RLock()
	if cached, exists := roleOverallCache.cache[key]; exists {
		roleOverallCache.RUnlock()
		return cached
	}
	roleOverallCache.RUnlock()

	// Calculate and cache
	overall := getPlayerOverallForRole(player, role, position)

	roleOverallCache.Lock()
	// Prevent cache from growing too large
	if len(roleOverallCache.cache) > 10000 {
		// Clear cache if it gets too large (simple eviction)
		roleOverallCache.cache = make(map[string]int)
	}
	roleOverallCache.cache[key] = overall
	roleOverallCache.Unlock()

	return overall
}

// sortPlayersByRole sorts players by role-specific overall with efficient caching
func sortPlayersByRole(players []Player, role, position string) {
	// Sort by role-specific overall (descending) - same logic as original but with caching
	sort.Slice(players, func(i, j int) bool {
		overallI := getCachedPlayerOverallForRole(players[i], role, position)
		overallJ := getCachedPlayerOverallForRole(players[j], role, position)
		return overallI > overallJ
	})
}

// ClearRoleOverallCache clears the role calculation cache (call periodically to prevent memory leaks)
func ClearRoleOverallCache() {
	roleOverallCache.Lock()
	roleOverallCache.cache = make(map[string]int)
	roleOverallCache.Unlock()
}
