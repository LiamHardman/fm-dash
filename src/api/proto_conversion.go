package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"api/proto"

	"go.opentelemetry.io/otel/attribute"
)

// safeIntToInt32 safely converts int to int32, clamping to int32 range
func safeIntToInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

// --- RoleOverallScore Conversion Functions ---

// --- Player Conversion Functions ---

// ToProto converts a Player struct to protobuf format (optimized for frontend)
func (p *Player) ToProto(ctx context.Context) (*proto.Player, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.player_to_proto", []attribute.KeyValue{
		attribute.String("conversion.type", "player"),
		attribute.String("conversion.direction", "to_protobuf"),
	})
	defer span.End()

	start := time.Now()

	if p == nil {
		RecordError(ctx, ErrNilPlayer, "Cannot convert nil Player to protobuf",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilPlayer
	}

	SetSpanAttributes(ctx,
		attribute.Int64("player.uid", p.UID),
		attribute.String("player.name", p.Name),
		attribute.String("player.position", p.Position),
		attribute.String("player.club", p.Club),
	)

	logDebug(ctx, "Converting Player to protobuf (optimized)",
		"player_uid", p.UID,
		"player_name", p.Name,
		"conversion_type", "player",
		"conversion_direction", "to_protobuf")

	// Create copies of position data to avoid race conditions
	parsedPositionsCopy := make([]string, len(p.ParsedPositions))
	copy(parsedPositionsCopy, p.ParsedPositions)

	shortPositionsCopy := make([]string, len(p.ShortPositions))
	copy(shortPositionsCopy, p.ShortPositions)

	positionGroupsCopy := make([]string, len(p.PositionGroups))
	copy(positionGroupsCopy, p.PositionGroups)

	// Only include essential attributes for display
	essentialAttributes := make(map[string]string)
	if p.Attributes != nil {
		// Only include the most important attributes for display
		importantAttrs := []string{"Av Rat", "Apps", "Mins", "Gls/90", "Asts/90", "Pas %", "Tck/90", "Int/90"}
		for _, attr := range importantAttrs {
			if value, exists := p.Attributes[attr]; exists {
				essentialAttributes[attr] = value
			}
		}
	}

	protoPlayer := &proto.Player{
		Uid:                 p.UID,
		Name:                p.Name,
		Position:            p.Position,
		Age:                 p.Age,
		Club:                p.Club,
		Division:            p.Division,
		TransferValue:       p.TransferValue,
		Wage:                p.Wage,
		Personality:         p.Personality,
		MediaHandling:       p.MediaHandling,
		Nationality:         p.Nationality,
		NationalityIso:      p.NationalityISO,
		NationalityFifaCode: p.NationalityFIFACode,
		AttributeMasked:     p.AttributeMasked,

		// Essential FIFA-style stats
		Pac:     safeIntToInt32(p.PAC),
		Sho:     safeIntToInt32(p.SHO),
		Pas:     safeIntToInt32(p.PAS),
		Dri:     safeIntToInt32(p.DRI),
		Def:     safeIntToInt32(p.DEF),
		Phy:     safeIntToInt32(p.PHY),
		Gk:      safeIntToInt32(p.GK),
		Div:     safeIntToInt32(p.DIV),
		Han:     safeIntToInt32(p.HAN),
		Ref:     safeIntToInt32(p.REF),
		Kic:     safeIntToInt32(p.KIC),
		Spd:     safeIntToInt32(p.SPD),
		Pos:     safeIntToInt32(p.POS),
		Overall: safeIntToInt32(p.Overall),

		// Position data
		ParsedPositions: parsedPositionsCopy,
		ShortPositions:  shortPositionsCopy,
		PositionGroups:  positionGroupsCopy,

		// Essential attributes only
		EssentialAttributes: essentialAttributes,

		// Best role for display
		BestRoleOverall: p.BestRoleOverall,

		// Numeric values for sorting
		TransferValueAmount: p.TransferValueAmount,
		WageAmount:          p.WageAmount,
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Int("conversion.essential_attributes_count", len(essentialAttributes)),
	)

	logDebug(ctx, "Player conversion to protobuf completed (optimized)",
		"player_uid", p.UID,
		"player_name", p.Name,
		"essential_attributes_count", len(essentialAttributes),
		"duration_ms", duration.Milliseconds())

	return protoPlayer, nil
}

// PlayerFromProto converts a protobuf Player to the native struct
func PlayerFromProto(ctx context.Context, protoPlayer *proto.Player) (*Player, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.player_from_proto", []attribute.KeyValue{
		attribute.String("conversion.type", "player"),
		attribute.String("conversion.direction", "from_protobuf"),
	})
	defer span.End()

	start := time.Now()

	if protoPlayer == nil {
		RecordError(ctx, ErrNilProtobufPlayer, "Cannot convert nil protobuf Player",
			WithErrorCategory("validation"),
			WithSeverity("medium"))
		return nil, ErrNilProtobufPlayer
	}

	SetSpanAttributes(ctx,
		attribute.Int64("player.uid", protoPlayer.GetUid()),
		attribute.String("player.name", protoPlayer.GetName()),
		attribute.String("player.position", protoPlayer.GetPosition()),
		attribute.String("player.club", protoPlayer.GetClub()),
	)

	logDebug(ctx, "Converting protobuf to Player (optimized)",
		"player_uid", protoPlayer.GetUid(),
		"player_name", protoPlayer.GetName(),
		"conversion_type", "player",
		"conversion_direction", "from_protobuf",
		"essential_attributes_count", len(protoPlayer.GetEssentialAttributes()))

	// Convert essential attributes from string
	attributes := make(map[string]string)
	for key, value := range protoPlayer.GetEssentialAttributes() {
		attributes[key] = value
	}

	player := &Player{
		UID:                 protoPlayer.GetUid(),
		Name:                protoPlayer.GetName(),
		Position:            protoPlayer.GetPosition(),
		Age:                 protoPlayer.GetAge(),
		Club:                protoPlayer.GetClub(),
		Division:            protoPlayer.GetDivision(),
		TransferValue:       protoPlayer.GetTransferValue(),
		Wage:                protoPlayer.GetWage(),
		Personality:         protoPlayer.GetPersonality(),
		MediaHandling:       protoPlayer.GetMediaHandling(),
		Nationality:         protoPlayer.GetNationality(),
		NationalityISO:      protoPlayer.GetNationalityIso(),
		NationalityFIFACode: protoPlayer.GetNationalityFifaCode(),
		AttributeMasked:     protoPlayer.GetAttributeMasked(),
		Attributes:          attributes,
		ParsedPositions:     protoPlayer.GetParsedPositions(),
		ShortPositions:      protoPlayer.GetShortPositions(),
		PositionGroups:      protoPlayer.GetPositionGroups(),
		PAC:                 int(protoPlayer.GetPac()),
		SHO:                 int(protoPlayer.GetSho()),
		PAS:                 int(protoPlayer.GetPas()),
		DRI:                 int(protoPlayer.GetDri()),
		DEF:                 int(protoPlayer.GetDef()),
		PHY:                 int(protoPlayer.GetPhy()),
		GK:                  int(protoPlayer.GetGk()),
		DIV:                 int(protoPlayer.GetDiv()),
		HAN:                 int(protoPlayer.GetHan()),
		REF:                 int(protoPlayer.GetRef()),
		KIC:                 int(protoPlayer.GetKic()),
		SPD:                 int(protoPlayer.GetSpd()),
		POS:                 int(protoPlayer.GetPos()),
		Overall:             int(protoPlayer.GetOverall()),
		BestRoleOverall:     protoPlayer.GetBestRoleOverall(),
		TransferValueAmount: protoPlayer.GetTransferValueAmount(),
		WageAmount:          protoPlayer.GetWageAmount(),
	}

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Int("conversion.essential_attributes_count", len(attributes)),
	)

	logDebug(ctx, "Protobuf Player conversion completed (optimized)",
		"player_uid", player.UID,
		"player_name", player.Name,
		"essential_attributes_count", len(attributes),
		"duration_ms", duration.Milliseconds())

	return player, nil
}

// --- DatasetData Conversion Functions ---

// ToProto converts a PlayerDataWithCurrency struct to protobuf format
func (d *PlayerDataWithCurrency) ToProto(ctx context.Context) (*proto.DatasetData, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.dataset_to_proto", []attribute.KeyValue{
		attribute.String("conversion.type", "dataset_data"),
		attribute.String("conversion.direction", "to_protobuf"),
	})
	defer span.End()

	start := time.Now()

	if d == nil {
		RecordError(ctx, ErrNilDatasetData, "Cannot convert nil DatasetData to protobuf",
			WithErrorCategory("validation"),
			WithSeverity("high"))
		return nil, ErrNilDatasetData
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.player_count", len(d.Players)),
		attribute.String("dataset.currency_symbol", d.CurrencySymbol),
	)

	logDebug(ctx, "Converting DatasetData to protobuf",
		"player_count", len(d.Players),
		"conversion_type", "dataset_data",
		"conversion_direction", "to_protobuf",
		"currency_symbol", d.CurrencySymbol)

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

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Int("conversion.players_converted", len(protoPlayers)),
	)

	logDebug(ctx, "DatasetData conversion to protobuf completed",
		"player_count", len(protoPlayers),
		"currency_symbol", d.CurrencySymbol,
		"duration_ms", duration.Milliseconds())

	return protoDataset, nil
}

// DatasetDataFromProto converts a protobuf DatasetData to the native struct
func DatasetDataFromProto(ctx context.Context, protoDataset *proto.DatasetData) (*PlayerDataWithCurrency, error) {
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.conversion.dataset_from_proto", []attribute.KeyValue{
		attribute.String("conversion.type", "dataset_data"),
		attribute.String("conversion.direction", "from_protobuf"),
	})
	defer span.End()

	start := time.Now()

	if protoDataset == nil {
		RecordError(ctx, ErrNilProtobufDatasetData, "Cannot convert nil protobuf DatasetData",
			WithErrorCategory("validation"),
			WithSeverity("high"))
		return nil, ErrNilProtobufDatasetData
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.player_count", len(protoDataset.GetPlayers())),
		attribute.String("dataset.currency_symbol", protoDataset.GetCurrencySymbol()),
	)

	logDebug(ctx, "Converting protobuf to DatasetData",
		"player_count", len(protoDataset.GetPlayers()),
		"conversion_type", "dataset_data",
		"conversion_direction", "from_protobuf",
		"currency_symbol", protoDataset.GetCurrencySymbol())

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

	duration := time.Since(start)
	SetSpanAttributes(ctx,
		attribute.Float64("conversion.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("conversion.success", true),
		attribute.Int("conversion.players_converted", len(players)),
	)

	logDebug(ctx, "Protobuf DatasetData conversion completed",
		"player_count", len(players),
		"currency_symbol", dataset.CurrencySymbol,
		"duration_ms", duration.Milliseconds())

	return dataset, nil
}
