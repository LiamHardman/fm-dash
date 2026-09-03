package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
)

func TestWriteParquetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	players := []PlayerRow{
		{
			PlayerID:                1,
			Name:                    "Test Player",
			Position:                "ST",
			Age:                     "24",
			Club:                    "Test FC",
			Division:                "Test Division",
			TransferValue:           "£1M",
			TransferValueAmount:     1_000_000,
			Wage:                    "£10K p/w",
			WageAmount:              10_000,
			Nationality:             "England",
			NationalityISO:          "ENG",
			NumericAttributes:       map[string]int{"Finishing": 18, "Pace": 15},
			PerformanceStatsNumeric: map[string]float64{"Goals": 22.0, "xG": 19.5},
			ParsedPositions:         []string{"ST", "AM"},
			ShortPositions:          []string{"ST"},
			PositionGroups:          []string{"Attack"},
		},
	}

	if err := writeParquet(ctx, dir, "test-dataset", players); err != nil {
		t.Fatalf("writeParquet failed: %v", err)
	}

	parquetPath := filepath.Join(dir, "test-dataset.parquet")

	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("opening duckdb: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM read_parquet(?)`, parquetPath).Scan(&count); err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if count != len(players) {
		t.Fatalf("expected %d rows, got %d", len(players), count)
	}

	var name string
	var finishing int
	if err := db.QueryRowContext(ctx,
		`SELECT name, numeric_attributes['Finishing'] FROM read_parquet(?) WHERE player_id = 1`,
		parquetPath,
	).Scan(&name, &finishing); err != nil {
		t.Fatalf("querying row: %v", err)
	}
	if name != "Test Player" || finishing != 18 {
		t.Fatalf("unexpected row values: name=%q finishing=%d", name, finishing)
	}
}
