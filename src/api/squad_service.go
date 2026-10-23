package main

import (
	_ "embed"
	"encoding/json"
	"math"
	"sort"
	"strings"
)

//go:embed public/cli_formations.json
var squadFormationJSON []byte

type FormationSlot struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}
type SquadAssignment struct {
	Player  Player `json:"player"`
	Overall int    `json:"overallInRole"`
	Exact   bool   `json:"exactMatch"`
}
type SquadComposition struct {
	Slots   map[string][]SquadAssignment `json:"squadComposition"`
	Average int                          `json:"bestTeamAverageOverall"`
}

func squadRoleScore(p Player, role string) int {
	if len(p.RoleSpecificOveralls) == 0 {
		return p.Overall
	}
	exact, fallback, matchers, prefixes := nationSelectionRoleMaps()
	best := func(positions []string) int {
		score := 0
		for _, r := range p.RoleSpecificOveralls {
			base := strings.TrimSpace(strings.Split(r.RoleName, " - ")[0])
			for _, pos := range positions {
				if pos == base {
					score = max(score, r.Score)
				}
			}
		}
		return score
	}
	intersect := func(positions []string) []string {
		found := []string{}
		for _, pos := range p.ShortPositions {
			for _, v := range positions {
				if pos == v {
					found = append(found, pos)
				}
			}
		}
		return found
	}
	role = strings.ToUpper(role)
	if matches := intersect(exact[role]); len(matches) > 0 {
		if score := best(matches); score > 0 {
			return score
		}
		return max(10, p.Overall)
	}
	if matches := intersect(fallback[role]); len(matches) > 0 {
		if score := best(matches); score > 0 {
			return score
		}
		return max(0, p.Overall-5)
	}
	targets := []string{}
	for _, m := range matchers[role] {
		if prefix := prefixes[strings.ToUpper(m)]; prefix != "" {
			targets = append(targets, prefix)
		}
	}
	if score := best(targets); score > 0 {
		return score
	}
	return max(0, p.Overall-10)
}

// ComputeSquadComposition follows useBestXI.js: scarce exact positions first,
// exact/fallback passes at each of three depths, then an orphan-slot pass.
func ComputeSquadComposition(players []Player, slots []FormationSlot) SquadComposition {
	exact, fallback, _, _ := nationSelectionRoleMaps()
	result := SquadComposition{Slots: map[string][]SquadAssignment{}}
	type assignment struct {
		slot   string
		player int
		score  int
		exact  bool
		sort   int
	}
	assignments := []assignment{}
	intersects := func(positions, allowed []string) bool {
		for _, p := range positions {
			for _, v := range allowed {
				if p == v {
					return true
				}
			}
		}
		return false
	}
	// The GUI keys position/assignment identity by name. Preserve that contract
	// here; durable Player references continue to use Dataset ID + unique UID.
	positions := map[string][]string{}
	for _, p := range players {
		positions[p.Name] = p.ShortPositions
	}
	for _, slot := range slots {
		result.Slots[slot.ID] = []SquadAssignment{}
		for i, p := range players {
			score := squadRoleScore(p, slot.Role)
			if score < 10 {
				continue
			}
			e := intersects(positions[p.Name], exact[strings.ToUpper(slot.Role)])
			f := intersects(positions[p.Name], fallback[strings.ToUpper(slot.Role)])
			if e || f {
				priority := score - 5000
				if e {
					priority = score + 10000
				}
				assignments = append(assignments, assignment{slot.ID, i, score, e, priority})
			}
		}
	}
	sort.SliceStable(assignments, func(i, j int) bool { return assignments[i].sort > assignments[j].sort })
	ordered := append([]FormationSlot{}, slots...)
	count := func(slot FormationSlot) int {
		n := 0
		for _, p := range players {
			if intersects(p.ShortPositions, exact[strings.ToUpper(slot.Role)]) {
				n++
			}
		}
		return n
	}
	sort.SliceStable(ordered, func(i, j int) bool { return count(ordered[i]) < count(ordered[j]) })
	used := map[string]bool{}
	for depth := 0; depth < 3; depth++ {
		for _, exactPass := range []bool{true, false} {
			for _, slot := range ordered {
				if len(result.Slots[slot.ID]) != depth {
					continue
				}
				for _, a := range assignments {
					p := players[a.player]
					if a.slot == slot.ID && a.exact == exactPass && !used[p.Name] {
						result.Slots[slot.ID] = append(result.Slots[slot.ID], SquadAssignment{p, a.score, a.exact})
						used[p.Name] = true
						break
					}
				}
			}
		}
	}
	for id := range result.Slots {
		sort.SliceStable(result.Slots[id], func(i, j int) bool {
			a, b := result.Slots[id][i], result.Slots[id][j]
			if a.Exact != b.Exact {
				return a.Exact
			}
			return a.Overall > b.Overall
		})
	}
	for _, slot := range slots {
		if len(result.Slots[slot.ID]) > 0 {
			continue
		}
		bestIndex, bestScore := -1, -1
		for i, p := range players {
			if used[p.Name] || !intersects(p.ShortPositions, fallback[strings.ToUpper(slot.Role)]) {
				continue
			}
			if score := squadRoleScore(p, slot.Role); score > bestScore {
				bestIndex, bestScore = i, score
			}
		}
		if bestIndex >= 0 {
			p := players[bestIndex]
			result.Slots[slot.ID] = append(result.Slots[slot.ID], SquadAssignment{p, bestScore, false})
			used[p.Name] = true
		}
	}
	sum, n := 0, 0
	for _, slot := range slots {
		if len(result.Slots[slot.ID]) > 0 {
			sum += result.Slots[slot.ID][0].Overall
			n++
		}
	}
	if n > 0 {
		result.Average = int(math.Floor(float64(sum)/float64(n) + .5))
	}
	return result
}

func cliSquad(players []Player, o cliOptions) (any, error) {
	var formations map[string]struct {
		Name   string `json:"name"`
		Layout []struct {
			Positions []FormationSlot `json:"positions"`
		} `json:"layout"`
	}
	if err := json.Unmarshal(squadFormationJSON, &formations); err != nil {
		return nil, err
	}
	if o.name == "" {
		return formations, nil
	}
	f, ok := formations[o.name]
	if !ok {
		return nil, cliFail(2, "unknown formation %q; omit --name to list formations", o.name)
	}
	slots := []FormationSlot{}
	for _, row := range f.Layout {
		slots = append(slots, row.Positions...)
	}
	if o.query != "" {
		var filters map[string]any
		if err := cliDecodeJSON([]byte(o.query), &filters); err != nil {
			return nil, err
		}
		var err error
		players, err = cliFilterPlayers(players, filters)
		if err != nil {
			return nil, err
		}
	}
	return ComputeSquadComposition(players, slots), nil
}
