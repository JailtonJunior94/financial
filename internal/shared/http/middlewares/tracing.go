package middlewares

import (
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracingMiddleware interface {
	Tracing(next http.Handler) http.Handler
}

type tracingMiddleware struct {
	observability observability.Observability
}

func NewTracingMiddleware(observability observability.Observability) TracingMiddleware {
	return &tracingMiddleware{observability: observability}
}

func (m *tracingMiddleware) Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := m.observability.Tracer().Start(r.Context(), spanName)
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
