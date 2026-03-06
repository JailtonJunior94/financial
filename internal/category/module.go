package category

import (
"database/sql"
"fmt"

"github.com/jailtonjunior94/financial/internal/category/application/usecase"
"github.com/jailtonjunior94/financial/internal/category/infrastructure/http"
"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
"github.com/jailtonjunior94/financial/pkg/api/httperrors"
"github.com/jailtonjunior94/financial/pkg/api/middlewares"
"github.com/jailtonjunior94/financial/pkg/auth"
"github.com/jailtonjunior94/financial/pkg/observability/metrics"

"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type CategoryModule struct {
CategoryRouter *http.CategoryRouter
}

func NewCategoryModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) (CategoryModule, error) {
errorHandler := httperrors.NewErrorHandler(o11y)
fm := metrics.NewFinancialMetrics(o11y)

unitOfWork, err := uow.NewUnitOfWork(db)
if err != nil {
return CategoryModule{}, fmt.Errorf("category module: %w", err)
}

categoryRepo := repositories.NewCategoryRepository(db, o11y, fm)
subcategoryRepo := repositories.NewSubcategoryRepository(db, o11y, fm)

createCategory := usecase.NewCreateCategoryUseCase(o11y, fm, categoryRepo)
findCategoryPaginated := usecase.NewFindCategoryPaginatedUseCase(o11y, fm, categoryRepo)
findCategoryBy := usecase.NewFindCategoryByUseCase(o11y, fm, categoryRepo, subcategoryRepo)
updateCategory := usecase.NewUpdateCategoryUseCase(o11y, fm, categoryRepo)
removeCategory := usecase.NewRemoveCategoryUseCase(o11y, fm, unitOfWork, categoryRepo)

createSubcategory := usecase.NewCreateSubcategoryUseCase(o11y, fm, categoryRepo, subcategoryRepo)
findSubcategoryBy := usecase.NewFindSubcategoryByUseCase(o11y, fm, categoryRepo, subcategoryRepo)
findSubcategoriesPaginated := usecase.NewFindSubcategoriesPaginatedUseCase(o11y, fm, categoryRepo, subcategoryRepo)
updateSubcategory := usecase.NewUpdateSubcategoryUseCase(o11y, fm, categoryRepo, subcategoryRepo)
removeSubcategory := usecase.NewRemoveSubcategoryUseCase(o11y, fm, categoryRepo, subcategoryRepo)

authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

categoryHandler := http.NewCategoryHandler(http.CategoryHandlerDeps{
O11y:                         o11y,
FM:                           fm,
ErrorHandler:                 errorHandler,
CreateCategoryUseCase:        createCategory,
FindCategoryPaginatedUseCase: findCategoryPaginated,
FindCategoryByUseCase:        findCategoryBy,
UpdateCategoryUseCase:        updateCategory,
RemoveCategoryUseCase:        removeCategory,
})

subcategoryHandler := http.NewSubcategoryHandler(http.SubcategoryHandlerDeps{
O11y:                              o11y,
FM:                                fm,
ErrorHandler:                      errorHandler,
CreateSubcategoryUseCase:          createSubcategory,
FindSubcategoryByUseCase:          findSubcategoryBy,
FindSubcategoriesPaginatedUseCase: findSubcategoriesPaginated,
UpdateSubcategoryUseCase:          updateSubcategory,
RemoveSubcategoryUseCase:          removeSubcategory,
})

router := http.NewCategoryRouter(categoryHandler, subcategoryHandler, authMiddleware)
return CategoryModule{CategoryRouter: router}, nil
}
