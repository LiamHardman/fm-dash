//go:build !no_otel

package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

var (
	serviceName  = getEnvWithDefault("SERVICE_NAME", "v2fmdash-api")
	collectorURL = getEnvWithDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "signoz.signoz:4317")
	insecure     = getEnvWithDefault("INSECURE_MODE", "true")
	sampleRate   = getEnvWithDefault("OTEL_TRACE_SAMPLE_RATE", "1.0") // Default to 100% sampling
)

// getSampler returns the appropriate sampler based on configuration
func getSampler() sdktrace.Sampler {
	if rate := sampleRate; rate != "" {
		if floatRate, err := strconv.ParseFloat(rate, 64); err == nil {
			switch {
			case floatRate <= 0.0:
				return sdktrace.NeverSample()
			case floatRate >= 1.0:
				return sdktrace.AlwaysSample()
			default:
				return sdktrace.TraceIDRatioBased(floatRate)
			}
		}
	}
	return sdktrace.AlwaysSample() // Default fallback
}

func initOTel() func(context.Context) error {
	var secureOption otlptracegrpc.Option

	if strings.EqualFold(insecure, "false") || insecure == "0" || strings.EqualFold(insecure, "f") {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	// Create trace exporter
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)
	if err != nil {
		log.Printf("Warning: Failed to create trace exporter: %v. OTEL tracing disabled.", err)
		return func(ctx context.Context) error { return nil }
	}

	// Create shared resources
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", getEnvWithDefault("SERVICE_VERSION", "v1.0.0")),
			attribute.String("service.environment", getEnvWithDefault("ENVIRONMENT", "development")),
			attribute.String("service.namespace", getEnvWithDefault("SERVICE_NAMESPACE", "fmdash")),
			attribute.String("library.language", "go"),
			attribute.String("library.name", "v2fmdash-api"),
			attribute.String("deployment.environment", getEnvWithDefault("DEPLOYMENT_ENV", "local")),
			// Add instance information
			attribute.String("service.instance.id", getEnvWithDefault("HOSTNAME", "unknown")),
			// Add application-specific attributes
			attribute.String("application.type", "football-manager-dashboard"),
			attribute.String("application.component", "api-server"),
		),
	)
	if err != nil {
		log.Printf("Warning: Could not set resources: %v. OTEL disabled.", err)
		return func(ctx context.Context) error { return nil }
	}

	// Initialize tracing
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(getSampler()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	// Initialize metrics
	var metricSecureOption otlpmetricgrpc.Option
	if strings.EqualFold(insecure, "false") || insecure == "0" || strings.EqualFold(insecure, "f") {
		metricSecureOption = otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		metricSecureOption = otlpmetricgrpc.WithInsecure()
	}

	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		metricSecureOption,
		otlpmetricgrpc.WithEndpoint(collectorURL),
	)
	if err != nil {
		log.Printf("Warning: Failed to create metric exporter: %v. OTEL metrics disabled.", err)
		// Continue without metrics, only traces
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resources),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)
	otel.SetMeterProvider(meterProvider)

	// Initialize logs
	var logSecureOption otlploggrpc.Option
	if strings.EqualFold(insecure, "false") || insecure == "0" || strings.EqualFold(insecure, "f") {
		logSecureOption = otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		logSecureOption = otlploggrpc.WithInsecure()
	}

	logExporter, err := otlploggrpc.New(
		context.Background(),
		logSecureOption,
		otlploggrpc.WithEndpoint(collectorURL),
	)
	if err != nil {
		log.Printf("Warning: Failed to create log exporter: %v. OTEL logs disabled.", err)
		// Continue without log streaming
		if exporter != nil {
			return func(ctx context.Context) error {
				if err := exporter.Shutdown(ctx); err != nil {
					log.Printf("Error shutting down trace exporter: %v", err)
				}
				return nil
			}
		}
		return func(ctx context.Context) error { return nil }
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resources),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)

	// Set up structured logging with OTLP
	handler := NewOTLPHandler(loggerProvider)
	slog.SetDefault(slog.New(handler))

	return func(ctx context.Context) error {
		if err := exporter.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down trace exporter: %v", err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
		if err := loggerProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down log provider: %v", err)
		}
		return nil
	}
}

func wrapHandler(handler http.Handler, operationName string) http.Handler {
	return otelhttp.NewHandler(handler, operationName)
}
