package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/go-chi/chi/v5"
)

type MetricsMiddleware struct {
	requestDuration observability.Histogram
	requestTotal    observability.Counter
	activeRequests  observability.UpDownCounter
}

func NewMetricsMiddleware(o11y observability.Observability) *MetricsMiddleware {
	return &MetricsMiddleware{
		requestDuration: o11y.Metrics().HistogramWithBuckets(
			"financial.http.request.duration.seconds",
			"HTTP request latency in seconds",
			"s",
			[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		),
		requestTotal: o11y.Metrics().Counter(
			"financial.http.requests.total",
			"Total HTTP requests by method, route, and status",
			"request",
		),
		activeRequests: o11y.Metrics().UpDownCounter(
			"financial.http.active_requests",
			"Currently active HTTP requests",
			"request",
		),
	}
}

func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m.activeRequests.Add(r.Context(), 1)
		defer m.activeRequests.Add(r.Context(), -1)

		ww := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		route := m.extractRoute(r)
		method := r.Method
		statusClass := statusClass(ww.statusCode)

		m.requestDuration.Record(r.Context(), duration,
			observability.String("method", method),
			observability.String("route", route),
			observability.String("status_class", statusClass),
		)

		m.requestTotal.Increment(r.Context(),
			observability.String("method", method),
			observability.String("route", route),
			observability.String("status_class", statusClass),
		)
	})
}

func (m *MetricsMiddleware) extractRoute(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
		return rctx.RoutePattern()
	}
	return r.URL.Path
}

func statusClass(code int) string {
	return fmt.Sprintf("%dxx", code/100)
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
