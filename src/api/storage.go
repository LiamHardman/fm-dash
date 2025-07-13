package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	apperrors "api/errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
)

// compressData compresses data using gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		if closeErr := gz.Close(); closeErr != nil {
			log.Printf("Failed to close gzip writer: %v", closeErr)
		}
	}()

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Failed to close gzip reader: %v", closeErr)
		}
	}()

	return io.ReadAll(reader)
}

type StorageInterface interface {
	Store(datasetID string, data DatasetData) error
	Retrieve(datasetID string) (DatasetData, error)
	Delete(datasetID string) error
	List() ([]string, error)
	CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error
}

type DatasetData struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currency_symbol"`
}

type InMemoryStorage struct {
	data  map[string]DatasetData
	mutex sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]DatasetData),
	}
}

func (s *InMemoryStorage) Store(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.memory.store")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "memory"),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[datasetID] = data

	RecordDBOperation(ctx, "store", "datasets", 0, 1)
	return nil
}

func (s *InMemoryStorage) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.memory.retrieve")
	defer span.End()

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("storage.type", "memory"),
	)

	start := time.Now()
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, exists := s.data[datasetID]
	if !exists {
		err := apperrors.WrapErrDatasetNotFound(datasetID)
		RecordError(ctx, err, "Dataset not found in memory storage")
		return DatasetData{}, err
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(data.Players)))
	RecordDBOperation(ctx, "retrieve", "datasets", time.Since(start), 1)
	return data, nil
}

func (s *InMemoryStorage) Delete(datasetID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, datasetID)
	return nil
}

func (s *InMemoryStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *InMemoryStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	// In-memory storage doesn't persist data, so no cleanup needed
	log.Println("CleanupOldDatasets called on in-memory storage - no action needed")
	return nil
}

type S3Storage struct {
	client     *minio.Client
	bucketName string
	fallback   StorageInterface
	// Connection pool for async operations
	operationsChan chan storageOperation
	workerPool     sync.WaitGroup
	shutdown       chan struct{}
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

// NewS3Storage creates a new S3 storage instance with fallback
func NewS3Storage(endpoint, accessKey, secretKey, bucketName string, useSSL bool, fallback StorageInterface) *S3Storage {
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
		log.Printf("Warning: Failed to create S3 client: %v. Using fallback storage.", err)
		return &S3Storage{fallback: fallback}
	}

	ctx := context.Background()

	// Check if bucket exists (this tests authentication)
	LogDebug("Testing S3 connection by checking bucket existence: %s", bucketName)
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Printf("Warning: S3 bucket check failed - %v. Using fallback storage.", err)
		return &S3Storage{fallback: fallback}
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Warning: Failed to create bucket %s: %v. Using fallback storage.", bucketName, err)
			return &S3Storage{fallback: fallback}
		}
		log.Printf("Created S3 bucket: %s", bucketName)
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
			log.Printf("PANIC during JSON marshaling for dataset %s: %v", sanitizeForLogging(datasetID), r)
		}
	}()

	jsonData, err := json.Marshal(data)
	if err != nil {
		RecordError(ctx, err, "Failed to marshal dataset data")
		log.Printf("JSON marshal error for dataset %s: %v", sanitizeForLogging(datasetID), err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Compress the JSON data
	compressedData, err := compressData(jsonData)
	if err != nil {
		RecordError(ctx, err, "Failed to compress dataset data")
		return fmt.Errorf("failed to compress data: %w", err)
	}

	objectName := fmt.Sprintf("datasets/%s.json.gz", datasetID)
	reader := bytes.NewReader(compressedData)

	SetSpanAttributes(ctx,
		attribute.String("s3.bucket", s.bucketName),
		attribute.String("s3.object", objectName),
		attribute.Int("data.size_bytes", len(jsonData)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
		attribute.Float64("compression.ratio", float64(len(jsonData))/float64(len(compressedData))),
	)

	_, err = s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(compressedData)), minio.PutObjectOptions{
		ContentType: "application/gzip",
		UserMetadata: map[string]string{
			"compression":   "gzip",
			"original-size": fmt.Sprintf("%d", len(jsonData)),
		},
	})
	if err != nil {
		RecordError(ctx, err, "Failed to store to S3")
		log.Printf("Warning: Failed to store to S3: %v. Using fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_store_failed"))
		return s.fallback.Store(datasetID, data)
	}

	RecordDBOperation(ctx, "store", "s3_datasets", time.Since(start), 1)
	LogDebug("Stored dataset %s to S3", sanitizeForLogging(datasetID))
	return nil
}

// Retrieve is the public synchronous interface
func (s *S3Storage) Retrieve(datasetID string) (DatasetData, error) {
	return s.retrieveSync(datasetID)
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

	// Try compressed file first, then fall back to uncompressed
	objectName := fmt.Sprintf("datasets/%s.json.gz", datasetID)
	isCompressed := true

	SetSpanAttributes(ctx,
		attribute.String("s3.bucket", s.bucketName),
		attribute.String("s3.object", objectName),
	)

	start := time.Now()
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		// Try uncompressed file
		objectName = fmt.Sprintf("datasets/%s.json", datasetID)
		isCompressed = false
		SetSpanAttributes(ctx, attribute.String("s3.object", objectName))
		object, err = s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			RecordError(ctx, err, "Failed to retrieve from S3")
			log.Printf("Warning: Failed to retrieve from S3: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_get_failed"))
			return s.fallback.Retrieve(datasetID)
		}
	}
	defer func() {
		if closeErr := object.Close(); closeErr != nil {
			log.Printf("Failed to close S3 object: %v", closeErr)
		}
	}()

	data, err := io.ReadAll(object)
	if err != nil {
		RecordError(ctx, err, "Failed to read S3 object data")
		log.Printf("Warning: Failed to read from S3 object: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "s3_read_failed"))
		return s.fallback.Retrieve(datasetID)
	}

	SetSpanAttributes(ctx,
		attribute.Int("data.size_bytes", len(data)),
		attribute.Bool("data.compressed", isCompressed),
	)

	// Decompress if necessary
	var jsonData []byte
	if isCompressed {
		jsonData, err = decompressData(data)
		if err != nil {
			RecordError(ctx, err, "Failed to decompress S3 data")
			log.Printf("Warning: Failed to decompress S3 data: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "decompress_failed"))
			return s.fallback.Retrieve(datasetID)
		}
		SetSpanAttributes(ctx, attribute.Int("data.decompressed_size_bytes", len(jsonData)))
	} else {
		jsonData = data
	}

	var dataset DatasetData
	if err := json.Unmarshal(jsonData, &dataset); err != nil {
		RecordError(ctx, err, "Failed to unmarshal S3 data")
		log.Printf("Warning: Failed to unmarshal S3 data: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "unmarshal_failed"))
		return s.fallback.Retrieve(datasetID)
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(dataset.Players)))
	RecordDBOperation(ctx, "retrieve", "s3_datasets", time.Since(start), 1)
	log.Printf("Retrieved dataset %s from S3", sanitizeForLogging(datasetID))
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

	objectName := fmt.Sprintf("datasets/%s.json", datasetID)
	ctx := context.Background()

	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Warning: Failed to delete from S3: %v. Using fallback storage.", err)
		return s.fallback.Delete(datasetID)
	}

	if err := s.fallback.Delete(datasetID); err != nil {
		log.Printf("Warning: Failed to delete from fallback storage: %v", err)
		// Don't return error since S3 deletion succeeded
	}
	log.Printf("Deleted dataset %s from S3", sanitizeForLogging(datasetID))
	return nil
}

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
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Warning: Error listing S3 objects: %v. Using fallback storage.", object.Err)
			return s.fallback.List()
		}

		if strings.HasSuffix(object.Key, ".json") {
			id := strings.TrimPrefix(object.Key, "datasets/")
			id = strings.TrimSuffix(id, ".json")
			ids = append(ids, id)
		}
	}

	log.Printf("Listed %d datasets from S3", len(ids))
	return ids, nil
}

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
			log.Printf("Warning: Error listing S3 objects during cleanup: %v", object.Err)
			continue
		}

		if !strings.HasSuffix(object.Key, ".json") {
			continue
		}

		// Extract dataset ID from object key
		datasetID := strings.TrimPrefix(object.Key, "datasets/")
		datasetID = strings.TrimSuffix(datasetID, ".json")

		// Skip excluded datasets (like demo)
		if excludeSet[datasetID] {
			log.Printf("Skipping cleanup for excluded dataset: %s", sanitizeForLogging(datasetID))
			continue
		}

		// Check if object is older than cutoff time
		if object.LastModified.Before(cutoffTime) {
			log.Printf("Deleting old dataset: %s (last modified: %s)", sanitizeForLogging(datasetID), object.LastModified.Format(time.RFC3339))

			err := s.client.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete old dataset %s from S3: %v", sanitizeForLogging(datasetID), err)
				continue
			}

			deletedCount++
		}
	}

	LogDebug("Cleanup completed: deleted %d old datasets from S3", deletedCount)
	return nil
}

// getFaceImage retrieves a face image from S3 and writes it to the response writer
func (s *S3Storage) getFaceImage(ctx context.Context, filename string, w http.ResponseWriter) error {
	if s.client == nil {
		return apperrors.ErrS3ClientNotAvailable
	}

	// Get the faces bucket name from environment, default to the main bucket + "/faces" prefix
	facesBucketName := os.Getenv("S3_FACES_BUCKET")
	if facesBucketName == "" {
		facesBucketName = s.bucketName // Use same bucket as datasets
	}

	// Construct the object key for faces
	objectKey := "faces/" + filename

	// If we're using a separate faces bucket, don't add the prefix
	if facesBucketName != s.bucketName {
		objectKey = filename
	}

	// Get object from S3
	reader, err := s.client.GetObject(ctx, facesBucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get face image from S3: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Failed to close S3 reader: %v", closeErr)
		}
	}()

	// Copy the image data to the response writer
	_, err = io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("failed to write face image to response: %w", err)
	}

	return nil
}

// getTeamLogo retrieves a team logo image from S3 and writes it to the response writer
func (s *S3Storage) getTeamLogo(ctx context.Context, filename string, w http.ResponseWriter) error {
	if s.client == nil {
		return apperrors.ErrS3ClientNotAvailable
	}

	// Get the logos bucket name from environment, default to the main bucket + "/logos/clubs" prefix
	logosBucketName := os.Getenv("S3_LOGOS_BUCKET")
	if logosBucketName == "" {
		logosBucketName = s.bucketName // Use same bucket as datasets
	}

	// Construct the object key for logos
	objectKey := "logos/Clubs/Normal/Normal/" + filename

	// If we're using a separate logos bucket, don't add the prefix
	if logosBucketName != s.bucketName {
		objectKey = filename
	}

	// Get object from S3
	reader, err := s.client.GetObject(ctx, logosBucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get team logo from S3: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Failed to close S3 reader: %v", closeErr)
		}
	}()

	// Copy the image data to the response writer
	_, err = io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("failed to write team logo to response: %w", err)
	}

	return nil
}

// LocalFileStorage stores datasets as JSON files in a local directory
type LocalFileStorage struct {
	datasetDir string
	mutex      sync.RWMutex
}

func NewLocalFileStorage(datasetDir string) (*LocalFileStorage, error) {
	// Create datasets directory if it doesn't exist
	if err := os.MkdirAll(datasetDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create datasets directory %s: %w", datasetDir, err)
	}

	log.Printf("Initialized local file storage at: %s", datasetDir)
	return &LocalFileStorage{
		datasetDir: datasetDir,
	}, nil
}

func (s *LocalFileStorage) Store(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.local_file.store")
	defer span.End()

	// Validate dataset ID format
	if err := validateID(datasetID, 100); err != nil {
		err := apperrors.WrapErrInvalidDatasetID(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Invalid dataset ID format")
		return err
	}

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", sanitizeForLogging(datasetID)),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "local_file"),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Safely construct the filename to prevent path injection
	filename, err := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s.json.gz", datasetID))
	if err != nil {
		err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Path validation failed")
		return err
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		RecordError(ctx, err, "Failed to marshal dataset data")
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Compress the data
	compressedData, err := compressData(jsonData)
	if err != nil {
		RecordError(ctx, err, "Failed to compress dataset data")
		return fmt.Errorf("failed to compress data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, compressedData, 0o600); err != nil {
		RecordError(ctx, err, "Failed to write dataset file")
		return fmt.Errorf("failed to write dataset file: %w", err)
	}

	SetSpanAttributes(ctx,
		attribute.String("file.path", sanitizeForLogging(filename)),
		attribute.Int("data.size_bytes", len(jsonData)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
	)

	log.Printf("Stored dataset %s to local file: %s", sanitizeForLogging(datasetID), sanitizeForLogging(filename))
	return nil
}

func (s *LocalFileStorage) Retrieve(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.local_file.retrieve")
	defer span.End()

	// Validate dataset ID format
	if err := validateID(datasetID, 100); err != nil {
		err := apperrors.WrapErrInvalidDatasetID(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Invalid dataset ID format")
		return DatasetData{}, err
	}

	SetSpanAttributes(ctx,
		attribute.String("dataset.id", sanitizeForLogging(datasetID)),
		attribute.String("storage.type", "local_file"),
	)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Try compressed file first - safely construct the filename
	filename, err := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s.json.gz", datasetID))
	if err != nil {
		err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Path validation failed")
		return DatasetData{}, err
	}
	isCompressed := true

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := apperrors.WrapErrDatasetNotFound(sanitizeForLogging(datasetID))
		RecordError(ctx, err, "Path validation failed")
		return DatasetData{}, err
	}

	// Read and decompress the file
	data, err := os.ReadFile(filename)
	if err != nil {
		err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Path validation failed")
		return DatasetData{}, err
	}

	SetSpanAttributes(ctx,
		attribute.String("file.path", sanitizeForLogging(filename)),
		attribute.Int("data.size_bytes", len(data)),
		attribute.Bool("data.compressed", isCompressed),
	)

	// Decompress if necessary
	var jsonData []byte
	if isCompressed {
		jsonData, err = decompressData(data)
		if err != nil {
			RecordError(ctx, err, "Failed to decompress dataset data")
			return DatasetData{}, fmt.Errorf("failed to decompress data: %w", err)
		}
		SetSpanAttributes(ctx, attribute.Int("data.decompressed_size_bytes", len(jsonData)))
	} else {
		jsonData = data
	}

	var dataset DatasetData
	if err := json.Unmarshal(jsonData, &dataset); err != nil {
		RecordError(ctx, err, "Failed to unmarshal dataset data")
		return DatasetData{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(dataset.Players)))
	log.Printf("Retrieved dataset %s from local file: %s", sanitizeForLogging(datasetID), sanitizeForLogging(filename))
	return dataset, nil
}

func (s *LocalFileStorage) Delete(datasetID string) error {
	// Validate dataset ID format
	if err := validateID(datasetID, 100); err != nil {
		return fmt.Errorf("invalid dataset ID: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Try to delete both compressed and uncompressed versions - safely construct paths
	compressedFile, err := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s.json.gz", datasetID))
	if err != nil {
		return fmt.Errorf("invalid file path for compressed file: %w", err)
	}

	uncompressedFile, err := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s.json", datasetID))
	if err != nil {
		return fmt.Errorf("invalid file path for uncompressed file: %w", err)
	}

	// Don't treat "file not found" as an error
	if err := os.Remove(compressedFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove compressed file %s: %v", sanitizeForLogging(compressedFile), err)
	}
	if err := os.Remove(uncompressedFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove uncompressed file %s: %v", sanitizeForLogging(uncompressedFile), err)
	}

	log.Printf("Deleted dataset %s from local storage", sanitizeForLogging(datasetID))
	return nil
}

func (s *LocalFileStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entries, err := os.ReadDir(s.datasetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read datasets directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".json.gz") {
			id := strings.TrimSuffix(name, ".json.gz")
			ids = append(ids, id)
		} else if strings.HasSuffix(name, ".json") {
			id := strings.TrimSuffix(name, ".json")
			// Only add if we don't already have the compressed version
			found := false
			for _, existingID := range ids {
				if existingID == id {
					found = true
					break
				}
			}
			if !found {
				ids = append(ids, id)
			}
		}
	}

	log.Printf("Listed %d datasets from local storage", len(ids))
	return ids, nil
}

func (s *LocalFileStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entries, err := os.ReadDir(s.datasetDir)
	if err != nil {
		return fmt.Errorf("failed to read datasets directory: %w", err)
	}

	cutoffTime := time.Now().Add(-maxAge)
	excludeSet := make(map[string]bool)
	for _, dataset := range excludeDatasets {
		excludeSet[dataset] = true
	}

	var deletedCount int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".json.gz") {
			continue
		}

		// Extract dataset ID
		var datasetID string
		if strings.HasSuffix(name, ".json.gz") {
			datasetID = strings.TrimSuffix(name, ".json.gz")
		} else {
			datasetID = strings.TrimSuffix(name, ".json")
		}

		// Skip excluded datasets
		if excludeSet[datasetID] {
			log.Printf("Skipping cleanup for excluded dataset: %s", sanitizeForLogging(datasetID))
			continue
		}

		// Get file info to check modification time - safely construct the file path
		filePath, err := validateAndJoinPath(s.datasetDir, name)
		if err != nil {
			log.Printf("Warning: Invalid file path for %s: %v", sanitizeForLogging(name), err)
			continue
		}

		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Warning: Failed to get file info for %s: %v", sanitizeForLogging(filePath), err)
			continue
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
			log.Printf("Deleting old dataset file: %s (last modified: %s)", sanitizeForLogging(name), info.ModTime().Format(time.RFC3339))

			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: Failed to delete old dataset file %s: %v", sanitizeForLogging(filePath), err)
				continue
			}

			deletedCount++
		}
	}

	log.Printf("Cleanup completed: deleted %d old dataset files from local storage", deletedCount)
	return nil
}

// HybridStorage combines in-memory storage with local file fallback
type HybridStorage struct {
	memory StorageInterface
	local  StorageInterface
}

func NewHybridStorage(datasetDir string) (*HybridStorage, error) {
	memory := NewInMemoryStorage()
	local, err := NewLocalFileStorage(datasetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file storage: %w", err)
	}

	log.Println("Initialized hybrid storage (in-memory + local file fallback)")
	return &HybridStorage{
		memory: memory,
		local:  local,
	}, nil
}

func (s *HybridStorage) Store(datasetID string, data DatasetData) error {
	// Store in both memory and local file
	if err := s.memory.Store(datasetID, data); err != nil {
		log.Printf("Warning: Failed to store dataset %s in memory: %v", sanitizeForLogging(datasetID), err)
	}

	return s.local.Store(datasetID, data)
}

func (s *HybridStorage) Retrieve(datasetID string) (DatasetData, error) {
	// Try memory first for fastest access
	data, err := s.memory.Retrieve(datasetID)
	if err == nil {
		log.Printf("Retrieved dataset %s from memory", sanitizeForLogging(datasetID))
		return data, nil
	}

	// Try persistent storage (critical for multi-replica consistency)
	log.Printf("Dataset %s not found in memory, checking persistent storage", sanitizeForLogging(datasetID))
	data, err = s.local.Retrieve(datasetID)
	if err != nil {
		return DatasetData{}, err
	}

	// Store in memory for future access (warm up the cache)
	go func() {
		if storeErr := s.memory.Store(datasetID, data); storeErr != nil {
			log.Printf("Warning: Failed to cache dataset %s in memory: %v", sanitizeForLogging(datasetID), storeErr)
		} else {
			log.Printf("Successfully cached dataset %s in memory for future access", sanitizeForLogging(datasetID))
		}
	}()

	log.Printf("Retrieved dataset %s from persistent storage and cached in memory", sanitizeForLogging(datasetID))
	return data, nil
}

func (s *HybridStorage) Delete(datasetID string) error {
	// Delete from both memory and local file
	if err := s.memory.Delete(datasetID); err != nil {
		// Ignore error since it might not exist in memory
		log.Printf("Note: Dataset %s not found in memory during deletion", sanitizeForLogging(datasetID))
	}
	return s.local.Delete(datasetID)
}

func (s *HybridStorage) List() ([]string, error) {
	// Get datasets from both memory and local storage, then merge
	memoryIDs, _ := s.memory.List() // Ignore error since memory might be empty
	localIDs, err := s.local.List()
	if err != nil {
		return memoryIDs, err
	}

	// Merge and deduplicate
	idSet := make(map[string]bool)
	for _, id := range memoryIDs {
		idSet[id] = true
	}
	for _, id := range localIDs {
		idSet[id] = true
	}

	var allIDs []string
	for id := range idSet {
		allIDs = append(allIDs, id)
	}

	return allIDs, nil
}

func (s *HybridStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
	// Cleanup both memory and local storage
	if err := s.memory.CleanupOldDatasets(maxAge, excludeDatasets); err != nil {
		log.Printf("Warning: Memory cleanup failed: %v", err)
		// Continue with local cleanup even if memory cleanup fails
	}
	return s.local.CleanupOldDatasets(maxAge, excludeDatasets)
}

func InitializeStorage() StorageInterface {
	inMemory := NewInMemoryStorage()

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		log.Println("No S3 endpoint configured. Using hybrid storage (in-memory + local file fallback).")

		// Use configurable datasets directory, default to "./datasets"
		datasetDir := os.Getenv("DATASETS_DIR")
		if datasetDir == "" {
			datasetDir = "./datasets"
		}

		hybrid, err := NewHybridStorage(datasetDir)
		if err != nil {
			log.Printf("Failed to initialize hybrid storage: %v. Falling back to in-memory only.", err)
			return inMemory
		}

		return hybrid
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	useSSL := strings.ToLower(os.Getenv("S3_USE_SSL")) == "true"

	if accessKey == "" || secretKey == "" {
		log.Println("S3 credentials not provided. Using hybrid storage (in-memory + local file fallback).")

		// Use configurable datasets directory, default to "./datasets"
		datasetDir := os.Getenv("DATASETS_DIR")
		if datasetDir == "" {
			datasetDir = "./datasets"
		}

		hybrid, err := NewHybridStorage(datasetDir)
		if err != nil {
			log.Printf("Failed to initialize hybrid storage: %v. Falling back to in-memory only.", err)
			return inMemory
		}

		return hybrid
	}

	// Debug logging (only show first few chars for security)
	// Log configuration without sensitive credentials
	LogDebug("S3 Config: endpoint=%s, useSSL=%v, credentials_provided=%t",
		s3Endpoint,
		useSSL,
		accessKey != "" && secretKey != "")

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "v2fmdash"
	}
	s3Storage := NewS3Storage(s3Endpoint, accessKey, secretKey, bucketName, useSSL, inMemory)

	LogDebug("Initialized S3 storage with in-memory fallback")
	return s3Storage
}
