package http

import (
	"github.com/go-chi/chi/v5"
)

type CategoryRouter struct {
	handlers *CategoryHandler
}

func NewCategoryRouter(handlers *CategoryHandler) *CategoryRouter {
	return &CategoryRouter{handlers: handlers}
}

func (r CategoryRouter) Register(router chi.Router) {
	router.Get("/api/v1/categories", r.handlers.Find)
	router.Get("/api/v1/categories/{id}", r.handlers.FindBy)
	router.Post("/api/v1/categories", r.handlers.Create)
	router.Put("/api/v1/categories/{id}", r.handlers.Update)
	router.Delete("/api/v1/categories/{id}", r.handlers.Delete)
}
