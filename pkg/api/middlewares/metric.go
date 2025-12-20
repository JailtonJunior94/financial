package middlewares

import (
	"net/http"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	HTTPMetricsMiddleware interface {
		Metrics(next http.Handler) http.Handler
	}

	httpMetricsMiddleware struct {
		o11y o11y.Telemetry
	}

	responseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func NewHTTPMetricsMiddleware(o11y o11y.Telemetry) (HTTPMetricsMiddleware, error) {

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

		m.o11y.Metrics().AddCounter(ctx, "http.requests", 1, o11y.Attribute{
			Key:   "method",
			Value: r.Method,
		}, o11y.Attribute{
			Key:   "uri",
			Value: r.RequestURI,
		}, o11y.Attribute{
			Key:   "statusCode",
			Value: rw.statusCode,
		})

		m.o11y.Metrics().RecordHistogram(
			ctx,
			"http.request.duration",
			float64(time.Since(start).Nanoseconds()),
			o11y.Attribute{
				Key:   "method",
				Value: r.Method,
			},
			o11y.Attribute{
				Key:   "uri",
				Value: r.RequestURI,
			},
			o11y.Attribute{
				Key:   "statusCode",
				Value: rw.statusCode,
			},
		)
	})
}
