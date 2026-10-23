package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type cliHeaders []string

func (h *cliHeaders) String() string     { return "" }
func (h *cliHeaders) Set(v string) error { *h = append(*h, v); return nil }

func cliRemoteOperation(command string, o cliOptions) (string, string, bool) {
	paths := map[string][2]string{
		"scout report": {"/api/scout-report/{dataset}", "POST"}, "scout stars": {"/api/scout-report-stars/{dataset}", "GET"},
		"scout reports": {"/api/scout-reports/{dataset}", "GET"}, "assistant ask": {"/api/chatbot/{dataset}", "POST"},
		"recruitment who-to-sign": {"/api/who-to-sign/{dataset}", "POST"},
		"cache status":            {"/api/cache-status", "GET"}, "cache clear": {"/api/cache/" + url.PathEscape(o.name), "DELETE"},
		"debug player": {"/api/debug/player", "GET"}, "club-match resolve": {"/api/club-logo/resolve", "GET"},
		"club-match overrides": {"/api/club-logo/overrides", "GET"},
	}
	p, ok := paths[command]
	return p[0], p[1], ok
}

// cliAPI provides explicit access to every existing backend operation, including
// provider-backed scouting, without embedding a server in the CLI process.
func cliAPI(ctx context.Context, args []string, o cliOptions, stdin io.Reader) (any, error) {
	if len(args) != 1 || !strings.HasPrefix(args[0], "/api/") {
		return nil, cliFail(2, "api requires one /api/... path")
	}
	base, err := url.Parse(o.baseURL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, cliFail(2, "base-url must be an HTTP(S) URL")
	}
	relative, err := url.Parse(args[0])
	if err != nil || relative.IsAbs() || relative.Host != "" {
		return nil, cliFail(2, "invalid API path")
	}
	endpoint := base.ResolveReference(relative)
	if o.query != "" {
		values, err := url.ParseQuery(o.query)
		if err != nil {
			return nil, cliFail(2, "query: %v", err)
		}
		endpoint.RawQuery = values.Encode()
	}
	var body []byte
	contentType := "application/json"
	if o.body != "" {
		body, err = cliBody(ctx, o, stdin)
		if err != nil {
			return nil, err
		}
		if !json.Valid(body) {
			return nil, cliFail(3, "body is not valid JSON")
		}
	}
	if o.upload != "" {
		if o.body != "" {
			return nil, cliFail(2, "upload and body cannot be combined")
		}
		file, err := cliReadInput(ctx, o.upload, stdin, o.maxBytes)
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		field := "playerFile"
		if relative.Path == "/api/fm-save-import" {
			field = "saveFile"
		}
		part, err := writer.CreateFormFile(field, filepath.Base(o.upload))
		if err != nil {
			return nil, err
		}
		if _, err = part.Write(file); err != nil {
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		body = buffer.Bytes()
		contentType = writer.FormDataContentType()
	}
	method := strings.ToUpper(o.method)
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
	default:
		return nil, cliFail(2, "unsupported HTTP method")
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", contentType)
	for _, header := range o.headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, cliFail(2, "header must have NAME:VALUE form")
		}
		value := strings.TrimSpace(parts[1])
		if strings.HasPrefix(value, "@") {
			value = os.Getenv(value[1:])
			if value == "" {
				return nil, cliFail(2, "header environment variable is empty")
			}
		}
		request.Header.Set(strings.TrimSpace(parts[0]), value)
	}
	if o.id != "" {
		request.Header.Set("If-Match", o.id)
	}
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return nil, cliFail(6, "API request: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(response.Body, o.maxBytes+1))
	if err != nil {
		return nil, cliFail(6, "API response: %v", err)
	}
	if int64(len(data)) > o.maxBytes {
		return nil, cliFail(6, "API response exceeds max-bytes")
	}
	var value any
	if len(data) > 0 {
		if json.Valid(data) {
			if err = cliDecodeJSON(data, &value); err != nil {
				return nil, err
			}
		} else {
			value = string(data)
		}
	}
	result := map[string]any{"httpStatus": response.StatusCode, "etag": response.Header.Get("ETag"), "data": value}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		code := 6
		switch response.StatusCode {
		case 400, 422:
			code = 3
		case 404:
			code = 5
		case 409, 412, 428:
			code = 7
		}
		return result, cliFail(code, "API returned HTTP %d", response.StatusCode)
	}
	return result, nil
}
