//go:build !no_otel

package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// OpenTelemetry metrics
	meter                  metric.Meter
	uploadDuration         metric.Float64Histogram
	uploadRowsPerSecond    metric.Float64Gauge
	uploadFileSize         metric.Float64Gauge
	uploadPlayersProcessed metric.Float64Gauge
	uploadMemoryUsage      metric.Float64Gauge
	uploadsTotal           metric.Int64Counter
	uploadWorkers          metric.Float64Gauge
)

// initMetrics initializes OpenTelemetry metrics instruments
func initMetrics() {
	meter = otel.Meter("v2fmdash-api")

	var err error
	uploadDuration, err = meter.Float64Histogram(
		"fm24_upload_duration_seconds",
		metric.WithDescription("Time taken to process uploads"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Printf("Failed to create upload duration histogram: %v", err)
	}

	uploadRowsPerSecond, err = meter.Float64Gauge(
		"fm24_upload_rows_per_second",
		metric.WithDescription("Rows processed per second in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload rows per second gauge: %v", err)
	}

	uploadFileSize, err = meter.Float64Gauge(
		"fm24_upload_file_size_bytes",
		metric.WithDescription("Size of last uploaded file in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		log.Printf("Failed to create upload file size gauge: %v", err)
	}

	uploadPlayersProcessed, err = meter.Float64Gauge(
		"fm24_upload_players_processed",
		metric.WithDescription("Number of players processed in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload players processed gauge: %v", err)
	}

	uploadMemoryUsage, err = meter.Float64Gauge(
		"fm24_upload_memory_usage_mb",
		metric.WithDescription("Memory usage during last upload in MB"),
		metric.WithUnit("MB"),
	)
	if err != nil {
		log.Printf("Failed to create upload memory usage gauge: %v", err)
	}

	uploadsTotal, err = meter.Int64Counter(
		"fm24_uploads_total",
		metric.WithDescription("Total number of file uploads processed"),
	)
	if err != nil {
		log.Printf("Failed to create uploads total counter: %v", err)
	}

	uploadWorkers, err = meter.Float64Gauge(
		"fm24_upload_workers",
		metric.WithDescription("Number of workers used in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload workers gauge: %v", err)
	}

	log.Printf("OpenTelemetry metrics initialized successfully")
}

// recordUploadMetrics stores metrics for a completed upload if metrics are enabled.
func recordUploadMetrics(filename string, fileSizeBytes int64, totalDuration, parseDuration time.Duration,
	playersProcessed int, memoryAllocMB float64, numWorkers, numGoroutines int) {
	if !metricsEnabled {
		return
	}

	ctx := context.Background()
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(playersProcessed) / parseDuration.Seconds()
	}

	// Record to OpenTelemetry metrics
	if uploadDuration != nil {
		uploadDuration.Record(ctx, totalDuration.Seconds(), metric.WithAttributes(
			attribute.String("type", "total"),
		))
		uploadDuration.Record(ctx, parseDuration.Seconds(), metric.WithAttributes(
			attribute.String("type", "parse"),
		))
	}

	if uploadRowsPerSecond != nil {
		uploadRowsPerSecond.Record(ctx, rowsPerSecond)
	}

	if uploadFileSize != nil {
		uploadFileSize.Record(ctx, float64(fileSizeBytes))
	}

	if uploadPlayersProcessed != nil {
		uploadPlayersProcessed.Record(ctx, float64(playersProcessed))
	}

	if uploadMemoryUsage != nil {
		uploadMemoryUsage.Record(ctx, memoryAllocMB)
	}

	if uploadWorkers != nil {
		uploadWorkers.Record(ctx, float64(numWorkers))
	}

	if uploadsTotal != nil {
		uploadsTotal.Add(ctx, 1)
	}
}
