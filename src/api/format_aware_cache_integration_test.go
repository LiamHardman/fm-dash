package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFormatAwareCacheIntegration(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Enable format-aware caching for the test
	formatAwareCacheEnabled = true
	formatAwareCacheConfig.EnableProtobufOptimization = true
	formatAwareCacheConfig.SeparateCacheKeys = true
	
	// We don't need a context for this test
	
	// Create test data
	const baseCacheKey = "test_integration_key"
	testData := map[string]interface{}{
		"name": "test",
		"value": 123,
	}
	
	// Test JSON format
	SetFormatAwareCacheItem(baseCacheKey, FormatTypeJSON, testData, 5*time.Minute)
	
	// Test Protobuf format with different data
	protoData := "mock_protobuf_data"
	SetFormatAwareCacheItem(baseCacheKey, FormatTypeProtobuf, protoData, 5*time.Minute)
	
	// Verify that both formats are stored separately
	jsonData, foundJSON := GetFormatAwareCacheItem(baseCacheKey, FormatTypeJSON)
	protoDataRetrieved, foundProto := GetFormatAwareCacheItem(baseCacheKey, FormatTypeProtobuf)
	
	if !foundJSON || !foundProto {
		t.Errorf("Failed to retrieve cached data: JSON found=%v, Proto found=%v", foundJSON, foundProto)
	}
	
	// Verify that the data is different for each format
	if jsonData == protoDataRetrieved {
		t.Errorf("Expected different data for different formats")
	}
	
	// Test format detection from request
	reqJSON := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqJSON.Header.Set("Accept", "application/json")
	
	reqProto := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqProto.Header.Set("Accept", "application/x-protobuf")
	
	jsonFormat := GetCacheFormatFromRequest(reqJSON)
	protoFormat := GetCacheFormatFromRequest(reqProto)
	
	if jsonFormat != FormatTypeJSON {
		t.Errorf("Expected JSON format, got %s", jsonFormat)
	}
	
	if protoFormat != FormatTypeProtobuf {
		t.Errorf("Expected Protobuf format, got %s", protoFormat)
	}
	
	// Test cache key differentiation
	jsonKey := FormatAwareCacheKey(baseCacheKey, FormatTypeJSON)
	protoKey := FormatAwareCacheKey(baseCacheKey, FormatTypeProtobuf)
	
	if jsonKey == protoKey {
		t.Errorf("Expected different cache keys for different formats")
	}
	
	// Test memory optimization
	memBefore := GetMemCacheDetailedStats()
	
	// Add a large number of items to test memory optimization
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("test_key_%d", i)
		SetFormatAwareCacheItem(key, FormatTypeJSON, testData, 5*time.Minute)
		SetFormatAwareCacheItem(key, FormatTypeProtobuf, protoData, 5*time.Minute)
	}
	
	memAfter := GetMemCacheDetailedStats()
	
	t.Logf("Memory usage before: %v, after: %v", memBefore["size_mb"], memAfter["size_mb"])
	
	// Test DeleteAllFormatVariants
	DeleteAllFormatVariants(baseCacheKey)
	
	_, foundJSONAfter := GetFormatAwareCacheItem(baseCacheKey, FormatTypeJSON)
	_, foundProtoAfter := GetFormatAwareCacheItem(baseCacheKey, FormatTypeProtobuf)
	
	if foundJSONAfter || foundProtoAfter {
		t.Errorf("Expected all format variants to be deleted, but found JSON=%v, Proto=%v", 
			foundJSONAfter, foundProtoAfter)
	}
}