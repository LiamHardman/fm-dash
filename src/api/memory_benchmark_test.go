package main

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

// BenchmarkPlayerCreation benchmarks player creation with and without object pooling
func BenchmarkPlayerCreation(b *testing.B) {
	b.Run("WithObjectPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			player := GetPlayer()
			// Simulate some work
			player.Name = "Test Player"
			player.Position = "ST"
			player.Club = "Test Club"
			ReturnPlayer(player)
		}
	})

	b.Run("WithoutObjectPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			player := &Player{
				Attributes:              make(map[string]string, defaultAttributeCapacity),
				NumericAttributes:       make(map[string]int, defaultAttributeCapacity),
				PerformanceStatsNumeric: make(map[string]float64, 50),
				PerformancePercentiles:  make(map[string]map[string]float64),
				ParsedPositions:         make([]string, 0, 4),
				ShortPositions:          make([]string, 0, 4),
				PositionGroups:          make([]string, 0, 2),
				RoleSpecificOveralls:    make([]RoleOverallScore, 0, 10),
			}
			// Simulate some work
			player.Name = "Test Player"
			player.Position = "ST"
			player.Club = "Test Club"
		}
	})
}

// BenchmarkStringInterning benchmarks string interning performance
func BenchmarkStringInterning(b *testing.B) {
	players := createTestPlayers(1000)

	b.Run("WithStringInterning", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			applyStringInterning(ctx, players)
		}
	})

	b.Run("WithoutStringInterning", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Just copy strings without interning
			for j := range players {
				_ = players[j].Club + ""
				_ = players[j].Position + ""
				_ = players[j].Nationality + ""
			}
		}
	})
}

// BenchmarkPlayerDeepCopy benchmarks deep copy operations
func BenchmarkPlayerDeepCopy(b *testing.B) {
	players := createTestPlayers(1000)

	b.Run("OptimizedDeepCopy", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = OptimizedDeepCopyPlayers(players)
		}
	})

	b.Run("RegularDeepCopy", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = deepCopyPlayers(players)
		}
	})
}

// BenchmarkMemoryOptimizations benchmarks the full optimization pipeline
func BenchmarkMemoryOptimizations(b *testing.B) {
	players := createTestPlayers(5000)

	b.Run("FullOptimizationPipeline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := OptimizePlayerData(ctx, players)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NoOptimizations", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Just deep copy without optimizations
			_ = deepCopyPlayers(players)
		}
	})
}

// BenchmarkCacheOperations benchmarks cache operations
func BenchmarkCacheOperations(b *testing.B) {
	InitInMemoryCache()

	b.Run("CacheSet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "test_key_" + string(rune(i))
			value := createTestPlayers(100)
			setInMemCache(key, value, 5*time.Minute)
		}
	})

	b.Run("CacheGet", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := "test_key_" + string(rune(i))
			value := createTestPlayers(100)
			setInMemCache(key, value, 5*time.Minute)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "test_key_" + string(rune(i%1000))
			_, _ = getFromMemCache(key)
		}
	})
}

// BenchmarkStructSizes benchmarks memory usage of different struct types
func BenchmarkStructSizes(b *testing.B) {
	b.Run("RegularPlayer", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		players := make([]Player, b.N)
		for i := range players {
			players[i] = createTestPlayer()
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/player")
	})

	// Note: OptimizedPlayer benchmark would require implementing the conversion
	// b.Run("OptimizedPlayer", func(b *testing.B) {
	//     // Similar to above but with OptimizedPlayer struct
	// })
}

// BenchmarkProtobufSerialization benchmarks protobuf vs JSON serialization
func BenchmarkProtobufSerialization(b *testing.B) {
	players := createTestPlayers(1000)
	ctx := context.Background()

	b.Run("ProtobufSerialization", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, player := range players {
				_, err := player.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("JSONSerialization", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, player := range players {
				_, err := json.Marshal(player)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// BenchmarkMemoryPressure benchmarks performance under memory pressure
func BenchmarkMemoryPressure(b *testing.B) {
	// Create memory pressure by allocating large datasets
	largePlayers := createTestPlayers(50000)

	b.Run("UnderMemoryPressure", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := OptimizePlayerData(ctx, largePlayers)
			if err != nil {
				b.Fatal(err)
			}
			// Force GC to simulate memory pressure
			runtime.GC()
		}
	})

	b.Run("NormalMemoryConditions", func(b *testing.B) {
		smallPlayers := createTestPlayers(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := OptimizePlayerData(ctx, smallPlayers)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper functions for benchmarks

func createTestPlayers(count int) []Player {
	players := make([]Player, count)
	for i := range players {
		players[i] = createTestPlayer()
	}
	return players
}

func createTestPlayer() Player {
	return Player{
		UID:            int64(123456 + len("test")),
		Name:           "Test Player",
		Position:       "ST",
		Age:            "25",
		Club:           "Test FC",
		Division:       "Premier League",
		TransferValue:  "£10M",
		Wage:           "£50K",
		Nationality:    "England",
		NationalityISO: "ENG",
		Overall:        85,
		PAC:            90,
		SHO:            85,
		PAS:            75,
		DRI:            80,
		DEF:            35,
		PHY:            85,
		Attributes: map[string]string{
			"Corners":     "15",
			"Crossing":    "12",
			"Dribbling":   "16",
			"Finishing":   "18",
			"First Touch": "16",
		},
		NumericAttributes: map[string]int{
			"Corners":     15,
			"Crossing":    12,
			"Dribbling":   16,
			"Finishing":   18,
			"First Touch": 16,
		},
		PerformanceStatsNumeric: map[string]float64{
			"Goals":   25.0,
			"Assists": 10.0,
			"Matches": 35.0,
		},
		PerformancePercentiles: map[string]map[string]float64{
			"Attacking": {
				"Goals":   95.0,
				"Assists": 80.0,
			},
		},
		ParsedPositions: []string{"ST", "CF"},
		ShortPositions:  []string{"ST"},
		PositionGroups:  []string{"Attack"},
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Advanced Forward", Score: 87},
			{RoleName: "Complete Forward", Score: 85},
		},
		TransferValueAmount: 10000000,
		WageAmount:          50000,
	}
}

// MemoryBenchmarkResult holds the results of a memory benchmark
type MemoryBenchmarkResult struct {
	TestName         string
	AllocsBefore     uint64
	AllocsAfter      uint64
	AllocDiff        uint64
	MemoryBefore     uint64
	MemoryAfter      uint64
	MemoryDiff       uint64
	Duration         time.Duration
	ObjectsProcessed int
	MemoryPerObject  float64
}

// RunMemoryBenchmark runs a comprehensive memory benchmark
func RunMemoryBenchmark(testName string, objectCount int, testFunc func()) MemoryBenchmarkResult {
	var m1, m2 runtime.MemStats

	// Clean up before measuring
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()
	testFunc()
	duration := time.Since(start)

	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	return MemoryBenchmarkResult{
		TestName:         testName,
		AllocsBefore:     m1.TotalAlloc,
		AllocsAfter:      m2.TotalAlloc,
		AllocDiff:        m2.TotalAlloc - m1.TotalAlloc,
		MemoryBefore:     m1.Alloc,
		MemoryAfter:      m2.Alloc,
		MemoryDiff:       m2.Alloc - m1.Alloc,
		Duration:         duration,
		ObjectsProcessed: objectCount,
		MemoryPerObject:  float64(m2.TotalAlloc-m1.TotalAlloc) / float64(objectCount),
	}
}

// BenchmarkWithMemoryTracking provides detailed memory usage tracking
func BenchmarkWithMemoryTracking(b *testing.B) {
	tests := []struct {
		name        string
		playerCount int
		testFunc    func([]Player)
	}{
		{
			name:        "SmallDataset_1K",
			playerCount: 1000,
			testFunc: func(players []Player) {
				ctx := context.Background()
				OptimizePlayerData(ctx, players)
			},
		},
		{
			name:        "MediumDataset_10K",
			playerCount: 10000,
			testFunc: func(players []Player) {
				ctx := context.Background()
				OptimizePlayerData(ctx, players)
			},
		},
		{
			name:        "LargeDataset_50K",
			playerCount: 50000,
			testFunc: func(players []Player) {
				ctx := context.Background()
				OptimizePlayerData(ctx, players)
			},
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			players := createTestPlayers(test.playerCount)

			result := RunMemoryBenchmark(test.name, test.playerCount, func() {
				for i := 0; i < b.N; i++ {
					test.testFunc(players)
				}
			})

			b.ReportMetric(result.MemoryPerObject, "bytes/player")
			b.ReportMetric(float64(result.Duration.Microseconds()), "μs/op")
			b.ReportMetric(float64(result.AllocDiff), "total-allocs")
		})
	}
}
