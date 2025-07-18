package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "api/proto"
	"go.opentelemetry.io/otel/attribute"
)

// CachedPlayerDataResponse represents a cached player data response with format-specific data
type CachedPlayerDataResponse struct {
	Format         FormatType
	JSONData       []Player
	ProtobufData   *pb.PlayerDataResponse
	CurrencySymbol string
	CacheTime      time.Time
	FilterHash     string
}

// GetCachedPlayerData retrieves player data from the format-aware cache
func GetCachedPlayerData(ctx context.Context, r *http.Request, cacheKey string) (*CachedPlayerDataResponse, bool) {
	format := GetCacheFormatFromRequest(r)
	
	// Try to get from format-specific cache
	if cached, found := GetFormatAwareCacheItem(cacheKey, format); found {
		if cachedResponse, ok := cached.(*CachedPlayerDataResponse); ok {
			AddSpanEvent(ctx, "cache.hit", 
				attribute.String("cache.key", cacheKey),
				attribute.String("cache.format", string(format)),
				attribute.Float64("cache.age_seconds", time.Since(cachedResponse.CacheTime).Seconds()))
			
			logDebug(ctx, "Cache hit for player data",
				"cache_key", cacheKey,
				"format", format,
				"age_seconds", time.Since(cachedResponse.CacheTime).Seconds())
			
			return cachedResponse, true
		}
	}
	
	AddSpanEvent(ctx, "cache.miss", 
		attribute.String("cache.key", cacheKey),
		attribute.String("cache.format", string(format)))
	
	return nil, false
}

// CachePlayerData stores player data in the format-aware cache
func CachePlayerData(ctx context.Context, cacheKey string, players []Player, currencySymbol string, filterHash string, expiration time.Duration) {
	// Create JSON format cache entry
	jsonResponse := &CachedPlayerDataResponse{
		Format:         FormatTypeJSON,
		JSONData:       players,
		CurrencySymbol: currencySymbol,
		CacheTime:      time.Now(),
		FilterHash:     filterHash,
	}
	
	// Create Protobuf format cache entry
	protoResponse := &CachedPlayerDataResponse{
		Format:         FormatTypeProtobuf,
		CurrencySymbol: currencySymbol,
		CacheTime:      time.Now(),
		FilterHash:     filterHash,
	}
	
	// Create protobuf response directly without intermediate conversion
	
	// Create protobuf response
	requestID := GetTraceID(ctx)
	protoPlayerResponse := &pb.PlayerDataResponse{
		Players:        make([]*pb.Player, 0, len(players)),
		CurrencySymbol: currencySymbol,
		Metadata:       CreateResponseMetadata(requestID, int32(len(players)), true),
	}
	
	// Convert each player to protobuf
	for _, player := range players {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			logError(ctx, "Failed to convert player to protobuf for caching",
				"error", err,
				"player_uid", player.UID,
				"player_name", player.Name)
			continue
		}
		protoPlayerResponse.Players = append(protoPlayerResponse.Players, protoPlayer)
	}
	
	// Set the protobuf data in the cache response
	protoResponse.ProtobufData = protoPlayerResponse
	
	// Store both formats in cache
	SetFormatAwareCacheItem(cacheKey, FormatTypeJSON, jsonResponse, expiration)
	
	// Optimize memory usage for protobuf cached data
	optimizedProtoResponse := OptimizeMemoryForProtobuf(protoResponse).(*CachedPlayerDataResponse)
	SetFormatAwareCacheItem(cacheKey, FormatTypeProtobuf, optimizedProtoResponse, expiration)
	
	logDebug(ctx, "Cached player data in both formats",
		"cache_key", cacheKey,
		"player_count", len(players),
		"json_size", estimateSize(jsonResponse),
		"protobuf_size", estimateSize(optimizedProtoResponse))
	
	AddSpanEvent(ctx, "cache.store", 
		attribute.String("cache.key", cacheKey),
		attribute.Int("cache.player_count", len(players)),
		attribute.Int64("cache.json_size", estimateSize(jsonResponse)),
		attribute.Int64("cache.protobuf_size", estimateSize(optimizedProtoResponse)))
}

// GeneratePlayerCacheKey creates a cache key for player data based on filters
func GeneratePlayerCacheKey(datasetID string, filters map[string]string) string {
	baseKey := fmt.Sprintf("players:%s", datasetID)
	
	// If there are no filters, return the base key
	if len(filters) == 0 {
		return baseKey
	}
	
	// Create a filter hash for the cache key
	filterHash := GenerateFilterHash(filters)
	return fmt.Sprintf("%s:filter:%s", baseKey, filterHash)
}

// GenerateFilterHash creates a hash of the filter parameters
func GenerateFilterHash(filters map[string]string) string {
	// Simple implementation - in a real system, use a proper hash function
	hash := ""
	for k, v := range filters {
		hash += fmt.Sprintf("%s=%s;", k, v)
	}
	return hash
}

// OptimizeProtobufPlayerData optimizes memory usage for protobuf player data
func OptimizeProtobufPlayerData(ctx context.Context, protoResponse *pb.PlayerDataResponse) *pb.PlayerDataResponse {
	if protoResponse == nil {
		return nil
	}
	
	// Log the original size
	originalSize := estimateSize(protoResponse)
	
	// Apply memory optimizations:
	
	// 1. Remove redundant data that can be recalculated
	for _, player := range protoResponse.Players {
		// Clear fields that can be recalculated if needed
		player.PerformancePercentiles = nil
	}
	
	// 2. Optimize string storage for common values
	optimizeCommonStrings(protoResponse)
	
	// Log the optimized size
	optimizedSize := estimateSize(protoResponse)
	
	logDebug(ctx, "Optimized protobuf player data memory usage",
		"original_size_bytes", originalSize,
		"optimized_size_bytes", optimizedSize,
		"reduction_percent", float64(originalSize-optimizedSize)/float64(originalSize)*100)
	
	return protoResponse
}

// optimizeCommonStrings optimizes memory usage by reusing common string values
func optimizeCommonStrings(protoResponse *pb.PlayerDataResponse) {
	// This would implement string interning for common values
	// For now, this is a placeholder for the actual implementation
}

// WritePlayerDataResponse writes the player data response using the appropriate format
func WritePlayerDataResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, 
	cachedResponse *CachedPlayerDataResponse) error {
	
	format := GetCacheFormatFromRequest(r)
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	
	if format == FormatTypeProtobuf && cachedResponse.ProtobufData != nil {
		// Write protobuf response
		responseData, err := serializer.Serialize(cachedResponse.ProtobufData)
		if err != nil {
			// Fallback to JSON on serialization error
			logError(ctx, "Failed to serialize protobuf player data, falling back to JSON",
				"error", err,
				"player_count", len(cachedResponse.ProtobufData.GetPlayers()))
			
			// Write JSON response instead
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("X-Cache-Format", "json")
			w.Header().Set("X-Format-Fallback", "true")
			
			return WriteJSONPlayerResponse(w, cachedResponse.JSONData, cachedResponse.CurrencySymbol)
		}
		
		w.Header().Set("Content-Type", serializer.ContentType())
		w.Header().Set("X-Cache-Source", "memory")
		w.Header().Set("X-Cache-Format", "protobuf")
		w.Write(responseData)
		return nil
	} else {
		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Source", "memory")
		w.Header().Set("X-Cache-Format", "json")
		
		return WriteJSONPlayerResponse(w, cachedResponse.JSONData, cachedResponse.CurrencySymbol)
	}
}

// WriteJSONPlayerResponse writes a JSON player response
func WriteJSONPlayerResponse(w http.ResponseWriter, players []Player, currencySymbol string) error {
	response := map[string]interface{}{
		"players":        players,
		"currencySymbol": currencySymbol,
	}
	
	// Use the standard JSON encoder
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(response)
}