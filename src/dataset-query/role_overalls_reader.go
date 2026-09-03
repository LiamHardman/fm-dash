package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// RoleOverallScore mirrors src/api/types.go's RoleOverallScore.
type RoleOverallScore struct {
	RoleName string
	Score    int
}

// RoleOveralls holds one player's full role-based Overall computation:
// every applicable role's score, the mean-of-categories Overall, and the
// deterministically tie-broken BestRoleOverall.
type RoleOveralls struct {
	Overall              int
	BestRoleOverall      string
	RoleSpecificOveralls []RoleOverallScore
}

// roleOverallsSQL replicates, exactly:
//   - CalculateOverallForRoleGo (src/api/calculations.go): attributes <=0 or
//     absent are skipped entirely; present positive values are CLAMPED into
//     [1,20] (not excluded if outside that range -- contrast with
//     calculateWeightedAverage/fifaStatsSQL's exclude-if-outside-[1,20] rule).
//     linearScore := weightedAvg * 5.9 (overallScalingFactor, NOT the FIFA
//     path's 5.3), then the identical applyNonLinearScaling piecewise curve
//     already implemented in fifaStatsSQL (verbatim, unchanged) is reused.
//   - roleStyleCategory + CalculateMeanCategoryOverall (src/api/calculations.go):
//     group a player's applicable role scores by style_category (precomputed
//     in Go at seed time, see role_weights_seed.go), keep the MAX score per
//     category, then take the TRUNCATING integer average across all distinct
//     categories present -- DuckDB's `/` on two integers is floating-point
//     division, so `//` (DuckDB's integer-division operator) is used
//     deliberately here, matching Go's `total/len` truncation.
//   - BestRoleOverall selection (player_processing.go): the original
//     tie-break iterates in nondeterministic Go map order, so which named
//     role "wins" an exact-score tie is not reproducible run-to-run even in
//     the original system. This query picks, deterministically, the highest
//     score then alphabetically-first role name (ROW_NUMBER, matching the
//     RANK()/ROW_NUMBER() idiom already established in percentiles_reader.go).
//     If the winning score is 0, BestRoleOverall is forced to the empty
//     string -- replicating the Go loop's bestRoleName default, which a strict `>` comparison
//     starting from 0 never overwrites when every candidate score is also 0.
const roleOverallsSQL = "WITH " + overallCTEsSQL + `,
ranked_roles AS (
    SELECT
        player_id, role_name, score,
        ROW_NUMBER() OVER (PARTITION BY player_id ORDER BY score DESC, role_name ASC) AS rn
    FROM role_scores
),
best_role AS (
    SELECT player_id, CASE WHEN score > 0 THEN role_name ELSE '' END AS best_role_overall
    FROM ranked_roles
    WHERE rn = 1
)
SELECT
    p.player_id,
    COALESCE(o.overall, 0) AS overall,
    COALESCE(b.best_role_overall, '') AS best_role_overall,
    rs.role_name,
    rs.score
FROM players p
LEFT JOIN overalls o ON o.player_id = p.player_id
LEFT JOIN best_role b ON b.player_id = p.player_id
LEFT JOIN role_scores rs ON rs.player_id = p.player_id
ORDER BY p.player_id, rs.score DESC NULLS LAST, rs.role_name ASC;
`

// computeRoleOveralls reads datasetID's Query artifact and computes every
// applicable role's score, the mean-of-categories Overall, and the
// deterministic BestRoleOverall for every player, joined against store's
// currently active role_weights. Runs against store's persistent DuckDB
// connection directly (role_weights lives in weights.duckdb), same pattern
// as computeFifaStats.
func computeRoleOveralls(ctx context.Context, store *WeightsStore, storageDir, datasetID string) (map[int64]RoleOveralls, error) {
	artifactPath := filepath.Join(storageDir, datasetID+".parquet")
	if _, err := os.Stat(artifactPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("checking artifact: %w", err)
	}

	rows, err := store.db.QueryContext(ctx, roleOverallsSQL, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("querying role overalls: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int64]RoleOveralls)
	for rows.Next() {
		var playerID int64
		var overall int
		var bestRoleOverall string
		var roleName sql.NullString
		var score sql.NullInt64
		if err := rows.Scan(&playerID, &overall, &bestRoleOverall, &roleName, &score); err != nil {
			return nil, fmt.Errorf("scanning role overalls row: %w", err)
		}
		entry, ok := result[playerID]
		if !ok {
			entry = RoleOveralls{
				Overall:              overall,
				BestRoleOverall:      bestRoleOverall,
				RoleSpecificOveralls: []RoleOverallScore{},
			}
		}
		if roleName.Valid {
			entry.RoleSpecificOveralls = append(entry.RoleSpecificOveralls,
				RoleOverallScore{RoleName: roleName.String, Score: int(score.Int64)})
		}
		result[playerID] = entry
	}
	return result, rows.Err()
}
