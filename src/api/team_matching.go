package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// TeamMatch represents a team matching result
type TeamMatch struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// TeamsData holds the loaded team data
var (
	teamsData   map[string]string           // id -> name
	teamsIndex  map[string][]TeamIndexEntry // normalized name -> list of teams
	teamsMutex  sync.RWMutex
	teamsLoaded bool
)

// TeamIndexEntry represents an indexed team entry
type TeamIndexEntry struct {
	ID             string
	Name           string
	NormalizedName string
	Words          []string
}

// initTeamsData loads and indexes the teams data
func initTeamsData() error {
	teamsMutex.Lock()
	defer teamsMutex.Unlock()

	if teamsLoaded {
		return nil
	}

	LogInfo("Team data initialization: Starting teams data loading process")

	// Load teams data from JSON file
	// In containerized environment, the file is at utils/teams_data.json relative to working directory
	teamsFilePath := filepath.Join("utils", "teams_data.json")
	if _, err := os.Stat(teamsFilePath); os.IsNotExist(err) {
		LogDebug("Team data initialization: Primary path not found: %s", teamsFilePath)
		// Try alternative paths (for different deployment scenarios)
		alternativePaths := []string{
			filepath.Join("src", "api", "utils", "teams_data.json"), // Local development
			filepath.Join("src", "utils", "teams_data.json"),        // Alternative local path
			"teams_data.json", // Root fallback
		}

		found := false
		for _, altPath := range alternativePaths {
			LogDebug("Team data initialization: Trying alternative path: %s", altPath)
			if _, err := os.Stat(altPath); err == nil {
				teamsFilePath = altPath
				found = true
				LogInfo("Team data initialization: Found teams data at: %s", altPath)
				break
			}
		}

		if !found {
			LogWarn("Team data initialization: Teams data file not found at any location: %v", err)
			return err
		}
	} else {
		LogInfo("Team data initialization: Found teams data at primary path: %s", teamsFilePath)
	}

	data, err := os.ReadFile(teamsFilePath)
	if err != nil {
		LogWarn("Team data initialization: Error reading teams data file: %v", err)
		return err
	}

	LogDebug("Team data initialization: Successfully read teams data file (%d bytes)", len(data))

	if err := json.Unmarshal(data, &teamsData); err != nil {
		LogWarn("Team data initialization: Error parsing teams data JSON: %v", err)
		return err
	}

	LogInfo("Team data initialization: Successfully parsed JSON, found %d teams", len(teamsData))

	// Build search index
	teamsIndex = make(map[string][]TeamIndexEntry)
	indexEntries := 0

	for id, name := range teamsData {
		normalized := normalizeTeamName(name)
		words := extractWords(normalized)

		entry := TeamIndexEntry{
			ID:             id,
			Name:           name,
			NormalizedName: normalized,
			Words:          words,
		}

		// Index by normalized full name
		teamsIndex[normalized] = append(teamsIndex[normalized], entry)
		indexEntries++

		// Index by individual words for partial matching
		for _, word := range words {
			if len(word) > 2 { // Only index meaningful words
				teamsIndex[word] = append(teamsIndex[word], entry)
				indexEntries++
			}
		}

		// Index by name prefixes for auto-complete style matching
		if len(normalized) > 3 {
			for i := 3; i <= len(normalized) && i <= 8; i++ {
				prefix := normalized[:i]
				teamsIndex[prefix] = append(teamsIndex[prefix], entry)
				indexEntries++
			}
		}
	}

	teamsLoaded = true
	LogInfo("Team data initialization: Loaded %d teams and built search index with %d total index entries", len(teamsData), indexEntries)
	LogDebug("Team data initialization: Index has %d unique keys", len(teamsIndex))

	// Log some sample teams for verification
	sampleCount := 0
	for id, name := range teamsData {
		if sampleCount < 5 {
			LogDebug("Team data initialization: Sample team - ID: %s, Name: '%s'", id, name)
			sampleCount++
		} else {
			break
		}
	}

	return nil
}

// normalizeTeamName normalizes a team name for comparison
func normalizeTeamName(name string) string {
	if name == "" {
		return ""
	}

	original := name

	// Convert to lowercase and trim
	normalized := strings.ToLower(strings.TrimSpace(name))

	// Remove common prefixes and suffixes
	prefixSuffixPatterns := []string{
		"fc ", "cf ", "ac ", "sc ", "as ", "ca ", "cs ", "rc ", "rs ", "cd ", "ud ",
		"rcd ", "rsd ", "rfc ", "afc ", "cfc ", "sfc ",
		" fc", " cf", " ac", " sc", " as", " ca", " cs", " rc", " rs", " cd", " ud",
		" rcd", " rsd", " rfc", " afc", " cfc", " sfc", " f.c.", " a.c.", " s.c.",
	}

	beforePatterns := normalized
	for _, pattern := range prefixSuffixPatterns {
		if strings.HasPrefix(normalized, pattern) {
			normalized = strings.TrimSpace(normalized[len(pattern):])
		}
		if strings.HasSuffix(normalized, pattern) {
			normalized = strings.TrimSpace(normalized[:len(normalized)-len(pattern)])
		}
	}

	// Remove punctuation and normalize spaces
	var result strings.Builder
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) {
			result.WriteRune(' ')
		}
	}

	// Normalize multiple spaces to single space
	normalized = strings.Join(strings.Fields(result.String()), " ")

	// Log normalization steps if there were significant changes
	if original != normalized {
		LogDebug("Team normalization: '%s' -> '%s'", sanitizeForLogging(original), sanitizeForLogging(normalized))
		if beforePatterns != normalized {
			LogDebug("Team normalization: Removed patterns: '%s' -> '%s'", sanitizeForLogging(beforePatterns), sanitizeForLogging(normalized))
		}
	}

	return normalized
}

// extractWords extracts meaningful words from a normalized team name
func extractWords(normalized string) []string {
	if normalized == "" {
		return []string{}
	}

	// Split by common separators
	words := strings.FieldsFunc(normalized, func(c rune) bool {
		return c == ' ' || c == '-' || c == '.' || c == '/'
	})

	var result []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		// Include words that are at least 2 characters OR known abbreviations
		if len(word) >= 2 || isKnownAbbreviation(word) {
			result = append(result, word)
		}
	}

	LogDebug("Team matching: Extracted words from '%s': %d words", sanitizeForLogging(normalized), len(result))
	return result
}

// isKnownAbbreviation checks if a short string is a known football abbreviation
func isKnownAbbreviation(word string) bool {
	knownAbbrs := map[string]bool{
		"sg": true, // Saint-Germain
		"fc": true, // Football Club
		"cf": true, // Club de Fútbol
		"ac": true, // Athletic Club / Association Club
		"sc": true, // Sport Club
		"rc": true, // Racing Club
		"ca": true, // Club Atlético
		"cd": true, // Club Deportivo
		"ud": true, // Unión Deportiva
		"ba": true, // Borussia
		"sv": true, // Sport Verein
		"bk": true, // Boldklub
		"if": true, // Idrettsforening
		"fk": true, // Fotballklubb
		"pk": true, // Piłkarski Klub
		"sk": true, // Sportklub
	}
	return knownAbbrs[word]
}

// calculateSimilarity calculates similarity between two strings with abbreviation awareness
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Check for abbreviation matches
	if isAbbreviationMatch(s1, s2) {
		return 0.9 // High score for abbreviation matches
	}

	// Use Jaro-Winkler for general similarity
	return jaroWinkler(s1, s2)
}

// isAbbreviationMatch checks if one string could be an abbreviation of another
func isAbbreviationMatch(abbr, full string) bool {
	abbr = strings.ToLower(strings.TrimSpace(abbr))
	full = strings.ToLower(strings.TrimSpace(full))

	// Known mappings
	abbreviations := map[string][]string{
		"sg":  {"saint-germain", "saint germain", "st-germain", "st germain"},
		"utd": {"united"},
		"fc":  {"football club", "futbol club"},
		"ac":  {"athletic club", "atletico club"},
		"rc":  {"racing club", "real club"},
		"cf":  {"club de futbol"},
		"sc":  {"sport club", "sporting club"},
	}

	if expansions, exists := abbreviations[abbr]; exists {
		for _, expansion := range expansions {
			if strings.Contains(full, expansion) {
				return true
			}
		}
	}

	// Reverse check
	if expansions, exists := abbreviations[full]; exists {
		for _, expansion := range expansions {
			if strings.Contains(abbr, expansion) {
				return true
			}
		}
	}

	return false
}

// checkAbbreviationBonus checks if a word could be an abbreviation found in the team name
func checkAbbreviationBonus(word, teamName string) bool {
	word = strings.ToLower(strings.TrimSpace(word))
	teamName = strings.ToLower(strings.TrimSpace(teamName))

	// Check enhanced abbreviations first
	if expansions, exists := enhancedAbbreviations[word]; exists {
		for _, expansion := range expansions {
			if strings.Contains(teamName, expansion) {
				return true
			}
		}
	}

	// Fallback to original mappings
	abbreviationMappings := map[string][]string{
		"sg":  {"saint-germain", "saint germain", "st-germain", "st germain"},
		"utd": {"united"},
		"fc":  {"football club", "futbol club"},
		"ac":  {"athletic club", "atletico club"},
		"rc":  {"racing club", "real club"},
		"cf":  {"club de futbol"},
		"sc":  {"sport club", "sporting club"},
	}

	if expansions, exists := abbreviationMappings[word]; exists {
		for _, expansion := range expansions {
			if strings.Contains(teamName, expansion) {
				return true
			}
		}
	}

	return false
}

// getAbbreviationExpansions returns possible expansions for a known abbreviation
func getAbbreviationExpansions(word string) []string {
	word = strings.ToLower(strings.TrimSpace(word))

	// Check enhanced abbreviations first
	if expansions, exists := enhancedAbbreviations[word]; exists {
		return expansions
	}

	// Fallback to original mappings
	abbreviationMappings := map[string][]string{
		"sg":  {"saint-germain", "saint germain", "st-germain", "st germain"},
		"utd": {"united"},
		"fc":  {"football club", "futbol club"},
		"ac":  {"athletic club", "atletico club"},
		"rc":  {"racing club", "real club"},
		"cf":  {"club de futbol"},
		"sc":  {"sport club", "sporting club"},
	}

	if expansions, exists := abbreviationMappings[word]; exists {
		return expansions
	}

	return []string{}
}

// jaroWinkler calculates the Jaro-Winkler similarity between two strings
func jaroWinkler(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	jaro := jaroSimilarity(s1, s2)
	if jaro < 0.7 {
		return jaro
	}

	// Calculate common prefix (up to 4 characters)
	prefixLen := 0
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}
	if minLen > 4 {
		minLen = 4
	}

	for i := 0; i < minLen && s1[i] == s2[i]; i++ {
		prefixLen++
	}

	return jaro + (0.1 * float64(prefixLen) * (1.0 - jaro))
}

// jaroSimilarity calculates the Jaro similarity between two strings
func jaroSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate the match window
	matchWindow := (maxInt(len1, len2) / 2) - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < len1; i++ {
		start := maxInt(0, i-matchWindow)
		end := minInt(i+matchWindow+1, len2)

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Find transpositions
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	jaro := (float64(matches)/float64(len1) + float64(matches)/float64(len2) + float64(matches-transpositions/2)/float64(matches)) / 3.0
	return jaro
}

// Helper functions for jaroSimilarity
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// findTeamMatches finds matching teams for a given team name
func findTeamMatches(teamName string) []TeamMatch {
	// Ensure teams data is loaded
	if err := initTeamsData(); err != nil {
		LogWarn("Error initializing teams data: %v", err)
		return []TeamMatch{}
	}

	teamsMutex.RLock()
	defer teamsMutex.RUnlock()

	if len(teamsData) == 0 {
		LogWarn("Team matching: No teams data available")
		return []TeamMatch{}
	}

	LogDebug("Team matching: Starting search for '%s'", sanitizeForLogging(teamName))

	// Check for common team name mappings first
	if mappedName, found := checkCommonTeamName(teamName); found {
		LogDebug("Team matching: Found common team name mapping '%s' -> '%s'", sanitizeForLogging(teamName), sanitizeForLogging(mappedName))
		teamName = mappedName
	}

	normalized := normalizeTeamName(teamName)
	words := extractWords(normalized)

	LogDebug("Team matching: Normalized '%s' -> '%s'", sanitizeForLogging(teamName), sanitizeForLogging(normalized))
	LogDebug("Team matching: Extracted %d words", len(words))

	// Track candidates and their scores
	candidates := make(map[string]*TeamMatch)

	// 1. Exact match on normalized name
	if entries, exists := teamsIndex[normalized]; exists {
		LogDebug("Team matching: Found %d exact matches for normalized name", len(entries))
		for _, entry := range entries {
			candidates[entry.ID] = &TeamMatch{
				ID:    entry.ID,
				Name:  entry.Name,
				Score: 1.0,
			}
			LogDebug("Team matching: Exact match - ID: %s, Name: '%s', Score: 1.0", entry.ID, sanitizeForLogging(entry.Name))
		}
	} else {
		LogDebug("Team matching: No exact matches found for normalized name")
	}

	// 2. Partial word matches
	wordMatchCount := 0
	for _, word := range words {
		if entries, exists := teamsIndex[word]; exists {
			LogDebug("Team matching: Found %d entries for word", len(entries))
			for _, entry := range entries {
				if existing, found := candidates[entry.ID]; found {
					// Boost score for multiple word matches
					oldScore := existing.Score
					existing.Score = math.Min(1.0, existing.Score+0.2)
					LogDebug("Team matching: Boosted score for ID %s from %.3f to %.3f",
						entry.ID, oldScore, existing.Score)
				} else {
					// Calculate similarity for partial matches
					score := calculateSimilarity(normalized, entry.NormalizedName)

					// Special bonus for abbreviation matches
					if checkAbbreviationBonus(word, entry.NormalizedName) {
						score = math.Min(1.0, score+0.3)
						LogDebug("Team matching: Applied abbreviation bonus for word in team '%s'", sanitizeForLogging(entry.NormalizedName))
					}

					// Apply disambiguation scoring to penalize partial matches
					score = calculateDisambiguationScore(teamName, entry.Name, score)

					if score > 0.3 { // Only include reasonably similar matches
						candidates[entry.ID] = &TeamMatch{
							ID:    entry.ID,
							Name:  entry.Name,
							Score: score,
						}
						LogDebug("Team matching: Word match - ID: %s, Name: '%s', Score: %.3f",
							entry.ID, sanitizeForLogging(entry.Name), score)
						wordMatchCount++
					}
				}
			}
		}

		// Also check for abbreviation expansion matches
		if expansions := getAbbreviationExpansions(word); len(expansions) > 0 {
			for _, expansion := range expansions {
				if entries, exists := teamsIndex[expansion]; exists {
					LogDebug("Team matching: Found %d entries for abbreviation expansion", len(entries))
					for _, entry := range entries {
						if existing, found := candidates[entry.ID]; found {
							// Boost existing scores for abbreviation matches
							oldScore := existing.Score
							existing.Score = math.Min(1.0, existing.Score+0.3)
							LogDebug("Team matching: Boosted score for abbreviation match ID %s from %.3f to %.3f",
								entry.ID, oldScore, existing.Score)
						} else {
							// High score for abbreviation matches
							score := 0.85 // High base score for abbreviation matches
							// Apply disambiguation scoring
							score = calculateDisambiguationScore(teamName, entry.Name, score)
							candidates[entry.ID] = &TeamMatch{
								ID:    entry.ID,
								Name:  entry.Name,
								Score: score,
							}
							LogDebug("Team matching: Abbreviation match - ID: %s, Name: '%s', Score: %.3f",
								entry.ID, sanitizeForLogging(entry.Name), score)
							wordMatchCount++
						}
					}
				}
			}
		}

		// Check enhanced abbreviation mappings
		if enhancedExpansions, exists := enhancedAbbreviations[word]; exists {
			for _, expansion := range enhancedExpansions {
				if entries, exists := teamsIndex[expansion]; exists {
					LogDebug("Team matching: Found %d entries for enhanced abbreviation expansion", len(entries))
					for _, entry := range entries {
						if existing, found := candidates[entry.ID]; found {
							// Boost existing scores for enhanced abbreviation matches
							oldScore := existing.Score
							existing.Score = math.Min(1.0, existing.Score+0.4) // Higher boost for enhanced mappings
							LogDebug("Team matching: Boosted score for enhanced abbreviation match ID %s from %.3f to %.3f",
								entry.ID, oldScore, existing.Score)
						} else {
							// Very high score for enhanced abbreviation matches
							score := 0.9 // Very high base score for enhanced abbreviation matches
							// Apply disambiguation scoring
							score = calculateDisambiguationScore(teamName, entry.Name, score)
							candidates[entry.ID] = &TeamMatch{
								ID:    entry.ID,
								Name:  entry.Name,
								Score: score,
							}
							LogDebug("Team matching: Enhanced abbreviation match - ID: %s, Name: '%s', Score: %.3f",
								entry.ID, sanitizeForLogging(entry.Name), score)
							wordMatchCount++
						}
					}
				}
			}
		}
	}
	LogDebug("Team matching: Added %d new candidates from word matching", wordMatchCount)

	// 3. Prefix matches for auto-complete style matching
	prefixMatchCount := 0
	if len(normalized) > 2 {
		for i := 3; i <= len(normalized) && i <= 8; i++ {
			prefix := normalized[:i]
			if entries, exists := teamsIndex[prefix]; exists {
				LogDebug("Team matching: Found %d entries for prefix", len(entries))
				for _, entry := range entries {
					if _, found := candidates[entry.ID]; !found {
						score := calculateSimilarity(normalized, entry.NormalizedName)
						// Apply disambiguation scoring to penalize prefix-only matches
						score = calculateDisambiguationScore(teamName, entry.Name, score)
						if score > 0.4 { // Slightly higher threshold for prefix matches
							candidates[entry.ID] = &TeamMatch{
								ID:    entry.ID,
								Name:  entry.Name,
								Score: score,
							}
							LogDebug("Team matching: Prefix match - ID: %s, Name: '%s', Score: %.3f",
								entry.ID, sanitizeForLogging(entry.Name), score)
							prefixMatchCount++
						}
					}
				}
			}
		}
	}
	LogDebug("Team matching: Added %d new candidates from prefix matching", prefixMatchCount)

	// 4. Fallback: similarity search against all teams (limited to prevent performance issues)
	fallbackMatchCount := 0
	if len(candidates) < 5 {
		LogDebug("Team matching: Only %d candidates so far, performing fallback similarity search", len(candidates))
		searchCount := 0
		maxSearches := 1000 // Limit to prevent performance issues

		for id, name := range teamsData {
			if searchCount >= maxSearches {
				LogDebug("Team matching: Reached maximum searches limit (%d), stopping fallback search", maxSearches)
				break
			}

			if _, found := candidates[id]; !found {
				entryNormalized := normalizeTeamName(name)
				score := calculateSimilarity(normalized, entryNormalized)
				// Apply disambiguation scoring
				score = calculateDisambiguationScore(teamName, name, score)
				if score > 0.6 { // Higher threshold for fallback search
					candidates[id] = &TeamMatch{
						ID:    id,
						Name:  name,
						Score: score,
					}
					LogDebug("Team matching: Fallback match - ID: %s, Name: '%s', Score: %.3f",
						id, sanitizeForLogging(name), score)
					fallbackMatchCount++
				}
			}
			searchCount++
		}
		LogDebug("Team matching: Searched %d teams in fallback, added %d new candidates", searchCount, fallbackMatchCount)
	} else {
		LogDebug("Team matching: Skipping fallback search, already have %d candidates", len(candidates))
	}

	// Convert to slice and sort by score
	matches := make([]TeamMatch, 0, len(candidates))
	for _, match := range candidates {
		matches = append(matches, *match)
	}

	LogDebug("Team matching: Total candidates found: %d", len(matches))

	// Sort by score descending, then by ID numerically (lower IDs preferred) for consistency
	sort.Slice(matches, func(i, j int) bool {
		if math.Abs(matches[i].Score-matches[j].Score) < 0.05 {
			// When scores are similar (within 0.05), prefer lower numeric IDs
			idI, errI := strconv.ParseInt(matches[i].ID, 10, 64)
			idJ, errJ := strconv.ParseInt(matches[j].ID, 10, 64)

			// If both IDs can be parsed as numbers, compare numerically
			if errI == nil && errJ == nil {
				return idI < idJ
			}

			// Fallback to string comparison if parsing fails
			return matches[i].ID < matches[j].ID
		}
		return matches[i].Score > matches[j].Score
	})

	// Limit results to top 10 matches
	originalCount := len(matches)
	if len(matches) > 10 {
		matches = matches[:10]
		LogDebug("Team matching: Limited results from %d to 10 matches", originalCount)
	}

	// Log final results summary
	if len(matches) > 0 {
		LogDebug("Team matching: Final results for '%s' (%d matches):", sanitizeForLogging(teamName), len(matches))
		for i, match := range matches {
			LogDebug("Team matching: Result %d - ID: %s, Name: '%s', Score: %.3f",
				i+1, match.ID, sanitizeForLogging(match.Name), match.Score)
		}
	} else {
		LogDebug("Team matching: No matches found for '%s'", sanitizeForLogging(teamName))
	}

	return matches
}

// Common team name mappings for better disambiguation
var teamNameMappings = map[string]string{
	"man utd":           "man utd",                   // Maps to existing "Man Utd" (ID 680)
	"man united":        "man utd",                   // Maps to existing "Man Utd" (ID 680)
	"manchester united": "man utd",                   // Maps to existing "Man Utd" (ID 680)
	"a. madrid":         "atletico madrid",           // Will need to be added to data
	"atletico":          "atletico madrid",           // Will need to be added to data
	"paris sg":          "paris saint germain (psg)", // Maps to existing "Paris Saint-Germain (PSG)" (ID 868)
	"psg":               "paris saint germain (psg)", // Maps to existing "Paris Saint-Germain (PSG)" (ID 868)
	"real madrid":       "real madrid",               // Maps to existing "Real Madrid" (ID 1736)
	"barcelona":         "barcelona",
	"bayern":            "bayern munich",
	"bayern mun":        "bayern munich",
	"arsenal":           "arsenal",
	"chelsea":           "chelsea",
	"liverpool":         "liverpool",
	"tottenham":         "tottenham hotspur",
	"spurs":             "tottenham hotspur",
	"ac milan":          "milan",
	"inter":             "inter milan",
	"juventus":          "juventus",
	"roma":              "as roma",
	"lazio":             "lazio",
	"napoli":            "napoli",
	"ajax":              "ajax",
	"psv":               "psv eindhoven",
	"feyenoord":         "feyenoord",
	"porto":             "fc porto",
	"benfica":           "sl benfica",
	"sporting":          "sporting cp",
	"celtic":            "celtic",
	"rangers":           "rangers",
	"dortmund":          "borussia dortmund",
	"schalke":           "schalke 04",
	"leipzig":           "rb leipzig",
	"leverkusen":        "bayer leverkusen",
	"monaco":            "as monaco",
	"lyon":              "olympique lyonnais",
	"marseille":         "olympique marseille",
	"sevilla":           "sevilla",
	"valencia":          "valencia",
	"villarreal":        "villarreal",
	"athletic":          "athletic bilbao",
	"atletico bilbao":   "athletic bilbao",
}

// Enhanced abbreviation mappings
var enhancedAbbreviations = map[string][]string{
	"sg":     {"saint germain", "saint-germain", "st germain", "st-germain"},
	"utd":    {"united"},
	"fc":     {"football club", "futbol club"},
	"ac":     {"athletic club", "atletico club"},
	"rc":     {"racing club", "real club"},
	"cf":     {"club de futbol"},
	"sc":     {"sport club", "sporting club"},
	"a.":     {"atletico"},
	"madrid": {"madrid"},
	"psg":    {"paris saint germain"},
	"man":    {"manchester"},
	"city":   {"city"},
	"united": {"united"},
}

// calculateDisambiguationScore calculates a score that penalizes partial matches
// when the query contains more specific information
func calculateDisambiguationScore(query, candidate string, baseScore float64) float64 {
	queryWords := strings.Fields(strings.ToLower(query))
	candidateWords := strings.Fields(strings.ToLower(candidate))

	// Count how many query words are found in candidate
	matchedWords := 0
	totalQueryWords := len(queryWords)

	for _, queryWord := range queryWords {
		for _, candidateWord := range candidateWords {
			if strings.Contains(candidateWord, queryWord) || strings.Contains(queryWord, candidateWord) {
				matchedWords++
				break
			}
		}
	}

	// If we matched all query words, boost the score
	if matchedWords == totalQueryWords && totalQueryWords > 0 {
		return math.Min(1.0, baseScore+0.2)
	}

	// If we only matched some words, penalize the score
	if matchedWords < totalQueryWords && totalQueryWords > 1 {
		penalty := float64(totalQueryWords-matchedWords) * 0.1
		return math.Max(0.0, baseScore-penalty)
	}

	return baseScore
}

// checkCommonTeamName checks if the query matches a common team name pattern
func checkCommonTeamName(query string) (string, bool) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	// Check direct mappings
	if mapped, exists := teamNameMappings[normalizedQuery]; exists {
		return mapped, true
	}

	// Check for partial matches in mappings
	for pattern, mapped := range teamNameMappings {
		if strings.Contains(normalizedQuery, pattern) || strings.Contains(pattern, normalizedQuery) {
			return mapped, true
		}
	}

	return "", false
}
