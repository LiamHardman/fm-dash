package main

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"runtime/debug"
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
func SetMemoryOptimizationConfig(ctx context.Context, config MemoryOptimizationConfig) {
	logDebug(ctx, "Updating memory optimization configuration",
		"string_interning", config.UseStringInterning,
		"optimized_structs", config.UseOptimizedStructs,
		"copy_on_write", config.UseCopyOnWrite,
		"object_pooling", config.UseObjectPooling,
		"monitor_memory", config.MonitorMemoryUsage)

	memoryOptConfig = config

	logDebug(ctx, "Memory optimization configuration updated successfully")
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
		ctx := r.Context()
		start := time.Now()

		logDebug(ctx, "Processing memory optimization report request")

		if r.Method != "GET" {
			logWarn(ctx, "Invalid HTTP method for memory optimization endpoint", "method", r.Method)
			WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
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
			logDebug(ctx, "No real player data found, using sample data for estimation")
			// Create a representative sample for estimation
			samplePlayers = []Player{{
				UID: 123456789, Name: "Sample Player", Position: "Centre Back",
				Age: "25", Club: "Sample FC", Division: "Premier League",
				Nationality: "England", Attributes: make(map[string]string, 100),
				NumericAttributes: make(map[string]int, 100),
			}}
			totalPlayers = 10000 // Estimate for 10k players
		}

		logDebug(ctx, "Generating memory optimization report", "total_players", totalPlayers)
		report := GenerateMemoryOptimizationReport(samplePlayers)
		report.PlayerCount = totalPlayers // Override with actual count

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(report); err != nil {
			logError(ctx, "Failed to encode memory optimization report", "error", err)
			WriteErrorResponse(w, r, "response_encoding_failed", "Failed to encode report", nil, http.StatusInternalServerError)
			return
		}

		logDebug(ctx, "Memory optimization report request completed successfully",
			"total_players", totalPlayers,
			"duration_ms", time.Since(start).Milliseconds())
	}
}

// InitializeMemoryOptimizations sets up all memory optimizations
func InitializeMemoryOptimizations() {
	ctx := context.Background()
	logInfo(ctx, "Initializing memory optimizations")
	start := time.Now()

	// Initialize global configuration with optimized settings
	config := DefaultMemoryOptimizationConfig()
	memOptConfig = config

	// Set default configuration
	SetMemoryOptimizationConfig(ctx, config)

	// Start background monitoring for memory pressure detection
	if memOptConfig.MonitorMemoryUsage {
		go startMemoryMonitoring()
		logDebug(ctx, "Memory pressure monitoring started")
	}

	logDebug(ctx, "Memory optimizations initialized successfully",
		"string_interning", config.UseStringInterning,
		"optimized_structs", config.UseOptimizedStructs,
		"copy_on_write", config.UseCopyOnWrite,
		"object_pooling", config.UseObjectPooling,
		"monitor_memory", config.MonitorMemoryUsage,
		"duration_ms", time.Since(start).Milliseconds())
}

// Global variables to track memory monitoring state
var (
	lastLoggedMemoryMB float64
	lastLoggedGCCount  uint32
)

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// startMemoryMonitoring runs background memory monitoring
func startMemoryMonitoring() {
	// Create background context for monitoring
	ctx := context.Background()
	logDebug(ctx, "Starting memory monitoring background process")

	// More frequent monitoring for better memory management
	ticker := time.NewTicker(30 * time.Second) // Changed back to 30s for better responsiveness
	defer ticker.Stop()

	for range ticker.C {
		stats := GetCurrentMemoryStats()

		// Only log memory stats if there's been significant change or new GC activity
		memoryChanged := abs(stats.TotalAllocMB-lastLoggedMemoryMB) > 10.0 // Log if memory changed by more than 10MB
		gcActivityChanged := stats.NumGC != lastLoggedGCCount

		if stats.TotalAllocMB > 64 && (memoryChanged || gcActivityChanged) {
			logInfo(ctx, "Memory stats",
				"allocated_mb", stats.TotalAllocMB,
				"system_mb", stats.SysMemoryMB,
				"gc_count", stats.NumGC,
				"gc_pause_ms", stats.GCPauseMS)

			// Update last logged values
			lastLoggedMemoryMB = stats.TotalAllocMB
			lastLoggedGCCount = stats.NumGC
		}

		// Enhanced memory pressure levels with automatic responses
		switch {
		case stats.TotalAllocMB > 8192: // Critical level
			logError(ctx, "CRITICAL: Memory usage triggering aggressive cleanup",
				"memory_mb", stats.TotalAllocMB,
				"action", "aggressive_cleanup")

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
				logInfo(ctx, "Emergency cache cleanup completed", "items_removed", cleared)
			}

			// Adjust GOGC for more aggressive collection
			debug.SetGCPercent(25) // Much more aggressive GC
		case stats.TotalAllocMB > 4096: // Warning level
			logWarn(ctx, "High memory usage triggering cache cleanup",
				"memory_mb", stats.TotalAllocMB,
				"action", "cache_cleanup")

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

// LogImmediateMemoryStats logs memory stats immediately (for use after parsing completion)
func LogImmediateMemoryStats(ctx context.Context) {
	start := time.Now()
	logDebug(ctx, "Logging immediate memory stats")

	stats := GetCurrentMemoryStats()

	if stats.TotalAllocMB > 64 {
		logInfo(ctx, "Memory stats",
			"allocated_mb", stats.TotalAllocMB,
			"system_mb", stats.SysMemoryMB,
			"gc_count", stats.NumGC,
			"gc_pause_ms", stats.GCPauseMS)

		// Update last logged values to prevent duplicate logging
		lastLoggedMemoryMB = stats.TotalAllocMB
		lastLoggedGCCount = stats.NumGC
	}

	logDebug(ctx, "Memory stats logging completed", "duration_ms", time.Since(start).Milliseconds())
}
