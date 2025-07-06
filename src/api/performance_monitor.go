package main

import (
	"context"
	"fmt"
	"log"
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

// NewParseTimer creates a new timer for tracking parsing performance
func NewParseTimer(operation string) *ParseTimer {
	return &ParseTimer{
		startTime: time.Now(),
		operation: operation,
		context:   context.Background(),
	}
}

// NewParseTimerWithContext creates a timer with context
func NewParseTimerWithContext(ctx context.Context, operation string) *ParseTimer {
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

	log.Printf("Parse operation '%s' completed: %d rows in %v (avg: %v/row, errors: %d)",
		pt.operation, rowsProcessed, duration,
		time.Duration(duration.Nanoseconds()/maxInt64(rowsProcessed, 1)), errors)
}

// JSONTimer tracks JSON operation performance
type JSONTimer struct {
	startTime   time.Time
	isUnmarshal bool
}

// NewJSONTimer creates a timer for JSON operations
func NewJSONTimer(isUnmarshal bool) *JSONTimer {
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
func LogPerformanceReport() {
	metrics := GetMetricsSnapshot()

	log.Printf("=== PERFORMANCE REPORT ===")
	log.Printf("Parsing: %d rows, %d players, %d errors",
		metrics.TotalRowsParsed, metrics.TotalPlayersProcessed, metrics.TotalParseErrors)
	log.Printf("Average parse time per row: %v", time.Duration(metrics.AverageParseTimePerRow))
	log.Printf("Memory: Current=%s, Peak=%s, GC=%d",
		formatBytes(metrics.CurrentMemoryUsage),
		formatBytes(metrics.PeakMemoryUsage),
		metrics.GCCollections)
	log.Printf("Workers: Active=%d, Goroutines=%d",
		metrics.ActiveWorkers, metrics.CurrentGoroutines)
	log.Printf("JSON: Marshal=%d ops (%v avg), Unmarshal=%d ops (%v avg)",
		metrics.JSONMarshalOperations,
		time.Duration(metrics.JSONMarshalTime/maxInt64(metrics.JSONMarshalOperations, 1)),
		metrics.JSONUnmarshalOperations,
		time.Duration(metrics.JSONUnmarshalTime/maxInt64(metrics.JSONUnmarshalOperations, 1)))
	log.Printf("Channels: Backpressure=%d, Timeouts=%d",
		metrics.ChannelBackpressureEvents, metrics.ChannelTimeouts)
	log.Printf("=========================")
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

// StartPerformanceMonitoring starts a background goroutine to periodically log performance metrics
func StartPerformanceMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			UpdateMemoryMetrics()
			// Log summary every interval
			metrics := GetMetricsSnapshot()
			if metrics.TotalRowsParsed > 0 {
				log.Printf("Performance: %d rows parsed, %s memory, %d goroutines, %d workers",
					metrics.TotalRowsParsed,
					formatBytes(metrics.CurrentMemoryUsage),
					metrics.CurrentGoroutines,
					metrics.ActiveWorkers)
			}
		}
	}()
}
