package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// uploadHandler handles POST requests for uploading HTML player files.
// It parses the file, processes player data concurrently, and stores the results.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	startTime := time.Now()

	// Limit file size (e.g., 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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

	parseStartTime := time.Now()
	playersList := make([]Player, 0, defaultPlayerCapacity) // From config.go
	var processingError error

	numWorkers := runtime.NumCPU()
	if numWorkers == 0 {
		numWorkers = 1 // Ensure at least one worker
	}
	const rowBufferMultiplier = 10 // Buffer size for channels
	rowCellsChan := make(chan []string, numWorkers*rowBufferMultiplier)
	resultsChan := make(chan PlayerParseResult, numWorkers*rowBufferMultiplier)
	var wg sync.WaitGroup // To wait for all parser workers to complete

	var headersSnapshot []string // To store the definitive headers for workers

	// Goroutine to collect results from workers
	doneConsumingResults := make(chan struct{})
	go func() {
		defer close(doneConsumingResults)
		for result := range resultsChan {
			if result.Err == nil {
				playersList = append(playersList, result.Player)
			} else {
				// Log errors from workers, but don't necessarily stop the whole process for one bad row
				log.Printf("Skipping row due to error from worker: %v", result.Err)
			}
		}
		log.Println("Finished collecting results from resultsChan.")
	}()

	// Start HTML parsing (from parsing.go)
	// ParseHTMLPlayerTable will send data to rowCellsChan and start workers via PlayerParserWorker
	processingError = ParseHTMLPlayerTable(file, &headersSnapshot, rowCellsChan, numWorkers, resultsChan, &wg)

	// Close rowCellsChan after ParseHTMLPlayerTable has finished attempting to send all rows.
	// This signals to workers that no more rows will be sent.
	close(rowCellsChan)
	log.Println("Row cells channel closed (HTML parsing attempt finished).")

	if processingError != nil {
		log.Printf("Error during HTML parsing or worker setup: %v", processingError)
		// Ensure workers are waited for even if there was a setup error,
		// in case some were started before the error was caught.
		// wg.Wait() might be problematic if workers never started, so check headersSnapshot.
		if len(headersSnapshot) > 0 { // Implies workers were likely started or attempted
			log.Println("Waiting for any potentially started workers after parsing error...")
			wg.Wait()
		}
		close(resultsChan) // Close results channel to unblock collector goroutine
		<-doneConsumingResults
		http.Error(w, processingError.Error(), http.StatusInternalServerError)
		return
	}

	if len(headersSnapshot) == 0 { // If no headers were found, it's a critical parsing failure
		log.Println("Critical: No headers were parsed from the HTML file.")
		close(resultsChan) // Close results channel
		<-doneConsumingResults
		http.Error(w, "Could not parse table headers, no data processed.", http.StatusInternalServerError)
		return
	}

	log.Println("Waiting for all player data parser workers to finish...")
	wg.Wait() // Wait for all PlayerParserWorker goroutines to complete
	log.Println("All workers have completed (wg.Wait() returned).")

	// Close resultsChan after all workers are done. This signals the collector goroutine to finish.
	close(resultsChan)
	log.Println("ResultsChan closed after all workers finished.")

	<-doneConsumingResults // Wait for the collector goroutine to process all items from resultsChan
	log.Println("Results consumer goroutine finished processing all items.")

	// Determine currency symbol from parsed data (if any players were parsed)
	finalDatasetCurrencySymbol := "$" // Default
	if len(playersList) > 0 {
		var foundSymbol bool
		// Check transfer value and wage of the first few players for a symbol
		for _, p := range playersList {
			_, _, tvSymbol := ParseMonetaryValueGo(p.TransferValue) // Use the raw string
			if tvSymbol != "" {
				finalDatasetCurrencySymbol = tvSymbol
				foundSymbol = true
				break
			}
			_, _, wSymbol := ParseMonetaryValueGo(p.Wage) // Use the raw string
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

	// Calculate performance percentiles after all players are parsed and enhanced
	if len(playersList) > 0 {
		log.Println("Calculating player performance percentiles...")
		CalculatePlayerPerformancePercentiles(playersList) // From performance_stats.go
		log.Println("Finished calculating percentiles.")
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

	// Store the processed player data (from store.go)
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{Players: playersList, CurrencySymbol: finalDatasetCurrencySymbol}
	storeMutex.Unlock()

	log.Printf("Stored %d players with DatasetID: %s. Detected Currency: %s", len(playersList), datasetID, finalDatasetCurrencySymbol)
	if len(playersList) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", playersList[0].Name, playersList[0].Overall, playersList[0].ParsedPositions, playersList[0].ShortPositions, playersList[0].PositionGroups)
		if globalPercentiles, ok := playersList[0].PerformancePercentiles["Global"]; ok && len(globalPercentiles) > 0 {
			log.Printf("DEBUG: Sample Player 1 Performance Percentile Keys (Global): %v", GetMapKeysStringFloat(globalPercentiles))
		}
	} else {
		log.Println("No players were successfully parsed or list is empty after processing.")
	}

	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Consider making this configurable
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response for upload: %v", err)
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
	}

	// Performance logging
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(playersList)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)
	log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Parsed Players: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---",
		handler.Filename, fileSize/1024, totalDuration, parseDuration, len(playersList), rowsPerSecond, BToMb(memStats.Alloc), numWorkers, runtime.NumGoroutine())
}

// playerDataHandler handles GET requests to retrieve processed player data by dataset ID.
func playerDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract dataset ID from URL path, e.g., /api/players/{datasetID}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	storeMutex.RLock() // Use RLock for read access
	data, found := playerDataStore[datasetID]
	storeMutex.RUnlock()

	if !found {
		log.Printf("Player data not found for DatasetID: %s", datasetID)
		http.Error(w, "Player data not found for the given ID. It might have expired, been cleared, or the ID is incorrect.", http.StatusNotFound)
		return
	}

	response := PlayerDataWithCurrency{Players: data.Players, CurrencySymbol: data.CurrencySymbol}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Consider configurability
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response for playerData (DatasetID: %s): %v", datasetID, err)
		// Don't write http.Error here if headers have already been partially written by NewEncoder
	}
}
