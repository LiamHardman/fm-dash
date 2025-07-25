package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFormatAwareCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		baseKey  string
		format   FormatType
		expected string
	}{
		{
			name:     "JSON format",
			baseKey:  "test_key",
			format:   FormatTypeJSON,
			expected: "test_key:json",
		},
		{
			name:     "Protobuf format",
			baseKey:  "test_key",
			format:   FormatTypeProtobuf,
			expected: "test_key:protobuf",
		},
		{
			name:     "Complex key with JSON format",
			baseKey:  "players:dataset123:filtered",
			format:   FormatTypeJSON,
			expected: "players:dataset123:filtered:json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAwareCacheKey(tt.baseKey, tt.format)
			if result != tt.expected {
				t.Errorf("FormatAwareCacheKey(%q, %q) = %q, want %q", 
					tt.baseKey, tt.format, result, tt.expected)
			}
		})
	}
}

func TestParseFormatAwareCacheKey(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		expectedBase   string
		expectedFormat FormatType
	}{
		{
			name:           "JSON format",
			key:            "test_key:json",
			expectedBase:   "test_key",
			expectedFormat: FormatTypeJSON,
		},
		{
			name:           "Protobuf format",
			key:            "test_key:protobuf",
			expectedBase:   "test_key",
			expectedFormat: FormatTypeProtobuf,
		},
		{
			name:           "Complex key with JSON format",
			key:            "players:dataset123:filtered:json",
			expectedBase:   "players:dataset123:filtered",
			expectedFormat: FormatTypeJSON,
		},
		{
			name:           "Invalid format defaults to JSON",
			key:            "test_key:invalid",
			expectedBase:   "test_key",
			expectedFormat: FormatTypeJSON,
		},
		{
			name:           "No format defaults to JSON",
			key:            "test_key",
			expectedBase:   "test_key",
			expectedFormat: FormatTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, format := ParseFormatAwareCacheKey(tt.key)
			if base != tt.expectedBase || format != tt.expectedFormat {
				t.Errorf("ParseFormatAwareCacheKey(%q) = (%q, %q), want (%q, %q)", 
					tt.key, base, format, tt.expectedBase, tt.expectedFormat)
			}
		})
	}
}

func TestGetCacheFormatFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedFormat FormatType
	}{
		{
			name:           "No Accept header",
			acceptHeader:   "",
			expectedFormat: FormatTypeJSON,
		},
		{
			name:           "JSON Accept header",
			acceptHeader:   "application/json",
			expectedFormat: FormatTypeJSON,
		},
		{
			name:           "Protobuf Accept header",
			acceptHeader:   "application/x-protobuf",
			expectedFormat: FormatTypeProtobuf,
		},
		{
			name:           "Multiple Accept headers with Protobuf",
			acceptHeader:   "application/json, application/x-protobuf",
			expectedFormat: FormatTypeProtobuf,
		},
		{
			name:           "Multiple Accept headers with quality and Protobuf",
			acceptHeader:   "application/json;q=0.9, application/x-protobuf;q=1.0",
			expectedFormat: FormatTypeProtobuf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			
			format := GetCacheFormatFromRequest(req)
			if format != tt.expectedFormat {
				t.Errorf("GetCacheFormatFromRequest() with Accept: %q = %q, want %q", 
					tt.acceptHeader, format, tt.expectedFormat)
			}
		})
	}
}

func TestFormatAwareCacheOperations(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Test data
	const baseKey = "test_cache_key"
	jsonValue := map[string]interface{}{"name": "test", "value": 123}
	protobufValue := "mock_protobuf_data"
	
	// Test setting and getting JSON format
	SetFormatAwareCacheItem(baseKey, FormatTypeJSON, jsonValue, 1*time.Minute)
	cachedJSON, foundJSON := GetFormatAwareCacheItem(baseKey, FormatTypeJSON)
	
	if !foundJSON {
		t.Errorf("JSON value not found in cache")
	}
	
	if cachedJSON == nil {
		t.Errorf("JSON value is nil")
	}
	
	// Test setting and getting Protobuf format
	SetFormatAwareCacheItem(baseKey, FormatTypeProtobuf, protobufValue, 1*time.Minute)
	cachedProtobuf, foundProtobuf := GetFormatAwareCacheItem(baseKey, FormatTypeProtobuf)
	
	if !foundProtobuf {
		t.Errorf("Protobuf value not found in cache")
	}
	
	if cachedProtobuf == nil {
		t.Errorf("Protobuf value is nil")
	}
	
	// Test deleting specific format
	DeleteFormatAwareCacheItem(baseKey, FormatTypeJSON)
	_, foundJSONAfterDelete := GetFormatAwareCacheItem(baseKey, FormatTypeJSON)
	_, foundProtobufAfterJSONDelete := GetFormatAwareCacheItem(baseKey, FormatTypeProtobuf)
	
	if foundJSONAfterDelete {
		t.Errorf("JSON value still found after deletion")
	}
	
	if !foundProtobufAfterJSONDelete {
		t.Errorf("Protobuf value not found after JSON deletion")
	}
	
	// Test deleting all format variants
	DeleteAllFormatVariants(baseKey)
	_, foundJSONAfterDeleteAll := GetFormatAwareCacheItem(baseKey, FormatTypeJSON)
	_, foundProtobufAfterDeleteAll := GetFormatAwareCacheItem(baseKey, FormatTypeProtobuf)
	
	if foundJSONAfterDeleteAll {
		t.Errorf("JSON value still found after DeleteAllFormatVariants")
	}
	
	if foundProtobufAfterDeleteAll {
		t.Errorf("Protobuf value still found after DeleteAllFormatVariants")
	}
}