package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"runtime" // Added for MemStats and NumCPU
	"strings"
	"time" // Added for timing

	"golang.org/x/net/html"
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

// getNodeText (remains the same)
func getNodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	var parts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text := getNodeText(c)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
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
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now() // Start timing the whole upload and parse process

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

	// --- Start of core parsing time measurement ---
	parseStartTime := time.Now()

	doc, err := html.Parse(file)
	if err != nil {
		http.Error(w, "Error parsing HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var players []Player
	var headers []string
	var tableNode *html.Node

	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTable(c)
			if tableNode != nil {
				return
			}
		}
	}
	findTable(doc)

	if tableNode == nil {
		http.Error(w, "No table found in the HTML", http.StatusInternalServerError)
		return
	}

	rowsProcessed := 0 // Counter for all rows attempted (including header)
	for tr := tableNode.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && tr.Data == "tbody" {
			for rowNode := tr.FirstChild; rowNode != nil; rowNode = rowNode.NextSibling {
				if rowNode.Type == html.ElementNode && rowNode.Data == "tr" {
					processRow(rowNode, &headers, &players)
					rowsProcessed++
				}
			}
			break
		} else if tr.Type == html.ElementNode && tr.Data == "tr" {
			processRow(tr, &headers, &players)
			rowsProcessed++
		}
	}

	parseDuration := time.Since(parseStartTime)
	// --- End of core parsing time measurement ---

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	log.Printf("Total Player Rows Successfully Parsed: %d (out of %d rows scanned in table body)", len(players), rowsProcessed)
	log.Printf("Player Rows Parsed Per Second (core parsing): %.2f", rowsPerSecond)
	log.Printf("Memory - Alloc: %v MiB", bToMb(memStats.Alloc))
	log.Printf("Memory - TotalAlloc (cumulative): %v MiB", bToMb(memStats.TotalAlloc))
	log.Printf("Memory - Sys (obtained from OS): %v MiB", bToMb(memStats.Sys))
	log.Printf("Memory - NumGC: %d", memStats.NumGC)
	log.Printf("System - NumCPU: %d", runtime.NumCPU())
	log.Printf("System - NumGoroutine: %d", runtime.NumGoroutine())
	log.Printf("----------------------------------------------")
}

// processRow (remains largely the same, ensure logging for skipped rows is clear)
func processRow(tr *html.Node, headers *[]string, players *[]Player) {
	var cells []string
	isHeaderRow := false

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode {
			cellText := getNodeText(td)
			if td.Data == "th" {
				isHeaderRow = true
				cells = append(cells, cellText)
			} else if td.Data == "td" {
				cells = append(cells, cellText)
			}
		}
	}

	if isHeaderRow {
		if len(*headers) == 0 { // Only log headers once
			*headers = cells
			log.Printf("Parsed Headers: %v", *headers)
		}
		return // Don't process header row as a player
	}

	if len(*headers) == 0 {
		// This can happen if the first row isn't a header row or table structure is unexpected
		log.Println("Skipping data row: Headers not yet parsed or not found.")
		return
	}

	nameIdx := getHeaderIndex(*headers, "Name")
	posIdx := getHeaderIndex(*headers, "Position")
	ageIdx := getHeaderIndex(*headers, "Age")
	clubIdx := getHeaderIndex(*headers, "Club")
	transferValueIdx := getHeaderIndex(*headers, "Transfer Value")
	wageIdx := getHeaderIndex(*headers, "Wage")

	if nameIdx == -1 || nameIdx >= len(cells) || strings.TrimSpace(cells[nameIdx]) == "" {
		// Reduce noise: only log if there's some content in the row, suggesting it's not just a formatting spacer
		isPotentiallyMeaningfulRow := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isPotentiallyMeaningfulRow = true
				break
			}
		}
		if isPotentiallyMeaningfulRow {
			log.Printf("Skipping row: 'Name' field is missing, empty, or index out of bounds. First few cells: %v", getFirstNCells(cells, 5))
		}
		return
	}

	playerNameForLog := safeGet(cells, nameIdx) // Get name early for logs

	requiredIndices := []int{posIdx, ageIdx, clubIdx, transferValueIdx, wageIdx}
	for _, idx := range requiredIndices {
		if idx != -1 && idx >= len(cells) {
			log.Printf("Skipping row for player '%s': Not enough cells for essential data. Cell count: %d, Required index: %d", playerNameForLog, len(cells), idx)
			return
		}
	}

	player := Player{
		Name:          playerNameForLog,
		Position:      safeGet(cells, posIdx),
		Age:           safeGet(cells, ageIdx),
		Club:          safeGet(cells, clubIdx),
		TransferValue: parseTransferValue(safeGet(cells, transferValueIdx)),
		Wage:          safeGet(cells, wageIdx),
		Attributes:    make(map[string]string),
	}

	attrStartIndex := getHeaderIndex(*headers, "Acc")
	attrEndIndex := getHeaderIndex(*headers, "Pen")

	if attrStartIndex != -1 {
		if attrEndIndex == -1 || attrEndIndex < attrStartIndex {
			fallbackEndIndex := getHeaderIndex(*headers, "Wor")
			if fallbackEndIndex != -1 && fallbackEndIndex >= attrStartIndex {
				attrEndIndex = fallbackEndIndex
			} else {
				attrEndIndex = -1
			}
		}
		if attrEndIndex != -1 {
			for i := attrStartIndex; i <= attrEndIndex && i < len(cells) && i < len(*headers); i++ {
				attrName := (*headers)[i]
				attrValue := cells[i]
				if attrName != "" && attrValue != "" && attrValue != "-" {
					player.Attributes[attrName] = attrValue
				}
			}
		} else {
			log.Printf("Could not determine a valid attribute end for player: %s (Acc idx: %d)", player.Name, attrStartIndex)
		}
	} else {
		log.Printf("'Acc' attribute header not found. Cannot parse attributes for player: %s", player.Name)
	}
	*players = append(*players, player)
}

// getFirstNCells returns the first N cells or fewer if the slice is smaller.
func getFirstNCells(slice []string, n int) []string {
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
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// main (remains the same)
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
