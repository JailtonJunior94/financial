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

const (
	defaultInvoiceLimit = 20
	maxInvoiceLimit     = 100
)

type (
	ListInvoicesByCardPaginatedUseCase interface {
		Execute(ctx context.Context, input ListInvoicesByCardPaginatedInput) (*ListInvoicesByCardPaginatedOutput, error)
	}

	ListInvoicesByCardPaginatedInput struct {
		UserID string
		CardID string
		Status string // optional filter: "" | "open" | "closed" | "paid"
		Limit  int
		Cursor string
	}

	ListInvoicesByCardPaginatedOutput struct {
		Invoices   []dtos.InvoiceOutput
		NextCursor *string
	}

	listInvoicesByCardPaginatedUseCase struct {
		invoiceRepository interfaces.InvoiceRepository
		o11y              observability.Observability
	}
)

func NewListInvoicesByCardPaginatedUseCase(
	invoiceRepository interfaces.InvoiceRepository,
	o11y observability.Observability,
) ListInvoicesByCardPaginatedUseCase {
	return &listInvoicesByCardPaginatedUseCase{
		invoiceRepository: invoiceRepository,
		o11y:              o11y,
	}
}

func (u *listInvoicesByCardPaginatedUseCase) Execute(
	ctx context.Context,
	input ListInvoicesByCardPaginatedInput,
) (*ListInvoicesByCardPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_invoices_by_card_paginated_usecase.execute")
	defer span.End()
	limit := clampLimit(input.Limit, defaultInvoiceLimit, maxInvoiceLimit)
	userID, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		return nil, err
	}
	cardID, err := vos.NewUUIDFromString(input.CardID)
	if err != nil {
		return nil, err
	}
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}
	invoices, err := u.invoiceRepository.ListByCard(ctx, interfaces.ListInvoicesByCardParams{
		UserID: userID,
		CardID: cardID,
		Status: input.Status,
		Limit:  limit + 1,
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}
	hasNext := len(invoices) > limit
	if hasNext {
		invoices = invoices[:limit]
	}
	var nextCursor *string
	if hasNext && len(invoices) > 0 {
		lastInvoice := invoices[len(invoices)-1]
		newCursor := pagination.Cursor{
			Fields: map[string]any{
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

func clampLimit(limit, defaultVal, maxVal int) int {
	if limit <= 0 {
		return defaultVal
	}
	if limit > maxVal {
		return maxVal
	}
	return limit
}
