package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FastSearchService provides optimized search functionality using pre-built indexes
type FastSearchService struct {
	index          *SearchIndex
	indexedDataset string
	indexMutex     sync.RWMutex
	lastIndexTime  time.Time
}

// NewFastSearchService creates a new fast search service
func NewFastSearchService() *FastSearchService {
	return &FastSearchService{
		index: NewSearchIndex(),
	}
}

// Search performs optimized search using pre-built indexes
func (fss *FastSearchService) Search(ctx context.Context, datasetID, query string, maxResults int) ([]SearchResult, error) {
	// Ensure index is built for the dataset
	if err := fss.ensureIndexBuilt(ctx, datasetID); err != nil {
		return nil, err
	}

	// Perform fast search
	return fss.index.FastSearch(query, maxResults), nil
}

// ensureIndexBuilt ensures the search index is built for the given dataset
func (fss *FastSearchService) ensureIndexBuilt(ctx context.Context, datasetID string) error {
	fss.indexMutex.RLock()
	if fss.indexedDataset == datasetID && !fss.lastIndexTime.IsZero() && time.Since(fss.lastIndexTime) < 30*time.Minute {
		// Index is fresh and valid
		fss.indexMutex.RUnlock()
		return nil
	}
	fss.indexMutex.RUnlock()

	// Need to rebuild index
	fss.indexMutex.Lock()
	defer fss.indexMutex.Unlock()

	// Double-check after acquiring write lock
	if fss.indexedDataset == datasetID && !fss.lastIndexTime.IsZero() && time.Since(fss.lastIndexTime) < 30*time.Minute {
		return nil
	}

	// Get player data (lightweight for search indexing - no deep copy or enhancements needed)
	dataStart := time.Now()
	players, _, found := GetPlayerDataForIndexing(datasetID)
	dataLoadTime := time.Since(dataStart)
	if !found {
		return fmt.Errorf("dataset not found: %s", datasetID)
	}

	logInfo(ctx, "Player data loaded for search index",
		"dataset_id", datasetID,
		"player_count", len(players),
		"data_load_time_ms", dataLoadTime.Milliseconds())

	// Build index
	startTime := time.Now()
	fss.index.BuildIndex(players)
	fss.indexedDataset = datasetID
	fss.lastIndexTime = time.Now()

	// Log index build time
	buildTime := time.Since(startTime)
	stats := fss.index.GetIndexStats()

	logInfo(ctx, "Search index built",
		"dataset_id", datasetID,
		"build_time_ms", buildTime.Milliseconds(),
		"players", stats["players"],
		"teams", stats["teams"],
		"leagues", stats["leagues"],
		"nations", stats["nations"],
		"total_terms", stats["name_terms"].(int)+stats["club_terms"].(int)+
			stats["position_terms"].(int)+stats["nationality_terms"].(int)+
			stats["division_terms"].(int))

	return nil
}

// InvalidateIndex forces the index to be rebuilt on next search
func (fss *FastSearchService) InvalidateIndex(datasetID string) {
	fss.indexMutex.Lock()
	defer fss.indexMutex.Unlock()

	if fss.indexedDataset == datasetID {
		fss.indexedDataset = ""
		fss.lastIndexTime = time.Time{}
	}
}

// GetIndexStats returns statistics about the current search index
func (fss *FastSearchService) GetIndexStats() map[string]interface{} {
	fss.indexMutex.RLock()
	defer fss.indexMutex.RUnlock()

	stats := fss.index.GetIndexStats()
	stats["indexed_dataset"] = fss.indexedDataset
	stats["last_index_time"] = fss.lastIndexTime

	return stats
}

// Global fast search service instance
var globalFastSearchService = NewFastSearchService()

// GetFastSearchService returns the global fast search service
func GetFastSearchService() *FastSearchService {
	return globalFastSearchService
}
