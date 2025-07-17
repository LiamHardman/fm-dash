package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	pbuf "api/proto"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
)

// Static errors for better error handling
var (
	ErrInvalidOptimizedProtobufFormat = errors.New("invalid optimized protobuf dataset format")
	ErrNotOptimizedProtobufDataset    = errors.New("not an optimized protobuf dataset")
	ErrDecompressionBomb              = errors.New("decompression bomb detected")
)

// Maximum size for decompression to prevent DoS attacks
const maxDecompressedSize = 100 * 1024 * 1024 // 100MB limit

// OptimizedProtobufStorage provides enhanced performance with simple, effective optimizations
type OptimizedProtobufStorage struct {
	backend          StorageInterface
	cache            sync.Map // Simple in-memory cache for frequently accessed data
	compressionLevel int      // 1 = fastest, 9 = best compression
	enableCaching    bool
}

// CreateOptimizedProtobufStorage creates a new optimized protobuf storage wrapper
func CreateOptimizedProtobufStorage(backend StorageInterface, compressionLevel int, enableCaching bool) *OptimizedProtobufStorage {
	if compressionLevel < 0 || compressionLevel > 9 {
		compressionLevel = gzip.BestSpeed // Default to fastest for throughput
	}
	return &OptimizedProtobufStorage{
		backend:          backend,
		compressionLevel: compressionLevel,
		enableCaching:    enableCaching,
	}
}

// StoreWithContext saves a dataset using optimized protobuf serialization
func (s *OptimizedProtobufStorage) StoreWithContext(ctx context.Context, datasetID string, data PlayerDataWithCurrency) error {
	ctx, span := StartSpanWithAttributes(ctx, "storage.store_optimized_protobuf", []attribute.KeyValue{
		attribute.String("storage.type", "optimized_protobuf"),
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.Int("compression.level", s.compressionLevel),
		attribute.Bool("caching.enabled", s.enableCaching),
	})
	defer span.End()

	start := time.Now()

	protoDataset, err := data.ToProtoOptimized(ctx)
	if err != nil {
		marshalErr := NewProtobufError(ctx, "marshal_optimized", datasetID, "Failed to marshal optimized protobuf data", err)
		RecordError(ctx, marshalErr, "Failed to marshal optimized protobuf data")
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	protoBytes, err := proto.Marshal(protoDataset)
	if err != nil {
		marshalErr := NewProtobufError(ctx, "marshal", datasetID, "Failed to marshal protobuf data", err)
		RecordError(ctx, marshalErr, "Failed to marshal protobuf data")
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	SetSpanAttributes(ctx, attribute.Int("protobuf.marshaled_size_bytes", len(protoBytes)))

	compressedData, err := s.compressProtobufDataOptimized(protoBytes)
	if err != nil {
		compErr := NewProtobufCompressionError(ctx, "compress_optimized", datasetID, err)
		RecordError(ctx, compErr, "Failed to compress protobuf data")
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	SetSpanAttributes(ctx,
		attribute.Int("protobuf.size_bytes", len(protoBytes)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
		attribute.Float64("compression.ratio", float64(len(protoBytes))/float64(len(compressedData))),
	)

	err = s.storeProtobufBytesOptimized(ctx, datasetID, compressedData)
	if err != nil {
		storageErr := NewProtobufError(ctx, "store_optimized", datasetID, "Failed to store protobuf data", err)
		RecordError(ctx, storageErr, "Failed to store protobuf data")
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	if s.enableCaching {
		s.cache.Store(datasetID, &CachedDataset{
			Data:      data,
			Timestamp: time.Now(),
			Size:      len(compressedData),
		})
	}

	duration := time.Since(start)
	compressionRatio := float64(len(protoBytes)) / float64(len(compressedData))
	spaceSavedPercent := (1.0 - 1.0/compressionRatio) * 100

	RecordDBOperation(ctx, "store", "optimized_protobuf_datasets", duration, 1)

	logInfo(ctx, "Successfully stored dataset using optimized protobuf serialization",
		"dataset_id", datasetID,
		"player_count", len(data.Players),
		"duration_ms", duration.Milliseconds(),
		"protobuf_size_bytes", len(protoBytes),
		"compressed_size_bytes", len(compressedData),
		"compression_ratio", fmt.Sprintf("%.2f", compressionRatio),
		"space_saved_percent", fmt.Sprintf("%.1f%%", spaceSavedPercent),
		"storage_format", "optimized_protobuf",
		"compression_level", s.compressionLevel,
		"caching_enabled", s.enableCaching)

	return nil
}

// RetrieveWithContext retrieves a dataset using optimized protobuf deserialization
func (s *OptimizedProtobufStorage) RetrieveWithContext(ctx context.Context, datasetID string) (*PlayerDataWithCurrency, error) {
	ctx, span := StartSpanWithAttributes(ctx, "storage.retrieve_optimized_protobuf", []attribute.KeyValue{
		attribute.String("storage.type", "optimized_protobuf"),
		attribute.String("dataset.id", datasetID),
		attribute.Bool("caching.enabled", s.enableCaching),
	})
	defer span.End()

	start := time.Now()

	// Check cache first if enabled
	if s.enableCaching {
		if cached, ok := s.cache.Load(datasetID); ok {
			cachedData := cached.(*CachedDataset)
			// Check if cache is still valid (e.g., less than 1 hour old)
			if time.Since(cachedData.Timestamp) < time.Hour {
				AddSpanEvent(ctx, "cache.hit", attribute.String("dataset_id", datasetID))
				logDebug(ctx, "Retrieved dataset from cache", "dataset_id", datasetID)
				return &cachedData.Data, nil
			}
			// Remove stale cache entry
			s.cache.Delete(datasetID)
		}
	}

	compressedData, err := s.retrieveProtobufBytesOptimized(ctx, datasetID)
	if err != nil {
		retrieveErr := NewProtobufError(ctx, "retrieve_optimized", datasetID, "Failed to retrieve protobuf data", err)
		RecordError(ctx, retrieveErr, "Failed to retrieve protobuf data")
		return nil, retrieveErr
	}

	protoBytes, err := s.decompressProtobufDataOptimized(compressedData)
	if err != nil {
		decompErr := NewProtobufCompressionError(ctx, "decompress_optimized", datasetID, err)
		RecordError(ctx, decompErr, "Failed to decompress protobuf data")
		return nil, decompErr
	}

	var protoDataset pbuf.DatasetData
	err = proto.Unmarshal(protoBytes, &protoDataset)
	if err != nil {
		unmarshalErr := NewProtobufError(ctx, "unmarshal_optimized", datasetID, "Failed to unmarshal protobuf data", err)
		RecordError(ctx, unmarshalErr, "Failed to unmarshal protobuf data")
		return nil, unmarshalErr
	}

	dataset, err := DatasetDataFromProtoOptimized(ctx, &protoDataset)
	if err != nil {
		convertErr := NewProtobufError(ctx, "convert_optimized", datasetID, "Failed to convert protobuf to dataset", err)
		RecordError(ctx, convertErr, "Failed to convert protobuf to dataset")
		return nil, convertErr
	}

	if s.enableCaching {
		s.cache.Store(datasetID, &CachedDataset{
			Data:      *dataset,
			Timestamp: time.Now(),
			Size:      len(compressedData),
		})
	}

	duration := time.Since(start)
	compressionRatio := float64(len(protoBytes)) / float64(len(compressedData))

	RecordDBOperation(ctx, "retrieve", "optimized_protobuf_datasets", duration, 1)

	logInfo(ctx, "Successfully retrieved dataset using optimized protobuf deserialization",
		"dataset_id", datasetID,
		"player_count", len(dataset.Players),
		"duration_ms", duration.Milliseconds(),
		"protobuf_size_bytes", len(protoBytes),
		"compressed_size_bytes", len(compressedData),
		"compression_ratio", fmt.Sprintf("%.2f", compressionRatio),
		"storage_format", "optimized_protobuf",
		"compression_level", s.compressionLevel,
		"caching_enabled", s.enableCaching)

	return dataset, nil
}

// compressProtobufDataOptimized uses fast compression for throughput
func (s *OptimizedProtobufStorage) compressProtobufDataOptimized(data []byte) ([]byte, error) {
	ctx := context.Background()
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.compression.compress_optimized", []attribute.KeyValue{
		attribute.String("compression.algorithm", "gzip"),
		attribute.Int("compression.level", s.compressionLevel),
		attribute.Int("compression.input_size_bytes", len(data)),
	})
	defer span.End()

	start := time.Now()
	var buf bytes.Buffer

	gz, err := gzip.NewWriterLevel(&buf, s.compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := gz.Write(data); err != nil {
		if closeErr := gz.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to write to gzip writer: %w, close error: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	result := buf.Bytes()
	duration := time.Since(start)
	compressionRatio := float64(len(data)) / float64(len(result))

	SetSpanAttributes(ctx,
		attribute.Int("compression.output_size_bytes", len(result)),
		attribute.Float64("compression.ratio", compressionRatio),
		attribute.Float64("compression.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("compression.success", true),
	)

	return result, nil
}

// decompressProtobufDataOptimized uses simple decompression with size limits
func (s *OptimizedProtobufStorage) decompressProtobufDataOptimized(data []byte) ([]byte, error) {
	ctx := context.Background()
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.compression.decompress_optimized", []attribute.KeyValue{
		attribute.String("compression.algorithm", "gzip"),
		attribute.Int("compression.input_size_bytes", len(data)),
	})
	defer span.End()

	start := time.Now()

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := gz.Close(); closeErr != nil {
			logError(ctx, "Failed to close gzip reader", "error", closeErr)
		}
	}()

	var buf bytes.Buffer
	// Use io.Copy with a size limit to prevent decompression bombs
	written, err := io.Copy(&buf, io.LimitReader(gz, maxDecompressedSize))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	if written >= maxDecompressedSize {
		return nil, fmt.Errorf("%w: decompressed size exceeds limit", ErrDecompressionBomb)
	}

	result := buf.Bytes()
	duration := time.Since(start)
	compressionRatio := float64(len(result)) / float64(len(data))

	SetSpanAttributes(ctx,
		attribute.Int("compression.output_size_bytes", len(result)),
		attribute.Float64("compression.ratio", compressionRatio),
		attribute.Float64("compression.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("compression.success", true),
	)

	return result, nil
}

// storeProtobufBytesOptimized stores protobuf bytes with optimized encoding
func (s *OptimizedProtobufStorage) storeProtobufBytesOptimized(ctx context.Context, datasetID string, data []byte) error {
	logDebug(ctx, "Storing optimized protobuf bytes", "dataset_id", datasetID, "size_bytes", len(data))

	encodedData := base64.StdEncoding.EncodeToString(data)

	protobufDataset := DatasetData{
		Players: []Player{{
			UID:                     -1, // Special marker UID
			Name:                    "__OPTIMIZED_PROTOBUF_DATA__",
			Position:                encodedData,
			Age:                     fmt.Sprintf("%d", len(data)),
			Club:                    "OPTIMIZED_PROTOBUF_STORAGE",
			Division:                "v2", // Version marker for optimized
			TransferValue:           "",
			Wage:                    "",
			Nationality:             "OPTIMIZED_PB",
			NationalityISO:          "OPB",
			NationalityFIFACode:     "OPB",
			Attributes:              make(map[string]string),
			NumericAttributes:       make(map[string]int),
			PerformanceStatsNumeric: make(map[string]float64),
			PerformancePercentiles:  make(map[string]map[string]float64),
			ParsedPositions:         []string{},
			ShortPositions:          []string{},
			PositionGroups:          []string{},
			RoleSpecificOveralls:    []RoleOverallScore{},
		}},
		CurrencySymbol: "__OPTIMIZED_PROTOBUF_MARKER__",
	}

	err := s.backend.Store(datasetID, protobufDataset)
	if err != nil {
		logError(ctx, "Failed to store optimized protobuf dataset", "error", err, "dataset_id", datasetID)
		return fmt.Errorf("failed to store optimized protobuf dataset: %w", err)
	}

	logDebug(ctx, "Successfully stored optimized protobuf bytes", "dataset_id", datasetID)
	return nil
}

// retrieveProtobufBytesOptimized retrieves raw protobuf bytes with optimized decoding
func (s *OptimizedProtobufStorage) retrieveProtobufBytesOptimized(ctx context.Context, datasetID string) ([]byte, error) {
	logDebug(ctx, "Retrieving optimized protobuf bytes", "dataset_id", datasetID)

	dataset, err := s.backend.Retrieve(datasetID)
	if err != nil {
		logError(ctx, "Failed to retrieve optimized protobuf dataset", "error", err, "dataset_id", datasetID)
		return nil, fmt.Errorf("failed to retrieve optimized protobuf dataset: %w", err)
	}

	if len(dataset.Players) != 1 {
		return nil, fmt.Errorf("%w: expected 1 player, got %d", ErrInvalidOptimizedProtobufFormat, len(dataset.Players))
	}

	player := dataset.Players[0]
	if player.Name != "__OPTIMIZED_PROTOBUF_DATA__" {
		return nil, fmt.Errorf("%w: %s", ErrNotOptimizedProtobufDataset, player.Name)
	}

	encodedData := player.Position
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		logError(ctx, "Failed to decode optimized protobuf data", "error", err, "dataset_id", datasetID)
		return nil, fmt.Errorf("failed to decode optimized protobuf data: %w", err)
	}

	logDebug(ctx, "Successfully retrieved optimized protobuf bytes", "dataset_id", datasetID, "size_bytes", len(decodedData))
	return decodedData, nil
}

// storeWithJSONFallback provides fallback to JSON storage
func (s *OptimizedProtobufStorage) storeWithJSONFallback(ctx context.Context, datasetID string, data PlayerDataWithCurrency) error {
	logWarn(ctx, "Falling back to JSON storage for optimized protobuf", "dataset_id", datasetID)
	// Convert PlayerDataWithCurrency to DatasetData
	datasetData := DatasetData{
		Players:        data.Players,
		CurrencySymbol: data.CurrencySymbol,
		CacheData:      "", // PlayerDataWithCurrency doesn't have CacheData field
	}
	return s.backend.Store(datasetID, datasetData)
}

// Implement StorageInterface methods

// Store implements StorageInterface.Store
func (s *OptimizedProtobufStorage) Store(datasetID string, data DatasetData) error {
	// Convert DatasetData to PlayerDataWithCurrency
	playerData := PlayerDataWithCurrency{
		Players:        data.Players,
		CurrencySymbol: data.CurrencySymbol,
	}

	ctx := context.Background()
	return s.StoreWithContext(ctx, datasetID, playerData)
}

// Retrieve implements StorageInterface.Retrieve
func (s *OptimizedProtobufStorage) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	playerData, err := s.RetrieveWithContext(ctx, datasetID)
	if err != nil {
		return DatasetData{}, err
	}

	// Convert PlayerDataWithCurrency to DatasetData
	return DatasetData{
		Players:        playerData.Players,
		CurrencySymbol: playerData.CurrencySymbol,
		CacheData:      "",
	}, nil
}

// Delete implements StorageInterface.Delete
func (s *OptimizedProtobufStorage) Delete(datasetID string) error {
	return s.backend.Delete(datasetID)
}

// List implements StorageInterface.List
func (s *OptimizedProtobufStorage) List() ([]string, error) {
	return s.backend.List()
}

// CleanupOldDatasets implements StorageInterface.CleanupOldDatasets
func (s *OptimizedProtobufStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	return s.backend.CleanupOldDatasets(maxAge, excludeDatasets)
}

// CachedDataset represents a cached dataset entry
type CachedDataset struct {
	Data      PlayerDataWithCurrency
	Timestamp time.Time
	Size      int
}

// GetCacheStats returns cache statistics
func (s *OptimizedProtobufStorage) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})
	var count int
	var totalSize int64

	s.cache.Range(func(_, value interface{}) bool {
		count++
		if cached, ok := value.(*CachedDataset); ok {
			totalSize += int64(cached.Size)
		}
		return true
	})

	stats["cache_entries"] = count
	stats["total_cache_size_bytes"] = totalSize
	stats["cache_enabled"] = s.enableCaching
	stats["compression_level"] = s.compressionLevel

	return stats
}

// ClearCache clears the cache
func (s *OptimizedProtobufStorage) ClearCache() {
	s.cache = sync.Map{}
	logInfo(context.Background(), "Optimized protobuf cache cleared")
}
