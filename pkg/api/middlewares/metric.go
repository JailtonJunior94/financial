package middlewares

import (
	"net/http"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type (
	HTTPMetricsMiddleware interface {
		Metrics(next http.Handler) http.Handler
	}

	httpMetricsMiddleware struct {
		meter           metric.Meter
		requestCounter  metric.Int64Counter
		requestDuration metric.Float64Histogram
	}

	responseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func NewHTTPMetricsMiddleware(o11y o11y.Observability) (HTTPMetricsMiddleware, error) {
	meter := o11y.MeterProvider().Meter("financial")

	counter, err := meter.Int64Counter("http.requests", metric.WithDescription("HTTP Requests Counter"))
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram("http.request.duration", metric.WithDescription("HTTP Request Duration"))
	if err != nil {
		return nil, err
	}

	return &httpMetricsMiddleware{
		meter:           meter,
		requestCounter:  counter,
		requestDuration: duration,
	}, nil
}

func (m *httpMetricsMiddleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rw := &responseWriter{w, http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rw, r.WithContext(ctx))

		m.requestDuration.Record(
			ctx,
			float64(time.Since(start).Nanoseconds()),
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("uri", r.RequestURI),
				attribute.Int("statusCode", rw.statusCode),
			),
		)

		m.requestCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("uri", r.RequestURI),
				attribute.Int("statusCode", rw.statusCode),
			),
		)
	})
}
