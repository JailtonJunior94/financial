package middlewares

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type TracingMiddleware interface {
	Tracing(next http.Handler) http.Handler
}

type tracingMiddleware struct {
	tracer trace.Tracer
}

func NewTracingMiddleware(tracer trace.Tracer) TracingMiddleware {
	return &tracingMiddleware{tracer: tracer}
}

func (m *tracingMiddleware) Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := m.tracer.Start(r.Context(), spanName)
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
