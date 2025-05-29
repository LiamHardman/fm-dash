package main

import (
	"context"
	"log/slog"
)

// DemoStructuredLogging shows examples of structured logging that will be streamed to OTLP when enabled
func DemoStructuredLogging() {
	ctx := context.Background()
	// These logs will be streamed to your OTLP endpoint when OTEL_ENABLED=true
	slog.InfoContext(ctx, "Application started successfully", 
		"version", "2.0.0", 
		"environment", "production",
		"startup_time_ms", 150)
	
	slog.WarnContext(ctx, "Configuration warning", 
		"component", "storage", 
		"message", "S3 not configured, using memory storage",
		"impact", "data_loss_on_restart")
	
	slog.ErrorContext(ctx, "Operation failed", 
		"operation", "file_upload", 
		"error", "file_too_large",
		"max_size_mb", 50,
		"actual_size_mb", 75)
	
	slog.DebugContext(ctx, "Performance metrics", 
		"component", "parser",
		"players_processed", 1500,
		"processing_time_ms", 2340,
		"memory_usage_mb", 128.5)
}