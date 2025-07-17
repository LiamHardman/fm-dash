package main

import (
	"context"
	"fmt"
)

// ProtobufError represents errors specific to protobuf operations
type ProtobufError struct {
	Operation string // The operation that failed (e.g., "marshal", "unmarshal", "compress")
	DatasetID string // The dataset ID involved in the operation
	Cause     error  // The underlying error
	Message   string // Additional context message
}

func (e *ProtobufError) Error() string {
	if e.DatasetID != "" {
		return fmt.Sprintf("protobuf %s failed for dataset %s: %s", e.Operation, e.DatasetID, e.Message)
	}
	return fmt.Sprintf("protobuf %s failed: %s", e.Operation, e.Message)
}

func (e *ProtobufError) Unwrap() error {
	return e.Cause
}

// NewProtobufError creates a new ProtobufError
func NewProtobufError(ctx context.Context, operation, datasetID, message string, cause error) *ProtobufError {
	logError(ctx, "Creating protobuf error",
		"error", cause,
		"operation", operation,
		"dataset_id", datasetID,
		"message", message)
	
	return &ProtobufError{
		Operation: operation,
		DatasetID: datasetID,
		Message:   message,
		Cause:     cause,
	}
}

// NewProtobufErrorLegacy creates a new ProtobufError without context (for backward compatibility)
func NewProtobufErrorLegacy(operation, datasetID, message string, cause error) *ProtobufError {
	return &ProtobufError{
		Operation: operation,
		DatasetID: datasetID,
		Message:   message,
		Cause:     cause,
	}
}

// ProtobufConversionError represents errors during protobuf conversion
type ProtobufConversionError struct {
	Direction string // "to_protobuf" or "from_protobuf"
	DataType  string // "Player", "DatasetData", "RoleOverallScore"
	DatasetID string
	Cause     error
}

func (e *ProtobufConversionError) Error() string {
	return fmt.Sprintf("protobuf conversion error (%s %s) for dataset %s: %v", e.Direction, e.DataType, e.DatasetID, e.Cause)
}

func (e *ProtobufConversionError) Unwrap() error {
	return e.Cause
}

// NewProtobufConversionError creates a new ProtobufConversionError
func NewProtobufConversionError(ctx context.Context, direction, dataType, datasetID string, cause error) *ProtobufConversionError {
	logError(ctx, "Creating protobuf conversion error",
		"error", cause,
		"direction", direction,
		"data_type", dataType,
		"dataset_id", datasetID)
	
	return &ProtobufConversionError{
		Direction: direction,
		DataType:  dataType,
		DatasetID: datasetID,
		Cause:     cause,
	}
}

// NewProtobufConversionErrorLegacy creates a new ProtobufConversionError without context (for backward compatibility)
func NewProtobufConversionErrorLegacy(direction, dataType, datasetID string, cause error) *ProtobufConversionError {
	return &ProtobufConversionError{
		Direction: direction,
		DataType:  dataType,
		DatasetID: datasetID,
		Cause:     cause,
	}
}

// ProtobufCompressionError represents errors during compression/decompression
type ProtobufCompressionError struct {
	Operation string // "compress" or "decompress"
	DatasetID string
	Cause     error
}

func (e *ProtobufCompressionError) Error() string {
	return fmt.Sprintf("protobuf %s error for dataset %s: %v", e.Operation, e.DatasetID, e.Cause)
}

func (e *ProtobufCompressionError) Unwrap() error {
	return e.Cause
}

// NewProtobufCompressionError creates a new ProtobufCompressionError
func NewProtobufCompressionError(ctx context.Context, operation, datasetID string, cause error) *ProtobufCompressionError {
	logError(ctx, "Creating protobuf compression error",
		"error", cause,
		"operation", operation,
		"dataset_id", datasetID)
	
	return &ProtobufCompressionError{
		Operation: operation,
		DatasetID: datasetID,
		Cause:     cause,
	}
}

// NewProtobufCompressionErrorLegacy creates a new ProtobufCompressionError without context (for backward compatibility)
func NewProtobufCompressionErrorLegacy(operation, datasetID string, cause error) *ProtobufCompressionError {
	return &ProtobufCompressionError{
		Operation: operation,
		DatasetID: datasetID,
		Cause:     cause,
	}
}



// ProtobufFallbackReason represents the reason for falling back to JSON
type ProtobufFallbackReason string

const (
	// FallbackReasonConversionFailed indicates protobuf conversion failed
	FallbackReasonConversionFailed    ProtobufFallbackReason = "protobuf_conversion_failed"
	// FallbackReasonMarshalFailed indicates protobuf marshaling failed
	FallbackReasonMarshalFailed       ProtobufFallbackReason = "protobuf_marshal_failed"
	// FallbackReasonUnmarshalFailed indicates protobuf unmarshaling failed
	FallbackReasonUnmarshalFailed     ProtobufFallbackReason = "protobuf_unmarshal_failed"
	// FallbackReasonCompressionFailed indicates protobuf compression failed
	FallbackReasonCompressionFailed   ProtobufFallbackReason = "protobuf_compression_failed"
	// FallbackReasonDecompressionFailed indicates protobuf decompression failed
	FallbackReasonDecompressionFailed ProtobufFallbackReason = "protobuf_decompression_failed"
	// FallbackReasonStorageFailed indicates protobuf storage operation failed
	FallbackReasonStorageFailed       ProtobufFallbackReason = "protobuf_storage_failed"
	// FallbackReasonRetrievalFailed indicates protobuf data retrieval failed
	FallbackReasonRetrievalFailed     ProtobufFallbackReason = "protobuf_retrieval_failed"
)

// ProtobufFallbackEvent represents a fallback event for logging and monitoring
type ProtobufFallbackEvent struct {
	DatasetID string
	Reason    ProtobufFallbackReason
	Error     error
	Message   string
}

func (e *ProtobufFallbackEvent) String() string {
	return fmt.Sprintf("Protobuf fallback for dataset %s: %s - %s (error: %v)",
		e.DatasetID, e.Reason, e.Message, e.Error)
}
