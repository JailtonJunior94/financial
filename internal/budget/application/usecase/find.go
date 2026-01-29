package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
)

type (
	FindBudgetUseCase interface {
		Execute(ctx context.Context, budgetID string) (*dtos.BudgetOutput, error)
	}

	findBudgetUseCase struct {
		budgetRepository interfaces.BudgetRepository
		o11y             observability.Observability
	}
)

func NewFindBudgetUseCase(
	budgetRepository interfaces.BudgetRepository,
	o11y observability.Observability,
) FindBudgetUseCase {
	return &findBudgetUseCase{
		budgetRepository: budgetRepository,
		o11y:             o11y,
	}
}

func (u *findBudgetUseCase) Execute(ctx context.Context, budgetID string) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_budget_usecase.execute")
	defer span.End()

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return nil, fmt.Errorf("invalid budget_id: %w", err)
	}

	// Find budget
	budget, err := u.budgetRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if budget == nil {
		return nil, domain.ErrBudgetNotFound
	}

	// Build items output
	items := make([]dtos.BudgetItemOutput, len(budget.Items))
	for i, item := range budget.Items {
		items[i] = dtos.BudgetItemOutput{
			ID:              item.ID.String(),
			BudgetID:        item.BudgetID.String(),
			CategoryID:      item.CategoryID.String(),
			PercentageGoal:  item.PercentageGoal.String(),
			PlannedAmount:   item.PlannedAmount.String(),
			SpentAmount:     item.SpentAmount.String(),
			RemainingAmount: item.RemainingAmount().String(),
			PercentageSpent: item.PercentageSpent().String(),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt.ValueOr(time.Time{}),
		}
	}

	// Build budget output
	return &dtos.BudgetOutput{
		ID:             budget.ID.String(),
		UserID:         budget.UserID.String(),
		ReferenceMonth: budget.ReferenceMonth.String(),
		TotalAmount:    budget.TotalAmount.String(),
		SpentAmount:    budget.SpentAmount.String(),
		PercentageUsed: budget.PercentageUsed.String(),
		Currency:       string(budget.TotalAmount.Currency()),
		Items:          items,
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      budget.UpdatedAt.ValueOr(time.Time{}),
	}, nil
}
