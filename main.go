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
	defaultAttributeCapacity = 64   // Default capacity for the attributes map
	defaultCellCapacity      = 64   // Default capacity for cells slice if headers unknown
)

// Player struct
type Player struct {
	Name          string            `json:"name"`
	Position      string            `json:"position"`
	Age           string            `json:"age"`
	Club          string            `json:"club"`
	TransferValue string            `json:"transfer_value"`
	Wage          string            `json:"wage"`
	Nationality   string            `json:"nationality"` // Added Nationality field
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

// parseTransferValue extracts the higher-end or single value from a transfer value string.
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
				}
			}
			if isHeader && len(tempHeaders) > 0 {
				headers = tempHeaders
				headerRowNode = n
				log.Printf("Parsed Headers: %v", headers)
				return true // Header found
			}
		}
		// Check within tbody or directly under table or thead
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
	if numRowsToProcess < numWorkers {
		numWorkers = numRowsToProcess
	}
	if numWorkers == 0 && numRowsToProcess > 0 {
		numWorkers = 1
	}

	rowNodeChan := make(chan *html.Node, numRowsToProcess)        // Buffered channel
	resultsChan := make(chan PlayerParseResult, numRowsToProcess) // Buffered channel
	var wg sync.WaitGroup

	headersSnapshot := make([]string, len(headers))
	copy(headersSnapshot, headers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rowNode := range rowNodeChan {
				player, err := parseRowToPlayer(rowNode, headersSnapshot) // Pass headersSnapshot
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
	log.Printf("File: %s, Size: %d bytes (%.2f KB)", handler.Filename, fileSize, float64(fileSize)/1024.0)
	log.Printf("Total Request Time: %v, Core Parsing Time: %v", totalDuration, parseDuration)
	log.Printf("Data Rows to Workers: %d, Parsed Players: %d, Rows/Sec: %.2f", numRowsToProcess, len(players), rowsPerSecond)
	log.Printf("Memory - Alloc: %.2f MiB, TotalAlloc: %.2f MiB, Sys: %.2f MiB, NumGC: %d", bToMb(memStats.Alloc), bToMb(memStats.TotalAlloc), bToMb(memStats.Sys), memStats.NumGC)
	log.Printf("System - Workers: %d, Goroutines: %d", numWorkers, runtime.NumGoroutine())
	log.Printf("----------------------------------------------")
}

// parseRowToPlayer processes a single <tr> node into a Player object.
func parseRowToPlayer(tr *html.Node, headers []string) (Player, error) {
	var cells []string
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
		return Player{}, errors.New("cannot process row: headers are empty")
	}
	if len(cells) == 0 {
		// This case might occur if a row is completely empty or malformed.
		return Player{}, errors.New("skipped row: no cells found in row")
	}

	// Initialize player with default attribute capacity
	player := Player{
		Attributes: make(map[string]string, defaultAttributeCapacity),
		// Nationality will be empty by default, to be filled if a "Nat" column is found
	}

	// Define known non-attribute column headers that are NOT the primary player fields.
	// "Nat" is intentionally excluded here because we handle it specially below.
	knownNonAttributeHeaders := map[string]bool{
		"Inf": true, // Common for player status icons/info
		// Add any other specific non-attribute columns that might appear in your tables
		// e.g., "UID", "Personality", "Media Handling", etc.
		// DO NOT add "Nat" here.
	}

	// Iterate through all headers and cells to populate player fields and attributes
	foundName := false
	for i, headerName := range headers {
		if i < len(cells) { // Ensure we have a corresponding cell for the header
			cellValue := strings.TrimSpace(cells[i])
			isAnAttributeField := true // Assume it's an attribute unless explicitly handled by a case

			switch headerName {
			case "Name":
				player.Name = cellValue
				if cellValue != "" {
					foundName = true
				}
				isAnAttributeField = false // This is a main field, not an attribute
			case "Position":
				player.Position = cellValue
				isAnAttributeField = false
			case "Age":
				player.Age = cellValue
				isAnAttributeField = false
			case "Club":
				player.Club = cellValue
				isAnAttributeField = false
			case "Transfer Value":
				player.TransferValue = parseTransferValue(cellValue)
				isAnAttributeField = false
			case "Wage":
				player.Wage = cellValue
				isAnAttributeField = false
			case "Nat": // Special handling for "Nat"
				// The first "Nat" column encountered is assumed to be Nationality.
				// Any subsequent "Nat" column will be treated as the "Natural Fitness" attribute.
				if player.Nationality == "" { // If Nationality field is not yet set, this is it.
					player.Nationality = cellValue
					isAnAttributeField = false // This "Nat" is handled as the main Nationality.
				} else {
					// If player.Nationality is already set, this "Nat" must be the "Natural Fitness" attribute.
					// isAnAttributeField remains true, so it will be added to Attributes map below.
					// No action needed here, the logic below will handle it.
				}
				// No default case needed here, logic below handles attributes
			}

			// If isAnAttributeField is true, it means the header was not one of the main player fields
			// (Name, Position, Age, Club, Transfer Value, Wage) OR the first "Nat" (Nationality).
			// It could be an actual attribute (like "Acc", "Fin", or the second "Nat" for Natural Fitness)
			// or something we explicitly want to ignore (like "Inf").
			if isAnAttributeField {
				// Check if it's a known non-attribute header we want to ignore (e.g., "Inf")
				if _, isKnownNonAttr := knownNonAttributeHeaders[headerName]; !isKnownNonAttr {
					// It's not a main field, not the first "Nat", and not in the explicit ignore list.
					// So, treat it as a player attribute.
					// Only add if the attribute name and value are not empty and value is not "-"
					if headerName != "" && cellValue != "" && cellValue != "-" {
						player.Attributes[headerName] = cellValue
					}
				}
			}
		}
	}

	// If no name was found, this row might be invalid or empty.
	if !foundName {
		// Check if the row is potentially meaningful before logging it as a skip
		isPotentiallyMeaningfulRow := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isPotentiallyMeaningfulRow = true
				break
			}
		}
		if isPotentiallyMeaningfulRow {
			return Player{}, errors.New("skipped row: 'Name' field is missing or empty. First few cells: " + strings.Join(getFirstNCells(cells, 5), ", "))
		}
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty") // Less verbose for truly empty rows
	}

	return player, nil
}

// getFirstNCells returns the first N cells or fewer if the slice is smaller.
func getFirstNCells(slice []string, n int) []string {
	if n < 0 {
		n = 0
	}
	if n > len(slice) {
		n = len(slice)
	}
	return slice[:n]
}

// bToMb converts bytes to megabytes for logging.
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// main function to start the server.
func main() {
	// Serve index.html at the root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	http.HandleFunc("/upload", uploadHandler)

	port := "8091" // Ensure this matches your frontend proxy and desired port
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
