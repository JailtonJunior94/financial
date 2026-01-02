package http

import (
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/go-chi/chi/v5"
)

type CardRouter struct {
	handlers       *CardHandler
	authMiddleware middlewares.Authorization
}

func NewCardRouter(handlers *CardHandler, authMiddleware middlewares.Authorization) *CardRouter {
	return &CardRouter{
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}
}

func (r CardRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		protected.Get("/api/v1/cards", r.handlers.Find)
		protected.Get("/api/v1/cards/{id}", r.handlers.FindBy)
		protected.Post("/api/v1/cards", r.handlers.Create)
		protected.Put("/api/v1/cards/{id}", r.handlers.Update)
		protected.Delete("/api/v1/cards/{id}", r.handlers.Delete)
	})
}
