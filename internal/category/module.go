package category

import (
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repository"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/rest"
	"github.com/jailtonjunior94/financial/internal/category/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
)

func RegisterCategoryModule(ioc *bundle.Container, router *chi.Mux) {
	categoryRepository := repository.NewCategoryRepository(ioc.DB)
	categoryCreateUseCase := usecase.NewCreateCategoryUseCase(ioc.Logger, categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryCreateUseCase)
	rest.NewCategoryRoutes(router, ioc.MiddlewareAuth, rest.WithCreateCategoryHandler(categoryHandler.Create))
}
