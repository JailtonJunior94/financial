package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	// Authorization define a interface para o middleware de autenticação.
	Authorization interface {
		Authorization(next http.Handler) http.Handler
	}

	authorization struct {
		validator    auth.TokenValidator
		o11y         observability.Observability
		errorHandler httperrors.ErrorHandler
	}

	// contextKey é um tipo privado para evitar colisões no contexto.
	contextKey struct {
		name string
	}
)

var userCtxKey = &contextKey{"authenticated_user"}

// NewAuthorization cria uma nova instância do middleware de autenticação.
// Requer um TokenValidator para validar tokens e observability para logs/métricas.
func NewAuthorization(validator auth.TokenValidator, o11y observability.Observability, errorHandler httperrors.ErrorHandler) Authorization {
	return &authorization{
		validator:    validator,
		o11y:         o11y,
		errorHandler: errorHandler,
	}
}

// Authorization é o middleware que valida tokens JWT e injeta o usuário autenticado no contexto.
// Fluxo:
// 1. Extrai o Bearer Token do header Authorization
// 2. Valida o formato do header
// 3. Valida o token usando o TokenValidator
// 4. Injeta o usuário autenticado no contexto
// 5. Chama o próximo handler.
func (a *authorization) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extrai o header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.errorHandler.HandleError(w, r, customerrors.ErrMissingAuthHeader)
			return
		}

		// Valida o formato "Bearer <token>"
		token, err := extractBearerToken(authHeader)
		if err != nil {
			a.errorHandler.HandleError(w, r, err)
			return
		}

		// Valida o token e obtém o usuário
		user, err := a.validator.Validate(ctx, token)
		if err != nil {
			a.errorHandler.HandleError(w, r, err)
			return
		}

		// Injeta o usuário no contexto e chama o próximo handler
		ctx = AddUserToContext(ctx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken extrai o token do header Authorization.
// Espera o formato: "Bearer <token>"
// Retorna erro se:
// - O formato não for "Bearer <token>"
// - O token estiver vazio.
func extractBearerToken(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	// Verifica se o header começa com "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", customerrors.ErrInvalidAuthFormat
	}

	// Extrai o token (tudo após "Bearer ")
	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", customerrors.ErrEmptyToken
	}

	return token, nil
}

// GetUserFromContext recupera o usuário autenticado do contexto.
// Retorna erro se o usuário não estiver presente no contexto.
func GetUserFromContext(ctx context.Context) (*auth.AuthenticatedUser, error) {
	user, ok := ctx.Value(userCtxKey).(*auth.AuthenticatedUser)
	if !ok || user == nil {
		return nil, customerrors.ErrUnauthorized
	}
	return user, nil
}

// AddUserToContext adiciona um usuário autenticado ao contexto.
// Útil para testes e situações onde o usuário já foi validado externamente.
func AddUserToContext(ctx context.Context, user *auth.AuthenticatedUser) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}
