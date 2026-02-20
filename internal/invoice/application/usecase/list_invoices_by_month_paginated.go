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
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type (
	// ListInvoicesByMonthPaginatedUseCase lista faturas de um usuário em um mês com paginação cursor-based.
	ListInvoicesByMonthPaginatedUseCase interface {
		Execute(ctx context.Context, input ListInvoicesByMonthPaginatedInput) (*ListInvoicesByMonthPaginatedOutput, error)
	}

	// ListInvoicesByMonthPaginatedInput representa a entrada do use case.
	ListInvoicesByMonthPaginatedInput struct {
		UserID         string
		ReferenceMonth string
		Limit          int
		Cursor         string
	}

	// ListInvoicesByMonthPaginatedOutput representa a saída do use case.
	ListInvoicesByMonthPaginatedOutput struct {
		Invoices   []*dtos.InvoiceListOutput
		NextCursor *string
	}

	listInvoicesByMonthPaginatedUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

// NewListInvoicesByMonthPaginatedUseCase cria uma nova instância do use case.
func NewListInvoicesByMonthPaginatedUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) ListInvoicesByMonthPaginatedUseCase {
	return &listInvoicesByMonthPaginatedUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

// Execute executa o use case de listagem paginada de faturas por mês.
func (u *listInvoicesByMonthPaginatedUseCase) Execute(
	ctx context.Context,
	input ListInvoicesByMonthPaginatedInput,
) (*ListInvoicesByMonthPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_invoices_by_month_paginated_usecase.execute")
	defer span.End()

	// Parse userID
	user, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse referenceMonth
	refMonth, err := invoiceVos.NewReferenceMonth(input.ReferenceMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid reference month: %w", err)
	}

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}

	// List invoices (paginado)
	invoices, err := u.invoiceRepository.ListByUserAndMonthPaginated(ctx, interfaces.ListInvoicesByMonthParams{
		UserID:         user,
		ReferenceMonth: refMonth,
		Limit:          input.Limit + 1, // +1 para detectar has_next
		Cursor:         cursor,
	})
	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to list invoices", observability.Error(err))
		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(invoices) > input.Limit
	if hasNext {
		invoices = invoices[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(invoices) > 0 {
		lastInvoice := invoices[len(invoices)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"due_date": lastInvoice.DueDate.Format("2006-01-02"),
				"id":       lastInvoice.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}

		nextCursor = &encoded
	}

	// Convert to DTOs
	result := make([]*dtos.InvoiceListOutput, len(invoices))
	for i, invoice := range invoices {
		result[i] = u.toInvoiceListOutput(invoice)
	}

	return &ListInvoicesByMonthPaginatedOutput{
		Invoices:   result,
		NextCursor: nextCursor,
	}, nil
}

func (u *listInvoicesByMonthPaginatedUseCase) toInvoiceListOutput(invoice *entities.Invoice) *dtos.InvoiceListOutput {
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
