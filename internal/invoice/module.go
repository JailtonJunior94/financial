package invoice

import (
	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

// InvoiceModule wires the invoice bounded context.
type InvoiceModule struct {
	InvoiceRouter                *http.InvoiceRouter
	InvoiceTotalProvider         pkginterfaces.InvoiceTotalProvider
	InvoiceCategoryTotalProvider pkginterfaces.InvoiceCategoryTotalProvider
	InvoiceProviderAdapter       *adapters.InvoiceProviderAdapter
}

// NewInvoiceModule creates and wires all dependencies for the invoice module.
func NewInvoiceModule(
	db database.DBTX,
	o11y observability.Observability,
	tokenValidator auth.TokenValidator,
) InvoiceModule {
	errorHandler := httperrors.NewErrorHandler(o11y, ErrorMappings())
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	financialMetrics := metrics.NewFinancialMetrics(o11y)
	invoiceRepository := repositories.NewInvoiceRepository(db, o11y, financialMetrics)

	getInvoiceUseCase := usecase.NewGetInvoiceUseCase(invoiceRepository, o11y)
	listInvoicesByCardPaginatedUseCase := usecase.NewListInvoicesByCardPaginatedUseCase(invoiceRepository, o11y)

	invoiceHandler := http.NewInvoiceHandler(
		o11y,
		errorHandler,
		listInvoicesByCardPaginatedUseCase,
		getInvoiceUseCase,
	)

	invoiceRouter := http.NewInvoiceRouter(invoiceHandler, authMiddleware)

	invoiceTotalProvider := adapters.NewInvoiceTotalProviderAdapter(invoiceRepository)
	invoiceCategoryTotalProvider := adapters.NewInvoiceCategoryTotalAdapter(invoiceRepository)
	invoiceProviderAdapter := adapters.NewInvoiceProviderAdapter(invoiceRepository, o11y)

	return InvoiceModule{
		InvoiceRouter:                invoiceRouter,
		InvoiceTotalProvider:         invoiceTotalProvider,
		InvoiceCategoryTotalProvider: invoiceCategoryTotalProvider,
		InvoiceProviderAdapter:       invoiceProviderAdapter,
	}
}
