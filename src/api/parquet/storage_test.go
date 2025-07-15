package parquet

import (
	"testing"
)

func TestNewParquetStorage(t *testing.T) {
	storage := NewParquetStorage("/tmp/test", nil)
	if storage == nil {
		t.Fatal("Expected non-nil ParquetStorage instance")
	}
	
	if storage.basePath != "/tmp/test" {
		t.Errorf("Expected basePath to be '/tmp/test', got '%s'", storage.basePath)
	}
	
	if storage.config == nil {
		t.Fatal("Expected non-nil config")
	}
	
	if storage.config.CompressionType != "SNAPPY" {
		t.Errorf("Expected default compression type to be 'SNAPPY', got '%s'", storage.config.CompressionType)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("Expected non-nil config")
	}
	
	if config.CompressionType != "SNAPPY" {
		t.Errorf("Expected compression type to be 'SNAPPY', got '%s'", config.CompressionType)
	}
	
	if config.RowGroupSize != 128*1024*1024 {
		t.Errorf("Expected row group size to be 128MB, got %d", config.RowGroupSize)
	}
	
	if !config.EnableBloomFilter {
		t.Error("Expected bloom filter to be enabled by default")
	}
}