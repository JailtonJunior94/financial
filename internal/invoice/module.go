package invoice

import (
	"database/sql"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	invoicedomain "github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type InvoiceModule struct {
	InvoiceRouter                *http.InvoiceRouter
	InvoiceTotalProvider         pkginterfaces.InvoiceTotalProvider
	InvoiceCategoryTotalProvider pkginterfaces.InvoiceCategoryTotalProvider
}

func NewInvoiceModule(
	db *sql.DB,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
	cardProvider interfaces.CardProvider,
	outboxService outbox.Service,
) (InvoiceModule, error) {
	errorHandler := httperrors.NewErrorHandler(o11y, invoicedomain.ErrorMappings())
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	// Create Unit of Work for transaction management
	uow, err := uow.NewUnitOfWork(db)
	if err != nil {
		return InvoiceModule{}, fmt.Errorf("invoice module: failed to create unit of work: %v", err)
	}

	financialMetrics := metrics.NewFinancialMetrics(o11y)

	// Create repository (will be created inside UoW transactions)
	invoiceRepository := repositories.NewInvoiceRepository(db, o11y, financialMetrics)

	// Create use cases
	createPurchaseUseCase := usecase.NewCreatePurchaseUseCase(uow, cardProvider, outboxService, o11y, financialMetrics)
	updatePurchaseUseCase := usecase.NewUpdatePurchaseUseCase(uow, outboxService, o11y, financialMetrics)
	deletePurchaseUseCase := usecase.NewDeletePurchaseUseCase(uow, outboxService, o11y, financialMetrics)
	getInvoiceUseCase := usecase.NewGetInvoiceUseCase(invoiceRepository, o11y)
	listInvoicesByMonthPaginatedUseCase := usecase.NewListInvoicesByMonthPaginatedUseCase(invoiceRepository, o11y)
	listInvoicesByCardPaginatedUseCase := usecase.NewListInvoicesByCardPaginatedUseCase(invoiceRepository, o11y)

	// Create handler
	invoiceHandler := http.NewInvoiceHandler(
		o11y,
		errorHandler,
		createPurchaseUseCase,
		updatePurchaseUseCase,
		deletePurchaseUseCase,
		getInvoiceUseCase,
		listInvoicesByMonthPaginatedUseCase,
		listInvoicesByCardPaginatedUseCase,
	)

	// Create router
	invoiceRouter := http.NewInvoiceRouter(invoiceHandler, authMiddleware)

	// Create provider adapters for cross-module integration
	invoiceTotalProvider := adapters.NewInvoiceTotalProviderAdapter(invoiceRepository)
	invoiceCategoryTotalProvider := adapters.NewInvoiceCategoryTotalAdapter(invoiceRepository)

	return InvoiceModule{
		InvoiceRouter:                invoiceRouter,
		InvoiceTotalProvider:         invoiceTotalProvider,
		InvoiceCategoryTotalProvider: invoiceCategoryTotalProvider,
	}, nil
}
