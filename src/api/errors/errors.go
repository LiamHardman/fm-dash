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

	// Configuration errors
	ErrEmptyWeightsFile       = errors.New("loaded weights file is empty")
	ErrConfigInitTimeout      = errors.New("configuration initialization timed out")
	ErrInvalidOtelEndpoint    = errors.New("invalid OTEL_EXPORTER_OTLP_ENDPOINT")
	ErrInvalidS3Endpoint      = errors.New("invalid S3_ENDPOINT format")
	ErrInvalidServiceName     = errors.New("invalid SERVICE_NAME: contains unsafe characters")
	ErrServiceNameEmpty       = errors.New("service name cannot be empty")
	ErrCollectorURLEmpty      = errors.New("collector URL cannot be empty")
	ErrInvalidTraceSampleRate = errors.New("trace sample rate must be between -1.0 and 1.0")
	ErrInvalidBatchSize       = errors.New("batch size must be positive")
	ErrInvalidMaxQueueSize    = errors.New("max queue size must be positive")

	// Processing errors
	ErrEmptyString              = errors.New("empty string")
	ErrInvalidCharacter         = errors.New("invalid character")
	ErrEmptyHeaders             = errors.New("headers found but workers were not started")
	ErrNoTableHeaders           = errors.New("no table headers found in HTML file")
	ErrTokenizingHTML           = errors.New("error tokenizing HTML")
	ErrFailedToParseAppearances = errors.New("failed to parse appearances")
	ErrInvalidAppearancesFormat = errors.New("invalid appearances format")
	ErrCannotProcessRow         = errors.New("cannot process row: headers are empty")
	ErrSkippedRowNameMissing    = errors.New("skipped row: 'Name' field is missing or empty, but other data present")
	ErrSkippedRowEmpty          = errors.New("skipped row: 'Name' field missing and row appears empty or is likely a non-player row")

	// Security errors
	ErrFilenameEmpty               = errors.New("filename cannot be empty")
	ErrFilenameTooLong             = errors.New("filename too long (max 255 characters)")
	ErrFilenamePathTraversal       = errors.New("filename contains path traversal sequence")
	ErrFilenameDirectorySeparators = errors.New("filename contains directory separators")
	ErrFilenameNullBytes           = errors.New("filename contains null bytes")
	ErrFilenameInvalidChars        = errors.New("filename contains invalid characters")
	ErrIDEmpty                     = errors.New("ID cannot be empty")
	ErrIDTooLong                   = errors.New("ID too long")
	ErrIDPathTraversal             = errors.New("ID contains path traversal sequence")
	ErrIDDirectorySeparators       = errors.New("ID contains directory separators")
	ErrIDNullBytes                 = errors.New("ID contains null bytes")
	ErrIDInvalidChars              = errors.New("ID contains invalid characters")
	ErrPathEscapesBase             = errors.New("path escapes base directory")

	// Service errors
	ErrDatasetIDEmpty         = errors.New("dataset ID cannot be empty")
	ErrDatasetNotFound        = errors.New("dataset not found")
	ErrNoPlayersToStore       = errors.New("no players to store")
	ErrPlayerNoName           = errors.New("player has no name")
	ErrNoPlayersToProcess     = errors.New("no players to process")
	ErrNoPlayersForValidation = errors.New("no players provided for validation")
	ErrValidationFailed       = errors.New("validation failed")
	ErrFileContentEmpty       = errors.New("file content is empty")
	ErrUnsupportedFileFormat  = errors.New("unsupported file format, expected .html file")
	ErrInvalidHTML            = errors.New("file does not appear to be valid HTML")

	// Storage errors
	ErrS3ClientNotAvailable = errors.New("S3 client not available")
	ErrJSONMarshalPanic     = errors.New("panic during JSON marshal")

	// HTTP errors
	ErrHTTPStatus     = errors.New("HTTP error")
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

// Error wrapper functions that maintain context while using static errors
func WrapErrInvalidCacheDataFormat(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidCacheDataFormat, err)
}

func WrapErrConfigInitTimeout(timeout interface{}) error {
	return fmt.Errorf("%w after %v", ErrConfigInitTimeout, timeout)
}

func WrapErrInvalidOtelEndpoint(endpoint string) error {
	return fmt.Errorf("%w: %s (must include port)", ErrInvalidOtelEndpoint, endpoint)
}

func WrapErrInvalidS3Endpoint(endpoint string) error {
	return fmt.Errorf("%w: %s (should include port or be full URL)", ErrInvalidS3Endpoint, endpoint)
}

func WrapErrInvalidTraceSampleRate(rate float64) error {
	return fmt.Errorf("%w, got %f", ErrInvalidTraceSampleRate, rate)
}

func WrapErrInvalidBatchSize(size int) error {
	return fmt.Errorf("%w, got %d", ErrInvalidBatchSize, size)
}

func WrapErrInvalidMaxQueueSize(size int) error {
	return fmt.Errorf("%w, got %d", ErrInvalidMaxQueueSize, size)
}

func WrapErrInvalidUID(uid string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidUID, uid, err)
}

func WrapErrInvalidFilePath(uid string, err error) error {
	return fmt.Errorf("%w for UID %s: %v", ErrInvalidFilePath, uid, err)
}

func WrapErrInvalidTeamID(teamID string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidTeamID, teamID, err)
}

func WrapErrInvalidClubsDirPath(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidClubsDirPath, err)
}

func WrapErrInvalidNormalDirPath(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidNormalDirPath, err)
}

func WrapErrInvalidFilePathForTeamID(teamID string, err error) error {
	return fmt.Errorf("%w for team ID %s: %v", ErrInvalidFilePath, teamID, err)
}

func WrapErrHTTPStatus(statusCode int, status string) error {
	return fmt.Errorf("%w %d: %s", ErrHTTPStatus, statusCode, status)
}

func WrapErrHandlerPanic(panic interface{}) error {
	return fmt.Errorf("%w: %v", ErrHandlerPanic, panic)
}

func WrapErrPanicRecovered(err error) error {
	return fmt.Errorf("%w: %v", ErrPanicRecovered, err)
}

func WrapErrWorkerPanic(workerID int, panic interface{}) error {
	return fmt.Errorf("worker %d panicked: %v", workerID, panic)
}

func WrapErrSkippedRowNameMissing(cells []string) error {
	return fmt.Errorf("%w. First few cells: %s", ErrSkippedRowNameMissing, cells)
}

func WrapErrIDTooLong(maxLength int) error {
	return fmt.Errorf("%w (max %d characters)", ErrIDTooLong, maxLength)
}

func WrapErrInvalidFilename(err error) error {
	return fmt.Errorf("invalid filename: %v", err)
}

func WrapErrFailedToGetAbsPath(err error) error {
	return fmt.Errorf("failed to get absolute path: %v", err)
}

func WrapErrDatasetNotFound(datasetID string) error {
	return fmt.Errorf("dataset %s not found", datasetID)
}

func WrapErrJSONMarshalPanic(panic interface{}) error {
	return fmt.Errorf("%w: %v", ErrJSONMarshalPanic, panic)
}

func WrapErrInvalidDatasetID(datasetID string, err error) error {
	return fmt.Errorf("%w: %s, error: %v", ErrInvalidDatasetID, datasetID, err)
}

func WrapErrInvalidFilePathForDataset(datasetID string, err error) error {
	return fmt.Errorf("%w for dataset %s: %v", ErrInvalidFilePath, datasetID, err)
}

func WrapErrPlayerNoName(index int) error {
	return fmt.Errorf("%w at index %d", ErrPlayerNoName, index)
}

func WrapErrValidationFailed(errors interface{}) error {
	return fmt.Errorf("%w: %v", ErrValidationFailed, errors)
}

func WrapErrAttributeAndRoleError(attrErr, roleErr error) error {
	return fmt.Errorf("attribute error: %v, role error: %v", attrErr, roleErr)
}

func WrapErrTokenizingHTML(err error) error {
	return fmt.Errorf("%w: %v", ErrTokenizingHTML, err)
}

func WrapErrFailedToParseAppearances() error {
	return fmt.Errorf("%w", ErrFailedToParseAppearances)
}

func WrapErrInvalidAppearancesFormat() error {
	return fmt.Errorf("%w", ErrInvalidAppearancesFormat)
}

// Validation errors
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       CodeValidationFailed,
		Message:    message,
		HTTPStatus: 400,
	}
}

// Not found errors
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		HTTPStatus: 404,
	}
}

// File upload errors
func NewFileTooLargeError(maxSize int64) *AppError {
	return &AppError{
		Code:       CodeFileTooLarge,
		Message:    fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", maxSize),
		HTTPStatus: 413,
	}
}

// Unsupported format errors
func NewUnsupportedFormatError(format string, supported []string) *AppError {
	return &AppError{
		Code:       CodeUnsupportedFormat,
		Message:    fmt.Sprintf("Unsupported file format: %s. Supported formats: %v", format, supported),
		HTTPStatus: 415,
	}
}

// Processing errors
func NewProcessingError(message string) *AppError {
	return &AppError{
		Code:       CodeProcessingFailed,
		Message:    message,
		HTTPStatus: 422,
	}
}

// Storage errors
func NewStorageError(message string) *AppError {
	return &AppError{
		Code:       CodeStorageError,
		Message:    message,
		HTTPStatus: 500,
	}
}

// Internal server errors
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    message,
		HTTPStatus: 500,
	}
}

// Bad request errors
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: 400,
	}
}

// Helper function to check if error is of specific type
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}
