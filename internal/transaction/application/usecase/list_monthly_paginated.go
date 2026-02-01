package usecase

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type (
	// ListMonthlyPaginatedUseCase lista monthly transactions de um usuário com paginação cursor-based.
	ListMonthlyPaginatedUseCase interface {
		Execute(ctx context.Context, input ListMonthlyPaginatedInput) (*ListMonthlyPaginatedOutput, error)
	}

	// ListMonthlyPaginatedInput representa a entrada do use case.
	ListMonthlyPaginatedInput struct {
		UserID string
		Limit  int
		Cursor string
	}

	// ListMonthlyPaginatedOutput representa a saída do use case.
	ListMonthlyPaginatedOutput struct {
		MonthlyTransactions []*dtos.MonthlyTransactionOutput
		NextCursor          *string
	}

	listMonthlyPaginatedUseCase struct {
		o11y       observability.Observability
		repository interfaces.TransactionRepository
	}
)

// NewListMonthlyPaginatedUseCase cria uma nova instância do use case.
func NewListMonthlyPaginatedUseCase(
	o11y observability.Observability,
	repository interfaces.TransactionRepository,
) ListMonthlyPaginatedUseCase {
	return &listMonthlyPaginatedUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

// Execute executa o use case de listagem paginada de monthly transactions.
func (u *listMonthlyPaginatedUseCase) Execute(
	ctx context.Context,
	input ListMonthlyPaginatedInput,
) (*ListMonthlyPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_monthly_paginated_usecase.execute")
	defer span.End()

	// Parse user ID
	userID, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		return nil, err
	}

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}

	// List monthly transactions (paginado)
	monthlyList, err := u.repository.ListMonthlyPaginated(ctx, interfaces.ListMonthlyParams{
		UserID: userID,
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(monthlyList) > input.Limit
	if hasNext {
		monthlyList = monthlyList[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(monthlyList) > 0 {
		lastMonthly := monthlyList[len(monthlyList)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"date": lastMonthly.ReferenceMonth.String(),
				"id":   lastMonthly.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	output := make([]*dtos.MonthlyTransactionOutput, len(monthlyList))
	for i, monthly := range monthlyList {
		items := make([]*dtos.TransactionItemOutput, 0)
		for _, item := range monthly.Items {
			// Ignorar itens deletados
			if item.IsDeleted() {
				continue
			}

			items = append(items, &dtos.TransactionItemOutput{
				ID:          item.ID.String(),
				CategoryID:  item.CategoryID.String(),
				Title:       item.Title,
				Description: item.Description,
				Amount:      item.Amount.String(),
				Direction:   item.Direction.String(),
				Type:        item.Type.String(),
				IsPaid:      item.IsPaid,
				CreatedAt:   item.CreatedAt.ValueOr(time.Time{}),
				UpdatedAt:   item.UpdatedAt.ValueOr(time.Time{}),
			})
		}

		output[i] = &dtos.MonthlyTransactionOutput{
			ID:             monthly.ID.String(),
			ReferenceMonth: monthly.ReferenceMonth.String(),
			TotalIncome:    monthly.TotalIncome.String(),
			TotalExpense:   monthly.TotalExpense.String(),
			TotalAmount:    monthly.TotalAmount.String(),
			Items:          items,
			CreatedAt:      monthly.CreatedAt.ValueOr(time.Time{}),
			UpdatedAt:      monthly.UpdatedAt.ValueOr(time.Time{}),
		}
	}

	return &ListMonthlyPaginatedOutput{
		MonthlyTransactions: output,
		NextCursor:          nextCursor,
	}, nil
}
