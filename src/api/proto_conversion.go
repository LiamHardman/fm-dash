package main

import (
	"context"
	"fmt"
	"time"

	"api/proto"
)

// --- RoleOverallScore Conversion Functions ---

// ToProto converts a RoleOverallScore struct to protobuf format
func (r *RoleOverallScore) ToProto(ctx context.Context) (*proto.RoleOverallScore, error) {
	start := time.Now()

	if r == nil {
		logError(ctx, "Cannot convert nil RoleOverallScore to protobuf", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil RoleOverallScore to protobuf")
	}

	logDebug(ctx, "Converting RoleOverallScore to protobuf", "role_name", r.RoleName)

	protoRole := &proto.RoleOverallScore{
		RoleName: r.RoleName,
		Score:    int32(r.Score),
	}

	logDebug(ctx, "RoleOverallScore conversion completed", 
		"role_name", r.RoleName, 
		"score", r.Score,
		"duration_ms", time.Since(start).Milliseconds())

	return protoRole, nil
}

// FromProto converts a protobuf RoleOverallScore to the native struct
func RoleOverallScoreFromProto(ctx context.Context, protoRole *proto.RoleOverallScore) (*RoleOverallScore, error) {
	start := time.Now()

	if protoRole == nil {
		logError(ctx, "Cannot convert nil protobuf RoleOverallScore", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil protobuf RoleOverallScore")
	}

	logDebug(ctx, "Converting protobuf to RoleOverallScore", "role_name", protoRole.GetRoleName())

	role := &RoleOverallScore{
		RoleName: protoRole.GetRoleName(),
		Score:    int(protoRole.GetScore()),
	}

	logDebug(ctx, "Protobuf RoleOverallScore conversion completed", 
		"role_name", role.RoleName, 
		"score", role.Score,
		"duration_ms", time.Since(start).Milliseconds())

	return role, nil
}

// --- Player Conversion Functions ---

// ToProto converts a Player struct to protobuf format
func (p *Player) ToProto(ctx context.Context) (*proto.Player, error) {
	start := time.Now()

	if p == nil {
		logError(ctx, "Cannot convert nil Player to protobuf", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil Player to protobuf")
	}

	logDebug(ctx, "Converting Player to protobuf", "player_uid", p.UID, "player_name", p.Name)

	// Convert RoleSpecificOveralls
	var protoRoles []*proto.RoleOverallScore
	for _, role := range p.RoleSpecificOveralls {
		protoRole, err := role.ToProto(ctx)
		if err != nil {
			logError(ctx, "Failed to convert role to protobuf", 
				"error", err, 
				"player_uid", p.UID, 
				"role_name", role.RoleName)
			return nil, fmt.Errorf("failed to convert role %s to protobuf: %w", role.RoleName, err)
		}
		protoRoles = append(protoRoles, protoRole)
	}

	// Convert performance percentiles nested map
	protoPercentiles := make(map[string]*proto.PerformancePercentileMap)
	for key, innerMap := range p.PerformancePercentiles {
		protoPercentiles[key] = &proto.PerformancePercentileMap{
			Percentiles: innerMap,
		}
	}

	// Convert numeric attributes to int32
	protoNumericAttrs := make(map[string]int32)
	for key, value := range p.NumericAttributes {
		protoNumericAttrs[key] = int32(value)
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
		Pac:                     int32(p.PAC),
		Sho:                     int32(p.SHO),
		Pas:                     int32(p.PAS),
		Dri:                     int32(p.DRI),
		Def:                     int32(p.DEF),
		Phy:                     int32(p.PHY),
		Gk:                      int32(p.GK),
		Div:                     int32(p.DIV),
		Han:                     int32(p.HAN),
		Ref:                     int32(p.REF),
		Kic:                     int32(p.KIC),
		Spd:                     int32(p.SPD),
		Pos:                     int32(p.POS),
		Overall:                 int32(p.Overall),
		BestRoleOverall:         p.BestRoleOverall,
		RoleSpecificOveralls:    protoRoles,
		TransferValueAmount:     p.TransferValueAmount,
		WageAmount:              p.WageAmount,
	}

	logDebug(ctx, "Player conversion to protobuf completed", 
		"player_uid", p.UID, 
		"player_name", p.Name,
		"role_count", len(protoRoles),
		"duration_ms", time.Since(start).Milliseconds())

	return protoPlayer, nil
}

// FromProto converts a protobuf Player to the native struct
func PlayerFromProto(ctx context.Context, protoPlayer *proto.Player) (*Player, error) {
	start := time.Now()

	if protoPlayer == nil {
		logError(ctx, "Cannot convert nil protobuf Player", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil protobuf Player")
	}

	logDebug(ctx, "Converting protobuf to Player", "player_uid", protoPlayer.GetUid(), "player_name", protoPlayer.GetName())

	// Convert RoleSpecificOveralls
	var roles []RoleOverallScore
	for _, protoRole := range protoPlayer.GetRoleSpecificOveralls() {
		role, err := RoleOverallScoreFromProto(ctx, protoRole)
		if err != nil {
			logError(ctx, "Failed to convert protobuf role", 
				"error", err, 
				"player_uid", protoPlayer.GetUid(), 
				"role_name", protoRole.GetRoleName())
			return nil, fmt.Errorf("failed to convert protobuf role %s: %w", protoRole.GetRoleName(), err)
		}
		roles = append(roles, *role)
	}

	// Convert performance percentiles nested map
	percentiles := make(map[string]map[string]float64)
	for key, protoMap := range protoPlayer.GetPerformancePercentiles() {
		percentiles[key] = protoMap.GetPercentiles()
	}

	// Convert numeric attributes from int32
	numericAttrs := make(map[string]int)
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

	logDebug(ctx, "Protobuf Player conversion completed", 
		"player_uid", player.UID, 
		"player_name", player.Name,
		"role_count", len(roles),
		"duration_ms", time.Since(start).Milliseconds())

	return player, nil
}

// --- DatasetData Conversion Functions ---

// ToProto converts a PlayerDataWithCurrency struct to protobuf format
func (d *PlayerDataWithCurrency) ToProto(ctx context.Context) (*proto.DatasetData, error) {
	start := time.Now()

	if d == nil {
		logError(ctx, "Cannot convert nil DatasetData to protobuf", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil DatasetData to protobuf")
	}

	logDebug(ctx, "Converting DatasetData to protobuf", "player_count", len(d.Players))

	var protoPlayers []*proto.Player
	for i, player := range d.Players {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			logError(ctx, "Failed to convert player to protobuf", 
				"error", err, 
				"player_index", i, 
				"player_uid", player.UID)
			return nil, fmt.Errorf("failed to convert player %d (UID: %d) to protobuf: %w", i, player.UID, err)
		}
		protoPlayers = append(protoPlayers, protoPlayer)
	}

	protoDataset := &proto.DatasetData{
		Players:        protoPlayers,
		CurrencySymbol: d.CurrencySymbol,
	}

	logDebug(ctx, "DatasetData conversion to protobuf completed", 
		"player_count", len(protoPlayers),
		"currency_symbol", d.CurrencySymbol,
		"duration_ms", time.Since(start).Milliseconds())

	return protoDataset, nil
}

// FromProto converts a protobuf DatasetData to the native struct
func DatasetDataFromProto(ctx context.Context, protoDataset *proto.DatasetData) (*PlayerDataWithCurrency, error) {
	start := time.Now()

	if protoDataset == nil {
		logError(ctx, "Cannot convert nil protobuf DatasetData", "error", "nil_input")
		return nil, fmt.Errorf("cannot convert nil protobuf DatasetData")
	}

	logDebug(ctx, "Converting protobuf to DatasetData", "player_count", len(protoDataset.GetPlayers()))

	var players []Player
	for i, protoPlayer := range protoDataset.GetPlayers() {
		player, err := PlayerFromProto(ctx, protoPlayer)
		if err != nil {
			logError(ctx, "Failed to convert protobuf player", 
				"error", err, 
				"player_index", i, 
				"player_uid", protoPlayer.GetUid())
			return nil, fmt.Errorf("failed to convert protobuf player %d (UID: %d): %w", i, protoPlayer.GetUid(), err)
		}
		players = append(players, *player)
	}

	dataset := &PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: protoDataset.GetCurrencySymbol(),
	}

	logDebug(ctx, "Protobuf DatasetData conversion completed", 
		"player_count", len(players),
		"currency_symbol", dataset.CurrencySymbol,
		"duration_ms", time.Since(start).Milliseconds())

	return dataset, nil
}