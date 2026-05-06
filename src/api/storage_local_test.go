package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestLocalFileStorageListIncludesProtobufAndDeduplicates(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := CreateLocalFileStorage(tempDir)
	if err != nil {
		t.Fatalf("CreateLocalFileStorage() error = %v", err)
	}

	files := map[string][]byte{
		"alpha" + datasetExtProtobuf: []byte("protobuf"),
		"alpha" + datasetExtJSON:     []byte(`{"players":[],"currencySymbol":"£"}`),
		"bravo" + datasetExtJSONGzip: []byte("gzip"),
		"notes.txt":                  []byte("ignored"),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(tempDir, name), data, 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	got, err := storage.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	sort.Strings(got)

	want := []string{"alpha", "bravo"}
	if len(got) != len(want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("List() = %v, want %v", got, want)
		}
	}
}

func TestLocalFileStorageCleanupDeletesOldProtobufDatasets(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := CreateLocalFileStorage(tempDir)
	if err != nil {
		t.Fatalf("CreateLocalFileStorage() error = %v", err)
	}

	oldPB := filepath.Join(tempDir, "old"+datasetExtProtobuf)
	excludedPB := filepath.Join(tempDir, "excluded"+datasetExtProtobuf)
	newJSON := filepath.Join(tempDir, "new"+datasetExtJSON)
	for _, path := range []string{oldPB, excludedPB, newJSON} {
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", path, err)
		}
	}

	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldPB, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(oldPB) error = %v", err)
	}
	if err := os.Chtimes(excludedPB, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(excludedPB) error = %v", err)
	}

	if err := storage.CleanupOldDatasets(24*time.Hour, []string{"excluded"}); err != nil {
		t.Fatalf("CleanupOldDatasets() error = %v", err)
	}

	if _, err := os.Stat(oldPB); !os.IsNotExist(err) {
		t.Fatalf("old protobuf file still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(excludedPB); err != nil {
		t.Fatalf("excluded protobuf file should remain: %v", err)
	}
	if _, err := os.Stat(newJSON); err != nil {
		t.Fatalf("new JSON file should remain: %v", err)
	}
}

func TestLocalFileStorageRetrievePrefersProtobuf(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := CreateLocalFileStorage(tempDir)
	if err != nil {
		t.Fatalf("CreateLocalFileStorage() error = %v", err)
	}

	raw := []byte("protobuf bytes")
	if err := os.WriteFile(filepath.Join(tempDir, "mixed"+datasetExtProtobuf), raw, 0o600); err != nil {
		t.Fatalf("WriteFile(pb) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "mixed"+datasetExtJSON), []byte(`{"players":[],"currencySymbol":"£"}`), 0o600); err != nil {
		t.Fatalf("WriteFile(json) error = %v", err)
	}

	got, err := storage.Retrieve("mixed")
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}
	if string(got.RawBytes) != string(raw) {
		t.Fatalf("Retrieve() RawBytes = %q, want %q", got.RawBytes, raw)
	}
}
