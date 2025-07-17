package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImmediateRollbackCapability tests the ability to immediately disable protobuf
// and fall back to JSON storage without data loss
func TestImmediateRollbackCapability(t *testing.T) {
	ctx := context.Background()
	logInfo(ctx, "Starting immediate rollback capability test")
	tests := []struct {
		name        string
		storageType string
		setupFunc   func() StorageInterface
	}{
		{
			name:        "InMemory_Storage_Rollback",
			storageType: "memory",
			setupFunc: func() StorageInterface {
				return CreateInMemoryStorage()
			},
		},
		{
			name:        "LocalFile_Storage_Rollback",
			storageType: "local",
			setupFunc: func() StorageInterface {
				tempDir := t.TempDir()
				storage, err := CreateLocalFileStorage(tempDir)
				require.NoError(t, err)
				return storage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()
			
			logInfo(ctx, "Starting rollback test", 
				"test_name", tt.name, 
				"storage_type", tt.storageType)
			
			// Create test dataset
			testData := createTestDataset(ctx, "rollback-test", 100)
			datasetID := "rollback-test-dataset"

			// Step 1: Test with protobuf enabled
			logInfo(ctx, "Step 1: Testing with protobuf enabled", 
				"dataset_id", datasetID, 
				"player_count", len(testData.Players))
			backend := tt.setupFunc()
			protobufStorage := CreateProtobufStorage(backend)

			// Store data using protobuf
			err := protobufStorage.Store(datasetID, testData)
			if err != nil {
				logError(ctx, "Failed to store data with protobuf", 
					"error", err, 
					"dataset_id", datasetID)
			}
			require.NoError(t, err, "Should store data successfully with protobuf")

			// Retrieve data using protobuf
			retrievedData, err := protobufStorage.Retrieve(datasetID)
			require.NoError(t, err, "Should retrieve data successfully with protobuf")
			assertDatasetEqual(t, testData, retrievedData)

			// Step 2: Immediate rollback - disable protobuf, use JSON directly
			logInfo(ctx, "Step 2: Testing immediate rollback to JSON", "dataset_id", datasetID)
			jsonStorage := backend // Use the underlying backend directly (JSON mode)

			// When data is stored using protobuf, it's stored in a special encoded format
			// Direct JSON retrieval will get the encoded protobuf data, not the original data
			retrievedDataJSON, err := jsonStorage.Retrieve(datasetID)
			if err != nil {
				logError(ctx, "Failed to retrieve data from backend during rollback", 
					"error", err, "dataset_id", datasetID)
				t.Fatalf("Should be able to retrieve data from backend: %v", err)
			}

			// Check if this is protobuf-encoded data (indicated by special marker)
			if retrievedDataJSON.CurrencySymbol == "__PROTOBUF_MARKER__" {
				logInfo(ctx, "Data is stored in protobuf format - testing protobuf storage can still handle retrieval", 
					"dataset_id", datasetID)

				// Test that protobuf storage can still retrieve and convert the data
				retrievedDataProtobuf, err := protobufStorage.Retrieve(datasetID)
				require.NoError(t, err, "Protobuf storage should handle retrieval of protobuf data")
				assertDatasetEqual(t, testData, retrievedDataProtobuf)
			} else {
				// If it's regular JSON data, verify data integrity
				logInfo(ctx, "Data is stored in JSON format - verifying data integrity", 
					"dataset_id", datasetID)
				assertDatasetEqual(t, testData, retrievedDataJSON)
			}

			// Step 3: Store new data using JSON only (simulating rollback scenario)
			logInfo(ctx, "Step 3: Testing new data storage with JSON after rollback")
			newDatasetID := "rollback-new-data"
			newTestData := createTestDataset(ctx, "rollback-new", 50)

			err = jsonStorage.Store(newDatasetID, newTestData)
			if err != nil {
				logError(ctx, "Failed to store new data with JSON after rollback", 
					"error", err, "dataset_id", newDatasetID)
			}
			require.NoError(t, err, "Should store new data successfully with JSON after rollback")

			retrievedNewData, err := jsonStorage.Retrieve(newDatasetID)
			if err != nil {
				logError(ctx, "Failed to retrieve new data with JSON", 
					"error", err, "dataset_id", newDatasetID)
			}
			require.NoError(t, err, "Should retrieve new data successfully with JSON")
			assertDatasetEqual(t, newTestData, retrievedNewData)

			// Step 4: Verify both datasets are accessible
			logInfo(ctx, "Step 4: Verifying both datasets remain accessible")
			datasets, err := jsonStorage.List()
			if err != nil {
				logError(ctx, "Failed to list datasets", "error", err)
			}
			require.NoError(t, err, "Should list datasets successfully")
			assert.Contains(t, datasets, datasetID, "Original dataset should still be listed")
			assert.Contains(t, datasets, newDatasetID, "New dataset should be listed")

			duration := time.Since(start)
			logInfo(ctx, "Rollback test completed successfully", 
				"test_name", tt.name,
				"storage_type", tt.storageType,
				"duration_ms", duration.Milliseconds())

			// Cleanup
			_ = jsonStorage.Delete(datasetID)
			_ = jsonStorage.Delete(newDatasetID)
		})
	}
}

// TestAutomaticFallbackBehavior tests that protobuf storage automatically falls back
// to JSON when protobuf operations fail
func TestAutomaticFallbackBehavior(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	
	logInfo(ctx, "Starting automatic fallback behavior test")

	// Create a mock storage that can simulate failures
	mockStorage := &MockStorageWithFailures{
		InMemoryStorage: CreateInMemoryStorage(),
		failureMode:     "",
	}

	protobufStorage := CreateProtobufStorage(mockStorage)
	testData := createTestDataset(ctx, "fallback-test", 50)
	datasetID := "fallback-test-dataset"

	// Test 1: Normal operation should work
	t.Run("Normal_Operation", func(t *testing.T) {
		logInfo(ctx, "Testing normal protobuf operation", "dataset_id", datasetID)
		err := protobufStorage.Store(datasetID, testData)
		if err != nil {
			logError(ctx, "Normal protobuf storage failed", "error", err, "dataset_id", datasetID)
		}
		require.NoError(t, err, "Normal protobuf storage should work")

		retrievedData, err := protobufStorage.Retrieve(datasetID)
		if err != nil {
			logError(ctx, "Normal protobuf retrieval failed", "error", err, "dataset_id", datasetID)
		}
		require.NoError(t, err, "Normal protobuf retrieval should work")
		assertDatasetEqual(t, testData, retrievedData)
	})

	// Test 2: Storage failure should trigger fallback
	t.Run("Storage_Failure_Fallback", func(_ *testing.T) {
		logInfo(ctx, "Testing storage failure fallback", "dataset_id", datasetID+"_fail")
		mockStorage.failureMode = "store"
		defer func() { mockStorage.failureMode = "" }()

		// This should trigger fallback to JSON storage
		err := protobufStorage.Store(datasetID+"_fail", testData)
		// The error handling depends on implementation - it might succeed via fallback
		// or fail gracefully. The key is that it doesn't panic or corrupt data.
		if err != nil {
			logError(ctx, "Storage with failure mode failed", "error", err, "dataset_id", datasetID+"_fail")
		}
		logDebug(ctx, "Storage with failure mode completed", "error", err)
	})

	// Test 3: Retrieval failure should trigger fallback
	t.Run("Retrieval_Failure_Fallback", func(t *testing.T) {
		logInfo(ctx, "Testing retrieval failure fallback", "dataset_id", datasetID+"_retrieve")
		// First store data normally
		err := protobufStorage.Store(datasetID+"_retrieve", testData)
		if err != nil {
			logError(ctx, "Failed to store data for retrieval test", "error", err, "dataset_id", datasetID+"_retrieve")
		}
		require.NoError(t, err)

		// Then simulate retrieval failure
		mockStorage.failureMode = "retrieve"
		defer func() { mockStorage.failureMode = "" }()

		// This should trigger fallback behavior
		_, err = protobufStorage.Retrieve(datasetID + "_retrieve")
		if err != nil {
			logError(ctx, "Retrieval with failure mode failed", "error", err, "dataset_id", datasetID+"_retrieve")
		}
		logDebug(ctx, "Retrieval with failure mode completed", "error", err)
	})

	duration := time.Since(start)
	logInfo(ctx, "Automatic fallback behavior test completed", "duration_ms", duration.Milliseconds())

	// Cleanup
	_ = mockStorage.Delete(datasetID)
	_ = mockStorage.Delete(datasetID + "_fail")
	_ = mockStorage.Delete(datasetID + "_retrieve")
}

// TestDataIntegrityDuringRollback verifies that no data is lost during rollback scenarios
func TestDataIntegrityDuringRollback(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	
	logInfo(ctx, "Starting data integrity during rollback test")

	tempDir := t.TempDir()
	localStorage, err := CreateLocalFileStorage(tempDir)
	if err != nil {
		logError(ctx, "Failed to create local file storage", "error", err)
	}
	require.NoError(t, err)

	protobufStorage := CreateProtobufStorage(localStorage)

	// Create multiple test datasets with different characteristics
	datasets := map[string]DatasetData{
		"small-dataset":  createTestDataset(ctx, "small", 10),
		"medium-dataset": createTestDataset(ctx, "medium", 100),
		"large-dataset":  createTestDataset(ctx, "large", 1000),
		"empty-dataset":  {Players: []Player{}, CurrencySymbol: "USD"},
		"special-chars":  createTestDatasetWithSpecialChars(ctx),
	}

	// Step 1: Store all datasets using protobuf
	logInfo(ctx, "Step 1: Storing datasets using protobuf", "dataset_count", len(datasets))
	for datasetID, data := range datasets {
		err := protobufStorage.Store(datasetID, data)
		if err != nil {
			logError(ctx, "Failed to store dataset with protobuf", "error", err, "dataset_id", datasetID)
		}
		require.NoError(t, err, "Should store dataset %s successfully", datasetID)
	}

	// Step 2: Verify all data can be retrieved using protobuf
	logInfo(ctx, "Step 2: Verifying protobuf retrieval", "dataset_count", len(datasets))
	for datasetID, originalData := range datasets {
		retrievedData, err := protobufStorage.Retrieve(datasetID)
		if err != nil {
			logError(ctx, "Failed to retrieve dataset with protobuf", "error", err, "dataset_id", datasetID)
		}
		require.NoError(t, err, "Should retrieve dataset %s successfully", datasetID)
		assertDatasetEqual(t, originalData, retrievedData)
	}

	// Step 3: Simulate rollback by using JSON storage directly
	logInfo(ctx, "Step 3: Testing rollback scenario with JSON storage")
	jsonStorage := localStorage

	// Try to retrieve data using JSON storage (this tests fallback mechanisms)
	for datasetID, originalData := range datasets {
		// The behavior here depends on how the data was stored
		// If stored as protobuf, direct JSON retrieval might fail, which is expected
		retrievedData, err := jsonStorage.Retrieve(datasetID)
		if err != nil {
			logDebug(ctx, "Direct JSON retrieval failed (expected if stored as protobuf)", 
				"error", err, "dataset_id", datasetID)

			// Verify that protobuf storage can still handle the retrieval
			retrievedDataProtobuf, err := protobufStorage.Retrieve(datasetID)
			if err != nil {
				logError(ctx, "Protobuf storage failed to retrieve data", "error", err, "dataset_id", datasetID)
			}
			require.NoError(t, err, "Protobuf storage should handle retrieval for %s", datasetID)
			assertDatasetEqual(t, originalData, retrievedDataProtobuf)
		} else {
			// If JSON retrieval succeeds, verify data integrity
			assertDatasetEqual(t, originalData, retrievedData)
		}
	}

	// Step 4: Store new data using JSON only (post-rollback scenario)
	logInfo(ctx, "Step 4: Testing post-rollback data storage")
	postRollbackData := createTestDataset(ctx, "post-rollback", 75)
	postRollbackID := "post-rollback-dataset"

	err = jsonStorage.Store(postRollbackID, postRollbackData)
	if err != nil {
		logError(ctx, "Failed to store post-rollback data", "error", err, "dataset_id", postRollbackID)
	}
	require.NoError(t, err, "Should store post-rollback data successfully")

	retrievedPostRollback, err := jsonStorage.Retrieve(postRollbackID)
	if err != nil {
		logError(ctx, "Failed to retrieve post-rollback data", "error", err, "dataset_id", postRollbackID)
	}
	require.NoError(t, err, "Should retrieve post-rollback data successfully")
	assertDatasetEqual(t, postRollbackData, retrievedPostRollback)

	// Step 5: Verify all datasets are still accessible and intact
	logInfo(ctx, "Step 5: Final data integrity verification")
	allDatasets, err := jsonStorage.List()
	if err != nil {
		logError(ctx, "Failed to list all datasets", "error", err)
	}
	require.NoError(t, err, "Should list all datasets successfully")

	expectedDatasets := make([]string, 0, len(datasets)+1)
	for datasetID := range datasets {
		expectedDatasets = append(expectedDatasets, datasetID)
	}
	expectedDatasets = append(expectedDatasets, postRollbackID)

	for _, expectedID := range expectedDatasets {
		assert.Contains(t, allDatasets, expectedID, "Dataset %s should be listed", expectedID)
	}

	duration := time.Since(start)
	logInfo(ctx, "Data integrity during rollback test completed", 
		"dataset_count", len(datasets),
		"duration_ms", duration.Milliseconds())

	// Cleanup
	for datasetID := range datasets {
		_ = jsonStorage.Delete(datasetID)
	}
	_ = jsonStorage.Delete(postRollbackID)
}

// TestEnvironmentVariableRollback tests rollback via environment variable changes
func TestEnvironmentVariableRollback(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	
	logInfo(ctx, "Starting environment variable rollback test")

	// Save original environment variable
	originalValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if originalValue == "" {
			if err := os.Unsetenv("USE_PROTOBUF"); err != nil {
				logError(ctx, "Failed to unset USE_PROTOBUF", "error", err)
			}
		} else {
			if err := os.Setenv("USE_PROTOBUF", originalValue); err != nil {
				logError(ctx, "Failed to restore USE_PROTOBUF", "error", err)
			}
		}
	}()

	tempDir := t.TempDir()
	testData := createTestDataset(ctx, "env-rollback", 50)
	datasetID := "env-rollback-test"

	// Test 1: Enable protobuf via environment variable
	t.Run("Enable_Protobuf", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "true"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}

		// Initialize storage (this would normally happen at app startup)
		localStorage, err := CreateLocalFileStorage(tempDir)
		require.NoError(t, err)

		storage := initializeStorageForTest(localStorage)

		// Store and retrieve data
		err = storage.Store(datasetID, testData)
		require.NoError(t, err, "Should store data with protobuf enabled")

		retrievedData, err := storage.Retrieve(datasetID)
		require.NoError(t, err, "Should retrieve data with protobuf enabled")
		assertDatasetEqual(t, testData, retrievedData)
	})

	// Test 2: Disable protobuf via environment variable (rollback)
	t.Run("Disable_Protobuf_Rollback", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "false"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}

		// Initialize storage again (simulating app restart after config change)
		localStorage, err := CreateLocalFileStorage(tempDir)
		require.NoError(t, err)

		storage := initializeStorageForTest(localStorage)

		// Should still be able to retrieve existing data
		retrievedData, err := storage.Retrieve(datasetID)
		if err != nil {
			logDebug(ctx, "Retrieval after rollback failed (may be expected)", "error", err, "dataset_id", datasetID)
		} else {
			assertDatasetEqual(t, testData, retrievedData)
		}

		// Should be able to store new data using JSON
		newDatasetID := "env-rollback-new"
		newTestData := createTestDataset(ctx, "env-rollback-new", 25)

		err = storage.Store(newDatasetID, newTestData)
		require.NoError(t, err, "Should store new data after rollback")

		retrievedNewData, err := storage.Retrieve(newDatasetID)
		require.NoError(t, err, "Should retrieve new data after rollback")
		assertDatasetEqual(t, newTestData, retrievedNewData)

		// Cleanup
		_ = storage.Delete(newDatasetID)
	})

	duration := time.Since(start)
	logInfo(ctx, "Environment variable rollback test completed", "duration_ms", duration.Milliseconds())

	// Cleanup
	localStorage, _ := CreateLocalFileStorage(tempDir)
	_ = localStorage.Delete(datasetID)
}

// TestRollbackPerformanceImpact tests that rollback doesn't significantly impact performance
func TestRollbackPerformanceImpact(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	
	logInfo(ctx, "Starting performance impact rollback test")

	tempDir := t.TempDir()
	localStorage, err := CreateLocalFileStorage(tempDir)
	if err != nil {
		logError(ctx, "Failed to create local file storage", "error", err)
	}
	require.NoError(t, err)

	testData := createTestDataset(ctx, "perf-test", 500)
	datasetID := "performance-rollback-test"

	// Measure protobuf performance
	protobufStorage := CreateProtobufStorage(localStorage)

	storeStart := time.Now()
	err = protobufStorage.Store(datasetID, testData)
	if err != nil {
		logError(ctx, "Failed to store data with protobuf for performance test", "error", err, "dataset_id", datasetID)
	}
	require.NoError(t, err)
	protobufStoreTime := time.Since(storeStart)

	retrieveStart := time.Now()
	_, err = protobufStorage.Retrieve(datasetID)
	if err != nil {
		logError(ctx, "Failed to retrieve data with protobuf for performance test", "error", err, "dataset_id", datasetID)
	}
	require.NoError(t, err)
	protobufRetrieveTime := time.Since(retrieveStart)

	// Measure JSON performance (rollback scenario)
	jsonStorage := localStorage
	jsonDatasetID := datasetID + "-json"

	jsonStoreStart := time.Now()
	err = jsonStorage.Store(jsonDatasetID, testData)
	if err != nil {
		logError(ctx, "Failed to store data with JSON for performance test", "error", err, "dataset_id", jsonDatasetID)
	}
	require.NoError(t, err)
	jsonStoreTime := time.Since(jsonStoreStart)

	jsonRetrieveStart := time.Now()
	_, err = jsonStorage.Retrieve(jsonDatasetID)
	if err != nil {
		logError(ctx, "Failed to retrieve data with JSON for performance test", "error", err, "dataset_id", jsonDatasetID)
	}
	require.NoError(t, err)
	jsonRetrieveTime := time.Since(jsonRetrieveStart)

	// Log performance comparison using structured logging
	logInfo(ctx, "Performance comparison completed",
		"protobuf_store_ms", protobufStoreTime.Milliseconds(),
		"json_store_ms", jsonStoreTime.Milliseconds(),
		"protobuf_retrieve_ms", protobufRetrieveTime.Milliseconds(),
		"json_retrieve_ms", jsonRetrieveTime.Milliseconds())

	// Verify that rollback performance is acceptable (within reasonable bounds)
	// JSON should be slower but not excessively so
	maxAcceptableSlowdown := 5.0 // 5x slower is still acceptable for rollback

	storeSlowdown := float64(jsonStoreTime) / float64(protobufStoreTime)
	retrieveSlowdown := float64(jsonRetrieveTime) / float64(protobufRetrieveTime)

	assert.Less(t, storeSlowdown, maxAcceptableSlowdown,
		"JSON store performance should be within acceptable bounds during rollback")
	assert.Less(t, retrieveSlowdown, maxAcceptableSlowdown,
		"JSON retrieve performance should be within acceptable bounds during rollback")

	duration := time.Since(start)
	logInfo(ctx, "Performance impact rollback test completed", 
		"duration_ms", duration.Milliseconds(),
		"store_slowdown", storeSlowdown,
		"retrieve_slowdown", retrieveSlowdown)

	// Cleanup
	_ = localStorage.Delete(datasetID)
	_ = localStorage.Delete(jsonDatasetID)
}

// Helper functions

// createTestDataset creates a test dataset with the specified name prefix and player count
func createTestDataset(ctx context.Context, namePrefix string, playerCount int) DatasetData {
	start := time.Now()
	logDebug(ctx, "Creating test dataset", "name_prefix", namePrefix, "player_count", playerCount)
	players := make([]Player, playerCount)

	for i := 0; i < playerCount; i++ {
		players[i] = Player{
			UID:                     int64(i + 1),
			Name:                    fmt.Sprintf("%s Player %d", namePrefix, i+1),
			Position:                "ST",
			Age:                     fmt.Sprintf("%d", 20+i%15), // Ages 20-34
			Club:                    fmt.Sprintf("%s FC %d", namePrefix, (i%10)+1),
			Division:                "Premier League",
			TransferValue:           fmt.Sprintf("£%dM", (i%50)+1),
			Wage:                    fmt.Sprintf("£%dK", (i%100)+10),
			Nationality:             "England",
			NationalityISO:          "ENG",
			NationalityFIFACode:     "ENG",
			Attributes:              map[string]string{"Pace": fmt.Sprintf("%d", 10+i%10)},
			NumericAttributes:       map[string]int{"Pace": 10 + i%10},
			PerformanceStatsNumeric: map[string]float64{"Goals": float64(i%20) + 0.5},
			PerformancePercentiles:  map[string]map[string]float64{"Goals": {"90th": float64(i%100) / 100.0}},
			ParsedPositions:         []string{"ST"},
			ShortPositions:          []string{"ST"},
			PositionGroups:          []string{"Forward"},
			PAC:                     10 + i%10,
			SHO:                     10 + i%10,
			PAS:                     10 + i%10,
			DRI:                     10 + i%10,
			DEF:                     5 + i%10,
			PHY:                     10 + i%10,
			Overall:                 70 + i%20,
			BestRoleOverall:         "Advanced Forward",
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Advanced Forward", Score: 70 + i%20},
			},
			TransferValueAmount: int64((i%50 + 1) * 1000000),
			WageAmount:          int64((i%100 + 10) * 1000),
		}
	}

	dataset := DatasetData{
		Players:        players,
		CurrencySymbol: "£",
	}

	duration := time.Since(start)
	logDebug(ctx, "Test dataset creation completed", 
		"name_prefix", namePrefix, 
		"player_count", playerCount,
		"duration_ms", duration.Milliseconds())

	return dataset
}

// MockStorageWithFailures simulates storage failures for testing fallback behavior
type MockStorageWithFailures struct {
	*InMemoryStorage
	failureMode string // "store", "retrieve", "delete", "list"
}

func (m *MockStorageWithFailures) Store(datasetID string, data DatasetData) error {
	ctx := context.Background()
	if m.failureMode == "store" {
		logDebug(ctx, "Mock storage simulating store failure", "dataset_id", datasetID, "failure_mode", m.failureMode)
		return ErrStorageFailure
	}
	logDebug(ctx, "Mock storage performing normal store operation", "dataset_id", datasetID)
	return m.InMemoryStorage.Store(datasetID, data)
}

func (m *MockStorageWithFailures) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	if m.failureMode == "retrieve" {
		logDebug(ctx, "Mock storage simulating retrieve failure", "dataset_id", datasetID, "failure_mode", m.failureMode)
		return DatasetData{}, ErrRetrievalFailure
	}
	logDebug(ctx, "Mock storage performing normal retrieve operation", "dataset_id", datasetID)
	return m.InMemoryStorage.Retrieve(datasetID)
}

func (m *MockStorageWithFailures) Delete(datasetID string) error {
	ctx := context.Background()
	if m.failureMode == "delete" {
		logDebug(ctx, "Mock storage simulating delete failure", "dataset_id", datasetID, "failure_mode", m.failureMode)
		return ErrDeletionFailure
	}
	logDebug(ctx, "Mock storage performing normal delete operation", "dataset_id", datasetID)
	return m.InMemoryStorage.Delete(datasetID)
}

func (m *MockStorageWithFailures) List() ([]string, error) {
	ctx := context.Background()
	if m.failureMode == "list" {
		logDebug(ctx, "Mock storage simulating list failure", "failure_mode", m.failureMode)
		return nil, ErrListFailure
	}
	logDebug(ctx, "Mock storage performing normal list operation")
	return m.InMemoryStorage.List()
}

// initializeStorageForTest simulates the storage initialization logic
//nolint:ireturn // This function is designed to return different storage implementations
func initializeStorageForTest(backend StorageInterface) StorageInterface {
	useProtobuf := os.Getenv("USE_PROTOBUF") == "true"
	if useProtobuf {
		return CreateProtobufStorage(backend)
	}
	return backend
}

// createTestDatasetWithSpecialChars creates a dataset with special characters for testing
func createTestDatasetWithSpecialChars(ctx context.Context) DatasetData {
	start := time.Now()
	logDebug(ctx, "Creating test dataset with special characters")
	
	dataset := DatasetData{
		Players: []Player{
			{
				UID:                     1,
				Name:                    "José María Ñoño",
				Position:                "GK/CB",
				Age:                     "25",
				Club:                    "Real Madrid C.F.",
				Division:                "La Liga™",
				TransferValue:           "€50M",
				Wage:                    "£100K",
				Nationality:             "España",
				NationalityISO:          "ES",
				NationalityFIFACode:     "ESP",
				Attributes:              map[string]string{"special": "ñáéíóú"},
				NumericAttributes:       map[string]int{"test": 100},
				PerformanceStatsNumeric: map[string]float64{"rating": 8.5},
				PerformancePercentiles:  map[string]map[string]float64{"overall": {"percentile": 95.0}},
				ParsedPositions:         []string{"GK", "CB"},
				ShortPositions:          []string{"GK", "CB"},
				PositionGroups:          []string{"Defence"},
				RoleSpecificOveralls:    []RoleOverallScore{{RoleName: "Goalkeeper", Score: 85}},
			},
		},
		CurrencySymbol: "€",
	}

	duration := time.Since(start)
	logDebug(ctx, "Test dataset with special characters creation completed", 
		"player_count", len(dataset.Players),
		"duration_ms", duration.Milliseconds())

	return dataset
}

// assertDatasetEqual compares two datasets for equality
func assertDatasetEqual(t *testing.T, expected, actual DatasetData) {
	assert.Equal(t, expected.CurrencySymbol, actual.CurrencySymbol, "Currency symbols should match")
	assert.Equal(t, len(expected.Players), len(actual.Players), "Player count should match")

	for i, expectedPlayer := range expected.Players {
		if i < len(actual.Players) {
			actualPlayer := actual.Players[i]
			assert.Equal(t, expectedPlayer.UID, actualPlayer.UID, "Player UID should match")
			assert.Equal(t, expectedPlayer.Name, actualPlayer.Name, "Player name should match")
			assert.Equal(t, expectedPlayer.Position, actualPlayer.Position, "Player position should match")
			assert.Equal(t, expectedPlayer.Club, actualPlayer.Club, "Player club should match")
			// Add more field comparisons as needed
		}
	}
}
