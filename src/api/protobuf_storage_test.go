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

// TestProtobufStorage_AllStorageBackends tests protobuf storage with different backend types
func TestProtobufStorage_AllStorageBackends(t *testing.T) {
	testCases := []struct {
		name    string
		backend StorageInterface
	}{
		{
			name:    "InMemoryStorage",
			backend: CreateInMemoryStorage(),
		},
		// Note: S3 and LocalFile storage would require more setup for testing
		// but the protobuf wrapper should work with any StorageInterface
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := CreateProtobufStorage(tc.backend)

			// Create comprehensive test data
			testData := DatasetData{
				Players: []Player{
					{
						UID:                     1,
						Name:                    "Backend Test Player",
						Position:                "CM",
						Age:                     "26",
						Club:                    "Test United",
						Division:                "Championship",
						TransferValue:           "£5M",
						Wage:                    "£30K",
						Nationality:             "Scotland",
						NationalityISO:          "SCO",
						NationalityFIFACode:     "SCO",
						AttributeMasked:         false,
						Attributes:              map[string]string{"Passing": "18", "Vision": "17"},
						NumericAttributes:       map[string]int{"Passing": 18, "Vision": 17},
						PerformanceStatsNumeric: map[string]float64{"Assists": 12.3, "KeyPasses": 3.1},
						PerformancePercentiles: map[string]map[string]float64{
							"Passing": {"Assists": 89.2, "KeyPasses": 91.5},
							"Vision":  {"ThroughBalls": 85.7},
						},
						ParsedPositions: []string{"CM", "CAM"},
						ShortPositions:  []string{"CM"},
						PositionGroups:  []string{"Midfielder"},
						PAC:             12,
						SHO:             10,
						PAS:             18,
						DRI:             15,
						DEF:             14,
						PHY:             13,
						GK:              5,
						Overall:         78,
						BestRoleOverall: "Deep Lying Playmaker",
						RoleSpecificOveralls: []RoleOverallScore{
							{RoleName: "Deep Lying Playmaker", Score: 82},
							{RoleName: "Central Midfielder", Score: 78},
						},
						TransferValueAmount: 5000000,
						WageAmount:          30000,
					},
				},
				CurrencySymbol: "£",
			}

			// Test complete storage workflow
			err := storage.Store("backend-test", testData)
			if err != nil {
				t.Fatalf("Failed to store data with %s backend: %v", tc.name, err)
			}

			// Test retrieval
			retrievedData, err := storage.Retrieve("backend-test")
			if err != nil {
				t.Fatalf("Failed to retrieve data with %s backend: %v", tc.name, err)
			}

			// Verify data integrity
			if len(retrievedData.Players) != len(testData.Players) {
				t.Errorf("Player count mismatch with %s backend: expected %d, got %d", 
					tc.name, len(testData.Players), len(retrievedData.Players))
			}

			if retrievedData.CurrencySymbol != testData.CurrencySymbol {
				t.Errorf("Currency symbol mismatch with %s backend: expected %s, got %s", 
					tc.name, testData.CurrencySymbol, retrievedData.CurrencySymbol)
			}

			// Test list functionality
			datasets, err := storage.List()
			if err != nil {
				t.Fatalf("Failed to list datasets with %s backend: %v", tc.name, err)
			}

			found := false
			for _, dataset := range datasets {
				if dataset == "backend-test" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Dataset not found in list with %s backend", tc.name)
			}

			// Test deletion
			err = storage.Delete("backend-test")
			if err != nil {
				t.Fatalf("Failed to delete dataset with %s backend: %v", tc.name, err)
			}

			// Verify deletion
			_, err = storage.Retrieve("backend-test")
			if err == nil {
				t.Errorf("Dataset should be deleted with %s backend", tc.name)
			}
		})
	}
}

// TestProtobufStorage_DataPersistenceAccuracy tests that data persists accurately
func TestProtobufStorage_DataPersistenceAccuracy(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create comprehensive test data with all field types
	originalData := DatasetData{
		Players: []Player{
			{
				UID:                 123456789,
				Name:                "Persistence Test Player",
				Position:            "RW",
				Age:                 "23",
				Club:                "Accuracy FC",
				Division:            "Premier League",
				TransferValue:       "£25M",
				Wage:                "£100K",
				Personality:         "Determined",
				MediaHandling:       "Confident",
				Nationality:         "Brazil",
				NationalityISO:      "BRA",
				NationalityFIFACode: "BRA",
				AttributeMasked:     true,
				Attributes: map[string]string{
					"Pace":        "19",
					"Dribbling":   "18",
					"Crossing":    "16",
					"Finishing":   "15",
					"Technique":   "17",
				},
				NumericAttributes: map[string]int{
					"Pace":        19,
					"Dribbling":   18,
					"Crossing":    16,
					"Finishing":   15,
					"Technique":   17,
				},
				PerformanceStatsNumeric: map[string]float64{
					"Goals":              8.5,
					"Assists":            14.2,
					"KeyPasses":          2.8,
					"DribblesPerGame":    4.3,
					"CrossAccuracy":      78.5,
				},
				PerformancePercentiles: map[string]map[string]float64{
					"Attacking": {
						"Goals":     82.3,
						"Assists":   94.1,
						"KeyPasses": 87.6,
					},
					"Dribbling": {
						"DribblesPerGame":    91.2,
						"DribbleSuccessRate": 85.7,
					},
					"Crossing": {
						"CrossAccuracy": 89.4,
						"Crosses":       76.8,
					},
				},
				ParsedPositions: []string{"RW", "RM", "RF", "AMR"},
				ShortPositions:  []string{"RW", "RM"},
				PositionGroups:  []string{"Winger", "Midfielder"},
				PAC:             19,
				SHO:             15,
				PAS:             16,
				DRI:             18,
				DEF:             8,
				PHY:             14,
				GK:              3,
				DIV:             2,
				HAN:             1,
				REF:             4,
				KIC:             12,
				SPD:             18,
				POS:             16,
				Overall:         84,
				BestRoleOverall: "Inside Forward (Attack)",
				RoleSpecificOveralls: []RoleOverallScore{
					{RoleName: "Inside Forward (Attack)", Score: 87},
					{RoleName: "Winger (Attack)", Score: 85},
					{RoleName: "Winger (Support)", Score: 82},
					{RoleName: "Wide Midfielder (Attack)", Score: 79},
				},
				TransferValueAmount: 25000000,
				WageAmount:          100000,
			},
		},
		CurrencySymbol: "£",
	}

	// Store the data
	err := storage.Store("persistence-test", originalData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve the data
	retrievedData, err := storage.Retrieve("persistence-test")
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Perform comprehensive data accuracy checks
	if len(retrievedData.Players) != len(originalData.Players) {
		t.Fatalf("Player count mismatch: expected %d, got %d", 
			len(originalData.Players), len(retrievedData.Players))
	}

	if retrievedData.CurrencySymbol != originalData.CurrencySymbol {
		t.Errorf("Currency symbol mismatch: expected %s, got %s", 
			originalData.CurrencySymbol, retrievedData.CurrencySymbol)
	}

	// Check detailed player data
	original := originalData.Players[0]
	retrieved := retrievedData.Players[0]

	// Basic fields
	if retrieved.UID != original.UID {
		t.Errorf("UID mismatch: expected %d, got %d", original.UID, retrieved.UID)
	}
	if retrieved.Name != original.Name {
		t.Errorf("Name mismatch: expected %s, got %s", original.Name, retrieved.Name)
	}
	if retrieved.AttributeMasked != original.AttributeMasked {
		t.Errorf("AttributeMasked mismatch: expected %t, got %t", original.AttributeMasked, retrieved.AttributeMasked)
	}

	// Maps
	if len(retrieved.Attributes) != len(original.Attributes) {
		t.Errorf("Attributes map length mismatch: expected %d, got %d", 
			len(original.Attributes), len(retrieved.Attributes))
	}
	for key, value := range original.Attributes {
		if retrieved.Attributes[key] != value {
			t.Errorf("Attributes[%s] mismatch: expected %s, got %s", key, value, retrieved.Attributes[key])
		}
	}

	if len(retrieved.NumericAttributes) != len(original.NumericAttributes) {
		t.Errorf("NumericAttributes map length mismatch: expected %d, got %d", 
			len(original.NumericAttributes), len(retrieved.NumericAttributes))
	}
	for key, value := range original.NumericAttributes {
		if retrieved.NumericAttributes[key] != value {
			t.Errorf("NumericAttributes[%s] mismatch: expected %d, got %d", key, value, retrieved.NumericAttributes[key])
		}
	}

	// Performance stats
	if len(retrieved.PerformanceStatsNumeric) != len(original.PerformanceStatsNumeric) {
		t.Errorf("PerformanceStatsNumeric map length mismatch: expected %d, got %d", 
			len(original.PerformanceStatsNumeric), len(retrieved.PerformanceStatsNumeric))
	}
	for key, value := range original.PerformanceStatsNumeric {
		if retrieved.PerformanceStatsNumeric[key] != value {
			t.Errorf("PerformanceStatsNumeric[%s] mismatch: expected %f, got %f", 
				key, value, retrieved.PerformanceStatsNumeric[key])
		}
	}

	// Nested performance percentiles
	if len(retrieved.PerformancePercentiles) != len(original.PerformancePercentiles) {
		t.Errorf("PerformancePercentiles map length mismatch: expected %d, got %d", 
			len(original.PerformancePercentiles), len(retrieved.PerformancePercentiles))
	}
	for category, percentiles := range original.PerformancePercentiles {
		retrievedPercentiles, exists := retrieved.PerformancePercentiles[category]
		if !exists {
			t.Errorf("PerformancePercentiles category %s missing", category)
			continue
		}
		for stat, value := range percentiles {
			if retrievedPercentiles[stat] != value {
				t.Errorf("PerformancePercentiles[%s][%s] mismatch: expected %f, got %f", 
					category, stat, value, retrievedPercentiles[stat])
			}
		}
	}

	// Slices
	if len(retrieved.ParsedPositions) != len(original.ParsedPositions) {
		t.Errorf("ParsedPositions length mismatch: expected %d, got %d", 
			len(original.ParsedPositions), len(retrieved.ParsedPositions))
	}
	for i, pos := range original.ParsedPositions {
		if i < len(retrieved.ParsedPositions) && retrieved.ParsedPositions[i] != pos {
			t.Errorf("ParsedPositions[%d] mismatch: expected %s, got %s", i, pos, retrieved.ParsedPositions[i])
		}
	}

	// Role-specific overalls
	if len(retrieved.RoleSpecificOveralls) != len(original.RoleSpecificOveralls) {
		t.Errorf("RoleSpecificOveralls length mismatch: expected %d, got %d", 
			len(original.RoleSpecificOveralls), len(retrieved.RoleSpecificOveralls))
	}
	for i, role := range original.RoleSpecificOveralls {
		if i < len(retrieved.RoleSpecificOveralls) {
			retrievedRole := retrieved.RoleSpecificOveralls[i]
			if retrievedRole.RoleName != role.RoleName {
				t.Errorf("RoleSpecificOveralls[%d].RoleName mismatch: expected %s, got %s", 
					i, role.RoleName, retrievedRole.RoleName)
			}
			if retrievedRole.Score != role.Score {
				t.Errorf("RoleSpecificOveralls[%d].Score mismatch: expected %d, got %d", 
					i, role.Score, retrievedRole.Score)
			}
		}
	}

	// Individual stats
	statsToCheck := []struct {
		name     string
		original int
		retrieved int
	}{
		{"PAC", original.PAC, retrieved.PAC},
		{"SHO", original.SHO, retrieved.SHO},
		{"PAS", original.PAS, retrieved.PAS},
		{"DRI", original.DRI, retrieved.DRI},
		{"DEF", original.DEF, retrieved.DEF},
		{"PHY", original.PHY, retrieved.PHY},
		{"Overall", original.Overall, retrieved.Overall},
	}

	for _, stat := range statsToCheck {
		if stat.retrieved != stat.original {
			t.Errorf("%s mismatch: expected %d, got %d", stat.name, stat.original, stat.retrieved)
		}
	}

	// Amount fields
	if retrieved.TransferValueAmount != original.TransferValueAmount {
		t.Errorf("TransferValueAmount mismatch: expected %d, got %d", 
			original.TransferValueAmount, retrieved.TransferValueAmount)
	}
	if retrieved.WageAmount != original.WageAmount {
		t.Errorf("WageAmount mismatch: expected %d, got %d", 
			original.WageAmount, retrieved.WageAmount)
	}
}

// TestProtobufStorage_APIResponseCompatibility tests that API responses remain unchanged
func TestProtobufStorage_APIResponseCompatibility(t *testing.T) {
	// Create both JSON and protobuf storage with the same backend data
	jsonBackend := CreateInMemoryStorage()
	protobufBackend := CreateInMemoryStorage()
	protobufStorage := CreateProtobufStorage(protobufBackend)

	// Create test data
	testData := DatasetData{
		Players: []Player{
			{
				UID:                     1,
				Name:                    "API Test Player",
				Position:                "ST",
				Age:                     "27",
				Club:                    "API FC",
				Division:                "Premier League",
				TransferValue:           "£15M",
				Wage:                    "£75K",
				Nationality:             "France",
				NationalityISO:          "FRA",
				NationalityFIFACode:     "FRA",
				Attributes:              map[string]string{"Finishing": "17", "Pace": "16"},
				NumericAttributes:       map[string]int{"Finishing": 17, "Pace": 16},
				PerformanceStatsNumeric: map[string]float64{"Goals": 18.0, "Assists": 6.5},
				PerformancePercentiles: map[string]map[string]float64{
					"Attacking": {"Goals": 92.1, "Assists": 78.3},
				},
				ParsedPositions: []string{"ST", "CF"},
				ShortPositions:  []string{"ST"},
				PositionGroups:  []string{"Forward"},
				PAC:             16,
				SHO:             17,
				PAS:             13,
				DRI:             14,
				DEF:             6,
				PHY:             15,
				Overall:         83,
				BestRoleOverall: "Advanced Forward",
				RoleSpecificOveralls: []RoleOverallScore{
					{RoleName: "Advanced Forward", Score: 86},
					{RoleName: "Poacher", Score: 84},
				},
				TransferValueAmount: 15000000,
				WageAmount:          75000,
			},
		},
		CurrencySymbol: "£",
	}

	// Store data in both storages
	err := jsonBackend.Store("api-test", testData)
	if err != nil {
		t.Fatalf("Failed to store data in JSON backend: %v", err)
	}

	err = protobufStorage.Store("api-test", testData)
	if err != nil {
		t.Fatalf("Failed to store data in protobuf storage: %v", err)
	}

	// Retrieve data from both storages
	jsonData, err := jsonBackend.Retrieve("api-test")
	if err != nil {
		t.Fatalf("Failed to retrieve data from JSON backend: %v", err)
	}

	protobufData, err := protobufStorage.Retrieve("api-test")
	if err != nil {
		t.Fatalf("Failed to retrieve data from protobuf storage: %v", err)
	}

	// Compare the retrieved data - they should be identical for API compatibility
	if len(jsonData.Players) != len(protobufData.Players) {
		t.Errorf("Player count mismatch: JSON %d vs Protobuf %d", 
			len(jsonData.Players), len(protobufData.Players))
	}

	if jsonData.CurrencySymbol != protobufData.CurrencySymbol {
		t.Errorf("Currency symbol mismatch: JSON %s vs Protobuf %s", 
			jsonData.CurrencySymbol, protobufData.CurrencySymbol)
	}

	// Compare player data in detail
	if len(jsonData.Players) > 0 && len(protobufData.Players) > 0 {
		jsonPlayer := jsonData.Players[0]
		protobufPlayer := protobufData.Players[0]

		// Basic fields
		if jsonPlayer.UID != protobufPlayer.UID {
			t.Errorf("Player UID mismatch: JSON %d vs Protobuf %d", jsonPlayer.UID, protobufPlayer.UID)
		}
		if jsonPlayer.Name != protobufPlayer.Name {
			t.Errorf("Player Name mismatch: JSON %s vs Protobuf %s", jsonPlayer.Name, protobufPlayer.Name)
		}
		if jsonPlayer.Overall != protobufPlayer.Overall {
			t.Errorf("Player Overall mismatch: JSON %d vs Protobuf %d", jsonPlayer.Overall, protobufPlayer.Overall)
		}

		// Maps should be identical
		for key, value := range jsonPlayer.Attributes {
			if protobufPlayer.Attributes[key] != value {
				t.Errorf("Attributes[%s] mismatch: JSON %s vs Protobuf %s", 
					key, value, protobufPlayer.Attributes[key])
			}
		}

		// Role-specific overalls
		if len(jsonPlayer.RoleSpecificOveralls) != len(protobufPlayer.RoleSpecificOveralls) {
			t.Errorf("RoleSpecificOveralls length mismatch: JSON %d vs Protobuf %d", 
				len(jsonPlayer.RoleSpecificOveralls), len(protobufPlayer.RoleSpecificOveralls))
		}
	}

	t.Log("API response compatibility verified - JSON and Protobuf storage return identical data")
}

// TestProtobufStorage_ConcurrentOperations tests protobuf storage under concurrent access
func TestProtobufStorage_ConcurrentOperations(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create test data
	createTestData := func(id int) DatasetData {
		return DatasetData{
			Players: []Player{
				{
					UID:                     int64(id),
					Name:                    fmt.Sprintf("Concurrent Player %d", id),
					Position:                "CM",
					Age:                     "25",
					Club:                    fmt.Sprintf("Concurrent FC %d", id),
					Overall:                 75 + (id % 10),
					Attributes:              map[string]string{"Passing": fmt.Sprintf("%d", 15+(id%5))},
					NumericAttributes:       map[string]int{"Passing": 15 + (id % 5)},
					PerformanceStatsNumeric: map[string]float64{"Assists": float64(id) * 0.5},
					RoleSpecificOveralls: []RoleOverallScore{
						{RoleName: "Central Midfielder", Score: 75 + (id % 10)},
					},
				},
			},
			CurrencySymbol: "£",
		}
	}

	// Test concurrent stores
	const numGoroutines = 10
	const numOperations = 5

	// Store operations
	storeCh := make(chan error, numGoroutines*numOperations)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				datasetID := fmt.Sprintf("concurrent-%d-%d", goroutineID, j)
				testData := createTestData(goroutineID*numOperations + j)
				err := storage.Store(datasetID, testData)
				storeCh <- err
			}
		}(i)
	}

	// Check store results
	for i := 0; i < numGoroutines*numOperations; i++ {
		err := <-storeCh
		if err != nil {
			t.Errorf("Concurrent store operation failed: %v", err)
		}
	}

	// Test concurrent retrieves
	retrieveCh := make(chan error, numGoroutines*numOperations)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				datasetID := fmt.Sprintf("concurrent-%d-%d", goroutineID, j)
				_, err := storage.Retrieve(datasetID)
				retrieveCh <- err
			}
		}(i)
	}

	// Check retrieve results
	for i := 0; i < numGoroutines*numOperations; i++ {
		err := <-retrieveCh
		if err != nil {
			t.Errorf("Concurrent retrieve operation failed: %v", err)
		}
	}

	// Test concurrent list operations
	listCh := make(chan error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := storage.List()
			listCh <- err
		}()
	}

	// Check list results
	for i := 0; i < numGoroutines; i++ {
		err := <-listCh
		if err != nil {
			t.Errorf("Concurrent list operation failed: %v", err)
		}
	}

	t.Logf("Successfully completed %d concurrent operations", numGoroutines*numOperations*2+numGoroutines)
}

// TestProtobufStorage_LargeDatasetIntegration tests protobuf storage with very large datasets
func TestProtobufStorage_LargeDatasetIntegration(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Create a very large dataset
	const numPlayers = 5000
	players := make([]Player, numPlayers)

	for i := 0; i < numPlayers; i++ {
		players[i] = Player{
			UID:                     int64(i + 1),
			Name:                    fmt.Sprintf("Large Dataset Player %d", i+1),
			Position:                []string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8],
			Age:                     fmt.Sprintf("%d", 18+(i%15)),
			Club:                    fmt.Sprintf("Large Club %d", (i%100)+1),
			Division:                []string{"Premier League", "Championship", "League One", "League Two"}[i%4],
			TransferValue:           fmt.Sprintf("£%dM", (i%50)+1),
			Wage:                    fmt.Sprintf("£%dK", (i%100)+10),
			Nationality:             []string{"England", "Spain", "Germany", "France", "Italy", "Brazil", "Argentina"}[i%7],
			NationalityISO:          []string{"ENG", "ESP", "GER", "FRA", "ITA", "BRA", "ARG"}[i%7],
			NationalityFIFACode:     []string{"ENG", "ESP", "GER", "FRA", "ITA", "BRA", "ARG"}[i%7],
			AttributeMasked:         i%2 == 0,
			Attributes: map[string]string{
				"Pace":      fmt.Sprintf("%d", 10+(i%11)),
				"Shooting":  fmt.Sprintf("%d", 8+(i%13)),
				"Passing":   fmt.Sprintf("%d", 12+(i%9)),
				"Dribbling": fmt.Sprintf("%d", 9+(i%12)),
				"Defending": fmt.Sprintf("%d", 7+(i%14)),
				"Physical":  fmt.Sprintf("%d", 11+(i%10)),
			},
			NumericAttributes: map[string]int{
				"Pace":      10 + (i % 11),
				"Shooting":  8 + (i % 13),
				"Passing":   12 + (i % 9),
				"Dribbling": 9 + (i % 12),
				"Defending": 7 + (i % 14),
				"Physical":  11 + (i % 10),
			},
			PerformanceStatsNumeric: map[string]float64{
				"Goals":         float64(i%30) + 0.5,
				"Assists":       float64(i%20) + 0.3,
				"KeyPasses":     float64(i%10) + 1.2,
				"Tackles":       float64(i%15) + 0.8,
				"Interceptions": float64(i%12) + 0.4,
			},
			PerformancePercentiles: map[string]map[string]float64{
				"Attacking": {
					"Goals":     float64((i*7)%100) + 0.1,
					"Assists":   float64((i*11)%100) + 0.2,
					"KeyPasses": float64((i*13)%100) + 0.3,
				},
				"Defending": {
					"Tackles":       float64((i*17)%100) + 0.4,
					"Interceptions": float64((i*19)%100) + 0.5,
					"Clearances":    float64((i*23)%100) + 0.6,
				},
			},
			ParsedPositions: []string{[]string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8]},
			ShortPositions:  []string{[]string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8]},
			PositionGroups:  []string{[]string{"Forward", "Midfielder", "Defender", "Goalkeeper"}[i%4]},
			PAC:             10 + (i % 11),
			SHO:             8 + (i % 13),
			PAS:             12 + (i % 9),
			DRI:             9 + (i % 12),
			DEF:             7 + (i % 14),
			PHY:             11 + (i % 10),
			GK:              (i % 20) + 1,
			Overall:         50 + (i % 40),
			BestRoleOverall: fmt.Sprintf("Best Role %d", i%10),
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: fmt.Sprintf("Role A %d", i%5), Score: 60 + (i % 30)},
				{RoleName: fmt.Sprintf("Role B %d", i%5), Score: 55 + (i % 35)},
			},
			TransferValueAmount: int64((i%50+1) * 1000000),
			WageAmount:          int64((i%100+10) * 1000),
		}
	}

	largeDataset := DatasetData{
		Players:        players,
		CurrencySymbol: "£",
	}

	// Measure storage time
	start := time.Now()
	err := storage.Store("large-integration-test", largeDataset)
	storeTime := time.Since(start)
	if err != nil {
		t.Fatalf("Failed to store large dataset: %v", err)
	}

	// Measure retrieval time
	start = time.Now()
	retrievedDataset, err := storage.Retrieve("large-integration-test")
	retrieveTime := time.Since(start)
	if err != nil {
		t.Fatalf("Failed to retrieve large dataset: %v", err)
	}

	// Verify data integrity
	if len(retrievedDataset.Players) != len(largeDataset.Players) {
		t.Errorf("Player count mismatch: expected %d, got %d", 
			len(largeDataset.Players), len(retrievedDataset.Players))
	}

	if retrievedDataset.CurrencySymbol != largeDataset.CurrencySymbol {
		t.Errorf("Currency symbol mismatch: expected %s, got %s", 
			largeDataset.CurrencySymbol, retrievedDataset.CurrencySymbol)
	}

	// Spot check random players for data integrity
	checkIndices := []int{0, 100, 1000, 2500, 4999}
	for _, idx := range checkIndices {
		if idx < len(retrievedDataset.Players) {
			original := largeDataset.Players[idx]
			retrieved := retrievedDataset.Players[idx]

			if retrieved.UID != original.UID {
				t.Errorf("Player %d UID mismatch: expected %d, got %d", idx, original.UID, retrieved.UID)
			}
			if retrieved.Name != original.Name {
				t.Errorf("Player %d Name mismatch: expected %s, got %s", idx, original.Name, retrieved.Name)
			}
			if retrieved.Overall != original.Overall {
				t.Errorf("Player %d Overall mismatch: expected %d, got %d", idx, original.Overall, retrieved.Overall)
			}
		}
	}

	t.Logf("Large dataset integration test completed successfully:")
	t.Logf("- Players: %d", numPlayers)
	t.Logf("- Store time: %v", storeTime)
	t.Logf("- Retrieve time: %v", retrieveTime)
	t.Logf("- Data integrity verified")
}

// TestProtobufStorage_ErrorRecoveryIntegration tests error recovery scenarios
func TestProtobufStorage_ErrorRecoveryIntegration(t *testing.T) {
	backend := CreateInMemoryStorage()
	storage := CreateProtobufStorage(backend)

	// Test 1: Store data that might cause protobuf issues but should fallback gracefully
	problematicData := DatasetData{
		Players: []Player{
			{
				UID:  1,
				Name: "Error Recovery Player",
				// Add potentially problematic data
				PerformancePercentiles: map[string]map[string]float64{
					"test": nil, // This might cause issues
				},
				Attributes: map[string]string{
					"": "empty_key_test", // Empty key
				},
			},
		},
		CurrencySymbol: "£",
	}

	// Should succeed with fallback
	err := storage.Store("error-recovery-test", problematicData)
	if err != nil {
		t.Fatalf("Store should succeed with fallback: %v", err)
	}

	// Should be able to retrieve
	retrievedData, err := storage.Retrieve("error-recovery-test")
	if err != nil {
		t.Fatalf("Retrieve should succeed with fallback: %v", err)
	}

	// Basic data should be preserved
	if len(retrievedData.Players) != len(problematicData.Players) {
		t.Errorf("Player count mismatch after error recovery: expected %d, got %d", 
			len(problematicData.Players), len(retrievedData.Players))
	}

	// Test 2: Corrupt protobuf data recovery
	corruptProtobufData := DatasetData{
		Players: []Player{{
			UID:      -1,
			Name:     "__PROTOBUF_DATA__",
			Position: "corrupted-base64-data-!@#$%^&*()", // Invalid base64
			Club:     "PROTOBUF_STORAGE",
		}},
		CurrencySymbol: "__PROTOBUF_MARKER__",
	}

	// Store corrupt data directly to backend
	err = backend.Store("corrupt-protobuf-test", corruptProtobufData)
	if err != nil {
		t.Fatalf("Failed to store corrupt protobuf data: %v", err)
	}

	// Should fallback to JSON when retrieving corrupt protobuf data
	retrievedCorruptData, err := storage.Retrieve("corrupt-protobuf-test")
	if err != nil {
		t.Fatalf("Should be able to retrieve corrupt protobuf data with fallback: %v", err)
	}

	// Should get the raw corrupt data back (JSON fallback)
	if retrievedCorruptData.CurrencySymbol != "__PROTOBUF_MARKER__" {
		t.Error("Should fallback to JSON for corrupt protobuf data")
	}

	t.Log("Error recovery integration tests completed successfully")
}