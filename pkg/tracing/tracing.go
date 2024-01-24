package tracing

import (
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Provider struct {
	ServiceName      string
	ServiceVersion   string
	ExporterEndpoint string
}

func NewProvider(serviceName, serviceVersion, exporterEndpoint string) *Provider {
	return &Provider{
		ServiceName:      serviceName,
		ServiceVersion:   serviceVersion,
		ExporterEndpoint: exporterEndpoint,
	}
}

func (o *Provider) GetTracer() trace.Tracer {
	var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)
	exporter, err := zipkin.New(
		o.ExporterEndpoint,
		zipkin.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(o.ServiceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer := otel.Tracer(o.ServiceName)
	return tracer
}
