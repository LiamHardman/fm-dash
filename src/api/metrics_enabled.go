package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
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

	// LLM-call metrics, shared across all three LLM features via callResponsesWithRetry
	// (tracing map ticket 01).
	llmInputTokens  metric.Float64Histogram
	llmOutputTokens metric.Float64Histogram
	llmCallDuration metric.Float64Histogram

	// Who to Sign business metrics (tracing map ticket 03).
	whoToSignPositionRecommendations metric.Float64Histogram
	whoToSignPlayersRecommendedTotal metric.Float64Histogram
	whoToSignMainPickSigningScore    metric.Float64Histogram

	// Chatbot business metrics (tracing map ticket 04).
	chatbotSearchResultsCount metric.Float64Histogram
	chatbotSearchesPerTurn    metric.Float64Histogram

	// Scout Report business metrics (tracing map ticket 05).
	scoutReportSearchResultsCount   metric.Float64Histogram
	scoutReportSubjectGrade         metric.Float64Histogram
	scoutReportComparableGrade      metric.Float64Histogram
	scoutReportComparableSquadStars metric.Float64Histogram
	scoutReportComparableDivStars   metric.Float64Histogram
	scoutReportSubjectSquadStars    metric.Float64Histogram
	scoutReportSubjectDivisionStars metric.Float64Histogram
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// llmTokenBuckets sizes histogram buckets for per-call token counts (tracing map ticket 01)
// — a different scale than durationBuckets, which is seconds.
var llmTokenBuckets = []float64{100, 500, 1000, 2500, 5000, 10000, 25000, 50000}

// initMetrics initializes all metrics
func initMetrics() {
	LogInfo("OpenTelemetry metrics initialized successfully")
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

	initLLMMetrics(meter)

	slog.Info("Enhanced OpenTelemetry metrics initialized")
}

// initLLMMetrics creates every LLM-related metric instrument added by the tracing map
// (tickets 01, 03, 04, 05) — kept separate from initEnhancedMetrics' pre-existing
// generic instruments for readability.
func initLLMMetrics(meter metric.Meter) {
	var err error

	llmInputTokens, err = meter.Float64Histogram(
		"fm24_llm_input_tokens",
		metric.WithDescription("Input tokens per LLM call"),
		metric.WithUnit("token"),
		metric.WithExplicitBucketBoundaries(llmTokenBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create LLM input tokens histogram", "error", err)
	}

	llmOutputTokens, err = meter.Float64Histogram(
		"fm24_llm_output_tokens",
		metric.WithDescription("Output tokens per LLM call"),
		metric.WithUnit("token"),
		metric.WithExplicitBucketBoundaries(llmTokenBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create LLM output tokens histogram", "error", err)
	}

	llmCallDuration, err = meter.Float64Histogram(
		"fm24_llm_call_duration_seconds",
		metric.WithDescription("Duration of a single LLM Responses API call"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		slog.Error("Failed to create LLM call duration histogram", "error", err)
	}

	whoToSignPositionRecommendations, err = meter.Float64Histogram(
		"fm24_who_to_sign_position_recommendations",
		metric.WithDescription("Number of players recommended (mainPick + runnersUp) per position"),
	)
	if err != nil {
		slog.Error("Failed to create who-to-sign position recommendations histogram", "error", err)
	}

	whoToSignPlayersRecommendedTotal, err = meter.Float64Histogram(
		"fm24_who_to_sign_players_recommended_total",
		metric.WithDescription("Total number of players recommended per request, summed across positions"),
	)
	if err != nil {
		slog.Error("Failed to create who-to-sign players recommended total histogram", "error", err)
	}

	whoToSignMainPickSigningScore, err = meter.Float64Histogram(
		"fm24_who_to_sign_main_pick_signing_score",
		metric.WithDescription("Model's own 0-100 signingScore for each position's main pick"),
	)
	if err != nil {
		slog.Error("Failed to create who-to-sign main pick signing score histogram", "error", err)
	}

	chatbotSearchResultsCount, err = meter.Float64Histogram(
		"fm24_chatbot_search_results_count",
		metric.WithDescription("Number of players returned per search_players tool call"),
	)
	if err != nil {
		slog.Error("Failed to create chatbot search results count histogram", "error", err)
	}

	chatbotSearchesPerTurn, err = meter.Float64Histogram(
		"fm24_chatbot_searches_per_turn",
		metric.WithDescription("Number of search_players tool calls made to answer a single chat turn"),
	)
	if err != nil {
		slog.Error("Failed to create chatbot searches per turn histogram", "error", err)
	}

	scoutReportSearchResultsCount, err = meter.Float64Histogram(
		"fm24_scout_report_search_results_count",
		metric.WithDescription("Number of players returned per find_comparable_players tool call"),
	)
	if err != nil {
		slog.Error("Failed to create scout report search results count histogram", "error", err)
	}

	scoutReportSubjectGrade, err = meter.Float64Histogram(
		"fm24_scout_report_subject_grade",
		metric.WithDescription("Subject player's own grade, mapped D=1..A+=7"),
	)
	if err != nil {
		slog.Error("Failed to create scout report subject grade histogram", "error", err)
	}

	scoutReportComparableGrade, err = meter.Float64Histogram(
		"fm24_scout_report_comparable_grade",
		metric.WithDescription("Each comparable player's grade, mapped D=1..A+=7"),
	)
	if err != nil {
		slog.Error("Failed to create scout report comparable grade histogram", "error", err)
	}

	scoutReportComparableSquadStars, err = meter.Float64Histogram(
		"fm24_scout_report_comparable_squad_stars",
		metric.WithDescription("Each comparable player's squad-scope star rating"),
	)
	if err != nil {
		slog.Error("Failed to create scout report comparable squad stars histogram", "error", err)
	}

	scoutReportComparableDivStars, err = meter.Float64Histogram(
		"fm24_scout_report_comparable_division_stars",
		metric.WithDescription("Each comparable player's division-scope star rating"),
	)
	if err != nil {
		slog.Error("Failed to create scout report comparable division stars histogram", "error", err)
	}

	scoutReportSubjectSquadStars, err = meter.Float64Histogram(
		"fm24_scout_report_subject_squad_stars",
		metric.WithDescription("Subject player's own squad-scope star rating"),
	)
	if err != nil {
		slog.Error("Failed to create scout report subject squad stars histogram", "error", err)
	}

	scoutReportSubjectDivisionStars, err = meter.Float64Histogram(
		"fm24_scout_report_subject_division_stars",
		metric.WithDescription("Subject player's own division-scope star rating"),
	)
	if err != nil {
		slog.Error("Failed to create scout report subject division stars histogram", "error", err)
	}
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

// RecordLLMCall records per-call token/duration metrics for one Responses API round —
// shared by all three LLM features via callResponsesWithRetry (tracing map ticket 01).
func RecordLLMCall(ctx context.Context, feature string, round int, inputTokens, outputTokens int64, duration time.Duration) {
	if !otelEnabled {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("llm.feature", feature),
		attribute.Int("llm.round", round),
	)
	if llmInputTokens != nil {
		llmInputTokens.Record(ctx, float64(inputTokens), attrs)
	}
	if llmOutputTokens != nil {
		llmOutputTokens.Record(ctx, float64(outputTokens), attrs)
	}
	if llmCallDuration != nil {
		llmCallDuration.Record(ctx, duration.Seconds(), attrs)
	}
}

// RecordWhoToSignPositionOutcome records, once per requested position, how many players
// were recommended (mainPick + runnersUp) and the main pick's own signingScore
// (tracing map ticket 03).
func RecordWhoToSignPositionOutcome(ctx context.Context, recommendationsCount, signingScore int) {
	if !otelEnabled {
		return
	}
	if whoToSignPositionRecommendations != nil {
		whoToSignPositionRecommendations.Record(ctx, float64(recommendationsCount))
	}
	if whoToSignMainPickSigningScore != nil {
		whoToSignMainPickSigningScore.Record(ctx, float64(signingScore))
	}
}

// RecordWhoToSignPlayersRecommendedTotal records, once per request, the total number of
// players recommended across every position (tracing map ticket 03).
func RecordWhoToSignPlayersRecommendedTotal(ctx context.Context, total int) {
	if !otelEnabled || whoToSignPlayersRecommendedTotal == nil {
		return
	}
	whoToSignPlayersRecommendedTotal.Record(ctx, float64(total))
}

// RecordChatbotSearchResults records, per search_players tool call, how many players it
// returned (tracing map ticket 04).
func RecordChatbotSearchResults(ctx context.Context, resultCount int) {
	if !otelEnabled || chatbotSearchResultsCount == nil {
		return
	}
	chatbotSearchResultsCount.Record(ctx, float64(resultCount))
}

// RecordChatbotSearchesPerTurn records, once per chat turn, how many search_players
// calls the model made to answer it (tracing map ticket 04).
func RecordChatbotSearchesPerTurn(ctx context.Context, searchCount int) {
	if !otelEnabled || chatbotSearchesPerTurn == nil {
		return
	}
	chatbotSearchesPerTurn.Record(ctx, float64(searchCount))
}

// RecordScoutReportSearchResults records, per find_comparable_players tool call, how
// many players it returned (tracing map ticket 05).
func RecordScoutReportSearchResults(ctx context.Context, resultCount int) {
	if !otelEnabled || scoutReportSearchResultsCount == nil {
		return
	}
	scoutReportSearchResultsCount.Record(ctx, float64(resultCount))
}

// RecordScoutReportSubjectGrade records, once per report, the subject player's own
// grade (already mapped D=1..A+=7 by the caller) (tracing map ticket 05).
func RecordScoutReportSubjectGrade(ctx context.Context, gradeOrdinal int) {
	if !otelEnabled || scoutReportSubjectGrade == nil {
		return
	}
	scoutReportSubjectGrade.Record(ctx, float64(gradeOrdinal))
}

// RecordScoutReportComparable records, once per comparable player in the report, its
// grade (mapped D=1..A+=7) plus its squad/division star ratings (tracing map ticket 05).
func RecordScoutReportComparable(ctx context.Context, gradeOrdinal int, squadStars, divisionStars float64) {
	if !otelEnabled {
		return
	}
	if scoutReportComparableGrade != nil {
		scoutReportComparableGrade.Record(ctx, float64(gradeOrdinal))
	}
	if scoutReportComparableSquadStars != nil {
		scoutReportComparableSquadStars.Record(ctx, squadStars)
	}
	if scoutReportComparableDivStars != nil {
		scoutReportComparableDivStars.Record(ctx, divisionStars)
	}
}

// RecordScoutReportSubjectStars records the subject player's own squad/division star
// ratings, from the separate non-LLM scoutReportStarsHandler endpoint (tracing map
// ticket 05's scope amendment).
func RecordScoutReportSubjectStars(ctx context.Context, squadStars, divisionStars float64) {
	if !otelEnabled {
		return
	}
	if scoutReportSubjectSquadStars != nil {
		scoutReportSubjectSquadStars.Record(ctx, squadStars)
	}
	if scoutReportSubjectDivisionStars != nil {
		scoutReportSubjectDivisionStars.Record(ctx, divisionStars)
	}
}
