package budget

import (
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	unitOfWork "github.com/jailtonjunior94/financial/pkg/database/uow"

	"github.com/go-chi/chi/v5"
)

func RegisterBudgetModule(ioc *bundle.Container, router *chi.Mux) {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)

	budgetRepository := repositories.NewBudgetRepository(uow.Executor(), ioc.Observability)
	createBudgetUseCase := usecase.NewCreateBudgetUseCase(uow, ioc.Observability, budgetRepository)
	budgetHandler := http.NewBudgetHandler(ioc.Observability, createBudgetUseCase)

	http.NewBudgetRoutes(
		router,
		ioc.MiddlewareAuth,
		http.WithCreateBudgetHandler(budgetHandler.Create),
	)
}
