// src/api/services/player_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Player represents a football player with all their attributes
type Player struct {
	ID                     string                            `json:"id"`
	Name                   string                            `json:"name"`
	Age                    int                               `json:"age"`
	Club                   string                            `json:"club"`
	Position               string                            `json:"position"`
	Overall                int                               `json:"overall"`
	Potential              int                               `json:"potential"`
	TransferValue          string                            `json:"transferValue"`
	Salary                 string                            `json:"salary"`
	Nationality            string                            `json:"nationality"`
	Attributes             map[string]interface{}            `json:"attributes"`
	NumericAttributes      map[string]int                    `json:"numericAttributes"`
	PerformanceStats       map[string]string                 `json:"performanceStats"`
	PerformanceStatsNumeric map[string]float64               `json:"performanceStatsNumeric"`
	PerformancePercentiles map[string]map[string]float64     `json:"performancePercentiles"`
	RoleSpecificOveralls   []RoleSpecificOverall             `json:"roleSpecificOveralls"`
	ParsedPositions        []string                          `json:"parsedPositions"`
	ShortPositions         []string                          `json:"shortPositions"`
	PositionGroups         []string                          `json:"positionGroups"`
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

// NewPlayerService creates a new player service
func NewPlayerService(storage StorageInterface) *PlayerService {
	return &PlayerService{
		storage: storage,
	}
}

// GetPlayersByDatasetID retrieves players for a specific dataset
func (s *PlayerService) GetPlayersByDatasetID(ctx context.Context, datasetID string) ([]Player, string, error) {
	if datasetID == "" {
		return nil, "", fmt.Errorf("dataset ID cannot be empty")
	}

	players, currencySymbol, found := s.storage.GetPlayerData(datasetID)
	if !found {
		return nil, "", fmt.Errorf("dataset not found: %s", datasetID)
	}

	log.Printf("Retrieved %d players for dataset %s", len(players), datasetID)
	return players, currencySymbol, nil
}

// StorePlayerData stores player data with the given dataset ID
func (s *PlayerService) StorePlayerData(ctx context.Context, datasetID string, players []Player, currencySymbol string) error {
	if datasetID == "" {
		return fmt.Errorf("dataset ID cannot be empty")
	}

	if len(players) == 0 {
		return fmt.Errorf("no players to store")
	}

	// Validate players before storing
	for i, player := range players {
		if player.Name == "" {
			return fmt.Errorf("player at index %d has no name", i)
		}
	}

	err := s.storage.SetPlayerData(datasetID, players, currencySymbol)
	if err != nil {
		return fmt.Errorf("failed to store player data: %w", err)
	}

	log.Printf("Stored %d players for dataset %s with currency %s", len(players), datasetID, currencySymbol)
	return nil
}

// DeleteDataset removes a dataset and all its data
func (s *PlayerService) DeleteDataset(ctx context.Context, datasetID string) error {
	if datasetID == "" {
		return fmt.Errorf("dataset ID cannot be empty")
	}

	err := s.storage.DeleteDataset(datasetID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset %s: %w", datasetID, err)
	}

	log.Printf("Deleted dataset %s", datasetID)
	return nil
}

// GetAllDatasets returns all available dataset IDs
func (s *PlayerService) GetAllDatasets(ctx context.Context) []string {
	datasets := s.storage.GetAllDatasetIDs()
	log.Printf("Retrieved %d datasets", len(datasets))
	return datasets
}

// ProcessPlayerPercentiles calculates performance percentiles for players
func (s *PlayerService) ProcessPlayerPercentiles(ctx context.Context, players []Player) error {
	if len(players) == 0 {
		return fmt.Errorf("no players to process")
	}

	// This would call the existing percentile calculation logic
	// For now, this is a placeholder for the business logic
	log.Printf("Processing percentiles for %d players", len(players))
	
	// TODO: Move the actual percentile calculation logic here
	// CalculatePlayerPerformancePercentiles(players)
	
	return nil
}

// ValidatePlayerData performs validation on player data
func (s *PlayerService) ValidatePlayerData(ctx context.Context, players []Player) error {
	if len(players) == 0 {
		return fmt.Errorf("no players provided for validation")
	}

	var errors []string

	for i, player := range players {
		// Basic validation
		if player.Name == "" {
			errors = append(errors, fmt.Sprintf("player %d: missing name", i))
		}

		if player.Age < 15 || player.Age > 50 {
			errors = append(errors, fmt.Sprintf("player %d (%s): invalid age %d", i, player.Name, player.Age))
		}

		if player.Overall < 1 || player.Overall > 100 {
			errors = append(errors, fmt.Sprintf("player %d (%s): invalid overall rating %d", i, player.Name, player.Overall))
		}

		if player.Club == "" {
			errors = append(errors, fmt.Sprintf("player %d (%s): missing club", i, player.Name))
		}

		if player.Position == "" {
			errors = append(errors, fmt.Sprintf("player %d (%s): missing position", i, player.Name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	log.Printf("Validated %d players successfully", len(players))
	return nil
}

// GetPlayerStatistics calculates basic statistics for a set of players
func (s *PlayerService) GetPlayerStatistics(ctx context.Context, players []Player) map[string]interface{} {
	if len(players) == 0 {
		return map[string]interface{}{
			"total": 0,
		}
	}

	stats := map[string]interface{}{
		"total":           len(players),
		"timestamp":       time.Now().Unix(),
	}

	// Calculate basic stats
	totalOverall := 0
	maxOverall := 0
	minOverall := 100
	positions := make(map[string]int)
	clubs := make(map[string]int)

	for _, player := range players {
		totalOverall += player.Overall
		
		if player.Overall > maxOverall {
			maxOverall = player.Overall
		}
		
		if player.Overall < minOverall {
			minOverall = player.Overall
		}

		positions[player.Position]++
		clubs[player.Club]++
	}

	stats["average_overall"] = totalOverall / len(players)
	stats["max_overall"] = maxOverall
	stats["min_overall"] = minOverall
	stats["unique_positions"] = len(positions)
	stats["unique_clubs"] = len(clubs)

	return stats
}