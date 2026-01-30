package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type (
	// ListBudgetsPaginatedUseCase lista budgets de um usuário com paginação cursor-based.
	ListBudgetsPaginatedUseCase interface {
		Execute(ctx context.Context, input ListBudgetsPaginatedInput) (*ListBudgetsPaginatedOutput, error)
	}

	// ListBudgetsPaginatedInput representa a entrada do use case.
	ListBudgetsPaginatedInput struct {
		UserID string
		Limit  int
		Cursor string
	}

	// ListBudgetsPaginatedOutput representa a saída do use case.
	ListBudgetsPaginatedOutput struct {
		Budgets    []*dtos.BudgetOutput
		NextCursor *string
	}

	listBudgetsPaginatedUseCase struct {
		o11y       observability.Observability
		repository interfaces.BudgetRepository
	}
)

// NewListBudgetsPaginatedUseCase cria uma nova instância do use case.
func NewListBudgetsPaginatedUseCase(
	o11y observability.Observability,
	repository interfaces.BudgetRepository,
) ListBudgetsPaginatedUseCase {
	return &listBudgetsPaginatedUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

// Execute executa o use case de listagem paginada de budgets.
func (u *listBudgetsPaginatedUseCase) Execute(
	ctx context.Context,
	input ListBudgetsPaginatedInput,
) (*ListBudgetsPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_budgets_paginated_usecase.execute")
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

	// List budgets (paginado)
	budgets, err := u.repository.ListPaginated(ctx, interfaces.ListBudgetsParams{
		UserID: userID,
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(budgets) > input.Limit
	if hasNext {
		budgets = budgets[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(budgets) > 0 {
		lastBudget := budgets[len(budgets)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"date": lastBudget.ReferenceMonth.ToTime().Format("2006-01-02"),
				"id":   lastBudget.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	output := make([]*dtos.BudgetOutput, len(budgets))
	for i, budget := range budgets {
		items := make([]dtos.BudgetItemOutput, len(budget.Items))
		for j, item := range budget.Items {
			items[j] = dtos.BudgetItemOutput{
				ID:              item.ID.String(),
				BudgetID:        item.BudgetID.String(),
				CategoryID:      item.CategoryID.String(),
				PercentageGoal:  item.PercentageGoal.String(),
				PlannedAmount:   item.PlannedAmount.String(),
				SpentAmount:     item.SpentAmount.String(),
				RemainingAmount: item.RemainingAmount().String(),
				PercentageSpent: item.PercentageSpent().String(),
				CreatedAt:       item.CreatedAt,
				UpdatedAt:       item.UpdatedAt.ValueOr(item.CreatedAt),
			}
		}

		output[i] = &dtos.BudgetOutput{
			ID:             budget.ID.String(),
			UserID:         budget.UserID.String(),
			ReferenceMonth: budget.ReferenceMonth.String(),
			TotalAmount:    budget.TotalAmount.String(),
			SpentAmount:    budget.SpentAmount.String(),
			PercentageUsed: budget.PercentageUsed.String(),
			Currency:       string(budget.TotalAmount.Currency()),
			Items:          items,
			CreatedAt:      budget.CreatedAt,
			UpdatedAt:      budget.UpdatedAt.ValueOr(budget.CreatedAt),
		}
	}

	return &ListBudgetsPaginatedOutput{
		Budgets:    output,
		NextCursor: nextCursor,
	}, nil
}
