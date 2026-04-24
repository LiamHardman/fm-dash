package main

import (
	"sync"
	"time"
)

// pendingPercentileCalcs maps datasetID → done channel.
// The upload handler registers a channel before starting the background percentile
// calculation goroutine. The GET handler waits on the channel rather than re-running
// the calculation from scratch, which would otherwise happen concurrently and double
// the work for large datasets.
var pendingPercentileCalcs sync.Map

func registerPendingPercentileCalc(datasetID string) chan struct{} {
	doneCh := make(chan struct{})
	pendingPercentileCalcs.Store(datasetID, doneCh)
	return doneCh
}

// completePendingPercentileCalc must be called after SetPlayerData so that any
// waiter that unblocks immediately sees the updated stored data.
func completePendingPercentileCalc(datasetID string, doneCh chan struct{}) {
	close(doneCh)
	pendingPercentileCalcs.Delete(datasetID)
}

// waitForPendingPercentileCalc blocks until the in-progress background calculation
// for datasetID finishes, or until the timeout elapses. Returns true if the
// calculation completed (caller should reload player data), false on timeout.
func waitForPendingPercentileCalc(datasetID string) bool {
	val, ok := pendingPercentileCalcs.Load(datasetID)
	if !ok {
		return false
	}
	doneCh := val.(chan struct{})
	select {
	case <-doneCh:
		return true
	case <-time.After(60 * time.Second):
		return false
	}
}
