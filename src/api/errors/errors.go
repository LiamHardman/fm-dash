// src/api/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
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

// Validation errors
func NewValidationError(message, details string) *AppError {
	return &AppError{
		Code:       CodeValidationFailed,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Not found errors
func NewNotFoundError(resource, id string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    fmt.Sprintf("ID: %s", id),
		HTTPStatus: http.StatusNotFound,
	}
}

// File upload errors
func NewFileTooLargeError(size, maxSize int64) *AppError {
	return &AppError{
		Code:       CodeFileTooLarge,
		Message:    "File size exceeds maximum allowed",
		Details:    fmt.Sprintf("Size: %d bytes, Max: %d bytes", size, maxSize),
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}
}

func NewUnsupportedFormatError(format string, supported []string) *AppError {
	return &AppError{
		Code:       CodeUnsupportedFormat,
		Message:    "Unsupported file format",
		Details:    fmt.Sprintf("Got: %s, Supported: %v", format, supported),
		HTTPStatus: http.StatusBadRequest,
	}
}

// Processing errors
func NewProcessingError(operation string, err error) *AppError {
	return &AppError{
		Code:       CodeProcessingFailed,
		Message:    fmt.Sprintf("Failed to %s", operation),
		Details:    "Please check your file format and try again",
		HTTPStatus: http.StatusUnprocessableEntity,
		Internal:   err,
	}
}

// Storage errors
func NewStorageError(operation string, err error) *AppError {
	return &AppError{
		Code:       CodeStorageError,
		Message:    fmt.Sprintf("Storage %s failed", operation),
		Details:    "Please try again later",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   err,
	}
}

// Internal server errors
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    message,
		Details:    "An unexpected error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   err,
	}
}

// Bad request errors
func NewBadRequestError(message, details string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Helper function to check if error is of specific type
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
