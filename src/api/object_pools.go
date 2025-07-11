package main

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PlayerPool manages a pool of Player structs to reduce allocations
type PlayerPool struct {
	pool  sync.Pool
	stats struct {
		gets    int64
		puts    int64
		hits    int64
		misses  int64
		created int64
	}
}

// Global player pool instance
var globalPlayerPool = &PlayerPool{}

// Initialize the player pool
func init() {
	globalPlayerPool.pool = sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&globalPlayerPool.stats.created, 1)
			return &Player{
				Attributes:              make(map[string]string, defaultAttributeCapacity),
				NumericAttributes:       make(map[string]int, defaultAttributeCapacity),
				PerformanceStatsNumeric: make(map[string]float64, 50),
				PerformancePercentiles:  make(map[string]map[string]float64),
				ParsedPositions:         make([]string, 0, 4),
				ShortPositions:          make([]string, 0, 4),
				PositionGroups:          make([]string, 0, 2),
				RoleSpecificOveralls:    make([]RoleOverallScore, 0, 10),
			}
		},
	}
}

// GetPlayer returns a Player from the pool or creates a new one
func GetPlayer() *Player {
	atomic.AddInt64(&globalPlayerPool.stats.gets, 1)

	player := globalPlayerPool.pool.Get().(*Player)
	if player == nil {
		atomic.AddInt64(&globalPlayerPool.stats.misses, 1)
		return globalPlayerPool.pool.New().(*Player)
	}

	atomic.AddInt64(&globalPlayerPool.stats.hits, 1)
	return player
}

// ReturnPlayer returns a Player to the pool after resetting it
func ReturnPlayer(player *Player) {
	if player == nil {
		return
	}

	atomic.AddInt64(&globalPlayerPool.stats.puts, 1)

	// Reset the player to clean state
	resetPlayer(player)

	globalPlayerPool.pool.Put(player)
}

// resetPlayer clears all data from a Player struct for reuse
func resetPlayer(player *Player) {
	// Reset string fields
	player.UID = 0
	player.Name = ""
	player.Position = ""
	player.Age = ""
	player.Club = ""
	player.Division = ""
	player.TransferValue = ""
	player.Wage = ""
	player.Personality = ""
	player.MediaHandling = ""
	player.Nationality = ""
	player.NationalityISO = ""
	player.NationalityFIFACode = ""
	player.BestRoleOverall = ""

	// Reset numeric fields
	player.Overall = 0
	player.PAC = 0
	player.SHO = 0
	player.PAS = 0
	player.DRI = 0
	player.DEF = 0
	player.PHY = 0
	player.GK = 0
	player.DIV = 0
	player.HAN = 0
	player.REF = 0
	player.KIC = 0
	player.SPD = 0
	player.POS = 0
	player.TransferValueAmount = 0
	player.WageAmount = 0

	// Reset boolean fields
	player.AttributeMasked = false

	// Clear maps but keep capacity
	for k := range player.Attributes {
		delete(player.Attributes, k)
	}
	for k := range player.NumericAttributes {
		delete(player.NumericAttributes, k)
	}
	for k := range player.PerformanceStatsNumeric {
		delete(player.PerformanceStatsNumeric, k)
	}
	for k := range player.PerformancePercentiles {
		delete(player.PerformancePercentiles, k)
	}

	// Reset slices but keep capacity
	player.ParsedPositions = player.ParsedPositions[:0]
	player.ShortPositions = player.ShortPositions[:0]
	player.PositionGroups = player.PositionGroups[:0]
	player.RoleSpecificOveralls = player.RoleSpecificOveralls[:0]
}

// GetPlayerPoolStats returns statistics about the player pool
func GetPlayerPoolStats() map[string]int64 {
	return map[string]int64{
		"gets":    atomic.LoadInt64(&globalPlayerPool.stats.gets),
		"puts":    atomic.LoadInt64(&globalPlayerPool.stats.puts),
		"hits":    atomic.LoadInt64(&globalPlayerPool.stats.hits),
		"misses":  atomic.LoadInt64(&globalPlayerPool.stats.misses),
		"created": atomic.LoadInt64(&globalPlayerPool.stats.created),
		"hit_rate": func() int64 {
			gets := atomic.LoadInt64(&globalPlayerPool.stats.gets)
			if gets == 0 {
				return 0
			}
			hits := atomic.LoadInt64(&globalPlayerPool.stats.hits)
			return (hits * 100) / gets
		}(),
	}
}

// SlicePool manages pools of various slice types to reduce allocations
type SlicePool struct {
	stringSlicePool sync.Pool
	playerSlicePool sync.Pool
	intSlicePool    sync.Pool
}

var globalSlicePool = &SlicePool{}

func init() {
	globalSlicePool.stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 10)
		},
	}

	globalSlicePool.playerSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]Player, 0, 100)
		},
	}

	globalSlicePool.intSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]int, 0, 10)
		},
	}
}

// GetStringSlice returns a string slice from the pool
func GetStringSlice() []string {
	return globalSlicePool.stringSlicePool.Get().([]string)
}

// ReturnStringSlice returns a string slice to the pool
func ReturnStringSlice(slice []string) {
	if slice == nil {
		return
	}
	slice = slice[:0] // Reset length but keep capacity
	//nolint:staticcheck // SA6002: Slices should be stored as values, not pointers in sync.Pool
	globalSlicePool.stringSlicePool.Put(slice)
}

// GetPlayerSlice returns a player slice from the pool
func GetPlayerSlice() []Player {
	return globalSlicePool.playerSlicePool.Get().([]Player)
}

// ReturnPlayerSlice returns a player slice to the pool
func ReturnPlayerSlice(slice []Player) {
	if slice == nil {
		return
	}
	slice = slice[:0] // Reset length but keep capacity
	//nolint:staticcheck // SA6002: Slices should be stored as values, not pointers in sync.Pool
	globalSlicePool.playerSlicePool.Put(slice)
}

// GetIntSlice returns an int slice from the pool
func GetIntSlice() []int {
	return globalSlicePool.intSlicePool.Get().([]int)
}

// ReturnIntSlice returns an int slice to the pool
func ReturnIntSlice(slice []int) {
	if slice == nil {
		return
	}
	slice = slice[:0] // Reset length but keep capacity
	//nolint:staticcheck // SA6002: Slices should be stored as values, not pointers in sync.Pool
	globalSlicePool.intSlicePool.Put(slice)
}

// MapPool manages pools of map types to reduce allocations
type MapPool struct {
	stringMapPool     sync.Pool
	intMapPool        sync.Pool
	float64MapPool    sync.Pool
	nestedFloat64Pool sync.Pool
}

var globalMapPool = &MapPool{}

func init() {
	globalMapPool.stringMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, defaultAttributeCapacity)
		},
	}

	globalMapPool.intMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]int, defaultAttributeCapacity)
		},
	}

	globalMapPool.float64MapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]float64, 50)
		},
	}

	globalMapPool.nestedFloat64Pool = sync.Pool{
		New: func() interface{} {
			return make(map[string]map[string]float64)
		},
	}
}

// GetStringMap returns a string map from the pool
func GetStringMap() map[string]string {
	return globalMapPool.stringMapPool.Get().(map[string]string)
}

// ReturnStringMap returns a string map to the pool
func ReturnStringMap(m map[string]string) {
	if m == nil {
		return
	}
	for k := range m {
		delete(m, k)
	}
	globalMapPool.stringMapPool.Put(m)
}

// GetIntMap returns an int map from the pool
func GetIntMap() map[string]int {
	return globalMapPool.intMapPool.Get().(map[string]int)
}

// ReturnIntMap returns an int map to the pool
func ReturnIntMap(m map[string]int) {
	if m == nil {
		return
	}
	for k := range m {
		delete(m, k)
	}
	globalMapPool.intMapPool.Put(m)
}

// GetFloat64Map returns a float64 map from the pool
func GetFloat64Map() map[string]float64 {
	return globalMapPool.float64MapPool.Get().(map[string]float64)
}

// ReturnFloat64Map returns a float64 map to the pool
func ReturnFloat64Map(m map[string]float64) {
	if m == nil {
		return
	}
	for k := range m {
		delete(m, k)
	}
	globalMapPool.float64MapPool.Put(m)
}

// GetNestedFloat64Map returns a nested float64 map from the pool
func GetNestedFloat64Map() map[string]map[string]float64 {
	return globalMapPool.nestedFloat64Pool.Get().(map[string]map[string]float64)
}

// ReturnNestedFloat64Map returns a nested float64 map to the pool
func ReturnNestedFloat64Map(m map[string]map[string]float64) {
	if m == nil {
		return
	}
	for k := range m {
		delete(m, k)
	}
	globalMapPool.nestedFloat64Pool.Put(m)
}

// PeriodicPoolMaintenance performs periodic maintenance on object pools
func PeriodicPoolMaintenance() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Force GC to clean up any unused pool objects
		runtime.GC()

		stats := GetPlayerPoolStats()
		if stats["gets"] > 0 {
			log.Printf("Player pool stats: %d gets, %d hits (%d%% hit rate), %d created",
				stats["gets"], stats["hits"], stats["hit_rate"], stats["created"])
		}
	}
}

// InitializeObjectPools starts the object pooling system
func InitializeObjectPools() {
	log.Println("Initializing enhanced object pooling system...")

	// Start periodic maintenance
	go PeriodicPoolMaintenance()

	log.Println("Object pooling system initialized with Player, Slice, and Map pools")
}
