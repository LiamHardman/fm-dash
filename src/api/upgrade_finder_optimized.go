package main

import (
	"container/heap"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PlayerCandidate represents a player candidate with pre-calculated role overall
type PlayerCandidate struct {
	Player      *Player // Use pointer to avoid copying mutex
	RoleOverall int
}

// TopKPlayerHeap implements a min-heap for efficient top-K player selection
type TopKPlayerHeap []PlayerCandidate

func (h TopKPlayerHeap) Len() int           { return len(h) }
func (h TopKPlayerHeap) Less(i, j int) bool { return h[i].RoleOverall < h[j].RoleOverall } // Min-heap
func (h TopKPlayerHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TopKPlayerHeap) Push(x interface{}) {
	*h = append(*h, x.(PlayerCandidate))
}

func (h *TopKPlayerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// RoleOverallIndex pre-computes role-specific overalls for faster lookup
type RoleOverallIndex struct {
	data        map[int64]map[string]int // playerUID -> roleName -> overall
	mu          sync.RWMutex
	lastBuilt   time.Time
	playerCount int // Track if we need to rebuild
}

func NewRoleOverallIndex() *RoleOverallIndex {
	return &RoleOverallIndex{
		data: make(map[int64]map[string]int),
	}
}

func (idx *RoleOverallIndex) BuildIndex(players []Player) {
	// Check if we need to rebuild (avoid unnecessary rebuilds)
	idx.mu.RLock()
	if len(idx.data) > 0 && idx.playerCount == len(players) && time.Since(idx.lastBuilt) < 10*time.Minute {
		idx.mu.RUnlock()
		return // Index is fresh
	}
	idx.mu.RUnlock()

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Double-check after acquiring write lock
	if len(idx.data) > 0 && idx.playerCount == len(players) && time.Since(idx.lastBuilt) < 10*time.Minute {
		return
	}

	// Clear existing data
	idx.data = make(map[int64]map[string]int)

	for i := range players {
		playerRoles := make(map[string]int)
		for _, roleOverall := range players[i].RoleSpecificOveralls {
			playerRoles[roleOverall.RoleName] = roleOverall.Score
		}
		idx.data[players[i].UID] = playerRoles
	}

	idx.playerCount = len(players)
	idx.lastBuilt = time.Now()
}

func (idx *RoleOverallIndex) GetRoleOverall(playerUID int64, role, position string) int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	playerRoles, exists := idx.data[playerUID]
	if !exists {
		return 0
	}

	// If role is specified, try to get role-specific overall
	if role != "" {
		if overall, found := playerRoles[role]; found {
			return overall
		}
	}

	// If position is specified, find the best role-specific overall for that position
	if position != "" {
		bestOverall := 0
		for roleName, overall := range playerRoles {
			if strings.HasPrefix(roleName, position+" - ") {
				if overall > bestOverall {
					bestOverall = overall
				}
			}
		}
		if bestOverall > 0 {
			return bestOverall
		}
	}

	return 0 // Will fall back to main overall in calling code
}

// Global role overall index
var globalRoleIndex = NewRoleOverallIndex()

// OptimizedFindPlayerUpgrades uses parallel processing, pre-indexing, and top-K heap for much better performance
func OptimizedFindPlayerUpgrades(players []Player, teamPlayers []Player, req UpgradeFinderRequest) []Player {
	// Debug logging
	LogDebug("OptimizedFindPlayerUpgrades starting - players: %d, teamPlayers: %d, team: %s, position: %s, minOverall: %d",
		len(players), len(teamPlayers), req.Team, req.Position, req.MinOverall)

	// Build role overall index for fast lookups (only if not already built)
	buildStart := time.Now()
	globalRoleIndex.BuildIndex(players)
	LogDebug("Role index built in %v", time.Since(buildStart))

	// Early filtering - create team exclusion set
	teamPlayerMap := make(map[string]bool)
	if req.TeamDatasetID != "" && req.TeamDatasetID != req.DatasetID {
		LogDebug("Using separate team dataset for filtering")
		for i := range teamPlayers {
			if teamPlayers[i].Club == req.Team {
				playerKey := teamPlayers[i].Name + "|" + teamPlayers[i].Position
				teamPlayerMap[playerKey] = true
			}
		}
		LogDebug("Found %d team players to exclude", len(teamPlayerMap))
	}

	// Use heap for top-K selection instead of sorting all candidates
	const maxResults = 1000 // Limit results for better performance
	candidatesHeap := make(TopKPlayerHeap, 0, maxResults)
	heap.Init(&candidatesHeap)

	// Mutex to protect heap from concurrent access
	var heapMutex sync.Mutex
	candidatesFound := 0

	// Calculate optimal chunk size for parallel processing
	numWorkers := Min(8, len(players)/1000+1)
	chunkSize := (len(players) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup

	// Launch parallel filtering workers with early termination
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := min(start+chunkSize, len(players))

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			localCandidates := make([]PlayerCandidate, 0, 100)

			for j := start; j < end; j++ {
				// Early termination if we have enough candidates
				heapMutex.Lock()
				found := candidatesFound
				heapMutex.Unlock()

				if found > maxResults*2 { // Allow some buffer for better results
					break
				}

				player := &players[j]

				// Early rejection filters (cheapest first)
				if !passesEarlyFiltersOptimized(player, req, teamPlayerMap) {
					continue
				}

				// Calculate role overall using index (much faster)
				roleOverall := globalRoleIndex.GetRoleOverall(player.UID, req.Role, req.Position)
				if roleOverall == 0 {
					roleOverall = player.Overall // Fallback to main overall
				}

				// Check minimum overall requirement
				if roleOverall < req.MinOverall {
					continue
				}

				// More expensive filters only after role check passes
				if !passesAdvancedFiltersOptimized(player, req, roleOverall) {
					continue
				}

				localCandidates = append(localCandidates, PlayerCandidate{
					Player:      player,
					RoleOverall: roleOverall,
				})
			}

			// Add local candidates to heap in batch
			if len(localCandidates) > 0 {
				heapMutex.Lock()
				for _, candidate := range localCandidates {
					if candidatesHeap.Len() < maxResults {
						heap.Push(&candidatesHeap, candidate)
					} else if candidate.RoleOverall > candidatesHeap[0].RoleOverall {
						// Replace worst candidate with better one
						heap.Pop(&candidatesHeap)
						heap.Push(&candidatesHeap, candidate)
					}
				}
				candidatesFound += len(localCandidates)
				heapMutex.Unlock()
			}
		}(start, end)
	}

	wg.Wait()

	// Convert heap to sorted slice (best to worst)
	result := make([]Player, 0, candidatesHeap.Len())

	// Extract all from heap and sort in descending order
	candidates := make([]PlayerCandidate, 0, candidatesHeap.Len())
	for candidatesHeap.Len() > 0 {
		candidates = append(candidates, heap.Pop(&candidatesHeap).(PlayerCandidate))
	}

	// Reverse to get descending order (best first)
	for i := len(candidates) - 1; i >= 0; i-- {
		result = append(result, *candidates[i].Player)
	}

	LogDebug("OptimizedFindPlayerUpgrades completed - found %d upgrades", len(result))
	if len(result) > 0 {
		// Log top 5 upgrades with their role-specific ratings for verification
		for i := 0; i < Min(5, len(result)); i++ {
			roleOverall := globalRoleIndex.GetRoleOverall(result[i].UID, req.Role, req.Position)
			if roleOverall == 0 {
				roleOverall = result[i].Overall
			}
			LogDebug("Upgrade %d: %s (Club: %s, Role Overall: %d, Main Overall: %d)",
				i+1, result[i].Name, result[i].Club, roleOverall, result[i].Overall)
		}
	}

	return result
}

// passesEarlyFilters applies the cheapest filters first to quickly eliminate players
//
//nolint:unused // Retained as the non-indexed upgrade filter implementation.
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
	if !CachedMatchesPositionForUpgradePtr(player, req.Position) {
		return false
	}

	// Quick overall check (use main overall as approximation for early filtering)
	if player.Overall < req.MinOverall-5 { // Allow some tolerance for role-specific calculations
		return false
	}

	return true
}

// passesAdvancedFilters applies more expensive filters after early filtering
//
//nolint:unused // Retained as the non-indexed upgrade filter implementation.
func passesAdvancedFilters(player *Player, req UpgradeFinderRequest) bool {
	// Precise role-based overall check (use pointer to avoid copying)
	playerOverall := getPlayerOverallForRolePtr(player, req.Role, req.Position)
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

// passesEarlyFiltersOptimized applies the cheapest filters first with optimizations
func passesEarlyFiltersOptimized(player *Player, req UpgradeFinderRequest, teamPlayerMap map[string]bool) bool {
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
	if !CachedMatchesPositionForUpgradePtr(player, req.Position) {
		return false
	}

	// Quick overall check (use main overall as approximation for early filtering)
	if player.Overall < req.MinOverall-10 { // Increased tolerance since we'll check role overall later
		return false
	}

	return true
}

// passesAdvancedFiltersOptimized applies more expensive filters with pre-calculated role overall
func passesAdvancedFiltersOptimized(player *Player, req UpgradeFinderRequest, roleOverall int) bool {
	// Role overall is already calculated and checked in main function

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

// roleOverallCache caches role calculations to avoid repeated computation
var roleOverallCache = struct {
	sync.RWMutex
	cache map[string]int
}{cache: make(map[string]int)}

// getCachedPlayerOverallForRole gets player overall with caching
//
//nolint:unused // Retained for non-pointer role scoring callers.
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
//
//nolint:unused // Retained for callers that sort full player values instead of heap candidates.
func sortPlayersByRole(players []Player, role, position string) {
	// Sort by role-specific overall (descending) - same logic as original but with caching
	sort.Slice(players, func(i, j int) bool {
		overallI := getCachedPlayerOverallForRolePtr(&players[i], role, position)
		overallJ := getCachedPlayerOverallForRolePtr(&players[j], role, position)
		return overallI > overallJ
	})
}

// ClearRoleOverallCache clears the role calculation cache (call periodically to prevent memory leaks)
func ClearRoleOverallCache() {
	roleOverallCache.Lock()
	roleOverallCache.cache = make(map[string]int)
	roleOverallCache.Unlock()
}

// InvalidateRoleIndex clears the global role index (call when datasets change)
func InvalidateRoleIndex() {
	globalRoleIndex.mu.Lock()
	globalRoleIndex.data = make(map[int64]map[string]int)
	globalRoleIndex.playerCount = 0
	globalRoleIndex.lastBuilt = time.Time{}
	globalRoleIndex.mu.Unlock()
	LogDebug("Role index invalidated")
}

// CachedMatchesPositionForUpgradePtr works with pointer to avoid copying mutex
// This function still copies but only for the final position check which is acceptable
func CachedMatchesPositionForUpgradePtr(player *Player, position string) bool {
	// Create cache key directly from pointer data
	key := player.Position + "|" + position

	// Try to get from cache
	positionMatchCache.mutex.RLock()
	if cached, exists := positionMatchCache.cache[key]; exists {
		positionMatchCache.mutex.RUnlock()
		return cached
	}
	positionMatchCache.mutex.RUnlock()

	// Calculate using pointer-safe method
	matches := matchesPositionForUpgradePtr(player, position)

	positionMatchCache.mutex.Lock()
	// Prevent cache from growing too large
	if len(positionMatchCache.cache) > 5000 {
		positionMatchCache.cache = make(map[string]bool)
	}
	positionMatchCache.cache[key] = matches
	positionMatchCache.mutex.Unlock()

	return matches
}

// matchesPositionForUpgradePtr is a pointer-safe version of position matching
func matchesPositionForUpgradePtr(player *Player, position string) bool {
	// Check short positions first
	for _, pos := range player.ShortPositions {
		if pos == position {
			return true
		}
	}

	// Check parsed positions
	for _, pos := range player.ParsedPositions {
		if pos == position {
			return true
		}
	}

	return false
}

// getPlayerOverallForRolePtr works with pointer and uses the global index when possible
//
//nolint:unused // Retained by the non-indexed advanced filter path.
func getPlayerOverallForRolePtr(player *Player, role, position string) int {
	// Try to use the global index first (much faster)
	if roleOverall := globalRoleIndex.GetRoleOverall(player.UID, role, position); roleOverall > 0 {
		return roleOverall
	}

	// Fallback to direct calculation if not in index
	return player.Overall
}

// getCachedPlayerOverallForRolePtr uses the global index for efficiency
//
//nolint:unused // Retained by the full-slice role sorting path.
func getCachedPlayerOverallForRolePtr(player *Player, role, position string) int {
	// Use the global index which is already optimized
	if roleOverall := globalRoleIndex.GetRoleOverall(player.UID, role, position); roleOverall > 0 {
		return roleOverall
	}

	// Fallback to main overall
	return player.Overall
}
