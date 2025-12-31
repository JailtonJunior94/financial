package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type CategoryRouter struct {
	handlers       *CategoryHandler
	authMiddleware middlewares.Authorization
}

func NewCategoryRouter(handlers *CategoryHandler, authMiddleware middlewares.Authorization) *CategoryRouter {
	return &CategoryRouter{
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}
}

func (r CategoryRouter) Register(router chi.Router) {
	// Aplica middleware de autenticação APENAS nas rotas de categorias
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		protected.Get("/api/v1/categories", r.handlers.Find)
		protected.Get("/api/v1/categories/{id}", r.handlers.FindBy)
		protected.Post("/api/v1/categories", r.handlers.Create)
		protected.Put("/api/v1/categories/{id}", r.handlers.Update)
		protected.Delete("/api/v1/categories/{id}", r.handlers.Delete)
	})
}
