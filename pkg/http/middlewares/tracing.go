package middlewares

import (
	"fmt"
	"net/http"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracingMiddleware interface {
	Tracing(next http.Handler) http.Handler
}

type tracingMiddleware struct {
	o11y o11y.Observability
}

func NewTracingMiddleware(o11y o11y.Observability) TracingMiddleware {
	return &tracingMiddleware{o11y: o11y}
}

func (m *tracingMiddleware) Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := m.o11y.Tracer().Start(r.Context(), spanName)
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
