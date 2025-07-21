// Package services provides search-related service functionality
package services

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	apperrors "api/errors"
)

// SearchResult represents a search result
type SearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`        // "player", "team", "league", "nation"
	Description string `json:"description"` // Additional context
	URL         string `json:"url"`         // URL to navigate to
	Overall     int    `json:"overall"`     // Include overall rating for sorting
}

// SearchService handles search functionality
type SearchService struct {
	playerService *PlayerService
}

// CreateSearchService creates a new search service
func CreateSearchService(playerService *PlayerService) *SearchService {
	return &SearchService{
		playerService: playerService,
	}
}

// Search performs a search across players, teams, leagues, and nations
func (s *SearchService) Search(ctx context.Context, query string, datasetID string) ([]SearchResult, error) {
	if datasetID == "" {
		return nil, apperrors.ErrDatasetIDEmpty
	}

	if query == "" {
		return []SearchResult{}, nil
	}

	// Get players from the dataset
	players, _, err := s.playerService.GetPlayersByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player data: %w", err)
	}

	// Perform search
	allResults := s.performSearch(players, query)

	// Use global logging function
	fmt.Printf("Search for '%s' in dataset %s returned %d results", query, datasetID, len(allResults))

	return allResults, nil
}

// performSearch performs the actual search logic
func (s *SearchService) performSearch(players []Player, query string) []SearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []SearchResult{}
	}

	var results []SearchResult

	// Search through players
	for _, player := range players {
		if s.matchesPlayer(player, query) {
			results = append(results, SearchResult{
				ID:          fmt.Sprintf("%d", player.UID),
				Name:        player.Name,
				Type:        "player",
				Description: fmt.Sprintf("%s - %s", player.Club, player.Division),
				URL:         fmt.Sprintf("/player/%d", player.UID),
				Overall:     player.Overall,
			})
		}
	}

	// Search through teams (extract unique teams from players)
	teams := make(map[string]bool)
	for _, player := range players {
		if player.Club != "" {
			teams[player.Club] = true
		}
	}

	for team := range teams {
		if s.matchesTeam(team, query) {
			results = append(results, SearchResult{
				ID:          team,
				Name:        team,
				Type:        "team",
				Description: "Team",
				URL:         fmt.Sprintf("/team/%s", team),
				Overall:     0, // Teams don't have overall ratings
			})
		}
	}

	// Search through divisions/leagues (extract unique divisions from players)
	divisions := make(map[string]bool)
	for _, player := range players {
		if player.Division != "" {
			divisions[player.Division] = true
		}
	}

	for division := range divisions {
		if s.matchesDivision(division, query) {
			results = append(results, SearchResult{
				ID:          division,
				Name:        division,
				Type:        "league",
				Description: "League/Division",
				URL:         fmt.Sprintf("/league/%s", division),
				Overall:     0, // Leagues don't have overall ratings
			})
		}
	}

	// Search through nations (extract unique nations from players)
	nations := make(map[string]bool)
	for _, player := range players {
		if player.Nationality != "" {
			nations[player.Nationality] = true
		}
	}

	for nation := range nations {
		if s.matchesNation(nation, query) {
			results = append(results, SearchResult{
				ID:          nation,
				Name:        nation,
				Type:        "nation",
				Description: "Nationality",
				URL:         fmt.Sprintf("/nation/%s", nation),
				Overall:     0, // Nations don't have overall ratings
			})
		}
	}

	return results
}

// matchesPlayer checks if a player matches the search query
func (s *SearchService) matchesPlayer(player Player, query string) bool {
	// Check name (case-insensitive)
	if strings.Contains(strings.ToLower(player.Name), query) {
		return true
	}

	// Check club (case-insensitive)
	if strings.Contains(strings.ToLower(player.Club), query) {
		return true
	}

	// Check position (case-insensitive)
	if strings.Contains(strings.ToLower(player.Position), query) {
		return true
	}

	// Check nationality (case-insensitive)
	if strings.Contains(strings.ToLower(player.Nationality), query) {
		return true
	}

	return false
}

// matchesTeam checks if a team matches the search query
func (s *SearchService) matchesTeam(team, query string) bool {
	return strings.Contains(strings.ToLower(team), query)
}

// matchesDivision checks if a division matches the search query
func (s *SearchService) matchesDivision(division, query string) bool {
	return strings.Contains(strings.ToLower(division), query)
}

// matchesNation checks if a nation matches the search query
func (s *SearchService) matchesNation(nation, query string) bool {
	return strings.Contains(strings.ToLower(nation), query)
}

// normalizeSearchQuery normalizes a search query for better matching
func (s *SearchService) normalizeSearchQuery(query string) string {
	// Convert to lowercase
	query = strings.ToLower(query)

	// Remove extra whitespace
	query = strings.TrimSpace(query)

	// Remove special characters (keep alphanumeric and spaces)
	var result strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}
