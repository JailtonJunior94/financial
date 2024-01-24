package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/jailtonjunior94/financial/configs"

	"github.com/dgrijalva/jwt-go"
)

type Authorization interface {
	Authorization(next http.Handler) http.Handler
}

type authorization struct {
	config *configs.Config
}

func NewAuthorization(config *configs.Config) Authorization {
	return &authorization{config: config}
}

func (a *authorization) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := a.tokenFromHeader(r)
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		jwtKey := []byte(a.config.AuthSecretKey)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user := NewUser(claims["sub"].(string), claims["email"].(string))
		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *authorization) tokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

var UserCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewUser(id, email string) *User {
	return &User{ID: id, Email: email}
}
