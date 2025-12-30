package middlewares

import (
	"context"
	"net/http"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type (
	Authorization interface {
		Authorization(next http.Handler) http.Handler
	}

	authorization struct {
		jwt    auth.JwtAdapter
		config *configs.Config
		o11y   observability.Observability
	}

	contextKey struct {
		name string
	}
)

var userCtxKey = &contextKey{"user"}

func NewAuthorization(config *configs.Config, jwt auth.JwtAdapter) Authorization {
	return &authorization{config: config, jwt: jwt}
}

func (a *authorization) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := a.jwt.ValidateToken(ctx, r.Header.Get("Authorization"))
		if err != nil {
			if a.o11y != nil {
				a.o11y.Logger().Error(ctx, "unauthorized: invalid or missing token", observability.Error(err))
			}
			responses.Error(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*auth.User, error) {
	raw, ok := ctx.Value(userCtxKey).(*auth.User)
	if !ok || raw == nil {
		return nil, customerrors.ErrUnauthorized
	}
	return raw, nil
}

// AddUserToContext adiciona um usuário ao contexto (útil para testes)
func AddUserToContext(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}
