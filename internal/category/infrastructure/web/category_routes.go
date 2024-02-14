package web

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/infrastructure/web/middlewares"

	"github.com/go-chi/chi/v5"
)

type (
	Routes        func(categoryRoute *categoryRoute)
	categoryRoute struct {
		Authorization         middlewares.Authorization
		CreateCategoryHandler func(w http.ResponseWriter, r *http.Request)
	}
)

func NewCategoryRoutes(router *chi.Mux, middleware middlewares.Authorization, categoryRoutes ...Routes) *categoryRoute {
	route := &categoryRoute{}
	for _, categoryRoute := range categoryRoutes {
		categoryRoute(route)
	}
	route.Register(middleware, router)
	return route
}

func (u *categoryRoute) Register(middleware middlewares.Authorization, router *chi.Mux) {
	router.Route("/api/v1/categories", func(r chi.Router) {
		r.Use(middleware.Authorization)
		r.Post("/", u.CreateCategoryHandler)
	})
}

func WithCreateCategoryHandler(handler func(w http.ResponseWriter, r *http.Request)) Routes {
	return func(categoryRouter *categoryRoute) {
		categoryRouter.CreateCategoryHandler = handler
	}
}
