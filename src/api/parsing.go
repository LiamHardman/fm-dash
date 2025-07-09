package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// sendRowWithBackpressure attempts to send row data to channel with optimized timeout handling
func sendRowWithBackpressure(rowCellsChan chan []string, cells []string, timeout time.Duration) {
	// Pre-allocate with exact capacity to avoid slice growth
	cellsCopy := make([]string, len(cells))
	copy(cellsCopy, cells)

	// Try immediate send first (fast path)
	select {
	case rowCellsChan <- cellsCopy:
		// Successfully sent immediately
		return
	default:
		// Channel is full, record backpressure
		RecordBackpressure()
	}

	// Slower path with timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case rowCellsChan <- cellsCopy:
		// Successfully sent with timeout
	case <-timer.C:
		RecordChannelTimeout() // Record timeout event
		log.Printf("Warning: Channel send timeout after %v, dropping row with %d cells", timeout, len(cells))
	}
}

// Optimized regex compilation for currency parsing - compiled once at package level
var currencySymbolRegex = regexp.MustCompile(
	`([€£$¥₹₽₺₩])|` + // Single character symbols (Euro, Pound, Dollar, Yen, Rupee, Ruble, Lira, Won)
		`(R\$)|` + // Brazilian Real
		`(CHF)|` + // Swiss Franc
		`(A\$)|` + // Australian Dollar
		`(CA\$)|` + // Canadian Dollar (use CA$ to distinguish from A$)
		`(Mex\$)|` + // Mexican Peso
		`(kr)|` + // Krona (Swedish, Norwegian, Danish) - might be too generic, consider specific like SEK, NOK, DKK if available in data
		`(zł)|` + // Polish Zloty
		`(R)`, // South African Rand (must be last if 'R$' is also used to avoid partial match)
)

// Optimized buffer pool for string operations
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func getStringBuilder() *strings.Builder {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

func putStringBuilder(sb *strings.Builder) {
	if sb.Cap() <= 1024 { // Don't pool extremely large builders
		stringBuilderPool.Put(sb)
	}
}

// Enhanced ParseMonetaryValueGo with better string handling
func ParseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	if rawValue == "" {
		return "", 0, ""
	}

	cleanedValue := strings.TrimSpace(rawValue)
	originalDisplay = cleanedValue

	// Use pre-compiled regex for better performance
	matches := currencySymbolRegex.FindStringSubmatch(cleanedValue)
	if len(matches) > 1 {
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				detectedSymbol = matches[i]
				break
			}
		}
	}

	if detectedSymbol == "" {
		// Fast fallback checks using Contains (faster than regex for simple cases)
		switch {
		case strings.Contains(cleanedValue, "€"):
			detectedSymbol = "€"
		case strings.Contains(cleanedValue, "£"):
			detectedSymbol = "£"
		case strings.Contains(cleanedValue, "$"):
			detectedSymbol = "$"
		}
	}

	// Handle ranges like "£10M - £15M" more efficiently
	if idx := strings.Index(cleanedValue, " - "); idx != -1 {
		cleanedValue = strings.TrimSpace(cleanedValue[idx+3:])
		// Re-detect symbol efficiently
		if detectedSymbol == "" {
			rangeSymbolMatches := currencySymbolRegex.FindStringSubmatch(cleanedValue)
			if len(rangeSymbolMatches) > 1 {
				for i := 1; i < len(rangeSymbolMatches); i++ {
					if rangeSymbolMatches[i] != "" {
						detectedSymbol = rangeSymbolMatches[i]
						break
					}
				}
			}
		}
	}

	// More efficient symbol removal using strings.Replacer
	var replacer *strings.Replacer
	if detectedSymbol != "" {
		// Create replacer for detected symbol plus common cleanup
		replacer = strings.NewReplacer(
			detectedSymbol, "",
			"p/w", "",
			"/w", "",
			",", "",
		)
	} else {
		// Fallback replacer for all symbols
		replacer = strings.NewReplacer(
			"€", "", "£", "", "$", "", "¥", "", "₹", "", "₽", "", "₺", "", "₩", "",
			"R", "", "CHF", "", "kr", "", "zł", "",
			"p/w", "", "/w", "", ",", "",
		)
	}

	valueToParse := strings.TrimSpace(replacer.Replace(cleanedValue))

	// Optimized multiplier detection
	multiplier := int64(1)
	valueLen := len(valueToParse)
	if valueLen > 0 {
		lastChar := valueToParse[valueLen-1]
		switch lastChar {
		case 'M', 'm':
			multiplier = 1000000
			valueToParse = valueToParse[:valueLen-1]
		case 'K', 'k':
			multiplier = 1000
			valueToParse = valueToParse[:valueLen-1]
		}
	}

	// Fast path for simple integer parsing
	if val, err := strconv.ParseInt(strings.TrimSpace(valueToParse), 10, 64); err == nil {
		numericValue = val * multiplier
	} else if valFloat, err := strconv.ParseFloat(strings.TrimSpace(valueToParse), 64); err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	} else {
		numericValue = 0
	}

	// Set default symbol if needed
	if detectedSymbol == "" && numericValue != 0 {
		detectedSymbol = "$"
	}

	return originalDisplay, numericValue, detectedSymbol
}

// ParseHTMLPlayerTable tokenizes an HTML file stream (typically a player squad view)
// and sends extracted table headers and row data (as slices of strings) to respective channels.
// It manages the HTML parsing state machine.
func ParseHTMLPlayerTable(file io.Reader, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) (processingError error) {
	bufferedReader := bufio.NewReaderSize(file, maxTokenBufferSize) // maxTokenBufferSize from config.go
	tokenizer := html.NewTokenizer(bufferedReader)

	var currentHeaders []string // Temporary headers collected
	var currentCells []string   // Cells for the current <tr> being processed
	inHeaderRow := false        // True if currently inside a <tr> identified as a header row
	inDataRow := false          // True if currently inside a <tr> identified as a data row
	inTable := false            // True if currently inside a <table> element
	inTHead := false            // True if currently inside a <thead> element
	inTBody := false            // True if currently inside a <tbody> element

	// Use pooled string builder for better performance
	cellBuilder := getStringBuilder()
	defer putStringBuilder(cellBuilder)

	workersStarted := false
	channelClosed := false             // Track channel state to prevent double-close
	var localHeadersForWorker []string // To pass to workers once finalized

	// Helper function to safely close the channel
	closeChannelOnce := func() {
		if !channelClosed {
			channelClosed = true
			close(rowCellsChan)
		}
	}

tokenLoop:
	for {
		tt := tokenizer.Next()
		token := tokenizer.Token()

		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				if inDataRow && len(currentCells) > 0 && workersStarted && !channelClosed {
					sendRowWithBackpressure(rowCellsChan, currentCells, 5*time.Second)
				}
				break tokenLoop
			}
			log.Printf("HTML tokenization error occurred during parsing")
			processingError = errors.New("Error tokenizing HTML: " + err.Error())
			break tokenLoop
		case html.StartTagToken:
			switch token.Data {
			case "table":
				inTable = true
			case "thead":
				if inTable {
					inTHead = true
				}
			case "tbody":
				if inTable {
					inTBody = true
					if !workersStarted && len(currentHeaders) > 0 {
						localHeadersForWorker = make([]string, len(currentHeaders))
						copy(localHeadersForWorker, currentHeaders)
						*headersSnapshot = localHeadersForWorker
						LogDebug("Headers found (tbody start), launching %d workers with %d headers", numWorkers, len(localHeadersForWorker))
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
					}
				}
			case "tr":
				currentCells = make([]string, 0, defaultCellCapacity) // defaultCellCapacity from config.go
				// Determine if this TR is a header row or data row
				switch {
				case inTHead:
					inHeaderRow = true
					inDataRow = false
				case inTable && !workersStarted && len(currentHeaders) == 0 && !inTBody:
					// If in a table, workers haven't started, no headers collected yet, AND not in <tbody>:
					// Treat this as a potential header row.
					inHeaderRow = true
					inDataRow = false
				default:
					// Otherwise (e.g., in tbody, or workers started, or headers already found), it's a data row.
					inHeaderRow = false
					inDataRow = true
				}
			case "th":
				if inHeaderRow || inDataRow {
					cellBuilder.Reset()
				}
			case "td":
				if inHeaderRow || inDataRow { // Cell content applies to both header and data rows
					cellBuilder.Reset()
				}
			}
		case html.TextToken:
			if inHeaderRow || inDataRow {
				cellBuilder.WriteString(token.Data)
			}
		case html.EndTagToken:
			switch token.Data {
			case "th":
				if inHeaderRow {
					headerContent := strings.TrimSpace(cellBuilder.String())
					if headerContent != "" { // Only add non-empty headers
						currentHeaders = append(currentHeaders, headerContent)
					}
					cellBuilder.Reset()
				} else if inDataRow {
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "td":
				if inHeaderRow {
					headerContent := strings.TrimSpace(cellBuilder.String())
					if headerContent != "" { // Only add non-empty headers
						currentHeaders = append(currentHeaders, headerContent)
					}
					cellBuilder.Reset()
				} else if inDataRow {
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "tr":
				if inHeaderRow && len(currentHeaders) > 0 {
					// Header row finished, potentially start workers if not started
					if !workersStarted && inTable && !inTBody {
						// Headers found and no <tbody> yet, so start workers now
						localHeadersForWorker = make([]string, len(currentHeaders))
						copy(localHeadersForWorker, currentHeaders)
						*headersSnapshot = localHeadersForWorker
						LogDebug("Headers found (tr end), launching %d workers with %d headers", numWorkers, len(localHeadersForWorker))
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
					}
				} else if inDataRow && len(currentCells) > 0 && workersStarted && !channelClosed {
					// Data row finished, send to workers with backpressure handling (only if channel not closed)
					sendRowWithBackpressure(rowCellsChan, currentCells, 5*time.Second)
				}
				// Reset row state
				inHeaderRow = false
				inDataRow = false
			case "table":
				inTable = false
				// End of table, ensure workers are started if headers were found
				if !workersStarted && len(currentHeaders) > 0 {
					localHeadersForWorker = make([]string, len(currentHeaders))
					copy(localHeadersForWorker, currentHeaders)
					*headersSnapshot = localHeadersForWorker
					LogDebug("Headers found (table end), launching %d workers with %d headers", numWorkers, len(localHeadersForWorker))
					wg.Add(numWorkers)
					for i := 0; i < numWorkers; i++ {
						go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
					}
					workersStarted = true
				}
			case "thead":
				inTHead = false
			case "tbody":
				inTBody = false
			}
		}
	}

	// Final cleanup and worker notification
	if workersStarted {
		LogDebug("HTML parsing finished, closing rowCellsChan to signal %d workers", numWorkers)
		closeChannelOnce() // Safe channel close
	} else {
		log.Printf("Warning: No workers were started during HTML parsing. Headers found: %v", len(currentHeaders) > 0)
		if len(currentHeaders) == 0 {
			processingError = errors.New("no table headers found in HTML file")
		} else {
			processingError = errors.New("headers found but workers were not started")
		}
		closeChannelOnce() // Ensure channel is closed even on error
	}

	return processingError
}
