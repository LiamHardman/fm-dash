package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func NewCacheStorageWrapper(storage StorageInterface) *CacheStorageWrapper {
	return &CacheStorageWrapper{storage: storage}
}

// StoreCacheData stores cache data using the existing storage interface
func (c *CacheStorageWrapper) StoreCacheData(cacheKey string, data *NationRatingsCache) error {
	// Convert cache data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Store using existing storage interface
	dummyData := DatasetData{
		Players:        []Player{},
		CurrencySymbol: string(jsonData), // Store JSON data in currency symbol field
	}

	return c.storage.Store(cacheKey, dummyData)
}

// RetrieveCacheData retrieves cache data using the existing storage interface
func (c *CacheStorageWrapper) RetrieveCacheData(cacheKey string) (NationRatingsCache, error) {
	cacheDatasetID := fmt.Sprintf("cache_%s", cacheKey)

	data, err := c.storage.Retrieve(cacheDatasetID)
	if err != nil {
		return NationRatingsCache{}, err
	}

	// Extract cache data from the dummy player
	if len(data.Players) == 0 || data.Players[0].Name != "__CACHE_DATA__" {
		return NationRatingsCache{}, fmt.Errorf("invalid cache data format")
	}

	var cacheData NationRatingsCache
	if err := json.Unmarshal([]byte(data.Players[0].Personality), &cacheData); err != nil {
		return NationRatingsCache{}, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return cacheData, nil
}

// DeleteCacheData deletes cache data
func (c *CacheStorageWrapper) DeleteCacheData(cacheKey string) error {
	return c.storage.Delete(cacheKey)
}

func InitCacheStorage() {
	log.Println("Cache storage system initialized")
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

	logInfo(ctx, "Processing cache request",
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
	cacheDatasetID := fmt.Sprintf("cache_nation_ratings_%s", cacheKey)

	switch r.Method {
	case http.MethodGet:
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if _, err := w.Write([]byte(data.CurrencySymbol)); err != nil {
			log.Printf("Error writing response: %v", err)
		}

	case http.MethodPost:
		var cacheData json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&cacheData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		dummyData := DatasetData{
			Players:        []Player{},
			CurrencySymbol: string(cacheData),
		}

		if err := storage.Store(cacheDatasetID, dummyData); err != nil {
			log.Printf("Error storing nation ratings cache: %v", err)
			http.Error(w, "Failed to store cache", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "cached"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}

	case http.MethodDelete:
		if err := storage.Delete(cacheDatasetID); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePercentilesCache handles percentiles cache operations
func handlePercentilesCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	switch r.Method {
	case http.MethodGet:
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if _, err := w.Write([]byte(data.CurrencySymbol)); err != nil {
			log.Printf("Error writing response: %v", err)
		}

	case http.MethodDelete:
		if err := storage.Delete(cacheDatasetID); err != nil {
			log.Printf("Error deleting percentiles cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}

	default:
		http.Error(w, "Method not allowed for percentiles cache", http.StatusMethodNotAllowed)
	}
}

// handleBargainHunterCache handles bargain hunter cache operations
func handleBargainHunterCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	switch r.Method {
	case http.MethodGet:
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if _, err := w.Write([]byte(data.CurrencySymbol)); err != nil {
			log.Printf("Error writing response: %v", err)
		}

	case http.MethodDelete:
		if err := storage.Delete(cacheDatasetID); err != nil {
			log.Printf("Error deleting bargain hunter cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}

	default:
		http.Error(w, "Method not allowed for bargain hunter cache", http.StatusMethodNotAllowed)
	}
}

// handleSearchCache handles search cache operations
func handleSearchCache(w http.ResponseWriter, r *http.Request, cacheKey string) {
	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	switch r.Method {
	case http.MethodGet:
		data, err := storage.Retrieve(cacheDatasetID)
		if err != nil {
			http.Error(w, "Cache not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if _, err := w.Write([]byte(data.CurrencySymbol)); err != nil {
			log.Printf("Error writing response: %v", err)
		}

	case http.MethodDelete:
		if err := storage.Delete(cacheDatasetID); err != nil {
			log.Printf("Error deleting search cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
			log.Printf("Error encoding response: %v", err)
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
func generatePercentilesCacheKey(datasetID, playerName, divisionFilter, targetDivision string, players []Player) string {
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

	return cacheKey
}

// savePercentilesToCache saves percentiles calculation to cache
func savePercentilesToCache(cacheKey, datasetID, playerName, divisionFilter, targetDivision string, players []Player, percentiles map[string]map[string]float64) {
	cacheData := PercentilesCacheData{
		Version:     cacheVersion,
		GeneratedAt: time.Now(),
		CacheKey: PercentilesCacheKey{
			DatasetID:      datasetID,
			PlayerName:     playerName,
			DivisionFilter: divisionFilter,
			TargetDivision: targetDivision,
			PlayerCount:    len(players),
			DataHash:       generateDataHash(players),
		},
		Percentiles: percentiles,
	}

	// Use the existing storage interface through a custom cache storage path
	// We'll create a separate cache dataset ID for this
	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	// Create a dummy DatasetData to work with existing storage interface
	dummyData := DatasetData{
		Players:        []Player{}, // Empty since we're storing cache data in CurrencySymbol
		CurrencySymbol: "",         // We'll encode our cache data here as JSON
	}

	// Encode our cache data as JSON and store it in the currency symbol field
	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		log.Printf("⚠️ Error marshaling percentiles cache data: %v", err)
		return
	}

	dummyData.CurrencySymbol = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, dummyData); err != nil {
		log.Printf("⚠️ Error storing percentiles cache: %v", err)
		return
	}

	log.Printf("✅ Percentiles cached successfully as %s", cacheKey)
}

// loadPercentilesFromCache loads percentiles calculation from cache
func loadPercentilesFromCache(cacheKey, datasetID, playerName, divisionFilter, targetDivision string, players []Player) (map[string]map[string]float64, bool) {
	cacheDatasetID := fmt.Sprintf("cache_percentiles_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		return nil, false
	}

	// Decode our cache data from the currency symbol field
	var cacheData PercentilesCacheData
	if err := json.Unmarshal([]byte(dummyData.CurrencySymbol), &cacheData); err != nil {
		log.Printf("⚠️ Error unmarshaling percentiles cache data: %v", err)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		log.Printf("♻️ Percentiles cache version mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.PlayerName != playerName ||
		cacheData.CacheKey.DivisionFilter != divisionFilter ||
		cacheData.CacheKey.TargetDivision != targetDivision {
		log.Printf("♻️ Percentiles cache key mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		log.Printf("♻️ Player count changed (%d vs %d), recalculating percentiles...",
			cacheData.CacheKey.PlayerCount, len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(players) {
		log.Printf("♻️ Dataset hash changed, recalculating percentiles...")
		return nil, false
	}

	log.Printf("✅ Loaded percentiles from cache (generated %s)", cacheData.GeneratedAt.Format(time.RFC3339))
	return cacheData.Percentiles, true
}

// generateDataHash creates a simple hash from player data to detect changes
func generateDataHash(players []Player) string {
	if len(players) == 0 {
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

	return fmt.Sprintf("%x", hash)[:8]
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
func generateBargainHunterCacheKey(datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player) string {
	playerCount := len(players)
	dataHash := generateDataHash(players)

	// Simple hash function
	cacheInput := fmt.Sprintf("%s:%d:%d:%d:%d:%d:%d:%s",
		datasetID, maxBudget, maxSalary, minAge, maxAge, minOverall, playerCount, dataHash)

	hash := 0
	for i := 0; i < len(cacheInput); i++ {
		char := int(cacheInput[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash
	}

	return fmt.Sprintf("bargain_hunter_%s", fmt.Sprintf("%x", hash)[:12])
}

// saveBargainHunterToCache saves bargain hunter calculation to cache
func saveBargainHunterToCache(cacheKey, datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player, results []BargainHunterResponse) {
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
			DataHash:    generateDataHash(players),
		},
		Results: results,
	}

	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	dummyData := DatasetData{
		Players:        []Player{},
		CurrencySymbol: "",
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		log.Printf("⚠️ Error marshaling bargain hunter cache data: %v", err)
		return
	}

	dummyData.CurrencySymbol = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, dummyData); err != nil {
		log.Printf("⚠️ Error storing bargain hunter cache: %v", err)
		return
	}

	log.Printf("✅ Bargain hunter results cached successfully as %s", cacheKey)
}

// loadBargainHunterFromCache loads bargain hunter calculation from cache
func loadBargainHunterFromCache(cacheKey, datasetID string, maxBudget, maxSalary int64, minAge, maxAge, minOverall int, players []Player) ([]BargainHunterResponse, bool) {
	cacheDatasetID := fmt.Sprintf("cache_bargain_hunter_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		return nil, false
	}

	var cacheData BargainHunterCacheData
	if err := json.Unmarshal([]byte(dummyData.CurrencySymbol), &cacheData); err != nil {
		log.Printf("⚠️ Error unmarshaling bargain hunter cache data: %v", err)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		log.Printf("♻️ Bargain hunter cache version mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.MaxBudget != maxBudget ||
		cacheData.CacheKey.MaxSalary != maxSalary ||
		cacheData.CacheKey.MinAge != minAge ||
		cacheData.CacheKey.MaxAge != maxAge ||
		cacheData.CacheKey.MinOverall != minOverall {
		log.Printf("♻️ Bargain hunter cache key mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		log.Printf("♻️ Player count changed (%d vs %d), recalculating bargain hunter...",
			cacheData.CacheKey.PlayerCount, len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(players) {
		log.Printf("♻️ Dataset hash changed, recalculating bargain hunter...")
		return nil, false
	}

	log.Printf("✅ Loaded bargain hunter results from cache (generated %s)", cacheData.GeneratedAt.Format(time.RFC3339))
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
func generateSearchCacheKey(datasetID, query string, players []Player) string {
	playerCount := len(players)
	dataHash := generateDataHash(players)

	// Simple hash function
	cacheInput := fmt.Sprintf("%s:%s:%d:%s",
		datasetID, strings.ToLower(strings.TrimSpace(query)), playerCount, dataHash)

	hash := 0
	for i := 0; i < len(cacheInput); i++ {
		char := int(cacheInput[i])
		hash = ((hash << 5) - hash) + char
		hash &= hash
	}

	return fmt.Sprintf("search_%s", fmt.Sprintf("%x", hash)[:12])
}

// saveSearchToCache saves search results to cache
func saveSearchToCache(cacheKey, datasetID, query string, players []Player, results []SearchResult) {
	cacheData := SearchCacheData{
		Version:     cacheVersion,
		GeneratedAt: time.Now(),
		CacheKey: SearchCacheKey{
			DatasetID:   datasetID,
			Query:       query,
			PlayerCount: len(players),
			DataHash:    generateDataHash(players),
		},
		Results: results,
	}

	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	dummyData := DatasetData{
		Players:        []Player{},
		CurrencySymbol: "",
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		log.Printf("⚠️ Error marshaling search cache data: %v", err)
		return
	}

	dummyData.CurrencySymbol = string(cacheJSON)

	if err := storage.Store(cacheDatasetID, dummyData); err != nil {
		log.Printf("⚠️ Error storing search cache: %v", err)
		return
	}

	log.Printf("✅ Search results cached successfully as %s", cacheKey)
}

// loadSearchFromCache loads search results from cache
func loadSearchFromCache(cacheKey, datasetID, query string, players []Player) ([]SearchResult, bool) {
	cacheDatasetID := fmt.Sprintf("cache_search_%s", cacheKey)

	dummyData, err := storage.Retrieve(cacheDatasetID)
	if err != nil {
		return nil, false
	}

	var cacheData SearchCacheData
	if err := json.Unmarshal([]byte(dummyData.CurrencySymbol), &cacheData); err != nil {
		log.Printf("⚠️ Error unmarshaling search cache data: %v", err)
		return nil, false
	}

	// Validate cache data
	if cacheData.Version != cacheVersion {
		log.Printf("♻️ Search cache version mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.DatasetID != datasetID ||
		cacheData.CacheKey.Query != query {
		log.Printf("♻️ Search cache key mismatch, recalculating...")
		return nil, false
	}

	if cacheData.CacheKey.PlayerCount != len(players) {
		log.Printf("♻️ Player count changed (%d vs %d), recalculating search...",
			cacheData.CacheKey.PlayerCount, len(players))
		return nil, false
	}

	if cacheData.CacheKey.DataHash != generateDataHash(players) {
		log.Printf("♻️ Dataset hash changed, recalculating search...")
		return nil, false
	}

	log.Printf("✅ Loaded search results from cache (generated %s)", cacheData.GeneratedAt.Format(time.RFC3339))
	return cacheData.Results, true
}
