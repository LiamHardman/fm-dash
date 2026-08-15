package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

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
