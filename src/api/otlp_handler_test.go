package main

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

type recordingLogExporter struct {
	records []sdklog.Record
}

func (e *recordingLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	for i := range records {
		e.records = append(e.records, records[i].Clone())
	}
	return nil
}

func (*recordingLogExporter) Shutdown(context.Context) error   { return nil }
func (*recordingLogExporter) ForceFlush(context.Context) error { return nil }

func TestOTLPHandlerPreservesLogContent(t *testing.T) {
	previousEnabled := otelEnabled
	t.Cleanup(func() { otelEnabled = previousEnabled })

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{2},
	})

	for _, tc := range []struct {
		name    string
		enabled bool
		traced  bool
	}{
		{name: "enabled with trace", enabled: true, traced: true},
		{name: "enabled without trace", enabled: true},
		{name: "disabled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			otelEnabled = tc.enabled
			exporter := &recordingLogExporter{}
			provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			var console bytes.Buffer
			handler := CreateOTLPHandler(provider)
			handler.fallbackHandler = slog.NewTextHandler(&console, nil)
			ctx := context.Background()
			if tc.traced {
				ctx = trace.ContextWithSpanContext(ctx, spanContext)
			}
			now := time.Now()
			record := slog.NewRecord(now, slog.LevelWarn, "Dataset processing warning", 0)
			record.AddAttrs(slog.String("dataset_id", "dataset-123"), slog.Int("player_count", 42))

			require.NoError(t, handler.Handle(ctx, record))
			require.Contains(t, console.String(), "Dataset processing warning")
			require.Contains(t, console.String(), "dataset_id=dataset-123")
			if !tc.enabled {
				require.Empty(t, exporter.records)
				return
			}

			require.Len(t, exporter.records, 1)
			exported := exporter.records[0]
			require.Equal(t, "Dataset processing warning", exported.Body().AsString())
			require.Equal(t, log.SeverityWarn, exported.Severity())
			require.Equal(t, now, exported.Timestamp())
			attrs := make(map[string]string)
			exported.WalkAttributes(func(kv attribute.KeyValue) bool {
				attrs[string(kv.Key)] = kv.Value.AsString()
				return true
			})
			require.Equal(t, "dataset-123", attrs["dataset_id"])
			require.Equal(t, "42", attrs["player_count"])
			if tc.traced {
				require.Equal(t, spanContext.TraceID().String(), attrs["trace_id"])
				require.Equal(t, spanContext.SpanID().String(), attrs["span_id"])
			} else {
				require.Equal(t, "no_valid_span_context", attrs["debug_context"])
			}
		})
	}
}
