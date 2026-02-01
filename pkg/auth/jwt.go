package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrGenerateToken = errors.New("error trying to generate token")
)

type (
	// JwtAdapter é responsável pela geração e validação de tokens JWT.
	// Implementa a interface TokenValidator.
	JwtAdapter interface {
		GenerateToken(ctx context.Context, id, email string) (string, error)
		TokenValidator
	}

	jwtAdapter struct {
		config *configs.Config
		obs    observability.Observability
	}

	// User representa um usuário (mantido para compatibilidade com código existente).
	// Deprecated: Use AuthenticatedUser da interface TokenValidator.
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
)

// NewUser cria uma nova instância de User.
// Deprecated: Use NewAuthenticatedUser.
func NewUser(id, email string) *User {
	return &User{ID: id, Email: email}
}

// NewJwtAdapter cria uma nova instância do adaptador JWT.
func NewJwtAdapter(config *configs.Config, obs observability.Observability) JwtAdapter {
	return &jwtAdapter{config: config, obs: obs}
}

func (j *jwtAdapter) GenerateToken(ctx context.Context, id, email string) (string, error) {
	_, span := j.obs.Tracer().Start(ctx, "jwt_adapter.generate_token")
	defer span.End()

	claims := jwt.MapClaims{
		"sub":   id,
		"email": email,
		"exp":   time.Now().Add(time.Hour * time.Duration(j.config.AuthConfig.AuthTokenDuration)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenSigned, err := token.SignedString([]byte(j.config.AuthConfig.AuthSecretKey))
	if err != nil {
		span.AddEvent(
			"error trying to generate token",
			observability.Field{Key: "e-mail", Value: email},
			observability.Field{Key: "error", Value: err.Error()},
		)
		j.obs.Logger().Error(ctx, "error trying to generate token", observability.Error(err), observability.String("e-mail", email))
		return "", ErrGenerateToken
	}
	return tokenSigned, nil
}

// Validate implementa a interface TokenValidator.
// Valida um token JWT e retorna o usuário autenticado.
// Retorna erros específicos para diferentes cenários de falha.
func (j *jwtAdapter) Validate(ctx context.Context, token string) (*AuthenticatedUser, error) {
	_, span := j.obs.Tracer().Start(ctx, "jwt_adapter.validate")
	defer span.End()

	if token == "" {
		return nil, customerrors.ErrEmptyToken
	}

	secret := []byte(j.config.AuthConfig.AuthSecretKey)
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		// Valida o método de assinatura
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			span.AddEvent(
				"invalid token signing method",
				observability.Field{Key: "method", Value: fmt.Sprintf("%v", t.Header["alg"])},
			)
			j.obs.Logger().Error(
				ctx,
				"invalid token signing method",
				observability.Error(customerrors.ErrInvalidSigningMethod),
				observability.String("method", fmt.Sprintf("%v", t.Header["alg"])),
			)
			return nil, customerrors.ErrInvalidSigningMethod
		}
		return secret, nil
	})

	if err != nil {
		// Verifica se é erro de expiração usando errors.Is (v5 API)
		if errors.Is(err, jwt.ErrTokenExpired) {
			j.obs.Logger().Error(ctx, "token expired", observability.Error(customerrors.ErrTokenExpired))
			return nil, customerrors.ErrTokenExpired
		}
		j.obs.Logger().Error(ctx, "invalid token", observability.Error(customerrors.ErrInvalidToken))
		return nil, customerrors.ErrInvalidToken
	}

	// Valida se o token é válido
	if parsedToken == nil || !parsedToken.Valid {
		return nil, customerrors.ErrInvalidToken
	}

	// Extrai claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, customerrors.ErrInvalidTokenClaims
	}

	// Valida claim "sub" (user ID)
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		span.AddEvent("invalid sub claim", observability.Field{Key: "sub", Value: claims["sub"]})
		j.obs.Logger().Error(ctx, "invalid sub claim", observability.Error(customerrors.ErrInvalidTokenClaims))
		return nil, customerrors.ErrInvalidTokenClaims
	}

	// Valida claim "email"
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		span.AddEvent("invalid email claim", observability.Field{Key: "email", Value: claims["email"]})
		j.obs.Logger().Error(ctx, "invalid email claim", observability.Error(customerrors.ErrInvalidTokenClaims))
		return nil, customerrors.ErrInvalidTokenClaims
	}

	// Extrai roles (opcional, pode não existir em tokens antigos)
	var roles []string
	if rolesInterface, ok := claims["roles"]; ok {
		if rolesList, ok := rolesInterface.([]any); ok {
			for _, r := range rolesList {
				if roleStr, ok := r.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	user := NewAuthenticatedUser(sub, email, roles)
	return user, nil
}
