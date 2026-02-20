package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// ListBudgetsParams representa os parâmetros para paginação de budgets.
type ListBudgetsParams struct {
	UserID vos.UUID
	Limit  int
	Cursor pagination.Cursor
}

type BudgetRepository interface {
	Insert(ctx context.Context, budget *entities.Budget) error
	InsertItems(ctx context.Context, items []*entities.BudgetItem) error
	FindByID(ctx context.Context, userID vos.UUID, id vos.UUID) (*entities.Budget, error)
	FindByUserIDAndReferenceMonth(ctx context.Context, userID vos.UUID, referenceMonth pkgVos.ReferenceMonth) (*entities.Budget, error)
	ListPaginated(ctx context.Context, params ListBudgetsParams) ([]*entities.Budget, error)
	Update(ctx context.Context, budget *entities.Budget) error
	UpdateItem(ctx context.Context, item *entities.BudgetItem) error
	Delete(ctx context.Context, id vos.UUID) error
}
