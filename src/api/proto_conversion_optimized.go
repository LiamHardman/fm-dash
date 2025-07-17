package main

import (
	"context"
	"fmt"
	"time"

	"api/proto"

	"go.opentelemetry.io/otel/attribute"
)

// Optimized protobuf conversion with minimal allocations and efficient data structures

// ToProtoOptimized converts a RoleOverallScore struct to protobuf format with minimal allocations
func (r *RoleOverallScore) ToProtoOptimized(ctx context.Context) (*proto.RoleOverallScore, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.role_to_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "role_overall_score"),
		attribute.String("conversion.direction", "to_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if r == nil {
		RecordError(ctx, ErrNilRoleOverallScore, "Cannot convert nil RoleOverallScore to protobuf",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilRoleOverallScore
	}

	protoRole := &proto.RoleOverallScore{
		RoleName: r.RoleName,
		Score:    safeIntToInt32(r.Score),
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
	)

	return protoRole, nil
}

// RoleOverallScoreFromProtoOptimized converts a protobuf RoleOverallScore to the native struct
func RoleOverallScoreFromProtoOptimized(ctx context.Context, protoRole *proto.RoleOverallScore) (*RoleOverallScore, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.role_from_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "role_overall_score"),
		attribute.String("conversion.direction", "from_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if protoRole == nil {
		RecordError(ctx, ErrNilProtobufRoleOverallScore, "Cannot convert nil protobuf RoleOverallScore",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilProtobufRoleOverallScore
	}

	role := &RoleOverallScore{
		RoleName: protoRole.GetRoleName(),
		Score:    int(protoRole.GetScore()),
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
	)

	return role, nil
}

// ToProtoOptimized converts a Player struct to protobuf format with minimal allocations
func (p *Player) ToProtoOptimized(ctx context.Context) (*proto.Player, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.player_to_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "player"),
		attribute.String("conversion.direction", "to_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if p == nil {
		RecordError(ctx, ErrNilPlayer, "Cannot convert nil Player to protobuf",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilPlayer
	}

	// Pre-allocate slices and maps with correct capacity to avoid reallocations
	roleCount := len(p.RoleSpecificOveralls)
	protoRoles := make([]*proto.RoleOverallScore, 0, roleCount)

	// Convert RoleSpecificOveralls with pre-allocated slice
	for _, role := range p.RoleSpecificOveralls {
		protoRole, err := role.ToProtoOptimized(ctx)
		if err != nil {
			logError(ctx, "Failed to convert role to protobuf",
				"error", err,
				"player_uid", p.UID,
				"role_name", role.RoleName)
			return nil, fmt.Errorf("failed to convert role %s to protobuf: %w", role.RoleName, err)
		}
		protoRoles = append(protoRoles, protoRole)
	}

	// Pre-allocate performance percentiles map
	percentileCount := len(p.PerformancePercentiles)
	protoPercentiles := make(map[string]*proto.PerformancePercentileMap, percentileCount)

	// Convert performance percentiles with pre-allocated map
	for key, innerMap := range p.PerformancePercentiles {
		protoPercentiles[key] = &proto.PerformancePercentileMap{
			Percentiles: innerMap,
		}
	}

	// Pre-allocate numeric attributes map
	numericAttrCount := len(p.NumericAttributes)
	protoNumericAttrs := make(map[string]int32, numericAttrCount)

	// Convert numeric attributes with pre-allocated map
	for key, value := range p.NumericAttributes {
		protoNumericAttrs[key] = safeIntToInt32(value)
	}

	protoPlayer := &proto.Player{
		Uid:                     p.UID,
		Name:                    p.Name,
		Position:                p.Position,
		Age:                     p.Age,
		Club:                    p.Club,
		Division:                p.Division,
		TransferValue:           p.TransferValue,
		Wage:                    p.Wage,
		Personality:             p.Personality,
		MediaHandling:           p.MediaHandling,
		Nationality:             p.Nationality,
		NationalityIso:          p.NationalityISO,
		NationalityFifaCode:     p.NationalityFIFACode,
		AttributeMasked:         p.AttributeMasked,
		Attributes:              p.Attributes,
		NumericAttributes:       protoNumericAttrs,
		PerformanceStatsNumeric: p.PerformanceStatsNumeric,
		PerformancePercentiles:  protoPercentiles,
		ParsedPositions:         p.ParsedPositions,
		ShortPositions:          p.ShortPositions,
		PositionGroups:          p.PositionGroups,
		Pac:                     safeIntToInt32(p.PAC),
		Sho:                     safeIntToInt32(p.SHO),
		Pas:                     safeIntToInt32(p.PAS),
		Dri:                     safeIntToInt32(p.DRI),
		Def:                     safeIntToInt32(p.DEF),
		Phy:                     safeIntToInt32(p.PHY),
		Gk:                      safeIntToInt32(p.GK),
		Div:                     safeIntToInt32(p.DIV),
		Han:                     safeIntToInt32(p.HAN),
		Ref:                     safeIntToInt32(p.REF),
		Kic:                     safeIntToInt32(p.KIC),
		Spd:                     safeIntToInt32(p.SPD),
		Pos:                     safeIntToInt32(p.POS),
		Overall:                 safeIntToInt32(p.Overall),
		BestRoleOverall:         p.BestRoleOverall,
		RoleSpecificOveralls:    protoRoles,
		TransferValueAmount:     p.TransferValueAmount,
		WageAmount:              p.WageAmount,
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
		attribute.Int("conversion.attributes_count", len(p.Attributes)),
		attribute.Int("conversion.numeric_attributes_count", len(p.NumericAttributes)),
		attribute.Int("conversion.performance_stats_count", len(p.PerformanceStatsNumeric)),
	)

	return protoPlayer, nil
}

// PlayerFromProtoOptimized converts a protobuf Player to the native struct with minimal allocations
func PlayerFromProtoOptimized(ctx context.Context, protoPlayer *proto.Player) (*Player, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.player_from_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "player"),
		attribute.String("conversion.direction", "from_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if protoPlayer == nil {
		RecordError(ctx, ErrNilProtobufPlayer, "Cannot convert nil protobuf Player",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilProtobufPlayer
	}

	// Pre-allocate slices with correct capacity
	roleCount := len(protoPlayer.GetRoleSpecificOveralls())
	roles := make([]RoleOverallScore, 0, roleCount)

	// Convert RoleSpecificOveralls with pre-allocated slice
	for _, protoRole := range protoPlayer.GetRoleSpecificOveralls() {
		role, err := RoleOverallScoreFromProtoOptimized(ctx, protoRole)
		if err != nil {
			logError(ctx, "Failed to convert protobuf role",
				"error", err,
				"player_uid", protoPlayer.GetUid(),
				"role_name", protoRole.GetRoleName())
			return nil, fmt.Errorf("failed to convert protobuf role %s: %w", protoRole.GetRoleName(), err)
		}
		roles = append(roles, *role)
	}

	// Pre-allocate performance percentiles map
	percentileCount := len(protoPlayer.GetPerformancePercentiles())
	percentiles := make(map[string]map[string]float64, percentileCount)

	// Convert performance percentiles with pre-allocated map
	for key, protoMap := range protoPlayer.GetPerformancePercentiles() {
		percentiles[key] = protoMap.GetPercentiles()
	}

	// Pre-allocate numeric attributes map
	numericAttrCount := len(protoPlayer.GetNumericAttributes())
	numericAttrs := make(map[string]int, numericAttrCount)

	// Convert numeric attributes with pre-allocated map
	for key, value := range protoPlayer.GetNumericAttributes() {
		numericAttrs[key] = int(value)
	}

	player := &Player{
		UID:                     protoPlayer.GetUid(),
		Name:                    protoPlayer.GetName(),
		Position:                protoPlayer.GetPosition(),
		Age:                     protoPlayer.GetAge(),
		Club:                    protoPlayer.GetClub(),
		Division:                protoPlayer.GetDivision(),
		TransferValue:           protoPlayer.GetTransferValue(),
		Wage:                    protoPlayer.GetWage(),
		Personality:             protoPlayer.GetPersonality(),
		MediaHandling:           protoPlayer.GetMediaHandling(),
		Nationality:             protoPlayer.GetNationality(),
		NationalityISO:          protoPlayer.GetNationalityIso(),
		NationalityFIFACode:     protoPlayer.GetNationalityFifaCode(),
		AttributeMasked:         protoPlayer.GetAttributeMasked(),
		Attributes:              protoPlayer.GetAttributes(),
		NumericAttributes:       numericAttrs,
		PerformanceStatsNumeric: protoPlayer.GetPerformanceStatsNumeric(),
		PerformancePercentiles:  percentiles,
		ParsedPositions:         protoPlayer.GetParsedPositions(),
		ShortPositions:          protoPlayer.GetShortPositions(),
		PositionGroups:          protoPlayer.GetPositionGroups(),
		PAC:                     int(protoPlayer.GetPac()),
		SHO:                     int(protoPlayer.GetSho()),
		PAS:                     int(protoPlayer.GetPas()),
		DRI:                     int(protoPlayer.GetDri()),
		DEF:                     int(protoPlayer.GetDef()),
		PHY:                     int(protoPlayer.GetPhy()),
		GK:                      int(protoPlayer.GetGk()),
		DIV:                     int(protoPlayer.GetDiv()),
		HAN:                     int(protoPlayer.GetHan()),
		REF:                     int(protoPlayer.GetRef()),
		KIC:                     int(protoPlayer.GetKic()),
		SPD:                     int(protoPlayer.GetSpd()),
		POS:                     int(protoPlayer.GetPos()),
		Overall:                 int(protoPlayer.GetOverall()),
		BestRoleOverall:         protoPlayer.GetBestRoleOverall(),
		RoleSpecificOveralls:    roles,
		TransferValueAmount:     protoPlayer.GetTransferValueAmount(),
		WageAmount:              protoPlayer.GetWageAmount(),
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
		attribute.Int("conversion.attributes_count", len(player.Attributes)),
		attribute.Int("conversion.numeric_attributes_count", len(player.NumericAttributes)),
		attribute.Int("conversion.performance_stats_count", len(player.PerformanceStatsNumeric)),
	)

	return player, nil
}

// ToProtoOptimized converts a PlayerDataWithCurrency struct to protobuf format with minimal allocations
func (d *PlayerDataWithCurrency) ToProtoOptimized(ctx context.Context) (*proto.DatasetData, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.dataset_to_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "dataset_data"),
		attribute.String("conversion.direction", "to_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if d == nil {
		RecordError(ctx, ErrNilDatasetData, "Cannot convert nil DatasetData to protobuf",
			WithErrorCategory("validation"),
			WithSeverity("high"))
		return nil, ErrNilDatasetData
	}

	// Pre-allocate players slice with correct capacity
	playerCount := len(d.Players)
	protoPlayers := make([]*proto.Player, 0, playerCount)

	// Convert players with pre-allocated slice
	for i, player := range d.Players {
		protoPlayer, err := player.ToProtoOptimized(ctx)
		if err != nil {
			logError(ctx, "Failed to convert player to protobuf",
				"error", err,
				"player_index", i,
				"player_uid", player.UID)
			return nil, fmt.Errorf("failed to convert player %d to protobuf: %w", i, err)
		}
		protoPlayers = append(protoPlayers, protoPlayer)
	}

	protoDataset := &proto.DatasetData{
		Players:        protoPlayers,
		CurrencySymbol: d.CurrencySymbol,
		CacheData:      "", // PlayerDataWithCurrency doesn't have CacheData field
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
		attribute.Int("dataset.player_count", len(d.Players)),
		attribute.String("dataset.currency_symbol", d.CurrencySymbol),
	)

	return protoDataset, nil
}

// DatasetDataFromProtoOptimized converts a protobuf DatasetData to PlayerDataWithCurrency with minimal allocations
func DatasetDataFromProtoOptimized(ctx context.Context, protoDataset *proto.DatasetData) (*PlayerDataWithCurrency, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.dataset_from_proto_optimized", []attribute.KeyValue{
		attribute.String("conversion.type", "dataset_data"),
		attribute.String("conversion.direction", "from_protobuf"),
		attribute.String("conversion.optimization", "minimal_allocations"),
	})
	defer span.End()

	start := time.Now()

	if protoDataset == nil {
		RecordError(ctx, ErrNilProtobufDatasetData, "Cannot convert nil protobuf DatasetData",
			WithErrorCategory("validation"),
			WithSeverity("high"))
		return nil, ErrNilProtobufDatasetData
	}

	// Pre-allocate players slice with correct capacity
	playerCount := len(protoDataset.GetPlayers())
	players := make([]Player, 0, playerCount)

	// Convert players with pre-allocated slice
	for i, protoPlayer := range protoDataset.GetPlayers() {
		player, err := PlayerFromProtoOptimized(ctx, protoPlayer)
		if err != nil {
			logError(ctx, "Failed to convert protobuf player",
				"error", err,
				"player_index", i,
				"player_uid", protoPlayer.GetUid())
			return nil, fmt.Errorf("failed to convert protobuf player %d: %w", i, err)
		}
		players = append(players, *player)
	}

	dataset := &PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: protoDataset.GetCurrencySymbol(),
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Bool("conversion.optimized", true),
		attribute.Int("dataset.player_count", len(players)),
		attribute.String("dataset.currency_symbol", dataset.CurrencySymbol),
	)

	return dataset, nil
}
