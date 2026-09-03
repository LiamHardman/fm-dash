package main

// overallCTEsSQL is the WITH-chain (without the leading "WITH" keyword and
// without a trailing comma) that computes each player's mean-of-categories
// Overall -- CalculateOverallForRoleGo + roleStyleCategory +
// CalculateMeanCategoryOverall (src/api/calculations.go) -- from
// role_weights joined against numeric_attributes/short_positions. Shared
// verbatim between role_overalls_reader.go (which layers BestRoleOverall/
// RoleSpecificOveralls on top of this) and mbr_reader.go (which only needs
// the scalar `overall` produced by the final `overalls` CTE), so the two
// endpoints can never silently diverge on how Overall is computed.
//
// The `players` CTE selects more columns than role-overalls itself needs
// (age, personality, transfer_value_amount, wage_amount) so that mbr_reader.go
// can consume them from the same CTE without a second read_parquet(?) call.
const overallCTEsSQL = `
players AS (
    SELECT player_id, short_positions, numeric_attributes,
           age, personality, transfer_value_amount, wage_amount
    FROM read_parquet(?)
),
role_positions AS (
    SELECT DISTINCT role_name, position, style_category FROM role_weights
),
applicable_roles AS (
    SELECT DISTINCT p.player_id, rp.role_name, rp.style_category
    FROM players p, UNNEST(p.short_positions) AS u(short_pos)
    JOIN role_positions rp ON rp.position = u.short_pos
),
attr_matches AS (
    SELECT
        ar.player_id,
        ar.role_name,
        ar.style_category,
        map_extract_value(p.numeric_attributes, rw.attribute) AS attr_value,
        rw.weight
    FROM applicable_roles ar
    JOIN players p ON p.player_id = ar.player_id
    JOIN role_weights rw ON rw.role_name = ar.role_name
),
weighted_averages AS (
    SELECT
        player_id, role_name, style_category,
        SUM(LEAST(GREATEST(attr_value, 1), 20) * weight)
            FILTER (WHERE attr_value IS NOT NULL AND attr_value > 0) AS weighted_sum,
        SUM(weight)
            FILTER (WHERE attr_value IS NOT NULL AND attr_value > 0) AS weight_total
    FROM attr_matches
    GROUP BY player_id, role_name, style_category
),
linear_scores AS (
    SELECT
        player_id, role_name, style_category,
        COALESCE(CAST(weighted_sum AS DOUBLE) / NULLIF(weight_total, 0), 0.0) * 5.9 AS linear_score
    FROM weighted_averages
),
role_scores AS (
    SELECT
        player_id, role_name, style_category,
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
        END AS score
    FROM linear_scores
),
category_max_scores AS (
    SELECT player_id, style_category, MAX(score) AS max_score
    FROM role_scores
    GROUP BY player_id, style_category
),
overalls AS (
    SELECT
        player_id,
        CAST(SUM(max_score) AS BIGINT) // CAST(COUNT(*) AS BIGINT) AS overall
    FROM category_max_scores
    GROUP BY player_id
)`
