package invoice

import (
	"database/sql"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type InvoiceModule struct {
	InvoiceRouter        *http.InvoiceRouter
	InvoiceTotalProvider transactionInterfaces.InvoiceTotalProvider
}

func NewInvoiceModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	cardProvider interfaces.CardProvider,
	outboxService outbox.Service,
) (InvoiceModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	// Create Unit of Work for transaction management
	uow, err := uow.NewUnitOfWork(db)
	if err != nil {
		return InvoiceModule{}, fmt.Errorf("invoice module: failed to create unit of work: %v", err)
	}

	// Create repository (will be created inside UoW transactions)
	invoiceRepository := repositories.NewInvoiceRepository(db, o11y)

	// Create use cases
	createPurchaseUseCase := usecase.NewCreatePurchaseUseCase(uow, cardProvider, outboxService, o11y)
	updatePurchaseUseCase := usecase.NewUpdatePurchaseUseCase(uow, outboxService, o11y)
	deletePurchaseUseCase := usecase.NewDeletePurchaseUseCase(uow, outboxService, o11y)
	getInvoiceUseCase := usecase.NewGetInvoiceUseCase(invoiceRepository, o11y)
	listInvoicesByMonthUseCase := usecase.NewListInvoicesByMonthUseCase(invoiceRepository, o11y)
	listInvoicesByMonthPaginatedUseCase := usecase.NewListInvoicesByMonthPaginatedUseCase(invoiceRepository, o11y)
	listInvoicesByCardUseCase := usecase.NewListInvoicesByCardUseCase(invoiceRepository, o11y)
	listInvoicesByCardPaginatedUseCase := usecase.NewListInvoicesByCardPaginatedUseCase(invoiceRepository, o11y)

	// Create handler
	invoiceHandler := http.NewInvoiceHandler(
		o11y,
		errorHandler,
		createPurchaseUseCase,
		updatePurchaseUseCase,
		deletePurchaseUseCase,
		getInvoiceUseCase,
		listInvoicesByMonthUseCase,
		listInvoicesByMonthPaginatedUseCase,
		listInvoicesByCardUseCase,
		listInvoicesByCardPaginatedUseCase,
	)

	// Create router
	invoiceRouter := http.NewInvoiceRouter(invoiceHandler, authMiddleware)

	// Create InvoiceTotalProvider adapter for Transaction module integration
	invoiceTotalProvider := adapters.NewInvoiceTotalProviderAdapter(invoiceRepository)

	return InvoiceModule{
		InvoiceRouter:        invoiceRouter,
		InvoiceTotalProvider: invoiceTotalProvider,
	}, nil
}
