package middlewares

import (
	"context"
	"net/http"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/auth"
)

type (
	Authorization interface {
		Authorization(next http.Handler) http.Handler
	}

	authorization struct {
		jwt    auth.JwtAdapter
		config *configs.Config
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
		user, err := a.jwt.ValidateToken(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *auth.User {
	raw, _ := ctx.Value(userCtxKey).(*auth.User)
	return raw
}
