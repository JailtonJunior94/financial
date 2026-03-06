package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type ListCategoriesParams struct {
	UserID vos.UUID
	Limit  int
	Cursor pagination.Cursor
}

type CategoryRepository interface {
	ListPaginated(ctx context.Context, params ListCategoriesParams) ([]*entities.Category, error)
	FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error)
	Save(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error
	SoftDelete(ctx context.Context, id vos.UUID) error
}
