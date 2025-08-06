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

			// Clear search cache to prevent memory leaks
			GetOptimizedSearch().ClearCache()

			// Clear position matching cache to prevent memory leaks
			ClearPositionMatchCache()

			LogDebug("Periodic cache cleanup completed - cleared role, search, and position caches")
			SetSpanAttributes(ctx,
				attribute.String("cache.cleanup.status", "completed"),
				attribute.String("cache.types_cleaned", "role_overall,search,position_match"),
			)
			span.End()
		}
	}()

	LogInfo("Cache cleanup scheduler started - cleaning every 10 minutes")
}
