package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// General OTEL meter
	meter metric.Meter

	// Upload metrics
	uploadDuration         metric.Float64Histogram
	uploadRowsPerSecond    metric.Float64Gauge
	uploadFileSize         metric.Float64Histogram
	uploadPlayersProcessed metric.Float64Gauge
	uploadMemoryUsage      metric.Float64Gauge
	uploadsTotal           metric.Int64Counter
	uploadWorkers          metric.Float64Gauge

	// Application-specific metrics
	// activeUploads   metric.Int64UpDownCounter
	// datasetCount    metric.Int64UpDownCounter
	// cacheSizeBytes  metric.Int64Gauge
	// searchLatency   metric.Float64Histogram
	// apiLatency      metric.Float64Histogram
	// dbQueryDuration metric.Float64Histogram
	// concurrentUsers metric.Int64UpDownCounter

	// Business metrics
	// playersPerTeam          metric.Float64Histogram
	// teamRatingsDistribution metric.Float64Histogram
	// userEngagementScore     metric.Float64Gauge
	// dataQualityScore        metric.Float64Gauge

	// Enhanced API/DB metrics
	apiRequestDuration      metric.Float64Histogram
	apiRequestsTotal        metric.Int64Counter
	apiRequestsActive       metric.Int64UpDownCounter
	dbOperationDuration     metric.Float64Histogram
	dbOperationsTotal       metric.Int64Counter
	fileProcessingDuration  metric.Float64Histogram
	errorEventsTotal        metric.Int64Counter
	businessOperationsTotal metric.Int64Counter

	// Enhanced business-specific metrics
	playersProcessedTotal metric.Int64Counter
	datasetsActiveGauge   metric.Int64UpDownCounter
	parsingErrorsTotal    metric.Int64Counter
	searchRequestsTotal   metric.Int64Counter
	cacheHitsTotal        metric.Int64Counter
	cacheMissesTotal      metric.Int64Counter
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// Size buckets for file sizes in bytes
var sizeBuckets = []float64{1024, 10240, 102400, 1048576, 10485760, 52428800} // 1KB to 50MB

// initMetrics initializes OpenTelemetry metrics instruments
func initMetrics() {
	if !otelEnabled {
		return
	}

	meter = otel.Meter("v2fmdash-api")

	var err error
	uploadDuration, err = meter.Float64Histogram(
		"fm24_upload_duration_seconds",
		metric.WithDescription("Time taken to process uploads"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
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

	uploadFileSize, err = meter.Float64Histogram(
		"fm24_upload_file_size_bytes",
		metric.WithDescription("Size of uploaded files in bytes"),
		metric.WithUnit("By"),
		metric.WithExplicitBucketBoundaries(sizeBuckets...),
	)
	if err != nil {
		log.Printf("Failed to create upload file size histogram: %v", err)
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

	// Additional metrics removed to reduce commented code warnings
	// These can be re-enabled later if needed: activeUploads, datasetCount, cacheSizeBytes

	// Additional latency metrics removed to reduce commented code warnings
	// These can be re-enabled later if needed: searchLatency, apiLatency, dbQueryDuration

	// Additional business metrics removed to reduce commented code warnings
	// These can be re-enabled later if needed: concurrentUsers, playersPerTeam, teamRatingsDistribution, userEngagementScore, dataQualityScore

	initEnhancedMetrics()

	log.Printf("OpenTelemetry metrics initialized successfully")
}

// recordUploadMetrics stores metrics for a completed upload if metrics are enabled.
func recordUploadMetrics(ctx context.Context, fileSizeBytes int64, totalDuration, parseDuration time.Duration,
	playersProcessed int, memoryAllocMB float64, numWorkers int) {
	if !otelEnabled {
		return
	}

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

// initEnhancedMetrics initializes additional metrics instruments
func initEnhancedMetrics() {
	if !otelEnabled {
		return
	}

	meter := otel.Meter("v2fmdash-api")

	var err error

	apiRequestDuration, err = meter.Float64Histogram(
		"fm24_api_request_duration_seconds",
		metric.WithDescription("Duration of API requests"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create API request duration histogram", "error", err)
	}

	apiRequestsTotal, err = meter.Int64Counter(
		"fm24_api_requests_total",
		metric.WithDescription("Total number of API requests"),
	)
	if err != nil {
		slog.Error("Failed to create API requests total counter", "error", err)
	}

	apiRequestsActive, err = meter.Int64UpDownCounter(
		"fm24_api_requests_active",
		metric.WithDescription("Number of active API requests"),
	)
	if err != nil {
		slog.Error("Failed to create active API requests gauge", "error", err)
	}

	dbOperationDuration, err = meter.Float64Histogram(
		"fm24_db_operation_duration_seconds",
		metric.WithDescription("Duration of database operations"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create DB operation duration histogram", "error", err)
	}

	dbOperationsTotal, err = meter.Int64Counter(
		"fm24_db_operations_total",
		metric.WithDescription("Total number of database operations"),
	)
	if err != nil {
		slog.Error("Failed to create DB operations total counter", "error", err)
	}

	fileProcessingDuration, err = meter.Float64Histogram(
		"fm24_file_processing_duration_seconds",
		metric.WithDescription("Duration of file processing operations"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create file processing duration histogram", "error", err)
	}

	errorEventsTotal, err = meter.Int64Counter(
		"fm24_error_events_total",
		metric.WithDescription("Total number of error events"),
	)
	if err != nil {
		slog.Error("Failed to create error events total counter", "error", err)
	}

	businessOperationsTotal, err = meter.Int64Counter(
		"fm24_business_operations_total",
		metric.WithDescription("Total number of business operations"),
	)
	if err != nil {
		slog.Error("Failed to create business operations total counter", "error", err)
	}

	playersProcessedTotal, err = meter.Int64Counter(
		"fm24_players_processed_total",
		metric.WithDescription("Total number of players processed"),
	)
	if err != nil {
		slog.Error("Failed to create players processed total counter", "error", err)
	}

	datasetsActiveGauge, err = meter.Int64UpDownCounter(
		"fm24_datasets_active_gauge",
		metric.WithDescription("Number of active datasets"),
	)
	if err != nil {
		slog.Error("Failed to create datasets active gauge", "error", err)
	}

	parsingErrorsTotal, err = meter.Int64Counter(
		"fm24_parsing_errors_total",
		metric.WithDescription("Total number of parsing errors"),
	)
	if err != nil {
		slog.Error("Failed to create parsing errors total counter", "error", err)
	}

	searchRequestsTotal, err = meter.Int64Counter(
		"fm24_search_requests_total",
		metric.WithDescription("Total number of search requests"),
	)
	if err != nil {
		slog.Error("Failed to create search requests total counter", "error", err)
	}

	cacheHitsTotal, err = meter.Int64Counter(
		"fm24_cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		slog.Error("Failed to create cache hits total counter", "error", err)
	}

	cacheMissesTotal, err = meter.Int64Counter(
		"fm24_cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		slog.Error("Failed to create cache misses total counter", "error", err)
	}

	slog.Info("Enhanced OpenTelemetry metrics initialized")
}

// RecordAPIOperation records metrics for API operations
func RecordAPIOperation(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration) {
	if !otelEnabled {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", endpoint),
		attribute.Int("http.status_code", statusCode),
	)

	if apiRequestDuration != nil {
		apiRequestDuration.Record(ctx, duration.Seconds(), attrs)
	}

	if apiRequestsTotal != nil {
		apiRequestsTotal.Add(ctx, 1, attrs)
	}
}

// RecordDBOperation records metrics for database operations
func RecordDBOperation(ctx context.Context, operation, table string, duration time.Duration, rowsAffected int) {
	if !otelEnabled {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.Int("db.rows_affected", rowsAffected),
	)

	if dbOperationDuration != nil {
		dbOperationDuration.Record(ctx, duration.Seconds(), attrs)
	}

	if dbOperationsTotal != nil {
		dbOperationsTotal.Add(ctx, 1, attrs)
	}
}

// RecordBusinessOperation records metrics for business operations
func RecordBusinessOperation(ctx context.Context, operation string, success bool, details map[string]interface{}) {
	if !otelEnabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("business.operation", operation),
		attribute.Bool("business.success", success),
	}

	// Add detail attributes
	for k, v := range details {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	if businessOperationsTotal != nil {
		businessOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// IncrementActiveRequests increments the active requests counter
func IncrementActiveRequests(ctx context.Context, endpoint string) {
	if !otelEnabled || apiRequestsActive == nil {
		return
	}

	apiRequestsActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String("http.route", endpoint),
	))
}

// DecrementActiveRequests decrements the active requests counter
func DecrementActiveRequests(ctx context.Context, endpoint string) {
	if !otelEnabled || apiRequestsActive == nil {
		return
	}

	apiRequestsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String("http.route", endpoint),
	))
}

// RecordPlayersProcessed records the number of players processed
func RecordPlayersProcessed(ctx context.Context, count int, operation string) {
	if !otelEnabled || playersProcessedTotal == nil {
		return
	}

	playersProcessedTotal.Add(ctx, int64(count), metric.WithAttributes(
		attribute.String("operation", operation),
	))
}

// RecordDatasetChange tracks dataset creation/deletion
func RecordDatasetChange(ctx context.Context, operation, datasetID string, delta int64) {
	if !otelEnabled || datasetsActiveGauge == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("dataset.id", datasetID),
	)

	datasetsActiveGauge.Add(ctx, delta, attrs)
}

// RecordUploadSize records the size of file uploads
func RecordUploadSize(ctx context.Context, sizeBytes int64, fileType string) {
	if !otelEnabled || uploadFileSize == nil {
		return
	}

	uploadFileSize.Record(ctx, float64(sizeBytes), metric.WithAttributes(
		attribute.String("file.type", fileType),
	))
}

// RecordParsingError records parsing errors with context
func RecordParsingError(ctx context.Context, errorType, filename string) {
	if !otelEnabled || parsingErrorsTotal == nil {
		return
	}

	parsingErrorsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error.type", errorType),
		attribute.String("file.name", filename),
	))
}

// RecordSearchRequest records search operations
func RecordSearchRequest(ctx context.Context, searchType, query string, resultsCount int) {
	if !otelEnabled || searchRequestsTotal == nil {
		return
	}

	searchRequestsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("search.type", searchType),
		attribute.String("search.query", query),
		attribute.Int("search.results.count", resultsCount),
	))
}

// RecordCacheHit records cache hit events
func RecordCacheHit(ctx context.Context, cacheType, key string) {
	if !otelEnabled || cacheHitsTotal == nil {
		return
	}

	cacheHitsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("cache.type", cacheType),
		attribute.String("cache.key", key),
	))
}

// RecordCacheMiss records cache miss events
func RecordCacheMiss(ctx context.Context, cacheType, key string) {
	if !otelEnabled || cacheMissesTotal == nil {
		return
	}

	cacheMissesTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("cache.type", cacheType),
		attribute.String("cache.key", key),
	))
}
