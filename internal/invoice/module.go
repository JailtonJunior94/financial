package invoice

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	unitOfWork "github.com/jailtonjunior94/financial/pkg/database/uow"
)

type InvoiceModule struct {
	InvoiceRouter *http.InvoiceRouter
}

func NewInvoiceModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	cardProvider interfaces.CardProvider, // ✅ Injected adapter from cards module
) InvoiceModule {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	// Create Unit of Work for transaction management
	uow := unitOfWork.NewUnitOfWork(db)

	// Create repository (will be created inside UoW transactions)
	invoiceRepository := repositories.NewInvoiceRepository(db, o11y)

	// Create use cases
	createPurchaseUseCase := usecase.NewCreatePurchaseUseCase(uow, cardProvider, o11y)
	updatePurchaseUseCase := usecase.NewUpdatePurchaseUseCase(uow, o11y)
	deletePurchaseUseCase := usecase.NewDeletePurchaseUseCase(uow, o11y)
	getInvoiceUseCase := usecase.NewGetInvoiceUseCase(invoiceRepository, o11y)
	listInvoicesByMonthUseCase := usecase.NewListInvoicesByMonthUseCase(invoiceRepository, o11y)
	listInvoicesByCardUseCase := usecase.NewListInvoicesByCardUseCase(invoiceRepository, o11y)

	// Create handler
	invoiceHandler := http.NewInvoiceHandler(
		o11y,
		errorHandler,
		createPurchaseUseCase,
		updatePurchaseUseCase,
		deletePurchaseUseCase,
		getInvoiceUseCase,
		listInvoicesByMonthUseCase,
		listInvoicesByCardUseCase,
	)

	// Create router
	invoiceRouter := http.NewInvoiceRouter(invoiceHandler, authMiddleware)

	return InvoiceModule{
		InvoiceRouter: invoiceRouter,
	}
}
