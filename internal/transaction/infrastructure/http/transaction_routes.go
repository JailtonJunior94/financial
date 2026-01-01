package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type TransactionRouter struct {
	handlers       *TransactionHandler
	authMiddleware middlewares.Authorization
}

func NewTransactionRouter(handlers *TransactionHandler, authMiddleware middlewares.Authorization) *TransactionRouter {
	return &TransactionRouter{
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}
}

func (r TransactionRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		protected.Post("/api/v1/transactions", r.handlers.Register)
		protected.Put("/api/v1/transactions/items/{id}", r.handlers.UpdateItem)
		protected.Delete("/api/v1/transactions/items/{id}", r.handlers.DeleteItem)
	})
}
