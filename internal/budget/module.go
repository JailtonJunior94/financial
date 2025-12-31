package budget

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	unitOfWork "github.com/jailtonjunior94/financial/pkg/database/uow"
)

type BudgetModule struct {
	BudgetRouter *http.BudgetRouter
}

func NewBudgetModule(db *sql.DB, o11y observability.Observability) BudgetModule {
	// Create error handler once for the module
	errorHandler := httperrors.NewErrorHandler(o11y)

	uow := unitOfWork.NewUnitOfWork(db)

	createBudgetUseCase := usecase.NewCreateBudgetUseCase(uow, o11y)
	budgetHandler := http.NewBudgetHandler(o11y, errorHandler, createBudgetUseCase)

	budgetRoutes := http.NewBudgetRouter(budgetHandler)

	return BudgetModule{
		BudgetRouter: budgetRoutes,
	}
}
