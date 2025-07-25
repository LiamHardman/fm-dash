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

	apperrors "api/errors"
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

	// Load teams data from JSON file using secure path validation
	// Try multiple possible locations for the teams data file
	possiblePaths := []string{
		"utils/teams_data.json",         // Containerized environment
		"src/api/utils/teams_data.json", // Local development
		"src/utils/teams_data.json",     // Alternative local path
		"teams_data.json",               // Root fallback
	}

	var teamsFilePath string
	var found bool

	for _, path := range possiblePaths {
		// Validate the path components for security
		pathParts := strings.Split(path, "/")
		isValidPath := true

		for _, part := range pathParts {
			if err := validateFileName(part); err != nil {
				LogWarn("Team data initialization: Invalid path component '%s' in path '%s': %v", sanitizeForLogging(part), sanitizeForLogging(path), err)
				isValidPath = false
				break
			}
		}

		if !isValidPath {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			teamsFilePath = path
			found = true
			LogInfo("Team data initialization: Found teams data at: %s", sanitizeForLogging(path))
			break
		}
		LogDebug("Team data initialization: Path not found: %s", sanitizeForLogging(path))
	}

	if !found {
		LogWarn("Team data initialization: Teams data file not found at any valid location")
		return apperrors.ErrDatasetNotFound
	}

	// Validate the file path for security before reading
	if err := validateFileName(filepath.Base(teamsFilePath)); err != nil {
		LogWarn("Team data initialization: Invalid file path for security: %v", err)
		return apperrors.ErrInvalidFilePath
	}

	// Additional security validation: ensure path is within allowed directories
	cleanPath := filepath.Clean(teamsFilePath)
	if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "\\") {
		LogWarn("Team data initialization: Path contains unsafe characters: %s", sanitizeForLogging(cleanPath))
		return apperrors.ErrInvalidFilePath
	}

	// Only allow specific file extensions
	if !strings.HasSuffix(cleanPath, ".json") {
		LogWarn("Team data initialization: Invalid file extension: %s", sanitizeForLogging(cleanPath))
		return apperrors.ErrInvalidFilePath
	}

	// Use secure file reading with additional validation
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
			if len(word) > 2 { // Only index meaningful words (3+ characters)
				teamsIndex[word] = append(teamsIndex[word], entry)
				indexEntries++
			}
		}

		// Index by multi-word combinations for better multi-word matching
		if len(words) > 1 {
			wordCombinations := generateWordCombinations(words)
			for _, combination := range wordCombinations {
				teamsIndex[combination] = append(teamsIndex[combination], entry)
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
			LogDebug("Team data initialization: Sample team - ID: %s, Name: '%s'", id, sanitizeForLogging(name))
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
		// Include words that are at least 3 characters OR known abbreviations
		// Increased minimum length to avoid partial word matches like "man" from "manchester"
		if len(word) >= 3 || isKnownAbbreviation(word) {
			result = append(result, word)
		}
	}

	LogDebug("Team matching: Extracted words from '%s': %d words", sanitizeForLogging(normalized), len(result))
	return result
}

// isKnownAbbreviation checks if a short string is a known football abbreviation
func isKnownAbbreviation(word string) bool {
	knownAbbrs := map[string]bool{
		"sg":  true, // Saint-Germain
		"fc":  true, // Football Club
		"cf":  true, // Club de Fútbol
		"ac":  true, // Athletic Club / Association Club
		"sc":  true, // Sport Club
		"rc":  true, // Racing Club
		"ca":  true, // Club Atlético
		"cd":  true, // Club Deportivo
		"ud":  true, // Unión Deportiva
		"ba":  true, // Borussia
		"sv":  true, // Sport Verein
		"bk":  true, // Boldklub
		"if":  true, // Idrettsforening
		"fk":  true, // Fotballklubb
		"pk":  true, // Piłkarski Klub
		"sk":  true, // Sportklub
		"utd": true, // United
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

	// Calculate base similarity using Jaro-Winkler
	baseSimilarity := jaroWinkler(s1, s2)

	// Apply length-based penalty for significantly different lengths
	len1 := len(s1)
	len2 := len(s2)

	if len1 > 0 && len2 > 0 {
		lengthRatio := float64(minInt(len1, len2)) / float64(maxInt(len1, len2))

		// If the length ratio is very low (one string is much shorter), apply penalty
		if lengthRatio < 0.5 {
			// Apply penalty for length mismatch
			penalty := (0.5 - lengthRatio) * 0.3 // Up to 0.15 penalty
			baseSimilarity -= penalty
		}

		// Additional penalty for very short strings when comparing to longer ones
		if minInt(len1, len2) < 5 && maxInt(len1, len2) > 10 {
			baseSimilarity -= 0.1 // Additional penalty for very short vs very long
		}
	}

	return math.Max(0.0, baseSimilarity) // Ensure we don't go below 0
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
		"man": {"manchester"}, // Only allow "man" to match "manchester" in specific contexts
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

	// Special case for "Manchester City" -> "Man City" and similar patterns
	if strings.Contains(full, "manchester") && strings.Contains(abbr, "man") {
		return true
	}

	return false
}

// checkAbbreviationBonus checks if a word could be an abbreviation found in the team name
func checkAbbreviationBonus(word, teamName string) bool {
	word = strings.ToLower(strings.TrimSpace(word))
	teamName = strings.ToLower(strings.TrimSpace(teamName))

	// Check common football abbreviations
	abbreviationMappings := map[string][]string{
		"sg":  {"saint-germain", "saint germain", "st-germain", "st germain"},
		"utd": {"united"},
		"fc":  {"football club", "futbol club"},
		"ac":  {"athletic club", "atletico club"},
		"rc":  {"racing club", "real club"},
		"cf":  {"club de futbol"},
		"sc":  {"sport club", "sporting club"},
		"man": {"manchester"}, // Add man -> manchester mapping
	}

	if expansions, exists := abbreviationMappings[word]; exists {
		for _, expansion := range expansions {
			if strings.Contains(teamName, expansion) {
				return true
			}
		}
	}

	// Special case for "Manchester City" -> "Man City" pattern
	if word == "man" && strings.Contains(teamName, "manchester") {
		return true
	}

	return false
}

// getAbbreviationExpansions returns possible expansions for a known abbreviation
func getAbbreviationExpansions(word string) []string {
	word = strings.ToLower(strings.TrimSpace(word))

	abbreviationMappings := map[string][]string{
		"sg":  {"saint-germain", "saint germain", "st-germain", "st germain"},
		"utd": {"united"},
		"fc":  {"football club", "futbol club"},
		"ac":  {"athletic club", "atletico club"},
		"rc":  {"racing club", "real club"},
		"cf":  {"club de futbol"},
		"sc":  {"sport club", "sporting club"},
		"man": {"manchester"}, // Add man -> manchester mapping
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

	normalized := normalizeTeamName(teamName)
	words := extractWords(normalized)

	LogDebug("Team matching: Normalized '%s' -> '%s'", sanitizeForLogging(teamName), sanitizeForLogging(normalized))
	LogDebug("Team matching: Extracted %d words", len(words))

	// Track candidates and their scores
	candidates := make(map[string]*TeamMatch)

	// 1. Exact match on normalized name (highest priority)
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

	// 2. Multi-word exact matches (high priority)
	if len(words) > 1 {
		// Create all possible word combinations for multi-word matching
		wordCombinations := generateWordCombinations(words)
		multiWordMatchCount := 0

		for _, combination := range wordCombinations {
			if entries, exists := teamsIndex[combination]; exists {
				LogDebug("Team matching: Found %d entries for multi-word combination '%s'", len(entries), combination)
				for _, entry := range entries {
					if existing, found := candidates[entry.ID]; found {
						// Boost score for multi-word matches
						oldScore := existing.Score
						existing.Score = math.Min(1.0, existing.Score+0.3)
						LogDebug("Team matching: Boosted score for multi-word match ID %s from %.3f to %.3f",
							entry.ID, oldScore, existing.Score)
					} else {
						// High score for multi-word matches
						score := 0.95 // Very high score for multi-word matches
						candidates[entry.ID] = &TeamMatch{
							ID:    entry.ID,
							Name:  entry.Name,
							Score: score,
						}
						LogDebug("Team matching: Multi-word match - ID: %s, Name: '%s', Score: %.3f",
							entry.ID, sanitizeForLogging(entry.Name), score)
						multiWordMatchCount++
					}
				}
			}
		}
		LogDebug("Team matching: Added %d new candidates from multi-word matching", multiWordMatchCount)
	}

	// 3. Partial word matches (lower priority, with improved scoring)
	wordMatchCount := 0
	for _, word := range words {
		if entries, exists := teamsIndex[word]; exists {
			LogDebug("Team matching: Found %d entries for word '%s'", len(entries), word)
			for _, entry := range entries {
				if existing, found := candidates[entry.ID]; found {
					// Boost score for multiple word matches, but with diminishing returns
					oldScore := existing.Score
					boost := 0.1 // Reduced boost for single word matches
					existing.Score = math.Min(1.0, existing.Score+boost)
					LogDebug("Team matching: Boosted score for ID %s from %.3f to %.3f",
						entry.ID, oldScore, existing.Score)
				} else {
					// Calculate similarity for partial matches with stricter scoring
					score := calculateSimilarity(normalized, entry.NormalizedName)

					// Apply word match penalty for single word matches
					if len(words) == 1 {
						// For single word searches, be more lenient
						if score > 0.4 {
							candidates[entry.ID] = &TeamMatch{
								ID:    entry.ID,
								Name:  entry.Name,
								Score: score,
							}
							LogDebug("Team matching: Single word match - ID: %s, Name: '%s', Score: %.3f",
								entry.ID, sanitizeForLogging(entry.Name), score)
							wordMatchCount++
						}
					} else {
						// For multi-word searches, apply penalty for partial matches
						wordMatchPenalty := 0.2
						adjustedScore := score - wordMatchPenalty

						// Additional penalty for word boundary mismatches
						// Check if the word appears as a complete word in the team name
						entryWords := strings.Fields(entry.NormalizedName)
						wordFoundAsComplete := false
						for _, entryWord := range entryWords {
							if entryWord == word {
								wordFoundAsComplete = true
								break
							}
						}

						// If the word doesn't appear as a complete word, apply additional penalty
						if !wordFoundAsComplete {
							adjustedScore -= 0.15 // Additional penalty for partial word matches
						}

						// Special bonus for abbreviation matches
						if checkAbbreviationBonus(word, entry.NormalizedName) {
							adjustedScore = math.Min(1.0, adjustedScore+0.2)
							LogDebug("Team matching: Applied abbreviation bonus for word in team '%s'", sanitizeForLogging(entry.NormalizedName))
						}

						// Special handling for "Manchester City" -> "Man City" pattern
						if len(words) == 2 && word == "man" {
							// Check if the other word is "city" and if this team contains "city"
							otherWord := ""
							if words[0] == "man" {
								otherWord = words[1]
							} else if words[1] == "man" {
								otherWord = words[0]
							}

							if otherWord == "city" && strings.Contains(entry.NormalizedName, "city") {
								adjustedScore = math.Min(1.0, adjustedScore+0.3) // High bonus for this specific pattern
								LogDebug("Team matching: Applied Manchester City pattern bonus for team '%s'", sanitizeForLogging(entry.NormalizedName))
							}
						}

						if adjustedScore > 0.4 { // Higher threshold for multi-word searches
							candidates[entry.ID] = &TeamMatch{
								ID:    entry.ID,
								Name:  entry.Name,
								Score: adjustedScore,
							}
							LogDebug("Team matching: Word match (penalized) - ID: %s, Name: '%s', Score: %.3f",
								entry.ID, sanitizeForLogging(entry.Name), adjustedScore)
							wordMatchCount++
						}
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
							existing.Score = math.Min(1.0, existing.Score+0.2)
							LogDebug("Team matching: Boosted score for abbreviation match ID %s from %.3f to %.3f",
								entry.ID, oldScore, existing.Score)
						} else {
							// High score for abbreviation matches
							score := 0.85 // High base score for abbreviation matches
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

		// Special case: Check for "Man City" -> "Manchester City" pattern
		if word == "man" {
			// Look for teams that contain "manchester" and also contain the other word from the search
			for id, name := range teamsData {
				entryNormalized := normalizeTeamName(name)
				if strings.Contains(entryNormalized, "manchester") {
					// Check if this team also contains the other word from our search
					otherWords := []string{}
					for _, w := range words {
						if w != "man" {
							otherWords = append(otherWords, w)
						}
					}

					allOtherWordsFound := true
					for _, otherWord := range otherWords {
						if !strings.Contains(entryNormalized, otherWord) {
							allOtherWordsFound = false
							break
						}
					}

					if allOtherWordsFound && len(otherWords) > 0 {
						// This is a potential "Manchester City" match
						if existing, found := candidates[id]; found {
							// Boost existing score
							oldScore := existing.Score
							existing.Score = math.Min(1.0, existing.Score+0.25)
							LogDebug("Team matching: Boosted score for Manchester pattern ID %s from %.3f to %.3f",
								id, oldScore, existing.Score)
						} else {
							// High score for this pattern
							score := 0.9
							candidates[id] = &TeamMatch{
								ID:    id,
								Name:  name,
								Score: score,
							}
							LogDebug("Team matching: Manchester pattern match - ID: %s, Name: '%s', Score: %.3f",
								id, sanitizeForLogging(name), score)
							wordMatchCount++
						}
					}
				}
			}
		}

		// Special case: Check for "Manchester City" -> "Man City" pattern
		if word == "manchester" {
			// Look for teams that contain "man" and also contain the other word from the search
			for id, name := range teamsData {
				entryNormalized := normalizeTeamName(name)
				if strings.Contains(entryNormalized, "man") {
					// Check if this team also contains the other word from our search
					otherWords := []string{}
					for _, w := range words {
						if w != "manchester" {
							otherWords = append(otherWords, w)
						}
					}

					allOtherWordsFound := true
					for _, otherWord := range otherWords {
						if !strings.Contains(entryNormalized, otherWord) {
							allOtherWordsFound = false
							break
						}
					}

					if allOtherWordsFound && len(otherWords) > 0 {
						// This is a potential "Man City" match
						if existing, found := candidates[id]; found {
							// Boost existing score
							oldScore := existing.Score
							existing.Score = math.Min(1.0, existing.Score+0.25)
							LogDebug("Team matching: Boosted score for Man City pattern ID %s from %.3f to %.3f",
								id, oldScore, existing.Score)
						} else {
							// High score for this pattern
							score := 0.9
							candidates[id] = &TeamMatch{
								ID:    id,
								Name:  name,
								Score: score,
							}
							LogDebug("Team matching: Man City pattern match - ID: %s, Name: '%s', Score: %.3f",
								id, sanitizeForLogging(name), score)
							wordMatchCount++
						}
					}
				}
			}
		}
	}
	LogDebug("Team matching: Added %d new candidates from word matching", wordMatchCount)

	// 4. Prefix matches for auto-complete style matching
	prefixMatchCount := 0
	if len(normalized) > 2 {
		for i := 3; i <= len(normalized) && i <= 8; i++ {
			prefix := normalized[:i]
			if entries, exists := teamsIndex[prefix]; exists {
				LogDebug("Team matching: Found %d entries for prefix", len(entries))
				for _, entry := range entries {
					if _, found := candidates[entry.ID]; !found {
						score := calculateSimilarity(normalized, entry.NormalizedName)
						if score > 0.5 { // Higher threshold for prefix matches
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

	// 5. Fallback: similarity search against all teams (limited to prevent performance issues)
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
				if score > 0.7 { // Higher threshold for fallback search
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

// generateWordCombinations generates all possible word combinations for multi-word matching
func generateWordCombinations(words []string) []string {
	if len(words) <= 1 {
		return words
	}

	var combinations []string

	// Add the full phrase
	fullPhrase := strings.Join(words, " ")
	combinations = append(combinations, fullPhrase)

	// Add combinations of 2+ words (for cases like "Real Madrid" -> "real madrid")
	if len(words) >= 2 {
		// Add all consecutive word pairs
		for i := 0; i < len(words)-1; i++ {
			pair := strings.Join(words[i:i+2], " ")
			combinations = append(combinations, pair)
		}

		// Add all consecutive word triplets
		if len(words) >= 3 {
			for i := 0; i < len(words)-2; i++ {
				triplet := strings.Join(words[i:i+3], " ")
				combinations = append(combinations, triplet)
			}
		}
	}

	return combinations
}
