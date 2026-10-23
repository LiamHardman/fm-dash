package main

// playersWithClubMetadata returns a private projection for club-level grouping.
// Only string fields change; shared maps and slices must remain read-only.
// Player records in parsing, storage and player responses keep their source values.
func playersWithClubMetadata(players []Player) []Player {
	projection := make([]Player, len(players))
	copy(projection, players)
	CanonicalizePlayerClubMetadata(projection)
	return projection
}

// CanonicalizePlayerClubMetadata normalizes a private aggregation projection to
// each club's most common Division and BasedIn values. Loan rows can disagree
// with club-level metadata; never apply this transformation to source records.
func CanonicalizePlayerClubMetadata(players []Player) {
	type voteCounts struct {
		counts    map[string]int
		bestValue string
		bestCount int
	}

	divisionVotes := make(map[string]*voteCounts)
	basedInVotes := make(map[string]*voteCounts)

	tally := func(votes map[string]*voteCounts, club, value string) {
		if value == "" {
			return
		}
		v, ok := votes[club]
		if !ok {
			v = &voteCounts{counts: make(map[string]int)}
			votes[club] = v
		}
		v.counts[value]++
		// Strict ">" means the first value to reach a given count keeps priority on ties,
		// so the result is deterministic regardless of map iteration order.
		if v.counts[value] > v.bestCount {
			v.bestCount = v.counts[value]
			v.bestValue = value
		}
	}

	for i := range players {
		club := players[i].Club
		if club == "" {
			continue
		}
		tally(divisionVotes, club, players[i].Division)
		tally(basedInVotes, club, players[i].BasedIn)
	}

	for i := range players {
		club := players[i].Club
		if club == "" {
			continue
		}
		if v, ok := divisionVotes[club]; ok && v.bestValue != "" {
			players[i].Division = v.bestValue
		}
		if v, ok := basedInVotes[club]; ok && v.bestValue != "" {
			players[i].BasedIn = v.bestValue
		}
	}
}
