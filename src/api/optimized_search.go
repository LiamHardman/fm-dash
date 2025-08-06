package main

import (
	"strings"
	"sync"
)

// OptimizedPlayerSearch provides parallel search functions for better performance
type OptimizedPlayerSearch struct {
	cache map[string][]Player
	mutex sync.RWMutex
}

// NewOptimizedPlayerSearch creates a new optimized search instance
func NewOptimizedPlayerSearch() *OptimizedPlayerSearch {
	return &OptimizedPlayerSearch{
		cache: make(map[string][]Player),
	}
}

// ParallelSearchPlayers searches players using parallel processing for large datasets
func (ops *OptimizedPlayerSearch) ParallelSearchPlayers(players []Player, query string, searchFields []string) []Player {
	if len(players) == 0 || query == "" {
		return []Player{}
	}

	// Use cache for repeated searches
	cacheKey := query + "|" + strings.Join(searchFields, ",")
	ops.mutex.RLock()
	if cached, exists := ops.cache[cacheKey]; exists {
		ops.mutex.RUnlock()
		return cached
	}
	ops.mutex.RUnlock()

	// For smaller datasets, use sequential search
	if len(players) < 2000 {
		results := sequentialSearch(players, query, searchFields)
		ops.cacheResults(cacheKey, results)
		return results
	}

	// For larger datasets, use parallel search
	results := parallelSearch(players, query, searchFields)
	ops.cacheResults(cacheKey, results)
	return results
}

// parallelSearch performs search using multiple goroutines
func parallelSearch(players []Player, query string, searchFields []string) []Player {
	queryLower := strings.ToLower(query)
	numWorkers := Min(8, len(players)/1000+1)
	chunkSize := (len(players) + numWorkers - 1) / numWorkers

	resultsChan := make(chan []Player, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := Min(start+chunkSize, len(players))

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var matches []Player
			for j := start; j < end; j++ {
				if playerMatchesQuery(&players[j], queryLower, searchFields) {
					matches = append(matches, players[j])
				}
			}

			if len(matches) > 0 {
				resultsChan <- matches
			}
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allResults []Player
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	return allResults
}

// sequentialSearch performs traditional sequential search
func sequentialSearch(players []Player, query string, searchFields []string) []Player {
	queryLower := strings.ToLower(query)
	var results []Player

	for i := range players {
		if playerMatchesQuery(&players[i], queryLower, searchFields) {
			results = append(results, players[i])
		}
	}

	return results
}

// playerMatchesQuery checks if a player matches the search query
func playerMatchesQuery(player *Player, queryLower string, searchFields []string) bool {
	for _, field := range searchFields {
		var fieldValue string

		switch field {
		case "name":
			fieldValue = strings.ToLower(player.Name)
		case "club":
			fieldValue = strings.ToLower(player.Club)
		case "position":
			fieldValue = strings.ToLower(player.Position)
		case "nationality":
			fieldValue = strings.ToLower(player.Nationality)
		case "division":
			fieldValue = strings.ToLower(player.Division)
		default:
			continue
		}

		if strings.Contains(fieldValue, queryLower) {
			return true
		}
	}

	return false
}

// cacheResults stores search results in cache with size limit
func (ops *OptimizedPlayerSearch) cacheResults(key string, results []Player) {
	ops.mutex.Lock()
	defer ops.mutex.Unlock()

	// Prevent cache from growing too large
	if len(ops.cache) > 100 {
		// Simple LRU eviction - clear half the cache
		for k := range ops.cache {
			delete(ops.cache, k)
			if len(ops.cache) <= 50 {
				break
			}
		}
	}

	ops.cache[key] = results
}

// ClearCache clears the search cache
func (ops *OptimizedPlayerSearch) ClearCache() {
	ops.mutex.Lock()
	ops.cache = make(map[string][]Player)
	ops.mutex.Unlock()
}

// ParallelFilterPlayers filters players using parallel processing
func ParallelFilterPlayers(players []Player, filterFunc func(*Player) bool) []Player {
	if len(players) == 0 {
		return []Player{}
	}

	// For smaller datasets, use sequential filtering
	if len(players) < 2000 {
		var results []Player
		for i := range players {
			if filterFunc(&players[i]) {
				results = append(results, players[i])
			}
		}
		return results
	}

	// For larger datasets, use parallel filtering
	numWorkers := Min(8, len(players)/1000+1)
	chunkSize := (len(players) + numWorkers - 1) / numWorkers

	resultsChan := make(chan []Player, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := Min(start+chunkSize, len(players))

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var matches []Player
			for j := start; j < end; j++ {
				if filterFunc(&players[j]) {
					matches = append(matches, players[j])
				}
			}

			if len(matches) > 0 {
				resultsChan <- matches
			}
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allResults []Player
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	return allResults
}

// Global optimized search instance
var globalOptimizedSearch = NewOptimizedPlayerSearch()

// GetOptimizedSearch returns the global optimized search instance
func GetOptimizedSearch() *OptimizedPlayerSearch {
	return globalOptimizedSearch
}
