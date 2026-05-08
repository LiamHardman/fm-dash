package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	pb "api/proto"
)

const (
	// Streaming constants
	defaultStreamChunkSize = 1000   // Players per chunk
	defaultStreamBuffer    = 8192   // 8KB buffer for JSON streaming
	maxStreamingPlayers    = 100000 // Don't stream for small datasets
)

type streamingContextKey string

const streamingEnabledKey streamingContextKey = "streaming_enabled"

// StreamingSerializer provides memory-efficient serialization for large datasets
type StreamingSerializer struct {
	chunkSize int
	buffer    *bufio.Writer
	gzWriter  *gzip.Writer
}

// StreamingPlayerResponse represents a chunk of players in a streaming response
type StreamingPlayerResponse struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currency_symbol"`
	ChunkIndex     int      `json:"chunk_index"`
	TotalChunks    int      `json:"total_chunks"`
	IsLast         bool     `json:"is_last"`
	Metadata       struct {
		Timestamp     int64   `json:"timestamp"`
		RequestID     string  `json:"request_id"`
		PlayerCount   int     `json:"player_count"`
		MemoryUsageMB float64 `json:"memory_usage_mb"`
	} `json:"metadata"`
}

// CreateStreamingSerializer creates a new streaming serializer
func CreateStreamingSerializer(w io.Writer, enableCompression bool) *StreamingSerializer {
	buffer := bufio.NewWriterSize(w, defaultStreamBuffer)

	var gzWriter *gzip.Writer
	var finalWriter io.Writer = buffer

	if enableCompression {
		gzWriter = gzip.NewWriter(buffer)
		finalWriter = gzWriter
	}

	return &StreamingSerializer{
		chunkSize: defaultStreamChunkSize,
		buffer:    bufio.NewWriter(finalWriter),
		gzWriter:  gzWriter,
	}
}

// StreamPlayers streams players in chunks to reduce memory pressure
func (s *StreamingSerializer) StreamPlayers(ctx context.Context, w http.ResponseWriter, players []Player, currencySymbol, requestID string) error {
	ctx, span := StartSpan(ctx, "streaming.serialize_players")
	defer span.End()

	playerCount := len(players)

	SetSpanAttributes(ctx,
		attribute.Int("streaming.player_count", playerCount),
		attribute.Int("streaming.chunk_size", s.chunkSize),
		attribute.Bool("streaming.enabled", playerCount > maxStreamingPlayers),
	)

	// For small datasets, use regular serialization
	if playerCount <= maxStreamingPlayers {
		return s.streamSmallDataset(ctx, w, players, currencySymbol, requestID)
	}

	// Set streaming headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Streaming", "true")
	w.Header().Set("Cache-Control", "no-cache")

	// Calculate chunks
	totalChunks := (playerCount + s.chunkSize - 1) / s.chunkSize

	logInfo(ctx, "Starting streaming serialization",
		"total_players", playerCount,
		"chunk_size", s.chunkSize,
		"total_chunks", totalChunks)

	// Stream header
	if err := s.writeStreamHeader(w, playerCount, totalChunks, requestID); err != nil {
		return fmt.Errorf("failed to write stream header: %w", err)
	}

	// Stream chunks
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		start := chunkIndex * s.chunkSize
		end := start + s.chunkSize
		if end > playerCount {
			end = playerCount
		}

		chunk := players[start:end]
		isLast := chunkIndex == totalChunks-1

		// Monitor memory during streaming
		memStats := GetCurrentMemoryStats()

		if err := s.writePlayerChunk(ctx, w, chunk, currencySymbol, chunkIndex, totalChunks, isLast, requestID, memStats.TotalAllocMB); err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", chunkIndex, err)
		}

		// Force flush every 10 chunks to reduce buffering
		if chunkIndex%10 == 0 {
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}

		// Yield CPU every 50 chunks for better concurrency
		if chunkIndex%50 == 0 {
			runtime.Gosched()
		}

		SetSpanAttributes(ctx,
			attribute.Int("streaming.chunk_processed", chunkIndex+1),
			attribute.Float64("streaming.memory_mb", memStats.TotalAllocMB),
		)
	}

	// Stream footer
	if err := s.writeStreamFooter(w); err != nil {
		return fmt.Errorf("failed to write stream footer: %w", err)
	}

	logInfo(ctx, "Streaming serialization completed",
		"total_chunks_sent", totalChunks,
		"total_players", playerCount)

	return nil
}

// streamSmallDataset handles small datasets with regular serialization
func (s *StreamingSerializer) streamSmallDataset(ctx context.Context, w http.ResponseWriter, players []Player, currencySymbol, requestID string) error {
	response := StreamingPlayerResponse{
		Players:        players,
		CurrencySymbol: currencySymbol,
		ChunkIndex:     0,
		TotalChunks:    1,
		IsLast:         true,
	}

	response.Metadata.Timestamp = time.Now().Unix()
	response.Metadata.RequestID = requestID
	response.Metadata.PlayerCount = len(players)
	response.Metadata.MemoryUsageMB = GetCurrentMemoryStats().TotalAllocMB

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Streaming", "false")

	return json.NewEncoder(w).Encode(response)
}

// writeStreamHeader writes the initial stream metadata
func (s *StreamingSerializer) writeStreamHeader(w io.Writer, playerCount, totalChunks int, requestID string) error {
	header := map[string]interface{}{
		"stream_info": map[string]interface{}{
			"total_players": playerCount,
			"total_chunks":  totalChunks,
			"chunk_size":    s.chunkSize,
			"request_id":    requestID,
			"timestamp":     time.Now().Unix(),
		},
	}

	data, err := json.Marshal(header)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// writePlayerChunk writes a chunk of players
func (s *StreamingSerializer) writePlayerChunk(ctx context.Context, w io.Writer, players []Player, currencySymbol string, chunkIndex, totalChunks int, isLast bool, requestID string, memoryMB float64) error {
	chunk := StreamingPlayerResponse{
		Players:        players,
		CurrencySymbol: currencySymbol,
		ChunkIndex:     chunkIndex,
		TotalChunks:    totalChunks,
		IsLast:         isLast,
	}

	chunk.Metadata.Timestamp = time.Now().Unix()
	chunk.Metadata.RequestID = requestID
	chunk.Metadata.PlayerCount = len(players)
	chunk.Metadata.MemoryUsageMB = memoryMB

	data, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// writeStreamFooter writes the final stream metadata
func (s *StreamingSerializer) writeStreamFooter(w io.Writer) error {
	footer := map[string]interface{}{
		"stream_end": map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"status":    "completed",
		},
	}

	data, err := json.Marshal(footer)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// ProtobufStreamingSerializer handles protobuf streaming
type ProtobufStreamingSerializer struct {
	chunkSize int
}

// CreateProtobufStreamingSerializer creates a protobuf streaming serializer
func CreateProtobufStreamingSerializer() *ProtobufStreamingSerializer {
	return &ProtobufStreamingSerializer{
		chunkSize: defaultStreamChunkSize,
	}
}

// StreamPlayersProtobuf streams players using protobuf format
func (ps *ProtobufStreamingSerializer) StreamPlayersProtobuf(ctx context.Context, w http.ResponseWriter, players []Player, currencySymbol, requestID string) error {
	ctx, span := StartSpan(ctx, "streaming.protobuf.serialize_players")
	defer span.End()

	playerCount := len(players)

	// For small datasets, use regular protobuf serialization
	if playerCount <= maxStreamingPlayers {
		return ps.streamSmallProtobufDataset(ctx, w, players, currencySymbol, requestID)
	}

	// Set protobuf streaming headers
	w.Header().Set("Content-Type", "application/x-protobuf-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Streaming", "true")

	totalChunks := (playerCount + ps.chunkSize - 1) / ps.chunkSize

	// Stream protobuf chunks
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		start := chunkIndex * ps.chunkSize
		end := start + ps.chunkSize
		if end > playerCount {
			end = playerCount
		}

		chunk := players[start:end]

		if err := ps.writeProtobufChunk(ctx, w, chunk, currencySymbol, chunkIndex, totalChunks, requestID); err != nil {
			return fmt.Errorf("failed to write protobuf chunk %d: %w", chunkIndex, err)
		}

		// Flush every 10 chunks
		if chunkIndex%10 == 0 {
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}

	return nil
}

// streamSmallProtobufDataset handles small datasets with regular protobuf
func (ps *ProtobufStreamingSerializer) streamSmallProtobufDataset(ctx context.Context, w http.ResponseWriter, players []Player, currencySymbol, requestID string) error {
	// Convert to protobuf
	protoPlayers := make([]*pb.Player, len(players))
	for i, player := range players {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			return fmt.Errorf("failed to convert player %d to protobuf: %w", i, err)
		}
		protoPlayers[i] = protoPlayer
	}

	response := &pb.PlayerDataResponse{
		Players:        protoPlayers,
		CurrencySymbol: currencySymbol,
		Metadata: &pb.ResponseMetadata{
			Timestamp:  time.Now().Unix(),
			RequestId:  requestID,
			TotalCount: int32(len(players)),
		},
	}

	data, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("X-Streaming", "false")

	_, err = w.Write(data)
	return err
}

// writeProtobufChunk writes a protobuf chunk
func (ps *ProtobufStreamingSerializer) writeProtobufChunk(ctx context.Context, w io.Writer, players []Player, currencySymbol string, chunkIndex, totalChunks int, requestID string) error {
	// Convert chunk to protobuf
	protoPlayers := make([]*pb.Player, len(players))
	for i, player := range players {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			return fmt.Errorf("failed to convert player %d to protobuf: %w", i, err)
		}
		protoPlayers[i] = protoPlayer
	}

	chunk := &pb.PlayerDataResponse{
		Players:        protoPlayers,
		CurrencySymbol: currencySymbol,
		Metadata: &pb.ResponseMetadata{
			Timestamp:  time.Now().Unix(),
			RequestId:  requestID,
			TotalCount: int32(len(players)),
		},
	}

	data, err := proto.Marshal(chunk)
	if err != nil {
		return err
	}

	// Write chunk size prefix (for delimiting chunks)
	chunkSize := len(data)
	sizeBytes := make([]byte, 4)
	sizeBytes[0] = byte(chunkSize >> 24)
	sizeBytes[1] = byte(chunkSize >> 16)
	sizeBytes[2] = byte(chunkSize >> 8)
	sizeBytes[3] = byte(chunkSize)

	if _, err := w.Write(sizeBytes); err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// ShouldUseStreaming determines if streaming should be used based on dataset size
func ShouldUseStreaming(playerCount int) bool {
	return playerCount > maxStreamingPlayers
}

// GetOptimalChunkSize calculates optimal chunk size based on available memory
func GetOptimalChunkSize(playerCount int) int {
	memStats := GetCurrentMemoryStats()

	// Base chunk size on available memory
	availableMemoryMB := 1024 - memStats.TotalAllocMB // Assume 1GB limit
	if availableMemoryMB < 100 {
		return 500 // Small chunks for low memory
	} else if availableMemoryMB < 500 {
		return 1000 // Medium chunks
	}

	return defaultStreamChunkSize // Default chunks for good memory
}

// StreamingHandler wraps existing handlers to use streaming when appropriate
func StreamingHandler(originalHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if client supports streaming
		acceptsStreaming := r.Header.Get("Accept-Streaming") == "true" ||
			r.Header.Get("X-Supports-Streaming") == "true"

		if acceptsStreaming {
			// Add streaming context
			ctx = context.WithValue(ctx, streamingEnabledKey, true)
			r = r.WithContext(ctx)
		}

		originalHandler(w, r)
	}
}
