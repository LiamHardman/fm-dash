package main

import (
	"net/http"
	"strings"
	"time"

	pb "api/proto"

	"go.opentelemetry.io/otel/attribute"
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
	stageStart = time.Now()

	// Use the storage interface to get player data
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
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
		if p.OverallLower != 0 && len(p.NumericAttributes) > 0 && caValid {
			shouldRecalc = false
		}
	}
	if shouldRecalc {
		players = RecalculateAllPlayersRatings(players)
	} else {
		logDebug(ctx, "Skipping ratings recalculation; dataset appears precomputed")
	}
	recalcSpan.End()
	logInfo(ctx, "PERF ratings_recalc", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "skipped", !shouldRecalc)
	stageStart = time.Now()

	// Make a deep copy of players to avoid modifying the stored data and prevent race conditions
	playersCopy := FastDeepCopyPlayers(players)
	logInfo(ctx, "PERF deep_copy", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "player_count", len(playersCopy))
	stageStart = time.Now()

	// Calculate percentiles with appropriate filtering using optimized algorithm.
	// Skip if the background goroutine already wrote percentiles into the stored data
	// (which we just deep-copied above), detected by checking the first player.
	ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate")
	hasGlobalPercentiles := len(playersCopy) > 0 &&
		playersCopy[0].PerformancePercentiles != nil &&
		len(playersCopy[0].PerformancePercentiles["Global"]) > 0

	if !hasGlobalPercentiles || divisionFilter != DivisionFilterAll {
		filteredPlayersForPercentiles := ApplyDivisionFilter(playersCopy, divisionFilter, targetDivision)
		CalculatePlayerPerformancePercentiles(filteredPlayersForPercentiles)
	} else {
		logDebug(ctx, "Skipping percentile calculation; already present on stored data", "dataset_id", datasetID)
	}
	percentileSpan.End()
	logInfo(ctx, "PERF percentile_calc", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "skipped", hasGlobalPercentiles && divisionFilter == DivisionFilterAll)
	stageStart = time.Now()

	// Apply all filters to get the final result set
	ctx, filterSpan := StartSpan(ctx, "filters.apply")
	filteredPlayers := ApplyAllFilters(ctx, playersCopy, filterPosition, filterRole, minAgeStr, maxAgeStr,
		minTransferValueStr, maxTransferValueStr, maxSalaryStr, divisionFilter, targetDivision, positionCompare)
	filterSpan.End()
	logInfo(ctx, "PERF filter_apply", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "filtered_count", len(filteredPlayers))
	stageStart = time.Now()

	SetSpanAttributes(ctx,
		attribute.Int("dataset.filtered_player_count", len(filteredPlayers)),
		attribute.String("dataset.currency", currencySymbol),
	)

	// Generate filter hash for cache key
	filterHash := GenerateFilterHash(filters)

	// Cache the filtered player data in both formats
	CachePlayerData(ctx, cacheKey, filteredPlayers, currencySymbol, filterHash, 5*time.Minute)

	// Create a cached response object for immediate use
	cachedResponse := &CachedPlayerDataResponse{
		Format:         format,
		JSONData:       filteredPlayers,
		CurrencySymbol: currencySymbol,
		CacheTime:      time.Now(),
		FilterHash:     filterHash,
	}

	// If protobuf format is requested, create the protobuf data
	if format == FormatTypeProtobuf {
		// Create protobuf response
		protoPlayerResponse := &pb.PlayerDataResponse{
			Players:        make([]*pb.Player, 0, len(filteredPlayers)),
			CurrencySymbol: currencySymbol,
			Metadata:       CreateResponseMetadata(requestID, safeInt32(len(filteredPlayers)), false),
		}

		// Convert each player to protobuf
		for _, player := range filteredPlayers {
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

		// Set the protobuf data in the cache response
		cachedResponse.ProtobufData = protoPlayerResponse
	}
	logInfo(ctx, "PERF proto_conversion", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds(), "format", string(format))
	stageStart = time.Now()

	// Write the response using the appropriate format
	if err := WritePlayerDataResponse(ctx, w, r, cachedResponse); err != nil {
		logError(ctx, "Error writing player data response",
			"error", err,
			"dataset_id", datasetID,
			"format", format)
		WriteErrorResponse(w, r, "response_error", "Error writing response", nil, http.StatusInternalServerError)
	}
	logInfo(ctx, "PERF response_write", "ms", time.Since(stageStart).Milliseconds(), "total_ms", time.Since(startTime).Milliseconds())
}
