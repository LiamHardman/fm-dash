package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// mbrSQL replicates CalculateMoneyballRating/getAgeModifier/
// getMentalityModifier/getExpectedValuePerRating/getSalaryPenalty
// (src/api/calculations.go + src/api/handlers.go:3178) and
// NormalizeMBRScoresRelativeToMax (src/api/calculations.go) exactly,
// bug-for-bug -- including the known age-modifier/value-score skew a
// previous session flagged but never fixed; a redesign is a deliberate
// future effort, not part of this migration.
//
// Reuses overallCTEsSQL (see overall_cte.go) for the scalar Overall each
// player needs -- this is the same mean-of-categories value role-overalls
// computes, so it's shared verbatim rather than hand-copied a second time.
//
// player_overall LEFT JOINs (not INNER JOINs) against the shared `overalls`
// CTE and COALESCEs to 0: a player with ShortPositions matching no seeded
// role produces no `overalls` row at all, and an INNER JOIN would silently
// drop such players from the result instead of computing their MBR with
// Overall=0 (matching Go, where Overall simply stays its zero value).
//
// Six raw-MBR terms, summed in raw_mbr:
//  1. baseRating = overall // 3 (Go: int division, truncating; overall>=0
//     so floor and truncation coincide, matching the `//` idiom already
//     established in overallCTEsSQL's own `overalls` CTE).
//  2. age_modifier: flat per-age-integer table. TRY_CAST/COALESCE-to-26
//     mirrors strconv.Atoi's parse-failure default.
//  3. mentality_modifier: a flat CASE over five *reachable* Go switch
//     tiers. Go's original has a SIXTH "Okay personalities" switch between
//     Good and Poor sharing the exact same 9-name list as Good -- since Go
//     evaluates switches top-to-bottom and Good already returns first,
//     that sixth switch is dead code. This CASE deliberately does NOT
//     reintroduce it as a live 0-returning branch: those 9 names map to 8,
//     full stop.
//  4. value_score_calc -> TRUNC(calculated_value_score*0.5): a multi-stage
//     tiered calculation, gated on transfer_value_millions=0 (->0) and
//     is_infinite_price_multiplier (overall=0 with transfer_value_millions
//     != 0 -- Go's IEEE-754 float division produces +Inf here, landing in
//     every tiered switch's >=5.0-equivalent branch; DuckDB's own
//     divide-by-zero behavior through an equivalent arithmetic chain is
//     NOT relied upon -- this is short-circuited explicitly to the known
//     Go-equivalent output, calculated_value_score=0).
//  5. transfer_value_penalty_calc: an independent 8-tier table on the same
//     price_multiplier computed once in price_multiplier_calc (Go
//     recomputes priceMultiplier from scratch in a separate code block;
//     this SQL computes it once and reuses it for both this and #4).
//     is_infinite_price_multiplier short-circuits straight to -40 here too.
//  6. salary_penalty_calc: transfer_value_amount=0 (with wage_amount!=0
//     already guaranteed by that point) is the same
//     Go-divides-by-zero-produces-+Inf case, landing in the >=3.0 tier ->
//     -15, short-circuited explicitly rather than relied upon.
//
// TRUNC() is used (not a bare CAST(DOUBLE AS INTEGER)) for both raw_mbr's
// value-score term and the final normalization step: DuckDB's own CAST
// from DOUBLE to INTEGER rounds to nearest, it does not truncate toward
// zero like Go's int(floatVal) -- confirmed empirically by
// TestDuckDBTruncBehavior before this query was written to depend on it.
//
// normalized/the final SELECT replicate NormalizeMBRScoresRelativeToMax:
// a MAX(mbr) OVER () window (the one genuinely dataset-relative step in
// this whole query), max_mbr<=0 forcing every player to 0, otherwise
// TRUNC((mbr/max_mbr)*100.0) clamped up to 0 if negative.
const mbrSQL = "WITH " + overallCTEsSQL + `,
player_overall AS (
    SELECT p.player_id, COALESCE(o.overall, 0) AS overall
    FROM players p
    LEFT JOIN overalls o ON o.player_id = p.player_id
),
age_ints AS (
    SELECT player_id, COALESCE(TRY_CAST(age AS INTEGER), 26) AS age_int
    FROM players
),
age_modifiers AS (
    SELECT player_id,
        CASE
            WHEN age_int BETWEEN 16 AND 18 THEN 30
            WHEN age_int = 19 THEN 25
            WHEN age_int = 20 THEN 22
            WHEN age_int = 21 THEN 18
            WHEN age_int = 22 THEN 15
            WHEN age_int = 23 THEN 10
            WHEN age_int = 24 THEN 6
            WHEN age_int = 25 THEN 3
            WHEN age_int = 26 THEN -1
            WHEN age_int = 27 THEN -3
            WHEN age_int = 28 THEN -5
            WHEN age_int = 29 THEN -10
            WHEN age_int = 30 THEN -15
            WHEN age_int = 31 THEN -20
            WHEN age_int > 31 THEN -20 - (age_int - 31) * 2
            ELSE 0
        END AS age_modifier
    FROM age_ints
),
mentality_modifiers AS (
    SELECT player_id,
        CASE
            WHEN personality IN ('Model Citizen', 'Model Professional') THEN 20
            WHEN personality IN (
                'Perfectionist', 'Resolute', 'Professional', 'Fairly Professional',
                'Iron Willed', 'Resillient', 'Spirited', 'Driven', 'Determined',
                'Fairly Determined', 'Charismatic Leader', 'Born Leader', 'Leader',
                'Very Ambitious', 'Ambitious', 'Fairly Ambitious'
            ) THEN 15
            WHEN personality IN (
                'Balanced', 'Light-Hearted', 'Jovial', 'Very Loyal', 'Loyal',
                'Fairly Loyal', 'Honest', 'Sporting', 'Fairly Sporting'
            ) THEN 8
            WHEN personality IN ('Fickle', 'Mercenary', 'Unambitious', 'Unsporting', 'Realist') THEN -15
            WHEN personality IN (
                'Slack', 'Casual', 'Temperamental', 'Spineless', 'Low Self-Belief',
                'Easily Discouraged', 'Low Determination'
            ) THEN -35
            ELSE 0
        END AS mentality_modifier
    FROM players
),
value_calc AS (
    SELECT po.player_id, po.overall,
        CAST(p.transfer_value_amount AS DOUBLE) / 1000000.0 AS transfer_value_millions
    FROM player_overall po
    JOIN players p ON p.player_id = po.player_id
),
expected_value AS (
    SELECT player_id, overall, transfer_value_millions,
        CASE
            WHEN overall >= 85 THEN 1.2
            WHEN overall >= 80 THEN 0.8
            WHEN overall >= 75 THEN 0.5
            WHEN overall >= 70 THEN 0.3
            WHEN overall >= 65 THEN 0.2
            WHEN overall >= 60 THEN 0.15
            ELSE 0.1
        END AS expected_value_per_rating
    FROM value_calc
),
price_multiplier_calc AS (
    SELECT player_id, overall, transfer_value_millions, expected_value_per_rating,
        (overall = 0 AND transfer_value_millions != 0) AS is_infinite_price_multiplier,
        CASE WHEN overall > 0 THEN transfer_value_millions / overall ELSE NULL END AS value_per_rating,
        CASE WHEN overall > 0 THEN (transfer_value_millions / overall) / expected_value_per_rating ELSE NULL END AS price_multiplier
    FROM expected_value
),
value_score_calc AS (
    SELECT player_id,
        CASE
            WHEN transfer_value_millions = 0 THEN 0.0
            WHEN is_infinite_price_multiplier THEN 0.0
            ELSE
                (overall / (LOG10(value_per_rating + 1) + 1))
                * CASE
                    WHEN overall >= 80 THEN 1.2
                    WHEN overall >= 70 THEN 1.0
                    WHEN overall >= 60 THEN 0.9
                    WHEN overall >= 55 THEN 0.8
                    ELSE 0.6
                  END
                * CASE
                    WHEN price_multiplier >= 5.0 THEN 0.1
                    WHEN price_multiplier >= 4.0 THEN 0.2
                    WHEN price_multiplier >= 3.0 THEN 0.3
                    WHEN price_multiplier >= 2.5 THEN 0.4
                    WHEN price_multiplier >= 2.0 THEN 0.5
                    WHEN price_multiplier >= 1.5 THEN 0.7
                    WHEN price_multiplier <= 0.5 THEN 1.5
                    WHEN price_multiplier <= 0.7 THEN 1.2
                    ELSE 1.0
                  END
                * CASE
                    WHEN price_multiplier > 1.5 THEN 1.0
                    WHEN value_per_rating < expected_value_per_rating * 0.7 THEN 1.3
                    WHEN value_per_rating < expected_value_per_rating * 0.85 THEN 1.15
                    ELSE 1.0
                  END
        END AS calculated_value_score
    FROM price_multiplier_calc
),
transfer_value_penalty_calc AS (
    SELECT player_id,
        CASE
            WHEN transfer_value_millions = 0 THEN 0
            WHEN is_infinite_price_multiplier THEN -40
            WHEN price_multiplier >= 5.0 THEN -40
            WHEN price_multiplier >= 4.0 THEN -30
            WHEN price_multiplier >= 3.0 THEN -20
            WHEN price_multiplier >= 2.5 THEN -15
            WHEN price_multiplier >= 2.0 THEN -10
            WHEN price_multiplier >= 1.5 THEN -5
            WHEN price_multiplier <= 0.5 THEN 5
            WHEN price_multiplier <= 0.7 THEN 2
            ELSE 0
        END AS transfer_value_penalty
    FROM price_multiplier_calc
),
salary_ratio_calc AS (
    SELECT player_id, wage_amount, transfer_value_amount,
        CASE
            WHEN wage_amount = 0 OR transfer_value_amount = 0 THEN NULL
            ELSE (CAST(wage_amount AS DOUBLE) / 1000.0)
                / ((CAST(transfer_value_amount AS DOUBLE) / 1000000.0 * 0.12 * 1000) / 52)
        END AS salary_ratio
    FROM players
),
salary_penalty_calc AS (
    SELECT player_id,
        CASE
            WHEN wage_amount = 0 THEN 0
            WHEN transfer_value_amount = 0 THEN -15
            WHEN salary_ratio >= 3.0 THEN -15
            WHEN salary_ratio >= 2.5 THEN -10
            WHEN salary_ratio >= 2.0 THEN -7
            WHEN salary_ratio >= 1.5 THEN -3
            WHEN salary_ratio <= 0.5 THEN 5
            WHEN salary_ratio <= 0.7 THEN 2
            ELSE 0
        END AS salary_penalty
    FROM salary_ratio_calc
),
raw_mbr AS (
    SELECT
        po.player_id,
        (po.overall // 3)
        + am.age_modifier
        + mm.mentality_modifier
        + CAST(TRUNC(vs.calculated_value_score * 0.5) AS INTEGER)
        + tvp.transfer_value_penalty
        + sp.salary_penalty
        AS mbr
    FROM player_overall po
    JOIN age_modifiers am ON am.player_id = po.player_id
    JOIN mentality_modifiers mm ON mm.player_id = po.player_id
    JOIN value_score_calc vs ON vs.player_id = po.player_id
    JOIN transfer_value_penalty_calc tvp ON tvp.player_id = po.player_id
    JOIN salary_penalty_calc sp ON sp.player_id = po.player_id
),
normalized AS (
    SELECT player_id, mbr, MAX(mbr) OVER () AS max_mbr
    FROM raw_mbr
)
SELECT
    player_id,
    CASE
        WHEN max_mbr <= 0 THEN 0
        ELSE GREATEST(0, CAST(TRUNC((CAST(mbr AS DOUBLE) / CAST(max_mbr AS DOUBLE)) * 100.0) AS INTEGER))
    END AS mbr
FROM normalized
ORDER BY player_id;
`

// computeMBR reads datasetID's Query artifact and computes the final,
// dataset-normalized Moneyball Rating for every player, joined against
// store's currently active role_weights (needed for the shared Overall
// computation). Runs against store's persistent DuckDB connection directly,
// same pattern as computeRoleOveralls/computeFifaStats.
func computeMBR(ctx context.Context, store *WeightsStore, storageDir, datasetID string) (map[int64]int, error) {
	artifactPath := filepath.Join(storageDir, datasetID+".parquet")
	if _, err := os.Stat(artifactPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("checking artifact: %w", err)
	}

	rows, err := store.db.QueryContext(ctx, mbrSQL, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("querying mbr: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int64]int)
	for rows.Next() {
		var playerID int64
		var mbr int
		if err := rows.Scan(&playerID, &mbr); err != nil {
			return nil, fmt.Errorf("scanning mbr row: %w", err)
		}
		result[playerID] = mbr
	}
	return result, rows.Err()
}
