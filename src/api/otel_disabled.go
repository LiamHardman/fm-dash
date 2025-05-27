//go:build no_otel

package main

import (
	"context"
	"net/http"
)

func initOTel() func(context.Context) error {
	// Return a no-op cleanup function when OTEL is disabled
	return func(ctx context.Context) error {
		return nil
	}
}

func wrapHandler(handler http.Handler, operationName string) http.Handler {
	// Return the handler as-is when OTEL is disabled
	return handler
}