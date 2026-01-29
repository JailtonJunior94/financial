package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
)

type (
	GetInvoiceUseCase interface {
		Execute(ctx context.Context, invoiceID string) (*dtos.InvoiceOutput, error)
	}

	getInvoiceUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

func NewGetInvoiceUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) GetInvoiceUseCase {
	return &getInvoiceUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

func (u *getInvoiceUseCase) Execute(ctx context.Context, invoiceID string) (*dtos.InvoiceOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "get_invoice_usecase.execute")
	defer span.End()

	// Parse invoiceID
	id, err := vos.NewUUIDFromString(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := u.invoiceRepository.FindByID(ctx, id)
	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to find invoice", observability.Error(err))
		return nil, err
	}

	if invoice == nil {
		return nil, domain.ErrInvoiceNotFound
	}

	// Convert to DTO
	return u.toInvoiceOutput(invoice), nil
}

func (u *getInvoiceUseCase) toInvoiceOutput(invoice *entities.Invoice) *dtos.InvoiceOutput {
	items := make([]dtos.InvoiceItemOutput, len(invoice.Items))
	for i, item := range invoice.Items {
		items[i] = dtos.InvoiceItemOutput{
			ID:                item.ID.String(),
			InvoiceID:         item.InvoiceID.String(),
			CategoryID:        item.CategoryID.String(),
			PurchaseDate:      item.PurchaseDate.Format("2006-01-02"),
			Description:       item.Description,
			TotalAmount:       item.TotalAmount.String(),
			InstallmentNumber: item.InstallmentNumber,
			InstallmentTotal:  item.InstallmentTotal,
			InstallmentAmount: item.InstallmentAmount.String(),
			InstallmentLabel:  item.InstallmentLabel(),
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt.ValueOr(item.CreatedAt),
		}
	}

	return &dtos.InvoiceOutput{
		ID:             invoice.ID.String(),
		UserID:         invoice.UserID.String(),
		CardID:         invoice.CardID.String(),
		ReferenceMonth: invoice.ReferenceMonth.String(),
		DueDate:        invoice.DueDate.Format("2006-01-02"),
		TotalAmount:    invoice.TotalAmount.String(),
		Currency:       string(invoice.TotalAmount.Currency()),
		ItemCount:      len(invoice.Items),
		Items:          items,
		CreatedAt:      invoice.CreatedAt,
		UpdatedAt:      invoice.UpdatedAt.ValueOr(invoice.CreatedAt),
	}
}
