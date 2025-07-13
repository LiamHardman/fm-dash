package main

import (
	"os"
	"testing"
)

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_KEY_EXISTS",
			defaultValue: "default_value",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "TEST_KEY_NOT_EXISTS",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "empty default value",
			key:          "TEST_KEY_EMPTY_DEFAULT",
			defaultValue: "",
			envValue:     "",
			expected:     "",
		},
		{
			name:         "empty environment value should use default",
			key:          "TEST_KEY_EMPTY_ENV",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before and after
			originalValue := os.Getenv(tt.key)
			defer func() {
				if originalValue != "" {
					if err := os.Setenv(tt.key, originalValue); err != nil {
						t.Logf("Failed to restore environment variable %s: %v", tt.key, err)
					}
				} else {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Failed to unset environment variable %s: %v", tt.key, err)
					}
				}
			}()

			// Set up test environment
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Logf("Failed to set environment variable %s: %v", tt.key, err)
				}
			} else {
				if err := os.Unsetenv(tt.key); err != nil {
					t.Logf("Failed to unset environment variable %s: %v", tt.key, err)
				}
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithDefault(%s, %s) = %s; want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestValidateEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		shouldErr bool
		errString string
	}{
		{
			name:      "no environment variables set",
			envVars:   map[string]string{},
			shouldErr: false,
		},
		{
			name: "valid OTEL_EXPORTER_OTLP_ENDPOINT",
			envVars: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "localhost:4317",
			},
			shouldErr: false,
		},
		{
			name: "invalid OTEL_EXPORTER_OTLP_ENDPOINT - no port",
			envVars: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "localhost",
			},
			shouldErr: true,
			errString: "invalid OTEL_EXPORTER_OTLP_ENDPOINT",
		},
		{
			name: "valid S3_ENDPOINT with port",
			envVars: map[string]string{
				"S3_ENDPOINT": "minio:9000",
			},
			shouldErr: false,
		},
		{
			name: "valid S3_ENDPOINT with http",
			envVars: map[string]string{
				"S3_ENDPOINT": "http://localhost:9000",
			},
			shouldErr: false,
		},
		{
			name: "valid S3_ENDPOINT with https",
			envVars: map[string]string{
				"S3_ENDPOINT": "https://s3.amazonaws.com",
			},
			shouldErr: false,
		},
		{
			name: "invalid S3_ENDPOINT",
			envVars: map[string]string{
				"S3_ENDPOINT": "invalid-endpoint",
			},
			shouldErr: true,
			errString: "invalid S3_ENDPOINT format",
		},
		{
			name: "valid SERVICE_NAME",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24-api",
			},
			shouldErr: false,
		},
		{
			name: "invalid SERVICE_NAME with space",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24 api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "invalid SERVICE_NAME with semicolon",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24;api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "invalid SERVICE_NAME with pipe",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24|api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "invalid SERVICE_NAME with ampersand",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24&api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "invalid SERVICE_NAME with dollar sign",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24$api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "invalid SERVICE_NAME with backtick",
			envVars: map[string]string{
				"SERVICE_NAME": "fm24`api",
			},
			shouldErr: true,
			errString: "invalid SERVICE_NAME: contains unsafe characters",
		},
		{
			name: "multiple valid environment variables",
			envVars: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "localhost:4317",
				"S3_ENDPOINT":                 "http://minio:9000",
				"SERVICE_NAME":                "fm24-api",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original environment variables
			originalEnvVars := make(map[string]string)
			for key := range tt.envVars {
				originalEnvVars[key] = os.Getenv(key)
			}

			// Clean up after test
			defer func() {
				for key, originalValue := range originalEnvVars {
					if originalValue != "" {
						if err := os.Setenv(key, originalValue); err != nil {
							t.Logf("Failed to restore environment variable %s: %v", key, err)
						}
					} else {
						if err := os.Unsetenv(key); err != nil {
							t.Logf("Failed to unset environment variable %s: %v", key, err)
						}
					}
				}
				// Clean up any test env vars not in original
				for key := range tt.envVars {
					if _, exists := originalEnvVars[key]; !exists {
						if err := os.Unsetenv(key); err != nil {
							t.Logf("Failed to unset environment variable %s: %v", key, err)
						}
					}
				}
			}()

			// Set up test environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Logf("Failed to set environment variable %s: %v", key, err)
				}
			}

			err := validateEnvironmentVariables()

			if tt.shouldErr {
				if err == nil {
					t.Errorf("validateEnvironmentVariables() should have returned an error")
				} else if tt.errString != "" && err.Error() != tt.errString {
					// Check if error message contains expected substring
					if len(tt.errString) > 0 && len(err.Error()) > 0 {
						// More flexible error message matching
						expected := tt.errString
						actual := err.Error()
						// Check if actual error contains the expected substring
						found := false
						if len(expected) <= len(actual) {
							for i := 0; i <= len(actual)-len(expected); i++ {
								if actual[i:i+len(expected)] == expected {
									found = true
									break
								}
							}
						}
						if !found {
							t.Errorf("validateEnvironmentVariables() error = %v; want error containing %v", err, expected)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("validateEnvironmentVariables() should not have returned an error, got: %v", err)
				}
			}
		})
	}
}
