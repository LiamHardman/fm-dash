package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"testing"

	"google.golang.org/protobuf/proto"
)

// setLogLevelForBenchmarks sets the log level to INFO to reduce benchmark noise
func setLogLevelForBenchmarks(ctx context.Context) func() {
	logInfo(ctx, "Setting log level for benchmarks", "target_level", "INFO")
	
	// Set LOG_LEVEL environment variable to INFO to suppress DEBUG messages
	originalLogLevel := os.Getenv("LOG_LEVEL")
	if err := os.Setenv("LOG_LEVEL", "INFO"); err != nil {
		logError(ctx, "Failed to set LOG_LEVEL environment variable", "error", err)
		panic(fmt.Sprintf("Failed to set LOG_LEVEL: %v", err))
	}

	// Also set slog level to INFO
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(handler))

	logDebug(ctx, "Log level configured for benchmarks", "original_level", originalLogLevel)

	// Return a function to restore original settings
	return func() {
		logInfo(ctx, "Restoring original log level", "original_level", originalLogLevel)
		
		if originalLogLevel == "" {
			if err := os.Unsetenv("LOG_LEVEL"); err != nil {
				logError(ctx, "Failed to unset LOG_LEVEL environment variable", "error", err)
				panic(fmt.Sprintf("Failed to unset LOG_LEVEL: %v", err))
			}
		} else {
			if err := os.Setenv("LOG_LEVEL", originalLogLevel); err != nil {
				logError(ctx, "Failed to restore LOG_LEVEL environment variable", "error", err, "original_level", originalLogLevel)
				panic(fmt.Sprintf("Failed to restore LOG_LEVEL: %v", err))
			}
		}
		slog.SetDefault(originalLogger)
		
		logDebug(ctx, "Log level restored successfully", "restored_level", originalLogLevel)
	}
}

// BenchmarkProtobufVsJSON_Serialization benchmarks serialization performance
func BenchmarkProtobufVsJSON_Serialization(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	// Create test data with varying sizes
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
		{"XLarge_5000_Players", 5000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("JSON_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Protobuf_%s", tc.name), func(b *testing.B) {
			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkProtobufVsJSON_Deserialization benchmarks deserialization performance
func BenchmarkProtobufVsJSON_Deserialization(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
		{"XLarge_5000_Players", 5000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		// Pre-serialize data
		jsonData, err := json.Marshal(dataset)
		if err != nil {
			b.Fatal(err)
		}

		ctx := context.Background()
		protoDataset, err := dataset.ToProto(ctx)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("JSON_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result PlayerDataWithCurrency
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Protobuf_%s", tc.name), func(b *testing.B) {
			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkProtobufVsJSON_StorageSize benchmarks storage size comparison
func BenchmarkProtobufVsJSON_StorageSize(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
		{"XLarge_5000_Players", 5000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		// Serialize with JSON
		jsonData, err := json.Marshal(dataset)
		if err != nil {
			b.Fatal(err)
		}

		// Serialize with Protobuf
		ctx := context.Background()
		protoDataset, err := dataset.ToProto(ctx)
		if err != nil {
			b.Fatal(err)
		}
		protoData, err := proto.Marshal(protoDataset)
		if err != nil {
			b.Fatal(err)
		}

		// Calculate sizes and compression ratios
		jsonSize := len(jsonData)
		protoSize := len(protoData)
		compressionRatio := float64(jsonSize) / float64(protoSize)

		b.Logf("%s - JSON: %d bytes, Protobuf: %d bytes, Compression: %.2fx",
			tc.name, jsonSize, protoSize, compressionRatio)

		// Benchmark is just for logging - no actual operations needed
		b.Run(fmt.Sprintf("Size_%s", tc.name), func(b *testing.B) {
			// This benchmark is for size comparison logging only
			b.ReportMetric(float64(jsonSize), "json_bytes")
			b.ReportMetric(float64(protoSize), "protobuf_bytes")
			b.ReportMetric(compressionRatio, "compression_ratio")
		})
	}
}

// BenchmarkProtobufVsJSON_MemoryUsage benchmarks memory usage during operations
func BenchmarkProtobufVsJSON_MemoryUsage(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("JSON_Memory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				jsonData, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}

				var result PlayerDataWithCurrency
				err = json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)
			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
		})

		b.Run(fmt.Sprintf("Protobuf_Memory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)
			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
		})
	}
}

// BenchmarkProtobufStorage_Operations benchmarks complete storage operations
func BenchmarkProtobufStorage_Operations(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := DatasetData{
			Players:        players,
			CurrencySymbol: "£",
		}

		// JSON Storage Benchmark
		b.Run(fmt.Sprintf("JSON_Storage_%s", tc.name), func(b *testing.B) {
			jsonStorage := CreateInMemoryStorage()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				datasetID := fmt.Sprintf("benchmark-json-%d", i)

				// Store
				err := jsonStorage.Store(datasetID, dataset)
				if err != nil {
					b.Fatal(err)
				}

				// Retrieve
				_, err = jsonStorage.Retrieve(datasetID)
				if err != nil {
					b.Fatal(err)
				}

				// Clean up
				err = jsonStorage.Delete(datasetID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// Protobuf Storage Benchmark
		b.Run(fmt.Sprintf("Protobuf_Storage_%s", tc.name), func(b *testing.B) {
			backend := CreateInMemoryStorage()
			protobufStorage := CreateProtobufStorage(backend)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				datasetID := fmt.Sprintf("benchmark-protobuf-%d", i)

				// Store
				err := protobufStorage.Store(datasetID, dataset)
				if err != nil {
					b.Fatal(err)
				}

				// Retrieve
				_, err = protobufStorage.Retrieve(datasetID)
				if err != nil {
					b.Fatal(err)
				}

				// Clean up
				err = protobufStorage.Delete(datasetID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkProtobufVsJSON_RoundTrip benchmarks complete round-trip operations
func BenchmarkProtobufVsJSON_RoundTrip(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()
	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("JSON_RoundTrip_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Serialize
				jsonData, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				var result PlayerDataWithCurrency
				err = json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Protobuf_RoundTrip_%s", tc.name), func(b *testing.B) {
			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Serialize
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// createBenchmarkPlayers creates test players for benchmarking
func createBenchmarkPlayers(count int) []Player {
	players := make([]Player, count)

	for i := 0; i < count; i++ {
		players[i] = Player{
			UID:                 int64(i + 1),
			Name:                fmt.Sprintf("Benchmark Player %d", i+1),
			Position:            []string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8],
			Age:                 fmt.Sprintf("%d", 18+(i%15)),
			Club:                fmt.Sprintf("Benchmark Club %d", (i%50)+1),
			Division:            []string{"Premier League", "Championship", "League One", "League Two"}[i%4],
			TransferValue:       fmt.Sprintf("£%dM", (i%50)+1),
			Wage:                fmt.Sprintf("£%dK", (i%100)+10),
			Personality:         []string{"Determined", "Ambitious", "Professional", "Casual"}[i%4],
			MediaHandling:       []string{"Confident", "Evasive", "Outspoken", "Reserved"}[i%4],
			Nationality:         []string{"England", "Spain", "Germany", "France", "Italy", "Brazil", "Argentina"}[i%7],
			NationalityISO:      []string{"ENG", "ESP", "GER", "FRA", "ITA", "BRA", "ARG"}[i%7],
			NationalityFIFACode: []string{"ENG", "ESP", "GER", "FRA", "ITA", "BRA", "ARG"}[i%7],
			AttributeMasked:     i%2 == 0,
			Attributes: map[string]string{
				"Pace":      fmt.Sprintf("%d", 10+(i%11)),
				"Shooting":  fmt.Sprintf("%d", 8+(i%13)),
				"Passing":   fmt.Sprintf("%d", 12+(i%9)),
				"Dribbling": fmt.Sprintf("%d", 9+(i%12)),
				"Defending": fmt.Sprintf("%d", 7+(i%14)),
				"Physical":  fmt.Sprintf("%d", 11+(i%10)),
				"Crossing":  fmt.Sprintf("%d", 6+(i%15)),
				"Finishing": fmt.Sprintf("%d", 13+(i%8)),
				"Heading":   fmt.Sprintf("%d", 14+(i%7)),
				"Technique": fmt.Sprintf("%d", 15+(i%6)),
			},
			NumericAttributes: map[string]int{
				"Pace":      10 + (i % 11),
				"Shooting":  8 + (i % 13),
				"Passing":   12 + (i % 9),
				"Dribbling": 9 + (i % 12),
				"Defending": 7 + (i % 14),
				"Physical":  11 + (i % 10),
				"Crossing":  6 + (i % 15),
				"Finishing": 13 + (i % 8),
				"Heading":   14 + (i % 7),
				"Technique": 15 + (i % 6),
			},
			PerformanceStatsNumeric: map[string]float64{
				"Goals":         float64(i%30) + 0.5,
				"Assists":       float64(i%20) + 0.3,
				"KeyPasses":     float64(i%10) + 1.2,
				"Tackles":       float64(i%15) + 0.8,
				"Interceptions": float64(i%12) + 0.4,
				"Saves":         float64(i%25) + 2.1,
				"CleanSheets":   float64(i%8) + 0.1,
			},
			PerformancePercentiles: map[string]map[string]float64{
				"Attacking": {
					"Goals":     float64((i*7)%100) + 0.1,
					"Assists":   float64((i*11)%100) + 0.2,
					"KeyPasses": float64((i*13)%100) + 0.3,
					"Shots":     float64((i*17)%100) + 0.4,
				},
				"Defending": {
					"Tackles":       float64((i*19)%100) + 0.5,
					"Interceptions": float64((i*23)%100) + 0.6,
					"Clearances":    float64((i*29)%100) + 0.7,
					"Blocks":        float64((i*31)%100) + 0.8,
				},
				"Passing": {
					"PassAccuracy":  float64((i*37)%100) + 0.9,
					"LongPasses":    float64((i*41)%100) + 0.1,
					"ThroughBalls":  float64((i*43)%100) + 0.2,
					"CrossAccuracy": float64((i*47)%100) + 0.3,
				},
				"Goalkeeping": {
					"Saves":        float64((i*53)%100) + 0.4,
					"CleanSheets":  float64((i*59)%100) + 0.5,
					"Distribution": float64((i*61)%100) + 0.6,
				},
			},
			ParsedPositions: []string{[]string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8]},
			ShortPositions:  []string{[]string{"ST", "CM", "CB", "GK", "RW", "LW", "CAM", "CDM"}[i%8]},
			PositionGroups:  []string{[]string{"Forward", "Midfielder", "Defender", "Goalkeeper"}[i%4]},
			PAC:             10 + (i % 11),
			SHO:             8 + (i % 13),
			PAS:             12 + (i % 9),
			DRI:             9 + (i % 12),
			DEF:             7 + (i % 14),
			PHY:             11 + (i % 10),
			GK:              (i % 20) + 1,
			DIV:             (i % 18) + 2,
			HAN:             (i % 16) + 3,
			REF:             (i % 14) + 4,
			KIC:             (i % 12) + 5,
			SPD:             (i % 10) + 6,
			POS:             (i % 8) + 7,
			Overall:         50 + (i % 40),
			BestRoleOverall: fmt.Sprintf("Best Role %d", i%10),
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: fmt.Sprintf("Primary Role %d", i%5), Score: 60 + (i % 30)},
				{RoleName: fmt.Sprintf("Secondary Role %d", i%5), Score: 55 + (i % 35)},
				{RoleName: fmt.Sprintf("Tertiary Role %d", i%3), Score: 50 + (i % 40)},
			},
			TransferValueAmount: int64((i%50 + 1) * 1000000),
			WageAmount:          int64((i%100 + 10) * 1000),
		}
	}

	return players
}

// BenchmarkProtobufVsJSON_CompressionEfficiency benchmarks compression efficiency
func BenchmarkProtobufVsJSON_CompressionEfficiency(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()

	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
		{"XLarge_5000_Players", 5000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		// Serialize with JSON
		jsonData, err := json.Marshal(dataset)
		if err != nil {
			b.Fatal(err)
		}

		// Serialize with Protobuf
		ctx := context.Background()
		protoDataset, err := dataset.ToProto(ctx)
		if err != nil {
			b.Fatal(err)
		}
		protoData, err := proto.Marshal(protoDataset)
		if err != nil {
			b.Fatal(err)
		}

		// Test compression with gzip
		b.Run(fmt.Sprintf("JSON_Compression_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := compressData(jsonData)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(len(jsonData)), "uncompressed_bytes")
		})

		b.Run(fmt.Sprintf("Protobuf_Compression_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := compressData(protoData)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(len(protoData)), "uncompressed_bytes")
		})

		// Test decompression
		compressedJSON, _ := compressData(jsonData)
		compressedProto, _ := compressData(protoData)

		b.Run(fmt.Sprintf("JSON_Decompression_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decompressData(compressedJSON)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(len(compressedJSON)), "compressed_bytes")
		})

		b.Run(fmt.Sprintf("Protobuf_Decompression_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decompressData(compressedProto)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(len(compressedProto)), "compressed_bytes")
		})
	}
}

// BenchmarkProtobufVsJSON_DetailedMemoryProfile provides detailed memory profiling
func BenchmarkProtobufVsJSON_DetailedMemoryProfile(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()

	testCases := []struct {
		name        string
		playerCount int
	}{
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("JSON_DetailedMemory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats

			// Force garbage collection and get baseline
			runtime.GC()
			runtime.GC() // Call twice to ensure clean state
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Serialize
				jsonData, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				var result PlayerDataWithCurrency
				err = json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}

				// Force intermediate cleanup to measure peak usage
				if i%100 == 0 {
					runtime.GC()
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)

			// Report detailed memory metrics
			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
			b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs_per_op")
			b.ReportMetric(float64(m2.Frees-m1.Frees)/float64(b.N), "frees_per_op")
			b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap_bytes_per_op")
		})

		b.Run(fmt.Sprintf("Protobuf_DetailedMemory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats

			// Force garbage collection and get baseline
			runtime.GC()
			runtime.GC() // Call twice to ensure clean state
			runtime.ReadMemStats(&m1)

			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Serialize
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}

				// Force intermediate cleanup to measure peak usage
				if i%100 == 0 {
					runtime.GC()
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)

			// Report detailed memory metrics
			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
			b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs_per_op")
			b.ReportMetric(float64(m2.Frees-m1.Frees)/float64(b.N), "frees_per_op")
			b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap_bytes_per_op")
		})
	}
}

// BenchmarkProtobufVsJSON_ConcurrentOperations benchmarks concurrent serialization/deserialization
func BenchmarkProtobufVsJSON_ConcurrentOperations(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()

	// Create test dataset
	players := createBenchmarkPlayers(100)
	dataset := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	b.Run("JSON_Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Serialize
				jsonData, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				var result PlayerDataWithCurrency
				err = json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Protobuf_Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			ctx := context.Background()
			for pb.Next() {
				// Serialize
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Deserialize
				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// BenchmarkProtobufVsJSON_LargeDatasetScaling tests performance scaling with very large datasets
func BenchmarkProtobufVsJSON_LargeDatasetScaling(b *testing.B) {
	ctx := context.Background()
	// Set log level to INFO to reduce benchmark noise
	restoreLogging := setLogLevelForBenchmarks(ctx)
	defer restoreLogging()

	testCases := []struct {
		name        string
		playerCount int
	}{
		{"XSmall_1_Player", 1},
		{"Small_10_Players", 10},
		{"Medium_100_Players", 100},
		{"Large_1000_Players", 1000},
		{"XLarge_5000_Players", 5000},
		{"XXLarge_10000_Players", 10000},
	}

	for _, tc := range testCases {
		// Create test dataset
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("JSON_Scaling_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				jsonData, err := json.Marshal(dataset)
				if err != nil {
					b.Fatal(err)
				}

				var result PlayerDataWithCurrency
				err = json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(tc.playerCount), "player_count")
		})

		b.Run(fmt.Sprintf("Protobuf_Scaling_%s", tc.name), func(b *testing.B) {
			ctx := context.Background()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(tc.playerCount), "player_count")
		})
	}
}
