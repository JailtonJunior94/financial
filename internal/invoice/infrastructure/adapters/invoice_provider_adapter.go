package adapters

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	invoiceEntities "github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceProviderAdapter implements transactionInterfaces.InvoiceProvider using the invoice repository.
type InvoiceProviderAdapter struct {
	repo invoiceInterfaces.InvoiceRepository
	o11y observability.Observability
}

// NewInvoiceProviderAdapter creates a new InvoiceProviderAdapter.
func NewInvoiceProviderAdapter(repo invoiceInterfaces.InvoiceRepository, o11y observability.Observability) *InvoiceProviderAdapter {
	return &InvoiceProviderAdapter{repo: repo, o11y: o11y}
}

// FindOrCreate atomically finds or creates an invoice for the given card and month.
func (a *InvoiceProviderAdapter) FindOrCreate(
	ctx context.Context,
	userID, cardID vos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
	dueDate time.Time,
) (*transactionInterfaces.InvoiceInfo, error) {
	ctx, span := a.o11y.Tracer().Start(ctx, "invoice_provider_adapter.find_or_create")
	defer span.End()

	zeroMoney, _ := vos.NewMoney(0, vos.CurrencyBRL)
	newInvoice := &invoiceEntities.Invoice{}
	newInvoice.UserID = userID
	newInvoice.CardID = cardID
	newInvoice.ReferenceMonth = referenceMonth
	newInvoice.DueDate = dueDate
	newInvoice.TotalAmount = zeroMoney
	newInvoice.Status = "open"
	newInvoice.CreatedAt = time.Now().UTC()

	id, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}
	newInvoice.SetID(id)

	invoice, err := a.repo.UpsertInvoice(ctx, newInvoice)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	status := invoice.Status
	if status == "" {
		status = "open"
	}

	return &transactionInterfaces.InvoiceInfo{
		ID:     invoice.ID,
		Status: status,
	}, nil
}

// GetStatus returns the current status of an invoice.
func (a *InvoiceProviderAdapter) GetStatus(ctx context.Context, invoiceID vos.UUID) (string, error) {
	ctx, span := a.o11y.Tracer().Start(ctx, "invoice_provider_adapter.get_status")
	defer span.End()

	status, err := a.repo.FindStatus(ctx, invoiceID)
	if err != nil {
		span.RecordError(err)
		return "", err
	}
	return status, nil
}
