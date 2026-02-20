package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type (
	// ListInvoicesByCardPaginatedUseCase lista faturas de um cartão com paginação cursor-based.
	ListInvoicesByCardPaginatedUseCase interface {
		Execute(ctx context.Context, input ListInvoicesByCardPaginatedInput) (*ListInvoicesByCardPaginatedOutput, error)
	}

	// ListInvoicesByCardPaginatedInput representa a entrada do use case.
	ListInvoicesByCardPaginatedInput struct {
		CardID string
		Limit  int
		Cursor string
	}

	// ListInvoicesByCardPaginatedOutput representa a saída do use case.
	ListInvoicesByCardPaginatedOutput struct {
		Invoices   []dtos.InvoiceOutput
		NextCursor *string
	}

	listInvoicesByCardPaginatedUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

// NewListInvoicesByCardPaginatedUseCase cria uma nova instância do use case.
func NewListInvoicesByCardPaginatedUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) ListInvoicesByCardPaginatedUseCase {
	return &listInvoicesByCardPaginatedUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

// Execute executa o use case de listagem paginada de faturas por cartão.
func (u *listInvoicesByCardPaginatedUseCase) Execute(
	ctx context.Context,
	input ListInvoicesByCardPaginatedInput,
) (*ListInvoicesByCardPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_invoices_by_card_paginated_usecase.execute")
	defer span.End()

	// Parse card ID
	cardID, err := vos.NewUUIDFromString(input.CardID)
	if err != nil {
		return nil, err
	}

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}

	// List invoices (paginado)
	invoices, err := u.invoiceRepository.ListByCard(ctx, interfaces.ListInvoicesByCardParams{
		CardID: cardID,
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
	})
	if err != nil {
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
				"reference_month": lastInvoice.ReferenceMonth.String(),
				"id":              lastInvoice.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	output := make([]dtos.InvoiceOutput, len(invoices))
	for i, invoice := range invoices {
		items := make([]dtos.InvoiceItemOutput, len(invoice.Items))
		for j, item := range invoice.Items {
			items[j] = dtos.InvoiceItemOutput{
				ID:                item.ID.String(),
				InvoiceID:         item.InvoiceID.String(),
				CategoryID:        item.CategoryID.String(),
				PurchaseDate:      item.PurchaseDate.Format("2006-01-02"),
				Description:       item.Description,
				TotalAmount:       fmt.Sprintf("%.2f", item.TotalAmount.Float()),
				InstallmentNumber: item.InstallmentNumber,
				InstallmentTotal:  item.InstallmentTotal,
				InstallmentAmount: fmt.Sprintf("%.2f", item.InstallmentAmount.Float()),
				InstallmentLabel:  item.InstallmentLabel(),
				CreatedAt:         item.CreatedAt,
				UpdatedAt:         item.UpdatedAt.ValueOr(item.CreatedAt),
			}
		}

		output[i] = dtos.InvoiceOutput{
			ID:             invoice.ID.String(),
			UserID:         invoice.UserID.String(),
			CardID:         invoice.CardID.String(),
			ReferenceMonth: invoice.ReferenceMonth.String(),
			DueDate:        invoice.DueDate.Format("2006-01-02"),
			TotalAmount:    fmt.Sprintf("%.2f", invoice.TotalAmount.Float()),
			Currency:       string(invoice.TotalAmount.Currency()),
			ItemCount:      len(invoice.Items),
			Items:          items,
			CreatedAt:      invoice.CreatedAt,
			UpdatedAt:      invoice.UpdatedAt.ValueOr(invoice.CreatedAt),
		}
	}

	return &ListInvoicesByCardPaginatedOutput{
		Invoices:   output,
		NextCursor: nextCursor,
	}, nil
}
