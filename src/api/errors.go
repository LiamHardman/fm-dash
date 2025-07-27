package main

import (
	"errors"
	"fmt"
)

// Static errors for err113 compliance
var (
	// Protobuf conversion errors
	ErrNilRoleOverallScore         = errors.New("nil RoleOverallScore")
	ErrNilProtobufRoleOverallScore = errors.New("nil protobuf RoleOverallScore")
	ErrNilPlayer                   = errors.New("nil Player")
	ErrNilProtobufPlayer           = errors.New("nil protobuf Player")
	ErrNilDatasetData              = errors.New("nil DatasetData")
	ErrNilProtobufDatasetData      = errors.New("nil protobuf DatasetData")

	// Storage errors
	ErrNotProtobufFormat      = errors.New("dataset is not in protobuf format")
	ErrInvalidProtobufFormat  = errors.New("invalid protobuf dataset format")
	ErrProtobufStorageWrapper = errors.New("failed to create protobuf storage wrapper")
	ErrInvalidProtobufEnv     = errors.New("invalid protobuf environment value")
	ErrConfigValidationFailed = errors.New("configuration validation failed")

	// Serialization errors
	ErrUnsupportedDataType    = errors.New("unsupported data type for protobuf serialization")
	ErrInvalidQualityValue    = errors.New("invalid quality value")
	ErrSerializationFailed    = errors.New("both protobuf and JSON serialization failed")
	ErrSimulatedSerialization = errors.New("simulated serialization error")
	ErrSimulatedConversion    = errors.New("simulated conversion error")
	ErrSimulatedCompression   = errors.New("simulated compression error")

	// Request errors
	ErrFailedToCreateRequest       = errors.New("failed to create request")
	ErrFailedToSendRequest         = errors.New("failed to send request")
	ErrUnexpectedStatusCode        = errors.New("unexpected status code")
	ErrExpectedProtobufContentType = errors.New("expected protobuf content type")
	ErrFailedToReadResponseBody    = errors.New("failed to read response body")
	ErrFailedToUnmarshalProtobuf   = errors.New("failed to unmarshal protobuf response")
	ErrUnexpectedPlayerCount       = errors.New("unexpected player count")
	ErrExpectedJSONContentType     = errors.New("expected JSON content type")
	ErrFailedToDecodeJSON          = errors.New("failed to decode JSON response")
	ErrUnexpectedPlayerCountJSON   = errors.New("unexpected player count in JSON")
	ErrExpectedFallbackHeader      = errors.New("expected fallback header")

	// Service errors
	ErrDatasetIDRequired   = errors.New("dataset ID is required")
	ErrNoPlayersToStore    = errors.New("no players to store")
	ErrNoPlayersToProcess  = errors.New("no players to process")
	ErrNoPlayersToValidate = errors.New("no players to validate")
	ErrPlayerNoName        = errors.New("player has no name")
	ErrPlayerNoUID         = errors.New("player has no UID")
	ErrInvalidFileFormat   = errors.New("invalid file format: expected HTML content")
	ErrDatasetNotFound     = errors.New("dataset not found in store")

	// Configuration errors
	ErrInvalidOTELEndpoint = errors.New("invalid OTEL_EXPORTER_OTLP_ENDPOINT")
	ErrInvalidS3Endpoint   = errors.New("invalid S3_ENDPOINT")
	ErrInvalidServiceName  = errors.New("invalid SERVICE_NAME: contains unsafe characters")

	// Test errors
	ErrTestError             = errors.New("test error")
	ErrUnderlyingError       = errors.New("underlying error")
	ErrConversionFailed      = errors.New("conversion failed")
	ErrCompressionFailed     = errors.New("compression failed")
	ErrStorageFailure        = errors.New("simulated storage failure")
	ErrRetrievalFailure      = errors.New("simulated retrieval failure")
	ErrDeletionFailure       = errors.New("simulated deletion failure")
	ErrListFailure           = errors.New("simulated list failure")
	ErrWorkerRequestFailed   = errors.New("worker request failed")
	ErrPlayersRequestFailed  = errors.New("players request failed")
	ErrLeaguesRequestFailed  = errors.New("leagues request failed")
	ErrFilteredRequestFailed = errors.New("filtered request failed")
)

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorf wraps an error with formatted context
func WrapErrorf(err error, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}
