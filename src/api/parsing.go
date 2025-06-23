// src/api/parsing.go
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

// Object pools for reducing allocations during parsing
var (
	cellBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
	
	cellsSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, defaultCellCapacity)
		},
	}
)

// sendRowWithBackpressure attempts to send row data to channel with timeout handling
func sendRowWithBackpressure(rowCellsChan chan []string, cells []string, timeout time.Duration) {
	// Reuse the cells slice from pool instead of copying
	select {
	case rowCellsChan <- cells:
		// Successfully sent
	case <-time.After(timeout):
		log.Printf("Warning: Channel send timeout after %v, dropping row with %d cells", timeout, len(cells))
		// Return cells to pool since we're not using them
		cellsSlicePool.Put(cells[:0]) // Reset slice before returning to pool
	}
}

// Optimized currency symbol regex with better ordering and non-capturing groups
var currencySymbolRegex = regexp.MustCompile(
	`(?:R\$|CHF|A\$|CA\$|Mex\$|kr|zł)|` + // Multi-character codes first
	`[€£$¥₹₽₺₩R]`, // Single character symbols last
)

// ParseMonetaryValueGo extracts the original display string, a numeric value, and a detected currency symbol
// from a raw monetary string (e.g., "$1.5M", "£25K p/w", "¥100M").
func ParseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	cleanedValue := strings.TrimSpace(rawValue)
	originalDisplay = cleanedValue

	// Optimized currency detection - single regex match
	if match := currencySymbolRegex.FindString(cleanedValue); match != "" {
		detectedSymbol = match
	}

	// Handle ranges like "£10M - £15M", take the higher value
	if strings.Contains(cleanedValue, " - ") {
		parts := strings.SplitN(cleanedValue, " - ", 2)
		if len(parts) == 2 {
			cleanedValue = strings.TrimSpace(parts[1])
			// Re-detect symbol if it was in the second part
			if match := currencySymbolRegex.FindString(cleanedValue); match != "" {
				detectedSymbol = match
			}
		}
	}

	// Optimized symbol removal - single pass
	valueToParse := cleanedValue
	if detectedSymbol != "" {
		valueToParse = strings.ReplaceAll(valueToParse, detectedSymbol, "")
	}

	// Remove common suffixes in one pass
	valueToParse = strings.ReplaceAll(valueToParse, "p/w", "")
	valueToParse = strings.ReplaceAll(valueToParse, "/w", "")
	valueToParse = strings.TrimSpace(valueToParse)

	// Optimized multiplier detection
	multiplier := int64(1)
	upperValue := strings.ToUpper(valueToParse)
	if strings.HasSuffix(upperValue, "M") {
		multiplier = 1000000
		valueToParse = strings.TrimRight(valueToParse, "Mm")
	} else if strings.HasSuffix(upperValue, "K") {
		multiplier = 1000
		valueToParse = strings.TrimRight(valueToParse, "Kk")
	}

	// Remove commas and parse
	valueToParse = strings.ReplaceAll(valueToParse, ",", "")
	
	if valFloat, err := strconv.ParseFloat(strings.TrimSpace(valueToParse), 64); err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	}

	// Default to $ if no symbol detected but we have a value
	if detectedSymbol == "" && numericValue != 0 {
		detectedSymbol = "$"
	}

	return originalDisplay, numericValue, detectedSymbol
}

// ParseHTMLPlayerTable tokenizes an HTML file stream (typically a player squad view)
// and sends extracted table headers and row data (as slices of strings) to respective channels.
// It manages the HTML parsing state machine.
func ParseHTMLPlayerTable(file io.Reader, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) (processingError error) {
	bufferedReader := bufio.NewReaderSize(file, maxTokenBufferSize)
	tokenizer := html.NewTokenizer(bufferedReader)

	var currentHeaders []string
	var currentCells []string
	inHeaderRow := false
	inDataRow := false
	inTable := false
	inTHead := false
	inTBody := false
	
	// Get cellBuilder from pool
	cellBuilder := cellBuilderPool.Get().(*strings.Builder)
	defer cellBuilderPool.Put(cellBuilder)

	workersStarted := false
	var localHeadersForWorker []string

tokenLoop:
	for {
		tt := tokenizer.Next()
		token := tokenizer.Token()

		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				if inDataRow && len(currentCells) > 0 && workersStarted {
					sendRowWithBackpressure(rowCellsChan, currentCells, 5*time.Second)
				}
				break tokenLoop
			}
			log.Printf("HTML tokenization error: %v", err)
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
						log.Printf("Headers found (tbody start), launching %d workers: %v", numWorkers, localHeadersForWorker)
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
					}
				}
			case "tr":
				// Get cells slice from pool
				currentCells = cellsSlicePool.Get().([]string)
				currentCells = currentCells[:0] // Reset slice
				
				// Determine if this TR is a header row or data row
				switch {
				case inTHead:
					inHeaderRow = true
					inDataRow = false
				case inTable && !workersStarted && len(currentHeaders) == 0 && !inTBody:
					inHeaderRow = true
					inDataRow = false
				default:
					inHeaderRow = false
					inDataRow = true
				}
			case "th", "td":
				if inHeaderRow || inDataRow {
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
					if headerContent != "" {
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
					if headerContent != "" {
						currentHeaders = append(currentHeaders, headerContent)
					}
					cellBuilder.Reset()
				} else if inDataRow {
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "tr":
				if inHeaderRow {
					inHeaderRow = false
					if !workersStarted && len(currentHeaders) > 0 {
						localHeadersForWorker = make([]string, len(currentHeaders))
						copy(localHeadersForWorker, currentHeaders)
						*headersSnapshot = localHeadersForWorker
						log.Printf("Headers collected (end of header tr), launching %d workers: %v", numWorkers, localHeadersForWorker)
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
					}
				} else if inDataRow {
					inDataRow = false
					if len(currentCells) > 0 && workersStarted {
						sendRowWithBackpressure(rowCellsChan, currentCells, 5*time.Second)
					} else if len(currentCells) > 0 && !workersStarted && len(localHeadersForWorker) > 0 {
						log.Printf("Fallback: Headers exist, workers not started, data row found. Launching workers.")
						*headersSnapshot = localHeadersForWorker
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
						sendRowWithBackpressure(rowCellsChan, currentCells, 5*time.Second)
					} else {
						// Return unused cells to pool
						cellsSlicePool.Put(currentCells)
					}
				}
			case "thead":
				inTHead = false
				if !workersStarted && len(currentHeaders) > 0 {
					localHeadersForWorker = make([]string, len(currentHeaders))
					copy(localHeadersForWorker, currentHeaders)
					*headersSnapshot = localHeadersForWorker
					log.Printf("Headers found (thead end), launching %d workers: %v", numWorkers, localHeadersForWorker)
					wg.Add(numWorkers)
					for i := 0; i < numWorkers; i++ {
						go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
					}
					workersStarted = true
				}
			case "tbody":
				inTBody = false
			case "table":
				inTable = false
			}
		}
	}

	// Fallback after loop: if workers haven't started but headers were collected
	if !workersStarted && len(currentHeaders) > 0 {
		localHeadersForWorker = make([]string, len(currentHeaders))
		copy(localHeadersForWorker, currentHeaders)
		*headersSnapshot = localHeadersForWorker
		log.Printf("Headers found (fallback after token loop), launching %d workers: %v", numWorkers, localHeadersForWorker)
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
		}
		workersStarted = true
	}

	if !workersStarted && processingError == nil {
		log.Println("Critical: Workers were not started. No headers found or table structure unparsable.")
		if len(currentHeaders) == 0 && len(*headersSnapshot) == 0 {
			processingError = errors.New("could not parse table headers, no data processed")
		}
	}

	return processingError
}
