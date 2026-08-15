package main

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

// BenchmarkPlayerCreation benchmarks player creation
func BenchmarkPlayerCreation(b *testing.B) {
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

// BenchmarkPlayerDeepCopy benchmarks deep copy operations
func BenchmarkPlayerDeepCopy(b *testing.B) {
	players := createTestPlayers(1000)

	b.Run("RegularDeepCopy", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
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

