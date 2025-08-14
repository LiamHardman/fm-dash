package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// compressData compresses data using gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		if closeErr := gz.Close(); closeErr != nil {
			LogWarn("Failed to close gzip writer: %v", closeErr)
		}
	}()

	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := gz.Close(); err != nil {
		LogWarn("Failed to close gzip writer: %v", err)
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			LogWarn("Failed to close gzip reader: %v", closeErr)
		}
	}()

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip reader: %w", err)
	}

	return decompressedData, nil
}
