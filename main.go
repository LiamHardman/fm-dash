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
	Attributes    map[string]string `json:"attributes"` // This will store all other columns as attributes
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
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getNodeTextOptimized(c))
		if c.Type == html.ElementNode && c.NextSibling != nil {
			sb.WriteByte(' ')
		} else if c.Type == html.TextNode && c.NextSibling != nil && c.NextSibling.Type == html.ElementNode {
			sb.WriteByte(' ')
		}
	}
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
			if tableNode != nil {
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

	var headers []string
	var headerRowNode *html.Node

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
				return true
			}
		}
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table" || n.Data == "thead") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if findHeaderRow(c) {
					return true
				}
			}
		}
		return false
	}

	if !findHeaderRow(tableNode) {
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
							break
						}
					}
				}
				break
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
					break
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

	players := make([]Player, 0, defaultPlayerCapacity)
	rowNodesToProcess := make([]*html.Node, 0, defaultPlayerCapacity)

	var collectDataRows func(*html.Node)
	collectDataRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if n != headerRowNode {
				rowNodesToProcess = append(rowNodesToProcess, n)
			}
		}
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
	}

	numWorkers := runtime.NumCPU()
	if numRowsToProcess < numWorkers {
		numWorkers = numRowsToProcess
	}
	if numWorkers == 0 && numRowsToProcess > 0 {
		numWorkers = 1
	}

	rowNodeChan := make(chan *html.Node, numRowsToProcess)
	resultsChan := make(chan PlayerParseResult, numRowsToProcess)
	var wg sync.WaitGroup

	headersSnapshot := make([]string, len(headers))
	copy(headersSnapshot, headers)

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

	for _, rowNode := range rowNodesToProcess {
		rowNodeChan <- rowNode
	}
	close(rowNodeChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
		} else {
			log.Printf("Skipping row due to parsing error: %v", result.Err)
		}
	}

	parseDuration := time.Since(parseStartTime)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(players); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}

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
		if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
			cells = append(cells, getNodeTextOptimized(td))
		}
	}

	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty")
	}
	if len(cells) == 0 {
		return Player{}, errors.New("skipped row: no cells found in row")
	}

	// Initialize player with default attribute capacity
	player := Player{
		Attributes: make(map[string]string, defaultAttributeCapacity),
	}

	// Define known non-attribute column headers.
	// These should match exactly how they appear after getNodeTextOptimized.
	// IMPORTANT: Ensure these keys are exactly what you expect from your HTML headers.
	knownNonAttributeHeaders := map[string]bool{
		"Name":           true,
		"Position":       true,
		"Age":            true,
		"Club":           true,
		"Transfer Value": true,
		"Wage":           true,
		"Nat":            true, // If "Nat" is for Nationality column, not Natural Fitness attribute
		"Inf":            true, // Common for player status icons/info
		// Add any other specific non-attribute columns that might appear in your tables
		// e.g., "UID", "Personality", "Media Handling", etc.
	}

	// Iterate through all headers and cells to populate player fields and attributes
	foundName := false
	for i, headerName := range headers {
		if i < len(cells) { // Ensure we have a corresponding cell for the header
			cellValue := strings.TrimSpace(cells[i])

			switch headerName {
			case "Name":
				player.Name = cellValue
				if cellValue != "" {
					foundName = true
				}
			case "Position":
				player.Position = cellValue
			case "Age":
				player.Age = cellValue
			case "Club":
				player.Club = cellValue
			case "Transfer Value":
				player.TransferValue = parseTransferValue(cellValue)
			case "Wage":
				player.Wage = cellValue
			default:
				// If the header is not one of the known non-attribute fields, treat it as an attribute.
				if _, isNonAttribute := knownNonAttributeHeaders[headerName]; !isNonAttribute {
					// Only add if the attribute name and value are not empty and value is not "-"
					if headerName != "" && cellValue != "" && cellValue != "-" {
						player.Attributes[headerName] = cellValue
					}
				}
				// If it IS a known non-attribute but not one of the main fields (e.g. "Nat" for Nationality)
				// and you want to store it elsewhere on the Player struct, add that logic here.
				// For now, we only explicitly handle the main fields and dynamically add others to Attributes.
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
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty")
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

// getHeaderIndex finds the index of a header string in a slice of headers.
// Not strictly needed with the new attribute parsing logic but can be kept for other uses.
func getHeaderIndex(headers []string, headerName string) int {
	for i, h := range headers {
		if h == headerName {
			return i
		}
	}
	return -1
}

// safeGet safely retrieves an element from a slice by index, returning "" if out of bounds.
// Not strictly needed with the new attribute parsing logic.
func safeGet(slice []string, index int) string {
	if index >= 0 && index < len(slice) {
		return strings.TrimSpace(slice[index])
	}
	return ""
}

// bToMb converts bytes to megabytes.
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// main function to start the server.
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	http.HandleFunc("/upload", uploadHandler)

	port := "8091" // Changed from 8080 to 8091 as per your vite.config.js
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
