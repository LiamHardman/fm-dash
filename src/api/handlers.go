// src/api/handlers.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	// Ensure you have a Go runtime package or similar for NumCPU if it's not standard.
	// For this example, assuming 'runtime' is the standard Go package.
	"runtime"

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
				log.Printf("Skipping row due to error from worker: %v", result.Err)
			}
		}
		log.Println("Finished collecting results from resultsChan.")
	}()

	processingError = ParseHTMLPlayerTable(file, &headersSnapshot, rowCellsChan, numWorkers, resultsChan, &wg)

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

	finalDatasetCurrencySymbol := "$"
	if len(playersList) > 0 {
		var foundSymbol bool
		for _, p := range playersList {
			_, _, tvSymbol := ParseMonetaryValueGo(p.TransferValue)
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
		CalculatePlayerPerformancePercentiles(playersList)
		log.Println("Finished calculating percentiles.")
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{Players: playersList, CurrencySymbol: finalDatasetCurrencySymbol}
	storeMutex.Unlock()

	log.Printf("Stored %d players with DatasetID: %s. Detected Currency: %s", len(playersList), datasetID, finalDatasetCurrencySymbol)
	if len(playersList) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", playersList[0].Name, playersList[0].Overall, playersList[0].ParsedPositions, playersList[0].ShortPositions, playersList[0].PositionGroups)
	} else {
		log.Println("No players were successfully parsed or list is empty after processing.")
	}

	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Parsed Players: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---",
		handler.Filename, fileSize/1024, totalDuration, parseDuration, len(playersList), rowsPerSecond, BToMb(memStats.Alloc), numWorkers, runtime.NumGoroutine())
}

// playerDataHandler handles GET requests to retrieve processed player data by dataset ID.
// It now accepts 'position' and 'role' query parameters for filtering and Overall adjustment.
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

	// Get filter parameters from query
	filterPosition := r.URL.Query().Get("position") // e.g., "DC"
	filterRole := r.URL.Query().Get("role")         // e.g., "DC - Central Defender - Defend"

	log.Printf("playerDataHandler: DatasetID=%s, PositionFilter=%s, RoleFilter=%s", datasetID, filterPosition, filterRole)

	storeMutex.RLock()
	data, found := playerDataStore[datasetID]
	storeMutex.RUnlock()

	if !found {
		log.Printf("Player data not found for DatasetID: %s", datasetID)
		http.Error(w, "Player data not found for the given ID. It might have expired, been cleared, or the ID is incorrect.", http.StatusNotFound)
		return
	}

	// Make a copy of players to modify.
	processedPlayers := make([]Player, 0, len(data.Players))

	for _, p := range data.Players {
		playerCopy := p // Create a copy to modify for this response

		// 1. Filter by Position
		if filterPosition != "" {
			canPlayPosition := false
			for _, shortPos := range playerCopy.ShortPositions {
				if shortPos == filterPosition {
					canPlayPosition = true
					break
				}
			}
			if !canPlayPosition {
				continue // Skip player if they can't play the selected position
			}
		}

		// 2. Modify Overall based on selected Role
		if filterRole != "" {
			roleMatched := false
			for _, roleOverall := range playerCopy.RoleSpecificOveralls {
				if roleOverall.RoleName == filterRole {
					playerCopy.Overall = roleOverall.Score // Set the main Overall to this role's score
					roleMatched = true
					break
				}
			}
			if !roleMatched {
				// If the player can play the position (checked above) but the specific role
				// is not found in their RoleSpecificOveralls, set their Overall for this request to 0.
				// This indicates they are not rated for this specific role.
				playerCopy.Overall = 0
			}
		}
		// If no role filter is applied, playerCopy.Overall remains their default calculated overall.
		// If a position filter is applied but no role filter, playerCopy.Overall is their default overall.

		processedPlayers = append(processedPlayers, playerCopy)
	}

	log.Printf("playerDataHandler: Returning %d players after processing for DatasetID=%s", len(processedPlayers), datasetID)

	response := PlayerDataWithCurrency{Players: processedPlayers, CurrencySymbol: data.CurrencySymbol}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

	muRoleSpecificOverallWeights.RLock() // Protects access to roleSpecificOverallWeights
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()

	sort.Strings(roleNames) // Optional: sort for consistent frontend display

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Consider configurability
	if err := json.NewEncoder(w).Encode(roleNames); err != nil {
		log.Printf("Error encoding JSON response for roles: %v", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

// Helper to convert bytes to MB for logging (already in fm_analysis_tool_updated_backend)
// func BToMb(b uint64) float64
// Helper to get first N cells for logging (already in fm_analysis_tool_updated_backend)
// func GetFirstNCells(slice []string, n int)
// Helper to get map keys for logging (already in fm_analysis_tool_updated_backend)
// func GetMapKeysStringFloat(m map[string]float64)

// Global variables (assuming they are defined in other files within the same package 'main')
// var defaultPlayerCapacity int (from config.go)
// var playerDataStore map[string]struct { Players []Player; CurrencySymbol string } (from store.go)
// var storeMutex sync.RWMutex (from store.go)
// var muRoleSpecificOverallWeights sync.RWMutex (from config.go)
// var roleSpecificOverallWeights map[string]map[string]int (from config.go)

// Functions from other files (assuming they are in the same package 'main')
// func ParseHTMLPlayerTable(...) (from parsing.go)
// func CalculatePlayerPerformancePercentiles(...) (from performance_stats.go)
// func ParseMonetaryValueGo(...) (from parsing.go)

// Structs from types.go (assuming they are in the same package 'main')
// type Player struct { ... }
// type PlayerParseResult struct { ... }
// type UploadResponse struct { ... }
// type PlayerDataWithCurrency struct { ... }
