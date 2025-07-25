package main

import (
	"testing"
	"time"
)

// TestFormatAwareCacheDeleteDataset tests the DeleteDataset function with format-aware cache
func TestFormatAwareCacheDeleteDataset(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Create test data
	datasetID := "test_delete_dataset"
	players := []Player{
		{
			UID:  1,
			Name: "Test Player",
		},
	}
	currencySymbol := "$"
	
	// Store the test dataset
	SetPlayerData(datasetID, players, currencySymbol)
	
	// Verify that the dataset exists
	storedPlayers, storedCurrency, found := GetPlayerData(datasetID)
	if !found {
		t.Errorf("Expected dataset to exist after SetPlayerData")
	}
	
	if len(storedPlayers) != len(players) {
		t.Errorf("Expected %d players, got %d", len(players), len(storedPlayers))
	}
	
	if storedCurrency != currencySymbol {
		t.Errorf("Expected currency symbol %s, got %s", currencySymbol, storedCurrency)
	}
	
	// Delete the dataset
	_ = DeleteDataset(datasetID)
	
	// Verify that the dataset no longer exists
	_, _, found = GetPlayerData(datasetID)
	if found {
		t.Errorf("Expected dataset to be deleted after DeleteDataset")
	}
	
	// Test with format-aware cache
	cacheKey := "players:" + datasetID
	SetFormatAwareCacheItem(cacheKey, FormatTypeJSON, "test_json_data", 5*time.Minute)
	SetFormatAwareCacheItem(cacheKey, FormatTypeProtobuf, "test_protobuf_data", 5*time.Minute)
	
	// Delete the dataset
	DeletePlayerData(datasetID)
	
	// Verify that the cache entries are deleted
	_, foundJSON := GetFormatAwareCacheItem(cacheKey, FormatTypeJSON)
	_, foundProtobuf := GetFormatAwareCacheItem(cacheKey, FormatTypeProtobuf)
	
	if foundJSON || foundProtobuf {
		t.Errorf("Expected cache entries to be deleted after DeletePlayerData, found JSON=%v, Protobuf=%v",
			foundJSON, foundProtobuf)
	}
}