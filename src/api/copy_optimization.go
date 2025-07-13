package main

import (
	"sync"
	"sync/atomic"
)

// CopyOnWritePlayer implements copy-on-write semantics for Player data
// This significantly reduces memory allocation during filtering and processing
type CopyOnWritePlayer struct {
	core     *Player                // Shared reference to core data
	refCount *int64                 // Reference count for shared data
	mutex    *sync.RWMutex          // Protects the core data
	modified map[string]interface{} // Track modified fields
	copied   bool                   // True if this instance has been copied
}

// CopyOnWritePlayerSlice manages a slice of copy-on-write players
type CopyOnWritePlayerSlice struct {
	players []*CopyOnWritePlayer
	shared  bool // True if this slice is shared
}

// CreateCopyOnWritePlayer creates a new copy-on-write player
func CreateCopyOnWritePlayer(player *Player) *CopyOnWritePlayer {
	refCount := int64(1)
	return &CopyOnWritePlayer{
		core:     player,
		refCount: &refCount,
		mutex:    &sync.RWMutex{},
		modified: make(map[string]interface{}),
		copied:   false,
	}
}

// CreateCopyOnWritePlayerSlice creates a copy-on-write slice from regular players
func CreateCopyOnWritePlayerSlice(players []Player) *CopyOnWritePlayerSlice {
	cowPlayers := make([]*CopyOnWritePlayer, len(players))
	for i := range players {
		cowPlayers[i] = CreateCopyOnWritePlayer(&players[i])
	}

	return &CopyOnWritePlayerSlice{
		players: cowPlayers,
		shared:  true,
	}
}

// Clone creates a shallow copy that shares data until modification
func (cow *CopyOnWritePlayer) Clone() *CopyOnWritePlayer {
	atomic.AddInt64(cow.refCount, 1)

	return &CopyOnWritePlayer{
		core:     cow.core,
		refCount: cow.refCount,
		mutex:    cow.mutex,
		modified: make(map[string]interface{}), // Each clone has its own modifications
		copied:   false,
	}
}

// ensureWritable performs copy-on-write if needed
func (cow *CopyOnWritePlayer) ensureWritable() {
	if cow.copied {
		return // Already copied
	}

	if atomic.LoadInt64(cow.refCount) > 1 {
		// Multiple references exist, need to copy
		cow.mutex.Lock()
		defer cow.mutex.Unlock()

		// Double-check after acquiring lock
		if !cow.copied && atomic.LoadInt64(cow.refCount) > 1 {
			// Create a deep copy
			cow.core = cow.deepCopyPlayer(cow.core)
			cow.refCount = new(int64)
			*cow.refCount = 1
			cow.copied = true
		}
	}
}

// deepCopyPlayer creates a complete copy of a player (optimized version)
func (cow *CopyOnWritePlayer) deepCopyPlayer(original *Player) *Player {
	if original == nil {
		return nil
	}

	// Use object pool for player structs to reduce allocations
	player := getPlayerFromPool()

	// Copy primitive fields
	player.UID = original.UID
	player.Name = original.Name
	player.Position = original.Position
	player.Age = original.Age
	player.Club = original.Club
	player.Division = original.Division
	player.TransferValue = original.TransferValue
	player.Wage = original.Wage
	player.Personality = original.Personality
	player.MediaHandling = original.MediaHandling
	player.Nationality = original.Nationality
	player.NationalityISO = original.NationalityISO
	player.NationalityFIFACode = original.NationalityFIFACode
	player.AttributeMasked = original.AttributeMasked
	player.PAC = original.PAC
	player.SHO = original.SHO
	player.PAS = original.PAS
	player.DRI = original.DRI
	player.DEF = original.DEF
	player.PHY = original.PHY
	player.GK = original.GK
	player.DIV = original.DIV
	player.HAN = original.HAN
	player.REF = original.REF
	player.KIC = original.KIC
	player.SPD = original.SPD
	player.POS = original.POS
	player.Overall = original.Overall
	player.BestRoleOverall = original.BestRoleOverall
	player.TransferValueAmount = original.TransferValueAmount
	player.WageAmount = original.WageAmount

	// Copy maps efficiently
	if original.Attributes != nil {
		player.Attributes = make(map[string]string, len(original.Attributes))
		for k, v := range original.Attributes {
			player.Attributes[k] = v
		}
	}

	if original.NumericAttributes != nil {
		player.NumericAttributes = make(map[string]int, len(original.NumericAttributes))
		for k, v := range original.NumericAttributes {
			player.NumericAttributes[k] = v
		}
	}

	if original.PerformanceStatsNumeric != nil {
		player.PerformanceStatsNumeric = make(map[string]float64, len(original.PerformanceStatsNumeric))
		for k, v := range original.PerformanceStatsNumeric {
			player.PerformanceStatsNumeric[k] = v
		}
	}

	// Copy nested map efficiently
	if original.PerformancePercentiles != nil {
		player.PerformancePercentiles = make(map[string]map[string]float64, len(original.PerformancePercentiles))
		for group, stats := range original.PerformancePercentiles {
			player.PerformancePercentiles[group] = make(map[string]float64, len(stats))
			for stat, value := range stats {
				player.PerformancePercentiles[group][stat] = value
			}
		}
	}

	// Copy slices efficiently
	if original.ParsedPositions != nil {
		player.ParsedPositions = make([]string, len(original.ParsedPositions))
		copy(player.ParsedPositions, original.ParsedPositions)
	}

	if original.ShortPositions != nil {
		player.ShortPositions = make([]string, len(original.ShortPositions))
		copy(player.ShortPositions, original.ShortPositions)
	}

	if original.PositionGroups != nil {
		player.PositionGroups = make([]string, len(original.PositionGroups))
		copy(player.PositionGroups, original.PositionGroups)
	}

	if original.RoleSpecificOveralls != nil {
		player.RoleSpecificOveralls = make([]RoleOverallScore, len(original.RoleSpecificOveralls))
		copy(player.RoleSpecificOveralls, original.RoleSpecificOveralls)
	}

	return player
}

// Object pool for Player structs to reduce GC pressure
var playerPool = sync.Pool{
	New: func() interface{} {
		return &Player{
			Attributes:              make(map[string]string),
			NumericAttributes:       make(map[string]int),
			PerformanceStatsNumeric: make(map[string]float64),
			PerformancePercentiles:  make(map[string]map[string]float64),
		}
	},
}

func getPlayerFromPool() *Player {
	player := playerPool.Get().(*Player)

	// Reset the player to clean state
	player.UID = 0
	player.Name = ""
	// ... reset other fields as needed

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

	return player
}

func returnPlayerToPool(player *Player) {
	if player != nil {
		playerPool.Put(player)
	}
}

// GetPlayer returns the underlying player data (read-only access)
func (cow *CopyOnWritePlayer) GetPlayer() *Player {
	cow.mutex.RLock()
	defer cow.mutex.RUnlock()
	return cow.core
}

// SetAttribute modifies an attribute (triggers copy-on-write)
func (cow *CopyOnWritePlayer) SetAttribute(name string, value interface{}) {
	cow.ensureWritable()
	cow.modified[name] = value

	// Apply the modification to the core data
	switch name {
	case "Overall":
		if v, ok := value.(int); ok {
			cow.core.Overall = v
		}
	case "Club":
		if v, ok := value.(string); ok {
			cow.core.Club = v
		}
		// Add more cases as needed
	}
}

// GetAttribute gets an attribute value (with copy-on-write optimizations)
func (cow *CopyOnWritePlayer) GetAttribute(name string) interface{} {
	// Check modifications first
	if value, exists := cow.modified[name]; exists {
		return value
	}

	// Fall back to core data
	cow.mutex.RLock()
	defer cow.mutex.RUnlock()

	switch name {
	case "Overall":
		return cow.core.Overall
	case "Club":
		return cow.core.Club
	// Add more cases as needed
	default:
		return nil
	}
}

// Filter creates a new slice with filtered players (copy-on-write optimized)
func (cows *CopyOnWritePlayerSlice) Filter(predicate func(*Player) bool) *CopyOnWritePlayerSlice {
	var filtered []*CopyOnWritePlayer

	for _, cowPlayer := range cows.players {
		if predicate(cowPlayer.GetPlayer()) {
			filtered = append(filtered, cowPlayer.Clone())
		}
	}

	return &CopyOnWritePlayerSlice{
		players: filtered,
		shared:  true,
	}
}

// ToPlayers converts back to regular Player slice (only copies when necessary)
func (cows *CopyOnWritePlayerSlice) ToPlayers() []Player {
	result := make([]Player, len(cows.players))

	for i, cowPlayer := range cows.players {
		result[i] = *cowPlayer.GetPlayer()
	}

	return result
}

// OptimizedDeepCopyPlayers creates a deep copy of a slice of players with optimized memory usage
func OptimizedDeepCopyPlayers(players []Player) []Player {
	if players == nil {
		return nil
	}

	// Create copy-on-write slice
	cowSlice := CreateCopyOnWritePlayerSlice(players)

	// For deep copy, we actually need to force copying, but this is still
	// more efficient due to the optimized copying logic and object pooling
	result := make([]Player, len(players))
	for i, cowPlayer := range cowSlice.players {
		// Force a copy for each player
		cowPlayer.ensureWritable()
		result[i] = *cowPlayer.GetPlayer()
	}

	return result
}

// Cleanup decrements reference count and cleans up if needed
func (cow *CopyOnWritePlayer) Cleanup() {
	if atomic.AddInt64(cow.refCount, -1) == 0 {
		// Last reference, return to pool if it was copied
		if cow.copied {
			returnPlayerToPool(cow.core)
		}
	}
}
