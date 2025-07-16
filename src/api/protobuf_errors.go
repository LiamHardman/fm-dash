package main

import (
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
func NewProtobufError(operation, datasetID, message string, cause error) *ProtobufError {
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
func NewProtobufConversionError(direction, dataType, datasetID string, cause error) *ProtobufConversionError {
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
func NewProtobufCompressionError(operation, datasetID string, cause error) *ProtobufCompressionError {
	return &ProtobufCompressionError{
		Operation: operation,
		DatasetID: datasetID,
		Cause:     cause,
	}
}

// ProtobufFallbackReason represents the reason for falling back to JSON
type ProtobufFallbackReason string

const (
	FallbackReasonConversionFailed    ProtobufFallbackReason = "protobuf_conversion_failed"
	FallbackReasonMarshalFailed       ProtobufFallbackReason = "protobuf_marshal_failed"
	FallbackReasonUnmarshalFailed     ProtobufFallbackReason = "protobuf_unmarshal_failed"
	FallbackReasonCompressionFailed   ProtobufFallbackReason = "protobuf_compression_failed"
	FallbackReasonDecompressionFailed ProtobufFallbackReason = "protobuf_decompression_failed"
	FallbackReasonStorageFailed       ProtobufFallbackReason = "protobuf_storage_failed"
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
