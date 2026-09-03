package main

import (
	"context"
	"math"
	"testing"
)

// computeNonLinearScalingRef is an independent Go re-derivation of
// computeNonLinearScaling (src/api/calculations.go) from its documented
// formula, used to compute expected test values programmatically instead of
// hand-transcribing POWER()-curve arithmetic (error-prone to do by hand).
func computeNonLinearScalingRef(linearRating float64) int {
	if linearRating <= 0 {
		return 0
	}
	if linearRating >= 99 {
		return 99
	}
	const inflectionPoint = 75.0
	if linearRating >= inflectionPoint {
		return int(math.Round(inflectionPoint + (linearRating-inflectionPoint)*0.95))
	}
	normalized := linearRating / inflectionPoint
	scaled := math.Pow(normalized, 1.8) * inflectionPoint
	if scaled < 10 && linearRating > 20 {
		scaled = 10 + (linearRating-20)*0.15
	}
	return int(math.Round(scaled))
}

// weightedAverageRef is an independent Go re-derivation of
// calculateWeightedAverage (src/api/calculations.go).
func weightedAverageRef(attrs map[string]int, weights map[string]int) float64 {
	var weightedSum, totalWeight int64
	for attr, weight := range weights {
		v, ok := attrs[attr]
		if ok && v >= 1 && v <= 20 {
			weightedSum += int64(v * weight)
			totalWeight += int64(weight)
		}
	}
	if totalWeight == 0 {
		return 0
	}
	return float64(weightedSum) / float64(totalWeight)
}

func fifaStatRef(attrs map[string]int, weights map[string]int) int {
	avg := weightedAverageRef(attrs, weights)
	if avg == 0 {
		return 0
	}
	return computeNonLinearScalingRef(avg * 5.3)
}

func TestComputeFifaStatsGoalkeeperOutfieldSplit(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	pacWeights := map[string]int{"Acc": 10}
	gkWeights := map[string]int{"Han": 10}
	if err := store.SetCategories(ctx, map[string]map[string]int{
		"PAC": pacWeights,
		"GK":  gkWeights,
	}); err != nil {
		t.Fatalf("SetCategories failed: %v", err)
	}

	attrs := map[string]int{"Acc": 15, "Han": 20}
	players := []PlayerRow{
		{PlayerID: 1, PositionGroups: []string{"Defenders"}, NumericAttributes: attrs},
		{PlayerID: 2, PositionGroups: []string{"Goalkeepers"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "fifa-test-split", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeFifaStats(ctx, store, dir, "fifa-test-split")
	if err != nil {
		t.Fatalf("computeFifaStats failed: %v", err)
	}

	wantPAC := fifaStatRef(attrs, pacWeights)
	wantGK := fifaStatRef(attrs, gkWeights)

	if result[1].PAC != wantPAC {
		t.Errorf("player 1 (outfield) PAC = %d, want %d", result[1].PAC, wantPAC)
	}
	if result[1].GK != 0 {
		t.Errorf("player 1 (outfield) GK = %d, want 0", result[1].GK)
	}
	if result[2].GK != wantGK {
		t.Errorf("player 2 (goalkeeper) GK = %d, want %d", result[2].GK, wantGK)
	}
	if result[2].PAC != 0 {
		t.Errorf("player 2 (goalkeeper) PAC = %d, want 0", result[2].PAC)
	}
}

func TestComputeFifaStatsPasThreeWayMax(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// PAS_standard and PAS_no_set_pieces reference attributes the player
	// lacks (score forced to 0); PAS_no_off_ball references one they have.
	stdWeights := map[string]int{"Cor": 10}
	noSetPiecesWeights := map[string]int{"Vis": 10}
	noOffBallWeights := map[string]int{"Pas": 10}
	if err := store.SetCategories(ctx, map[string]map[string]int{
		"PAS_standard":      stdWeights,
		"PAS_no_set_pieces": noSetPiecesWeights,
		"PAS_no_off_ball":   noOffBallWeights,
	}); err != nil {
		t.Fatalf("SetCategories failed: %v", err)
	}

	attrs := map[string]int{"Pas": 12}
	players := []PlayerRow{
		{PlayerID: 1, PositionGroups: []string{"Midfielders"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "fifa-test-pas", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeFifaStats(ctx, store, dir, "fifa-test-pas")
	if err != nil {
		t.Fatalf("computeFifaStats failed: %v", err)
	}

	wantStd := fifaStatRef(attrs, stdWeights)                 // 0 - player has no "Cor"
	wantNoSetPieces := fifaStatRef(attrs, noSetPiecesWeights) // 0 - player has no "Vis"
	wantNoOffBall := fifaStatRef(attrs, noOffBallWeights)     // nonzero - player has "Pas"

	if wantStd != 0 || wantNoSetPieces != 0 {
		t.Fatalf("test setup invalid: expected the two decoy sub-scores to be 0, got std=%d noSetPieces=%d", wantStd, wantNoSetPieces)
	}
	if wantNoOffBall == 0 {
		t.Fatalf("test setup invalid: expected PAS_no_off_ball to be nonzero, got 0")
	}

	if result[1].PAS != wantNoOffBall {
		t.Errorf("PAS = %d, want %d (max of %d, %d, %d)", result[1].PAS, wantNoOffBall, wantStd, wantNoSetPieces, wantNoOffBall)
	}
}

func TestComputeFifaStatsMissingCategoryYieldsZero(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Only seed PAC - every other category has zero weight rows.
	if err := store.SetCategories(ctx, map[string]map[string]int{"PAC": {"Acc": 10}}); err != nil {
		t.Fatalf("SetCategories failed: %v", err)
	}

	players := []PlayerRow{
		{PlayerID: 1, PositionGroups: []string{"Defenders"}, NumericAttributes: map[string]int{"Acc": 15, "Fin": 10}},
	}
	if err := writeParquet(ctx, dir, "fifa-test-missing", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeFifaStats(ctx, store, dir, "fifa-test-missing")
	if err != nil {
		t.Fatalf("computeFifaStats failed: %v", err)
	}

	if result[1].SHO != 0 {
		t.Errorf("SHO (no weight rows at all) = %d, want 0", result[1].SHO)
	}
	if result[1].DRI != 0 {
		t.Errorf("DRI (no weight rows at all) = %d, want 0", result[1].DRI)
	}
}
