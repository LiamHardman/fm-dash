package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// HybridSearchService combines fast index-based search with parallel processing for large result sets
type HybridSearchService struct {
	fastSearch     *FastSearchService
	parallelSearch *OptimizedPlayerSearch
}

// NewHybridSearchService creates a new hybrid search service
func NewHybridSearchService() *HybridSearchService {
	return &HybridSearchService{
		fastSearch:     NewFastSearchService(),
		parallelSearch: NewOptimizedPlayerSearch(),
	}
}

// Search performs optimized search using the best strategy based on dataset size and query
func (hss *HybridSearchService) Search(ctx context.Context, datasetID, query string, maxResults int) ([]SearchResult, error) {
	searchStart := time.Now()

	// For most searches, use the fast index-based search
	results, err := hss.fastSearch.Search(ctx, datasetID, query, maxResults)
	searchTime := time.Since(searchStart)

	logInfo(ctx, "Hybrid search completed",
		"dataset_id", datasetID,
		"query", query,
		"results_count", len(results),
		"search_time_ms", searchTime.Milliseconds(),
		"search_method", "fast_index")

	if err != nil {
		return nil, err
	}

	// If we have too few results and the query is very specific,
	// fall back to parallel search for comprehensive coverage
	if len(results) < maxResults/4 && len(query) >= 4 {
		return hss.comprehensiveSearch(ctx, datasetID, query, maxResults)
	}

	return results, nil
}

// comprehensiveSearch performs a thorough search using parallel processing
func (hss *HybridSearchService) comprehensiveSearch(ctx context.Context, datasetID, query string, maxResults int) ([]SearchResult, error) {
	// Get player data
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil, fmt.Errorf("dataset not found: %s", datasetID)
	}

	// Use parallel search for comprehensive results
	searchFields := []string{"name", "club", "position", "nationality", "division"}
	playerMatches := hss.parallelSearch.ParallelSearchPlayers(players, query, searchFields)

	// Convert to search results
	var results []SearchResult
	for i := range playerMatches {
		if len(results) >= maxResults {
			break
		}

		player := &playerMatches[i]
		results = append(results, SearchResult{
			ID:          player.Name,
			Name:        player.Name,
			Type:        "player",
			Description: formatPlayerDescriptionFromPtr(player),
			URL:         formatPlayerURLFromPtr(player),
			Overall:     player.Overall,
		})
	}

	// Add team/league/nation results if we have room
	if len(results) < maxResults {
		additionalResults := hss.searchNonPlayerEntities(players, query, maxResults-len(results))
		results = append(results, additionalResults...)
	}

	return results, nil
}

// searchNonPlayerEntities searches for teams, leagues, and nations
func (hss *HybridSearchService) searchNonPlayerEntities(players []Player, query string, maxResults int) []SearchResult {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	// Collect unique entities
	teams := make(map[string]TeamInfo)
	leagues := make(map[string]LeagueInfo)
	nations := make(map[string]NationInfo)

	for i := range players {
		player := &players[i]
		// Collect teams
		if player.Club != "" {
			if team, exists := teams[player.Club]; exists {
				team.Players = append(team.Players, i)
				teams[player.Club] = team
			} else {
				teams[player.Club] = TeamInfo{
					Name:     player.Club,
					Division: player.Division,
					Players:  []int{i},
				}
			}
		}

		// Collect leagues
		if player.Division != "" {
			if league, exists := leagues[player.Division]; exists {
				league.Players = append(league.Players, i)
				leagues[player.Division] = league
			} else {
				leagues[player.Division] = LeagueInfo{
					Name:    player.Division,
					Players: []int{i},
				}
			}
		}

		// Collect nations
		if player.Nationality != "" {
			if nation, exists := nations[player.Nationality]; exists {
				nation.Players = append(nation.Players, i)
				nations[player.Nationality] = nation
			} else {
				nations[player.Nationality] = NationInfo{
					Name:    player.Nationality,
					Players: []int{i},
				}
			}
		}
	}

	// Search teams
	for teamName, team := range teams {
		if len(results) >= maxResults {
			break
		}
		if strings.Contains(strings.ToLower(teamName), queryLower) {
			results = append(results, SearchResult{
				ID:          teamName,
				Name:        teamName,
				Type:        "team",
				Description: formatTeamDescription(team),
				URL:         formatTeamURL(teamName),
				Overall:     0,
			})
		}
	}

	// Search leagues
	for leagueName, league := range leagues {
		if len(results) >= maxResults {
			break
		}
		if strings.Contains(strings.ToLower(leagueName), queryLower) {
			results = append(results, SearchResult{
				ID:          leagueName,
				Name:        leagueName,
				Type:        "league",
				Description: formatLeagueDescription(league),
				URL:         formatLeagueURL(leagueName),
				Overall:     0,
			})
		}
	}

	// Search nations
	for nationName, nation := range nations {
		if len(results) >= maxResults {
			break
		}
		if matchesNationality(nationName, queryLower) {
			results = append(results, SearchResult{
				ID:          nationName,
				Name:        nationName,
				Type:        "nation",
				Description: formatNationDescription(nation),
				URL:         formatNationURL(nationName),
				Overall:     0,
			})
		}
	}

	return results
}

// Helper formatting functions (matching the search_index.go implementations)
func formatPlayerDescriptionFromPtr(player *Player) string {
	return player.Club + " • " + player.Division + " (" + formatOverall(player.Overall) + " OVR)"
}

func formatPlayerURLFromPtr(player *Player) string {
	return "/dataset/?search=" + player.Name
}

func formatTeamDescription(team TeamInfo) string {
	return team.Division + " • " + formatInt(len(team.Players)) + " players"
}

func formatTeamURL(teamName string) string {
	return "/dataset/?team=" + teamName
}

func formatLeagueDescription(league LeagueInfo) string {
	return formatInt(len(league.Players)) + " players"
}

func formatLeagueURL(leagueName string) string {
	return "/leagues?league=" + leagueName
}

func formatNationDescription(nation NationInfo) string {
	return formatInt(len(nation.Players)) + " players"
}

func formatNationURL(nationName string) string {
	return "/nations?nation=" + nationName
}

func matchesNationality(nationName, query string) bool {
	nationLower := strings.ToLower(nationName)

	// Direct substring match
	if strings.Contains(nationLower, query) {
		return true
	}

	// Common nationality variations
	switch query {
	case "fra", "france":
		return strings.Contains(nationLower, "fran") || strings.Contains(nationLower, "french")
	case "eng", "england":
		return strings.Contains(nationLower, "eng") || strings.Contains(nationLower, "british")
	case "ger", "germany":
		return strings.Contains(nationLower, "ger") || strings.Contains(nationLower, "deutsch")
	case "spa", "spain":
		return strings.Contains(nationLower, "spa") || strings.Contains(nationLower, "spanish")
	case "ita", "italy":
		return strings.Contains(nationLower, "ita") || strings.Contains(nationLower, "italian")
	case "por", "portugal":
		return strings.Contains(nationLower, "por") || strings.Contains(nationLower, "portuguese")
	case "bra", "brazil":
		return strings.Contains(nationLower, "bra") || strings.Contains(nationLower, "brazilian")
	case "arg", "argentina":
		return strings.Contains(nationLower, "arg") || strings.Contains(nationLower, "argentine")
	case "net", "netherlands":
		return strings.Contains(nationLower, "net") || strings.Contains(nationLower, "dutch")
	case "bel", "belgium":
		return strings.Contains(nationLower, "bel") || strings.Contains(nationLower, "belgian")
	default:
		// For 3-letter queries, also try prefix matching
		if len(query) == 3 && len(nationLower) >= 3 {
			return strings.HasPrefix(nationLower, query)
		}
	}

	return false
}

// InvalidateIndex invalidates the search index for a dataset
func (hss *HybridSearchService) InvalidateIndex(datasetID string) {
	hss.fastSearch.InvalidateIndex(datasetID)
	hss.parallelSearch.ClearCache()
}

// GetSearchStats returns statistics about the search performance
func (hss *HybridSearchService) GetSearchStats() map[string]interface{} {
	stats := hss.fastSearch.GetIndexStats()
	stats["parallel_cache_size"] = len(hss.parallelSearch.cache)

	return stats
}

// Global hybrid search service instance
var globalHybridSearchService = NewHybridSearchService()

// GetHybridSearchService returns the global hybrid search service
func GetHybridSearchService() *HybridSearchService {
	return globalHybridSearchService
}
