package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// BenchmarkParseMonetaryValueGoOptimized benchmarks the optimized monetary value parsing
func BenchmarkParseMonetaryValueGoOptimized(b *testing.B) {
	testCases := []string{
		"$1.5M",
		"£25K p/w",
		"€500K",
		"R$2.5M",
		"CHF1.2M",
		"A$800K",
		"CA$1M",
		"Mex$5M",
		"kr500K",
		"zł2M",
		"R1.5M",
		"¥100M",
		"₹50M",
		"₽25M",
		"₺10M",
		"₩1000M",
		"$1,500,000",
		"£2.75M",
		"1.5M",
		"£10M - £15M",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range testCases {
			ParseMonetaryValueGo(testCase)
		}
	}
}

// BenchmarkHTMLParsingOptimized benchmarks the HTML parsing performance
func BenchmarkHTMLParsingOptimized(b *testing.B) {
	// Create a sample HTML table for testing
	htmlContent := generateSampleHTMLTable(1000) // 1000 rows
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(htmlContent)
		var headers []string
		rowCellsChan := make(chan []string, 100)
		resultsChan := make(chan PlayerParseResult, 100)
		var wg sync.WaitGroup
		
		ParseHTMLPlayerTable(reader, &headers, rowCellsChan, 4, resultsChan, &wg)
	}
}

// BenchmarkPlayerProcessingOptimized benchmarks the player processing performance
func BenchmarkPlayerProcessingOptimized(b *testing.B) {
	// Create sample player data
	headers := []string{"Name", "Position", "Age", "Club", "Division", "Transfer Value", "Wage", "PAC", "SHO", "PAS", "DRI", "DEF", "PHY"}
	sampleCells := []string{"John Doe", "ST", "25", "Test Club", "Premier League", "$10M", "£50K p/w", "15", "16", "12", "14", "8", "13"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseCellsToPlayer(sampleCells, headers)
	}
}

// BenchmarkPercentileCalculationOptimized benchmarks the percentile calculation performance
func BenchmarkPercentileCalculationOptimized(b *testing.B) {
	// Create sample players for testing
	players := generateSamplePlayers(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a copy to avoid modifying the original
		playersCopy := make([]Player, len(players))
		copy(playersCopy, players)
		CalculatePlayerPerformancePercentiles(playersCopy)
	}
}

// BenchmarkUploadHandlerOptimized benchmarks the upload handler performance
func BenchmarkUploadHandlerOptimized(b *testing.B) {
	// Create a sample HTML file content
	htmlContent := generateSampleHTMLTable(500) // 500 rows
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the upload processing pipeline
		processUploadContent(htmlContent)
	}
}

// Helper functions for benchmarks

func generateSampleHTMLTable(rowCount int) string {
	var buf bytes.Buffer
	
	buf.WriteString(`<html><body><table><thead><tr>`)
	buf.WriteString(`<th>Name</th><th>Position</th><th>Age</th><th>Club</th><th>Division</th>`)
	buf.WriteString(`<th>Transfer Value</th><th>Wage</th><th>PAC</th><th>SHO</th><th>PAS</th><th>DRI</th><th>DEF</th><th>PHY</th>`)
	buf.WriteString(`</tr></thead><tbody>`)
	
	for i := 0; i < rowCount; i++ {
		buf.WriteString(fmt.Sprintf(`<tr><td>Player %d</td><td>ST</td><td>%d</td><td>Club %d</td><td>Division %d</td>`, 
			i, 20+(i%20), i%10, i%5))
		buf.WriteString(fmt.Sprintf(`<td>$%dM</td><td>£%dK p/w</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`,
			1+(i%10), 10+(i%50), 10+(i%10), 10+(i%10), 10+(i%10), 10+(i%10), 10+(i%10), 10+(i%10)))
	}
	
	buf.WriteString(`</tbody></table></body></html>`)
	return buf.String()
}

func generateSamplePlayers(count int) []Player {
	players := make([]Player, count)
	
	for i := 0; i < count; i++ {
		players[i] = Player{
			Name:     fmt.Sprintf("Player %d", i),
			Position: "ST",
			Age:      fmt.Sprintf("%d", 20+(i%20)),
			Club:     fmt.Sprintf("Club %d", i%10),
			Division: fmt.Sprintf("Division %d", i%5),
			TransferValue: fmt.Sprintf("$%dM", 1+(i%10)),
			Wage:     fmt.Sprintf("£%dK p/w", 10+(i%50)),
			Attributes: map[string]string{
				"PAC": fmt.Sprintf("%d", 10+(i%10)),
				"SHO": fmt.Sprintf("%d", 10+(i%10)),
				"PAS": fmt.Sprintf("%d", 10+(i%10)),
				"DRI": fmt.Sprintf("%d", 10+(i%10)),
				"DEF": fmt.Sprintf("%d", 10+(i%10)),
				"PHY": fmt.Sprintf("%d", 10+(i%10)),
			},
			NumericAttributes: map[string]int{
				"PAC": 10 + (i % 10),
				"SHO": 10 + (i % 10),
				"PAS": 10 + (i % 10),
				"DRI": 10 + (i % 10),
				"DEF": 10 + (i % 10),
				"PHY": 10 + (i % 10),
			},
			PerformanceStatsNumeric: map[string]float64{
				"Gls/90": float64(i % 20),
				"Asts/90": float64(i % 15),
				"Av Rat": float64(6.0 + float64(i%40)/10.0),
			},
			PerformancePercentiles: make(map[string]map[string]float64),
			ParsedPositions: []string{"ST"},
			ShortPositions: []string{"ST"},
			PositionGroups: []string{"Attackers"},
		}
	}
	
	return players
}

func processUploadContent(htmlContent string) {
	// Simulate the upload processing pipeline
	reader := strings.NewReader(htmlContent)
	var headers []string
	rowCellsChan := make(chan []string, 100)
	resultsChan := make(chan PlayerParseResult, 100)
	var wg sync.WaitGroup
	
	// Parse HTML
	ParseHTMLPlayerTable(reader, &headers, rowCellsChan, 4, resultsChan, &wg)
	
	// Collect results
	var players []Player
	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
		}
	}
	
	// Calculate percentiles
	if len(players) > 0 {
		CalculatePlayerPerformancePercentiles(players)
	}
}

// Performance comparison tests

func TestPerformanceImprovements(t *testing.T) {
	// Test monetary parsing performance
	start := time.Now()
	for i := 0; i < 10000; i++ {
		ParseMonetaryValueGo("$1.5M")
		ParseMonetaryValueGo("£25K p/w")
		ParseMonetaryValueGo("€500K")
	}
	duration := time.Since(start)
	
	t.Logf("Monetary parsing performance: %d operations in %v", 30000, duration)
	
	// Test HTML parsing performance
	htmlContent := generateSampleHTMLTable(100)
	start = time.Now()
	for i := 0; i < 100; i++ {
		reader := strings.NewReader(htmlContent)
		var headers []string
		rowCellsChan := make(chan []string, 50)
		resultsChan := make(chan PlayerParseResult, 50)
		var wg sync.WaitGroup
		
		ParseHTMLPlayerTable(reader, &headers, rowCellsChan, 4, resultsChan, &wg)
	}
	duration = time.Since(start)
	
	t.Logf("HTML parsing performance: %d operations in %v", 100, duration)
	
	// Test percentile calculation performance
	players := generateSamplePlayers(500)
	start = time.Now()
	CalculatePlayerPerformancePercentiles(players)
	duration = time.Since(start)
	
	t.Logf("Percentile calculation performance: %d players in %v", 500, duration)
}

// Memory usage tests

func TestMemoryUsage(t *testing.T) {
	// Test memory usage for large datasets
	players := generateSamplePlayers(10000)
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc
	
	CalculatePlayerPerformancePercentiles(players)
	
	runtime.ReadMemStats(&m)
	afterAlloc := m.Alloc
	
	memoryUsed := afterAlloc - beforeAlloc
	t.Logf("Memory usage for 10,000 players: %d bytes (%.2f MB)", memoryUsed, float64(memoryUsed)/1024/1024)
	
	// Force garbage collection to see actual memory usage
	runtime.GC()
	runtime.ReadMemStats(&m)
	t.Logf("Memory after GC: %d bytes (%.2f MB)", m.Alloc, float64(m.Alloc)/1024/1024)
} 