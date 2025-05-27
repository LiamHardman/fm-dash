// src/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof" // For profiling, if needed
	"os"
	"path/filepath"
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

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

func main() {
	// Initialize OpenTelemetry (tracing and metrics)
	cleanup := initOTel()
	defer cleanup(context.Background())

	// Initialize storage system
	InitStore()
	
	// Start automatic cleanup scheduler for old datasets
	StartCleanupScheduler()
	// Serve the main index.html page (assuming it's built into a 'public' or 'dist' folder by Vue)
	// Adjust the path according to your frontend build output.
	// If Vue serves on a different port (e.g., 3000) and proxies API calls, this might not be needed here.
	// However, if Go is serving everything, ensure this path is correct.
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Fallback to serving static files if not root, or let fsPublic handle it
			// For a Single Page Application, usually all non-API routes serve index.html
			// http.ServeFile(w, r, filepath.Join(".", "public", "index.html")) // Example
			// For now, let's keep it simple: only "/" serves index.html directly.
			// Other static assets are handled by fsPublic.
			// If you have client-side routing, you'll need a more sophisticated setup here.
			http.NotFound(w, r)
			return
		}
		// This path assumes your Go executable is run from the root of the Go API module,
		// and 'public/index.html' is relative to that.
		// If your Vue build output is elsewhere, adjust this path.
		// For example, if Vue builds to '../dist', it might be filepath.Join("..", "dist", "index.html")
		http.ServeFile(w, r, filepath.Join(".", "public", "index.html"))
	}), "index"))

	// Serve static files from the "public" directory (e.g., CSS, JS from Vue build, weight JSONs)
	// This path also assumes the 'public' folder is relative to where the Go executable is run.
	fsPublic := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fsPublic))
	// If your Vue assets are in 'dist/assets', you might need:
	// fsAssets := http.FileServer(http.Dir("./dist/assets"))
	// http.Handle("/assets/", http.StripPrefix("/assets/", fsAssets))

	// API endpoint for file uploads
	http.Handle("/upload", otelhttp.NewHandler(http.HandlerFunc(uploadHandler), "upload"))

	// API endpoint for retrieving player data
	http.Handle("/api/players/", otelhttp.NewHandler(http.HandlerFunc(playerDataHandler), "player-data"))

	// New API endpoint for retrieving available roles
	http.Handle("/api/roles", otelhttp.NewHandler(http.HandlerFunc(rolesHandler), "roles"))

	// API endpoint for retrieving leagues data
	http.Handle("/api/leagues/", otelhttp.NewHandler(http.HandlerFunc(leaguesHandler), "leagues"))

	// API endpoint for retrieving teams data for a specific league
	http.Handle("/api/teams/", otelhttp.NewHandler(http.HandlerFunc(teamsHandler), "teams"))

	// API endpoint for updating player percentiles with division filtering
	http.Handle("/api/percentiles/", otelhttp.NewHandler(http.HandlerFunc(percentilesHandler), "percentiles"))

	// Determine port for HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091" // Default port for the Go API
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
