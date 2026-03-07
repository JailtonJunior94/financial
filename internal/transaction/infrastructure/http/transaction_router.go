package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

// TransactionRouter registers transaction HTTP routes.
type TransactionRouter struct {
	handlers       *TransactionHandler
	authMiddleware middlewares.Authorization
}

// NewTransactionRouter creates a new TransactionRouter.
func NewTransactionRouter(handlers *TransactionHandler, authMiddleware middlewares.Authorization) *TransactionRouter {
	return &TransactionRouter{handlers: handlers, authMiddleware: authMiddleware}
}

// Register registers routes on the provided chi.Router.
func (r TransactionRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)
		protected.Post("/api/v1/transactions", r.handlers.Create)
		protected.Get("/api/v1/transactions", r.handlers.List)
		protected.Get("/api/v1/transactions/{id}", r.handlers.Get)
		protected.Put("/api/v1/transactions/{id}", r.handlers.Update)
		protected.Post("/api/v1/transactions/{id}/reverse", r.handlers.Reverse)
	})
}
