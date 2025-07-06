package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("v2fmdash-api")

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

// RecordError records an error on the current span and sets status with enhanced context
func RecordError(ctx context.Context, err error, description string, opts ...ErrorOption) {
	if !otelEnabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, description)

	// Apply error options for enhanced context
	errorInfo := &ErrorInfo{
		ErrorType:    fmt.Sprintf("%T", err),
		ErrorMessage: err.Error(),
		Description:  description,
	}

	for _, opt := range opts {
		opt(errorInfo)
	}

	// Enhanced error attributes
	attrs := []attribute.KeyValue{
		attribute.String("error.type", errorInfo.ErrorType),
		attribute.String("error.message", errorInfo.ErrorMessage),
		attribute.String("error.description", errorInfo.Description),
		attribute.String("error.stacktrace", string(debug.Stack())),
	}

	if errorInfo.ErrorCode != "" {
		attrs = append(attrs, attribute.String("error.code", errorInfo.ErrorCode))
	}
	if errorInfo.ErrorCategory != "" {
		attrs = append(attrs, attribute.String("error.category", errorInfo.ErrorCategory))
	}
	if errorInfo.UserID != "" {
		attrs = append(attrs, attribute.String("error.user_id", errorInfo.UserID))
	}
	if errorInfo.RequestID != "" {
		attrs = append(attrs, attribute.String("error.request_id", errorInfo.RequestID))
	}
	if errorInfo.Severity != "" {
		attrs = append(attrs, attribute.String("error.severity", errorInfo.Severity))
	}

	// Record error metric with enhanced attributes
	if errorEventsTotal != nil {
		errorEventsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	// Log error with trace correlation and enhanced context
	logAttrs := []any{
		"error", err.Error(),
		"description", description,
		"error_type", errorInfo.ErrorType,
	}

	if errorInfo.ErrorCode != "" {
		logAttrs = append(logAttrs, "error_code", errorInfo.ErrorCode)
	}
	if errorInfo.ErrorCategory != "" {
		logAttrs = append(logAttrs, "error_category", errorInfo.ErrorCategory)
	}
	if errorInfo.Severity != "" {
		logAttrs = append(logAttrs, "error_severity", errorInfo.Severity)
	}

	slog.ErrorContext(ctx, "Operation failed", logAttrs...)
}

// ErrorInfo holds enhanced error context
type ErrorInfo struct {
	ErrorType     string
	ErrorMessage  string
	Description   string
	ErrorCode     string
	ErrorCategory string
	UserID        string
	RequestID     string
	Severity      string
}

// ErrorOption allows customizing error recording
type ErrorOption func(*ErrorInfo)

// WithErrorCode adds an error code
func WithErrorCode(code string) ErrorOption {
	return func(info *ErrorInfo) {
		info.ErrorCode = code
	}
}

// WithErrorCategory categorizes the error (e.g., "validation", "network", "database")
func WithErrorCategory(category string) ErrorOption {
	return func(info *ErrorInfo) {
		info.ErrorCategory = category
	}
}

// WithUserContext adds user context to error
func WithUserContext(userID string) ErrorOption {
	return func(info *ErrorInfo) {
		info.UserID = userID
	}
}

// WithSeverity sets error severity (e.g., "low", "medium", "high", "critical")
func WithSeverity(severity string) ErrorOption {
	return func(info *ErrorInfo) {
		info.Severity = severity
	}
}

// WithRequestID adds request context
func WithRequestID(requestID string) ErrorOption {
	return func(info *ErrorInfo) {
		info.RequestID = requestID
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

// PerformanceMonitor provides comprehensive performance tracking
type PerformanceMonitor struct {
	operationName string
	startTime     time.Time
	ctx           context.Context
	span          trace.Span
	attributes    []attribute.KeyValue
}

// StartPerformanceMonitor begins tracking performance for an operation
func StartPerformanceMonitor(ctx context.Context, operationName string, attrs ...attribute.KeyValue) *PerformanceMonitor {
	if !otelEnabled {
		return &PerformanceMonitor{ctx: ctx}
	}

	ctx, span := StartSpanWithAttributes(ctx, operationName, attrs)

	return &PerformanceMonitor{
		operationName: operationName,
		startTime:     time.Now(),
		ctx:           ctx,
		span:          span,
		attributes:    attrs,
	}
}

// RecordCheckpoint records an intermediate checkpoint with timing
func (pm *PerformanceMonitor) RecordCheckpoint(name string, attrs ...attribute.KeyValue) {
	if !otelEnabled || pm.span == nil {
		return
	}

	elapsed := time.Since(pm.startTime)
	checkpointAttrs := append([]attribute.KeyValue{
		attribute.Float64("checkpoint.duration_ms", float64(elapsed.Nanoseconds())/1e6),
		attribute.String("checkpoint.name", name),
	}, attrs...)

	pm.span.AddEvent("performance.checkpoint", trace.WithAttributes(checkpointAttrs...))
}

// RecordMetric records a custom metric for this operation
func (pm *PerformanceMonitor) RecordMetric(name string, value float64, unit string, attrs ...attribute.KeyValue) {
	if !otelEnabled {
		return
	}

	pm.attributes = append(pm.attributes,
		append(attrs,
			attribute.String("metric.unit", unit),
			attribute.String("metric.name", name),
			attribute.Float64("metric.value", value))...)

	// Record as span event for immediate visibility
	pm.span.AddEvent("performance.metric", trace.WithAttributes(pm.attributes...))

	// Log the metric for debugging
	log.Printf("📊 Performance Metric: %s = %.2f %s", name, value, unit)
}

// Finish completes the performance monitoring and records final metrics
func (pm *PerformanceMonitor) Finish(success bool, itemsProcessed int, bytesProcessed int64) {
	if !otelEnabled || pm.span == nil {
		return
	}

	duration := time.Since(pm.startTime)

	// Add final performance attributes
	finalAttrs := []attribute.KeyValue{
		attribute.Bool("operation.success", success),
		attribute.Float64("operation.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Int("operation.items_processed", itemsProcessed),
		attribute.Int64("operation.bytes_processed", bytesProcessed),
	}

	if itemsProcessed > 0 && duration.Seconds() > 0 {
		throughput := float64(itemsProcessed) / duration.Seconds()
		finalAttrs = append(finalAttrs, attribute.Float64("operation.throughput_per_sec", throughput))
	}

	if bytesProcessed > 0 && duration.Seconds() > 0 {
		byteThroughput := float64(bytesProcessed) / duration.Seconds()
		finalAttrs = append(finalAttrs, attribute.Float64("operation.bytes_per_sec", byteThroughput))
	}

	pm.span.SetAttributes(finalAttrs...)

	if success {
		pm.span.SetStatus(codes.Ok, "Operation completed successfully")
	} else {
		pm.span.SetStatus(codes.Error, "Operation failed")
	}

	pm.span.End()

	// Record operation metrics
	if success {
		RecordBusinessOperation(pm.ctx, pm.operationName, true, map[string]interface{}{
			"duration_ms":     float64(duration.Nanoseconds()) / 1e6,
			"items_processed": itemsProcessed,
			"bytes_processed": bytesProcessed,
		})
	}
}

// TraceSlowQuery identifies and traces slow database/processing operations
func TraceSlowQuery(ctx context.Context, operation string, threshold time.Duration, fn func(context.Context) error) error {
	start := time.Now()

	ctx, span := StartSpan(ctx, "slow_query."+operation)
	defer span.End()

	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Float64("query.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.Bool("query.is_slow", duration > threshold),
		attribute.Float64("query.threshold_ms", float64(threshold.Nanoseconds())/1e6),
	)

	if duration > threshold {
		span.AddEvent("slow_query_detected", trace.WithAttributes(
			attribute.String("query.operation", operation),
			attribute.Float64("query.duration_ms", float64(duration.Nanoseconds())/1e6),
		))

		// Log slow query warning
		slog.WarnContext(ctx, "Slow query detected",
			"operation", operation,
			"duration_ms", float64(duration.Nanoseconds())/1e6,
			"threshold_ms", float64(threshold.Nanoseconds())/1e6,
		)
	}

	if err != nil {
		RecordError(ctx, err, "Query operation failed",
			WithErrorCategory("database"),
			WithSeverity(func() string {
				if duration > threshold*2 {
					return "high"
				}
				return "medium"
			}()),
		)
	}

	return err
}
