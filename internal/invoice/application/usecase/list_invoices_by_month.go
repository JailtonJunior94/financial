package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
)

type (
	ListInvoicesByMonthUseCase interface {
		Execute(ctx context.Context, userID string, referenceMonth string) ([]*dtos.InvoiceListOutput, error)
	}

	listInvoicesByMonthUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

func NewListInvoicesByMonthUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) ListInvoicesByMonthUseCase {
	return &listInvoicesByMonthUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

func (u *listInvoicesByMonthUseCase) Execute(ctx context.Context, userID string, referenceMonth string) ([]*dtos.InvoiceListOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_invoices_by_month_usecase.execute")
	defer span.End()

	// Parse userID
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse referenceMonth
	refMonth, err := invoiceVos.NewReferenceMonth(referenceMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid reference month: %w", err)
	}

	// Find invoices
	invoices, err := u.invoiceRepository.FindByUserAndMonth(ctx, user, refMonth)
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

func (u *listInvoicesByMonthUseCase) toInvoiceListOutput(invoice *entities.Invoice) *dtos.InvoiceListOutput {
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
