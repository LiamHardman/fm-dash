package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"testing"

	duckdb "github.com/duckdb/duckdb-go/v2"
)

// casRef is an independent Go re-derivation of CalculateCAS/
// calculateCASForPosition (src/api/ca_calculation.go:151-227), built on the
// real math.Round -- not a hand-reimplementation of rounding. Every
// attribute with a nonzero weight always contributes: missing/absent or
// <=0 values default to 1, present values are used as-is if in [1,20] or
// clamped down to 20 if >20 (never clamped up).
func casRef(attrs map[string]int, weights map[string]float64) int {
	var weightedSum, totalWeight float64
	for attr, weight := range weights {
		effective := 1.0
		if v, ok := attrs[attr]; ok && v > 0 {
			effective = float64(v)
			if effective > 20 {
				effective = 20
			}
		}
		weightedSum += effective * weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return 0
	}
	weightedAvg := weightedSum / totalWeight
	ca := 19.8*weightedAvg - 117.8
	rounded := math.Round(ca)
	if rounded < 1 {
		rounded = 1
	}
	if rounded > 200 {
		rounded = 200
	}
	return int(rounded)
}

// casMaxRef is an independent Go re-derivation of CalculateCAS's
// max-across-ShortPositions step: positions with no cas_weights entry
// contribute nothing (not a 0 candidate), and an empty positionScores
// (e.g. from empty ShortPositions) yields 0.
func casMaxRef(positionScores map[string]int) int {
	max := 0
	for _, s := range positionScores {
		if s > max {
			max = s
		}
	}
	return max
}

// seedCasWeightsForTest inserts a custom cas-weights set directly,
// bypassing seedCasWeightsIfEmpty's "only if empty" / embedded-JSON path --
// used to test computeCAS against controlled, hand-designed positions.
func seedCasWeightsForTest(t *testing.T, ctx context.Context, store *WeightsStore, casWeights map[string]map[string]float64) {
	t.Helper()
	for position, weights := range casWeights {
		for attr, weight := range weights {
			if _, err := store.db.ExecContext(ctx,
				`INSERT INTO cas_weights (position, attribute, weight) VALUES (?, ?, ?)`,
				position, attr, weight,
			); err != nil {
				t.Fatalf("seeding cas weight %s.%s: %v", position, attr, err)
			}
		}
	}
}

func TestComputeCASDefaultToOneNotExclude(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// A is present, B is entirely absent from the player's attributes.
	// Excluding B (like fifaStatsSQL/roleOverallsSQL would) gives
	// weightedAvg=20 (A alone); defaulting B to 1 (the correct CAS
	// behavior) gives weightedAvg=(20*10+1*10)/20=10.5 -- these produce
	// different final scores, so the test can actually catch a
	// copy-pasted exclude/skip filter from the other two endpoints.
	weights := map[string]float64{"A": 10, "B": 10}
	attrs := map[string]int{"A": 20}
	seedCasWeightsForTest(t, ctx, store, map[string]map[string]float64{"TESTPOS": weights})

	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"TESTPOS"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "cas-test-default", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeCAS(ctx, store, dir, "cas-test-default")
	if err != nil {
		t.Fatalf("computeCAS failed: %v", err)
	}

	want := casRef(attrs, weights)
	excludeWant := casRef(map[string]int{"A": 20}, map[string]float64{"A": 10}) // sanity-check: what excluding B entirely would give
	if want == excludeWant {
		t.Fatalf("test setup invalid: default-to-1 (%d) and exclude (%d) baselines coincide", want, excludeWant)
	}
	if got := result[1]; got != want {
		t.Errorf("CAS = %d, want %d (default-missing-to-1, not exclude)", got, want)
	}
}

func TestComputeCASClampAboveTwentyNotExclude(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// B=25 is outside the normal [1,20] attribute domain. Clamping B to 20
	// (correct) vs. using it unclamped (incorrect) produce different final
	// scores after the [1,200] output clamp -- chosen so the two paths
	// don't collide at the output clamp boundary like a naive single-
	// attribute test would.
	weights := map[string]float64{"A": 10, "B": 10}
	attrs := map[string]int{"A": 10, "B": 25}
	seedCasWeightsForTest(t, ctx, store, map[string]map[string]float64{"TESTPOS": weights})

	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"TESTPOS"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "cas-test-clamp", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeCAS(ctx, store, dir, "cas-test-clamp")
	if err != nil {
		t.Fatalf("computeCAS failed: %v", err)
	}

	want := casRef(attrs, weights)
	if got := result[1]; got != want {
		t.Errorf("CAS = %d, want %d (clamp down to 20, not exclude)", got, want)
	}
}

// TestCasRoundHalfAwayFromZeroSQLFragment runs the exact
// casRoundHalfAwayFromZeroSQL text (the same fragment embedded verbatim
// into casSQL) against a set of values including an exact negative
// half-integer tie, asserting against math.Round directly. This is the
// only test that can actually distinguish a correct sign-aware rounding
// implementation from an incorrect FLOOR(x+0.5)-only one, since the
// end-to-end [1,200] output clamp makes every negative near-tie collapse
// to the same clamped answer regardless of which rounding rule produced it.
func TestCasRoundHalfAwayFromZeroSQLFragment(t *testing.T) {
	ctx := context.Background()
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		t.Fatalf("creating duckdb connector: %v", err)
	}
	defer func() { _ = connector.Close() }()
	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

	values := []float64{-48.5, 48.5, -0.5, 0.5, -2.5, 2.5, -117.8, 0, -1, 1, -99.4, 99.4}

	expr := fmt.Sprintf(casRoundHalfAwayFromZeroSQL, "val", "val", "val")
	query := fmt.Sprintf("SELECT val, %s AS rounded FROM (VALUES %s) AS t(val)",
		expr, valuesListSQL(values))

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		t.Fatalf("querying rounding fragment: %v", err)
	}
	defer func() { _ = rows.Close() }()

	got := make(map[float64]int)
	for rows.Next() {
		var val float64
		var rounded int
		if err := rows.Scan(&val, &rounded); err != nil {
			t.Fatalf("scanning rounding row: %v", err)
		}
		got[val] = rounded
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterating rounding rows: %v", err)
	}

	for _, v := range values {
		want := int(math.Round(v))
		if got[v] != want {
			t.Errorf("round-half-away-from-zero(%v) = %d, want %d (math.Round)", v, got[v], want)
		}
	}
}

// valuesListSQL renders a []float64 as a DuckDB VALUES-list body, e.g.
// "(-48.5), (48.5)".
func valuesListSQL(values []float64) string {
	s := ""
	for i, v := range values {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("(%v)", v)
	}
	return s
}

// TestComputeCASEndToEndDeepNegativeClamp confirms a real player whose
// weightedAvg bottoms out at 1 (every weighted attribute missing, so all
// default to 1) produces ca_raw=19.8*1-117.8=-98, which clamps to the
// output floor of 1. This is a sanity/regression test on the full pipeline,
// not itself a tie-break test -- see
// TestCasRoundHalfAwayFromZeroSQLFragment for that.
func TestComputeCASEndToEndDeepNegativeClamp(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	weights := map[string]float64{"A": 10, "B": 5}
	attrs := map[string]int{} // both A and B entirely absent -> default to 1
	seedCasWeightsForTest(t, ctx, store, map[string]map[string]float64{"TESTPOS": weights})

	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"TESTPOS"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "cas-test-deep-negative", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeCAS(ctx, store, dir, "cas-test-deep-negative")
	if err != nil {
		t.Fatalf("computeCAS failed: %v", err)
	}

	want := casRef(attrs, weights)
	if want != 1 {
		t.Fatalf("test setup invalid: expected reference score of 1 (clamped from deep negative), got %d", want)
	}
	if got := result[1]; got != want {
		t.Errorf("CAS = %d, want %d (deep-negative ca_raw clamped to output floor of 1)", got, want)
	}
}

func TestComputeCASMultiPositionTakesMax(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Two positions keyed on different attributes so their scores diverge
	// unambiguously (chosen to not tie) for the same player.
	weightsLow := map[string]float64{"A": 10}
	weightsHigh := map[string]float64{"B": 10}
	seedCasWeightsForTest(t, ctx, store, map[string]map[string]float64{
		"POSLOW":  weightsLow,
		"POSHIGH": weightsHigh,
	})

	attrs := map[string]int{"A": 5, "B": 18}

	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"POSLOW", "POSHIGH"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "cas-test-multipos", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeCAS(ctx, store, dir, "cas-test-multipos")
	if err != nil {
		t.Fatalf("computeCAS failed: %v", err)
	}

	scoreLow := casRef(attrs, weightsLow)
	scoreHigh := casRef(attrs, weightsHigh)
	if scoreLow == scoreHigh {
		t.Fatalf("test setup invalid: expected POSLOW (%d) and POSHIGH (%d) scores to differ", scoreLow, scoreHigh)
	}
	want := casMaxRef(map[string]int{"POSLOW": scoreLow, "POSHIGH": scoreHigh})

	if got := result[1]; got != want {
		t.Errorf("CAS = %d, want %d (max across ShortPositions: POSLOW=%d POSHIGH=%d)", got, want, scoreLow, scoreHigh)
	}
}

func TestComputeCASEmptyAndUnrecognizedPositions(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	seedCasWeightsForTest(t, ctx, store, map[string]map[string]float64{
		"KNOWNPOS": {"A": 10},
	})

	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{}, NumericAttributes: map[string]int{"A": 15}},
		{PlayerID: 2, ShortPositions: []string{"UNKNOWNPOS"}, NumericAttributes: map[string]int{"A": 15}},
	}
	if err := writeParquet(ctx, dir, "cas-test-empty-unknown", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeCAS(ctx, store, dir, "cas-test-empty-unknown")
	if err != nil {
		t.Fatalf("computeCAS failed: %v", err)
	}

	if got := result[1]; got != 0 {
		t.Errorf("player with empty ShortPositions: CAS = %d, want 0", got)
	}
	if got := result[2]; got != 0 {
		t.Errorf("player with an unrecognized position: CAS = %d, want 0", got)
	}
}

// TestCasWeightsSpotCheck is a permanent drift tripwire: cas_weights.json
// has no independent source of truth to diff against later (unlike
// role_weights.json, which is a verbatim copy of a real repo file), so this
// asserts a handful of values hand-read directly from the
// casPositionWeights Go literal in src/api/ca_calculation.go at the time of
// extraction. If this ever fails, casPositionWeights was edited without
// re-running the extraction program that produced cas_weights.json.
func TestCasWeightsSpotCheck(t *testing.T) {
	weights, err := parseEmbeddedCasWeights()
	if err != nil {
		t.Fatalf("parseEmbeddedCasWeights failed: %v", err)
	}

	wantPositions := []string{"AMC", "AML", "AMR", "DC", "DL", "DM", "DR", "GK", "MC", "ML", "MR", "ST", "SW", "WBL", "WBR"}
	if len(weights) != len(wantPositions) {
		t.Errorf("cas_weights.json has %d positions, want %d", len(weights), len(wantPositions))
	}
	for _, pos := range wantPositions {
		if _, ok := weights[pos]; !ok {
			t.Errorf("cas_weights.json is missing position %q", pos)
		}
	}

	spotChecks := []struct {
		position, attribute string
		wantWeight          float64
	}{
		{"GK", "Cmd", 6},
		{"ST", "Fin", 8},
		{"DC", "Str", 6},
		{"AMC", "OtB", 3},
	}
	for _, sc := range spotChecks {
		got, ok := weights[sc.position][sc.attribute]
		if !ok {
			t.Errorf("cas_weights.json missing %s.%s", sc.position, sc.attribute)
			continue
		}
		if got != sc.wantWeight {
			t.Errorf("%s.%s weight = %v, want %v", sc.position, sc.attribute, got, sc.wantWeight)
		}
	}

	if _, ok := weights["GK"]["Ecc"]; ok {
		t.Errorf(`cas_weights.json should NOT contain GK.Ecc (weight was 0 in casPositionWeights, zero-weight entries are omitted)`)
	}

	if weights["DR"] == nil || weights["DL"] == nil {
		t.Fatalf("expected both DR and DL to be present")
	}
	if len(weights["DR"]) != len(weights["DL"]) {
		t.Errorf("DR and DL should have identical weight maps (DR mirrors DL in casPositionWeights), got %d vs %d entries", len(weights["DR"]), len(weights["DL"]))
	}
	for attr, w := range weights["DL"] {
		if weights["DR"][attr] != w {
			t.Errorf("DR.%s = %v, want %v (should mirror DL)", attr, weights["DR"][attr], w)
		}
	}

	if weights["SW"] == nil || weights["DC"] == nil {
		t.Fatalf("expected both SW and DC to be present")
	}
	for attr, w := range weights["DC"] {
		if weights["SW"][attr] != w {
			t.Errorf("SW.%s = %v, want %v (should mirror DC)", attr, weights["SW"][attr], w)
		}
	}
}
