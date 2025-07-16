package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	pb "api/proto"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
)

// ProtobufStorage wraps an existing storage backend and adds protobuf serialization
type ProtobufStorage struct {
	backend StorageInterface
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
	protoBytes, err := proto.Marshal(protoData)
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

	RecordDBOperation(ctx, "store", "protobuf_datasets", time.Since(start), 1)
	logInfo(ctx, "Successfully stored dataset using protobuf serialization", 
		"dataset_id", datasetID, 
		"player_count", len(data.Players),
		"duration_ms", time.Since(start).Milliseconds())
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
	var protoData pb.DatasetData
	if err := proto.Unmarshal(protoBytes, &protoData); err != nil {
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

	RecordDBOperation(ctx, "retrieve", "protobuf_datasets", time.Since(start), 1)
	logInfo(ctx, "Successfully retrieved dataset using protobuf deserialization", 
		"dataset_id", datasetID, 
		"player_count", len(result.Players),
		"duration_ms", time.Since(start).Milliseconds())
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
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		if closeErr := gz.Close(); closeErr != nil {
			// Note: We can't use context here as it's in a defer, but this is a minor cleanup error
		}
	}()

	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressProtobufData decompresses gzip-compressed protobuf data
func (s *ProtobufStorage) decompressProtobufData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			// Note: We can't use context here as it's in a defer, but this is a minor cleanup error
		}
	}()

	result, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip reader: %w", err)
	}

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
	logInfo(ctx, "Protobuf fallback event occurred", 
		"dataset_id", event.DatasetID,
		"reason", string(event.Reason),
		"message", event.Message,
		"error", event.Error)
	
	// In a production environment, you might want to send this to a monitoring system
	// such as metrics collection, alerting, or structured logging
}