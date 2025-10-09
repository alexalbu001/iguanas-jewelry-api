package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Telemetry struct {
	tracerProvider *trace.TracerProvider
	metricProvider *metric.MeterProvider
}

func InitTelemetry(ctx context.Context, serviceName, version, environment, otelEndpoint string) (*Telemetry, error) {
	// 1. Create Resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// 2. Create exporters
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// 3. Create trace provider - only connect to OTEL collector if endpoint is configured
	var tp *trace.TracerProvider

	if otelEndpoint != "" {
		// OTEL collector endpoint is configured, create exporter
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(otelEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}

		tp = trace.NewTracerProvider(
			trace.WithBatcher(traceExporter),
			trace.WithResource(res),
		)
	} else {
		// No OTEL collector configured, use no-op provider (no errors, no export attempts)
		tp = trace.NewTracerProvider(
			trace.WithResource(res),
		)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)

	// 4. Register globally
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	// 5. Set up propagatio. Only once needed
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{
		tracerProvider: tp,
		metricProvider: mp,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	err := t.tracerProvider.Shutdown(ctx)
	if err != nil {
		return err
	}

	if err := t.metricProvider.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
