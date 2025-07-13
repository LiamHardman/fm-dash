package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	apperrors "api/errors"

	"github.com/google/uuid"
	"github.com/klauspost/compress/gzip"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// RequestContextMiddleware adds essential request context to the root span
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		// Add standard HTTP attributes
		span.SetAttributes(
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.URLFull(r.URL.String()),
			semconv.HTTPRequestBodySize(int(r.ContentLength)),
			attribute.String("http.host", r.Host),
			attribute.String("http.scheme", r.URL.Scheme),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("net.peer.ip", r.RemoteAddr),
		)

		// Add a unique request ID for easier tracking
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		span.SetAttributes(attribute.String("http.request_id", requestID))

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// RequestTimeoutMiddleware wraps handlers with configurable timeout
func RequestTimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a span for timeout middleware
			ctx, span := StartSpan(r.Context(), "http.middleware.timeout",
				trace.WithAttributes(
					attribute.Float64("timeout_seconds", timeout.Seconds()),
				))
			defer span.End()

			// Create context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Create a channel to signal when handler is done
			done := make(chan struct{})
			var handlerPanic interface{}

			// Run handler in goroutine
			go func() {
				defer func() {
					if p := recover(); p != nil {
						handlerPanic = p
					}
					close(done)
				}()
				next.ServeHTTP(w, r.WithContext(timeoutCtx))
			}()

			// Wait for handler completion or timeout
			select {
			case <-done:
				// Check if handler panicked
				if handlerPanic != nil {
					span.RecordError(apperrors.WrapErrHandlerPanic(handlerPanic))
					span.SetStatus(codes.Error, "Handler panicked")
					panic(handlerPanic) // Re-panic to let other middleware handle it
				}
				span.SetStatus(codes.Ok, "Request completed within timeout")
				return
			case <-timeoutCtx.Done():
				// Request timed out
				if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
					span.SetStatus(codes.Error, "Request timeout")
					span.AddEvent("request.timeout", trace.WithAttributes(
						attribute.Float64("timeout_seconds", timeout.Seconds()),
					))

					if GetMinLogLevel() <= LogLevelWarn {
						slog.WarnContext(ctx, "Request timeout",
							"timeout_seconds", timeout.Seconds(),
							"method", r.Method,
							"path", r.URL.Path,
							"trace_id", GetTraceID(ctx),
						)
					}

					http.Error(w, fmt.Sprintf("Request timeout after %v", timeout), http.StatusRequestTimeout)
				} else {
					// Context was cancelled for other reasons
					span.SetStatus(codes.Error, "Request cancelled")
					span.AddEvent("request.cancelled")
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

// LoggingMiddleware logs requests with timing information and OpenTelemetry integration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a span for this middleware
		ctx, span := StartSpan(r.Context(), "http.middleware.logging",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
				attribute.Int64("http.request.content_length", r.ContentLength),
			))
		defer span.End()

		// Add request size if available
		if r.ContentLength > 0 {
			span.SetAttributes(attribute.Int64("http.request.size", r.ContentLength))
		}

		// Wrap the response writer to capture status code and response size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}

		// Update request context with the span context
		r = r.WithContext(ctx)

		// Track active requests
		IncrementActiveRequests(ctx, r.URL.Path)
		defer DecrementActiveRequests(ctx, r.URL.Path)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Set span attributes with response details
		span.SetAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
			attribute.Int64("http.response.size", int64(wrapped.responseSize)),
		)

		// Set span status based on HTTP status code
		if wrapped.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
		} else {
			span.SetStatus(codes.Ok, "Request completed successfully")
		}

		// Add span event for request completion
		span.AddEvent("request.completed", trace.WithAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
			attribute.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		))

		// Enhanced structured logging with trace correlation
		// Only log non-200 responses or when explicitly configured to log all requests
		shouldLog := wrapped.statusCode != http.StatusOK || logAllRequests

		if shouldLog && GetMinLogLevel() <= LogLevelDebug {
			slog.DebugContext(ctx, "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", float64(duration.Nanoseconds())/1e6,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
				"request_size", r.ContentLength,
				"response_size", wrapped.responseSize,
				"trace_id", GetTraceID(ctx),
				"span_id", GetSpanID(ctx),
			)
		}

		// Record API operation metrics
		RecordAPIOperation(ctx, r.Method, r.URL.Path, wrapped.statusCode, duration)

		// Legacy log for backward compatibility (only for non-200 responses)
		if wrapped.statusCode != http.StatusOK {
			log.Printf("%s %s %d %v", r.Method, sanitizeForLogging(r.URL.Path), wrapped.statusCode, duration)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.responseSize += size
	return size, err
}

// Size returns the current response size
func (rw *responseWriter) Size() int {
	return rw.responseSize
}

// Status returns the status code
func (rw *responseWriter) Status() int {
	return rw.statusCode
}

// PanicRecoveryMiddleware recovers from panics and records them in OpenTelemetry
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := StartSpan(r.Context(), "http.middleware.panic_recovery")
		defer span.End()

		defer func() {
			if err := recover(); err != nil {
				var panicErr error
				// Convert interface{} to error without using fmt.Errorf or errors.New
				panicErr = apperrors.WrapErrPanicRecoveredNonError(err)
				RecordError(ctx, panicErr, "Panic recovered in middleware")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		// Update request context
		r = r.WithContext(ctx)
		span.SetStatus(codes.Ok, "Handler executed without panic")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers with tracing
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := StartSpan(r.Context(), "http.middleware.cors")
		defer span.End()

		origin := r.Header.Get("Origin")
		span.SetAttributes(
			attribute.String("cors.origin", origin),
			attribute.String("cors.method", r.Method),
		)

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			span.AddEvent("preflight.request")
			span.SetAttributes(attribute.Bool("cors.preflight", true))
			w.WriteHeader(http.StatusOK)
			return
		}

		span.SetAttributes(attribute.Bool("cors.preflight", false))
		span.SetStatus(codes.Ok, "CORS headers set")
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// CompressionMiddleware provides gzip compression for HTTP responses
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create a span for compression middleware
		ctx, span := StartSpan(r.Context(), "http.middleware.compression")
		defer span.End()

		// Wrap the response writer with gzip compression
		gw := gzip.NewWriter(w)
		defer func() {
			if err := gw.Close(); err != nil {
				span.RecordError(err)
				if GetMinLogLevel() <= LogLevelWarn {
					slog.WarnContext(ctx, "Failed to close gzip writer", "error", err)
				}
			}
		}()

		// Set compression headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create a custom response writer that writes to gzip writer
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gw,
		}

		// Add compression metrics
		span.SetAttributes(
			attribute.String("compression.algorithm", "gzip"),
			attribute.Bool("compression.enabled", true),
		)

		next.ServeHTTP(gzipWriter, r.WithContext(ctx))
	})
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gzw *gzipResponseWriter) Write(data []byte) (int, error) {
	return gzw.Writer.Write(data)
}

func (gzw *gzipResponseWriter) WriteHeader(statusCode int) {
	// Remove Content-Length header as it will be incorrect after compression
	gzw.ResponseWriter.Header().Del("Content-Length")
	gzw.ResponseWriter.WriteHeader(statusCode)
}
