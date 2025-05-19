package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	defaultPlayerCapacity    = 1024 // Default capacity for the players slice
	defaultAttributeCapacity = 64  // Default capacity for the attributes map
	defaultCellCapacity      = 64  // Default capacity for cells slice if headers unknown
)

// Player struct (remains the same)
type Player struct {
	Name          string            `json:"name"`
	Position      string            `json:"position"`
	Age           string            `json:"age"`
	Club          string            `json:"club"`
	TransferValue string            `json:"transfer_value"`
	Wage          string            `json:"wage"`
	Attributes    map[string]string `json:"attributes"`
}

// PlayerParseResult is used to send results (or errors) from worker goroutines.
type PlayerParseResult struct {
	Player Player
	Err    error
}

// getNodeTextOptimized uses a strings.Builder and strings.Fields for efficient text extraction and normalization.
func getNodeTextOptimized(n *html.Node) string {
	if n == nil {
		return ""
	}
	// If it's a text node, return its data directly. Normalization will happen at a higher level.
	if n.Type == html.TextNode {
		return n.Data
	}

	// For element nodes, recursively gather text from children.
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getNodeTextOptimized(c))
		// Add a space after processing a child node if it's an element node
		// and it's not the last child. This helps separate words from different tags.
		// strings.Fields will handle multiple spaces or leading/trailing spaces later.
		if c.Type == html.ElementNode && c.NextSibling != nil {
			sb.WriteByte(' ')
		} else if c.Type == html.TextNode && c.NextSibling != nil && c.NextSibling.Type == html.ElementNode {
			// Add a space if text node is followed by an element node
			sb.WriteByte(' ')
		}
	}

	// strings.Fields splits the string by whitespace and removes empty strings.
	// strings.Join then joins them with a single space. This normalizes all whitespace.
	return strings.Join(strings.Fields(sb.String()), " ")
}

// parseTransferValue (remains the same)
func parseTransferValue(rawValue string) string {
	rawValue = strings.TrimSpace(rawValue)
	if strings.Contains(rawValue, " - ") {
		parts := strings.Split(rawValue, " - ")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	return rawValue
}

// uploadHandler handles file uploads, parses HTML, and returns player data as JSON.
// It now uses goroutines for concurrent row processing.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
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

	doc, err := html.Parse(file)
	if err != nil {
		http.Error(w, "Error parsing HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var tableNode *html.Node
	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if tableNode != nil { // Optimization: stop searching if already found
				return
			}
			findTable(c)
		}
	}
	findTable(doc)

	if tableNode == nil {
		http.Error(w, "No table found in the HTML", http.StatusInternalServerError)
		return
	}

	// --- Header Parsing (Sequential) ---
	var headers []string
	var headerRowNode *html.Node // Keep track of the header row to skip it later

	// Find the header row
	var findHeaderRow func(n *html.Node) bool
	findHeaderRow = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "tr" {
			isHeader := false
			tempHeaders := make([]string, 0, defaultCellCapacity)
			for cell := n.FirstChild; cell != nil; cell = cell.NextSibling {
				if cell.Type == html.ElementNode && cell.Data == "th" {
					isHeader = true
					tempHeaders = append(tempHeaders, getNodeTextOptimized(cell))
				} else if cell.Type == html.ElementNode && cell.Data == "td" && !isHeader {
					// If we see a td before any th, this row is unlikely a primary header.
					// However, some tables might mix. For simplicity, first row with <th> wins.
					// Or, if no <th> at all, the first row's <td>s might be considered headers.
					// This part might need adjustment based on HTML variations.
					// For now, we require at least one <th> for it to be a header row.
				}
			}
			if isHeader && len(tempHeaders) > 0 {
				headers = tempHeaders
				headerRowNode = n
				log.Printf("Parsed Headers: %v", headers)
				return true // Header found
			}
		}
		// Check within tbody or directly under table
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table" || n.Data == "thead") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if findHeaderRow(c) {
					return true
				}
			}
		}
		return false
	}

	// Search for headers starting from the table node
	if !findHeaderRow(tableNode) {
		// Fallback: if no <th> based header found, try to use the very first row.
		// This is a heuristic and might not always be correct.
		firstRow := true
		for tr := tableNode.FirstChild; tr != nil; tr = tr.NextSibling {
			if tr.Type == html.ElementNode && tr.Data == "tbody" {
				for rowNode := tr.FirstChild; rowNode != nil; rowNode = rowNode.NextSibling {
					if rowNode.Type == html.ElementNode && rowNode.Data == "tr" {
						if firstRow {
							headerRowNode = rowNode
							for td := rowNode.FirstChild; td != nil; td = td.NextSibling {
								if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
									headers = append(headers, getNodeTextOptimized(td))
								}
							}
							log.Printf("Warning: No <th> header row found. Using first row as header: %v", headers)
							firstRow = false
							break // Found first row
						}
					}
				}
				break // Processed tbody
			} else if tr.Type == html.ElementNode && tr.Data == "tr" {
				if firstRow {
					headerRowNode = tr
					for td := tr.FirstChild; td != nil; td = td.NextSibling {
						if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
							headers = append(headers, getNodeTextOptimized(td))
						}
					}
					log.Printf("Warning: No <th> header row found. Using first row as header (direct tr): %v", headers)
					firstRow = false
					break // Found first row
				}
			}
			if !firstRow {
				break
			}
		}
	}

	if len(headers) == 0 {
		log.Println("Critical: Headers could not be parsed. Aborting player data processing.")
		http.Error(w, "Could not parse table headers", http.StatusInternalServerError)
		return
	}

	// --- Concurrent Row Processing ---
	players := make([]Player, 0, defaultPlayerCapacity)
	rowNodesToProcess := make([]*html.Node, 0, defaultPlayerCapacity) // Collect data rows first

	// Collect all data row nodes (excluding the identified header row)
	var collectDataRows func(*html.Node)
	collectDataRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if n != headerRowNode { // Skip the already processed header row
				rowNodesToProcess = append(rowNodesToProcess, n)
			}
		}
		// Traverse into tbody or directly look for trs
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				collectDataRows(c)
			}
		}
	}
	collectDataRows(tableNode)

	numRowsToProcess := len(rowNodesToProcess)
	if numRowsToProcess == 0 {
		log.Println("No data rows found to process after header parsing.")
		// Proceed to send empty player list if no data rows
	}

	numWorkers := runtime.NumCPU()
	if numRowsToProcess < numWorkers { // Don't spin up more workers than rows
		numWorkers = numRowsToProcess
	}
	if numWorkers == 0 && numRowsToProcess > 0 { // Ensure at least one worker if there are rows
		numWorkers = 1
	}

	rowNodeChan := make(chan *html.Node, numRowsToProcess)        // Buffered channel
	resultsChan := make(chan PlayerParseResult, numRowsToProcess) // Buffered channel
	var wg sync.WaitGroup

	// Create a snapshot of headers for workers to prevent race conditions if headers were mutable
	// (though in this flow, they are set once before workers start).
	headersSnapshot := make([]string, len(headers))
	copy(headersSnapshot, headers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rowNode := range rowNodeChan {
				player, err := parseRowToPlayer(rowNode, headersSnapshot)
				resultsChan <- PlayerParseResult{Player: player, Err: err}
			}
		}()
	}

	// Distribute row nodes to workers
	for _, rowNode := range rowNodesToProcess {
		rowNodeChan <- rowNode
	}
	close(rowNodeChan) // Signal workers that no more rows will be sent

	// Closer goroutine for resultsChan
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
		} else {
			// Log errors from parsing individual rows, but don't stop the whole process
			log.Printf("Skipping row due to parsing error: %v", result.Err)
		}
	}

	parseDuration := time.Since(parseStartTime)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For local testing
	if err := json.NewEncoder(w).Encode(players); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}

	// Performance Logging
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(players)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)

	log.Printf("--- Parsing & Processing Performance Metrics ---")
	log.Printf("File: %s", handler.Filename)
	log.Printf("File Size Processed: %d bytes (%.2f KB)", fileSize, float64(fileSize)/1024.0)
	log.Printf("Total Request Time (Upload + Parse + Response): %v", totalDuration)
	log.Printf("Core Parsing Time (html.Parse + row processing): %v", parseDuration)
	log.Printf("Total Data Rows Sent to Workers: %d", numRowsToProcess)
	log.Printf("Total Player Rows Successfully Parsed: %d", len(players))
	log.Printf("Player Rows Parsed Per Second (core parsing): %.2f", rowsPerSecond)
	log.Printf("Memory - Alloc: %.2f MiB", bToMb(memStats.Alloc))
	log.Printf("Memory - TotalAlloc (cumulative): %.2f MiB", bToMb(memStats.TotalAlloc))
	log.Printf("Memory - Sys (obtained from OS): %.2f MiB", bToMb(memStats.Sys))
	log.Printf("Memory - NumGC: %d", memStats.NumGC)
	log.Printf("System - NumCPU for Workers: %d (Max Available: %d)", numWorkers, runtime.NumCPU())
	log.Printf("System - NumGoroutine at end of request: %d", runtime.NumGoroutine())
	log.Printf("----------------------------------------------")
}

// parseRowToPlayer processes a single <tr> node into a Player object.
// It's designed to be called by worker goroutines.
func parseRowToPlayer(tr *html.Node, headers []string) (Player, error) {
	var cells []string
	// Estimate cell capacity based on headers length, or use a default.
	cellCap := defaultCellCapacity
	if len(headers) > 0 {
		cellCap = len(headers)
	}
	cells = make([]string, 0, cellCap)

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") { // Also consider th if data rows might have them
			cells = append(cells, getNodeTextOptimized(td))
		}
	}

	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty (should have been caught earlier)")
	}

	nameIdx := getHeaderIndex(headers, "Name")
	posIdx := getHeaderIndex(headers, "Position")
	ageIdx := getHeaderIndex(headers, "Age")
	clubIdx := getHeaderIndex(headers, "Club")
	transferValueIdx := getHeaderIndex(headers, "Transfer Value")
	wageIdx := getHeaderIndex(headers, "Wage")

	playerName := safeGet(cells, nameIdx)
	if nameIdx == -1 || playerName == "" {
		// Check if the row is potentially meaningful before logging it as a skip
		isPotentiallyMeaningfulRow := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isPotentiallyMeaningfulRow = true
				break
			}
		}
		if isPotentiallyMeaningfulRow {
			return Player{}, errors.New("skipped row: 'Name' field is missing, empty, or index out of bounds. First few cells: " + strings.Join(getFirstNCells(cells, 5), ", "))
		}
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty") // Less verbose for truly empty rows
	}

	// Validate that essential indices are within bounds of actual cells for this row
	requiredIndices := map[string]int{
		"Position": posIdx, "Age": ageIdx, "Club": clubIdx,
		"Transfer Value": transferValueIdx, "Wage": wageIdx,
	}
	for headerName, idx := range requiredIndices {
		if idx != -1 && idx >= len(cells) {
			return Player{}, errors.New("skipped row for player '" + playerName + "': Not enough cells for essential data '" + headerName + "'. Cell count: " + string(len(cells)) + ", Required index: " + string(idx))
		}
	}

	player := Player{
		Name:          playerName,
		Position:      safeGet(cells, posIdx),
		Age:           safeGet(cells, ageIdx),
		Club:          safeGet(cells, clubIdx),
		TransferValue: parseTransferValue(safeGet(cells, transferValueIdx)),
		Wage:          safeGet(cells, wageIdx),
		Attributes:    make(map[string]string, defaultAttributeCapacity),
	}

	attrStartIndex := getHeaderIndex(headers, "Acc") // Example start attribute
	attrEndIndex := getHeaderIndex(headers, "Pen")   // Example end attribute, adjust as needed

	if attrStartIndex != -1 {
		// Fallback if "Pen" is not found, try "Wor" or a sensible last known attribute
		if attrEndIndex == -1 || attrEndIndex < attrStartIndex {
			fallbackEndIndex := getHeaderIndex(headers, "Wor")
			if fallbackEndIndex != -1 && fallbackEndIndex >= attrStartIndex {
				attrEndIndex = fallbackEndIndex
			} else {
				// If no clear end, go up to the end of known headers or available cells
				attrEndIndex = len(headers) - 1
			}
		}

		for i := attrStartIndex; i <= attrEndIndex && i < len(cells) && i < len(headers); i++ {
			attrName := headers[i]
			attrValue := strings.TrimSpace(cells[i])
			if attrName != "" && attrValue != "" && attrValue != "-" {
				player.Attributes[attrName] = attrValue
			}
		}
	} else {
		// Log this as a warning, but don't fail the player if attributes are missing
		// log.Printf("Warning: 'Acc' attribute header not found. Cannot parse attributes for player: %s", player.Name)
	}
	return player, nil
}

// getFirstNCells returns the first N cells or fewer if the slice is smaller.
// (Unchanged from original, useful for logging)
func getFirstNCells(slice []string, n int) []string {
	if n < 0 {
		n = 0
	}
	if n > len(slice) {
		n = len(slice)
	}
	return slice[:n]
}

// getHeaderIndex (remains the same)
func getHeaderIndex(headers []string, headerName string) int {
	for i, h := range headers {
		if h == headerName {
			return i
		}
	}
	return -1
}

// safeGet (remains the same)
func safeGet(slice []string, index int) string {
	if index >= 0 && index < len(slice) {
		return strings.TrimSpace(slice[index])
	}
	return ""
}

// bToMb converts bytes to megabytes for logging.
// (Unchanged from original)
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// main (remains the same)
func main() {
	// Serve index.html at the root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Determine the path to index.html relative to the executable
		// This assumes index.html is in the same directory as the compiled binary.
		// For development, running `go run .` from the directory containing main.go and index.html works.
		// For deployment, ensure index.html is placed alongside the executable.
		// _, currentFilePath, _, ok := runtime.Caller(0)
		// if !ok {
		// 	http.Error(w, "Could not determine server file path", http.StatusInternalServerError)
		// 	return
		// }
		// dir := filepath.Dir(currentFilePath)
		// http.ServeFile(w, r, filepath.Join(dir, "index.html"))

		// Simpler: serve from current working directory.
		// Ensure your `index.html` is in the directory where you run the server.
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	http.HandleFunc("/upload", uploadHandler)

	port := "8091"
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
