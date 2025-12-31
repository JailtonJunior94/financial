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

		// Purchase routes (create, update, delete purchases)
		protected.Post("/api/v1/purchases", r.handlers.CreatePurchase)
		protected.Put("/api/v1/purchases/{id}", r.handlers.UpdatePurchase)
		protected.Delete("/api/v1/purchases/{id}", r.handlers.DeletePurchase)

		// Invoice routes (read-only - invoices are calculated from purchases)
		protected.Get("/api/v1/invoices", r.handlers.ListInvoicesByMonth)         // ?month=YYYY-MM
		protected.Get("/api/v1/invoices/{id}", r.handlers.GetInvoice)             // Get specific invoice
		protected.Get("/api/v1/invoices/cards/{cardId}", r.handlers.ListInvoicesByCard) // List by card
	})
}
