package authentication

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/logger"

	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrGenerateToken = errors.New("error trying to generate token")
)

type (
	JwtAdapter interface {
		GenerateToken(id, email string) (string, error)
		ValidateToken(tokenRequest string) (*User, error)
	}

	jwtAdapter struct {
		logger logger.Logger
		config *configs.Config
	}

	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
)

func NewUser(id, email string) *User {
	return &User{ID: id, Email: email}
}

func NewJwtAdapter(logger logger.Logger, config *configs.Config) JwtAdapter {
	return &jwtAdapter{logger: logger, config: config}
}

func (j *jwtAdapter) GenerateToken(id, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   id,
		"email": email,
		"exp":   time.Now().Add(time.Hour * time.Duration(j.config.AuthExpirationAt)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenSigned, err := token.SignedString([]byte(j.config.AuthSecretKey))
	if err != nil {
		j.logger.Error("error trying to generate token",
			logger.Field{Key: "e-mail", Value: email},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return "", ErrGenerateToken
	}
	return tokenSigned, nil
}

func (j *jwtAdapter) ValidateToken(tokenRequest string) (*User, error) {
	tokenString := j.removeBearerPrefix(tokenRequest)
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	secret := []byte(j.config.AuthSecretKey)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			j.logger.Error(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
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

	user := NewUser(claims["sub"].(string), claims["email"].(string))
	return user, nil
}

func (j *jwtAdapter) removeBearerPrefix(tokenString string) string {
	if len(tokenString) > 7 && strings.ToUpper(tokenString[0:6]) == "BEARER" {
		return tokenString[7:]
	}
	return ""
}
