//go:build no_otel

package main

import "context"

// BaggageKeys defines common baggage keys (no-op when disabled)
var BaggageKeys = struct {
	UserID       string
	TenantID     string
	RequestID    string
	FeatureFlag  string
	Environment  string
	DatasetID    string
}{
	UserID:      "user.id",
	TenantID:    "tenant.id", 
	RequestID:   "request.id",
	FeatureFlag: "feature.flag",
	Environment: "environment",
	DatasetID:   "dataset.id",
}

// SetBaggageValue is a no-op when OTEL is disabled
func SetBaggageValue(ctx context.Context, key, value string) context.Context {
	return ctx
}

// GetBaggageValue returns empty string when OTEL is disabled
func GetBaggageValue(ctx context.Context, key string) string {
	return ""
}

// SetUserContext is a no-op when OTEL is disabled
func SetUserContext(ctx context.Context, userID, tenantID string) context.Context {
	return ctx
}

// SetDatasetContext is a no-op when OTEL is disabled
func SetDatasetContext(ctx context.Context, datasetID string) context.Context {
	return ctx
}

// SetFeatureFlag is a no-op when OTEL is disabled
func SetFeatureFlag(ctx context.Context, flagName string, enabled bool) context.Context {
	return ctx
}

// GetFeatureFlag returns false when OTEL is disabled
func GetFeatureFlag(ctx context.Context, flagName string) bool {
	return false
}

// EnrichSpanWithBaggage is a no-op when OTEL is disabled
func EnrichSpanWithBaggage(ctx context.Context) {
	// No-op
}

// LogBaggageValues is a no-op when OTEL is disabled
func LogBaggageValues(ctx context.Context, message string) {
	// No-op
} 