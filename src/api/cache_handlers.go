// cache_handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

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

// NewCacheStorageWrapper creates a new cache storage wrapper
func NewCacheStorageWrapper(storage StorageInterface) *CacheStorageWrapper {
	return &CacheStorageWrapper{storage: storage}
}

// StoreCacheData stores cache data using the existing storage interface
func (c *CacheStorageWrapper) StoreCacheData(cacheKey string, data NationRatingsCache) error {
	// Convert cache data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Store as a special "cache" dataset using the existing storage infrastructure
	// We'll use a dummy player with the cache data stored in Personality field (which is a string)
	cacheDatasetID := fmt.Sprintf("cache_%s", cacheKey)

	dummyPlayer := Player{
		Name:        "__CACHE_DATA__",
		Personality: string(jsonData), // Store JSON as string in Personality field
		Overall:     0,
		Age:         "0",
	}

	cacheDataset := DatasetData{
		Players:        []Player{dummyPlayer},
		CurrencySymbol: "",
	}

	return c.storage.Store(cacheDatasetID, cacheDataset)
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
	cacheDatasetID := fmt.Sprintf("cache_%s", cacheKey)
	return c.storage.Delete(cacheDatasetID)
}

// Global cache storage wrapper
var cacheStorage *CacheStorageWrapper

// InitCacheStorage initializes the cache storage wrapper
func InitCacheStorage() {
	cacheStorage = NewCacheStorageWrapper(storage)
	log.Println("Initialized nation ratings cache storage")
}

// extractCacheKeyFromPath extracts cache key from URL path
func extractCacheKeyFromPath(path string) string {
	// Expected path: /api/cache/nation-ratings/{cacheKey}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "cache" && parts[2] == "nation-ratings" {
		return parts[3]
	}
	return ""
}

// handleStoreNationRatingsCache handles POST /api/cache/nation-ratings/{cacheKey}
func handleStoreNationRatingsCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "cache.store_nation_ratings")
	defer span.End()

	cacheKey := extractCacheKeyFromPath(r.URL.Path)
	if cacheKey == "" {
		http.Error(w, "Invalid cache key in URL", http.StatusBadRequest)
		return
	}

	// Validate cache key format
	if !strings.HasPrefix(cacheKey, "nation_ratings_") {
		http.Error(w, "Invalid cache key format", http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("cache.key", cacheKey),
		attribute.String("cache.operation", "store"),
	)

	var cacheData NationRatingsCache
	if err := json.NewDecoder(r.Body).Decode(&cacheData); err != nil {
		RecordError(ctx, err, "Failed to decode cache data")
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate cache data
	if cacheData.Version == "" || cacheData.DatasetID == "" || len(cacheData.NationsData) == 0 {
		http.Error(w, "Invalid cache data: missing required fields", http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("cache.version", cacheData.Version),
		attribute.String("cache.dataset_id", cacheData.DatasetID),
		attribute.Int("cache.nations_count", len(cacheData.NationsData)),
		attribute.Int("cache.player_count", cacheData.PlayerCount),
	)

	// Store cache data
	if err := cacheStorage.StoreCacheData(cacheKey, cacheData); err != nil {
		RecordError(ctx, err, "Failed to store cache data")
		http.Error(w, "Failed to store cache data", http.StatusInternalServerError)
		return
	}

	SetSpanAttributes(ctx, attribute.String("cache.status", "stored"))
	log.Printf("Stored nation ratings cache: %s (dataset: %s, nations: %d)",
		cacheKey, cacheData.DatasetID, len(cacheData.NationsData))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "cached",
		"key":    cacheKey,
	})
}

// handleRetrieveNationRatingsCache handles GET /api/cache/nation-ratings/{cacheKey}
func handleRetrieveNationRatingsCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "cache.retrieve_nation_ratings")
	defer span.End()

	cacheKey := extractCacheKeyFromPath(r.URL.Path)
	if cacheKey == "" {
		http.Error(w, "Invalid cache key in URL", http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("cache.key", cacheKey),
		attribute.String("cache.operation", "retrieve"),
	)

	cacheData, err := cacheStorage.RetrieveCacheData(cacheKey)
	if err != nil {
		SetSpanAttributes(ctx, attribute.String("cache.status", "miss"))
		http.Error(w, "Cache not found", http.StatusNotFound)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("cache.status", "hit"),
		attribute.String("cache.version", cacheData.Version),
		attribute.String("cache.dataset_id", cacheData.DatasetID),
		attribute.Int("cache.nations_count", len(cacheData.NationsData)),
		attribute.String("cache.generated_at", cacheData.GeneratedAt.Format(time.RFC3339)),
	)

	log.Printf("Retrieved nation ratings cache: %s (dataset: %s, generated: %s)",
		cacheKey, cacheData.DatasetID, cacheData.GeneratedAt.Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cacheData)
}

// handleDeleteNationRatingsCache handles DELETE /api/cache/nation-ratings/{cacheKey}
func handleDeleteNationRatingsCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "cache.delete_nation_ratings")
	defer span.End()

	cacheKey := extractCacheKeyFromPath(r.URL.Path)
	if cacheKey == "" {
		http.Error(w, "Invalid cache key in URL", http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("cache.key", cacheKey),
		attribute.String("cache.operation", "delete"),
	)

	if err := cacheStorage.DeleteCacheData(cacheKey); err != nil {
		RecordError(ctx, err, "Failed to delete cache data")
		http.Error(w, "Failed to delete cache", http.StatusInternalServerError)
		return
	}

	SetSpanAttributes(ctx, attribute.String("cache.status", "deleted"))
	log.Printf("Deleted nation ratings cache: %s", cacheKey)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"key":    cacheKey,
	})
}

// cacheHandler routes cache requests to the appropriate handler
func cacheHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a nation-ratings cache request
	if strings.Contains(r.URL.Path, "/api/cache/nation-ratings/") {
		switch r.Method {
		case http.MethodPost:
			handleStoreNationRatingsCache(w, r)
		case http.MethodGet:
			handleRetrieveNationRatingsCache(w, r)
		case http.MethodDelete:
			handleDeleteNationRatingsCache(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, "Cache endpoint not found", http.StatusNotFound)
}
