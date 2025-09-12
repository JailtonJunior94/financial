package category

import (
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

func RegisterCategoryModule(ioc *bundle.Container) []httpserver.Route {
	categoryRepository := repositories.NewCategoryRepository(ioc.DB, ioc.Observability)
	findCategoryUsecase := usecase.NewFindCategoryUseCase(ioc.Observability, categoryRepository)
	findCategoryByUsecase := usecase.NewFindCategoryByUseCase(ioc.Observability, categoryRepository)
	createCategoryUsecase := usecase.NewCreateCategoryUseCase(ioc.Observability, categoryRepository)
	updateCategoryUsecase := usecase.NewUpdateCategoryUseCase(ioc.Observability, categoryRepository)

	categoryHandler := http.NewCategoryHandler(
		ioc.Observability,
		createCategoryUsecase,
		findCategoryUsecase,
		findCategoryByUsecase,
		updateCategoryUsecase,
	)

	categoryRoutes := http.NewCategoryRoutes()
	categoryRoutes.Register(
		httpserver.NewRoute(
			"GET",
			"/api/v1/categories",
			categoryHandler.Find,
			ioc.MiddlewareAuth.Authorization,
		),
	)
	categoryRoutes.Register(
		httpserver.NewRoute(
			"GET",
			"/api/v1/categories/{id}",
			categoryHandler.FindBy,
			ioc.MiddlewareAuth.Authorization,
		),
	)
	categoryRoutes.Register(
		httpserver.NewRoute(
			"POST",
			"/api/v1/categories",
			categoryHandler.Create,
			ioc.MiddlewareAuth.Authorization,
		),
	)
	categoryRoutes.Register(
		httpserver.NewRoute(
			"PUT",
			"/api/v1/categories/{id}",
			categoryHandler.Update,
			ioc.MiddlewareAuth.Authorization,
		),
	)

	return categoryRoutes.Routes()
}
