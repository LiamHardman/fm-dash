package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalConfigStorageStoreCreatesNestedSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	storage := &localConfigStorage{dir: tempDir}

	// Keys like "wishlists/<datasetID>.json" (see wishlistStorageKey in
	// wishlist_handler.go) place the file in a subdirectory of tempDir that never
	// gets created otherwise — StoreConfig must create it on demand rather than
	// failing with "The system cannot find the path specified."
	want := []byte(`{"players":[]}`)
	if err := storage.StoreConfig("wishlists/some-dataset-id.json", want); err != nil {
		t.Fatalf("StoreConfig() with nested key error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "wishlists", "some-dataset-id.json")); err != nil {
		t.Fatalf("expected stored file to exist: %v", err)
	}

	got, err := storage.RetrieveConfig("wishlists/some-dataset-id.json")
	if err != nil {
		t.Fatalf("RetrieveConfig() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("RetrieveConfig() = %q, want %q", got, want)
	}

	if err := storage.DeleteConfig("wishlists/some-dataset-id.json"); err != nil {
		t.Fatalf("DeleteConfig() error = %v", err)
	}
	if _, err := storage.RetrieveConfig("wishlists/some-dataset-id.json"); !os.IsNotExist(err) {
		t.Fatalf("RetrieveConfig() after delete error = %v, want os.IsNotExist", err)
	}
}
