package main

import (
	"api/errors"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Cache version - increment when calculation logic changes
const cacheVersion = "1.1"

// NationRatingsCache represents cached nation rating data
type NationRatingsCache struct {
	Version     string         `json:"version"`
	DatasetID   string         `json:"datasetId"`
	GeneratedAt time.Time      `json:"generatedAt"`
	PlayerCount int            `json:"playerCount"`
	NationsData []CachedNation `json:"nationsData"`
}

// CachedNation represents a nation's cached rating data
type CachedNation struct {
	Name                 string `json:"name"`
	NationalityISO       string `json:"nationality_iso"`
	PlayerCount          int    `json:"playerCount"`
	BestFormationOverall int    `json:"bestFormationOverall"`
	AttRating            int    `json:"attRating"`
	MidRating            int    `json:"midRating"`
	DefRating            int    `json:"defRating"`
}

// CacheStorageWrapper wraps the existing storage interface for cache operations
type CacheStorageWrapper struct {
	storage StorageInterface
}

// CreateCacheStorageWrapper creates a new cache storage wrapper
func CreateCacheStorageWrapper(storage StorageInterface) *CacheStorageWrapper {
	return &CacheStorageWrapper{storage: storage}
}

// StoreCacheData stores cache data using the existing storage interface
func (c *CacheStorageWrapper) StoreCacheData(ctx context.Context, cacheKey string, data *NationRatingsCache) error {
	logInfo(ctx, "Starting cache data storage", "cache_key", cacheKey, "player_count", data.PlayerCount)
	start := time.Now()

	// Convert cache data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		logError(ctx, "Failed to marshal cache data", "error", err, "cache_key", cacheKey)
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Store using existing storage interface
	cacheDataset := DatasetData{
		Players:   []Player{},
		CacheData: string(jsonData), // Store JSON data in cache data field
	}

	err = c.storage.Store(cacheKey, cacheDataset)
	if err != nil {
		logError(ctx, "Failed to store cache data", "error", err, "cache_key", cacheKey)
		return err
	}

	logDebug(ctx, "Cache data stored successfully", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

// RetrieveCacheData retrieves cache data using the existing storage interface
func (c *CacheStorageWrapper) RetrieveCacheData(ctx context.Context, cacheKey string) (NationRatingsCache, error) {
	logInfo(ctx, "Starting cache data retrieval", "cache_key", cacheKey)
	start := time.Now()

	cacheDatasetID := fmt.Sprintf("cache_%s", cacheKey)

	data, err := c.storage.Retrieve(cacheDatasetID)
	if err != nil {
		logError(ctx, "Failed to retrieve cache data", "error", err, "cache_key", cacheKey, "cache_dataset_id", cacheDatasetID)
		return NationRatingsCache{}, err
	}

	// Extract cache data from the dummy player
	if len(data.Players) == 0 || data.Players[0].Name != "__CACHE_DATA__" {
		logError(ctx, "Invalid cache data format", "cache_key", cacheKey, "player_count", len(data.Players))
		return NationRatingsCache{}, errors.ErrInvalidCacheDataFormat
	}

	var cacheData NationRatingsCache
	if err := json.Unmarshal([]byte(data.Players[0].Personality), &cacheData); err != nil {
		logError(ctx, "Failed to unmarshal cache data", "error", err, "cache_key", cacheKey)
		return NationRatingsCache{}, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	logDebug(ctx, "Cache data retrieved successfully", "cache_key", cacheKey, "player_count", cacheData.PlayerCount, "duration_ms", time.Since(start).Milliseconds())
	return cacheData, nil
}

// DeleteCacheData deletes cache data
func (c *CacheStorageWrapper) DeleteCacheData(ctx context.Context, cacheKey string) error {
	logInfo(ctx, "Starting cache data deletion", "cache_key", cacheKey)
	start := time.Now()

	err := c.storage.Delete(cacheKey)
	if err != nil {
		logError(ctx, "Failed to delete cache data", "error", err, "cache_key", cacheKey)
		return err
	}

	logDebug(ctx, "Cache data deleted successfully", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

// InitCacheStorage initializes the cache storage system
func InitCacheStorage(ctx context.Context) {
	logDebug(ctx, "Initializing cache storage system", "cache_version", cacheVersion)
	start := time.Now()

	// Cache storage initialization logic would go here

	logDebug(ctx, "Cache storage system initialized successfully",
		"cache_version", cacheVersion,
		"duration_ms", time.Since(start).Milliseconds())
}

// cacheHandler handles cache operations for various cache types
func cacheHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse URL path to determine cache type and operation
	path := strings.TrimPrefix(r.URL.Path, "/api/cache/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 2 {
		http.Error(w, "Invalid cache endpoint format", http.StatusBadRequest)
		return
	}

	cacheType := pathParts[0]
	cacheKey := pathParts[1]

	logDebug(ctx, "Processing cache request",
		"cache_type", cacheType,
		"cache_key", cacheKey,
		"method", r.Method)

	switch cacheType {
	case "nation-ratings":
		handleNationRatingsCache(w, r, cacheKey)
	case "percentiles":
		handlePercentilesCache(w, r, cacheKey)
	case "bargain-hunter":
		handleBargainHunterCache(w, r, cacheKey)
	case "search":
		handleSearchCache(w, r, cacheKey)
	default:
		http.Error(w, "Unknown cache type", http.StatusBadRequest)
	}
}

// handleNationRatingsCache handles nation ratings cache operations
func handleNationRatingsCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	ctx := r.Context()
	cacheDatasetID := fmt.Sprintf("cache_nation_ratings_%s", cacheKey)

	logInfo(ctx, "Handling nation ratings cache operation",
		"method", r.Method,
		"cache_key", cacheKey,
		"cache_dataset_id", cacheDatasetID)

	switch r.Method {
	case http.MethodGet:
		start := time.Now()
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			logError(ctx, "Cache retrieval failed",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		logDebug(ctx, "Cache hit for nation ratings",
			"cache_key", cacheKey,
			"retrieval_duration_ms", time.Since(start).Milliseconds())

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)

		// Use CacheData field, fallback to CurrencySymbol for backward compatibility
		cacheContent := data.CacheData
		if cacheContent == "" {
			cacheContent = data.CurrencySymbol
			logDebug(ctx, "Using fallback CurrencySymbol for cache content", "cache_key", cacheKey)
		}

		if _, err := w.Write([]byte(cacheContent)); err != nil {
			logError(ctx, "Error writing cache response",
				"error", err,
				"cache_key", cacheKey)
		}

	case http.MethodPost:
		start := time.Now()
		var cacheData json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&cacheData); err != nil {
			logError(ctx, "Invalid JSON body in cache POST request",
				"error", err,
				"cache_key", cacheKey)
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		logInfo(ctx, "Storing nation ratings cache",
			"cache_key", cacheKey,
			"data_size_bytes", len(cacheData))

		cacheDataset := DatasetData{
			Players:   []Player{},
			CacheData: string(cacheData),
		}

		if err := storage.Store(cacheDatasetID, cacheDataset); err != nil {
			logError(ctx, "Error storing nation ratings cache",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
			http.Error(w, "Failed to store cache", http.StatusInternalServerError)
			return
		}

		logDebug(ctx, "Nation ratings cache stored successfully",
			"cache_key", cacheKey,
			"storage_duration_ms", time.Since(start).Milliseconds())

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "cached"}); err != nil {
			logError(ctx, "Error encoding cache response",
				"error", err,
				"cache_key", cacheKey)
		}

	case http.MethodDelete:
		start := time.Now()
		if err := storage.Delete(cacheDatasetID); err != nil {
			logError(ctx, "Error deleting nation ratings cache",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
		} else {
			logDebug(ctx, "Nation ratings cache deleted successfully",
				"cache_key", cacheKey,
				"deletion_duration_ms", time.Since(start).Milliseconds())
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			logError(ctx, "Error encoding cache deletion response",
				"error", err,
				"cache_key", cacheKey)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePercentilesCache handles percentiles cache operations
func handlePercentilesCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	ctx := r.Context()
	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	logInfo(ctx, "Handling percentiles cache operation",
		"method", r.Method,
		"cache_key", cacheKey,
		"cache_dataset_id", cacheDatasetID)

	switch r.Method {
	case http.MethodGet:
		start := time.Now()
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			logError(ctx, "Percentiles cache retrieval failed",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		logDebug(ctx, "Cache hit for percentiles",
			"cache_key", cacheKey,
			"retrieval_duration_ms", time.Since(start).Milliseconds())

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)

		// Use CacheData field, fallback to CurrencySymbol for backward compatibility
		cacheContent := data.CacheData
		if cacheContent == "" {
			cacheContent = data.CurrencySymbol
			logDebug(ctx, "Using fallback CurrencySymbol for percentiles cache content", "cache_key", cacheKey)
		}

		if _, err := w.Write([]byte(cacheContent)); err != nil {
			logError(ctx, "Error writing percentiles cache response",
				"error", err,
				"cache_key", cacheKey)
		}

	case http.MethodDelete:
		start := time.Now()
		if err := storage.Delete(cacheDatasetID); err != nil {
			logError(ctx, "Error deleting percentiles cache",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
		} else {
			logDebug(ctx, "Percentiles cache deleted successfully",
				"cache_key", cacheKey,
				"deletion_duration_ms", time.Since(start).Milliseconds())
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			logError(ctx, "Error encoding percentiles cache deletion response",
				"error", err,
				"cache_key", cacheKey)
		}

	default:
		http.Error(w, "Method not allowed for percentiles cache", http.StatusMethodNotAllowed)
	}
}

// handleBargainHunterCache handles bargain hunter cache operations
func handleBargainHunterCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	ctx := r.Context()
	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	logInfo(ctx, "Handling bargain hunter cache operation",
		"method", r.Method,
		"cache_key", cacheKey,
		"cache_dataset_id", cacheDatasetID)

	switch r.Method {
	case http.MethodGet:
		start := time.Now()
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			logError(ctx, "Bargain hunter cache retrieval failed",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		logDebug(ctx, "Cache hit for bargain hunter",
			"cache_key", cacheKey,
			"retrieval_duration_ms", time.Since(start).Milliseconds())

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)

		// Use CacheData field, fallback to CurrencySymbol for backward compatibility
		cacheContent := data.CacheData
		if cacheContent == "" {
			cacheContent = data.CurrencySymbol
			logDebug(ctx, "Using fallback CurrencySymbol for bargain hunter cache content", "cache_key", cacheKey)
		}

		if _, err := w.Write([]byte(cacheContent)); err != nil {
			logError(ctx, "Error writing bargain hunter cache response",
				"error", err,
				"cache_key", cacheKey)
		}

	case http.MethodDelete:
		start := time.Now()
		if err := storage.Delete(cacheDatasetID); err != nil {
			logError(ctx, "Error deleting bargain hunter cache",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
		} else {
			logDebug(ctx, "Bargain hunter cache deleted successfully",
				"cache_key", cacheKey,
				"deletion_duration_ms", time.Since(start).Milliseconds())
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			logError(ctx, "Error encoding bargain hunter cache deletion response",
				"error", err,
				"cache_key", cacheKey)
		}

	default:
		http.Error(w, "Method not allowed for bargain hunter cache", http.StatusMethodNotAllowed)
	}
}

// handleSearchCache handles search cache operations
func handleSearchCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	ctx := r.Context()
	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	logInfo(ctx, "Handling search cache operation",
		"method", r.Method,
		"cache_key", cacheKey,
		"cache_dataset_id", cacheDatasetID)

	switch r.Method {
	case http.MethodGet:
		start := time.Now()
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			logError(ctx, "Search cache retrieval failed",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		logDebug(ctx, "Cache hit for search",
			"cache_key", cacheKey,
			"retrieval_duration_ms", time.Since(start).Milliseconds())

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)

		// Use CacheData field, fallback to CurrencySymbol for backward compatibility
		cacheContent := data.CacheData
		if cacheContent == "" {
			cacheContent = data.CurrencySymbol
			logDebug(ctx, "Using fallback CurrencySymbol for search cache content", "cache_key", cacheKey)
		}

		if _, err := w.Write([]byte(cacheContent)); err != nil {
			logError(ctx, "Error writing search cache response",
				"error", err,
				"cache_key", cacheKey)
		}

	case http.MethodDelete:
		start := time.Now()
		if err := storage.Delete(cacheDatasetID); err != nil {
			logError(ctx, "Error deleting search cache",
				"error", err,
				"cache_dataset_id", cacheDatasetID,
				"cache_key", cacheKey)
		} else {
			logDebug(ctx, "Search cache deleted successfully",
				"cache_key", cacheKey,
				"deletion_duration_ms", time.Since(start).Milliseconds())
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			logError(ctx, "Error encoding search cache deletion response",
				"error", err,
				"cache_key", cacheKey)
		}

	default:
		http.Error(w, "Method not allowed for search cache", http.StatusMethodNotAllowed)
	}
}

// PercentilesCacheKey represents the cache key for percentiles calculation
type PercentilesCacheKey struct {
	DatasetID      string `json:"datasetId"`
	PlayerName     string `json:"playerName"`
	DivisionFilter string `json:"divisionFilter"`
	TargetDivision string `json:"targetDivision"`
	PlayerCount    int    `json:"playerCount"`
	DataHash       string `json:"dataHash"`
}

// PercentilesCacheData represents cached percentiles data
type PercentilesCacheData struct {
	Version     string                        `json:"version"`
	GeneratedAt time.Time                     `json:"generatedAt"`
	CacheKey    PercentilesCacheKey           `json:"cacheKey"`
	Percentiles map[string]map[string]float64 `json:"percentiles"`
}

// generatePercentilesCacheKey generates a cache key for percentiles calculation
func generatePercentilesCacheKey(ctx context.Context, datasetID, playerName, divisionFilter, targetDivision string, players []Player) string {
	logDebug(ctx, "Generating percentiles cache key", "dataset_id", datasetID, "player_name", playerName, "player_count", len(players))
	start := time.Now()

	// Create a simple hash based on key parameters and dataset state
	playerCount := len(players)

	// Create a small hash from player data to detect changes
	samplePlayerData := ""
	if len(players) > 0 {
		// Sample first 10 players' names and overalls
		sampleSize := 10
		if len(players) < sampleSize {
			sampleSize = len(players)
		}
		for i := 0; i < sampleSize; i++ {
			samplePlayerData += fmt.Sprintf("%s:%d;", players[i].Name, players[i].Overall)
		}
	}

	// Simple hash function
	cacheInput := fmt.Sprintf("%s:%s:%s:%s:%d:%s",
		datasetID, playerName, divisionFilter, targetDivision, playerCount, samplePlayerData)

	hash := 0
	for i := 0; i < len(cacheInput); i++ {
		char := int(cacheInput[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash // Convert to 32-bit integer
	}

	cacheKey := fmt.Sprintf("percentiles_%s", fmt.Sprintf("%x", hash)[:12])

	logDebug(ctx, "Generated percentiles cache key", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
	return cacheKey
}

// savePercentilesToCache saves percentiles calculation to cache
func savePercentilesToCache(ctx context.Context, cacheKey, datasetID, playerName, divisionFilter, targetDivision string, players []Player, percentiles map[string]map[string]float64) {
	logInfo(ctx, "Starting percentiles cache save", "cache_key", cacheKey, "dataset_id", datasetID, "player_count", len(players))
	start := time.Now()

	cacheData := PercentilesCacheData{
		Version:     cacheVersion,
		GeneratedAt: time.Now(),
		CacheKey: PercentilesCacheKey{
			DatasetID:      datasetID,
			PlayerName:     playerName,
			DivisionFilter: divisionFilter,
			TargetDivision: targetDivision,
			PlayerCount:    len(players),
			DataHash:       generateDataHash(ctx, players),
		},
		Percentiles: percentiles,
	}

	// Use the existing storage interface through a custom cache storage path
	// We'll create a separate cache dataset ID for this
	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	// Create a cache DatasetData using the dedicated cache field
	cacheDataset := DatasetData{
		Players:   []Player{}, // Empty since we're storing cache data separately
		CacheData: "",         // We'll encode our cache data here as JSON
	}

	// Encode our cache data as JSON and store it in the cache data field
	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		logError(ctx, "Error marshaling percentiles cache data", "error", err, "cache_key", cacheKey)
		return
	}

	cacheDataset.CacheData = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, cacheDataset); err != nil {
		logError(ctx, "Error storing percentiles cache", "error", err, "cache_key", cacheKey, "cache_dataset_id", cacheDatasetID)
		return
	}

	logDebug(ctx, "Percentiles cached successfully", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
}

// loadPercentilesFromCache loads percentiles calculation from cache
func loadPercentilesFromCache(ctx context.Context, cacheKey, datasetID, playerName, divisionFilter, targetDivision string, players []Player) (map[string]map[string]float64, bool) {
	logInfo(ctx, "Starting percentiles cache load", "cache_key", cacheKey, "dataset_id", datasetID, "player_count", len(players))
	start := time.Now()

	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		logDebug(ctx, "Percentiles cache miss", "cache_key", cacheKey, "error", err.Error())
		return nil, false
	}

	// Decode our cache data from the cache data field
	var cacheData PercentilesCacheData
	cacheSource := dummyData.CacheData
	if cacheSource == "" {
		// Fallback to currency symbol for backward compatibility
		cacheSource = dummyData.CurrencySymbol
	}
	if err := json.Unmarshal([]byte(cacheSource), &cacheData); err != nil {
		logError(ctx, "Error unmarshaling percentiles cache data", "error", err, "cache_key", cacheKey)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		logDebug(ctx, "Percentiles cache version mismatch, recalculating", "cache_version", cacheData.Version, "expected_version", cacheVersion)
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.PlayerName != playerName ||
		cacheData.CacheKey.DivisionFilter != divisionFilter ||
		cacheData.CacheKey.TargetDivision != targetDivision {
		logDebug(ctx, "Percentiles cache key mismatch, recalculating", "cache_key", cacheKey)
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		logDebug(ctx, "Player count changed, recalculating percentiles",
			"cached_count", cacheData.CacheKey.PlayerCount,
			"current_count", len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(ctx, players) {
		logDebug(ctx, "Dataset hash changed, recalculating percentiles", "cache_key", cacheKey)
		return nil, false
	}

	logDebug(ctx, "Loaded percentiles from cache",
		"cache_key", cacheKey,
		"generated_at", cacheData.GeneratedAt.Format(time.RFC3339),
		"duration_ms", time.Since(start).Milliseconds())
	return cacheData.Percentiles, true
}

// generateDataHash creates a simple hash from player data to detect changes
func generateDataHash(ctx context.Context, players []Player) string {
	logDebug(ctx, "Generating data hash", "player_count", len(players))
	start := time.Now()

	if len(players) == 0 {
		logDebug(ctx, "Empty player list, returning empty hash", "duration_ms", time.Since(start).Milliseconds())
		return ""
	}

	// Sample first and last 5 players for efficiency
	sampleData := ""
	sampleSize := 5

	// First 5 players
	for i := 0; i < sampleSize && i < len(players); i++ {
		sampleData += fmt.Sprintf("%s:%d:%s;", players[i].Name, players[i].Overall, players[i].Club)
	}

	// Last 5 players (if dataset is large enough)
	if len(players) > sampleSize*2 {
		for i := len(players) - sampleSize; i < len(players); i++ {
			sampleData += fmt.Sprintf("%s:%d:%s;", players[i].Name, players[i].Overall, players[i].Club)
		}
	}

	// Simple hash
	hash := 0
	for i := 0; i < len(sampleData); i++ {
		char := int(sampleData[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash
	}

	hashResult := fmt.Sprintf("%x", hash)[:8]
	logDebug(ctx, "Generated data hash", "hash", hashResult, "duration_ms", time.Since(start).Milliseconds())
	return hashResult
}

// BargainHunterCacheKey represents the cache key for bargain hunter calculation
type BargainHunterCacheKey struct {
	DatasetID   string `json:"datasetId"`
	MaxBudget   int64  `json:"maxBudget"`
	MaxSalary   int64  `json:"maxSalary"`
	MinAge      int    `json:"minAge"`
	MaxAge      int    `json:"maxAge"`
	MinOverall  int    `json:"minOverall"`
	PlayerCount int    `json:"playerCount"`
	DataHash    string `json:"dataHash"`
}

// BargainHunterCacheData represents cached bargain hunter data
type BargainHunterCacheData struct {
	Version     string                  `json:"version"`
	GeneratedAt time.Time               `json:"generatedAt"`
	CacheKey    BargainHunterCacheKey   `json:"cacheKey"`
	Results     []BargainHunterResponse `json:"results"`
}

// generateBargainHunterCacheKey generates a cache key for bargain hunter calculation
func generateBargainHunterCacheKey(ctx context.Context, datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player) string {
	logDebug(ctx, "Generating bargain hunter cache key", "dataset_id", datasetID, "player_count", len(players), "max_budget", maxBudget)
	start := time.Now()

	playerCount := len(players)
	dataHash := generateDataHash(ctx, players)

	// Simple hash function
	cacheInput := fmt.Sprintf("%s:%d:%d:%d:%d:%d:%d:%s",
		datasetID, maxBudget, maxSalary, minAge, maxAge, minOverall, playerCount, dataHash)

	hash := 0
	for i := 0; i < len(cacheInput); i++ {
		char := int(cacheInput[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash
	}

	cacheKey := fmt.Sprintf("bargain_hunter_%s", fmt.Sprintf("%x", hash)[:12])

	logDebug(ctx, "Generated bargain hunter cache key", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
	return cacheKey
}

// saveBargainHunterToCache saves bargain hunter calculation to cache
func saveBargainHunterToCache(ctx context.Context, cacheKey, datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player, results []BargainHunterResponse) {
	logInfo(ctx, "Starting bargain hunter cache save", "cache_key", cacheKey, "dataset_id", datasetID, "player_count", len(players), "results_count", len(results))
	start := time.Now()

	cacheData := BargainHunterCacheData{
		Version:     cacheVersion,
		GeneratedAt: time.Now(),
		CacheKey: BargainHunterCacheKey{
			DatasetID:   datasetID,
			MaxBudget:   maxBudget,
			MaxSalary:   maxSalary,
			MinAge:      minAge,
			MaxAge:      maxAge,
			MinOverall:  minOverall,
			PlayerCount: len(players),
			DataHash:    generateDataHash(ctx, players),
		},
		Results: results,
	}

	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	cacheDataset := DatasetData{
		Players:   []Player{},
		CacheData: "",
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		logError(ctx, "Error marshaling bargain hunter cache data", "error", err, "cache_key", cacheKey)
		return
	}

	cacheDataset.CacheData = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, cacheDataset); err != nil {
		logError(ctx, "Error storing bargain hunter cache", "error", err, "cache_key", cacheKey, "cache_dataset_id", cacheDatasetID)
		return
	}

	logDebug(ctx, "Bargain hunter results cached successfully", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
}

// loadBargainHunterFromCache loads bargain hunter calculation from cache
func loadBargainHunterFromCache(ctx context.Context, cacheKey, datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player) ([]BargainHunterResponse, bool) {
	logInfo(ctx, "Starting bargain hunter cache load", "cache_key", cacheKey, "dataset_id", datasetID, "player_count", len(players))
	start := time.Now()

	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		logDebug(ctx, "Bargain hunter cache miss", "cache_key", cacheKey, "error", err.Error())
		return nil, false
	}

	var cacheData BargainHunterCacheData
	cacheSource := dummyData.CacheData
	if cacheSource == "" {
		// Fallback to currency symbol for backward compatibility
		cacheSource = dummyData.CurrencySymbol
	}
	if err := json.Unmarshal([]byte(cacheSource), &cacheData); err != nil {
		logError(ctx, "Error unmarshaling bargain hunter cache data", "error", err, "cache_key", cacheKey)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		logDebug(ctx, "Bargain hunter cache version mismatch, recalculating", "cache_version", cacheData.Version, "expected_version", cacheVersion)
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.MaxBudget != maxBudget ||
		cacheData.CacheKey.MaxSalary != maxSalary ||
		cacheData.CacheKey.MinAge != minAge ||
		cacheData.CacheKey.MaxAge != maxAge ||
		cacheData.CacheKey.MinOverall != minOverall {
		logDebug(ctx, "Bargain hunter cache key mismatch, recalculating", "cache_key", cacheKey)
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		logDebug(ctx, "Player count changed, recalculating bargain hunter",
			"cached_count", cacheData.CacheKey.PlayerCount,
			"current_count", len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(ctx, players) {
		logDebug(ctx, "Dataset hash changed, recalculating bargain hunter", "cache_key", cacheKey)
		return nil, false
	}

	logDebug(ctx, "Loaded bargain hunter results from cache",
		"cache_key", cacheKey,
		"generated_at", cacheData.GeneratedAt.Format(time.RFC3339),
		"duration_ms", time.Since(start).Milliseconds())
	return cacheData.Results, true
}

// SearchCacheKey represents the cache key for search calculation
type SearchCacheKey struct {
	DatasetID   string `json:"datasetId"`
	Query       string `json:"query"`
	PlayerCount int    `json:"playerCount"`
	DataHash    string `json:"dataHash"`
}

// SearchCacheData represents cached search data
type SearchCacheData struct {
	Version     string         `json:"version"`
	GeneratedAt time.Time      `json:"generatedAt"`
	CacheKey    SearchCacheKey `json:"cacheKey"`
	Results     []SearchResult `json:"results"`
}

// generateSearchCacheKey generates a cache key for search calculation
func generateSearchCacheKey(ctx context.Context, datasetID, query string, players []Player) string {
	logDebug(ctx, "Generating search cache key", "dataset_id", datasetID, "query", query, "player_count", len(players))
	start := time.Now()

	playerCount := len(players)
	dataHash := generateDataHash(ctx, players)

	// Simple hash function
	cacheInput := fmt.Sprintf("%s:%s:%d:%s",
		datasetID, strings.ToLower(strings.TrimSpace(query)), playerCount, dataHash)

	hash := 0
	for i := 0; i < len(cacheInput); i++ {
		char := int(cacheInput[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash
	}

	cacheKey := fmt.Sprintf("search_%s", fmt.Sprintf("%x", hash)[:12])

	logDebug(ctx, "Generated search cache key", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
	return cacheKey
}

// saveSearchToCache saves search results to cache
func saveSearchToCache(ctx context.Context, cacheKey, datasetID, query string, players []Player, results []SearchResult) {
	logInfo(ctx, "Starting search cache save", "cache_key", cacheKey, "dataset_id", datasetID, "query", query, "player_count", len(players), "results_count", len(results))
	start := time.Now()

	cacheData := SearchCacheData{
		Version:     cacheVersion,
		GeneratedAt: time.Now(),
		CacheKey: SearchCacheKey{
			DatasetID:   datasetID,
			Query:       query,
			PlayerCount: len(players),
			DataHash:    generateDataHash(ctx, players),
		},
		Results: results,
	}

	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	cacheDataset := DatasetData{
		Players:   []Player{},
		CacheData: "",
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		logError(ctx, "Error marshaling search cache data", "error", err, "cache_key", cacheKey)
		return
	}

	cacheDataset.CacheData = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, cacheDataset); err != nil {
		logError(ctx, "Error storing search cache", "error", err, "cache_key", cacheKey, "cache_dataset_id", cacheDatasetID)
		return
	}

	logDebug(ctx, "Search results cached successfully", "cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())
}

// loadSearchFromCache loads search results from cache
func loadSearchFromCache(ctx context.Context, cacheKey, datasetID, query string, players []Player) ([]SearchResult, bool) {
	logInfo(ctx, "Starting search cache load", "cache_key", cacheKey, "dataset_id", datasetID, "query", query, "player_count", len(players))
	start := time.Now()

	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		logDebug(ctx, "Search cache miss", "cache_key", cacheKey, "error", err.Error())
		return nil, false
	}

	var cacheData SearchCacheData
	cacheSource := dummyData.CacheData
	if cacheSource == "" {
		// Fallback to currency symbol for backward compatibility
		cacheSource = dummyData.CurrencySymbol
	}
	if err := json.Unmarshal([]byte(cacheSource), &cacheData); err != nil {
		logError(ctx, "Error unmarshaling search cache data", "error", err, "cache_key", cacheKey)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		logDebug(ctx, "Search cache version mismatch, recalculating", "cache_version", cacheData.Version, "expected_version", cacheVersion)
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.Query != query {
		logDebug(ctx, "Search cache key mismatch, recalculating", "cache_key", cacheKey)
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		logDebug(ctx, "Player count changed, recalculating search",
			"cached_count", cacheData.CacheKey.PlayerCount,
			"current_count", len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(ctx, players) {
		logDebug(ctx, "Dataset hash changed, recalculating search", "cache_key", cacheKey)
		return nil, false
	}

	logDebug(ctx, "Loaded search results from cache",
		"cache_key", cacheKey,
		"generated_at", cacheData.GeneratedAt.Format(time.RFC3339),
		"duration_ms", time.Since(start).Milliseconds())
	return cacheData.Results, true
}
