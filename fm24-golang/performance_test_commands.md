# Performance Testing Commands

# 1. Run existing benchmarks
go test -bench=BenchmarkParseMonetaryValueGo -benchmem ./src/api/

# 2. Test parsing performance with a sample file
go test -run=TestParsing -v ./src/api/

# 3. Run memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./src/api/

# 4. Run CPU profiling  
go test -bench=. -cpuprofile=cpu.prof ./src/api/

# 5. Compare before/after with pprof
go tool pprof mem.prof
go tool pprof cpu.prof
