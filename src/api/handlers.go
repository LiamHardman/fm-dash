// src/api/handlers.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// setCORSHeaders sets secure CORS headers based on the request origin
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	
	// Define allowed origins for production
	allowedOrigins := []string{
		"http://localhost:3000",   // Development frontend
		"http://localhost:8080",   // Production nginx
		"https://localhost:8080",  // Production nginx with SSL
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

// processPercentilesAsync calculates percentiles asynchronously and updates stored dataset
func processPercentilesAsync(datasetID string, players []Player) {
	go func() {
		log.Printf("Starting async percentile calculation for dataset %s with %d players", datasetID, len(players))
		
		// Calculate percentiles
		CalculatePlayerPerformancePercentiles(players)
		
		// Get currency from stored data to preserve it
		_, currencySymbol, found := GetPlayerData(datasetID)
		if !found {
			log.Printf("Warning: Dataset %s not found when updating percentiles", datasetID)
			return
		}
		
		// Update stored dataset with percentiles
		SetPlayerData(datasetID, players, currencySymbol)
		log.Printf("Completed async percentile calculation and storage update for dataset %s", datasetID)
	}()
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

// Structured logging helpers with trace context
func logInfo(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func logWarn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func logError(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

// League represents a division/league with its teams
type League struct {
	Name            string `json:"name"`
	TeamCount       int    `json:"teamCount"`
	PlayerCount     int    `json:"playerCount"`
	BestOverall     int    `json:"bestOverall"`
	AttRating       int    `json:"attRating"`
	MidRating       int    `json:"midRating"`
	DefRating       int    `json:"defRating"`
}

// Team represents a team with its ratings and stats
type Team struct {
	Name            string   `json:"name"`
	Division        string   `json:"division"`
	PlayerCount     int      `json:"playerCount"`
	BestOverall     int      `json:"bestOverall"`
	AttRating       int      `json:"attRating"`
	MidRating       int      `json:"midRating"`
	DefRating       int      `json:"defRating"`
	Players         []Player `json:"players,omitempty"`
}

// getMaxUploadSize reads the MAX_UPLOAD_SIZE environment variable and returns the size in bytes.
// If not set or invalid, defaults to 15MB.
func getMaxUploadSize() int64 {
	envValue := os.Getenv("MAX_UPLOAD_SIZE")
	if envValue == "" {
		return 15 * 1024 * 1024 // Default 15MB
	}
	
	// Parse as integer (expecting value in MB)
	sizeInMB, err := strconv.Atoi(envValue)
	if err != nil || sizeInMB <= 0 {
		log.Printf("Invalid MAX_UPLOAD_SIZE environment variable '%s', defaulting to 15MB", envValue)
		return 15 * 1024 * 1024 // Default 15MB
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
		if span.SpanContext().IsValid() {
			// Check if there was an error by inspecting span status
			// We'll update this in error cases below
		}
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
	// We'll check the actual file handler size after getting the file.
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
	defer file.Close()

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
	
	logInfo(ctx, "File uploaded", 
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

	parseStartTime := time.Now()
	playersList := make([]Player, 0, defaultPlayerCapacity) // Assumes defaultPlayerCapacity is defined in config.go
	var processingError error

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
		log.Println("Finished collecting results from resultsChan.")
	}()

	// Wrap file parsing in a child span using the already-read content
	err = TraceFileProcessing(ctx, handler.Filename, actualFileSize, func(spanCtx context.Context) error {
		contentReader := strings.NewReader(string(fileContent))
		return ParseHTMLPlayerTable(contentReader, &headersSnapshot, rowCellsChan, numWorkers, resultsChan, &wg)
	})
	processingError = err

	close(rowCellsChan)
	log.Println("Row cells channel closed (HTML parsing attempt finished).")

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
	log.Println("Waiting for all player data parser workers to finish...")
	wg.Wait()
	AddSpanEvent(ctx, "workers.completed")
	log.Println("All workers have completed (wg.Wait() returned).")

	close(resultsChan)
	log.Println("ResultsChan closed after all workers finished.")

	<-doneConsumingResults
	log.Println("Results consumer goroutine finished processing all items.")

	finalDatasetCurrencySymbol := "$" // Default
	if len(playersList) > 0 {
		var foundSymbol bool
		for _, p := range playersList {
			_, _, tvSymbol := ParseMonetaryValueGo(p.TransferValue) // Assumes ParseMonetaryValueGo is in parsing.go
			if tvSymbol != "" {
				finalDatasetCurrencySymbol = tvSymbol
				foundSymbol = true
				break
			}
			_, _, wSymbol := ParseMonetaryValueGo(p.Wage)
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

	// Store using the new storage interface (with MinIO support and fallback)
	ctx, storageSpan := StartSpan(ctx, "storage.save_dataset")
	SetSpanAttributes(ctx, 
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(playersList)),
		attribute.String("dataset.currency", finalDatasetCurrencySymbol),
	)
	SetPlayerData(datasetID, playersList, finalDatasetCurrencySymbol)
	storageSpan.End()

	// Process percentiles asynchronously to avoid blocking the upload response
	if len(playersList) > 0 {
		// Make a deep copy for async processing to avoid data races
		playersListCopy := make([]Player, len(playersList))
		for i, player := range playersList {
			// Deep copy each player to avoid concurrent map access
			playersListCopy[i] = Player{
				Name:                    player.Name,
				Position:                player.Position,
				Age:                     player.Age,
				Club:                    player.Club,
				Division:                player.Division,
				TransferValue:           player.TransferValue,
				Wage:                    player.Wage,
				Personality:             player.Personality,
				MediaHandling:           player.MediaHandling,
				Nationality:             player.Nationality,
				NationalityISO:          player.NationalityISO,
				NationalityFIFACode:     player.NationalityFIFACode,
				TransferValueAmount:     player.TransferValueAmount,
				WageAmount:              player.WageAmount,
				PAC:                     player.PAC,
				SHO:                     player.SHO,
				PAS:                     player.PAS,
				DRI:                     player.DRI,
				DEF:                     player.DEF,
				PHY:                     player.PHY,
				GK:                      player.GK,
				DIV:                     player.DIV,
				HAN:                     player.HAN,
				REF:                     player.REF,
				KIC:                     player.KIC,
				SPD:                     player.SPD,
				POS:                     player.POS,
				Overall:                 player.Overall,
				// Deep copy maps
				Attributes:              make(map[string]string),
				NumericAttributes:       make(map[string]int),
				PerformanceStatsNumeric: make(map[string]float64),
				PerformancePercentiles:  make(map[string]map[string]float64),
				// Copy slices
				ParsedPositions:      append([]string(nil), player.ParsedPositions...),
				ShortPositions:       append([]string(nil), player.ShortPositions...),
				PositionGroups:       append([]string(nil), player.PositionGroups...),
				RoleSpecificOveralls: append([]RoleOverallScore(nil), player.RoleSpecificOveralls...),
			}
			
			// Copy map contents
			for k, v := range player.Attributes {
				playersListCopy[i].Attributes[k] = v
			}
			for k, v := range player.NumericAttributes {
				playersListCopy[i].NumericAttributes[k] = v
			}
			for k, v := range player.PerformanceStatsNumeric {
				playersListCopy[i].PerformanceStatsNumeric[k] = v
			}
			for k, v := range player.PerformancePercentiles {
				playersListCopy[i].PerformancePercentiles[k] = make(map[string]float64)
				for kk, vv := range v {
					playersListCopy[i].PerformancePercentiles[k][kk] = vv
				}
			}
		}
		processPercentilesAsync(datasetID, playersListCopy)
	}

	logInfo(ctx, "Player data stored successfully", 
		"dataset_id", datasetID,
		"player_count", len(playersList),
		"detected_currency", finalDatasetCurrencySymbol)
	if len(playersList) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", playersList[0].Name, playersList[0].Overall, playersList[0].ParsedPositions, playersList[0].ShortPositions, playersList[0].PositionGroups)
	} else {
		log.Println("No players were successfully parsed or list is empty after processing.")
	}

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
		"filename": handler.Filename,
		"file_size_bytes": actualFileSize,
		"players_processed": len(playersList),
		"workers_used": numWorkers,
		"currency_detected": finalDatasetCurrencySymbol,
		"rows_per_second": rowsPerSecond,
		"memory_mb": memAllocMB,
	})
	
	// Record metrics if enabled
	recordUploadMetrics(handler.Filename, actualFileSize, totalDuration, parseDuration, 
		len(playersList), memAllocMB, numWorkers, runtime.NumGoroutine())
	
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
	logInfo(ctx, "Upload processing completed", 
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
}

// playerDataHandler handles GET requests to retrieve processed player data by dataset ID.
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
		// We'll update this in error cases below
		RecordAPIOperation(ctx, r.Method, "/api/players", status, time.Since(startTime))
	}()
	
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	
	SetSpanAttributes(ctx, 
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/players"),
	)

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		logWarn(ctx, "Dataset ID missing in request path")
		SetSpanAttributes(ctx, attribute.String("error.type", "missing_dataset_id"))
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
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

	logInfo(ctx, "Processing player data request",
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

	// Use the new storage interface to get player data
	ctx, dataSpan := StartSpan(ctx, "storage.get_dataset")
	players, currencySymbol, found := GetPlayerData(datasetID)
	dataSpan.End()

	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		SetSpanAttributes(ctx, attribute.String("error.type", "dataset_not_found"))
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	SetSpanAttributes(ctx, 
		attribute.Int("dataset.initial_player_count", len(players)),
		attribute.String("dataset.currency", currencySymbol),
	)

	// Parse division filter
	var divisionFilter DivisionFilter = DivisionFilterAll
	switch divisionFilterStr {
	case "same":
		divisionFilter = DivisionFilterSame
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	}

	// Recalculate percentiles with division filtering if not "all"
	if divisionFilter != DivisionFilterAll {
		// Make a copy of players to avoid modifying the stored data
		playersCopy := make([]Player, len(players))
		copy(playersCopy, players)
		
		// Recalculate percentiles with division filter
		CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, targetDivision)
		players = playersCopy
	}
	
	data := struct {
		Players        []Player
		CurrencySymbol string
	}{Players: players, CurrencySymbol: currencySymbol}

	processedPlayers := make([]Player, 0, len(data.Players))

	var minAge, maxAge int = -1, -1
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

	for _, p := range data.Players {
		playerCopy := p

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
				// If the role filter is active but the player doesn't have that specific role calculated,
				// you might want to either exclude them or set their overall to 0 or a special value.
				// For now, let's assume if a role filter is active, we only want players matching that role's overall.
				// So, if not matched, their 'Overall' (which might be their best general overall) might not be relevant.
				// Depending on requirements, you might set playerCopy.Overall = 0 or skip.
				// For this example, if a role is filtered, we are primarily interested in the players' score for THAT role.
				// If they don't have that role, they are effectively 0 for that role.
				// However, the primary filtering for display happens on the frontend based on the RoleSpecificOveralls array.
				// The backend here is just adjusting the main 'Overall' field if a role is specified.
				// If no match, the original player.Overall (best general) remains.
			}
		}
		processedPlayers = append(processedPlayers, playerCopy)
	}

	logInfo(ctx, "Returning processed players", "dataset_id", datasetID, "player_count", len(processedPlayers))

	response := PlayerDataWithCurrency{Players: processedPlayers, CurrencySymbol: currencySymbol}
	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response for playerData (DatasetID: %s): %v", datasetID, err)
	}
}

// rolesHandler returns a list of all available role names.
func rolesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure config is initialized with timeout
	if err := EnsureConfigInitialized(5 * time.Second); err != nil {
		log.Printf("Configuration not ready for roles request: %v", err)
		http.Error(w, "Configuration loading, please try again", http.StatusServiceUnavailable)
		return
	}

	muRoleSpecificOverallWeights.RLock()
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()

	sort.Strings(roleNames) // Sort for consistent frontend display

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(roleNames); err != nil {
		log.Printf("Error encoding JSON response for roles: %v", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

// leaguesHandler returns league data with teams and their ratings
func leaguesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/leagues/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	logInfo(ctx, "Processing leagues request", "dataset_id", datasetID)

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Process leagues data with concurrent processing
	processor := NewConcurrentLeagueProcessor(runtime.NumCPU())
	leaguesData := processor.ProcessLeaguesAsync(ctx, players)

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(leaguesData); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		log.Printf("Error encoding JSON response for leagues (DatasetID: %s): %v", datasetID, err)
	}
}

// teamsHandler returns detailed team data for a specific league
func teamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/teams/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" || pathParts[1] == "" {
		http.Error(w, "Dataset ID and Division are required in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]
	division := pathParts[1]

	logInfo(ctx, "Processing teams request", "dataset_id", datasetID, "division", division)

	// Get player data from storage
	players, _, found := GetPlayerData(datasetID)
	if !found {
		logWarn(ctx, "Player data not found", "dataset_id", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
	}

	// Process teams data for the specific division
	teamsData := processTeamsData(players, division)

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(teamsData); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		log.Printf("Error encoding JSON response for teams (DatasetID: %s, Division: %s): %v", datasetID, division, err)
	}
}

// processLeaguesData groups players by division and calculates league statistics
func processLeaguesData(players []Player) []League {
	divisionMap := make(map[string][]Player)
	
	// Group players by division
	for _, player := range players {
		if player.Division != "" {
			divisionMap[player.Division] = append(divisionMap[player.Division], player)
		}
	}
	
	var leagues []League
	for divisionName, divisionPlayers := range divisionMap {
		league := League{
			Name:        divisionName,
			PlayerCount: len(divisionPlayers),
		}
		
		// Group players by team within this division
		teamMap := make(map[string][]Player)
		for _, player := range divisionPlayers {
			if player.Club != "" {
				teamMap[player.Club] = append(teamMap[player.Club], player)
			}
		}
		
		league.TeamCount = len(teamMap)
		
		// Calculate league ratings based on best teams
		var teamOveralls []int
		var allAttRatings []int
		var allMidRatings []int
		var allDefRatings []int
		
		for _, teamPlayers := range teamMap {
			if len(teamPlayers) >= 11 { // Only consider teams with enough players
				teamRatings := calculateTeamRatings(teamPlayers)
				if teamRatings.BestOverall > 0 {
					teamOveralls = append(teamOveralls, teamRatings.BestOverall)
					allAttRatings = append(allAttRatings, teamRatings.AttRating)
					allMidRatings = append(allMidRatings, teamRatings.MidRating)
					allDefRatings = append(allDefRatings, teamRatings.DefRating)
				}
			}
		}
		
		// League overall is the average of the top teams
		if len(teamOveralls) > 0 {
			sort.Ints(teamOveralls)
			sort.Ints(allAttRatings)
			sort.Ints(allMidRatings)
			sort.Ints(allDefRatings)
			
			// Take top 50% of teams or at least 3 teams
			topTeamsCount := len(teamOveralls) / 2
			if topTeamsCount < 3 && len(teamOveralls) >= 3 {
				topTeamsCount = 3
			} else if topTeamsCount < 1 {
				topTeamsCount = len(teamOveralls)
			}
			
			league.BestOverall = calculateAverage(teamOveralls[len(teamOveralls)-topTeamsCount:])
			league.AttRating = calculateAverage(allAttRatings[len(allAttRatings)-topTeamsCount:])
			league.MidRating = calculateAverage(allMidRatings[len(allMidRatings)-topTeamsCount:])
			league.DefRating = calculateAverage(allDefRatings[len(allDefRatings)-topTeamsCount:])
		}
		
		leagues = append(leagues, league)
	}
	
	// Sort leagues by overall rating
	sort.Slice(leagues, func(i, j int) bool {
		return leagues[i].BestOverall > leagues[j].BestOverall
	})
	
	return leagues
}

// processTeamsData returns detailed team data for a specific division
func processTeamsData(players []Player, division string) []Team {
	// Filter players by division
	var divisionPlayers []Player
	for _, player := range players {
		if player.Division == division {
			divisionPlayers = append(divisionPlayers, player)
		}
	}
	
	// Group by team
	teamMap := make(map[string][]Player)
	for _, player := range divisionPlayers {
		if player.Club != "" {
			teamMap[player.Club] = append(teamMap[player.Club], player)
		}
	}
	
	var teams []Team
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
	PlayerName      string `json:"playerName"`
	DivisionFilter  string `json:"divisionFilter"`
	TargetDivision  string `json:"targetDivision"`
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
	for i, player := range players {
		if player.Name == req.PlayerName {
			targetPlayerIndex = i
			break
		}
	}

	if targetPlayerIndex == -1 {
		http.Error(w, "Player not found in dataset", http.StatusNotFound)
		return
	}

	// Parse division filter
	var divisionFilter DivisionFilter = DivisionFilterAll
	switch req.DivisionFilter {
	case "same":
		divisionFilter = DivisionFilterSame
	case "top5":
		divisionFilter = DivisionFilterTop5
	case "all", "":
		divisionFilter = DivisionFilterAll
	}

	// Create a copy of the dataset and recalculate percentiles with division filtering
	playersCopy := make([]Player, len(players))
	copy(playersCopy, players)
	
	CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, req.TargetDivision)

	// Return only the updated percentiles for the target player
	updatedPercentiles := playersCopy[targetPlayerIndex].PerformancePercentiles

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(updatedPercentiles); err != nil {
		log.Printf("Error encoding JSON response for percentiles (DatasetID: %s): %v", datasetID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
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
	for _, player := range bestXI {
		totalOverall += player.Overall
	}
	bestOverall := totalOverall / 11
	
	// Calculate section ratings based on positions
	var attPlayers, midPlayers, defPlayers []Player
	
	for _, player := range bestXI {
		// Categorize based on position groups
		isAttacker := false
		isMidfielder := false
		isDefender := false
		
		for _, posGroup := range player.PositionGroups {
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
			attPlayers = append(attPlayers, player)
		}
		if isMidfielder {
			midPlayers = append(midPlayers, player)
		}
		if isDefender {
			defPlayers = append(defPlayers, player)
		}
	}
	
	// Calculate section averages
	attRating := 0
	if len(attPlayers) > 0 {
		var attSum int
		for _, player := range attPlayers {
			attSum += player.Overall
		}
		attRating = attSum / len(attPlayers)
	}
	
	midRating := 0
	if len(midPlayers) > 0 {
		var midSum int
		for _, player := range midPlayers {
			midSum += player.Overall
		}
		midRating = midSum / len(midPlayers)
	}
	
	defRating := 0
	if len(defPlayers) > 0 {
		var defSum int
		for _, player := range defPlayers {
			defSum += player.Overall
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
}

// searchHandler handles GET requests for searching players, teams, leagues, and nations
func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Start tracing
	ctx, span := StartSpan(ctx, "search.handler")
	defer span.End()

	// Track active requests
	IncrementActiveRequests(ctx, "/search")
	defer DecrementActiveRequests(ctx, "/search")

	// Extract dataset ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[0] != "api" || pathParts[1] != "search" {
		http.Error(w, "Invalid URL format. Expected /api/search/{datasetId}", http.StatusBadRequest)
		return
	}

	datasetID := pathParts[2]
	if datasetID == "" {
		http.Error(w, "Dataset ID is required", http.StatusBadRequest)
		return
	}

	// Get search query from URL parameters
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		// Return empty results for empty query
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SearchResult{})
		return
	}

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("search.query", query),
	)

	// Get player data
	players, _, found := GetPlayerData(datasetID)
	if !found {
		http.Error(w, "Dataset not found", http.StatusNotFound)
		return
	}

	logInfo(ctx, "Performing search", "dataset_id", datasetID, "query", query, "player_count", len(players))

	// Perform search
	results := performSearch(players, query)

	SetSpanAttributes(ctx, attribute.Int("search.results_count", len(results)))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		RecordError(ctx, err, "Failed to encode search results")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	logInfo(ctx, "Search completed", "results_count", len(results))
}

// performSearch searches through players, teams, leagues, and nations
func performSearch(players []Player, query string) []SearchResult {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	
	// Collect unique teams, leagues, and nations
	teams := make(map[string]struct {
		division string
		players  int
	})
	leagues := make(map[string]int)  // league -> player count
	nations := make(map[string]int)  // nation -> player count

	// Search players and collect team/league/nation data
	for _, player := range players {
		// Search players by name
		if strings.Contains(strings.ToLower(player.Name), queryLower) {
			results = append(results, SearchResult{
				ID:          player.Name, // Using name as ID for players
				Name:        player.Name,
				Type:        "player",
				Description: fmt.Sprintf("%s • %s", player.Club, player.Division),
				URL:         fmt.Sprintf("/dataset/%s?search=%s", "", player.Name), // Frontend will fill dataset ID
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
				URL:         fmt.Sprintf("/leagues?league=%s", leagueName),
			})
		}
	}

	// Search nations
	for nationName, playerCount := range nations {
		if strings.Contains(strings.ToLower(nationName), queryLower) {
			results = append(results, SearchResult{
				ID:          nationName,
				Name:        nationName,
				Type:        "nation",
				Description: fmt.Sprintf("%d players", playerCount),
				URL:         fmt.Sprintf("/nations?nation=%s", nationName),
			})
		}
	}

	// Sort results by type priority (players first, then teams, leagues, nations) and then by name
	sort.Slice(results, func(i, j int) bool {
		typePriority := map[string]int{"player": 1, "team": 2, "league": 3, "nation": 4}
		if typePriority[results[i].Type] != typePriority[results[j].Type] {
			return typePriority[results[i].Type] < typePriority[results[j].Type]
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	// Limit results to avoid overwhelming the UI
	maxResults := 50
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// configHandler returns configuration values to the frontend
func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	config := struct {
		MaxUploadSizeMB int64 `json:"maxUploadSizeMB"`
		MaxUploadSizeBytes int64 `json:"maxUploadSizeBytes"`
	}{
		MaxUploadSizeMB: getMaxUploadSize() / (1024 * 1024),
		MaxUploadSizeBytes: getMaxUploadSize(),
	}

	w.Header().Set("Content-Type", "application/json")
	setCORSHeaders(w, r)
	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Error encoding JSON response for config: %v", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

