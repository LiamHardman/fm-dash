package main

// CanonicalizePlayerClubMetadata normalizes each player's Division and BasedIn fields to
// their club's most common (majority) values.
//
// FM exports routinely contain a handful of rows per club whose Division/BasedIn disagree
// with the rest of the squad - most commonly players out on loan, whose row still lists the
// parent Club but a Division/BasedIn belonging to wherever they're actually playing. Left
// alone, that per-row noise leaks into every place that treats "division" as a club-level
// fact: team aggregation picks whichever stray row it happens to see first, and league
// grouping treats a division as "ambiguous" (splitting it into bogus per-country fragments)
// the moment a single noisy row disagrees on country. Canonicalizing once, right after
// parsing and before storage, means every downstream consumer (leagues, teams, team view,
// Team of the Season, the player table) sees one consistent Division/BasedIn per club.
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
