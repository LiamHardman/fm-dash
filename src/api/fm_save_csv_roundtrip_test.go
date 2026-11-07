package main

import (
	"bytes"
	"context"
	"os"
	"testing"
)

// The browser FM27 save parser (src/utils/fmSave/exportCsv.js) hands players to this
// backend as an FM-style CSV. testdata/fmsave/save_export.csv is that exporter's output
// (pinned by src/utils/fmSave/exportCsv.test.js); this test proves the existing CSV
// pipeline reads it unchanged.
func TestFMSaveExportCSVRoundTrip(t *testing.T) {
	content, err := os.ReadFile("testdata/fmsave/save_export.csv")
	if err != nil {
		t.Fatal(err)
	}

	response := runUploadPreflight(t, "My Save.csv", string(content))
	if !response.Valid {
		t.Fatalf("preflight rejected the save export, missing %v", response.MissingColumns)
	}

	var headers []string
	var players []Player
	err = scanExport(context.Background(), bytes.NewReader(content), "csv", func(_ int, cells []string, header bool) error {
		if header {
			headers = normalizeCSVHeaders(cells)
			return nil
		}
		player, err := parseCellsToPlayer(cells, headers)
		if err != nil {
			return err
		}
		EnhancePlayerWithCalculations(&player)
		players = append(players, player)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 2 {
		t.Fatalf("got %d players, want 2", len(players))
	}

	outfield, keeper := players[0], players[1]
	checks := []struct {
		field, got, want string
	}{
		{"name", outfield.Name, "Bruno Costa"},
		{"club", outfield.Club, "Arsenal"},
		{"division", outfield.Division, "Premier League"},
		{"based in", outfield.BasedIn, "England"},
		{"nationality", outfield.Nationality, "England"},
		{"FIFA code", outfield.NationalityFIFACode, "ENG"},
		{"position", outfield.Position, "D (RC), DM"},
		{"age", outfield.Age, "26"},
		{"keeper nationality", keeper.Nationality, "Zanzibar"},
		{"keeper position", keeper.Position, "GK"},
		{"personality", outfield.Personality, "Leader"},
		{"media handling", outfield.MediaHandling, "Volatile, Media-friendly"},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s: got %q, want %q", check.field, check.got, check.want)
		}
	}
	if outfield.UID != 2002330039 {
		t.Errorf("UID: got %d", outfield.UID)
	}
	if outfield.TransferValueAmount != 153_400_000 {
		t.Errorf("transfer value: got %d", outfield.TransferValueAmount)
	}
	if outfield.WageAmount != 110_000 {
		t.Errorf("wage: got %d", outfield.WageAmount)
	}
	// Corners is the first attribute column (value 6); Natural Fitness must stay an
	// attribute even though "Nation" also normalises to "Nat".
	if outfield.NumericAttributes["Cor"] != 6 {
		t.Errorf("Cor: got %d", outfield.NumericAttributes["Cor"])
	}
	if outfield.NumericAttributes["Nat"] < 1 || outfield.NumericAttributes["Nat"] > 20 {
		t.Errorf("Natural Fitness: got %d", outfield.NumericAttributes["Nat"])
	}
	if outfield.AttributeMasked {
		t.Error("save exports carry exact attributes and must not be marked masked")
	}
	if len(outfield.ParsedPositions) == 0 || outfield.Overall == 0 {
		t.Errorf("ratings were not calculated: positions %v, overall %d", outfield.ParsedPositions, outfield.Overall)
	}

	// Save-only columns.
	if outfield.SaveCA != 152 || outfield.PA != 171 || len(outfield.HiddenAttributes) != 13 {
		t.Errorf("save-only fields: CA %d PA %d hidden %v", outfield.SaveCA, outfield.PA, outfield.HiddenAttributes)
	}

	// Season stats: the save export's 42 stat headers land in PerformanceStatsNumeric.
	statChecks := []struct {
		key  string
		want float64
	}{{"Mins", 1800}, {"Ps A/90", 45}, {"Pas %", 90}, {"Mins/Gl", 360}, {"Dist/90", 10}}
	for _, check := range statChecks {
		if got, ok := outfield.PerformanceStatsNumeric[check.key]; !ok || got != check.want {
			t.Errorf("stat %s: got %v (present %v), want %v", check.key, got, ok, check.want)
		}
	}
	// The keeper's advanced block is missing: blank cells must be absent, not zero.
	for _, key := range []string{"Dist/90", "Sv %", "Saves/90", "Poss Lost/90"} {
		if value, ok := keeper.PerformanceStatsNumeric[key]; ok {
			t.Errorf("blank stat %s should be missing, got %v", key, value)
		}
	}
	if keeper.PerformanceStatsNumeric["Mins"] != 900 {
		t.Errorf("keeper minutes: got %v", keeper.PerformanceStatsNumeric["Mins"])
	}
	if _, ok := keeper.PerformanceStatsNumeric["Average Rating Club"]; ok {
		t.Error("a '-' rating should be missing")
	}
}

// Manual: FM27_EXPORT_CSV=<csv written by the browser parser from a real save> go test -run RealSave
func TestFMSaveExportCSVRealSave(t *testing.T) {
	path := os.Getenv("FM27_EXPORT_CSV")
	if path == "" {
		t.Skip("set FM27_EXPORT_CSV to a CSV exported from a real FM27 save")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var headers []string
	rows, failures, unrated := 0, 0, 0
	err = scanExport(context.Background(), bytes.NewReader(content), "csv", func(_ int, cells []string, header bool) error {
		if header {
			headers = normalizeCSVHeaders(cells)
			return nil
		}
		rows++
		player, err := parseCellsToPlayer(cells, headers)
		if err != nil {
			failures++
			return nil
		}
		EnhancePlayerWithCalculations(&player)
		if len(player.ParsedPositions) == 0 || player.Overall == 0 {
			unrated++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rows %d, parse failures %d, players without positions or overall %d", rows, failures, unrated)
	if failures > 0 || unrated > rows/100 {
		t.Fatalf("too many bad rows: %d failures, %d unrated of %d", failures, unrated, rows)
	}
}
