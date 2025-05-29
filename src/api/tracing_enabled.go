//go:build !no_otel

package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("v2fmdash-api")

	// Enhanced metrics
	apiRequestDuration      metric.Float64Histogram
	apiRequestsTotal        metric.Int64Counter
	apiRequestsActive       metric.Int64UpDownCounter
	dbOperationDuration     metric.Float64Histogram
	dbOperationsTotal       metric.Int64Counter
	fileProcessingDuration  metric.Float64Histogram
	errorEventsTotal        metric.Int64Counter
	businessOperationsTotal metric.Int64Counter
)

// StartSpan creates a new span with standard attributes
func StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !otelEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	// Add caller information
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		fn := runtime.FuncForPC(pc)
		opts = append(opts, trace.WithAttributes(
			attribute.String("code.function", fn.Name()),
			attribute.String("code.filepath", file),
			attribute.Int("code.lineno", line),
		))
	}

	return tracer.Start(ctx, operationName, opts...)
}

// StartSpanWithAttributes creates a span with custom attributes
func StartSpanWithAttributes(ctx context.Context, operationName string, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !otelEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	opts = append(opts, trace.WithAttributes(attrs...))
	return StartSpan(ctx, operationName, opts...)
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, eventName string, attrs ...attribute.KeyValue) {
	if !otelEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// SetSpanAttributes adds attributes to the current span
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !otelEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span and sets status
func RecordError(ctx context.Context, err error, description string) {
	if !otelEnabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, description)

	// Also record error metric
	if errorEventsTotal != nil {
		errorEventsTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("error.type", fmt.Sprintf("%T", err)),
			attribute.String("error.message", err.Error()),
		))
	}

	// Log error with trace correlation
	slog.ErrorContext(ctx, "Operation failed",
		"error", err.Error(),
		"description", description,
	)
}

// RecordAPIOperation records metrics for API operations
func RecordAPIOperation(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration) {
	if !otelEnabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", endpoint),
		attribute.Int("http.status_code", statusCode),
	}

	if apiRequestDuration != nil {
		apiRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	if apiRequestsTotal != nil {
		apiRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordDBOperation records metrics for database operations
func RecordDBOperation(ctx context.Context, operation, table string, duration time.Duration, rowsAffected int) {
	if !otelEnabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.Int("db.rows_affected", rowsAffected),
	}

	if dbOperationDuration != nil {
		dbOperationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	if dbOperationsTotal != nil {
		dbOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordBusinessOperation records metrics for business operations
func RecordBusinessOperation(ctx context.Context, operation string, success bool, details map[string]interface{}) {
	if !otelEnabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("business.operation", operation),
		attribute.Bool("business.success", success),
	}

	// Add detail attributes
	for k, v := range details {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	if businessOperationsTotal != nil {
		businessOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// TraceFileProcessing wraps file processing operations with tracing
func TraceFileProcessing(ctx context.Context, filename string, fileSize int64, fn func(context.Context) error) error {
	if !otelEnabled {
		return fn(ctx)
	}

	ctx, span := StartSpanWithAttributes(ctx, "file.processing", []attribute.KeyValue{
		attribute.String("file.name", filename),
		attribute.Int64("file.size", fileSize),
	})
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if fileProcessingDuration != nil {
			fileProcessingDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
				attribute.String("file.name", filename),
			))
		}
	}()

	if err := fn(ctx); err != nil {
		RecordError(ctx, err, "File processing failed")
		return err
	}

	span.SetStatus(codes.Ok, "File processed successfully")
	return nil
}

// IncrementActiveRequests increments the active requests counter
func IncrementActiveRequests(ctx context.Context, endpoint string) {
	if !otelEnabled || apiRequestsActive == nil {
		return
	}

	apiRequestsActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String("http.route", endpoint),
	))
}

// DecrementActiveRequests decrements the active requests counter
func DecrementActiveRequests(ctx context.Context, endpoint string) {
	if !otelEnabled || apiRequestsActive == nil {
		return
	}

	apiRequestsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String("http.route", endpoint),
	))
}

// GetTraceID returns the trace ID from the current context
func GetTraceID(ctx context.Context) string {
	if !otelEnabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID from the current context
func GetSpanID(ctx context.Context) string {
	if !otelEnabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// initEnhancedMetrics initializes additional metrics instruments
func initEnhancedMetrics() {
	if !otelEnabled {
		return
	}

	meter := otel.Meter("v2fmdash-api")

	var err error

	apiRequestDuration, err = meter.Float64Histogram(
		"fm24_api_request_duration_seconds",
		metric.WithDescription("Duration of API requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("Failed to create API request duration histogram", "error", err)
	}

	apiRequestsTotal, err = meter.Int64Counter(
		"fm24_api_requests_total",
		metric.WithDescription("Total number of API requests"),
	)
	if err != nil {
		slog.Error("Failed to create API requests total counter", "error", err)
	}

	apiRequestsActive, err = meter.Int64UpDownCounter(
		"fm24_api_requests_active",
		metric.WithDescription("Number of active API requests"),
	)
	if err != nil {
		slog.Error("Failed to create active API requests gauge", "error", err)
	}

	dbOperationDuration, err = meter.Float64Histogram(
		"fm24_db_operation_duration_seconds",
		metric.WithDescription("Duration of database operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("Failed to create DB operation duration histogram", "error", err)
	}

	dbOperationsTotal, err = meter.Int64Counter(
		"fm24_db_operations_total",
		metric.WithDescription("Total number of database operations"),
	)
	if err != nil {
		slog.Error("Failed to create DB operations total counter", "error", err)
	}

	fileProcessingDuration, err = meter.Float64Histogram(
		"fm24_file_processing_duration_seconds",
		metric.WithDescription("Duration of file processing operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("Failed to create file processing duration histogram", "error", err)
	}

	errorEventsTotal, err = meter.Int64Counter(
		"fm24_error_events_total",
		metric.WithDescription("Total number of error events"),
	)
	if err != nil {
		slog.Error("Failed to create error events total counter", "error", err)
	}

	businessOperationsTotal, err = meter.Int64Counter(
		"fm24_business_operations_total",
		metric.WithDescription("Total number of business operations"),
	)
	if err != nil {
		slog.Error("Failed to create business operations total counter", "error", err)
	}

	slog.Info("Enhanced OpenTelemetry metrics initialized")
}
