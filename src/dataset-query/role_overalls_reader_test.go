package main

import (
	"context"
	"testing"
)

// roleOverallRef is an independent Go re-derivation of
// CalculateOverallForRoleGo (src/api/calculations.go): unlike
// weightedAverageRef/fifaStatRef (the FIFA-stat path), attributes with
// value <=0 are skipped entirely, but present positive values are CLAMPED
// into [1,20] rather than excluded if outside that range.
func roleOverallRef(attrs, weights map[string]int) int {
	var weightedSum, totalWeight float64
	for attr, weight := range weights {
		v, ok := attrs[attr]
		if !ok || v <= 0 {
			continue
		}
		clamped := float64(v)
		if clamped < 1 {
			clamped = 1
		} else if clamped > 20 {
			clamped = 20
		}
		weightedSum += clamped * float64(weight)
		totalWeight += float64(weight)
	}
	if totalWeight == 0 {
		return 0
	}
	linear := (weightedSum / totalWeight) * 5.9
	final := computeNonLinearScalingRef(linear)
	if final > 99 {
		return 99
	}
	if final < 0 {
		return 0
	}
	return final
}

// meanCategoryOverallRef is an independent Go re-derivation of
// CalculateMeanCategoryOverall (src/api/calculations.go).
func meanCategoryOverallRef(roleScores map[string]int) int {
	bestByCategory := make(map[string]int)
	for roleName, score := range roleScores {
		category := roleStyleCategory(roleName)
		if existing, ok := bestByCategory[category]; !ok || score > existing {
			bestByCategory[category] = score
		}
	}
	if len(bestByCategory) == 0 {
		return 0
	}
	total := 0
	for _, s := range bestByCategory {
		total += s
	}
	return total / len(bestByCategory)
}

// seedRoleWeightsForTest inserts a small custom role-weights set directly,
// bypassing seedRoleWeightsIfEmpty's "only if empty" / embedded-JSON path --
// used to test computeRoleOveralls against controlled, hand-designed roles.
func seedRoleWeightsForTest(t *testing.T, ctx context.Context, store *WeightsStore, roleWeights map[string]map[string]int) {
	t.Helper()
	for roleName, weights := range roleWeights {
		position := roleStylePosition(roleName)
		styleCategory := roleStyleCategory(roleName)
		for attr, weight := range weights {
			if _, err := store.db.ExecContext(ctx,
				`INSERT INTO role_weights (role_name, position, style_category, attribute, weight) VALUES (?, ?, ?, ?, ?)`,
				roleName, position, styleCategory, attr, weight,
			); err != nil {
				t.Fatalf("seeding role weight %s.%s: %v", roleName, attr, err)
			}
		}
	}
}

func TestComputeRoleOverallsMultiCategoryAggregation(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	roleWeights := map[string]map[string]int{
		"DC - Ball Playing Defender - Defend": {"Tck": 10},
		"DC - Ball Playing Defender - Cover":  {"Tck": 10, "Pas": 10},
		"DC - Central Defender - Defend":      {"Hea": 10},
	}
	seedRoleWeightsForTest(t, ctx, store, roleWeights)

	attrs := map[string]int{"Tck": 15, "Pas": 10, "Hea": 5}
	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"DC"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "role-test-multi", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeRoleOveralls(ctx, store, dir, "role-test-multi")
	if err != nil {
		t.Fatalf("computeRoleOveralls failed: %v", err)
	}

	wantScores := make(map[string]int, len(roleWeights))
	for roleName, weights := range roleWeights {
		wantScores[roleName] = roleOverallRef(attrs, weights)
	}
	wantOverall := meanCategoryOverallRef(wantScores)

	got := result[1]
	if got.Overall != wantOverall {
		t.Errorf("Overall = %d, want %d", got.Overall, wantOverall)
	}

	gotScores := make(map[string]int, len(got.RoleSpecificOveralls))
	for _, rs := range got.RoleSpecificOveralls {
		gotScores[rs.RoleName] = rs.Score
	}
	if len(gotScores) != len(wantScores) {
		t.Fatalf("RoleSpecificOveralls has %d entries, want %d (got=%v want=%v)", len(gotScores), len(wantScores), gotScores, wantScores)
	}
	for roleName, want := range wantScores {
		if got := gotScores[roleName]; got != want {
			t.Errorf("role %q score = %d, want %d", roleName, got, want)
		}
	}
}

func TestComputeRoleOverallsDeterministicTieBreak(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Two roles in DIFFERENT style categories with identical weights ->
	// identical scores for a player with the matching attribute -> a
	// genuine cross-category tie that CalculateMeanCategoryOverall cannot
	// collapse away (each survives into the mean as its own category).
	roleA := "AMC - Advanced Playmaker - Support"
	roleB := "AMC - Attacking Midfielder - Support"
	roleWeights := map[string]map[string]int{
		roleA: {"Pas": 10},
		roleB: {"Pas": 10},
	}
	seedRoleWeightsForTest(t, ctx, store, roleWeights)

	attrs := map[string]int{"Pas": 15}
	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"AMC"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "role-test-tie", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeRoleOveralls(ctx, store, dir, "role-test-tie")
	if err != nil {
		t.Fatalf("computeRoleOveralls failed: %v", err)
	}

	got := result[1]
	scoreA := roleOverallRef(attrs, roleWeights[roleA])
	scoreB := roleOverallRef(attrs, roleWeights[roleB])
	if scoreA != scoreB {
		t.Fatalf("test setup invalid: expected roleA and roleB to tie, got %d and %d", scoreA, scoreB)
	}
	if scoreA == 0 {
		t.Fatalf("test setup invalid: expected a nonzero tied score, got 0")
	}

	// "AMC - Advanced Playmaker - Support" < "AMC - Attacking Midfielder - Support"
	// alphabetically ('d' < 't' at the first differing character).
	wantBest := roleA
	if got.BestRoleOverall != wantBest {
		t.Errorf("BestRoleOverall = %q, want %q (deterministic alphabetical tie-break)", got.BestRoleOverall, wantBest)
	}

	gotScores := make(map[string]int, len(got.RoleSpecificOveralls))
	for _, rs := range got.RoleSpecificOveralls {
		gotScores[rs.RoleName] = rs.Score
	}
	if gotScores[roleA] != scoreA || gotScores[roleB] != scoreB {
		t.Errorf("both tied roles should appear with equal scores, got %v", gotScores)
	}
}

func TestComputeRoleOverallsClampNotExclude(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	roleName := "ST - Poacher - Attack"
	weights := map[string]int{"Str": 10}
	seedRoleWeightsForTest(t, ctx, store, map[string]map[string]int{roleName: weights})

	// Str=25 is outside the normal [1,20] attribute domain -- the role
	// path must CLAMP it to 20, not exclude it (which is what the
	// FIFA-stat path's BETWEEN 1 AND 20 filter would incorrectly do if
	// mistakenly reused here).
	attrs := map[string]int{"Str": 25}
	players := []PlayerRow{
		{PlayerID: 1, ShortPositions: []string{"ST"}, NumericAttributes: attrs},
	}
	if err := writeParquet(ctx, dir, "role-test-clamp", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	result, err := computeRoleOveralls(ctx, store, dir, "role-test-clamp")
	if err != nil {
		t.Fatalf("computeRoleOveralls failed: %v", err)
	}

	want := roleOverallRef(attrs, weights)
	if want == 0 {
		t.Fatalf("test setup invalid: expected a nonzero reference score, got 0")
	}

	got := result[1]
	if len(got.RoleSpecificOveralls) != 1 || got.RoleSpecificOveralls[0].Score != want {
		t.Errorf("role score = %v, want a single entry with score %d (clamped, not excluded)", got.RoleSpecificOveralls, want)
	}
	if len(got.RoleSpecificOveralls) == 1 && got.RoleSpecificOveralls[0].Score == 0 {
		t.Errorf("score was 0 -- attribute was likely excluded (BETWEEN 1 AND 20) instead of clamped")
	}
}
