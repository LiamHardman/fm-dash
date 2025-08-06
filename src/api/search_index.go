package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// SearchIndex provides fast text search capabilities using inverted indexes
type SearchIndex struct {
	// Inverted indexes for different fields
	nameIndex        map[string][]int // word -> player indices
	clubIndex        map[string][]int
	positionIndex    map[string][]int
	nationalityIndex map[string][]int
	divisionIndex    map[string][]int

	// Pre-processed player data for quick access
	players []Player

	// Pre-computed lowercase strings to avoid repeated ToLower calls
	playerNames         []string
	playerClubs         []string
	playerPositions     []string
	playerNationalities []string
	playerDivisions     []string

	// Aggregated data for teams, leagues, nations
	teams   map[string]TeamInfo
	leagues map[string]LeagueInfo
	nations map[string]NationInfo

	// Trie structures for prefix matching (for autocomplete/fuzzy search)
	nameTrie        *TrieNode
	clubTrie        *TrieNode
	positionTrie    *TrieNode
	nationalityTrie *TrieNode
	divisionTrie    *TrieNode

	mutex sync.RWMutex
}

// TeamInfo stores aggregated team information
type TeamInfo struct {
	Name     string
	Division string
	Players  []int // player indices
}

// LeagueInfo stores aggregated league information
type LeagueInfo struct {
	Name    string
	Players []int // player indices
	Teams   []string
}

// NationInfo stores aggregated nation information
type NationInfo struct {
	Name    string
	Players []int // player indices
}

// TrieNode represents a node in a trie data structure for prefix matching
type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
	playerIds   []int // Store player indices that match this prefix
}

// NewSearchIndex creates a new search index
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		nameIndex:        make(map[string][]int),
		clubIndex:        make(map[string][]int),
		positionIndex:    make(map[string][]int),
		nationalityIndex: make(map[string][]int),
		divisionIndex:    make(map[string][]int),
		teams:            make(map[string]TeamInfo),
		leagues:          make(map[string]LeagueInfo),
		nations:          make(map[string]NationInfo),
		nameTrie:         NewTrieNode(),
		clubTrie:         NewTrieNode(),
		positionTrie:     NewTrieNode(),
		nationalityTrie:  NewTrieNode(),
		divisionTrie:     NewTrieNode(),
	}
}

// NewTrieNode creates a new trie node
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children:  make(map[rune]*TrieNode),
		playerIds: make([]int, 0),
	}
}

// BuildIndex constructs the search index from player data
func (si *SearchIndex) BuildIndex(players []Player) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	si.players = players
	playerCount := len(players)

	// Pre-allocate arrays for better performance
	si.playerNames = make([]string, playerCount)
	si.playerClubs = make([]string, playerCount)
	si.playerPositions = make([]string, playerCount)
	si.playerNationalities = make([]string, playerCount)
	si.playerDivisions = make([]string, playerCount)

	// Clear existing indexes
	si.clearIndexes()

	// Build indexes for each player
	for i := range players {
		player := &players[i]
		// Pre-compute and store lowercase strings
		si.playerNames[i] = strings.ToLower(player.Name)
		si.playerClubs[i] = strings.ToLower(player.Club)
		si.playerPositions[i] = strings.ToLower(player.Position)
		si.playerNationalities[i] = strings.ToLower(player.Nationality)
		si.playerDivisions[i] = strings.ToLower(player.Division)

		// Index player name
		si.indexField(si.playerNames[i], i, si.nameIndex)
		si.addToTrie(si.nameTrie, si.playerNames[i], i)

		// Index club
		if player.Club != "" {
			si.indexField(si.playerClubs[i], i, si.clubIndex)
			si.addToTrie(si.clubTrie, si.playerClubs[i], i)
		}

		// Index position
		if player.Position != "" {
			si.indexField(si.playerPositions[i], i, si.positionIndex)
			si.addToTrie(si.positionTrie, si.playerPositions[i], i)
		}

		// Index nationality
		if player.Nationality != "" {
			si.indexField(si.playerNationalities[i], i, si.nationalityIndex)
			si.addToTrie(si.nationalityTrie, si.playerNationalities[i], i)
		}

		// Index division
		if player.Division != "" {
			si.indexField(si.playerDivisions[i], i, si.divisionIndex)
			si.addToTrie(si.divisionTrie, si.playerDivisions[i], i)
		}

		// Build aggregated data
		si.buildAggregatedData(player, i)
	}

	// Sort player indices in aggregated data for better performance
	si.sortAggregatedData()
}

// clearIndexes clears all existing indexes
func (si *SearchIndex) clearIndexes() {
	si.nameIndex = make(map[string][]int)
	si.clubIndex = make(map[string][]int)
	si.positionIndex = make(map[string][]int)
	si.nationalityIndex = make(map[string][]int)
	si.divisionIndex = make(map[string][]int)
	si.teams = make(map[string]TeamInfo)
	si.leagues = make(map[string]LeagueInfo)
	si.nations = make(map[string]NationInfo)
	si.nameTrie = NewTrieNode()
	si.clubTrie = NewTrieNode()
	si.positionTrie = NewTrieNode()
	si.nationalityTrie = NewTrieNode()
	si.divisionTrie = NewTrieNode()
}

// indexField adds words from a field to the inverted index
func (si *SearchIndex) indexField(fieldValue string, playerIndex int, index map[string][]int) {
	words := si.extractSearchableWords(fieldValue)
	for _, word := range words {
		if len(word) >= 2 { // Only index words with 2+ characters
			index[word] = append(index[word], playerIndex)
		}
	}
}

// extractSearchableWords extracts searchable words from text
func (si *SearchIndex) extractSearchableWords(text string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	// Add the last word if any
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	// Also add the full text for substring matching
	if len(text) > 0 {
		words = append(words, text)
	}

	return words
}

// addToTrie adds a term to the trie for prefix matching
func (si *SearchIndex) addToTrie(root *TrieNode, term string, playerIndex int) {
	node := root

	for _, r := range term {
		if _, exists := node.children[r]; !exists {
			node.children[r] = NewTrieNode()
		}
		node = node.children[r]
		node.playerIds = append(node.playerIds, playerIndex)
	}

	node.isEndOfWord = true
}

// buildAggregatedData builds team, league, and nation information
func (si *SearchIndex) buildAggregatedData(player *Player, playerIndex int) {
	// Build team data
	if player.Club != "" {
		if team, exists := si.teams[player.Club]; exists {
			team.Players = append(team.Players, playerIndex)
			si.teams[player.Club] = team
		} else {
			si.teams[player.Club] = TeamInfo{
				Name:     player.Club,
				Division: player.Division,
				Players:  []int{playerIndex},
			}
		}
	}

	// Build league data
	if player.Division != "" {
		if league, exists := si.leagues[player.Division]; exists {
			league.Players = append(league.Players, playerIndex)
			if player.Club != "" {
				// Add team to league if not already present
				teamExists := false
				for _, team := range league.Teams {
					if team == player.Club {
						teamExists = true
						break
					}
				}
				if !teamExists {
					league.Teams = append(league.Teams, player.Club)
				}
			}
			si.leagues[player.Division] = league
		} else {
			teams := make([]string, 0, 1)
			if player.Club != "" {
				teams = append(teams, player.Club)
			}
			si.leagues[player.Division] = LeagueInfo{
				Name:    player.Division,
				Players: []int{playerIndex},
				Teams:   teams,
			}
		}
	}

	// Build nation data
	if player.Nationality != "" {
		if nation, exists := si.nations[player.Nationality]; exists {
			nation.Players = append(nation.Players, playerIndex)
			si.nations[player.Nationality] = nation
		} else {
			si.nations[player.Nationality] = NationInfo{
				Name:    player.Nationality,
				Players: []int{playerIndex},
			}
		}
	}
}

// sortAggregatedData sorts player indices in aggregated data for better performance
func (si *SearchIndex) sortAggregatedData() {
	for teamName, team := range si.teams {
		sort.Ints(team.Players)
		si.teams[teamName] = team
	}

	for leagueName, league := range si.leagues {
		sort.Ints(league.Players)
		sort.Strings(league.Teams)
		si.leagues[leagueName] = league
	}

	for nationName, nation := range si.nations {
		sort.Ints(nation.Players)
		si.nations[nationName] = nation
	}
}

// FastSearch performs optimized search using the built indexes
func (si *SearchIndex) FastSearch(query string, maxResults int) []SearchResult {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	if query == "" {
		return []SearchResult{}
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult

	// For PLAYERS: Only search by name (not nationality, club, etc.)
	playerResultSet := make(map[int]bool)
	si.searchInIndex(queryLower, si.nameIndex, playerResultSet)

	// Search nations FIRST (highest priority)
	for nationName, nation := range si.nations {
		if si.matchesNationality(nationName, queryLower) {
			results = append(results, SearchResult{
				ID:          nationName,
				Name:        nationName,
				Type:        "nation",
				Description: si.formatNationDescription(nation),
				URL:         si.formatNationURL(nationName),
				Overall:     0,
			})
		}
	}

	// Search leagues SECOND (search in division index)
	for leagueName, league := range si.leagues {
		if strings.Contains(strings.ToLower(leagueName), queryLower) {
			results = append(results, SearchResult{
				ID:          leagueName,
				Name:        leagueName,
				Type:        "league",
				Description: si.formatLeagueDescription(league),
				URL:         si.formatLeagueURL(leagueName),
				Overall:     0,
			})
		}
	}

	// Search teams THIRD (search in club index)
	for teamName, team := range si.teams {
		if strings.Contains(strings.ToLower(teamName), queryLower) {
			results = append(results, SearchResult{
				ID:          teamName,
				Name:        teamName,
				Type:        "team",
				Description: si.formatTeamDescription(team),
				URL:         si.formatTeamURL(teamName),
				Overall:     0,
			})
		}
	}

	// Convert player matches to results LAST (only players with matching names)
	for playerIndex := range playerResultSet {
		player := &si.players[playerIndex]
		results = append(results, SearchResult{
			ID:          player.Name,
			Name:        player.Name,
			Type:        "player",
			Description: si.formatPlayerDescription(player),
			URL:         si.formatPlayerURL(player),
			Overall:     player.Overall,
		})
	}

	// Apply proper sorting: Nations → Leagues → Teams → Players (by overall rating)
	sort.Slice(results, func(i, j int) bool {
		// Define type priority: nations (1), leagues (2), teams (3), players (4)
		typePriority := map[string]int{"nation": 1, "league": 2, "team": 3, "player": 4}

		// First sort by type priority
		if typePriority[results[i].Type] != typePriority[results[j].Type] {
			return typePriority[results[i].Type] < typePriority[results[j].Type]
		}

		// For players, sort by highest overall rating first
		if results[i].Type == "player" && results[j].Type == "player" {
			if results[i].Overall != results[j].Overall {
				return results[i].Overall > results[j].Overall // Highest overall first
			}
		}

		// For non-players or same overall rating, sort alphabetically by name
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	// Limit to maxResults for performance
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// searchInIndex searches for a query in a specific inverted index
func (si *SearchIndex) searchInIndex(query string, index map[string][]int, resultSet map[int]bool) {
	// Direct word match
	if playerIndices, exists := index[query]; exists {
		for _, idx := range playerIndices {
			resultSet[idx] = true
		}
	}

	// Prefix matching for longer queries
	if len(query) >= 3 {
		for term, playerIndices := range index {
			if strings.HasPrefix(term, query) || strings.Contains(term, query) {
				for _, idx := range playerIndices {
					resultSet[idx] = true
				}
			}
		}
	}
}

// Helper methods for formatting results
func (si *SearchIndex) formatPlayerDescription(player *Player) string {
	return player.Club + " • " + player.Division + " (" + formatOverall(player.Overall) + " OVR)"
}

func (si *SearchIndex) formatPlayerURL(player *Player) string {
	return "/dataset/?search=" + player.Name
}

func (si *SearchIndex) formatTeamDescription(team TeamInfo) string {
	return team.Division + " • " + formatInt(len(team.Players)) + " players"
}

func (si *SearchIndex) formatTeamURL(teamName string) string {
	return "/dataset/?team=" + teamName
}

func (si *SearchIndex) formatLeagueDescription(league LeagueInfo) string {
	return formatInt(len(league.Players)) + " players"
}

func (si *SearchIndex) formatLeagueURL(leagueName string) string {
	return "/leagues?league=" + leagueName
}

func (si *SearchIndex) formatNationDescription(nation NationInfo) string {
	return formatInt(len(nation.Players)) + " players"
}

func (si *SearchIndex) formatNationURL(nationName string) string {
	return "/nations?nation=" + nationName
}

// matchesNationality provides enhanced nationality matching
func (si *SearchIndex) matchesNationality(nationName, query string) bool {
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

// Utility functions
func formatOverall(overall int) string {
	if overall > 0 {
		return formatInt(overall)
	}
	return "N/A"
}

func formatInt(value int) string {
	return strings.Replace(strings.Replace(fmt.Sprintf("%d", value), "000000", "M", 1), "000", "K", 1)
}

// GetIndexStats returns statistics about the search index
func (si *SearchIndex) GetIndexStats() map[string]interface{} {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	return map[string]interface{}{
		"players":           len(si.players),
		"teams":             len(si.teams),
		"leagues":           len(si.leagues),
		"nations":           len(si.nations),
		"name_terms":        len(si.nameIndex),
		"club_terms":        len(si.clubIndex),
		"position_terms":    len(si.positionIndex),
		"nationality_terms": len(si.nationalityIndex),
		"division_terms":    len(si.divisionIndex),
	}
}

// Global search index instance
var globalSearchIndex = NewSearchIndex()

// GetSearchIndex returns the global search index
func GetSearchIndex() *SearchIndex {
	return globalSearchIndex
}
