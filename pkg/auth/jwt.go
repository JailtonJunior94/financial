package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/configs"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrGenerateToken = errors.New("error trying to generate token")
)

type (
	JwtAdapter interface {
		GenerateToken(ctx context.Context, id, email string) (string, error)
		ValidateToken(ctx context.Context, tokenRequest string) (*User, error)
	}

	jwtAdapter struct {
		config *configs.Config
		obs    observability.Observability
	}

	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
)

func NewUser(id, email string) *User {
	return &User{ID: id, Email: email}
}

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

func (j *jwtAdapter) ValidateToken(ctx context.Context, tokenRequest string) (*User, error) {
	_, span := j.obs.Tracer().Start(ctx, "jwt_adapter.validate_token")
	defer span.End()

	tokenString := j.removeBearerPrefix(tokenRequest)
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	secret := []byte(j.config.AuthConfig.AuthSecretKey)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			span.AddEvent(
				"invalid token signing method",
				observability.Field{Key: "method", Value: fmt.Sprintf("%v", token.Header["alg"])},
			)
			j.obs.Logger().Error(ctx, "invalid token signing method", observability.Error(ErrInvalidToken), observability.String("method", fmt.Sprintf("%v", token.Header["alg"])))
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		span.AddEvent("invalid sub claim", observability.Field{Key: "sub", Value: claims["sub"]})
		j.obs.Logger().Error(ctx, "invalid sub claim", observability.Error(ErrInvalidToken))
		return nil, ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		span.AddEvent("invalid email claim", observability.Field{Key: "email", Value: claims["email"]})
		j.obs.Logger().Error(ctx, "invalid email claim", observability.Error(ErrInvalidToken))
		return nil, ErrInvalidToken
	}

	user := NewUser(sub, email)
	return user, nil
}

func (j *jwtAdapter) removeBearerPrefix(tokenString string) string {
	const bearerPrefix = "BEARER "
	if len(tokenString) > len(bearerPrefix) && strings.ToUpper(tokenString[:len(bearerPrefix)]) == bearerPrefix {
		return tokenString[len(bearerPrefix):]
	}
	return tokenString
}
