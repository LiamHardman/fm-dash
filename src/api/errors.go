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
	ErrNotProtobufFormat     = errors.New("dataset is not in protobuf format")
	ErrInvalidProtobufFormat = errors.New("invalid protobuf dataset format")
	ErrProtobufStorageWrapper = errors.New("failed to create protobuf storage wrapper")
	ErrInvalidProtobufEnv     = errors.New("invalid protobuf environment value")
	ErrConfigValidationFailed = errors.New("configuration validation failed")
	
	// Test errors
	ErrTestError           = errors.New("test error")
	ErrUnderlyingError     = errors.New("underlying error")
	ErrConversionFailed    = errors.New("conversion failed")
	ErrCompressionFailed   = errors.New("compression failed")
	ErrStorageFailure      = errors.New("simulated storage failure")
	ErrRetrievalFailure    = errors.New("simulated retrieval failure")
	ErrDeletionFailure     = errors.New("simulated deletion failure")
	ErrListFailure         = errors.New("simulated list failure")
	ErrWorkerRequestFailed = errors.New("worker request failed")
	ErrPlayersRequestFailed = errors.New("players request failed")
	ErrLeaguesRequestFailed = errors.New("leagues request failed")
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