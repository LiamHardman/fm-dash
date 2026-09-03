package main

// createPlayersTableDDL defines the Query artifact schema: raw per-player
// facts only. Every weight-derived field (Overall, TotalStats, MBR, CA,
// BestRoleOverall, RoleSpecificOveralls, the FIFA-style PAC/SHO/PAS/DRI/DEF/
// PHY/GK/DIV/HAN/REF/KIC/SPD/POS, and PerformancePercentiles) is deliberately
// excluded — those are computed at query time in a later phase, not stored,
// per the Dataset Query Service design map (.scratch/duckdb-query-service/
// issues/02-query-artifact-schema.md).
const createPlayersTableDDL = `
CREATE TABLE players (
	player_id                  BIGINT,
	name                       VARCHAR,
	position                   VARCHAR,
	age                        VARCHAR,
	club                       VARCHAR,
	division                   VARCHAR,
	based_in                   VARCHAR,
	transfer_value             VARCHAR,
	transfer_value_amount      BIGINT,
	wage                       VARCHAR,
	wage_amount                BIGINT,
	personality                VARCHAR,
	media_handling              VARCHAR,
	nationality                VARCHAR,
	nationality_iso             VARCHAR,
	nationality_fifa_code        VARCHAR,
	attribute_masked            BOOLEAN,
	numeric_attributes          MAP(VARCHAR, INTEGER),
	performance_stats_numeric    MAP(VARCHAR, DOUBLE),
	parsed_positions             VARCHAR[],
	short_positions              VARCHAR[],
	position_groups              VARCHAR[]
);`
