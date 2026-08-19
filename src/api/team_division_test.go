package main

import "testing"

func TestCanonicalizePlayerClubMetadata_MajorityWins(t *testing.T) {
	players := []Player{
		{Name: "A", Club: "Montana", Division: "Prva Liga", BasedIn: "Bulgaria"},
		{Name: "B", Club: "Montana", Division: "Prva Liga", BasedIn: "Bulgaria"},
		{Name: "C", Club: "Montana", Division: "Prva Liga", BasedIn: "Bulgaria"},
		// Noisy rows, e.g. players out on loan, that disagree with the rest of the squad.
		{Name: "D", Club: "Montana", Division: "Saudi Premier Division", BasedIn: "Bulgaria"},
		{Name: "E", Club: "Montana", Division: "Bundesliga", BasedIn: "Germany"},
	}

	CanonicalizePlayerClubMetadata(players)

	for _, p := range players {
		if p.Division != "Prva Liga" {
			t.Errorf("player %s: expected canonical Division %q, got %q", p.Name, "Prva Liga", p.Division)
		}
		if p.BasedIn != "Bulgaria" {
			t.Errorf("player %s: expected canonical BasedIn %q, got %q", p.Name, "Bulgaria", p.BasedIn)
		}
	}
}

func TestCanonicalizePlayerClubMetadata_DifferentClubsUnaffected(t *testing.T) {
	players := []Player{
		{Name: "A", Club: "Club X", Division: "League 1", BasedIn: "England"},
		{Name: "B", Club: "Club Y", Division: "League 2", BasedIn: "Wales"},
	}

	CanonicalizePlayerClubMetadata(players)

	if players[0].Division != "League 1" || players[0].BasedIn != "England" {
		t.Errorf("unexpected mutation for Club X: %+v", players[0])
	}
	if players[1].Division != "League 2" || players[1].BasedIn != "Wales" {
		t.Errorf("unexpected mutation for Club Y: %+v", players[1])
	}
}

func TestCanonicalizePlayerClubMetadata_NoClubUntouched(t *testing.T) {
	players := []Player{
		{Name: "Free Agent", Club: "", Division: "", BasedIn: ""},
	}

	CanonicalizePlayerClubMetadata(players)

	if players[0].Division != "" || players[0].BasedIn != "" {
		t.Errorf("expected no-club player to remain untouched, got %+v", players[0])
	}
}

func TestCanonicalizePlayerClubMetadata_TieBreaksOnFirstSeen(t *testing.T) {
	players := []Player{
		{Name: "A", Club: "Split FC", Division: "League A", BasedIn: "Country A"},
		{Name: "B", Club: "Split FC", Division: "League B", BasedIn: "Country B"},
	}

	CanonicalizePlayerClubMetadata(players)

	for _, p := range players {
		if p.Division != "League A" {
			t.Errorf("player %s: expected tie-break to favor first-seen value %q, got %q", p.Name, "League A", p.Division)
		}
	}
}
