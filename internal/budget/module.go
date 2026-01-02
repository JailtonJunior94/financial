package budget

import (
	"database/sql"

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

func NewBudgetModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) BudgetModule {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	uow := uow.NewUnitOfWork(db)
	budgetRepository := repositories.NewBudgetRepository(db, o11y)

	createBudgetUseCase := usecase.NewCreateBudgetUseCase(uow, o11y)
	findBudgetUseCase := usecase.NewFindBudgetUseCase(budgetRepository, o11y)
	updateBudgetUseCase := usecase.NewUpdateBudgetUseCase(uow, o11y)
	deleteBudgetUseCase := usecase.NewDeleteBudgetUseCase(uow, o11y)

	budgetHandler := http.NewBudgetHandler(
		o11y,
		errorHandler,
		createBudgetUseCase,
		findBudgetUseCase,
		updateBudgetUseCase,
		deleteBudgetUseCase,
	)

	budgetRoutes := http.NewBudgetRouter(budgetHandler, authMiddleware)

	return BudgetModule{BudgetRouter: budgetRoutes}
}
