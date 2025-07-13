// Package errors provides error handling and middleware functionality
package errors

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// sanitizeForLogging sanitizes input for safe logging
func sanitizeForLogging(input string) string {
	// Remove or escape newlines and carriage returns
	sanitized := strings.ReplaceAll(input, "\n", "\\n")
	sanitized = strings.ReplaceAll(sanitized, "\r", "\\r")
	sanitized = strings.ReplaceAll(sanitized, "\t", "\\t")

	// Truncate if too long
	if len(sanitized) > 200 {
		sanitized = sanitized[:200] + "..."
	}

	return sanitized
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"requestId,omitempty"`
}

// ErrorDetail contains the error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ResponseWriter wraps http.ResponseWriter to capture status codes
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.Size += size
	return size, err
}

// ErrorHandlerMiddleware provides centralized error handling
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer
		wrapped := &ResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		// Add panic recovery
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				// Send internal server error
				SendErrorResponse(wrapped, r, NewInternalError(
					"An unexpected error occurred",
				))
			}
		}()

		// Execute the handler
		next.ServeHTTP(wrapped, r)

		// Log the request with sanitized URL path
		log.Printf("%s %s %d %d bytes",
			r.Method,
			sanitizeForLogging(r.URL.Path),
			wrapped.StatusCode,
			wrapped.Size,
		)
	})
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// Check if it's an AppError, otherwise create a generic one
	if !IsAppError(err) {
		err = NewInternalError("An error occurred")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		appErr = NewInternalError("An error occurred")
	}

	// Get request ID from context if available
	requestID := r.Header.Get("X-Request-ID")

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
	}

	// Set content type and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	// Encode and send response
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("Failed to encode error response: %v", encodeErr)
		// Fallback to plain text
		http.Error(w, appErr.Message, appErr.HTTPStatus)
	}

	// Log error details (but not sensitive info)
	if appErr.HTTPStatus >= 500 {
		log.Printf("Internal error [%s]: %s", appErr.Code, appErr.Error())
	} else {
		log.Printf("Client error [%s]: %s", appErr.Code, appErr.Message)
	}
}

// ValidateRequest performs common request validation
func ValidateRequest(r *http.Request, allowedMethods []string) error {
	// Check method
	methodAllowed := false
	for _, method := range allowedMethods {
		if r.Method == method {
			methodAllowed = true
			break
		}
	}

	if !methodAllowed {
		return NewBadRequestError(
			"Method not allowed. Allowed methods: " + joinStrings(allowedMethods, ", "),
		)
	}

	// Check content type for POST/PUT requests
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			return NewBadRequestError(
				"Content-Type header required. Specify appropriate Content-Type for request body",
			)
		}
	}

	return nil
}

// Helper function to join strings (since strings.Join isn't imported)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
