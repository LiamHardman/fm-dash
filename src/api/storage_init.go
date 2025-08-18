package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// InitializeStorage creates and returns the appropriate storage implementation
//
//nolint:ireturn // StorageInterface is the intended return type for this factory function
func InitializeStorage(ctx context.Context) StorageInterface {
	logInfo(ctx, "Initializing storage system")
	start := time.Now()

	// Validate and determine storage configuration
	config := validateStorageConfiguration(ctx)

	// Initialize the base storage backend
	baseStorage := initializeBaseStorage(ctx)

	// Apply protobuf wrapper if enabled and available
	if config.UseProtobuf {
		protobufStorage, err := createProtobufStorageWithFallback(ctx, baseStorage)
		if err != nil {
			logError(ctx, "Failed to initialize protobuf storage, falling back to JSON serialization",
				"error", err, "duration_ms", time.Since(start).Milliseconds())
			return baseStorage
		}
		logInfo(ctx, "Storage initialization completed",
			"storage_type", "protobuf",
			"duration_ms", time.Since(start).Milliseconds())
		return protobufStorage
	}

	logDebug(ctx, "Storage initialization completed",
		"storage_type", "json",
		"duration_ms", time.Since(start).Milliseconds())
	return baseStorage
}

// CreateProtobufEnabledStorage creates a storage instance with protobuf serialization enabled
// This factory method allows explicit creation of protobuf storage for testing or specific use cases
//
//nolint:ireturn // StorageInterface is the intended return type for this factory function
func CreateProtobufEnabledStorage(ctx context.Context, backend StorageInterface) StorageInterface {
	logDebug(ctx, "Creating protobuf-enabled storage wrapper")
	return CreateProtobufStorage(backend)
}

// CreateJSONStorage creates a storage instance with JSON serialization (no protobuf wrapper)
// This factory method allows explicit creation of JSON-only storage for testing or specific use cases
//
//nolint:ireturn // StorageInterface is the intended return type for this factory function
func CreateJSONStorage(ctx context.Context) StorageInterface {
	logInfo(ctx, "Creating JSON-only storage without protobuf wrapper")
	return initializeBaseStorage(ctx)
}

// StorageConfiguration holds the validated storage configuration
type StorageConfiguration struct {
	UseProtobuf bool
	ConfigValid bool
	Errors      []string
}

// validateStorageConfiguration validates the storage configuration from environment variables
func validateStorageConfiguration(ctx context.Context) StorageConfiguration {
	logDebug(ctx, "Validating storage configuration")
	start := time.Now()

	config := StorageConfiguration{
		UseProtobuf: false,
		ConfigValid: true,
		Errors:      []string{},
	}

	// Check USE_PROTOBUF environment variable
	protobufEnv := strings.ToLower(strings.TrimSpace(os.Getenv("USE_PROTOBUF")))

	switch protobufEnv {
	case "true", "1", "yes", "on":
		config.UseProtobuf = true
		logInfo(ctx, "Protobuf storage configuration enabled", "env_value", protobufEnv)
	case "false", "0", "no", "off", "":
		config.UseProtobuf = false
		logInfo(ctx, "Protobuf storage configuration disabled", "env_value", protobufEnv)
	default:
		config.ConfigValid = false
		config.Errors = append(config.Errors, fmt.Sprintf("invalid USE_PROTOBUF value: %s (expected: true/false)", protobufEnv))
		logError(ctx, "Invalid USE_PROTOBUF environment variable value",
			"error", WrapErrorf(ErrInvalidProtobufEnv, "invalid value: %s", protobufEnv),
			"env_value", protobufEnv,
			"expected_values", "true/false")
	}

	// Log configuration validation results
	if config.ConfigValid {
		logDebug(ctx, "Storage configuration validation completed",
			"use_protobuf", config.UseProtobuf,
			"duration_ms", time.Since(start).Milliseconds())
	} else {
		logError(ctx, "Storage configuration validation failed",
			"error", WrapErrorf(ErrConfigValidationFailed, "validation failed with %d errors", len(config.Errors)),
			"error_count", len(config.Errors),
			"errors", config.Errors,
			"duration_ms", time.Since(start).Milliseconds())
	}

	return config
}

// createProtobufStorageWithFallback creates protobuf storage with graceful fallback handling
//
//nolint:ireturn // This function is designed to return different storage implementations
func createProtobufStorageWithFallback(ctx context.Context, baseStorage StorageInterface) (StorageInterface, error) {
	logDebug(ctx, "Creating protobuf storage with fallback handling")
	start := time.Now()

	// Create protobuf storage wrapper
	protobufStorage := CreateProtobufStorage(baseStorage)
	if protobufStorage == nil {
		err := ErrProtobufStorageWrapper
		logError(ctx, "Failed to create protobuf storage wrapper", "error", err)
		return nil, err
	}

	logDebug(ctx, "Protobuf storage created and validated successfully",
		"duration_ms", time.Since(start).Milliseconds())
	return protobufStorage, nil
}

// initializeBaseStorage creates the underlying storage backend (S3, hybrid, or in-memory)
//
//nolint:ireturn // This function is designed to return different storage implementations
func initializeBaseStorage(ctx context.Context) StorageInterface {
	logDebug(ctx, "Initializing base storage backend")
	start := time.Now()

	inMemory := CreateInMemoryStorage()

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		logInfo(ctx, "No S3 endpoint configured, using hybrid storage", "storage_type", "hybrid")

		// Use configurable datasets directory, default to "./datasets"
		datasetDir := os.Getenv("DATASETS_DIR")
		if datasetDir == "" {
			datasetDir = "./datasets"
		}

		hybrid, err := CreateHybridStorage(datasetDir)
		if err != nil {
			logError(ctx, "Failed to initialize hybrid storage, falling back to in-memory only",
				"error", err, "dataset_dir", datasetDir)
			logInfo(ctx, "Base storage initialization completed",
				"storage_type", "in-memory",
				"duration_ms", time.Since(start).Milliseconds())
			return inMemory
		}

		logInfo(ctx, "Base storage initialization completed",
			"storage_type", "hybrid",
			"dataset_dir", datasetDir,
			"duration_ms", time.Since(start).Milliseconds())
		return hybrid
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	useSSL := strings.ToLower(os.Getenv("S3_USE_SSL")) == "true"

	if accessKey == "" || secretKey == "" {
		logInfo(ctx, "S3 credentials not provided, using hybrid storage", "storage_type", "hybrid")

		// Use configurable datasets directory, default to "./datasets"
		datasetDir := os.Getenv("DATASETS_DIR")
		if datasetDir == "" {
			datasetDir = "./datasets"
		}

		hybrid, err := CreateHybridStorage(datasetDir)
		if err != nil {
			logError(ctx, "Failed to initialize hybrid storage, falling back to in-memory only",
				"error", err, "dataset_dir", datasetDir)
			logInfo(ctx, "Base storage initialization completed",
				"storage_type", "in-memory",
				"duration_ms", time.Since(start).Milliseconds())
			return inMemory
		}

		logInfo(ctx, "Base storage initialization completed",
			"storage_type", "hybrid",
			"dataset_dir", datasetDir,
			"duration_ms", time.Since(start).Milliseconds())
		return hybrid
	}

	// Log configuration without sensitive credentials
	logDebug(ctx, "S3 configuration detected",
		"endpoint", s3Endpoint,
		"use_ssl", useSSL,
		"credentials_provided", accessKey != "" && secretKey != "")

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "v2fmdash"
	}
	s3Storage := CreateS3Storage(s3Endpoint, accessKey, secretKey, bucketName, useSSL, inMemory)

	logInfo(ctx, "Base storage initialization completed",
		"storage_type", "s3",
		"endpoint", s3Endpoint,
		"bucket_name", bucketName,
		"duration_ms", time.Since(start).Milliseconds())
	return s3Storage
}
