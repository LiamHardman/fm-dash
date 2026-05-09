package main

import (
	"context"
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

// When true, bypasses the in-process dataset cache (playerDataStore) entirely.
// This reduces resident memory by avoiding duplication of full player slices in RAM.
// Set via environment variable: DISABLE_IN_MEMORY_DATASET_CACHE=true
var disableInMemoryDatasetCache = strings.ToLower(os.Getenv("DISABLE_IN_MEMORY_DATASET_CACHE")) == "true"

// InitStore initializes the global storage instance and shared config storage.
func InitStore() {
	ctx := context.Background()
	storage = InitializeStorage(ctx)
	InitConfigStorage()
	LoadOverridesFromStorage()
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

	// Invalidate search index when dataset is updated
	GetHybridSearchService().InvalidateIndex(datasetID)

	// Invalidate role index for upgrade finder optimization
	InvalidateRoleIndex()

	RecordBusinessOperation(ctx, "dataset_store", true, map[string]interface{}{
		"dataset_id":   datasetID,
		"player_count": len(players),
		"currency":     currencySymbol,
	})

	return nil
}

// StoreDatasetAsync stores player data using the storage interface asynchronously
func StoreDatasetAsync(datasetID string, players []Player, currencySymbol string) {
	// Create a proper deep copy of players slice to avoid race conditions during async storage
	// PERFORMANCE: Use FastDeepCopyPlayers for much better performance while staying thread-safe
	playersCopy := FastDeepCopyPlayers(players)

	// Ensure all players have their NumericAttributes populated before storage
	// This is critical because the data might be stored before the workers finish enhancement
	for i := range playersCopy {
		if len(playersCopy[i].NumericAttributes) == 0 {
			EnhancePlayerWithCalculations(&playersCopy[i])
		}
	}

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

		// Invalidate role index for upgrade finder optimization
		InvalidateRoleIndex()

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
	LogDebug("Started automatic dataset cleanup scheduler (runs daily)")
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
	LogDebug("Starting automatic cleanup of old datasets...")

	// Define datasets to exclude from cleanup
	excludeDatasets := []string{"demo", "1e0c8dcd-f6b8-4874-a72e-a2a3bdf20038"}

	// Get configured retention period
	maxAge := getRetentionPeriod()
	LogDebug("Using retention period of %.0f days", maxAge.Hours()/24)

	err := CleanupOldDatasets(maxAge, excludeDatasets)
	if err != nil {
		LogWarn("Error during automatic cleanup: %v", err)
	} else {
		LogDebug("Automatic cleanup completed successfully")
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

	if !disableInMemoryDatasetCache {
		// Try fast in-memory cache first for performance
		storeMutex.RLock()
		if data, exists := playerDataStore[datasetID]; exists {
			storeMutex.RUnlock()
			AddSpanEvent(ctx, "store.memory_cache_hit")
			SetSpanAttributes(ctx,
				attribute.Int("dataset.player_count", len(data.Players)),
				attribute.String("data.source", "memory_fast"),
			)
			// Return an optimized, safe deep copy to prevent race conditions
			// PERFORMANCE: Use FastDeepCopyPlayers for much better performance while staying thread-safe
			players := FastDeepCopyPlayers(data.Players)
			// Skip re-enhancement when data already enhanced (typical path).
			// Iterate all players rather than sampling only players[0]: if any single
			// player is missing computed fields the whole slice is treated as needing
			// enhancement so we never silently serve partial data.
			needsEnhancement := false
			for i := range players {
				if len(players[i].NumericAttributes) == 0 || players[i].PerformanceStatsNumeric == nil {
					needsEnhancement = true
					break
				}
			}
			if needsEnhancement {
				enhancedCount := 0
				for i := range players {
					EnhancePlayerWithCalculations(&players[i])
					enhancedCount++
				}
				if enhancedCount > 0 {
					LogDebug("Enhanced %d retrieved players for dataset %s (memory cache path)", enhancedCount, datasetID)
				}
			}
			return players, data.CurrencySymbol, true
		}
		storeMutex.RUnlock()
	}

	// Fallback to persistent storage only if not in memory
	AddSpanEvent(ctx, "store.fallback_to_persistent")
	players, currency, err := RetrieveDataset(datasetID)
	if err == nil {
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(players)),
			attribute.String("data.source", "persistent_fallback"),
		)
		// Return a deep copy to prevent race conditions
		// PERFORMANCE: Use FastDeepCopyPlayers for much better performance while staying thread-safe
		enhancedPlayers := FastDeepCopyPlayers(players)
		// Avoid redundant enhancement if already stored enhanced.
		// Iterate all players rather than sampling only enhancedPlayers[0] so that
		// partially-enhanced datasets (e.g. after a mid-flight crash) are fully recovered.
		needsEnhancement := false
		for i := range enhancedPlayers {
			if len(enhancedPlayers[i].NumericAttributes) == 0 || enhancedPlayers[i].PerformanceStatsNumeric == nil {
				needsEnhancement = true
				break
			}
		}
		if needsEnhancement {
			enhancedCount := 0
			for i := range enhancedPlayers {
				EnhancePlayerWithCalculations(&enhancedPlayers[i])
				enhancedCount++
			}
			if enhancedCount > 0 {
				LogDebug("Enhanced %d retrieved players for dataset %s (persistent path)", enhancedCount, datasetID)
			}
		}
		if !disableInMemoryDatasetCache {
			// Store the enhanced data in memory cache for future fast access
			storeMutex.Lock()
			playerDataStore[datasetID] = struct {
				Players        []Player
				CurrencySymbol string
			}{
				Players:        enhancedPlayers,
				CurrencySymbol: currency,
			}
			storeMutex.Unlock()
			LogInfo("Cached %d enhanced players from storage in memory for dataset %s", len(enhancedPlayers), datasetID)
		}

		return enhancedPlayers, currency, true
	}

	SetSpanAttributes(ctx, attribute.String("result", "not_found"))
	return nil, "", false
}

// GetPlayerDataForUpgradeFinder retrieves player data optimized for upgrade finder (minimal copy, no enhancements)
func GetPlayerDataForUpgradeFinder(datasetID string) ([]Player, string, bool) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.get_player_data_for_upgrade_finder")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("store.type", "upgrade_finder_optimized"),
	)

	// Try fast in-memory cache first
	storeMutex.RLock()
	if data, exists := playerDataStore[datasetID]; exists {
		storeMutex.RUnlock()
		AddSpanEvent(ctx, "store.memory_cache_hit_upgrade_finder")
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(data.Players)),
			attribute.String("data.source", "memory_fast_upgrade_finder"),
		)

		// Return minimal copy - only fields needed for upgrade finder
		// Skip deep copy and enhancement calculations for massive performance gain
		players := make([]Player, len(data.Players))
		for i, player := range data.Players {
			players[i] = Player{
				// Core identity fields
				UID:      player.UID,
				Name:     player.Name,
				Club:     player.Club,
				Position: player.Position,
				Age:      player.Age,

				// Position matching fields
				ShortPositions:  player.ShortPositions,
				ParsedPositions: player.ParsedPositions,

				// Overall rating fields
				Overall:              player.Overall,
				RoleSpecificOveralls: player.RoleSpecificOveralls,

				// Filter fields
				TransferValue:       player.TransferValue,
				TransferValueAmount: player.TransferValueAmount,
				WageAmount:          player.WageAmount,

				// FIFA Stats (outfield and goalkeeper) - uppercase and lowercase versions
				PAC: player.PAC,
				DRI: player.DRI,
				SHO: player.SHO,
				PAS: player.PAS,
				DEF: player.DEF,
				PHY: player.PHY,
				GK:  player.GK,
				DIV: player.DIV,
				HAN: player.HAN,
				REF: player.REF,
				KIC: player.KIC,
				SPD: player.SPD,
				POS: player.POS,
				Pac: player.Pac,
				Dri: player.Dri,
				Sho: player.Sho,
				Pas: player.Pas,
				Def: player.Def,
				Phy: player.Phy,
				Gk:  player.Gk,
				Div: player.Div,
				Han: player.Han,
				Ref: player.Ref,
				Kic: player.Kic,
				Spd: player.Spd,
				Pos: player.Pos,

				// Additional fields for UI display
				Nationality:    player.Nationality,
				NationalityISO: player.NationalityISO,
				Division:       player.Division,
				Wage:           player.Wage,
				Personality:    player.Personality,
				MediaHandling:  player.MediaHandling,

				// FM Attributes map (contains detailed attributes like crossing, finishing, etc.)
				Attributes:        player.Attributes,
				NumericAttributes: player.NumericAttributes,

				// Performance stats and other display fields
				PerformanceStatsNumeric: player.PerformanceStatsNumeric,
				PerformancePercentiles:  player.PerformancePercentiles,
				TotalStats:              player.TotalStats,
				TotalStatsLower:         player.TotalStatsLower,
				MBR:                     player.MBR,
				Mbr:                     player.Mbr,
				OverallLower:            player.OverallLower,
				BestRoleOverall:         player.BestRoleOverall,
			}
		}

		logDebug(ctx, "Returned lightweight player data for upgrade finder",
			"dataset_id", datasetID,
			"player_count", len(players))

		return players, data.CurrencySymbol, true
	}
	storeMutex.RUnlock()

	// Fallback to persistent storage - but use lightweight loading
	AddSpanEvent(ctx, "store.fallback_to_persistent_upgrade_finder")
	players, currency, err := RetrieveDataset(datasetID)
	if err == nil {
		SetSpanAttributes(ctx,
			attribute.Int("dataset.fallback_player_count", len(players)),
			attribute.String("data.source", "persistent_upgrade_finder"),
		)

		logDebug(ctx, "Returned player data from persistent storage for upgrade finder",
			"dataset_id", datasetID,
			"player_count", len(players))

		return players, currency, true
	}

	SetSpanAttributes(ctx,
		attribute.String("error.type", "dataset_not_found"),
		attribute.String("error.message", err.Error()),
	)

	return nil, "", false
}

// GetPlayerDataForIndexing retrieves player data optimized for search indexing (no deep copy, no enhancements)
func GetPlayerDataForIndexing(datasetID string) ([]Player, string, bool) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "store.get_player_data_for_indexing")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("store.type", "indexing_optimized"),
	)

	// Try fast in-memory cache first
	storeMutex.RLock()
	if data, exists := playerDataStore[datasetID]; exists {
		storeMutex.RUnlock()
		AddSpanEvent(ctx, "store.memory_cache_hit_indexing")
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(data.Players)),
			attribute.String("data.source", "memory_fast_indexing"),
		)

		// Return shallow copy - safe for read-only indexing, much faster than deep copy
		// For search indexing, we only need name, club, position, nationality, division fields
		players := make([]Player, len(data.Players))
		for i, player := range data.Players {
			players[i] = Player{
				UID:         player.UID,
				Name:        player.Name,
				Club:        player.Club,
				Position:    player.Position,
				Nationality: player.Nationality,
				Division:    player.Division,
				Overall:     player.Overall,
				Age:         player.Age,
			}
		}

		logDebug(ctx, "Returned lightweight player data for indexing",
			"dataset_id", datasetID,
			"player_count", len(players))

		return players, data.CurrencySymbol, true
	}
	storeMutex.RUnlock()

	// Fallback to persistent storage - but use lightweight loading
	AddSpanEvent(ctx, "store.fallback_to_persistent_indexing")
	players, currency, err := RetrieveDataset(datasetID)
	if err == nil {
		SetSpanAttributes(ctx,
			attribute.Int("dataset.player_count", len(players)),
			attribute.String("data.source", "persistent_fallback_indexing"),
		)

		// Return lightweight copy for indexing
		lightPlayers := make([]Player, len(players))
		for i, player := range players {
			lightPlayers[i] = Player{
				UID:         player.UID,
				Name:        player.Name,
				Club:        player.Club,
				Position:    player.Position,
				Nationality: player.Nationality,
				Division:    player.Division,
				Overall:     player.Overall,
				Age:         player.Age,
			}
		}

		return lightPlayers, currency, true
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

	// Ensure configuration is loaded before enhancing players
	// This is crucial for calculating role overall ratings and FM attributes
	if err := EnsureConfigInitialized(10 * time.Second); err != nil {
		LogWarn("Configuration initialization timed out during SetPlayerData, proceeding with default weights - error: %v, dataset_id: %s", err, datasetID)
		// Continue with default weights rather than failing the operation
	}

	// Ensure all players have their NumericAttributes populated before storage
	enhancedPlayers := make([]Player, len(players))
	enhancedCount := 0
	for i, player := range players {
		enhancedPlayers[i] = player
		// Always enhance players to ensure TotalStats and other calculated fields are populated
		LogDebug("Enhancing player %s (UID: %d) before legacy storage - NumericAttributes count: %d",
			player.Name, player.UID, len(player.NumericAttributes))
		EnhancePlayerWithCalculations(&enhancedPlayers[i])
		enhancedCount++
		LogDebug("Enhanced player %s (UID: %d) - NumericAttributes count after: %d",
			enhancedPlayers[i].Name, enhancedPlayers[i].UID, len(enhancedPlayers[i].NumericAttributes))
	}

	if enhancedCount > 0 {
		LogDebug("Enhanced %d players before legacy storage for dataset %s", enhancedCount, datasetID)
	}

	// Store in legacy format
	AddSpanEvent(ctx, "store.legacy_store")
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{
		Players:        enhancedPlayers,
		CurrencySymbol: currencySymbol,
	}
	storeMutex.Unlock()

	// Store in new storage system
	AddSpanEvent(ctx, "store.new_storage_store")
	if err := StoreDataset(datasetID, enhancedPlayers, currencySymbol); err != nil {
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

	// Ensure configuration is loaded before enhancing players
	// This is crucial for calculating role overall ratings and FM attributes
	if err := EnsureConfigInitialized(10 * time.Second); err != nil {
		LogWarn("Configuration initialization timed out during SetPlayerDataAsync, proceeding with default weights - error: %v, dataset_id: %s", err, datasetID)
		// Continue with default weights rather than failing the operation
	}

	// Ensure all players have their NumericAttributes populated before storage
	// This is critical because the data is stored before the workers finish enhancement
	enhancedPlayers := make([]Player, len(players))
	enhancedCount := 0
	for i, player := range players {
		enhancedPlayers[i] = player
		// Always enhance players to ensure TotalStats and other calculated fields are populated
		LogDebug("Enhancing player %s (UID: %d) before storage - NumericAttributes count: %d",
			player.Name, player.UID, len(player.NumericAttributes))
		EnhancePlayerWithCalculations(&enhancedPlayers[i])
		enhancedCount++
		LogDebug("Enhanced player %s (UID: %d) - NumericAttributes count after: %d",
			enhancedPlayers[i].Name, enhancedPlayers[i].UID, len(enhancedPlayers[i].NumericAttributes))
	}

	if enhancedCount > 0 {
		LogDebug("Enhanced %d players before storage for dataset %s", enhancedCount, datasetID)
	}

	// Store in legacy format immediately (fast, in-memory operation)
	AddSpanEvent(ctx, "store.legacy_store_immediate")
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{
		Players:        enhancedPlayers,
		CurrencySymbol: currencySymbol,
	}
	storeMutex.Unlock()

	// Store in new storage system asynchronously (potentially slow S3/disk operation)
	AddSpanEvent(ctx, "store.new_storage_async_queued")

	// Copy the enhanced data immediately to avoid races during async storage.
	dataSnapshot := DatasetData{
		Players:        FastDeepCopyPlayers(enhancedPlayers),
		CurrencySymbol: currencySymbol,
	}

	go func() {
		asyncCtx := context.Background()
		asyncCtx, asyncSpan := StartSpan(asyncCtx, "store.new_storage_async_operation")
		defer asyncSpan.End()

		SetSpanAttributes(asyncCtx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.player_count", len(dataSnapshot.Players)),
			attribute.String("operation.type", "async_persistent_storage"),
		)

		startTime := time.Now()

		if err := StoreDataset(datasetID, dataSnapshot.Players, dataSnapshot.CurrencySymbol); err != nil {
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
