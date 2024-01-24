package authentication

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/configs"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type JwtAdapter interface {
	GenerateTokenJWT(id, email string) (string, error)
	ExtractClaims(tokenString string) (string, error)
}

type jwtAdapter struct {
	config *configs.Config
}

func NewJwtAdapter(config *configs.Config) JwtAdapter {
	return &jwtAdapter{config: config}
}

func (j *jwtAdapter) GenerateTokenJWT(id, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = id
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(j.config.AuthExpirationAt)).Unix()

	t, err := token.SignedString([]byte(j.config.AuthSecretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (j *jwtAdapter) ExtractClaims(tokenString string) (string, error) {
	tokenString = strings.Split(tokenString, " ")[1]
	hmacSecret := []byte(j.config.AuthSecretKey)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	sub := fmt.Sprintf("%v", claims["sub"])
	return sub, nil
}
