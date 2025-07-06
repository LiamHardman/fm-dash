package main

import (
	"bytes"
	"encoding/json"
	"sync"
)

// JSON buffer pool for reducing allocations during encoding/decoding
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// getJSONBuffer gets a buffer from the pool
func getJSONBuffer() *bytes.Buffer {
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// putJSONBuffer returns a buffer to the pool
func putJSONBuffer(buf *bytes.Buffer) {
	if buf.Cap() <= 64*1024 { // Don't pool extremely large buffers
		jsonBufferPool.Put(buf)
	}
}

// FastJSONMarshal provides optimized JSON marshaling with buffer pooling
func FastJSONMarshal(v interface{}) ([]byte, error) {
	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	encoder := json.NewEncoder(buf)
	// Disable HTML escaping for better performance
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// Remove the trailing newline that json.Encoder adds
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Return a copy since we're returning the buffer to the pool
	output := make([]byte, len(result))
	copy(output, result)
	return output, nil
}

// FastJSONUnmarshal provides optimized JSON unmarshaling
func FastJSONUnmarshal(data []byte, v interface{}) error {
	// For small data, use standard unmarshaling
	if len(data) < 1024 {
		return json.Unmarshal(data, v)
	}

	// For larger data, use streaming decoder
	buf := bytes.NewReader(data)
	decoder := json.NewDecoder(buf)
	return decoder.Decode(v)
}

// PlayerJSONOptimizer provides specialized JSON operations for Player structs
type PlayerJSONOptimizer struct {
	buf     *bytes.Buffer
	encoder *json.Encoder
}

// NewPlayerJSONOptimizer creates a new optimizer instance
func NewPlayerJSONOptimizer() *PlayerJSONOptimizer {
	buf := getJSONBuffer()
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)

	return &PlayerJSONOptimizer{
		buf:     buf,
		encoder: encoder,
	}
}

// MarshalPlayer efficiently marshals a Player struct
func (opt *PlayerJSONOptimizer) MarshalPlayer(player *Player) ([]byte, error) {
	opt.buf.Reset()

	if err := opt.encoder.Encode(player); err != nil {
		return nil, err
	}

	// Remove trailing newline
	result := opt.buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	output := make([]byte, len(result))
	copy(output, result)
	return output, nil
}

// MarshalPlayers efficiently marshals a slice of Player structs
func (opt *PlayerJSONOptimizer) MarshalPlayers(players []Player) ([]byte, error) {
	opt.buf.Reset()

	if err := opt.encoder.Encode(players); err != nil {
		return nil, err
	}

	// Remove trailing newline
	result := opt.buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	output := make([]byte, len(result))
	copy(output, result)
	return output, nil
}

// Close returns buffers to pools
func (opt *PlayerJSONOptimizer) Close() {
	putJSONBuffer(opt.buf)
}

// Global optimizer instance for reuse
var globalPlayerJSONOptimizer = NewPlayerJSONOptimizer()

// OptimizedPlayerMarshal is a convenience function using the global optimizer
func OptimizedPlayerMarshal(player *Player) ([]byte, error) {
	return globalPlayerJSONOptimizer.MarshalPlayer(player)
}

// OptimizedPlayersMarshal is a convenience function for marshaling player slices
func OptimizedPlayersMarshal(players []Player) ([]byte, error) {
	return globalPlayerJSONOptimizer.MarshalPlayers(players)
}
