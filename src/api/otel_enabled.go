//go:build !no_otel

package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

var (
	serviceName  = getEnvWithDefault("SERVICE_NAME", "v2fmdash-api")
	collectorURL = getEnvWithDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "signoz.signoz:4317")
	insecure     = getEnvWithDefault("INSECURE_MODE", "true")
)

func initOTel() func(context.Context) error {
	var secureOption otlptracegrpc.Option

	if strings.ToLower(insecure) == "false" || insecure == "0" || strings.ToLower(insecure) == "f" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Could not set resources: %v", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	// Initialize metrics
	var metricSecureOption otlpmetricgrpc.Option
	if strings.ToLower(insecure) == "false" || insecure == "0" || strings.ToLower(insecure) == "f" {
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
		log.Fatalf("Failed to create metric exporter: %v", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resources),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)
	otel.SetMeterProvider(meterProvider)

	return func(ctx context.Context) error {
		if err := exporter.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down trace exporter: %v", err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
		return nil
	}
}

func wrapHandler(handler http.Handler, operationName string) http.Handler {
	return otelhttp.NewHandler(handler, operationName)
}