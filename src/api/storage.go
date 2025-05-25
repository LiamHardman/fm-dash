package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageInterface interface {
	Store(datasetID string, data DatasetData) error
	Retrieve(datasetID string) (DatasetData, error)
	Delete(datasetID string) error
	List() ([]string, error)
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[datasetID] = data
	return nil
}

func (s *InMemoryStorage) Retrieve(datasetID string) (DatasetData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	data, exists := s.data[datasetID]
	if !exists {
		return DatasetData{}, fmt.Errorf("dataset %s not found", datasetID)
	}
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

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	fallback   StorageInterface
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool, fallback StorageInterface) (*MinIOStorage, error) {
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
	return &MinIOStorage{
		client:     client,
		bucketName: bucketName,
		fallback:   fallback,
	}, nil
}

func (s *MinIOStorage) Store(datasetID string, data DatasetData) error {
	if s.client == nil {
		return s.fallback.Store(datasetID, data)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	objectName := fmt.Sprintf("datasets/%s.json", datasetID)
	reader := bytes.NewReader(jsonData)
	
	ctx := context.Background()
	_, err = s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(jsonData)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		log.Printf("Warning: Failed to store to MinIO: %v. Using fallback storage.", err)
		return s.fallback.Store(datasetID, data)
	}

	log.Printf("Stored dataset %s to MinIO", datasetID)
	return nil
}

func (s *MinIOStorage) Retrieve(datasetID string) (DatasetData, error) {
	if s.client == nil {
		return s.fallback.Retrieve(datasetID)
	}

	objectName := fmt.Sprintf("datasets/%s.json", datasetID)
	ctx := context.Background()
	
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Warning: Failed to retrieve from MinIO: %v. Trying fallback storage.", err)
		return s.fallback.Retrieve(datasetID)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		log.Printf("Warning: Failed to read from MinIO object: %v. Trying fallback storage.", err)
		return s.fallback.Retrieve(datasetID)
	}

	var dataset DatasetData
	if err := json.Unmarshal(data, &dataset); err != nil {
		log.Printf("Warning: Failed to unmarshal MinIO data: %v. Trying fallback storage.", err)
		return s.fallback.Retrieve(datasetID)
	}

	log.Printf("Retrieved dataset %s from MinIO", datasetID)
	return dataset, nil
}

func (s *MinIOStorage) Delete(datasetID string) error {
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