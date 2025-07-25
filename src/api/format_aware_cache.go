package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FormatType represents the serialization format type
type FormatType string

const (
	// FormatTypeJSON represents JSON serialization format
	FormatTypeJSON FormatType = "json"
	// FormatTypeProtobuf represents protobuf serialization format
	FormatTypeProtobuf FormatType = "protobuf"
)

// FormatAwareCacheKey generates a cache key that includes the response format
func FormatAwareCacheKey(baseKey string, format FormatType) string {
	return fmt.Sprintf("%s:%s", baseKey, format)
}

// ParseFormatAwareCacheKey parses a format-aware cache key into base key and format
func ParseFormatAwareCacheKey(key string) (string, FormatType) {
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return key, FormatTypeJSON // Default to JSON if no format specified
	}
	
	format := FormatType(parts[len(parts)-1])
	baseKey := strings.Join(parts[:len(parts)-1], ":")
	
	// Validate format type
	if format != FormatTypeJSON && format != FormatTypeProtobuf {
		return key, FormatTypeJSON // Default to JSON for invalid formats
	}
	
	return baseKey, format
}

// FormatAwareCache provides format-specific caching operations
type FormatAwareCache struct {
	// Underlying cache implementation
}

// GetFormatAwareCacheItem retrieves a cached item with format awareness
func GetFormatAwareCacheItem(baseKey string, format FormatType) (interface{}, bool) {
	cacheKey := FormatAwareCacheKey(baseKey, format)
	return getFromMemCache(cacheKey)
}

// SetFormatAwareCacheItem stores an item in the cache with format awareness
func SetFormatAwareCacheItem(baseKey string, format FormatType, value interface{}, expiration time.Duration) {
	cacheKey := FormatAwareCacheKey(baseKey, format)
	setInMemCache(cacheKey, value, expiration)
}

// DeleteFormatAwareCacheItem removes an item from the cache with format awareness
func DeleteFormatAwareCacheItem(baseKey string, format FormatType) {
	cacheKey := FormatAwareCacheKey(baseKey, format)
	deleteFromMemCache(cacheKey)
}

// DeleteAllFormatVariants removes all format variants of a cache item
func DeleteAllFormatVariants(baseKey string) {
	// Delete JSON variant
	DeleteFormatAwareCacheItem(baseKey, FormatTypeJSON)
	// Delete Protobuf variant
	DeleteFormatAwareCacheItem(baseKey, FormatTypeProtobuf)
}

// GetCacheFormatFromRequest determines the appropriate cache format based on the request
func GetCacheFormatFromRequest(r *http.Request) FormatType {
	negotiator := NewContentNegotiator(r)
	if negotiator.SupportsProtobuf() {
		return FormatTypeProtobuf
	}
	return FormatTypeJSON
}

// OptimizeMemoryForProtobuf optimizes memory usage for protobuf cached data
// by applying specific memory optimization techniques for protobuf data
func OptimizeMemoryForProtobuf(data interface{}) interface{} {
	// For now, just return the data as is
	// In a real implementation, this would apply protobuf-specific optimizations
	return data
}