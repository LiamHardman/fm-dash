package main

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// Deterministic star ratings — ticket 06 of the Scout Report map
// (.scratch/scout-report/issues/06-star-rating-scaling-formula.md). Pure Go, no LLM call,
// so this is served by its own fast GET endpoint (ticket 04) rather than bundled into the
// LLM response — the frontend fires it the instant the tab opens, and again on every
// position-selector change, independent of the "Regenerate" button that gates the LLM call.

// computeStars implements the corrected 0-5 scale (originally specified as 0-3, corrected
// live while reacting to ticket 07's prototype — "N stars = equal to the best player" was
// the user's example point on a 5-star scale, not the scale's ceiling):
//
//	pool = players at `position`, in `scope`, excluding the subject player
//	if pool is empty: stars = 5.0   // trivially "the best" — no rival to compare against
//	else: stars = round_to_nearest_half(clamp(subject.CA / bestCA_in_pool * 5, 0.5, 5.0))
func computeStars(subjectCA int, pool []Player) float64 {
	bestCA := 0
	for _, p := range pool {
		if p.CA > bestCA {
			bestCA = p.CA
		}
	}
	if bestCA == 0 {
		return 5.0
	}
	raw := (float64(subjectCA) / float64(bestCA)) * 5.0
	if raw > 5.0 {
		raw = 5.0
	}
	rounded := math.Round(raw*2) / 2
	if rounded < 0.5 {
		rounded = 0.5
	}
	return rounded
}

// playersAtPosition filters players to those with the given short position, excluding
// excludeUID (the subject never gets compared against themselves).
func playersAtPosition(players []Player, shortPosition string, excludeUID int64) []Player {
	pool := make([]Player, 0, len(players))
	for _, p := range players {
		if p.UID == excludeUID {
			continue
		}
		for _, pos := range p.ShortPositions {
			if strings.EqualFold(pos, shortPosition) {
				pool = append(pool, p)
				break
			}
		}
	}
	return pool
}

// scoutReportStars computes both scopes for a given subject player. squadClub is the
// managed team's own club (not necessarily the subject's own club, since the subject may
// be an external transfer target being compared against the manager's own squad).
func scoutReportStars(players []Player, subject Player, position string, squadClub string) (squadStars, divisionStars float64) {
	squadPool := playersAtPosition(filterByClub(players, squadClub), position, subject.UID)
	divisionPool := playersAtPosition(filterByDivision(players, subject.Division), position, subject.UID)
	return computeStars(subject.CA, squadPool), computeStars(subject.CA, divisionPool)
}

func filterByClub(players []Player, club string) []Player {
	out := make([]Player, 0, len(players))
	for _, p := range players {
		if p.Club == club {
			out = append(out, p)
		}
	}
	return out
}

func filterByDivision(players []Player, division string) []Player {
	out := make([]Player, 0, len(players))
	for _, p := range players {
		if p.Division == division {
			out = append(out, p)
		}
	}
	return out
}

type scoutReportStarsResponse struct {
	SquadStars    float64 `json:"squadStars"`
	DivisionStars float64 `json:"divisionStars"`
}

// scoutReportStarsHandler serves GET /api/scout-report-stars/{datasetId}?playerUid=&position=
func scoutReportStarsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	datasetID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/scout-report-stars/"), "/")
	if datasetID == "" {
		WriteErrorResponse(w, r, "invalid_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}

	playerUIDStr := r.URL.Query().Get("playerUid")
	position := r.URL.Query().Get("position")
	if playerUIDStr == "" || position == "" {
		WriteErrorResponse(w, r, "missing_params", "playerUid and position are both required", nil, http.StatusBadRequest)
		return
	}
	playerUID, err := strconv.ParseInt(playerUIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid_player_uid", "playerUid must be an integer", nil, http.StatusBadRequest)
		return
	}

	players, _, found := GetPlayerData(datasetID)
	if !found {
		WriteErrorResponse(w, r, "dataset_not_found", "Dataset not found", nil, http.StatusNotFound)
		return
	}
	players = RecalculateAllPlayersRatings(players)

	var subject *Player
	for i := range players {
		if players[i].UID == playerUID {
			subject = &players[i]
			break
		}
	}
	if subject == nil {
		WriteErrorResponse(w, r, "player_not_found", "Player not found in dataset", nil, http.StatusNotFound)
		return
	}

	managedTeam, err := loadManagedTeam(datasetID)
	if err != nil {
		WriteErrorResponse(w, r, "managed_team_not_set", "Set your managed team before generating a Scout Report", nil, http.StatusBadRequest)
		return
	}

	squadStars, divisionStars := scoutReportStars(players, *subject, position, managedTeam.Club)
	RecordScoutReportSubjectStars(r.Context(), squadStars, divisionStars)

	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(scoutReportStarsResponse{SquadStars: squadStars, DivisionStars: divisionStars}); err != nil {
		LogWarn("scout-report-stars: failed to encode response: %v", err)
	}
}

// loadManagedTeam reads the managed team for a dataset, or returns an error if unset —
// shared by the stars endpoint and the LLM-backed scout-report endpoint.
func loadManagedTeam(datasetID string) (*ManagedTeam, error) {
	if configStorage == nil {
		return nil, http.ErrNoLocation
	}
	data, err := configStorage.RetrieveConfig(managedTeamStorageKey(datasetID))
	if err != nil {
		return nil, err
	}
	var team ManagedTeam
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, err
	}
	return &team, nil
}
