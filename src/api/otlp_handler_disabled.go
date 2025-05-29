//go:build no_otel

package main

import (
	"log/slog"
	"os"
)

// OTLPHandler is a no-op when OTEL is disabled
type OTLPHandler struct {
	*slog.TextHandler
}

// NewOTLPHandler creates a standard text handler when OTEL is disabled
func NewOTLPHandler(_ interface{}) *OTLPHandler {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return &OTLPHandler{
		TextHandler: handler,
	}
}
