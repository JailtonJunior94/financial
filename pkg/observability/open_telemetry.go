package observability

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	Observability interface {
		Tracer() trace.Tracer
		MeterProvider() *metric.MeterProvider
		TracerProvider() *sdktrace.TracerProvider
	}

	Option        func(observability *observability)
	observability struct {
		serviceName    string
		serviceVersion string
		tracer         trace.Tracer
		resource       *resource.Resource
		meterProvider  *metric.MeterProvider
		tracerProvider *sdktrace.TracerProvider
	}
)

func NewObservability(options ...Option) Observability {
	observability := &observability{}
	for _, option := range options {
		option(observability)
	}
	return observability
}

func NewDevelopmentObservability(serviceName string) Observability {
	return NewObservability(
		WithServiceName(serviceName),
		WithServiceVersion("1.0.0"),
		WithResource(),
		WithStdoutTracerProvider(),
		WithStdoutMeterProvider(),
	)
}

func (o *observability) Tracer() trace.Tracer {
	return o.tracer
}

func (o *observability) MeterProvider() *metric.MeterProvider {
	return o.meterProvider
}

func (o *observability) TracerProvider() *sdktrace.TracerProvider {
	return o.tracerProvider
}

func WithServiceName(serviceName string) Option {
	return func(observability *observability) {
		observability.serviceName = serviceName
	}
}

func WithServiceVersion(serviceVersion string) Option {
	return func(observability *observability) {
		observability.serviceVersion = serviceVersion
	}
}

func WithResource() Option {
	return func(observability *observability) {
		resource, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(observability.serviceName),
				semconv.ServiceVersion(observability.serviceVersion),
			),
		)

		if err != nil {
			log.Fatal(err)
		}
		observability.resource = resource
	}
}

func WithTracerProvider(ctx context.Context, endpoint string) Option {
	return func(observability *observability) {
		traceExporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint),
		)

		if err != nil {
			log.Fatal(err)
		}

		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(traceExporter),
			sdktrace.WithResource(observability.resource),
		)

		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		observability.tracer = tracerProvider.Tracer(observability.serviceName)
		observability.tracerProvider = tracerProvider
	}
}

func WithStdoutTracerProvider() Option {
	return func(observability *observability) {
		exporter, err := stdouttrace.New()
		if err != nil {
			log.Fatalf("failed to initialize stdout export pipeline: %v", err)
		}

		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(exporter),
		)

		observability.tracer = tracerProvider.Tracer(observability.serviceName)
		observability.tracerProvider = tracerProvider
	}
}

func WithMeterProvider(ctx context.Context, endpoint string) Option {
	return func(observability *observability) {
		metricExporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpoint(endpoint),
		)

		if err != nil {
			log.Fatal(err)
		}

		meterProvider := metric.NewMeterProvider(
			metric.WithResource(observability.resource),
			metric.WithReader(metric.NewPeriodicReader(
				metricExporter,
				metric.WithInterval(2*time.Second)),
			),
		)

		otel.SetMeterProvider(meterProvider)
		observability.meterProvider = meterProvider
	}
}

func WithStdoutMeterProvider() Option {
	return func(observability *observability) {
		exporter, err := stdoutmetric.New()
		if err != nil {
			log.Fatalf("failed to initialize stdout export pipeline: %v", err)
		}

		meterProvider := metric.NewMeterProvider(
			metric.WithResource(observability.resource),
			metric.WithReader(metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(2*time.Second)),
			),
		)

		otel.SetMeterProvider(meterProvider)
		observability.meterProvider = meterProvider
	}
}
