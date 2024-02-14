package middlewares

import (
	"context"
	"net/http"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/authentication"
)

type (
	Authorization interface {
		Authorization(next http.Handler) http.Handler
		GetUserFromContext(ctx context.Context) *authentication.User
	}

	authorization struct {
		jwt    authentication.JwtAdapter
		config *configs.Config
	}

	contextKey struct {
		name string
	}
)

var UserCtxKey = &contextKey{"user"}

func NewAuthorization(config *configs.Config, jwt authentication.JwtAdapter) Authorization {
	return &authorization{config: config, jwt: jwt}
}

func (a *authorization) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.jwt.ValidateToken(r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *authorization) GetUserFromContext(ctx context.Context) *authentication.User {
	raw, _ := ctx.Value(UserCtxKey).(*authentication.User)
	return raw
}
