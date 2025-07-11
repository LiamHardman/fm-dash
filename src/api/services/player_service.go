// src/api/services/player_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Player represents a football player with all their attributes
type Player struct {
	ID                      string                        `json:"id"`
	UID                     int64                         `json:"uid"` // Unique identifier for the player
	Name                    string                        `json:"name"`
	Age                     int                           `json:"age"`
	Club                    string                        `json:"club"`
	Position                string                        `json:"position"`
	Overall                 int                           `json:"overall"`
	Potential               int                           `json:"potential"`
	TransferValue           string                        `json:"transferValue"`
	Salary                  string                        `json:"salary"`
	Nationality             string                        `json:"nationality"`
	Attributes              map[string]interface{}        `json:"attributes"`
	NumericAttributes       map[string]int                `json:"numericAttributes"`
	PerformanceStats        map[string]string             `json:"performanceStats"`
	PerformanceStatsNumeric map[string]float64            `json:"performanceStatsNumeric"`
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"`
	RoleSpecificOveralls    []RoleSpecificOverall         `json:"roleSpecificOveralls"`
	ParsedPositions         []string                      `json:"parsedPositions"`
	ShortPositions          []string                      `json:"shortPositions"`
	PositionGroups          []string                      `json:"positionGroups"`
}

// RoleSpecificOverall represents a player's rating for a specific role
type RoleSpecificOverall struct {
	Role  string `json:"role"`
	Score int    `json:"score"`
}

// PlayerService handles all player-related business logic
type PlayerService struct {
	storage StorageInterface
}

// StorageInterface defines the storage contract
type StorageInterface interface {
	GetPlayerData(datasetID string) ([]Player, string, bool)
	SetPlayerData(datasetID string, players []Player, currencySymbol string) error
	DeleteDataset(datasetID string) error
	GetAllDatasetIDs() []string
}

var (
	tracer = otel.Tracer("v2fmdash-player-service")
)

// NewPlayerService creates a new player service
func NewPlayerService(storage StorageInterface) *PlayerService {
	return &PlayerService{
		storage: storage,
	}
}

// GetPlayersByDatasetID retrieves players for a specific dataset
func (s *PlayerService) GetPlayersByDatasetID(ctx context.Context, datasetID string) ([]Player, string, error) {
	_, span := tracer.Start(ctx, "player_service.get_players_by_dataset_id",
		trace.WithAttributes(
			attribute.String("dataset.id", datasetID),
		))
	defer span.End()

	if datasetID == "" {
		err := fmt.Errorf("dataset ID cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid dataset ID")
		return nil, "", err
	}

	players, currencySymbol, found := s.storage.GetPlayerData(datasetID)
	if !found {
		err := fmt.Errorf("dataset not found: %s", datasetID)
		span.RecordError(err)
		span.SetStatus(codes.Error, "dataset not found")
		return nil, "", err
	}

	span.SetAttributes(
		attribute.Int("players.count", len(players)),
		attribute.String("currency.symbol", currencySymbol),
	)
	span.SetStatus(codes.Ok, "players retrieved successfully")

	log.Printf("Retrieved %d players for dataset %s", len(players), datasetID)
	return players, currencySymbol, nil
}

// StorePlayerData stores player data with the given dataset ID
func (s *PlayerService) StorePlayerData(ctx context.Context, datasetID string, players []Player, currencySymbol string) error {
	_, span := tracer.Start(ctx, "player_service.store_player_data",
		trace.WithAttributes(
			attribute.String("dataset.id", datasetID),
			attribute.Int("players.count", len(players)),
			attribute.String("currency.symbol", currencySymbol),
		))
	defer span.End()

	if datasetID == "" {
		err := fmt.Errorf("dataset ID cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid dataset ID")
		return err
	}

	if len(players) == 0 {
		err := fmt.Errorf("no players to store")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no players provided")
		return err
	}

	// Validate players before storing
	span.AddEvent("validation.start")
	for i := range players {
		if players[i].Name == "" {
			err := fmt.Errorf("player at index %d has no name", i)
			span.RecordError(err)
			span.SetStatus(codes.Error, "player validation failed")
			return err
		}
	}
	span.AddEvent("validation.complete")

	span.AddEvent("storage.start")
	err := s.storage.SetPlayerData(datasetID, players, currencySymbol)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "storage operation failed")
		return fmt.Errorf("failed to store player data: %w", err)
	}
	span.AddEvent("storage.complete")

	span.SetStatus(codes.Ok, "player data stored successfully")
	log.Printf("Stored %d players for dataset %s with currency %s", len(players), datasetID, currencySymbol)
	return nil
}

// DeleteDataset removes a dataset and all its data
func (s *PlayerService) DeleteDataset(ctx context.Context, datasetID string) error {
	_, span := tracer.Start(ctx, "player_service.delete_dataset",
		trace.WithAttributes(
			attribute.String("dataset.id", datasetID),
		))
	defer span.End()

	if datasetID == "" {
		err := fmt.Errorf("dataset ID cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid dataset ID")
		return err
	}

	err := s.storage.DeleteDataset(datasetID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "delete operation failed")
		return fmt.Errorf("failed to delete dataset %s: %w", datasetID, err)
	}

	span.SetStatus(codes.Ok, "dataset deleted successfully")
	log.Printf("Deleted dataset %s", datasetID)
	return nil
}

// GetAllDatasets returns all available dataset IDs
func (s *PlayerService) GetAllDatasets(ctx context.Context) []string {
	_, span := tracer.Start(ctx, "player_service.get_all_datasets")
	defer span.End()

	datasets := s.storage.GetAllDatasetIDs()

	span.SetAttributes(
		attribute.Int("datasets.count", len(datasets)),
	)
	span.SetStatus(codes.Ok, "datasets retrieved successfully")

	log.Printf("Retrieved %d datasets", len(datasets))
	return datasets
}

// ProcessPlayerPercentiles calculates performance percentiles for players
func (s *PlayerService) ProcessPlayerPercentiles(ctx context.Context, players []Player) error {
	_, span := tracer.Start(ctx, "player_service.process_player_percentiles",
		trace.WithAttributes(
			attribute.Int("players.count", len(players)),
		))
	defer span.End()

	if len(players) == 0 {
		err := fmt.Errorf("no players to process")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no players provided")
		return err
	}

	// This would call the existing percentile calculation logic
	// For now, this is a placeholder for the business logic
	log.Printf("Processing percentiles for %d players", len(players))

	// TODO: Move the actual percentile calculation logic here
	// CalculatePlayerPerformancePercentiles(players)

	span.SetStatus(codes.Ok, "percentiles processed successfully")
	return nil
}

// ValidatePlayerData performs validation on player data
func (s *PlayerService) ValidatePlayerData(ctx context.Context, players []Player) error {
	_, span := tracer.Start(ctx, "player_service.validate_player_data",
		trace.WithAttributes(
			attribute.Int("players.count", len(players)),
		))
	defer span.End()

	if len(players) == 0 {
		err := fmt.Errorf("no players provided for validation")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no players provided")
		return err
	}

	var errors []string

	for i := range players {
		// Basic validation
		if players[i].Name == "" {
			errors = append(errors, fmt.Sprintf("player %d: missing name", i))
		}

		if players[i].Age < 15 || players[i].Age > 50 {
			errors = append(errors, fmt.Sprintf("player %d (%s): invalid age %d", i, players[i].Name, players[i].Age))
		}

		if players[i].Overall < 1 || players[i].Overall > 100 {
			errors = append(errors, fmt.Sprintf("player %d (%s): invalid overall rating %d", i, players[i].Name, players[i].Overall))
		}

		if players[i].Club == "" {
			errors = append(errors, fmt.Sprintf("player %d (%s): missing club", i, players[i].Name))
		}

		if players[i].Position == "" {
			errors = append(errors, fmt.Sprintf("player %d (%s): missing position", i, players[i].Name))
		}
	}

	if len(errors) > 0 {
		err := fmt.Errorf("validation failed: %v", errors)
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		span.SetAttributes(
			attribute.Int("validation.errors.count", len(errors)),
		)
		return err
	}

	span.SetStatus(codes.Ok, "validation successful")
	log.Printf("Validated %d players successfully", len(players))
	return nil
}

// GetPlayerStatistics calculates basic statistics for a set of players
func (s *PlayerService) GetPlayerStatistics(ctx context.Context, players []Player) map[string]interface{} {
	_, span := tracer.Start(ctx, "player_service.get_player_statistics",
		trace.WithAttributes(
			attribute.Int("players.count", len(players)),
		))
	defer span.End()

	if len(players) == 0 {
		span.SetStatus(codes.Ok, "no players provided")
		return map[string]interface{}{
			"total": 0,
		}
	}

	stats := map[string]interface{}{
		"total":     len(players),
		"timestamp": time.Now().Unix(),
	}

	// Calculate basic stats
	totalOverall := 0
	maxOverall := 0
	minOverall := 100
	positions := make(map[string]int)
	clubs := make(map[string]int)

	for i := range players {
		totalOverall += players[i].Overall

		if players[i].Overall > maxOverall {
			maxOverall = players[i].Overall
		}

		if players[i].Overall < minOverall {
			minOverall = players[i].Overall
		}

		positions[players[i].Position]++
		clubs[players[i].Club]++
	}

	stats["average_overall"] = totalOverall / len(players)
	stats["max_overall"] = maxOverall
	stats["min_overall"] = minOverall
	stats["unique_positions"] = len(positions)
	stats["unique_clubs"] = len(clubs)

	return stats
}
