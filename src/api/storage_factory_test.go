package main

import (
	"os"
	"testing"
)

func TestValidateStorageConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedUse    bool
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "protobuf enabled with true",
			envValue:       "true",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf enabled with 1",
			envValue:       "1",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf enabled with yes",
			envValue:       "yes",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf enabled with on",
			envValue:       "on",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf disabled with false",
			envValue:       "false",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf disabled with 0",
			envValue:       "0",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf disabled with no",
			envValue:       "no",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf disabled with off",
			envValue:       "off",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "protobuf disabled with empty string",
			envValue:       "",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "invalid protobuf value",
			envValue:       "invalid",
			expectedUse:    false,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "case insensitive TRUE",
			envValue:       "TRUE",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "case insensitive FALSE",
			envValue:       "FALSE",
			expectedUse:    false,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "whitespace handling",
			envValue:       "  true  ",
			expectedUse:    true,
			expectedValid:  true,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			oldValue := os.Getenv("USE_PROTOBUF")
			defer func() {
				if oldValue == "" {
					os.Unsetenv("USE_PROTOBUF")
				} else {
					os.Setenv("USE_PROTOBUF", oldValue)
				}
			}()

			os.Setenv("USE_PROTOBUF", tt.envValue)

			// Test the validation function
			config := validateStorageConfiguration()

			if config.UseProtobuf != tt.expectedUse {
				t.Errorf("Expected UseProtobuf=%v, got %v", tt.expectedUse, config.UseProtobuf)
			}

			if config.ConfigValid != tt.expectedValid {
				t.Errorf("Expected ConfigValid=%v, got %v", tt.expectedValid, config.ConfigValid)
			}

			if len(config.Errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(config.Errors), config.Errors)
			}
		})
	}
}

func TestCreateProtobufEnabledStorage(t *testing.T) {
	// Create a base storage for testing
	baseStorage := CreateInMemoryStorage()

	// Test creating protobuf-enabled storage
	protobufStorage := CreateProtobufEnabledStorage(baseStorage)

	if protobufStorage == nil {
		t.Fatal("Expected protobuf storage to be created, got nil")
	}

	// Verify it's a ProtobufStorage wrapper by checking the type
	if _, ok := protobufStorage.(*ProtobufStorage); !ok {
		t.Errorf("Expected ProtobufStorage type, got %T", protobufStorage)
	}
}

func TestCreateJSONStorage(t *testing.T) {
	// Test creating JSON-only storage
	jsonStorage := CreateJSONStorage()

	if jsonStorage == nil {
		t.Fatal("Expected JSON storage to be created, got nil")
	}

	// Verify it's not a ProtobufStorage wrapper
	if _, ok := jsonStorage.(*ProtobufStorage); ok {
		t.Error("Expected non-ProtobufStorage type, got ProtobufStorage")
	}
}

func TestInitializeStorageWithProtobufEnabled(t *testing.T) {
	// Save original environment
	oldValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("USE_PROTOBUF")
		} else {
			os.Setenv("USE_PROTOBUF", oldValue)
		}
	}()

	// Test with protobuf enabled
	os.Setenv("USE_PROTOBUF", "true")

	storage := InitializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be initialized, got nil")
	}

	// Verify it's a ProtobufStorage wrapper
	if _, ok := storage.(*ProtobufStorage); !ok {
		t.Errorf("Expected ProtobufStorage type when USE_PROTOBUF=true, got %T", storage)
	}
}

func TestInitializeStorageWithProtobufDisabled(t *testing.T) {
	// Save original environment
	oldValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("USE_PROTOBUF")
		} else {
			os.Setenv("USE_PROTOBUF", oldValue)
		}
	}()

	// Test with protobuf disabled
	os.Setenv("USE_PROTOBUF", "false")

	storage := InitializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be initialized, got nil")
	}

	// Verify it's not a ProtobufStorage wrapper
	if _, ok := storage.(*ProtobufStorage); ok {
		t.Error("Expected non-ProtobufStorage type when USE_PROTOBUF=false, got ProtobufStorage")
	}
}

func TestInitializeStorageWithInvalidProtobufValue(t *testing.T) {
	// Save original environment
	oldValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("USE_PROTOBUF")
		} else {
			os.Setenv("USE_PROTOBUF", oldValue)
		}
	}()

	// Test with invalid protobuf value
	os.Setenv("USE_PROTOBUF", "invalid")

	storage := InitializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be initialized, got nil")
	}

	// Should fall back to JSON storage when invalid value is provided
	if _, ok := storage.(*ProtobufStorage); ok {
		t.Error("Expected non-ProtobufStorage type when USE_PROTOBUF has invalid value, got ProtobufStorage")
	}
}

func TestStorageFactoryFallbackBehavior(t *testing.T) {
	// Save original environment
	oldValue := os.Getenv("USE_PROTOBUF")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("USE_PROTOBUF")
		} else {
			os.Setenv("USE_PROTOBUF", oldValue)
		}
	}()

	// Test that storage factory always returns a valid storage instance
	testCases := []string{"true", "false", "invalid", ""}

	for _, testCase := range testCases {
		t.Run("USE_PROTOBUF="+testCase, func(t *testing.T) {
			if testCase == "" {
				os.Unsetenv("USE_PROTOBUF")
			} else {
				os.Setenv("USE_PROTOBUF", testCase)
			}

			storage := InitializeStorage()

			if storage == nil {
				t.Fatal("Expected storage to be initialized, got nil")
			}

			// Test basic storage operations to ensure it's functional
			testData := DatasetData{
				Players:        []Player{},
				CurrencySymbol: "USD",
			}

			// Test store operation
			err := storage.Store("test-dataset", testData)
			if err != nil {
				t.Errorf("Storage store operation failed: %v", err)
			}

			// Test retrieve operation
			retrievedData, err := storage.Retrieve("test-dataset")
			if err != nil {
				t.Errorf("Storage retrieve operation failed: %v", err)
			}

			if retrievedData.CurrencySymbol != testData.CurrencySymbol {
				t.Errorf("Expected currency symbol %s, got %s", testData.CurrencySymbol, retrievedData.CurrencySymbol)
			}

			// Test delete operation
			err = storage.Delete("test-dataset")
			if err != nil {
				t.Errorf("Storage delete operation failed: %v", err)
			}
		})
	}
}