package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	duckdb "github.com/duckdb/duckdb-go/v2"
)

// ErrArtifactNotFound indicates the requested dataset has no materialized
// Query artifact (Parquet file) on disk.
var ErrArtifactNotFound = errors.New("query artifact not found")

// StatPercentiles is player_id -> group_name -> stat_key -> percentile
// (0-100, or -1 when there is no eligible data for that player/stat).
type StatPercentiles map[int64]map[string]map[string]float64

// percentilesSQL replicates calculatePercentileValue and
// CalculatePlayerPerformancePercentiles from src/api/performance_stats.go
// exactly: rank-based percentile with tie-midpoint handling, computed
// across three populations (Global, broad position groups, detailed
// position groups) for every known performance stat key. See
// .claude/plans (Phase C1) for the full derivation.
const percentilesSQL = `
WITH players AS (
    SELECT * FROM read_parquet(?)
),
stat_keys(stat_key) AS (
    VALUES ('Asts/90'), ('Av Rat'), ('Blk/90'), ('Ch C/90'), ('Clr/90'), ('Cr C/90'), ('Drb/90'),
        ('xA/90'), ('xG/90'), ('Gls/90'), ('Hdrs W/90'), ('Int/90'), ('K Ps/90'), ('Ps C/90'),
        ('Shot/90'), ('Tck/90'), ('Poss Won/90'), ('ShT/90'), ('Pres C/90'), ('Poss Lost/90'),
        ('Pr passes/90'), ('Conv %'), ('Tck R'), ('Pas %'), ('Cr C/A'),
        ('Fls'), ('Apps'), ('NP-xG/90'), ('Ps A/90'), ('Mins'), ('Clean Sheets'), ('FA'), ('CRS A/90'),
        ('Con/90'), ('Cln/90'), ('xGP/90'), ('Sv %'),
        ('Pres A/90'), ('Dist/90'), ('Saves/90'), ('Mins/Gl'), ('K Tck/90')
),
broad_groups(group_name) AS (
    VALUES ('Goalkeepers'), ('Defenders'), ('Midfielders'), ('Attackers')
),
detailed_group_members(group_name, short_pos) AS (
    VALUES ('Full-backs', 'DR'), ('Full-backs', 'DL'), ('Centre-backs', 'DC'),
        ('Wing-backs', 'WBR'), ('Wing-backs', 'WBL'), ('Defensive Midfielders', 'DM'),
        ('Central Midfielders', 'MC'), ('Wide Midfielders', 'MR'), ('Wide Midfielders', 'ML'),
        ('Attacking Midfielders (Central)', 'AMC'), ('Wingers', 'AMR'), ('Wingers', 'AML'),
        ('Strikers', 'ST')
),
memberships AS (
    SELECT DISTINCT player_id, 'Global' AS group_name FROM players
    UNION ALL
    SELECT DISTINCT p.player_id, g.group_name
    FROM players p, UNNEST(p.position_groups) AS u(group_name)
    JOIN broad_groups g ON g.group_name = u.group_name
    UNION ALL
    SELECT DISTINCT p.player_id, d.group_name
    FROM players p, UNNEST(p.short_positions) AS u(short_pos)
    JOIN detailed_group_members d ON d.short_pos = u.short_pos
),
player_stats AS (
    SELECT m.player_id, m.group_name, k.stat_key,
           map_extract_value(p.performance_stats_numeric, k.stat_key) AS stat_value
    FROM memberships m
    JOIN players p ON p.player_id = m.player_id
    CROSS JOIN stat_keys k
),
eligible_stats AS (
    SELECT player_id, group_name, stat_key, stat_value
    FROM player_stats
    WHERE stat_value IS NOT NULL AND NOT isnan(stat_value)
),
ranked AS (
    SELECT player_id, group_name, stat_key,
        FLOOR(
            ((RANK() OVER (PARTITION BY group_name, stat_key ORDER BY stat_value) - 1)
             + 0.5 * COUNT(*) OVER (PARTITION BY group_name, stat_key, stat_value))
            / COUNT(*) OVER (PARTITION BY group_name, stat_key) * 100.0 + 0.5
        ) AS percentile
    FROM eligible_stats
)
SELECT ps.player_id, ps.group_name, ps.stat_key, COALESCE(r.percentile, -1.0) AS percentile
FROM player_stats ps
LEFT JOIN ranked r
    ON r.player_id = ps.player_id AND r.group_name = ps.group_name AND r.stat_key = ps.stat_key
ORDER BY ps.player_id, ps.group_name, ps.stat_key;
`

// computePercentiles reads datasetID's Query artifact and computes the
// 3-tier (Global / broad position group / detailed position group)
// performance-stat percentiles for every player.
func computePercentiles(ctx context.Context, storageDir, datasetID string) (StatPercentiles, error) {
	artifactPath := filepath.Join(storageDir, datasetID+".parquet")
	if _, err := os.Stat(artifactPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("checking artifact: %w", err)
	}

	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		return nil, fmt.Errorf("creating duckdb connector: %w", err)
	}
	defer func() { _ = connector.Close() }()

	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, percentilesSQL, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("querying percentiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(StatPercentiles)
	for rows.Next() {
		var playerID int64
		var groupName, statKey string
		var percentile float64
		if err := rows.Scan(&playerID, &groupName, &statKey, &percentile); err != nil {
			return nil, fmt.Errorf("scanning percentile row: %w", err)
		}
		groups, ok := result[playerID]
		if !ok {
			groups = make(map[string]map[string]float64)
			result[playerID] = groups
		}
		stats, ok := groups[groupName]
		if !ok {
			stats = make(map[string]float64)
			groups[groupName] = stats
		}
		stats[statKey] = percentile
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating percentile rows: %w", err)
	}
	return result, nil
}
