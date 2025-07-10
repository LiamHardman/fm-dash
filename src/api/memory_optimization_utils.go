package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"runtime"
	"time"
	"unsafe"

	"runtime/debug"

	"go.opentelemetry.io/otel/attribute"
)

// MemoryOptimizationConfig controls which optimizations to use
type MemoryOptimizationConfig struct {
	UseStringInterning  bool // Enable string interning
	UseOptimizedStructs bool // Use OptimizedPlayer struct
	UseCopyOnWrite      bool // Use copy-on-write for operations
	UseObjectPooling    bool // Use object pools
	MonitorMemoryUsage  bool // Track memory usage metrics
}

// Global configuration for memory optimizations
var memOptConfig MemoryOptimizationConfig

// DefaultMemoryOptimizationConfig returns recommended settings
func DefaultMemoryOptimizationConfig() MemoryOptimizationConfig {
	return MemoryOptimizationConfig{
		UseStringInterning:  true,
		UseOptimizedStructs: true,  // ENABLED - provides significant memory savings (~66% reduction)
		UseCopyOnWrite:      false, // Keep disabled - race conditions and overhead
		UseObjectPooling:    true,  // ENABLED - reduces GC pressure with proper lifecycle management
		MonitorMemoryUsage:  true,  // ENABLED - lightweight monitoring for memory pressure detection
	}
}

// MemoryStats tracks memory usage statistics
type MemoryStats struct {
	TotalAllocMB      float64
	SysMemoryMB       float64
	NumGC             uint32
	GCPauseMS         float64
	PlayerMemoryMB    float64
	StringInterningMB float64
	LastUpdated       time.Time
}

var (
	memoryOptConfig = DefaultMemoryOptimizationConfig()
)

// SetMemoryOptimizationConfig updates the global configuration
func SetMemoryOptimizationConfig(config MemoryOptimizationConfig) {
	memoryOptConfig = config
	log.Printf("Memory optimization config updated: %+v", config)
}

// GetCurrentMemoryStats returns current memory usage
func GetCurrentMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := MemoryStats{
		TotalAllocMB: float64(m.Alloc) / 1024 / 1024,
		SysMemoryMB:  float64(m.Sys) / 1024 / 1024,
		NumGC:        m.NumGC,
		LastUpdated:  time.Now(),
	}

	// Calculate GC pause time
	if m.NumGC > 0 {
		stats.GCPauseMS = float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000
	}

	return stats
}

// OptimizePlayerData applies all enabled optimizations to player data
func OptimizePlayerData(ctx context.Context, players []Player) ([]Player, error) {
	ctx, span := StartSpan(ctx, "memory.optimize_player_data")
	defer span.End()

	startTime := time.Now()
	originalMemory := GetCurrentMemoryStats()

	SetSpanAttributes(ctx,
		attribute.Int("players.count", len(players)),
		attribute.Bool("string_interning.enabled", memoryOptConfig.UseStringInterning),
		attribute.Bool("copy_on_write.enabled", memoryOptConfig.UseCopyOnWrite),
		attribute.Bool("object_pooling.enabled", memoryOptConfig.UseObjectPooling),
	)

	// Apply string interning if enabled
	if memoryOptConfig.UseStringInterning {
		applyStringInterning(ctx, players)
	}

	// Use copy-on-write for operations if enabled
	var result []Player
	if memoryOptConfig.UseCopyOnWrite {
		// Use optimized deep copy
		result = OptimizedDeepCopyPlayers(players)
	} else {
		// Fall back to original deep copy
		result = deepCopyPlayers(players)
	}

	// Record memory optimization metrics
	if memoryOptConfig.MonitorMemoryUsage {
		finalMemory := GetCurrentMemoryStats()
		recordMemoryOptimizationMetrics(ctx, originalMemory, finalMemory, len(players), time.Since(startTime))
	}

	return result, nil
}

// applyStringInterning applies string interning to all players
func applyStringInterning(ctx context.Context, players []Player) {
	ctx, span := StartSpan(ctx, "memory.string_interning")
	defer span.End()

	for i := range players {
		SafeOptimizePlayerStrings(&players[i])
	}

	// Record string interning statistics
	stats := GetSafeStringInterningStats()
	SetSpanAttributes(ctx,
		attribute.Int64("string_interning.clubs.unique", stats["clubs"]["unique_strings"]),
		attribute.Int64("string_interning.positions.unique", stats["positions"]["unique_strings"]),
		attribute.Int64("string_interning.nationalities.unique", stats["nationalities"]["unique_strings"]),
		attribute.Int64("string_interning.total_memory_saved", getTotalMemorySaved(stats)),
	)
}

// getTotalMemorySaved calculates total memory saved across all interning pools
func getTotalMemorySaved(stats map[string]map[string]int64) int64 {
	total := int64(0)
	for _, poolStats := range stats {
		total += poolStats["memory_saved"]
	}
	return total
}

// recordMemoryOptimizationMetrics records detailed memory optimization metrics
func recordMemoryOptimizationMetrics(ctx context.Context, before, after MemoryStats, playerCount int, duration time.Duration) {
	memoryReduction := before.TotalAllocMB - after.TotalAllocMB

	SetSpanAttributes(ctx,
		attribute.Float64("memory.before_mb", before.TotalAllocMB),
		attribute.Float64("memory.after_mb", after.TotalAllocMB),
		attribute.Float64("memory.reduction_mb", memoryReduction),
		attribute.Float64("memory.optimization_duration_ms", float64(duration.Milliseconds())),
		attribute.Int("memory.player_count", playerCount),
	)

	// Record as business metric
	RecordBusinessOperation(ctx, "memory_optimization", true, map[string]interface{}{
		"memory_reduction_mb":   memoryReduction,
		"player_count":          playerCount,
		"optimization_duration": duration.Milliseconds(),
		"memory_reduction_pct":  (memoryReduction / before.TotalAllocMB) * 100,
	})
}

// EstimatePlayerMemoryUsage estimates memory usage for a slice of players
func EstimatePlayerMemoryUsage(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}

	// Sample a few players for estimation
	sampleSize := 10
	if len(players) < sampleSize {
		sampleSize = len(players)
	}

	totalSize := 0
	for i := 0; i < sampleSize; i++ {
		totalSize += estimateSinglePlayerMemory(&players[i])
	}

	averageSize := float64(totalSize) / float64(sampleSize)
	totalEstimate := averageSize * float64(len(players))

	return totalEstimate / 1024 / 1024 // Convert to MB
}

// estimateSinglePlayerMemory estimates memory usage for a single player
func estimateSinglePlayerMemory(player *Player) int {
	size := int(unsafe.Sizeof(*player))

	// Add string lengths
	size += len(player.UID) + len(player.Name) + len(player.Position) + len(player.Age) +
		len(player.Club) + len(player.Division) + len(player.TransferValue) + len(player.Wage) +
		len(player.Personality) + len(player.MediaHandling) + len(player.Nationality) +
		len(player.NationalityISO) + len(player.NationalityFIFACode) + len(player.BestRoleOverall)

	// Add map overhead and contents
	size += len(player.Attributes) * 32 // Rough estimate for map entries
	size += len(player.NumericAttributes) * 24
	size += len(player.PerformanceStatsNumeric) * 24

	// Add nested map overhead
	for _, innerMap := range player.PerformancePercentiles {
		size += len(innerMap) * 24
	}
	size += len(player.PerformancePercentiles) * 32

	// Add slice overhead
	for _, pos := range player.ParsedPositions {
		size += len(pos)
	}
	for _, pos := range player.ShortPositions {
		size += len(pos)
	}
	for _, group := range player.PositionGroups {
		size += len(group)
	}

	size += len(player.RoleSpecificOveralls) * int(unsafe.Sizeof(RoleOverallScore{}))

	return size
}

// MemoryOptimizationReport generates a comprehensive memory optimization report
type MemoryOptimizationReport struct {
	PlayerCount                int                         `json:"player_count"`
	EstimatedOriginalMemoryMB  float64                     `json:"estimated_original_memory_mb"`
	EstimatedOptimizedMemoryMB float64                     `json:"estimated_optimized_memory_mb"`
	EstimatedSavingsPercent    float64                     `json:"estimated_savings_percent"`
	StringInterningStats       map[string]map[string]int64 `json:"string_interning_stats"`
	CurrentMemoryStats         MemoryStats                 `json:"current_memory_stats"`
	OptimizationsEnabled       MemoryOptimizationConfig    `json:"optimizations_enabled"`
	GeneratedAt                time.Time                   `json:"generated_at"`
}

// GenerateMemoryOptimizationReport creates a detailed optimization report
func GenerateMemoryOptimizationReport(players []Player) MemoryOptimizationReport {
	originalMB, optimizedMB, savingsPercent := EstimateMemorySavings(len(players))

	return MemoryOptimizationReport{
		PlayerCount:                len(players),
		EstimatedOriginalMemoryMB:  originalMB,
		EstimatedOptimizedMemoryMB: optimizedMB,
		EstimatedSavingsPercent:    savingsPercent,
		StringInterningStats:       GetSafeStringInterningStats(),
		CurrentMemoryStats:         GetCurrentMemoryStats(),
		OptimizationsEnabled:       memoryOptConfig,
		GeneratedAt:                time.Now(),
	}
}

// GetMemoryOptimizationHandler returns an HTTP handler for memory optimization reports
func GetMemoryOptimizationHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get all datasets for analysis
		totalPlayers := 0

		// Use the storage interface to get all datasets
		if storage != nil {
			if datasetIDs, err := storage.List(); err == nil {
				for _, datasetID := range datasetIDs {
					if players, _, found := GetPlayerData(datasetID); found {
						totalPlayers += len(players)
					}
				}
			}
		}

		// Generate mock player data for estimation if no real data
		var samplePlayers []Player
		if totalPlayers == 0 {
			// Create a representative sample for estimation
			samplePlayers = []Player{{
				UID: "sample", Name: "Sample Player", Position: "Centre Back",
				Age: "25", Club: "Sample FC", Division: "Premier League",
				Nationality: "England", Attributes: make(map[string]string, 100),
				NumericAttributes: make(map[string]int, 100),
			}}
			totalPlayers = 10000 // Estimate for 10k players
		}

		report := GenerateMemoryOptimizationReport(samplePlayers)
		report.PlayerCount = totalPlayers // Override with actual count

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(report); err != nil {
			http.Error(w, "Failed to encode report", http.StatusInternalServerError)
		}
	}
}

// InitializeMemoryOptimizations sets up all memory optimizations
func InitializeMemoryOptimizations() {
	log.Println("Initializing enhanced memory optimizations...")

	// Initialize global configuration with optimized settings
	memOptConfig = DefaultMemoryOptimizationConfig()

	// Set default configuration
	SetMemoryOptimizationConfig(DefaultMemoryOptimizationConfig())

	// Start background monitoring for memory pressure detection
	if memOptConfig.MonitorMemoryUsage {
		go startMemoryMonitoring()
		log.Println("Memory pressure monitoring started")
	}

	log.Printf("Enhanced memory optimizations initialized with config: %+v", memOptConfig)
}

// startMemoryMonitoring runs background memory monitoring
func startMemoryMonitoring() {
	// More frequent monitoring for better memory management
	ticker := time.NewTicker(30 * time.Second) // Changed back to 30s for better responsiveness
	defer ticker.Stop()

	for range ticker.C {
		stats := GetCurrentMemoryStats()

		// Log memory stats periodically with INFO level
		if stats.TotalAllocMB > 64 { // Reduced from 128MB to 64MB for much earlier detection
			slog.Info("Memory stats",
				"allocated_mb", stats.TotalAllocMB,
				"system_mb", stats.SysMemoryMB,
				"gc_count", stats.NumGC,
				"gc_pause_ms", stats.GCPauseMS)
		}

		// Enhanced memory pressure levels with automatic responses
		switch {
		case stats.TotalAllocMB > 4096: // Critical level - increased to 4096MB
			log.Printf("CRITICAL: Memory usage at %.1fMB - triggering aggressive cleanup", stats.TotalAllocMB)

			// Force garbage collection multiple times for more aggressive cleanup
			runtime.GC()
			runtime.GC()
			runtime.GC() // Triple GC for critical situations

			// Emergency cache cleanup - clear 90% of cache
			if memCache != nil {
				memCache.mutex.Lock()
				targetSize := len(memCache.items) / 10 // Keep only 10%
				cleared := 0
				for key := range memCache.items {
					if cleared >= len(memCache.items)-targetSize {
						break
					}
					memCache.removeLRU(key)
					cleared++
				}
				memCache.mutex.Unlock()
				log.Printf("Emergency cache cleanup: removed %d items", cleared)
			}

			// Adjust GOGC for more aggressive collection
			debug.SetGCPercent(25) // Much more aggressive GC
		case stats.TotalAllocMB > 1024: // Warning level - increased to 1024MB
			log.Printf("WARNING: High memory usage at %.1fMB - triggering cache cleanup", stats.TotalAllocMB)

			// Trigger cache cleanup - clear 50% of cache
			if memCache != nil {
				memCache.mutex.Lock()
				targetSize := len(memCache.items) / 2 // Keep 50%
				cleared := 0
				for key := range memCache.items {
					if cleared >= len(memCache.items)-targetSize {
						break
					}
					memCache.removeLRU(key)
					cleared++
				}
				memCache.mutex.Unlock()
			}

			// Double GC for warning level
			runtime.GC()
			runtime.GC()
		case stats.TotalAllocMB < 1024:
			// Memory usage is low, relax GC pressure
			debug.SetGCPercent(100) // Default GC behavior
		}
	}
}
