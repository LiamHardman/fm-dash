package main

import (
	"context"
	"sync"
	"time"

	apperrors "api/errors"

	"go.opentelemetry.io/otel/attribute"
)

// InMemoryStorage provides in-memory storage for datasets
type InMemoryStorage struct {
	data  map[string]DatasetData
	mutex sync.RWMutex
}

// CreateInMemoryStorage creates a new in-memory storage instance
func CreateInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]DatasetData),
	}
}

// Store saves a dataset to in-memory storage
func (s *InMemoryStorage) Store(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.memory.store")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "memory"),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[datasetID] = data

	RecordDBOperation(ctx, "store", "datasets", 0, 1)
	return nil
}

// Retrieve retrieves a dataset from in-memory storage
func (s *InMemoryStorage) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.memory.retrieve")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("storage.type", "memory"),
	)

	start := time.Now()
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, exists := s.data[datasetID]
	if !exists {
		err := apperrors.WrapErrDatasetNotFound(datasetID)
		RecordError(ctx, err, "Dataset not found in memory storage")
		return DatasetData{}, err
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(data.Players)))
	RecordDBOperation(ctx, "retrieve", "datasets", time.Since(start), 1)
	return data, nil
}

// Delete removes a dataset from in-memory storage
func (s *InMemoryStorage) Delete(datasetID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, datasetID)
	return nil
}

// List returns all dataset IDs stored in memory
func (s *InMemoryStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}

// CleanupOldDatasets removes datasets older than the specified duration
func (s *InMemoryStorage) CleanupOldDatasets(_ time.Duration, _ []string) error {
	// In-memory storage doesn't persist data, so no cleanup needed
	LogInfo("CleanupOldDatasets called on in-memory storage - no action needed")
	return nil
}
