package auth

import "context"

// AuthenticatedUser representa um usuário autenticado no sistema.
// Contém as informações mínimas necessárias para identificação e autorização.
type AuthenticatedUser struct {
	ID    string
	Email string
	Roles []string
}

// NewAuthenticatedUser cria uma nova instância de AuthenticatedUser.
func NewAuthenticatedUser(id, email string, roles []string) *AuthenticatedUser {
	return &AuthenticatedUser{
		ID:    id,
		Email: email,
		Roles: roles,
	}
}

// TokenValidator define a interface para validação de tokens de autenticação.
// Permite desacoplar o middleware de autenticação de implementações específicas
// (JWT, OAuth2, etc.), facilitando testes e mudanças futuras.
type TokenValidator interface {
	// Validate valida um token e retorna o usuário autenticado.
	// Retorna erro se o token for inválido, expirado ou malformado.
	Validate(ctx context.Context, token string) (*AuthenticatedUser, error)
}
