package http

import (
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/go-chi/chi/v5"
)

type UserRouter struct {
	authHandler         *AuthHandler
	userHandler         *UserHandler
	authMiddleware      middlewares.Authorization
	ownershipMiddleware middlewares.ResourceOwnership
}

func NewUserRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	authMiddleware middlewares.Authorization,
	ownershipMiddleware middlewares.ResourceOwnership,
) *UserRouter {
	return &UserRouter{
		authHandler:         authHandler,
		userHandler:         userHandler,
		authMiddleware:      authMiddleware,
		ownershipMiddleware: ownershipMiddleware,
	}
}

func (r UserRouter) Register(router chi.Router) {
	router.Post("/api/v1/token", r.authHandler.Token)
	router.Post("/api/v1/users", r.userHandler.Create)
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)
		protected.Get("/api/v1/users", r.userHandler.List)
		protected.Group(func(owned chi.Router) {
			owned.Use(r.ownershipMiddleware.Ownership("id"))
			owned.Get("/api/v1/users/{id}", r.userHandler.GetByID)
			owned.Put("/api/v1/users/{id}", r.userHandler.Update)
			owned.Delete("/api/v1/users/{id}", r.userHandler.Delete)
		})
	})
}
