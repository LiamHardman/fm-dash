// Package main provides configuration management for the FM24 application.
// It handles loading configuration from files and environment variables,
// with support for YAML configuration files and comprehensive validation.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Features      FeaturesConfig      `yaml:"features"`
	Observability ObservabilityConfig `yaml:"observability"`
	Upload        UploadConfig        `yaml:"upload"`
	Storage       StorageConfig       `yaml:"storage"`
	Images        ImagesConfig        `yaml:"images"`
	Frontend      FrontendConfig      `yaml:"frontend"`
	Build         BuildConfig         `yaml:"build"`
	Security      SecurityConfig      `yaml:"security"`
	RateLimiting  RateLimitingConfig  `yaml:"rate_limiting"`
	Cache         CacheConfig         `yaml:"cache"`
	Performance   PerformanceConfig   `yaml:"performance"`
}

// ServerConfig defines server-related configuration settings
type ServerConfig struct {
	Port         int    `yaml:"port"`
	PortNginx    int    `yaml:"port_nginx"`
	Host         string `yaml:"host"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

// FeaturesConfig defines feature flags and toggles
type FeaturesConfig struct {
	MetricsEnabled bool `yaml:"metrics_enabled"`
	UploadEnabled  bool `yaml:"upload_enabled"`
	ExportEnabled  bool `yaml:"export_enabled"`
}

// ObservabilityConfig defines observability and monitoring settings
type ObservabilityConfig struct {
	Otel OtelConfig `yaml:"otel"`
}

// OtelConfig defines OpenTelemetry configuration
type OtelConfig struct {
	Enabled           bool           `yaml:"enabled"`
	InsecureMode      bool           `yaml:"insecure_mode"`
	TelemetryDisabled bool           `yaml:"telemetry_disabled"`
	ServiceName       string         `yaml:"service_name"`
	ServiceVersion    string         `yaml:"service_version"`
	ServiceNamespace  string         `yaml:"service_namespace"`
	Environment       string         `yaml:"environment"`
	Exporter          ExporterConfig `yaml:"exporter"`
	Tracing           TracingConfig  `yaml:"tracing"`
	Metrics           MetricsConfig  `yaml:"metrics"`
	Logging           LoggingConfig  `yaml:"logging"`
	Batch             BatchConfig    `yaml:"batch"`
	Resource          ResourceConfig `yaml:"resource"`
}

// ExporterConfig defines OpenTelemetry exporter settings
type ExporterConfig struct {
	Endpoint string `yaml:"endpoint"`
	Timeout  string `yaml:"timeout"`
}

// TracingConfig defines OpenTelemetry tracing configuration
type TracingConfig struct {
	Enabled    bool    `yaml:"enabled"`
	SampleRate float64 `yaml:"sample_rate"`
}

// MetricsConfig defines OpenTelemetry metrics configuration
type MetricsConfig struct {
	Enabled               bool   `yaml:"enabled"`
	ExportInterval        string `yaml:"export_interval"`
	RuntimeMetricsEnabled bool   `yaml:"runtime_metrics_enabled"`
	CustomMetricsEnabled  bool   `yaml:"custom_metrics_enabled"`
}

// LoggingConfig defines OpenTelemetry logging configuration
type LoggingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	LogLevel string `yaml:"log_level"`
}

// BatchConfig defines OpenTelemetry batch processing settings
type BatchConfig struct {
	ScheduleDelay      string `yaml:"schedule_delay"`
	MaxExportBatchSize int    `yaml:"max_export_batch_size"`
	MaxQueueSize       int    `yaml:"max_queue_size"`
	ExportTimeout      string `yaml:"export_timeout"`
}

// ResourceConfig defines OpenTelemetry resource detection settings
type ResourceConfig struct {
	DetectionTimeout string `yaml:"detection_timeout"`
}

// UploadConfig defines file upload settings and constraints
type UploadConfig struct {
	MaxSizeMB         int      `yaml:"max_size_mb"`
	AllowedExtensions []string `yaml:"allowed_extensions"`
	TempDirectory     string   `yaml:"temp_directory"`
	Timeout           string   `yaml:"timeout"`
}

// StorageConfig defines data storage and retention settings
type StorageConfig struct {
	DatasetRetentionDays int      `yaml:"dataset_retention_days"`
	S3                   S3Config `yaml:"s3"`
}

// S3Config defines S3 storage configuration
type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
}

// ImagesConfig defines image-related configuration settings
type ImagesConfig struct {
	APIURL         string `yaml:"api_url"`
	FacesDirectory string `yaml:"faces_directory"`
	LogosDirectory string `yaml:"logos_directory"`
}

// FrontendConfig defines frontend application settings
type FrontendConfig struct {
	APIEndpoint  string `yaml:"api_endpoint"`
	GATrackingID string `yaml:"ga_tracking_id"`
}

// BuildConfig defines build and deployment information
type BuildConfig struct {
	RepositoryURL  string `yaml:"repository_url"`
	Branch         string `yaml:"branch"`
	CommitSHA      string `yaml:"commit_sha"`
	BuildID        string `yaml:"build_id"`
	BuildTimestamp string `yaml:"build_timestamp"`
	DeploymentEnv  string `yaml:"deployment_env"`
}

// SecurityConfig defines security and CORS settings
type SecurityConfig struct {
	EnableCors     bool     `yaml:"enable_cors"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
}

// RateLimitingConfig defines rate limiting settings
type RateLimitingConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	UploadPerHour     int  `yaml:"upload_per_hour"`
	ExportPerHour     int  `yaml:"export_per_hour"`
}

// CacheConfig defines caching settings
type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
	MaxSizeMB  int  `yaml:"max_size_mb"`
}

// PerformanceConfig defines performance and processing settings
type PerformanceConfig struct {
	WorkerCount       int    `yaml:"worker_count"`
	BatchSize         int    `yaml:"batch_size"`
	ProcessingTimeout string `yaml:"processing_timeout"`
	MemoryLimitMB     int    `yaml:"memory_limit_mb"`
}

var (
	appConfig *Config
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Set defaults
	setDefaults(config)

	// Load from file if specified
	configPath := os.Getenv("CONFIG_FILE_PATH")
	if configPath == "" {
		configPath = "/app/config/app-config.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		log.Printf("Loading configuration from: %s", configPath)
		if err := loadFromFile(config, configPath); err != nil {
			log.Printf("Warning: Failed to load config file: %v", err)
		}
	} else {
		log.Printf("Config file not found at %s, using defaults and environment variables", configPath)
	}

	// Override with environment variables
	applyEnvironmentOverrides(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Printf("Configuration loaded successfully")
	return config, nil
}

func setDefaults(config *Config) {
	config.Server.Port = 8091
	config.Server.PortNginx = 8080
	config.Server.Host = "0.0.0.0"
	config.Server.ReadTimeout = "30s"
	config.Server.WriteTimeout = "30s"
	config.Server.IdleTimeout = "120s"

	config.Features.MetricsEnabled = true
	config.Features.UploadEnabled = true
	config.Features.ExportEnabled = true

	config.Upload.MaxSizeMB = 50
	config.Upload.AllowedExtensions = []string{".html", ".htm"}
	config.Upload.TempDirectory = "/tmp/uploads"
	config.Upload.Timeout = "300s"

	config.Storage.DatasetRetentionDays = 30
	config.Storage.S3.Bucket = "v2fmdash-data"
	config.Storage.S3.Region = "us-east-1"

	config.Performance.WorkerCount = 4
	config.Performance.BatchSize = 100
	config.Performance.ProcessingTimeout = "600s"
	config.Performance.MemoryLimitMB = 512
}

func loadFromFile(config *Config, path string) error {
	// Only allow config files in /app/config/ or ./config/
	if !strings.HasPrefix(path, "/app/config/") && !strings.HasPrefix(path, "./config/") {
		return ErrInvalidConfigPath
	}

	data, err := os.ReadFile(path) // #nosec G304 -- path is validated above
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func applyEnvironmentOverrides(config *Config) {
	// Server configuration
	if port := os.Getenv("PORT_GO_API"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if portNginx := os.Getenv("PORT_NGINX"); portNginx != "" {
		if p, err := strconv.Atoi(portNginx); err == nil {
			config.Server.PortNginx = p
		}
	}

	// Features
	if metrics := os.Getenv("ENABLE_METRICS"); metrics != "" {
		config.Features.MetricsEnabled = metrics == "true"
	}

	// OTEL configuration
	if enabled := os.Getenv("OTEL_ENABLED"); enabled != "" {
		config.Observability.Otel.Enabled = enabled == "true"
	}
	if insecureMode := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); insecureMode != "" {
		config.Observability.Otel.InsecureMode = insecureMode == "true"
	}
	if telemetryDisabled := os.Getenv("OTEL_TELEMETRY_DISABLED"); telemetryDisabled != "" {
		config.Observability.Otel.TelemetryDisabled = telemetryDisabled == "true"
	}
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.Observability.Otel.ServiceName = serviceName
	}
	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.Observability.Otel.ServiceVersion = serviceVersion
	}
	if serviceNamespace := os.Getenv("SERVICE_NAMESPACE"); serviceNamespace != "" {
		config.Observability.Otel.ServiceNamespace = serviceNamespace
	}
	if environment := os.Getenv("ENVIRONMENT"); environment != "" {
		config.Observability.Otel.Environment = environment
	}
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		config.Observability.Otel.Exporter.Endpoint = endpoint
	}

	// OTEL Tracing
	if sampleRate := os.Getenv("OTEL_TRACE_SAMPLE_RATE"); sampleRate != "" {
		if rate, err := strconv.ParseFloat(sampleRate, 64); err == nil {
			config.Observability.Otel.Tracing.SampleRate = rate
		}
	}

	// OTEL Metrics
	if metricsEnabled := os.Getenv("OTEL_METRICS_ENABLED"); metricsEnabled != "" {
		config.Observability.Otel.Metrics.Enabled = metricsEnabled == "true"
	}
	if exportInterval := os.Getenv("OTEL_METRIC_EXPORT_INTERVAL"); exportInterval != "" {
		// Convert milliseconds to duration string
		if ms, err := strconv.Atoi(exportInterval); err == nil {
			config.Observability.Otel.Metrics.ExportInterval = fmt.Sprintf("%dms", ms)
		}
	}
	if runtimeMetrics := os.Getenv("OTEL_RUNTIME_METRICS_ENABLED"); runtimeMetrics != "" {
		config.Observability.Otel.Metrics.RuntimeMetricsEnabled = runtimeMetrics == "true"
	}
	if customMetrics := os.Getenv("OTEL_CUSTOM_METRICS_ENABLED"); customMetrics != "" {
		config.Observability.Otel.Metrics.CustomMetricsEnabled = customMetrics == "true"
	}

	// OTEL Logging
	if logsEnabled := os.Getenv("OTEL_LOGS_ENABLED"); logsEnabled != "" {
		config.Observability.Otel.Logging.Enabled = logsEnabled == "true"
	}
	if logLevel := os.Getenv("OTEL_LOG_LEVEL"); logLevel != "" {
		config.Observability.Otel.Logging.LogLevel = logLevel
	}

	// OTEL Tracing toggle
	if tracesEnabled := os.Getenv("OTEL_TRACES_ENABLED"); tracesEnabled != "" {
		config.Observability.Otel.Tracing.Enabled = tracesEnabled == "true"
	}

	// OTEL Batch configuration
	if scheduleDelay := os.Getenv("OTEL_BSP_SCHEDULE_DELAY"); scheduleDelay != "" {
		config.Observability.Otel.Batch.ScheduleDelay = scheduleDelay
	}
	if maxBatchSize := os.Getenv("OTEL_BSP_MAX_EXPORT_BATCH_SIZE"); maxBatchSize != "" {
		if size, err := strconv.Atoi(maxBatchSize); err == nil {
			config.Observability.Otel.Batch.MaxExportBatchSize = size
		}
	}
	if maxQueueSize := os.Getenv("OTEL_BSP_MAX_QUEUE_SIZE"); maxQueueSize != "" {
		if size, err := strconv.Atoi(maxQueueSize); err == nil {
			config.Observability.Otel.Batch.MaxQueueSize = size
		}
	}
	if exportTimeout := os.Getenv("OTEL_BSP_EXPORT_TIMEOUT"); exportTimeout != "" {
		config.Observability.Otel.Batch.ExportTimeout = exportTimeout
	}

	// OTEL Resource configuration
	if detectionTimeout := os.Getenv("OTEL_RESOURCE_DETECTION_TIMEOUT"); detectionTimeout != "" {
		config.Observability.Otel.Resource.DetectionTimeout = detectionTimeout
	}

	// Upload configuration
	if maxUploadSize := os.Getenv("MAX_UPLOAD_SIZE"); maxUploadSize != "" {
		if size, err := strconv.Atoi(maxUploadSize); err == nil {
			config.Upload.MaxSizeMB = size
		}
	}

	// Images configuration
	if imageAPIURL := os.Getenv("IMAGE_API_URL"); imageAPIURL != "" {
		config.Images.APIURL = imageAPIURL
	}

	// Frontend Analytics configuration
	if gaTrackingID := os.Getenv("VITE_GA_TRACKING_ID"); gaTrackingID != "" {
		config.Frontend.GATrackingID = gaTrackingID
	}

	// Storage configuration
	if retentionDays := os.Getenv("DATASET_RETENTION_DAYS"); retentionDays != "" {
		if days, err := strconv.Atoi(retentionDays); err == nil {
			config.Storage.DatasetRetentionDays = days
		}
	}

	// S3 configuration from secrets/environment
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		config.Storage.S3.Endpoint = endpoint
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		config.Storage.S3.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		config.Storage.S3.SecretKey = secretKey
	}
	if useSSL := os.Getenv("S3_USE_SSL"); useSSL != "" {
		config.Storage.S3.UseSSL = strings.ToLower(useSSL) == "true"
	}

	// Build information
	if repoURL := os.Getenv("GIT_REPOSITORY_URL"); repoURL != "" {
		config.Build.RepositoryURL = repoURL
	}
	if branch := os.Getenv("GIT_BRANCH"); branch != "" {
		config.Build.Branch = branch
	}
	if sha := os.Getenv("GIT_COMMIT_SHA"); sha != "" {
		config.Build.CommitSHA = sha
	}
	if buildID := os.Getenv("BUILD_ID"); buildID != "" {
		config.Build.BuildID = buildID
	}
	if buildTimestamp := os.Getenv("BUILD_TIMESTAMP"); buildTimestamp != "" {
		config.Build.BuildTimestamp = buildTimestamp
	}
	if deploymentEnv := os.Getenv("DEPLOYMENT_ENV"); deploymentEnv != "" {
		config.Build.DeploymentEnv = deploymentEnv
	}

	// Frontend configuration
	if apiEndpoint := os.Getenv("API_ENDPOINT"); apiEndpoint != "" {
		config.Frontend.APIEndpoint = apiEndpoint
	}
}

// Define static errors for validation
var (
	ErrInvalidServerPort    = errors.New("invalid server port")
	ErrInvalidNginxPort     = errors.New("invalid nginx port")
	ErrInvalidUploadSize    = errors.New("invalid upload max size")
	ErrInvalidRetentionDays = errors.New("invalid dataset retention days")
	ErrInvalidConfigPath    = errors.New("invalid config file path")
)

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidServerPort, config.Server.Port)
	}

	if config.Server.PortNginx <= 0 || config.Server.PortNginx > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidNginxPort, config.Server.PortNginx)
	}

	if config.Upload.MaxSizeMB <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidUploadSize, config.Upload.MaxSizeMB)
	}

	if config.Storage.DatasetRetentionDays <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidRetentionDays, config.Storage.DatasetRetentionDays)
	}

	return nil
}

// GetConfig returns the loaded configuration
func GetConfig() *Config {
	return appConfig
}

// InitConfig initializes the global configuration
func InitConfig() error {
	var err error
	appConfig, err = LoadConfig()
	return err
}

// GetServerPort returns the server port from configuration or environment
func GetServerPort() int {
	if appConfig != nil {
		return appConfig.Server.Port
	}
	// Fallback to environment variable
	if port := os.Getenv("PORT_GO_API"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return 8091
}

// GetNginxPort returns the nginx port from configuration or environment
func GetNginxPort() int {
	if appConfig != nil {
		return appConfig.Server.PortNginx
	}
	// Fallback to environment variable
	if port := os.Getenv("PORT_NGINX"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return 8080
}

// IsMetricsEnabled returns whether metrics are enabled
func IsMetricsEnabled() bool {
	if appConfig != nil {
		return appConfig.Features.MetricsEnabled
	}
	return os.Getenv("ENABLE_METRICS") == "true"
}

// IsOTelEnabled returns whether OpenTelemetry is enabled
func IsOTelEnabled() bool {
	if appConfig != nil {
		return appConfig.Observability.Otel.Enabled
	}
	return os.Getenv("OTEL_ENABLED") == "true"
}
