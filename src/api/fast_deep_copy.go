package main

import (
	"sync"
)

// FastDeepCopyPlayers creates an optimized deep copy that's safe but much faster than deepCopyPlayers
// This avoids the COW race conditions while being significantly more efficient
func FastDeepCopyPlayers(players []Player) []Player {
	if len(players) == 0 {
		return nil
	}

	// Pre-allocate result slice
	result := make([]Player, len(players))

	// Use worker goroutines for parallel copying of large datasets
	if len(players) > 1000 {
		return fastDeepCopyParallel(players)
	}

	// For smaller datasets, use optimized sequential copy
	for i := range players {
		result[i] = fastDeepCopyPlayer(&players[i])
	}

	return result
}

// fastDeepCopyParallel uses goroutines to copy players in parallel for large datasets
func fastDeepCopyParallel(players []Player) []Player {
	result := make([]Player, len(players))

	// Calculate optimal chunk size based on CPU cores and data size
	numWorkers := Min(8, (len(players)/500)+1) // Adaptive worker count
	chunkSize := (len(players) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := Min(start+chunkSize, len(players))

		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				result[j] = fastDeepCopyPlayer(&players[j])
			}
		}(start, end)
	}

	wg.Wait()
	return result
}

// fastDeepCopyPlayer creates an optimized deep copy of a single player
func fastDeepCopyPlayer(original *Player) Player {
	player := Player{
		// Copy all primitive fields efficiently
		UID:                 original.UID,
		Name:                original.Name,
		Position:            original.Position,
		Age:                 original.Age,
		Club:                original.Club,
		Division:            original.Division,
		TransferValue:       original.TransferValue,
		TransferValueAmount: original.TransferValueAmount,
		Wage:                original.Wage,
		WageAmount:          original.WageAmount,
		Personality:         original.Personality,
		MediaHandling:       original.MediaHandling,
		Nationality:         original.Nationality,
		NationalityISO:      original.NationalityISO,
		NationalityFIFACode: original.NationalityFIFACode,
		AttributeMasked:     original.AttributeMasked,
		PAC:                 original.PAC,
		SHO:                 original.SHO,
		PAS:                 original.PAS,
		DRI:                 original.DRI,
		DEF:                 original.DEF,
		PHY:                 original.PHY,
		GK:                  original.GK,
		DIV:                 original.DIV,
		HAN:                 original.HAN,
		REF:                 original.REF,
		KIC:                 original.KIC,
		SPD:                 original.SPD,
		POS:                 original.POS,
		Overall:             original.Overall,
		TotalStats:          original.TotalStats,
		MBR:                 original.MBR,
		Mbr:                 original.Mbr,
		BasedIn:             original.BasedIn,

		// Initialize new mutex
		mu: sync.RWMutex{},
	}

	// Efficiently copy maps only if they exist
	if original.NumericAttributes != nil {
		player.NumericAttributes = make(map[string]int, len(original.NumericAttributes))
		for k, v := range original.NumericAttributes {
			player.NumericAttributes[k] = v
		}
	}

	if original.Attributes != nil {
		player.Attributes = make(map[string]string, len(original.Attributes))
		for k, v := range original.Attributes {
			player.Attributes[k] = v
		}
	}

	if original.PerformanceStatsNumeric != nil {
		player.PerformanceStatsNumeric = make(map[string]float64, len(original.PerformanceStatsNumeric))
		for k, v := range original.PerformanceStatsNumeric {
			player.PerformanceStatsNumeric[k] = v
		}
	}

	// Copy nested maps efficiently
	if original.PerformancePercentiles != nil {
		player.PerformancePercentiles = make(map[string]map[string]float64, len(original.PerformancePercentiles))
		for group, stats := range original.PerformancePercentiles {
			player.PerformancePercentiles[group] = make(map[string]float64, len(stats))
			for stat, value := range stats {
				player.PerformancePercentiles[group][stat] = value
			}
		}
	}

	// Efficiently copy slices
	if original.ParsedPositions != nil {
		player.ParsedPositions = make([]string, len(original.ParsedPositions))
		copy(player.ParsedPositions, original.ParsedPositions)
	}

	if original.RoleSpecificOveralls != nil {
		player.RoleSpecificOveralls = make([]RoleOverallScore, len(original.RoleSpecificOveralls))
		copy(player.RoleSpecificOveralls, original.RoleSpecificOveralls)
	}

	return player
}
