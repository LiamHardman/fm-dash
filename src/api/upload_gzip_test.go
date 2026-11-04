package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"testing"
)

func gzipBytes(t *testing.T, content []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestDecodeCompressedUpload(t *testing.T) {
	name, content, err := decodeCompressedUpload("My Save.CSV.GZ", gzipBytes(t, []byte("Name;Age\n")))
	if err != nil {
		t.Fatal(err)
	}
	if name != "My Save.CSV" || string(content) != "Name;Age\n" {
		t.Fatalf("got %q / %q", name, content)
	}

	plain := []byte("Name;Age\n")
	name, content, err = decodeCompressedUpload("players.csv", plain)
	if err != nil || name != "players.csv" || !bytes.Equal(content, plain) {
		t.Fatalf("plain upload changed: %q %q %v", name, content, err)
	}

	if _, _, err = decodeCompressedUpload("players.csv.gz", []byte("not gzip")); err == nil {
		t.Fatal("expected an error for corrupt gzip")
	}
}

func TestDecodeCompressedUploadRefusesBombs(t *testing.T) {
	bomb := gzipBytes(t, make([]byte, maxDecompressedUploadBytes+1))
	if _, _, err := decodeCompressedUpload("bomb.csv.gz", bomb); !errors.Is(err, errDecompressedUploadTooLarge) {
		t.Fatalf("expected the size cap, got %v", err)
	}
}

func TestUploadPreflightAcceptsGzippedSaveExport(t *testing.T) {
	content, err := os.ReadFile("testdata/fmsave/save_export.csv")
	if err != nil {
		t.Fatal(err)
	}
	response := runUploadPreflight(t, "My Save.csv.gz", string(gzipBytes(t, content)))
	if !response.Valid || response.Format != "csv" {
		t.Fatalf("gzipped save export rejected: %+v", response)
	}
}
