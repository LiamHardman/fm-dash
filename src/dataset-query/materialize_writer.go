package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	duckdb "github.com/duckdb/duckdb-go/v2"
)

// PlayerRow is the raw, per-player shape materialized into the Query
// artifact. It intentionally excludes every weight-derived field (see
// schema.go) — those are computed at query time in a later phase.
type PlayerRow struct {
	PlayerID                int64              `json:"playerId"`
	Name                    string             `json:"name"`
	Position                string             `json:"position"`
	Age                     string             `json:"age"`
	Club                    string             `json:"club"`
	Division                string             `json:"division"`
	BasedIn                 string             `json:"basedIn"`
	TransferValue           string             `json:"transferValue"`
	TransferValueAmount     int64              `json:"transferValueAmount"`
	Wage                    string             `json:"wage"`
	WageAmount              int64              `json:"wageAmount"`
	Personality             string             `json:"personality"`
	MediaHandling           string             `json:"mediaHandling"`
	Nationality             string             `json:"nationality"`
	NationalityISO          string             `json:"nationalityIso"`
	NationalityFIFACode     string             `json:"nationalityFifaCode"`
	AttributeMasked         bool               `json:"attributeMasked"`
	NumericAttributes       map[string]int     `json:"numericAttributes"`
	PerformanceStatsNumeric map[string]float64 `json:"performanceStatsNumeric"`
	ParsedPositions         []string           `json:"parsedPositions"`
	ShortPositions          []string           `json:"shortPositions"`
	PositionGroups          []string           `json:"positionGroups"`
}

// writeParquet materializes players as the Query artifact for datasetID,
// writing it atomically (via a temp file + rename) to
// <storageDir>/<datasetID>.parquet.
func writeParquet(ctx context.Context, storageDir, datasetID string, players []PlayerRow) (err error) {
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		return fmt.Errorf("creating duckdb connector: %w", err)
	}
	defer func() { _ = connector.Close() }()

	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

	// Pin a single physical connection for the whole operation, so the
	// CREATE TABLE, Appender writes, and COPY all observe the same
	// in-memory database state.
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquiring duckdb connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err = conn.ExecContext(ctx, createPlayersTableDDL); err != nil {
		return fmt.Errorf("creating players table: %w", err)
	}

	if err = appendPlayers(conn, players); err != nil {
		return err
	}

	finalPath := filepath.Join(storageDir, datasetID+".parquet")
	tmpPath := filepath.Join(storageDir, "."+datasetID+".parquet.tmp")

	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	// The path is inlined as an escaped SQL string literal rather than a
	// bound parameter — COPY's parameter-binding behavior isn't reliably
	// documented for this driver version. datasetID is already validated
	// by the caller (materializeHandler) to exclude path-traversal
	// characters, and both path components below are otherwise built
	// entirely from server-controlled values, so this is not
	// attacker-controlled input; the quote-doubling escape is defense in
	// depth, not the primary safeguard.
	copySQL := fmt.Sprintf(`COPY players TO '%s' (FORMAT PARQUET)`, escapeSQLStringLiteral(tmpPath))
	if _, err = conn.ExecContext(ctx, copySQL); err != nil {
		return fmt.Errorf("copying to parquet: %w", err)
	}

	if err = os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("renaming temp parquet to final path: %w", err)
	}
	return nil
}

func appendPlayers(conn *sql.Conn, players []PlayerRow) error {
	return conn.Raw(func(driverConn any) error {
		dc, ok := driverConn.(driver.Conn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type %T", driverConn)
		}

		appender, err := duckdb.NewAppenderFromConn(dc, "", "players")
		if err != nil {
			return fmt.Errorf("creating appender: %w", err)
		}
		defer func() { _ = appender.Close() }()

		for _, p := range players {
			numericAttrs := duckdb.OrderedMap{}
			for k, v := range p.NumericAttributes {
				numericAttrs.Set(k, int32(v))
			}

			perfStats := duckdb.OrderedMap{}
			for k, v := range p.PerformanceStatsNumeric {
				perfStats.Set(k, v)
			}

			if err := appender.AppendRow(
				p.PlayerID, p.Name, p.Position, p.Age, p.Club, p.Division, p.BasedIn,
				p.TransferValue, p.TransferValueAmount, p.Wage, p.WageAmount,
				p.Personality, p.MediaHandling,
				p.Nationality, p.NationalityISO, p.NationalityFIFACode,
				p.AttributeMasked,
				numericAttrs, perfStats,
				p.ParsedPositions, p.ShortPositions, p.PositionGroups,
			); err != nil {
				return fmt.Errorf("appending row for player %d: %w", p.PlayerID, err)
			}
		}
		return nil
	})
}

func escapeSQLStringLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
