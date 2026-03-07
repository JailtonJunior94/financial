package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

// InvoiceRouter registers invoice HTTP routes.
type InvoiceRouter struct {
	handlers       *InvoiceHandler
	authMiddleware middlewares.Authorization
}

// NewInvoiceRouter creates a new InvoiceRouter.
func NewInvoiceRouter(handlers *InvoiceHandler, authMiddleware middlewares.Authorization) *InvoiceRouter {
	return &InvoiceRouter{handlers: handlers, authMiddleware: authMiddleware}
}

// Register registers routes on the provided chi.Router.
func (r InvoiceRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)
		protected.Get("/api/v1/cards/{cardId}/invoices", r.handlers.ListByCard)
		protected.Get("/api/v1/cards/{cardId}/invoices/{invoiceId}", r.handlers.GetByCard)
	})
}
