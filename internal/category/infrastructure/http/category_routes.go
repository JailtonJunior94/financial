package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type CategoryRouter struct {
	categoryHandler    *CategoryHandler
	subcategoryHandler *SubcategoryHandler
	authMiddleware     middlewares.Authorization
}

func NewCategoryRouter(
	categoryHandler *CategoryHandler,
	subcategoryHandler *SubcategoryHandler,
	authMiddleware middlewares.Authorization,
) *CategoryRouter {
	return &CategoryRouter{
		categoryHandler:    categoryHandler,
		subcategoryHandler: subcategoryHandler,
		authMiddleware:     authMiddleware,
	}
}

func (r CategoryRouter) Register(router chi.Router) {
	router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware.Authorization)

		protected.Get("/api/v1/categories", r.categoryHandler.Find)
		protected.Post("/api/v1/categories", r.categoryHandler.Create)
		protected.Get("/api/v1/categories/{id}", r.categoryHandler.FindBy)
		protected.Put("/api/v1/categories/{id}", r.categoryHandler.Update)
		protected.Delete("/api/v1/categories/{id}", r.categoryHandler.Delete)

		protected.Route("/api/v1/categories/{categoryId}/subcategories", func(sub chi.Router) {
			sub.Get("/", r.subcategoryHandler.List)
			sub.Post("/", r.subcategoryHandler.Create)
			sub.Get("/{id}", r.subcategoryHandler.FindBy)
			sub.Put("/{id}", r.subcategoryHandler.Update)
			sub.Delete("/{id}", r.subcategoryHandler.Delete)
		})
	})
}
