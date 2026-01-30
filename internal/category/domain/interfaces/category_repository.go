package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// ListCategoriesParams representa os parâmetros para paginação de categorias.
type ListCategoriesParams struct {
	UserID vos.UUID
	Limit  int
	Cursor pagination.Cursor
}

type CategoryRepository interface {
	List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error)
	ListPaginated(ctx context.Context, params ListCategoriesParams) ([]*entities.Category, error)
	FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error)
	Save(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error
	CheckCycleExists(ctx context.Context, userID, categoryID, parentID vos.UUID) (bool, error)
}
