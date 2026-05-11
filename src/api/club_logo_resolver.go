package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	clubLogoStatusExact       = "exact"
	clubLogoStatusConfident   = "confident"
	clubLogoStatusAmbiguous   = "ambiguous"
	clubLogoStatusMissingLogo = "missing_logo"
	clubLogoStatusUnmatched   = "unmatched"

	clubLogoConfidenceThreshold = 0.70
	clubLogoAmbiguousDelta      = 0.05
	clubLogoExactThreshold      = 0.98
)

// ClubLogoCandidate describes a possible club logo resolution.
type ClubLogoCandidate struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Score         float64 `json:"score"`
	LogoAvailable bool    `json:"logoAvailable"`
	LogoURL       string  `json:"logoUrl,omitempty"`
}

// ClubLogoResolution is the structured result returned to the frontend.
type ClubLogoResolution struct {
	Status       string              `json:"status"`
	Query        string              `json:"query"`
	TeamID       string              `json:"teamId,omitempty"`
	TeamName     string              `json:"teamName,omitempty"`
	Score        float64             `json:"score,omitempty"`
	LogoURL      string              `json:"logoUrl,omitempty"`
	Reason       string              `json:"reason,omitempty"`
	Alternatives []ClubLogoCandidate `json:"alternatives,omitempty"`
}

var clubLogoAliases = map[string]string{
	"manchester city":         "679",
	"man city":                "679",
	"mancity":                 "679",
	"manchester united":       "680",
	"man utd":                 "680",
	"man united":              "680",
	"manchester utd":          "680",
	"manutd":                  "680",
	"nottingham forest":       "692",
	"nottm forest":            "692",
	"notts forest":            "692",
	"valencia":                "1775",
	"valencia cf":             "1775",
	"valencia c f":            "1775",
	"tottenham":               "728",
	"tottenham hotspur":       "728",
	"spurs":                   "728",
	"brighton":                "618",
	"brighton hove":           "618",
	"brighton hove albion":    "618",
	"west ham":                "735",
	"west ham united":         "735",
	"newcastle":               "688",
	"newcastle united":        "688",
	"wolves":                  "740",
	"wolverhampton":           "740",
	"wolverhampton wanderers": "740",
	"qpr":                     "701",
	"queens park rangers":     "701",
	"psg":                     "868",
	"paris saint germain":     "868",
}

func resolveClubLogo(query string) ClubLogoResolution {
	query = strings.TrimSpace(query)
	if query == "" {
		return ClubLogoResolution{
			Status: clubLogoStatusUnmatched,
			Query:  query,
			Reason: "empty query",
		}
	}

	if err := initTeamsData(); err != nil {
		LogWarn("Club logo resolver: failed to initialize teams data: %v", err)
		return ClubLogoResolution{
			Status: clubLogoStatusUnmatched,
			Query:  query,
			Reason: "teams data unavailable",
		}
	}

	// User overrides take priority over everything
	if teamID, isRejection, found := getLogoOverride(query); found {
		if isRejection {
			return ClubLogoResolution{
				Status: clubLogoStatusUnmatched,
				Query:  query,
				Reason: "user-rejected",
			}
		}
		if candidate, ok := clubLogoCandidateForID(teamID, 1.0); ok {
			if !candidate.LogoAvailable {
				return resolutionFromCandidate(query, clubLogoStatusMissingLogo, "user override but logo unavailable", candidate, nil)
			}
			return resolutionFromCandidate(query, clubLogoStatusExact, "user override", candidate, nil)
		}
	}

	if aliasID, ok := clubLogoAliases[normalizeLogoAlias(query)]; ok {
		if candidate, ok := clubLogoCandidateForID(aliasID, 1.0); ok {
			if !candidate.LogoAvailable {
				return resolutionFromCandidate(query, clubLogoStatusMissingLogo, "alias matched but logo is not available", candidate, nil)
			}
			return resolutionFromCandidate(query, clubLogoStatusExact, "alias match", candidate, nil)
		}
	}

	matches := findTeamMatches(query)
	if len(matches) == 0 {
		return ClubLogoResolution{
			Status: clubLogoStatusUnmatched,
			Query:  query,
			Reason: "no team match found",
		}
	}

	candidates := make([]ClubLogoCandidate, 0, len(matches))
	for _, match := range matches {
		candidates = append(candidates, clubLogoCandidateFromMatch(match))
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if math.Abs(candidates[i].Score-candidates[j].Score) <= clubLogoAmbiguousDelta {
			return numericIDLess(candidates[i].ID, candidates[j].ID)
		}
		return candidates[i].Score > candidates[j].Score
	})

	best := candidates[0]
	alternatives := candidates
	if len(alternatives) > 5 {
		alternatives = alternatives[:5]
	}

	if best.Score < clubLogoConfidenceThreshold {
		return ClubLogoResolution{
			Status:       clubLogoStatusUnmatched,
			Query:        query,
			Reason:       "best match below confidence threshold",
			Alternatives: alternatives,
		}
	}

	if !best.LogoAvailable {
		return resolutionFromCandidate(query, clubLogoStatusMissingLogo, "team matched but logo is not available", best, alternatives)
	}

	if len(candidates) > 1 && isAmbiguousLogoMatch(best, candidates[1]) {
		return resolutionFromCandidate(query, clubLogoStatusAmbiguous, "top candidates are too close to choose automatically", best, alternatives)
	}

	status := clubLogoStatusConfident
	reason := "confident fuzzy match"
	if best.Score >= clubLogoExactThreshold {
		status = clubLogoStatusExact
		reason = "exact normalized match"
	}

	return resolutionFromCandidate(query, status, reason, best, alternatives)
}

func normalizeLogoAlias(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "&", " and ")
	replacer := strings.NewReplacer(".", " ", "-", " ", "_", " ", "'", "")
	normalized = replacer.Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}

func clubLogoCandidateForID(teamID string, score float64) (ClubLogoCandidate, bool) {
	teamsMutex.RLock()
	name, ok := teamsData[teamID]
	teamsMutex.RUnlock()
	if !ok {
		return ClubLogoCandidate{}, false
	}
	return clubLogoCandidateFromMatch(TeamMatch{ID: teamID, Name: name, Score: score}), true
}

func clubLogoCandidateFromMatch(match TeamMatch) ClubLogoCandidate {
	available := isTeamLogoAvailable(match.ID)
	candidate := ClubLogoCandidate{
		ID:            match.ID,
		Name:          match.Name,
		Score:         match.Score,
		LogoAvailable: available,
	}
	if available {
		candidate.LogoURL = teamLogoAPIURL(match.ID)
	}
	return candidate
}

func resolutionFromCandidate(query, status, reason string, candidate ClubLogoCandidate, alternatives []ClubLogoCandidate) ClubLogoResolution {
	resolution := ClubLogoResolution{
		Status:       status,
		Query:        query,
		TeamID:       candidate.ID,
		TeamName:     candidate.Name,
		Score:        candidate.Score,
		LogoURL:      candidate.LogoURL,
		Reason:       reason,
		Alternatives: alternatives,
	}
	if status == clubLogoStatusAmbiguous || status == clubLogoStatusMissingLogo {
		resolution.LogoURL = ""
	}
	return resolution
}

func isAmbiguousLogoMatch(best, second ClubLogoCandidate) bool {
	if best.Score >= clubLogoExactThreshold {
		return false
	}
	return math.Abs(best.Score-second.Score) <= clubLogoAmbiguousDelta
}

func teamLogoAPIURL(teamID string) string {
	values := url.Values{}
	values.Set("teamId", teamID)
	values.Set("size", "256")
	return "/api/logos?" + values.Encode()
}

func isTeamLogoAvailable(teamID string) bool {
	if os.Getenv("IMAGE_API_URL") != "" {
		// The external provider is not enumerable from this app. Treat syntactically
		// valid team IDs as displayable and let the image request fail closed if the
		// provider does not have the asset.
		return true
	}

	if s3Storage := underlyingS3Storage(); s3Storage != nil && s3Storage.client != nil {
		return true
	}

	_, err := findLocalTeamLogoPath(teamID)
	return err == nil
}

func findLocalTeamLogoPath(teamID string) (string, error) {
	if err := validateID(teamID, 100); err != nil {
		return "", err
	}

	logoFileName := teamID + ".png"
	logosDir := getLogosDirectory()
	candidates := [][]string{
		{"Clubs", "Normal", "Normal", logoFileName},
		{"clubs", "Normal", "Normal", logoFileName},
	}

	for _, parts := range candidates {
		path := logosDir
		var err error
		for _, part := range parts {
			path, err = validateAndJoinPath(path, part)
			if err != nil {
				return "", err
			}
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", os.ErrNotExist
}

func numericIDLess(left, right string) bool {
	leftID, leftErr := strconv.ParseInt(left, 10, 64)
	rightID, rightErr := strconv.ParseInt(right, 10, 64)
	if leftErr == nil && rightErr == nil {
		return leftID < rightID
	}
	return left < right
}

func externalTeamLogoURL(teamID string) string {
	imageAPIURL := strings.TrimRight(os.Getenv("IMAGE_API_URL"), "/")
	if imageAPIURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/team/%s.png?width=256", imageAPIURL, teamID)
}

func localTeamLogoPathForHandler(teamID string) (string, error) {
	path, err := findLocalTeamLogoPath(teamID)
	if err == nil {
		return path, nil
	}

	path = getLogosDirectory()
	for _, part := range []string{"Clubs", "Normal", "Normal", teamID + ".png"} {
		path, err = validateAndJoinPath(path, part)
		if err != nil {
			return "", err
		}
	}
	return path, os.ErrNotExist
}
