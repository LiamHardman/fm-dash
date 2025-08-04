package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// MemoryTestResult represents the result of a memory optimization test
type MemoryTestResult struct {
	TestName              string                 `json:"test_name"`
	PlayerCount           int                    `json:"player_count"`
	BeforeOptimization    MemoryStats            `json:"before_optimization"`
	AfterOptimization     MemoryStats            `json:"after_optimization"`
	RegularPlayerSizeMB   float64                `json:"regular_player_size_mb"`
	OptimizedPlayerSizeMB float64                `json:"optimized_player_size_mb"`
	MemorySavedMB         float64                `json:"memory_saved_mb"`
	ReductionPercent      float64                `json:"reduction_percent"`
	PerPlayerSavings      float64                `json:"per_player_savings_bytes"`
	OptimizationStats     map[string]interface{} `json:"optimization_stats"`
	TestDurationMS        int64                  `json:"test_duration_ms"`
}

// memoryTestHandler runs memory optimization tests
func memoryTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.memory_test")
	defer span.End()

	startTime := time.Now()

	// Get dataset ID from query parameter
	datasetID := r.URL.Query().Get("dataset_id")
	if datasetID == "" {
		http.Error(w, "dataset_id parameter required", http.StatusBadRequest)
		return
	}

	// Retrieve dataset
	players, _, found := GetPlayerData(datasetID)
	if !found {
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
	}

	logInfo(ctx, "Starting memory optimization test",
		"dataset_id", datasetID,
		"player_count", len(players))

	// Take before snapshot
	runtime.GC() // Clean up before measuring
	beforeStats := GetCurrentMemoryStats()

	// Test regular player memory
	regularSize := GlobalAdvancedOptimizer.estimatePlayerSliceMemory(players)

	// Apply advanced optimization
	optimizedPlayers, savedBytes := GlobalAdvancedOptimizer.OptimizePlayersAdvanced(ctx, players)

	// Take after snapshot
	runtime.GC() // Clean up after optimization
	afterStats := GetCurrentMemoryStats()

	// Calculate optimized size
	optimizedSize := GlobalAdvancedOptimizer.estimateOptimizedSliceMemory(optimizedPlayers)

	// Get optimization stats
	optStats := GlobalAdvancedOptimizer.GetOptimizationStats()

	result := MemoryTestResult{
		TestName:              fmt.Sprintf("Dataset %s Optimization Test", datasetID),
		PlayerCount:           len(players),
		BeforeOptimization:    beforeStats,
		AfterOptimization:     afterStats,
		RegularPlayerSizeMB:   float64(regularSize) / 1024 / 1024,
		OptimizedPlayerSizeMB: float64(optimizedSize) / 1024 / 1024,
		MemorySavedMB:         float64(savedBytes) / 1024 / 1024,
		ReductionPercent:      float64(savedBytes) / float64(regularSize) * 100,
		PerPlayerSavings:      float64(savedBytes) / float64(len(players)),
		OptimizationStats:     optStats,
		TestDurationMS:        time.Since(startTime).Milliseconds(),
	}

	SetSpanAttributes(ctx,
		attribute.Int("test.player_count", len(players)),
		attribute.Float64("test.memory_saved_mb", result.MemorySavedMB),
		attribute.Float64("test.reduction_percent", result.ReductionPercent),
		attribute.Float64("test.per_player_savings", result.PerPlayerSavings),
	)

	logInfo(ctx, "Memory optimization test completed",
		"player_count", len(players),
		"memory_saved_mb", result.MemorySavedMB,
		"reduction_percent", result.ReductionPercent,
		"per_player_savings_bytes", result.PerPlayerSavings,
		"test_duration_ms", result.TestDurationMS)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logError(ctx, "Failed to encode test result", "error", err)
		http.Error(w, "Failed to encode test result", http.StatusInternalServerError)
	}
}

// memoryBenchmarkHandler runs comprehensive memory benchmarks
func memoryBenchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.memory_benchmark")
	defer span.End()

	// Create test data with different sizes
	testSizes := []int{100, 1000, 5000, 10000}
	results := make([]MemoryTestResult, 0, len(testSizes))

	for _, size := range testSizes {
		// Create test players
		testPlayers := createTestPlayersForBenchmark(size)

		startTime := time.Now()

		// Take before snapshot
		runtime.GC()
		beforeStats := GetCurrentMemoryStats()

		// Test optimization
		regularSize := GlobalAdvancedOptimizer.estimatePlayerSliceMemory(testPlayers)
		optimizedPlayers, savedBytes := GlobalAdvancedOptimizer.OptimizePlayersAdvanced(ctx, testPlayers)
		optimizedSize := GlobalAdvancedOptimizer.estimateOptimizedSliceMemory(optimizedPlayers)

		// Take after snapshot
		runtime.GC()
		afterStats := GetCurrentMemoryStats()

		result := MemoryTestResult{
			TestName:              fmt.Sprintf("Benchmark Test %d Players", size),
			PlayerCount:           size,
			BeforeOptimization:    beforeStats,
			AfterOptimization:     afterStats,
			RegularPlayerSizeMB:   float64(regularSize) / 1024 / 1024,
			OptimizedPlayerSizeMB: float64(optimizedSize) / 1024 / 1024,
			MemorySavedMB:         float64(savedBytes) / 1024 / 1024,
			ReductionPercent:      float64(savedBytes) / float64(regularSize) * 100,
			PerPlayerSavings:      float64(savedBytes) / float64(size),
			TestDurationMS:        time.Since(startTime).Milliseconds(),
		}

		results = append(results, result)

		logInfo(ctx, "Benchmark test completed",
			"test_size", size,
			"memory_saved_mb", result.MemorySavedMB,
			"reduction_percent", result.ReductionPercent,
			"per_player_savings_bytes", result.PerPlayerSavings)
	}

	// Calculate summary stats
	summary := map[string]interface{}{
		"total_tests":            len(results),
		"optimization_stats":     GlobalAdvancedOptimizer.GetOptimizationStats(),
		"benchmark_results":      results,
		"avg_reduction_percent":  calculateAverageReduction(results),
		"avg_per_player_savings": calculateAveragePerPlayerSavings(results),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		logError(ctx, "Failed to encode benchmark results", "error", err)
		http.Error(w, "Failed to encode benchmark results", http.StatusInternalServerError)
	}
}

// createTestPlayersForBenchmark creates test players for benchmarking
func createTestPlayersForBenchmark(count int) []Player {
	players := make([]Player, count)

	clubs := []string{"Manchester United", "Barcelona", "Real Madrid", "Liverpool", "Arsenal", "Chelsea"}
	positions := []string{"ST", "CM", "CB", "GK", "LW", "RW", "CDM", "CAM"}
	nationalities := []string{"England", "Spain", "Brazil", "Argentina", "France", "Germany"}

	for i := 0; i < count; i++ {
		player := &players[i]
		player.UID = int64(1000000 + i)
		player.Name = fmt.Sprintf("Test Player %d", i)
		player.Club = clubs[i%len(clubs)]
		player.Position = positions[i%len(positions)]
		player.Division = "Premier League"
		player.Nationality = nationalities[i%len(nationalities)]
		player.NationalityISO = "ENG"
		player.Age = "25"
		player.Overall = 70 + (i % 30)
		player.TransferValueAmount = int64(1000000 + (i%50)*100000)
		player.WageAmount = int64(50000 + (i%20)*5000)

		// FIFA stats
		player.PAC = 70 + (i % 30)
		player.SHO = 65 + (i % 35)
		player.PAS = 75 + (i % 25)
		player.DRI = 70 + (i % 30)
		player.DEF = 60 + (i % 40)
		player.PHY = 75 + (i % 25)

		// Initialize maps
		player.Attributes = make(map[string]string)
		player.NumericAttributes = make(map[string]int)
		player.PerformanceStatsNumeric = make(map[string]float64)
		player.PerformancePercentiles = make(map[string]map[string]float64)

		// Add some attributes
		attrs := []string{"Cor", "Cro", "Dri", "Fin", "Fir", "Hea", "Pas", "Tck"}
		for j, attr := range attrs {
			val := 10 + ((i + j) % 11) // Values 10-20
			player.Attributes[attr] = fmt.Sprintf("%d", val)
			player.NumericAttributes[attr] = val
		}

		// Positions
		player.ParsedPositions = []string{player.Position}
		player.ShortPositions = []string{player.Position}
		player.PositionGroups = []string{"Attack"}

		// Role overalls
		player.RoleSpecificOveralls = []RoleOverallScore{
			{RoleName: "Advanced Forward", Score: player.Overall},
		}
	}

	return players
}

// Helper functions
func calculateAverageReduction(results []MemoryTestResult) float64 {
	if len(results) == 0 {
		return 0
	}

	total := 0.0
	for _, result := range results {
		total += result.ReductionPercent
	}
	return total / float64(len(results))
}

func calculateAveragePerPlayerSavings(results []MemoryTestResult) float64 {
	if len(results) == 0 {
		return 0
	}

	total := 0.0
	for _, result := range results {
		total += result.PerPlayerSavings
	}
	return total / float64(len(results))
}

// RegisterMemoryTestEndpoints registers memory testing endpoints
func RegisterMemoryTestEndpoints(mux *http.ServeMux) {
	mux.Handle("/api/debug/memory/test", wrapHandler(http.HandlerFunc(memoryTestHandler), "memory-test"))
	mux.Handle("/api/debug/memory/benchmark", wrapHandler(http.HandlerFunc(memoryBenchmarkHandler), "memory-benchmark"))
}
