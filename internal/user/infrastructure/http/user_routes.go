package http

import (
	"github.com/go-chi/chi/v5"
)

type UserRouter struct {
	authHandler *AuthHandler
	userHandler *UserHandler
}

func NewUserRouter(authHandler *AuthHandler, userHandler *UserHandler) *UserRouter {
	return &UserRouter{
		authHandler: authHandler,
		userHandler: userHandler,
	}
}

func (r UserRouter) Register(router chi.Router) {
	router.Post("/api/v1/token", r.authHandler.Token)
	router.Post("/api/v1/users", r.userHandler.Create)
}
