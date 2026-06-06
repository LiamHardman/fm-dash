package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "api/proto"

	"google.golang.org/protobuf/proto"

	"go.opentelemetry.io/otel/attribute"
)

// CachedPlayerDataResponse represents a cached player data response with format-specific data
type CachedPlayerDataResponse struct {
	Format         FormatType
	JSONData       []Player
	ProtobufData   *pb.PlayerDataResponse
	ProtoBytes     []byte // pre-marshaled protobuf response for zero-copy write
	CurrencySymbol string
	CacheTime      time.Time
	FilterHash     string
}

// CachedSerializedResponse stores only the serialized bytes for a response
// to minimize in-memory duplication of large player slices.
type CachedSerializedResponse struct {
	Format         FormatType
	Bytes          []byte
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
		if serializedResponse, ok := cached.(*CachedSerializedResponse); ok && serializedResponse != nil {
			cachedResponse := &CachedPlayerDataResponse{
				Format:         serializedResponse.Format,
				CurrencySymbol: serializedResponse.CurrencySymbol,
				CacheTime:      serializedResponse.CacheTime,
				FilterHash:     serializedResponse.FilterHash,
			}

			switch format {
			case FormatTypeJSON:
				var jsonResponse struct {
					Players        []Player `json:"players"`
					CurrencySymbol string   `json:"currencySymbol"`
				}
				if err := json.Unmarshal(serializedResponse.Bytes, &jsonResponse); err != nil {
					logError(ctx, "Failed to deserialize cached JSON player data", "error", err)
					break
				}
				cachedResponse.JSONData = jsonResponse.Players
				if cachedResponse.CurrencySymbol == "" {
					cachedResponse.CurrencySymbol = jsonResponse.CurrencySymbol
				}
			case FormatTypeProtobuf:
				// Store raw bytes directly — WritePlayerDataResponse will write them without re-marshaling
				cachedResponse.ProtoBytes = serializedResponse.Bytes
			}

			AddSpanEvent(ctx, "cache.hit",
				attribute.String("cache.key", cacheKey),
				attribute.String("cache.format", string(format)),
				attribute.Float64("cache.age_seconds", time.Since(serializedResponse.CacheTime).Seconds()))

			return cachedResponse, true
		}
	}

	AddSpanEvent(ctx, "cache.miss",
		attribute.String("cache.key", cacheKey),
		attribute.String("cache.format", string(format)))

	return nil, false
}

// CachePlayerData stores player data in the format-aware cache.
// protoBytes should be the pre-marshaled protobuf response; pass nil to skip proto caching.
func CachePlayerData(ctx context.Context, cacheKey string, players []Player, currencySymbol string, filterHash string, protoBytes []byte, expiration time.Duration) {
	jsonBytes, jsonErr := json.Marshal(map[string]interface{}{
		"players":        players,
		"currencySymbol": currencySymbol,
	})
	if jsonErr == nil {
		SetFormatAwareCacheItem(cacheKey, FormatTypeJSON, &CachedSerializedResponse{
			Format:         FormatTypeJSON,
			Bytes:          jsonBytes,
			CurrencySymbol: currencySymbol,
			CacheTime:      time.Now(),
			FilterHash:     filterHash,
		}, expiration)
	} else {
		logError(ctx, "Failed to serialize JSON for cache", "error", jsonErr)
	}

	if protoBytes == nil {
		// No pre-built bytes provided — build proto from players as a fallback
		requestID := GetTraceID(ctx)
		protoPlayerResponse := &pb.PlayerDataResponse{
			Players:        make([]*pb.Player, 0, len(players)),
			CurrencySymbol: currencySymbol,
			Metadata:       CreateResponseMetadata(requestID, safeInt32(len(players)), true),
		}
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
		if data, err := proto.Marshal(protoPlayerResponse); err == nil {
			protoBytes = data
		} else {
			logError(ctx, "Failed to serialize protobuf for cache", "error", err)
		}
	}
	if protoBytes != nil {
		SetFormatAwareCacheItem(cacheKey, FormatTypeProtobuf, &CachedSerializedResponse{
			Format:         FormatTypeProtobuf,
			Bytes:          protoBytes,
			CurrencySymbol: currencySymbol,
			CacheTime:      time.Now(),
			FilterHash:     filterHash,
		}, expiration)
	}

	AddSpanEvent(ctx, "cache.store",
		attribute.String("cache.key", cacheKey),
		attribute.Int("cache.player_count", len(players)))
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

	originalSize := estimateSize(protoResponse)
	optimizeCommonStrings(protoResponse)
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
	if cachedResponse == nil {
		return fmt.Errorf("cached player data is unavailable")
	}

	format := GetCacheFormatFromRequest(r)

	// Fast path: pre-built bytes require no serialization work
	if format == FormatTypeProtobuf && len(cachedResponse.ProtoBytes) > 0 {
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Header().Set("X-Cache-Source", "memory")
		w.Header().Set("X-Cache-Format", "protobuf")
		w.Header().Set("Cache-Control", "private, max-age=300")
		if _, err := w.Write(cachedResponse.ProtoBytes); err != nil {
			logError(r.Context(), "Error writing protobuf response", "error", err)
		}
		return nil
	}

	if format == FormatTypeProtobuf {
		return fmt.Errorf("protobuf response requested but protobuf data is unavailable")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Source", "memory")
	w.Header().Set("X-Cache-Format", "json")
	w.Header().Set("Cache-Control", "private, max-age=300")
	return WriteJSONPlayerResponse(w, cachedResponse.JSONData, cachedResponse.CurrencySymbol)
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
