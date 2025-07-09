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
		return func(ctx context.Context) error { return nil }
	}

	cfg := LoadOTelConfig()
	if err := cfg.Validate(); err != nil {
		log.Printf("Warning: Invalid OpenTelemetry configuration: %v. OTEL disabled.", err)
		return func(ctx context.Context) error { return nil }
	}

	cfg.Print()

	ctx := context.Background()

	// Create enhanced resources
	resources, err := cfg.CreateEnhancedResource()
	if err != nil {
		log.Printf("Warning: Could not set resources: %v. Using minimal resources.", err)
		// Create a minimal resource instead of failing completely
		resources, _ = resource.New(ctx,
			resource.WithAttributes(
				attribute.String("service.name", cfg.ServiceName),
				attribute.String("service.version", cfg.ServiceVersion),
			),
		)
	}

	// Set up composite propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize tracing (independent of other components)
	var traceExporter *otlptrace.Exporter
	if cfg.EnableTracing {
		traceExporter, err = createTraceExporter(ctx, cfg)
		if err != nil {
			log.Printf("Warning: Failed to create trace exporter: %v. OTEL tracing disabled.", err)
		} else {
			bsp := sdktrace.NewBatchSpanProcessor(traceExporter, cfg.GetSpanProcessorOptions()...)
			tracerProvider := sdktrace.NewTracerProvider(
				sdktrace.WithSampler(cfg.GetSampler()),
				sdktrace.WithSpanProcessor(bsp),
				sdktrace.WithResource(resources),
			)
			otel.SetTracerProvider(tracerProvider)
			log.Printf("✓ OTEL tracing initialized")
		}
	}

	// Initialize metrics (independent of tracing)
	var meterProvider *sdkmetric.MeterProvider
	if cfg.EnableMetrics {
		meterProvider, err = createMeterProvider(ctx, cfg, resources)
		if err != nil {
			log.Printf("Warning: Failed to create meter provider: %v. OTEL metrics disabled.", err)
		} else {
			otel.SetMeterProvider(meterProvider)
			initMetrics()
			log.Printf("✓ OTEL metrics initialized")
		}
	}

	// Initialize runtime metrics (independent of other components)
	if cfg.EnableRuntimeMetrics {
		initRuntimeMetrics()
		log.Printf("✓ OTEL runtime metrics initialized")
	}

	// Initialize logs (independent of tracing and metrics)
	var loggerProvider *sdklog.LoggerProvider
	if cfg.EnableLogging {
		loggerProvider, err = createLoggerProvider(ctx, cfg, resources)
		if err != nil {
			log.Printf("Warning: Failed to create logger provider: %v. OTEL logs disabled.", err)
		} else {
			slog.SetDefault(slog.New(NewOTLPHandler(loggerProvider)))
			log.Printf("✓ OTEL logging initialized - logs will be sent to SignOz")
		}
	}

	return func(ctx context.Context) error {
		if traceExporter != nil {
			if err := traceExporter.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down trace exporter: %v", err)
			}
		}
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down meter provider: %v", err)
			}
		}
		if loggerProvider != nil {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down log provider: %v", err)
			}
		}
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
