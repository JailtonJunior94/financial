package http

import (
	"github.com/go-chi/chi/v5"
)

type BudgetRouter struct {
	handlers *BudgetHandler
}

func NewBudgetRouter(handlers *BudgetHandler) *BudgetRouter {
	return &BudgetRouter{handlers: handlers}
}

func (r BudgetRouter) Register(router chi.Router) {
	router.Post("/budgets", r.handlers.Create)
}
