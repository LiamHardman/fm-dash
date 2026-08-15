package main

import (
	"testing"
	"time"

	apperrors "api/errors"
)

// Mock storage implementation for testing
type MockStorage struct {
	data map[string]DatasetData
}

// CreateMockStorage creates a new mock storage for testing
func CreateMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]DatasetData),
	}
}

func (m *MockStorage) Store(datasetID string, data DatasetData) error {
	m.data[datasetID] = data
	return nil
}

func (m *MockStorage) Retrieve(datasetID string) (DatasetData, error) {
	data, exists := m.data[datasetID]
	if !exists {
		return DatasetData{}, apperrors.ErrDatasetNotFound
	}
	return data, nil
}

func (m *MockStorage) Delete(datasetID string) error {
	delete(m.data, datasetID)
	return nil
}

func (m *MockStorage) List() ([]string, error) {
	var ids []string
	for id := range m.data {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MockStorage) CleanupOldDatasets(_ time.Duration, _ []string) error {
	// Mock implementation - just return nil for testing
	return nil
}

func TestStoreDataset(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	tests := []struct {
		name           string
		datasetID      string
		players        []Player
		currencySymbol string
		expectError    bool
	}{
		{
			name:      "store valid dataset",
			datasetID: "test-dataset-1",
			players: []Player{
				{
					UID:  1,
					Name: "Test Player 1",
					Age:  "25",
				},
				{
					UID:  2,
					Name: "Test Player 2",
					Age:  "28",
				},
			},
			currencySymbol: "£",
			expectError:    false,
		},
		{
			name:           "store empty dataset",
			datasetID:      "empty-dataset",
			players:        []Player{},
			currencySymbol: "$",
			expectError:    false,
		},
		{
			name:      "store dataset with no currency",
			datasetID: "no-currency",
			players: []Player{
				{
					UID:  3,
					Name: "Test Player 3",
					Age:  "30",
				},
			},
			currencySymbol: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StoreDataset(tt.datasetID, tt.players, tt.currencySymbol)

			if tt.expectError && err == nil {
				t.Errorf("StoreDataset() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("StoreDataset() unexpected error: %v", err)
			}

			// Verify data was stored correctly
			if !tt.expectError {
				retrievedPlayers, retrievedCurrency, err := RetrieveDataset(tt.datasetID)
				if err != nil {
					t.Errorf("Failed to retrieve stored dataset: %v", err)
				}

				if len(retrievedPlayers) != len(tt.players) {
					t.Errorf("Retrieved %d players, expected %d", len(retrievedPlayers), len(tt.players))
				}

				if retrievedCurrency != tt.currencySymbol {
					t.Errorf("Retrieved currency %s, expected %s", retrievedCurrency, tt.currencySymbol)
				}
			}
		})
	}
}

func TestRetrieveDataset(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	mockStorage := CreateMockStorage()
	storage = mockStorage

	// Pre-populate with test data
	testPlayers := []Player{
		{UID: 1, Name: "Test Player", Age: "25"},
	}
	testCurrency := "€"
	testDatasetID := "test-retrieve"

	err := StoreDataset(testDatasetID, testPlayers, testCurrency)
	if err != nil {
		t.Fatalf("Failed to store test data: %v", err)
	}

	tests := []struct {
		name             string
		datasetID        string
		expectError      bool
		expectedPlayers  int
		expectedCurrency string
	}{
		{
			name:             "retrieve existing dataset",
			datasetID:        testDatasetID,
			expectError:      false,
			expectedPlayers:  1,
			expectedCurrency: testCurrency,
		},
		{
			name:        "retrieve non-existent dataset",
			datasetID:   "non-existent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players, currency, err := RetrieveDataset(tt.datasetID)

			if tt.expectError && err == nil {
				t.Errorf("RetrieveDataset() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("RetrieveDataset() unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(players) != tt.expectedPlayers {
					t.Errorf("Retrieved %d players, expected %d", len(players), tt.expectedPlayers)
				}

				if currency != tt.expectedCurrency {
					t.Errorf("Retrieved currency %s, expected %s", currency, tt.expectedCurrency)
				}
			}
		})
	}
}

func TestDeleteDataset(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	// Store test data
	testDatasetID := "test-delete"
	testPlayers := []Player{{UID: 1, Name: "Test Player"}}

	err := StoreDataset(testDatasetID, testPlayers, "$")
	if err != nil {
		t.Fatalf("Failed to store test data: %v", err)
	}

	// Verify data exists
	_, _, err = RetrieveDataset(testDatasetID)
	if err != nil {
		t.Fatalf("Test data should exist before deletion: %v", err)
	}

	// Delete the dataset
	err = DeleteDataset(testDatasetID)
	if err != nil {
		t.Errorf("DeleteDataset() unexpected error: %v", err)
	}

	// Verify data no longer exists
	_, _, err = RetrieveDataset(testDatasetID)
	if err == nil {
		t.Errorf("Dataset should not exist after deletion")
	}
}

func TestListDatasets(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	// Store multiple datasets
	datasets := []string{"dataset1", "dataset2", "dataset3"}
	for _, id := range datasets {
		err := StoreDataset(id, []Player{{UID: 123}}, "$")
		if err != nil {
			t.Fatalf("Failed to store test dataset %s: %v", id, err)
		}
	}

	// List datasets
	listedDatasets, err := ListDatasets()
	if err != nil {
		t.Errorf("ListDatasets() unexpected error: %v", err)
	}

	if len(listedDatasets) != len(datasets) {
		t.Errorf("Listed %d datasets, expected %d", len(listedDatasets), len(datasets))
	}

	// Check that all expected datasets are in the list
	datasetMap := make(map[string]bool)
	for _, id := range listedDatasets {
		datasetMap[id] = true
	}

	for _, expectedID := range datasets {
		if !datasetMap[expectedID] {
			t.Errorf("Expected dataset %s not found in list", expectedID)
		}
	}
}

func TestStoreDatasetAsync(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	testDatasetID := "test-async"
	testPlayers := []Player{{UID: 1, Name: "Async Test Player"}}
	testCurrency := "£"

	// Store dataset asynchronously
	StoreDatasetAsync(testDatasetID, testPlayers, testCurrency)

	// Wait a bit for the async operation to complete
	time.Sleep(100 * time.Millisecond)

	// Verify data was stored
	retrievedPlayers, retrievedCurrency, err := RetrieveDataset(testDatasetID)
	if err != nil {
		t.Errorf("Failed to retrieve async stored dataset: %v", err)
	}

	if len(retrievedPlayers) != len(testPlayers) {
		t.Errorf("Retrieved %d players, expected %d", len(retrievedPlayers), len(testPlayers))
	}

	if retrievedCurrency != testCurrency {
		t.Errorf("Retrieved currency %s, expected %s", retrievedCurrency, testCurrency)
	}
}

func TestCleanupOldDatasets(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	// Test cleanup function (mock implementation just returns nil)
	maxAge := 24 * time.Hour
	excludeDatasets := []string{"demo", "important"}

	err := CleanupOldDatasets(maxAge, excludeDatasets)
	if err != nil {
		t.Errorf("CleanupOldDatasets() unexpected error: %v", err)
	}
}

// Test data validation
func TestStoreDatasetValidation(t *testing.T) {
	// Setup mock storage
	originalStorage := storage
	defer func() { storage = originalStorage }()

	storage = CreateMockStorage()

	tests := []struct {
		name      string
		datasetID string
		players   []Player
		currency  string
	}{
		{
			name:      "valid dataset with special characters in currency",
			datasetID: "special-currency",
			players:   []Player{{UID: 1}},
			currency:  "₹", // Indian Rupee
		},
		{
			name:      "dataset with large number of players",
			datasetID: "large-dataset",
			players:   make([]Player, 1000), // Large dataset
			currency:  "$",
		},
		{
			name:      "dataset with unicode player names",
			datasetID: "unicode-names",
			players: []Player{
				{UID: 1, Name: "José María"},
				{UID: 2, Name: "Müller"},
				{UID: 3, Name: "Žižek"},
			},
			currency: "€",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StoreDataset(tt.datasetID, tt.players, tt.currency)
			if err != nil {
				t.Errorf("StoreDataset() unexpected error: %v", err)
			}

			// Verify retrieval works
			retrievedPlayers, retrievedCurrency, err := RetrieveDataset(tt.datasetID)
			if err != nil {
				t.Errorf("RetrieveDataset() unexpected error: %v", err)
			}

			if len(retrievedPlayers) != len(tt.players) {
				t.Errorf("Retrieved %d players, expected %d", len(retrievedPlayers), len(tt.players))
			}

			if retrievedCurrency != tt.currency {
				t.Errorf("Retrieved currency %s, expected %s", retrievedCurrency, tt.currency)
			}
		})
	}
}

func TestNumericAttributesPopulation(t *testing.T) {
	// Create a player with string attributes but no numeric attributes
	player := Player{
		UID:      123,
		Name:     "Test Player",
		Position: "ST",
		Attributes: map[string]string{
			"Pac": "15",
			"Fin": "18",
			"Pas": "12",
			"Dri": "16",
			"Tck": "8",
			"Str": "14",
		},
		NumericAttributes: make(map[string]int), // Empty initially
	}

	// Verify that NumericAttributes are empty initially
	if len(player.NumericAttributes) != 0 {
		t.Errorf("Expected empty NumericAttributes initially, got %d", len(player.NumericAttributes))
	}

	// Test SetPlayerData ensures NumericAttributes are populated
	testDatasetID := "test-numeric-attributes"
	testCurrency := "£"

	SetPlayerData(testDatasetID, []Player{player}, testCurrency)

	// Retrieve the data to verify NumericAttributes were populated
	players, _, err := RetrieveDataset(testDatasetID)
	if err != nil {
		t.Fatalf("Failed to retrieve dataset: %v", err)
	}

	if len(players) != 1 {
		t.Fatalf("Expected 1 player, got %d", len(players))
	}

	retrievedPlayer := players[0]
	if len(retrievedPlayer.NumericAttributes) == 0 {
		t.Error("NumericAttributes should be populated after storage")
	}

	// Verify specific attributes were populated
	expectedAttributes := []string{"Pac", "Fin", "Pas", "Dri", "Tck", "Str"}
	for _, attr := range expectedAttributes {
		if value, exists := retrievedPlayer.NumericAttributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be populated", attr)
		} else if value == 0 {
			t.Errorf("Expected attribute %s to have non-zero value, got %d", attr, value)
		}
	}

	// Clean up
	DeleteDataset(testDatasetID)
}

func TestSetPlayerDataAsyncNumericAttributes(t *testing.T) {
	// Create a player with string attributes but no numeric attributes
	player := Player{
		UID:      456,
		Name:     "Async Test Player",
		Position: "CM",
		Attributes: map[string]string{
			"Pac": "12",
			"Fin": "14",
			"Pas": "18",
			"Dri": "15",
			"Tck": "16",
			"Str": "13",
		},
		NumericAttributes: make(map[string]int), // Empty initially
	}

	// Verify that NumericAttributes are empty initially
	if len(player.NumericAttributes) != 0 {
		t.Errorf("Expected empty NumericAttributes initially, got %d", len(player.NumericAttributes))
	}

	// Test SetPlayerDataAsync ensures NumericAttributes are populated
	testDatasetID := "test-async-numeric-attributes"
	testCurrency := "€"

	SetPlayerDataAsync(testDatasetID, []Player{player}, testCurrency)

	// Wait a bit for async storage to complete
	time.Sleep(100 * time.Millisecond)

	// Retrieve the data to verify NumericAttributes were populated
	players, _, err := RetrieveDataset(testDatasetID)
	if err != nil {
		t.Fatalf("Failed to retrieve dataset: %v", err)
	}

	if len(players) != 1 {
		t.Fatalf("Expected 1 player, got %d", len(players))
	}

	retrievedPlayer := players[0]
	if len(retrievedPlayer.NumericAttributes) == 0 {
		t.Error("NumericAttributes should be populated after async storage")
	}

	// Verify specific attributes were populated
	expectedAttributes := []string{"Pac", "Fin", "Pas", "Dri", "Tck", "Str"}
	for _, attr := range expectedAttributes {
		if value, exists := retrievedPlayer.NumericAttributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be populated", attr)
		} else if value == 0 {
			t.Errorf("Expected attribute %s to have non-zero value, got %d", attr, value)
		}
	}

	// Clean up
	DeleteDataset(testDatasetID)
}
