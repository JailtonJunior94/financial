package transaction

import (
	"database/sql"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/transaction/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type TransactionModule struct {
	TransactionRouter *http.TransactionRouter
}

func NewTransactionModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	invoiceTotalProvider interfaces.InvoiceTotalProvider,
) (TransactionModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)
	unitOfWork, err := uow.NewUnitOfWork(db)
	if err != nil {
		return TransactionModule{}, fmt.Errorf("transaction module: failed to create unit of work: %v", err)
	}

	// Repository
	transactionRepository := repositories.NewTransactionRepository(db, o11y)

	// Use Cases
	registerTransactionUseCase := usecase.NewRegisterTransactionUseCase(unitOfWork, transactionRepository, invoiceTotalProvider, o11y)
	updateTransactionItemUseCase := usecase.NewUpdateTransactionItemUseCase(unitOfWork, transactionRepository, o11y)
	deleteTransactionItemUseCase := usecase.NewDeleteTransactionItemUseCase(unitOfWork, transactionRepository, o11y)
	listMonthlyPaginatedUseCase := usecase.NewListMonthlyPaginatedUseCase(o11y, transactionRepository)
	getMonthlyUseCase := usecase.NewGetMonthlyUseCase(o11y, transactionRepository)

	// Handler
	transactionHandler := http.NewTransactionHandler(
		o11y,
		errorHandler,
		registerTransactionUseCase,
		updateTransactionItemUseCase,
		deleteTransactionItemUseCase,
		listMonthlyPaginatedUseCase,
		getMonthlyUseCase,
	)

	// Router
	transactionRouter := http.NewTransactionRouter(transactionHandler, authMiddleware)

	return TransactionModule{TransactionRouter: transactionRouter}, nil
}
