package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		log.Printf("Cleaned up %d stale duplicate mappings", len(staleMappings))
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
// If not set or invalid, defaults to 15MB.
func getMaxUploadSize() int64 {
	envValue := os.Getenv("MAX_UPLOAD_SIZE")
	if envValue == "" {
		return 20 * 1024 * 1024 // Default 15MB
	}

	// Parse as integer (expecting value in MB)
	sizeInMB, err := strconv.Atoi(envValue)
	if err != nil || sizeInMB <= 0 {
		log.Printf("Invalid MAX_UPLOAD_SIZE environment variable '%s', defaulting to 20MB", envValue)
		return 20 * 1024 * 1024 // Default 15MB
	}

	return int64(sizeInMB) * 1024 * 1024
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

	parseStartTime := time.Now()
	// Optimized pre-allocation based on file size estimation
	estimatedPlayerCount := int(actualFileSize / 2048) // Rough estimation: ~2KB per player row
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
				log.Printf("Skipping row due to error from worker: %v", result.Err)
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
		log.Printf("Error during HTML parsing or worker setup: %v", processingError)
		if len(headersSnapshot) > 0 {
			log.Println("Waiting for any potentially started workers after parsing error...")
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
			log.Println("No currency symbol detected from parsed player monetary values, using default '$'.")
		}
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

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
		"detected_currency", finalDatasetCurrencySymbol)

	// Send response immediately - percentiles will be calculated asynchronously
	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)

	ctx, responseSpan := StartSpan(ctx, "response.encode")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		RecordError(ctx, err, "Failed to encode JSON response")
		log.Printf("Error encoding JSON response for upload: %v", err)
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	responseSpan.End()

	// Calculate percentiles asynchronously after response is sent
	go func() {
		percentileCtx := context.Background()
		percentileCtx, percentileSpan := StartSpan(percentileCtx, "percentiles.calculate_upload_async")
		defer percentileSpan.End()

		SetSpanAttributes(percentileCtx,
			attribute.String("dataset.id", datasetID),
			attribute.Int("dataset.player_count", len(playersList)),
			attribute.String("operation.type", "async_percentile_calculation"),
		)

		startTime := time.Now()
		logDebug(percentileCtx, "Starting async percentile calculation after response sent",
			"dataset_id", datasetID,
			"player_count", len(playersList))

		// Get the stored data and calculate percentiles
		storedPlayers, storedCurrency, found := GetPlayerData(datasetID)
		if !found {
			LogWarn("Could not retrieve stored dataset %s for percentile calculation", sanitizeForLogging(datasetID))
			return
		}

		// Make a deep copy to avoid race conditions during calculation
		// Use the original deep copy to avoid any concurrency issues with optimizations
		playersCopyForPercentiles := make([]Player, len(storedPlayers))
		for i := range storedPlayers {
			playersCopyForPercentiles[i] = storedPlayers[i]
			// Deep copy all the maps to prevent race conditions
			if storedPlayers[i].Attributes != nil {
				playersCopyForPercentiles[i].Attributes = make(map[string]string)
				for k, v := range storedPlayers[i].Attributes {
					playersCopyForPercentiles[i].Attributes[k] = v
				}
			}
			if storedPlayers[i].NumericAttributes != nil {
				playersCopyForPercentiles[i].NumericAttributes = make(map[string]int)
				for k, v := range storedPlayers[i].NumericAttributes {
					playersCopyForPercentiles[i].NumericAttributes[k] = v
				}
			}
			if storedPlayers[i].PerformanceStatsNumeric != nil {
				playersCopyForPercentiles[i].PerformanceStatsNumeric = make(map[string]float64)
				for k, v := range storedPlayers[i].PerformanceStatsNumeric {
					playersCopyForPercentiles[i].PerformanceStatsNumeric[k] = v
				}
			}
			if storedPlayers[i].PerformancePercentiles != nil {
				playersCopyForPercentiles[i].PerformancePercentiles = make(map[string]map[string]float64)
				for group, stats := range storedPlayers[i].PerformancePercentiles {
					playersCopyForPercentiles[i].PerformancePercentiles[group] = make(map[string]float64)
					for stat, value := range stats {
						playersCopyForPercentiles[i].PerformancePercentiles[group][stat] = value
					}
				}
			}
			if storedPlayers[i].RoleSpecificOveralls != nil {
				playersCopyForPercentiles[i].RoleSpecificOveralls = make([]RoleOverallScore, len(storedPlayers[i].RoleSpecificOveralls))
				copy(playersCopyForPercentiles[i].RoleSpecificOveralls, storedPlayers[i].RoleSpecificOveralls)
			}
			if storedPlayers[i].ShortPositions != nil {
				playersCopyForPercentiles[i].ShortPositions = make([]string, len(storedPlayers[i].ShortPositions))
				copy(playersCopyForPercentiles[i].ShortPositions, storedPlayers[i].ShortPositions)
			}
			if storedPlayers[i].ParsedPositions != nil {
				playersCopyForPercentiles[i].ParsedPositions = make([]string, len(storedPlayers[i].ParsedPositions))
				copy(playersCopyForPercentiles[i].ParsedPositions, storedPlayers[i].ParsedPositions)
			}
			if storedPlayers[i].PositionGroups != nil {
				playersCopyForPercentiles[i].PositionGroups = make([]string, len(storedPlayers[i].PositionGroups))
				copy(playersCopyForPercentiles[i].PositionGroups, storedPlayers[i].PositionGroups)
			}
		}

		// Calculate percentiles for all division filters to ensure stability
		CalculatePlayerPerformancePercentiles(playersCopyForPercentiles)

		// Update the stored data with calculated percentiles (use sync SetPlayerData to avoid additional async operations)
		SetPlayerData(datasetID, playersCopyForPercentiles, storedCurrency)

		duration := time.Since(startTime)
		SetSpanAttributes(percentileCtx,
			attribute.Int64("percentile_calculation.duration_ms", duration.Milliseconds()),
			attribute.String("percentile_calculation.status", "success"),
		)

		LogInfo("Completed async percentile calculation for dataset %s in %v", sanitizeForLogging(datasetID), duration)
	}()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(playersList)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)
	memAllocMB := BToMb(memStats.Alloc) // Assumes BToMb is defined in utils.go

	// Record comprehensive business operation metrics
	RecordBusinessOperation(ctx, "file_upload", true, map[string]interface{}{
		"filename":          handler.Filename,
		"file_size_bytes":   actualFileSize,
		"players_processed": len(playersList),
		"workers_used":      numWorkers,
		"currency_detected": finalDatasetCurrencySymbol,
		"rows_per_second":   rowsPerSecond,
		"memory_mb":         memAllocMB,
	})

	// Record metrics if enabled
	recordUploadMetrics(ctx, actualFileSize, totalDuration, parseDuration,
		len(playersList), memAllocMB, numWorkers)

	// Add final span attributes with performance metrics
	SetSpanAttributes(ctx,
		attribute.String("upload.status", "success"),
		attribute.String("dataset.id", datasetID),
		attribute.Int("players.processed", len(playersList)),
		attribute.Float64("performance.rows_per_second", rowsPerSecond),
		attribute.Float64("performance.memory_mb", memAllocMB),
		attribute.Int64("performance.total_duration_ms", totalDuration.Milliseconds()),
		attribute.Int64("performance.parse_duration_ms", parseDuration.Milliseconds()),
	)

	// Log performance metrics with trace context
	logDebug(ctx, "Upload processing completed",
		"filename", handler.Filename,
		"file_size_kb", actualFileSize/1024,
		"total_duration_ms", totalDuration.Milliseconds(),
		"parse_duration_ms", parseDuration.Milliseconds(),
		"players_parsed", len(playersList),
		"rows_per_second", rowsPerSecond,
		"memory_alloc_mb", memAllocMB,
		"workers", numWorkers,
		"goroutines", runtime.NumGoroutine(),
		"trace_id", GetTraceID(ctx),
		"span_id", GetSpanID(ctx))

	// Log immediate performance and memory stats after parsing completion
	LogImmediatePerformanceStats()
	LogImmediateMemoryStats(ctx)
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

		// Recalculate all player ratings based on the current calculation method setting
		ctx, recalcSpan := StartSpan(ctx, "ratings.recalculate")
		players = RecalculateAllPlayersRatings(players)
		recalcSpan.End()

		// Calculate percentiles with appropriate filtering using optimized algorithm
		ctx, percentileSpan := StartSpan(ctx, "percentiles.calculate")
		// Make a deep copy of players to avoid modifying the stored data and prevent race conditions
		// Use optimized deep copy for better memory efficiency
		playersCopy := OptimizedDeepCopyPlayers(players)

		if divisionFilter != DivisionFilterAll {
			// Recalculate percentiles with division filter
			CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, targetDivision)
		} else {
			// Calculate global percentiles using optimized algorithm
			CalculatePlayerPerformancePercentiles(playersCopy)
		}

		players = playersCopy
		percentileSpan.End()

		// Cache the percentile-calculated data for future requests
		cacheData := struct {
			Players        []Player
			CurrencySymbol string
		}{
			Players:        players,
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
				log.Printf("Error writing cached response: %v", err)
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
			log.Printf("Skipping player %s due to unparsable age '%s' while age filters are active.", playerCopy.Name, playerCopy.Age)
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
		metadata := CreateResponseMetadata(requestID, int32(len(processedPlayers)), found)
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

// rolesHandler returns a list of all available role names.
func rolesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Start comprehensive tracing
	ctx, span := StartSpan(ctx, "api.roles.get")
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
		attribute.String("http.route", "/api/roles"),
		attribute.String("response.format", serializer.ContentType()),
		attribute.Bool("client.supports_protobuf", supportsProtobuf),
		attribute.String("request.id", requestID),
	)

	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	// Ensure config is initialized with timeout
	if err := EnsureConfigInitialized(5 * time.Second); err != nil {
		logError(ctx, "Configuration not ready for roles request", "error", err)
		WriteErrorResponse(w, r, "config_not_ready", "Configuration loading, please try again", nil, http.StatusServiceUnavailable)
		return
	}

	muRoleSpecificOverallWeights.RLock()
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()

	sort.Strings(roleNames) // Sort for consistent frontend display

	// Set CORS headers
	setCORSHeaders(w, r)

	// Create response metadata
	metadata := CreateResponseMetadata(requestID, int32(len(roleNames)), false)

	if supportsProtobuf {
		// Create protobuf response
		protoResponse := &pb.RolesResponse{
			Roles:    roleNames,
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

			logDebug(ctx, "Roles served as protobuf",
				"role_count", len(roleNames),
				"response_size_bytes", len(responseBytes),
				"processing_time_ms", time.Since(startTime).Milliseconds())
			return
		}

		// Log protobuf serialization failure
		logWarn(ctx, "Protobuf serialization failed for roles, falling back to JSON", "error", err)
	}

	// Fallback to JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roleNames); err != nil {
		logError(ctx, "Error encoding JSON response for roles", "error", err)
		WriteErrorResponse(w, r, "serialization_error", "Error encoding response", nil, http.StatusInternalServerError)
		return
	}

	logDebug(ctx, "Roles served as JSON",
		"role_count", len(roleNames),
		"processing_time_ms", time.Since(startTime).Milliseconds())
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
			metadata := CreateResponseMetadata(requestID, int32(len(leaguesData)), true)

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
	metadata := CreateResponseMetadata(requestID, int32(len(leaguesData)), false)

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
			metadata := CreateResponseMetadata(requestID, int32(len(teamsData)), true)

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
		log.Printf("Error encoding JSON response for teams (DatasetID: %s, Division: %s): %v", sanitizeForLogging(datasetID), sanitizeForLogging(division), err)
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
			log.Printf("Error encoding JSON response for cached percentiles (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - perform calculation
	logDebug(ctx, "💫 CACHE MISS - calculating percentiles",
		"dataset_id", datasetID,
		"player_name", req.PlayerName,
		"division_filter", req.DivisionFilter,
		"target_division", req.TargetDivision,
		"cache_key", cacheKey)

	// Create a copy of the dataset and recalculate percentiles with division filtering
	playersCopy := make([]Player, len(players))
	copy(playersCopy, players)

	CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, req.TargetDivision)

	// Get the updated percentiles for the target player
	updatedPercentiles := playersCopy[targetPlayerIndex].PerformancePercentiles

	// NEW: Save to cache for future requests
	go func() {
		savePercentilesToCache(ctx, cacheKey, datasetID, req.PlayerName, req.DivisionFilter, req.TargetDivision, players, updatedPercentiles)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(updatedPercentiles); err != nil {
		log.Printf("Error encoding JSON response for percentiles (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
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
			log.Printf("Error encoding JSON response for cached player percentiles (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss - perform calculation
	logDebug(ctx, "💫 CACHE MISS - calculating player percentiles",
		"dataset_id", datasetID,
		"player_uid", req.PlayerUID,
		"compare_division", req.CompareDivision,
		"compare_position", req.ComparePosition,
		"cache_key", cacheKey)

	// Create a copy of the dataset and recalculate percentiles with division filtering
	playersCopy := make([]Player, len(players))
	copy(playersCopy, players)

	CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, targetDivision)

	// Find the target player in the copy
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
		log.Printf("Error encoding JSON response for player percentiles (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
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
		log.Printf("Error encoding JSON response for percentiles status (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
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
func calculateTeamRatings(players []Player) TeamRatings {
	if len(players) < 11 {
		return TeamRatings{}
	}

	// Sort players by overall rating
	sort.Slice(players, func(i, j int) bool {
		return players[i].Overall > players[j].Overall
	})

	// Take the top 11 players as the best XI
	bestXI := players[:11]

	// Calculate average overall
	var totalOverall int
	for i := range bestXI {
		totalOverall += bestXI[i].Overall
	}
	bestOverall := totalOverall / 11

	// Calculate section ratings based on positions
	// Pre-allocate with estimated capacities based on typical team formations
	attPlayers := make([]Player, 0, 4) // Typically 3-4 attackers in best XI
	midPlayers := make([]Player, 0, 4) // Typically 3-4 midfielders in best XI
	defPlayers := make([]Player, 0, 5) // Typically 4-5 defenders in best XI

	for i := range bestXI {
		// Categorize based on position groups
		isAttacker := false
		isMidfielder := false
		isDefender := false

		for _, posGroup := range bestXI[i].PositionGroups {
			switch posGroup {
			case "Attackers":
				isAttacker = true
			case "Midfielders":
				isMidfielder = true
			case "Defenders":
				isDefender = true
			}
		}

		// Assign to categories (players can be in multiple)
		if isAttacker {
			attPlayers = append(attPlayers, bestXI[i])
		}
		if isMidfielder {
			midPlayers = append(midPlayers, bestXI[i])
		}
		if isDefender {
			defPlayers = append(defPlayers, bestXI[i])
		}
	}

	// Calculate section averages
	attRating := 0
	if len(attPlayers) > 0 {
		var attSum int
		for i := range attPlayers {
			attSum += attPlayers[i].Overall
		}
		attRating = attSum / len(attPlayers)
	}

	midRating := 0
	if len(midPlayers) > 0 {
		var midSum int
		for i := range midPlayers {
			midSum += midPlayers[i].Overall
		}
		midRating = midSum / len(midPlayers)
	}

	defRating := 0
	if len(defPlayers) > 0 {
		var defSum int
		for i := range defPlayers {
			defSum += defPlayers[i].Overall
		}
		defRating = defSum / len(defPlayers)
	}

	return TeamRatings{
		BestOverall: bestOverall,
		AttRating:   attRating,
		MidRating:   midRating,
		DefRating:   defRating,
	}
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
	players, _, found := GetPlayerData(datasetID)
	if !found {
		WriteErrorResponse(w, r, "dataset_not_found", "Dataset not found", nil, http.StatusNotFound)
		return
	}

	// Recalculate all player ratings based on the current calculation method setting
	players = RecalculateAllPlayersRatings(players)

	logDebug(ctx, "Performing search", "dataset_id", datasetID, "query", query, "player_count", len(players))

	// NEW: Generate cache key and try to load from cache first
	cacheKey := generateSearchCacheKey(ctx, datasetID, query, players)

	// Try to load from cache
	if cachedResults, found := loadSearchFromCache(ctx, cacheKey, datasetID, query, players); found {
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
		metadata := CreateResponseMetadata(requestID, int32(len(cachedResults)), true)

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

	// Perform search
	results := performSearch(players, query)

	// NEW: Save to cache for future requests (only cache if results are not too large)
	if len(results) <= 1000 { // Reasonable limit to avoid caching huge result sets
		go func() {
			saveSearchToCache(ctx, cacheKey, datasetID, query, players, results)
		}()
	}

	SetSpanAttributes(ctx,
		attribute.Int("search.results_count", len(results)),
		attribute.String("search.cache_status", "MISS"))

	// Set CORS headers
	setCORSHeaders(w, r)

	// Create response metadata
	metadata := CreateResponseMetadata(requestID, int32(len(results)), false)

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
			log.Printf("Error encoding JSON response for cached bargain hunter (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
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
		log.Printf("Error encoding JSON response for bargain hunter (DatasetID: %s): %v", sanitizeForLogging(datasetID), err)
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

		// Skip free transfers entirely
		if player.TransferValueAmount == 0 {
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

	// Use optimized deep copy if memory optimizations are enabled
	if memOptConfig.UseCopyOnWrite {
		return OptimizedDeepCopyPlayers(players)
	}

	// Fallback to original implementation for compatibility
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

	// Get dataset
	players, currencySymbol, found := GetPlayerData(datasetID)
	if !found {
		logError(ctx, "Failed to get dataset for full player stats",
			"dataset_id", datasetID,
			"player_uid", playerUID)
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
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

	// Ensure the player has percentile data
	if targetPlayer.PerformancePercentiles == nil || len(targetPlayer.PerformancePercentiles) == 0 {
		logInfo(ctx, "Player missing percentiles, calculating them",
			"dataset_id", datasetID,
			"player_uid", playerUID,
			"player_name", targetPlayer.Name)

		// Create a copy of the dataset and calculate percentiles
		playersCopy := make([]Player, len(players))
		copy(playersCopy, players)

		// Calculate percentiles for the entire dataset
		CalculatePlayerPerformancePercentiles(playersCopy)

		// Find the updated player with percentiles
		for _, player := range playersCopy {
			if player.UID == playerUID {
				targetPlayer = &player
				break
			}
		}

		logInfo(ctx, "Successfully calculated percentiles for player",
			"dataset_id", datasetID,
			"player_uid", playerUID,
			"player_name", targetPlayer.Name,
			"percentile_groups", len(targetPlayer.PerformancePercentiles))
	}

	// Create response with full player data
	response := map[string]interface{}{
		"player":          targetPlayer,
		"dataset_id":      datasetID,
		"currency_symbol": currencySymbol,
	}

	// Use content negotiation to determine response format
	if err := WriteResponse(w, r, response); err != nil {
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
		"player_name", targetPlayer.Name)
}
