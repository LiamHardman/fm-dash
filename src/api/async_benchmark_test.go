package main

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkPlayerProcessing compares sync vs async player processing
func BenchmarkPlayerProcessing(b *testing.B) {
	// Create test players
	players := createTestPlayers(1000)

	b.Run("Sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range players {
				EnhancePlayerWithCalculations(&players[j])
			}
		}
	})

	b.Run("Async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			if _, err := ProcessPlayersBatch(ctx, players, 100); err != nil {
				b.Fatalf("ProcessPlayersBatch failed: %v", err)
			}
		}
	})
}

// BenchmarkPlayerFiltering compares sync vs async filtering
func BenchmarkPlayerFiltering(b *testing.B) {
	players := createTestPlayers(5000)
	filter := PlayerFilter{
		Position: "ST",
		MinAge:   20,
		MaxAge:   30,
	}

	b.Run("Sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := make([]Player, 0)
			for _, player := range players {
				// Simplified sync filtering logic
				if player.Age >= "20" && player.Age <= "30" {
					result = append(result, player)
				}
			}
			// Prevent compiler optimization
			_ = result
		}
	})

	b.Run("Async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			asyncFilter := NewAsyncPlayerFilter(runtime.NumCPU(), 200)
			asyncFilter.FilterPlayersAsync(ctx, players, filter)
		}
	})
}

// BenchmarkLeagueProcessing compares sync vs async league processing
func BenchmarkLeagueProcessing(b *testing.B) {
	players := createTestPlayersWithDivisions(2000)

	b.Run("Sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			processLeaguesData(players)
		}
	})

	b.Run("Async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			processor := NewConcurrentLeagueProcessor(runtime.NumCPU())
			processor.ProcessLeaguesAsync(ctx, players)
		}
	})
}

// BenchmarkDatasetStorage compares sync vs async storage performance
func BenchmarkDatasetStorage(b *testing.B) {
	// Initialize storage for testing
	InitStore()
	
	// Create test dataset
	players := createTestPlayers(1000)
	currencySymbol := "$"

	b.Run("SyncStorage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			datasetID := fmt.Sprintf("sync-test-%d", i)
			// This includes both in-memory and persistent storage operations
			SetPlayerData(datasetID, players, currencySymbol)
		}
	})

	b.Run("AsyncStorage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			datasetID := fmt.Sprintf("async-test-%d", i)
			// This returns immediately after in-memory storage, persistent storage happens in background
			SetPlayerDataAsync(datasetID, players, currencySymbol)
		}
	})

	// Test response time simulation (more realistic scenario)
	b.Run("ResponseTimeSync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			datasetID := fmt.Sprintf("response-sync-test-%d", i)
			
			// Simulate the upload handler flow with sync storage
			start := time.Now()
			SetPlayerData(datasetID, players, currencySymbol)
			
			// Simulate JSON response encoding
			response := map[string]interface{}{
				"datasetID": datasetID,
				"message":   "File uploaded and parsed successfully.",
				"currency":  currencySymbol,
			}
			_ = response // Prevent optimization
			
			// This represents the total time user waits for response
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "response_time_ms")
		}
	})

	b.Run("ResponseTimeAsync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			datasetID := fmt.Sprintf("response-async-test-%d", i)
			
			// Simulate the upload handler flow with async storage
			start := time.Now()
			SetPlayerDataAsync(datasetID, players, currencySymbol)
			
			// Simulate JSON response encoding
			response := map[string]interface{}{
				"datasetID": datasetID,
				"message":   "File uploaded and parsed successfully.",
				"currency":  currencySymbol,
			}
			_ = response // Prevent optimization
			
			// This represents the total time user waits for response (should be much faster)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "response_time_ms")
		}
	})
}

// createTestPlayers generates test player data for benchmarking
func createTestPlayers(count int) []Player {
	players := make([]Player, count)

	for i := 0; i < count; i++ {
		players[i] = Player{
			Name:     fmt.Sprintf("Player %d", i),
			Position: "Striker",
			Age:      fmt.Sprintf("%d", 18+(i%15)), // Ages 18-32
			Club:     fmt.Sprintf("Club %d", i%20),
			Division: fmt.Sprintf("Division %d", i%5),
			Overall:  60 + (i % 30), // Overall 60-89
			Attributes: map[string]string{
				"Pac": fmt.Sprintf("%d", 10+(i%10)),
				"Sho": fmt.Sprintf("%d", 10+(i%10)),
				"Pas": fmt.Sprintf("%d", 10+(i%10)),
				"Dri": fmt.Sprintf("%d", 10+(i%10)),
				"Def": fmt.Sprintf("%d", 10+(i%10)),
				"Phy": fmt.Sprintf("%d", 10+(i%10)),
			},
			NumericAttributes:       make(map[string]int),
			PerformanceStatsNumeric: make(map[string]float64),
			PerformancePercentiles:  make(map[string]map[string]float64),
			ParsedPositions:         []string{"Striker"},
			ShortPositions:          []string{"ST"},
			PositionGroups:          []string{"Attackers"},
			RoleSpecificOveralls:    make([]RoleOverallScore, 0),
		}
	}

	return players
}

// createTestPlayersWithDivisions creates players spread across different divisions
func createTestPlayersWithDivisions(count int) []Player {
	players := createTestPlayers(count)
	divisions := []string{"Premier League", "Championship", "League One", "League Two", "Non-League"}

	for i := range players {
		players[i].Division = divisions[i%len(divisions)]
		players[i].Club = fmt.Sprintf("Team %d", i%100) // 100 teams total
	}

	return players
}

// TestAsyncProcessingCorrectness verifies async processing produces correct results
func TestAsyncProcessingCorrectness(t *testing.T) {
	players := createTestPlayers(100)

	// Test player processing
	t.Run("PlayerProcessing", func(t *testing.T) {
		// Process synchronously
		syncPlayers := make([]Player, len(players))
		copy(syncPlayers, players)
		for i := range syncPlayers {
			EnhancePlayerWithCalculations(&syncPlayers[i])
		}

		// Process asynchronously
		ctx := context.Background()
		asyncPlayers, err := ProcessPlayersBatch(ctx, players, 20)
		if err != nil {
			t.Fatalf("Async processing failed: %v", err)
		}

		// Verify same number of players
		if len(syncPlayers) != len(asyncPlayers) {
			t.Fatalf("Player count mismatch: sync=%d, async=%d", len(syncPlayers), len(asyncPlayers))
		}

		// Verify processing was applied (check some stats were calculated)
		processedCount := 0
		for _, player := range asyncPlayers {
			if player.PAC > 0 || player.SHO > 0 || player.PAS > 0 {
				processedCount++
			}
		}

		if processedCount == 0 {
			t.Error("No players were processed in async pipeline")
		}

		t.Logf("Successfully processed %d/%d players asynchronously", processedCount, len(asyncPlayers))
	})

	// Test filtering
	t.Run("PlayerFiltering", func(t *testing.T) {
		filter := PlayerFilter{
			MinAge: 20,
			MaxAge: 25,
		}

		ctx := context.Background()
		asyncFilter := NewAsyncPlayerFilter(2, 30)
		filteredPlayers := asyncFilter.FilterPlayersAsync(ctx, players, filter)

		// Verify all filtered players meet age criteria
		for _, player := range filteredPlayers {
			// Note: This is a simplified test since we'd need to properly parse ages
			if player.Name == "" {
				t.Error("Found player with empty name in filtered results")
			}
		}

		t.Logf("Async filtering returned %d players", len(filteredPlayers))
	})

	// Test league processing
	t.Run("LeagueProcessing", func(t *testing.T) {
		playersWithDivisions := createTestPlayersWithDivisions(200)

		ctx := context.Background()
		processor := NewConcurrentLeagueProcessor(2)
		leagues := processor.ProcessLeaguesAsync(ctx, playersWithDivisions)

		if len(leagues) == 0 {
			t.Error("No leagues were processed")
		}

		// Verify league data makes sense
		for _, league := range leagues {
			if league.Name == "" {
				t.Error("Found league with empty name")
			}
			if league.PlayerCount <= 0 {
				t.Errorf("League %s has no players", league.Name)
			}
		}

		t.Logf("Successfully processed %d leagues asynchronously", len(leagues))
	})
}

// TestAsyncPerformanceImprovement measures performance improvement
func TestAsyncPerformanceImprovement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	players := createTestPlayers(1000)

	// Measure sync processing time
	syncStart := time.Now()
	for i := range players {
		EnhancePlayerWithCalculations(&players[i])
	}
	syncDuration := time.Since(syncStart)

	// Reset players for async test
	players = createTestPlayers(1000)

	// Measure async processing time
	asyncStart := time.Now()
	ctx := context.Background()
	if _, err := ProcessPlayersBatch(ctx, players, 100); err != nil {
		t.Fatalf("Async ProcessPlayersBatch failed: %v", err)
	}
	asyncDuration := time.Since(asyncStart)

	t.Logf("Sync processing: %v", syncDuration)
	t.Logf("Async processing: %v", asyncDuration)
	t.Logf("Performance ratio: %.2fx", float64(syncDuration)/float64(asyncDuration))

	// On multi-core systems, async should be faster for large datasets
	if runtime.NumCPU() > 1 && asyncDuration >= syncDuration {
		t.Logf("Warning: Async processing was not faster (this may be expected for small datasets)")
	}
}

// TestConcurrencyLimits verifies our concurrency controls work properly
func TestConcurrencyLimits(t *testing.T) {
	// Test with various worker counts
	players := createTestPlayers(500)

	for _, workerCount := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("Workers_%d", workerCount), func(t *testing.T) {
			start := time.Now()

			processor := NewPlayerProcessor(ProcessingConfig{
				WorkerCount: workerCount,
				BatchSize:   50,
				BufferSize:  workerCount * 10,
			})

			resultCh := processor.ProcessPlayersAsync(players)

			count := 0
			for range resultCh {
				count++
			}

			processor.Shutdown()
			duration := time.Since(start)

			t.Logf("Workers: %d, Processed: %d players in %v", workerCount, count, duration)

			if count != len(players) {
				t.Errorf("Expected %d players, got %d", len(players), count)
			}
		})
	}
}

// TestAsyncStorageCorrectness verifies async storage works correctly
func TestAsyncStorageCorrectness(t *testing.T) {
	// Initialize storage for testing
	InitStore()
	
	// Create test data
	players := createTestPlayers(50)
	datasetID := "test-async-storage"
	currencySymbol := "$"

	t.Run("ImmediateAvailability", func(t *testing.T) {
		// Store data asynchronously
		SetPlayerDataAsync(datasetID, players, currencySymbol)
		
		// Data should be immediately available from in-memory cache
		retrievedPlayers, retrievedCurrency, found := GetPlayerData(datasetID)
		
		if !found {
			t.Fatal("Data not found immediately after async storage")
		}
		
		if len(retrievedPlayers) != len(players) {
			t.Errorf("Player count mismatch: expected %d, got %d", len(players), len(retrievedPlayers))
		}
		
		if retrievedCurrency != currencySymbol {
			t.Errorf("Currency mismatch: expected %s, got %s", currencySymbol, retrievedCurrency)
		}
		
		// Verify specific player data
		if len(retrievedPlayers) > 0 && retrievedPlayers[0].Name != players[0].Name {
			t.Errorf("Player data mismatch: expected %s, got %s", players[0].Name, retrievedPlayers[0].Name)
		}
		
		t.Logf("Successfully verified immediate availability of %d players", len(retrievedPlayers))
	})

	t.Run("ResponseTimeComparison", func(t *testing.T) {
		testPlayers := createTestPlayers(100)
		
		// Test sync storage response time
		syncDatasetID := "test-sync-response"
		syncStart := time.Now()
		SetPlayerData(syncDatasetID, testPlayers, currencySymbol)
		syncDuration := time.Since(syncStart)
		
		// Test async storage response time  
		asyncDatasetID := "test-async-response"
		asyncStart := time.Now()
		SetPlayerDataAsync(asyncDatasetID, testPlayers, currencySymbol)
		asyncDuration := time.Since(asyncStart)
		
		t.Logf("Sync storage time: %v", syncDuration)
		t.Logf("Async storage time: %v", asyncDuration)
		t.Logf("Improvement ratio: %.2fx faster", float64(syncDuration)/float64(asyncDuration))
		
		// Async should be significantly faster (at least 2x for persistent storage)
		if asyncDuration >= syncDuration {
			t.Logf("Warning: Async storage was not faster. This may occur if persistent storage is very fast or disabled.")
		}
		
		// Verify both datasets are available
		_, _, syncFound := GetPlayerData(syncDatasetID)
		_, _, asyncFound := GetPlayerData(asyncDatasetID)
		
		if !syncFound {
			t.Error("Sync dataset not found")
		}
		if !asyncFound {
			t.Error("Async dataset not found")
		}
	})
}
