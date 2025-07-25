package main

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"google.golang.org/protobuf/proto"
)

// BenchmarkOptimizedProtobufVsOriginal benchmarks the optimized protobuf implementation against the original
func BenchmarkOptimizedProtobufVsOriginal(b *testing.B) {
	ctx := context.Background()

	// Create test data
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

		b.Run(fmt.Sprintf("Original_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use original protobuf conversion
				protoDataset, err := dataset.ToProto(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Convert back
				_, err = DatasetDataFromProto(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Optimized_%s", tc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use optimized protobuf conversion
				protoDataset, err := dataset.ToProtoOptimized(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Convert back
				_, err = DatasetDataFromProtoOptimized(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkOptimizedProtobufMemoryUsage benchmarks memory usage of optimized protobuf
func BenchmarkOptimizedProtobufMemoryUsage(b *testing.B) {
	ctx := context.Background()

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
		players := createBenchmarkPlayers(tc.playerCount)
		dataset := PlayerDataWithCurrency{
			Players:        players,
			CurrencySymbol: "£",
		}

		b.Run(fmt.Sprintf("Original_Memory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats

			runtime.GC()
			runtime.GC()
			runtime.ReadMemStats(&m1)

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

				if i%100 == 0 {
					runtime.GC()
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
			b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs_per_op")
			b.ReportMetric(float64(m2.Frees-m1.Frees)/float64(b.N), "frees_per_op")
			b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap_bytes_per_op")
		})

		b.Run(fmt.Sprintf("Optimized_Memory_%s", tc.name), func(b *testing.B) {
			var m1, m2 runtime.MemStats

			runtime.GC()
			runtime.GC()
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				protoDataset, err := dataset.ToProtoOptimized(ctx)
				if err != nil {
					b.Fatal(err)
				}

				_, err = DatasetDataFromProtoOptimized(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}

				if i%100 == 0 {
					runtime.GC()
				}
			}
			b.StopTimer()

			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
			b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs_per_op")
			b.ReportMetric(float64(m2.Frees-m1.Frees)/float64(b.N), "frees_per_op")
			b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap_bytes_per_op")
		})
	}
}

// BenchmarkOptimizedProtobufConcurrent benchmarks concurrent operations
func BenchmarkOptimizedProtobufConcurrent(b *testing.B) {
	ctx := context.Background()

	players := createBenchmarkPlayers(100)
	dataset := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	// Test concurrent serialization/deserialization
	b.Run("Concurrent_Serialization", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				protoDataset, err := dataset.ToProtoOptimized(ctx)
				if err != nil {
					b.Fatal(err)
				}

				_, err = DatasetDataFromProtoOptimized(ctx, protoDataset)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// BenchmarkOptimizedProtobufObjectPools benchmarks object pool performance
func BenchmarkOptimizedProtobufObjectPools(b *testing.B) {
	ctx := context.Background()

	players := createBenchmarkPlayers(100)
	dataset := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	b.Run("With_Object_Pools", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Use optimized conversion with object pools
			protoDataset, err := dataset.ToProtoOptimized(ctx)
			if err != nil {
				b.Fatal(err)
			}

			_, err = DatasetDataFromProtoOptimized(ctx, protoDataset)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Without_Object_Pools", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Use original conversion without object pools
			protoDataset, err := dataset.ToProto(ctx)
			if err != nil {
				b.Fatal(err)
			}

			_, err = DatasetDataFromProto(ctx, protoDataset)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestOptimizedProtobufCompatibility tests that optimized protobuf is compatible with original
func TestOptimizedProtobufCompatibility(t *testing.T) {
	ctx := context.Background()

	// Create test data
	players := createBenchmarkPlayers(100)
	dataset := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	// Test that optimized and original produce the same results
	originalProto, err := dataset.ToProto(ctx)
	if err != nil {
		t.Fatal(err)
	}

	optimizedProto, err := dataset.ToProtoOptimized(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the protobuf messages
	if !proto.Equal(originalProto, optimizedProto) {
		t.Error("Optimized protobuf conversion produces different results than original")
	}

	// Test round-trip conversion
	originalResult, err := DatasetDataFromProto(ctx, originalProto)
	if err != nil {
		t.Fatal(err)
	}

	optimizedResult, err := DatasetDataFromProtoOptimized(ctx, optimizedProto)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the results
	if len(originalResult.Players) != len(optimizedResult.Players) {
		t.Errorf("Player count mismatch: original=%d, optimized=%d",
			len(originalResult.Players), len(optimizedResult.Players))
	}

	if originalResult.CurrencySymbol != optimizedResult.CurrencySymbol {
		t.Errorf("Currency symbol mismatch: original=%s, optimized=%s",
			originalResult.CurrencySymbol, optimizedResult.CurrencySymbol)
	}
}
