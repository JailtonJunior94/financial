package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	CreateBudgetItemUseCase interface {
		Execute(ctx context.Context, userID string, budgetID, input *dtos.BudgetItem) (*dtos.BudgetOutput, error)
	}

	createBudgetItemUseCase struct {
		o11y             o11y.Observability
		budgetRepository interfaces.BudgetRepository
	}
)

func NewCreateBudgetItemUseCase(
	o11y o11y.Observability,
	budgetRepository interfaces.BudgetRepository,
) CreateBudgetItemUseCase {
	return &createBudgetItemUseCase{
		o11y:             o11y,
		budgetRepository: budgetRepository,
	}
}

func (u *createBudgetItemUseCase) Execute(ctx context.Context, userID string, budgetID, input *dtos.BudgetItem) (*dtos.BudgetOutput, error) {
	return nil, nil
}
