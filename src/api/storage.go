package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
)

// compressData compresses data using gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()
	
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
	defer reader.Close()
	
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
		RecordError(ctx, fmt.Errorf("dataset %s not found", datasetID), "Dataset not found in memory storage")
		return DatasetData{}, fmt.Errorf("dataset %s not found", datasetID)
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

type MinIOStorage struct {
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

func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool, fallback StorageInterface) (*MinIOStorage, error) {
	const workerPoolSize = 5
	const operationsBuffer = 100
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Warning: Failed to create MinIO client: %v. Using fallback storage.", err)
		return &MinIOStorage{fallback: fallback}, nil
	}

	ctx := context.Background()
	
	// Check if bucket exists (this tests authentication)
	log.Printf("Testing MinIO connection by checking bucket existence: %s", bucketName)
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Printf("Warning: MinIO bucket check failed - %v. Using fallback storage.", err)
		return &MinIOStorage{fallback: fallback}, nil
	}
	
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Warning: Failed to create bucket %s: %v. Using fallback storage.", bucketName, err)
			return &MinIOStorage{fallback: fallback}, nil
		}
		log.Printf("Created MinIO bucket: %s", bucketName)
	} else {
		log.Printf("MinIO bucket %s already exists", bucketName)
	}

	log.Printf("Successfully connected to MinIO at %s with bucket %s", endpoint, bucketName)
	
	// Initialize async storage with worker pool
	storage := &MinIOStorage{
		client:         client,
		bucketName:     bucketName,
		fallback:       fallback,
		operationsChan: make(chan storageOperation, operationsBuffer),
		shutdown:       make(chan struct{}),
	}
	
	// Start worker pool for async operations
	for i := 0; i < workerPoolSize; i++ {
		storage.workerPool.Add(1)
		go storage.asyncWorker()
	}
	
	return storage, nil
}

// asyncWorker processes storage operations asynchronously
func (s *MinIOStorage) asyncWorker() {
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
func (s *MinIOStorage) Shutdown() {
	close(s.shutdown)
	s.workerPool.Wait()
	close(s.operationsChan)
}

// StoreAsync performs asynchronous storage operation
func (s *MinIOStorage) StoreAsync(datasetID string, data DatasetData) <-chan error {
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
func (s *MinIOStorage) Store(datasetID string, data DatasetData) error {
	return s.storeSync(datasetID, data)
}

// storeSync performs synchronous storage operation
func (s *MinIOStorage) storeSync(datasetID string, data DatasetData) error {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.minio.store")
	defer span.End()
	
	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.Int("dataset.player_count", len(data.Players)),
		attribute.String("storage.type", "minio"),
	)
	
	if s.client == nil {
		AddSpanEvent(ctx, "storage.fallback_to_memory")
		return s.fallback.Store(datasetID, data)
	}

	start := time.Now()
	jsonData, err := json.Marshal(data)
	if err != nil {
		RecordError(ctx, err, "Failed to marshal dataset data")
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
		attribute.String("minio.bucket", s.bucketName),
		attribute.String("minio.object", objectName),
		attribute.Int("data.size_bytes", len(jsonData)),
		attribute.Int("compressed.size_bytes", len(compressedData)),
		attribute.Float64("compression.ratio", float64(len(jsonData))/float64(len(compressedData))),
	)
	
	_, err = s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(compressedData)), minio.PutObjectOptions{
		ContentType:     "application/gzip",
		ContentEncoding: "gzip",
	})
	if err != nil {
		RecordError(ctx, err, "Failed to store to MinIO")
		log.Printf("Warning: Failed to store to MinIO: %v. Using fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "minio_store_failed"))
		return s.fallback.Store(datasetID, data)
	}

	RecordDBOperation(ctx, "store", "minio_datasets", time.Since(start), 1)
	log.Printf("Stored dataset %s to MinIO", datasetID)
	return nil
}

// Retrieve is the public synchronous interface  
func (s *MinIOStorage) Retrieve(datasetID string) (DatasetData, error) {
	return s.retrieveSync(datasetID)
}

// retrieveSync performs synchronous retrieval operation
func (s *MinIOStorage) retrieveSync(datasetID string) (DatasetData, error) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "storage.minio.retrieve")
	defer span.End()
	
	SetSpanAttributes(ctx,
		attribute.String("dataset.id", datasetID),
		attribute.String("storage.type", "minio"),
	)
	
	if s.client == nil {
		AddSpanEvent(ctx, "storage.fallback_to_memory")
		return s.fallback.Retrieve(datasetID)
	}

	// Try compressed file first, then fall back to uncompressed
	objectName := fmt.Sprintf("datasets/%s.json.gz", datasetID)
	isCompressed := true
	
	SetSpanAttributes(ctx,
		attribute.String("minio.bucket", s.bucketName),
		attribute.String("minio.object", objectName),
	)
	
	start := time.Now()
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		// Try uncompressed file
		objectName = fmt.Sprintf("datasets/%s.json", datasetID)
		isCompressed = false
		SetSpanAttributes(ctx, attribute.String("minio.object", objectName))
		object, err = s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			RecordError(ctx, err, "Failed to retrieve from MinIO")
			log.Printf("Warning: Failed to retrieve from MinIO: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "minio_get_failed"))
			return s.fallback.Retrieve(datasetID)
		}
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		RecordError(ctx, err, "Failed to read MinIO object data")
		log.Printf("Warning: Failed to read from MinIO object: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "minio_read_failed"))
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
			RecordError(ctx, err, "Failed to decompress MinIO data")
			log.Printf("Warning: Failed to decompress MinIO data: %v. Trying fallback storage.", err)
			AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "decompress_failed"))
			return s.fallback.Retrieve(datasetID)
		}
		SetSpanAttributes(ctx, attribute.Int("data.decompressed_size_bytes", len(jsonData)))
	} else {
		jsonData = data
	}

	var dataset DatasetData
	if err := json.Unmarshal(jsonData, &dataset); err != nil {
		RecordError(ctx, err, "Failed to unmarshal MinIO data")
		log.Printf("Warning: Failed to unmarshal MinIO data: %v. Trying fallback storage.", err)
		AddSpanEvent(ctx, "storage.fallback_to_memory", attribute.String("reason", "unmarshal_failed"))
		return s.fallback.Retrieve(datasetID)
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(dataset.Players)))
	RecordDBOperation(ctx, "retrieve", "minio_datasets", time.Since(start), 1)
	log.Printf("Retrieved dataset %s from MinIO", datasetID)
	return dataset, nil
}

// Delete is the public synchronous interface
func (s *MinIOStorage) Delete(datasetID string) error {
	return s.deleteSync(datasetID)
}

// deleteSync performs synchronous deletion operation
func (s *MinIOStorage) deleteSync(datasetID string) error {
	if s.client == nil {
		return s.fallback.Delete(datasetID)
	}

	objectName := fmt.Sprintf("datasets/%s.json", datasetID)
	ctx := context.Background()
	
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Warning: Failed to delete from MinIO: %v. Using fallback storage.", err)
		return s.fallback.Delete(datasetID)
	}

	s.fallback.Delete(datasetID)
	log.Printf("Deleted dataset %s from MinIO", datasetID)
	return nil
}

func (s *MinIOStorage) List() ([]string, error) {
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
			log.Printf("Warning: Error listing MinIO objects: %v. Using fallback storage.", object.Err)
			return s.fallback.List()
		}
		
		if strings.HasSuffix(object.Key, ".json") {
			id := strings.TrimPrefix(object.Key, "datasets/")
			id = strings.TrimSuffix(id, ".json")
			ids = append(ids, id)
		}
	}

	log.Printf("Listed %d datasets from MinIO", len(ids))
	return ids, nil
}

func (s *MinIOStorage) CleanupOldDatasets(maxAge time.Duration, excludeDatasets []string) error {
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
			log.Printf("Warning: Error listing MinIO objects during cleanup: %v", object.Err)
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
			log.Printf("Skipping cleanup for excluded dataset: %s", datasetID)
			continue
		}

		// Check if object is older than cutoff time
		if object.LastModified.Before(cutoffTime) {
			log.Printf("Deleting old dataset: %s (last modified: %s)", datasetID, object.LastModified.Format(time.RFC3339))
			
			err := s.client.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete old dataset %s from MinIO: %v", datasetID, err)
				continue
			}
			
			deletedCount++
		}
	}

	log.Printf("Cleanup completed: deleted %d old datasets from MinIO", deletedCount)
	return nil
}

func InitializeStorage() StorageInterface {
	inMemory := NewInMemoryStorage()

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		log.Println("No MinIO endpoint configured. Using in-memory storage only.")
		return inMemory
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := strings.ToLower(os.Getenv("MINIO_USE_SSL")) == "true"

	if accessKey == "" || secretKey == "" {
		log.Println("MinIO credentials not provided. Using in-memory storage only.")
		return inMemory
	}

	// Debug logging (only show first few chars for security)
	accessKeyPrefix := accessKey
	if len(accessKey) > 4 {
		accessKeyPrefix = accessKey[:4]
	}
	log.Printf("MinIO Config: endpoint=%s, accessKey=%s..., useSSL=%v", 
		minioEndpoint, 
		accessKeyPrefix, 
		useSSL)

	minioStorage, err := NewMinIOStorage(minioEndpoint, accessKey, secretKey, "v2fmdash", useSSL, inMemory)
	if err != nil {
		log.Printf("Failed to initialize MinIO storage: %v. Using in-memory storage only.", err)
		return inMemory
	}

	log.Println("Initialized MinIO storage with in-memory fallback")
	return minioStorage
}