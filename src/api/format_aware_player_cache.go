package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
				protoResponse := &pb.PlayerDataResponse{}
				if err := proto.Unmarshal(serializedResponse.Bytes, protoResponse); err != nil {
					logError(ctx, "Failed to deserialize cached protobuf player data", "error", err)
					break
				}
				cachedResponse.ProtobufData = protoResponse
				if cachedResponse.CurrencySymbol == "" {
					cachedResponse.CurrencySymbol = protoResponse.GetCurrencySymbol()
				}
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

// CachePlayerData stores player data in the format-aware cache
func CachePlayerData(ctx context.Context, cacheKey string, players []Player, currencySymbol string, filterHash string, expiration time.Duration) {
	// Serialize and cache only bytes to reduce memory usage
	// JSON bytes
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

	// Protobuf bytes
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
		SetFormatAwareCacheItem(cacheKey, FormatTypeProtobuf, &CachedSerializedResponse{
			Format:         FormatTypeProtobuf,
			Bytes:          data,
			CurrencySymbol: currencySymbol,
			CacheTime:      time.Now(),
			FilterHash:     filterHash,
		}, expiration)
	} else {
		logError(ctx, "Failed to serialize protobuf for cache", "error", err)
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
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()

	// Attempt to serve from serialized-byte cache first
	baseKey := cacheKeyFromRequest(r)
	if serialized, ok := GetFormatAwareCacheItem(baseKey, format); ok {
		if csr, ok := serialized.(*CachedSerializedResponse); ok && csr != nil {
			ct := serializer.ContentType()
			if format == FormatTypeJSON {
				ct = "application/json"
			}
			w.Header().Set("Content-Type", ct)
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("X-Cache-Format", string(format))
			w.Header().Set("Cache-Control", "private, max-age=300")
			if _, err := w.Write(csr.Bytes); err != nil {
				logError(r.Context(), "Error writing serialized response", "error", err)
			}
			return nil
		}
	}

	// Fallback to existing behavior if serialized bytes are not present
	if format == FormatTypeProtobuf && cachedResponse.ProtobufData != nil {
		responseData, err := serializer.Serialize(cachedResponse.ProtobufData)
		if err != nil {
			logError(ctx, "Failed to serialize protobuf player data",
				"error", err,
				"player_count", len(cachedResponse.ProtobufData.GetPlayers()))
			return fmt.Errorf("failed to serialize protobuf player data: %w", err)
		}
		w.Header().Set("Content-Type", serializer.ContentType())
		w.Header().Set("X-Cache-Source", "memory")
		w.Header().Set("X-Cache-Format", "protobuf")
		w.Header().Set("Cache-Control", "private, max-age=300")
		if _, err := w.Write(responseData); err != nil {
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

// cacheKeyFromRequest reconstructs the base cache key as used by GeneratePlayerCacheKey in the handler
func cacheKeyFromRequest(r *http.Request) string {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	datasetID := ""
	if len(parts) > 0 {
		datasetID = parts[0]
	}
	q := r.URL.Query()
	filters := map[string]string{}
	for _, key := range []string{
		"position", "role", "minAge", "maxAge", "minTransferValue",
		"maxTransferValue", "maxSalary", "divisionFilter", "targetDivision",
		"positionCompare",
	} {
		if value := q.Get(key); value != "" {
			filters[key] = value
		}
	}
	return GeneratePlayerCacheKey(datasetID, filters)
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
