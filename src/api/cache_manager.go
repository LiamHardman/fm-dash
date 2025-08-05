package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// StartCacheCleanupScheduler starts a background goroutine that periodically clears caches
func StartCacheCleanupScheduler() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute) // Clear caches every 10 minutes
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			ctx, span := StartSpan(ctx, "cache.cleanup")

			// Clear role calculation cache to prevent memory leaks
			ClearRoleOverallCache()

			LogDebug("Periodic cache cleanup completed")
			SetSpanAttributes(ctx, attribute.String("cache.cleanup.status", "completed"))
			span.End()
		}
	}()

	LogInfo("Cache cleanup scheduler started - cleaning every 10 minutes")
}
