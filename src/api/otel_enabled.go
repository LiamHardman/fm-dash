package main

import (
	"context"
	"log"
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
		log.Printf("🔴 OTEL: OpenTelemetry is disabled (OTEL_ENABLED=false)")
		return func(ctx context.Context) error { return nil }
	}

	log.Printf("🟡 OTEL: Starting OpenTelemetry initialization...")

	cfg := LoadOTelConfig()
	if err := cfg.Validate(); err != nil {
		log.Printf("🔴 OTEL: Invalid OpenTelemetry configuration: %v. OTEL disabled.", err)
		return func(ctx context.Context) error { return nil }
	}

	log.Printf("🟢 OTEL: Configuration validated successfully")
	cfg.Print()

	ctx := context.Background()

	// Create enhanced resources
	log.Printf("🟡 OTEL: Creating enhanced resources...")
	resources, err := cfg.CreateEnhancedResource()
	if err != nil {
		log.Printf("🟡 OTEL: Warning - Could not set resources: %v. Using minimal resources.", err)
		// Create a minimal resource instead of failing completely
		resources, _ = resource.New(ctx,
			resource.WithAttributes(
				attribute.String("service.name", cfg.ServiceName),
				attribute.String("service.version", cfg.ServiceVersion),
			),
		)
	}
	log.Printf("🟢 OTEL: Resources created successfully")

	// Set up composite propagator
	log.Printf("🟡 OTEL: Setting up text map propagator...")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	log.Printf("🟢 OTEL: Text map propagator configured")

	// Initialize tracing
	var tracerProvider *sdktrace.TracerProvider
	if cfg.EnableTracing {
		log.Printf("🟡 OTEL: Initializing tracing (endpoint: %s)...", cfg.CollectorURL)
		traceExporter, err := createTraceExporter(ctx, cfg)
		if err != nil {
			log.Printf("🟡 OTEL: Warning - Failed to create trace exporter: %v. Tracing disabled.", err)
		} else {
			log.Printf("🟢 OTEL: Trace exporter created successfully")

			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter, cfg.GetSpanProcessorOptions()...),
				sdktrace.WithResource(resources),
			)
			otel.SetTracerProvider(tracerProvider)
			log.Printf("🟢 OTEL: ✓ OTEL tracing initialized")
		}
	} else {
		log.Printf("🟡 OTEL: Tracing disabled by configuration")
	}

	// Initialize metrics
	var meterProvider *sdkmetric.MeterProvider
	if cfg.EnableMetrics {
		log.Printf("🟡 OTEL: Initializing metrics (endpoint: %s)...", cfg.CollectorURL)
		meterProvider, err := createMeterProvider(ctx, cfg, resources)
		if err != nil {
			log.Printf("🟡 OTEL: Warning - Failed to create meter provider: %v. Metrics disabled.", err)
		} else {
			log.Printf("🟢 OTEL: Meter provider created successfully")
			otel.SetMeterProvider(meterProvider)
			initMetrics()
			log.Printf("🟢 OTEL: Enhanced OpenTelemetry metrics initialized")
			log.Printf("🟢 OTEL: OpenTelemetry metrics initialized successfully")
			log.Printf("🟢 OTEL: ✓ OTEL metrics initialized")

			// Initialize runtime metrics if enabled
			if cfg.EnableRuntimeMetrics {
				log.Printf("🟡 OTEL: Initializing runtime metrics...")
				initRuntimeMetrics()
				log.Printf("🟢 OTEL: Go runtime metrics initialized")
				log.Printf("🟢 OTEL: ✓ OTEL runtime metrics initialized")
			}
		}
	} else {
		log.Printf("🟡 OTEL: Metrics disabled by configuration")
	}

	// Initialize logging
	var loggerProvider *sdklog.LoggerProvider
	if cfg.EnableLogging {
		log.Printf("🟡 OTEL: Initializing logging (endpoint: %s)...", cfg.CollectorURL)
		loggerProvider, err := createLoggerProvider(ctx, cfg, resources)
		if err != nil {
			log.Printf("🟡 OTEL: Warning - Failed to create logger provider: %v. Logging disabled.", err)
		} else {
			log.Printf("🟢 OTEL: Logger provider created successfully")

			// Set up slog with OTLP handler
			handler := NewOTLPHandler(loggerProvider)
			logger := slog.New(handler)
			slog.SetDefault(logger)

			log.Printf("🟢 OTEL: slog configured with OTLP handler")
			slog.Info("✓ OTEL logging initialized - logs will be sent to SignOz")
		}
	} else {
		log.Printf("🟡 OTEL: Logging disabled by configuration")
	}

	log.Printf("🟢 OTEL: OpenTelemetry initialization completed!")

	// Return cleanup function
	return func(ctx context.Context) error {
		log.Printf("🟡 OTEL: Starting cleanup...")
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				log.Printf("🔴 OTEL: Error shutting down tracer provider: %v", err)
			} else {
				log.Printf("🟢 OTEL: Tracer provider shut down successfully")
			}
		}
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				log.Printf("🔴 OTEL: Error shutting down meter provider: %v", err)
			} else {
				log.Printf("🟢 OTEL: Meter provider shut down successfully")
			}
		}
		if loggerProvider != nil {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				log.Printf("🔴 OTEL: Error shutting down log provider: %v", err)
			} else {
				log.Printf("🟢 OTEL: Logger provider shut down successfully")
			}
		}
		log.Printf("🟢 OTEL: Cleanup completed")
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
