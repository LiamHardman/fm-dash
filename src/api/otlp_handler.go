package main

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// OTLPHandler implements slog.Handler to stream logs to OTLP
type OTLPHandler struct {
	logger          log.Logger
	fallbackHandler slog.Handler
}

// CreateOTLPHandler creates a new OTLP handler
func CreateOTLPHandler(loggerProvider *sdklog.LoggerProvider) *OTLPHandler {
	if !otelEnabled {
		// Return a standard text handler when OTEL is disabled
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		return &OTLPHandler{
			logger:          nil,
			fallbackHandler: handler,
		}
	}

	logger := loggerProvider.Logger("v2fmdash-api")

	// Create a fallback handler for local console output
	fallbackHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return &OTLPHandler{
		logger:          logger,
		fallbackHandler: fallbackHandler,
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *OTLPHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

// Handle processes a log record
//
//nolint:gocritic // hugeParam: required by slog.Handler interface
func (h *OTLPHandler) Handle(ctx context.Context, record slog.Record) error {
	// Send to fallback handler for local logging
	if err := h.fallbackHandler.Handle(ctx, record); err != nil {
		// Don't fail if local logging fails, but log the issue
		// Using stderr directly to avoid recursion
		_, _ = os.Stderr.WriteString("OTLP: fallback logging failed: " + err.Error() + "\n")
	}

	// If OTEL is disabled, just use fallback handler
	if !otelEnabled || h.logger == nil {
		return nil
	}

	// Convert slog level to OTEL severity
	severity := convertLevel(record.Level)

	// Create OTEL log record
	var logRecord log.Record
	logRecord.SetTimestamp(record.Time)
	logRecord.SetSeverity(severity)
	logRecord.SetBody(log.StringValue(record.Message))

	// Add trace context if available
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		logRecord.AddAttributes(
			log.String("trace_id", spanCtx.TraceID().String()),
			log.String("span_id", spanCtx.SpanID().String()),
		)
	} else {
		// Add debugging - check if we're getting context but no span
		logRecord.AddAttributes(
			log.String("debug_context", "no_valid_span_context"),
		)
	}

	// Add attributes from slog record
	record.Attrs(func(attr slog.Attr) bool {
		logRecord.AddAttributes(log.String(attr.Key, attr.Value.String()))
		return true
	})

	// Emit the log record
	h.logger.Emit(ctx, logRecord)

	return nil
}

// WithAttrs returns a new handler with the given attributes
func (h *OTLPHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OTLPHandler{
		logger:          h.logger,
		fallbackHandler: h.fallbackHandler.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the given group
func (h *OTLPHandler) WithGroup(name string) slog.Handler {
	return &OTLPHandler{
		logger:          h.logger,
		fallbackHandler: h.fallbackHandler.WithGroup(name),
	}
}

// convertLevel converts slog.Level to log.Severity
func convertLevel(level slog.Level) log.Severity {
	switch {
	case level >= slog.LevelError:
		return log.SeverityError
	case level >= slog.LevelWarn:
		return log.SeverityWarn
	case level >= slog.LevelInfo:
		return log.SeverityInfo
	default:
		return log.SeverityDebug
	}
}
