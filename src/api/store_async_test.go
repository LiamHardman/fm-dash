package main

import (
	"sync"
	"testing"
	"time"
)

type waitableStorage struct {
	mu     sync.Mutex
	data   map[string]DatasetData
	stored chan struct{}
}

func newWaitableStorage() *waitableStorage {
	return &waitableStorage{
		data:   make(map[string]DatasetData),
		stored: make(chan struct{}, 1),
	}
}

func (s *waitableStorage) Store(datasetID string, data DatasetData) error {
	s.mu.Lock()
	s.data[datasetID] = data
	s.mu.Unlock()
	s.stored <- struct{}{}
	return nil
}

func (s *waitableStorage) Retrieve(datasetID string) (DatasetData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[datasetID], nil
}

func (s *waitableStorage) Delete(datasetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, datasetID)
	return nil
}

func (s *waitableStorage) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *waitableStorage) CleanupOldDatasets(time.Duration, []string) error {
	return nil
}

func TestSetPlayerDataAsyncUsesDeepCopySnapshotForPersistentStorage(t *testing.T) {
	originalStorage := storage
	originalCacheDisabled := disableInMemoryDatasetCache
	originalConfigInitialized := configInitialized
	defer func() {
		storage = originalStorage
		disableInMemoryDatasetCache = originalCacheDisabled
		configInitialized = originalConfigInitialized
	}()

	waitStorage := newWaitableStorage()
	storage = waitStorage
	disableInMemoryDatasetCache = true
	configInitialized = true

	players := []Player{
		{
			UID:      123,
			Name:     "Snapshot Player",
			Position: "CM",
			Attributes: map[string]string{
				"Pac": "12",
			},
			NumericAttributes: map[string]int{
				"Pac": 12,
			},
		},
	}

	SetPlayerDataAsync("snapshot-test", players, "£")

	players[0].Name = "Mutated Player"
	players[0].Attributes["Pac"] = "1"
	players[0].NumericAttributes["Pac"] = 1

	select {
	case <-waitStorage.stored:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async storage")
	}

	got, err := waitStorage.Retrieve("snapshot-test")
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}
	if len(got.Players) != 1 {
		t.Fatalf("stored player count = %d, want 1", len(got.Players))
	}
	if got.Players[0].Name != "Snapshot Player" {
		t.Fatalf("stored name = %q, want %q", got.Players[0].Name, "Snapshot Player")
	}
	if got.Players[0].Attributes["Pac"] == "1" {
		t.Fatal("stored Attributes map reflected caller mutation")
	}
	if got.Players[0].NumericAttributes["Pac"] == 1 {
		t.Fatal("stored NumericAttributes map reflected caller mutation")
	}
}
