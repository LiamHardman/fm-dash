package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	apperrors "api/errors"

	"go.opentelemetry.io/otel/attribute"
)

// LocalFileStorage stores datasets as JSON files in a local directory
type LocalFileStorage struct {
	datasetDir string
	mutex      sync.RWMutex
}

// CreateLocalFileStorage creates a new local file storage instance
func CreateLocalFileStorage(datasetDir string) (*LocalFileStorage, error) {
	// Create datasets directory if it doesn't exist
	if err := os.MkdirAll(datasetDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create datasets directory %s: %w", datasetDir, err)
	}

	LogInfo("Initialized local file storage at: %s", datasetDir)
	return &LocalFileStorage{
		datasetDir: datasetDir,
	}, nil
}

// Store saves a dataset to local file storage
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

	var fileData []byte
	var ext string
	if len(data.RawBytes) > 0 {
		fileData = data.RawBytes
		ext = ".pb.gz"
	} else {
		var err error
		fileData, err = json.Marshal(data)
		if err != nil {
			RecordError(ctx, err, "Failed to marshal dataset data")
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		ext = ".json"
	}

	// Safely construct the filename to prevent path injection
	filename, err := validateAndJoinPath(s.datasetDir, datasetID+ext)
	if err != nil {
		err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Path validation failed")
		return err
	}

	// Write file data
	if err := os.WriteFile(filename, fileData, 0o600); err != nil {
		RecordError(ctx, err, "Failed to write dataset file")
		return fmt.Errorf("failed to write dataset file: %w", err)
	}

	SetSpanAttributes(ctx,
		attribute.String("file.path", sanitizeForLogging(filename)),
		attribute.Int("data.size_bytes", len(fileData)),
	)

	LogDebug("Stored dataset %s to local file: %s", sanitizeForLogging(datasetID), sanitizeForLogging(filename))
	return nil
}

// Retrieve retrieves a dataset from local file storage
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

	// Try protobuf first, then current JSON, then legacy compressed JSON.
	filename, err := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtProtobuf))
	if err != nil {
		err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err)
		RecordError(ctx, err, "Path validation failed")
		return DatasetData{}, err
	}
	isCompressed := false
	isProtobuf := false

	// Check if file exists; if not, try others
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		jsonFilename, errJSON := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtJSON))
		if errJSON != nil {
			err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), errJSON)
			RecordError(ctx, err, "Path validation failed")
			return DatasetData{}, err
		}
		if _, statErr := os.Stat(jsonFilename); statErr == nil {
			filename = jsonFilename
		} else {
			legacyFilename, err2 := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtJSONGzip))
			if err2 != nil {
				err := apperrors.WrapErrInvalidFilePathForDataset(sanitizeForLogging(datasetID), err2)
				RecordError(ctx, err, "Path validation failed")
				return DatasetData{}, err
			}
			if _, statErr := os.Stat(legacyFilename); os.IsNotExist(statErr) {
				err := apperrors.WrapErrDatasetNotFound(sanitizeForLogging(datasetID))
				RecordError(ctx, err, "Dataset not found")
				return DatasetData{}, err
			}
			filename = legacyFilename
			isCompressed = true
		}
	} else {
		isProtobuf = true
	}

	// Read the file
	//nolint:gosec // filename is validated by validateAndJoinPath
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

	// Decompress only if legacy compressed file was used
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
	if isProtobuf {
		dataset.RawBytes = jsonData
	} else {
		if err := json.Unmarshal(jsonData, &dataset); err != nil {
			RecordError(ctx, err, "Failed to unmarshal dataset data")
			return DatasetData{}, fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	SetSpanAttributes(ctx, attribute.Int("dataset.player_count", len(dataset.Players)))
	LogDebug("Retrieved dataset %s from local file: %s", sanitizeForLogging(datasetID), sanitizeForLogging(filename))
	return dataset, nil
}

// Delete removes a dataset from local file storage
func (s *LocalFileStorage) Delete(datasetID string) error {
	// Validate dataset ID format
	if err := validateID(datasetID, 100); err != nil {
		return fmt.Errorf("invalid dataset ID: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Try to delete both compressed and uncompressed versions - safely construct paths
	pbFile, _ := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtProtobuf))
	compressedFile, _ := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtJSONGzip))
	uncompressedFile, _ := validateAndJoinPath(s.datasetDir, fmt.Sprintf("%s%s", datasetID, datasetExtJSON))

	// Don't treat "file not found" as an error
	if pbFile != "" {
		if err := os.Remove(pbFile); err != nil && !os.IsNotExist(err) {
			LogWarn("Warning: Failed to remove pb file %s: %v", sanitizeForLogging(pbFile), err)
		}
	}
	if compressedFile != "" {
		if err := os.Remove(compressedFile); err != nil && !os.IsNotExist(err) {
			LogWarn("Warning: Failed to remove compressed file %s: %v", sanitizeForLogging(compressedFile), err)
		}
	}
	if uncompressedFile != "" {
		if err := os.Remove(uncompressedFile); err != nil && !os.IsNotExist(err) {
			LogWarn("Warning: Failed to remove uncompressed file %s: %v", sanitizeForLogging(uncompressedFile), err)
		}
	}

	LogDebug("Deleted dataset %s from local storage", sanitizeForLogging(datasetID))
	return nil
}

// List returns all dataset IDs stored in local files
func (s *LocalFileStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entries, err := os.ReadDir(s.datasetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read datasets directory: %w", err)
	}

	var ids []string
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		id, ok := datasetIDFromStorageName(name)
		if ok {
			ids = appendDatasetIDOnce(ids, seen, id)
		}
	}

	LogDebug("Listed %d datasets from local storage", len(ids))
	return ids, nil
}

// CleanupOldDatasets removes datasets older than the specified duration from local files
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
		datasetID, ok := datasetIDFromStorageName(name)
		if !ok {
			continue
		}

		// Skip excluded datasets
		if excludeSet[datasetID] {
			LogDebug("Skipping cleanup for excluded dataset: %s", sanitizeForLogging(datasetID))
			continue
		}

		// Get file info to check modification time - safely construct the file path
		filePath, err := validateAndJoinPath(s.datasetDir, name)
		if err != nil {
			LogWarn("Warning: Invalid file path for %s: %v", sanitizeForLogging(name), err)
			continue
		}

		info, err := os.Stat(filePath)
		if err != nil {
			LogWarn("Warning: Failed to get file info for %s: %v", sanitizeForLogging(filePath), err)
			continue
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
			LogDebug("Deleting old dataset file: %s (last modified: %s)", sanitizeForLogging(name), info.ModTime().Format(time.RFC3339))

			if err := os.Remove(filePath); err != nil {
				LogWarn("Warning: Failed to delete old dataset file %s: %v", sanitizeForLogging(filePath), err)
				continue
			}

			deletedCount++
		}
	}

	LogDebug("Cleanup completed: deleted %d old dataset files from local storage", deletedCount)
	return nil
}
