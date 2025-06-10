package main

import (
	"context"
	"log"
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

	RecordBusinessOperation(ctx, "dataset_store", true, map[string]interface{}{
		"dataset_id":   datasetID,
		"player_count": len(players),
		"currency":     currencySymbol,
	})

	return nil
}

// StoreDatasetAsync stores player data using the storage interface asynchronously
func StoreDatasetAsync(datasetID string, players []Player, currencySymbol string) {
	go func() {
		ctx := context.Background()
		ctx, span := StartSpan(ctx, "store.dataset_async")
		defer span.End()

		SetSpanAttributes(ctx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.player_count", len(players)),
			attribute.String("dataset.currency", currencySymbol),
			attribute.String("operation.type", "async"),
		)

		startTime := time.Now()

		data := DatasetData{
			Players:        players,
			CurrencySymbol: currencySymbol,
		}

		err := storage.Store(datasetID, data)
		duration := time.Since(startTime)

		if err != nil {
			RecordError(ctx, err, "Failed to store dataset asynchronously")
			log.Printf("Error storing dataset %s asynchronously: %v", datasetID, err)
			return
		}

		SetSpanAttributes(ctx,
			attribute.Int64("operation.duration_ms", duration.Milliseconds()),
			attribute.String("operation.status", "success"),
		)

		RecordBusinessOperation(ctx, "dataset_store_async", true, map[string]interface{}{
			"dataset_id":     datasetID,
			"player_count":   len(players),
			"currency":       currencySymbol,
			"duration_ms":    duration.Milliseconds(),
			"operation_type": "async",
		})

		log.Printf("Successfully stored dataset %s asynchronously in %v", datasetID, duration)
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
	log.Println("Started automatic dataset cleanup scheduler (runs daily)")
}

func runCleanup() {
	log.Println("Starting automatic cleanup of old datasets...")

	// Define datasets to exclude from cleanup
	excludeDatasets := []string{"demo"}

	// Clean up datasets older than 30 days
	maxAge := 30 * 24 * time.Hour

	err := CleanupOldDatasets(maxAge, excludeDatasets)
	if err != nil {
		log.Printf("Error during automatic cleanup: %v", err)
	} else {
		log.Println("Automatic cleanup completed successfully")
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
	go func() {
		asyncCtx := context.Background()
		asyncCtx, asyncSpan := StartSpan(asyncCtx, "store.new_storage_async_operation")
		defer asyncSpan.End()

		SetSpanAttributes(asyncCtx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.player_count", len(players)),
			attribute.String("operation.type", "async_persistent_storage"),
		)

		startTime := time.Now()

		if err := StoreDataset(datasetID, players, currencySymbol); err != nil {
			RecordError(asyncCtx, err, "Failed to store in new storage system asynchronously")
			log.Printf("Error storing dataset %s to persistent storage asynchronously: %v", datasetID, err)
			return
		}

		duration := time.Since(startTime)
		SetSpanAttributes(asyncCtx,
			attribute.Int64("async_storage.duration_ms", duration.Milliseconds()),
			attribute.String("async_storage.status", "success"),
		)

		log.Printf("Successfully stored dataset %s to persistent storage asynchronously in %v", datasetID, duration)
	}()
}
