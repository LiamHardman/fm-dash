// src/api/services/search_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

// SearchResult represents a search result item
type SearchResult struct {
	Type      string      `json:"type"` // "player", "team", "league", "nation"
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	SubText   string      `json:"subText"` // Additional context (position, league, etc.)
	Overall   int         `json:"overall,omitempty"`
	Data      interface{} `json:"data"`      // Full object data
	Relevance float64     `json:"relevance"` // Search relevance score
}

// SearchService handles search functionality
type SearchService struct {
	playerService *PlayerService
}

// NewSearchService creates a new search service
func NewSearchService(playerService *PlayerService) *SearchService {
	return &SearchService{
		playerService: playerService,
	}
}

// SearchAll performs a comprehensive search across all data types
func (s *SearchService) SearchAll(ctx context.Context, datasetID, query string, maxResults int) ([]SearchResult, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}

	if query == "" {
		return []SearchResult{}, nil
	}

	// Get player data for the dataset
	players, _, err := s.playerService.GetPlayersByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players for search: %w", err)
	}

	var allResults []SearchResult

	// Search players
	playerResults := s.searchPlayers(players, query)
	allResults = append(allResults, playerResults...)

	// Search teams/clubs
	teamResults := s.searchTeams(players, query)
	allResults = append(allResults, teamResults...)

	// Search leagues (extracted from club data if available)
	leagueResults := s.searchLeagues(players, query)
	allResults = append(allResults, leagueResults...)

	// Search nationalities
	nationResults := s.searchNationalities(players, query)
	allResults = append(allResults, nationResults...)

	// Sort by relevance (highest first)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Relevance > allResults[j].Relevance
	})

	// Limit results
	if maxResults > 0 && len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	log.Printf("Search for '%s' in dataset %s returned %d results", query, datasetID, len(allResults))
	return allResults, nil
}

// searchPlayers searches for players by name
func (s *SearchService) searchPlayers(players []Player, query string) []SearchResult {
	var results []SearchResult
	lowerQuery := strings.ToLower(query)

	for i := range players {
		relevance := s.calculatePlayerRelevance(&players[i], lowerQuery)
		if relevance > 0 {
			// Use the player's UID as the ID since it's the unique identifier
			playerID := strconv.FormatInt(players[i].UID, 10)
			if players[i].UID == 0 {
				// Fallback to composite ID if UID is somehow missing
				playerID = fmt.Sprintf("player_%s_%s_%d_%s",
					strings.ReplaceAll(players[i].Name, " ", "_"),
					strings.ReplaceAll(players[i].Club, " ", "_"),
					players[i].Age,
					strings.ReplaceAll(players[i].Position, " ", "_"))
			}

			result := SearchResult{
				Type:      "player",
				ID:        playerID,
				Name:      players[i].Name,
				SubText:   fmt.Sprintf("%s - %s (%d OVR)", players[i].Position, players[i].Club, players[i].Overall),
				Overall:   players[i].Overall,
				Data:      players[i],
				Relevance: relevance,
			}
			results = append(results, result)
		}
	}

	return results
}

// searchTeams searches for teams/clubs
func (s *SearchService) searchTeams(players []Player, query string) []SearchResult {
	lowerQuery := strings.ToLower(query)
	teamMap := make(map[string][]Player)

	// Group players by club
	for i := range players {
		if players[i].Club != "" {
			teamMap[players[i].Club] = append(teamMap[players[i].Club], players[i])
		}
	}

	var results []SearchResult
	for teamName, teamPlayers := range teamMap {
		relevance := s.calculateStringRelevance(teamName, lowerQuery)
		if relevance > 0 {
			// Calculate team statistics
			totalOverall := 0
			bestPlayer := ""
			maxOverall := 0

			for i := range teamPlayers {
				totalOverall += teamPlayers[i].Overall
				if teamPlayers[i].Overall > maxOverall {
					maxOverall = teamPlayers[i].Overall
					bestPlayer = teamPlayers[i].Name
				}
			}

			avgOverall := totalOverall / len(teamPlayers)

			result := SearchResult{
				Type:      "team",
				ID:        fmt.Sprintf("team_%s", strings.ReplaceAll(teamName, " ", "_")),
				Name:      teamName,
				SubText:   fmt.Sprintf("%d players, Best: %s (%d)", len(teamPlayers), bestPlayer, maxOverall),
				Overall:   avgOverall,
				Data:      teamPlayers,
				Relevance: relevance,
			}
			results = append(results, result)
		}
	}

	return results
}

// searchLeagues searches for leagues (if available in data)
func (s *SearchService) searchLeagues(players []Player, query string) []SearchResult {
	// This is a placeholder - leagues would need to be extracted from player data
	// or stored separately. For now, we'll return empty results.
	return []SearchResult{}
}

// searchNationalities searches for nationalities
func (s *SearchService) searchNationalities(players []Player, query string) []SearchResult {
	lowerQuery := strings.ToLower(query)
	nationalityMap := make(map[string][]Player)

	// Group players by nationality
	for i := range players {
		if players[i].Nationality != "" {
			nationalityMap[players[i].Nationality] = append(nationalityMap[players[i].Nationality], players[i])
		}
	}

	var results []SearchResult
	for nationality, nationalPlayers := range nationalityMap {
		relevance := s.calculateStringRelevance(nationality, lowerQuery)
		if relevance > 0 {
			// Calculate nationality statistics
			totalOverall := 0
			bestPlayer := ""
			maxOverall := 0

			for i := range nationalPlayers {
				totalOverall += nationalPlayers[i].Overall
				if nationalPlayers[i].Overall > maxOverall {
					maxOverall = nationalPlayers[i].Overall
					bestPlayer = nationalPlayers[i].Name
				}
			}

			avgOverall := totalOverall / len(nationalPlayers)

			result := SearchResult{
				Type:      "nation",
				ID:        fmt.Sprintf("nation_%s", strings.ReplaceAll(nationality, " ", "_")),
				Name:      nationality,
				SubText:   fmt.Sprintf("%d players, Best: %s (%d)", len(nationalPlayers), bestPlayer, maxOverall),
				Overall:   avgOverall,
				Data:      nationalPlayers,
				Relevance: relevance,
			}
			results = append(results, result)
		}
	}

	return results
}

// calculatePlayerRelevance calculates how relevant a player is to the search query
func (s *SearchService) calculatePlayerRelevance(player *Player, lowerQuery string) float64 {
	lowerName := strings.ToLower(player.Name)
	lowerClub := strings.ToLower(player.Club)
	lowerPosition := strings.ToLower(player.Position)
	lowerNationality := strings.ToLower(player.Nationality)

	var relevance float64

	// Use switch statement instead of if-else chain
	switch {
	case lowerName == lowerQuery:
		relevance += 100.0
	case strings.HasPrefix(lowerName, lowerQuery):
		relevance += 80.0
	case strings.Contains(lowerName, lowerQuery):
		relevance += 60.0
	}

	// Club matches
	if strings.Contains(lowerClub, lowerQuery) {
		relevance += 20.0
	}

	// Position matches
	if strings.Contains(lowerPosition, lowerQuery) {
		relevance += 15.0
	}

	// Nationality matches
	if strings.Contains(lowerNationality, lowerQuery) {
		relevance += 10.0
	}

	return relevance
}

// calculateStringRelevance calculates how relevant a string is to the search query
func (s *SearchService) calculateStringRelevance(text, query string) float64 {
	lowerText := strings.ToLower(text)

	if lowerText == query {
		return 100.0
	}
	if strings.HasPrefix(lowerText, query) {
		return 80.0
	}
	if strings.Contains(lowerText, query) {
		return 50.0
	}
	return 0
}
