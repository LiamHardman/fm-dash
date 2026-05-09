package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ConfigStorage handles simple named config files that need to be shared across replicas.
// These are stored alongside datasets in whatever backend is configured.
type ConfigStorage interface {
	StoreConfig(name string, data []byte) error
	RetrieveConfig(name string) ([]byte, error)
}

var configStorage ConfigStorage

// InitConfigStorage creates the appropriate config storage backend using the same
// environment variables as the main dataset storage.
func InitConfigStorage() {
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	if s3Endpoint != "" && accessKey != "" && secretKey != "" {
		useSSL := strings.ToLower(os.Getenv("S3_USE_SSL")) == "true"
		bucketName := os.Getenv("S3_BUCKET_NAME")
		if bucketName == "" {
			bucketName = "v2fmdash"
		}
		cs, err := newS3ConfigStorage(s3Endpoint, accessKey, secretKey, bucketName, useSSL)
		if err == nil {
			configStorage = cs
			LogInfo("Config storage initialized: S3 (bucket=%s)", bucketName)
			return
		}
		LogWarn("Failed to initialize S3 config storage, falling back to local: %v", err)
	}

	datasetDir := os.Getenv("DATASETS_DIR")
	if datasetDir == "" {
		datasetDir = "./datasets"
	}
	configStorage = &localConfigStorage{dir: datasetDir}
	LogInfo("Config storage initialized: local (dir=%s)", datasetDir)
}

// ---- local backend ----

type localConfigStorage struct {
	dir   string
	mutex sync.RWMutex
}

func (s *localConfigStorage) StoreConfig(name string, data []byte) error {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return err
	}
	path := filepath.Join(s.dir, name)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return os.WriteFile(path, data, 0o644)
}

func (s *localConfigStorage) RetrieveConfig(name string) ([]byte, error) {
	path := filepath.Join(s.dir, name)
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return os.ReadFile(path)
}

// ---- S3 backend ----

type s3ConfigStorage struct {
	client     *minio.Client
	bucketName string
}

func newS3ConfigStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*s3ConfigStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &s3ConfigStorage{client: client, bucketName: bucketName}, nil
}

func (s *s3ConfigStorage) StoreConfig(name string, data []byte) error {
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucketName, name, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/json"})
	return err
}

func (s *s3ConfigStorage) RetrieveConfig(name string) ([]byte, error) {
	ctx := context.Background()
	obj, err := s.client.GetObject(ctx, s.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := obj.Close(); closeErr != nil {
			LogWarn("Config storage: failed to close S3 object: %v", closeErr)
		}
	}()
	return io.ReadAll(obj)
}
