package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestProtobufStorage_Store(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create test data
	testData := DatasetData{
		Players: []Player{
			{
				UID:                     1,
				Name:                    "Test Player",
				Position:                "ST",
				Age:                     "25",
				Club:                    "Test FC",
				Division:                "Premier League",
				TransferValue:           "£10M",
				Wage:                    "£50K",
				Nationality:             "England",
				NationalityISO:          "ENG",
				NationalityFIFACode:     "ENG",
				Attributes:              map[string]string{"Pace": "15"},
				NumericAttributes:       map[string]int{"Pace": 15},
				PerformanceStatsNumeric: map[string]float64{"Goals": 10.5},
				PerformancePercentiles:  map[string]map[string]float64{"Goals": {"90th": 0.9}},
				ParsedPositions:         []string{"ST"},
				ShortPositions:          []string{"ST"},
				PositionGroups:          []string{"Forward"},
				PAC:                     15,
				SHO:                     14,
				PAS:                     12,
				DRI:                     13,
				DEF:                     8,
				PHY:                     16,
				Overall:                 85,
				BestRoleOverall:         "Advanced Forward",
				RoleSpecificOveralls: []RoleOverallScore{
					{RoleName: "Advanced Forward", Score: 85},
				},
				TransferValueAmount: 10000000,
				WageAmount:          50000,
			},
		},
		CurrencySymbol: "£",
	}

	// Test storing data
	err := storage.Store("test-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Verify data was stored by checking the backend
	storedData, err := backend.Retrieve("test-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve stored data from backend: %v", err)
	}

	// Check if it's stored as protobuf marker
	if storedData.CurrencySymbol != "__PROTOBUF_MARKER__" {
		t.Errorf("Expected protobuf marker, got: %s", storedData.CurrencySymbol)
	}

	if len(storedData.Players) != 1 || storedData.Players[0].UID != -1 {
		t.Errorf("Expected protobuf marker player, got: %+v", storedData.Players)
	}
}

func TestProtobufStorage_Retrieve(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create test data
	testData := DatasetData{
		Players: []Player{
			{
				UID:                     1,
				Name:                    "Test Player",
				Position:                "ST",
				Age:                     "25",
				Club:                    "Test FC",
				Division:                "Premier League",
				TransferValue:           "£10M",
				Wage:                    "£50K",
				Nationality:             "England",
				NationalityISO:          "ENG",
				NationalityFIFACode:     "ENG",
				Attributes:              map[string]string{"Pace": "15"},
				NumericAttributes:       map[string]int{"Pace": 15},
				PerformanceStatsNumeric: map[string]float64{"Goals": 10.5},
				PerformancePercentiles:  map[string]map[string]float64{"Goals": {"90th": 0.9}},
				ParsedPositions:         []string{"ST"},
				ShortPositions:          []string{"ST"},
				PositionGroups:          []string{"Forward"},
				PAC:                     15,
				SHO:                     14,
				PAS:                     12,
				DRI:                     13,
				DEF:                     8,
				PHY:                     16,
				Overall:                 85,
				BestRoleOverall:         "Advanced Forward",
				RoleSpecificOveralls: []RoleOverallScore{
					{RoleName: "Advanced Forward", Score: 85},
				},
				TransferValueAmount: 10000000,
				WageAmount:          50000,
			},
		},
		CurrencySymbol: "£",
	}

	// Store data first
	err := storage.Store("test-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve data
	retrievedData, err := storage.Retrieve("test-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Verify data integrity
	if len(retrievedData.Players) != len(testData.Players) {
		t.Errorf("Expected %d players, got %d", len(testData.Players), len(retrievedData.Players))
	}

	if retrievedData.CurrencySymbol != testData.CurrencySymbol {
		t.Errorf("Expected currency symbol %s, got %s", testData.CurrencySymbol, retrievedData.CurrencySymbol)
	}

	// Check first player data
	if len(retrievedData.Players) > 0 {
		player := retrievedData.Players[0]
		expectedPlayer := testData.Players[0]

		if player.UID != expectedPlayer.UID {
			t.Errorf("Expected UID %d, got %d", expectedPlayer.UID, player.UID)
		}

		if player.Name != expectedPlayer.Name {
			t.Errorf("Expected name %s, got %s", expectedPlayer.Name, player.Name)
		}

		if player.Overall != expectedPlayer.Overall {
			t.Errorf("Expected overall %d, got %d", expectedPlayer.Overall, player.Overall)
		}
	}
}

func TestProtobufStorage_Delete(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create and store test data
	testData := DatasetData{
		Players:        []Player{{UID: 1, Name: "Test Player"}},
		CurrencySymbol: "£",
	}

	err := storage.Store("test-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Delete the dataset
	err = storage.Delete("test-dataset")
	if err != nil {
		t.Fatalf("Failed to delete dataset: %v", err)
	}

	// Verify deletion
	_, err = storage.Retrieve("test-dataset")
	if err == nil {
		t.Error("Expected error when retrieving deleted dataset")
	}
}

func TestProtobufStorage_List(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create and store test data
	testData := DatasetData{
		Players:        []Player{{UID: 1, Name: "Test Player"}},
		CurrencySymbol: "£",
	}

	err := storage.Store("test-dataset-1", testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	err = storage.Store("test-dataset-2", testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// List datasets
	datasets, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list datasets: %v", err)
	}

	if len(datasets) != 2 {
		t.Errorf("Expected 2 datasets, got %d", len(datasets))
	}
}

func TestProtobufStorage_CleanupOldDatasets(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Test cleanup (should delegate to backend)
	err := storage.CleanupOldDatasets(24*time.Hour, []string{"demo"})
	if err != nil {
		t.Fatalf("Failed to cleanup old datasets: %v", err)
	}
}

func TestProtobufStorage_JSONFallback(t *testing.T) {
	// Create an in-memory backend for testing
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create test data with invalid protobuf conversion (this should trigger fallback)
	testData := DatasetData{
		Players: []Player{
			{
				UID:  1,
				Name: "Test Player",
				// Add some fields that might cause protobuf issues
				PerformancePercentiles: map[string]map[string]float64{
					"test": nil, // This might cause issues
				},
			},
		},
		CurrencySymbol: "£",
	}

	// Store data (should fall back to JSON if protobuf fails)
	err := storage.Store("test-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store data with fallback: %v", err)
	}

	// Retrieve data (should fall back to JSON if protobuf fails)
	retrievedData, err := storage.Retrieve("test-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve data with fallback: %v", err)
	}

	// Verify basic data integrity
	if len(retrievedData.Players) != len(testData.Players) {
		t.Errorf("Expected %d players, got %d", len(testData.Players), len(retrievedData.Players))
	}
}

func TestProtobufStorage_CompressionDecompression(t *testing.T) {
	storage := CreateProtobufStorage(CreateInMemoryStorage())

	// Test data
	testData := []byte("This is test data for compression")

	// Test compression
	compressed, err := storage.compressProtobufData(testData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("Compressed data should not be empty")
	}

	// Test decompression
	decompressed, err := storage.decompressProtobufData(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	if string(decompressed) != string(testData) {
		t.Errorf("Expected decompressed data %s, got %s", string(testData), string(decompressed))
	}
}

func TestProtobufStorage_ProtobufBytesStorage(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)
	ctx := context.Background()

	// Test data
	testData := []byte("This is test protobuf data")

	// Store protobuf bytes
	err := storage.storeProtobufBytes(ctx, "test-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store protobuf bytes: %v", err)
	}

	// Retrieve protobuf bytes
	retrievedData, err := storage.retrieveProtobufBytes(ctx, "test-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve protobuf bytes: %v", err)
	}

	if string(retrievedData) != string(testData) {
		t.Errorf("Expected retrieved data %s, got %s", string(testData), string(retrievedData))
	}
}

func TestProtobufStorage_CompressionRatio(t *testing.T) {
	storage := CreateProtobufStorage(CreateInMemoryStorage())

	// Create test data with repetitive content that should compress well
	testData := make([]byte, 1000)
	for i := range testData {
		testData[i] = byte(i % 10) // Repetitive pattern
	}

	// Test compression
	compressed, err := storage.compressProtobufData(testData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	// Verify compression actually reduces size
	if len(compressed) >= len(testData) {
		t.Errorf("Compression should reduce size. Original: %d, Compressed: %d", len(testData), len(compressed))
	}

	// Calculate compression ratio
	ratio := float64(len(testData)) / float64(len(compressed))
	if ratio < 1.1 { // Should achieve at least 10% compression
		t.Errorf("Expected compression ratio > 1.1, got %.2f", ratio)
	}

	t.Logf("Compression ratio: %.2f (Original: %d bytes, Compressed: %d bytes)", ratio, len(testData), len(compressed))

	// Test decompression
	decompressed, err := storage.decompressProtobufData(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Verify data integrity
	if len(decompressed) != len(testData) {
		t.Errorf("Decompressed data length mismatch. Expected: %d, Got: %d", len(testData), len(decompressed))
	}

	for i, b := range decompressed {
		if b != testData[i] {
			t.Errorf("Data mismatch at index %d. Expected: %d, Got: %d", i, testData[i], b)
			break
		}
	}
}

func TestProtobufStorage_CompressionWithLargeDataset(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create a large dataset with multiple players
	players := make([]Player, 100)
	for i := 0; i < 100; i++ {
		players[i] = Player{
			UID:                     int64(i + 1),
			Name:                    fmt.Sprintf("Player %d", i+1),
			Position:                "ST",
			Age:                     "25",
			Club:                    "Test FC",
			Division:                "Premier League",
			TransferValue:           "£10M",
			Wage:                    "£50K",
			Nationality:             "England",
			NationalityISO:          "ENG",
			NationalityFIFACode:     "ENG",
			Attributes:              map[string]string{"Pace": "15", "Shooting": "14"},
			NumericAttributes:       map[string]int{"Pace": 15, "Shooting": 14},
			PerformanceStatsNumeric: map[string]float64{"Goals": 10.5, "Assists": 5.2},
			PerformancePercentiles:  map[string]map[string]float64{"Goals": {"90th": 0.9}},
			ParsedPositions:         []string{"ST", "CF"},
			ShortPositions:          []string{"ST"},
			PositionGroups:          []string{"Forward"},
			PAC:                     15,
			SHO:                     14,
			PAS:                     12,
			DRI:                     13,
			DEF:                     8,
			PHY:                     16,
			Overall:                 85,
			BestRoleOverall:         "Advanced Forward",
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Advanced Forward", Score: 85},
				{RoleName: "Poacher", Score: 82},
			},
			TransferValueAmount: 10000000,
			WageAmount:          50000,
		}
	}

	testData := DatasetData{
		Players:        players,
		CurrencySymbol: "£",
	}

	// Store the large dataset
	err := storage.Store("large-dataset", testData)
	if err != nil {
		t.Fatalf("Failed to store large dataset: %v", err)
	}

	// Retrieve the large dataset
	retrievedData, err := storage.Retrieve("large-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve large dataset: %v", err)
	}

	// Verify data integrity
	if len(retrievedData.Players) != len(testData.Players) {
		t.Errorf("Expected %d players, got %d", len(testData.Players), len(retrievedData.Players))
	}

	if retrievedData.CurrencySymbol != testData.CurrencySymbol {
		t.Errorf("Expected currency symbol %s, got %s", testData.CurrencySymbol, retrievedData.CurrencySymbol)
	}

	// Spot check a few players
	for i := 0; i < 5; i++ {
		if retrievedData.Players[i].UID != testData.Players[i].UID {
			t.Errorf("Player %d UID mismatch. Expected: %d, Got: %d", i, testData.Players[i].UID, retrievedData.Players[i].UID)
		}
		if retrievedData.Players[i].Name != testData.Players[i].Name {
			t.Errorf("Player %d Name mismatch. Expected: %s, Got: %s", i, testData.Players[i].Name, retrievedData.Players[i].Name)
		}
	}

	t.Logf("Successfully stored and retrieved dataset with %d players", len(retrievedData.Players))
}

func TestProtobufStorage_CompressionErrorHandling(t *testing.T) {
	storage := CreateProtobufStorage(CreateInMemoryStorage())

	// Test decompression with invalid data
	invalidData := []byte("this is not gzip compressed data")
	_, err := storage.decompressProtobufData(invalidData)
	if err == nil {
		t.Error("Expected error when decompressing invalid data")
	}

	// Test compression with empty data
	emptyData := []byte{}
	compressed, err := storage.compressProtobufData(emptyData)
	if err != nil {
		t.Fatalf("Failed to compress empty data: %v", err)
	}

	// Should be able to decompress empty data
	decompressed, err := storage.decompressProtobufData(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress empty data: %v", err)
	}

	if len(decompressed) != 0 {
		t.Errorf("Expected empty decompressed data, got %d bytes", len(decompressed))
	}
}

func TestProtobufStorage_ErrorHandling(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Test with invalid data that should trigger fallback
	testData := DatasetData{
		Players: []Player{
			{
				UID:  1,
				Name: "Test Player",
				// Create a scenario that might cause protobuf issues
				PerformancePercentiles: map[string]map[string]float64{
					"test": nil, // This might cause issues during conversion
				},
			},
		},
		CurrencySymbol: "£",
	}

	// Store should succeed (with fallback if needed)
	err := storage.Store("error-test-dataset", testData)
	if err != nil {
		t.Fatalf("Store should succeed even with fallback: %v", err)
	}

	// Retrieve should succeed (with fallback if needed)
	retrievedData, err := storage.Retrieve("error-test-dataset")
	if err != nil {
		t.Fatalf("Retrieve should succeed even with fallback: %v", err)
	}

	// Verify basic data integrity
	if len(retrievedData.Players) != len(testData.Players) {
		t.Errorf("Expected %d players, got %d", len(testData.Players), len(retrievedData.Players))
	}
}

func TestProtobufStorage_CustomErrors(t *testing.T) {
	// Test ProtobufError
	err := NewProtobufError("marshal", "test-dataset", "test message", fmt.Errorf("underlying error"))
	expectedMsg := "protobuf marshal failed for dataset test-dataset: test message"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %s, got %s", expectedMsg, err.Error())
	}

	// Test ProtobufConversionError
	convErr := NewProtobufConversionError("to_protobuf", "Player", "test-dataset", fmt.Errorf("conversion failed"))
	expectedConvMsg := "protobuf conversion error (to_protobuf Player) for dataset test-dataset: conversion failed"
	if convErr.Error() != expectedConvMsg {
		t.Errorf("Expected conversion error message %s, got %s", expectedConvMsg, convErr.Error())
	}

	// Test ProtobufCompressionError
	compErr := NewProtobufCompressionError("compress", "test-dataset", fmt.Errorf("compression failed"))
	expectedCompMsg := "protobuf compress error for dataset test-dataset: compression failed"
	if compErr.Error() != expectedCompMsg {
		t.Errorf("Expected compression error message %s, got %s", expectedCompMsg, compErr.Error())
	}
}

func TestProtobufStorage_FallbackEvent(t *testing.T) {
	event := ProtobufFallbackEvent{
		DatasetID: "test-dataset",
		Reason:    FallbackReasonConversionFailed,
		Error:     fmt.Errorf("test error"),
		Message:   "test message",
	}

	expectedStr := "Protobuf fallback for dataset test-dataset: protobuf_conversion_failed - test message (error: test error)"
	if event.String() != expectedStr {
		t.Errorf("Expected fallback event string %s, got %s", expectedStr, event.String())
	}
}

func TestProtobufStorage_RetrieveNonExistentDataset(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Try to retrieve a dataset that doesn't exist
	_, err := storage.Retrieve("non-existent-dataset")
	if err == nil {
		t.Error("Expected error when retrieving non-existent dataset")
	}
}

func TestProtobufStorage_RetrieveNonProtobufDataset(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Store a regular JSON dataset directly to the backend
	regularData := DatasetData{
		Players: []Player{
			{
				UID:  1,
				Name: "Regular Player",
			},
		},
		CurrencySymbol: "£",
	}

	// Store directly to backend (bypassing protobuf)
	err := backend.Store("regular-dataset", regularData)
	if err != nil {
		t.Fatalf("Failed to store regular dataset: %v", err)
	}

	// Try to retrieve through protobuf storage (should fallback to JSON)
	retrievedData, err := storage.Retrieve("regular-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve regular dataset through protobuf storage: %v", err)
	}

	// Verify data integrity
	if len(retrievedData.Players) != len(regularData.Players) {
		t.Errorf("Expected %d players, got %d", len(regularData.Players), len(retrievedData.Players))
	}

	if retrievedData.Players[0].Name != regularData.Players[0].Name {
		t.Errorf("Expected player name %s, got %s", regularData.Players[0].Name, retrievedData.Players[0].Name)
	}
}

func TestProtobufStorage_InvalidProtobufData(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create a dataset that looks like protobuf but has invalid data
	invalidProtobufDataset := DatasetData{
		Players: []Player{{
			UID:      -1,
			Name:     "__PROTOBUF_DATA__",
			Position: "invalid-base64-data-!@#$%", // Invalid base64
			Club:     "PROTOBUF_STORAGE",
		}},
		CurrencySymbol: "__PROTOBUF_MARKER__",
	}

	// Store the invalid dataset directly to backend
	err := backend.Store("invalid-protobuf-dataset", invalidProtobufDataset)
	if err != nil {
		t.Fatalf("Failed to store invalid protobuf dataset: %v", err)
	}

	// Try to retrieve through protobuf storage (should fallback to JSON)
	retrievedData, err := storage.Retrieve("invalid-protobuf-dataset")
	if err != nil {
		t.Fatalf("Failed to retrieve invalid protobuf dataset: %v", err)
	}

	// Should get the raw invalid dataset back (JSON fallback)
	if retrievedData.CurrencySymbol != "__PROTOBUF_MARKER__" {
		t.Errorf("Expected fallback to return raw dataset")
	}
}

func TestProtobufStorage_LogFallbackEvent(t *testing.T) {
	storage := CreateProtobufStorage(CreateInMemoryStorage())

	// Test that logFallbackEvent doesn't panic
	event := ProtobufFallbackEvent{
		DatasetID: "test-dataset",
		Reason:    FallbackReasonMarshalFailed,
		Error:     fmt.Errorf("test error"),
		Message:   "test message",
	}

	// This should not panic
	storage.logFallbackEvent(event)
}

func TestProtobufStorage_ErrorWrapping(t *testing.T) {
	// Test that custom errors properly wrap underlying errors
	underlyingErr := fmt.Errorf("underlying error")
	
	protobufErr := NewProtobufError("test", "dataset", "message", underlyingErr)
	if protobufErr.Unwrap() != underlyingErr {
		t.Error("ProtobufError should properly wrap underlying error")
	}

	convErr := NewProtobufConversionError("to_protobuf", "Player", "dataset", underlyingErr)
	if convErr.Unwrap() != underlyingErr {
		t.Error("ProtobufConversionError should properly wrap underlying error")
	}

	compErr := NewProtobufCompressionError("compress", "dataset", underlyingErr)
	if compErr.Unwrap() != underlyingErr {
		t.Error("ProtobufCompressionError should properly wrap underlying error")
	}
}