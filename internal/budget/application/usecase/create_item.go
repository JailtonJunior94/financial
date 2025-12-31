package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateBudgetItemUseCase interface {
		Execute(ctx context.Context, budgetID string, input *dtos.BudgetItemInput) error
	}

	createBudgetItemUseCase struct {
		o11y             observability.Observability
		budgetRepository interfaces.BudgetRepository
	}
)

func NewCreateBudgetItemUseCase(
	o11y observability.Observability,
	budgetRepository interfaces.BudgetRepository,
) CreateBudgetItemUseCase {
	return &createBudgetItemUseCase{
		o11y:             o11y,
		budgetRepository: budgetRepository,
	}
}

func (u *createBudgetItemUseCase) Execute(ctx context.Context, budgetID string, input *dtos.BudgetItemInput) error {
	// TODO: Implement create budget item use case
	// This should:
	// 1. Find the budget by ID
	// 2. Create a new BudgetItem entity
	// 3. Call budget.AddItem() to validate and add
	// 4. Persist the item
	return nil
}
