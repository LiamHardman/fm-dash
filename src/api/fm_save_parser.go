package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/klauspost/compress/zstd"
)

const (
	// fmSaveHeaderSize is the size in bytes of the custom pre-zstd header found at the start
	// of a Football Manager .fm save file. Its fields beyond the magic bytes are undocumented
	// (see .scratch/fm-save-parsing/issues/02-research-findings.md) and are not otherwise used.
	fmSaveHeaderSize = 26
	fmSaveMagic      = "fmf."

	// fmSaveScanWindow bounds how much of the decompressed archive is scanned, per the map's
	// decision to keep server-side cost low rather than scanning the full ~800MB archive.
	fmSaveScanWindow = 2 * 1024 * 1024

	fmSaveMinStringLen = 3
	fmSaveMaxStringLen = 200
)

// FMSaveField is one string recovered from an .fm save by the length-prefixed-string scanner.
type FMSaveField struct {
	Offset int64  `json:"offset"`
	Value  string `json:"value"`
}

// FMSaveInfo is the basic information extracted from a Football Manager .fm save file.
type FMSaveInfo struct {
	Fields []FMSaveField `json:"fields"`
}

// ParseFMSave decompresses an .fm save file and scans a window near the start of the
// decompressed archive for length-prefixed ASCII strings (u32 LE length + raw ASCII bytes, no
// terminator) -- the only confirmed decoding primitive for this undocumented, proprietary
// format (see .scratch/fm-save-parsing/map.md). There is no known record schema, so results are
// an ordered list of whatever plausible strings were found, not a fixed set of named fields.
func ParseFMSave(r io.Reader) (*FMSaveInfo, error) {
	header := make([]byte, fmSaveHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("reading fm save header: %w", err)
	}
	if string(header[2:6]) != fmSaveMagic {
		return nil, fmt.Errorf("not a recognized .fm save file (missing %q magic)", fmSaveMagic)
	}

	dec, err := zstd.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("opening zstd stream: %w", err)
	}
	defer dec.Close()

	window := make([]byte, fmSaveScanWindow)
	n, err := io.ReadFull(dec, window)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("decompressing fm save: %w", err)
	}
	window = window[:n]

	return &FMSaveInfo{Fields: scanLengthPrefixedStrings(window)}, nil
}

// scanLengthPrefixedStrings walks data looking for the confirmed encoding: a little-endian
// u32 length immediately followed by that many printable-ASCII bytes. On a match, it skips past
// the consumed string rather than re-scanning inside it, to avoid overlapping duplicate hits.
func scanLengthPrefixedStrings(data []byte) []FMSaveField {
	var fields []FMSaveField
	for i := 0; i+4 < len(data); i++ {
		length := binary.LittleEndian.Uint32(data[i : i+4])
		if length < fmSaveMinStringLen || length > fmSaveMaxStringLen {
			continue
		}
		start := i + 4
		end := start + int(length)
		if end > len(data) {
			continue
		}
		candidate := data[start:end]
		if !isPlausibleString(candidate) {
			continue
		}
		fields = append(fields, FMSaveField{
			Offset: int64(i),
			Value:  string(candidate),
		})
		i = end - 1 // skip past the consumed string; loop's i++ lands us right after it
	}
	return fields
}

// isPlausibleString accepts printable ASCII plus valid UTF-8 (many player/place names in this
// format carry real diacritics -- "Fernández", "Koundé", "García" -- encoded as UTF-8, which an
// ASCII-only check silently drops). Purely numeric/punctuation strings (like a version number
// "26.2.0+0") are kept -- rejecting them would also reject real fields, since the format's
// version string is exactly this shape.
func isPlausibleString(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r < 0x20 || (r < 0x80 && r > 0x7e) {
			return false
		}
		b = b[size:]
	}
	return true
}
