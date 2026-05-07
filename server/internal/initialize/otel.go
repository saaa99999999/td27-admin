package initialize

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"server/internal/global"
)

// InitTracerProvider initializes the OpenTelemetry tracer provider if tracing is enabled
func InitTracerProvider() (*sdktrace.TracerProvider, error) {
	otelCfg := global.TD27_CONFIG.Observability.Otel

	if !otelCfg.Enabled {
		global.TD27_LOG.Info("OpenTelemetry tracing is disabled")
		return nil, nil
	}

	ctx := context.Background()

	// Create gRPC connection to Jaeger OTLP endpoint
	conn, err := grpc.NewClient(otelCfg.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		global.TD27_LOG.Error("Failed to create OTel gRPC connection", "error", err)
		return nil, err
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		global.TD27_LOG.Error("Failed to create OTel trace exporter", "error", err)
		return nil, err
	}

	// Create resource with service attributes
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", otelCfg.ServiceName),
			attribute.String("deployment.environment", global.TD27_CONFIG.System.Env),
		),
	)
	if err != nil {
		global.TD27_LOG.Error("Failed to create OTel resource", "error", err)
		return nil, err
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(otelCfg.SamplingRate)),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	global.TD27_LOG.Info("OpenTelemetry tracer provider initialized",
		"endpoint", otelCfg.Endpoint,
		"service", otelCfg.ServiceName,
		"sampling_rate", otelCfg.SamplingRate,
	)

	return tp, nil
}
