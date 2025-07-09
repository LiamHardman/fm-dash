package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Global storage instance
var storage StorageInterface

// Deprecated: Use storage interface instead
// playerDataStore holds the parsed player data in memory, keyed by a dataset ID.
// Each dataset also stores the detected currency symbol.
var playerDataStore = make(map[string]struct {
	Players        []Player
	CurrencySymbol string
})

// storeMutex protects concurrent access to playerDataStore.
var storeMutex sync.RWMutex

// InitStore initializes the global storage instance
func InitStore() {
	storage = InitializeStorage()
}

// StoreDataset stores player data using the storage interface
func StoreDataset(datasetID string, players []Player, currencySymbol string) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.dataset")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
	)

	data := DatasetData{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}

	err := storage.Store(datasetID, data)
	if err != nil {
		RecordError(ctx, err, "Failed to store dataset")
		return err
	}

	// Invalidate related cache entries when dataset is updated
	invalidateDatasetCache(datasetID)

	RecordBusinessOperation(ctx, "dataset_store", true, map[string]interface{}{
		"dataset_id":   datasetID,
		"player_count": len(players),
		"currency":     currencySymbol,
	})

	return nil
}

// StoreDatasetAsync stores player data using the storage interface asynchronously
func StoreDatasetAsync(datasetID string, players []Player, currencySymbol string) {
	// Create a deep copy of players slice to avoid race conditions during async storage
	playersCopy := make([]Player, len(players))
	copy(playersCopy, players)

	go func() {
		ctx := context.Background()
		ctx, span := StartSpan(ctx, "store.dataset_async")
		defer span.End()

		SetSpanAttributes(ctx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.player_count", len(playersCopy)),
			attribute.String("dataset.currency", currencySymbol),
			attribute.String("operation.type", "async"),
		)

		startTime := time.Now()

		data := DatasetData{
			Players:        playersCopy,
			CurrencySymbol: currencySymbol,
		}

		err := storage.Store(datasetID, data)
		duration := time.Since(startTime)

		if err != nil {
			RecordError(ctx, err, "Failed to store dataset asynchronously")
			LogWarn("Error storing dataset %s asynchronously: %v", sanitizeForLogging(datasetID), err)
			return
		}

		// Invalidate related cache entries when dataset is updated
		invalidateDatasetCache(datasetID)

		SetSpanAttributes(ctx,
			attribute.Int64("operation.duration_ms", duration.Milliseconds()),
			attribute.String("operation.status", "success"),
		)

		RecordBusinessOperation(ctx, "dataset_store_async", true, map[string]interface{}{
			"dataset_id":     datasetID,
			"player_count":   len(playersCopy),
			"currency":       currencySymbol,
			"duration_ms":    duration.Milliseconds(),
			"operation_type": "async",
		})

		LogInfo("Successfully stored dataset %s asynchronously in %v", sanitizeForLogging(datasetID), duration)
	}()
}

// RetrieveDataset retrieves player data using the storage interface
func RetrieveDataset(datasetID string) ([]Player, string, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.retrieve_dataset")
	defer span.End()

	SetSpanAttributes(ctx, attribute.String("dataset.id", datasetID))

	data, err := storage.Retrieve(datasetID)
	if err != nil {
		RecordError(ctx, err, "Failed to retrieve dataset")
		return nil, "", err
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("dataset.currency", data.CurrencySymbol),
	)

	RecordBusinessOperation(ctx, "dataset_retrieve", true, map[string]interface{}{
		"dataset_id":   datasetID,
		"player_count": len(data.Players),
	})

	return data.Players, data.CurrencySymbol, nil
}

// DeleteDataset deletes player data using the storage interface
func DeleteDataset(datasetID string) error {
	err := storage.Delete(datasetID)
	if err != nil {
		return err
	}

	// Remove the duplicate mapping after successful deletion
	removeDuplicateMapping(datasetID)

	return nil
}

// ListDatasets lists all available dataset IDs
func ListDatasets() ([]string, error) {
	return storage.List()
}

// CleanupOldDatasets removes datasets older than the specified duration, excluding specified datasets
func CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	return storage.CleanupOldDatasets(maxAge, excludeDatasets)
}

// StartCleanupScheduler starts a background goroutine that periodically cleans up old datasets
func StartCleanupScheduler() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run cleanup once per day
		defer ticker.Stop()

		// Run initial cleanup after 1 minute (to allow server to fully start)
		time.Sleep(1 * time.Minute)
		runCleanup()

		// Then run cleanup every 24 hours
		for range ticker.C {
			runCleanup()
		}
	}()
	LogInfo("Started automatic dataset cleanup scheduler (runs daily)")
}

// getRetentionPeriod returns the configured retention period for datasets
func getRetentionPeriod() time.Duration {
	// Default to 30 days if not configured
	defaultRetention := 30 * 24 * time.Hour

	// Get retention period from environment variable (in days)
	retentionDays := os.Getenv("DATASET_RETENTION_DAYS")
	if retentionDays == "" {
		return defaultRetention
	}

	// Parse the retention period
	days, err := strconv.Atoi(retentionDays)
	if err != nil || days <= 0 {
		LogWarn("Invalid DATASET_RETENTION_DAYS value: %s. Using default of 30 days.", retentionDays)
		return defaultRetention
	}

	return time.Duration(days) * 24 * time.Hour
}

func runCleanup() {
	LogInfo("Starting automatic cleanup of old datasets...")

	// Define datasets to exclude from cleanup
	excludeDatasets := []string{"demo", "1e0c8dcd-f6b8-4874-a72e-a2a3bdf20038"}

	// Get configured retention period
	maxAge := getRetentionPeriod()
	LogInfo("Using retention period of %.0f days", maxAge.Hours()/24)

	err := CleanupOldDatasets(maxAge, excludeDatasets)
	if err != nil {
		LogWarn("Error during automatic cleanup: %v", err)
	} else {
		LogInfo("Automatic cleanup completed successfully")
	}

	// Clean up stale duplicate mappings
	cleanupStaleDuplicateMappings()
}

// Legacy functions for backward compatibility

// GetPlayerData retrieves player data from the legacy store (for backward compatibility)
func GetPlayerData(datasetID string) ([]Player, string, bool) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.get_player_data_legacy")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("store.type", "legacy_compatible"),
	)

	// Try fast in-memory cache first for performance
	storeMutex.RLock()
	if data, exists := playerDataStore[datasetID]; exists {
		storeMutex.RUnlock()
		AddSpanEvent(ctx, "store.memory_cache_hit")
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(data.Players)),
			attribute.String("data.source", "memory_fast"),
		)
		return data.Players, data.CurrencySymbol, true
	}
	storeMutex.RUnlock()

	// Fallback to persistent storage only if not in memory
	AddSpanEvent(ctx, "store.fallback_to_persistent")
	players, currency, err := RetrieveDataset(datasetID)
	if err == nil {
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(players)),
			attribute.String("data.source", "persistent_fallback"),
		)
		return players, currency, true
	}

	SetSpanAttributes(ctx, attribute.String("result", "not_found"))
	return nil, "", false
}

// SetPlayerData stores player data in the legacy store and new storage
func SetPlayerData(datasetID string, players []Player, currencySymbol string) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.set_player_data_legacy")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
		attribute.String("store.type", "legacy_compatible"),
	)

	// Store in legacy format
	AddSpanEvent(ctx, "store.legacy_store")
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}
	storeMutex.Unlock()

	// Store in new storage system
	AddSpanEvent(ctx, "store.new_storage_store")
	if err := StoreDataset(datasetID, players, currencySymbol); err != nil {
		RecordError(ctx, err, "Failed to store in new storage system")
		// Log error but don't fail - legacy store still works
		// (error logging is handled in storage implementation)
	}
}

// SetPlayerDataAsync stores player data in both legacy store (immediately) and new storage (asynchronously)
func SetPlayerDataAsync(datasetID string, players []Player, currencySymbol string) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.set_player_data_async")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
		attribute.String("store.type", "legacy_compatible_async"),
	)

	// Store in legacy format immediately (fast, in-memory operation)
	AddSpanEvent(ctx, "store.legacy_store_immediate")
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}
	storeMutex.Unlock()

	// Store in new storage system asynchronously (potentially slow S3/disk operation)
	AddSpanEvent(ctx, "store.new_storage_async_queued")

	// Serialize the data immediately to avoid race conditions during async storage
	// This way the goroutine only works with immutable JSON data
	data := DatasetData{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}

	serializedData, err := json.Marshal(data)
	if err != nil {
		RecordError(ctx, err, "Failed to serialize data for async storage")
		LogWarn("Error serializing dataset %s for async storage: %v", sanitizeForLogging(datasetID), err)
		return
	}

	go func() {
		asyncCtx := context.Background()
		asyncCtx, asyncSpan := StartSpan(asyncCtx, "store.new_storage_async_operation")
		defer asyncSpan.End()

		SetSpanAttributes(asyncCtx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.serialized_bytes", len(serializedData)),
			attribute.String("operation.type", "async_persistent_storage"),
		)

		startTime := time.Now()

		// Deserialize and store using the existing storage interface
		var deserializedData DatasetData
		if err := json.Unmarshal(serializedData, &deserializedData); err != nil {
			RecordError(asyncCtx, err, "Failed to deserialize data for async storage")
			LogWarn("Error deserializing dataset %s for async storage: %v", sanitizeForLogging(datasetID), err)
			return
		}

		if err := StoreDataset(datasetID, deserializedData.Players, deserializedData.CurrencySymbol); err != nil {
			RecordError(asyncCtx, err, "Failed to store in new storage system asynchronously")
			LogWarn("Error storing dataset %s to persistent storage asynchronously: %v", sanitizeForLogging(datasetID), err)
			return
		}

		duration := time.Since(startTime)
		SetSpanAttributes(asyncCtx,
			attribute.Int64("async_storage.duration_ms", duration.Milliseconds()),
			attribute.String("async_storage.status", "success"),
		)

		LogInfo("Successfully stored dataset %s to persistent storage asynchronously in %v", sanitizeForLogging(datasetID), duration)
	}()
}

// invalidateDatasetCache clears cache entries related to a specific dataset
func invalidateDatasetCache(datasetID string) {
	// Invalidate memory cache entries for this dataset
	patterns := []string{
		fmt.Sprintf("leagues_%s", datasetID),
		fmt.Sprintf("teams_%s_*", datasetID), // Note: This is a pattern, we'll need to iterate through keys
		fmt.Sprintf("players_%s", datasetID),
		fmt.Sprintf("percentiles:%s:*", datasetID), // Percentile cache entries
		fmt.Sprintf("filtered:%s:*", datasetID),    // Filtered result cache entries
	}

	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			// For wildcard patterns, we need to iterate through cache and match
			invalidatePatternCache(pattern)
		} else {
			deleteFromMemCache(pattern)
			LogDebug("Invalidated cache key: %s", pattern)
		}
	}

	LogDebug("Invalidated all cache entries for dataset: %s", sanitizeForLogging(datasetID))
}

// invalidatePatternCache removes cache entries matching a pattern
func invalidatePatternCache(pattern string) {
	// Convert pattern to prefix (remove the * wildcard)
	prefix := strings.TrimSuffix(pattern, "*")

	// Get current cache stats to iterate
	if memCache == nil {
		return
	}

	memCache.mutex.RLock()
	keysToDelete := make([]string, 0)
	for key := range memCache.items {
		if strings.HasPrefix(key, prefix) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	memCache.mutex.RUnlock()

	// Delete matching keys
	for _, key := range keysToDelete {
		deleteFromMemCache(key)
		LogDebug("Invalidated cache key (pattern): %s", sanitizeForLogging(key))
	}
}
