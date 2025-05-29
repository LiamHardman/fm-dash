// src/api/middleware.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// RequestTimeoutMiddleware wraps handlers with configurable timeout
func RequestTimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal when handler is done
			done := make(chan struct{})

			// Run handler in goroutine
			go func() {
				defer close(done)
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			// Wait for handler completion or timeout
			select {
			case <-done:
				// Handler completed successfully
				return
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, fmt.Sprintf("Request timeout after %v", timeout), http.StatusRequestTimeout)
				} else {
					// Context was cancelled for other reasons
					http.Error(w, "Request cancelled", http.StatusInternalServerError)
				}
				return
			}
		})
	}
}

// RetryMiddleware provides automatic retry for failed requests (mainly for outbound calls)
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig provides sensible defaults for retry behavior
var DefaultRetryConfig = RetryConfig{
	MaxRetries: 3,
	BaseDelay:  100 * time.Millisecond,
	MaxDelay:   5 * time.Second,
}

// ExponentialBackoff calculates the delay for a given attempt
func (c RetryConfig) ExponentialBackoff(attempt int) time.Duration {
	// Prevent integer overflow by limiting the attempt value
	if attempt > 30 { // 2^30 is already very large
		attempt = 30
	}
	delay := c.BaseDelay * time.Duration(1<<attempt) // 2^attempt
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}
	return delay
}

// ShouldRetry determines if an HTTP status code warrants a retry
func ShouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	default:
		return false
	}
}

// LoggingMiddleware logs requests with timing information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request details
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
