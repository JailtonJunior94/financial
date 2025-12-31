package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

type (
	PanicRecoverMiddleware interface {
		Recover(next http.Handler) http.Handler
	}

	panicRecoverMiddleware struct {
		o11y         observability.Observability
		errorHandler httperrors.ErrorHandler
	}
)

func NewPanicRecoverMiddleware(o11y observability.Observability, errorHandler httperrors.ErrorHandler) PanicRecoverMiddleware {
	return &panicRecoverMiddleware{
		o11y:         o11y,
		errorHandler: errorHandler,
	}
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

				// Log stacktrace for panic (critical information)
				stacktrace := string(debug.Stack())
				m.o11y.Logger().Error(
					ctx,
					"panic recovered in middleware",
					observability.Error(err),
					observability.String("stacktrace", stacktrace),
				)

				// Delegate HTTP error response to error handler
				m.errorHandler.HandleError(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
