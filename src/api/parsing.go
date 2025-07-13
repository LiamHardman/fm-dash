package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// Comprehensive list of world currencies used in Football Manager
var worldCurrencies = []string{
	// Major currencies (most common in FM)
	"$", "€", "£", "¥", "₹", "₽", "₺", "₩", "R$", "CHF", "A$", "CA$", "Mex$", "kr", "zł", "R",

	// European currencies
	"SEK", "NOK", "DKK", "ISK", "CZK", "PLN", "HUF", "RON", "BGN", "HRK", "RSD", "MKD", "ALL", "BAM", "MDL", "UAH", "BYN", "GEL", "AMD", "AZN", "KZT", "KGS", "TJS", "TMT", "UZS", "TND", "MAD", "EGP", "LYD", "DZD", "TND", "MAD", "EGP", "LYD", "DZD",

	// Asian currencies
	"CNY", "HKD", "TWD", "SGD", "MYR", "THB", "IDR", "PHP", "VND", "KHR", "LAK", "MMK", "BDT", "LKR", "NPR", "PKR", "AFN", "IRR", "IQD", "JOD", "KWD", "LBP", "OMR", "QAR", "SAR", "SYP", "AED", "YER", "BHD", "KWD", "OMR", "QAR", "SAR", "AED", "BHD",

	// African currencies
	"ZAR", "NGN", "KES", "GHS", "UGX", "TZS", "MWK", "ZMW", "BWP", "NAD", "SZL", "LSL", "MUR", "SCR", "KMF", "DJF", "ETB", "SDG", "SSP", "SOS", "TND", "MAD", "EGP", "LYD", "DZD", "MAD", "EGP", "LYD", "DZD",

	// American currencies
	"ARS", "BRL", "CLP", "COP", "PEN", "UYU", "PYG", "BOB", "GTQ", "HNL", "NIO", "CRC", "PAB", "BBD", "BZD", "GYD", "SRD", "TTD", "XCD", "ANG", "AWG", "BMD", "KYD", "JMD", "HTG", "DOP", "CUC", "CUP",

	// Oceanian currencies
	"FJD", "PGK", "SBD", "VUV", "WST", "TOP", "TVD", "NZD", "AUD",

	// Middle Eastern currencies
	"ILS", "TRY", "IRR", "IQD", "JOD", "KWD", "LBP", "OMR", "QAR", "SAR", "SYP", "AED", "YER", "BHD",

	// Other major currencies
	"INR", "RUB", "KRW", "JPY", "CNY", "HKD", "TWD", "SGD", "MYR", "THB", "IDR", "PHP", "VND", "KHR", "LAK", "MMK", "BDT", "LKR", "NPR", "PKR", "AFN", "IRR", "IQD", "JOD", "KWD", "LBP", "OMR", "QAR", "SAR", "SYP", "AED", "YER", "BHD",
}

// Optimized string builder pool with size management
var optimizedStringBuilderPool = sync.Pool{
	New: func() interface{} {
		sb := &strings.Builder{}
		sb.Grow(stringBuilderInitSize)
		return sb
	},
}

func getOptimizedStringBuilder() *strings.Builder {
	sb := optimizedStringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

func putOptimizedStringBuilder(sb *strings.Builder) {
	if sb.Cap() <= maxStringBuilderSize { // Don't pool extremely large builders
		optimizedStringBuilderPool.Put(sb)
	}
}

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

// Enhanced currency detection with comprehensive world currency support
func detectCurrencySymbol(rawValue string) string {
	// First check for the most common currencies (faster path)
	commonCurrencies := []string{"$", "€", "£", "¥", "₹", "₽", "₺", "₩", "R$", "CHF", "A$", "CA$", "Mex$", "kr", "zł", "R"}

	for _, sym := range commonCurrencies {
		if strings.Contains(rawValue, sym) {
			return sym
		}
	}

	// If no common currency found, check the full world currency list
	for _, sym := range worldCurrencies {
		if strings.Contains(rawValue, sym) {
			return sym
		}
	}

	return ""
}

// FastParseMonetaryValue parses monetary values, handling ranges and decimals correctly.
func FastParseMonetaryValue(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	originalDisplay = rawValue
	cleanValue := rawValue
	detectedSymbol = detectCurrencySymbol(rawValue)

	// Remove all currency symbols and whitespace
	for _, sym := range worldCurrencies {
		cleanValue = strings.ReplaceAll(cleanValue, sym, "")
	}
	cleanValue = strings.TrimSpace(cleanValue)

	// If it's a range (e.g., '£140M - £183M'), extract the upper bound
	if strings.Contains(cleanValue, "-") {
		parts := strings.Split(cleanValue, "-")
		if len(parts) > 1 {
			cleanValue = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	// Detect multiplier
	multiplier := float64(1)
	if cleanValue != "" {
		lastChar := cleanValue[len(cleanValue)-1]
		switch lastChar {
		case 'M', 'm':
			multiplier = 1000000
			cleanValue = cleanValue[:len(cleanValue)-1]
		case 'K', 'k':
			multiplier = 1000
			cleanValue = cleanValue[:len(cleanValue)-1]
		}
	}

	// Parse as float to handle decimals
	valFloat, err := strconv.ParseFloat(strings.TrimSpace(cleanValue), 64)
	if err == nil {
		numericValue = int64(valFloat * multiplier)
	} else {
		numericValue = 0
	}

	// Set default symbol if needed
	if detectedSymbol == "" && numericValue != 0 {
		detectedSymbol = "$"
	}

	return originalDisplay, numericValue, detectedSymbol
}

// Enhanced ParseMonetaryValueGo with optimized byte operations (maintains compatibility)
func ParseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	return FastParseMonetaryValue(rawValue)
}

// ParseHTMLPlayerTable tokenizes an HTML file stream (typically a player squad view)
// and sends extracted table headers and row data (as slices of strings) to respective channels.
// It manages the HTML parsing state machine with optimized performance.
func ParseHTMLPlayerTable(file io.Reader, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) (processingError error) {
	// Use larger buffered reader for better performance
	bufferedReader := bufio.NewReaderSize(file, maxTokenBufferSize)
	tokenizer := html.NewTokenizer(bufferedReader)

	var currentHeaders []string // Temporary headers collected
	var currentCells []string   // Cells for the current <tr> being processed
	inHeaderRow := false        // True if currently inside a <tr> identified as a header row
	inDataRow := false          // True if currently inside a <tr> identified as a data row
	inTable := false            // True if currently inside a <table> element
	inTHead := false            // True if currently inside a <thead> element
	inTBody := false            // True if currently inside a <tbody> element

	// Use optimized string builder for better performance
	cellBuilder := getOptimizedStringBuilder()
	defer putOptimizedStringBuilder(cellBuilder)

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
