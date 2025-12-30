package budget

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	unitOfWork "github.com/jailtonjunior94/financial/pkg/database/uow"
)

type BudgetModule struct {
	BudgetRouter *http.BudgetRouter
}

func NewBudgetModule(db *sql.DB, o11y observability.Observability) BudgetModule {
	uow := unitOfWork.NewUnitOfWork(db)

	createBudgetUseCase := usecase.NewCreateBudgetUseCase(uow, o11y)
	budgetHandler := http.NewBudgetHandler(o11y, createBudgetUseCase)

	budgetRoutes := http.NewBudgetRouter(budgetHandler)

	return BudgetModule{
		BudgetRouter: budgetRoutes,
	}
}
