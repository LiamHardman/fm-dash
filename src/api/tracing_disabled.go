//go:build no_otel

package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan is a no-op when OTEL is disabled
func StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

// StartSpanWithAttributes is a no-op when OTEL is disabled
func StartSpanWithAttributes(ctx context.Context, operationName string, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

// AddSpanEvent is a no-op when OTEL is disabled
func AddSpanEvent(ctx context.Context, eventName string, attrs ...attribute.KeyValue) {
	// No-op
}

// SetSpanAttributes is a no-op when OTEL is disabled
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	// No-op
}

// RecordError is a no-op when OTEL is disabled
func RecordError(ctx context.Context, err error, description string) {
	// No-op
}

// RecordAPIOperation is a no-op when OTEL is disabled
func RecordAPIOperation(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration) {
	// No-op
}

// RecordDBOperation is a no-op when OTEL is disabled
func RecordDBOperation(ctx context.Context, operation, table string, duration time.Duration, rowsAffected int) {
	// No-op
}

// RecordBusinessOperation is a no-op when OTEL is disabled
func RecordBusinessOperation(ctx context.Context, operation string, success bool, details map[string]interface{}) {
	// No-op
}

// TraceFileProcessing is a pass-through when OTEL is disabled
func TraceFileProcessing(ctx context.Context, filename string, fileSize int64, fn func(context.Context) error) error {
	return fn(ctx)
}

// IncrementActiveRequests is a no-op when OTEL is disabled
func IncrementActiveRequests(ctx context.Context, endpoint string) {
	// No-op
}

// DecrementActiveRequests is a no-op when OTEL is disabled
func DecrementActiveRequests(ctx context.Context, endpoint string) {
	// No-op
}

// GetTraceID returns empty string when OTEL is disabled
func GetTraceID(ctx context.Context) string {
	return ""
}

// GetSpanID returns empty string when OTEL is disabled
func GetSpanID(ctx context.Context) string {
	return ""
}

// initEnhancedMetrics is a no-op when OTEL is disabled
func initEnhancedMetrics() {
	// No-op
}
