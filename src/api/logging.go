package main

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/baggage"
)

// Logger provides enhanced structured logging with automatic trace correlation
type Logger struct {
	baseLogger *slog.Logger
}

// CreateLogger creates a new enhanced logger
func CreateLogger() *Logger {
	return &Logger{
		baseLogger: slog.Default(),
	}
}

// enrichLogContext automatically adds trace and baggage context to log attributes
func (l *Logger) enrichLogContext(ctx context.Context, attrs []any) []any {
	if !otelEnabled {
		return attrs
	}

	// Add trace information
	if traceID := GetTraceID(ctx); traceID != "" {
		attrs = append(attrs, "trace_id", traceID)
	}
	if spanID := GetSpanID(ctx); spanID != "" {
		attrs = append(attrs, "span_id", spanID)
	}

	// Add baggage values
	bag := baggage.FromContext(ctx)
	for _, member := range bag.Members() {
		attrs = append(attrs, "baggage."+member.Key(), member.Value())
	}

	return attrs
}

// InfoContext logs at info level with automatic context enrichment
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...any) {
	enrichedAttrs := l.enrichLogContext(ctx, attrs)
	l.baseLogger.InfoContext(ctx, msg, enrichedAttrs...)
}

// WarnContext logs at warn level with automatic context enrichment
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...any) {
	enrichedAttrs := l.enrichLogContext(ctx, attrs)
	l.baseLogger.WarnContext(ctx, msg, enrichedAttrs...)
}

// ErrorContext logs at error level with automatic context enrichment
func (l *Logger) ErrorContext(ctx context.Context, msg string, attrs ...any) {
	enrichedAttrs := l.enrichLogContext(ctx, attrs)
	l.baseLogger.ErrorContext(ctx, msg, enrichedAttrs...)
}

// DebugContext logs at debug level with automatic context enrichment
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...any) {
	enrichedAttrs := l.enrichLogContext(ctx, attrs)
	l.baseLogger.DebugContext(ctx, msg, enrichedAttrs...)
}

// LogOperation logs the start and end of an operation with performance metrics
func (l *Logger) LogOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	start := time.Now()

	l.InfoContext(ctx, "Operation started",
		"operation", operation,
		"start_time", start.Format(time.RFC3339),
	)

	err := fn(ctx)
	duration := time.Since(start)

	if err != nil {
		l.ErrorContext(ctx, "Operation failed",
			"operation", operation,
			"duration_ms", float64(duration.Nanoseconds())/1e6,
			"error", err.Error(),
		)
	} else {
		l.InfoContext(ctx, "Operation completed",
			"operation", operation,
			"duration_ms", float64(duration.Nanoseconds())/1e6,
		)
	}

	return err
}

// LogSlowOperation logs operations that exceed a threshold
func (l *Logger) LogSlowOperation(ctx context.Context, operation string, threshold, duration time.Duration, attrs ...any) {
	if duration <= threshold {
		return
	}

	baseAttrs := []any{
		"operation", operation,
		"duration", duration,
		"threshold", threshold,
		"slow_operation", true,
	}
	baseAttrs = append(baseAttrs, attrs...)

	l.baseLogger.With(baseAttrs...).WarnContext(ctx, "Slow operation detected", "operation", operation)
}

// Global enhanced logger instance
var enhancedLogger = CreateLogger()

// LogInfoContext logs an info message with context
func LogInfoContext(ctx context.Context, msg string, attrs ...any) {
	enhancedLogger.InfoContext(ctx, msg, attrs...)
}

// LogWarnContext logs a warning message with context
func LogWarnContext(ctx context.Context, msg string, attrs ...any) {
	enhancedLogger.WarnContext(ctx, msg, attrs...)
}

// LogErrorContext logs an error message with context
func LogErrorContext(ctx context.Context, msg string, attrs ...any) {
	enhancedLogger.ErrorContext(ctx, msg, attrs...)
}

// LogDebugContext logs a debug message with context
func LogDebugContext(ctx context.Context, msg string, attrs ...any) {
	enhancedLogger.DebugContext(ctx, msg, attrs...)
}
