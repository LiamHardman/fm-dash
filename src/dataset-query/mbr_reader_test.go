package main

import (
	"context"
	"database/sql"
	"math"
	"strconv"
	"testing"

	duckdb "github.com/duckdb/duckdb-go/v2"
)

// TestDuckDBTruncBehavior confirms DuckDB's TRUNC() truncates toward zero
// (matching Go's int(floatVal) conversion) rather than rounding to nearest
// -- unlike a bare CAST(DOUBLE AS INTEGER), which DuckDB rounds. mbrSQL
// depends on this for both the value-score term and the final
// normalization step, since (unlike role_overalls/fifa_stats/cas) neither
// has a FLOOR/CEIL applied before the cast in the original Go source.
func TestDuckDBTruncBehavior(t *testing.T) {
	ctx := context.Background()
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		t.Fatalf("creating duckdb connector: %v", err)
	}
	defer func() { _ = connector.Close() }()
	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

	cases := []struct {
		input float64
		want  int
	}{
		{2.7, 2},
		{-2.7, -2},
		{2.999999, 2},
		{-2.999999, -2},
		{0.5, 0},
		{-0.5, 0},
		{0, 0},
		{5.0, 5},
		{-5.0, -5},
	}
	for _, c := range cases {
		var got int
		if err := db.QueryRowContext(ctx, `SELECT CAST(TRUNC(?) AS INTEGER)`, c.input).Scan(&got); err != nil {
			t.Fatalf("querying TRUNC(%v): %v", c.input, err)
		}
		if got != c.want {
			t.Errorf("TRUNC(%v) = %d, want %d (truncate toward zero)", c.input, got, c.want)
		}
	}
}

// ageModifierRef is an independent Go re-derivation of getAgeModifier
// (src/api/calculations.go).
func ageModifierRef(age int) int {
	switch {
	case age >= 16 && age <= 18:
		return 30
	case age == 19:
		return 25
	case age == 20:
		return 22
	case age == 21:
		return 18
	case age == 22:
		return 15
	case age == 23:
		return 10
	case age == 24:
		return 6
	case age == 25:
		return 3
	case age == 26:
		return -1
	case age == 27:
		return -3
	case age == 28:
		return -5
	case age == 29:
		return -10
	case age == 30:
		return -15
	case age == 31:
		return -20
	default:
		if age > 31 {
			return -20 - (age-31)*2
		}
		return 0
	}
}

// mentalityModifierRef is an independent Go re-derivation of
// getMentalityModifier (src/api/calculations.go), INCLUDING its dead
// "Okay personalities" switch (identical case list to "Good", so it can
// never fire since Good already returns first) -- this reference
// deliberately reproduces that shadowing rather than "fixing" it, since
// mbrSQL must match the real (buggy) behavior bug-for-bug.
func mentalityModifierRef(personality string) int {
	switch personality {
	case "Model Citizen", "Model Professional":
		return 20
	}
	switch personality {
	case "Perfectionist", "Resolute", "Professional", "Fairly Professional",
		"Iron Willed", "Resillient", "Spirited", "Driven", "Determined",
		"Fairly Determined", "Charismatic Leader", "Born Leader", "Leader",
		"Very Ambitious", "Ambitious", "Fairly Ambitious":
		return 15
	}
	switch personality {
	case "Balanced", "Light-Hearted", "Jovial", "Very Loyal", "Loyal",
		"Fairly Loyal", "Honest", "Sporting", "Fairly Sporting":
		return 8
	}
	switch personality {
	case "Fickle", "Mercenary", "Unambitious", "Unsporting", "Realist":
		return -15
	}
	switch personality {
	case "Slack", "Casual", "Temperamental", "Spineless", "Low Self-Belief",
		"Easily Discouraged", "Low Determination":
		return -35
	}
	return 0
}

// expectedValuePerRatingRef is an independent Go re-derivation of
// getExpectedValuePerRating (src/api/handlers.go:3178).
func expectedValuePerRatingRef(overall float64) float64 {
	switch {
	case overall >= 85:
		return 1.2
	case overall >= 80:
		return 0.8
	case overall >= 75:
		return 0.5
	case overall >= 70:
		return 0.3
	case overall >= 65:
		return 0.2
	case overall >= 60:
		return 0.15
	default:
		return 0.1
	}
}

// salaryPenaltyRef is an independent Go re-derivation of getSalaryPenalty
// (src/api/calculations.go). Uses ordinary float64 division throughout --
// when transferValueAmount=0 (and wageAmount!=0, guaranteed by the caller
// at that point), expectedWeeklySalary is 0 and Go's IEEE-754 float
// division naturally produces +Inf for salaryRatio, landing in the >=3.0
// tier below -- exactly like the real production code, with no special-
// casing needed in this reference.
func salaryPenaltyRef(transferValueAmount, wageAmount int64) int {
	if wageAmount == 0 {
		return 0
	}
	transferValueMillions := float64(transferValueAmount) / 1000000.0
	wageThousandsPerWeek := float64(wageAmount) / 1000.0
	expectedWeeklySalary := (transferValueMillions * 0.12 * 1000) / 52
	salaryRatio := wageThousandsPerWeek / expectedWeeklySalary
	switch {
	case salaryRatio >= 3.0:
		return -15
	case salaryRatio >= 2.5:
		return -10
	case salaryRatio >= 2.0:
		return -7
	case salaryRatio >= 1.5:
		return -3
	case salaryRatio <= 0.5:
		return 5
	case salaryRatio <= 0.7:
		return 2
	default:
		return 0
	}
}

// mbrRef is an independent Go re-derivation of CalculateMoneyballRating
// (src/api/calculations.go), built directly from the formula rather than
// hand-copied magic numbers. Like salaryPenaltyRef, this relies on Go's
// ordinary IEEE-754 float64 division-by-zero semantics (+Inf, no panic) to
// naturally reproduce the overall=0 edge case with no special-casing here
// -- mbrSQL's explicit is_infinite_price_multiplier flag exists only
// because that DuckDB behavior isn't something to rely on unverified, not
// because the Go formula itself needs special-casing.
func mbrRef(overall int, ageStr, personality string, transferValueAmount, wageAmount int64) int {
	baseRating := overall / 3

	ageInt, err := strconv.Atoi(ageStr)
	if err != nil {
		ageInt = 26
	}
	ageModifier := ageModifierRef(ageInt)

	mentalityModifier := mentalityModifierRef(personality)

	overallF := float64(overall)
	transferValueMillions := float64(transferValueAmount) / 1000000.0

	var calculatedValueScore float64
	if transferValueMillions != 0 {
		valuePerRating := transferValueMillions / overallF
		expectedValuePerRating := expectedValuePerRatingRef(overallF)
		priceMultiplier := valuePerRating / expectedValuePerRating

		var valuePenalty float64
		switch {
		case priceMultiplier >= 5.0:
			valuePenalty = 0.1
		case priceMultiplier >= 4.0:
			valuePenalty = 0.2
		case priceMultiplier >= 3.0:
			valuePenalty = 0.3
		case priceMultiplier >= 2.5:
			valuePenalty = 0.4
		case priceMultiplier >= 2.0:
			valuePenalty = 0.5
		case priceMultiplier >= 1.5:
			valuePenalty = 0.7
		case priceMultiplier <= 0.5:
			valuePenalty = 1.5
		case priceMultiplier <= 0.7:
			valuePenalty = 1.2
		default:
			valuePenalty = 1.0
		}

		logValuePerRating := math.Log10(valuePerRating + 1)
		baseEfficiency := overallF / (logValuePerRating + 1)

		switch {
		case overallF >= 80:
			calculatedValueScore = baseEfficiency * 1.2
		case overallF >= 70:
			calculatedValueScore = baseEfficiency * 1.0
		case overallF >= 60:
			calculatedValueScore = baseEfficiency * 0.9
		case overallF >= 55:
			calculatedValueScore = baseEfficiency * 0.8
		default:
			calculatedValueScore = baseEfficiency * 0.6
		}

		calculatedValueScore *= valuePenalty

		if priceMultiplier <= 1.5 {
			if valuePerRating < expectedValuePerRating*0.7 {
				calculatedValueScore *= 1.3
			} else if valuePerRating < expectedValuePerRating*0.85 {
				calculatedValueScore *= 1.15
			}
		}
	}

	valueScoreContribution := int(calculatedValueScore * 0.5)

	var transferValuePenalty int
	if transferValueMillions > 0 {
		valuePerRating := transferValueMillions / overallF
		expectedValuePerRating := expectedValuePerRatingRef(overallF)
		priceMultiplier := valuePerRating / expectedValuePerRating
		switch {
		case priceMultiplier >= 5.0:
			transferValuePenalty = -40
		case priceMultiplier >= 4.0:
			transferValuePenalty = -30
		case priceMultiplier >= 3.0:
			transferValuePenalty = -20
		case priceMultiplier >= 2.5:
			transferValuePenalty = -15
		case priceMultiplier >= 2.0:
			transferValuePenalty = -10
		case priceMultiplier >= 1.5:
			transferValuePenalty = -5
		case priceMultiplier <= 0.5:
			transferValuePenalty = 5
		case priceMultiplier <= 0.7:
			transferValuePenalty = 2
		default:
			transferValuePenalty = 0
		}
	}

	salaryPenalty := salaryPenaltyRef(transferValueAmount, wageAmount)

	return baseRating + ageModifier + mentalityModifier + valueScoreContribution + transferValuePenalty + salaryPenalty
}

// normalizeMBRRef is an independent Go re-derivation of
// NormalizeMBRScoresRelativeToMax (src/api/calculations.go).
func normalizeMBRRef(rawMBRs map[int64]int) map[int64]int {
	result := make(map[int64]int, len(rawMBRs))
	if len(rawMBRs) == 0 {
		return result
	}
	maxMBR := math.MinInt64
	for _, v := range rawMBRs {
		if v > maxMBR {
			maxMBR = v
		}
	}
	if maxMBR <= 0 {
		for k := range rawMBRs {
			result[k] = 0
		}
		return result
	}
	for k, v := range rawMBRs {
		n := int((float64(v) / float64(maxMBR)) * 100.0)
		if n < 0 {
			n = 0
		}
		result[k] = n
	}
	return result
}

// mbrPlayer is the per-scenario input to runMBRScenario: a single-attribute
// role/weight pair (shared "TESTPOS - TestRole - TestDuty" role across all
// players in a scenario, weight {"Attr": 10}) controls Overall via attr,
// keeping every other MBR term isolated to exactly what the test varies.
type mbrPlayer struct {
	playerID            int64
	attr                int
	age                 string
	personality         string
	transferValueAmount int64
	wageAmount          int64
}

// runMBRScenario seeds a single-role role_weights fixture, materializes the
// given players, queries computeMBR, and asserts the result matches
// normalizeMBRRef applied to mbrRef's raw values for the same players --
// end-to-end, matching what the real endpoint actually exposes (only the
// final normalized value, never raw intermediate terms).
func runMBRScenario(t *testing.T, datasetName string, players []mbrPlayer) map[int64]int {
	t.Helper()
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	weights := map[string]int{"Attr": 10}
	seedRoleWeightsForTest(t, ctx, store, map[string]map[string]int{
		"TESTPOS - TestRole - TestDuty": weights,
	})

	rows := make([]PlayerRow, 0, len(players))
	rawMBRs := make(map[int64]int, len(players))
	for _, p := range players {
		attrs := map[string]int{"Attr": p.attr}
		overall := roleOverallRef(attrs, weights)
		rows = append(rows, PlayerRow{
			PlayerID:            p.playerID,
			ShortPositions:      []string{"TESTPOS"},
			NumericAttributes:   attrs,
			Age:                 p.age,
			Personality:         p.personality,
			TransferValueAmount: p.transferValueAmount,
			WageAmount:          p.wageAmount,
		})
		rawMBRs[p.playerID] = mbrRef(overall, p.age, p.personality, p.transferValueAmount, p.wageAmount)
	}

	if err := writeParquet(ctx, dir, datasetName, rows); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	got, err := computeMBR(ctx, store, dir, datasetName)
	if err != nil {
		t.Fatalf("computeMBR failed: %v", err)
	}

	want := normalizeMBRRef(rawMBRs)
	if len(got) != len(want) {
		t.Errorf("result has %d players, want %d", len(got), len(want))
	}
	for playerID, wantMBR := range want {
		if gotMBR, ok := got[playerID]; !ok {
			t.Errorf("player %d missing from result", playerID)
		} else if gotMBR != wantMBR {
			t.Errorf("player %d: MBR = %d, want %d (raw=%d)", playerID, gotMBR, wantMBR, rawMBRs[playerID])
		}
	}
	return got
}

func TestComputeMBRAgeModifierSweep(t *testing.T) {
	players := []mbrPlayer{
		{playerID: 1, attr: 15, age: "17"}, // 16-18 bracket
		{playerID: 2, attr: 15, age: "19"},
		{playerID: 3, attr: 15, age: "25"},
		{playerID: 4, attr: 15, age: "26"},
		{playerID: 5, attr: 15, age: "31"},
		{playerID: 6, attr: 15, age: "35"},  // >31 formula
		{playerID: 7, attr: 15, age: "10"},  // <16 default 0
		{playerID: 8, attr: 15, age: "N/A"}, // malformed -> defaults to 26, same as player 4
	}
	got := runMBRScenario(t, "mbr-test-age-sweep", players)
	if got[4] != got[8] {
		t.Errorf("malformed age (%d) should default to the same modifier as age 26 (%d)", got[8], got[4])
	}
}

func TestComputeMBRMentalityModifierSweep(t *testing.T) {
	players := []mbrPlayer{
		{playerID: 1, attr: 15, age: "26", personality: "Model Citizen"}, // Elite: 20
		{playerID: 2, attr: 15, age: "26", personality: "Ambitious"},     // Very Good: 15
		{playerID: 3, attr: 15, age: "26", personality: "Balanced"},      // Good: 8 (also in the dead "Okay" list)
		{playerID: 4, attr: 15, age: "26", personality: "Realist"},       // Poor: -15
		{playerID: 5, attr: 15, age: "26", personality: "Slack"},         // Very Poor: -35
		{playerID: 6, attr: 15, age: "26", personality: "Unknown Trait"}, // unrecognized: 0
	}
	got := runMBRScenario(t, "mbr-test-mentality-sweep", players)

	// Regression: "Balanced" must score as the reachable Good tier (8), not
	// the dead "Okay" switch's implied 0 -- both share player 3's
	// personality, so if the dead branch were ever mistakenly reintroduced
	// live, player 3 would collapse toward player 6 (unrecognized, 0)
	// instead of standing apart at the Good tier.
	if got[3] == got[6] {
		t.Errorf(`"Balanced" (Good tier, want modifier 8) produced the same MBR as an unrecognized personality (0) -- the dead "Okay" switch may have been reintroduced live`)
	}
}

func TestComputeMBRValueScoreAndTransferPenaltyTiers(t *testing.T) {
	const attr = 15 // -> a fixed Overall shared by every player in this scenario
	overall := roleOverallRef(map[string]int{"Attr": attr}, map[string]int{"Attr": 10})
	expected := expectedValuePerRatingRef(float64(overall))

	amountFor := func(priceMultiplier float64) int64 {
		valuePerRating := priceMultiplier * expected
		millions := valuePerRating * float64(overall)
		return int64(millions * 1_000_000)
	}

	players := []mbrPlayer{
		{playerID: 1, attr: attr, age: "26", transferValueAmount: amountFor(1.0)}, // reasonably priced
		{playerID: 2, attr: attr, age: "26", transferValueAmount: amountFor(2.0)}, // moderately overpriced
		{playerID: 3, attr: attr, age: "26", transferValueAmount: amountFor(6.0)}, // severely overpriced
		{playerID: 4, attr: attr, age: "26", transferValueAmount: amountFor(0.4)}, // undervalued
	}
	runMBRScenario(t, "mbr-test-price-tiers", players)
}

func TestComputeMBROverallZeroDivideByZeroEdgeCase(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	weights := map[string]int{"Attr": 10}
	seedRoleWeightsForTest(t, ctx, store, map[string]map[string]int{
		"TESTPOS - TestRole - TestDuty": weights,
	})

	// Player 1: ShortPositions matches no seeded role -> Overall=0, with a
	// nonzero transfer value -> the is_infinite_price_multiplier edge case.
	// Player 2: a normal control player, included so normalization has a
	// nonzero max to scale against.
	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"ZZZUNKNOWN"}, NumericAttributes: map[string]int{},
			Age: "26", TransferValueAmount: 50_000_000},
		{PlayerID: 2, ShortPositions: []string{"TESTPOS"}, NumericAttributes: map[string]int{"Attr": 15},
			Age: "26"},
	}
	if err := writeParquet(ctx, dir, "mbr-test-overall-zero", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	got, err := computeMBR(ctx, store, dir, "mbr-test-overall-zero")
	if err != nil {
		t.Fatalf("computeMBR failed: %v", err)
	}

	overall2 := roleOverallRef(map[string]int{"Attr": 15}, weights)
	rawMBRs := map[int64]int{
		1: mbrRef(0, "26", "", 50_000_000, 0),
		2: mbrRef(overall2, "26", "", 0, 0),
	}
	want := normalizeMBRRef(rawMBRs)
	for playerID, wantMBR := range want {
		if got[playerID] != wantMBR {
			t.Errorf("player %d: MBR = %d, want %d (raw=%d)", playerID, got[playerID], wantMBR, rawMBRs[playerID])
		}
	}
}

func TestComputeMBRSalaryPenaltyTiers(t *testing.T) {
	const attr = 15
	overall := roleOverallRef(map[string]int{"Attr": attr}, map[string]int{"Attr": 10})
	expected := expectedValuePerRatingRef(float64(overall))
	// priceMultiplier=1.0 -> "reasonably priced" (0 penalty on both the
	// value-score and transfer-value-penalty terms), isolating each
	// player's MBR difference to the salary-penalty term alone.
	reasonableTransferValue := int64(1.0 * expected * float64(overall) * 1_000_000)
	transferValueMillions := float64(reasonableTransferValue) / 1_000_000.0
	expectedWeeklySalary := (transferValueMillions * 0.12 * 1000) / 52

	wageFor := func(salaryRatio float64) int64 {
		wageThousandsPerWeek := salaryRatio * expectedWeeklySalary
		return int64(wageThousandsPerWeek * 1000)
	}

	players := []mbrPlayer{
		{playerID: 1, attr: attr, age: "26", transferValueAmount: reasonableTransferValue, wageAmount: 0}, // WageAmount=0 -> 0
		{playerID: 2, attr: attr, age: "26", transferValueAmount: 0, wageAmount: 100_000},                 // TransferValue=0 & wage>0 -> -15
		{playerID: 3, attr: attr, age: "26", transferValueAmount: reasonableTransferValue, wageAmount: wageFor(2.2)},
		{playerID: 4, attr: attr, age: "26", transferValueAmount: reasonableTransferValue, wageAmount: wageFor(0.4)},
	}
	runMBRScenario(t, "mbr-test-salary-tiers", players)
}

func TestComputeMBRNormalizationTopAndClamp(t *testing.T) {
	players := []mbrPlayer{
		{playerID: 1, attr: 15, age: "17"}, // best age bonus -> should be the dataset max -> normalizes to 100
		{playerID: 2, attr: 15, age: "26"}, // a middling raw MBR
		{playerID: 3, attr: 15, age: "35", personality: "Slack", transferValueAmount: 500_000_000}, // very negative raw MBR -> must clamp to 0, not go negative
	}
	got := runMBRScenario(t, "mbr-test-normalization-clamp", players)

	if got[1] != 100 {
		t.Errorf("the dataset's max raw-MBR player should normalize to exactly 100, got %d", got[1])
	}
	if got[3] != 0 {
		t.Errorf("a very negative raw-MBR player should clamp to 0 (not go negative), got %d", got[3])
	}
}

func TestComputeMBRAllNonPositiveNormalizesToZero(t *testing.T) {
	players := []mbrPlayer{
		{playerID: 1, attr: 1, age: "35", personality: "Slack"},
		{playerID: 2, attr: 1, age: "40", personality: "Slack"},
	}
	got := runMBRScenario(t, "mbr-test-all-nonpositive", players)

	for playerID, mbr := range got {
		if mbr != 0 {
			t.Errorf("player %d: MBR = %d, want 0 (every player has raw MBR<=0)", playerID, mbr)
		}
	}
}

func TestComputeMBREndToEndScenario(t *testing.T) {
	const attr = 16
	overall := roleOverallRef(map[string]int{"Attr": attr}, map[string]int{"Attr": 10})
	expected := expectedValuePerRatingRef(float64(overall))
	overpriced := int64(2.5 * expected * float64(overall) * 1_000_000)

	players := []mbrPlayer{
		{
			playerID:            1,
			attr:                attr,
			age:                 "23",
			personality:         "Ambitious",
			transferValueAmount: overpriced,
			wageAmount:          200_000,
		},
		{playerID: 2, attr: attr, age: "26"}, // neutral baseline for normalization
	}
	runMBRScenario(t, "mbr-test-end-to-end", players)
}
