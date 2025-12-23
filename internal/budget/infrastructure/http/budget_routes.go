package http

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/go-chi/chi/v5"
)

type (
	Routes      func(budgetRoute *budgetRoute)
	budgetRoute struct {
		Authorization       middlewares.Authorization
		CreateBudgetHandler func(w http.ResponseWriter, r *http.Request) error
	}
)

func NewBudgetRoutes(router *chi.Mux, middleware middlewares.Authorization, budgetRoutes ...Routes) *budgetRoute {
	route := &budgetRoute{}
	for _, budgetRoute := range budgetRoutes {
		budgetRoute(route)
	}
	route.Register(middleware, router)
	return route
}

func (u *budgetRoute) Register(middleware middlewares.Authorization, router *chi.Mux) {
	router.Route("/api/v1/budgets", func(r chi.Router) {
		r.Use(middleware.Authorization)
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			if err := u.CreateBudgetHandler(w, r); err != nil {
				// Error is already handled by the error handler middleware
				return
			}
		})
	})
}

func WithCreateBudgetHandler(handler func(w http.ResponseWriter, r *http.Request) error) Routes {
	return func(categoryRouter *budgetRoute) {
		categoryRouter.CreateBudgetHandler = handler
	}
}
