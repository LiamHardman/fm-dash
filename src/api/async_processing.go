// Package main provides async processing functionality for the FM24 API
package main

import (
	"context"
	"runtime"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// ProcessingConfig holds configuration for async processing
type ProcessingConfig struct {
	WorkerCount   int
	BatchSize     int
	BufferSize    int
	MaxGoroutines int
}

// DefaultProcessingConfig returns optimized default configuration
func DefaultProcessingConfig() ProcessingConfig {
	numCPU := runtime.NumCPU()
	return ProcessingConfig{
		WorkerCount:   numCPU * 3,  // Increased worker count for better CPU utilization
		BatchSize:     1000,        // Larger batch size for better throughput
		BufferSize:    numCPU * 32, // Larger buffer for better concurrency
		MaxGoroutines: numCPU * 6,  // Allow more concurrent goroutines
	}
}

// OptimizedProcessingConfig returns configuration optimized for large datasets
func OptimizedProcessingConfig(datasetSize int) ProcessingConfig {
	numCPU := runtime.NumCPU()

	// Adjust configuration based on dataset size with more aggressive scaling
	var batchSize, bufferSize, workerMultiplier int
	switch {
	case datasetSize > 20000:
		batchSize = 2000
		bufferSize = numCPU * 80
		workerMultiplier = 4
	case datasetSize > 10000:
		batchSize = 1500
		bufferSize = numCPU * 60
		workerMultiplier = 3
	case datasetSize > 5000:
		batchSize = 1000
		bufferSize = numCPU * 40
		workerMultiplier = 3
	default:
		batchSize = 500
		bufferSize = numCPU * 32
		workerMultiplier = 2
	}

	return ProcessingConfig{
		WorkerCount:   numCPU * workerMultiplier,
		BatchSize:     batchSize,
		BufferSize:    bufferSize,
		MaxGoroutines: numCPU * (workerMultiplier + 2),
	}
}

// PlayerProcessor handles concurrent player enhancement
type PlayerProcessor struct {
	config   ProcessingConfig
	inputCh  chan Player
	outputCh chan Player
	errorCh  chan error
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewPlayerProcessor creates a new concurrent player processor
func NewPlayerProcessor(config ProcessingConfig) *PlayerProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	processor := &PlayerProcessor{
		config:   config,
		inputCh:  make(chan Player, config.BufferSize),
		outputCh: make(chan Player, config.BufferSize),
		errorCh:  make(chan error, config.BufferSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	return processor
}

// worker processes players concurrently with improved efficiency
func (p *PlayerProcessor) worker(workerID int) {
	defer p.wg.Done()

	// Process players in batches to reduce context switching
	playerBatch := make([]Player, 0, 10)

	for {
		select {
		case player, ok := <-p.inputCh:
			if !ok {
				// Process remaining batch before exiting
				if len(playerBatch) > 0 {
					p.processBatch(playerBatch, workerID)
				}
				return
			}

			playerBatch = append(playerBatch, player)

			// Process batch when it's full or channel is empty
			if len(playerBatch) >= 10 {
				p.processBatch(playerBatch, workerID)
				playerBatch = playerBatch[:0] // Reset slice but keep capacity
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// processBatch processes a batch of players efficiently
func (p *PlayerProcessor) processBatch(batch []Player, workerID int) {
	start := time.Now()

	for i := range batch {
		// Create a copy to avoid data races
		playerCopy := batch[i]

		// Enhance player with calculations
		EnhancePlayerWithCalculations(&playerCopy)

		select {
		case p.outputCh <- playerCopy:
			// Successfully sent
		case <-p.ctx.Done():
			return
		}
	}

	// Track batch processing metrics
	RecordBusinessOperation(p.ctx, "batch_processing", true, map[string]interface{}{
		"worker_id":          workerID,
		"batch_size":         len(batch),
		"processing_time":    time.Since(start).Milliseconds(),
		"players_per_second": float64(len(batch)) / time.Since(start).Seconds(),
	})
}

// ProcessPlayersAsync processes a slice of players concurrently
func (p *PlayerProcessor) ProcessPlayersAsync(players []Player) <-chan Player {
	for i := 0; i < p.config.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	go func() {
		defer close(p.inputCh)
		for i := range players {
			select {
			case p.inputCh <- players[i]:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	go func() {
		p.wg.Wait()
		close(p.outputCh)
	}()

	return p.outputCh
}

// ProcessPlayersBatch processes players in batches for memory efficiency
func ProcessPlayersBatch(ctx context.Context, players []Player, batchSize int) ([]Player, error) {
	ctx, span := StartSpan(ctx, "async.process_players_batch")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.Int("players.total", len(players)),
		attribute.Int("batch.size", batchSize),
	)

	if len(players) == 0 {
		return players, nil
	}

	config := DefaultProcessingConfig()
	config.BatchSize = batchSize

	result := make([]Player, 0, len(players))

	for i := 0; i < len(players); i += batchSize {
		end := i + batchSize
		if end > len(players) {
			end = len(players)
		}

		batch := players[i:end]
		processor := NewPlayerProcessor(config)

		resultCh := processor.ProcessPlayersAsync(batch)

		batchResults := make([]Player, 0, len(batch))
		for player := range resultCh {
			batchResults = append(batchResults, player)
		}

		processor.Shutdown()

		result = append(result, batchResults...)

		SetSpanAttributes(ctx, attribute.Int("players.processed", len(result)))
	}

	return result, nil
}

// ProcessPercentilesAsync calculates percentiles for multiple datasets concurrently
func ProcessPercentilesAsync(datasets map[string][]Player) <-chan PercentileResult {
	resultCh := make(chan PercentileResult, len(datasets))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())

	for datasetID, players := range datasets {
		wg.Add(1)
		go func(id string, playerList []Player) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()

			playersCopy := make([]Player, len(playerList))
			copy(playersCopy, playerList)

			CalculatePlayerPerformancePercentiles(playersCopy)

			resultCh <- PercentileResult{
				DatasetID:      id,
				Players:        playersCopy,
				ProcessingTime: time.Since(start),
			}
		}(datasetID, players)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

// PercentileResult holds the result of percentile calculation
type PercentileResult struct {
	DatasetID      string
	Players        []Player
	ProcessingTime time.Duration
}

// Shutdown gracefully stops the processor
func (p *PlayerProcessor) Shutdown() {
	p.cancel()
	p.wg.Wait()
}

// ConcurrentLeagueProcessor processes league data with parallel team calculations
type ConcurrentLeagueProcessor struct {
	workerCount int
	semaphore   chan struct{}
}

// NewConcurrentLeagueProcessor creates a new league processor
func NewConcurrentLeagueProcessor(workerCount int) *ConcurrentLeagueProcessor {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	return &ConcurrentLeagueProcessor{
		workerCount: workerCount,
		semaphore:   make(chan struct{}, workerCount),
	}
}

// ProcessLeaguesAsync processes league data with concurrent team calculations
func (p *ConcurrentLeagueProcessor) ProcessLeaguesAsync(ctx context.Context, players []Player) []League {
	ctx, span := StartSpan(ctx, "async.process_leagues")
	defer span.End()

	divisionMap := make(map[string][]Player)
	for i := range players {
		division := players[i].Division
		if division == "" {
			division = "Unknown"
		}
		divisionMap[division] = append(divisionMap[division], players[i])
	}

	SetSpanAttributes(ctx, attribute.Int("divisions.count", len(divisionMap)))

	type leagueResult struct {
		league League
		err    error
	}

	resultCh := make(chan leagueResult, len(divisionMap))
	var wg sync.WaitGroup

	for divisionName, divisionPlayers := range divisionMap {
		wg.Add(1)
		go func(name string, playerList []Player) {
			defer wg.Done()

			p.semaphore <- struct{}{}
			defer func() { <-p.semaphore }()

			league := p.processLeagueSync(name, playerList)
			resultCh <- leagueResult{league: league}
		}(divisionName, divisionPlayers)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var leagues []League
	for result := range resultCh {
		if result.err == nil {
			leagues = append(leagues, result.league)
		}
	}

	// Sort leagues by overall rating
	// (sorting logic same as original)
	for i := 0; i < len(leagues)-1; i++ {
		for j := i + 1; j < len(leagues); j++ {
			if leagues[i].BestOverall < leagues[j].BestOverall {
				leagues[i], leagues[j] = leagues[j], leagues[i]
			}
		}
	}

	return leagues
}

// processLeagueSync processes a single league synchronously
func (p *ConcurrentLeagueProcessor) processLeagueSync(divisionName string, divisionPlayers []Player) League {
	league := League{
		Name:        divisionName,
		PlayerCount: len(divisionPlayers),
	}

	// Group players by team within this division
	teamMap := make(map[string][]Player)
	for i := range divisionPlayers {
		if divisionPlayers[i].Club != "" {
			teamMap[divisionPlayers[i].Club] = append(teamMap[divisionPlayers[i].Club], divisionPlayers[i])
		}
	}

	league.TeamCount = len(teamMap)

	// Calculate league ratings based on best teams (concurrent team processing)
	teamRatingsCh := make(chan TeamRatings, len(teamMap))
	var teamWg sync.WaitGroup

	for _, teamPlayers := range teamMap {
		if len(teamPlayers) >= 11 { // Only consider teams with enough players
			teamWg.Add(1)
			go func(players []Player) {
				defer teamWg.Done()
				ratings := calculateTeamRatings(players)
				if ratings.BestOverall > 0 {
					teamRatingsCh <- ratings
				}
			}(teamPlayers)
		}
	}

	go func() {
		teamWg.Wait()
		close(teamRatingsCh)
	}()

	// Collect team ratings
	var teamOveralls []int
	var allAttRatings []int
	var allMidRatings []int
	var allDefRatings []int

	for ratings := range teamRatingsCh {
		teamOveralls = append(teamOveralls, ratings.BestOverall)
		allAttRatings = append(allAttRatings, ratings.AttRating)
		allMidRatings = append(allMidRatings, ratings.MidRating)
		allDefRatings = append(allDefRatings, ratings.DefRating)
	}

	// Calculate league averages (same logic as original)
	if len(teamOveralls) > 0 {
		// Sort ratings
		for i := 0; i < len(teamOveralls)-1; i++ {
			for j := i + 1; j < len(teamOveralls); j++ {
				if teamOveralls[i] < teamOveralls[j] {
					teamOveralls[i], teamOveralls[j] = teamOveralls[j], teamOveralls[i]
				}
			}
		}

		// Take top 50% of teams or at least 3 teams
		topTeamsCount := len(teamOveralls) / 2
		if topTeamsCount < 3 && len(teamOveralls) >= 3 {
			topTeamsCount = 3
		} else if topTeamsCount < 1 {
			topTeamsCount = len(teamOveralls)
		}

		league.BestOverall = calculateAverage(teamOveralls[:topTeamsCount])
		league.AttRating = calculateAverage(allAttRatings[:topTeamsCount])
		league.MidRating = calculateAverage(allMidRatings[:topTeamsCount])
		league.DefRating = calculateAverage(allDefRatings[:topTeamsCount])
	}

	return league
}

// PlayerFilter holds filtering criteria
type PlayerFilter struct {
	Position         string
	Role             string
	MinAge           int
	MaxAge           int
	MinTransferValue int64
	MaxTransferValue int64
	MaxSalary        int64
}

// FilterResult holds filtered player and whether it matched
type FilterResult struct {
	Player  Player
	Matched bool
}

// AsyncPlayerFilter processes player filtering concurrently
type AsyncPlayerFilter struct {
	workerCount int
	chunkSize   int
}

// NewAsyncPlayerFilter creates a new async player filter
func NewAsyncPlayerFilter(workerCount, chunkSize int) *AsyncPlayerFilter {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	if chunkSize <= 0 {
		chunkSize = 100
	}

	return &AsyncPlayerFilter{
		workerCount: workerCount,
		chunkSize:   chunkSize,
	}
}

// FilterPlayersAsync filters players concurrently using chunked processing
func (f *AsyncPlayerFilter) FilterPlayersAsync(ctx context.Context, players []Player, filter PlayerFilter) []Player {
	ctx, span := StartSpan(ctx, "async.filter_players")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.Int("players.total", len(players)),
		attribute.Int("filter.chunk_size", f.chunkSize),
		attribute.Int("filter.workers", f.workerCount),
	)

	if len(players) == 0 {
		return players
	}

	// If dataset is small, use synchronous filtering
	if len(players) < f.chunkSize {
		return f.filterPlayersSync(players, filter)
	}

	// Process in chunks
	resultCh := make(chan []Player, (len(players)/f.chunkSize)+1)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, f.workerCount)

	// Split players into chunks and process each chunk concurrently
	for i := 0; i < len(players); i += f.chunkSize {
		end := i + f.chunkSize
		if end > len(players) {
			end = len(players)
		}

		chunk := players[i:end]

		wg.Add(1)
		go func(playerChunk []Player) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Filter this chunk
			filtered := f.filterPlayersSync(playerChunk, filter)
			resultCh <- filtered
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var result []Player
	for chunk := range resultCh {
		result = append(result, chunk...)
	}

	SetSpanAttributes(ctx, attribute.Int("players.filtered", len(result)))
	return result
}

// filterPlayersSync filters players synchronously (used for chunks)
func (f *AsyncPlayerFilter) filterPlayersSync(players []Player, filter PlayerFilter) []Player {
	var result []Player

	for i := range players {
		if f.matchesFilter(&players[i], filter) {
			// Apply role-specific overall if role filter is active
			playerCopy := players[i]
			if filter.Role != "" {
				for _, roleOverall := range playerCopy.RoleSpecificOveralls {
					if roleOverall.RoleName == filter.Role {
						playerCopy.Overall = roleOverall.Score
						break
					}
				}
			}
			result = append(result, playerCopy)
		}
	}

	return result
}

// matchesFilter checks if a player matches the filter criteria
func (f *AsyncPlayerFilter) matchesFilter(player *Player, filter PlayerFilter) bool {
	// Position filter
	if filter.Position != "" {
		canPlayPosition := false
		for _, shortPos := range player.ShortPositions {
			if shortPos == filter.Position {
				canPlayPosition = true
				break
			}
		}
		if !canPlayPosition {
			return false
		}
	}

	// Age filter
	if filter.MinAge > 0 || filter.MaxAge > 0 {
		playerAge, err := strconv.Atoi(player.Age)
		if err != nil {
			return false // Skip players with unparseable age
		}
		if filter.MinAge > 0 && playerAge < filter.MinAge {
			return false
		}
		if filter.MaxAge > 0 && playerAge > filter.MaxAge {
			return false
		}
	}

	// Transfer value filter
	if filter.MinTransferValue > 0 && player.TransferValueAmount < filter.MinTransferValue {
		return false
	}
	if filter.MaxTransferValue > 0 && player.TransferValueAmount > filter.MaxTransferValue {
		return false
	}

	// Salary filter
	if filter.MaxSalary > 0 && player.WageAmount > filter.MaxSalary {
		return false
	}

	// Role filter (check if player has this role)
	if filter.Role != "" {
		hasRole := false
		for _, roleOverall := range player.RoleSpecificOveralls {
			if roleOverall.RoleName == filter.Role {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return false
		}
	}

	return true
}

// ConcurrentPercentileProcessor handles percentile calculations for multiple players
type ConcurrentPercentileProcessor struct {
	workerCount int
}

// NewConcurrentPercentileProcessor creates a new percentile processor
func NewConcurrentPercentileProcessor(workerCount int) *ConcurrentPercentileProcessor {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	return &ConcurrentPercentileProcessor{
		workerCount: workerCount,
	}
}

// ProcessPercentilesWithDivisionFilterAsync processes percentiles with division filtering concurrently
func (p *ConcurrentPercentileProcessor) ProcessPercentilesWithDivisionFilterAsync(
	ctx context.Context,
	players []Player,
	divisionFilter DivisionFilter,
	targetDivision string,
) []Player {
	ctx, span := StartSpan(ctx, "async.process_percentiles_division_filter")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.Int("players.count", len(players)),
		attribute.Int("division.filter", int(divisionFilter)),
		attribute.String("division.target", targetDivision),
	)

	// Make a copy to avoid modifying original data
	playersCopy := make([]Player, len(players))
	copy(playersCopy, players)

	// Process in a goroutine to make it non-blocking
	resultCh := make(chan []Player, 1)
	go func() {
		defer close(resultCh)

		// Calculate percentiles with division filter
		CalculatePlayerPerformancePercentilesWithDivisionFilter(playersCopy, divisionFilter, targetDivision)
		resultCh <- playersCopy
	}()

	// Return processed players
	return <-resultCh
}
