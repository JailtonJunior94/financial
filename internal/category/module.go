package category

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
)

type CategoryModule struct {
	CategoryRouter *http.CategoryRouter
}

func NewCategoryModule(db *sql.DB, o11y observability.Observability) CategoryModule {
	categoryRepository := repositories.NewCategoryRepository(db, o11y)
	findCategoryUsecase := usecase.NewFindCategoryUseCase(o11y, categoryRepository)
	findCategoryByUsecase := usecase.NewFindCategoryByUseCase(o11y, categoryRepository)
	createCategoryUsecase := usecase.NewCreateCategoryUseCase(o11y, categoryRepository)
	updateCategoryUsecase := usecase.NewUpdateCategoryUseCase(o11y, categoryRepository)
	removeCategoryUsecase := usecase.NewRemoveCategoryUseCase(o11y, categoryRepository)

	categoryHandler := http.NewCategoryHandler(
		o11y,
		findCategoryUsecase,
		createCategoryUsecase,
		findCategoryByUsecase,
		updateCategoryUsecase,
		removeCategoryUsecase,
	)

	categoryRouter := http.NewCategoryRouter(categoryHandler)

	return CategoryModule{
		CategoryRouter: categoryRouter,
	}
}
