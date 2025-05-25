// src/api/handlers.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime" // Standard Go runtime package
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxUploadSize defines the maximum allowed file size for uploads (15MB)
	// This is an approximation for about 10,000 players.
	MaxUploadSize = 15 * 1024 * 1024
	// User-facing error message for file size limit.
	FileSizeLimitErrorMessage = "Only 10,000 players or less can be in a given dataset. (Max file size: 15MB)"
)

// uploadHandler handles POST requests for uploading HTML player files.
// It parses the file, processes player data concurrently, and stores the results.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	startTime := time.Now()

	// Check Content-Length header first for a quick check, though it can be spoofed.
	// r.ContentLength is an int64
	if r.ContentLength > MaxUploadSize {
		log.Printf("Upload rejected: Content-Length (%d bytes) exceeds limit (%d bytes)", r.ContentLength, MaxUploadSize)
		http.Error(w, FileSizeLimitErrorMessage, http.StatusRequestEntityTooLarge)
		return
	}

	// ParseMultipartForm will also respect the maxMemory argument for in-memory parts,
	// but the total request size is what we're primarily concerned with for the file part.
	// We'll check the actual file handler size after getting the file.
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB for other form data, not the file itself immediately
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("playerFile")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileSize := handler.Size
	log.Printf("Uploaded File: %s (Size: %d bytes)", handler.Filename, fileSize)

	// Enforce the 15MB limit on the actual file size
	if fileSize > MaxUploadSize {
		log.Printf("Upload rejected: Actual file size (%d bytes) exceeds limit (%d bytes)", fileSize, MaxUploadSize)
		http.Error(w, FileSizeLimitErrorMessage, http.StatusRequestEntityTooLarge)
		return
	}

	parseStartTime := time.Now()
	playersList := make([]Player, 0, defaultPlayerCapacity) // Assumes defaultPlayerCapacity is defined in config.go
	var processingError error

	numWorkers := runtime.NumCPU()
	if numWorkers == 0 {
		numWorkers = 1
	}
	const rowBufferMultiplier = 10
	rowCellsChan := make(chan []string, numWorkers*rowBufferMultiplier)
	resultsChan := make(chan PlayerParseResult, numWorkers*rowBufferMultiplier)
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

	processingError = ParseHTMLPlayerTable(file, &headersSnapshot, rowCellsChan, numWorkers, resultsChan, &wg) // Assumes ParseHTMLPlayerTable is in parsing.go

	close(rowCellsChan)
	log.Println("Row cells channel closed (HTML parsing attempt finished).")

	if processingError != nil {
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
		log.Println("Critical: No headers were parsed from the HTML file.")
		close(resultsChan)
		<-doneConsumingResults
		http.Error(w, "Could not parse table headers, no data processed.", http.StatusInternalServerError)
		return
	}

	log.Println("Waiting for all player data parser workers to finish...")
	wg.Wait()
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

	if len(playersList) > 0 {
		log.Println("Calculating player performance percentiles...")
		CalculatePlayerPerformancePercentiles(playersList) // Assumes CalculatePlayerPerformancePercentiles is in performance_stats.go
		log.Println("Finished calculating percentiles.")
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

	// Store using the new storage interface (with MinIO support and fallback)
	SetPlayerData(datasetID, playersList, finalDatasetCurrencySymbol)

	log.Printf("Stored %d players with DatasetID: %s. Detected Currency: %s", len(playersList), datasetID, finalDatasetCurrencySymbol)
	if len(playersList) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", playersList[0].Name, playersList[0].Overall, playersList[0].ParsedPositions, playersList[0].ShortPositions, playersList[0].PositionGroups)
	} else {
		log.Println("No players were successfully parsed or list is empty after processing.")
	}

	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS is set if frontend is on a different domain/port
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response for upload: %v", err)
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(playersList)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)
	memAllocMB := BToMb(memStats.Alloc) // Assumes BToMb is defined in utils.go
	
	// Record metrics if enabled
	recordUploadMetrics(handler.Filename, fileSize, totalDuration, parseDuration, 
		len(playersList), memAllocMB, numWorkers, runtime.NumGoroutine())
	
	// Keep existing log for backwards compatibility
	log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Parsed Players: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---",
		handler.Filename, fileSize/1024, totalDuration, parseDuration, len(playersList), rowsPerSecond, memAllocMB, numWorkers, runtime.NumGoroutine())
}

// playerDataHandler handles GET requests to retrieve processed player data by dataset ID.
func playerDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	queryValues := r.URL.Query()
	filterPosition := queryValues.Get("position")
	filterRole := queryValues.Get("role")
	minAgeStr := queryValues.Get("minAge")
	maxAgeStr := queryValues.Get("maxAge")
	minTransferValueStr := queryValues.Get("minTransferValue")
	maxTransferValueStr := queryValues.Get("maxTransferValue")
	maxSalaryStr := queryValues.Get("maxSalary")

	log.Printf("playerDataHandler: DatasetID=%s, PositionFilter=%s, RoleFilter=%s, MinAge=%s, MaxAge=%s, MinVal=%s, MaxVal=%s, MaxSalary=%s",
		datasetID, filterPosition, filterRole, minAgeStr, maxAgeStr, minTransferValueStr, maxTransferValueStr, maxSalaryStr)

	// Use the new storage interface to get player data
	players, currencySymbol, found := GetPlayerData(datasetID)
	if !found {
		log.Printf("Player data not found for DatasetID: %s", datasetID)
		http.Error(w, "Player data not found for the given ID.", http.StatusNotFound)
		return
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

	log.Printf("playerDataHandler: Returning %d players after processing for DatasetID=%s", len(processedPlayers), datasetID)

	response := PlayerDataWithCurrency{Players: processedPlayers, CurrencySymbol: data.CurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS
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

	// Assumes muRoleSpecificOverallWeights and roleSpecificOverallWeights are defined in config.go
	muRoleSpecificOverallWeights.RLock()
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()

	sort.Strings(roleNames) // Sort for consistent frontend display

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS
	if err := json.NewEncoder(w).Encode(roleNames); err != nil {
		log.Printf("Error encoding JSON response for roles: %v", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

