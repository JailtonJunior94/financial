package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type InvoiceRouter struct {
	handlers       *InvoiceHandler
	authMiddleware middlewares.Authorization
}

func NewInvoiceRouter(handlers *InvoiceHandler, authMiddleware middlewares.Authorization) *InvoiceRouter {
	return &InvoiceRouter{
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}
}

func (r InvoiceRouter) Register(router chi.Router) {
	// Aplica middleware de autenticação nas rotas de faturas e compras
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		// Change 8: Renamed /purchases → /invoice-items for semantic clarity
		protected.Post("/api/v1/invoice-items", r.handlers.CreatePurchase)
		protected.Put("/api/v1/invoice-items/{id}", r.handlers.UpdatePurchase)
		protected.Delete("/api/v1/invoice-items/{id}", r.handlers.DeletePurchase)

		// Invoice routes (read-only - invoices are calculated from purchases)
		// Change 6: Unified route with query params ?month= or ?cardId=
		protected.Get("/api/v1/invoices", r.handlers.ListInvoices)    // ?month=YYYY-MM or ?cardId=uuid
		protected.Get("/api/v1/invoices/{id}", r.handlers.GetInvoice) // Get specific invoice
	})
}
