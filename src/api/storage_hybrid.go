package main

import (
	"fmt"
	"time"
)

// HybridStorage combines in-memory storage with local file fallback
type HybridStorage struct {
	memory StorageInterface
	local  StorageInterface
}

// CreateHybridStorage creates a new hybrid storage instance
func CreateHybridStorage(datasetDir string) (*HybridStorage, error) {
	memory := CreateInMemoryStorage()
	local, err := CreateLocalFileStorage(datasetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file storage: %w", err)
	}

	LogInfo("Initialized hybrid storage (in-memory + local file fallback)")
	return &HybridStorage{
		memory: memory,
		local:  local,
	}, nil
}

// Store saves a dataset using hybrid storage (memory + local file)
func (s *HybridStorage) Store(datasetID string, data DatasetData) error {
	// Store in both memory and local file
	if err := s.memory.Store(datasetID, data); err != nil {
		LogWarn("Warning: Failed to store dataset %s in memory: %v", sanitizeForLogging(datasetID), err)
	}

	return s.local.Store(datasetID, data)
}

// Retrieve retrieves a dataset from hybrid storage
func (s *HybridStorage) Retrieve(datasetID string) (DatasetData, error) {
	// Try memory first for fastest access
	data, err := s.memory.Retrieve(datasetID)
	if err == nil {
		LogDebug("Retrieved dataset %s from memory", sanitizeForLogging(datasetID))
		return data, nil
	}

	// Try persistent storage (critical for multi-replica consistency)
	LogDebug("Dataset %s not found in memory, checking persistent storage", sanitizeForLogging(datasetID))
	data, err = s.local.Retrieve(datasetID)
	if err != nil {
		return DatasetData{}, err
	}

	// Store in memory for future access (warm up the cache).
	// The caller receives the same underlying []Player array that came off disk.
	// If the goroutine stores a reference to that same array, InMemoryStorage.Store
	// will share the backing array with the returned value.  Any concurrent mutation
	// on either side (filters applied by handlers, future in-place operations in the
	// memory store) would then be an undetected data race.  Take an explicit copy of
	// the slice so the goroutine owns its own independent backing array.
	cacheData := DatasetData{
		CurrencySymbol: data.CurrencySymbol,
		CacheData:      data.CacheData,
		RawBytes:       data.RawBytes,
		Players:        make([]Player, len(data.Players)),
	}
	copy(cacheData.Players, data.Players)
	go func() {
		if storeErr := s.memory.Store(datasetID, cacheData); storeErr != nil {
			LogWarn("Warning: Failed to cache dataset %s in memory: %v", sanitizeForLogging(datasetID), storeErr)
		} else {
			LogDebug("Successfully cached dataset %s in memory for future access", sanitizeForLogging(datasetID))
		}
	}()

	LogDebug("Retrieved dataset %s from persistent storage and cached in memory", sanitizeForLogging(datasetID))
	return data, nil
}

// Delete removes a dataset from hybrid storage
func (s *HybridStorage) Delete(datasetID string) error {
	// Delete from both memory and local file
	if err := s.memory.Delete(datasetID); err != nil {
		// Ignore error since it might not exist in memory
		LogDebug("Note: Dataset %s not found in memory during deletion", sanitizeForLogging(datasetID))
	}
	return s.local.Delete(datasetID)
}

// List returns all dataset IDs from hybrid storage
func (s *HybridStorage) List() ([]string, error) {
	// Get datasets from both memory and local storage, then merge
	memoryIDs, _ := s.memory.List() // Ignore error since memory might be empty
	localIDs, err := s.local.List()
	if err != nil {
		return memoryIDs, err
	}

	// Merge and deduplicate
	idSet := make(map[string]bool)
	for _, id := range memoryIDs {
		idSet[id] = true
	}
	for _, id := range localIDs {
		idSet[id] = true
	}

	var allIDs []string
	for id := range idSet {
		allIDs = append(allIDs, id)
	}

	return allIDs, nil
}

// CleanupOldDatasets removes datasets older than the specified duration from hybrid storage
func (s *HybridStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	// Cleanup both memory and local storage
	if err := s.memory.CleanupOldDatasets(maxAge, excludeDatasets); err != nil {
		LogWarn("Warning: Memory cleanup failed: %v", err)
		// Continue with local cleanup even if memory cleanup fails
	}
	return s.local.CleanupOldDatasets(maxAge, excludeDatasets)
}
