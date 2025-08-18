package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	apperrors "api/errors"
	pb "api/proto"
)

// Note: gzip compression middleware already exists in middleware.go

// Pre-allocation helper for better memory efficiency
func calculateOptimalSliceCapacity(expectedSize int) int {
	// Use power of 2 growth with reasonable limits
	if expectedSize <= 0 {
		return defaultPlayerCapacity
	}

	// Round up to next power of 2 for better memory allocation
	capacity := 1
	for capacity < expectedSize {
		capacity <<= 1
	}

	// Cap at reasonable maximum
	if capacity > defaultPlayerCapacity*4 {
		capacity = defaultPlayerCapacity * 4
	}

	return capacity
}

// FileHashToDatasetMap stores the mapping of file content hashes to dataset IDs
var fileHashToDatasetMap = make(map[string]string)
var hashMapMutex sync.RWMutex

func calculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func checkForDuplicateUpload(fileHash string) (string, bool) {
	hashMapMutex.RLock()
	defer hashMapMutex.RUnlock()
	datasetID, exists := fileHashToDatasetMap[fileHash]
	return datasetID, exists
}

func storeDuplicateMapping(fileHash, datasetID string) {
	hashMapMutex.Lock()
	defer hashMapMutex.Unlock()
	fileHashToDatasetMap[fileHash] = datasetID
}

func removeDuplicateMapping(datasetID string) {
	hashMapMutex.Lock()
	defer hashMapMutex.Unlock()
	// Find and remove the mapping by dataset ID
	for hash, id := range fileHashToDatasetMap {
		if id == datasetID {
			delete(fileHashToDatasetMap, hash)
			break
		}
	}
}

func cleanupStaleDuplicateMappings() {
	hashMapMutex.Lock()
	defer hashMapMutex.Unlock()

	var staleMappings []string
	for hash, datasetID := range fileHashToDatasetMap {
		// Check if the dataset still exists
		if _, _, found := GetPlayerData(datasetID); !found {
			staleMappings = append(staleMappings, hash)
		}
	}

	// Remove stale mappings
	for _, hash := range staleMappings {
		delete(fileHashToDatasetMap, hash)
	}

	if len(staleMappings) > 0 {
		logInfo(context.Background(), "Cleaned up %d stale duplicate mappings", len(staleMappings))
	}
}

// setCORSHeaders sets secure CORS headers based on the request origin
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Define allowed origins for production
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if corsOrigins != "" {
		allowedOrigins = strings.Split(corsOrigins, ",")
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		allowedOrigins = []string{
			"http://localhost:3000",  // Development frontend
			"http://localhost:8080",  // Production nginx
			"https://localhost:8080", // Production nginx with SSL
		}
	}

	// Check if the origin is in our allowed list
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			return
		}
	}

	// If no allowed origin matches, don't set CORS headers (more secure default)
	// For completely public APIs, you might set a restrictive default here
}

// calculateOptimalBufferSize determines optimal channel buffer size based on system resources
func calculateOptimalBufferSize(numWorkers int, fileSize int64) int {
	const baseBufferMultiplier = 10
	const maxBufferSize = 1000
	const minBufferSize = 20

	// Base calculation
	baseSize := numWorkers * baseBufferMultiplier

	// Adjust based on file size - larger files get bigger buffers
	sizeAdjustment := int(fileSize / (1024 * 1024)) // MB
	if sizeAdjustment > 50 {
		sizeAdjustment = 50 // Cap the adjustment
	}

	adjustedSize := baseSize + sizeAdjustment

	// Ensure within bounds
	if adjustedSize > maxBufferSize {
		return maxBufferSize
	}
	if adjustedSize < minBufferSize {
		return minBufferSize
	}

	return adjustedSize
}

// Structured logging helpers with trace context that respect LOG_LEVEL
func logDebug(ctx context.Context, msg string, args ...any) {
	if shouldLog(LogLevelDebug) {
		slog.DebugContext(ctx, msg, args...)
	}
}

func logInfo(ctx context.Context, msg string, args ...any) {
	if shouldLog(LogLevelInfo) {
		slog.InfoContext(ctx, msg, args...)
	}
}

func logWarn(ctx context.Context, msg string, args ...any) {
	if shouldLog(LogLevelWarn) {
		slog.WarnContext(ctx, msg, args...)
	}
}

func logError(ctx context.Context, msg string, args ...any) {
	if shouldLog(LogLevelCritical) {
		slog.ErrorContext(ctx, msg, args...)
	}
}

// League represents a division/league with its teams
type League struct {
	Name        string `json:"name"`
	TeamCount   int    `json:"teamCount"`
	PlayerCount int    `json:"playerCount"`
	BestOverall int    `json:"bestOverall"`
	AttRating   int    `json:"attRating"`
	MidRating   int    `json:"midRating"`
	DefRating   int    `json:"defRating"`
}

// Team represents a team with its ratings and stats
type Team struct {
	Name        string   `json:"name"`
	Division    string   `json:"division"`
	PlayerCount int      `json:"playerCount"`
	BestOverall int      `json:"bestOverall"`
	AttRating   int      `json:"attRating"`
	MidRating   int      `json:"midRating"`
	DefRating   int      `json:"defRating"`
	Players     []Player `json:"players,omitempty"`
}

// getMaxUploadSize reads the MAX_UPLOAD_SIZE environment variable and returns the size in bytes.
// If not set or invalid, defaults to 20MB.
func getMaxUploadSize() int64 {
	envValue := os.Getenv("MAX_UPLOAD_SIZE")

	if envValue == "" {
		return 20 * 1024 * 1024 // Default 20MB
	}

	// Parse as integer (expecting value in MB)
	sizeInMB, err := strconv.Atoi(envValue)
	if err != nil || sizeInMB <= 0 {
		logWarn(context.Background(), "Invalid MAX_UPLOAD_SIZE environment variable '%s', defaulting to 20MB", envValue)
		return 20 * 1024 * 1024 // Default 20MB
	}

	result := int64(sizeInMB) * 1024 * 1024
	return result
}

// getFileSizeLimitErrorMessage returns the user-facing error message with the current upload limit
func getFileSizeLimitErrorMessage() string {
	maxUploadSizeMB := getMaxUploadSize() / (1024 * 1024)
	return fmt.Sprintf("Only 10,000 players or less can be in a given dataset. (Max file size: %dMB)", maxUploadSizeMB)
}

// uploadHandler handles POST requests for uploading HTML player files.
// It parses the file, processes player data concurrently, and stores the results.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "upload.handler")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/upload")
	defer DecrementActiveRequests(ctx, "/upload")

	// Record API operation metrics at the end
	defer func() {
		status := http.StatusOK
		//nolint:staticcheck // SA9003: empty branch is intentional for future use
		// TODO: Add span status checking logic here when needed
		RecordAPIOperation(ctx, r.Method, "/upload", status, time.Since(startTime))
	}()

	// Check Content-Length header first for a quick check, though it can be spoofed.
	// r.ContentLength is an int64
	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/upload"),
		attribute.Int64("http.request.content_length", r.ContentLength),
	)

	if r.ContentLength > getMaxUploadSize() {
		logWarn(ctx, "Upload rejected: Content-Length exceeds limit",
			"content_length_bytes", r.ContentLength,
			"max_size_bytes", getMaxUploadSize())
		SetSpanAttributes(ctx, attribute.String("upload.rejection_reason", "content_length_exceeded"))
		http.Error(w, getFileSizeLimitErrorMessage(), http.StatusRequestEntityTooLarge)
		return
	}

	// ParseMultipartForm will also respect the maxMemory argument for in-memory parts,
	// but the total request size is what we're primarily concerned with for the file part.
	// We'll check the actual file size after getting the file.
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB for other form data, not the file itself immediately
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	AddSpanEvent(ctx, "multipart.form.parsed")

	file, handler, err := r.FormFile("playerFile")
	if err != nil {
		RecordError(ctx, err, "Failed to retrieve uploaded file")
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			RecordError(ctx, closeErr, "Failed to close uploaded file")
		}
	}()

	// Validate actual file size by reading content (more secure than relying on headers)
	limitedReader := http.MaxBytesReader(w, file, getMaxUploadSize())
	fileContent, err := io.ReadAll(limitedReader)
	if err != nil {
		RecordError(ctx, err, "File size validation failed - file too large or read error")
		logWarn(ctx, "Upload rejected: File content exceeds size limit or read error",
			"max_size_bytes", getMaxUploadSize(),
			"filename", handler.Filename)
		http.Error(w, getFileSizeLimitErrorMessage(), http.StatusRequestEntityTooLarge)
		return
	}

	actualFileSize := int64(len(fileContent))
	SetSpanAttributes(ctx,
		attribute.String("file.name", handler.Filename),
		attribute.Int64("file.size", actualFileSize),
		attribute.String("file.content_type", handler.Header.Get("Content-Type")),
		attribute.Int64("file.size_from_header", handler.Size),
	)

	logDebug(ctx, "File uploaded",
		"filename", handler.Filename,
		"size_bytes", actualFileSize)

	// Enforce the 50MB limit on the actual file size
	if actualFileSize > getMaxUploadSize() {
		logWarn(ctx, "Upload rejected: File size exceeds limit",
			"filename", handler.Filename,
			"file_size_bytes", actualFileSize,
			"max_size_bytes", getMaxUploadSize())
		SetSpanAttributes(ctx, attribute.String("upload.rejection_reason", "file_size_exceeded"))
		http.Error(w, getFileSizeLimitErrorMessage(), http.StatusRequestEntityTooLarge)
		return
	}

	// Check for duplicate upload before processing
	ctx, duplicateSpan := StartSpan(ctx, "duplicate.check")
	fileHash := calculateFileHash(fileContent)
	existingDatasetID, isDuplicate := checkForDuplicateUpload(fileHash)

	SetSpanAttributes(ctx,
		attribute.String("file.hash", fileHash[:16]+"..."), // Only log first 16 chars for security
		attribute.Bool("duplicate.found", isDuplicate),
	)

	if isDuplicate {
		// Verify the existing dataset still exists in storage
		if _, _, found := GetPlayerData(existingDatasetID); found {
			logInfo(ctx, "Duplicate upload detected, redirecting to existing dataset",
				"filename", handler.Filename,
				"existing_dataset_id", existingDatasetID,
				"file_hash", fileHash[:16]+"...")

			SetSpanAttributes(ctx,
				attribute.String("duplicate.existing_dataset_id", existingDatasetID),
				attribute.String("duplicate.action", "redirect_to_existing"),
			)
			duplicateSpan.End()

			// Return the existing dataset info without reprocessing
			response := UploadResponse{
				DatasetID:              existingDatasetID,
				Message:                "Duplicate file detected. Redirected to existing dataset.",
				DetectedCurrencySymbol: "$", // Will be updated with actual value from existing dataset
			}

			// Get the actual currency symbol from the existing dataset
			if _, currencySymbol, found := GetPlayerData(existingDatasetID); found {
				response.DetectedCurrencySymbol = currencySymbol
			}

			w.Header().Set("Content-Type", "application/json")
			setCORSHeaders(w, r)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				RecordError(ctx, err, "Failed to encode duplicate response")
				http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
				return
			}

			RecordBusinessOperation(ctx, "duplicate_upload_detected", true, map[string]interface{}{
				"filename":            handler.Filename,
				"file_size_bytes":     actualFileSize,
				"existing_dataset_id": existingDatasetID,
				"file_hash":           fileHash[:16] + "...",
			})

			return
		}

		// Dataset no longer exists, remove the stale mapping and continue processing
		logWarn(ctx, "Stale duplicate mapping found, dataset no longer exists",
			"filename", handler.Filename,
			"stale_dataset_id", existingDatasetID)
		removeDuplicateMapping(existingDatasetID)
	}

	AddSpanEvent(ctx, "duplicate.check.completed", attribute.Bool("is_duplicate", isDuplicate))
	duplicateSpan.End()

	// Process all files synchronously
	estimatedPlayerCount := int(actualFileSize / 2048) // Rough estimation: ~2KB per player row
	parseStartTime := time.Now()
	// Optimized pre-allocation based on file size estimation
	if estimatedPlayerCount == 0 {
		estimatedPlayerCount = 100 // Minimum reasonable estimate
	}
	optimalCapacity := calculateOptimalSliceCapacity(estimatedPlayerCount)
	playersList := make([]Player, 0, optimalCapacity)
	var processingError error

	// Ensure configuration is initialized before processing players to avoid slow fallback path
	if err := EnsureConfigInitialized(5 * time.Second); err != nil {
		logWarn(ctx, "Configuration initialization timeout, proceeding with defaults", "error", err)
	}

	numWorkers := runtime.NumCPU()
	if numWorkers == 0 {
		numWorkers = 1
	}
	SetSpanAttributes(ctx,
		attribute.Int("workers.count", numWorkers),
		attribute.String("processing.phase", "setup"),
	)

	// Dynamic buffer sizing based on available memory and system resources
	bufferSize := calculateOptimalBufferSize(numWorkers, actualFileSize)
	rowCellsChan := make(chan []string, bufferSize)
	resultsChan := make(chan PlayerParseResult, bufferSize)
	var wg sync.WaitGroup

	var headersSnapshot []string

	doneConsumingResults := make(chan struct{})
	go func() {
		defer close(doneConsumingResults)
		for result := range resultsChan {
			if result.Err == nil {
				playersList = append(playersList, result.Player)
			} else {
				logWarn(ctx, "Skipping row due to error from worker", "error", result.Err)
			}
		}
		LogDebug("Finished collecting results from resultsChan.")
	}()

	// Start performance timer for parsing
	parseTimer := CreateParseTimerWithContext(ctx, "html_parsing")

	// Wrap file parsing in a child span using the already-read content
	err = TraceFileProcessing(ctx, handler.Filename, actualFileSize, func(_ context.Context) error {
		contentReader := strings.NewReader(string(fileContent))
		return ParseHTMLPlayerTable(contentReader, &headersSnapshot, rowCellsChan, numWorkers, resultsChan, &wg)
	})
	processingError = err

	// Note: rowCellsChan is now closed by ParseHTMLPlayerTable function to prevent race conditions
	LogDebug("HTML parsing attempt finished - channel closed by parser.")

	if processingError != nil {
		RecordError(ctx, processingError, "HTML parsing failed")
		logError(ctx, "Error during HTML parsing or worker setup", "error", processingError)
		if len(headersSnapshot) > 0 {
			logInfo(ctx, "Waiting for any potentially started workers after parsing error...")
			wg.Wait()
		}
		close(resultsChan)
		<-doneConsumingResults
		http.Error(w, processingError.Error(), http.StatusInternalServerError)
		return
	}

	if len(headersSnapshot) == 0 {
		logError(ctx, "Critical: No headers were parsed from the HTML file")
		SetSpanAttributes(ctx, attribute.String("error.type", "no_headers_parsed"))
		close(resultsChan)
		<-doneConsumingResults
		http.Error(w, "Could not parse table headers, no data processed.", http.StatusInternalServerError)
		return
	}

	AddSpanEvent(ctx, "workers.waiting_for_completion")
	LogDebug("Waiting for all player data parser workers to finish...")
	wg.Wait()
	AddSpanEvent(ctx, "workers.completed")
	LogDebug("All workers have completed (wg.Wait() returned).")

	close(resultsChan)
	LogDebug("ResultsChan closed after all workers finished.")

	<-doneConsumingResults
	LogDebug("Results consumer goroutine finished processing all items.")

	// Finish performance timing
	parseTimer.Finish(int64(len(playersList)), 0) // No errors counted here since workers handle errors

	finalDatasetCurrencySymbol := "$" // Default
	if len(playersList) > 0 {
		var foundSymbol bool
		for i := range playersList {
			_, _, tvSymbol := ParseMonetaryValueGo(playersList[i].TransferValue) // Assumes ParseMonetaryValueGo is in parsing.go
			if tvSymbol != "" {
				finalDatasetCurrencySymbol = tvSymbol
				foundSymbol = true
				break
			}
			_, _, wSymbol := ParseMonetaryValueGo(playersList[i].Wage)
			if wSymbol != "" {
				finalDatasetCurrencySymbol = wSymbol
				foundSymbol = true
				break
			}
		}
		if !foundSymbol {
			logInfo(ctx, "No currency symbol detected from parsed player monetary values, using default '$'")
		}
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

	// Debug logging for dataset upload
	logDebug(ctx, "Dataset upload processing",
		"dataset_id", datasetID,
		"player_count", len(playersList),
		"currency", finalDatasetCurrencySymbol,
		"parse_duration_ms", parseDuration.Milliseconds())

	// Store data immediately in memory for fast access (without percentiles initially)
	ctx, storageSpan := StartSpan(ctx, "storage.save_dataset_async")
	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(playersList)),
		attribute.String("dataset.currency", finalDatasetCurrencySymbol),
		attribute.String("storage.method", "async"),
	)

	// Use async storage for performance - data available immediately in memory
	SetPlayerDataAsync(datasetID, playersList, finalDatasetCurrencySymbol)
	storageSpan.End()

	// Store the file hash mapping for duplicate detection
	storeDuplicateMapping(fileHash, datasetID)

	logDebug(ctx, "Duplicate detection mapping stored",
		"dataset_id", datasetID,
		"file_hash", fileHash[:16]+"...")

	logDebug(ctx, "Player data stored successfully",
		"dataset_id", datasetID,
		"player_count", len(playersList),
		"parse_duration", parseDuration,
		"currency_symbol", finalDatasetCurrencySymbol)

	// Normalize MBR scores relative to the maximum (best player gets 100)
	NormalizeMBRScoresRelativeToMax(playersList)

	// Extract roles from processed players for immediate response
	roles := extractRolesFromPlayers(playersList)

	// Create enhanced response with processed data
	response := UploadResponse{
		DatasetID:              datasetID,
		Message:                "File uploaded and processed successfully.",
		DetectedCurrencySymbol: finalDatasetCurrencySymbol,
		Players:                processBasicPlayerData(playersList), // Return processed players
		Roles:                  roles,                               // Return available roles
		PercentilesReady:       false,                               // Percentiles will be calculated asynchronously

		PlayerCount: len(playersList),
	}

	// Debug logging for upload response
	logDebug(ctx, "Upload response generated",
		"dataset_id", datasetID,
		"player_count", len(playersList),
		"response_ready", true)

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		RecordError(ctx, err, "Failed to encode upload response")
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Start async percentile calculation
	go func() {
		ctx, asyncSpan := StartSpan(context.Background(), "async.percentile_calculation")
		defer asyncSpan.End()

		SetSpanAttributes(ctx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("player_count", len(playersList)),
		)

		logInfo(ctx, "Starting async percentile calculation for dataset %s", datasetID)
		startTime := time.Now()

		// Calculate percentiles for all division filters to ensure stability
		CalculatePlayerPerformancePercentiles(playersList)

		// Log top 25 MBR players after calculations are complete
		logTop25OverallPlayers(playersList)

		// Update the stored data with calculated percentiles
		SetPlayerData(datasetID, playersList, finalDatasetCurrencySymbol)

		calculationTime := time.Since(startTime)
		logInfo(ctx, "Completed async percentile calculation for dataset %s in %v", datasetID, calculationTime)

		RecordBusinessOperation(ctx, "async_percentile_calculation_completed", true, map[string]interface{}{
			"dataset_id":       datasetID,
			"player_count":     len(playersList),
			"calculation_time": calculationTime.Milliseconds(),
		})
	}()

	RecordBusinessOperation(ctx, "file_upload_completed", true, map[string]interface{}{
		"filename":             handler.Filename,
		"file_size_bytes":      actualFileSize,
		"dataset_id":           datasetID,
		"player_count":         len(playersList),
		"parse_duration_ms":    parseDuration.Milliseconds(),
		"currency_symbol":      finalDatasetCurrencySymbol,
		"roles_count":          len(roles),
		"total_upload_time_ms": time.Since(startTime).Milliseconds(),
	})
}

// playerDataHandler handles GET requests for retrieving player data by dataset ID.
func playerDataHandler(w http.ResponseWriter, r *http.Request) {
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
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/players"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
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

	queryValues := r.URL.Query()

	// Check for cache-busting parameter to fix field name compatibility issues
	if queryValues.Get("clear_cache") == "true" {
		invalidateDatasetCache(datasetID)
		logInfo(ctx, "Dataset cache cleared due to clear_cache parameter", "dataset_id", datasetID)
	}

	// OPTIMIZATION: Create query fingerprint for result caching
	queryFingerprint := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		datasetID, queryValues.Get("position"), queryValues.Get("role"),
		queryValues.Get("minAge"), queryValues.Get("maxAge"),
		queryValues.Get("minTransferValue"), queryValues.Get("maxTransferValue"),
		queryValues.Get("maxSalary"), queryValues.Get("division_filter"),
		queryValues.Get("target_division"))

	resultCacheKey := "result:" + queryFingerprint
	if cachedResult, found := getFromMemCache(resultCacheKey); found {
		if cacheData, ok := cachedResult.(map[string]interface{}); ok {
			SetSpanAttributes(ctx, attribute.Bool("result_cache.hit", true))
			logDebug(ctx, "Result cache hit", "cache_key", resultCacheKey)
			if err := WriteResponse(w, r, cacheData); err != nil {
				logWarn(ctx, "Failed to write cached response", "error", err)
			}
			return
		}
	}
	SetSpanAttributes(ctx, attribute.Bool("result_cache.hit", false))

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
		"position_compare", positionCompare)

	// Create cache key for percentile-calculated data (separate from final filtered result)
	percentileCacheKey := fmt.Sprintf("percentiles:%s:%s:%s", datasetID, divisionFilterStr, targetDivision)

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

	// Check cache for percentile-calculated players first
	var players []Player
	var currencySymbol string
	var found bool

	if cachedData, cacheFound := getFromMemCache(percentileCacheKey); cacheFound {
		if cachedResult, ok := cachedData.(struct {
			Players        []Player
			CurrencySymbol string
		}); ok {
			players = cachedResult.Players
			currencySymbol = cachedResult.CurrencySymbol
			found = true
			logDebug(ctx, "Using cached percentile data", "dataset_id", datasetID, "division_filter", divisionFilterStr)
			SetSpanAttributes(ctx, attribute.Bool("percentile_cache.hit", true))
		}
	}

	if !found {
		// Cache miss - need to load and calculate percentiles
		SetSpanAttributes(ctx, attribute.Bool("percentile_cache.hit", false))

		// Use the new storage interface to get player data
		ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
		players, currencySymbol, found = GetPlayerData(datasetID)
		dataSpan.End()

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

		// Make a deep copy of players to avoid modifying the stored data and prevent race conditions
		// CRITICAL: Use simple deepCopyPlayers instead of OptimizedDeepCopyPlayers (COW has race conditions)
		playersCopy := deepCopyPlayers(players)

		// Recalculate all player ratings based on the current calculation method setting
		ctx, recalcSpan := StartSpan(ctx, "ratings.recalculate")
		playersCopy = RecalculateAllPlayersRatings(playersCopy)
		recalcSpan.End()

		// Normalize MBR scores relative to the maximum (best player gets 100)
		NormalizeMBRScoresRelativeToMax(playersCopy)

		// Calculate percentiles with appropriate filtering using optimized algorithm
		ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate")

		// Fast path: if divisionFilter is ALL and players already have global percentiles, skip recomputation
		hasGlobal := false
		if len(playersCopy) > 0 && playersCopy[0].PerformancePercentiles != nil {
			if _, ok := playersCopy[0].PerformancePercentiles["Global"]; ok {
				hasGlobal = true
			}
		}

		// Try loading reusable percentile distributions from persistent cache
		distributions, foundDists := loadPercentileDistributionsFromCache(ctx, datasetID, divisionFilterStr, targetDivision)

		if divisionFilter == DivisionFilterAll && hasGlobal {
			// Skip recalculation entirely when global percentiles exist
		} else if foundDists {
			// Apply precomputed distributions
			ApplyPercentilesFromDistributions(playersCopy, distributions)
		} else {
			// Compute distributions, apply, and persist reusable cache
			dists := BuildPercentileDistributions(playersCopy, divisionFilter, targetDivision)
			ApplyPercentilesFromDistributions(playersCopy, dists)
			// Persist for reuse
			savePercentileDistributionsToCache(ctx, datasetID, divisionFilterStr, targetDivision, dists)
		}

		players = playersCopy
		percentileSpan.End()

		// Cache only light wrapper; heavy work is now in reusable distributions cache
		cacheData := struct {
			Players        []Player
			CurrencySymbol string
		}{
			Players:        playersCopy,
			CurrencySymbol: currencySymbol,
		}
		setInMemCacheForDataset(percentileCacheKey, cacheData, 10*time.Minute) // Cache for 10 minutes

		logDebug(ctx, "Calculated and cached percentiles",
			"dataset_id", datasetID,
			"division_filter", divisionFilterStr,
			"player_count", len(players))
	}

	// Create cache key for final filtered result
	finalCacheKey := fmt.Sprintf("filtered:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		datasetID, filterPosition, filterRole, minAgeStr, maxAgeStr,
		minTransferValueStr, maxTransferValueStr, maxSalaryStr, divisionFilterStr, targetDivision)

	// Check cache for final filtered result
	if cachedFiltered, cacheFound := getFromMemCache(finalCacheKey); cacheFound {
		if jsonData, ok := cachedFiltered.([]byte); ok {
			logDebug(ctx, "Serving filtered player data from cache", "dataset_id", datasetID)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=180") // Cache for 3 minutes
			setCORSHeaders(w, r)
			if _, err := w.Write(jsonData); err != nil {
				logError(ctx, "Error writing cached response", "error", err)
			}
			SetSpanAttributes(ctx, attribute.Bool("final_cache.hit", true))
			return
		}
	}

	SetSpanAttributes(ctx, attribute.Bool("final_cache.hit", false))

	data := struct {
		Players        []Player
		CurrencySymbol string
	}{Players: players, CurrencySymbol: currencySymbol}

	processedPlayers := make([]Player, 0, len(data.Players))

	var minAge, maxAge = -1, -1
	var minTransferValue, maxTransferValue int64 = -1, -1
	var maxSalary int64 = -1

	if val, err := strconv.Atoi(minAgeStr); err == nil {
		minAge = val
	}
	if val, err := strconv.Atoi(maxAgeStr); err == nil {
		maxAge = val
	}
	if val, err := strconv.ParseInt(minTransferValueStr, 10, 64); err == nil {
		minTransferValue = val
	}
	if val, err := strconv.ParseInt(maxTransferValueStr, 10, 64); err == nil {
		maxTransferValue = val
	}
	if val, err := strconv.ParseInt(maxSalaryStr, 10, 64); err == nil {
		maxSalary = val
	}

	for i := range data.Players {
		playerCopy := data.Players[i]

		if filterPosition != "" {
			canPlayPosition := false
			for _, shortPos := range playerCopy.ShortPositions {
				if shortPos == filterPosition {
					canPlayPosition = true
					break
				}
			}
			if !canPlayPosition {
				continue
			}
		}

		playerAgeVal, ageErr := strconv.Atoi(playerCopy.Age)
		if ageErr == nil {
			if minAge != -1 && playerAgeVal < minAge {
				continue
			}
			if maxAge != -1 && playerAgeVal > maxAge {
				continue
			}
		} else if minAge != -1 || maxAge != -1 {
			logWarn(ctx, "Skipping player due to unparsable age while age filters are active", "player_name", playerCopy.Name, "age", playerCopy.Age)
			continue
		}

		if minTransferValue != -1 && playerCopy.TransferValueAmount < minTransferValue {
			continue
		}
		if maxTransferValue != -1 && playerCopy.TransferValueAmount > maxTransferValue {
			continue
		}

		if maxSalary != -1 && playerCopy.WageAmount > maxSalary {
			continue
		}

		if filterRole != "" {
			roleMatched := false
			for _, roleOverall := range playerCopy.RoleSpecificOveralls {
				if roleOverall.RoleName == filterRole {
					playerCopy.Overall = roleOverall.Score // Update player's main overall to the role-specific one for display
					roleMatched = true
					break
				}
			}
			if !roleMatched {
				playerCopy.Overall = 0 // Set to 0 for unmatched roles when filtering by role
			}
		}
		processedPlayers = append(processedPlayers, playerCopy)
	}

	logDebug(ctx, "Returning processed players", "dataset_id", datasetID, "player_count", len(processedPlayers))

	// Create response data structure
	responseData := PlayerDataWithCurrency{Players: processedPlayers, CurrencySymbol: currencySymbol}

	// Use the content negotiation system to write the response
	if supportsProtobuf {
		// Create protobuf response with metadata
		metadata := CreateResponseMetadata(requestID, safeInt32(len(processedPlayers)), found)
		protoResponse := &pb.PlayerDataResponse{
			Players:        make([]*pb.Player, len(processedPlayers)),
			CurrencySymbol: currencySymbol,
			Metadata:       metadata,
		}

		// Convert players to protobuf format
		for i, player := range processedPlayers {
			protoPlayer, err := player.ToProto(ctx)
			if err != nil {
				logError(ctx, "Failed to convert player to protobuf",
					"error", err,
					"player_uid", player.UID,
					"player_name", player.Name)
				// Fallback to JSON on conversion error
				break
			}
			protoResponse.Players[i] = protoPlayer
		}

		// Try to write protobuf response using content negotiation
		if err := WriteResponse(w, r, protoResponse); err == nil {
			// Protobuf response written successfully
			SetSpanAttributes(ctx,
				attribute.String("response.serialization", "protobuf"),
				attribute.Int("response.size_bytes", len(processedPlayers)*1024), // Rough estimate
			)

			logDebug(ctx, "Protobuf response written successfully",
				"player_count", len(processedPlayers))
			return
		} else {
			logWarn(ctx, "Protobuf serialization failed, falling back to JSON",
				"error", err,
				"dataset_id", datasetID)
			SetSpanAttributes(ctx,
				attribute.String("fallback.reason", "protobuf_serialization_failed"),
				attribute.String("fallback.error", err.Error()),
			)
		}
	}

	// Fallback to JSON response
	SetSpanAttributes(ctx,
		attribute.String("response.serialization", "json"),
		attribute.Int("response.size_bytes", len(processedPlayers)*1024), // Rough estimate
	)

	logDebug(ctx, "Writing JSON response",
		"player_count", len(processedPlayers))

	// Write JSON response using content negotiation
	if err := WriteResponse(w, r, responseData); err != nil {
		RecordError(ctx, err, "Failed to write JSON response")
		WriteErrorResponse(w, r, "SERIALIZATION_ERROR",
			"Failed to serialize response data",
			[]string{err.Error()},
			http.StatusInternalServerError)
		return
	}

	logDebug(ctx, "Player data processed and served",
		"dataset_id", datasetID,
		"processed_count", len(processedPlayers),
		"original_count", len(data.Players),
		"response_format", serializer.ContentType(),
		"processing_time_ms", time.Since(startTime).Milliseconds())

	RecordBusinessOperation(ctx, "player_data_served", true, map[string]interface{}{
		"dataset_id":           datasetID,
		"player_count":         len(processedPlayers),
		"response_format":      serializer.ContentType(),
		"processing_time_ms":   time.Since(startTime).Milliseconds(),
		"percentile_cache_hit": found,
		"division_filter":      divisionFilterStr,
		"has_filters":          filterPosition != "" || filterRole != "" || minAgeStr != "" || maxAgeStr != "",
	})
}

// leaguesHandler returns league data with teams and their ratings
func leaguesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.leagues.get")
	defer span.End()

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/leagues"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/leagues/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	logInfo(ctx, "Processing leagues request", "dataset_id", datasetID)

	// Try to get leagues data from cache first
	cacheKey := fmt.Sprintf("leagues_%s", datasetID)
	if cached, found := getFromMemCache(cacheKey); found {
		if leaguesData, ok := cached.([]League); ok {
			logInfo(ctx, "Retrieved leagues data from memory cache", "dataset_id", datasetID)

			// Set CORS headers
			setCORSHeaders(w, r)

			// Create response metadata
			metadata := CreateResponseMetadata(requestID, safeInt32(len(leaguesData)), true)

			if supportsProtobuf {
				// Create protobuf response
				protoLeagues := make([]string, len(leaguesData))
				for i, league := range leaguesData {
					protoLeagues[i] = league.Name
				}

				protoResponse := &pb.LeaguesResponse{
					Leagues:  protoLeagues,
					Metadata: metadata,
				}

				// Serialize to protobuf
				responseBytes, err := serializer.Serialize(protoResponse)
				if err == nil {
					// Protobuf serialization successful
					w.Header().Set("Content-Type", serializer.ContentType())
					w.Header().Set("X-Cache-Source", "memory")
					w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
					if serializer.ShouldCompress() {
						w.Header().Set("Content-Encoding", "gzip")
					}

					if _, writeErr := w.Write(responseBytes); writeErr != nil {
						logError(ctx, "Error writing protobuf response", "error", writeErr)
					}

					logDebug(ctx, "Leagues served as protobuf from cache",
						"league_count", len(leaguesData),
						"response_size_bytes", len(responseBytes),
						"processing_time_ms", time.Since(startTime).Milliseconds())
					return
				}

				// Log protobuf serialization failure
				logWarn(ctx, "Protobuf serialization failed for cached leagues, falling back to JSON", "error", err)
			}

			// Fallback to JSON
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
			if err := json.NewEncoder(w).Encode(leaguesData); err != nil {
				WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
				logError(ctx, "Error encoding JSON response for cached leagues",
					"error", err,
					"dataset_id", datasetID)
				return
			}
			return
		}
	}

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		WriteErrorResponse(w, r, "dataset_not_found", "Player data not found for the given ID.", nil, http.StatusNotFound)
		return
	}

	// Recalculate all player ratings based on the current calculation method setting
	players = RecalculateAllPlayersRatings(players)

	// Process leagues data with concurrent processing
	processor := CreateConcurrentLeagueProcessor(runtime.NumCPU())
	leaguesData := processor.ProcessLeaguesAsync(ctx, players)

	// Cache the result for 5 minutes
	setInMemCache(cacheKey, leaguesData, 5*time.Minute)

	// Set CORS headers
	setCORSHeaders(w, r)

	// Create response metadata
	metadata := CreateResponseMetadata(requestID, safeInt32(len(leaguesData)), false)

	if supportsProtobuf {
		// Create protobuf response
		protoLeagues := make([]string, len(leaguesData))
		for i, league := range leaguesData {
			protoLeagues[i] = league.Name
		}

		protoResponse := &pb.LeaguesResponse{
			Leagues:  protoLeagues,
			Metadata: metadata,
		}

		// Serialize to protobuf
		responseBytes, err := serializer.Serialize(protoResponse)
		if err == nil {
			// Protobuf serialization successful
			w.Header().Set("Content-Type", serializer.ContentType())
			w.Header().Set("X-Cache-Source", "computed")
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
			if serializer.ShouldCompress() {
				w.Header().Set("Content-Encoding", "gzip")
			}

			if _, writeErr := w.Write(responseBytes); writeErr != nil {
				logError(ctx, "Error writing protobuf response", "error", writeErr)
			}

			logDebug(ctx, "Leagues served as protobuf",
				"league_count", len(leaguesData),
				"response_size_bytes", len(responseBytes),
				"processing_time_ms", time.Since(startTime).Milliseconds())
			return
		}

		// Log protobuf serialization failure
		logWarn(ctx, "Protobuf serialization failed for leagues, falling back to JSON", "error", err)
	}

	// Fallback to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Source", "computed")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
	if err := json.NewEncoder(w).Encode(leaguesData); err != nil {
		WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		logError(ctx, "Error encoding JSON response for leagues",
			"error", err,
			"dataset_id", datasetID)
		return
	}

	logDebug(ctx, "Leagues served as JSON",
		"league_count", len(leaguesData),
		"processing_time_ms", time.Since(startTime).Milliseconds())
}

// teamsHandler returns detailed team data for a specific league
func teamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.teams.get")
	defer span.End()

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/teams"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/teams/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" || pathParts[1] == "" {
		WriteErrorResponse(w, r, "missing_parameters", "Dataset ID and Division are required in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]
	division := pathParts[1]

	logInfo(ctx, "Processing teams request", "dataset_id", datasetID, "division", division)

	// Try to get teams data from cache first
	cacheKey := fmt.Sprintf("teams_%s_%s", datasetID, division)
	if cached, found := getFromMemCache(cacheKey); found {
		if teamsData, ok := cached.([]Team); ok {
			logInfo(ctx, "Retrieved teams data from memory cache", "dataset_id", datasetID, "division", division)

			// Set CORS headers
			setCORSHeaders(w, r)

			// Create response metadata
			metadata := CreateResponseMetadata(requestID, safeInt32(len(teamsData)), true)

			if supportsProtobuf {
				// Create protobuf response
				protoTeams := make([]string, len(teamsData))
				for i, team := range teamsData {
					protoTeams[i] = team.Name
				}

				protoResponse := &pb.TeamsResponse{
					Teams:    protoTeams,
					Metadata: metadata,
				}

				// Serialize to protobuf
				responseBytes, err := serializer.Serialize(protoResponse)
				if err == nil {
					// Protobuf serialization successful
					w.Header().Set("Content-Type", serializer.ContentType())
					w.Header().Set("X-Cache-Source", "memory")
					w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
					if serializer.ShouldCompress() {
						w.Header().Set("Content-Encoding", "gzip")
					}

					if _, writeErr := w.Write(responseBytes); writeErr != nil {
						logError(ctx, "Error writing protobuf response", "error", writeErr)
					}

					logDebug(ctx, "Teams served as protobuf from cache",
						"team_count", len(teamsData),
						"response_size_bytes", len(responseBytes),
						"processing_time_ms", time.Since(startTime).Milliseconds())
					return
				}

				// Log protobuf serialization failure
				logWarn(ctx, "Protobuf serialization failed for cached teams, falling back to JSON", "error", err)
			}

			// Fallback to JSON
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
			if err := json.NewEncoder(w).Encode(teamsData); err != nil {
				WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
				logError(ctx, "Error encoding JSON response for cached teams",
					"error", err,
					"dataset_id", datasetID,
					"division", division)
			}
			return
		}
	}

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Recalculate all player ratings based on the current calculation method setting
	players = RecalculateAllPlayersRatings(players)

	// Process teams data for the specific division
	teamsData := processTeamsData(players, division)

	// Cache the result for 5 minutes
	setInMemCache(cacheKey, teamsData, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Source", "computed")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(teamsData); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		logError(ctx, "Error encoding JSON response for teams", "dataset_id", sanitizeForLogging(datasetID), "division", sanitizeForLogging(division), "error", err)
	}
}

// processTeamsData returns detailed team data for a specific division
func processTeamsData(players []Player, division string) []Team {
	// Pre-allocate division players with estimated capacity
	estimatedDivisionPlayers := len(players) / 10 // Estimate ~10% of players in any given division
	if estimatedDivisionPlayers < 50 {
		estimatedDivisionPlayers = 50
	}
	divisionPlayers := make([]Player, 0, estimatedDivisionPlayers)

	// Filter players by division
	for i := range players {
		if players[i].Division == division {
			divisionPlayers = append(divisionPlayers, players[i])
		}
	}

	// Group by team with estimated team count
	estimatedTeams := len(divisionPlayers) / 25 // Estimate ~25 players per team
	if estimatedTeams < 10 {
		estimatedTeams = 10 // Minimum reasonable team count
	}
	if estimatedTeams > 30 {
		estimatedTeams = 30 // Maximum reasonable team count per division
	}

	teamMap := make(map[string][]Player, estimatedTeams)
	for i := range divisionPlayers {
		if divisionPlayers[i].Club != "" {
			teamMap[divisionPlayers[i].Club] = append(teamMap[divisionPlayers[i].Club], divisionPlayers[i])
		}
	}

	teams := make([]Team, 0, len(teamMap))
	for teamName, teamPlayers := range teamMap {
		team := Team{
			Name:        teamName,
			Division:    division,
			PlayerCount: len(teamPlayers),
			Players:     teamPlayers,
		}

		ratings := calculateTeamRatings(teamPlayers)
		team.BestOverall = ratings.BestOverall
		team.AttRating = ratings.AttRating
		team.MidRating = ratings.MidRating
		team.DefRating = ratings.DefRating

		teams = append(teams, team)
	}

	// Sort teams by overall rating
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].BestOverall > teams[j].BestOverall
	})

	return teams
}

// PercentileRequest represents the request body for percentile recalculation
type PercentileRequest struct {
	PlayerName     string `json:"playerName"`
	DivisionFilter string `json:"divisionFilter"`
	TargetDivision string `json:"targetDivision"`
}

// PlayerPercentilesRequest represents a request for player percentiles by UID
type PlayerPercentilesRequest struct {
	PlayerUID       string `json:"playerUID"`
	CompareDivision string `json:"compareDivision"` // "all", "same", "top5"
	ComparePosition string `json:"comparePosition"` // position group like "Global", "Midfielders", etc.
}

// percentilesHandler handles POST requests to recalculate percentiles for a specific player with division filtering
func percentilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/percentiles/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	var req PercentileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	logInfo(ctx, "Processing percentiles request",
		"dataset_id", datasetID,
		"player_name", req.PlayerName,
		"division_filter", req.DivisionFilter,
		"target_division", req.TargetDivision)

	// Get the full dataset
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Find the specific player
	var targetPlayerIndex = -1
	for i := range players {
		if players[i].Name == req.PlayerName {
			targetPlayerIndex = i
			break
		}
	}

	if targetPlayerIndex == -1 {
		http.Error(w, "Player not found in dataset", http.StatusNotFound)
		return
	}

	// Parse division filter
	var divisionFilter = DivisionFilterAll
	switch req.DivisionFilter {
	case "same":
		divisionFilter = DivisionFilterSame
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	}

	// NEW: Generate cache key and try to load from cache first
	cacheKey := generatePercentilesCacheKey(ctx, datasetID, req.PlayerName, req.DivisionFilter, req.TargetDivision, players)

	logDebug(ctx, "Generated cache key for percentiles request",
		"dataset_id", datasetID,
		"player_name", req.PlayerName,
		"division_filter", req.DivisionFilter,
		"target_division", req.TargetDivision,
		"cache_key", cacheKey,
		"player_count", len(players))

	// Try to load from cache
	if cachedPercentiles, found := loadPercentilesFromCache(ctx, cacheKey, datasetID, req.PlayerName, req.DivisionFilter, req.TargetDivision, players); found {
		logDebug(ctx, "🎯 CACHE HIT - Returning cached percentiles",
			"dataset_id", datasetID,
			"player_name", req.PlayerName,
			"division_filter", req.DivisionFilter,
			"target_division", req.TargetDivision,
			"cache_key", cacheKey)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Status", "HIT")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(cachedPercentiles); err != nil {
			logError(ctx, "Error encoding JSON response for cached percentiles", "dataset_id", sanitizeForLogging(datasetID), "error", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - perform optimized calculation
	logDebug(ctx, "💫 CACHE MISS - calculating percentiles",
		"dataset_id", datasetID,
		"player_name", req.PlayerName,
		"division_filter", req.DivisionFilter,
		"target_division", req.TargetDivision,
		"cache_key", cacheKey)

	// OPTIMIZATION: Filter players first, then only copy the filtered subset
	var filteredPlayers []Player
	var targetPlayerInFiltered bool

	// Filter players based on division criteria
	for i := range players {
		if isPlayerInTargetDivision(&players[i], divisionFilter, req.TargetDivision) {
			filteredPlayers = append(filteredPlayers, players[i])
			if i == targetPlayerIndex {
				targetPlayerInFiltered = true
			}
		}
	}

	// If target player is not in filtered set, add them
	if !targetPlayerInFiltered {
		filteredPlayers = append(filteredPlayers, players[targetPlayerIndex])
	}

	logDebug(ctx, "Filtered players for percentile calculation",
		"total_players", len(players),
		"filtered_players", len(filteredPlayers),
		"target_player_included", targetPlayerInFiltered)

	// Create a copy of only the filtered players (much smaller dataset)
	playersCopy := make([]Player, len(filteredPlayers))
	copy(playersCopy, filteredPlayers)

	// Calculate percentiles only for the filtered subset
	CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, DivisionFilterAll, "")

	// Find the target player in the filtered copy
	var updatedPlayer *Player
	for i := range playersCopy {
		if playersCopy[i].Name == req.PlayerName {
			updatedPlayer = &playersCopy[i]
			break
		}
	}

	if updatedPlayer == nil {
		http.Error(w, "Player not found after calculation", http.StatusInternalServerError)
		return
	}

	// Get the updated percentiles for the target player
	updatedPercentiles := updatedPlayer.PerformancePercentiles

	// NEW: Save to cache for future requests
	go func() {
		savePercentilesToCache(ctx, cacheKey, datasetID, req.PlayerName, req.DivisionFilter, req.TargetDivision, players, updatedPercentiles)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(updatedPercentiles); err != nil {
		logError(ctx, "Error encoding JSON response for percentiles", "dataset_id", sanitizeForLogging(datasetID), "error", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// playerPercentilesHandler handles POST requests to get percentiles for a specific player by UID
func playerPercentilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/player-percentiles/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	var req PlayerPercentilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	logInfo(ctx, "Processing player percentiles request",
		"dataset_id", datasetID,
		"player_uid", req.PlayerUID,
		"compare_division", req.CompareDivision,
		"compare_position", req.ComparePosition)

	// Get the full dataset
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Find the specific player by UID
	var targetPlayer *Player
	playerUID, err := strconv.ParseInt(req.PlayerUID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid player UID format", http.StatusBadRequest)
		return
	}

	for i := range players {
		if players[i].UID == playerUID {
			targetPlayer = &players[i]
			break
		}
	}

	if targetPlayer == nil {
		http.Error(w, "Player not found in dataset", http.StatusNotFound)
		return
	}

	// Parse division filter
	var divisionFilter = DivisionFilterAll
	var targetDivision = ""

	switch req.CompareDivision {
	case "same":
		divisionFilter = DivisionFilterSame
		// For 'same', we need to get the target player's division
		if targetPlayer != nil {
			targetDivision = targetPlayer.Division
		}
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	default:
		// If it's not a special filter type, treat it as a specific division
		divisionFilter = DivisionFilterSame
		targetDivision = req.CompareDivision
	}

	// Generate cache key for this specific request
	cacheKey := generatePlayerPercentilesCacheKey(ctx, datasetID, req.PlayerUID, req.CompareDivision, req.ComparePosition, players)

	logDebug(ctx, "Generated cache key for player percentiles request",
		"dataset_id", datasetID,
		"player_uid", req.PlayerUID,
		"compare_division", req.CompareDivision,
		"compare_position", req.ComparePosition,
		"cache_key", cacheKey,
		"player_count", len(players))

	// Try to load from cache
	if cachedPercentiles, found := loadPlayerPercentilesFromCache(ctx, cacheKey, datasetID, req.PlayerUID, req.CompareDivision, req.ComparePosition, players); found {
		logDebug(ctx, "🎯 CACHE HIT - Returning cached player percentiles",
			"dataset_id", datasetID,
			"player_uid", req.PlayerUID,
			"compare_division", req.CompareDivision,
			"compare_position", req.ComparePosition,
			"cache_key", cacheKey)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Status", "HIT")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(cachedPercentiles); err != nil {
			logError(ctx, "Error encoding JSON response for cached player percentiles", "dataset_id", sanitizeForLogging(datasetID), "error", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - perform optimized calculation
	logDebug(ctx, "💫 CACHE MISS - calculating player percentiles",
		"dataset_id", datasetID,
		"player_uid", req.PlayerUID,
		"compare_division", req.CompareDivision,
		"compare_position", req.ComparePosition,
		"cache_key", cacheKey)

	// OPTIMIZATION: Filter players first, then only copy the filtered subset
	var filteredPlayers []Player
	var targetPlayerInFiltered bool

	// Filter players based on division criteria
	for i := range players {
		if isPlayerInTargetDivision(&players[i], divisionFilter, targetDivision) {
			filteredPlayers = append(filteredPlayers, players[i])
			if players[i].UID == playerUID {
				targetPlayerInFiltered = true
			}
		}
	}

	// If target player is not in filtered set, add them
	if !targetPlayerInFiltered {
		filteredPlayers = append(filteredPlayers, *targetPlayer)
	}

	logDebug(ctx, "Filtered players for percentile calculation",
		"total_players", len(players),
		"filtered_players", len(filteredPlayers),
		"target_player_included", targetPlayerInFiltered)

	// Create a copy of only the filtered players (much smaller dataset)
	playersCopy := make([]Player, len(filteredPlayers))
	copy(playersCopy, filteredPlayers)

	// Calculate percentiles only for the filtered subset
	CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, DivisionFilterAll, "")

	// Find the target player in the filtered copy
	var updatedPlayer *Player
	for i := range playersCopy {
		if playersCopy[i].UID == playerUID {
			updatedPlayer = &playersCopy[i]
			break
		}
	}

	if updatedPlayer == nil {
		http.Error(w, "Player not found after calculation", http.StatusInternalServerError)
		return
	}

	// Get the percentiles for the specific position group
	var resultPercentiles map[string]interface{}
	if req.ComparePosition == "Global" || req.ComparePosition == "" {
		// Return all percentiles
		resultPercentiles = make(map[string]interface{})
		for group, percentiles := range updatedPlayer.PerformancePercentiles {
			resultPercentiles[group] = percentiles
		}
	} else {
		// Return only the specific position group
		if percentiles, exists := updatedPlayer.PerformancePercentiles[req.ComparePosition]; exists {
			resultPercentiles = map[string]interface{}{
				req.ComparePosition: percentiles,
			}
		} else {
			// Position group not found, return empty result
			resultPercentiles = map[string]interface{}{
				req.ComparePosition: map[string]int{},
			}
		}
	}

	// Save to cache for future requests
	go func() {
		savePlayerPercentilesToCache(ctx, cacheKey, datasetID, req.PlayerUID, req.CompareDivision, req.ComparePosition, players, resultPercentiles)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(resultPercentiles); err != nil {
		logError(ctx, "Error encoding JSON response for player percentiles", "dataset_id", sanitizeForLogging(datasetID), "error", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// percentilesStatusHandler handles GET requests to check if percentiles are ready for a dataset
func percentilesStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/percentiles-status/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	logInfo(ctx, "Checking percentiles status", "dataset_id", datasetID)

	// Get the dataset
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Check percentile readiness
	totalPlayers := len(players)
	playersWithPercentiles := 0
	playersWithValidPercentiles := 0

	for i := range players {
		if players[i].PerformancePercentiles != nil {
			playersWithPercentiles++

			// Check if the player has valid percentiles (not all -1 or 0)
			globalPercentiles := players[i].PerformancePercentiles["Global"]
			if globalPercentiles != nil {
				validCount := 0
				for _, percentile := range globalPercentiles {
					if percentile >= 0 {
						validCount++
					}
				}
				if validCount > 0 {
					playersWithValidPercentiles++
				}
			}
		}
	}

	// Calculate readiness percentages
	percentileInitialized := float64(playersWithPercentiles) / float64(totalPlayers) * 100
	percentileValid := float64(playersWithValidPercentiles) / float64(totalPlayers) * 100

	// Determine overall status
	var status string
	switch {
	case percentileValid >= 90:
		status = "ready"
	case percentileValid >= 50:
		status = "partial"
	case percentileInitialized >= 50:
		status = "calculating"
	default:
		status = "not_ready"
	}

	statusResponse := map[string]interface{}{
		"dataset_id":                     datasetID,
		"status":                         status,
		"total_players":                  totalPlayers,
		"players_with_percentiles":       playersWithPercentiles,
		"players_with_valid_percentiles": playersWithValidPercentiles,
		"percentile_initialized_percent": percentileInitialized,
		"percentile_valid_percent":       percentileValid,
		"message":                        getStatusMessage(status, percentileValid),
	}

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(statusResponse); err != nil {
		logError(ctx, "Error encoding JSON response for percentiles status", "dataset_id", sanitizeForLogging(datasetID), "error", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// getStatusMessage returns a human-readable message for the percentile status
func getStatusMessage(status string, validPercent float64) string {
	switch status {
	case "ready":
		return "Performance percentiles are ready"
	case "partial":
		return fmt.Sprintf("Performance percentiles are %.1f%% ready", validPercent)
	case "calculating":
		return "Performance percentiles are being calculated"
	default:
		return "Performance percentiles are not available"
	}
}

// TeamRatings holds the calculated ratings for a team
type TeamRatings struct {
	BestOverall int
	AttRating   int
	MidRating   int
	DefRating   int
}

// calculateTeamRatings calculates the overall and section ratings for a team
// using formation-based optimization similar to the frontend approach
func calculateTeamRatings(players []Player) TeamRatings {
	if len(players) < 11 {
		return TeamRatings{}
	}

	// Define common formations to test (same as frontend)
	formations := []struct {
		name   string
		layout [][]string
	}{
		{"4-4-2", [][]string{
			{"GK"},
			{"D (R)", "D (C)", "D (C)", "D (L)"},
			{"M (R)", "M (C)", "M (C)", "M (L)"},
			{"ST (C)", "ST (C)"},
		}},
		{"4-3-3", [][]string{
			{"GK"},
			{"D (R)", "D (C)", "D (C)", "D (L)"},
			{"M (C)", "M (C)", "M (C)"},
			{"AM (R)", "ST (C)", "AM (L)"},
		}},
		{"4-2-3-1", [][]string{
			{"GK"},
			{"D (R)", "D (C)", "D (C)", "D (L)"},
			{"DM (C)", "DM (C)"},
			{"AM (R)", "AM (C)", "AM (L)"},
			{"ST (C)"},
		}},
		{"3-5-2", [][]string{
			{"GK"},
			{"D (C)", "D (C)", "D (C)"},
			{"WB (R)", "M (C)", "M (C)", "M (C)", "WB (L)"},
			{"ST (C)", "ST (C)"},
		}},
		{"4-1-4-1", [][]string{
			{"GK"},
			{"D (R)", "D (C)", "D (C)", "D (L)"},
			{"DM (C)"},
			{"M (R)", "M (C)", "M (C)", "M (L)"},
			{"ST (C)"},
		}},
	}

	// Position mappings (EXACT same as frontend)
	positionSideMap := map[string][]string{
		"D (R)":  {"DR"},
		"D (L)":  {"DL"},
		"D (C)":  {"DC"},
		"WB (R)": {"WBR"},
		"WB (L)": {"WBL"},
		"DM (C)": {"DM"},
		"M (R)":  {"MR"},
		"M (L)":  {"ML"},
		"M (C)":  {"MC"},
		"AM (R)": {"AMR"},
		"AM (L)": {"AML"},
		"AM (C)": {"AMC"},
		"ST (C)": {"ST"},
		"GK":     {"GK"},
	}

	fallbackPositionMap := map[string][]string{
		"D (R)":  {"DR", "WBR", "MR"},
		"D (L)":  {"DL", "WBL", "ML"},
		"D (C)":  {"DC", "DM"},
		"WB (R)": {"WBR", "DR", "MR"},
		"WB (L)": {"WBL", "DL", "ML"},
		"DM (C)": {"DM", "DC", "MC"},
		"M (R)":  {"MR", "WBR", "AMR"},
		"M (L)":  {"ML", "WBL", "AML"},
		"M (C)":  {"MC", "DM"},
		"AM (R)": {"AMR", "MR"},
		"AM (L)": {"AML", "ML"},
		"AM (C)": {"AMC", "MC"},
		"ST (C)": {"ST", "AMC"},
		"GK":     {"GK"},
	}

	// FM slot role matcher (EXACT same as frontend)
	fmSlotRoleMatcher := map[string][]string{
		"GK":     {"Goalkeeper"},
		"D (R)":  {"Defender (Right)", "Right Back"},
		"D (L)":  {"Defender (Left)", "Left Back"},
		"D (C)":  {"Defender (Centre)", "Centre Back"},
		"WB (R)": {"Wing-Back (Right)", "Right Wing-Back"},
		"WB (L)": {"Wing-Back (Left)", "Left Wing-Back"},
		"DM (C)": {"Defensive Midfielder (Centre)", "Centre Defensive Midfielder"},
		"M (R)":  {"Midfielder (Right)", "Right Midfielder"},
		"M (L)":  {"Midfielder (Left)", "Left Midfielder"},
		"M (C)":  {"Midfielder (Centre)", "Centre Midfielder"},
		"AM (R)": {"Attacking Midfielder (Right)", "Right Attacking Midfielder", "Winger (Right)"},
		"AM (L)": {"Attacking Midfielder (Left)", "Left Attacking Midfielder", "Winger (Left)"},
		"AM (C)": {"Attacking Midfielder (Centre)", "Centre Attacking Midfielder"},
		"ST (C)": {"Striker (Centre)", "Striker"},
	}

	fmMatcherToRoleKeyPrefix := map[string]string{
		"GOALKEEPER":                    "GK",
		"SWEEPER":                       "DC",
		"DEFENDER (RIGHT)":              "DR",
		"RIGHT BACK":                    "DR",
		"DEFENDER (LEFT)":               "DL",
		"LEFT BACK":                     "DL",
		"DEFENDER (CENTRE)":             "DC",
		"CENTRE BACK":                   "DC",
		"WING-BACK (RIGHT)":             "WBR",
		"RIGHT WING-BACK":               "WBR",
		"WING-BACK (LEFT)":              "WBL",
		"LEFT WING-BACK":                "WBL",
		"DEFENSIVE MIDFIELDER (CENTRE)": "DM",
		"CENTRE DEFENSIVE MIDFIELDER":   "DM",
		"MIDFIELDER (RIGHT)":            "MR",
		"RIGHT MIDFIELDER":              "MR",
		"MIDFIELDER (LEFT)":             "ML",
		"LEFT MIDFIELDER":               "ML",
		"MIDFIELDER (CENTRE)":           "MC",
		"CENTRE MIDFIELDER":             "MC",
		"ATTACKING MIDFIELDER (RIGHT)":  "AMR",
		"RIGHT ATTACKING MIDFIELDER":    "AMR",
		"WINGER (RIGHT)":                "AMR",
		"ATTACKING MIDFIELDER (LEFT)":   "AML",
		"LEFT ATTACKING MIDFIELDER":     "AML",
		"WINGER (LEFT)":                 "AML",
		"ATTACKING MIDFIELDER (CENTRE)": "AMC",
		"CENTRE ATTACKING MIDFIELDER":   "AMC",
		"STRIKER (CENTRE)":              "ST",
		"STRIKER":                       "ST",
	}

	// Section position definitions (same as frontend)
	attackingPositions := []string{"AM (R)", "AM (L)", "ST (C)"}
	midfielderPositions := []string{"DM (C)", "M (R)", "M (L)", "M (C)", "AM (C)"}
	defensivePositions := []string{"GK", "D (R)", "D (L)", "D (C)", "WB (R)", "WB (L)"}

	const MIN_SUITABILITY_THRESHOLD = 10

	var bestOverall int
	var bestSectionRatings TeamRatings

	// Test each formation
	for _, formation := range formations {
		formationSlots := []string{}
		for _, row := range formation.layout {
			formationSlots = append(formationSlots, row...)
		}

		// Calculate best team for this formation
		teamComposition := make(map[string]Player)
		usedPlayers := make(map[string]bool)

		// Fill each position with the best available player (same logic as frontend)
		for _, slot := range formationSlots {
			var bestPlayer Player
			var bestRating int

			for _, player := range players {
				if usedPlayers[player.Name] {
					continue
				}

				// Get position-specific rating for this player in this role (EXACT same as frontend)
				rating := getPlayerOverallForRoleGo(player, slot, positionSideMap, fallbackPositionMap, fmSlotRoleMatcher, fmMatcherToRoleKeyPrefix)

				if rating >= MIN_SUITABILITY_THRESHOLD {
					// Check if player can play this position (EXACT same as frontend)
					slotPositions := positionSideMap[slot] // Note: frontend uses toUpperCase()
					fallbackPositions := fallbackPositionMap[slot]
					playerPositions := player.ShortPositions

					isExactMatch := false
					isFallbackMatch := false

					for _, pos := range playerPositions {
						for _, slotPos := range slotPositions {
							if pos == slotPos {
								isExactMatch = true
								break
							}
						}
						for _, fallbackPos := range fallbackPositions {
							if pos == fallbackPos {
								isFallbackMatch = true
								break
							}
						}
					}

					if isExactMatch || isFallbackMatch {
						// Sort score logic (same as frontend)
						sortScore := rating
						if isExactMatch {
							sortScore += 10000
						} else {
							sortScore -= 5000
						}

						if sortScore > bestRating {
							bestRating = sortScore
							bestPlayer = player
						}
					}
				}
			}

			if bestPlayer.Name != "" {
				teamComposition[slot] = bestPlayer
				usedPlayers[bestPlayer.Name] = true
			}
		}

		// Calculate average overall for this formation
		var totalRating int
		var playerCount int
		var attSum, midSum, defSum int
		var attCount, midCount, defCount int

		for slot, player := range teamComposition {
			rating := getPlayerOverallForRoleGo(player, slot, positionSideMap, fallbackPositionMap, fmSlotRoleMatcher, fmMatcherToRoleKeyPrefix)
			totalRating += rating
			playerCount++

			// Categorize by position (same as frontend)
			if contains(attackingPositions, slot) {
				attSum += rating
				attCount++
			} else if contains(midfielderPositions, slot) {
				midSum += rating
				midCount++
			} else if contains(defensivePositions, slot) {
				defSum += rating
				defCount++
			}
		}

		if playerCount >= 5 { // Minimum viable team
			averageOverall := totalRating / playerCount
			if averageOverall > bestOverall {
				bestOverall = averageOverall
				// Calculate section ratings
				attRating := 0
				if attCount > 0 {
					attRating = attSum / attCount
				}
				midRating := 0
				if midCount > 0 {
					midRating = midSum / midCount
				}
				defRating := 0
				if defCount > 0 {
					defRating = defSum / defCount
				}

				bestSectionRatings = TeamRatings{
					BestOverall: averageOverall,
					AttRating:   attRating,
					MidRating:   midRating,
					DefRating:   defRating,
				}
			}
		}
	}

	return bestSectionRatings
}

// getPlayerOverallForRoleGo calculates the rating for a player in a specific role (EXACT same logic as frontend)
func getPlayerOverallForRoleGo(player Player, role string, positionSideMap, fallbackPositionMap, fmSlotRoleMatcher map[string][]string, fmMatcherToRoleKeyPrefix map[string]string) int {
	if player.Name == "" || role == "" {
		return 0
	}

	bestScoreForRole := 0

	// Check if player has role-specific overalls
	if len(player.RoleSpecificOveralls) == 0 {
		// If no role-specific overalls, use player's general Overall as fallback
		return player.Overall
	}

	// Check if player has role-specific overalls (same logic as frontend)
	hasRoleOveralls := len(player.RoleSpecificOveralls) > 0
	if !hasRoleOveralls {
		// If no role-specific overalls, use player's general Overall as fallback
		return player.Overall
	}

	// Get required positions for this role (same as frontend)
	upperSlotRole := strings.ToUpper(role)
	requiredPositions := []string{}
	if positions, exists := positionSideMap[upperSlotRole]; exists {
		requiredPositions = positions
	}

	// Check exact position matches first (same as frontend)
	if player.ShortPositions != nil && len(player.ShortPositions) > 0 {
		exactPositionMatches := []string{}
		for _, pos := range player.ShortPositions {
			for _, requiredPos := range requiredPositions {
				if pos == requiredPos {
					exactPositionMatches = append(exactPositionMatches, pos)
				}
			}
		}

		if len(exactPositionMatches) > 0 {
			// Find best role-specific rating for exact matches
			for _, rso := range player.RoleSpecificOveralls {
				rsoBasePosition := strings.Split(rso.RoleName, " - ")[0]
				for _, exactMatch := range exactPositionMatches {
					if rsoBasePosition == exactMatch {
						if rso.Score > bestScoreForRole {
							bestScoreForRole = rso.Score
						}
					}
				}
			}

			if bestScoreForRole > 0 {
				return bestScoreForRole
			}
		}
	}

	// Fallback logic (same as frontend)
	if bestScoreForRole == 0 {
		upperSlotRole := strings.ToUpper(role)
		fmPositionMatchers := []string{}
		if matchers, exists := fmSlotRoleMatcher[upperSlotRole]; exists {
			fmPositionMatchers = matchers
		}
		if len(fmPositionMatchers) == 0 {
			fmPositionMatchers = []string{upperSlotRole}
		}

		// Build target role key prefixes (same as frontend)
		targetRoleKeyPrefixes := []string{}
		for _, matcher := range fmPositionMatchers {
			if prefix, exists := fmMatcherToRoleKeyPrefix[strings.ToUpper(matcher)]; exists {
				// Check if not already included
				found := false
				for _, existing := range targetRoleKeyPrefixes {
					if existing == prefix {
						found = true
						break
					}
				}
				if !found {
					targetRoleKeyPrefixes = append(targetRoleKeyPrefixes, prefix)
				}
			}
		}

		// Check role-specific overalls against target prefixes
		for _, rso := range player.RoleSpecificOveralls {
			rsoBasePosition := strings.Split(rso.RoleName, " - ")[0]
			for _, targetPrefix := range targetRoleKeyPrefixes {
				if rsoBasePosition == targetPrefix {
					if rso.Score > bestScoreForRole {
						bestScoreForRole = rso.Score
					}
				}
			}
		}

		// Final fallback to player's general Overall rating if still nothing found
		if bestScoreForRole == 0 {
			fallbackOverall := player.Overall - 10
			if fallbackOverall < 0 {
				fallbackOverall = 0
			}
			bestScoreForRole = fallbackOverall
		}
	}

	return bestScoreForRole
}

// calculateAverage calculates the average of a slice of integers
func calculateAverage(numbers []int) int {
	if len(numbers) == 0 {
		return 0
	}

	sum := 0
	for _, num := range numbers {
		sum += num
	}

	return sum / len(numbers)
}

// SearchResult represents a search result item
type SearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`        // "player", "team", "league", "nation"
	Description string `json:"description"` // Additional context (e.g., team/division for player)
	URL         string `json:"url"`         // URL to navigate to
	Overall     int    `json:"overall"`     // Include overall rating for sorting
}

// searchHandler handles GET requests for searching players, teams, leagues, and nations
func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start tracing
	ctx, span := StartSpan(ctx, "search.handler")
	defer span.End()

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/search"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Track active requests
	IncrementActiveRequests(ctx, "/search")
	defer DecrementActiveRequests(ctx, "/search")

	// Extract dataset ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[0] != "api" || pathParts[1] != "search" {
		WriteErrorResponse(w, r, "invalid_url", "Invalid URL format. Expected /api/search/{datasetId}", nil, http.StatusBadRequest)
		return
	}

	datasetID := pathParts[2]
	if datasetID == "" {
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is required", nil, http.StatusBadRequest)
		return
	}

	// Get search query from URL parameters
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		// Return empty results for empty query
		setCORSHeaders(w, r)

		// Create response metadata
		metadata := CreateResponseMetadata(requestID, 0, false)

		if supportsProtobuf {
			// Create empty protobuf response
			protoResponse := &pb.SearchResponse{
				Players:  []*pb.Player{},
				Query:    "",
				Metadata: metadata,
			}

			// Serialize to protobuf
			responseBytes, err := serializer.Serialize(protoResponse)
			if err == nil {
				// Protobuf serialization successful
				w.Header().Set("Content-Type", serializer.ContentType())
				if serializer.ShouldCompress() {
					w.Header().Set("Content-Encoding", "gzip")
				}

				if _, writeErr := w.Write(responseBytes); writeErr != nil {
					logError(ctx, "Error writing protobuf response", "error", writeErr)
				}

				logDebug(ctx, "Empty search results served as protobuf",
					"response_size_bytes", len(responseBytes),
					"processing_time_ms", time.Since(startTime).Milliseconds())
				return
			}

			// Log protobuf serialization failure
			logWarn(ctx, "Protobuf serialization failed for empty search results, falling back to JSON", "error", err)
		}

		// Fallback to JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]SearchResult{}); err != nil {
			logError(ctx, "Error encoding empty search results", "error", err)
			WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		}
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("search.query", query),
	)

	// Get player data
	// Quick dataset existence check (no expensive operations)
	_, _, found := GetPlayerDataForIndexing(datasetID)
	if !found {
		WriteErrorResponse(w, r, "dataset_not_found", "Dataset not found", nil, http.StatusNotFound)
		return
	}

	logDebug(ctx, "Performing search", "dataset_id", datasetID, "query", query)

	// NEW: Generate cache key (lightweight)
	cacheKey := generateSearchCacheKeyLightweight(ctx, datasetID, query)

	// Try to load from cache
	if cachedResults, found := loadSearchFromCacheLightweight(ctx, cacheKey, datasetID, query); found {
		logInfo(ctx, "Returning cached search results",
			"dataset_id", datasetID,
			"query", query,
			"cache_key", cacheKey,
			"result_count", len(cachedResults))

		SetSpanAttributes(ctx,
			attribute.Int("search.results_count", len(cachedResults)),
			attribute.String("search.cache_status", "HIT"))

		// Set CORS headers
		setCORSHeaders(w, r)

		// Create response metadata
		metadata := CreateResponseMetadata(requestID, safeInt32(len(cachedResults)), true)

		if supportsProtobuf {
			// Create protobuf response
			// Note: For search results we only include minimal player data in protobuf format
			protoPlayers := make([]*pb.Player, 0, len(cachedResults))
			for _, result := range cachedResults {
				if result.Type == "player" {
					// Create a minimal player representation for search results
					protoPlayer := &pb.Player{
						Uid:      int64(0), // We don't have UID in search results
						Name:     result.Name,
						Position: result.Description,
						Overall:  int32(result.Overall),
					}
					protoPlayers = append(protoPlayers, protoPlayer)
				}
			}

			protoResponse := &pb.SearchResponse{
				Players:  protoPlayers,
				Query:    query,
				Metadata: metadata,
			}

			// Serialize to protobuf
			responseBytes, err := serializer.Serialize(protoResponse)
			if err == nil {
				// Protobuf serialization successful
				w.Header().Set("Content-Type", serializer.ContentType())
				w.Header().Set("X-Cache-Status", "HIT")
				if serializer.ShouldCompress() {
					w.Header().Set("Content-Encoding", "gzip")
				}

				if _, writeErr := w.Write(responseBytes); writeErr != nil {
					logError(ctx, "Error writing protobuf response", "error", writeErr)
				}

				logDebug(ctx, "Search results served as protobuf from cache",
					"result_count", len(cachedResults),
					"response_size_bytes", len(responseBytes),
					"processing_time_ms", time.Since(startTime).Milliseconds())
				return
			}

			// Log protobuf serialization failure
			logWarn(ctx, "Protobuf serialization failed for cached search results, falling back to JSON", "error", err)
		}

		// Fallback to JSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Status", "HIT")
		if err := json.NewEncoder(w).Encode(cachedResults); err != nil {
			RecordError(ctx, err, "Failed to encode cached search results")
			WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
			return
		}
		return
	}

	// Cache miss - perform search
	logInfo(ctx, "Cache miss, performing search",
		"dataset_id", datasetID,
		"query", query,
		"cache_key", cacheKey)

	// Perform optimized search
	results, searchErr := GetHybridSearchService().Search(ctx, datasetID, query, 100)
	if searchErr != nil {
		logError(ctx, "Error performing optimized search", "error", searchErr)
		WriteErrorResponse(w, r, "search_error", "Search failed", nil, http.StatusInternalServerError)
		return
	}

	// NEW: Save to cache for future requests (only cache if results are not too large)
	if len(results) <= 1000 { // Reasonable limit to avoid caching huge result sets
		go func() {
			saveSearchToCacheLightweight(ctx, cacheKey, datasetID, query, results)
		}()
	}

	SetSpanAttributes(ctx,
		attribute.Int("search.results_count", len(results)),
		attribute.String("search.cache_status", "MISS"))

	// Set CORS headers
	setCORSHeaders(w, r)

	// Create response metadata
	metadata := CreateResponseMetadata(requestID, safeInt32(len(results)), false)

	if supportsProtobuf {
		// Create protobuf response
		// Note: For search results we only include minimal player data in protobuf format
		protoPlayers := make([]*pb.Player, 0, len(results))
		for _, result := range results {
			if result.Type == "player" {
				// Create a minimal player representation for search results
				protoPlayer := &pb.Player{
					Uid:      int64(0), // We don't have UID in search results
					Name:     result.Name,
					Position: result.Description,
					Overall:  safeInt32(result.Overall),
				}
				protoPlayers = append(protoPlayers, protoPlayer)
			}
		}

		protoResponse := &pb.SearchResponse{
			Players:  protoPlayers,
			Query:    query,
			Metadata: metadata,
		}

		// Serialize to protobuf
		responseBytes, err := serializer.Serialize(protoResponse)
		if err == nil {
			// Protobuf serialization successful
			w.Header().Set("Content-Type", serializer.ContentType())
			w.Header().Set("X-Cache-Status", "MISS")
			if serializer.ShouldCompress() {
				w.Header().Set("Content-Encoding", "gzip")
			}

			if _, writeErr := w.Write(responseBytes); writeErr != nil {
				logError(ctx, "Error writing protobuf response", "error", writeErr)
			}

			logDebug(ctx, "Search results served as protobuf",
				"result_count", len(results),
				"response_size_bytes", len(responseBytes),
				"processing_time_ms", time.Since(startTime).Milliseconds())
			return
		}

		// Log protobuf serialization failure
		logWarn(ctx, "Protobuf serialization failed for search results, falling back to JSON", "error", err)
	}

	// Fallback to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		RecordError(ctx, err, "Failed to encode search results")
		WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		return
	}

	logInfo(ctx, "Search completed", "results_count", len(results))
}

// performSearch searches through players, teams, leagues, and nations
func performSearch(players []Player, query string) []SearchResult {
	// Pre-allocate with reasonable capacity for search results
	estimatedResults := 100 // Most searches return < 100 results
	if len(players) > 5000 {
		estimatedResults = 200 // Larger datasets might have more matches
	}
	results := make([]SearchResult, 0, estimatedResults)
	queryLower := strings.ToLower(query)

	// Collect unique teams, leagues, and nations with estimated capacities
	estimatedTeams := len(players) / 25    // Rough estimate: ~25 players per team
	estimatedLeagues := len(players) / 100 // Rough estimate: ~100 players per league
	estimatedNations := len(players) / 50  // Rough estimate: ~50 players per nation

	teams := make(map[string]struct {
		division string
		players  int
	}, estimatedTeams)
	leagues := make(map[string]int, estimatedLeagues) // league -> player count
	nations := make(map[string]int, estimatedNations) // nation -> player count

	// Search players and collect team/league/nation data
	for i := range players {
		player := &players[i]
		// Search players by name only
		if strings.Contains(strings.ToLower(player.Name), queryLower) {
			results = append(results, SearchResult{
				ID:          player.Name, // Using name as ID for players
				Name:        player.Name,
				Type:        "player",
				Description: fmt.Sprintf("%s • %s (%d OVR)", player.Club, player.Division, player.Overall),
				URL:         fmt.Sprintf("/dataset/%s?search=%s", "", player.Name), // Frontend will fill dataset ID
				Overall:     player.Overall,                                        // Include overall rating for sorting
			})
		}

		// Collect team data
		if player.Club != "" {
			if _, exists := teams[player.Club]; !exists {
				teams[player.Club] = struct {
					division string
					players  int
				}{division: player.Division, players: 0}
			}
			teamData := teams[player.Club]
			teamData.players++
			teams[player.Club] = teamData
		}

		// Collect league data
		if player.Division != "" {
			leagues[player.Division]++
		}

		// Collect nation data
		if player.Nationality != "" {
			nations[player.Nationality]++
		}
	}

	// Search teams
	for teamName, teamData := range teams {
		if strings.Contains(strings.ToLower(teamName), queryLower) {
			results = append(results, SearchResult{
				ID:          teamName,
				Name:        teamName,
				Type:        "team",
				Description: fmt.Sprintf("%s • %d players", teamData.division, teamData.players),
				URL:         fmt.Sprintf("/dataset/%s?team=%s", "", teamName), // Frontend will fill dataset ID
			})
		}
	}

	// Search leagues
	for leagueName, playerCount := range leagues {
		if strings.Contains(strings.ToLower(leagueName), queryLower) {
			results = append(results, SearchResult{
				ID:          leagueName,
				Name:        leagueName,
				Type:        "league",
				Description: fmt.Sprintf("%d players", playerCount),
				URL:         fmt.Sprintf("/leagues?league=%s", url.QueryEscape(leagueName)),
			})
		}
	}

	// Search nations - more flexible matching
	for nationName, playerCount := range nations {
		nationLower := strings.ToLower(nationName)
		// Direct substring match
		directMatch := strings.Contains(nationLower, queryLower)

		// Also check common nationality variations
		var variationMatch bool
		switch queryLower {
		case "fra", "france":
			variationMatch = strings.Contains(nationLower, "fran") || strings.Contains(nationLower, "french")
		case "eng", "england":
			variationMatch = strings.Contains(nationLower, "eng") || strings.Contains(nationLower, "british")
		case "ger", "germany":
			variationMatch = strings.Contains(nationLower, "ger") || strings.Contains(nationLower, "deutsch")
		case "spa", "spain":
			variationMatch = strings.Contains(nationLower, "spa") || strings.Contains(nationLower, "spanish")
		case "ita", "italy":
			variationMatch = strings.Contains(nationLower, "ita") || strings.Contains(nationLower, "italian")
		case "por", "portugal":
			variationMatch = strings.Contains(nationLower, "por") || strings.Contains(nationLower, "portuguese")
		case "bra", "brazil":
			variationMatch = strings.Contains(nationLower, "bra") || strings.Contains(nationLower, "brazilian")
		case "arg", "argentina":
			variationMatch = strings.Contains(nationLower, "arg") || strings.Contains(nationLower, "argentine")
		case "net", "netherlands":
			variationMatch = strings.Contains(nationLower, "net") || strings.Contains(nationLower, "dutch")
		case "bel", "belgium":
			variationMatch = strings.Contains(nationLower, "bel") || strings.Contains(nationLower, "belgian")
		default:
			// For any 3-letter query, also try matching the first 3 letters of nation names
			if len(queryLower) == 3 {
				if len(nationLower) >= 3 {
					variationMatch = strings.HasPrefix(nationLower, queryLower)
				}
			}
		}

		if directMatch || variationMatch {
			results = append(results, SearchResult{
				ID:          nationName,
				Name:        nationName,
				Type:        "nation",
				Description: fmt.Sprintf("%d players", playerCount),
				URL:         fmt.Sprintf("/nations?nation=%s", url.QueryEscape(nationName)),
			})
		}
	}

	// Sort results by the new priority order: nations, leagues, teams, then players by highest overall
	sort.Slice(results, func(i, j int) bool {
		// Define new type priority: nations (1), leagues (2), teams (3), players (4)
		typePriority := map[string]int{"nation": 1, "league": 2, "team": 3, "player": 4}

		// First sort by type priority
		if typePriority[results[i].Type] != typePriority[results[j].Type] {
			return typePriority[results[i].Type] < typePriority[results[j].Type]
		}

		// For players, sort by highest overall rating first
		if results[i].Type == "player" && results[j].Type == "player" {
			if results[i].Overall != results[j].Overall {
				return results[i].Overall > results[j].Overall // Highest overall first
			}
		}

		// For non-players or same overall rating, sort alphabetically by name
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	// Limit to top 500 results for better performance and visualization
	maxResults := 500
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// BargainHunterRequest represents the request body for bargain hunter analysis
type BargainHunterRequest struct {
	MaxBudget  int64 `json:"maxBudget"`
	MaxSalary  int64 `json:"maxSalary"`
	MinAge     int   `json:"minAge"`
	MaxAge     int   `json:"maxAge"`
	MinOverall int   `json:"minOverall"`
}

// BargainHunterResponse represents a player with calculated value score
type BargainHunterResponse struct {
	Player     Player  `json:"player"`
	ValueScore float64 `json:"valueScore"`
}

// bargainHunterHandler handles POST requests to find the best value players within budget constraints
func bargainHunterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/bargain-hunter/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	var req BargainHunterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	logInfo(ctx, "Processing bargain hunter request",
		"dataset_id", datasetID,
		"max_budget", req.MaxBudget,
		"max_salary", req.MaxSalary,
		"min_age", req.MinAge,
		"max_age", req.MaxAge,
		"min_overall", req.MinOverall)

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Recalculate all player ratings based on the current calculation method setting
	players = RecalculateAllPlayersRatings(players)

	// NEW: Generate cache key and try to load from cache first
	cacheKey := generateBargainHunterCacheKey(ctx, datasetID, req.MaxBudget, req.MaxSalary, req.MinAge, req.MaxAge, req.MinOverall, players)

	// Try to load from cache
	if cachedResults, found := loadBargainHunterFromCache(ctx, cacheKey, datasetID, req.MaxBudget, req.MaxSalary, req.MinAge, req.MaxAge, req.MinOverall, players); found {
		logInfo(ctx, "Returning cached bargain hunter results",
			"dataset_id", datasetID,
			"cache_key", cacheKey,
			"result_count", len(cachedResults))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Status", "HIT")
		setCORSHeaders(w, r)
		if err := json.NewEncoder(w).Encode(cachedResults); err != nil {
			logError(ctx, "Error encoding JSON response for cached bargain hunter", "dataset_id", sanitizeForLogging(datasetID), "error", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - perform calculation
	logInfo(ctx, "Cache miss, calculating bargain hunter results",
		"dataset_id", datasetID,
		"cache_key", cacheKey)

	// Process bargain hunter analysis
	bargainPlayers := processBargainHunter(players, req.MaxBudget, req.MaxSalary, int64(req.MinAge), int64(req.MaxAge), int64(req.MinOverall))

	// NEW: Save to cache for future requests
	go func() {
		saveBargainHunterToCache(ctx, cacheKey, datasetID, req.MaxBudget, req.MaxSalary, req.MinAge, req.MaxAge, req.MinOverall, players, bargainPlayers)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(bargainPlayers); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		logError(ctx, "Error encoding JSON response for bargain hunter", "dataset_id", sanitizeForLogging(datasetID), "error", err)
	}
}

// getExpectedValuePerRating returns the expected value-per-rating ratio (millions per overall point)
// for players of different overall ratings, based on typical market pricing
func getExpectedValuePerRating(overall float64) float64 {
	// These values are based on typical FM market pricing patterns
	// Higher rated players typically cost more per rating point due to scarcity
	switch {
	case overall >= 85:
		return 1.2 // Elite players: ~£1.2m per overall point
	case overall >= 80:
		return 0.8 // World class players: ~£0.8m per overall point
	case overall >= 75:
		return 0.5 // Quality players: ~£0.5m per overall point
	case overall >= 70:
		return 0.3 // Good players: ~£0.3m per overall point
	case overall >= 65:
		return 0.2 // Decent players: ~£0.2m per overall point
	case overall >= 60:
		return 0.15 // Average players: ~£0.15m per overall point
	default:
		return 0.1 // Below average players: ~£0.1m per overall point
	}
}

// processBargainHunter calculates value scores and filters players by budget constraints
func processBargainHunter(players []Player, maxBudget, maxSalary, minAge, maxAge, minOverall int64) []BargainHunterResponse {
	// Pre-allocate with estimated capacity (typically 10-20% of players match criteria)
	estimatedResults := len(players) / 8 // Estimate ~12.5% match
	if estimatedResults < 20 {
		estimatedResults = 20
	}
	if estimatedResults > 500 {
		estimatedResults = 500 // Cap at max results limit
	}
	results := make([]BargainHunterResponse, 0, estimatedResults)

	for i := range players {
		player := players[i]

		// Skip free transfers and "Not for Sale" players entirely
		if player.TransferValueAmount == 0 ||
			player.TransferValue == "Not for Sale" ||
			strings.Contains(strings.ToLower(player.TransferValue), "not for sale") {
			continue
		}

		// Skip players outside budget constraints
		if maxBudget > 0 && player.TransferValueAmount > maxBudget {
			continue
		}
		if maxSalary > 0 && player.WageAmount > maxSalary {
			continue
		}

		// Skip players outside age constraints
		if minAge > 0 || maxAge > 0 {
			playerAge, ageErr := strconv.Atoi(player.Age)
			if ageErr != nil {
				continue // Skip players with unparseable age
			}
			if minAge > 0 && int64(playerAge) < minAge {
				continue
			}
			if maxAge > 0 && int64(playerAge) > maxAge {
				continue
			}
		}

		// Skip players below minimum overall
		if minOverall > 0 && int64(player.Overall) < minOverall {
			continue
		}

		// Calculate value score using improved algorithm
		var valueScore float64
		overall := float64(player.Overall)
		transferValueMillions := float64(player.TransferValueAmount) / 1000000.0

		// Prevent division by zero
		if transferValueMillions == 0 {
			continue
		}

		// Calculate value-per-rating ratio (millions per overall point)
		valuePerRating := transferValueMillions / overall

		// Use logarithmic scaling to reduce the penalty for expensive but valuable players
		// This helps ensure that an 85-rated £50m player can compete with a 75-rated £5m player
		logValuePerRating := math.Log10(valuePerRating + 1) // +1 to avoid log(0)

		// Base efficiency score: higher overall rating and lower value-per-rating is better
		baseEfficiency := overall / (logValuePerRating + 1)

		// Apply tier-based multipliers to maintain some differentiation
		switch {
		case overall >= 80:
			// Elite players (80+) - Expect premium pricing, moderate penalty for cost
			valueScore = baseEfficiency * 1.2
		case overall >= 70:
			// Quality players (70-79) - Good balance of quality and value
			valueScore = baseEfficiency * 1.0
		case overall >= 60:
			// Decent players (60-69) - Should be better value for money
			valueScore = baseEfficiency * 0.9
		case overall >= 55:
			// Budget players (55-59) - Expected to be cheap
			valueScore = baseEfficiency * 0.8
		default:
			// Youth/development players (<55) - Penalized for poor current ability
			valueScore = baseEfficiency * 0.6
		}

		// Apply bonus for exceptional value scenarios
		// If a player's value-per-rating is significantly below their tier average
		expectedValuePerRating := getExpectedValuePerRating(overall)
		if valuePerRating < expectedValuePerRating*0.7 { // 30% below expected
			valueScore *= 1.3 // 30% bonus for exceptional value
		} else if valuePerRating < expectedValuePerRating*0.85 { // 15% below expected
			valueScore *= 1.15 // 15% bonus for good value
		}

		results = append(results, BargainHunterResponse{
			Player:     player,
			ValueScore: valueScore,
		})
	}

	// Sort by value score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ValueScore > results[j].ValueScore
	})

	// Normalize value scores so highest = 100, lowest = 0
	if len(results) > 1 {
		// Find min and max scores
		maxScore := results[0].ValueScore              // Already sorted, so first is highest
		minScore := results[len(results)-1].ValueScore // Last is lowest

		// Apply normalization formula: ((score - min) / (max - min)) * 100
		if maxScore != minScore { // Avoid division by zero
			for i := range results {
				normalized := ((results[i].ValueScore - minScore) / (maxScore - minScore)) * 100
				results[i].ValueScore = normalized
			}
		} else {
			// If all scores are the same, set them all to 100
			for i := range results {
				results[i].ValueScore = 100
			}
		}
	} else if len(results) == 1 {
		// Single result gets score of 100
		results[0].ValueScore = 100
	}

	// Limit to top 500 results for better performance and visualization
	maxResults := 500
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// facesHandler serves player face images from external API, S3 or local storage
func facesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "handlers.faces")
	defer span.End()

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract UID from query parameters
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		logWarn(ctx, "Missing uid parameter in faces request")
		http.Error(w, "Missing 'uid' parameter", http.StatusBadRequest)
		return
	}

	// Validate UID format for security (prevent path injection)
	if err := validateID(uid, 100); err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidUID(sanitizeForLogging(uid), err), "Invalid UID format")
		http.Error(w, "Invalid UID format", http.StatusBadRequest)
		return
	}

	logDebug(ctx, "Processing face image request", "uid", sanitizeForLogging(uid))

	// Check if IMAGE_API_URL is configured - if so, redirect to external API
	if imageAPIURL := os.Getenv("IMAGE_API_URL"); imageAPIURL != "" {
		externalURL := fmt.Sprintf("%s/face/%s.png?width=256", imageAPIURL, uid)
		logInfo(ctx, "Redirecting to external image API", "url", externalURL)

		// Set appropriate headers for redirect
		setCORSHeaders(w, r)
		http.Redirect(w, r, externalURL, http.StatusFound)
		return
	}

	// Set appropriate headers for image response
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	setCORSHeaders(w, r)

	// Construct the face image filename
	faceFileName := uid + ".png"

	// Try S3 first if configured
	if s3Storage, ok := storage.(*S3Storage); ok && s3Storage.client != nil {
		logDebug(ctx, "Attempting to retrieve face from S3", "filename", sanitizeForLogging(faceFileName))

		// Get face image from S3
		if err := s3Storage.getFaceImage(ctx, faceFileName, w); err != nil {
			logWarn(ctx, "Failed to retrieve face from S3", "filename", sanitizeForLogging(faceFileName), "error", err)
			// Fall through to local storage
		} else {
			logInfo(ctx, "Successfully served face from S3", "filename", sanitizeForLogging(faceFileName))
			return
		}
	}

	// Try local storage as fallback
	facesDir := getFacesDirectory()

	// Safely construct the file path to prevent path injection
	faceFilePath, err := validateAndJoinPath(facesDir, faceFileName)
	if err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidFilePath(sanitizeForLogging(uid), err), "Path validation failed")
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	logDebug(ctx, "Attempting to retrieve face from local storage", "path", sanitizeForLogging(faceFilePath))

	// Check if file exists
	if _, err := os.Stat(faceFilePath); os.IsNotExist(err) {
		logWarn(ctx, "Face image not found", "path", sanitizeForLogging(faceFilePath))
		http.Error(w, "Face image not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, faceFilePath)
	logInfo(ctx, "Successfully served face from local storage", "path", sanitizeForLogging(faceFilePath))
}

// getFacesDirectory returns the directory path for local face storage
func getFacesDirectory() string {
	facesDir := os.Getenv("FACES_DIR")
	if facesDir == "" {
		facesDir = "./faces" // Default to "./faces" directory
	}
	return facesDir
}

// logosHandler serves team logo images from external API, S3 or local storage
func logosHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "handlers.logos")
	defer span.End()

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from query parameters
	teamID := r.URL.Query().Get("teamId")
	if teamID == "" {
		logWarn(ctx, "Missing teamId parameter in logos request")
		http.Error(w, "Missing 'teamId' parameter", http.StatusBadRequest)
		return
	}

	// Validate team ID format for security (prevent path injection)
	if err := validateID(teamID, 100); err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidTeamID(sanitizeForLogging(teamID), err), "Invalid team ID format")
		http.Error(w, "Invalid team ID format", http.StatusBadRequest)
		return
	}

	logDebug(ctx, "Processing team logo request", "teamId", sanitizeForLogging(teamID))

	// Check if IMAGE_API_URL is configured - if so, redirect to external API
	if imageAPIURL := os.Getenv("IMAGE_API_URL"); imageAPIURL != "" {
		externalURL := fmt.Sprintf("%s/team/%s.png?width=256", imageAPIURL, teamID)
		logInfo(ctx, "Redirecting to external image API", "url", externalURL)

		// Set appropriate headers for redirect
		setCORSHeaders(w, r)
		http.Redirect(w, r, externalURL, http.StatusFound)
		return
	}

	// Set appropriate headers for image response
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	setCORSHeaders(w, r)

	// Construct the logo image filename
	logoFileName := teamID + ".png"

	// Try S3 first if configured
	if s3Storage, ok := storage.(*S3Storage); ok && s3Storage.client != nil {
		logDebug(ctx, "Attempting to retrieve logo from S3", "filename", sanitizeForLogging(logoFileName))

		// Get logo image from S3
		if err := s3Storage.getTeamLogo(ctx, logoFileName, w); err != nil {
			logWarn(ctx, "Failed to retrieve logo from S3", "filename", sanitizeForLogging(logoFileName), "error", err)
			// Fall through to local storage
		} else {
			logDebug(ctx, "Successfully served logo from S3", "filename", sanitizeForLogging(logoFileName))
			return
		}
	}

	// Try local storage as fallback
	logosDir := getLogosDirectory()

	// Safely construct the file path to prevent path injection
	// Build the nested path: logosDir/Clubs/Normal/Normal/logoFileName
	clubsDir, err := validateAndJoinPath(logosDir, "Clubs")
	if err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidClubsDirPath(err), "Path validation failed")
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	normalDir1, err := validateAndJoinPath(clubsDir, "Normal")
	if err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidNormalDirPath(err), "Path validation failed")
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	normalDir2, err := validateAndJoinPath(normalDir1, "Normal")
	if err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidNormalDirPath(err), "Path validation failed")
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	logoFilePath, err := validateAndJoinPath(normalDir2, logoFileName)
	if err != nil {
		RecordError(ctx, apperrors.WrapErrInvalidFilePathForTeamID(sanitizeForLogging(teamID), err), "Path validation failed")
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	logInfo(ctx, "Attempting to retrieve logo from local storage", "path", sanitizeForLogging(logoFilePath))

	// Check if file exists
	if _, err := os.Stat(logoFilePath); os.IsNotExist(err) {
		logWarn(ctx, "Team logo not found", "path", sanitizeForLogging(logoFilePath))
		http.Error(w, "Team logo not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, logoFilePath)
	logInfo(ctx, "Successfully served logo from local storage", "path", sanitizeForLogging(logoFilePath))
}

// getLogosDirectory returns the directory path for local logo storage
func getLogosDirectory() string {
	logosDir := os.Getenv("LOGOS_DIR")
	if logosDir == "" {
		logosDir = "./logos" // Default to "./logos" directory
	}
	return logosDir
}

// cacheStatusHandler returns cache statistics and status
func cacheStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	cacheCount, totalItems := getMemCacheStats()

	status := map[string]interface{}{
		"memory_cache": map[string]interface{}{
			"items_count": cacheCount,
			"total_items": totalItems,
			"status":      "active",
		},
		"persistent_cache": map[string]interface{}{
			"version": cacheVersion,
			"types":   []string{"percentiles", "bargain_hunter", "search", "nation_ratings"},
		},
		"cache_config": map[string]interface{}{
			"default_ttl":      "5m",
			"cleanup_interval": "10m",
			"invalidation":     "automatic",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Error encoding cache status", http.StatusInternalServerError)
	}
}

// teamMatchHandler handles team name to ID matching
func teamMatchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := StartSpan(ctx, "handlers.teamMatch")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("name")
	if teamName == "" {
		logWarn(ctx, "Missing 'name' parameter in team match request")
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	logDebug(ctx, "Processing team match request", "teamName", teamName, "originalLength", len(teamName))

	// Get team match results
	matches := findTeamMatches(teamName)

	// Log detailed results for troubleshooting
	if len(matches) == 0 {
		logWarn(ctx, "No team matches found", "teamName", teamName)
	} else {
		logDebug(ctx, "Team matches found",
			"teamName", teamName,
			"matchCount", len(matches),
			"bestMatch", matches[0].Name,
			"bestScore", matches[0].Score,
			"bestID", matches[0].ID)

		// Log top 3 matches for debugging
		for i, match := range matches {
			if i >= 3 {
				break
			}
			logDebug(ctx, "Match result",
				"rank", i+1,
				"teamName", teamName,
				"matchName", match.Name,
				"matchID", match.ID,
				"score", match.Score)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	setCORSHeaders(w, r)

	if err := json.NewEncoder(w).Encode(matches); err != nil {
		logError(ctx, "Error encoding team match response", "error", err, "teamName", teamName)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// deepCopyPlayers creates a deep copy of the players slice including all nested maps
func deepCopyPlayers(players []Player) []Player {
	if players == nil {
		return nil
	}

	// CRITICAL: Use safe implementation instead of OptimizedDeepCopyPlayers
	// OptimizedDeepCopyPlayers has COW race conditions with concurrent access
	playersCopy := make([]Player, len(players))
	for i := range players {
		playersCopy[i] = players[i]

		// Deep copy PerformancePercentiles map
		if players[i].PerformancePercentiles != nil {
			playersCopy[i].PerformancePercentiles = make(map[string]map[string]float64)
			for group, stats := range players[i].PerformancePercentiles {
				playersCopy[i].PerformancePercentiles[group] = make(map[string]float64)
				for stat, value := range stats {
					playersCopy[i].PerformancePercentiles[group][stat] = value
				}
			}
		}

		// Deep copy NumericAttributes map - CRITICAL for race condition prevention
		if players[i].NumericAttributes != nil {
			playersCopy[i].NumericAttributes = make(map[string]int)
			for key, value := range players[i].NumericAttributes {
				playersCopy[i].NumericAttributes[key] = value
			}
		}

		// Deep copy Attributes map
		if players[i].Attributes != nil {
			playersCopy[i].Attributes = make(map[string]string)
			for key, value := range players[i].Attributes {
				playersCopy[i].Attributes[key] = value
			}
		}

		// Deep copy PerformanceStatsNumeric map
		if players[i].PerformanceStatsNumeric != nil {
			playersCopy[i].PerformanceStatsNumeric = make(map[string]float64)
			for key, value := range players[i].PerformanceStatsNumeric {
				playersCopy[i].PerformanceStatsNumeric[key] = value
			}
		}

		// Deep copy StringSlice fields (these are safe but let's be thorough)
		if players[i].ParsedPositions != nil {
			playersCopy[i].ParsedPositions = make([]string, len(players[i].ParsedPositions))
			copy(playersCopy[i].ParsedPositions, players[i].ParsedPositions)
		}

		if players[i].ShortPositions != nil {
			playersCopy[i].ShortPositions = make([]string, len(players[i].ShortPositions))
			copy(playersCopy[i].ShortPositions, players[i].ShortPositions)
		}

		if players[i].PositionGroups != nil {
			playersCopy[i].PositionGroups = make([]string, len(players[i].PositionGroups))
			copy(playersCopy[i].PositionGroups, players[i].PositionGroups)
		}

		// Initialize mutex for the copied player - CRITICAL for thread safety
		playersCopy[i].mu = sync.RWMutex{}

		// Deep copy RoleSpecificOveralls slice
		if players[i].RoleSpecificOveralls != nil {
			playersCopy[i].RoleSpecificOveralls = make([]RoleOverallScore, len(players[i].RoleSpecificOveralls))
			copy(playersCopy[i].RoleSpecificOveralls, players[i].RoleSpecificOveralls)
		}

		// Deep copy Attributes map
		if players[i].Attributes != nil {
			playersCopy[i].Attributes = make(map[string]string)
			for key, value := range players[i].Attributes {
				playersCopy[i].Attributes[key] = value
			}
		}

		// Deep copy NumericAttributes map
		if players[i].NumericAttributes != nil {
			playersCopy[i].NumericAttributes = make(map[string]int)
			for key, value := range players[i].NumericAttributes {
				playersCopy[i].NumericAttributes[key] = value
			}
		}
	}

	return playersCopy
}

// fullPlayerStatsHandler returns detailed stats for a single player
func fullPlayerStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.fullplayerstats.get")
	defer span.End()

	// Extract player UID and dataset ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format. Expected: /api/fullplayerstats/{datasetID}/{playerUID}", http.StatusBadRequest)
		return
	}

	datasetID := pathParts[3]
	playerUIDStr := pathParts[4]

	// Parse player UID
	playerUID, err := strconv.ParseInt(playerUIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid player UID", http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int64("player.uid", playerUID),
	)

	// Get dataset
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
	dataSpan.End()

	if !found {
		logError(ctx, "Failed to get dataset for full player stats",
			"dataset_id", datasetID,
			"player_uid", playerUID)
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
	}

	// Ensure configuration is loaded before processing player data
	// This is crucial for calculating role overall ratings and FM attributes
	if err := EnsureConfigInitialized(5 * time.Second); err != nil {
		logWarn(ctx, "Configuration initialization timed out, proceeding with default weights",
			"error", err,
			"dataset_id", datasetID,
			"player_uid", playerUID)
		// Continue with default weights rather than failing the request
	}

	// Find the specific player
	var targetPlayer *Player
	for _, player := range players {
		if player.UID == playerUID {
			targetPlayer = &player
			break
		}
	}

	if targetPlayer == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// OPTIMIZATION: Only calculate percentiles for this specific player if missing
	if targetPlayer.PerformancePercentiles == nil || len(targetPlayer.PerformancePercentiles) == 0 {
		logInfo(ctx, "Player missing percentiles, calculating for this player only",
			"dataset_id", datasetID,
			"player_uid", playerUID,
			"player_name", targetPlayer.Name)

		// OPTIMIZATION: Use division-based filtering for better performance
		// Default to "same" division instead of "all" for faster processing
		ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate_single_player")

		// Filter players by the same division as the target player
		var filteredPlayers []Player
		targetDivision := targetPlayer.Division

		// Add the target player first
		filteredPlayers = append(filteredPlayers, *targetPlayer)

		// Add players in the same division for comparison
		for _, player := range players {
			if player.UID != targetPlayer.UID && player.Division == targetDivision {
				filteredPlayers = append(filteredPlayers, player)
				// Limit to 50 players for faster processing
				if len(filteredPlayers) >= 50 {
					break
				}
			}
		}

		// If we don't have enough players in the same division, add some from similar positions
		if len(filteredPlayers) < 20 {
			for _, player := range players {
				if len(filteredPlayers) >= 50 {
					break
				}

				// Skip if already included
				alreadyIncluded := false
				for _, included := range filteredPlayers {
					if included.UID == player.UID {
						alreadyIncluded = true
						break
					}
				}

				if !alreadyIncluded && player.Position == targetPlayer.Position {
					filteredPlayers = append(filteredPlayers, player)
				}
			}
		}

		// Calculate percentiles only for the filtered comparison group
		CalculatePlayerPerformancePercentiles(filteredPlayers)

		// Find the updated player with percentiles
		for _, player := range filteredPlayers {
			if player.UID == playerUID {
				targetPlayer = &player
				break
			}
		}

		percentileSpan.End()

		logInfo(ctx, "Successfully calculated percentiles for player",
			"dataset_id", datasetID,
			"player_uid", playerUID,
			"player_name", targetPlayer.Name,
			"percentile_groups", len(targetPlayer.PerformancePercentiles),
			"comparison_players", len(filteredPlayers),
			"target_division", targetDivision)
	}

	// Create response with full player data
	response := map[string]interface{}{
		"player":          targetPlayer,
		"dataset_id":      datasetID,
		"currency_symbol": currencySymbol,
	}

	// Serialize to JSON string for protobuf compatibility
	jsonData, err := json.Marshal(response)
	if err != nil {
		logError(ctx, "Failed to marshal full player stats response",
			"error", err,
			"dataset_id", datasetID,
			"player_uid", playerUID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Use GenericResponse for protobuf compatibility
	protoResponse := &pb.GenericResponse{
		Data: string(jsonData),
	}

	// Use content negotiation to determine response format
	if err := WriteResponse(w, r, protoResponse); err != nil {
		logError(ctx, "Failed to write full player stats response",
			"error", err,
			"dataset_id", datasetID,
			"player_uid", playerUID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logInfo(ctx, "Successfully returned full player stats",
		"dataset_id", datasetID,
		"player_uid", playerUID,
		"player_name", targetPlayer.Name,
		"response_time_ms", time.Since(startTime).Milliseconds())
}

// getSimilarPlayersForComparison returns a subset of players for percentile calculation
// This avoids calculating percentiles for the entire dataset when only one player is needed
func getSimilarPlayersForComparison(allPlayers []Player, targetPlayer *Player) []Player {
	// Start with the target player
	similarPlayers := []Player{*targetPlayer}

	// Add players in the same position for comparison
	// This provides enough data for meaningful percentiles without processing the entire dataset
	position := targetPlayer.Position

	// Collect up to 100 similar players for comparison
	maxSimilarPlayers := 100
	count := 0

	for _, player := range allPlayers {
		if count >= maxSimilarPlayers {
			break
		}

		// Skip the target player (already added)
		if player.UID == targetPlayer.UID {
			continue
		}

		// Add players with similar position
		if position != "" && player.Position == position {
			similarPlayers = append(similarPlayers, player)
			count++
		}
	}

	// If we don't have enough similar players, add some random players for broader comparison
	if len(similarPlayers) < 50 {
		for _, player := range allPlayers {
			if len(similarPlayers) >= maxSimilarPlayers {
				break
			}

			// Skip if already included
			alreadyIncluded := false
			for _, included := range similarPlayers {
				if included.UID == player.UID {
					alreadyIncluded = true
					break
				}
			}

			if !alreadyIncluded {
				similarPlayers = append(similarPlayers, player)
			}
		}
	}

	return similarPlayers
}

// teamDataHandler returns detailed stats for all players in a team or nation
func teamDataHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Expected format: /api/team-data/{datasetID}/{type}/{name}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/team-data/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format. Expected: /api/team-data/{datasetID}/{type}/{name}", http.StatusBadRequest)
		return
	}

	datasetID := pathParts[0]
	dataType := pathParts[1] // "team" or "nation"
	teamOrNationName := pathParts[2]

	// Validate data type
	if dataType != "team" && dataType != "nation" {
		http.Error(w, "Invalid data type. Must be 'team' or 'nation'", http.StatusBadRequest)
		return
	}

	// Get dataset
	players, currencySymbol, found := GetPlayerData(datasetID)
	if !found {
		logError(ctx, "Failed to get dataset for team data",
			"dataset_id", datasetID,
			"type", dataType,
			"name", teamOrNationName)
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
	}

	// Filter players based on type
	var filteredPlayers []Player
	if dataType == "team" {
		filteredPlayers = make([]Player, 0)
		for _, player := range players {
			if player.Club == teamOrNationName {
				filteredPlayers = append(filteredPlayers, player)
			}
		}
	} else { // nation
		filteredPlayers = make([]Player, 0)
		for _, player := range players {
			if player.Nationality == teamOrNationName {
				filteredPlayers = append(filteredPlayers, player)
			}
		}
	}

	if len(filteredPlayers) == 0 {
		http.Error(w, fmt.Sprintf("No players found for %s: %s", dataType, teamOrNationName), http.StatusNotFound)
		return
	}

	// Ensure all players have percentile data and are properly enhanced
	playersWithPercentiles := make([]Player, len(filteredPlayers))
	copy(playersWithPercentiles, filteredPlayers)

	// Enhance players with calculations (FIFA stats, overall, position parsing)
	for i := range playersWithPercentiles {
		EnhancePlayerWithCalculations(&playersWithPercentiles[i])
	}

	// Recalculate all player ratings based on the current calculation method setting
	playersWithPercentiles = RecalculateAllPlayersRatings(playersWithPercentiles)

	// Calculate percentiles for the filtered players
	CalculatePlayerPerformancePercentiles(playersWithPercentiles)

	// Create response with full team/nation data
	response := map[string]interface{}{
		"players":         playersWithPercentiles,
		"dataset_id":      datasetID,
		"currency_symbol": currencySymbol,
		"type":            dataType,
		"name":            teamOrNationName,
		"player_count":    len(playersWithPercentiles),
	}

	// Serialize to JSON string for protobuf compatibility
	jsonData, err := json.Marshal(response)
	if err != nil {
		logError(ctx, "Failed to marshal team data response",
			"error", err,
			"dataset_id", datasetID,
			"type", dataType,
			"name", teamOrNationName)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Use GenericResponse for protobuf compatibility
	protoResponse := &pb.GenericResponse{
		Data: string(jsonData),
	}

	// Use content negotiation to determine response format
	if err := WriteResponse(w, r, protoResponse); err != nil {
		logError(ctx, "Failed to write team data response",
			"error", err,
			"dataset_id", datasetID,
			"type", dataType,
			"name", teamOrNationName)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logInfo(ctx, "Successfully returned team data",
		"dataset_id", datasetID,
		"type", dataType,
		"name", teamOrNationName,
		"player_count", len(playersWithPercentiles))
}

// performanceDataHandler handles GET requests for retrieving detailed performance data by dataset ID.
func performanceDataHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.performance.get")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/api/performance")
	defer DecrementActiveRequests(ctx, "/api/performance")

	// Record API operation metrics at the end
	defer func() {
		status := http.StatusOK
		RecordAPIOperation(ctx, r.Method, "/api/performance", status, time.Since(startTime))
	}()

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/performance"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/performance/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		logWarn(ctx, "Dataset ID missing in request path")
		SetSpanAttributes(ctx, attribute.String("error.type", "missing_dataset_id"))
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	SetSpanAttributes(ctx, attribute.String("dataset.id", datasetID))

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

	logDebug(ctx, "Processing performance data request",
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
		"position_compare", positionCompare)

	// Create cache key for performance data
	performanceCacheKey := fmt.Sprintf("performance:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		datasetID, filterPosition, filterRole, minAgeStr, maxAgeStr,
		minTransferValueStr, maxTransferValueStr, maxSalaryStr, divisionFilterStr, targetDivision)

	// Check cache for performance data
	if cachedPerformance, cacheFound := getFromMemCache(performanceCacheKey); cacheFound {
		if jsonData, ok := cachedPerformance.([]byte); ok {
			logDebug(ctx, "Serving performance data from cache", "dataset_id", datasetID)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
			setCORSHeaders(w, r)
			if _, err := w.Write(jsonData); err != nil {
				logError(ctx, "Error writing cached performance response", "error", err)
			}
			SetSpanAttributes(ctx, attribute.Bool("performance_cache.hit", true))
			return
		}
	}

	// Cache miss - need to load and process performance data
	SetSpanAttributes(ctx, attribute.Bool("performance_cache.hit", false))

	// Use the storage interface to get player data
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
	dataSpan.End()

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

	// Enhance only if necessary (most datasets are stored already enhanced)
	ctx, enhanceSpan := StartSpan(ctx, "players.enhance")
	needsEnhancement := false
	if len(players) > 0 {
		if len(players[0].NumericAttributes) == 0 || players[0].PerformanceStatsNumeric == nil {
			needsEnhancement = true
		}
	}
	if needsEnhancement {
		logInfo(ctx, "Enhancing players with calculations", "player_count", len(players))
		for i := range players {
			EnhancePlayerWithCalculations(&players[i])
		}
		logInfo(ctx, "Enhanced players with calculations", "player_count", len(players))
	} else {
		logDebug(ctx, "Skipping enhancement; dataset already enhanced", "player_count", len(players))
	}
	enhanceSpan.End()

	// Recalculate all player ratings based on the current calculation method setting
	ctx, recalcSpan := StartSpan(ctx, "ratings.recalculate")
	players = RecalculateAllPlayersRatings(players)
	recalcSpan.End()

	// Normalize MBR scores relative to the maximum (best player gets 100)
	NormalizeMBRScoresRelativeToMax(players)

	// Parse filters
	var minAge, maxAge int
	var minTransferValue, maxTransferValue, maxSalary int64

	if minAgeStr != "" {
		if val, err := strconv.Atoi(minAgeStr); err == nil {
			minAge = val
		}
	}
	if maxAgeStr != "" {
		if val, err := strconv.Atoi(maxAgeStr); err == nil {
			maxAge = val
		}
	}
	if minTransferValueStr != "" {
		if val, err := strconv.ParseInt(minTransferValueStr, 10, 64); err == nil {
			minTransferValue = val
		}
	}
	if maxTransferValueStr != "" {
		if val, err := strconv.ParseInt(maxTransferValueStr, 10, 64); err == nil {
			maxTransferValue = val
		}
	}
	if maxSalaryStr != "" {
		if val, err := strconv.ParseInt(maxSalaryStr, 10, 64); err == nil {
			maxSalary = val
		}
	}

	// Parse division filter
	var divisionFilter = DivisionFilterAll
	switch divisionFilterStr {
	case "same":
		divisionFilter = DivisionFilterSame
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	}

	// Calculate percentiles with appropriate filtering
	ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate")
	// Make a deep copy of players to avoid modifying the stored data
	// PERFORMANCE: Use FastDeepCopyPlayers for much better performance while staying thread-safe
	playersCopy := FastDeepCopyPlayers(players)

	if divisionFilter != DivisionFilterAll {
		// Recalculate percentiles with division filter
		CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, targetDivision)
	} else {
		// Calculate global percentiles using optimized algorithm
		CalculatePlayerPerformancePercentiles(playersCopy)
	}

	players = playersCopy
	percentileSpan.End()

	// Apply filters to get final performance data
	ctx, filterSpan := StartSpan(ctx, "performance.filter")
	filteredPlayers := filterPlayersForPerformance(players, filterPosition, filterRole, minAge, maxAge, minTransferValue, maxTransferValue, maxSalary, positionCompare)
	filterSpan.End()

	SetSpanAttributes(ctx, attribute.Int("performance.filtered_player_count", len(filteredPlayers)))

	// Create response with detailed performance data
	response := struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
		TotalCount     int      `json:"totalCount"`
		FilteredCount  int      `json:"filteredCount"`
	}{
		Players:        filteredPlayers,
		CurrencySymbol: currencySymbol,
		TotalCount:     len(players),
		FilteredCount:  len(filteredPlayers),
	}

	// Serialize response
	var responseData []byte
	var err error

	if supportsProtobuf {
		// For protobuf, we need to convert to JSON first, then wrap in protobuf
		_, jsonSpan := StartSpan(ctx, "serialization.json")
		jsonData, jsonErr := json.Marshal(response)
		jsonSpan.End()

		if jsonErr != nil {
			logError(ctx, "Failed to marshal performance response to JSON", "error", jsonErr)
			SetSpanAttributes(ctx, attribute.String("error.type", "json_marshal_failed"))
			WriteErrorResponse(w, r, "serialization_failed", "Failed to serialize response", nil, http.StatusInternalServerError)
			return
		}

		// Create protobuf response wrapper
		_, protoSpan := StartSpan(ctx, "serialization.protobuf")
		protoResponse := &pb.GenericResponse{
			Data: string(jsonData),
		}
		responseData, err = proto.Marshal(protoResponse)
		protoSpan.End()
	} else {
		// Use JSON serialization
		_, jsonSpan := StartSpan(ctx, "serialization.json")
		responseData, err = json.Marshal(response)
		jsonSpan.End()
	}

	if err != nil {
		logError(ctx, "Failed to serialize performance response", "error", err)
		SetSpanAttributes(ctx, attribute.String("error.type", "serialization_failed"))
		WriteErrorResponse(w, r, "serialization_failed", "Failed to serialize response", nil, http.StatusInternalServerError)
		return
	}

	// Cache the performance data
	setInMemCacheForDataset(performanceCacheKey, responseData, 5*time.Minute) // Cache for 5 minutes

	// Set response headers
	w.Header().Set("Content-Type", serializer.ContentType())
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	setCORSHeaders(w, r)

	// Write response
	if _, err := w.Write(responseData); err != nil {
		logError(ctx, "Error writing performance response", "error", err)
	}

	logDebug(ctx, "Performance data request completed",
		"dataset_id", datasetID,
		"total_players", len(players),
		"filtered_players", len(filteredPlayers),
		"response_size_bytes", len(responseData))
}

// exportDataHandler handles GET requests for exporting complete dataset data.
func exportDataHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.export.get")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/api/export")
	defer DecrementActiveRequests(ctx, "/api/export")

	// Record API operation metrics at the end
	defer func() {
		status := http.StatusOK
		RecordAPIOperation(ctx, r.Method, "/api/export", status, time.Since(startTime))
	}()

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/export"),
		attribute.String("request.id", requestID),
	)

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/export/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		logWarn(ctx, "Dataset ID missing in request path")
		SetSpanAttributes(ctx, attribute.String("error.type", "missing_dataset_id"))
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	SetSpanAttributes(ctx, attribute.String("dataset.id", datasetID))

	// Get format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default to JSON
	}

	if format != "json" && format != "csv" {
		logWarn(ctx, "Invalid export format", "format", format)
		SetSpanAttributes(ctx, attribute.String("error.type", "invalid_format"))
		WriteErrorResponse(w, r, "invalid_format", "Invalid export format. Supported formats: json, csv", nil, http.StatusBadRequest)
		return
	}

	SetSpanAttributes(ctx, attribute.String("export.format", format))

	logDebug(ctx, "Processing export request",
		"dataset_id", datasetID,
		"format", format)

	// Use the storage interface to get player data
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
	dataSpan.End()

	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		SetSpanAttributes(ctx, attribute.String("error.type", "dataset_not_found"))
		WriteErrorResponse(w, r, "dataset_not_found", "Player data not found for the given ID.", nil, http.StatusNotFound)
		return
	}

	SetSpanAttributes(ctx,
		attribute.Int("dataset.player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
	)

	// Enhance only if needed for export
	ctx, enhanceSpan := StartSpan(ctx, "players.enhance")
	needsEnhancement := false
	if len(players) > 0 {
		if len(players[0].NumericAttributes) == 0 || players[0].PerformanceStatsNumeric == nil {
			needsEnhancement = true
		}
	}
	if needsEnhancement {
		logInfo(ctx, "Enhancing players with calculations for export", "player_count", len(players))
		for i := range players {
			EnhancePlayerWithCalculations(&players[i])
		}
		logInfo(ctx, "Enhanced players with calculations for export", "player_count", len(players))
	} else {
		logDebug(ctx, "Skipping enhancement for export; dataset already enhanced", "player_count", len(players))
	}
	enhanceSpan.End()

	// Recalculate all player ratings
	ctx, recalcSpan := StartSpan(ctx, "ratings.recalculate")
	players = RecalculateAllPlayersRatings(players)
	recalcSpan.End()

	// Normalize MBR scores relative to the maximum (best player gets 100)
	NormalizeMBRScoresRelativeToMax(players)

	// Calculate performance percentiles
	ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate")
	CalculatePlayerPerformancePercentiles(players)
	percentileSpan.End()

	SetSpanAttributes(ctx, attribute.Int("export.player_count", len(players)))

	// Export based on format
	var responseData []byte

	if format == "csv" {
		// Generate CSV content
		_, csvSpan := StartSpan(ctx, "export.csv_generate")
		csvContent := generateCSVContent(players)
		csvSpan.End()

		// Set response headers for CSV download
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=players_export_%s.csv", datasetID))
		setCORSHeaders(w, r)

		responseData = []byte(csvContent)
	} else {
		// Generate JSON content
		_, jsonSpan := StartSpan(ctx, "export.json_generate")
		jsonContent := generateJSONContent(players, currencySymbol, datasetID)
		jsonSpan.End()

		// Set response headers for JSON
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=players_export_%s.json", datasetID))
		setCORSHeaders(w, r)

		responseData = jsonContent
	}

	// Write response
	if _, writeErr := w.Write(responseData); writeErr != nil {
		logError(ctx, "Error writing export response", "error", writeErr)
	}

	logDebug(ctx, "Export request completed",
		"dataset_id", datasetID,
		"format", format,
		"player_count", len(players),
		"response_size_bytes", len(responseData))
}

// filterPlayersForPerformance applies all filters to get performance data
func filterPlayersForPerformance(players []Player, position, role string, minAge, maxAge int, minTransferValue, maxTransferValue, maxSalary int64, positionCompare string) []Player {
	filtered := make([]Player, 0, len(players))

	for _, player := range players {
		// Age filter
		if minAge > 0 {
			if age, err := strconv.Atoi(player.Age); err == nil && age < minAge {
				continue
			}
		}
		if maxAge > 0 {
			if age, err := strconv.Atoi(player.Age); err == nil && age > maxAge {
				continue
			}
		}

		// Transfer value filter
		if minTransferValue > 0 || maxTransferValue > 0 {
			// Skip players who are "Not for Sale"
			if player.TransferValue == "Not for Sale" ||
				strings.Contains(strings.ToLower(player.TransferValue), "not for sale") {
				continue
			}

			// Use the parsed TransferValueAmount field instead of trying to parse the string
			if minTransferValue > 0 && player.TransferValueAmount < minTransferValue {
				continue
			}
			if maxTransferValue > 0 && player.TransferValueAmount > maxTransferValue {
				continue
			}
		}

		// Salary filter
		if maxSalary > 0 {
			if wage, err := strconv.ParseInt(player.Wage, 10, 64); err == nil && wage > maxSalary {
				continue
			}
		}

		// Position filter
		if position != "" && !matchesPosition(player, position, positionCompare) {
			continue
		}

		// Role filter
		if role != "" && !matchesRole(player, role) {
			continue
		}

		filtered = append(filtered, player)
	}

	return filtered
}

// matchesPosition checks if a player matches the given position filter
func matchesPosition(player Player, position, compare string) bool {
	if position == "" {
		return true
	}

	// Handle position groups
	switch position {
	case "Goalkeeper":
		return containsPosition(player.ShortPositions, "GK")
	case "Defender":
		return containsAnyPosition(player.ShortPositions, []string{"DC", "DR", "DL", "WBR", "WBL"})
	case "Midfielder":
		return containsAnyPosition(player.ShortPositions, []string{"DM", "MC", "MR", "ML", "AMR", "AMC", "AML"})
	case "Forward":
		return containsPosition(player.ShortPositions, "ST")
	}

	// Handle specific positions
	return containsPosition(player.ShortPositions, position)
}

// matchesRole checks if a player matches the given role filter
func matchesRole(player Player, role string) bool {
	if role == "" {
		return true
	}

	// Check if player has the specific role in their role-specific overalls
	if player.RoleSpecificOveralls != nil {
		for _, rso := range player.RoleSpecificOveralls {
			if strings.Contains(strings.ToLower(rso.RoleName), strings.ToLower(role)) {
				return true
			}
		}
	}

	return false
}

// containsPosition checks if a slice contains a specific position
func containsPosition(positions []string, position string) bool {
	for _, pos := range positions {
		if pos == position {
			return true
		}
	}
	return false
}

// containsAnyPosition checks if a slice contains any of the given positions
func containsAnyPosition(positions []string, targetPositions []string) bool {
	for _, pos := range positions {
		for _, target := range targetPositions {
			if pos == target {
				return true
			}
		}
	}
	return false
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getAttributeFullName returns the full name for an attribute key
func getAttributeFullName(attrKey string) string {
	attributeFullNameMap := map[string]string{
		"Cor":  "Corners",
		"Cro":  "Crossing",
		"Dri":  "Dribbling",
		"Fin":  "Finishing",
		"Fir":  "First Touch",
		"Fre":  "Free Kick Taking",
		"Hea":  "Heading",
		"Lon":  "Long Shots",
		"L Th": "Long Throws",
		"Mar":  "Marking",
		"Pas":  "Passing",
		"Pen":  "Penalty Taking",
		"Tck":  "Tackling",
		"Tec":  "Technique",
		"Agg":  "Aggression",
		"Ant":  "Anticipation",
		"Bra":  "Bravery",
		"Cmp":  "Composure",
		"Cnt":  "Concentration",
		"Dec":  "Decisions",
		"Det":  "Determination",
		"Fla":  "Flair",
		"Ldr":  "Leadership",
		"OtB":  "Off the Ball",
		"Pos":  "Positioning",
		"Tea":  "Teamwork",
		"Vis":  "Vision",
		"Wor":  "Work Rate",
		"Acc":  "Acceleration",
		"Agi":  "Agility",
		"Bal":  "Balance",
		"Jum":  "Jumping Reach",
		"Nat":  "Natural Fitness",
		"Pac":  "Pace",
		"Sta":  "Stamina",
		"Str":  "Strength",
		"Aer":  "Aerial Reach",
		"Cmd":  "Command of Area",
		"Com":  "Communication",
		"Ecc":  "Eccentricity",
		"Han":  "Handling",
		"Kic":  "Kicking",
		"1v1":  "One on Ones",
		"Pun":  "Punching (Tendency)",
		"Ref":  "Reflexes",
		"TRO":  "Rushing Out (Tendency)",
		"Thr":  "Throwing",
	}

	if fullName, exists := attributeFullNameMap[attrKey]; exists {
		return fullName
	}
	return attrKey // Return the original key if no mapping exists
}

// generateCSVContent generates CSV content from player data
func generateCSVContent(players []Player) string {
	if len(players) == 0 {
		return ""
	}

	// First, collect all unique role names and performance percentile groups
	roleNames := make(map[string]bool)
	percentileGroups := make(map[string]bool)
	attributeNames := make(map[string]bool)

	for _, player := range players {
		// Collect role names
		for _, role := range player.RoleSpecificOveralls {
			roleNames[role.RoleName] = true
		}

		// Collect performance percentile groups
		for group := range player.PerformancePercentiles {
			percentileGroups[group] = true
		}

		// Collect attribute names
		for attr := range player.Attributes {
			attributeNames[attr] = true
		}
	}

	// Convert maps to sorted slices for consistent column ordering
	roleNamesList := make([]string, 0, len(roleNames))
	for role := range roleNames {
		roleNamesList = append(roleNamesList, role)
	}
	sort.Strings(roleNamesList)

	percentileGroupsList := make([]string, 0, len(percentileGroups))
	for group := range percentileGroups {
		percentileGroupsList = append(percentileGroupsList, group)
	}
	sort.Strings(percentileGroupsList)

	attributeNamesList := make([]string, 0, len(attributeNames))
	for attr := range attributeNames {
		attributeNamesList = append(attributeNamesList, attr)
	}
	sort.Strings(attributeNamesList)

	// Define CSV headers
	headers := []string{
		"Name", "Age", "Nationality", "Nationality ISO", "Nationality FIFA Code",
		"Club", "Division", "Position", "Transfer Value", "Wage",
		"Overall", "PAC", "SHO", "PAS", "DRI", "DEF", "PHY", "TotalStats", "MBR", "GK",
		"DIV", "HAN", "REF", "KIC", "SPD", "POS", // Individual GK stats
		"Personality", "Media Handling", "Attributes Masked",
	}

	// Add FM attribute columns right after Attributes Masked (Technical -> Mental -> Physical -> Goalkeeper)
	// Technical attributes
	technicalAttrs := []string{"Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec"}
	for _, attrName := range technicalAttrs {
		if attributeNames[attrName] {
			headers = append(headers, getAttributeFullName(attrName))
		}
	}

	// Mental attributes
	mentalAttrs := []string{"Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor"}
	for _, attrName := range mentalAttrs {
		if attributeNames[attrName] {
			headers = append(headers, getAttributeFullName(attrName))
		}
	}

	// Physical attributes
	physicalAttrs := []string{"Acc", "Agi", "Bal", "Jum", "Nat", "Pac", "Sta", "Str"}
	for _, attrName := range physicalAttrs {
		if attributeNames[attrName] {
			headers = append(headers, getAttributeFullName(attrName))
		}
	}

	// Goalkeeper attributes
	gkAttrs := []string{"Aer", "Cmd", "Com", "Ecc", "Han", "Kic", "1v1", "Pun", "Ref", "TRO", "Thr"}
	for _, attrName := range gkAttrs {
		if attributeNames[attrName] {
			headers = append(headers, getAttributeFullName(attrName))
		}
	}

	// Add any remaining attributes that weren't in the predefined lists
	for _, attrName := range attributeNamesList {
		if !contains(technicalAttrs, attrName) && !contains(mentalAttrs, attrName) &&
			!contains(physicalAttrs, attrName) && !contains(gkAttrs, attrName) {
			headers = append(headers, getAttributeFullName(attrName))
		}
	}

	// Add role rating columns
	for _, roleName := range roleNamesList {
		headers = append(headers, fmt.Sprintf("Role_%s", roleName))
	}

	// Add performance percentile columns
	for _, group := range percentileGroupsList {
		headers = append(headers, fmt.Sprintf("Percentile_%s", group))
	}

	// Create CSV content
	var csvLines []string
	csvLines = append(csvLines, strings.Join(headers, ","))

	for _, player := range players {
		// Basic info
		row := []string{
			escapeCSVField(player.Name),
			escapeCSVField(player.Age),
			escapeCSVField(player.Nationality),
			escapeCSVField(player.NationalityISO),
			escapeCSVField(player.NationalityFIFACode),
			escapeCSVField(player.Club),
			escapeCSVField(player.Division),
			escapeCSVField(player.Position),
			escapeCSVField(player.TransferValue),
			escapeCSVField(player.Wage),
		}

		// FIFA stats (outfield)
		row = append(row, []string{
			fmt.Sprintf("%d", player.Overall),
			fmt.Sprintf("%d", player.PAC),
			fmt.Sprintf("%d", player.SHO),
			fmt.Sprintf("%d", player.PAS),
			fmt.Sprintf("%d", player.DRI),
			fmt.Sprintf("%d", player.DEF),
			fmt.Sprintf("%d", player.PHY),
			fmt.Sprintf("%d", player.TotalStats),
			fmt.Sprintf("%d", player.MBR),
			fmt.Sprintf("%d", player.GK),
		}...)

		// Individual GK stats
		row = append(row, []string{
			fmt.Sprintf("%d", player.DIV),
			fmt.Sprintf("%d", player.HAN),
			fmt.Sprintf("%d", player.REF),
			fmt.Sprintf("%d", player.KIC),
			fmt.Sprintf("%d", player.SPD),
			fmt.Sprintf("%d", player.POS),
		}...)

		// Personal info
		row = append(row, []string{
			escapeCSVField(player.Personality),
			escapeCSVField(player.MediaHandling),
			fmt.Sprintf("%t", player.AttributeMasked),
		}...)

		// FM Attributes (individual columns) right after Attributes Masked (Technical -> Mental -> Physical -> Goalkeeper)
		// Protect attribute map access with read lock
		player.mu.RLock()

		// Technical attributes
		technicalAttrs := []string{"Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec"}
		for _, attrName := range technicalAttrs {
			if attributeNames[attrName] {
				if value, exists := player.Attributes[attrName]; exists {
					row = append(row, escapeCSVField(value))
				} else {
					row = append(row, "") // Empty for players without this attribute
				}
			}
		}

		// Mental attributes
		mentalAttrs := []string{"Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor"}
		for _, attrName := range mentalAttrs {
			if attributeNames[attrName] {
				if value, exists := player.Attributes[attrName]; exists {
					row = append(row, escapeCSVField(value))
				} else {
					row = append(row, "") // Empty for players without this attribute
				}
			}
		}

		// Physical attributes
		physicalAttrs := []string{"Acc", "Agi", "Bal", "Jum", "Nat", "Pac", "Sta", "Str"}
		for _, attrName := range physicalAttrs {
			if attributeNames[attrName] {
				if value, exists := player.Attributes[attrName]; exists {
					row = append(row, escapeCSVField(value))
				} else {
					row = append(row, "") // Empty for players without this attribute
				}
			}
		}

		// Goalkeeper attributes
		gkAttrs := []string{"Aer", "Cmd", "Com", "Ecc", "Han", "Kic", "1v1", "Pun", "Ref", "TRO", "Thr"}
		for _, attrName := range gkAttrs {
			if attributeNames[attrName] {
				if value, exists := player.Attributes[attrName]; exists {
					row = append(row, escapeCSVField(value))
				} else {
					row = append(row, "") // Empty for players without this attribute
				}
			}
		}

		// Add any remaining attributes that weren't in the predefined lists
		for _, attrName := range attributeNamesList {
			if !contains(technicalAttrs, attrName) && !contains(mentalAttrs, attrName) &&
				!contains(physicalAttrs, attrName) && !contains(gkAttrs, attrName) {
				if value, exists := player.Attributes[attrName]; exists {
					row = append(row, escapeCSVField(value))
				} else {
					row = append(row, "") // Empty for players without this attribute
				}
			}
		}

		// Role ratings (individual columns)
		roleMap := make(map[string]int)
		for _, role := range player.RoleSpecificOveralls {
			roleMap[role.RoleName] = role.Score
		}
		for _, roleName := range roleNamesList {
			if score, exists := roleMap[roleName]; exists {
				row = append(row, fmt.Sprintf("%d", score))
			} else {
				row = append(row, "") // Empty for players without this role
			}
		}

		// Performance percentiles (individual columns)
		for _, group := range percentileGroupsList {
			if percentiles, exists := player.PerformancePercentiles[group]; exists {
				percentilesJSON, _ := json.Marshal(percentiles)
				row = append(row, escapeCSVField(string(percentilesJSON)))
			} else {
				row = append(row, "") // Empty for players without this percentile group
			}
		}

		// Release the read lock after all map access is complete
		player.mu.RUnlock()

		csvLines = append(csvLines, strings.Join(row, ","))
	}

	return strings.Join(csvLines, "\n")
}

// generateJSONContent generates JSON content from player data
func generateJSONContent(players []Player, currencySymbol, datasetID string) []byte {
	// Create export metadata
	exportData := struct {
		Metadata struct {
			ExportDate     string `json:"exportDate"`
			DatasetID      string `json:"datasetId"`
			TotalPlayers   int    `json:"totalPlayers"`
			CurrencySymbol string `json:"currencySymbol"`
			Format         string `json:"format"`
			Version        string `json:"version"`
		} `json:"metadata"`
		Players []Player `json:"players"`
	}{
		Players: players,
	}

	exportData.Metadata.ExportDate = time.Now().Format(time.RFC3339)
	exportData.Metadata.DatasetID = datasetID
	exportData.Metadata.TotalPlayers = len(players)
	exportData.Metadata.CurrencySymbol = currencySymbol
	exportData.Metadata.Format = "json"
	exportData.Metadata.Version = "1.0"

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		logError(context.Background(), "Error marshaling JSON export", "error", err)
		return []byte("{}")
	}

	return jsonData
}

// escapeCSVField escapes a field for CSV format
func escapeCSVField(field string) string {
	if strings.Contains(field, ",") || strings.Contains(field, "\"") || strings.Contains(field, "\n") {
		// Escape quotes by doubling them
		field = strings.ReplaceAll(field, "\"", "\"\"")
		// Wrap in quotes
		field = "\"" + field + "\""
	}
	return field
}

// extractRolesFromPlayers extracts unique roles from the player data
func extractRolesFromPlayers(players []Player) []string {
	roleSet := make(map[string]bool)

	for _, player := range players {
		// Extract roles from role-specific overalls
		for _, roleOverall := range player.RoleSpecificOveralls {
			if roleOverall.RoleName != "" {
				roleSet[roleOverall.RoleName] = true
			}
		}

		// Extract roles from position groups
		for _, positionGroup := range player.PositionGroups {
			if positionGroup != "" {
				roleSet[positionGroup] = true
			}
		}
	}

	// Convert map to slice
	roles := make([]string, 0, len(roleSet))
	for role := range roleSet {
		roles = append(roles, role)
	}

	// Sort for consistent output
	sort.Strings(roles)
	return roles
}

// processBasicPlayerData performs basic processing on players without heavy calculations
func processBasicPlayerData(players []Player) []Player {
	processedPlayers := make([]Player, len(players))

	for i, player := range players {
		// Copy the player
		processedPlayers[i] = player

		// Ensure basic fields are properly set
		if processedPlayers[i].Attributes == nil {
			processedPlayers[i].Attributes = make(map[string]string)
		}
		if processedPlayers[i].NumericAttributes == nil {
			processedPlayers[i].NumericAttributes = make(map[string]int)
		}
		if processedPlayers[i].PerformanceStatsNumeric == nil {
			processedPlayers[i].PerformanceStatsNumeric = make(map[string]float64)
		}
		if processedPlayers[i].PerformancePercentiles == nil {
			processedPlayers[i].PerformancePercentiles = make(map[string]map[string]float64)
		}
		if processedPlayers[i].ShortPositions == nil {
			processedPlayers[i].ShortPositions = make([]string, 0)
		}
		if processedPlayers[i].ParsedPositions == nil {
			processedPlayers[i].ParsedPositions = make([]string, 0)
		}
		if processedPlayers[i].PositionGroups == nil {
			processedPlayers[i].PositionGroups = make([]string, 0)
		}
		if processedPlayers[i].RoleSpecificOveralls == nil {
			processedPlayers[i].RoleSpecificOveralls = make([]RoleOverallScore, 0)
		}

		// Ensure numeric fields are properly set
		if processedPlayers[i].Overall == 0 {
			processedPlayers[i].Overall = 50 // Default value
		}
		processedPlayers[i].OverallLower = processedPlayers[i].Overall
	}

	return processedPlayers
}

// CalculatePlayerPercentilesAsync calculates percentiles for a dataset asynchronously
func CalculatePlayerPercentilesAsync(ctx context.Context, datasetID string, players []Player, currencySymbol string) error {
	ctx, span := StartSpan(ctx, "percentile_calculation_async")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("player_count", len(players)),
	)

	logInfo(ctx, "Starting async percentile calculation for dataset %s", datasetID)
	startTime := time.Now()

	// Get a fresh copy of the data from the store to avoid race conditions
	// This ensures we're working with the most up-to-date data and not a reference
	// that might be modified by other goroutines
	storeMutex.RLock()
	storedData, exists := playerDataStore[datasetID]
	storeMutex.RUnlock()

	if !exists {
		logError(ctx, "Dataset %s not found in store for percentile calculation", datasetID)
		return WrapErrorf(ErrDatasetNotFound, "dataset: %s", datasetID)
	}

	// Use the existing OptimizedDeepCopyPlayers function which is designed to handle this safely
	// CRITICAL: Use safe deepCopyPlayers instead of OptimizedDeepCopyPlayers (COW has race conditions)
	playersCopy := deepCopyPlayers(storedData.Players)

	// Calculate percentiles for all division filters to ensure stability
	CalculatePlayerPerformancePercentiles(playersCopy)

	// Log top 25 MBR players after calculations are complete
	logTop25OverallPlayers(playersCopy)

	// Update the stored data with calculated percentiles
	SetPlayerData(datasetID, playersCopy, storedData.CurrencySymbol)

	calculationTime := time.Since(startTime)
	logInfo(ctx, "Completed async percentile calculation for dataset %s in %v", datasetID, calculationTime)

	SetSpanAttributes(ctx,
		attribute.Int64("calculation_time_ms", calculationTime.Milliseconds()),
		attribute.String("calculation.status", "success"),
	)

	RecordBusinessOperation(ctx, "percentile_calculation_completed", true, map[string]interface{}{
		"dataset_id":       datasetID,
		"player_count":     len(players),
		"calculation_time": calculationTime.Milliseconds(),
	})

	return nil
}

// --- END: Struct Definitions ---

// UpgradeFinderRequest represents the request parameters for finding player upgrades
type UpgradeFinderRequest struct {
	DatasetID        string `json:"datasetId"`     // Dataset to search for upgrades (transfer market)
	TeamDatasetID    string `json:"teamDatasetId"` // Optional: Dataset containing the team (current squad)
	Team             string `json:"team"`
	Position         string `json:"position"`
	Role             string `json:"role"`
	MinOverall       int    `json:"minOverall"`
	MaxAge           int    `json:"maxAge"`
	MaxTransferValue int64  `json:"maxTransferValue"`
	MaxSalary        int64  `json:"maxSalary"`
	// Minimum attribute filters
	MinPAC int `json:"minPAC"`
	MinDRI int `json:"minDRI"`
	MinSHO int `json:"minSHO"`
	MinPAS int `json:"minPAS"`
	MinDEF int `json:"minDEF"`
	MinPHY int `json:"minPHY"`
	MinGK  int `json:"minGK"`
	MinDIV int `json:"minDIV"`
	MinHAN int `json:"minHAN"`
	MinREF int `json:"minREF"`
	MinKIC int `json:"minKIC"`
	MinSPD int `json:"minSPD"`
	MinPOS int `json:"minPOS"`
}

// UpgradeFinderResponse represents the response for upgrade finding
type UpgradeFinderResponse struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currencySymbol"`
	BaseOverall    int      `json:"baseOverall"`
	TargetOverall  int      `json:"targetOverall"`
	TotalFound     int      `json:"totalFound"`
}

// upgradeFinderHandler handles POST requests for finding player upgrades
func upgradeFinderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.upgrade-finder.post")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/api/upgrade-finder")
	defer DecrementActiveRequests(ctx, "/api/upgrade-finder")

	// Record API operation metrics at the end
	defer func() {
		status := http.StatusOK
		RecordAPIOperation(ctx, r.Method, "/api/upgrade-finder", status, time.Since(startTime))
	}()

	if r.Method != http.MethodPost {
		WriteErrorResponse(w, r, "method_not_allowed", "Only POST method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/upgrade-finder"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	// Parse request body
	var req UpgradeFinderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(ctx, "Error decoding upgrade finder request", "error", err)
		WriteErrorResponse(w, r, "invalid_request", "Invalid request body", nil, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.DatasetID == "" {
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is required", nil, http.StatusBadRequest)
		return
	}

	if req.Position == "" {
		WriteErrorResponse(w, r, "missing_position", "Position is required", nil, http.StatusBadRequest)
		return
	}

	if req.MinOverall <= 0 {
		WriteErrorResponse(w, r, "invalid_min_overall", "Minimum overall must be greater than 0", nil, http.StatusBadRequest)
		return
	}

	logDebug(ctx, "Processing upgrade finder request",
		"dataset_id", req.DatasetID,
		"team", req.Team,
		"position", req.Position,
		"role", req.Role,
		"min_overall", req.MinOverall,
		"max_age", req.MaxAge,
		"max_transfer_value", req.MaxTransferValue,
		"max_salary", req.MaxSalary)

	// Get players from the main dataset (transfer market) using optimized loader
	dataLoadStart := time.Now()
	players, currencySymbol, found := GetPlayerDataForUpgradeFinder(req.DatasetID)
	dataLoadTime := time.Since(dataLoadStart)

	if !found {
		logWarn(ctx, "No players found for main dataset", "dataset_id", req.DatasetID)
		WriteErrorResponse(w, r, "no_players_found", "No players found for the specified dataset", nil, http.StatusNotFound)
		return
	}

	logDebug(ctx, "Data loading completed for upgrade finder",
		"dataset_id", req.DatasetID,
		"player_count", len(players),
		"data_load_time_ms", dataLoadTime.Milliseconds())

	// If a team dataset is specified, get team players from it for validation
	var teamPlayers []Player
	if req.TeamDatasetID != "" && req.TeamDatasetID != req.DatasetID {
		teamLoadStart := time.Now()
		teamPlayersData, _, teamFound := GetPlayerDataForUpgradeFinder(req.TeamDatasetID)
		teamLoadTime := time.Since(teamLoadStart)

		if !teamFound {
			logWarn(ctx, "No players found for team dataset", "team_dataset_id", req.TeamDatasetID)
			WriteErrorResponse(w, r, "no_team_players_found", "No players found for the specified team dataset", nil, http.StatusNotFound)
			return
		}
		teamPlayers = teamPlayersData
		logDebug(ctx, "Using separate team dataset for team validation",
			"team_dataset_id", req.TeamDatasetID,
			"team_players_count", len(teamPlayers),
			"team_data_load_time_ms", teamLoadTime.Milliseconds())
	} else {
		// Use the same dataset for both team and upgrades
		teamPlayers = players
	}

	// Add debug logging
	logDebug(ctx, "Upgrade finder processing",
		"main_dataset_id", req.DatasetID,
		"team_dataset_id", req.TeamDatasetID,
		"using_separate_datasets", req.TeamDatasetID != "" && req.TeamDatasetID != req.DatasetID,
		"main_dataset_players", len(players),
		"team_dataset_players", len(teamPlayers),
		"team_name", req.Team,
		"position", req.Position,
		"min_overall", req.MinOverall)

	// Filter and find upgrades using optimized parallel processing
	filterStart := time.Now()
	upgrades := OptimizedFindPlayerUpgrades(players, teamPlayers, req)
	filterTime := time.Since(filterStart)

	logDebug(ctx, "Upgrade finder results",
		"upgrades_found", len(upgrades),
		"request_team", req.Team,
		"request_position", req.Position,
		"filter_time_ms", filterTime.Milliseconds(),
		"data_load_time_ms", dataLoadTime.Milliseconds(),
		"total_time_ms", time.Since(startTime).Milliseconds())

	// Set CORS headers
	setCORSHeaders(w, r)

	// Create response
	response := UpgradeFinderResponse{
		Players:        upgrades,
		CurrencySymbol: currencySymbol,
		BaseOverall:    req.MinOverall,
		TargetOverall:  req.MinOverall + 1, // Default upgrade by 1
		TotalFound:     len(upgrades),
	}

	if supportsProtobuf {
		// Create protobuf response (if protobuf types are available)
		// For now, fall back to JSON
		logDebug(ctx, "Protobuf not implemented for upgrade finder, using JSON")
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logError(ctx, "Error encoding JSON response for upgrade finder", "error", err)
		WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		return
	}

	logDebug(ctx, "Upgrade finder response served",
		"upgrades_found", len(upgrades),
		"processing_time_ms", time.Since(startTime).Milliseconds())
}

// matchesPositionForUpgrade checks if a player matches the required position
func matchesPositionForUpgrade(player Player, position string) bool {
	// Check short positions first
	for _, pos := range player.ShortPositions {
		if pos == position {
			return true
		}
	}

	// Check parsed positions
	for _, pos := range player.ParsedPositions {
		if pos == position {
			return true
		}
	}

	return false
}

// getPlayerOverallForRole gets the overall rating for a player in a specific role
func getPlayerOverallForRole(player Player, role, position string) int {
	// If role is specified, try to get role-specific overall
	if role != "" {
		for _, roleOverall := range player.RoleSpecificOveralls {
			if roleOverall.RoleName == role {
				return roleOverall.Score
			}
		}
	}

	// If position is specified, find the best role-specific overall for that position
	if position != "" {
		bestOverall := 0
		for _, roleOverall := range player.RoleSpecificOveralls {
			if strings.HasPrefix(roleOverall.RoleName, position+" - ") {
				if roleOverall.Score > bestOverall {
					bestOverall = roleOverall.Score
				}
			}
		}
		if bestOverall > 0 {
			return bestOverall
		}
	}

	// Fall back to main overall
	return player.Overall
}

// processTopTeamsData returns the top teams across all divisions
func processTopTeamsData(players []Player, limit int) []Team {
	// Group all players by team
	teamMap := make(map[string][]Player)
	for i := range players {
		if players[i].Club != "" {
			teamMap[players[i].Club] = append(teamMap[players[i].Club], players[i])
		}
	}

	teams := make([]Team, 0, len(teamMap))
	for teamName, teamPlayers := range teamMap {
		// Skip teams with too few players (likely data issues)
		if len(teamPlayers) < 5 {
			continue
		}

		team := Team{
			Name:        teamName,
			Division:    teamPlayers[0].Division, // Use first player's division
			PlayerCount: len(teamPlayers),
			Players:     teamPlayers,
		}

		ratings := calculateTeamRatings(teamPlayers)
		team.BestOverall = ratings.BestOverall
		team.AttRating = ratings.AttRating
		team.MidRating = ratings.MidRating
		team.DefRating = ratings.DefRating

		teams = append(teams, team)
	}

	// Sort teams by overall rating (best first)
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].BestOverall > teams[j].BestOverall
	})

	// Return top N teams
	if limit > 0 && len(teams) > limit {
		return teams[:limit]
	}

	return teams
}

// topTeamsHandler returns the top teams across all divisions
func topTeamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.top-teams.get")
	defer span.End()

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/top-teams"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/top-teams/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		WriteErrorResponse(w, r, "missing_parameters", "Dataset ID is required in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	// Get limit from query parameter, default to 100
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logInfo(ctx, "Processing top teams request", "dataset_id", datasetID, "limit", limit)

	// Try to get top teams data from cache first
	cacheKey := fmt.Sprintf("top_teams_%s_%d", datasetID, limit)
	if cached, found := getFromMemCache(cacheKey); found {
		if teamsData, ok := cached.([]Team); ok {
			logInfo(ctx, "Retrieved top teams data from memory cache", "dataset_id", datasetID, "limit", limit)

			// Set CORS headers
			setCORSHeaders(w, r)

			if supportsProtobuf {
				// Create protobuf response with full team data
				jsonData, err := json.Marshal(teamsData)
				if err == nil {
					protoResponse := &pb.GenericResponse{
						Data: string(jsonData),
					}

					// Serialize to protobuf
					responseBytes, err := serializer.Serialize(protoResponse)
					if err == nil {
						// Protobuf serialization successful
						w.Header().Set("Content-Type", serializer.ContentType())
						w.Header().Set("X-Cache-Source", "memory")
						w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
						if serializer.ShouldCompress() {
							w.Header().Set("Content-Encoding", "gzip")
						}

						if _, writeErr := w.Write(responseBytes); writeErr != nil {
							logError(ctx, "Error writing protobuf response", "error", writeErr)
						}

						logDebug(ctx, "Top teams served as protobuf from cache",
							"team_count", len(teamsData),
							"response_size_bytes", len(responseBytes),
							"processing_time_ms", time.Since(startTime).Milliseconds())
						return
					}

					// Log protobuf serialization failure
					logWarn(ctx, "Protobuf serialization failed for cached top teams, falling back to JSON", "error", err)
				}
			}

			// Fallback to JSON
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
			if err := json.NewEncoder(w).Encode(teamsData); err != nil {
				WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
				logError(ctx, "Error encoding JSON response for cached top teams",
					"error", err,
					"dataset_id", datasetID,
					"limit", limit)
			}
			return
		}
	}

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Recalculate all player ratings based on the current calculation method setting
	players = RecalculateAllPlayersRatings(players)

	// Normalize MBR scores relative to the maximum (best player gets 100)
	NormalizeMBRScoresRelativeToMax(players)

	// Process top teams data
	teamsData := processTopTeamsData(players, limit)

	// Cache the result for 5 minutes
	setInMemCache(cacheKey, teamsData, 5*time.Minute)

	// Set CORS headers
	setCORSHeaders(w, r)

	if supportsProtobuf {
		// Create protobuf response with full team data
		jsonData, err := json.Marshal(teamsData)
		if err == nil {
			protoResponse := &pb.GenericResponse{
				Data: string(jsonData),
			}

			// Serialize to protobuf
			responseBytes, err := serializer.Serialize(protoResponse)
			if err == nil {
				// Protobuf serialization successful
				w.Header().Set("Content-Type", serializer.ContentType())
				w.Header().Set("X-Cache-Source", "computed")
				w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
				if serializer.ShouldCompress() {
					w.Header().Set("Content-Encoding", "gzip")
				}

				if _, writeErr := w.Write(responseBytes); writeErr != nil {
					logError(ctx, "Error writing protobuf response", "error", writeErr)
				}

				logDebug(ctx, "Top teams served as protobuf",
					"team_count", len(teamsData),
					"response_size_bytes", len(responseBytes),
					"processing_time_ms", time.Since(startTime).Milliseconds())
				return
			}

			// Log protobuf serialization failure
			logWarn(ctx, "Protobuf serialization failed for top teams, falling back to JSON", "error", err)
		}
	}

	// Fallback to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Source", "computed")
	if err := json.NewEncoder(w).Encode(teamsData); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		logError(ctx, "Error encoding JSON response for top teams", "dataset_id", sanitizeForLogging(datasetID), "limit", limit, "error", err)
		return
	}

	logInfo(ctx, "Top teams request completed",
		"dataset_id", datasetID,
		"limit", limit,
		"team_count", len(teamsData),
		"processing_time_ms", time.Since(startTime).Milliseconds())
}

// divisionsHandler returns unique divisions for a dataset
func divisionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.divisions.get")
	defer span.End()

	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	supportsProtobuf := negotiator.SupportsProtobuf()

	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	SetSpanAttributes(ctx,
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/divisions"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/divisions/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		WriteErrorResponse(w, r, "missing_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	logInfo(ctx, "Processing divisions request", "dataset_id", datasetID)

	// Try to get divisions data from cache first
	cacheKey := fmt.Sprintf("divisions_%s", datasetID)
	if cached, found := getFromMemCache(cacheKey); found {
		if divisionsData, ok := cached.([]string); ok {
			logInfo(ctx, "Retrieved divisions data from memory cache", "dataset_id", datasetID)

			// Set CORS headers
			setCORSHeaders(w, r)

			// For now, only support JSON response for divisions
			// TODO: Add protobuf support when pb.DivisionsResponse is defined

			// Fallback to JSON
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
			if err := json.NewEncoder(w).Encode(divisionsData); err != nil {
				WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
				logError(ctx, "Error encoding JSON response for cached divisions",
					"error", err,
					"dataset_id", datasetID)
				return
			}

			logDebug(ctx, "Divisions served as JSON from cache",
				"division_count", len(divisionsData),
				"processing_time_ms", time.Since(startTime).Milliseconds())
			return
		}
	}

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		WriteErrorResponse(w, r, "dataset_not_found", "Player data not found for the given ID.", nil, http.StatusNotFound)
		return
	}

	// Extract unique divisions
	divisionsMap := make(map[string]bool)
	for _, player := range players {
		if player.Division != "" {
			divisionsMap[player.Division] = true
		}
	}

	// Convert to sorted slice
	divisions := make([]string, 0, len(divisionsMap))
	for division := range divisionsMap {
		divisions = append(divisions, division)
	}
	sort.Strings(divisions)

	// Cache the result
	setInMemCache(cacheKey, divisions, 300) // Cache for 5 minutes

	// Set CORS headers
	setCORSHeaders(w, r)

	// For now, only support JSON response for divisions
	// TODO: Add protobuf support when pb.DivisionsResponse is defined

	// Fallback to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
	if err := json.NewEncoder(w).Encode(divisions); err != nil {
		WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		logError(ctx, "Error encoding JSON response for divisions",
			"error", err,
			"dataset_id", datasetID)
		return
	}

	logDebug(ctx, "Divisions served as JSON",
		"division_count", len(divisions),
		"processing_time_ms", time.Since(startTime).Milliseconds())
}

// Debug endpoint to check JSON serialization
func debugPlayerDataHandler(w http.ResponseWriter, r *http.Request) {
	datasetID := r.URL.Query().Get("dataset")
	playerName := r.URL.Query().Get("player")

	if datasetID == "" || playerName == "" {
		http.Error(w, "Missing dataset or player parameter", http.StatusBadRequest)
		return
	}

	players, _, found := GetPlayerData(datasetID)
	if !found {
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
	}

	// Find the specific player
	for _, player := range players {
		if strings.Contains(strings.ToLower(player.Name), strings.ToLower(playerName)) {
			// Return the raw JSON for this player
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(player)
			return
		}
	}

	http.Error(w, "Player not found", http.StatusNotFound)
}

// logTop25OverallPlayers logs the top 25 players by overall rating with their MBR breakdown
func logTop25OverallPlayers(players []Player) {
	// Create a proper deep copy of players for sorting to prevent any potential race conditions
	// PERFORMANCE: Use FastDeepCopyPlayers for much better performance while staying thread-safe
	playersCopy := FastDeepCopyPlayers(players)

	// Sort by overall rating in descending order
	sort.Slice(playersCopy, func(i, j int) bool {
		return playersCopy[i].Overall > playersCopy[j].Overall
	})

	// Log top 25
	LogInfo("=== TOP 25 OVERALL RATING PLAYERS (WITH MBR BREAKDOWN) ===")
	for i := 0; i < 25 && i < len(playersCopy); i++ {
		player := playersCopy[i]
		ageInt, _ := strconv.Atoi(player.Age)
		valueScore := getExpectedValuePerRating(float64(player.Overall))
		baseRating := player.Overall / 2
		ageModifier := getAgeModifier(ageInt)
		mentalityModifier := getMentalityModifier(player.Personality)
		valueScoreContribution := int(valueScore * 1.5)

		// Calculate value score for logging
		var calculatedValueScore float64
		overall := float64(player.Overall)
		transferValueMillions := float64(player.TransferValueAmount) / 1000000.0

		if transferValueMillions > 0 {
			valuePerRating := transferValueMillions / overall
			logValuePerRating := math.Log10(valuePerRating + 1)
			baseEfficiency := overall / (logValuePerRating + 1)

			switch {
			case overall >= 80:
				calculatedValueScore = baseEfficiency * 1.2
			case overall >= 70:
				calculatedValueScore = baseEfficiency * 1.0
			case overall >= 60:
				calculatedValueScore = baseEfficiency * 0.9
			case overall >= 55:
				calculatedValueScore = baseEfficiency * 0.8
			default:
				calculatedValueScore = baseEfficiency * 0.6
			}

			expectedValuePerRating := getExpectedValuePerRating(overall)
			if valuePerRating < expectedValuePerRating*0.7 {
				calculatedValueScore *= 1.3
			} else if valuePerRating < expectedValuePerRating*0.85 {
				calculatedValueScore *= 1.15
			}
		}

		// Calculate salary penalty for logging
		salaryPenalty := getSalaryPenalty(player.TransferValueAmount, player.WageAmount)

		// Calculate transfer value penalty for logging
		var transferValuePenalty int
		if player.TransferValueAmount > 0 {
			transferValueMillions := float64(player.TransferValueAmount) / 1000000.0
			valuePerRating := transferValueMillions / overall
			expectedValuePerRating := getExpectedValuePerRating(overall)
			priceMultiplier := valuePerRating / expectedValuePerRating

			switch {
			case priceMultiplier >= 5.0:
				transferValuePenalty = -40
			case priceMultiplier >= 4.0:
				transferValuePenalty = -30
			case priceMultiplier >= 3.0:
				transferValuePenalty = -20
			case priceMultiplier >= 2.5:
				transferValuePenalty = -15
			case priceMultiplier >= 2.0:
				transferValuePenalty = -10
			case priceMultiplier >= 1.5:
				transferValuePenalty = -5
			case priceMultiplier <= 0.5:
				transferValuePenalty = 5
			case priceMultiplier <= 0.7:
				transferValuePenalty = 2
			default:
				transferValuePenalty = 0
			}
		}

		LogInfo("Rank %d: %s (Age: %d, Overall: %d, Club: %s) - MBR: %d",
			i+1, player.Name, ageInt, player.Overall, player.Club, player.MBR)
		LogInfo("  Breakdown: Base=%d, Age=%d, Mentality=%d, ValueScore=%.2f (contribution=%d), TransferValuePenalty=%d, SalaryPenalty=%d",
			baseRating, ageModifier, mentalityModifier, calculatedValueScore, valueScoreContribution, transferValuePenalty, salaryPenalty)
		LogInfo("  Transfer Value: %s (£%d), Wage: £%d/week", player.TransferValue, player.TransferValueAmount, player.WageAmount)
	}
	LogInfo("=== END TOP 25 OVERALL RATING PLAYERS ===")
}
