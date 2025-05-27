//go:build no_otel

package main

import (
	"time"
)

// initMetrics is a no-op when OTEL is disabled
func initMetrics() {
	// No-op implementation
}

// recordUploadMetrics is a no-op when OTEL is disabled
func recordUploadMetrics(filename string, fileSizeBytes int64, totalDuration, parseDuration time.Duration, 
	playersProcessed int, memoryAllocMB float64, numWorkers, numGoroutines int) {
	// No-op implementation
}