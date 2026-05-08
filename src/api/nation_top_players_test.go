package main

import "testing"

func nationTestPlayer(uid int64, name string, positions []string, overall int) Player {
	return Player{
		UID:            uid,
		Name:           name,
		Nationality:    "England",
		ShortPositions: positions,
		Overall:        overall,
	}
}

func TestSelectTopNationPlayersCapsAndDedupesPositionPool(t *testing.T) {
	players := []Player{
		{UID: 1, Name: "Keeper 1", Nationality: "England", ShortPositions: []string{"GK"}, Overall: 80},
		{UID: 2, Name: "Keeper 2", Nationality: "England", ShortPositions: []string{"GK"}, Overall: 79},
		{UID: 3, Name: "Centre Back", Nationality: "England", ShortPositions: []string{"DC"}, Overall: 82},
		{UID: 4, Name: "Versatile Mid", Nationality: "England", ShortPositions: []string{"MC", "AMC"}, Overall: 85},
		{UID: 5, Name: "Striker", Nationality: "England", ShortPositions: []string{"ST"}, Overall: 88},
		{UID: 6, Name: "Other Nation", Nationality: "France", ShortPositions: []string{"ST"}, Overall: 99},
	}

	selected := selectTopNationPlayers(players, "England", 4)

	if len(selected) != 4 {
		t.Fatalf("selected player count = %d, want 4", len(selected))
	}

	seen := map[int64]bool{}
	for _, player := range selected {
		if player.Nationality != "England" {
			t.Fatalf("selected player from wrong nation: %s", player.Nationality)
		}
		if seen[player.UID] {
			t.Fatalf("selected duplicate player UID %d", player.UID)
		}
		seen[player.UID] = true
	}
}

func TestSelectTopNationPlayersPreservesBestFormationDepth(t *testing.T) {
	players := []Player{}
	nextUID := int64(1)
	addPositionPlayers := func(position string, count int, startOverall int) {
		for i := 0; i < count; i++ {
			players = append(players, nationTestPlayer(
				nextUID,
				position+" Player",
				[]string{position},
				startOverall-i,
			))
			nextUID++
		}
	}

	addPositionPlayers("GK", 3, 70)
	addPositionPlayers("DR", 3, 69)
	addPositionPlayers("DL", 3, 68)
	addPositionPlayers("DC", 6, 67)
	addPositionPlayers("MR", 3, 66)
	addPositionPlayers("ML", 3, 65)
	addPositionPlayers("MC", 6, 64)
	addPositionPlayers("ST", 30, 99)

	selected := selectTopNationPlayers(players, "England", 35)

	if len(selected) != 35 {
		t.Fatalf("selected player count = %d, want 35", len(selected))
	}

	counts := map[string]int{}
	for _, player := range selected {
		for _, position := range player.ShortPositions {
			counts[position]++
		}
	}

	expectedMinimums := map[string]int{
		"GK": 2,
		"DR": 2,
		"DL": 2,
		"DC": 4,
		"MR": 2,
		"ML": 2,
		"MC": 4,
		"ST": 4,
	}
	for position, minimum := range expectedMinimums {
		if counts[position] < minimum {
			t.Fatalf("selected %s count = %d, want at least %d; counts=%v", position, counts[position], minimum, counts)
		}
	}
}

func TestSelectTopNationPlayersReturnsEmptyForUnknownNation(t *testing.T) {
	selected := selectTopNationPlayers([]Player{
		{UID: 1, Name: "Keeper", Nationality: "England", ShortPositions: []string{"GK"}, Overall: 80},
	}, "France", 35)

	if len(selected) != 0 {
		t.Fatalf("selected player count = %d, want 0", len(selected))
	}
}
