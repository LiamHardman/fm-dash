package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

// WorkerStats tracks performance metrics for worker pools
type WorkerStats struct {
	ProcessedCount int64
	ErrorCount     int64
	TotalDuration  int64 // in nanoseconds
}

// PlayerParserWorker is a goroutine that receives raw cell data (a table row) from a channel,
// parses it into a Player struct, enhances it with calculations, and sends the result (or error)
// to another channel.
func PlayerParserWorker(workerID int, rowCellsChan <-chan []string, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup, headers []string) {
	defer func() {
		RecordWorkerEnd() // Update global metrics
		if r := recover(); r != nil {
			// Log panic from worker to avoid silent failures
			log.Printf("Worker %d PANICKED: %v", workerID, r)
		}
		wg.Done() // Signal completion to the WaitGroup
	}()

	RecordWorkerStart() // Update global metrics

	if len(headers) == 0 {
		log.Printf("Worker %d started with NO headers. Draining rowCellsChan and exiting.", workerID)
		// Consume any rows sent before this worker realized headers were missing.
		for range rowCellsChan {
		}
		return
	}

	processedRows := 0
	for cells := range rowCellsChan { // Process rows until the channel is closed
		player, err := parseCellsToPlayer(cells, headers) // From player_processing.go
		if err == nil {
			// If initial parsing is successful, enhance the player with calculated stats.
			EnhancePlayerWithCalculations(&player) // From player_processing.go (modifies player in-place)
			processedRows++
			atomic.AddInt64(&globalMetrics.TotalPlayersProcessed, 1)
		}
		resultsChan <- PlayerParseResult{Player: player, Err: err}
	}

	log.Printf("Worker %d finished: processed %d rows", workerID, processedRows)
}

// OptimizedPlayerParserWorker is an enhanced version with better error handling and metrics
func OptimizedPlayerParserWorker(workerID int, rowCellsChan <-chan []string, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup, headers []string, stats *WorkerStats) {
	defer func() {
		RecordWorkerEnd() // Update global metrics
		if r := recover(); r != nil {
			// Log panic from worker to avoid silent failures
			log.Printf("Worker %d PANICKED: %v", workerID, r)
			// Send error result to prevent deadlock
			select {
			case resultsChan <- PlayerParseResult{Err: fmt.Errorf("worker %d panicked: %v", workerID, r)}:
			default:
				// Channel might be full, log the issue
				log.Printf("Worker %d: couldn't send panic error, channel full", workerID)
			}
		}
		wg.Done() // Signal completion to the WaitGroup
	}()

	RecordWorkerStart() // Update global metrics

	if len(headers) == 0 {
		log.Printf("Worker %d started with NO headers. Draining rowCellsChan and exiting.", workerID)
		// Consume any rows sent before this worker realized headers were missing.
		for range rowCellsChan {
		}
		return
	}

	processedRows := 0
	localErrorCount := 0

	// Batch processing to reduce channel overhead
	const batchSize = 10
	rowBatch := make([][]string, 0, batchSize)

	for {
		// Try to fill a batch
		rowBatch = rowBatch[:0] // Reset batch but keep capacity

		// First, get at least one row or exit if channel is closed
		cells, ok := <-rowCellsChan
		if !ok {
			break // Channel closed, no more work
		}
		rowBatch = append(rowBatch, cells)

		// Try to get more rows for batch processing (non-blocking)
		for len(rowBatch) < batchSize {
			select {
			case moreCells, moreOk := <-rowCellsChan:
				if !moreOk {
					// Channel closed, process what we have
					goto processBatch
				}
				rowBatch = append(rowBatch, moreCells)
			default:
				// No more rows available immediately, process current batch
				goto processBatch
			}
		}

	processBatch:
		// Process the batch
		for _, cellRow := range rowBatch {
			player, err := parseCellsToPlayer(cellRow, headers)
			if err == nil {
				// If initial parsing is successful, enhance the player with calculated stats.
				EnhancePlayerWithCalculations(&player)
				processedRows++
				atomic.AddInt64(&globalMetrics.TotalPlayersProcessed, 1)
			} else {
				localErrorCount++
			}

			// Send result with timeout to prevent blocking
			select {
			case resultsChan <- PlayerParseResult{Player: player, Err: err}:
				// Successfully sent
			default:
				// Channel full, record backpressure and try blocking send
				RecordBackpressure()
				log.Printf("Worker %d: result channel full, potential backpressure", workerID)
				// Try one more time with blocking send
				resultsChan <- PlayerParseResult{Player: player, Err: err}
			}
		}

		// Exit if channel was closed during batch collection
		if !ok {
			break
		}
	}

	// Update stats atomically
	if stats != nil {
		atomic.AddInt64(&stats.ProcessedCount, int64(processedRows))
		atomic.AddInt64(&stats.ErrorCount, int64(localErrorCount))
	}

	log.Printf("Worker %d finished: processed %d rows, %d errors", workerID, processedRows, localErrorCount)
}
