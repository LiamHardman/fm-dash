package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	apperrors "api/errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// OTelConfig holds OpenTelemetry configuration
type OTelConfig struct {
	ServiceName              string
	ServiceVersion           string
	Environment              string
	CollectorURL             string
	InsecureMode             bool
	TraceSampleRate          float64
	MetricExportInterval     time.Duration
	ResourceDetectionTimeout time.Duration
	BatchTimeout             time.Duration
	BatchSize                int
	MaxQueueSize             int
	ExportTimeout            time.Duration
	EnableLogging            bool
	EnableMetrics            bool
	EnableTracing            bool
	EnableRuntimeMetrics     bool
	LogLevel                 string
	MaxAttributeValueLength  int
	MaxAttributeCount        int
	MaxSpanEvents            int
	MaxSpanLinks             int
}

// LoadOTelConfig loads OpenTelemetry configuration from environment variables
func LoadOTelConfig() *OTelConfig {
	return &OTelConfig{
		ServiceName:              getEnvWithDefault("SERVICE_NAME", "v2fmdash-api"),
		ServiceVersion:           getEnvWithDefault("SERVICE_VERSION", "v1.0.0"),
		Environment:              getEnvWithDefault("ENVIRONMENT", "development"),
		CollectorURL:             getEnvWithDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "signoz-otel-collector.signoz-helm:4317"),
		InsecureMode:             getEnvBool("OTEL_EXPORTER_OTLP_INSECURE", false),
		TraceSampleRate:          getEnvFloat("OTEL_TRACE_SAMPLE_RATE", -1.0),      // -1 means use adaptive
		MetricExportInterval:     getEnvDuration("OTEL_METRIC_EXPORT_INTERVAL", 0), // 0 means use adaptive
		ResourceDetectionTimeout: getEnvDuration("OTEL_RESOURCE_DETECTION_TIMEOUT", 5*time.Second),
		BatchTimeout:             getEnvDuration("OTEL_BSP_SCHEDULE_DELAY", 5*time.Second),
		BatchSize:                getEnvInt("OTEL_BSP_MAX_EXPORT_BATCH_SIZE", 512),
		MaxQueueSize:             getEnvInt("OTEL_BSP_MAX_QUEUE_SIZE", 2048),
		ExportTimeout:            getEnvDuration("OTEL_BSP_EXPORT_TIMEOUT", 30*time.Second),
		EnableLogging:            getEnvBool("OTEL_LOGS_ENABLED", true),
		EnableMetrics:            getEnvBool("OTEL_METRICS_ENABLED", true),
		EnableTracing:            getEnvBool("OTEL_TRACES_ENABLED", true),
		EnableRuntimeMetrics:     getEnvBool("OTEL_RUNTIME_METRICS_ENABLED", true),
		LogLevel:                 getEnvWithDefault("OTEL_LOG_LEVEL", "info"),
		MaxAttributeValueLength:  getEnvInt("OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT", 1024),
		MaxAttributeCount:        getEnvInt("OTEL_ATTRIBUTE_COUNT_LIMIT", 128),
		MaxSpanEvents:            getEnvInt("OTEL_SPAN_EVENT_COUNT_LIMIT", 128),
		MaxSpanLinks:             getEnvInt("OTEL_SPAN_LINK_COUNT_LIMIT", 128),
	}
}

// getEnvBool returns a boolean from environment variable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvFloat returns a float64 from environment variable
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt returns an int from environment variable
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvDuration returns a time.Duration from environment variable
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetSampler returns the appropriate sampler based on configuration
func (c *OTelConfig) GetSampler() sdktrace.Sampler {
	// Use explicit sample rate if provided
	if c.TraceSampleRate >= 0 {
		switch {
		case c.TraceSampleRate <= 0.0:
			return sdktrace.NeverSample()
		case c.TraceSampleRate >= 1.0:
			return sdktrace.AlwaysSample()
		default:
			return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.TraceSampleRate))
		}
	}

	// Adaptive sampling based on environment
	switch c.Environment {
	case "production":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1)) // 10% sampling
	case "staging":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5)) // 50% sampling
	case "development", "local":
		return sdktrace.AlwaysSample() // Full sampling for development
	default:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))
	}
}

// GetMetricInterval returns appropriate metric collection interval
func (c *OTelConfig) GetMetricInterval() time.Duration {
	if c.MetricExportInterval > 0 {
		return c.MetricExportInterval
	}

	// Environment-based defaults
	switch c.Environment {
	case "production":
		return 30 * time.Second
	case "staging":
		return 15 * time.Second
	default:
		return 10 * time.Second
	}
}

// GetSpanProcessorOptions returns span processor options
func (c *OTelConfig) GetSpanProcessorOptions() []sdktrace.BatchSpanProcessorOption {
	return []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithBatchTimeout(c.BatchTimeout),
		sdktrace.WithMaxExportBatchSize(c.BatchSize),
		sdktrace.WithMaxQueueSize(c.MaxQueueSize),
		sdktrace.WithExportTimeout(c.ExportTimeout),
	}
}

// CreateEnhancedResource creates a resource with comprehensive attributes
func (c *OTelConfig) CreateEnhancedResource() (*resource.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.ResourceDetectionTimeout)
	defer cancel()

	// Detected resources
	detectedRes, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		fmt.Printf("Warning: Resource detection failed: %v\n", err)
		detectedRes = resource.Empty()
	}

	// Manual resources with comprehensive attributes
	manualRes := resource.NewWithAttributes(
		resource.Default().SchemaURL(),
		// Service identification
		attribute.String("service.name", c.ServiceName),
		attribute.String("service.version", c.ServiceVersion),
		attribute.String("service.environment", c.Environment),
		attribute.String("service.namespace", getEnvWithDefault("SERVICE_NAMESPACE", "fmdash")),
		attribute.String("service.instance.id", getEnvWithDefault("HOSTNAME", "unknown")),

		// Technology stack
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.name", "opentelemetry"),
		attribute.String("application.type", "football-manager-dashboard"),
		attribute.String("application.component", "api-server"),
		attribute.String("application.layer", "service"),

		// Deployment information
		attribute.String("deployment.environment", c.Environment),
		attribute.String("deployment.region", getEnvWithDefault("DEPLOYMENT_REGION", "")),
		attribute.String("deployment.zone", getEnvWithDefault("DEPLOYMENT_ZONE", "")),
		attribute.String("deployment.cluster", getEnvWithDefault("DEPLOYMENT_CLUSTER", "")),

		// Version control
		attribute.String("vcs.repository.url", getEnvWithDefault("GIT_REPOSITORY_URL", "")),
		attribute.String("vcs.revision", getEnvWithDefault("GIT_COMMIT_SHA", "")),
		attribute.String("vcs.branch", getEnvWithDefault("GIT_BRANCH", "")),
		attribute.String("vcs.tag", getEnvWithDefault("GIT_TAG", "")),

		// Build information
		attribute.String("build.id", getEnvWithDefault("BUILD_ID", "")),
		attribute.String("build.timestamp", getEnvWithDefault("BUILD_TIMESTAMP", "")),
		attribute.String("build.pipeline", getEnvWithDefault("BUILD_PIPELINE", "")),

		// Configuration
		attribute.Float64("otel.trace.sample_rate", c.TraceSampleRate),
		attribute.String("otel.metric.export_interval", c.GetMetricInterval().String()),
		attribute.Bool("otel.logging.enabled", c.EnableLogging),
		attribute.Bool("otel.metrics.enabled", c.EnableMetrics),
		attribute.Bool("otel.tracing.enabled", c.EnableTracing),
	)

	// Merge resources
	finalRes, err := resource.Merge(detectedRes, manualRes)
	if err != nil {
		return nil, fmt.Errorf("failed to merge resources: %w", err)
	}

	return finalRes, nil
}

// Validate checks if the configuration is valid
func (c *OTelConfig) Validate() error {
	if c.ServiceName == "" {
		return apperrors.ErrServiceNameEmpty
	}
	if c.CollectorURL == "" {
		return apperrors.ErrCollectorURLEmpty
	}
	if c.TraceSampleRate < -1.0 || c.TraceSampleRate > 1.0 {
		return apperrors.WrapErrInvalidTraceSampleRate(c.TraceSampleRate)
	}
	if c.BatchSize <= 0 {
		return apperrors.WrapErrInvalidBatchSize(c.BatchSize)
	}
	if c.MaxQueueSize <= 0 {
		return apperrors.WrapErrInvalidMaxQueueSize(c.MaxQueueSize)
	}
	return nil
}

// Print logs the configuration (sensitive values redacted)
func (c *OTelConfig) Print() {
	fmt.Printf("OpenTelemetry Configuration:\n")
	fmt.Printf("  Service Name: %s\n", c.ServiceName)
	fmt.Printf("  Service Version: %s\n", c.ServiceVersion)
	fmt.Printf("  Environment: %s\n", c.Environment)
	fmt.Printf("  Collector URL: %s\n", c.CollectorURL)
	fmt.Printf("  Insecure Mode: %t\n", c.InsecureMode)
	fmt.Printf("  Trace Sample Rate: %.3f\n", c.TraceSampleRate)
	fmt.Printf("  Metric Export Interval: %s\n", c.GetMetricInterval())
	fmt.Printf("  Logging Enabled: %t\n", c.EnableLogging)
	fmt.Printf("  Metrics Enabled: %t\n", c.EnableMetrics)
	fmt.Printf("  Tracing Enabled: %t\n", c.EnableTracing)
	fmt.Printf("  Runtime Metrics: %t\n", c.EnableRuntimeMetrics)
}
