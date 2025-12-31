package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
)

type BudgetRepository interface {
	Insert(ctx context.Context, budget *entities.Budget) error
	InsertItems(ctx context.Context, items []*entities.BudgetItem) error
	FindByID(ctx context.Context, id vos.UUID) (*entities.Budget, error)
	FindByUserIDAndReferenceMonth(ctx context.Context, userID vos.UUID, referenceMonth budgetVos.ReferenceMonth) (*entities.Budget, error)
	Update(ctx context.Context, budget *entities.Budget) error
	UpdateItem(ctx context.Context, item *entities.BudgetItem) error
}
