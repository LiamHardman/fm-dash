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
	defaultPlayerCapacity    = 1024
	defaultAttributeCapacity = 64
	defaultCellCapacity      = 64
)

// Player struct
type Player struct {
	Name          string            `json:"name"`
	Position      string            `json:"position"`
	Age           string            `json:"age"`
	Club          string            `json:"club"`
	TransferValue string            `json:"transfer_value"`
	Wage          string            `json:"wage"`
	Nationality   string            `json:"nationality"` // Will store the full country name
	Attributes    map[string]string `json:"attributes"`
}

// PlayerParseResult is used for concurrent processing.
type PlayerParseResult struct {
	Player Player
	Err    error
}

// START: FIFA Country Code Map
var fifaCountryCodes = map[string]string{
	"AFG": "Afghanistan", "ALB": "Albania", "ALG": "Algeria", "ASA": "American Samoa", "AND": "Andorra",
	"ANG": "Angola", "AIA": "Anguilla", "ATG": "Antigua and Barbuda", "ARG": "Argentina", "ARM": "Armenia",
	"ARU": "Aruba", "AUS": "Australia", "AUT": "Austria", "AZE": "Azerbaijan", "BAH": "Bahamas",
	"BHR": "Bahrain", "BAN": "Bangladesh", "BRB": "Barbados", "BLR": "Belarus", "BEL": "Belgium",
	"BLZ": "Belize", "BEN": "Benin", "BER": "Bermuda", "BHU": "Bhutan", "BOL": "Bolivia",
	"BIH": "Bosnia and Herzegovina", "BOT": "Botswana", "BRA": "Brazil", "VGB": "British Virgin Islands",
	"BRU": "Brunei Darussalam", "BUL": "Bulgaria", "BFA": "Burkina Faso", "BDI": "Burundi", "CAM": "Cambodia",
	"CMR": "Cameroon", "CAN": "Canada", "CPV": "Cape Verde", "CAY": "Cayman Islands", "CTA": "Central African Republic",
	"CHA": "Chad", "CHI": "Chile", "CHN": "China PR", "TPE": "Chinese Taipei", "COL": "Colombia",
	"COM": "Comoros", "CGO": "Congo", "COD": "DR Congo", "COK": "Cook Islands", "CRC": "Costa Rica",
	"CIV": "Ivory Coast", "CRO": "Croatia", "CUB": "Cuba", "CUW": "Curaçao", "CYP": "Cyprus",
	"CZE": "Czech Republic", "DEN": "Denmark", "DJI": "Djibouti", "DMA": "Dominica", "DOM": "Dominican Republic",
	"ECU": "Ecuador", "EGY": "Egypt", "SLV": "El Salvador", "ENG": "England", "EQG": "Equatorial Guinea",
	"ERI": "Eritrea", "EST": "Estonia", "SWZ": "Eswatini", "ETH": "Ethiopia", "FRO": "Faroe Islands",
	"FIJ": "Fiji", "FIN": "Finland", "FRA": "France", "GAB": "Gabon", "GAM": "Gambia",
	"GEO": "Georgia", "GER": "Germany", "GHA": "Ghana", "GIB": "Gibraltar", "GRE": "Greece",
	"GRN": "Grenada", "GUM": "Guam", "GUA": "Guatemala", "GUI": "Guinea", "GNB": "Guinea-Bissau",
	"GUY": "Guyana", "HAI": "Haiti", "HON": "Honduras", "HKG": "Hong Kong", "HUN": "Hungary",
	"ISL": "Iceland", "IND": "India", "IDN": "Indonesia", "IRN": "Iran", "IRQ": "Iraq",
	"IRL": "Republic of Ireland", "ISR": "Israel", "ITA": "Italy", "JAM": "Jamaica", "JPN": "Japan",
	"JOR": "Jordan", "KAZ": "Kazakhstan", "KEN": "Kenya", "PRK": "Korea DPR", "KOR": "Korea Republic",
	"KVX": "Kosovo", "KUW": "Kuwait", "KGZ": "Kyrgyzstan", "LAO": "Laos", "LVA": "Latvia",
	"LBN": "Lebanon", "LES": "Lesotho", "LBR": "Liberia", "LBY": "Libya", "LIE": "Liechtenstein",
	"LTU": "Lithuania", "LUX": "Luxembourg", "MAC": "Macau", "MAD": "Madagascar", "MWI": "Malawi",
	"MAS": "Malaysia", "MDV": "Maldives", "MLI": "Mali", "MLT": "Malta", "MTN": "Mauritania",
	"MRI": "Mauritius", "MEX": "Mexico", "MDA": "Moldova", "MNG": "Mongolia", "MNE": "Montenegro",
	"MSR": "Montserrat", "MAR": "Morocco", "MOZ": "Mozambique", "MYA": "Myanmar", "NAM": "Namibia",
	"NEP": "Nepal", "NED": "Netherlands", "NCL": "New Caledonia", "NZL": "New Zealand", "NCA": "Nicaragua",
	"NIG": "Niger", "NGA": "Nigeria", "MKD": "North Macedonia", "NIR": "Northern Ireland", "NOR": "Norway",
	"OMA": "Oman", "PAK": "Pakistan", "PLE": "Palestine", "PAN": "Panama", "PNG": "Papua New Guinea",
	"PAR": "Paraguay", "PER": "Peru", "PHI": "Philippines", "POL": "Poland", "POR": "Portugal",
	"PUR": "Puerto Rico", "QAT": "Qatar", "ROU": "Romania", "RUS": "Russia", "RWA": "Rwanda",
	"SKN": "St. Kitts and Nevis", "LCA": "St. Lucia", "VIN": "St. Vincent & Grenadines", "SAM": "Samoa",
	"SMR": "San Marino", "STP": "São Tomé e Príncipe", "KSA": "Saudi Arabia", "SCO": "Scotland", "SEN": "Senegal",
	"SRB": "Serbia", "SEY": "Seychelles", "SLE": "Sierra Leone", "SIN": "Singapore", "SVK": "Slovakia",
	"SVN": "Slovenia", "SOL": "Solomon Islands", "SOM": "Somalia", "RSA": "South Africa", "SSD": "South Sudan",
	"ESP": "Spain", "SRI": "Sri Lanka", "SDN": "Sudan", "SUR": "Suriname", "SWE": "Sweden",
	"SUI": "Switzerland", "SYR": "Syria", "TAH": "Tahiti", "TJK": "Tajikistan", "TAN": "Tanzania",
	"THA": "Thailand", "TLS": "Timor-Leste", "TOG": "Togo", "TGA": "Tonga", "TRI": "Trinidad and Tobago",
	"TUN": "Tunisia", "TUR": "Turkey", "TKM": "Turkmenistan", "TCA": "Turks and Caicos Islands",
	"UGA": "Uganda", "UKR": "Ukraine", "UAE": "United Arab Emirates", "USA": "USA", "URU": "Uruguay",
	"VIR": "US Virgin Islands", "UZB": "Uzbekistan", "VAN": "Vanuatu", "VEN": "Venezuela", "VIE": "Vietnam",
	"WAL": "Wales", "YEM": "Yemen", "ZAM": "Zambia", "ZIM": "Zimbabwe",
	// Add more codes if needed
}

// END: FIFA Country Code Map

// getNodeTextOptimized remains the same
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

// parseTransferValue remains the same
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

// uploadHandler remains largely the same, details omitted for brevity, ensure it's as per your latest version
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	startTime := time.Now()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
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
		log.Println("Critical: Headers could not be parsed.")
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
		log.Println("No data rows found.")
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
				player, err := parseRowToPlayer(rowNode, headersSnapshot)
				resultsChan <- PlayerParseResult{Player: player, Err: err}
			}
		}()
	}
	for _, rowNode := range rowNodesToProcess {
		rowNodeChan <- rowNode
	}
	close(rowNodeChan)
	go func() { wg.Wait(); close(resultsChan) }()
	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
		} else {
			log.Printf("Skipping row: %v", result.Err)
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
	log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Rows: %d, Parsed: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---", handler.Filename, fileSize/1024, totalDuration, parseDuration, numRowsToProcess, len(players), rowsPerSecond, bToMb(memStats.Alloc), numWorkers, runtime.NumGoroutine())
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

	player := Player{
		Attributes: make(map[string]string, defaultAttributeCapacity),
	}

	knownNonAttributeHeaders := map[string]bool{
		"Inf": true, // Common for player status icons/info
		// "Nat" is handled specially, so it's not listed here.
	}

	foundName := false
	for i, headerName := range headers {
		if i < len(cells) {
			cellValue := strings.TrimSpace(cells[i])
			isAnAttributeField := true

			switch headerName {
			case "Name":
				player.Name = cellValue
				if cellValue != "" {
					foundName = true
				}
				isAnAttributeField = false
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
			case "Nat":
				if player.Nationality == "" { // First "Nat" column is Nationality
					// Convert 3-letter code to full name
					if fullName, ok := fifaCountryCodes[strings.ToUpper(cellValue)]; ok {
						player.Nationality = fullName
					} else {
						player.Nationality = cellValue // Default to the code if not found in map
						log.Printf("Warning: FIFA country code '%s' not found in map. Using original value.", cellValue)
					}
					isAnAttributeField = false
				} else {
					// Subsequent "Nat" column is treated as "Natural Fitness" attribute
					// isAnAttributeField remains true
				}
			}

			if isAnAttributeField {
				if _, isKnownNonAttr := knownNonAttributeHeaders[headerName]; !isKnownNonAttr {
					if headerName != "" && cellValue != "" && cellValue != "-" {
						player.Attributes[headerName] = cellValue
					}
				}
			}
		}
	}

	if !foundName {
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

// bToMb converts bytes to megabytes for logging.
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
	port := "8091"
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
