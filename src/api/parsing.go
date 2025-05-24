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

	"golang.org/x/net/html"
)

// Updated currencySymbolRegex to include more symbols.
// This regex tries to match common currency symbols or specific multi-character codes.
// It's important to order alternatives carefully if some are substrings of others (e.g., R vs R$).
// Using non-capturing groups (?:...) for multi-character codes if needed, but simple alternation should work.
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

// ParseMonetaryValueGo extracts the original display string, a numeric value, and a detected currency symbol
// from a raw monetary string (e.g., "$1.5M", "£25K p/w", "¥100M").
func ParseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	cleanedValue := strings.TrimSpace(rawValue)
	originalDisplay = cleanedValue // Store the original formatting

	// Try to detect currency symbol from the raw value using the updated regex
	matches := currencySymbolRegex.FindStringSubmatch(cleanedValue)
	// FindStringSubmatch returns an array where matches[0] is the full match,
	// and subsequent elements are the captures of each group in the regex.
	// We iterate through the capture groups to find which one matched.
	if len(matches) > 1 {
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				detectedSymbol = matches[i]
				break
			}
		}
	}
	if detectedSymbol == "" {
		// Fallback if no specific symbol from the regex is found, but common ones might exist without being in the regex groups
		// This part might be redundant if the regex is comprehensive.
		if strings.Contains(cleanedValue, "€") {
			detectedSymbol = "€"
		} else if strings.Contains(cleanedValue, "£") {
			detectedSymbol = "£"
		} else if strings.Contains(cleanedValue, "$") { // This could be USD, CAD, AUD etc.
			detectedSymbol = "$" // Default to $ if only $ is found
		}
	}

	// Handle ranges like "£10M - £15M", take the higher value or average if needed.
	// Current implementation takes the part after " - " if present.
	if strings.Contains(cleanedValue, " - ") {
		parts := strings.Split(cleanedValue, " - ")
		if len(parts) == 2 {
			cleanedValue = strings.TrimSpace(parts[1]) // Use the second part of the range
			// Re-detect symbol if it was in the second part
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

	// Remove currency symbols and suffixes for parsing
	// Create a list of all symbols/codes from the regex to remove them for numeric parsing.
	// This needs to be robust.
	valueToParse := cleanedValue
	if detectedSymbol != "" { // If a symbol was detected, try removing it specifically
		valueToParse = strings.ReplaceAll(valueToParse, detectedSymbol, "")
	}
	// Generic removal of common symbols as a fallback or additional cleanup
	valueToParse = strings.ReplaceAll(valueToParse, "€", "")
	valueToParse = strings.ReplaceAll(valueToParse, "£", "")
	valueToParse = strings.ReplaceAll(valueToParse, "$", "") // Catches general dollars
	valueToParse = strings.ReplaceAll(valueToParse, "¥", "")
	valueToParse = strings.ReplaceAll(valueToParse, "₹", "")
	valueToParse = strings.ReplaceAll(valueToParse, "₽", "")
	valueToParse = strings.ReplaceAll(valueToParse, "₺", "")
	valueToParse = strings.ReplaceAll(valueToParse, "₩", "")
	valueToParse = strings.ReplaceAll(valueToParse, "R", "") // For R, R$, ensure R$ is handled first by regex
	valueToParse = strings.ReplaceAll(valueToParse, "CHF", "")
	valueToParse = strings.ReplaceAll(valueToParse, "kr", "")
	valueToParse = strings.ReplaceAll(valueToParse, "zł", "")
	// Ensure multi-character codes like CA$, A$, Mex$ are also removed if they were the detectedSymbol
	// The current logic of `strings.ReplaceAll(valueToParse, detectedSymbol, "")` should handle this.

	valueToParse = strings.TrimSpace(strings.ReplaceAll(valueToParse, "p/w", "")) // Per week
	valueToParse = strings.TrimSpace(strings.ReplaceAll(valueToParse, "/w", ""))  // Per week (alternative)

	multiplier := int64(1)
	if strings.HasSuffix(strings.ToUpper(valueToParse), "M") {
		multiplier = 1000000
		valueToParse = strings.TrimRight(valueToParse, "Mm")
	} else if strings.HasSuffix(strings.ToUpper(valueToParse), "K") {
		multiplier = 1000
		valueToParse = strings.TrimRight(valueToParse, "Kk")
	}

	valueToParse = strings.ReplaceAll(valueToParse, ",", "") // Remove commas

	valFloat, err := strconv.ParseFloat(strings.TrimSpace(valueToParse), 64)
	if err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	} else {
		numericValue = 0 // Default to 0 if parsing fails
	}

	// If no symbol was detected initially, but we have a value, set to default $
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
	var cellBuilder strings.Builder

	workersStarted := false
	var localHeadersForWorker []string // To pass to workers once finalized

tokenLoop:
	for {
		tt := tokenizer.Next()
		token := tokenizer.Token()

		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				if inDataRow && len(currentCells) > 0 && workersStarted {
					cellsCopy := make([]string, len(currentCells))
					copy(cellsCopy, currentCells)
					rowCellsChan <- cellsCopy
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
				currentCells = make([]string, 0, defaultCellCapacity) // defaultCellCapacity from config.go
				// Determine if this TR is a header row or data row
				if inTHead {
					inHeaderRow = true
					inDataRow = false
				} else if inTable && !workersStarted && len(currentHeaders) == 0 && !inTBody {
					// If in a table, workers haven't started, no headers collected yet, AND not in <tbody>:
					// Treat this as a potential header row.
					inHeaderRow = true
					inDataRow = false
				} else {
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
				if inHeaderRow { // If this row was marked as a header row, <td> can be a header
					headerContent := strings.TrimSpace(cellBuilder.String())
					if headerContent != "" { // Only add non-empty headers
						currentHeaders = append(currentHeaders, headerContent)
					}
					cellBuilder.Reset()
				} else if inDataRow { // If this row was marked as a data row
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "tr":
				if inHeaderRow { // This was a header row that just ended
					inHeaderRow = false // Reset for the next row
					// If workers haven't started AND we now have some headers, start them.
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
				} else if inDataRow { // This was a data row that just ended
					inDataRow = false // Reset for the next row
					if len(currentCells) > 0 && workersStarted {
						cellsCopy := make([]string, len(currentCells))
						copy(cellsCopy, currentCells)
						rowCellsChan <- cellsCopy
					} else if len(currentCells) > 0 && !workersStarted && len(localHeadersForWorker) > 0 {
						// Fallback: headers were found, but workers didn't start (e.g. no tbody, thead).
						// This data row ending might be the trigger.
						log.Printf("Fallback: Headers exist, workers not started, data row found. Launching workers.")
						*headersSnapshot = localHeadersForWorker // Ensure snapshot is set
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, localHeadersForWorker)
						}
						workersStarted = true
						cellsCopy := make([]string, len(currentCells)) // Send the current row
						copy(cellsCopy, currentCells)
						rowCellsChan <- cellsCopy
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

	// Fallback after loop: if workers haven't started but headers were collected (e.g. very short table)
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
