package main

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
)

// TestDuckDBRoundTieBehavior documents why percentilesSQL deliberately uses
// FLOOR(x+0.5) instead of DuckDB's ROUND(): DuckDB's docs don't pin down
// ROUND's tie-breaking rule, so this is diagnostic only, not an assertion —
// percentilesSQL never calls ROUND() at all.
func TestDuckDBRoundTieBehavior(t *testing.T) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("opening duckdb: %v", err)
	}
	defer func() { _ = db.Close() }()

	var round05, round15, round25, floor05, floor15, floor25 float64
	err = db.QueryRowContext(context.Background(),
		`SELECT round(0.5), round(1.5), round(2.5), floor(0.5+0.5), floor(1.5+0.5), floor(2.5+0.5)`,
	).Scan(&round05, &round15, &round25, &floor05, &floor15, &floor25)
	if err != nil {
		t.Fatalf("querying round/floor behavior: %v", err)
	}
	t.Logf("ROUND(0.5)=%v ROUND(1.5)=%v ROUND(2.5)=%v | FLOOR(x+0.5): %v %v %v (Go math.Round: 1 2 3)",
		round05, round15, round25, floor05, floor15, floor25)
}

// TestComputePercentiles seeds a small, hand-computable dataset via
// writeParquet and checks computePercentiles against values worked out by
// hand from calculatePercentileValue's exact formula, covering: the
// rank-with-ties formula, the -1 no-data sentinel, and multi-group
// membership (broad and detailed).
func TestComputePercentiles(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	players := []PlayerRow{
		{ // sorted position 1 of 4 by Sv %: countSmaller=0, countEqual=1 -> (0+0.5)/4*100=12.5 -> floor(13.0)=13
			PlayerID: 1, Name: "GK Low", ShortPositions: []string{"GK"}, PositionGroups: []string{"Goalkeepers"},
			PerformanceStatsNumeric: map[string]float64{"Sv %": 60},
		},
		{ // tied at 70 with player 3: countSmaller=1, countEqual=2 -> (1+1)/4*100=50.0 -> floor(50.5)=50
			PlayerID: 2, Name: "GK MidA", ShortPositions: []string{"GK"}, PositionGroups: []string{"Goalkeepers"},
			PerformanceStatsNumeric: map[string]float64{"Sv %": 70},
		},
		{
			PlayerID: 3, Name: "GK MidB", ShortPositions: []string{"GK"}, PositionGroups: []string{"Goalkeepers"},
			PerformanceStatsNumeric: map[string]float64{"Sv %": 70},
		},
		{ // countSmaller=3, countEqual=1 -> (3+0.5)/4*100=87.5 -> floor(88.0)=88
			PlayerID: 4, Name: "GK High", ShortPositions: []string{"GK"}, PositionGroups: []string{"Goalkeepers"},
			PerformanceStatsNumeric: map[string]float64{"Sv %": 80},
		},
		{ // no "Sv %" at all -> must be -1, and must NOT affect players 1-4's ranking
			PlayerID: 5, Name: "GK NoData", ShortPositions: []string{"GK"}, PositionGroups: []string{"Goalkeepers"},
			PerformanceStatsNumeric: map[string]float64{},
		},
		{ // multi-group membership: Defenders+Midfielders broad, Full-backs+Wingers detailed
			PlayerID: 6, Name: "Multi", ShortPositions: []string{"DR", "AMR"}, PositionGroups: []string{"Defenders", "Midfielders"},
			PerformanceStatsNumeric: map[string]float64{"Gls/90": 0.5},
		},
	}

	if err := writeParquet(ctx, dir, "percentile-test", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computePercentiles(ctx, dir, "percentile-test")
	if err != nil {
		t.Fatalf("computePercentiles failed: %v", err)
	}

	cases := []struct {
		playerID int64
		group    string
		stat     string
		want     float64
	}{
		{1, "Global", "Sv %", 13},
		{1, "Goalkeepers", "Sv %", 13},
		{2, "Goalkeepers", "Sv %", 50},
		{3, "Goalkeepers", "Sv %", 50},
		{4, "Goalkeepers", "Sv %", 88},
		{5, "Goalkeepers", "Sv %", -1}, // no data for this player -> -1
	}
	for _, c := range cases {
		groups, ok := result[c.playerID]
		if !ok {
			t.Fatalf("player %d missing from result entirely", c.playerID)
		}
		stats, ok := groups[c.group]
		if !ok {
			t.Fatalf("player %d missing group %q (has groups: %v)", c.playerID, c.group, keysOf(groups))
		}
		got, ok := stats[c.stat]
		if !ok {
			t.Fatalf("player %d group %q missing stat %q", c.playerID, c.group, c.stat)
		}
		if got != c.want {
			t.Errorf("player %d group %q stat %q = %v, want %v", c.playerID, c.group, c.stat, got, c.want)
		}
	}

	// Player 5 having no data must not shift players 1-4's percentiles -
	// already asserted above via the exact expected values, which were
	// hand-computed over the 4-player eligible population (n=4), not 5.

	// Multi-group membership: player 6 must appear under both broad groups
	// and both detailed groups it belongs to, plus Global.
	p6Groups := keysOf(result[6])
	wantGroups := []string{"Global", "Defenders", "Midfielders", "Full-backs", "Wingers"}
	for _, g := range wantGroups {
		if _, ok := result[6][g]; !ok {
			t.Errorf("player 6 missing expected group %q (has: %v)", g, p6Groups)
		}
	}

	// Sole member of a group -> percentile 50 for their stat there.
	if got := result[6]["Full-backs"]["Gls/90"]; got != 50 {
		t.Errorf("player 6 Full-backs Gls/90 = %v, want 50 (sole member)", got)
	}
}

func keysOf(m map[string]map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
