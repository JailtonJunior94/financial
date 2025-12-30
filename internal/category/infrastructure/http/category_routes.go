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
	router.Get("/categories", r.handlers.Find)
	router.Get("/categories/{id}", r.handlers.FindBy)
	router.Post("/categories", r.handlers.Create)
	router.Put("/categories/{id}", r.handlers.Update)
	router.Delete("/categories/{id}", r.handlers.Delete)
}
