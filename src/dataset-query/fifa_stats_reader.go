package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FifaStats holds one player's FIFA-style category scores, computed from
// their numeric_attributes and the currently active category_weights.
type FifaStats struct {
	PAC, SHO, PAS, DRI, DEF, PHY     int
	GK, DIV, HAN, REF, KIC, SPD, POS int
}

// fifaStatsSQL replicates calculateWeightedAverage, computeNonLinearScaling,
// and CalculateFifaStatGo from src/api/calculations.go exactly:
//   - weighted average only counts attributes present with value in [1, 20]
//     (calculateWeightedAverage)
//   - linearScore := weightedAverage * 5.3
//   - the >=75 / <75 piecewise power-curve compression, including the
//     "scaledRating < 10 && linearRating > 20" floor override
//     (computeNonLinearScaling)
//   - PAS is the max of three independently scaled methods keyed by
//     PAS_standard / PAS_no_set_pieces / PAS_no_off_ball (CalculateFifaStatGo's
//     special-cased branch)
//   - goalkeepers (position_groups contains 'Goalkeepers') get
//     GK/DIV/HAN/REF/KIC/SPD/POS and 0 for the outfield categories, and vice
//     versa (EnhancePlayerWithCalculations, player_processing.go)
//
// FLOOR(x + 0.5) is used instead of DuckDB's ROUND() for the same reason
// documented in percentiles_reader.go: ROUND's tie-breaking rule isn't
// pinned down by DuckDB's docs, whereas every value rounded here is
// non-negative, so FLOOR(x+0.5) is provably identical to Go's math.Round
// for this domain.
const fifaStatsSQL = `
WITH players AS (
    SELECT
        player_id,
        list_contains(position_groups, 'Goalkeepers') AS is_goalkeeper,
        numeric_attributes
    FROM read_parquet(?)
),
attr_matches AS (
    SELECT
        p.player_id,
        w.category,
        map_extract_value(p.numeric_attributes, w.attribute) AS attr_value,
        w.weight
    FROM players p
    CROSS JOIN category_weights w
),
weighted_averages AS (
    SELECT
        player_id,
        category,
        SUM(attr_value * weight) FILTER (WHERE attr_value BETWEEN 1 AND 20) AS weighted_sum,
        SUM(weight)               FILTER (WHERE attr_value BETWEEN 1 AND 20) AS weight_total
    FROM attr_matches
    GROUP BY player_id, category
),
linear_scores AS (
    SELECT
        player_id,
        category,
        COALESCE(CAST(weighted_sum AS DOUBLE) / NULLIF(weight_total, 0), 0.0) * 5.3 AS linear_score
    FROM weighted_averages
),
category_scores AS (
    SELECT
        player_id,
        category,
        CASE
            WHEN linear_score <= 0 THEN 0
            WHEN linear_score >= 99 THEN 99
            WHEN linear_score >= 75 THEN
                CAST(FLOOR(75.0 + (linear_score - 75.0) * 0.95 + 0.5) AS INTEGER)
            ELSE
                CASE
                    WHEN (POWER(linear_score / 75.0, 1.8) * 75.0) < 10.0 AND linear_score > 20.0
                        THEN CAST(FLOOR(10.0 + (linear_score - 20.0) * 0.15 + 0.5) AS INTEGER)
                    ELSE CAST(FLOOR(POWER(linear_score / 75.0, 1.8) * 75.0 + 0.5) AS INTEGER)
                END
        END AS final_score
    FROM linear_scores
),
pivoted AS (
    SELECT
        player_id,
        MAX(final_score) FILTER (WHERE category = 'PAC')               AS pac_score,
        MAX(final_score) FILTER (WHERE category = 'SHO')               AS sho_score,
        MAX(final_score) FILTER (WHERE category = 'PAS_standard')      AS pas_standard_score,
        MAX(final_score) FILTER (WHERE category = 'PAS_no_set_pieces') AS pas_no_set_pieces_score,
        MAX(final_score) FILTER (WHERE category = 'PAS_no_off_ball')   AS pas_no_off_ball_score,
        MAX(final_score) FILTER (WHERE category = 'DRI')               AS dri_score,
        MAX(final_score) FILTER (WHERE category = 'DEF')               AS def_score,
        MAX(final_score) FILTER (WHERE category = 'PHY')               AS phy_score,
        MAX(final_score) FILTER (WHERE category = 'GK')                AS gk_score,
        MAX(final_score) FILTER (WHERE category = 'DIV')               AS div_score,
        MAX(final_score) FILTER (WHERE category = 'HAN')               AS han_score,
        MAX(final_score) FILTER (WHERE category = 'REF')               AS ref_score,
        MAX(final_score) FILTER (WHERE category = 'KIC')               AS kic_score,
        MAX(final_score) FILTER (WHERE category = 'SPD')               AS spd_score,
        MAX(final_score) FILTER (WHERE category = 'POS')               AS pos_score
    FROM category_scores
    GROUP BY player_id
)
SELECT
    p.player_id,
    CASE WHEN p.is_goalkeeper THEN 0 ELSE COALESCE(v.pac_score, 0) END AS "pac",
    CASE WHEN p.is_goalkeeper THEN 0 ELSE COALESCE(v.sho_score, 0) END AS "sho",
    CASE WHEN p.is_goalkeeper THEN 0 ELSE GREATEST(
        COALESCE(v.pas_standard_score, 0),
        COALESCE(v.pas_no_set_pieces_score, 0),
        COALESCE(v.pas_no_off_ball_score, 0)
    ) END AS "pas",
    CASE WHEN p.is_goalkeeper THEN 0 ELSE COALESCE(v.dri_score, 0) END AS "dri",
    CASE WHEN p.is_goalkeeper THEN 0 ELSE COALESCE(v.def_score, 0) END AS "def",
    CASE WHEN p.is_goalkeeper THEN 0 ELSE COALESCE(v.phy_score, 0) END AS "phy",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.gk_score, 0)  ELSE 0 END AS "gk",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.div_score, 0) ELSE 0 END AS "div",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.han_score, 0) ELSE 0 END AS "han",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.ref_score, 0) ELSE 0 END AS "ref",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.kic_score, 0) ELSE 0 END AS "kic",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.spd_score, 0) ELSE 0 END AS "spd",
    CASE WHEN p.is_goalkeeper THEN COALESCE(v.pos_score, 0) ELSE 0 END AS "pos"
FROM players p
LEFT JOIN pivoted v ON v.player_id = p.player_id
ORDER BY p.player_id;
`

// computeFifaStats reads datasetID's Query artifact and computes FIFA-style
// category scores for every player, joined against store's currently active
// category_weights. Runs against store's persistent DuckDB connection
// directly (not an ephemeral in-memory one) -- read_parquet() works on any
// connection, so no ATTACH is needed to combine the per-dataset Parquet
// artifact with the persistent weights table.
func computeFifaStats(ctx context.Context, store *WeightsStore, storageDir, datasetID string) (map[int64]FifaStats, error) {
	artifactPath := filepath.Join(storageDir, datasetID+".parquet")
	if _, err := os.Stat(artifactPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("checking artifact: %w", err)
	}

	rows, err := store.db.QueryContext(ctx, fifaStatsSQL, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("querying fifa stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int64]FifaStats)
	for rows.Next() {
		var playerID int64
		var s FifaStats
		if err := rows.Scan(
			&playerID, &s.PAC, &s.SHO, &s.PAS, &s.DRI, &s.DEF, &s.PHY,
			&s.GK, &s.DIV, &s.HAN, &s.REF, &s.KIC, &s.SPD, &s.POS,
		); err != nil {
			return nil, fmt.Errorf("scanning fifa stats row: %w", err)
		}
		result[playerID] = s
	}
	return result, rows.Err()
}
