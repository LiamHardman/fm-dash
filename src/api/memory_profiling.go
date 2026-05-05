package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"sort"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// MemoryProfiler provides memory profiling and analysis capabilities
type MemoryProfiler struct {
	snapshots []MemorySnapshot
	maxSize   int
}

// MemorySnapshot captures a point-in-time memory state
type MemorySnapshot struct {
	Timestamp     time.Time `json:"timestamp"`
	TotalAllocMB  float64   `json:"total_alloc_mb"`
	SysMemoryMB   float64   `json:"sys_memory_mb"`
	HeapAllocMB   float64   `json:"heap_alloc_mb"`
	HeapSysMB     float64   `json:"heap_sys_mb"`
	HeapIdleMB    float64   `json:"heap_idle_mb"`
	HeapInuseMB   float64   `json:"heap_inuse_mb"`
	StackSysMB    float64   `json:"stack_sys_mb"`
	NumGC         uint32    `json:"num_gc"`
	GCPauseMS     float64   `json:"gc_pause_ms"`
	NumGoroutines int       `json:"num_goroutines"`

	// Application-specific metrics
	DatasetCount   int              `json:"dataset_count"`
	CacheItemCount int              `json:"cache_item_count"`
	CacheSizeMB    float64          `json:"cache_size_mb"`
	PoolStats      map[string]int64 `json:"pool_stats"`
}

// MemoryAnalysis provides detailed memory usage analysis
type MemoryAnalysis struct {
	Summary            MemorySnapshot         `json:"summary"`
	TopMemoryConsumers []MemoryConsumer       `json:"top_memory_consumers"`
	Recommendations    []MemoryRecommendation `json:"recommendations"`
	TrendAnalysis      MemoryTrendAnalysis    `json:"trend_analysis"`
	GCEfficiency       GCEfficiencyMetrics    `json:"gc_efficiency"`
}

type MemoryConsumer struct {
	Component   string  `json:"component"`
	EstimatedMB float64 `json:"estimated_mb"`
	Percentage  float64 `json:"percentage"`
	Description string  `json:"description"`
}

type MemoryRecommendation struct {
	Priority     string `json:"priority"` // HIGH, MEDIUM, LOW
	Component    string `json:"component"`
	Issue        string `json:"issue"`
	Suggestion   string `json:"suggestion"`
	ExpectedGain string `json:"expected_gain"`
}

type MemoryTrendAnalysis struct {
	MemoryGrowthRate   float64 `json:"memory_growth_rate_mb_per_hour"`
	GCFrequencyTrend   string  `json:"gc_frequency_trend"`
	PerformanceTrend   string  `json:"performance_trend"`
	PredictedPeakHours int     `json:"predicted_peak_hours"`
}

type GCEfficiencyMetrics struct {
	AverageGCPauseMS    float64 `json:"average_gc_pause_ms"`
	GCFrequencyPerHour  float64 `json:"gc_frequency_per_hour"`
	GCEfficiencyRating  string  `json:"gc_efficiency_rating"`
	MemoryReclamationMB float64 `json:"memory_reclamation_mb"`
}

var globalMemoryProfiler = &MemoryProfiler{
	snapshots: make([]MemorySnapshot, 0),
	maxSize:   1000, // Keep last 1000 snapshots
}

// InitializeMemoryProfiler sets up memory profiling
func InitializeMemoryProfiler() {
	LogInfo("Initializing memory profiler...")

	// Take initial snapshot
	globalMemoryProfiler.TakeSnapshot()

	// Start periodic snapshots every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			globalMemoryProfiler.TakeSnapshot()
		}
	}()

	LogInfo("Memory profiler initialized with 5-minute snapshots")
}

// TakeSnapshot captures current memory state
func (mp *MemoryProfiler) TakeSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:     time.Now(),
		TotalAllocMB:  float64(m.Alloc) / 1024 / 1024,
		SysMemoryMB:   float64(m.Sys) / 1024 / 1024,
		HeapAllocMB:   float64(m.HeapAlloc) / 1024 / 1024,
		HeapSysMB:     float64(m.HeapSys) / 1024 / 1024,
		HeapIdleMB:    float64(m.HeapIdle) / 1024 / 1024,
		HeapInuseMB:   float64(m.HeapInuse) / 1024 / 1024,
		StackSysMB:    float64(m.StackSys) / 1024 / 1024,
		NumGC:         m.NumGC,
		GCPauseMS:     float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000,
		NumGoroutines: runtime.NumGoroutine(),
		PoolStats:     GetPlayerPoolStats(),
	}

	// Add application-specific metrics
	if storage != nil {
		if datasets, err := storage.List(); err == nil {
			snapshot.DatasetCount = len(datasets)
		}
	}

	if memCache != nil {
		memCache.mutex.RLock()
		snapshot.CacheItemCount = len(memCache.items)
		snapshot.CacheSizeMB = float64(memCache.currentSize) / 1024 / 1024
		memCache.mutex.RUnlock()
	}

	// Store snapshot
	mp.snapshots = append(mp.snapshots, snapshot)
	if len(mp.snapshots) > mp.maxSize {
		mp.snapshots = mp.snapshots[1:]
	}

	return snapshot
}

// AnalyzeMemoryUsage provides comprehensive memory analysis
func (mp *MemoryProfiler) AnalyzeMemoryUsage() MemoryAnalysis {
	if len(mp.snapshots) == 0 {
		mp.TakeSnapshot()
	}

	latest := mp.snapshots[len(mp.snapshots)-1]

	analysis := MemoryAnalysis{
		Summary:            latest,
		TopMemoryConsumers: mp.identifyTopConsumers(latest),
		Recommendations:    mp.generateRecommendations(latest),
		TrendAnalysis:      mp.analyzeTrends(),
		GCEfficiency:       mp.analyzeGCEfficiency(),
	}

	return analysis
}

// identifyTopConsumers estimates memory usage by component
func (mp *MemoryProfiler) identifyTopConsumers(snapshot MemorySnapshot) []MemoryConsumer {
	consumers := []MemoryConsumer{
		{
			Component:   "Heap Allocated",
			EstimatedMB: snapshot.HeapAllocMB,
			Percentage:  (snapshot.HeapAllocMB / snapshot.SysMemoryMB) * 100,
			Description: "Currently allocated heap memory",
		},
		{
			Component:   "Cache Storage",
			EstimatedMB: snapshot.CacheSizeMB,
			Percentage:  (snapshot.CacheSizeMB / snapshot.SysMemoryMB) * 100,
			Description: fmt.Sprintf("In-memory cache (%d items)", snapshot.CacheItemCount),
		},
		{
			Component:   "Stack Memory",
			EstimatedMB: snapshot.StackSysMB,
			Percentage:  (snapshot.StackSysMB / snapshot.SysMemoryMB) * 100,
			Description: fmt.Sprintf("Goroutine stacks (%d goroutines)", snapshot.NumGoroutines),
		},
		{
			Component:   "Heap Idle",
			EstimatedMB: snapshot.HeapIdleMB,
			Percentage:  (snapshot.HeapIdleMB / snapshot.SysMemoryMB) * 100,
			Description: "Idle heap pages (can be reclaimed)",
		},
	}

	// Sort by memory usage
	sort.Slice(consumers, func(i, j int) bool {
		return consumers[i].EstimatedMB > consumers[j].EstimatedMB
	})

	return consumers
}

// generateRecommendations provides optimization suggestions
func (mp *MemoryProfiler) generateRecommendations(snapshot MemorySnapshot) []MemoryRecommendation {
	recommendations := []MemoryRecommendation{}

	// High memory usage recommendations
	if snapshot.HeapAllocMB > 1024 {
		recommendations = append(recommendations, MemoryRecommendation{
			Priority:     "HIGH",
			Component:    "Heap Memory",
			Issue:        "High heap allocation",
			Suggestion:   "Consider implementing memory-mapped files for large datasets",
			ExpectedGain: "50-70% heap reduction",
		})
	}

	// Cache size recommendations
	if snapshot.CacheSizeMB > 50 {
		recommendations = append(recommendations, MemoryRecommendation{
			Priority:     "MEDIUM",
			Component:    "Cache",
			Issue:        "Large cache size",
			Suggestion:   "Implement cache compression or reduce cache TTL",
			ExpectedGain: "20-30% cache reduction",
		})
	}

	// GC efficiency recommendations
	if snapshot.GCPauseMS > 10 {
		recommendations = append(recommendations, MemoryRecommendation{
			Priority:     "HIGH",
			Component:    "Garbage Collection",
			Issue:        "High GC pause times",
			Suggestion:   "Reduce GOGC value or implement streaming for large operations",
			ExpectedGain: "50-80% pause reduction",
		})
	}

	// Goroutine recommendations
	if snapshot.NumGoroutines > 1000 {
		recommendations = append(recommendations, MemoryRecommendation{
			Priority:     "MEDIUM",
			Component:    "Goroutines",
			Issue:        "High goroutine count",
			Suggestion:   "Review goroutine lifecycle and implement worker pools",
			ExpectedGain: "10-20% memory reduction",
		})
	}

	return recommendations
}

// analyzeTrends analyzes memory usage trends
func (mp *MemoryProfiler) analyzeTrends() MemoryTrendAnalysis {
	if len(mp.snapshots) < 2 {
		return MemoryTrendAnalysis{
			MemoryGrowthRate:   0,
			GCFrequencyTrend:   "INSUFFICIENT_DATA",
			PerformanceTrend:   "INSUFFICIENT_DATA",
			PredictedPeakHours: 0,
		}
	}

	// Calculate memory growth rate
	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]
	timeDiff := last.Timestamp.Sub(first.Timestamp).Hours()
	memoryDiff := last.TotalAllocMB - first.TotalAllocMB
	growthRate := memoryDiff / timeDiff

	// Analyze GC trends
	gcTrend := "STABLE"
	if len(mp.snapshots) >= 10 {
		recentGC := 0
		oldGC := 0
		midpoint := len(mp.snapshots) / 2

		for i := midpoint; i < len(mp.snapshots); i++ {
			recentGC += int(mp.snapshots[i].NumGC)
		}
		for i := 0; i < midpoint; i++ {
			oldGC += int(mp.snapshots[i].NumGC)
		}

		if recentGC > oldGC*2 {
			gcTrend = "INCREASING"
		} else if recentGC < oldGC/2 {
			gcTrend = "DECREASING"
		}
	}

	return MemoryTrendAnalysis{
		MemoryGrowthRate:   growthRate,
		GCFrequencyTrend:   gcTrend,
		PerformanceTrend:   "STABLE", // Could be calculated based on response times
		PredictedPeakHours: 24,       // Estimate based on growth rate
	}
}

// analyzeGCEfficiency analyzes garbage collection efficiency
func (mp *MemoryProfiler) analyzeGCEfficiency() GCEfficiencyMetrics {
	if len(mp.snapshots) == 0 {
		return GCEfficiencyMetrics{}
	}

	// Calculate average GC pause
	totalPause := 0.0
	validPauses := 0
	for _, snapshot := range mp.snapshots {
		if snapshot.GCPauseMS > 0 {
			totalPause += snapshot.GCPauseMS
			validPauses++
		}
	}

	avgPause := 0.0
	if validPauses > 0 {
		avgPause = totalPause / float64(validPauses)
	}

	// Calculate GC frequency
	gcFrequency := 0.0
	if len(mp.snapshots) >= 2 {
		first := mp.snapshots[0]
		last := mp.snapshots[len(mp.snapshots)-1]
		timeDiff := last.Timestamp.Sub(first.Timestamp).Hours()
		gcDiff := last.NumGC - first.NumGC
		gcFrequency = float64(gcDiff) / timeDiff
	}

	// Rate efficiency
	efficiency := "EXCELLENT"
	if avgPause > 50 {
		efficiency = "POOR"
	} else if avgPause > 20 {
		efficiency = "FAIR"
	} else if avgPause > 10 {
		efficiency = "GOOD"
	}

	return GCEfficiencyMetrics{
		AverageGCPauseMS:    avgPause,
		GCFrequencyPerHour:  gcFrequency,
		GCEfficiencyRating:  efficiency,
		MemoryReclamationMB: 0, // Could be calculated from heap differences
	}
}

// Memory profiling HTTP handlers
func memoryProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.memory_profile")
	defer span.End()

	analysis := globalMemoryProfiler.AnalyzeMemoryUsage()

	SetSpanAttributes(ctx,
		attribute.Float64("memory.total_alloc_mb", analysis.Summary.TotalAllocMB),
		attribute.Float64("memory.heap_alloc_mb", analysis.Summary.HeapAllocMB),
		attribute.Int("memory.dataset_count", analysis.Summary.DatasetCount),
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		logError(ctx, "Failed to encode memory analysis", "error", err)
		http.Error(w, "Failed to generate memory analysis", http.StatusInternalServerError)
	}
}

func memorySnapshotHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.memory_snapshot")
	defer span.End()

	snapshot := globalMemoryProfiler.TakeSnapshot()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		logError(ctx, "Failed to encode memory snapshot", "error", err)
		http.Error(w, "Failed to generate memory snapshot", http.StatusInternalServerError)
	}
}

func garbageCollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.force_gc")
	defer span.End()

	before := GetCurrentMemoryStats()

	// Force garbage collection
	runtime.GC()
	runtime.GC() // Second GC for better cleanup

	after := GetCurrentMemoryStats()
	freed := before.TotalAllocMB - after.TotalAllocMB

	result := map[string]interface{}{
		"before_mb": before.TotalAllocMB,
		"after_mb":  after.TotalAllocMB,
		"freed_mb":  freed,
		"timestamp": time.Now(),
	}

	SetSpanAttributes(ctx,
		attribute.Float64("gc.memory_freed_mb", freed),
		attribute.Float64("gc.before_mb", before.TotalAllocMB),
		attribute.Float64("gc.after_mb", after.TotalAllocMB),
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logError(ctx, "Failed to encode GC result", "error", err)
		http.Error(w, "Failed to encode GC result", http.StatusInternalServerError)
	}
}

// advancedOptimizationHandler provides advanced memory optimization
func advancedOptimizationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ctx, span := StartSpan(ctx, "api.advanced_optimization")
	defer span.End()

	// Get optimization stats
	stats := GlobalAdvancedOptimizer.GetOptimizationStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "enabled",
		"stats":   stats,
		"message": "Advanced memory optimization is active",
	}); err != nil {
		logError(ctx, "Failed to encode optimization stats", "error", err)
		http.Error(w, "Failed to encode optimization stats", http.StatusInternalServerError)
	}
}

// RegisterMemoryProfileEndpoints registers memory profiling endpoints on the main mux
// and starts a separate bare debug server for pprof (no middleware — pprof writes its
// own gzip binary format and must not be wrapped by CompressionMiddleware).
func RegisterMemoryProfileEndpoints(mux *http.ServeMux) {
	// Memory analysis endpoints go on the main mux (middleware is fine here)
	mux.Handle("/api/debug/memory/analysis", wrapHandler(http.HandlerFunc(memoryProfileHandler), "memory-analysis"))
	mux.Handle("/api/debug/memory/snapshot", wrapHandler(http.HandlerFunc(memorySnapshotHandler), "memory-snapshot"))
	mux.Handle("/api/debug/memory/gc", wrapHandler(http.HandlerFunc(garbageCollectHandler), "force-gc"))
	mux.Handle("/api/debug/memory/optimize", wrapHandler(http.HandlerFunc(advancedOptimizationHandler), "advanced-optimization"))

	// pprof runs on its own bare mux so that CompressionMiddleware never intercepts it.
	debugPort := os.Getenv("DEBUG_PORT")
	if debugPort == "" {
		debugPort = "8092"
	}
	debugMux := http.NewServeMux()
	debugMux.HandleFunc("/debug/pprof/", pprof.Index)
	debugMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	debugMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	debugMux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	debugMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	debugMux.Handle("/debug/pprof/block", pprof.Handler("block"))
	debugMux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	go func() {
		LogInfo("pprof debug server listening on :%s", debugPort)
		if err := http.ListenAndServe(":"+debugPort, debugMux); err != nil {
			LogWarn("pprof debug server stopped: %v", err)
		}
	}()

	LogInfo("Memory profiling endpoints registered at /api/debug/memory/* and pprof at :%s/debug/pprof/*", debugPort)
}
