package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
)

type (
	ListInvoicesByCardUseCase interface {
		Execute(ctx context.Context, cardID string) ([]*dtos.InvoiceListOutput, error)
	}

	listInvoicesByCardUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

func NewListInvoicesByCardUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) ListInvoicesByCardUseCase {
	return &listInvoicesByCardUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

func (u *listInvoicesByCardUseCase) Execute(ctx context.Context, cardID string) ([]*dtos.InvoiceListOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_invoices_by_card_usecase.execute")
	defer span.End()

	// Parse cardID
	card, err := vos.NewUUIDFromString(cardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card ID: %w", err)
	}

	// Find invoices
	invoices, err := u.invoiceRepository.FindByCard(ctx, card)
	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to list invoices", observability.Error(err))
		return nil, err
	}

	// Convert to DTOs
	result := make([]*dtos.InvoiceListOutput, len(invoices))
	for i, invoice := range invoices {
		result[i] = u.toInvoiceListOutput(invoice)
	}

	return result, nil
}

func (u *listInvoicesByCardUseCase) toInvoiceListOutput(invoice *entities.Invoice) *dtos.InvoiceListOutput {
	return &dtos.InvoiceListOutput{
		ID:             invoice.ID.String(),
		CardID:         invoice.CardID.String(),
		ReferenceMonth: invoice.ReferenceMonth.String(),
		DueDate:        invoice.DueDate.Format("2006-01-02"),
		TotalAmount:    fmt.Sprintf("%.2f", invoice.TotalAmount.Float()),
		Currency:       string(invoice.TotalAmount.Currency()),
		ItemCount:      len(invoice.Items),
		CreatedAt:      invoice.CreatedAt,
	}
}
