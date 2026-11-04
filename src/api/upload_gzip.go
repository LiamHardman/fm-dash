package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"
)

// maxDecompressedUploadBytes caps a gzipped upload once inflated. A whole-database
// export from an FM27 save is about 25 MB of CSV, so this leaves ample room while
// refusing decompression bombs.
const maxDecompressedUploadBytes = 256 * 1024 * 1024

var errDecompressedUploadTooLarge = errors.New("the compressed export inflates past the size limit")

// decodeCompressedUpload inflates a `.csv.gz` upload (the browser FM27 save import sends
// one, because a whole-database export is larger than the default upload limit). It
// returns the name without `.gz` so format detection sees `.csv`. Other uploads pass
// through untouched.
func decodeCompressedUpload(filename string, content []byte) (string, []byte, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".csv.gz") {
		return filename, content, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		return filename, nil, fmt.Errorf("the compressed export could not be read: %w", err)
	}
	defer func() { _ = reader.Close() }()
	inflated, err := io.ReadAll(io.LimitReader(reader, maxDecompressedUploadBytes+1))
	if err != nil {
		return filename, nil, fmt.Errorf("the compressed export could not be read: %w", err)
	}
	if len(inflated) > maxDecompressedUploadBytes {
		return filename, nil, errDecompressedUploadTooLarge
	}
	return filename[:len(filename)-len(".gz")], inflated, nil
}
