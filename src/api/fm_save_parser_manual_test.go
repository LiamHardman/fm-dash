package main

import (
	"os"
	"testing"
)

// TestParseFMSaveManual is a throwaway manual verification against the real test_save.fm
// fixture, gated behind an env var so it never runs in CI. Delete once the parser is trusted.
func TestParseFMSaveManual(t *testing.T) {
	if os.Getenv("FM_SAVE_MANUAL_TEST") == "" {
		t.Skip("set FM_SAVE_MANUAL_TEST=1 to run against test_save.fm")
	}
	f, err := os.Open("../../test_save.fm")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	info, err := ParseFMSave(f)
	if err != nil {
		t.Fatalf("ParseFMSave: %v", err)
	}
	t.Logf("found %d fields", len(info.Fields))
	for _, fld := range info.Fields {
		t.Logf("  offset=%d %q", fld.Offset, fld.Value)
	}
}
