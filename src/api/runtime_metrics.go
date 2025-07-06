package main

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// Go runtime metrics
	goMemoryUsage metric.Int64ObservableGauge
	goMemoryHeap  metric.Int64ObservableGauge
	goMemoryStack metric.Int64ObservableGauge
	goGCDuration  metric.Float64ObservableCounter
	goGCCount     metric.Int64ObservableCounter
	goGoroutines  metric.Int64ObservableGauge
	goNextGC      metric.Int64ObservableGauge
	goLastGC      metric.Int64ObservableGauge
	goCgoCalls    metric.Int64ObservableCounter
)

// initRuntimeMetrics initializes Go runtime metrics
func initRuntimeMetrics() {
	if !otelEnabled {
		return
	}

	meter := otel.Meter("v2fmdash-runtime")

	var err error

	// Memory metrics
	goMemoryUsage, err = meter.Int64ObservableGauge(
		"go.memory.used",
		metric.WithDescription("Total memory used by Go runtime"),
		metric.WithUnit("By"),
	)
	if err != nil {
		slog.Error("Failed to create Go memory usage gauge", "error", err)
	}

	goMemoryHeap, err = meter.Int64ObservableGauge(
		"go.memory.heap",
		metric.WithDescription("Heap memory statistics"),
		metric.WithUnit("By"),
	)
	if err != nil {
		slog.Error("Failed to create Go heap memory gauge", "error", err)
	}

	goMemoryStack, err = meter.Int64ObservableGauge(
		"go.memory.stack",
		metric.WithDescription("Stack memory used by Go runtime"),
		metric.WithUnit("By"),
	)
	if err != nil {
		slog.Error("Failed to create Go stack memory gauge", "error", err)
	}

	// GC metrics
	goGCDuration, err = meter.Float64ObservableCounter(
		"go.gc.duration",
		metric.WithDescription("Time spent in garbage collection"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("Failed to create Go GC duration counter", "error", err)
	}

	goGCCount, err = meter.Int64ObservableCounter(
		"go.gc.count",
		metric.WithDescription("Number of garbage collections"),
	)
	if err != nil {
		slog.Error("Failed to create Go GC count counter", "error", err)
	}

	goNextGC, err = meter.Int64ObservableGauge(
		"go.gc.next_gc",
		metric.WithDescription("Target heap size for next GC"),
		metric.WithUnit("By"),
	)
	if err != nil {
		slog.Error("Failed to create Go next GC gauge", "error", err)
	}

	goLastGC, err = meter.Int64ObservableGauge(
		"go.gc.last_gc",
		metric.WithDescription("Time of last garbage collection"),
		metric.WithUnit("ns"),
	)
	if err != nil {
		slog.Error("Failed to create Go last GC gauge", "error", err)
	}

	// Goroutine metrics
	goGoroutines, err = meter.Int64ObservableGauge(
		"go.goroutines",
		metric.WithDescription("Number of goroutines"),
	)
	if err != nil {
		slog.Error("Failed to create Go goroutines gauge", "error", err)
	}

	// CGO metrics
	goCgoCalls, err = meter.Int64ObservableCounter(
		"go.cgo_calls",
		metric.WithDescription("Number of CGO calls"),
	)
	if err != nil {
		slog.Error("Failed to create Go CGO calls counter", "error", err)
	}

	// Register callback to collect runtime metrics
	_, err = meter.RegisterCallback(
		collectRuntimeMetrics,
		goMemoryUsage,
		goMemoryHeap,
		goMemoryStack,
		goGCDuration,
		goGCCount,
		goNextGC,
		goLastGC,
		goGoroutines,
		goCgoCalls,
	)
	if err != nil {
		slog.Error("Failed to register runtime metrics callback", "error", err)
		return
	}

	slog.Info("Go runtime metrics initialized")
}

// safeUint64ToInt64 safely converts uint64 to int64, capping at max int64 to avoid overflow
func safeUint64ToInt64(val uint64) int64 {
	const maxInt64 = 9223372036854775807 // 2^63 - 1
	if val > maxInt64 {
		return maxInt64
	}
	return int64(val)
}

// collectRuntimeMetrics collects Go runtime statistics
func collectRuntimeMetrics(_ context.Context, o metric.Observer) error {
	if !otelEnabled {
		return nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	if goMemoryUsage != nil {
		o.ObserveInt64(goMemoryUsage, safeUint64ToInt64(m.Sys),
			metric.WithAttributes(attribute.String("memory.type", "total")))
		o.ObserveInt64(goMemoryUsage, safeUint64ToInt64(m.Alloc),
			metric.WithAttributes(attribute.String("memory.type", "allocated")))
		o.ObserveInt64(goMemoryUsage, safeUint64ToInt64(m.TotalAlloc),
			metric.WithAttributes(attribute.String("memory.type", "total_allocated")))
	}

	if goMemoryHeap != nil {
		o.ObserveInt64(goMemoryHeap, safeUint64ToInt64(m.HeapAlloc),
			metric.WithAttributes(attribute.String("heap.type", "allocated")))
		o.ObserveInt64(goMemoryHeap, safeUint64ToInt64(m.HeapSys),
			metric.WithAttributes(attribute.String("heap.type", "system")))
		o.ObserveInt64(goMemoryHeap, safeUint64ToInt64(m.HeapIdle),
			metric.WithAttributes(attribute.String("heap.type", "idle")))
		o.ObserveInt64(goMemoryHeap, safeUint64ToInt64(m.HeapInuse),
			metric.WithAttributes(attribute.String("heap.type", "in_use")))
		o.ObserveInt64(goMemoryHeap, safeUint64ToInt64(m.HeapReleased),
			metric.WithAttributes(attribute.String("heap.type", "released")))
	}

	if goMemoryStack != nil {
		o.ObserveInt64(goMemoryStack, safeUint64ToInt64(m.StackSys),
			metric.WithAttributes(attribute.String("stack.type", "system")))
		o.ObserveInt64(goMemoryStack, safeUint64ToInt64(m.StackInuse),
			metric.WithAttributes(attribute.String("stack.type", "in_use")))
	}

	// GC metrics
	if goGCDuration != nil {
		// Convert nanoseconds to seconds
		gcDuration := float64(m.PauseTotalNs) / 1e9
		o.ObserveFloat64(goGCDuration, gcDuration)
	}

	if goGCCount != nil {
		o.ObserveInt64(goGCCount, int64(m.NumGC))
	}

	if goNextGC != nil {
		o.ObserveInt64(goNextGC, safeUint64ToInt64(m.NextGC))
	}

	if goLastGC != nil {
		o.ObserveInt64(goLastGC, safeUint64ToInt64(m.LastGC))
	}

	// Goroutine metrics
	if goGoroutines != nil {
		o.ObserveInt64(goGoroutines, int64(runtime.NumGoroutine()))
	}

	// CGO metrics
	if goCgoCalls != nil {
		o.ObserveInt64(goCgoCalls, runtime.NumCgoCall())
	}

	return nil
}

// GetMemoryStats returns current memory statistics for logging/debugging
func GetMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_bytes":       m.Alloc,
		"total_alloc_bytes": m.TotalAlloc,
		"sys_bytes":         m.Sys,
		"heap_alloc_bytes":  m.HeapAlloc,
		"heap_sys_bytes":    m.HeapSys,
		"heap_inuse_bytes":  m.HeapInuse,
		"heap_idle_bytes":   m.HeapIdle,
		"stack_sys_bytes":   m.StackSys,
		"stack_inuse_bytes": m.StackInuse,
		"num_gc":            m.NumGC,
		"gc_pause_total_ns": m.PauseTotalNs,
		"next_gc_bytes":     m.NextGC,
		"last_gc_ns":        m.LastGC,
		"num_goroutine":     runtime.NumGoroutine(),
		"num_cgo_call":      runtime.NumCgoCall(),
	}
}

// LogMemoryStats logs current memory statistics
func LogMemoryStats(ctx context.Context) {
	if !otelEnabled {
		return
	}

	stats := GetMemoryStats()
	slog.InfoContext(ctx, "Runtime memory statistics", "stats", stats)
}

// StartMemoryMonitoring starts a goroutine that periodically logs memory stats
func StartMemoryMonitoring(ctx context.Context, interval time.Duration) {
	if !otelEnabled {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				LogMemoryStats(ctx)
			case <-ctx.Done():
				slog.InfoContext(ctx, "Memory monitoring stopped")
				return
			}
		}
	}()

	slog.InfoContext(ctx, "Memory monitoring started", "interval", interval)
}
