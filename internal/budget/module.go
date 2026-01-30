package budget

import (
	"database/sql"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type BudgetModule struct {
	BudgetRouter *http.BudgetRouter
}

func NewBudgetModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) (BudgetModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	uow, err := uow.NewUnitOfWork(db)
	if err != nil {
		return BudgetModule{}, fmt.Errorf("budget module: failed to create unit of work: %v", err)
	}

	budgetRepository := repositories.NewBudgetRepository(db, o11y)
	createBudgetUseCase := usecase.NewCreateBudgetUseCase(uow, o11y)
	updateBudgetUseCase := usecase.NewUpdateBudgetUseCase(uow, o11y)
	deleteBudgetUseCase := usecase.NewDeleteBudgetUseCase(uow, o11y)
	findBudgetUseCase := usecase.NewFindBudgetUseCase(budgetRepository, o11y)
	listBudgetsPaginatedUseCase := usecase.NewListBudgetsPaginatedUseCase(o11y, budgetRepository)

	budgetHandler := http.NewBudgetHandler(
		o11y,
		errorHandler,
		createBudgetUseCase,
		findBudgetUseCase,
		updateBudgetUseCase,
		deleteBudgetUseCase,
		listBudgetsPaginatedUseCase,
	)

	budgetRoutes := http.NewBudgetRouter(budgetHandler, authMiddleware)

	return BudgetModule{BudgetRouter: budgetRoutes}, nil
}
