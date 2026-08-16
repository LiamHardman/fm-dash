package main

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func lengthPrefixedString(s string) []byte {
	buf := make([]byte, 4+len(s))
	binary.LittleEndian.PutUint32(buf, uint32(len(s)))
	copy(buf[4:], s)
	return buf
}

func buildFakeFMSave(t *testing.T, decompressedContent []byte) []byte {
	t.Helper()

	var compressed bytes.Buffer
	enc, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatalf("creating zstd writer: %v", err)
	}
	if _, err := enc.Write(decompressedContent); err != nil {
		t.Fatalf("writing zstd data: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("closing zstd writer: %v", err)
	}

	header := make([]byte, fmSaveHeaderSize)
	copy(header[2:6], fmSaveMagic)

	var full bytes.Buffer
	full.Write(header)
	full.Write(compressed.Bytes())
	return full.Bytes()
}

func TestParseFMSave_ExtractsLengthPrefixedStrings(t *testing.T) {
	var content bytes.Buffer
	content.Write(lengthPrefixedString("26.2.0+0"))
	content.Write(lengthPrefixedString("WINHTTP.DLL"))
	content.Write([]byte{0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff}) // binary noise, no match expected

	save := buildFakeFMSave(t, content.Bytes())

	info, err := ParseFMSave(bytes.NewReader(save))
	if err != nil {
		t.Fatalf("ParseFMSave: %v", err)
	}

	if len(info.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d: %+v", len(info.Fields), info.Fields)
	}
	if info.Fields[0].Value != "26.2.0+0" || info.Fields[0].Offset != 0 {
		t.Errorf("unexpected first field: %+v", info.Fields[0])
	}
	if info.Fields[1].Value != "WINHTTP.DLL" {
		t.Errorf("unexpected second field: %+v", info.Fields[1])
	}
}

func TestParseFMSave_RejectsWrongMagic(t *testing.T) {
	header := make([]byte, fmSaveHeaderSize)
	copy(header[2:6], "xxxx")

	_, err := ParseFMSave(bytes.NewReader(header))
	if err == nil {
		t.Fatal("expected an error for a file missing the fmf. magic bytes")
	}
}

func TestScanLengthPrefixedStrings_SkipsNonPrintableAndOutOfRangeLengths(t *testing.T) {
	var data bytes.Buffer
	data.Write(lengthPrefixedString("ok"))           // too short (min 3)
	data.Write(lengthPrefixedString("valid string")) // should match
	binary.Write(&data, binary.LittleEndian, uint32(5))
	data.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05}) // non-printable, should not match

	fields := scanLengthPrefixedStrings(data.Bytes())
	if len(fields) != 1 || fields[0].Value != "valid string" {
		t.Fatalf("expected exactly one field %q, got %+v", "valid string", fields)
	}
}

func TestScanLengthPrefixedStrings_AcceptsUTF8Diacritics(t *testing.T) {
	// Real names found in test_save.fm carry UTF-8 diacritics (e.g. "Fernández", "Koundé") --
	// an ASCII-only check silently drops these, so this guards against regressing that.
	var data bytes.Buffer
	data.Write(lengthPrefixedString("Fernández"))
	data.Write(lengthPrefixedString("Koundé"))
	data.Write(lengthPrefixedString("García"))

	fields := scanLengthPrefixedStrings(data.Bytes())
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %+v", len(fields), fields)
	}
	want := []string{"Fernández", "Koundé", "García"}
	for i, w := range want {
		if fields[i].Value != w {
			t.Errorf("field %d: got %q, want %q", i, fields[i].Value, w)
		}
	}
}
