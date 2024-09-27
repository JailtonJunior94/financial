package category

import (
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/internal/category/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
)

func RegisterCategoryModule(ioc *bundle.Container, router *chi.Mux) {
	categoryRepository := repositories.NewCategoryRepository(ioc.DB, ioc.Observability)
	categoryCreateUseCase := usecase.NewCreateCategoryUseCase(ioc.Observability, categoryRepository)
	categoryHandler := http.NewCategoryHandler(ioc.Observability, categoryCreateUseCase)
	http.NewCategoryRoutes(
		router,
		ioc.MiddlewareAuth,
		http.WithCreateCategoryHandler(categoryHandler.Create),
	)
}
