package category

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type CategoryModule struct {
	CategoryRouter *http.CategoryRouter
}

func NewCategoryModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) CategoryModule {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	categoryRepository := repositories.NewCategoryRepository(db, o11y)
	findCategoryUsecase := usecase.NewFindCategoryUseCase(o11y, categoryRepository)
	findCategoryByUsecase := usecase.NewFindCategoryByUseCase(o11y, categoryRepository)
	createCategoryUsecase := usecase.NewCreateCategoryUseCase(o11y, categoryRepository)
	updateCategoryUsecase := usecase.NewUpdateCategoryUseCase(o11y, categoryRepository)
	removeCategoryUsecase := usecase.NewRemoveCategoryUseCase(o11y, categoryRepository)

	categoryHandler := http.NewCategoryHandler(
		o11y,
		errorHandler,
		findCategoryUsecase,
		createCategoryUsecase,
		findCategoryByUsecase,
		updateCategoryUsecase,
		removeCategoryUsecase,
	)

	categoryRouter := http.NewCategoryRouter(categoryHandler, authMiddleware)

	return CategoryModule{
		CategoryRouter: categoryRouter,
	}
}
