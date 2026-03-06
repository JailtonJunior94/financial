package adapters

import (
	"context"
	"time"

	cardInterfaces "github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type invoiceCheckerAdapter struct {
	invoiceRepo invoiceInterfaces.InvoiceRepository
	o11y        observability.Observability
}

func NewInvoiceCheckerAdapter(
	invoiceRepo invoiceInterfaces.InvoiceRepository,
	o11y observability.Observability,
) cardInterfaces.InvoiceChecker {
	return &invoiceCheckerAdapter{
		invoiceRepo: invoiceRepo,
		o11y:        o11y,
	}
}

func (a *invoiceCheckerAdapter) HasOpenInvoices(ctx context.Context, cardID vos.UUID) (bool, error) {
	ctx, span := a.o11y.Tracer().Start(ctx, "invoice_checker_adapter.has_open_invoices")
	defer span.End()

	invoices, err := a.invoiceRepo.FindByCard(ctx, cardID)
	if err != nil {
		span.RecordError(err)
		a.o11y.Logger().Error(ctx, "invoice_check_failed",
			observability.String("operation", "HasOpenInvoices"),
			observability.String("layer", "adapter"),
			observability.String("entity", "invoice"),
			observability.String("card_id", cardID.String()),
			observability.Error(err),
		)
		return false, err
	}

	now := time.Now()
	for _, inv := range invoices {
		if !inv.DueDate.Before(now) {
			return true, nil
		}
	}

	return false, nil
}
