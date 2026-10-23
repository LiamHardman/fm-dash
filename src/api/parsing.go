package main

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
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
		LogWarn("Channel send timeout, dropping row - timeout: %v, cell_count: %d", timeout, len(cells))
	}
}

// Enhanced currency detection with comprehensive world currency support
func detectCurrencySymbol(rawValue string) string {
	for _, sym := range []string{"Mex$", "CA$", "A$", "R$"} {
		if strings.Contains(rawValue, sym) {
			return sym
		}
	}

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
	originalDisplay = strings.TrimSpace(rawValue)
	cleanValue := rawValue
	detectedSymbol = detectCurrencySymbol(rawValue)

	// Remove all currency symbols and whitespace
	if detectedSymbol != "" {
		cleanValue = strings.ReplaceAll(cleanValue, detectedSymbol, "")
	}
	for _, sym := range worldCurrencies {
		cleanValue = strings.ReplaceAll(cleanValue, sym, "")
	}
	cleanValue = strings.TrimSpace(cleanValue)

	// Remove per week indicators (p/w, /w, per week, etc.)
	cleanValue = strings.ReplaceAll(cleanValue, "p/w", "")
	cleanValue = strings.ReplaceAll(cleanValue, "P/W", "")
	cleanValue = strings.ReplaceAll(cleanValue, "/w", "")
	cleanValue = strings.ReplaceAll(cleanValue, "/W", "")
	cleanValue = strings.ReplaceAll(cleanValue, "per week", "")
	cleanValue = strings.ReplaceAll(cleanValue, "Per Week", "")
	cleanValue = strings.TrimSpace(cleanValue)

	// Remove commas from numbers (e.g., "47,500" -> "47500")
	cleanValue = strings.ReplaceAll(cleanValue, ",", "")

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

// ParseMonetaryValueGo parses monetary values with optimized byte operations
func ParseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	return FastParseMonetaryValue(rawValue)
}

// ParseHTMLPlayerTable tokenizes an HTML file stream (typically a player squad view)
// and sends extracted table headers and row data (as slices of strings) to respective channels.
// It manages the HTML parsing state machine with optimized performance.
func ParseHTMLPlayerTable(file io.Reader, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) error {
	return parseLegacyTable(file, "html", headersSnapshot, rowCellsChan, numWorkers, resultsChan, wg)
}

func parseLegacyTable(file io.Reader, format string, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) error {
	defer close(rowCellsChan)
	return scanExport(context.Background(), file, format, func(_ int, cells []string, header bool) error {
		if header {
			*headersSnapshot = normalizeCSVHeaders(cells)
			wg.Add(numWorkers)
			for i := 0; i < numWorkers; i++ {
				go PlayerParserWorker(i, rowCellsChan, resultsChan, wg, *headersSnapshot)
			}
		} else {
			rowCellsChan <- cells
		}
		return nil
	})
}

// csvHeaderNormalize maps CSV-specific header names to the canonical header names
// used by parseCellsToPlayer, so the same worker logic handles both HTML and CSV exports.
// FM's CSV export uses full attribute names; the HTML view uses the short abbreviations
// that all downstream rating calculations expect.
var csvHeaderNormalize = map[string]string{
	// Column identity differences
	"Player":         "Name",
	"Nation":         "Nat",
	"League":         "Division",
	"Competition":    "Division",
	"Based In":       "Based in",
	"Rating":         "Av Rat",
	"Average Rating": "Av Rat",
	"Appearances":    "Apps",
	"Aps":            "Apps",

	// Physical attributes
	"Acceleration":    "Acc",
	"Pace":            "Pac",
	"Strength":        "Str",
	"Stamina":         "Sta",
	"Natural Fitness": "Nat",
	"Balance":         "Bal",
	"Jumping Reach":   "Jum",
	"Agility":         "Agi",

	// Mental attributes
	"Aggression":    "Agg",
	"Anticipation":  "Ant",
	"Bravery":       "Bra",
	"Composure":     "Cmp",
	"Concentration": "Cnt",
	"Decisions":     "Dec",
	"Determination": "Det",
	"Flair":         "Fla",
	"Leadership":    "Ldr",
	"Off The Ball":  "OtB",
	"Positioning":   "Pos",
	"Team Work":     "Tea",
	"Vision":        "Vis",
	"Work Rate":     "Wor",

	// Technical attributes
	"Corners":          "Cor",
	"Crossing":         "Cro",
	"Dribbling":        "Dri",
	"Finishing":        "Fin",
	"First Touch":      "Fir",
	"Free Kick Taking": "Fre",
	"Heading":          "Hea",
	"Long Shots":       "Lon",
	"Long Throws":      "L Th",
	"Marking":          "Mar",
	"Passing":          "Pas",
	"Penalty Taking":   "Pen",
	"Tackling":         "Tck",
	"Technique":        "Tec",

	// Performance stat name differences (CSV vs HTML)
	"KP/90":                "K Ps/90",
	"Crs A/90":             "CRS A/90",
	"Goals per 90 minutes": "Gls/90",
	"Average Rating Club":  "Av Rat",
	"Fouls Against":        "FA",

	// Goalkeeper attributes
	"Aerial Reach":           "Aer",
	"Command Of Area":        "Cmd",
	"Communication":          "Com",
	"Eccentricity":           "Ecc",
	"Handling":               "Han",
	"Kicking":                "Kic",
	"One On Ones":            "1v1",
	"Reflexes":               "Ref",
	"Rushing Out (Tendency)": "TRO",
	"Throwing":               "Thr",
	"Punching":               "Pun",
}

// ParseCSVPlayerTable parses a semicolon-delimited CSV player export (as produced by FM's
// "Export to CSV" view option) and feeds rows into the same worker pipeline used by
// ParseHTMLPlayerTable. The first non-empty line is treated as the header row.
func ParseCSVPlayerTable(file io.Reader, headersSnapshot *[]string, rowCellsChan chan []string, numWorkers int, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup) error {
	return parseLegacyTable(file, "csv", headersSnapshot, rowCellsChan, numWorkers, resultsChan, wg)
}
