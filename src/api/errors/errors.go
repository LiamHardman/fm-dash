// Package errors provides centralized error definitions for the FM24 API
package errors

import (
	"errors"
	"fmt"
)

// AppError represents a structured application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
	Internal   error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	CodeValidationFailed  = "VALIDATION_FAILED"
	CodeNotFound          = "NOT_FOUND"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeConflict          = "CONFLICT"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeBadRequest        = "BAD_REQUEST"
	CodeFileTooLarge      = "FILE_TOO_LARGE"
	CodeUnsupportedFormat = "UNSUPPORTED_FORMAT"
	CodeProcessingFailed  = "PROCESSING_FAILED"
	CodeStorageError      = "STORAGE_ERROR"
	CodeRateLimited       = "RATE_LIMITED"
)

// Static error definitions to avoid dynamic error creation (err113)
var (
	// Cache errors
	ErrInvalidCacheDataFormat = errors.New("invalid cache data format")
	ErrCacheNotFound          = errors.New("cache not found")
	ErrCacheStoreFailed       = errors.New("failed to store cache")
	ErrCacheDeleteFailed      = errors.New("failed to delete cache")

	// Processing errors
	ErrProcessingFailed                 = errors.New("processing failed")
	ErrInvalidDataFormat                = errors.New("invalid data format")
	ErrWorkerPanic                      = errors.New("worker panic")
	ErrHTMLTokenization                 = errors.New("HTML tokenization error")
	ErrNoTableHeadersFound              = errors.New("no table headers found in HTML file")
	ErrHeadersFoundButWorkersNotStarted = errors.New("headers found but workers were not started")
	ErrTokenizingHTML                   = errors.New("error tokenizing HTML")
	ErrFailedToParseAppearances         = errors.New("failed to parse appearances")
	ErrInvalidAppearancesFormat         = errors.New("invalid appearances format")
	ErrEmptyString                      = errors.New("empty string")
	ErrInvalidCharacter                 = errors.New("invalid character")
	ErrCannotProcessRow                 = errors.New("cannot process row: headers are empty")
	ErrSkippedRowEmpty                  = errors.New("skipped row: 'Name' field missing and row appears empty or is likely a non-player row (e.g., header repetition, spacer)")
	ErrSkippedRowNameMissing            = errors.New("skipped row: 'Name' field is missing or empty, but other data present")

	// Configuration errors
	ErrEmptyWeightsFile       = errors.New("loaded weights file is empty")
	ErrConfigInitTimeout      = errors.New("configuration initialization timed out")
	ErrInvalidOtelEndpoint    = errors.New("invalid OTEL_EXPORTER_OTLP_ENDPOINT")
	ErrInvalidS3Endpoint      = errors.New("invalid S3_ENDPOINT format")
	ErrInvalidServiceName     = errors.New("invalid SERVICE_NAME: contains unsafe characters")
	ErrServiceNameEmpty       = errors.New("service name cannot be empty")
	ErrCollectorURLEmpty      = errors.New("collector URL cannot be empty")
	ErrInvalidTraceSampleRate = errors.New("invalid trace sample rate")
	ErrInvalidBatchSize       = errors.New("invalid batch size")
	ErrInvalidMaxQueueSize    = errors.New("invalid max queue size")

	// Security errors
	ErrFilenameEmpty               = errors.New("filename cannot be empty")
	ErrFilenameTooLong             = errors.New("filename too long (max 255 characters)")
	ErrFilenamePathTraversal       = errors.New("filename contains path traversal sequence")
	ErrFilenameDirectorySeparators = errors.New("filename contains directory separators")
	ErrFilenameNullBytes           = errors.New("filename contains null bytes")
	ErrFilenameInvalidChars        = errors.New("filename contains invalid characters")

	// ID validation errors
	ErrIDEmpty               = errors.New("ID cannot be empty")
	ErrIDTooLong             = errors.New("ID too long")
	ErrIDPathTraversal       = errors.New("ID contains path traversal sequence")
	ErrIDDirectorySeparators = errors.New("ID contains directory separators")
	ErrIDNullBytes           = errors.New("ID contains null bytes")
	ErrIDInvalidChars        = errors.New("ID contains invalid characters")

	// Panic and recovery errors
	ErrPathEscapesBase        = errors.New("path escapes base directory")
	ErrPanicRecoveredNonError = errors.New("panic recovered (non-error value)")

	// Service errors
	ErrDatasetIDEmpty         = errors.New("dataset ID cannot be empty")
	ErrDatasetNotFound        = errors.New("dataset not found")
	ErrNoPlayersToStore       = errors.New("no players to store")
	ErrNoPlayersToProcess     = errors.New("no players to process")
	ErrNoPlayersForValidation = errors.New("no players provided for validation")
	ErrPlayerNoName           = errors.New("player has no name")
	ErrValidationFailed       = errors.New("validation failed")
	ErrFileContentEmpty       = errors.New("file content is empty")
	ErrUnsupportedFileFormat  = errors.New("unsupported file format, expected .html file")
	ErrInvalidHTML            = errors.New("file does not appear to be valid HTML")
	ErrS3ClientNotAvailable   = errors.New("S3 client not available")

	// Storage errors
	ErrJSONMarshalPanic = errors.New("panic during JSON marshal")

	// HTTP errors
	ErrHTTPStatus     = errors.New("HTTP status error")
	ErrHandlerPanic   = errors.New("handler panic")
	ErrPanicRecovered = errors.New("panic recovered")

	// File processing errors
	ErrInvalidUID           = errors.New("invalid UID")
	ErrInvalidFilePath      = errors.New("invalid file path")
	ErrInvalidTeamID        = errors.New("invalid team ID")
	ErrInvalidClubsDirPath  = errors.New("invalid clubs directory path")
	ErrInvalidNormalDirPath = errors.New("invalid normal directory path")
	ErrInvalidDatasetID     = errors.New("invalid dataset ID")
)

// WrapErrInvalidCacheDataFormat wraps an error with context about invalid cache data format
func WrapErrInvalidCacheDataFormat(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidCacheDataFormat, err)
}

// WrapErrConfigInitTimeout wraps a config initialization timeout error with context
func WrapErrConfigInitTimeout(timeout interface{}) error {
	return fmt.Errorf("%w after %v", ErrConfigInitTimeout, timeout)
}

// WrapErrInvalidOtelEndpoint wraps an invalid OpenTelemetry endpoint error with context
func WrapErrInvalidOtelEndpoint(endpoint string) error {
	return fmt.Errorf("%w: %s (must include port)", ErrInvalidOtelEndpoint, endpoint)
}

// WrapErrInvalidS3Endpoint wraps an invalid S3 endpoint error with context
func WrapErrInvalidS3Endpoint(endpoint string) error {
	return fmt.Errorf("%w: %s (should include port or be full URL)", ErrInvalidS3Endpoint, endpoint)
}

// WrapErrInvalidTraceSampleRate wraps an invalid trace sample rate error with context
func WrapErrInvalidTraceSampleRate(rate float64) error {
	return fmt.Errorf("%w, got %f", ErrInvalidTraceSampleRate, rate)
}

// WrapErrInvalidBatchSize wraps an invalid batch size error with context
func WrapErrInvalidBatchSize(size int) error {
	return fmt.Errorf("%w, got %d", ErrInvalidBatchSize, size)
}

// WrapErrInvalidMaxQueueSize wraps an invalid max queue size error with context
func WrapErrInvalidMaxQueueSize(size int) error {
	return fmt.Errorf("%w, got %d", ErrInvalidMaxQueueSize, size)
}

// WrapErrInvalidUID wraps an invalid UID error with context
func WrapErrInvalidUID(uid string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidUID, uid, err)
}

// WrapErrInvalidFilePath wraps an invalid file path error with context
func WrapErrInvalidFilePath(uid string, err error) error {
	return fmt.Errorf("%w for UID %s: %v", ErrInvalidFilePath, uid, err)
}

// WrapErrInvalidTeamID wraps an invalid team ID error with context
func WrapErrInvalidTeamID(teamID string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidTeamID, teamID, err)
}

// WrapErrInvalidClubsDirPath wraps an invalid clubs directory path error with context
func WrapErrInvalidClubsDirPath(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidClubsDirPath, err)
}

// WrapErrInvalidNormalDirPath wraps an invalid normal directory path error with context
func WrapErrInvalidNormalDirPath(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidNormalDirPath, err)
}

// WrapErrInvalidFilePathForTeamID wraps an invalid file path for team ID error with context
func WrapErrInvalidFilePathForTeamID(teamID string, err error) error {
	return fmt.Errorf("%w for team ID %s: %v", ErrInvalidFilePath, teamID, err)
}

// WrapErrHTTPStatus wraps an HTTP status error with context
func WrapErrHTTPStatus(statusCode int, status string) error {
	return fmt.Errorf("%w %d: %s", ErrHTTPStatus, statusCode, status)
}

// WrapErrHandlerPanic wraps a handler panic error with context
func WrapErrHandlerPanic(panicValue interface{}) error {
	return fmt.Errorf("%w: %v", ErrHandlerPanic, panicValue)
}

// WrapErrPanicRecovered wraps a panic recovered error with context
func WrapErrPanicRecovered(err error) error {
	return fmt.Errorf("%w: %v", ErrPanicRecovered, err)
}

// WrapErrWorkerPanic wraps a worker panic error with context.
func WrapErrWorkerPanic(workerID int, panicValue interface{}) error {
	return fmt.Errorf("%w: worker %d panicked: %v", ErrHandlerPanic, workerID, panicValue)
}

// WrapErrInvalidFilename wraps an invalid filename error with context.
func WrapErrInvalidFilename(err error) error {
	return fmt.Errorf("%w: %v", ErrFilenameInvalidChars, err)
}

// WrapErrFailedToGetAbsPath wraps a failed absolute path error with context.
func WrapErrFailedToGetAbsPath(err error) error {
	return fmt.Errorf("%w: %v", ErrPathEscapesBase, err)
}

// WrapErrDatasetNotFound wraps a dataset not found error with context.
func WrapErrDatasetNotFound(datasetID string) error {
	return fmt.Errorf("%w: %s", ErrDatasetNotFound, datasetID)
}

// WrapErrJSONMarshalPanic wraps a JSON marshal panic error with context
func WrapErrJSONMarshalPanic(panicValue interface{}) error {
	return fmt.Errorf("%w: %v", ErrJSONMarshalPanic, panicValue)
}

// WrapErrInvalidDatasetID wraps an invalid dataset ID error with context
func WrapErrInvalidDatasetID(datasetID string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidDatasetID, datasetID, err)
}

// WrapErrInvalidFilePathForDataset wraps an invalid file path for dataset error with context
func WrapErrInvalidFilePathForDataset(datasetID string, err error) error {
	return fmt.Errorf("%w for dataset %s: %v", ErrInvalidFilePath, datasetID, err)
}

// WrapErrPlayerNoName wraps a player no name error with context
func WrapErrPlayerNoName(index int) error {
	return fmt.Errorf("%w at index %d", ErrPlayerNoName, index)
}

// WrapErrValidationFailed wraps a validation failed error with context
func WrapErrValidationFailed(errors interface{}) error {
	return fmt.Errorf("%w: %v", ErrValidationFailed, errors)
}

// WrapErrAttributeAndRoleError wraps attribute and role errors with context
func WrapErrAttributeAndRoleError(attrErr, roleErr error) error {
	return fmt.Errorf("attribute error: %w, role error: %w", attrErr, roleErr)
}

// WrapErrTokenizingHTML wraps a tokenizing HTML error with context
func WrapErrTokenizingHTML(err error) error {
	return fmt.Errorf("%w: %v", ErrTokenizingHTML, err)
}

// WrapErrFailedToParseAppearances wraps a failed to parse appearances error with context
func WrapErrFailedToParseAppearances() error {
	return fmt.Errorf("%w", ErrFailedToParseAppearances)
}

// WrapErrInvalidAppearancesFormat wraps an invalid appearances format error with context
func WrapErrInvalidAppearancesFormat() error {
	return fmt.Errorf("%w", ErrInvalidAppearancesFormat)
}

// WrapErrSkippedRowNameMissing wraps a skipped row name missing error with context.
func WrapErrSkippedRowNameMissing(cells string) error {
	return fmt.Errorf("%w. First few cells: %s", ErrSkippedRowNameMissing, cells)
}

// WrapErrIDTooLong wraps an ID too long error with context.
func WrapErrIDTooLong(maxLength int) error {
	return fmt.Errorf("%w (max %d characters)", ErrIDTooLong, maxLength)
}

// CreateValidationError creates a new validation error
func CreateValidationError(message string) *AppError {
	return &AppError{
		Code:       CodeValidationFailed,
		Message:    message,
		HTTPStatus: 400,
	}
}

// CreateNotFoundError creates a new not found error
func CreateNotFoundError(message string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		HTTPStatus: 404,
	}
}

// CreateFileTooLargeError creates a new file too large error
func CreateFileTooLargeError(maxSize int64) *AppError {
	return &AppError{
		Code:       CodeFileTooLarge,
		Message:    fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", maxSize),
		HTTPStatus: 413,
	}
}

// CreateUnsupportedFormatError creates a new unsupported format error
func CreateUnsupportedFormatError(format string, supported []string) *AppError {
	return &AppError{
		Code:       CodeUnsupportedFormat,
		Message:    fmt.Sprintf("Unsupported format: %s. Supported formats: %v", format, supported),
		HTTPStatus: 400,
	}
}

// CreateProcessingError creates a new processing error
func CreateProcessingError(message string) *AppError {
	return &AppError{
		Code:       CodeProcessingFailed,
		Message:    message,
		HTTPStatus: 500,
	}
}

// CreateStorageError creates a new storage error
func CreateStorageError(message string) *AppError {
	return &AppError{
		Code:       CodeStorageError,
		Message:    message,
		HTTPStatus: 500,
	}
}

// CreateInternalError creates a new internal error
func CreateInternalError(message string) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    message,
		HTTPStatus: 500,
	}
}

// CreateBadRequestError creates a new bad request error
func CreateBadRequestError(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: 400,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// WrapErrPanicRecoveredNonError wraps a panic recovered non-error value with context
func WrapErrPanicRecoveredNonError(value interface{}) error {
	return fmt.Errorf("%w: %v", ErrPanicRecoveredNonError, value)
}
