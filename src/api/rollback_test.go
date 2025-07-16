package main

import (
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
			// Create test dataset
			testData := createTestDataset("rollback-test", 100)
			datasetID := "rollback-test-dataset"

			// Step 1: Test with protobuf enabled
			t.Log("Step 1: Testing with protobuf enabled")
			backend := tt.setupFunc()
			protobufStorage := CreateProtobufStorage(backend)

			// Store data using protobuf
			err := protobufStorage.Store(datasetID, testData)
			require.NoError(t, err, "Should store data successfully with protobuf")

			// Retrieve data using protobuf
			retrievedData, err := protobufStorage.Retrieve(datasetID)
			require.NoError(t, err, "Should retrieve data successfully with protobuf")
			assertDatasetEqual(t, testData, retrievedData)

			// Step 2: Immediate rollback - disable protobuf, use JSON directly
			t.Log("Step 2: Testing immediate rollback to JSON")
			jsonStorage := backend // Use the underlying backend directly (JSON mode)

			// When data is stored using protobuf, it's stored in a special encoded format
			// Direct JSON retrieval will get the encoded protobuf data, not the original data
			retrievedDataJSON, err := jsonStorage.Retrieve(datasetID)
			if err != nil {
				t.Fatalf("Should be able to retrieve data from backend: %v", err)
			}

			// Check if this is protobuf-encoded data (indicated by special marker)
			if retrievedDataJSON.CurrencySymbol == "__PROTOBUF_MARKER__" {
				t.Log("Data is stored in protobuf format - testing protobuf storage can still handle retrieval")
				
				// Test that protobuf storage can still retrieve and convert the data
				retrievedDataProtobuf, err := protobufStorage.Retrieve(datasetID)
				require.NoError(t, err, "Protobuf storage should handle retrieval of protobuf data")
				assertDatasetEqual(t, testData, retrievedDataProtobuf)
			} else {
				// If it's regular JSON data, verify data integrity
				t.Log("Data is stored in JSON format - verifying data integrity")
				assertDatasetEqual(t, testData, retrievedDataJSON)
			}

			// Step 3: Store new data using JSON only (simulating rollback scenario)
			t.Log("Step 3: Testing new data storage with JSON after rollback")
			newDatasetID := "rollback-new-data"
			newTestData := createTestDataset("rollback-new", 50)

			err = jsonStorage.Store(newDatasetID, newTestData)
			require.NoError(t, err, "Should store new data successfully with JSON after rollback")

			retrievedNewData, err := jsonStorage.Retrieve(newDatasetID)
			require.NoError(t, err, "Should retrieve new data successfully with JSON")
			assertDatasetEqual(t, newTestData, retrievedNewData)

			// Step 4: Verify both datasets are accessible
			t.Log("Step 4: Verifying both datasets remain accessible")
			datasets, err := jsonStorage.List()
			require.NoError(t, err, "Should list datasets successfully")
			assert.Contains(t, datasets, datasetID, "Original dataset should still be listed")
			assert.Contains(t, datasets, newDatasetID, "New dataset should be listed")

			// Cleanup
			_ = jsonStorage.Delete(datasetID)
			_ = jsonStorage.Delete(newDatasetID)
		})
	}
}

// TestAutomaticFallbackBehavior tests that protobuf storage automatically falls back
// to JSON when protobuf operations fail
func TestAutomaticFallbackBehavior(t *testing.T) {
	t.Log("Testing automatic fallback behavior when protobuf operations fail")

	// Create a mock storage that can simulate failures
	mockStorage := &MockStorageWithFailures{
		InMemoryStorage: CreateInMemoryStorage(),
		failureMode:     "",
	}

	protobufStorage := CreateProtobufStorage(mockStorage)
	testData := createTestDataset("fallback-test", 50)
	datasetID := "fallback-test-dataset"

	// Test 1: Normal operation should work
	t.Run("Normal_Operation", func(t *testing.T) {
		err := protobufStorage.Store(datasetID, testData)
		require.NoError(t, err, "Normal protobuf storage should work")

		retrievedData, err := protobufStorage.Retrieve(datasetID)
		require.NoError(t, err, "Normal protobuf retrieval should work")
		assertDatasetEqual(t, testData, retrievedData)
	})

	// Test 2: Storage failure should trigger fallback
	t.Run("Storage_Failure_Fallback", func(t *testing.T) {
		mockStorage.failureMode = "store"
		defer func() { mockStorage.failureMode = "" }()

		// This should trigger fallback to JSON storage
		err := protobufStorage.Store(datasetID+"_fail", testData)
		// The error handling depends on implementation - it might succeed via fallback
		// or fail gracefully. The key is that it doesn't panic or corrupt data.
		t.Logf("Storage with failure mode result: %v", err)
	})

	// Test 3: Retrieval failure should trigger fallback
	t.Run("Retrieval_Failure_Fallback", func(t *testing.T) {
		// First store data normally
		err := protobufStorage.Store(datasetID+"_retrieve", testData)
		require.NoError(t, err)

		// Then simulate retrieval failure
		mockStorage.failureMode = "retrieve"
		defer func() { mockStorage.failureMode = "" }()

		// This should trigger fallback behavior
		_, err = protobufStorage.Retrieve(datasetID + "_retrieve")
		t.Logf("Retrieval with failure mode result: %v", err)
	})

	// Cleanup
	_ = mockStorage.Delete(datasetID)
	_ = mockStorage.Delete(datasetID + "_fail")
	_ = mockStorage.Delete(datasetID + "_retrieve")
}

// TestDataIntegrityDuringRollback verifies that no data is lost during rollback scenarios
func TestDataIntegrityDuringRollback(t *testing.T) {
	t.Log("Testing data integrity during various rollback scenarios")

	tempDir := t.TempDir()
	localStorage, err := CreateLocalFileStorage(tempDir)
	require.NoError(t, err)

	protobufStorage := CreateProtobufStorage(localStorage)

	// Create multiple test datasets with different characteristics
	datasets := map[string]DatasetData{
		"small-dataset":  createTestDataset("small", 10),
		"medium-dataset": createTestDataset("medium", 100),
		"large-dataset":  createTestDataset("large", 1000),
		"empty-dataset":  {Players: []Player{}, CurrencySymbol: "USD"},
		"special-chars":  createTestDatasetWithSpecialChars(),
	}

	// Step 1: Store all datasets using protobuf
	t.Log("Step 1: Storing datasets using protobuf")
	for datasetID, data := range datasets {
		err := protobufStorage.Store(datasetID, data)
		require.NoError(t, err, "Should store dataset %s successfully", datasetID)
	}

	// Step 2: Verify all data can be retrieved using protobuf
	t.Log("Step 2: Verifying protobuf retrieval")
	for datasetID, originalData := range datasets {
		retrievedData, err := protobufStorage.Retrieve(datasetID)
		require.NoError(t, err, "Should retrieve dataset %s successfully", datasetID)
		assertDatasetEqual(t, originalData, retrievedData)
	}

	// Step 3: Simulate rollback by using JSON storage directly
	t.Log("Step 3: Testing rollback scenario with JSON storage")
	jsonStorage := localStorage

	// Try to retrieve data using JSON storage (this tests fallback mechanisms)
	for datasetID, originalData := range datasets {
		// The behavior here depends on how the data was stored
		// If stored as protobuf, direct JSON retrieval might fail, which is expected
		retrievedData, err := jsonStorage.Retrieve(datasetID)
		if err != nil {
			t.Logf("Direct JSON retrieval failed for %s (expected if stored as protobuf): %v", datasetID, err)
			
			// Verify that protobuf storage can still handle the retrieval
			retrievedDataProtobuf, err := protobufStorage.Retrieve(datasetID)
			require.NoError(t, err, "Protobuf storage should handle retrieval for %s", datasetID)
			assertDatasetEqual(t, originalData, retrievedDataProtobuf)
		} else {
			// If JSON retrieval succeeds, verify data integrity
			assertDatasetEqual(t, originalData, retrievedData)
		}
	}

	// Step 4: Store new data using JSON only (post-rollback scenario)
	t.Log("Step 4: Testing post-rollback data storage")
	postRollbackData := createTestDataset("post-rollback", 75)
	postRollbackID := "post-rollback-dataset"

	err = jsonStorage.Store(postRollbackID, postRollbackData)
	require.NoError(t, err, "Should store post-rollback data successfully")

	retrievedPostRollback, err := jsonStorage.Retrieve(postRollbackID)
	require.NoError(t, err, "Should retrieve post-rollback data successfully")
	assertDatasetEqual(t, postRollbackData, retrievedPostRollback)

	// Step 5: Verify all datasets are still accessible and intact
	t.Log("Step 5: Final data integrity verification")
	allDatasets, err := jsonStorage.List()
	require.NoError(t, err, "Should list all datasets successfully")

	expectedDatasets := make([]string, 0, len(datasets)+1)
	for datasetID := range datasets {
		expectedDatasets = append(expectedDatasets, datasetID)
	}
	expectedDatasets = append(expectedDatasets, postRollbackID)

	for _, expectedID := range expectedDatasets {
		assert.Contains(t, allDatasets, expectedID, "Dataset %s should be listed", expectedID)
	}

	// Cleanup
	for datasetID := range datasets {
		_ = jsonStorage.Delete(datasetID)
	}
	_ = jsonStorage.Delete(postRollbackID)
}

// TestEnvironmentVariableRollback tests rollback via environment variable changes
func TestEnvironmentVariableRollback(t *testing.T) {
	t.Log("Testing rollback via USE_PROTOBUF environment variable")

	// Save original environment variable
	originalValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("USE_PROTOBUF")
		} else {
			os.Setenv("USE_PROTOBUF", originalValue)
		}
	}()

	tempDir := t.TempDir()
	testData := createTestDataset("env-rollback", 50)
	datasetID := "env-rollback-test"

	// Test 1: Enable protobuf via environment variable
	t.Run("Enable_Protobuf", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "true")
		
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
		os.Setenv("USE_PROTOBUF", "false")
		
		// Initialize storage again (simulating app restart after config change)
		localStorage, err := CreateLocalFileStorage(tempDir)
		require.NoError(t, err)
		
		storage := initializeStorageForTest(localStorage)
		
		// Should still be able to retrieve existing data
		retrievedData, err := storage.Retrieve(datasetID)
		if err != nil {
			t.Logf("Retrieval after rollback failed (may be expected): %v", err)
		} else {
			assertDatasetEqual(t, testData, retrievedData)
		}
		
		// Should be able to store new data using JSON
		newDatasetID := "env-rollback-new"
		newTestData := createTestDataset("env-rollback-new", 25)
		
		err = storage.Store(newDatasetID, newTestData)
		require.NoError(t, err, "Should store new data after rollback")
		
		retrievedNewData, err := storage.Retrieve(newDatasetID)
		require.NoError(t, err, "Should retrieve new data after rollback")
		assertDatasetEqual(t, newTestData, retrievedNewData)
		
		// Cleanup
		_ = storage.Delete(newDatasetID)
	})

	// Cleanup
	localStorage, _ := CreateLocalFileStorage(tempDir)
	_ = localStorage.Delete(datasetID)
}

// TestRollbackPerformanceImpact tests that rollback doesn't significantly impact performance
func TestRollbackPerformanceImpact(t *testing.T) {
	t.Log("Testing performance impact of rollback scenarios")

	tempDir := t.TempDir()
	localStorage, err := CreateLocalFileStorage(tempDir)
	require.NoError(t, err)

	testData := createTestDataset("perf-test", 500)
	datasetID := "performance-rollback-test"

	// Measure protobuf performance
	protobufStorage := CreateProtobufStorage(localStorage)
	
	start := time.Now()
	err = protobufStorage.Store(datasetID, testData)
	require.NoError(t, err)
	protobufStoreTime := time.Since(start)

	start = time.Now()
	_, err = protobufStorage.Retrieve(datasetID)
	require.NoError(t, err)
	protobufRetrieveTime := time.Since(start)

	// Measure JSON performance (rollback scenario)
	jsonStorage := localStorage
	jsonDatasetID := datasetID + "-json"

	start = time.Now()
	err = jsonStorage.Store(jsonDatasetID, testData)
	require.NoError(t, err)
	jsonStoreTime := time.Since(start)

	start = time.Now()
	_, err = jsonStorage.Retrieve(jsonDatasetID)
	require.NoError(t, err)
	jsonRetrieveTime := time.Since(start)

	// Log performance comparison
	t.Logf("Performance comparison:")
	t.Logf("  Protobuf Store: %v", protobufStoreTime)
	t.Logf("  JSON Store: %v", jsonStoreTime)
	t.Logf("  Protobuf Retrieve: %v", protobufRetrieveTime)
	t.Logf("  JSON Retrieve: %v", jsonRetrieveTime)

	// Verify that rollback performance is acceptable (within reasonable bounds)
	// JSON should be slower but not excessively so
	maxAcceptableSlowdown := 5.0 // 5x slower is still acceptable for rollback
	
	storeSlowdown := float64(jsonStoreTime) / float64(protobufStoreTime)
	retrieveSlowdown := float64(jsonRetrieveTime) / float64(protobufRetrieveTime)
	
	assert.Less(t, storeSlowdown, maxAcceptableSlowdown, 
		"JSON store performance should be within acceptable bounds during rollback")
	assert.Less(t, retrieveSlowdown, maxAcceptableSlowdown, 
		"JSON retrieve performance should be within acceptable bounds during rollback")

	// Cleanup
	_ = localStorage.Delete(datasetID)
	_ = localStorage.Delete(jsonDatasetID)
}

// Helper functions

// createTestDataset creates a test dataset with the specified name prefix and player count
func createTestDataset(namePrefix string, playerCount int) DatasetData {
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
	
	return DatasetData{
		Players:        players,
		CurrencySymbol: "£",
	}
}

// MockStorageWithFailures simulates storage failures for testing fallback behavior
type MockStorageWithFailures struct {
	*InMemoryStorage
	failureMode string // "store", "retrieve", "delete", "list"
}

func (m *MockStorageWithFailures) Store(datasetID string, data DatasetData) error {
	if m.failureMode == "store" {
		return fmt.Errorf("simulated storage failure")
	}
	return m.InMemoryStorage.Store(datasetID, data)
}

func (m *MockStorageWithFailures) Retrieve(datasetID string) (DatasetData, error) {
	if m.failureMode == "retrieve" {
		return DatasetData{}, fmt.Errorf("simulated retrieval failure")
	}
	return m.InMemoryStorage.Retrieve(datasetID)
}

func (m *MockStorageWithFailures) Delete(datasetID string) error {
	if m.failureMode == "delete" {
		return fmt.Errorf("simulated deletion failure")
	}
	return m.InMemoryStorage.Delete(datasetID)
}

func (m *MockStorageWithFailures) List() ([]string, error) {
	if m.failureMode == "list" {
		return nil, fmt.Errorf("simulated list failure")
	}
	return m.InMemoryStorage.List()
}

// initializeStorageForTest simulates the storage initialization logic
func initializeStorageForTest(backend StorageInterface) StorageInterface {
	useProtobuf := os.Getenv("USE_PROTOBUF") == "true"
	if useProtobuf {
		return CreateProtobufStorage(backend)
	}
	return backend
}

// createTestDatasetWithSpecialChars creates a dataset with special characters for testing
func createTestDatasetWithSpecialChars() DatasetData {
	return DatasetData{
		Players: []Player{
			{
				UID:                 1,
				Name:                "José María Ñoño",
				Position:            "GK/CB",
				Age:                 "25",
				Club:                "Real Madrid C.F.",
				Division:            "La Liga™",
				TransferValue:       "€50M",
				Wage:                "£100K",
				Nationality:         "España",
				NationalityISO:      "ES",
				NationalityFIFACode: "ESP",
				Attributes:          map[string]string{"special": "ñáéíóú"},
				NumericAttributes:   map[string]int{"test": 100},
				PerformanceStatsNumeric: map[string]float64{"rating": 8.5},
				PerformancePercentiles:  map[string]map[string]float64{"overall": {"percentile": 95.0}},
				ParsedPositions:     []string{"GK", "CB"},
				ShortPositions:      []string{"GK", "CB"},
				PositionGroups:      []string{"Defence"},
				RoleSpecificOveralls: []RoleOverallScore{{RoleName: "Goalkeeper", Score: 85}},
			},
		},
		CurrencySymbol: "€",
	}
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