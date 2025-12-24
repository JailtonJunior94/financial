package test_helpers

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
)

// AddUserToRequest adiciona um usuário ao contexto da requisição usando o middleware
func AddUserToRequest(r *http.Request, user *auth.User) *http.Request {
	ctx := middlewares.AddUserToContext(r.Context(), user)
	return r.WithContext(ctx)
}
