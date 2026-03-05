package middlewares

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/go-chi/chi/v5"
)

type (
	// ResourceOwnership define a interface para o middleware de controle de acesso ao recurso.
	ResourceOwnership interface {
		Ownership(paramName string) func(next http.Handler) http.Handler
	}

	resourceOwnership struct {
		o11y         observability.Observability
		errorHandler httperrors.ErrorHandler
	}
)

// NewResourceOwnership cria uma nova instância do middleware de ownership.
func NewResourceOwnership(o11y observability.Observability, errorHandler httperrors.ErrorHandler) ResourceOwnership {
	return &resourceOwnership{
		o11y:         o11y,
		errorHandler: errorHandler,
	}
}

// Ownership retorna um middleware que verifica se o user_id do token corresponde ao {paramName} da rota.
// Retorna 403 Forbidden se os IDs forem diferentes.
func (m *resourceOwnership) Ownership(paramName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, err := GetUserFromContext(ctx)
			if err != nil {
				m.errorHandler.HandleError(w, r, customerrors.ErrUnauthorized)
				return
			}

			paramID := chi.URLParam(r, paramName)
			if user.ID != paramID {
				m.o11y.Logger().Warn(ctx, "ownership_denied",
					observability.String("user_id", user.ID),
					observability.String("resource_id", paramID),
					observability.String("param_name", paramName),
				)
				m.errorHandler.HandleError(w, r, customerrors.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
