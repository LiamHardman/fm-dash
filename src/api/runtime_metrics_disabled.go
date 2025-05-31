//go:build no_otel

package main

import (
	"context"
	"time"
)

// initRuntimeMetrics is a no-op when OTEL is disabled
func initRuntimeMetrics() {
	// No-op
}

// GetMemoryStats returns empty map when OTEL is disabled
func GetMemoryStats() map[string]interface{} {
	return make(map[string]interface{})
}

// LogMemoryStats is a no-op when OTEL is disabled
func LogMemoryStats(ctx context.Context) {
	// No-op
}

// StartMemoryMonitoring is a no-op when OTEL is disabled
func StartMemoryMonitoring(ctx context.Context, interval time.Duration) {
	// No-op
} 