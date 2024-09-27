package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
)

type BudgetRepository interface {
	Insert(ctx context.Context, budget *entities.Budget) error
	InsertItems(ctx context.Context, items []*entities.BudgetItem) error
}
