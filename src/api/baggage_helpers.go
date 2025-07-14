package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// BaggageKeys defines common baggage keys used across the application
var BaggageKeys = struct {
	UserID      string
	TenantID    string
	RequestID   string
	FeatureFlag string
	Environment string
	DatasetID   string
}{
	UserID:      "user.id",
	TenantID:    "tenant.id",
	RequestID:   "request.id",
	FeatureFlag: "feature.flag",
	Environment: "environment",
	DatasetID:   "dataset.id",
}

// SetBaggageValue sets a baggage value in the context
func SetBaggageValue(ctx context.Context, key, value string) context.Context {
	if !otelEnabled || value == "" {
		return ctx
	}

	bag := baggage.FromContext(ctx)
	member, err := baggage.NewMember(key, value)
	if err != nil {
		if GetMinLogLevel() <= LogLevelWarn {
			slog.WarnContext(ctx, "Failed to create baggage member",
				"key", key,
				"value", value,
				"error", err)
		}
		return ctx
	}

	newBag, err := bag.SetMember(member)
	if err != nil {
		if GetMinLogLevel() <= LogLevelWarn {
			slog.WarnContext(ctx, "Failed to set baggage member",
				"key", key,
				"error", err)
		}
		return ctx
	}

	return baggage.ContextWithBaggage(ctx, newBag)
}

// GetBaggageValue retrieves a baggage value from the context
func GetBaggageValue(ctx context.Context, key string) string {
	if !otelEnabled {
		return ""
	}

	bag := baggage.FromContext(ctx)
	member := bag.Member(key)
	return member.Value()
}

// SetUserContext sets user-related baggage values
func SetUserContext(ctx context.Context, userID, tenantID string) context.Context {
	if !otelEnabled {
		return ctx
	}

	ctx = SetBaggageValue(ctx, BaggageKeys.UserID, userID)
	ctx = SetBaggageValue(ctx, BaggageKeys.TenantID, tenantID)

	// Also add as span attributes for immediate visibility
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("user.id", userID),
			attribute.String("tenant.id", tenantID),
		)
	}

	return ctx
}

// SetDatasetContext sets dataset-related baggage
func SetDatasetContext(ctx context.Context, datasetID string) context.Context {
	if !otelEnabled {
		return ctx
	}

	ctx = SetBaggageValue(ctx, BaggageKeys.DatasetID, datasetID)

	// Also add as span attribute
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String("dataset.id", datasetID))
	}

	return ctx
}

// SetFeatureFlag sets a feature flag value in baggage
func SetFeatureFlag(ctx context.Context, flagName string, enabled bool) context.Context {
	if !otelEnabled {
		return ctx
	}

	value := "false"
	if enabled {
		value = "true"
	}

	key := "feature_flag." + flagName
	return SetBaggageValue(ctx, key, value)
}

// GetFeatureFlag retrieves a feature flag from baggage
func GetFeatureFlag(ctx context.Context, flagName string) bool {
	if !otelEnabled {
		return false
	}

	key := "feature_flag." + flagName
	value := GetBaggageValue(ctx, key)
	return value == "true"
}

// EnrichSpanWithBaggage adds all baggage members as span attributes
func EnrichSpanWithBaggage(ctx context.Context) {
	if !otelEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	bag := baggage.FromContext(ctx)
	for _, member := range bag.Members() {
		span.SetAttributes(attribute.String("baggage."+member.Key(), member.Value()))
	}
}
