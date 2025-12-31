package http

import (
	"github.com/go-chi/chi/v5"
)

type PaymentMethodRouter struct {
	handlers *PaymentMethodHandler
}

func NewPaymentMethodRouter(handlers *PaymentMethodHandler) *PaymentMethodRouter {
	return &PaymentMethodRouter{
		handlers: handlers,
	}
}

func (r PaymentMethodRouter) Register(router chi.Router) {
	router.Get("/api/v1/payment-methods", r.handlers.Find)
	router.Get("/api/v1/payment-methods/{id}", r.handlers.FindBy)
	router.Get("/api/v1/payment-methods/code/{code}", r.handlers.FindByCode)
	router.Post("/api/v1/payment-methods", r.handlers.Create)
	router.Put("/api/v1/payment-methods/{id}", r.handlers.Update)
	router.Delete("/api/v1/payment-methods/{id}", r.handlers.Delete)
}
