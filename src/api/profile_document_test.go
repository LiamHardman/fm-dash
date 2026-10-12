package main

import (
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"sync"
	"testing"
	"time"
)

type unavailableProfileStorage struct{ ConfigStorage }

func (unavailableProfileStorage) RetrieveConfig(string) ([]byte, error) {
	return nil, errors.New("storage offline")
}

func TestJSONFactoryReadsExistingProtobufDatasets(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATASETS_DIR", dir)
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("USE_PROTOBUF", "false")
	backend, err := CreateLocalFileStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = CreateProtobufStorage(backend).Store("rollback-factory", DatasetData{CurrencySymbol: "GBP", Players: []Player{{UID: 7, Name: "Preserved"}}}); err != nil {
		t.Fatal(err)
	}
	data, err := InitializeStorage(context.Background()).Retrieve("rollback-factory")
	if err != nil || data.CurrencySymbol != "GBP" || len(data.Players) != 1 || data.Players[0].Name != "Preserved" {
		t.Fatalf("rollback retrieval failed: %+v %v", data, err)
	}
}

func TestS3ProfileConditionalWrites(t *testing.T) {
	var mu sync.Mutex
	var stored []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		failure := func(status int, code string) {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(status)
			_, _ = w.Write([]byte("<Error><Code>" + code + "</Code><Message>fixture</Message></Error>"))
		}
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			if stored == nil {
				failure(404, "NoSuchKey")
				return
			}
			w.Header().Set("ETag", `"storage-etag"`)
			w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			_, _ = w.Write(stored)
		case http.MethodPut:
			if stored == nil && r.Header.Get("If-None-Match") != "*" {
				t.Error("create must be conditional")
			}
			if stored != nil && r.Header.Get("If-Match") != `"storage-etag"` {
				failure(412, "PreconditionFailed")
				return
			}
			var body io.Reader = r.Body
			// The SDK signs streaming uploads using AWS chunk framing.
			if strings.HasPrefix(r.Header.Get("X-Amz-Content-Sha256"), "STREAMING-") || r.Header.Get("X-Amz-Decoded-Content-Length") != "" {
				body = httputil.NewChunkedReader(r.Body)
			}
			stored, _ = io.ReadAll(body)
			w.Header().Set("ETag", `"storage-etag"`)
		}
	}))
	defer server.Close()
	client, err := minio.New(strings.TrimPrefix(server.URL, "http://"), &minio.Options{Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	s := &s3ConfigStorage{client: client, bucketName: "profile-test"}
	first := []byte(`{"version":1}`)
	if err = s.CompareAndSwapConfig("test.json", first, `"missing"`); err != nil {
		t.Fatal(err)
	}
	if err = s.CompareAndSwapConfig("test.json", []byte(`{"version":2}`), profileRevision(first)); err != nil {
		t.Fatal(err)
	}
	if err = s.CompareAndSwapConfig("test.json", first, profileRevision(first)); !errors.Is(err, errProfileConflict) {
		t.Fatalf("stale write: %v", err)
	}
}

func TestProfileStorageOutageIsNotEmptySuccess(t *testing.T) {
	previous := configStorage
	t.Cleanup(func() { configStorage = previous })
	configStorage = unavailableProfileStorage{}
	for _, handler := range []http.HandlerFunc{shortlistsHandler, savedSearchesHandler, progressionProjectsHandler} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("got %d, want 503", w.Code)
		}
	}
}

func TestProfileRevisionRejectsStaleWritesAcrossStorageInstances(t *testing.T) {
	dir := t.TempDir()
	stores := []*localConfigStorage{{dir: dir}, {dir: dir}}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, store := range stores {
		wg.Add(1)
		go func(s *localConfigStorage) {
			defer wg.Done()
			errs <- s.CompareAndSwapConfig("test.json", []byte(`{"version":1}`), `"missing"`)
		}(store)
	}
	wg.Wait()
	close(errs)
	success, conflict := 0, 0
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, errProfileConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}

func TestProfileHTTPRevisionContract(t *testing.T) {
	previous := configStorage
	configStorage = &localConfigStorage{dir: t.TempDir()}
	t.Cleanup(func() { configStorage = previous })
	get := httptest.NewRecorder()
	savedSearchesHandler(get, httptest.NewRequest(http.MethodGet, "/", nil))
	etag := get.Header().Get("ETag")
	for i, expectedCode := range []int{http.StatusOK, http.StatusPreconditionFailed} {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"version":1,"searches":[]}`))
		r.Header.Set("If-Match", etag)
		w := httptest.NewRecorder()
		savedSearchesHandler(w, r)
		if w.Code != expectedCode {
			t.Fatalf("write %d: got %d %s", i, w.Code, w.Body.String())
		}
	}
}
