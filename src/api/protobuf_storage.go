package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	pb "api/proto"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
)

// ProtobufStorage wraps an existing storage backend and adds protobuf serialization
type ProtobufStorage struct {
	backend StorageInterface
}

// ProtobufPerformanceMetrics holds performance metrics for protobuf operations
type ProtobufPerformanceMetrics struct {
	DatasetID           string
	Operation           string
	PlayerCount         int
	ProtobufSizeBytes   int
	CompressedSizeBytes int
	CompressionRatio    float64
	DurationMs          int64
	SpaceSavedPercent   float64
}

// CreateProtobufStorage creates a new protobuf storage wrapper
func CreateProtobufStorage(backend StorageInterface) *ProtobufStorage {
	return &ProtobufStorage{
		backend: backend,
	}
}

// Store saves a dataset using protobuf serialization
func (s *ProtobufStorage) Store(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.protobuf.store")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "protobuf"),
	)

	start := time.Now()

	// Convert DatasetData to PlayerDataWithCurrency for protobuf conversion
	playerData := &PlayerDataWithCurrency{
		Players:        data.Players,
		CurrencySymbol: data.CurrencySymbol,
	}

	// Convert to protobuf
	protoData, err := playerData.ToProto(ctx)
	if err != nil {
		convErr := NewProtobufConversionError("to_protobuf", "DatasetData", datasetID, err)
		RecordError(ctx, convErr, "Failed to convert data to protobuf")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonConversionFailed,
			Error:     convErr,
			Message:   "Protobuf conversion failed during store operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonConversionFailed)))
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	// Serialize protobuf to binary
	ctx, marshalSpan := StartSpanWithAttributes(ctx, "protobuf.serialization.marshal", []attribute.KeyValue{
		attribute.String("dataset.id", datasetID),
		attribute.String("serialization.format", "protobuf"),
	})
	protoBytes, err := proto.Marshal(protoData)
	marshalSpan.End()
	
	if err != nil {
		marshalErr := NewProtobufError("marshal", datasetID, "Failed to marshal protobuf data", err)
		RecordError(ctx, marshalErr, "Failed to marshal protobuf data")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonMarshalFailed,
			Error:     marshalErr,
			Message:   "Protobuf marshaling failed during store operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonMarshalFailed)))
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}
	
	SetSpanAttributes(ctx, attribute.Int("protobuf.marshaled_size_bytes", len(protoBytes)))

	// Compress protobuf data
	compressedData, err := s.compressProtobufData(protoBytes)
	if err != nil {
		compErr := NewProtobufCompressionError("compress", datasetID, err)
		RecordError(ctx, compErr, "Failed to compress protobuf data")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonCompressionFailed,
			Error:     compErr,
			Message:   "Protobuf compression failed during store operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonCompressionFailed)))
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	SetSpanAttributes(ctx,
		attribute.Int("protobuf.size_bytes", len(protoBytes)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
		attribute.Float64("compression.ratio", float64(len(protoBytes))/float64(len(compressedData))),
	)

	// Store using a custom approach since we need to store raw bytes
	err = s.storeProtobufBytes(ctx, datasetID, compressedData)
	if err != nil {
		storageErr := NewProtobufError("store", datasetID, "Failed to store protobuf data", err)
		RecordError(ctx, storageErr, "Failed to store protobuf data")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonStorageFailed,
			Error:     storageErr,
			Message:   "Protobuf storage failed during store operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonStorageFailed)))
		return s.storeWithJSONFallback(ctx, datasetID, data)
	}

	duration := time.Since(start)
	compressionRatio := float64(len(protoBytes)) / float64(len(compressedData))
	spaceSavedPercent := (1.0 - 1.0/compressionRatio) * 100

	RecordDBOperation(ctx, "store", "protobuf_datasets", duration, 1)
	
	// Log detailed performance and storage metrics
	logInfo(ctx, "Successfully stored dataset using protobuf serialization", 
		"dataset_id", datasetID, 
		"player_count", len(data.Players),
		"duration_ms", duration.Milliseconds(),
		"protobuf_size_bytes", len(protoBytes),
		"compressed_size_bytes", len(compressedData),
		"compression_ratio", fmt.Sprintf("%.2f", compressionRatio),
		"space_saved_percent", fmt.Sprintf("%.1f%%", spaceSavedPercent),
		"storage_format", "protobuf",
		"compression_algorithm", "gzip")
	
	// Log performance improvement metrics
	s.logPerformanceImprovement(ctx, ProtobufPerformanceMetrics{
		DatasetID:           datasetID,
		Operation:           "store",
		PlayerCount:         len(data.Players),
		ProtobufSizeBytes:   len(protoBytes),
		CompressedSizeBytes: len(compressedData),
		CompressionRatio:    compressionRatio,
		DurationMs:          duration.Milliseconds(),
		SpaceSavedPercent:   spaceSavedPercent,
	})
	
	return nil
}

// Retrieve retrieves a dataset using protobuf deserialization
func (s *ProtobufStorage) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.protobuf.retrieve")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("storage.type", "protobuf"),
	)

	start := time.Now()

	// Try to retrieve protobuf data first
	compressedData, err := s.retrieveProtobufBytes(ctx, datasetID)
	if err != nil {
		// If protobuf retrieval fails, try JSON fallback
		retrievalErr := NewProtobufError("retrieve", datasetID, "Failed to retrieve protobuf data", err)
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonRetrievalFailed,
			Error:     retrievalErr,
			Message:   "Protobuf retrieval failed during retrieve operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonRetrievalFailed)))
		return s.retrieveWithJSONFallback(ctx, datasetID)
	}

	// Decompress protobuf data
	protoBytes, err := s.decompressProtobufData(compressedData)
	if err != nil {
		decompErr := NewProtobufCompressionError("decompress", datasetID, err)
		RecordError(ctx, decompErr, "Failed to decompress protobuf data")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonDecompressionFailed,
			Error:     decompErr,
			Message:   "Protobuf decompression failed during retrieve operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonDecompressionFailed)))
		return s.retrieveWithJSONFallback(ctx, datasetID)
	}

	// Unmarshal protobuf data
	ctx, unmarshalSpan := StartSpanWithAttributes(ctx, "protobuf.serialization.unmarshal", []attribute.KeyValue{
		attribute.String("dataset.id", datasetID),
		attribute.String("serialization.format", "protobuf"),
		attribute.Int("protobuf.size_bytes", len(protoBytes)),
	})
	var protoData pb.DatasetData
	err = proto.Unmarshal(protoBytes, &protoData)
	unmarshalSpan.End()
	
	if err != nil {
		unmarshalErr := NewProtobufError("unmarshal", datasetID, "Failed to unmarshal protobuf data", err)
		RecordError(ctx, unmarshalErr, "Failed to unmarshal protobuf data")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonUnmarshalFailed,
			Error:     unmarshalErr,
			Message:   "Protobuf unmarshaling failed during retrieve operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonUnmarshalFailed)))
		return s.retrieveWithJSONFallback(ctx, datasetID)
	}

	// Convert from protobuf to native structs
	playerData, err := DatasetDataFromProto(ctx, &protoData)
	if err != nil {
		convErr := NewProtobufConversionError("from_protobuf", "DatasetData", datasetID, err)
		RecordError(ctx, convErr, "Failed to convert protobuf to native structs")
		s.logFallbackEvent(ProtobufFallbackEvent{
			DatasetID: datasetID,
			Reason:    FallbackReasonConversionFailed,
			Error:     convErr,
			Message:   "Protobuf conversion failed during retrieve operation",
		})
		AddSpanEvent(ctx, "storage.fallback_to_json", attribute.String("reason", string(FallbackReasonConversionFailed)))
		return s.retrieveWithJSONFallback(ctx, datasetID)
	}

	// Convert back to DatasetData format
	result := DatasetData{
		Players:        playerData.Players,
		CurrencySymbol: playerData.CurrencySymbol,
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.player_count", len(result.Players)),
		attribute.Int("protobuf.size_bytes", len(protoBytes)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
	)

	duration := time.Since(start)
	compressionRatio := float64(len(protoBytes)) / float64(len(compressedData))
	spaceSavedPercent := (1.0 - 1.0/compressionRatio) * 100

	RecordDBOperation(ctx, "retrieve", "protobuf_datasets", duration, 1)
	
	// Log detailed performance and storage metrics
	logInfo(ctx, "Successfully retrieved dataset using protobuf deserialization", 
		"dataset_id", datasetID, 
		"player_count", len(result.Players),
		"duration_ms", duration.Milliseconds(),
		"protobuf_size_bytes", len(protoBytes),
		"compressed_size_bytes", len(compressedData),
		"compression_ratio", fmt.Sprintf("%.2f", compressionRatio),
		"space_saved_percent", fmt.Sprintf("%.1f%%", spaceSavedPercent),
		"storage_format", "protobuf",
		"compression_algorithm", "gzip")
	
	// Log performance improvement metrics
	s.logPerformanceImprovement(ctx, ProtobufPerformanceMetrics{
		DatasetID:           datasetID,
		Operation:           "retrieve",
		PlayerCount:         len(result.Players),
		ProtobufSizeBytes:   len(protoBytes),
		CompressedSizeBytes: len(compressedData),
		CompressionRatio:    compressionRatio,
		DurationMs:          duration.Milliseconds(),
		SpaceSavedPercent:   spaceSavedPercent,
	})
	
	return result, nil
}

// Delete removes a dataset
func (s *ProtobufStorage) Delete(datasetID string) error {
	return s.backend.Delete(datasetID)
}

// List returns all dataset IDs
func (s *ProtobufStorage) List() ([]string, error) {
	return s.backend.List()
}

// CleanupOldDatasets removes old datasets
func (s *ProtobufStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	return s.backend.CleanupOldDatasets(maxAge, excludeDatasets)
}

// compressProtobufData compresses protobuf binary data using gzip
func (s *ProtobufStorage) compressProtobufData(data []byte) ([]byte, error) {
	ctx := context.Background()
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.compression.compress", []attribute.KeyValue{
		attribute.String("compression.algorithm", "gzip"),
		attribute.Int("compression.input_size_bytes", len(data)),
	})
	defer span.End()

	start := time.Now()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		if closeErr := gz.Close(); closeErr != nil {
			// Note: We can't use context here as it's in a defer, but this is a minor cleanup error
		}
	}()

	if _, err := gz.Write(data); err != nil {
		RecordError(ctx, err, "Failed to write to gzip writer",
			WithErrorCategory("compression"),
			WithSeverity("medium"))
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := gz.Close(); err != nil {
		RecordError(ctx, err, "Failed to close gzip writer",
			WithErrorCategory("compression"),
			WithSeverity("medium"))
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

	AddSpanEvent(ctx, "compression.completed", 
		attribute.Float64("compression.space_saved_percent", (1.0-1.0/compressionRatio)*100))

	return result, nil
}

// decompressProtobufData decompresses gzip-compressed protobuf data
func (s *ProtobufStorage) decompressProtobufData(data []byte) ([]byte, error) {
	ctx := context.Background()
	ctx, span := StartSpanWithAttributes(ctx, "protobuf.compression.decompress", []attribute.KeyValue{
		attribute.String("compression.algorithm", "gzip"),
		attribute.Int("compression.input_size_bytes", len(data)),
	})
	defer span.End()

	start := time.Now()
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		RecordError(ctx, err, "Failed to create gzip reader",
			WithErrorCategory("compression"),
			WithSeverity("medium"))
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			// Note: We can't use context here as it's in a defer, but this is a minor cleanup error
		}
	}()

	result, err := io.ReadAll(reader)
	if err != nil {
		RecordError(ctx, err, "Failed to read from gzip reader",
			WithErrorCategory("compression"),
			WithSeverity("medium"))
		return nil, fmt.Errorf("failed to read from gzip reader: %w", err)
	}

	duration := time.Since(start)
	decompressionRatio := float64(len(result)) / float64(len(data))

	SetSpanAttributes(ctx,
		attribute.Int("compression.output_size_bytes", len(result)),
		attribute.Float64("compression.ratio", decompressionRatio),
		attribute.Float64("compression.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("compression.success", true),
	)

	AddSpanEvent(ctx, "decompression.completed", 
		attribute.Float64("compression.expansion_ratio", decompressionRatio))

	return result, nil
}

// storeProtobufBytes stores raw protobuf bytes by encoding them as a special dataset
func (s *ProtobufStorage) storeProtobufBytes(ctx context.Context, datasetID string, data []byte) error {
	logDebug(ctx, "Storing protobuf bytes", "dataset_id", datasetID, "size_bytes", len(data))
	
	// Encode the protobuf bytes as a base64 string in a special player record
	// This allows us to use the existing storage interface without modification
	encodedData := base64.StdEncoding.EncodeToString(data)
	
	// Create a special dataset that contains the encoded protobuf data
	protobufDataset := DatasetData{
		Players: []Player{{
			UID:                 -1, // Special marker UID to indicate protobuf data
			Name:                "__PROTOBUF_DATA__",
			Position:            encodedData, // Store encoded data in position field
			Age:                 fmt.Sprintf("%d", len(data)), // Store original size
			Club:                "PROTOBUF_STORAGE",
			Division:            "v1", // Version marker
			TransferValue:       "",
			Wage:                "",
			Nationality:         "PROTOBUF",
			NationalityISO:      "PB",
			NationalityFIFACode: "PB",
			Attributes:          make(map[string]string),
			NumericAttributes:   make(map[string]int),
			PerformanceStatsNumeric: make(map[string]float64),
			PerformancePercentiles:  make(map[string]map[string]float64),
			ParsedPositions:     []string{},
			ShortPositions:      []string{},
			PositionGroups:      []string{},
			RoleSpecificOveralls: []RoleOverallScore{},
		}},
		CurrencySymbol: "__PROTOBUF_MARKER__",
	}
	
	err := s.backend.Store(datasetID, protobufDataset)
	if err != nil {
		logError(ctx, "Failed to store protobuf dataset", "error", err, "dataset_id", datasetID)
		return fmt.Errorf("failed to store protobuf dataset: %w", err)
	}
	
	logDebug(ctx, "Successfully stored protobuf bytes", "dataset_id", datasetID)
	return nil
}

// retrieveProtobufBytes retrieves raw protobuf bytes by decoding from the special dataset
func (s *ProtobufStorage) retrieveProtobufBytes(ctx context.Context, datasetID string) ([]byte, error) {
	logDebug(ctx, "Retrieving protobuf bytes", "dataset_id", datasetID)
	
	// Retrieve the special dataset
	dataset, err := s.backend.Retrieve(datasetID)
	if err != nil {
		logError(ctx, "Failed to retrieve protobuf dataset", "error", err, "dataset_id", datasetID)
		return nil, fmt.Errorf("failed to retrieve protobuf dataset: %w", err)
	}
	
	// Check if this is a protobuf dataset
	if dataset.CurrencySymbol != "__PROTOBUF_MARKER__" {
		logDebug(ctx, "Dataset is not in protobuf format", "dataset_id", datasetID, "currency_symbol", dataset.CurrencySymbol)
		return nil, fmt.Errorf("dataset is not in protobuf format")
	}
	
	if len(dataset.Players) != 1 || dataset.Players[0].UID != -1 || dataset.Players[0].Name != "__PROTOBUF_DATA__" {
		logError(ctx, "Invalid protobuf dataset format", "dataset_id", datasetID, "player_count", len(dataset.Players))
		return nil, fmt.Errorf("invalid protobuf dataset format")
	}
	
	// Decode the base64 data
	encodedData := dataset.Players[0].Position
	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		logError(ctx, "Failed to decode protobuf data", "error", err, "dataset_id", datasetID)
		return nil, fmt.Errorf("failed to decode protobuf data: %w", err)
	}
	
	logDebug(ctx, "Successfully retrieved protobuf bytes", "dataset_id", datasetID, "size_bytes", len(data))
	return data, nil
}

// storeWithJSONFallback falls back to JSON storage when protobuf fails
func (s *ProtobufStorage) storeWithJSONFallback(ctx context.Context, datasetID string, data DatasetData) error {
	logInfo(ctx, "Using JSON fallback for storing dataset", "dataset_id", datasetID)
	return s.backend.Store(datasetID, data)
}

// retrieveWithJSONFallback falls back to JSON retrieval when protobuf fails
func (s *ProtobufStorage) retrieveWithJSONFallback(ctx context.Context, datasetID string) (DatasetData, error) {
	logInfo(ctx, "Using JSON fallback for retrieving dataset", "dataset_id", datasetID)
	return s.backend.Retrieve(datasetID)
}

// logFallbackEvent logs protobuf fallback events for monitoring and debugging
func (s *ProtobufStorage) logFallbackEvent(event ProtobufFallbackEvent) {
	// Create a context for logging - in a real scenario this would be passed from the caller
	ctx := context.Background()
	
	// Log fallback event with detailed context
	logWarn(ctx, "Protobuf fallback event occurred - switching to JSON storage", 
		"dataset_id", event.DatasetID,
		"fallback_reason", string(event.Reason),
		"fallback_message", event.Message,
		"original_error", event.Error.Error(),
		"storage_format", "json",
		"fallback_triggered", true)
	
	// Log structured fallback metrics for monitoring
	logInfo(ctx, "Storage fallback metrics", 
		"dataset_id", event.DatasetID,
		"fallback_type", "protobuf_to_json",
		"failure_stage", string(event.Reason),
		"impact", "performance_degradation",
		"action_required", s.getFallbackActionRequired(event.Reason))
}

// logPerformanceImprovement logs detailed performance improvement metrics
func (s *ProtobufStorage) logPerformanceImprovement(ctx context.Context, metrics ProtobufPerformanceMetrics) {
	// Calculate estimated JSON size for comparison (rough estimate: protobuf is typically 20-50% smaller)
	estimatedJSONSize := float64(metrics.ProtobufSizeBytes) * 1.3 // Conservative estimate
	estimatedJSONSizeCompressed := estimatedJSONSize * 0.7 // Typical JSON gzip compression
	
	jsonSpaceSaved := (estimatedJSONSizeCompressed - float64(metrics.CompressedSizeBytes)) / estimatedJSONSizeCompressed * 100
	
	// Log comprehensive performance improvement metrics
	logInfo(ctx, "Protobuf performance improvement achieved",
		"dataset_id", metrics.DatasetID,
		"operation", metrics.Operation,
		"player_count", metrics.PlayerCount,
		"protobuf_size_bytes", metrics.ProtobufSizeBytes,
		"compressed_size_bytes", metrics.CompressedSizeBytes,
		"compression_ratio", fmt.Sprintf("%.2f", metrics.CompressionRatio),
		"space_saved_percent", fmt.Sprintf("%.1f%%", metrics.SpaceSavedPercent),
		"duration_ms", metrics.DurationMs,
		"estimated_json_space_saved_percent", fmt.Sprintf("%.1f%%", jsonSpaceSaved),
		"storage_efficiency", "optimized",
		"serialization_format", "protobuf+gzip")
	
	// Log performance benchmarking data
	throughputMBps := float64(metrics.ProtobufSizeBytes) / (float64(metrics.DurationMs) / 1000.0) / (1024 * 1024)
	
	// Calculate per-player metrics only if player count > 0
	var bytesPerPlayer int
	var msPerPlayer float64
	if metrics.PlayerCount > 0 {
		bytesPerPlayer = metrics.ProtobufSizeBytes / metrics.PlayerCount
		msPerPlayer = float64(metrics.DurationMs) / float64(metrics.PlayerCount)
	}
	
	logDebug(ctx, "Protobuf operation performance metrics",
		"dataset_id", metrics.DatasetID,
		"operation", metrics.Operation,
		"throughput_mbps", fmt.Sprintf("%.2f", throughputMBps),
		"bytes_per_player", bytesPerPlayer,
		"ms_per_player", msPerPlayer,
		"compression_efficiency", fmt.Sprintf("%.1f%%", metrics.SpaceSavedPercent))
}

// getFallbackActionRequired returns recommended action for different fallback reasons
func (s *ProtobufStorage) getFallbackActionRequired(reason ProtobufFallbackReason) string {
	switch reason {
	case FallbackReasonConversionFailed:
		return "check_data_schema_compatibility"
	case FallbackReasonMarshalFailed, FallbackReasonUnmarshalFailed:
		return "verify_protobuf_schema_version"
	case FallbackReasonCompressionFailed, FallbackReasonDecompressionFailed:
		return "check_compression_library_status"
	case FallbackReasonStorageFailed, FallbackReasonRetrievalFailed:
		return "verify_storage_backend_health"
	default:
		return "investigate_protobuf_implementation"
	}
}