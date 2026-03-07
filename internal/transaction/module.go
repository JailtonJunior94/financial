package transaction

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionhttp "github.com/jailtonjunior94/financial/internal/transaction/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/messaging"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

// TransactionModule wires the transaction bounded context.
type TransactionModule struct {
	TransactionRouter     *transactionhttp.TransactionRouter
	PurchaseEventConsumer *messaging.PurchaseEventConsumer
}

// NewTransactionModule creates and wires all dependencies for the transaction module.
func NewTransactionModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	invoiceProvider transactionInterfaces.InvoiceProvider,
	cardProvider invoiceInterfaces.CardProvider,
	outboxService outbox.Service,
) (TransactionModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y, transactionDomain.ErrorMappings())
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	transactionMetrics := metrics.NewTransactionMetrics(o11y)
	transactionRepository := repositories.NewTransactionRepository(db, o11y, transactionMetrics)

	unitOfWork, err := uow.NewUnitOfWork(db)
	if err != nil {
		return TransactionModule{}, err
	}

	createUC := usecase.NewCreateTransactionUseCase(o11y, unitOfWork, transactionRepository, invoiceProvider, cardProvider, outboxService)
	updateUC := usecase.NewUpdateTransactionUseCase(o11y, unitOfWork, transactionRepository, invoiceProvider)
	reverseUC := usecase.NewReverseTransactionUseCase(o11y, unitOfWork, transactionRepository, invoiceProvider)
	listUC := usecase.NewListTransactionsUseCase(o11y, transactionRepository)
	getUC := usecase.NewGetTransactionUseCase(o11y, transactionRepository)

	transactionHandler := transactionhttp.NewTransactionHandler(o11y, errorHandler, createUC, updateUC, reverseUC, listUC, getUC)
	transactionRouter := transactionhttp.NewTransactionRouter(transactionHandler, authMiddleware)

	purchaseEventConsumer := messaging.NewPurchaseEventConsumer()

	return TransactionModule{
		TransactionRouter:     transactionRouter,
		PurchaseEventConsumer: purchaseEventConsumer,
	}, nil
}
