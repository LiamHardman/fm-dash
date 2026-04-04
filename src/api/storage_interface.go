package main

import (
	"time"
)

// StorageInterface defines the interface for data storage operations
type StorageInterface interface {
	Store(datasetID string, data DatasetData) error
	Retrieve(datasetID string) (DatasetData, error)
	Delete(datasetID string) error
	List() ([]string, error)
	CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error
}

// DatasetData represents a dataset containing player information
type DatasetData struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currency_symbol"`
	CacheData      string   `json:"cache_data,omitempty"`
	RawBytes       []byte   `json:"-"`
}
