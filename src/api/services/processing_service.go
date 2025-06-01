// src/api/services/processing_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// ProcessingService handles file processing and data transformation
type ProcessingService struct {
	playerService *PlayerService
}

// ProcessingResult contains the results of file processing
type ProcessingResult struct {
	DatasetID      string        `json:"datasetId"`
	PlayersCount   int           `json:"playersCount"`
	CurrencySymbol string        `json:"currencySymbol"`
	ProcessingTime time.Duration `json:"processingTime"`
	Errors         []string      `json:"errors,omitempty"`
}

// ProcessingOptions configures how file processing should be performed
type ProcessingOptions struct {
	MaxWorkers    int
	BufferSize    int
	EnableMetrics bool
	EnableTracing bool
}

// NewProcessingService creates a new processing service
func NewProcessingService(playerService *PlayerService) *ProcessingService {
	return &ProcessingService{
		playerService: playerService,
	}
}

// ProcessPlayerFile processes an uploaded player file
func (s *ProcessingService) ProcessPlayerFile(ctx context.Context, fileContent []byte, filename string, options ProcessingOptions) (*ProcessingResult, error) {
	startTime := time.Now()

	if len(fileContent) == 0 {
		return nil, fmt.Errorf("file content is empty")
	}

	// Validate file format
	if err := s.validateFileFormat(filename, fileContent); err != nil {
		return nil, fmt.Errorf("invalid file format: %w", err)
	}

	// Set default processing options
	if options.MaxWorkers == 0 {
		options.MaxWorkers = runtime.NumCPU()
	}
	if options.BufferSize == 0 {
		options.BufferSize = s.calculateOptimalBufferSize(options.MaxWorkers, int64(len(fileContent)))
	}

	// Generate unique dataset ID
	datasetID := s.generateDatasetID()

	// Process the file content
	players, currencySymbol := s.parsePlayerData(ctx, fileContent, options)

	// Validate processed data
	if err := s.playerService.ValidatePlayerData(ctx, players); err != nil {
		log.Printf("Warning: Player data validation issues: %v", err)
		// Don't fail processing for validation warnings, just log them
	}

	// Store the processed data
	if err := s.playerService.StorePlayerData(ctx, datasetID, players, currencySymbol); err != nil {
		return nil, fmt.Errorf("failed to store player data: %w", err)
	}

	// Process percentiles asynchronously
	go func() {
		if err := s.processPercentilesAsync(datasetID, players, currencySymbol); err != nil {
			log.Printf("Error processing percentiles for dataset %s: %v", datasetID, err)
		}
	}()

	processingTime := time.Since(startTime)

	result := &ProcessingResult{
		DatasetID:      datasetID,
		PlayersCount:   len(players),
		CurrencySymbol: currencySymbol,
		ProcessingTime: processingTime,
	}

	log.Printf("Successfully processed file %s: %d players in %v", filename, len(players), processingTime)
	return result, nil
}

// validateFileFormat checks if the file is in a supported format
func (s *ProcessingService) validateFileFormat(filename string, content []byte) error {
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filename), ".html") {
		return fmt.Errorf("unsupported file format, expected .html file")
	}

	// Check content starts with HTML
	contentStr := string(content[:minInt(100, len(content))])
	if !strings.Contains(strings.ToLower(contentStr), "<html") &&
		!strings.Contains(strings.ToLower(contentStr), "<!doctype") {
		return fmt.Errorf("file does not appear to be valid HTML")
	}

	return nil
}

// parsePlayerData parses the HTML content and extracts player data
func (s *ProcessingService) parsePlayerData(_ctx context.Context, _content []byte, _options ProcessingOptions) (players []Player, currencySymbol string) {
	// TODO: Integrate with existing parsing logic
	// This would call the existing ParseHTMLPlayerTable function
	players = []Player{}
	currencySymbol = "$"

	log.Printf("Parsed %d players from HTML content", len(players))
	return players, currencySymbol
}

// PlayerParseResult represents the result of parsing a single player
type PlayerParseResult struct {
	Player Player
	Error  error
}

// processPercentilesAsync calculates percentiles in the background
func (s *ProcessingService) processPercentilesAsync(datasetID string, players []Player, currencySymbol string) error {
	log.Printf("Starting async percentile calculation for dataset %s", datasetID)

	// Calculate percentiles
	if err := s.playerService.ProcessPlayerPercentiles(context.Background(), players); err != nil {
		return fmt.Errorf("failed to process percentiles: %w", err)
	}

	// Update stored dataset with percentiles
	if err := s.playerService.StorePlayerData(context.Background(), datasetID, players, currencySymbol); err != nil {
		return fmt.Errorf("failed to update dataset with percentiles: %w", err)
	}

	log.Printf("Completed async percentile calculation for dataset %s", datasetID)
	return nil
}

// calculateOptimalBufferSize determines the optimal buffer size for processing
func (s *ProcessingService) calculateOptimalBufferSize(numWorkers int, fileSize int64) int {
	const baseBufferMultiplier = 10
	const maxBufferSize = 1000
	const minBufferSize = 20

	// Base calculation on number of workers
	baseBuffer := numWorkers * baseBufferMultiplier

	// Adjust based on file size (larger files need bigger buffers)
	sizeAdjustment := int(fileSize / (1024 * 1024)) // MB
	adjustedBuffer := baseBuffer + sizeAdjustment

	// Ensure within reasonable bounds
	if adjustedBuffer > maxBufferSize {
		return maxBufferSize
	}
	if adjustedBuffer < minBufferSize {
		return minBufferSize
	}

	return adjustedBuffer
}

// generateDatasetID creates a unique identifier for the dataset
func (s *ProcessingService) generateDatasetID() string {
	return fmt.Sprintf("dataset_%d", time.Now().UnixNano())
}

// GetProcessingStats returns statistics about processing performance
func (s *ProcessingService) GetProcessingStats(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"available_workers": runtime.NumCPU(),
		"max_buffer_size":   1000,
		"supported_formats": []string{"html"},
		"timestamp":         time.Now().Unix(),
	}

	return stats
}

// Helper function for min calculation
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
