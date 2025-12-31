package middlewares

import (
	"net/http"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	HTTPMetricsMiddleware interface {
		Metrics(next http.Handler) http.Handler
	}

	httpMetricsMiddleware struct {
		o11y observability.Observability
	}

	responseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func NewHTTPMetricsMiddleware(o11y observability.Observability) (HTTPMetricsMiddleware, error) {
	return &httpMetricsMiddleware{
		o11y: o11y,
	}, nil
}

func (m *httpMetricsMiddleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rw := &responseWriter{w, http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rw, r.WithContext(ctx))

		counter := m.o11y.Metrics().Counter("http.requests", "HTTP request count", "request")
		counter.Add(ctx, 1, observability.Field{
			Key:   "method",
			Value: r.Method,
		}, observability.Field{
			Key:   "uri",
			Value: r.RequestURI,
		}, observability.Field{
			Key:   "statusCode",
			Value: rw.statusCode,
		})

		histogram := m.o11y.Metrics().Histogram("http.request.duration", "HTTP request duration", "ns")
		histogram.Record(
			ctx,
			float64(time.Since(start).Nanoseconds()),
			observability.Field{
				Key:   "method",
				Value: r.Method,
			},
			observability.Field{
				Key:   "uri",
				Value: r.RequestURI,
			},
			observability.Field{
				Key:   "statusCode",
				Value: rw.statusCode,
			},
		)
	})
}
