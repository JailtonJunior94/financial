package auth

import (
	"context"
	"errors"
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
	// Implementa as interfaces TokenGenerator e TokenValidator.
	JwtAdapter interface {
		TokenGenerator
		TokenValidator
	}

	jwtAdapter struct {
		config *configs.Config
		obs    observability.Observability
	}
)

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
		span.RecordError(err)
		j.obs.Logger().Error(ctx, "error trying to generate token", observability.Error(err))
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
			span.RecordError(customerrors.ErrInvalidSigningMethod)
			j.obs.Logger().Error(
				ctx,
				"invalid token signing method",
				observability.Error(customerrors.ErrInvalidSigningMethod),
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
		span.RecordError(customerrors.ErrInvalidTokenClaims)
		j.obs.Logger().Error(ctx, "invalid sub claim", observability.Error(customerrors.ErrInvalidTokenClaims))
		return nil, customerrors.ErrInvalidTokenClaims
	}

	// Valida claim "email"
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		span.RecordError(customerrors.ErrInvalidTokenClaims)
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
