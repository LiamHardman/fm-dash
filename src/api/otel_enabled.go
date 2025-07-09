package main

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

func initOTel() func(context.Context) error {
	if !otelEnabled {
		LogInfo("🔴 OTEL: OpenTelemetry is disabled (OTEL_ENABLED=false)")
		return func(ctx context.Context) error { return nil }
	}

	LogInfo("🟡 OTEL: Starting OpenTelemetry initialization...")

	cfg := LoadOTelConfig()
	if err := cfg.Validate(); err != nil {
		LogWarn("🔴 OTEL: Invalid OpenTelemetry configuration: %v. OTEL disabled.", err)
		return func(ctx context.Context) error { return nil }
	}

	LogInfo("🟢 OTEL: Configuration validated successfully")
	cfg.Print()

	ctx := context.Background()

	// Create enhanced resources
	LogInfo("🟡 OTEL: Creating enhanced resources...")
	resources, err := cfg.CreateEnhancedResource()
	if err != nil {
		LogWarn("🟡 OTEL: Warning - Could not set resources: %v. Using minimal resources.", err)
		// Create a minimal resource instead of failing completely
		resources, _ = resource.New(ctx,
			resource.WithAttributes(
				attribute.String("service.name", cfg.ServiceName),
				attribute.String("service.version", cfg.ServiceVersion),
			),
		)
	}
	LogInfo("🟢 OTEL: Resources created successfully")

	// Set up composite propagator
	LogInfo("🟡 OTEL: Setting up text map propagator...")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	LogInfo("🟢 OTEL: Text map propagator configured")

	// Initialize tracing
	var tracerProvider *sdktrace.TracerProvider
	if cfg.EnableTracing {
		LogInfo("🟡 OTEL: Initializing tracing (endpoint: %s)...", cfg.CollectorURL)
		traceExporter, err := createTraceExporter(ctx, cfg)
		if err != nil {
			LogWarn("🟡 OTEL: Warning - Failed to create trace exporter: %v. Tracing disabled.", err)
		} else {
			LogInfo("🟢 OTEL: Trace exporter created successfully")

			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter, cfg.GetSpanProcessorOptions()...),
				sdktrace.WithResource(resources),
			)
			otel.SetTracerProvider(tracerProvider)
			LogInfo("🟢 OTEL: ✓ OTEL tracing initialized")
		}
	} else {
		LogInfo("🟡 OTEL: Tracing disabled by configuration")
	}

	// Initialize metrics
	var meterProvider *sdkmetric.MeterProvider
	if cfg.EnableMetrics {
		LogInfo("🟡 OTEL: Initializing metrics (endpoint: %s)...", cfg.CollectorURL)
		meterProvider, err := createMeterProvider(ctx, cfg, resources)
		if err != nil {
			LogWarn("🟡 OTEL: Warning - Failed to create meter provider: %v. Metrics disabled.", err)
		} else {
			LogInfo("🟢 OTEL: Meter provider created successfully")
			otel.SetMeterProvider(meterProvider)
			initMetrics()
			LogInfo("🟢 OTEL: Enhanced OpenTelemetry metrics initialized")
			LogInfo("🟢 OTEL: OpenTelemetry metrics initialized successfully")
			LogInfo("🟢 OTEL: ✓ OTEL metrics initialized")

			// Initialize runtime metrics if enabled
			if cfg.EnableRuntimeMetrics {
				LogInfo("🟡 OTEL: Initializing runtime metrics...")
				initRuntimeMetrics()
				LogInfo("🟢 OTEL: Go runtime metrics initialized")
				LogInfo("🟢 OTEL: ✓ OTEL runtime metrics initialized")
			}
		}
	} else {
		LogInfo("🟡 OTEL: Metrics disabled by configuration")
	}

	// Initialize logging
	var loggerProvider *sdklog.LoggerProvider
	if cfg.EnableLogging {
		LogInfo("🟡 OTEL: Initializing logging (endpoint: %s)...", cfg.CollectorURL)
		loggerProvider, err := createLoggerProvider(ctx, cfg, resources)
		if err != nil {
			LogWarn("🟡 OTEL: Warning - Failed to create logger provider: %v. Logging disabled.", err)
		} else {
			LogInfo("🟢 OTEL: Logger provider created successfully")

			// Set up slog with OTLP handler
			handler := NewOTLPHandler(loggerProvider)
			logger := slog.New(handler)
			slog.SetDefault(logger)

			LogInfo("🟢 OTEL: slog configured with OTLP handler")
			slog.Info("✓ OTEL logging initialized - logs will be sent to SignOz")
		}
	} else {
		LogInfo("🟡 OTEL: Logging disabled by configuration")
	}

	LogInfo("🟢 OTEL: OpenTelemetry initialization completed!")

	// Return cleanup function
	return func(ctx context.Context) error {
		LogInfo("🟡 OTEL: Starting cleanup...")
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				LogWarn("🔴 OTEL: Error shutting down tracer provider: %v", err)
			} else {
				LogInfo("🟢 OTEL: Tracer provider shut down successfully")
			}
		}
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				LogWarn("🔴 OTEL: Error shutting down meter provider: %v", err)
			} else {
				LogInfo("🟢 OTEL: Meter provider shut down successfully")
			}
		}
		if loggerProvider != nil {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				LogWarn("🔴 OTEL: Error shutting down log provider: %v", err)
			} else {
				LogInfo("🟢 OTEL: Logger provider shut down successfully")
			}
		}
		LogInfo("🟢 OTEL: Cleanup completed")
		return nil
	}
}

func createTraceExporter(ctx context.Context, cfg *OTelConfig) (*otlptrace.Exporter, error) {
	var grpcOpts []otlptracegrpc.Option
	if cfg.InsecureMode {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	} else {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}
	grpcOpts = append(grpcOpts, otlptracegrpc.WithEndpoint(cfg.CollectorURL))

	return otlptrace.New(ctx, otlptracegrpc.NewClient(grpcOpts...))
}

func createMeterProvider(ctx context.Context, cfg *OTelConfig, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var grpcOpts []otlpmetricgrpc.Option
	if cfg.InsecureMode {
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithInsecure())
	} else {
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}
	grpcOpts = append(grpcOpts, otlpmetricgrpc.WithEndpoint(cfg.CollectorURL))

	metricExporter, err := otlpmetricgrpc.New(ctx, grpcOpts...)
	if err != nil {
		return nil, err
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			metricExporter,
			sdkmetric.WithInterval(cfg.GetMetricInterval()),
		)),
	), nil
}

func createLoggerProvider(ctx context.Context, cfg *OTelConfig, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	var grpcOpts []otlploggrpc.Option
	if cfg.InsecureMode {
		grpcOpts = append(grpcOpts, otlploggrpc.WithInsecure())
	} else {
		grpcOpts = append(grpcOpts, otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}
	grpcOpts = append(grpcOpts, otlploggrpc.WithEndpoint(cfg.CollectorURL))

	logExporter, err := otlploggrpc.New(ctx, grpcOpts...)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	), nil
}

func wrapHandler(handler http.Handler, operationName string) http.Handler {
	if !otelEnabled {
		return handler
	}
	return otelhttp.NewHandler(handler, operationName)
}
