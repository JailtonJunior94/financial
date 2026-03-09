package budget

import (
	"database/sql"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	budgetdomain "github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	budgethttp "github.com/jailtonjunior94/financial/internal/budget/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/messaging"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/outbox"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type BudgetModule struct {
	BudgetRouter        *budgethttp.BudgetRouter
	BudgetEventConsumer *messaging.BudgetEventConsumer
}

func NewBudgetModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	invoiceCategoryTotal interfaces.InvoiceCategoryTotalProvider,
) (BudgetModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y, budgetdomain.ErrorMappings())
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	unitOfWork, err := uow.NewUnitOfWork(db)
	if err != nil {
		return BudgetModule{}, fmt.Errorf("budget module: failed to create unit of work: %v", err)
	}

	financialMetrics := metrics.NewFinancialMetrics(o11y)

	budgetRepository := repositories.NewBudgetRepository(db, o11y, financialMetrics)
	categoryProvider := adapters.NewCategoryProviderAdapter(db, o11y, financialMetrics)
	replicateBudgetUseCase := usecase.NewReplicateBudgetUseCase(o11y)
	createBudgetUseCase := usecase.NewCreateBudgetUseCase(unitOfWork, o11y, financialMetrics, budgetRepository, categoryProvider, replicateBudgetUseCase)
	updateBudgetUseCase := usecase.NewUpdateBudgetUseCase(unitOfWork, o11y, financialMetrics, budgetRepository, categoryProvider, replicateBudgetUseCase)
	deleteBudgetUseCase := usecase.NewDeleteBudgetUseCase(unitOfWork, o11y, financialMetrics, budgetRepository)
	findBudgetUseCase := usecase.NewFindBudgetUseCase(budgetRepository, o11y, financialMetrics)
	listBudgetsPaginatedUseCase := usecase.NewListBudgetsPaginatedUseCase(o11y, budgetRepository)

	budgetHandler := budgethttp.NewBudgetHandler(
		o11y,
		errorHandler,
		createBudgetUseCase,
		findBudgetUseCase,
		updateBudgetUseCase,
		deleteBudgetUseCase,
		listBudgetsPaginatedUseCase,
	)

	budgetRoutes := budgethttp.NewBudgetRouter(budgetHandler, authMiddleware)

	var budgetEventConsumer *messaging.BudgetEventConsumer
	if invoiceCategoryTotal != nil {
		syncUseCase := usecase.NewSyncBudgetSpentAmountUseCase(unitOfWork, invoiceCategoryTotal, budgetRepository, o11y, financialMetrics)
		processedEventsRepo := outbox.NewProcessedEventsRepository(db)
		budgetEventConsumer = messaging.NewBudgetEventConsumer(syncUseCase, processedEventsRepo, o11y)
	}

	return BudgetModule{
		BudgetRouter:        budgetRoutes,
		BudgetEventConsumer: budgetEventConsumer,
	}, nil
}
