package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	apperrors "api/errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
)

// S3Storage provides S3-based storage for datasets
type S3Storage struct {
	client     *minio.Client
	bucketName string
	fallback   StorageInterface
	// Connection pool for async operations
	operationsChan chan storageOperation
	workerPool     sync.WaitGroup
	shutdown       chan struct{}
	formatCache    map[string]string
	formatCacheMu  sync.RWMutex
}

type storageOperation struct {
	opType     string // "store", "retrieve", "delete"
	datasetID  string
	data       *DatasetData
	resultChan chan storageResult
}

type storageResult struct {
	data DatasetData
	err  error
}

// CreateS3Storage creates a new S3 storage instance with fallback
func CreateS3Storage(endpoint, accessKey, secretKey, bucketName string, useSSL bool, fallback StorageInterface) *S3Storage {
	const workerPoolSize = 10    // Increased worker pool size
	const operationsBuffer = 200 // Increased buffer size

	// Optimized MinIO client configuration
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: "us-east-1", // Set a default region

		// Custom transport for better performance
		Transport: &http.Transport{
			MaxIdleConns:          100,              // Increased connection pool
			MaxIdleConnsPerHost:   20,               // More connections per host
			IdleConnTimeout:       90 * time.Second, // Keep connections alive longer
			DisableKeepAlives:     false,            // Enable keep-alives
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	})
	if err != nil {
		LogWarn("Warning: Failed to create S3 client: %v. Using fallback storage.", err)
		return &S3Storage{fallback: fallback}
	}

	ctx := context.Background()

	// Check if bucket exists (this tests authentication)
	LogDebug("Testing S3 connection by checking bucket existence: %s", bucketName)
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		LogWarn("Warning: S3 bucket check failed - %v. Using fallback storage.", err)
		return &S3Storage{fallback: fallback}
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			LogWarn("Warning: Failed to create bucket %s: %v. Using fallback storage.", bucketName, err)
			return &S3Storage{fallback: fallback}
		}
		LogInfo("Created S3 bucket: %s", bucketName)
	} else {
		LogDebug("S3 bucket %s already exists", bucketName)
	}

	LogDebug("Successfully connected to S3 at %s with bucket %s", endpoint, bucketName)

	s3Storage := &S3Storage{
		client:         client,
		bucketName:     bucketName,
		fallback:       fallback,
		operationsChan: make(chan storageOperation, operationsBuffer),
		shutdown:       make(chan struct{}),
		formatCache:    make(map[string]string),
	}

	// Start optimized worker pool
	for i := 0; i < workerPoolSize; i++ {
		s3Storage.workerPool.Add(1)
		go s3Storage.asyncWorker()
	}

	LogDebug("S3 storage initialized with %d workers and %d operation buffer", workerPoolSize, operationsBuffer)
	return s3Storage
}

// asyncWorker processes storage operations asynchronously
func (s *S3Storage) asyncWorker() {
	defer s.workerPool.Done()

	for {
		select {
		case op := <-s.operationsChan:
			switch op.opType {
			case "store":
				err := s.storeSync(op.datasetID, *op.data)
				op.resultChan <- storageResult{err: err}
			case "retrieve":
				data, err := s.retrieveSync(op.datasetID)
				op.resultChan <- storageResult{data: data, err: err}
			case "delete":
				err := s.deleteSync(op.datasetID)
				op.resultChan <- storageResult{err: err}
			}
		case <-s.shutdown:
			return
		}
	}
}

// Shutdown gracefully stops the async workers
func (s *S3Storage) Shutdown() {
	close(s.shutdown)
	s.workerPool.Wait()
	close(s.operationsChan)
}

// StoreAsync performs asynchronous storage operation
func (s *S3Storage) StoreAsync(datasetID string, data DatasetData) <-chan error {
	resultChan := make(chan storageResult, 1)
	errorChan := make(chan error, 1)

	select {
	case s.operationsChan <- storageOperation{
		opType:     "store",
		datasetID:  datasetID,
		data:       &data,
		resultChan: resultChan,
	}:
		go func() {
			result := <-resultChan
			errorChan <- result.err
			close(errorChan)
		}()
		return errorChan
	default:
		// If channel is full, fall back to synchronous operation
		go func() {
			errorChan <- s.Store(datasetID, data)
			close(errorChan)
		}()
		return errorChan
	}
}

// Store is the public synchronous interface
func (s *S3Storage) Store(datasetID string, data DatasetData) error {
	return s.storeSync(datasetID, data)
}

// storeSync performs synchronous storage operation
func (s *S3Storage) storeSync(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.s3.store")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "s3"),
	)

	if s.client == nil {
		AddSpanEvent(ctx, "storage.fallback_to_memory")
		return s.fallback.Store(datasetID, data)
	}

	start := time.Now()

	// Add defensive programming for JSON marshaling to catch race conditions
	defer func() {
		if r := recover(); r != nil {
			RecordError(ctx, apperrors.WrapErrJSONMarshalPanic(r), "JSON marshal panic recovered")
			LogWarn("PANIC during JSON marshaling for dataset %s: %v", sanitizeForLogging(datasetID), r)
		}
	}()

	var fileData []byte
	var objectName string
	var contentType string

	if len(data.RawBytes) > 0 {
		fileData = data.RawBytes
		objectName = fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtProtobuf)
		contentType = "application/octet-stream"
	} else {
		var err error
		fileData, err = json.Marshal(data)
		if err != nil {
			RecordError(ctx, err, "Failed to marshal dataset data")
			LogWarn("JSON marshal error for dataset %s: %v", sanitizeForLogging(datasetID), err)
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		objectName = fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtJSON)
		contentType = "application/json"
	}

	reader := bytes.NewReader(fileData)

	SetSpanAttributes(ctx,
		attribute.String("s3.bucket", s.bucketName),
		attribute.String("s3.object", objectName),
		attribute.Int("data.size_bytes", len(fileData)),
	)

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(fileData)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		RecordError(ctx, err, "Failed to store to S3")
		LogWarn("Warning: Failed to store to S3: %v. Using fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_store_failed"))
		return s.fallback.Store(datasetID, data)
	}

	RecordDBOperation(ctx, "store", "s3_datasets", time.Since(start), 1)
	s.setCachedFormat(datasetID, objectName)
	LogDebug("Stored dataset %s to S3", sanitizeForLogging(datasetID))
	return nil
}

// Retrieve is the public synchronous interface
func (s *S3Storage) Retrieve(datasetID string) (DatasetData, error) {
	return s.retrieveSync(datasetID)
}

func (s *S3Storage) cachedFormat(datasetID string) (string, bool) {
	s.formatCacheMu.RLock()
	defer s.formatCacheMu.RUnlock()
	format, ok := s.formatCache[datasetID]
	return format, ok
}

func (s *S3Storage) setCachedFormat(datasetID, objectName string) {
	_, ok := datasetIDFromStorageName(objectName)
	if !ok {
		return
	}
	s.formatCacheMu.Lock()
	defer s.formatCacheMu.Unlock()
	if s.formatCache == nil {
		s.formatCache = make(map[string]string)
	}
	s.formatCache[datasetID] = objectName
}

func (s *S3Storage) clearCachedFormat(datasetID string) {
	s.formatCacheMu.Lock()
	defer s.formatCacheMu.Unlock()
	delete(s.formatCache, datasetID)
}

func (s *S3Storage) getDatasetObject(ctx context.Context, datasetID string) (*minio.Object, string, error) {
	objectNames := []string{
		fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtProtobuf),
		fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtJSON),
		fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtJSONGzip),
	}

	if cachedObjectName, ok := s.cachedFormat(datasetID); ok {
		objectNames = append([]string{cachedObjectName}, objectNames...)
	}

	var lastErr error
	probeCount := 0
	tried := make(map[string]struct{}, len(objectNames))
	for _, objectName := range objectNames {
		if _, exists := tried[objectName]; exists {
			continue
		}
		tried[objectName] = struct{}{}
		probeCount++

		SetSpanAttributes(ctx,
			attribute.String("s3.bucket", s.bucketName),
			attribute.String("s3.object", objectName),
		)

		object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			lastErr = err
			continue
		}
		if _, statErr := object.Stat(); statErr != nil {
			lastErr = statErr
			if closeErr := object.Close(); closeErr != nil {
				LogWarn("Failed to close missing S3 object: %v", closeErr)
			}
			continue
		}

		s.setCachedFormat(datasetID, objectName)
		SetSpanAttributes(ctx,
			attribute.Int("s3.object_probe_count", probeCount),
			attribute.String("storage.object_format", datasetFormatFromStorageName(objectName)),
		)
		return object, objectName, nil
	}

	s.clearCachedFormat(datasetID)
	SetSpanAttributes(ctx, attribute.Int("s3.object_probe_count", probeCount))
	if lastErr == nil {
		lastErr = fmt.Errorf("dataset %s not found in S3", datasetID)
	}
	return nil, "", lastErr
}

// retrieveSync performs synchronous retrieval operation
func (s *S3Storage) retrieveSync(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.s3.retrieve")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("storage.type", "s3"),
	)

	if s.client == nil {
		AddSpanEvent(ctx, "storage.fallback_to_memory")
		return s.fallback.Retrieve(datasetID)
	}

	start := time.Now()
	object, objectName, err := s.getDatasetObject(ctx, datasetID)
	if err != nil {
		RecordError(ctx, err, "Failed to retrieve from S3")
		LogWarn("Warning: Failed to retrieve from S3: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_get_failed"))
		return s.fallback.Retrieve(datasetID)
	}
	isCompressed := strings.HasSuffix(objectName, datasetExtJSONGzip)
	isProtobuf := strings.HasSuffix(objectName, datasetExtProtobuf)
	objectFormat := datasetFormatFromStorageName(objectName)

	defer func() {
		if closeErr := object.Close(); closeErr != nil {
			LogWarn("Failed to close S3 object: %v", closeErr)
		}
	}()

	data, err := io.ReadAll(object)
	if err != nil {
		RecordError(ctx, err, "Failed to read S3 object data")
		LogWarn("Warning: Failed to read from S3 object: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_read_failed"))
		return s.fallback.Retrieve(datasetID)
	}

	SetSpanAttributes(ctx,
		attribute.Int("data.size_bytes", len(data)),
		attribute.Bool("data.compressed", isCompressed),
		attribute.String("storage.object_format", objectFormat),
	)

	// Decompress only if legacy compressed object was used
	var jsonData []byte
	if isCompressed {
		jsonData, err = decompressData(data)
		if err != nil {
			RecordError(ctx, err, "Failed to decompress S3 data")
			LogWarn("Warning: Failed to decompress S3 data: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "decompress_failed"))
			return s.fallback.Retrieve(datasetID)
		}
		SetSpanAttributes(ctx, attribute.Int("data.decompressed_size_bytes", len(jsonData)))
	} else {
		jsonData = data
	}

	var dataset DatasetData
	if isProtobuf {
		dataset.RawBytes = jsonData
	} else {
		if err := json.Unmarshal(jsonData, &dataset); err != nil {
			RecordError(ctx, err, "Failed to unmarshal S3 data")
			LogWarn("Warning: Failed to unmarshal S3 data: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "unmarshal_failed"))
			return s.fallback.Retrieve(datasetID)
		}
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(dataset.Players)))
	RecordDBOperation(ctx, "retrieve", "s3_datasets", time.Since(start), 1)
	LogInfo("Retrieved dataset %s from S3 using %s format", sanitizeForLogging(datasetID), objectFormat)
	return dataset, nil
}

// Delete is the public synchronous interface
func (s *S3Storage) Delete(datasetID string) error {
	return s.deleteSync(datasetID)
}

// deleteSync performs synchronous deletion operation
func (s *S3Storage) deleteSync(datasetID string) error {
	if s.client == nil {
		return s.fallback.Delete(datasetID)
	}

	ctx := context.Background()

	// Attempt to delete pb, uncompressed, and legacy compressed objects
	pbObject := fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtProtobuf)
	jsonObject := fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtJSON)
	gzObject := fmt.Sprintf("%s%s%s", datasetObjectPrefix, datasetID, datasetExtJSONGzip)

	if err := s.client.RemoveObject(ctx, s.bucketName, pbObject, minio.RemoveObjectOptions{}); err != nil {
		LogWarn("Warning: Failed to delete pb from S3: %v", err)
	}
	if err := s.client.RemoveObject(ctx, s.bucketName, jsonObject, minio.RemoveObjectOptions{}); err != nil {
		LogWarn("Warning: Failed to delete JSON object from S3: %v", err)
	}
	if err := s.client.RemoveObject(ctx, s.bucketName, gzObject, minio.RemoveObjectOptions{}); err != nil {
		LogWarn("Warning: Failed to delete legacy GZIP object from S3: %v", err)
	}

	if err := s.fallback.Delete(datasetID); err != nil {
		LogWarn("Warning: Failed to delete from fallback storage: %v", err)
		// Don't return error since S3 deletion succeeded
	}
	s.clearCachedFormat(datasetID)
	LogDebug("Deleted dataset %s from S3", sanitizeForLogging(datasetID))
	return nil
}

// List returns all dataset IDs stored in S3
func (s *S3Storage) List() ([]string, error) {
	if s.client == nil {
		return s.fallback.List()
	}

	ctx := context.Background()
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "datasets/",
		Recursive: true,
	})

	var ids []string
	seen := make(map[string]struct{})
	for object := range objectCh {
		if object.Err != nil {
			LogWarn("Warning: Error listing S3 objects: %v. Using fallback storage.", object.Err)
			return s.fallback.List()
		}

		id, ok := datasetIDFromStorageName(object.Key)
		if ok {
			ids = appendDatasetIDOnce(ids, seen, id)
		}
	}

	LogDebug("Listed %d datasets from S3", len(ids))
	return ids, nil
}

// CleanupOldDatasets removes datasets older than the specified duration from S3
func (s *S3Storage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	if s.client == nil {
		return s.fallback.CleanupOldDatasets(maxAge, excludeDatasets)
	}

	ctx := context.Background()
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "datasets/",
		Recursive: true,
	})

	cutoffTime := time.Now().Add(-maxAge)
	excludeSet := make(map[string]bool)
	for _, dataset := range excludeDatasets {
		excludeSet[dataset] = true
	}

	var deletedCount int
	for object := range objectCh {
		if object.Err != nil {
			LogWarn("Warning: Error listing S3 objects during cleanup: %v", object.Err)
			continue
		}

		datasetID, ok := datasetIDFromStorageName(object.Key)
		if !ok {
			continue
		}

		// Skip excluded datasets (like demo)
		if excludeSet[datasetID] {
			LogDebug("Skipping cleanup for excluded dataset: %s", sanitizeForLogging(datasetID))
			continue
		}

		// Check if object is older than cutoff time
		if object.LastModified.Before(cutoffTime) {
			LogDebug("Deleting old dataset: %s (last modified: %s)", sanitizeForLogging(datasetID), object.LastModified.Format(time.RFC3339))

			err := s.client.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				LogWarn("Warning: Failed to delete old dataset %s from S3: %v", sanitizeForLogging(datasetID), err)
				continue
			}

			deletedCount++
		}
	}

	LogDebug("Cleanup completed: deleted %d old datasets from S3", deletedCount)
	return nil
}
