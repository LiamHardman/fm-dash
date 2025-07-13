package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDefaultProcessingConfig(t *testing.T) {
	config := DefaultProcessingConfig()

	// Check that default config has reasonable values
	if config.BatchSize <= 0 {
		t.Errorf("DefaultProcessingConfig BatchSize = %d; want > 0", config.BatchSize)
	}

	if config.WorkerCount <= 0 {
		t.Errorf("DefaultProcessingConfig WorkerCount = %d; want > 0", config.WorkerCount)
	}

	if config.BufferSize <= 0 {
		t.Errorf("DefaultProcessingConfig BufferSize = %d; want > 0", config.BufferSize)
	}

	if config.MaxGoroutines <= 0 {
		t.Errorf("DefaultProcessingConfig MaxGoroutines = %d; want > 0", config.MaxGoroutines)
	}

	// Check specific default values
	expectedBatchSize := 100
	if config.BatchSize != expectedBatchSize {
		t.Errorf("DefaultProcessingConfig BatchSize = %d; want %d", config.BatchSize, expectedBatchSize)
	}
}

func TestCreatePlayerProcessor(t *testing.T) {
	config := ProcessingConfig{WorkerCount: 2, BufferSize: 100}
	processor := CreatePlayerProcessor(config)

	if processor == nil {
		t.Fatal("CreatePlayerProcessor returned nil")
	}

	// Test that processor has the expected configuration
	// Since the fields might be private, we'll test functionality instead
	testPlayers := []Player{
		{Name: "Player1", Overall: 75},
		{Name: "Player2", Overall: 80},
	}

	// Test that we can process without errors
	resultCh := processor.ProcessPlayersAsync(testPlayers)
	if resultCh == nil {
		t.Error("ProcessPlayersAsync returned nil channel")
	}

	// Collect results
	var results []Player
	for player := range resultCh {
		results = append(results, player)
	}

	if len(results) != len(testPlayers) {
		t.Errorf("ProcessPlayersAsync returned %d results; want %d", len(results), len(testPlayers))
	}

	// Clean up
	processor.Shutdown()
}

func TestProcessPlayersAsync(t *testing.T) {
	config := DefaultProcessingConfig()
	config.WorkerCount = 2 // Use fewer workers for testing
	config.BatchSize = 1   // Small batch size for testing
	processor := CreatePlayerProcessor(config)
	defer processor.Shutdown()

	testPlayers := []Player{
		{Name: "Player1", Overall: 75, Age: "25"},
		{Name: "Player2", Overall: 80, Age: "27"},
		{Name: "Player3", Overall: 85, Age: "23"},
	}

	resultCh := processor.ProcessPlayersAsync(testPlayers)

	// Collect results
	var results []Player
	for player := range resultCh {
		results = append(results, player)
	}

	if len(results) != len(testPlayers) {
		t.Errorf("ProcessPlayersAsync returned %d results; want %d", len(results), len(testPlayers))
	}

	// Verify that all players were processed (names should match)
	resultNames := make(map[string]bool)
	for _, result := range results {
		resultNames[result.Name] = true
	}

	for _, original := range testPlayers {
		if !resultNames[original.Name] {
			t.Errorf("Player %s not found in results", original.Name)
		}
	}
}

func TestProcessPlayersAsyncEmpty(t *testing.T) {
	config := DefaultProcessingConfig()
	processor := CreatePlayerProcessor(config)
	defer processor.Shutdown()

	emptyPlayers := []Player{}

	resultCh := processor.ProcessPlayersAsync(emptyPlayers)

	// Collect results
	var results []Player
	for player := range resultCh {
		results = append(results, player)
	}

	if len(results) != 0 {
		t.Errorf("ProcessPlayersAsync with empty input returned %d results; want 0", len(results))
	}
}

func TestProcessPlayersBatch(t *testing.T) {
	ctx := context.Background()
	testPlayers := []Player{
		{Name: "Player1", Overall: 75},
		{Name: "Player2", Overall: 80},
	}

	results, err := ProcessPlayersBatch(ctx, testPlayers, 1)
	if err != nil {
		t.Errorf("ProcessPlayersBatch returned error: %v", err)
	}

	if len(results) != len(testPlayers) {
		t.Errorf("ProcessPlayersBatch returned %d results; want %d", len(results), len(testPlayers))
	}

	// Verify basic processing occurred
	resultNames := make(map[string]bool)
	for _, result := range results {
		resultNames[result.Name] = true
	}

	for _, original := range testPlayers {
		if !resultNames[original.Name] {
			t.Errorf("Player %s not found in batch results", original.Name)
		}
	}
}

func TestProcessPercentilesAsync(t *testing.T) {
	testDatasets := map[string][]Player{
		"dataset1": {
			{Name: "Player1", Overall: 75, PAC: 70, SHO: 65},
			{Name: "Player2", Overall: 80, PAC: 75, SHO: 70},
		},
		"dataset2": {
			{Name: "Player3", Overall: 85, PAC: 80, SHO: 75},
		},
	}

	resultCh := ProcessPercentilesAsync(testDatasets)

	// Collect results
	var results []PercentileResult
	for result := range resultCh {
		results = append(results, result)
	}

	// Should have one result per dataset
	if len(results) != len(testDatasets) {
		t.Errorf("ProcessPercentilesAsync returned %d results; want %d", len(results), len(testDatasets))
	}

	// Verify all datasets are represented
	datasetsSeen := make(map[string]bool)
	for _, result := range results {
		datasetsSeen[result.DatasetID] = true
	}

	for datasetID := range testDatasets {
		if !datasetsSeen[datasetID] {
			t.Errorf("Dataset %s not found in results", datasetID)
		}
	}
}

func TestCreateConcurrentLeagueProcessor(t *testing.T) {
	numWorkers := 4
	processor := CreateConcurrentLeagueProcessor(numWorkers)

	if processor == nil {
		t.Fatal("CreateConcurrentLeagueProcessor returned nil")
	}

	// Test functionality with sample data
	ctx := context.Background()
	testPlayers := []Player{
		{Name: "Player1", Division: "League1", Overall: 75},
		{Name: "Player2", Division: "League1", Overall: 80},
		{Name: "Player3", Division: "League2", Overall: 85},
	}

	leagues := processor.ProcessLeaguesAsync(ctx, testPlayers)

	if len(leagues) == 0 {
		t.Error("ProcessLeaguesAsync returned no leagues")
	}

	// Should have 2 leagues based on test data
	if len(leagues) != 2 {
		t.Errorf("ProcessLeaguesAsync returned %d leagues; want 2", len(leagues))
	}
}

func TestProcessLeaguesAsync(t *testing.T) {
	processor := CreateConcurrentLeagueProcessor(2)
	ctx := context.Background()

	// Create test players with enough players per team (11+ for each team)
	testPlayers := make([]Player, 0)

	// Arsenal players (Premier League)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:     fmt.Sprintf("Arsenal Player %d", i),
			Division: "Premier League",
			Club:     "Arsenal",
			Overall:  75 + i,
		})
	}

	// Chelsea players (Premier League)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:     fmt.Sprintf("Chelsea Player %d", i),
			Division: "Premier League",
			Club:     "Chelsea",
			Overall:  80 + i,
		})
	}

	// Leeds United players (Championship)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:     fmt.Sprintf("Leeds Player %d", i),
			Division: "Championship",
			Club:     "Leeds United",
			Overall:  70 + i,
		})
	}

	leagues := processor.ProcessLeaguesAsync(ctx, testPlayers)

	// Should return 2 leagues
	if len(leagues) != 2 {
		t.Errorf("ProcessLeaguesAsync returned %d leagues; want 2", len(leagues))
	}

	// Verify league data structure
	for _, league := range leagues {
		if league.Name == "" {
			t.Error("League name should not be empty")
		}
		if league.PlayerCount <= 0 {
			t.Error("League should have players")
		}
		if league.TeamCount <= 0 {
			t.Error("League should have teams")
		}
	}
}

func TestCreateAsyncPlayerFilter(t *testing.T) {
	numWorkers := 3
	chunkSize := 10
	filter := CreateAsyncPlayerFilter(numWorkers, chunkSize)

	if filter == nil {
		t.Fatal("CreateAsyncPlayerFilter returned nil")
	}
}

func TestFilterPlayersAsync(t *testing.T) {
	filter := CreateAsyncPlayerFilter(2, 10)
	ctx := context.Background()

	testPlayers := []Player{
		{Name: "Player1", Position: "ST", ShortPositions: []string{"ST"}, Age: "25", Overall: 75},
		{Name: "Player2", Position: "GK", ShortPositions: []string{"GK"}, Age: "30", Overall: 80},
		{Name: "Player3", Position: "ST", ShortPositions: []string{"ST"}, Age: "22", Overall: 70},
		{Name: "Player4", Position: "CB", ShortPositions: []string{"CB"}, Age: "28", Overall: 78},
	}

	// Filter for strikers only
	filterCriteria := PlayerFilter{
		Position: "ST",
		MinAge:   20,
		MaxAge:   30,
	}

	results := filter.FilterPlayersAsync(ctx, testPlayers, filterCriteria)

	// Should return only the strikers
	expectedCount := 2
	if len(results) != expectedCount {
		t.Errorf("FilterPlayersAsync returned %d players; want %d", len(results), expectedCount)
	}

	// Verify all results are strikers
	for _, player := range results {
		if player.Position != "ST" {
			t.Errorf("Filtered player has position %s; want ST", player.Position)
		}
	}
}

func TestFilterPlayersAsyncEmptyResult(t *testing.T) {
	filter := CreateAsyncPlayerFilter(2, 10)
	ctx := context.Background()

	testPlayers := []Player{
		{Name: "Player1", Position: "ST", ShortPositions: []string{"ST"}, Age: "25", Overall: 75},
		{Name: "Player2", Position: "GK", ShortPositions: []string{"GK"}, Age: "30", Overall: 80},
	}

	// Filter that should match no players
	filterCriteria := PlayerFilter{
		Position: "RW", // No right wingers in test data
	}

	results := filter.FilterPlayersAsync(ctx, testPlayers, filterCriteria)

	if len(results) != 0 {
		t.Errorf("FilterPlayersAsync with non-matching criteria returned %d players; want 0", len(results))
	}
}

func TestCreateConcurrentPercentileProcessor(t *testing.T) {
	numWorkers := 4
	processor := CreateConcurrentPercentileProcessor(numWorkers)

	if processor == nil {
		t.Fatal("CreateConcurrentPercentileProcessor returned nil")
	}
}

func TestProcessPercentilesWithDivisionFilterAsync(t *testing.T) {
	processor := CreateConcurrentPercentileProcessor(2)
	ctx := context.Background()

	testPlayers := []Player{
		{Name: "Player1", Division: "Premier League", Overall: 75, PAC: 70},
		{Name: "Player2", Division: "Premier League", Overall: 80, PAC: 75},
		{Name: "Player3", Division: "Championship", Overall: 70, PAC: 65},
		{Name: "Player4", Division: "Premier League", Overall: 85, PAC: 80},
	}

	// Test with different division filters
	divisionFilters := []DivisionFilter{
		DivisionFilterAll,
		DivisionFilterSame,
		DivisionFilterTop5,
	}

	for _, filter := range divisionFilters {
		t.Run(fmt.Sprintf("filter_%d", int(filter)), func(t *testing.T) {
			results := processor.ProcessPercentilesWithDivisionFilterAsync(
				ctx, testPlayers, filter, "Premier League")

			if results == nil {
				t.Errorf("ProcessPercentilesWithDivisionFilterAsync with filter %v returned nil", filter)
			}

			// Should return filtered players
			if len(results) == 0 {
				t.Errorf("ProcessPercentilesWithDivisionFilterAsync with filter %v returned no players", filter)
			}
		})
	}
}

// Test concurrent access and race conditions
func TestConcurrentProcessing(t *testing.T) {
	processor := CreateConcurrentLeagueProcessor(4)
	filter := CreateAsyncPlayerFilter(4, 10)
	ctx := context.Background()

	// Create test players with enough players per team
	testPlayers := make([]Player, 0)

	// Team1 players (League1)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:           fmt.Sprintf("Team1 Player %d", i),
			Division:       "League1",
			Club:           "Team1",
			Position:       "ST",
			ShortPositions: []string{"ST"},
			Overall:        75 + i,
		})
	}

	// Team2 players (League1)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:           fmt.Sprintf("Team2 Player %d", i),
			Division:       "League1",
			Club:           "Team2",
			Position:       "GK",
			ShortPositions: []string{"GK"},
			Overall:        80 + i,
		})
	}

	// Team3 players (League2)
	for i := 1; i <= 11; i++ {
		testPlayers = append(testPlayers, Player{
			Name:           fmt.Sprintf("Team3 Player %d", i),
			Division:       "League2",
			Club:           "Team3",
			Position:       "ST",
			ShortPositions: []string{"ST"},
			Overall:        85 + i,
		})
	}

	// Run multiple operations concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// Process leagues
			leagues := processor.ProcessLeaguesAsync(ctx, testPlayers)
			if len(leagues) == 0 {
				t.Error("Concurrent ProcessLeaguesAsync returned no results")
				return
			}

			// Filter players
			filterCriteria := PlayerFilter{Position: "ST"}
			filtered := filter.FilterPlayersAsync(ctx, testPlayers, filterCriteria)
			if len(filtered) == 0 {
				t.Error("Concurrent FilterPlayersAsync returned no results")
				return
			}
		}()
	}

	// Wait for all goroutines to complete
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-timeoutCtx.Done():
			t.Fatal("Concurrent processing test timed out")
		}
	}
}

// Benchmark tests for performance validation
func BenchmarkProcessPlayersAsync(b *testing.B) {
	config := DefaultProcessingConfig()

	// Create a reasonably sized dataset for benchmarking
	players := make([]Player, 100) // Smaller dataset for benchmarks
	for i := range players {
		players[i] = Player{
			Name:    fmt.Sprintf("Player%d", i),
			Overall: 70 + (i % 30),                // Vary between 70-99
			Age:     fmt.Sprintf("%d", 20+(i%15)), // Vary between 20-34
		}
	}

	// Helper to drain a channel
	drain := func(ch <-chan Player) {
		for v := range ch {
			_ = v
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor := CreatePlayerProcessor(config)
		resultCh := processor.ProcessPlayersAsync(players)

		drain(resultCh)
		processor.Shutdown()
	}
}

func BenchmarkFilterPlayersAsync(b *testing.B) {
	filter := CreateAsyncPlayerFilter(4, 25)
	ctx := context.Background()

	// Create test data
	players := make([]Player, 100) // Smaller dataset for benchmarks
	positions := []string{"ST", "GK", "CB", "LB", "RB", "CM", "LW", "RW"}
	for i := range players {
		players[i] = Player{
			Name:     fmt.Sprintf("Player%d", i),
			Position: positions[i%len(positions)],
			Overall:  70 + (i % 30),
			Age:      fmt.Sprintf("%d", 20+(i%15)),
		}
	}

	filterCriteria := PlayerFilter{
		Position: "ST",
		MinAge:   20,
		MaxAge:   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterPlayersAsync(ctx, players, filterCriteria)
	}
}
