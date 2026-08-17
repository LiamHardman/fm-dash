package main

import (
	"sort"
	"strconv"
	"strings"
)

// Shared player-search implementation backing both Who to Sign's find_players tool
// (who_to_sign.go) and Scout Report's find_comparable_players tool (scout_report.go).
// One filter/sort/trim code path, per the Scout Report map's ticket 02 decision — each
// feature keeps its own thin, minimal LLM-facing tool contract on top of this.

// PlayerSearchCriteria is the shared filter set. Zero-value fields mean "no filter",
// matching the existing convention each individual field already had in
// whoToSignFindPlayersArgs. MinCA/MaxCA are new — Who to Sign's find_players never sets
// them; Scout Report's find_comparable_players is the only caller that does, since
// ability-similarity for that feature is keyed on CA (what the star ratings and grading
// are based on), not Overall.
type PlayerSearchCriteria struct {
	ShortPosition string
	MinCA         int
	MaxCA         int
	MaxBudget     int64
	MaxSalary     int64
	MinAge        int
	MaxAge        int
	MinOverall    int
	// ExcludeUID, when non-zero, drops that player from results — used by Scout Report's
	// find_comparable_players so the subject never appears in their own comparables list.
	ExcludeUID int64
}

// searchPlayers filters the dataset's players against criteria, sorts by MBR descending,
// and trims to cap. Returns raw Player records — serialization to whatever summary shape
// a caller needs (ScoutPlayerSummary today) stays the caller's responsibility.
func searchPlayers(datasetID string, criteria PlayerSearchCriteria, cap int) []Player {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil
	}
	players = RecalculateAllPlayersRatings(players)

	matches := make([]Player, 0, 32)
	for i := range players {
		player := players[i]

		if criteria.ExcludeUID != 0 && player.UID == criteria.ExcludeUID {
			continue
		}
		if criteria.ShortPosition != "" {
			hasPosition := false
			for _, pos := range player.ShortPositions {
				if strings.EqualFold(pos, criteria.ShortPosition) {
					hasPosition = true
					break
				}
			}
			if !hasPosition {
				continue
			}
		}
		if criteria.MinCA > 0 && player.CA < criteria.MinCA {
			continue
		}
		if criteria.MaxCA > 0 && player.CA > criteria.MaxCA {
			continue
		}
		if criteria.MaxBudget > 0 && player.TransferValueAmount > criteria.MaxBudget {
			continue
		}
		if criteria.MaxSalary > 0 && player.WageAmount > criteria.MaxSalary {
			continue
		}
		if criteria.MinAge > 0 || criteria.MaxAge > 0 {
			playerAge, ageErr := strconv.Atoi(player.Age)
			if ageErr != nil {
				continue
			}
			if criteria.MinAge > 0 && playerAge < criteria.MinAge {
				continue
			}
			if criteria.MaxAge > 0 && playerAge > criteria.MaxAge {
				continue
			}
		}
		if criteria.MinOverall > 0 && player.Overall < criteria.MinOverall {
			continue
		}
		matches = append(matches, player)
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].MBR > matches[j].MBR })
	if cap > 0 && len(matches) > cap {
		matches = matches[:cap]
	}
	return matches
}
