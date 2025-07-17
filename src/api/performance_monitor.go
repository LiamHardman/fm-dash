package main

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// PerformanceMetrics tracks detailed performance statistics
type PerformanceMetrics struct {
	// Parsing metrics
	TotalRowsParsed        int64
	TotalPlayersProcessed  int64
	TotalParseErrors       int64
	TotalParseTime         int64 // nanoseconds
	AverageParseTimePerRow int64 // nanoseconds

	// Memory metrics
	CurrentGoroutines  int32
	PeakMemoryUsage    uint64
	CurrentMemoryUsage uint64
	GCCollections      uint32

	// Worker metrics
	ActiveWorkers     int32
	WorkerUtilization float64

	// JSON metrics
	JSONMarshalOperations   int64
	JSONMarshalTime         int64 // nanoseconds
	JSONUnmarshalOperations int64
	JSONUnmarshalTime       int64 // nanoseconds

	// Channel metrics
	ChannelBackpressureEvents int64
	ChannelTimeouts           int64
}

// Global performance metrics instance
var globalMetrics = &PerformanceMetrics{}

// ParseTimer tracks timing for parsing operations
type ParseTimer struct {
	startTime time.Time
	operation string
	context   context.Context
}

// CreateParseTimer creates a new timer for tracking parsing performance
func CreateParseTimer(operation string) *ParseTimer {
	return &ParseTimer{
		startTime: time.Now(),
		operation: operation,
		context:   context.Background(),
	}
}

// CreateParseTimerWithContext creates a timer with context
func CreateParseTimerWithContext(ctx context.Context, operation string) *ParseTimer {
	return &ParseTimer{
		startTime: time.Now(),
		operation: operation,
		context:   ctx,
	}
}

// Finish completes the timing and records metrics
func (pt *ParseTimer) Finish(rowsProcessed, errors int64) {
	duration := time.Since(pt.startTime)

	// Update metrics atomically
	atomic.AddInt64(&globalMetrics.TotalRowsParsed, rowsProcessed)
	atomic.AddInt64(&globalMetrics.TotalParseErrors, errors)
	atomic.AddInt64(&globalMetrics.TotalParseTime, duration.Nanoseconds())

	if rowsProcessed > 0 {
		avgTime := duration.Nanoseconds() / rowsProcessed
		atomic.StoreInt64(&globalMetrics.AverageParseTimePerRow, avgTime)
	}

	logInfo(pt.context, "Parse operation completed",
		"operation", pt.operation,
		"rows_processed", rowsProcessed,
		"duration_ms", duration.Milliseconds(),
		"avg_time_per_row_ns", duration.Nanoseconds()/maxInt64(rowsProcessed, 1),
		"errors", errors)
}

// JSONTimer tracks JSON operation performance
type JSONTimer struct {
	startTime   time.Time
	isUnmarshal bool
}

// CreateJSONTimer creates a new timer for tracking JSON operations
func CreateJSONTimer(isUnmarshal bool) *JSONTimer {
	return &JSONTimer{
		startTime:   time.Now(),
		isUnmarshal: isUnmarshal,
	}
}

// Finish completes JSON timing
func (jt *JSONTimer) Finish() {
	duration := time.Since(jt.startTime)

	if jt.isUnmarshal {
		atomic.AddInt64(&globalMetrics.JSONUnmarshalOperations, 1)
		atomic.AddInt64(&globalMetrics.JSONUnmarshalTime, duration.Nanoseconds())
	} else {
		atomic.AddInt64(&globalMetrics.JSONMarshalOperations, 1)
		atomic.AddInt64(&globalMetrics.JSONMarshalTime, duration.Nanoseconds())
	}
}

// RecordBackpressure records a backpressure event
func RecordBackpressure() {
	atomic.AddInt64(&globalMetrics.ChannelBackpressureEvents, 1)
}

// RecordChannelTimeout records a channel timeout
func RecordChannelTimeout() {
	atomic.AddInt64(&globalMetrics.ChannelTimeouts, 1)
}

// RecordWorkerStart increments active worker count
func RecordWorkerStart() {
	atomic.AddInt32(&globalMetrics.ActiveWorkers, 1)
}

// RecordWorkerEnd decrements active worker count
func RecordWorkerEnd() {
	atomic.AddInt32(&globalMetrics.ActiveWorkers, -1)
}

// UpdateMemoryMetrics updates memory usage statistics
func UpdateMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	atomic.StoreUint64(&globalMetrics.CurrentMemoryUsage, m.Alloc)
	atomic.StoreUint32(&globalMetrics.GCCollections, m.NumGC)
	// Safe conversion to avoid integer overflow
	numGoroutines := runtime.NumGoroutine()
	var safeGoroutines int32
	if numGoroutines > 2147483647 { // max int32
		safeGoroutines = 2147483647
	} else {
		safeGoroutines = int32(numGoroutines) // #nosec G115 - bounds checked above
	}
	atomic.StoreInt32(&globalMetrics.CurrentGoroutines, safeGoroutines)

	// Update peak memory if current usage is higher
	for {
		current := atomic.LoadUint64(&globalMetrics.PeakMemoryUsage)
		if m.Alloc <= current {
			break
		}
		if atomic.CompareAndSwapUint64(&globalMetrics.PeakMemoryUsage, current, m.Alloc) {
			break
		}
	}
}

// GetMetricsSnapshot returns a snapshot of current metrics
func GetMetricsSnapshot() PerformanceMetrics {
	UpdateMemoryMetrics()

	return PerformanceMetrics{
		TotalRowsParsed:           atomic.LoadInt64(&globalMetrics.TotalRowsParsed),
		TotalPlayersProcessed:     atomic.LoadInt64(&globalMetrics.TotalPlayersProcessed),
		TotalParseErrors:          atomic.LoadInt64(&globalMetrics.TotalParseErrors),
		TotalParseTime:            atomic.LoadInt64(&globalMetrics.TotalParseTime),
		AverageParseTimePerRow:    atomic.LoadInt64(&globalMetrics.AverageParseTimePerRow),
		CurrentGoroutines:         atomic.LoadInt32(&globalMetrics.CurrentGoroutines),
		PeakMemoryUsage:           atomic.LoadUint64(&globalMetrics.PeakMemoryUsage),
		CurrentMemoryUsage:        atomic.LoadUint64(&globalMetrics.CurrentMemoryUsage),
		GCCollections:             atomic.LoadUint32(&globalMetrics.GCCollections),
		ActiveWorkers:             atomic.LoadInt32(&globalMetrics.ActiveWorkers),
		JSONMarshalOperations:     atomic.LoadInt64(&globalMetrics.JSONMarshalOperations),
		JSONMarshalTime:           atomic.LoadInt64(&globalMetrics.JSONMarshalTime),
		JSONUnmarshalOperations:   atomic.LoadInt64(&globalMetrics.JSONUnmarshalOperations),
		JSONUnmarshalTime:         atomic.LoadInt64(&globalMetrics.JSONUnmarshalTime),
		ChannelBackpressureEvents: atomic.LoadInt64(&globalMetrics.ChannelBackpressureEvents),
		ChannelTimeouts:           atomic.LoadInt64(&globalMetrics.ChannelTimeouts),
	}
}

// LogPerformanceReport logs a detailed performance report
func LogPerformanceReport(ctx context.Context) {
	logInfo(ctx, "Generating performance report")
	start := time.Now()

	metrics := GetMetricsSnapshot()

	logInfo(ctx, "=== PERFORMANCE REPORT ===")
	logInfo(ctx, "Parsing metrics",
		"rows_parsed", metrics.TotalRowsParsed,
		"players_processed", metrics.TotalPlayersProcessed,
		"parse_errors", metrics.TotalParseErrors)
	logInfo(ctx, "Parse timing",
		"avg_parse_time_per_row_ns", metrics.AverageParseTimePerRow,
		"avg_parse_time_per_row", time.Duration(metrics.AverageParseTimePerRow).String())
	logInfo(ctx, "Memory metrics",
		"current_memory", formatBytes(metrics.CurrentMemoryUsage),
		"peak_memory", formatBytes(metrics.PeakMemoryUsage),
		"gc_collections", metrics.GCCollections)
	logInfo(ctx, "Worker metrics",
		"active_workers", metrics.ActiveWorkers,
		"goroutines", metrics.CurrentGoroutines)
	logInfo(ctx, "JSON operation metrics",
		"marshal_operations", metrics.JSONMarshalOperations,
		"marshal_avg_time", time.Duration(metrics.JSONMarshalTime/maxInt64(metrics.JSONMarshalOperations, 1)).String(),
		"unmarshal_operations", metrics.JSONUnmarshalOperations,
		"unmarshal_avg_time", time.Duration(metrics.JSONUnmarshalTime/maxInt64(metrics.JSONUnmarshalOperations, 1)).String())
	logInfo(ctx, "Channel metrics",
		"backpressure_events", metrics.ChannelBackpressureEvents,
		"timeouts", metrics.ChannelTimeouts)

	logDebug(ctx, "Performance report generation completed",
		"duration_ms", time.Since(start).Milliseconds())
}

// LogImmediatePerformanceStats logs performance stats immediately (for use after parsing completion)
func LogImmediatePerformanceStats() {
	UpdateMemoryMetrics()
	metrics := GetMetricsSnapshot()

	if metrics.TotalRowsParsed > 0 {
		totalTimeSeconds := float64(metrics.TotalParseTime) / 1e9 // Convert nanoseconds to seconds
		var rowsPerSecond float64
		if totalTimeSeconds > 0 {
			rowsPerSecond = float64(metrics.TotalRowsParsed) / totalTimeSeconds
		}

		LogInfo("Performance: %d rows parsed, %s memory, %d goroutines, %d workers, %.1fs total time, %.0f rows/sec",
			metrics.TotalRowsParsed,
			formatBytes(metrics.CurrentMemoryUsage),
			metrics.CurrentGoroutines,
			metrics.ActiveWorkers,
			totalTimeSeconds,
			rowsPerSecond)

		// Update the last logged count to prevent duplicate logging
		lastLoggedParseCount = metrics.TotalRowsParsed
	}
}

// formatBytes formats byte counts for human reading
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// maxInt64 returns the maximum of two int64 values
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Global variable to track last logged parse count
var lastLoggedParseCount int64

// StartPerformanceMonitoring starts a background goroutine to periodically log performance metrics
func StartPerformanceMonitoring(interval time.Duration) {
	ctx := context.Background()
	logDebug(ctx, "Starting performance monitoring", "interval_seconds", interval.Seconds())

	go func() {
		// Create a background context for the goroutine
		monitorCtx := context.Background()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logDebug(monitorCtx, "Performance monitoring goroutine started", "interval_seconds", interval.Seconds())

		for range ticker.C {
			UpdateMemoryMetrics()
			// Only log if there's been new parsing activity since last log
			metrics := GetMetricsSnapshot()
			currentParseCount := metrics.TotalRowsParsed

			if currentParseCount > 0 && currentParseCount != lastLoggedParseCount {
				totalTimeSeconds := float64(metrics.TotalParseTime) / 1e9 // Convert nanoseconds to seconds
				var rowsPerSecond float64
				if totalTimeSeconds > 0 {
					rowsPerSecond = float64(metrics.TotalRowsParsed) / totalTimeSeconds
				}

				logInfo(monitorCtx, "Performance monitoring update",
					"rows_parsed", metrics.TotalRowsParsed,
					"memory_usage_mb", fmt.Sprintf("%.1f", float64(metrics.CurrentMemoryUsage)/1024/1024),
					"memory_usage_formatted", formatBytes(metrics.CurrentMemoryUsage),
					"goroutines", metrics.CurrentGoroutines,
					"active_workers", metrics.ActiveWorkers,
					"total_time_seconds", totalTimeSeconds,
					"rows_per_second", rowsPerSecond,
					"gc_collections", metrics.GCCollections)

				// Update the last logged count
				lastLoggedParseCount = currentParseCount
			}
		}
	}()
}
