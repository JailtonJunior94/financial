package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jailtonjunior94/financial/pkg/o11y"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type (
	PanicRecoverMiddleware interface {
		Recover(next http.Handler) http.Handler
	}

	panicRecoverMiddleware struct {
		o11y o11y.Observability
	}
)

func NewPanicRecoverMiddleware(o11y o11y.Observability) PanicRecoverMiddleware {
	return &panicRecoverMiddleware{o11y: o11y}
}

func (m *panicRecoverMiddleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				_, span := m.o11y.Start(r.Context(), "panic_recover_middleware.recover")
				defer span.End()

				err, ok := err.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				errFormated := fmt.Sprintf("stacktrace from panic: \n" + string(debug.Stack()))
				span.AddStatus(o11y.Error, err.Error())
				span.AddAttributes(
					o11y.Attributes{Key: "stacktrace", Value: errFormated},
					o11y.Attributes{Key: "error", Value: err},
				)
				responses.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
