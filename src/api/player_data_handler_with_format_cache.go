package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "api/proto"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// formatAwarePlayerDataHandler is an enhanced version of playerDataHandler that uses format-aware caching
func formatAwarePlayerDataHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.players.get")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/api/players")
	defer DecrementActiveRequests(ctx, "/api/players")

	// Record API operation metrics at the end
	defer func() {
		status := http.StatusOK
		RecordAPIOperation(ctx, r.Method, "/api/players", status, time.Since(startTime))
	}()

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	format := GetCacheFormatFromRequest(r)

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/players"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.String("cache.format", string(format)),
		attribute.String("request.id", requestID),
	)

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		logWarn(ctx, "Dataset ID missing in request path")
		SetSpanAttributes(ctx, attribute.String("error.type", "missing_dataset_id"))
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	SetSpanAttributes(ctx, attribute.String("dataset.id", datasetID))

	// Parse query parameters
	queryValues := r.URL.Query()
	filterPosition := queryValues.Get("position")
	filterRole := queryValues.Get("role")
	minAgeStr := queryValues.Get("minAge")
	maxAgeStr := queryValues.Get("maxAge")
	minTransferValueStr := queryValues.Get("minTransferValue")
	maxTransferValueStr := queryValues.Get("maxTransferValue")
	maxSalaryStr := queryValues.Get("maxSalary")
	divisionFilterStr := queryValues.Get("divisionFilter") // "all", "same", "top5"
	targetDivision := queryValues.Get("targetDivision")
	positionCompare := queryValues.Get("positionCompare") // "all", "broad", "detailed"
	pageStr := queryValues.Get("page")
	perPageStr := queryValues.Get("perPage")

	logDebug(ctx, "Processing player data request",
		"dataset_id", datasetID,
		"position_filter", filterPosition,
		"role_filter", filterRole,
		"min_age", minAgeStr,
		"max_age", maxAgeStr,
		"min_transfer_value", minTransferValueStr,
		"max_transfer_value", maxTransferValueStr,
		"max_salary", maxSalaryStr,
		"division_filter", divisionFilterStr,
		"target_division", targetDivision,
		"position_compare", positionCompare,
		"response_format", format)

	// Create filter map for cache key generation
	filters := map[string]string{
		"position":         filterPosition,
		"role":             filterRole,
		"minAge":           minAgeStr,
		"maxAge":           maxAgeStr,
		"minTransferValue": minTransferValueStr,
		"maxTransferValue": maxTransferValueStr,
		"maxSalary":        maxSalaryStr,
		"divisionFilter":   divisionFilterStr,
		"targetDivision":   targetDivision,
		"positionCompare":  positionCompare,
		"page":             pageStr,
		"perPage":          perPageStr,
	}

	// Generate cache key based on dataset ID and filters
	cacheKey := GeneratePlayerCacheKey(datasetID, filters)

	// Try to get from format-aware cache
	if cachedData, found := GetCachedPlayerData(ctx, r, cacheKey); found {
		logDebug(ctx, "Using cached player data",
			"dataset_id", datasetID,
			"format", format,
			"player_count", len(cachedData.JSONData),
			"cache_age_seconds", time.Since(cachedData.CacheTime).Seconds())

		SetSpanAttributes(ctx,
			attribute.Bool("cache.hit", true),
			attribute.String("cache.key", cacheKey),
			attribute.Int("cache.player_count", len(cachedData.JSONData)),
			attribute.Float64("cache.age_seconds", time.Since(cachedData.CacheTime).Seconds()),
		)
		if playerDataCacheStatusTotal != nil {
			playerDataCacheStatusTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("cache_status", "hit")))
		}

		// Write the cached response using the appropriate format
		if err := WritePlayerDataResponse(ctx, w, r, cachedData); err != nil {
			logError(ctx, "Error writing cached player data response",
				"error", err,
				"dataset_id", datasetID,
				"format", format)
			WriteErrorResponse(w, r, "response_error", "Error writing response", nil, http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - need to load and process the data
	SetSpanAttributes(ctx, attribute.Bool("cache.hit", false))
	if playerDataCacheStatusTotal != nil {
		playerDataCacheStatusTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("cache_status", "miss")))
	}

	stageStart := time.Now()

	// Ensure configuration is loaded before processing player data
	// This is crucial for calculating role overall ratings and FM attributes
	if err := EnsureConfigInitialized(5 * time.Second); err != nil {
		logWarn(ctx, "Configuration initialization timed out, proceeding with default weights",
			"error", err,
			"dataset_id", datasetID)
		// Continue with default weights rather than failing the request
	}
	logInfo(ctx, "PERF config_init", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
	span.SetAttributes(attribute.Int64("stage.config_init_ms", time.Since(stageStart).Milliseconds()))
	stageStart = time.Now()

	// Parse division filter early
	var divisionFilter = DivisionFilterAll
	switch divisionFilterStr {
	case "same":
		divisionFilter = DivisionFilterSame
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	}

	// If the upload handler's background percentile calculation is still running for
	// this dataset, wait for it to finish. This avoids re-running the full calculation
	// synchronously here while the same work is already in progress on another goroutine.
	if waitForPendingPercentileCalc(datasetID) {
		logDebug(ctx, "Waited for background percentile calculation to complete", "dataset_id", datasetID)
	}
	logInfo(ctx, "PERF percentile_wait", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
	span.SetAttributes(attribute.Int64("stage.percentile_wait_ms", time.Since(stageStart).Milliseconds()))
	stageStart = time.Now()

	// Use the storage interface to get player data
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
	dataSpan.SetAttributes(attribute.Int64("stage.duration_ms", time.Since(stageStart).Milliseconds()))
	dataSpan.End()
	logInfo(ctx, "PERF storage_get", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
	stageStart = time.Now()

	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		SetSpanAttributes(ctx, attribute.String("error.type", "dataset_not_found"))
		WriteErrorResponse(w, r, "dataset_not_found", "Player data not found for the given ID.", nil, http.StatusNotFound)
		return
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.initial_player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
	)

	// Recalculate all player ratings only if necessary
	ctx, recalcSpan := StartSpan(ctx, "ratings.recalculate")
	shouldRecalc := true
	if len(players) > 0 {
		p := players[0]
		// Heuristic: skip recalculation only if Overall, NumericAttributes, and CA are all present.
		// CA = 0 is only valid when the player has no positions; otherwise it signals missing data
		// (e.g. datasets stored before the ca proto field was added).
		caValid := p.CA > 0 || len(p.ShortPositions) == 0
		if p.Overall != 0 && len(p.NumericAttributes) > 0 && caValid {
			shouldRecalc = false
		}
	}
	if shouldRecalc {
		players = RecalculateAllPlayersRatings(players)
	} else {
		logDebug(ctx, "Skipping ratings recalculation; dataset appears precomputed")
	}
	recalcSpan.SetAttributes(
		attribute.Int64("stage.duration_ms", time.Since(stageStart).Milliseconds()),
		attribute.Bool("stage.skipped", !shouldRecalc),
	)
	recalcSpan.End()
	logInfo(ctx, "PERF ratings_recalc", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "skipped", !shouldRecalc)
	stageStart = time.Now()

	// Percentiles are computed relative to the whole population in scope, so we need to know
	// up front (before any copying) whether this request will need to (re)compute them.
	// Checked against the stored `players` slice directly (a field read, not a struct copy) —
	// deep-copying never adds percentiles, so the answer is the same either way.
	hasGlobalPercentiles := len(players) > 0 &&
		players[0].PerformancePercentiles != nil &&
		len(players[0].PerformancePercentiles["Global"]) > 0
	needsPercentileCalc := !hasGlobalPercentiles || divisionFilter != DivisionFilterAll

	var filteredPlayers []Player
	var stepSpan trace.Span

	if needsPercentileCalc {
		// Percentiles must be computed over the full (division-scoped) population before any
		// position/role/age/etc. filter narrows the result, so we can't avoid copying
		// everything up front in this branch. See docs/PERFORMANCE_FIXES_2026-07-08.md #3.
		playersCopy := FastDeepCopyPlayers(players)
		logInfo(ctx, "PERF deep_copy", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "player_count", len(playersCopy))
		span.SetAttributes(attribute.Int64("stage.deep_copy_ms", time.Since(stageStart).Milliseconds()))
		stageStart = time.Now()

		ctx, stepSpan = StartSpan(ctx, "percentiles.calculate")
		filteredPlayersForPercentiles := ApplyDivisionFilter(playersCopy, divisionFilter, targetDivision)
		CalculatePlayerPerformancePercentiles(filteredPlayersForPercentiles)
		stepSpan.SetAttributes(
			attribute.Int64("stage.duration_ms", time.Since(stageStart).Milliseconds()),
			attribute.Bool("stage.skipped", false),
		)
		stepSpan.End()
		logInfo(ctx, "PERF percentile_calc", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "skipped", false)
		stageStart = time.Now()

		ctx, stepSpan = StartSpan(ctx, "filters.apply")
		filteredPlayers = ApplyAllFilters(ctx, playersCopy, filterPosition, filterRole, minAgeStr, maxAgeStr,
			minTransferValueStr, maxTransferValueStr, maxSalaryStr, divisionFilter, targetDivision, positionCompare)
		stepSpan.SetAttributes(attribute.Int64("stage.duration_ms", time.Since(stageStart).Milliseconds()))
		stepSpan.End()
	} else {
		// Fast path: percentiles are already correct on the stored data and divisionFilter is
		// "all" (a no-op), so filter the stored slice first and deep-copy only the survivors —
		// avoids copying players that the response will discard anyway.
		logDebug(ctx, "Skipping percentile calculation; already present on stored data", "dataset_id", datasetID)
		logInfo(ctx, "PERF percentile_calc", "ms", 0, "total_ms", time.Since(startTime).Milliseconds(), "skipped", true)
		span.SetAttributes(attribute.Bool("stage.percentile_calc_skipped", true))

		ctx, stepSpan = StartSpan(ctx, "filters.apply_and_copy")
		filteredPlayers = FastFilterAndCopyPlayers(players, filterPosition, filterRole, minAgeStr, maxAgeStr,
			minTransferValueStr, maxTransferValueStr, maxSalaryStr)
		stepSpan.SetAttributes(attribute.Int64("stage.duration_ms", time.Since(stageStart).Milliseconds()))
		stepSpan.End()
	}
	logInfo(ctx, "PERF filter_apply", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "filtered_count", len(filteredPlayers))
	stageStart = time.Now()

	SetSpanAttributes(ctx,
		attribute.Int("dataset.filtered_player_count", len(filteredPlayers)),
		attribute.String("dataset.currency", currencySymbol),
	)

	// Generate filter hash for cache key
	filterHash := GenerateFilterHash(filters)

	// Optional pagination: additive only. When page/perPage are absent (the default), the
	// full filtered result set is returned exactly as before — every existing consumer that
	// relies on getting everything back (wishlists, squad depth analyzer, CSV export,
	// division comparisons) keeps working unchanged. See docs/PERFORMANCE_FIXES_2026-07-08.md #1.
	totalFilteredCount := len(filteredPlayers)
	responsePlayers := filteredPlayers
	var paginationInfo *pb.PaginationInfo
	if page, perPage, ok := parsePagination(pageStr, perPageStr); ok {
		start := (page - 1) * perPage
		end := start + perPage
		if start > totalFilteredCount {
			start = totalFilteredCount
		}
		if end > totalFilteredCount {
			end = totalFilteredCount
		}
		responsePlayers = filteredPlayers[start:end]
		paginationInfo = CreatePaginationInfo(int32(page), int32(perPage), safeInt32(totalFilteredCount))
	}

	// Build proto response once — reused for both the current response and the cache entry
	protoPlayerResponse := &pb.PlayerDataResponse{
		Players:        make([]*pb.Player, 0, len(responsePlayers)),
		CurrencySymbol: currencySymbol,
		Metadata:       CreateResponseMetadata(requestID, safeInt32(totalFilteredCount), false),
		Pagination:     paginationInfo,
	}
	for _, player := range responsePlayers {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			logError(ctx, "Failed to convert player to protobuf",
				"error", err,
				"player_uid", player.UID,
				"player_name", player.Name)
			continue
		}
		protoPlayerResponse.Players = append(protoPlayerResponse.Players, protoPlayer)
	}
	var protoBytes []byte
	if data, err := proto.Marshal(protoPlayerResponse); err == nil {
		protoBytes = data
	} else {
		logError(ctx, "Failed to marshal protobuf for cache", "error", err)
	}
	logInfo(ctx, "PERF proto_conversion", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
	span.SetAttributes(attribute.Int64("stage.proto_conversion_ms", time.Since(stageStart).Milliseconds()))
	stageStart = time.Now()

	cachedResponse := &CachedPlayerDataResponse{
		Format:         format,
		JSONData:       responsePlayers,
		ProtoBytes:     protoBytes,
		CurrencySymbol: currencySymbol,
		CacheTime:      time.Now(),
		FilterHash:     filterHash,
	}

	// Write the response immediately using pre-built bytes
	if err := WritePlayerDataResponse(ctx, w, r, cachedResponse); err != nil {
		logError(ctx, "Error writing player data response",
			"error", err,
			"dataset_id", datasetID,
			"format", format)
		WriteErrorResponse(w, r, "response_error", "Error writing response", nil, http.StatusInternalServerError)
	}
	logInfo(ctx, "PERF response_write", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
	span.SetAttributes(attribute.Int64("stage.response_write_ms", time.Since(stageStart).Milliseconds()))

	// Cache asynchronously — don't block the response on JSON serialization
	go CachePlayerData(context.Background(), cacheKey, responsePlayers, currencySymbol, filterHash, protoBytes, 5*time.Minute)
}

// parsePagination parses page/perPage query params. Pagination only applies when both are
// present and valid (page >= 1, 0 < perPage <= 1000); otherwise ok is false and callers should
// return the full result set, preserving the pre-pagination default behavior.
func parsePagination(pageStr, perPageStr string) (page, perPage int, ok bool) {
	if pageStr == "" || perPageStr == "" {
		return 0, 0, false
	}
	p, err := strconv.Atoi(pageStr)
	if err != nil || p < 1 {
		return 0, 0, false
	}
	pp, err := strconv.Atoi(perPageStr)
	if err != nil || pp < 1 || pp > 1000 {
		return 0, 0, false
	}
	return p, pp, true
}
