package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

func cliInputFormat(path, override string) (string, error) {
	if override != "" {
		if override != "csv" && override != "html" && override != "fm" {
			return "", cliFail(2, "input-format must be csv, html or fm")
		}
		return override, nil
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv":
		return "csv", nil
	case ".html", ".htm":
		return "html", nil
	case ".fm":
		return "fm", nil
	default:
		return "", cliFail(2, "cannot infer input format; supply --input-format (required for stdin)")
	}
}

func cliParse(ctx context.Context, command string, args []string, o cliOptions, stdin io.Reader, e *cliEnvelope) (any, error) {
	if command == "parse diff" {
		return cliDiff(ctx, o, stdin)
	}
	if command == "parse validate" || command == "parse bench" {
		return cliEvaluate(ctx, command, o, stdin)
	}
	if len(args) != 1 {
		return nil, cliFail(2, "%s requires one input path or -", command)
	}
	format, err := cliInputFormat(args[0], o.inputFormat)
	if err != nil {
		return nil, err
	}
	readStart := time.Now()
	data, err := cliReadInput(ctx, args[0], stdin, o.maxBytes)
	if err != nil {
		return nil, err
	}
	e.Provenance["inputSha256"] = cliHash(data)
	e.Provenance["inputBytes"] = len(data)
	e.Provenance["readMs"] = float64(time.Since(readStart).Nanoseconds()) / 1e6
	if command == "parse fm" {
		if format != "fm" {
			return nil, cliFail(2, "parse fm requires .fm input or --input-format fm")
		}
		info, err := InspectFMSave(ctx, bytes.NewReader(data))
		if err != nil {
			return nil, cliFail(3, "native save: %v", err)
		}
		return info, nil
	}
	if format == "fm" {
		return nil, cliFail(2, "native saves support parse fm string inspection only")
	}
	if command == "parse preflight" {
		headersFormat, headers, err := extractExportHeaders("input."+format, data)
		if err != nil {
			return nil, cliFail(3, "preflight: %v", err)
		}
		missing := missingRequiredExportHeaders(headers)
		response := UploadPreflightResponse{Valid: len(missing) == 0, Format: headersFormat, Headers: headers, MissingColumns: missing, Message: "Export is ready to upload."}
		if !response.Valid {
			response.Message = "Export is missing columns from the FM Dash Search View."
			return response, cliFail(3, "%s", response.Message)
		}
		return response, nil
	}
	if command != "parse players" {
		return nil, cliFail(2, "unknown parser operation")
	}
	parsed, err := ParsePlayers(ctx, bytes.NewReader(data), ParseOptions{format, o.stage, o.workers, o.strict})
	if parsed != nil {
		e.Diagnostics = parsed.Diagnostics
		if parsed.Counts.Errors > 0 {
			e.Status = "partial"
		}
	}
	if err != nil && !errors.Is(err, errStrictParse) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		err = cliFail(3, "parse: %v", err)
	}
	return parsed, err
}

type FMSaveInspection struct {
	Mode            string        `json:"mode"`
	Fields          []FMSaveField `json:"fields"`
	ScannedBytes    int           `json:"scannedBytes"`
	ScanLimit       int           `json:"scanLimit"`
	Truncated       bool          `json:"truncated"`
	OffsetOrigin    string        `json:"offsetOrigin"`
	MinStringLength int           `json:"minStringLength"`
	MaxStringLength int           `json:"maxStringLength"`
}

func InspectFMSave(ctx context.Context, r io.Reader) (*FMSaveInspection, error) {
	return inspectFMSaveReader(contextReader{ctx, r})
}

func cliReadJSON(ctx context.Context, path string, stdin io.Reader, maxBytes int64, target any) error {
	data, err := cliReadInput(ctx, path, stdin, maxBytes)
	if err != nil {
		return err
	}
	return cliDecodeJSON(data, target)
}

func cliRequired(value, label string) error {
	if strings.TrimSpace(value) == "" {
		return cliFail(2, "%s is required", label)
	}
	return nil
}

func cliBody(ctx context.Context, o cliOptions, stdin io.Reader) ([]byte, error) {
	if err := cliRequired(o.body, "--body JSON or @path"); err != nil {
		return nil, err
	}
	if strings.HasPrefix(o.body, "@") {
		return cliReadInput(ctx, strings.TrimPrefix(o.body, "@"), stdin, o.maxBytes)
	}
	if !strings.HasPrefix(strings.TrimSpace(o.body), "{") && !strings.HasPrefix(strings.TrimSpace(o.body), "[") {
		return nil, cliFail(2, "body must be JSON or @path")
	}
	return []byte(o.body), nil
}

func cliDecodeJSON(data []byte, target any) error {
	decoder := newJSONDecoder(bytes.NewReader(data))
	if err := decoder.Decode(target); err != nil {
		return cliFail(3, "JSON: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return cliFail(3, "expected exactly one JSON value")
	}
	return nil
}

func newJSONDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.UseNumber()
	d.DisallowUnknownFields()
	return d
}

func cliString(v any) string { return fmt.Sprint(v) }
