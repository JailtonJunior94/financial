package transaction

import (
	"database/sql"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	appstrategies "github.com/jailtonjunior94/financial/internal/transaction/application/strategies"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	transactiondomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/messaging"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type TransactionModule struct {
	TransactionRouter     *http.TransactionRouter
	SyncMonthlyUseCase    usecase.SyncMonthlyFromInvoicesUseCase
	PurchaseEventConsumer *messaging.PurchaseEventConsumer
}

func NewTransactionModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	invoiceTotalProvider interfaces.InvoiceTotalProvider,
) (TransactionModule, error) {
	if invoiceTotalProvider == nil {
		invoiceTotalProvider = adapters.NewNoOpInvoiceTotalProvider()
	}

	errorHandler := httperrors.NewErrorHandler(o11y, transactiondomain.ErrorMappings())
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	unitOfWork, err := uow.NewUnitOfWork(db)
	if err != nil {
		return TransactionModule{}, fmt.Errorf("transaction module: failed to create unit of work: %v", err)
	}

	financialMetrics := metrics.NewFinancialMetrics(o11y)

	transactionRepository := repositories.NewTransactionRepository(db, o11y, financialMetrics)

	ccItemPersister := appstrategies.NewCreditCardItemPersister(transactionRepository)

	registerTransactionUseCase := usecase.NewRegisterTransactionUseCase(unitOfWork, transactionRepository, invoiceTotalProvider, ccItemPersister, o11y, financialMetrics)
	updateTransactionItemUseCase := usecase.NewUpdateTransactionItemUseCase(unitOfWork, transactionRepository, o11y, financialMetrics)
	deleteTransactionItemUseCase := usecase.NewDeleteTransactionItemUseCase(unitOfWork, transactionRepository, o11y, financialMetrics)
	listMonthlyPaginatedUseCase := usecase.NewListMonthlyPaginatedUseCase(o11y, transactionRepository, financialMetrics)
	getMonthlyUseCase := usecase.NewGetMonthlyUseCase(o11y, transactionRepository, financialMetrics)
	syncMonthlyFromInvoicesUseCase := usecase.NewSyncMonthlyFromInvoicesUseCase(unitOfWork, transactionRepository, invoiceTotalProvider, ccItemPersister, o11y, financialMetrics)

	transactionHandler := http.NewTransactionHandler(
		o11y,
		errorHandler,
		registerTransactionUseCase,
		updateTransactionItemUseCase,
		deleteTransactionItemUseCase,
		listMonthlyPaginatedUseCase,
		getMonthlyUseCase,
	)

	transactionRouter := http.NewTransactionRouter(transactionHandler, authMiddleware)

	purchaseEventConsumer := messaging.NewPurchaseEventConsumer(syncMonthlyFromInvoicesUseCase, db, o11y, financialMetrics)

	return TransactionModule{
		TransactionRouter:     transactionRouter,
		SyncMonthlyUseCase:    syncMonthlyFromInvoicesUseCase,
		PurchaseEventConsumer: purchaseEventConsumer,
	}, nil
}
