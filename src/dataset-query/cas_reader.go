package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// casRoundHalfAwayFromZeroSQL rounds a DOUBLE expression to the nearest
// integer using round-half-away-from-zero (matching Go's math.Round for all
// reals, not just x>=0 like the FLOOR(x+0.5) trick used elsewhere in this
// project). Extracted as its own fragment, embedded verbatim into
// casOverallSQL below, so the isolated rounding unit test exercises the
// exact same SQL text the real query runs rather than a hand-copied
// duplicate that could drift.
const casRoundHalfAwayFromZeroSQL = `
    CASE WHEN %s >= 0 THEN CAST(FLOOR(%s + 0.5) AS INTEGER)
         ELSE CAST(-FLOOR(-(%s) + 0.5) AS INTEGER)
    END`

// casSQL replicates CalculateCAS/calculateCASForPosition
// (src/api/ca_calculation.go:151-227) exactly:
//   - every attribute with a nonzero position weight always contributes
//     (contrast fifaStatsSQL's exclude-if-outside-[1,20] and
//     roleOverallsSQL's skip-if-missing rules): missing/absent or <=0
//     values default to 1, present values are clamped down to 20 if >20
//     (never clamped up -- a contributing value is already >=1 by
//     construction of the CASE below).
//   - ca := 19.8*weightedAvg - 117.8, rounded half-away-from-zero (this
//     formula's weightedAvg can be as low as 1, making ca as low as -98 --
//     unlike every other formula in this project, so the sign-aware
//     rounding fragment above is required here, not just FLOOR(x+0.5)),
//     then clamped to [1,200] (NOT [0,99] like the FIFA/role scores).
//   - CalculateCAS takes the MAX across the player's ShortPositions,
//     defaulting to 0 if no ShortPosition has a matching cas_weights entry
//     (mirrors the Go function's empty-ShortPositions early return AND its
//     per-position map lookup miss, both of which produce no contribution).
//
// No NULLIF/zero-division guard is needed on weighted_averages, unlike
// fifaStatsSQL/roleOverallsSQL: cas_weights never stores zero-weight rows
// (see cas_weights_seed.go), so every joined attr_matches row has weight>0
// by construction, making SUM(weight) provably >0 for any position that
// produced a row at all.
var casSQL = fmt.Sprintf(`
WITH players AS (
    SELECT player_id, short_positions, numeric_attributes FROM read_parquet(?)
),
cas_positions AS (
    SELECT DISTINCT position FROM cas_weights
),
applicable_positions AS (
    SELECT DISTINCT p.player_id, cp.position
    FROM players p, UNNEST(p.short_positions) AS u(short_pos)
    JOIN cas_positions cp ON cp.position = u.short_pos
),
attr_matches AS (
    SELECT ap.player_id, ap.position,
           map_extract_value(p.numeric_attributes, cw.attribute) AS attr_value, cw.weight
    FROM applicable_positions ap
    JOIN players p ON p.player_id = ap.player_id
    JOIN cas_weights cw ON cw.position = ap.position
),
effective_values AS (
    SELECT player_id, position,
        CASE WHEN attr_value IS NOT NULL AND attr_value > 0 THEN LEAST(attr_value, 20) ELSE 1 END AS effective_value,
        weight
    FROM attr_matches
),
weighted_averages AS (
    SELECT player_id, position, SUM(effective_value * weight) / SUM(weight) AS weighted_avg
    FROM effective_values
    GROUP BY player_id, position
),
linear_scores AS (
    SELECT player_id, position, 19.8 * weighted_avg - 117.8 AS ca_raw
    FROM weighted_averages
),
position_rounded AS (
    SELECT player_id, position, %s AS ca_rounded
    FROM linear_scores
),
position_clamped AS (
    SELECT player_id, GREATEST(1, LEAST(200, ca_rounded)) AS ca
    FROM position_rounded
),
player_max AS (
    SELECT player_id, MAX(ca) AS ca FROM position_clamped GROUP BY player_id
)
SELECT p.player_id, COALESCE(m.ca, 0) AS ca
FROM players p
LEFT JOIN player_max m ON m.player_id = p.player_id
ORDER BY p.player_id;
`, fmt.Sprintf(casRoundHalfAwayFromZeroSQL, "ca_raw", "ca_raw", "ca_raw"))

// computeCAS reads datasetID's Query artifact and computes CurrentAbility
// (CAS) for every player, joined against store's currently active
// cas_weights. Runs against store's persistent DuckDB connection directly
// (cas_weights lives in weights.duckdb), same pattern as computeFifaStats
// and computeRoleOveralls.
func computeCAS(ctx context.Context, store *WeightsStore, storageDir, datasetID string) (map[int64]int, error) {
	artifactPath := filepath.Join(storageDir, datasetID+".parquet")
	if _, err := os.Stat(artifactPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("checking artifact: %w", err)
	}

	rows, err := store.db.QueryContext(ctx, casSQL, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("querying cas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int64]int)
	for rows.Next() {
		var playerID int64
		var ca int
		if err := rows.Scan(&playerID, &ca); err != nil {
			return nil, fmt.Errorf("scanning cas row: %w", err)
		}
		result[playerID] = ca
	}
	return result, rows.Err()
}
