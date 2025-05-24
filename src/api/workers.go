package main

import (
	"log"
	"sync"
)

// PlayerParserWorker is a goroutine that receives raw cell data (a table row) from a channel,
// parses it into a Player struct, enhances it with calculations, and sends the result (or error)
// to another channel.
func PlayerParserWorker(workerID int, rowCellsChan <-chan []string, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup, headers []string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic from worker to avoid silent failures
			log.Printf("Worker %d PANICKED: %v", workerID, r)
			// Optionally, send an error result to resultsChan if appropriate,
			// but the primary goal is to log and ensure wg.Done() is called.
		}
		wg.Done() // Signal completion to the WaitGroup
	}()

	if len(headers) == 0 {
		log.Printf("Worker %d started with NO headers. Draining rowCellsChan and exiting.", workerID)
		// Consume any rows sent before this worker realized headers were missing.
		// This prevents deadlock if the producer (HTML parser) sends rows before all workers check headers.
		for range rowCellsChan {
			// Discard rows
		}
		return
	}

	// log.Printf("Worker %d started with headers: %v", workerID, headers) // Debug: confirm headers

	for cells := range rowCellsChan { // Process rows until the channel is closed
		player, err := parseCellsToPlayer(cells, headers) // From player_processing.go
		if err == nil {
			// If initial parsing is successful, enhance the player with calculated stats.
			EnhancePlayerWithCalculations(&player) // From player_processing.go (modifies player in-place)
		}
		// Send the result (player or error) to the results channel.
		// The error could be from parseCellsToPlayer. EnhancePlayerWithCalculations might log its own errors
		// but doesn't typically return one that would overwrite the parsing error here.
		resultsChan <- PlayerParseResult{Player: player, Err: err}
	}
	// log.Printf("Worker %d finished processing rows and is exiting.", workerID)
}
