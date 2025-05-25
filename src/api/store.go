package main

import "sync"

// Global storage instance
var storage StorageInterface

// Deprecated: Use storage interface instead
// playerDataStore holds the parsed player data in memory, keyed by a dataset ID.
// Each dataset also stores the detected currency symbol.
var playerDataStore = make(map[string]struct {
	Players        []Player
	CurrencySymbol string
})

// storeMutex protects concurrent access to playerDataStore.
var storeMutex sync.RWMutex

// InitStore initializes the global storage instance
func InitStore() {
	storage = InitializeStorage()
}

// StoreDataset stores player data using the storage interface
func StoreDataset(datasetID string, players []Player, currencySymbol string) error {
	data := DatasetData{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}
	return storage.Store(datasetID, data)
}

// RetrieveDataset retrieves player data using the storage interface
func RetrieveDataset(datasetID string) ([]Player, string, error) {
	data, err := storage.Retrieve(datasetID)
	if err != nil {
		return nil, "", err
	}
	return data.Players, data.CurrencySymbol, nil
}

// DeleteDataset deletes player data using the storage interface
func DeleteDataset(datasetID string) error {
	return storage.Delete(datasetID)
}

// ListDatasets lists all available dataset IDs
func ListDatasets() ([]string, error) {
	return storage.List()
}

// Legacy functions for backward compatibility

// GetPlayerData retrieves player data from the legacy store (for backward compatibility)
func GetPlayerData(datasetID string) ([]Player, string, bool) {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	
	if data, exists := playerDataStore[datasetID]; exists {
		return data.Players, data.CurrencySymbol, true
	}
	
	// Try to get from new storage system
	players, currency, err := RetrieveDataset(datasetID)
	if err == nil {
		return players, currency, true
	}
	
	return nil, "", false
}

// SetPlayerData stores player data in the legacy store and new storage
func SetPlayerData(datasetID string, players []Player, currencySymbol string) {
	// Store in legacy format
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{
		Players:        players,
		CurrencySymbol: currencySymbol,
	}
	storeMutex.Unlock()
	
	// Store in new storage system
	if err := StoreDataset(datasetID, players, currencySymbol); err != nil {
		// Log error but don't fail - legacy store still works
		// (error logging is handled in storage implementation)
	}
}
