package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	PanicRecoverMiddleware interface {
		Recover(next http.Handler) http.Handler
	}

	panicRecoverMiddleware struct {
		o11y o11y.Telemetry
	}
)

func NewPanicRecoverMiddleware(o11y o11y.Telemetry) PanicRecoverMiddleware {
	return &panicRecoverMiddleware{o11y: o11y}
}

func (m *panicRecoverMiddleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				ctx := r.Context()
				_, span := m.o11y.Tracer().Start(ctx, "panic_recover_middleware.recover")
				defer span.End()

				err, ok := panicErr.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", panicErr)
				}

				errFormated := fmt.Sprintf("stacktrace from panic: \n %s", string(debug.Stack()))
				m.o11y.Logger().Error(
					ctx, err,
					"panic recovered in middleware",
					o11y.Field{Key: "stacktrace", Value: errFormated},
				)
				responses.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
