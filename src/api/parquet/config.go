package parquet

import (
	"os"
	"strconv"
)

// Config holds configuration for Parquet/Arrow operations
type Config struct {
	// Storage configuration
	StorageBasePath   string
	CompressionType   string
	RowGroupSize      int64
	EnableBloomFilter bool
	
	// Arrow configuration
	MemoryPoolSize    int64
	BatchSize         int
	EnableVectorized  bool
	
	// Performance configuration
	MaxConcurrency    int
	CacheSize         int64
	EnableMetrics     bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		StorageBasePath:   "./data/parquet",
		CompressionType:   "SNAPPY",
		RowGroupSize:      128 * 1024 * 1024, // 128MB
		EnableBloomFilter: true,
		MemoryPoolSize:    512 * 1024 * 1024, // 512MB
		BatchSize:         10000,
		EnableVectorized:  true,
		MaxConcurrency:    4,
		CacheSize:         256 * 1024 * 1024, // 256MB
		EnableMetrics:     true,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := DefaultConfig()
	
	if basePath := os.Getenv("PARQUET_STORAGE_BASE_PATH"); basePath != "" {
		config.StorageBasePath = basePath
	}
	
	if compression := os.Getenv("PARQUET_COMPRESSION_TYPE"); compression != "" {
		config.CompressionType = compression
	}
	
	if rowGroupSize := os.Getenv("PARQUET_ROW_GROUP_SIZE"); rowGroupSize != "" {
		if size, err := strconv.ParseInt(rowGroupSize, 10, 64); err == nil {
			config.RowGroupSize = size
		}
	}
	
	if bloomFilter := os.Getenv("PARQUET_ENABLE_BLOOM_FILTER"); bloomFilter != "" {
		if enable, err := strconv.ParseBool(bloomFilter); err == nil {
			config.EnableBloomFilter = enable
		}
	}
	
	if memoryPool := os.Getenv("ARROW_MEMORY_POOL_SIZE"); memoryPool != "" {
		if size, err := strconv.ParseInt(memoryPool, 10, 64); err == nil {
			config.MemoryPoolSize = size
		}
	}
	
	if batchSize := os.Getenv("ARROW_BATCH_SIZE"); batchSize != "" {
		if size, err := strconv.Atoi(batchSize); err == nil {
			config.BatchSize = size
		}
	}
	
	if vectorized := os.Getenv("ARROW_ENABLE_VECTORIZED"); vectorized != "" {
		if enable, err := strconv.ParseBool(vectorized); err == nil {
			config.EnableVectorized = enable
		}
	}
	
	if concurrency := os.Getenv("PARQUET_MAX_CONCURRENCY"); concurrency != "" {
		if max, err := strconv.Atoi(concurrency); err == nil {
			config.MaxConcurrency = max
		}
	}
	
	if cacheSize := os.Getenv("PARQUET_CACHE_SIZE"); cacheSize != "" {
		if size, err := strconv.ParseInt(cacheSize, 10, 64); err == nil {
			config.CacheSize = size
		}
	}
	
	if metrics := os.Getenv("PARQUET_ENABLE_METRICS"); metrics != "" {
		if enable, err := strconv.ParseBool(metrics); err == nil {
			config.EnableMetrics = enable
		}
	}
	
	return config
}