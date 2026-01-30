package http

import (
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/go-chi/chi/v5"
)

type BudgetRouter struct {
	handlers       *BudgetHandler
	authMiddleware middlewares.Authorization
}

func NewBudgetRouter(handlers *BudgetHandler, authMiddleware middlewares.Authorization) *BudgetRouter {
	return &BudgetRouter{handlers: handlers, authMiddleware: authMiddleware}
}

func (r BudgetRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		protected.Get("/api/v1/budgets", r.handlers.List)
		protected.Post("/api/v1/budgets", r.handlers.Create)
		protected.Get("/api/v1/budgets/{id}", r.handlers.Find)
		protected.Put("/api/v1/budgets/{id}", r.handlers.Update)
		protected.Delete("/api/v1/budgets/{id}", r.handlers.Delete)
	})
}
